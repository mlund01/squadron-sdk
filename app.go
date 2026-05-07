package squadron

import (
	"context"
	"fmt"
)

type App struct {
	tools    map[string]*registeredTool
	onConfig func(map[string]string) error
}

func New() *App {
	return &App{tools: make(map[string]*registeredTool)}
}

func (a *App) Configure(fn func(settings map[string]string) error) {
	a.onConfig = fn
}

func (a *App) AsProvider() ToolProvider {
	return &appProvider{app: a}
}

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
