package main

import (
	"fmt"
	"log"
	"os"

	"tui-app-launcher/internal/config"
	"tui-app-launcher/internal/launcher"
	"tui-app-launcher/internal/logging"
	"tui-app-launcher/internal/scanner"
	"tui-app-launcher/internal/search"
	"tui-app-launcher/internal/tui"

	tea "github.com/charmbracelet/bubbletea"
)

func main() {
	// Initialize logging first
	if err := logging.InitGlobalLogger(); err != nil {
		log.Printf("Failed to initialize logger: %v", err)
		// Continue without logging rather than failing
	}

	// Ensure logger is closed on exit
	defer func() {
		if err := logging.CloseGlobalLogger(); err != nil {
			log.Printf("Failed to close logger: %v", err)
		}
	}()

	// Log application startup
	logging.Info("TUI App Launcher starting up")

	// Initialize core components
	configManager := config.NewManager()
	fuzzySearcher := search.NewFuzzySearcher()
	appScanner := scanner.NewScanner()
	appLauncher := launcher.NewLauncher()

	// Create TUI model
	model := tui.NewModel(fuzzySearcher, configManager, appLauncher, appScanner)

	// Create Bubble Tea program
	program := tea.NewProgram(model, tea.WithAltScreen(), tea.WithMouseCellMotion())

	// Run the program
	logging.Info("Starting TUI interface")
	if _, err := program.Run(); err != nil {
		logging.Error("Error running TUI: %v", err)
		fmt.Printf("Error running TUI: %v\n", err)
		os.Exit(1)
	}

	// Log application shutdown
	logging.Info("TUI App Launcher shutting down")
}
