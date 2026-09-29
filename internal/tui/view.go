package tui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/mattn/go-runewidth"
)

// View renders the launcher in the same terminal grid used by Bubble Tea.
func (m Model) View() string {
	if m.animations.help.active {
		return m.renderHelpTransition()
	}
	if m.showHelp {
		return m.renderHelp()
	}
	return m.renderMain()
}

func (m Model) renderMain() string {
	inner := m.width - 2
	rows := m.visibleRows()
	m.ensureSelectionVisible()

	listWidth := m.listWidth(inner)
	lines := make([]string, 0, rows)
	shift := m.entranceRows()
	for i := 0; i < shift; i++ {
		lines = append(lines, rowStyle.Copy().Width(listWidth).Render(""))
	}
	end := min(len(m.filteredApps), m.scrollOffset+rows-shift)
	for i := m.scrollOffset; i < end; i++ {
		lines = append(lines, m.renderApplication(i, listWidth))
	}
	for len(lines) < rows {
		lines = append(lines, rowStyle.Copy().Width(listWidth).Render(""))
	}

	parts := []string{m.renderHeader(inner), dividerRule(inner), m.renderSearch(inner), m.renderListPanel(lines, inner)}
	if m.errorMessage != "" {
		label := errorStyle.Render("Error:")
		offset := min(6, max(0, int(m.animations.error.value+0.5)))
		message := m.truncateText(m.errorMessage, inner-10-offset)
		parts = append(parts, searchTextStyle.Copy().Width(inner).Render(strings.Repeat(" ", offset+1)+label+" "+message))
	}
	parts = append(parts, m.renderFooter(inner, false))
	return containerStyle.Copy().Width(m.width - 2).Render(lipgloss.JoinVertical(lipgloss.Left, parts...))
}

func (m Model) title() string {
	title := "TUI App Launcher"
	if m.isRefreshing {
		title += " (refreshing...)"
	} else if m.searchQuery != "" {
		title += fmt.Sprintf(" (%d results)", len(m.filteredApps))
	} else if len(m.filteredApps) > 0 {
		favorites := 0
		for _, app := range m.applications {
			if app.IsFavorite {
				favorites++
			}
		}
		if favorites > 0 && len(m.filteredApps) == favorites {
			title += " (favorites)"
		}
	}
	return title
}

func (m Model) renderHeader(width int) string {
	if m.height < 14 {
		return headerTextStyle.Copy().Background(backgroundColor).Width(width).Render(" " + m.truncateText(m.title(), width-2))
	}
	// Keep the title accent on the same black surface as the banner.
	text := headerTextStyle.Copy().Background(backgroundColor).Render(m.truncateText(m.title(), width-5))
	return headerStyle.Copy().Width(width - 2).Render(headerAccentStyle.Render("▍") + " " + text)
}

func (m Model) renderSearch(width int) string {
	prefix := searchStyle.Render("  Search  ")
	available := max(0, width-lipgloss.Width(prefix)-3)
	query := m.searchQuery
	if runewidth.StringWidth(query) > available {
		// Retain the newest search input; do not split a UTF-8 rune.
		runes := []rune(query)
		start, used := len(runes), 0
		for start > 0 && used+runewidth.RuneWidth(runes[start-1]) <= max(0, available-3) {
			start--
			used += runewidth.RuneWidth(runes[start])
		}
		query = "..." + string(runes[start:])
	}
	line := " " + prefix + searchTextStyle.Render(query) + m.searchCursor()
	return line + searchTextStyle.Render(strings.Repeat(" ", max(0, width-lipgloss.Width(line))))
}

func (m Model) renderApplication(index, width int) string {
	app := m.filteredApps[index]
	selected := index == m.selectedIndex
	name := m.truncateText(app.Name, max(0, width-6))
	if m.searchQuery != "" {
		matches := m.fuzzySearcher.GetMatchPositions(m.searchQuery, app.Name)
		highlight := matchStyle
		if selected {
			highlight = selectedMatchStyle
		}
		if strings.HasSuffix(name, "...") && name != app.Name {
			name = m.highlightMatches(strings.TrimSuffix(name, "..."), matches, highlight) + "..."
		} else {
			name = m.highlightMatches(name, matches, highlight)
		}
	}
	favorite := ""
	if app.IsFavorite {
		star := favoriteStyle
		if selected {
			star = selectedFavoriteStyle
		}
		favorite = " " + star.Render("★")
	}
	prefix := m.applicationIndicator(index, selected)
	style := rowStyle
	if selected {
		style = selectedRowStyle
	}
	return style.Copy().Width(width).Render(prefix + name + favorite)
}

func (m Model) renderFooter(width int, help bool) string {
	footer := "↑/↓ Navigate  Enter Launch  Tab Favorite  Ctrl+R Refresh  ? Help  Esc Exit"
	if help {
		footer = "Press ? to return to main view"
	} else if m.errorMessage != "" {
		footer = "Ctrl+L Clear  Ctrl+R Refresh  ? Help  Esc Exit"
	} else if width < 79 {
		footer = "↑/↓ Navigate  Enter Launch  Tab Fav  ? Help  Esc Exit"
	}
	if width < 60 && !help && m.errorMessage == "" {
		footer = "↑/↓ Nav  Enter Launch  Tab Fav  ? Help  Esc Exit"
	}
	if width < 50 && !help && m.errorMessage == "" {
		footer = "↑/↓  Enter  Tab  ? Help  Esc Exit"
	}
	if width < 40 && !help {
		footer = "? Help  Esc Exit"
	}
	if m.height < 14 {
		return footerTextStyle.Copy().Width(width).Render(" " + m.truncateText(footer, width-2))
	}
	return footerStyle.Copy().Width(width - 2).Render(footerTextStyle.Render(m.truncateText(footer, width-4)))
}

// renderHelp keeps the same chrome as the main view and colors section headings.
func (m Model) renderHelp() string {
	inner := m.width - 2
	contentWidth := inner - 2
	helpLines := []string{
		"", "Keyboard Shortcuts:",
		"  ↑/↓           Navigate up/down through applications",
		"  Enter         Launch selected application",
		"  Tab           Toggle favorite status",
		"  Backspace     Delete character from search",
		"  Ctrl+U        Clear search query",
		"  Ctrl+L        Clear error message",
		"  Ctrl+R        Refresh application list",
		"  ?             Toggle this help screen",
		"  Esc/Ctrl+C    Exit application",
		"", "Search:",
		"  Type any characters to search for applications",
		"  Words match names, commands and metadata in any order",
		"  Clear search to see favorites list",
		"", "Favorites:",
		"  Use Tab to add/remove applications from favorites",
		"  Favorites appear first in search results",
		"  When no search query, only favorites are shown",
		"", "Error Handling:",
		"  If an application fails to launch, an error message will appear",
		"  Use Ctrl+L to clear error messages",
		"", "Press ? again to return to the main view.",
	}
	rows := m.visibleRows()
	lines := make([]string, 0, rows)
	if contentWidth >= 76 {
		// Two columns show all shortcut and help sections at standard heights.
		left := helpLines[1:11]
		right := append(append(append([]string{}, helpLines[12:16]...), helpLines[17:21]...), helpLines[22:25]...)
		right = append(right, helpLines[26])
		leftWidth := contentWidth / 2
		for i := 0; i < rows; i++ {
			var first, second string
			if i < len(left) {
				first = left[i]
			}
			if i < len(right) {
				second = right[i]
			}
			lines = append(lines, m.renderHelpLine(first, leftWidth)+m.renderHelpLine(second, contentWidth-leftWidth))
		}
	} else {
		for i := 0; i < rows; i++ {
			line := ""
			if i < len(helpLines) {
				line = helpLines[i]
			}
			lines = append(lines, m.renderHelpLine(line, contentWidth))
		}
	}
	parts := []string{m.renderHelpHeader(inner), dividerRule(inner), panelStyle.Copy().Width(inner - 2).Render(lipgloss.JoinVertical(lipgloss.Left, lines...)), m.renderFooter(inner, true)}
	return containerStyle.Copy().Width(m.width - 2).Render(lipgloss.JoinVertical(lipgloss.Left, parts...))
}

func (m Model) renderHelpHeader(width int) string {
	label := m.truncateText("TUI App Launcher - Help", width-5)
	if m.height < 14 {
		return headerTextStyle.Copy().Background(backgroundColor).Width(width).Render(" " + label)
	}
	return headerStyle.Copy().Width(width - 2).Render(headerAccentStyle.Render("▍") + " " + headerTextStyle.Copy().Background(backgroundColor).Render(label))
}

func (m Model) renderHelpLine(text string, width int) string {
	text = m.truncateText(text, width-2)
	switch text {
	case "Keyboard Shortcuts:", "Search:", "Favorites:", "Error Handling:":
		text = helpHeadingStyle.Render(text)
	}
	return rowStyle.Copy().Width(width).Render(" " + text)
}

// highlightMatches colors only visible rune positions; ellipses are never highlighted.
func (m *Model) highlightMatches(text string, matches []int, style lipgloss.Style) string {
	if len(matches) == 0 {
		return text
	}
	positions := make(map[int]bool, len(matches))
	for _, pos := range matches {
		positions[pos] = true
	}
	var result strings.Builder
	for i, r := range []rune(text) {
		if positions[i] {
			result.WriteString(style.Render(string(r)))
		} else {
			result.WriteRune(r)
		}
	}
	return result.String()
}
