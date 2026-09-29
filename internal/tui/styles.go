package tui

import (
	"strings"

	"github.com/charmbracelet/lipgloss"
)

const (
	backgroundColor lipgloss.Color = "#000000"
	selectionColor  lipgloss.Color = "#111111"
	accentColor     lipgloss.Color = "#7AA2F7"
	textColor       lipgloss.Color = "#E5E5E5"
	brightColor     lipgloss.Color = "#FFFFFF"
	mutedColor      lipgloss.Color = "#A3A3A3"
	borderColor     lipgloss.Color = "#737373"
	favoriteColor   lipgloss.Color = "#EBCB8B"
	errorColor      lipgloss.Color = "#E5A4A9"
)

var (
	containerStyle         = lipgloss.NewStyle().Border(lipgloss.DoubleBorder()).BorderForeground(borderColor).BorderBackground(backgroundColor).Background(backgroundColor).Foreground(textColor)
	headerStyle            = lipgloss.NewStyle().Border(lipgloss.NormalBorder()).BorderForeground(borderColor).BorderBackground(backgroundColor).Background(backgroundColor).Foreground(brightColor).Bold(true).Padding(0, 1)
	headerTextStyle        = lipgloss.NewStyle().Foreground(brightColor).Background(backgroundColor).Bold(true)
	headerAccentStyle      = lipgloss.NewStyle().Foreground(brightColor).Background(backgroundColor)
	searchStyle            = lipgloss.NewStyle().Foreground(brightColor).Background(backgroundColor).Bold(true)
	searchTextStyle        = lipgloss.NewStyle().Foreground(textColor).Background(backgroundColor)
	cursorStyle            = lipgloss.NewStyle().Foreground(accentColor).Background(backgroundColor).Bold(true)
	panelStyle             = lipgloss.NewStyle().Border(lipgloss.NormalBorder()).BorderForeground(borderColor).BorderBackground(backgroundColor).Background(backgroundColor).Foreground(textColor)
	rowStyle               = lipgloss.NewStyle().Background(backgroundColor).Foreground(textColor)
	selectedRowStyle       = lipgloss.NewStyle().Background(selectionColor).Foreground(brightColor).Bold(true)
	selectedIndicatorStyle = lipgloss.NewStyle().Foreground(accentColor).Background(selectionColor).Bold(true)
	favoriteStyle          = lipgloss.NewStyle().Foreground(favoriteColor).Background(backgroundColor).Bold(true)
	selectedFavoriteStyle  = favoriteStyle.Copy().Background(selectionColor)
	matchStyle             = lipgloss.NewStyle().Foreground(brightColor).Background(backgroundColor).Bold(true).Underline(true)
	selectedMatchStyle     = matchStyle.Copy().Background(selectionColor)
	errorStyle             = lipgloss.NewStyle().Foreground(errorColor).Background(backgroundColor).Bold(true)
	footerStyle            = lipgloss.NewStyle().Border(lipgloss.NormalBorder()).BorderForeground(borderColor).BorderBackground(backgroundColor).Background(backgroundColor).Foreground(mutedColor).Padding(0, 1)
	footerTextStyle        = lipgloss.NewStyle().Foreground(mutedColor).Background(backgroundColor)
	helpHeadingStyle       = lipgloss.NewStyle().Foreground(accentColor).Background(backgroundColor).Bold(true)
)

func dividerRule(width int) string {
	return lipgloss.NewStyle().Foreground(accentColor).Background(backgroundColor).Render(strings.Repeat("━", max(0, width)))
}
