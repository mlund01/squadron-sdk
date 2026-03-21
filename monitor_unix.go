//go:build !windows

package squadron

import (
	"os"
	"time"
)

// monitorParent checks if the parent process is still alive.
// On Unix, when the parent dies, PPID becomes 1 (init/launchd).
func monitorParent() {
	ppid := os.Getppid()
	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()

	for range ticker.C {
		currentPPID := os.Getppid()
		if currentPPID != ppid || currentPPID == 1 {
			os.Exit(0)
		}
	}
}
