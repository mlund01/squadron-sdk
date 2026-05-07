package squadron

import (
	"context"
	"fmt"
)

// App is the high-level, Go-generics-based plugin API.
//
// Construct one with New(), register tools with Tool[I, O], optionally set a
// configure handler with Configure, and call Serve() from main:
//
//	app := squadron.New()
//	squadron.Tool(app, "echo", "Echo a message",
//	    func(ctx context.Context, in EchoInput) (string, error) {
//	        return in.Message, nil
//	    })
//	app.Serve()
//
// App is a convenience layer over the lower-level ToolProvider interface:
// AsProvider() returns a ToolProvider that can be passed to anything
// expecting one.
type App struct {
	tools    map[string]*registeredTool
	onConfig func(map[string]string) error
}

// New returns an empty App.
func New() *App {
	return &App{tools: make(map[string]*registeredTool)}
}

// Configure registers a function to receive settings from the host's HCL
// config. The handler may return an error to fail configuration.
func (a *App) Configure(fn func(settings map[string]string) error) {
	a.onConfig = fn
}

// AsProvider returns a ToolProvider backed by this App.
func (a *App) AsProvider() ToolProvider {
	return &appProvider{app: a}
}

// Serve starts the plugin server. Call from main.
func (a *App) Serve() {
	Serve(a.AsProvider())
}

type registeredTool struct {
	info    *ToolInfo
	handler func(ctx context.Context, payload string) (string, error)
}

type appProvider struct {
	app *App
}

func (p *appProvider) Configure(settings map[string]string) error {
	if p.app.onConfig == nil {
		return nil
	}
	return p.app.onConfig(settings)
}

func (p *appProvider) Call(ctx context.Context, toolName string, payload string) (string, error) {
	rt, ok := p.app.tools[toolName]
	if !ok {
		return "", fmt.Errorf("unknown tool: %s", toolName)
	}
	return rt.handler(ctx, payload)
}

func (p *appProvider) GetToolInfo(toolName string) (*ToolInfo, error) {
	rt, ok := p.app.tools[toolName]
	if !ok {
		return nil, fmt.Errorf("unknown tool: %s", toolName)
	}
	return rt.info, nil
}

func (p *appProvider) ListTools() ([]*ToolInfo, error) {
	out := make([]*ToolInfo, 0, len(p.app.tools))
	for _, rt := range p.app.tools {
		out = append(out, rt.info)
	}
	return out, nil
}
