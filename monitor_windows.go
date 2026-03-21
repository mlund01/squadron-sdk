//go:build windows

package squadron

import (
	"os"
	"time"

	"golang.org/x/sys/windows"
)

// monitorParent checks if the parent process is still alive.
// On Windows, we open a handle to the parent process and poll for its exit.
func monitorParent() {
	ppid := os.Getppid()

	handle, err := windows.OpenProcess(windows.SYNCHRONIZE, false, uint32(ppid))
	if err != nil {
		// Can't open parent process — fall back to PPID polling
		ticker := time.NewTicker(5 * time.Second)
		defer ticker.Stop()
		for range ticker.C {
			if os.Getppid() != ppid {
				os.Exit(0)
			}
		}
		return
	}
	defer windows.CloseHandle(handle)

	// Wait for the parent process to exit
	windows.WaitForSingleObject(handle, windows.INFINITE)
	os.Exit(0)
}
