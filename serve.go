package squadron

import (
	"github.com/hashicorp/go-plugin"
)

// Serve starts the plugin server with the given ToolProvider implementation.
// This is the main entry point for plugin binaries.
// It also monitors the parent process and exits if the parent dies,
// preventing orphaned plugin processes.
func Serve(impl ToolProvider) {
	// Monitor parent process - exit if parent dies
	go monitorParent()

	plugin.Serve(&plugin.ServeConfig{
		HandshakeConfig: Handshake,
		Plugins: map[string]plugin.Plugin{
			"tool": &ToolPluginGRPCPlugin{Impl: impl},
		},
		GRPCServer: plugin.DefaultGRPCServer,
	})
}
