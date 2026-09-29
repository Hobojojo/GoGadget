//go:build ignore

// Standalone logging diagnostic: go run test_log_creation.go.
package main

import (
	"fmt"
	"os"
	"path/filepath"

	"tui-app-launcher/internal/logging"
)

func main() {
	// Initialize logging
	if err := logging.InitGlobalLogger(); err != nil {
		fmt.Printf("Failed to initialize logger: %v\n", err)
		os.Exit(1)
	}
	defer logging.CloseGlobalLogger()

	// Test various logging levels
	logging.Debug("This is a debug message")
	logging.Info("This is an info message")
	logging.Warn("This is a warning message")
	logging.Error("This is an error message")

	// Test structured error logging
	testErr := fmt.Errorf("test error")
	context := map[string]interface{}{
		"component": "test",
		"operation": "logging_test",
	}
	logging.LogError(testErr, context)

	// Get logger and test specific methods
	logger := logging.GetGlobalLogger()
	if logger != nil {
		logger.LogApplicationScan("/test/directory", 5, 0)
		logger.LogApplicationLaunch("test-app", "test-command", true, nil)
		logger.LogConfigOperation("test_operation", true, nil)
		logger.LogStartup()
		logger.LogShutdown()
	}

	// Check if log file was created
	homeDir, _ := os.UserHomeDir()
	logPath := filepath.Join(homeDir, ".config", "tui-launcher", "launcher.log")
	
	if _, err := os.Stat(logPath); err == nil {
		fmt.Printf("✓ Log file created successfully at: %s\n", logPath)
		
		// Read and display log contents
		content, err := os.ReadFile(logPath)
		if err == nil {
			fmt.Printf("✓ Log file contents:\n%s\n", string(content))
		}
	} else {
		fmt.Printf("✗ Log file not found at: %s\n", logPath)
	}
}