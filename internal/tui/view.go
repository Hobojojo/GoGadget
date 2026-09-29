package tui

import (
	"fmt"
	"strings"

	"github.com/mattn/go-runewidth"
	"github.com/muesli/reflow/ansi"
)

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
	padding := m.width - ansi.PrintableRuneWidth(titleLine) - 1
	if padding > 0 {
		titleLine += strings.Repeat(" ", padding)
	}
	titleLine += "│\n"
	s.WriteString(titleLine)

	s.WriteString("├" + strings.Repeat("─", m.width-2) + "┤\n")

	// Search input with cursor - handle long search queries
	cursor := "█" // Block cursor
	searchPrefix := "Search: "
	maxSearchWidth := m.width - ansi.PrintableRuneWidth("│ "+searchPrefix+"│") - 1

	searchDisplay := m.searchQuery
	if maxSearchWidth > 1 {
		if runewidth.StringWidth(searchDisplay)+1 > maxSearchWidth {
			// Keep a terminal-width-bounded suffix without splitting a UTF-8 rune.
			runes := []rune(searchDisplay)
			width := 0
			start := len(runes)
			for start > 0 && width+runewidth.RuneWidth(runes[start-1]) <= maxSearchWidth-4 {
				start--
				width += runewidth.RuneWidth(runes[start])
			}
			searchDisplay = "..." + string(runes[start:])
		}
		searchDisplay += cursor
	} else {
		searchDisplay = cursor
	}

	searchLine := fmt.Sprintf("│ %s%s", searchPrefix, searchDisplay)
	padding = m.width - ansi.PrintableRuneWidth(searchLine) - 1
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
		if app.IsFavorite {
			favorite = " \033[1;33m★\033[0m" // Yellow star for favorites
		}

		// Calculate available width for the application name
		maxNameWidth := m.getMaxAppNameWidth()

		// Match positions refer to the original name. Highlight only the visible
		// prefix, never the ellipsis occupying the end of a truncated name.
		displayName := m.truncateText(app.Name, maxNameWidth)
		if m.searchQuery != "" {
			matches := m.fuzzySearcher.GetMatchPositions(m.searchQuery, app.Name)
			if strings.HasSuffix(displayName, "...") && displayName != app.Name {
				prefix := strings.TrimSuffix(displayName, "...")
				displayName = m.highlightMatches(prefix, matches) + "..."
			} else {
				displayName = m.highlightMatches(displayName, matches)
			}
		}

		line := fmt.Sprintf("│%s%s%s", prefix, displayName, favorite)
		if isSelected {
			// Restore selection styling after nested highlight/favorite resets.
			line = "│\033[7m" + prefix + strings.ReplaceAll(displayName+favorite, "\033[0m", "\033[0m\033[7m") + "\033[0m"
		}

		padding = m.width - ansi.PrintableRuneWidth(line) - 1
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
		padding = m.width - ansi.PrintableRuneWidth(errorLine) - 1
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

		padding = m.width - ansi.PrintableRuneWidth(footer) - 1
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

		padding = m.width - ansi.PrintableRuneWidth(footer) - 1
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
	padding := m.width - ansi.PrintableRuneWidth(title) - 1
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

		helpLine := "│ " + m.truncateText(line, m.width-4)
		padding = m.width - ansi.PrintableRuneWidth(helpLine) - 1
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
	padding = m.width - ansi.PrintableRuneWidth(footer) - 1
	if padding > 0 {
		footer += strings.Repeat(" ", padding)
	}
	footer += "│\n"
	s.WriteString(footer)

	s.WriteString("└" + strings.Repeat("─", m.width-2) + "┘")

	return s.String()
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
