package tui

import (
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/mattn/go-runewidth"
	"github.com/muesli/termenv"
)

func TestNordViewsFitTerminal(t *testing.T) {
	for _, size := range []struct{ width, height int }{{40, 10}, {80, 24}, {100, 32}} {
		m := testModel(t, 35)
		m.width, m.height = size.width, size.height
		for _, state := range []string{"main", "error", "help"} {
			m.showHelp = state == "help"
			if state == "error" {
				m.errorMessage = "Something went wrong"
			} else {
				m.errorMessage = ""
			}
			view := m.View()
			lines := strings.Split(view, "\n")
			if len(lines) != m.height {
				t.Fatalf("%s at %dx%d: got %d lines", state, m.width, m.height, len(lines))
			}
			for _, line := range lines {
				if got := runewidth.StringWidth(stripANSIForTest(line)); got != m.width {
					t.Fatalf("%s at %dx%d: got line width %d: %q", state, m.width, m.height, got, line)
				}
			}
		}
	}
}

func TestNordColorStates(t *testing.T) {
	previous := lipgloss.ColorProfile()
	lipgloss.SetColorProfile(termenv.TrueColor)
	defer lipgloss.SetColorProfile(previous)
	m := testModel(t, 1)
	m.height = 44
	m.applications[0].IsFavorite = true
	m.filteredApps[0].IsFavorite = true
	m.searchQuery = "App"
	main := m.View()
	for _, color := range []string{"129;161;193", "163;190;140", "235;203;139", "191;97;105"} {
		if !strings.Contains(main, color) {
			t.Fatalf("missing Nord color %s in main view: %q", color, main)
		}
	}
	selected := m.renderApplication(0, 40)
	if !strings.Contains(selected, "38;2;46;52;64") || strings.Contains(selected, "38;2;163;190;140") || strings.Contains(selected, "38;2;235;203;139") {
		t.Fatal("selected row accents must remain legible on the blue background")
	}
	m.errorMessage = "Example error"
	if view := m.View(); !strings.Contains(view, "Example error") || !strings.Contains(view, "229;163;169") {
		t.Fatalf("error label or message lost readable red styling: %q", view)
	}
	m = updateModel(m, tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'?'}})
	if !m.showHelp {
		t.Fatal("help key did not open the guide")
	}
	help := m.View()
	for _, heading := range []string{"Keyboard Shortcuts:", "Search:", "Favorites:", "Error Handling:"} {
		if !strings.Contains(help, heading) {
			t.Fatalf("help missing %s", heading)
		}
	}
	m = updateModel(m, tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'?'}})
	if m.showHelp || !strings.Contains(m.View(), "Search") {
		t.Fatal("help key did not return to the main view")
	}
}
