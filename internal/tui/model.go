package tui

import (
	"fmt"
	"strings"
	"tui-app-launcher/internal/errors"
	"tui-app-launcher/internal/interfaces"

	tea "github.com/charmbracelet/bubbletea"
)

// Model represents the TUI application state
type Model struct {
	applications  []interfaces.Application
	filteredApps  []interfaces.Application
	searchQuery   string
	selectedIndex int
	scrollOffset  int
	fuzzySearcher interfaces.FuzzySearcher
	configManager interfaces.ConfigManager
	launcher      interfaces.ApplicationLauncher
	scanner       interfaces.ApplicationScanner
	showHelp      bool
	width         int
	height        int
	errorMessage  string
	errorRecovery *errors.ErrorRecovery
	isRefreshing  bool
}

// NewModel creates a new TUI model
func NewModel(fuzzySearcher interfaces.FuzzySearcher, configManager interfaces.ConfigManager, launcher interfaces.ApplicationLauncher, scanner interfaces.ApplicationScanner) Model {
	return Model{
		applications:  make([]interfaces.Application, 0),
		filteredApps:  make([]interfaces.Application, 0),
		searchQuery:   "",
		selectedIndex: 0,
		fuzzySearcher: fuzzySearcher,
		configManager: configManager,
		launcher:      launcher,
		scanner:       scanner,
		showHelp:      false,
		width:         80,
		height:        24,
		errorMessage:  "",
		errorRecovery: errors.NewErrorRecovery(),
		isRefreshing:  false,
	}
}

// getMinWidth returns the minimum width required for the interface
func (m *Model) getMinWidth() int {
	return 40 // Minimum width to display interface properly
}

// getMinHeight returns the minimum height required for the interface
func (m *Model) getMinHeight() int {
	return 10 // Minimum height to display header, search, one item, and footer
}

// truncateText truncates text to fit within the specified width, adding ellipsis if needed
func (m *Model) truncateText(text string, maxWidth int) string {
	if maxWidth <= 0 {
		return ""
	}

	runes := []rune(text)
	if len(runes) <= maxWidth {
		return text
	}

	if maxWidth <= 3 {
		return strings.Repeat(".", maxWidth)
	}

	return string(runes[:maxWidth-3]) + "..."
}

// wrapText wraps text to fit within the specified width
func (m *Model) wrapText(text string, maxWidth int) []string {
	if maxWidth <= 0 {
		return []string{}
	}

	words := strings.Fields(text)
	if len(words) == 0 {
		return []string{""}
	}

	var lines []string
	var currentLine strings.Builder

	for _, word := range words {
		// If adding this word would exceed the width, start a new line
		if currentLine.Len() > 0 && currentLine.Len()+1+len(word) > maxWidth {
			lines = append(lines, currentLine.String())
			currentLine.Reset()
		}

		// If the word itself is too long, truncate it
		if len(word) > maxWidth {
			word = m.truncateText(word, maxWidth)
		}

		if currentLine.Len() > 0 {
			currentLine.WriteString(" ")
		}
		currentLine.WriteString(word)
	}

	if currentLine.Len() > 0 {
		lines = append(lines, currentLine.String())
	}

	return lines
}

// getMaxAppNameWidth calculates the maximum width available for application names
func (m *Model) getMaxAppNameWidth() int {
	// Account for: border (2) + prefix (2) + favorite star (2) + padding (2)
	return m.width - 8
}

// SetApplications sets the applications list and updates favorites status
func (m *Model) SetApplications(apps []interfaces.Application) {
	// Update favorite status based on config
	for i := range apps {
		apps[i].IsFavorite = m.configManager.IsFavorite(apps[i].Name)
	}

	m.applications = apps
	m.fuzzySearcher.SetItems(apps)
	m.updateFilteredApps()
}

// Init initializes the TUI model
func (m Model) Init() tea.Cmd {
	// Load favorites and set up initial state
	return tea.Batch(
		tea.EnterAltScreen,
		func() tea.Msg {
			return initMsg{}
		},
	)
}

// initMsg is sent when the model is initialized
type initMsg struct{}

// Update handles TUI events and updates
func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
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
		// Initialize with favorites or all applications
		m.updateFilteredApps()
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
		return m, func() tea.Msg {
			return launchSuccessMsg{}
		}

	case launchSuccessMsg:
		// Application launched successfully, exit the launcher
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
			return m, tea.Quit
		}

		return m, nil

	case toggleFavoriteMsg:
		// Toggle favorite status
		app := msg.app
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
				selectedApp := &m.filteredApps[m.selectedIndex]
				return m, m.toggleFavorite(selectedApp)
			}
			return m, nil

		case "backspace":
			if len(m.searchQuery) > 0 {
				m.searchQuery = m.searchQuery[:len(m.searchQuery)-1]
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
			// Handle regular character input for search
			if len(msg.String()) == 1 {
				char := msg.String()
				// Only accept printable characters
				if char >= " " && char <= "~" {
					m.searchQuery += char
					m.updateFilteredApps()
				}
			}
			return m, nil
		}
	}

	return m, nil
}

// View renders the TUI interface
func (m Model) View() string {
	if m.showHelp {
		return m.renderHelp()
	}

	var s strings.Builder

	// Header
	s.WriteString("┌" + strings.Repeat("─", m.width-2) + "┐\n")

	// Title with results count - truncate if too long
	title := "TUI App Launcher"
	if m.isRefreshing {
		title += " (refreshing...)"
	} else if m.searchQuery != "" {
		title += fmt.Sprintf(" (%d results)", len(m.filteredApps))
	} else if len(m.filteredApps) > 0 {
		// Check if showing favorites
		favoriteCount := 0
		for _, app := range m.applications {
			if app.IsFavorite {
				favoriteCount++
			}
		}
		if favoriteCount > 0 && len(m.filteredApps) == favoriteCount {
			title += " (favorites)"
		}
	}

	// Truncate title if it's too long for the terminal width
	maxTitleWidth := m.width - 4 // Account for borders and padding
	if maxTitleWidth > 0 {
		title = m.truncateText(title, maxTitleWidth)
	}

	titleLine := "│ " + title
	padding := m.width - len(titleLine) - 1
	if padding > 0 {
		titleLine += strings.Repeat(" ", padding)
	}
	titleLine += "│\n"
	s.WriteString(titleLine)

	s.WriteString("├" + strings.Repeat("─", m.width-2) + "┤\n")

	// Search input with cursor - handle long search queries
	cursor := "█" // Block cursor
	searchPrefix := "Search: "
	maxSearchWidth := m.width - len("│ ") - len(searchPrefix) - len("│") - 1

	searchDisplay := m.searchQuery
	if maxSearchWidth > 1 {
		// If search query is too long, show the end part with ellipsis
		if len(searchDisplay) > maxSearchWidth-1 {
			searchDisplay = "..." + searchDisplay[len(searchDisplay)-(maxSearchWidth-4):]
		}
		searchDisplay += cursor

		// Ensure we don't exceed the available width
		if len(searchDisplay) > maxSearchWidth {
			searchDisplay = m.truncateText(searchDisplay, maxSearchWidth)
		}
	} else {
		searchDisplay = cursor
	}

	searchLine := fmt.Sprintf("│ %s%s", searchPrefix, searchDisplay)
	padding = m.width - len(searchLine) - 1
	if padding > 0 {
		searchLine += strings.Repeat(" ", padding)
	}
	searchLine += "│\n"
	s.WriteString(searchLine)

	s.WriteString("├" + strings.Repeat("─", m.width-2) + "┤\n")

	// Render only the visible window, including when an error reduces its height.
	m.ensureSelectionVisible()
	maxItems := m.visibleRows()
	end := min(len(m.filteredApps), m.scrollOffset+maxItems)
	for i := m.scrollOffset; i < end; i++ {
		app := m.filteredApps[i]

		prefix := "  "
		isSelected := i == m.selectedIndex
		if isSelected {
			prefix = "> "
		}

		favorite := ""
		favoritePlain := ""
		if app.IsFavorite {
			favorite = " \033[1;33m★\033[0m" // Yellow star for favorites
			favoritePlain = " ★"
		}

		// Calculate available width for the application name
		maxNameWidth := m.getMaxAppNameWidth()

		// Get highlighted name if there's a search query
		displayName := app.Name
		plainName := app.Name

		// Truncate name if it's too long
		if len(plainName) > maxNameWidth {
			plainName = m.truncateText(plainName, maxNameWidth)
			displayName = plainName
		}

		if m.searchQuery != "" && len(plainName) == len(app.Name) {
			// Only highlight if we didn't truncate
			matches := m.fuzzySearcher.GetMatchPositions(m.searchQuery, app.Name)
			if len(matches) > 0 {
				displayName = m.highlightMatches(plainName, matches)
			}
		}

		line := fmt.Sprintf("│%s%s%s", prefix, displayName, favorite)

		// Add selection highlighting
		if isSelected {
			// Highlight the entire line for selected item
			line = fmt.Sprintf("│\033[7m%s%s%s\033[0m", prefix, displayName, favorite)
		}

		// Calculate padding without ANSI codes for proper alignment
		plainLine := fmt.Sprintf("│%s%s%s", prefix, plainName, favoritePlain)
		padding = m.width - len(plainLine) - 1
		if padding > 0 {
			if isSelected {
				line += "\033[7m" + strings.Repeat(" ", padding) + "\033[0m"
			} else {
				line += strings.Repeat(" ", padding)
			}
		}
		line += "│\n"
		s.WriteString(line)
	}

	// Fill remaining space
	for i := end - m.scrollOffset; i < maxItems; i++ {
		s.WriteString("│" + strings.Repeat(" ", m.width-2) + "│\n")
	}

	// Footer
	s.WriteString("├" + strings.Repeat("─", m.width-2) + "┤\n")

	// Show error message if present
	if m.errorMessage != "" {
		// Truncate error message if it's too long
		maxErrorWidth := m.width - len("│ Error: ") - 1
		errorMsg := m.errorMessage
		if maxErrorWidth > 0 {
			errorMsg = m.truncateText(errorMsg, maxErrorWidth)
		}

		errorLine := fmt.Sprintf("│ \033[1;31mError:\033[0m %s", errorMsg)
		// Calculate padding without ANSI codes
		plainErrorLine := fmt.Sprintf("│ Error: %s", errorMsg)
		padding = m.width - len(plainErrorLine) - 1
		if padding > 0 {
			errorLine += strings.Repeat(" ", padding)
		}
		errorLine += "│\n"
		s.WriteString(errorLine)
		s.WriteString("├" + strings.Repeat("─", m.width-2) + "┤\n")

		// Responsive footer for error state
		footer := "│ Ctrl+L: Clear  Ctrl+R: Refresh  ?: Help  Esc: Exit"
		if m.width < 60 {
			footer = "│ Ctrl+L: Clear  ?: Help  Esc: Exit"
		}
		if m.width < 40 {
			footer = "│ ?: Help  Esc: Exit"
		}

		padding = m.width - len(footer) - 1
		if padding > 0 {
			footer += strings.Repeat(" ", padding)
		}
		footer += "│"
		s.WriteString(footer)
	} else {
		// Responsive footer for normal state
		footer := "│ ↑/↓: Navigate  Enter: Launch  Tab: Favorite  Ctrl+R: Refresh  ?: Help  Esc: Exit"
		if m.width < 80 {
			footer = "│ ↑/↓: Navigate  Enter: Launch  Tab: Fav  Ctrl+R: Refresh  ?: Help  Esc: Exit"
		}
		if m.width < 70 {
			footer = "│ ↑/↓: Navigate  Enter: Launch  Tab: Fav  ?: Help  Esc: Exit"
		}
		if m.width < 60 {
			footer = "│ ↑/↓: Nav  Enter: Launch  Tab: Fav  ?: Help  Esc: Exit"
		}
		if m.width < 50 {
			footer = "│ ↑/↓  Enter  Tab  ?: Help  Esc: Exit"
		}
		if m.width < 40 {
			footer = "│ ?: Help  Esc: Exit"
		}

		padding = m.width - len(footer) - 1
		if padding > 0 {
			footer += strings.Repeat(" ", padding)
		}
		footer += "│"
		s.WriteString(footer)
	}

	s.WriteString("\n└" + strings.Repeat("─", m.width-2) + "┘")

	return s.String()
}

// renderHelp displays the help screen with proper formatting
func (m Model) renderHelp() string {
	var s strings.Builder

	// Header
	s.WriteString("┌" + strings.Repeat("─", m.width-2) + "┐\n")

	// Title
	title := "│ TUI App Launcher - Help"
	padding := m.width - len(title) - 1
	if padding > 0 {
		title += strings.Repeat(" ", padding)
	}
	title += "│\n"
	s.WriteString(title)

	s.WriteString("├" + strings.Repeat("─", m.width-2) + "┤\n")

	// Help content
	helpLines := []string{
		"",
		"Keyboard Shortcuts:",
		"  ↑/↓           Navigate up/down through applications",
		"  Enter         Launch selected application",
		"  Tab           Toggle favorite status",
		"  Backspace     Delete character from search",
		"  Ctrl+U        Clear search query",
		"  Ctrl+L        Clear error message",
		"  Ctrl+R        Refresh application list",
		"  ?             Toggle this help screen",
		"  Esc/Ctrl+C    Exit application",
		"",
		"Search:",
		"  Type any characters to search for applications",
		"  Search uses fuzzy matching - you don't need exact names",
		"  Clear search to see favorites list",
		"",
		"Favorites:",
		"  Use Tab to add/remove applications from favorites",
		"  Favorites appear first in search results",
		"  When no search query, only favorites are shown",
		"",
		"Error Handling:",
		"  If an application fails to launch, an error message will appear",
		"  Use Ctrl+L to clear error messages",
		"",
		"Press ? again to return to the main view.",
	}

	// Calculate available height for content
	availableHeight := m.height - 4 // Header, separator, footer

	// Display help lines with proper padding
	for i, line := range helpLines {
		if i >= availableHeight {
			break
		}

		helpLine := "│ " + line
		padding = m.width - len(helpLine) - 1
		if padding > 0 {
			helpLine += strings.Repeat(" ", padding)
		}
		helpLine += "│\n"
		s.WriteString(helpLine)
	}

	// Fill remaining space if needed
	for i := len(helpLines); i < availableHeight; i++ {
		s.WriteString("│" + strings.Repeat(" ", m.width-2) + "│\n")
	}

	// Footer
	s.WriteString("├" + strings.Repeat("─", m.width-2) + "┤\n")
	footer := "│ Press ? to return to main view"
	padding = m.width - len(footer) - 1
	if padding > 0 {
		footer += strings.Repeat(" ", padding)
	}
	footer += "│\n"
	s.WriteString(footer)

	s.WriteString("└" + strings.Repeat("─", m.width-2) + "┘")

	return s.String()
}

// updateFilteredApps updates the filtered applications list
func (m *Model) updateFilteredApps() {
	if m.searchQuery == "" {
		// Show favorites when no search query
		m.filteredApps = make([]interfaces.Application, 0)
		for _, app := range m.applications {
			if app.IsFavorite {
				m.filteredApps = append(m.filteredApps, app)
			}
		}
		// If no favorites, show all applications
		if len(m.filteredApps) == 0 {
			m.filteredApps = m.applications
		}
	} else {
		// Use fuzzy search
		results := m.fuzzySearcher.Search(m.searchQuery, m.applications)

		// Separate favorites and non-favorites in search results
		var favoriteResults []interfaces.Application
		var regularResults []interfaces.Application

		for _, result := range results {
			if result.Application.IsFavorite {
				favoriteResults = append(favoriteResults, result.Application)
			} else {
				regularResults = append(regularResults, result.Application)
			}
		}

		// Combine with favorites first
		m.filteredApps = make([]interfaces.Application, 0, len(results))
		m.filteredApps = append(m.filteredApps, favoriteResults...)
		m.filteredApps = append(m.filteredApps, regularResults...)
	}

	// Reset selection if out of bounds
	if m.selectedIndex >= len(m.filteredApps) {
		m.selectedIndex = 0
	}
	m.ensureSelectionVisible()
}

// highlightMatches highlights matching characters in the text
func (m *Model) highlightMatches(text string, matches []int) string {
	if len(matches) == 0 {
		return text
	}

	runes := []rune(text)
	if len(runes) == 0 {
		return text
	}

	// Create a map for quick lookup of match positions
	matchMap := make(map[int]bool)
	for _, pos := range matches {
		if pos >= 0 && pos < len(runes) {
			matchMap[pos] = true
		}
	}

	var result strings.Builder
	for i, r := range runes {
		if matchMap[i] {
			// Add highlighting for matched character
			result.WriteString("\033[1;33m") // Bold yellow
			result.WriteRune(r)
			result.WriteString("\033[0m") // Reset
		} else {
			result.WriteRune(r)
		}
	}

	return result.String()
}

// launchApplication creates a command to launch the selected application
func (m *Model) launchApplication(app interfaces.Application) tea.Cmd {
	return func() tea.Msg {
		return launchMsg{app: app}
	}
}

// launchMsg is sent when an application should be launched
type launchMsg struct {
	app interfaces.Application
}

// launchSuccessMsg is sent when an application launches successfully
type launchSuccessMsg struct{}

// launchErrorMsg is sent when an application fails to launch
type launchErrorMsg struct {
	err error
}

// toggleFavorite creates a command to toggle favorite status
func (m *Model) toggleFavorite(app *interfaces.Application) tea.Cmd {
	return func() tea.Msg {
		return toggleFavoriteMsg{app: app}
	}
}

// refreshApplications creates a command to refresh the application list
func (m *Model) refreshApplications() tea.Cmd {
	return func() tea.Msg {
		apps, err := m.scanner.ScanApplications()
		return refreshCompleteMsg{apps: apps, err: err}
	}
}

// toggleFavoriteMsg is sent when favorite status should be toggled
type toggleFavoriteMsg struct {
	app *interfaces.Application
}

// refreshMsg is sent when applications should be refreshed
type refreshMsg struct{}

// refreshCompleteMsg is sent when refresh is complete
type refreshCompleteMsg struct {
	apps []interfaces.Application
	err  error
}
