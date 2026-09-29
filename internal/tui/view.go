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
	columns := m.columns()
	cellWidth := (listWidth - (columns - 1)) / columns
	end := min(len(m.filteredApps), m.scrollOffset+(rows-shift)*columns)
	for i := m.scrollOffset; i < end; i += columns {
		line := m.renderApplication(i, cellWidth)
		if columns == 2 {
			line += rowStyle.Render(" ")
			if i+1 < end {
				line += m.renderApplication(i+1, listWidth-cellWidth-1)
			} else {
				line += rowStyle.Copy().Width(listWidth - cellWidth - 1).Render("")
			}
		}
		lines = append(lines, line)
	}
	if len(m.filteredApps) == 0 && len(lines) < rows {
		message := "No applications found. Ctrl+R to rescan."
		if m.isRefreshing {
			message = "Scanning for applications..."
		} else if m.searchQuery != "" || m.category != "" {
			message = "No matches. Clear search or change category."
		}
		lines = append(lines, rowStyle.Copy().Width(listWidth).Render(m.truncateText(message, listWidth)))
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
	if m.category != "" {
		title += " [" + m.category + "]"
	}
	if m.caseSensitive {
		title += " [Aa]"
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
	runes := []rune(m.searchQuery)
	cursor := min(max(0, m.searchCursorIndex), len(runes))
	start := 0
	for start < cursor && runewidth.StringWidth(string(runes[start:cursor])) > available {
		start++
	}
	before := string(runes[start:cursor])
	after := m.truncateText(string(runes[cursor:]), max(0, available-runewidth.StringWidth(before)))
	line := " " + prefix + searchTextStyle.Render(before) + m.searchCursor() + searchTextStyle.Render(after)
	return line + searchTextStyle.Render(strings.Repeat(" ", max(0, width-lipgloss.Width(line))))
}

func (m Model) renderApplication(index, width int) string {
	app := m.filteredApps[index]
	selected := index == m.selectedIndex
	name := m.truncateText(app.Name, max(0, width-8))
	if m.searchQuery != "" {
		query, _ := m.searchScope()
		matches := m.fuzzySearcher.GetMatchPositions(query, app.Name)
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
	if number := index - m.scrollOffset + 1; number <= 9 {
		prefix += fmt.Sprintf("%d ", number)
	} else {
		prefix += "  "
	}
	style := rowStyle
	if selected {
		style = selectedRowStyle
	}
	return style.Copy().Width(width).Render(prefix + name + favorite)
}

func (m Model) renderFooter(width int, help bool) string {
	footer := "↑/↓ Navigate  Enter Launch  Ctrl+D Favorite  Ctrl+R Refresh  ? Help  Esc Exit"
	if help {
		footer = "Press ? to return to main view"
	} else if m.errorMessage != "" {
		footer = "Ctrl+L Clear  Ctrl+R Refresh  ? Help  Esc Exit"
	} else if width < 79 {
		footer = "↑/↓ Navigate  Enter Launch  ^D Fav  ? Help  Esc Exit"
	}
	if width < 60 && !help && m.errorMessage == "" {
		footer = "↑/↓ Nav  Enter Launch  ^D Fav  ? Help  Esc Exit"
	}
	if width < 50 && !help && m.errorMessage == "" {
		footer = "↑/↓  Enter  ^D  ? Help  Esc Exit"
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
	left := []string{
		"Keyboard Shortcuts:",
		"↑/↓ Navigate   Home/End First/Last",
		"PgUp/PgDn Move one page",
		"Enter Launch   Alt+1–9 Visible app",
		"Alt+Enter Force terminal launch",
		"Click Select   Double-click Launch",
		"Mouse wheel Navigate/scroll list",
		"Ctrl+D / Ctrl+Space Favorite",
		"Tab / Shift+Tab Cycle category",
		"←/→ Edit cursor   Ctrl+A/E Start/End",
		"Ctrl+W Delete word   Ctrl+U Clear",
		"Ctrl+\\ Toggle case sensitivity",
		"Ctrl+R Rescan   Ctrl+L Clear error",
		"? Help   Esc/Ctrl+C Exit",
	}
	right := []string{
		"Search:", "Words match name/Exec/metadata",
		"/Development query filters category",
		"Clear search to show favorites",
		"Favorites:", "Ctrl+D adds/removes favorites",
		"Favorites appear first in results",
		"Wide terminals show two columns",
		"Settings:", "Edit ~/.config/tui-launcher/config.json",
		"search_fields, ranking, multi_column",
		"case_sensitive (loaded on startup)",
		"Error Handling:", "Ctrl+L clears errors; Ctrl+R rescans",
	}
	rows := m.visibleRows()
	lines := make([]string, 0, rows)
	for i := 0; i < rows; i++ {
		if contentWidth >= 76 {
			first, second := "", ""
			if i < len(left) {
				first = left[i]
			}
			if i < len(right) {
				second = right[i]
			}
			leftWidth := contentWidth / 2
			lines = append(lines, m.renderHelpLine(first, leftWidth)+m.renderHelpLine(second, contentWidth-leftWidth))
		} else {
			all := append(append([]string{}, left...), right...)
			text := ""
			if i < len(all) {
				text = all[i]
			}
			lines = append(lines, m.renderHelpLine(text, contentWidth))
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
	case "Keyboard Shortcuts:", "Search:", "Favorites:", "Error Handling:", "Settings:":
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
