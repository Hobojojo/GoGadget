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

	assertBlack := func(view string, allowSelection bool) {
		t.Helper()
		backgrounds := regexp.MustCompile(`48;2;(\d+;\d+;\d+)`).FindAllStringSubmatch(view, -1)
		if len(backgrounds) == 0 {
			t.Fatal("view did not set an explicit black background")
		}
		for _, color := range backgrounds {
			if color[1] != "0;0;0" && !(allowSelection && color[1] == "17;17;17") {
				t.Fatalf("found an unexpected background: %s", color[1])
			}
		}
		if strings.Contains(view, "129;161;193") || strings.Contains(view, "136;192;208") {
			t.Fatal("teal and blue chrome remain in the black theme")
		}
	}

	assertBlack(m.View(), true)
	assertBlack(m.renderSearch(80), false)
	if !strings.Contains(cursorStyle.Render("█"), "121;162;247") || !strings.Contains(dividerRule(80), "121;162;247") {
		t.Fatalf("cursor and divider must use muted blue: cursor=%q divider=%q", cursorStyle.Render("█"), dividerRule(80))
	}
	selected := m.renderApplication(0, 40)
	assertBlack(selected, true)
	if !strings.Contains(selected, "255;255;255") || !strings.Contains(selected, "121;162;247") || !strings.Contains(stripANSIForTest(selected), "▎ App") {
		t.Fatal("selection lost its bright text or blue indicator")
	}
	for _, color := range regexp.MustCompile(`48;2;(\d+;\d+;\d+)`).FindAllStringSubmatch(selected, -1) {
		if color[1] != "17;17;17" {
			t.Fatal("selected matches and favorite star must share the row highlight")
		}
	}
	m.selectedIndex = -1 // Render the same favorite/search result without selection.
	assertBlack(m.renderApplication(0, 40), false)
	m.selectedIndex = 0
	m.errorMessage = "Example error"
	view := m.View()
	assertBlack(view, true)
	if !strings.Contains(view, "Example error") || !strings.Contains(view, "229;163;169") {
		t.Fatal("error label or message lost readable styling")
	}
	m = updateModel(m, tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'?'}})
	if !m.showHelp {
		t.Fatal("help key did not open the guide")
	}
	m = settleAnimations(t, m)
	help := m.View()
	assertBlack(help, false)
	for _, heading := range []string{"Keyboard Shortcuts:", "Search:", "Favorites:", "Error Handling:"} {
		if !strings.Contains(help, heading) {
			t.Fatalf("help missing %s", heading)
		}
		if !strings.Contains(m.renderHelpLine(heading, 50), "121;162;247") {
			t.Fatalf("help heading %s lost its muted blue accent", heading)
		}
	}
	m = updateModel(m, tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'?'}})
	m = settleAnimations(t, m)
	if m.showHelp || !strings.Contains(m.View(), "Search") {
		t.Fatal("help key did not return to the main view")
	}
}
