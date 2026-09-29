package tui

import (
	"context"
	"fmt"
	"time"
	"unicode"
	"unicode/utf8"

	"tui-app-launcher/internal/errors"
	"tui-app-launcher/internal/interfaces"
	"tui-app-launcher/internal/logging"

	tea "github.com/charmbracelet/bubbletea"
)

// update handles application events; Update adds presentation-only animation.
func (m Model) update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height

		// Ensure minimum dimensions
		if m.width < m.getMinWidth() {
			m.width = m.getMinWidth()
		}
		if m.height < m.getMinHeight() {
			m.height = m.getMinHeight()
		}

		// Resizing must preserve the selected application, not select another row.
		m.ensureSelectionVisible()

		return m, nil

	case initMsg:
		m.updateFilteredApps()
		if m.scanner != nil {
			m.isRefreshing = true
			return m, m.refreshApplications()
		}
		return m, nil

	case launchMsg:
		// Validate application before launching
		if err := m.launcher.ValidateApplication(msg.app); err != nil {
			return m, func() tea.Msg {
				return launchErrorMsg{err: err}
			}
		}

		// Launch the application
		err := m.launcher.LaunchApplication(msg.app)
		if err != nil {
			return m, func() tea.Msg {
				return launchErrorMsg{err: err}
			}
		}
		if history, ok := m.configManager.(interfaces.LaunchHistoryManager); ok {
			if err := history.RecordLaunch(msg.app); err != nil {
				// The app has already started: a history write must not invite a
				// second launch or turn success into a launch failure.
				logging.Error("Failed to record launch history: %v", err)
			}
		}
		return m, func() tea.Msg {
			return launchSuccessMsg{}
		}

	case launchSuccessMsg:
		// Application launched successfully, exit the launcher
		if m.cancelScan != nil {
			m.cancelScan()
		}
		return m, tea.Quit

	case launchErrorMsg:
		// Use error recovery system to handle the error gracefully
		recovery := m.errorRecovery.RecoverFromError(msg.err)

		// Check if it's a structured launcher error
		if launcherErr, ok := msg.err.(*errors.LauncherError); ok {
			m.errorMessage = launcherErr.GetUserFriendlyMessage()
		} else {
			m.errorMessage = recovery.Message
		}

		// If the error is not recoverable, we should exit
		if !recovery.CanContinue {
			if m.cancelScan != nil {
				m.cancelScan()
			}
			return m, tea.Quit
		}

		return m, nil

	case toggleFavoriteMsg:
		// Toggle favorite status
		app := msg.app
		// Resolve the current favorite state at delivery time, not keypress time.
		// Other updates may have rebuilt the filtered slice meanwhile.
		for _, current := range m.applications {
			if current.Name == app.Name {
				app.IsFavorite = current.IsFavorite
				break
			}
		}
		var err error

		if app.IsFavorite {
			// Remove from favorites
			err = m.configManager.RemoveFavorite(app.Name)
			if err == nil {
				app.IsFavorite = false
				// Update the corresponding app in the main applications list
				for i := range m.applications {
					if m.applications[i].Name == app.Name {
						m.applications[i].IsFavorite = false
						break
					}
				}
			}
		} else {
			// Add to favorites
			err = m.configManager.AddFavorite(app.Name)
			if err == nil {
				app.IsFavorite = true
				// Update the corresponding app in the main applications list
				for i := range m.applications {
					if m.applications[i].Name == app.Name {
						m.applications[i].IsFavorite = true
						break
					}
				}
			}
		}

		// Handle configuration errors gracefully
		if err != nil {
			recovery := m.errorRecovery.RecoverFromError(err)
			if launcherErr, ok := err.(*errors.LauncherError); ok {
				m.errorMessage = fmt.Sprintf("Favorites: %s", launcherErr.GetUserFriendlyMessage())
			} else {
				m.errorMessage = fmt.Sprintf("Favorites: %s", recovery.Message)
			}
		}

		// Update filtered apps if we're showing favorites
		if m.searchQuery == "" {
			m.updateFilteredApps()
		}

		return m, nil

	case refreshCompleteMsg:
		if m.cancelScan != nil {
			m.cancelScan()
			m.cancelScan = nil
		}
		m.isRefreshing = false
		if msg.err != nil {
			// Handle refresh error
			recovery := m.errorRecovery.RecoverFromError(msg.err)
			if launcherErr, ok := msg.err.(*errors.LauncherError); ok {
				m.errorMessage = fmt.Sprintf("Refresh failed: %s", launcherErr.GetUserFriendlyMessage())
			} else {
				m.errorMessage = fmt.Sprintf("Refresh failed: %s", recovery.Message)
			}
		} else {
			// Successfully refreshed applications
			m.SetApplications(msg.apps)
			m.errorMessage = ""
		}
		return m, nil

	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "esc":
			if m.cancelScan != nil {
				m.cancelScan()
			}
			return m, tea.Quit

		case "?":
			m.showHelp = !m.showHelp
			return m, nil

		case "up":
			if len(m.filteredApps) > 0 {
				m.selectedIndex--
				if m.selectedIndex < 0 {
					m.selectedIndex = len(m.filteredApps) - 1
				}
			}
			m.ensureSelectionVisible()
			return m, nil

		case "down":
			if len(m.filteredApps) > 0 {
				m.selectedIndex++
				if m.selectedIndex >= len(m.filteredApps) {
					m.selectedIndex = 0
				}
			}
			m.ensureSelectionVisible()
			return m, nil

		case "enter":
			if len(m.filteredApps) > 0 && m.selectedIndex < len(m.filteredApps) {
				selectedApp := m.filteredApps[m.selectedIndex]
				return m, m.launchApplication(selectedApp)
			}
			return m, nil

		case "tab":
			if len(m.filteredApps) > 0 && m.selectedIndex < len(m.filteredApps) {
				selectedApp := m.filteredApps[m.selectedIndex]
				return m, m.toggleFavorite(selectedApp)
			}
			return m, nil

		case "backspace":
			if len(m.searchQuery) > 0 {
				_, size := utf8.DecodeLastRuneInString(m.searchQuery)
				m.searchQuery = m.searchQuery[:len(m.searchQuery)-size]
				m.updateFilteredApps()
			}
			return m, nil

		case "ctrl+u":
			// Clear search query
			m.searchQuery = ""
			m.updateFilteredApps()
			return m, nil

		case "ctrl+l":
			// Clear error message
			m.errorMessage = ""
			return m, nil

		case "ctrl+r":
			// Refresh application list
			if !m.isRefreshing {
				m.isRefreshing = true
				m.errorMessage = ""
				return m, m.refreshApplications()
			}
			return m, nil

		default:
			if msg.Type == tea.KeyRunes {
				for _, r := range msg.Runes {
					if unicode.IsPrint(r) {
						m.searchQuery += string(r)
						m.updateFilteredApps()
					}
				}
			}
			return m, nil
		}
	}

	return m, nil
}

// launchApplication creates a command to launch the selected application
func (m *Model) launchApplication(app interfaces.Application) tea.Cmd {
	return func() tea.Msg {
		return launchMsg{app: app}
	}
}

// toggleFavorite creates a command to toggle favorite status
func (m *Model) toggleFavorite(app interfaces.Application) tea.Cmd {
	return func() tea.Msg {
		return toggleFavoriteMsg{app: app}
	}
}

// refreshApplications starts a bounded scan without blocking the TUI on slow I/O.
func (m *Model) refreshApplications() tea.Cmd {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	m.cancelScan = cancel
	scanner := m.scanner
	return func() tea.Msg {
		results := make(chan refreshCompleteMsg, 1)
		go func() {
			apps, err := scanner.ScanApplicationsContext(ctx)
			results <- refreshCompleteMsg{apps: apps, err: err}
		}()
		select {
		case msg := <-results:
			return msg
		case <-ctx.Done():
			return refreshCompleteMsg{err: ctx.Err()}
		}
	}
}
