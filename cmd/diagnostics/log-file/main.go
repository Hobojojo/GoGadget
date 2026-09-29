// Standalone manual logging diagnostic: go run ./cmd/diagnostics/log-file.
package main

import (
	"fmt"
	"os"
	"path/filepath"
)

func main() {
	// Get user home directory
	homeDir, err := os.UserHomeDir()
	if err != nil {
		fmt.Printf("Error getting home directory: %v\n", err)
		return
	}

	// Create log directory
	logDir := filepath.Join(homeDir, ".config", "tui-launcher")
	fmt.Printf("Creating log directory: %s\n", logDir)

	if err := os.MkdirAll(logDir, 0755); err != nil {
		fmt.Printf("Error creating log directory: %v\n", err)
		return
	}

	// Create log file
	logPath := filepath.Join(logDir, "launcher.log")
	fmt.Printf("Creating log file: %s\n", logPath)

	logFile, err := os.OpenFile(logPath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		fmt.Printf("Error creating log file: %v\n", err)
		return
	}
	defer logFile.Close()

	// Write test log entry
	testEntry := "[2024-01-01 12:00:00] INFO: Test log entry\n"
	if _, err := logFile.WriteString(testEntry); err != nil {
		fmt.Printf("Error writing to log file: %v\n", err)
		return
	}

	fmt.Printf("✓ Log file created and written successfully\n")

	// Verify file exists and read content
	if content, err := os.ReadFile(logPath); err == nil {
		fmt.Printf("✓ Log file content: %s", string(content))
	} else {
		fmt.Printf("Error reading log file: %v\n", err)
	}
}
