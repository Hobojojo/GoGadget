package launcher

import (
	"fmt"
	"os"
	"os/exec"
)

// terminalEmulator selects an emulator with a -e command/argument interface.
func terminalEmulator() (string, error) {
	candidates := []string{}
	if preferred := os.Getenv("TERMINAL"); preferred != "" {
		candidates = append(candidates, preferred)
	}
	candidates = append(candidates, "x-terminal-emulator", "xterm")
	for _, candidate := range candidates {
		if path, err := exec.LookPath(candidate); err == nil {
			return path, nil
		}
	}
	return "", fmt.Errorf("set TERMINAL or install x-terminal-emulator or xterm")
}
