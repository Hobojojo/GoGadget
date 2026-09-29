package tui

import (
	"strings"

	"tui-app-launcher/internal/errors"
	"tui-app-launcher/internal/interfaces"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/mattn/go-runewidth"
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

	if runewidth.StringWidth(text) <= maxWidth {
		return text
	}
	if maxWidth <= 3 {
		return strings.Repeat(".", maxWidth)
	}
	var result strings.Builder
	width := 0
	for _, r := range text {
		runeWidth := runewidth.RuneWidth(r)
		if width+runeWidth > maxWidth-3 {
			break
		}
		result.WriteRune(r)
		width += runeWidth
	}
	return result.String() + "..."
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
		if currentLine.Len() > 0 && runewidth.StringWidth(currentLine.String())+1+runewidth.StringWidth(word) > maxWidth {
			lines = append(lines, currentLine.String())
			currentLine.Reset()
		}

		// If the word itself is too long, truncate it
		if runewidth.StringWidth(word) > maxWidth {
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
