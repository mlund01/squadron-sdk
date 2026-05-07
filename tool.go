package squadron

import (
	"context"
	"encoding/json"
	"fmt"
	"reflect"

	"github.com/invopop/jsonschema"
)

type Handler[I, O any] func(ctx context.Context, in I) (O, error)

func Tool[I, O any](app *App, name, description string, handler Handler[I, O]) {
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

	app.tools[name] = &registeredTool{
		info: &ToolInfo{
			Name:         name,
			Description:  description,
			RawSchema:    rawSchema,
			OutputSchema: outputSchema,
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
