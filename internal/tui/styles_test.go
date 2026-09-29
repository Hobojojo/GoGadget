package tui

import (
	"regexp"
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/mattn/go-runewidth"
	"github.com/muesli/termenv"
)

func TestBlackViewsFitTerminal(t *testing.T) {
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

func TestBlackColorStates(t *testing.T) {
	previous := lipgloss.ColorProfile()
	lipgloss.SetColorProfile(termenv.TrueColor)
	defer lipgloss.SetColorProfile(previous)
	m := testModel(t, 1)
	m.height = 44
	m.applications[0].IsFavorite = true
	m.filteredApps[0].IsFavorite = true
	m.searchQuery = "App"

	assertBlack := func(view string) {
		t.Helper()
		backgrounds := regexp.MustCompile(`48;2;(\d+;\d+;\d+)`).FindAllStringSubmatch(view, -1)
		if len(backgrounds) == 0 {
			t.Fatal("view did not set an explicit black background")
		}
		for _, color := range backgrounds {
			if color[1] != "0;0;0" {
				t.Fatalf("found a non-black background: %s", color[1])
			}
		}
		if strings.Contains(view, "129;161;193") || strings.Contains(view, "136;192;208") {
			t.Fatal("teal and blue chrome remain in the black theme")
		}
	}

	assertBlack(m.View())
	selected := m.renderApplication(0, 40)
	assertBlack(selected)
	if !strings.Contains(selected, "255;255;255") || !strings.Contains(stripANSIForTest(selected), "> App") {
		t.Fatal("selection lost its bright text or indicator")
	}
	m.errorMessage = "Example error"
	view := m.View()
	assertBlack(view)
	if !strings.Contains(view, "Example error") || !strings.Contains(view, "229;163;169") {
		t.Fatal("error label or message lost readable styling")
	}
	m = updateModel(m, tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'?'}})
	if !m.showHelp {
		t.Fatal("help key did not open the guide")
	}
	help := m.View()
	assertBlack(help)
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
