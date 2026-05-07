package squadron

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/invopop/jsonschema"
)

// Handler is the signature of a typed tool implementation.
type Handler[I, O any] func(ctx context.Context, in I) (O, error)

// Tool registers a typed tool on the App. The input schema is reflected from
// I via github.com/invopop/jsonschema (driven by `json:` and `jsonschema:`
// struct tags); the host receives the full JSON Schema verbatim, with no
// lossy projection. Use I = struct{} for tools that take no input.
//
// Example:
//
//	type EchoInput struct {
//	    Message string `json:"message" jsonschema:"required,description=Text to echo"`
//	    AllCaps bool   `json:"all_caps,omitempty"`
//	}
//
//	squadron.Tool(app, "echo", "Echoes back a message",
//	    func(ctx context.Context, in EchoInput) (string, error) {
//	        if in.AllCaps {
//	            return strings.ToUpper(in.Message), nil
//	        }
//	        return in.Message, nil
//	    })
func Tool[I, O any](app *App, name, description string, handler Handler[I, O]) {
	if _, exists := app.tools[name]; exists {
		panic(fmt.Sprintf("squadron: tool %q is already registered", name))
	}

	rawSchema, err := reflectSchema[I]()
	if err != nil {
		panic(fmt.Sprintf("squadron: reflecting schema for tool %q: %v", name, err))
	}

	app.tools[name] = &registeredTool{
		info: &ToolInfo{
			Name:        name,
			Description: description,
			RawSchema:   rawSchema,
		},
		handler: func(ctx context.Context, payload string) (string, error) {
			var in I
			if payload != "" {
				if err := json.Unmarshal([]byte(payload), &in); err != nil {
					return "", fmt.Errorf("invalid payload for %s: %w", name, err)
				}
			}
			out, err := handler(ctx, in)
			if err != nil {
				return "", err
			}
			b, err := json.Marshal(out)
			if err != nil {
				return "", fmt.Errorf("marshal output for %s: %w", name, err)
			}
			return string(b), nil
		},
	}
}

// reflectSchema produces a JSON Schema (as raw bytes) for the type I. We
// strip $schema and $id since LLM tool callers don't need them.
func reflectSchema[I any]() (json.RawMessage, error) {
	r := &jsonschema.Reflector{
		Anonymous:      true,
		ExpandedStruct: true,
	}
	var zero I
	schema := r.Reflect(zero)
	schema.Version = ""
	schema.ID = ""
	return json.Marshal(schema)
}
