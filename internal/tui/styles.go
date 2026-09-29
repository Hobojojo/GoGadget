package tui

import (
	"strings"

	"github.com/charmbracelet/lipgloss"
)

// Nord's polar night, snow storm, frost and aurora palettes.
const (
	nord0        lipgloss.Color = "#2E3440"
	nord1        lipgloss.Color = "#3B4252"
	nord2        lipgloss.Color = "#434C5E"
	nord3        lipgloss.Color = "#4C566A"
	nord4        lipgloss.Color = "#D8DEE9"
	nord5        lipgloss.Color = "#ECEFF4"
	frostDeep    lipgloss.Color = "#5E81AC"
	frostBlue    lipgloss.Color = "#81A1C1"
	frostCyan    lipgloss.Color = "#88C0D0"
	frostIntense lipgloss.Color = "#8FBCBB"
	auroraRed    lipgloss.Color = "#BF616A"
	auroraOrange lipgloss.Color = "#D08770"
	auroraYellow lipgloss.Color = "#EBCB8B"
	auroraGreen  lipgloss.Color = "#A3BE8C"
	auroraPurple lipgloss.Color = "#B48EAD"
	// A lighter red keeps error text readable on Nord's dark surfaces.
	readableRed lipgloss.Color = "#E5A4A9"
)

var (
	containerStyle        = lipgloss.NewStyle().Border(lipgloss.DoubleBorder()).BorderForeground(frostBlue).Background(nord1).Foreground(nord5)
	headerStyle           = lipgloss.NewStyle().Border(lipgloss.NormalBorder()).BorderForeground(frostBlue).Background(nord1).Foreground(nord5).Bold(true).Padding(0, 1)
	headerTextStyle       = lipgloss.NewStyle().Foreground(nord5).Bold(true)
	headerAccentStyle     = lipgloss.NewStyle().Foreground(nord4).Background(nord1)
	searchStyle           = lipgloss.NewStyle().Foreground(frostCyan).Background(nord0).Bold(true)
	searchTextStyle       = lipgloss.NewStyle().Foreground(nord5).Background(nord0)
	cursorStyle           = lipgloss.NewStyle().Foreground(auroraYellow).Background(nord0).Bold(true)
	panelStyle            = lipgloss.NewStyle().Border(lipgloss.NormalBorder()).BorderForeground(frostBlue).Background(nord0).Foreground(nord5)
	rowStyle              = lipgloss.NewStyle().Background(nord0).Foreground(nord5)
	selectedRowStyle      = lipgloss.NewStyle().Background(frostBlue).Foreground(nord0).Bold(true)
	favoriteStyle         = lipgloss.NewStyle().Foreground(auroraGreen).Bold(true)
	selectedFavoriteStyle = lipgloss.NewStyle().Foreground(nord0).Background(frostBlue).Bold(true)
	matchStyle            = lipgloss.NewStyle().Foreground(auroraYellow).Bold(true)
	selectedMatchStyle    = lipgloss.NewStyle().Foreground(nord0).Background(frostBlue).Bold(true).Underline(true)
	errorStyle            = lipgloss.NewStyle().Foreground(readableRed).Background(nord0).Bold(true)
	footerStyle           = lipgloss.NewStyle().Border(lipgloss.NormalBorder()).BorderForeground(frostBlue).Background(nord1).Foreground(nord5).Padding(0, 1)
	footerTextStyle       = lipgloss.NewStyle().Foreground(nord5).Background(nord1)
	helpShortcutStyle     = lipgloss.NewStyle().Foreground(frostBlue).Bold(true)
	helpSearchStyle       = lipgloss.NewStyle().Foreground(auroraGreen).Bold(true)
	helpFavoriteStyle     = lipgloss.NewStyle().Foreground(auroraYellow).Bold(true)
	helpErrorStyle        = lipgloss.NewStyle().Foreground(readableRed).Bold(true)
)

// A one-cell-high terminal approximation of the four-stop aurora gradient.
func auroraRule(width int) string {
	colors := []lipgloss.Color{frostBlue, auroraGreen, auroraYellow, auroraRed}
	stops := []int{34, 57, 78, 100}
	var line strings.Builder
	start := 0
	for i, stop := range stops {
		end := (width*stop + 99) / 100
		end = min(width, end)
		line.WriteString(lipgloss.NewStyle().Foreground(colors[i]).Render(strings.Repeat("━", end-start)))
		start = end
	}
	return line.String()
}
