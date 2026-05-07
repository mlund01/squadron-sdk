package squadron

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/invopop/jsonschema"
)

type Handler[I, O any] func(ctx context.Context, in I) (O, error)

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

func reflectSchema[I any]() (json.RawMessage, error) {
	r := &jsonschema.Reflector{
		Anonymous:      true,
		ExpandedStruct: true,
	}
	var zero I
	schema := r.Reflect(zero)
	// Strip $schema and $id; LLM tool callers don't use them and they add noise.
	schema.Version = ""
	schema.ID = ""
	return json.Marshal(schema)
}
