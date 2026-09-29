package tui

import (
	"fmt"
	"math"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/muesli/reflow/truncate"
)

func unit(value float64) float64 {
	return math.Max(0, math.Min(1, value))
}

func (m Model) entranceRows() int {
	return min(max(0, m.visibleRows()-1-(m.selectedIndex-m.scrollOffset)), max(0, int(math.Round(m.animations.list.value))))
}

// Terminal rows cannot move by fractions of a cell. Crossfading adjacent markers
// makes the spring's fractional position visible while selection stays immediate.
func (m Model) applicationIndicator(index int, selected bool) string {
	position := float64(m.selectedIndex - m.scrollOffset)
	if m.animations.cursor.active {
		position = m.animations.cursor.value
	}
	strength := unit(1 - math.Abs(float64(index-m.scrollOffset)-position))
	if strength <= 0 {
		return "  "
	}
	style := selectedIndicatorStyle
	if !selected {
		style = style.Copy().Background(backgroundColor)
	}
	if strength < 1 {
		color := lipgloss.Color(fmt.Sprintf("#%02x%02x%02x", int(122*strength), int(162*strength), int(247*strength)))
		style = style.Copy().Foreground(color)
	}
	return style.Render("▎") + " "
}

func (m Model) searchCursor() string {
	style := cursorStyle
	if m.animations.pulse.active {
		amount := unit(m.animations.pulse.value)
		color := lipgloss.Color(fmt.Sprintf("#%02x%02x%02x", 122+int(133*amount), 162+int(93*amount), 247+int(8*amount)))
		style = style.Copy().Foreground(color)
	}
	return style.Render("█")
}

// Reveal the help view from the right, cropping ANSI-styled lines to the exact
// terminal grid. Progress is reversible, including repeated '?' presses.
func (m Model) renderHelpTransition() string {
	main := m
	main.showHelp = false
	help := m
	help.showHelp = true
	progress := unit(m.animations.help.value)
	if progress == 0 {
		return main.renderMain()
	}
	if progress == 1 {
		return help.renderHelp()
	}
	leftWidth := m.width - int(math.Round(float64(m.width)*progress))
	mainLines := strings.Split(main.renderMain(), "\n")
	helpLines := strings.Split(help.renderHelp(), "\n")
	for i := range mainLines {
		left := truncate.String(mainLines[i], uint(leftWidth))
		right := truncate.String(helpLines[i], uint(m.width-leftWidth))
		mainLines[i] = left + right
	}
	return strings.Join(mainLines, "\n")
}
