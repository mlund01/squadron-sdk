package squadron

import (
	"os"
	"time"

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

// monitorParent checks if the parent process is still alive.
// If the parent dies (PPID becomes 1 on Unix), the plugin exits.
func monitorParent() {
	ppid := os.Getppid()
	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()

	for range ticker.C {
		currentPPID := os.Getppid()
		// On Unix, when the parent dies, PPID becomes 1 (init/launchd)
		if currentPPID != ppid || currentPPID == 1 {
			os.Exit(0)
		}
	}
}
