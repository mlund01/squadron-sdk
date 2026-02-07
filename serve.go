package squad

import (
	"github.com/hashicorp/go-plugin"
)

// Serve starts the plugin server with the given ToolProvider implementation.
// This is the main entry point for plugin binaries.
func Serve(impl ToolProvider) {
	plugin.Serve(&plugin.ServeConfig{
		HandshakeConfig: Handshake,
		Plugins: map[string]plugin.Plugin{
			"tool": &ToolPluginGRPCPlugin{Impl: impl},
		},
		GRPCServer: plugin.DefaultGRPCServer,
	})
}
