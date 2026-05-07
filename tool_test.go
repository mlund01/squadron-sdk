package squadron

import (
	"context"
	"encoding/json"
	"strings"
	"testing"
)

type echoInput struct {
	Message string `json:"message" jsonschema:"required,description=Text to echo"`
	AllCaps bool   `json:"all_caps,omitempty" jsonschema:"description=Capitalize"`
}

func TestToolRegistersAndCalls(t *testing.T) {
	app := New()
	Tool(app, "echo", "Echo a message",
		func(ctx context.Context, in echoInput) (string, error) {
			if in.AllCaps {
				return strings.ToUpper(in.Message), nil
			}
			return in.Message, nil
		})

	provider := app.AsProvider()

	tools, err := provider.ListTools()
	if err != nil {
		t.Fatalf("ListTools: %v", err)
	}
	if len(tools) != 1 || tools[0].Name != "echo" {
		t.Fatalf("expected one echo tool, got %+v", tools)
	}

	if len(tools[0].RawSchema) == 0 {
		t.Fatal("expected RawSchema to be populated")
	}

	var schema map[string]any
	if err := json.Unmarshal(tools[0].RawSchema, &schema); err != nil {
		t.Fatalf("schema unmarshal: %v", err)
	}
	props, ok := schema["properties"].(map[string]any)
	if !ok {
		t.Fatalf("schema missing properties: %v", schema)
	}
	if _, ok := props["message"]; !ok {
		t.Fatalf("schema missing message property: %v", schema)
	}
	required, _ := schema["required"].([]any)
	if len(required) != 1 || required[0] != "message" {
		t.Fatalf("expected message required, got %v", required)
	}

	result, err := provider.Call(context.Background(), "echo",
		`{"message":"hi","all_caps":true}`)
	if err != nil {
		t.Fatalf("Call: %v", err)
	}
	var got string
	if err := json.Unmarshal([]byte(result), &got); err != nil {
		t.Fatalf("decode result: %v", err)
	}
	if got != "HI" {
		t.Fatalf("got %q, want HI", got)
	}
}

func TestToolEmptyInput(t *testing.T) {
	app := New()
	Tool(app, "ping", "Returns pong",
		func(ctx context.Context, _ struct{}) (string, error) {
			return "pong", nil
		})

	out, err := app.AsProvider().Call(context.Background(), "ping", "")
	if err != nil {
		t.Fatalf("Call: %v", err)
	}
	if out != `"pong"` {
		t.Fatalf("got %q, want \"pong\"", out)
	}
}

func TestConfigureHandler(t *testing.T) {
	app := New()
	var captured map[string]string
	app.Configure(func(settings map[string]string) error {
		captured = settings
		return nil
	})

	if err := app.AsProvider().Configure(map[string]string{"k": "v"}); err != nil {
		t.Fatalf("Configure: %v", err)
	}
	if captured["k"] != "v" {
		t.Fatalf("expected handler to receive settings, got %v", captured)
	}
}

func TestRawSchemaWinsOverTypedOnSerialize(t *testing.T) {
	info := &ToolInfo{
		Name:        "test",
		Description: "",
		Schema: Schema{
			Type:       TypeObject,
			Properties: PropertyMap{"x": {Type: TypeString}},
		},
		RawSchema: json.RawMessage(`{"type":"object","custom":true}`),
	}
	pb := toolInfoToProto(info)
	if !strings.Contains(pb.SchemaJson, `"custom":true`) {
		t.Fatalf("RawSchema should win, got %q", pb.SchemaJson)
	}
}
