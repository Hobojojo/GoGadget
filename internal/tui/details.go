package tui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
)

// Wide terminals share the existing list panel with a read-only details column.
// Narrow terminals retain the full-width list and the same viewport height.
func (m Model) listWidth(inner int) int {
	if m.width <= 80 {
		return inner - 2
	}
	return (inner - 5) / 2
}

func (m Model) renderListPanel(lines []string, inner int) string {
	if m.width > 80 {
		left := m.listWidth(inner)
		right := inner - 5 - left
		details := m.detailLines(right, len(lines))
		for i := range lines {
			lines[i] += rowStyle.Render(" ") + helpHeadingStyle.Render("│") + rowStyle.Render(" ") + details[i]
		}
	}
	return panelStyle.Copy().Width(inner - 2).Render(lipgloss.JoinVertical(lipgloss.Left, lines...))
}

func (m Model) detailLines(width, rows int) []string {
	text := []string{"App details"}
	if len(m.filteredApps) == 0 {
		text = append(text, "", "No application selected")
	} else {
		app := m.filteredApps[m.selectedIndex]
		text = append(text, "", app.Name)
		for _, field := range []struct{ label, value string }{
			{"Description", app.Comment},
			{"Generic name", app.GenericName},
			{"Exec", app.Exec},
			{"Terminal", fmt.Sprint(app.Terminal)},
			{"Categories", strings.Join(app.Categories, ", ")},
			{"Keywords", strings.Join(app.Keywords, ", ")},
			{"Working directory", app.Path},
			{"Desktop file", app.DesktopFile},
		} {
			if field.value != "" {
				text = append(text, field.label+": "+field.value)
			}
		}
	}
	lines := make([]string, rows)
	for i := range lines {
		value := ""
		if i < len(text) {
			// Flatten control whitespace before rendering terminal cells.
			value = m.truncateText(strings.Join(strings.Fields(text[i]), " "), width)
		}
		style := rowStyle
		if i == 0 {
			style = helpHeadingStyle
		}
		lines[i] = style.Copy().Width(width).Render(value)
	}
	return lines
}
