package squadron

import (
	"context"
	"encoding/json"
	"fmt"
	"reflect"

	"github.com/invopop/jsonschema"
)

type Handler[I, O any] func(ctx context.Context, in I) (O, error)

type ToolOption func(*ToolInfo)

// Idempotent declares that replaying a call with the same invocation identity
// returns the original effect/result without producing another side effect.
func Idempotent() ToolOption {
	return func(info *ToolInfo) { info.Idempotent = true }
}

func Tool[I, O any](app *App, name, description string, handler Handler[I, O], options ...ToolOption) {
	if _, exists := app.tools[name]; exists {
		panic(fmt.Sprintf("squadron: tool %q is already registered", name))
	}

	rawSchema, err := reflectSchema[I]()
	if err != nil {
		panic(fmt.Sprintf("squadron: reflecting input schema for tool %q: %v", name, err))
	}

	outputSchema, err := reflectSchema[O]()
	if err != nil {
		panic(fmt.Sprintf("squadron: reflecting output schema for tool %q: %v", name, err))
	}

	info := &ToolInfo{
		Name:         name,
		Description:  description,
		RawSchema:    rawSchema,
		OutputSchema: outputSchema,
	}
	for _, option := range options {
		option(info)
	}
	app.tools[name] = &registeredTool{
		info: info,
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
			return marshalOutput(out)
		},
	}
}

func marshalOutput[O any](out O) (string, error) {
	if s, ok := any(out).(string); ok {
		return s, nil
	}
	b, err := json.Marshal(out)
	if err != nil {
		return "", fmt.Errorf("marshal output: %w", err)
	}
	return string(b), nil
}

func reflectSchema[T any]() (json.RawMessage, error) {
	var zero T
	if reflect.TypeOf(zero) == nil {
		return nil, nil
	}
	r := &jsonschema.Reflector{
		Anonymous:      true,
		ExpandedStruct: true,
	}
	schema := r.Reflect(zero)
	schema.Version = ""
	schema.ID = ""
	return json.Marshal(schema)
}
