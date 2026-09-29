package tui

import (
	"fmt"
	"math"
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/mattn/go-runewidth"
	"github.com/muesli/termenv"
)

func settleAnimations(t *testing.T, m Model) Model {
	t.Helper()
	for i := 0; i < animationFPS*3; i++ {
		if !m.animations.active() && !m.animations.pending {
			return m
		}
		if !m.animations.pending {
			m.animationTick()
		}
		m = updateModel(m, animationFrameMsg{id: m.animations.frameID})
	}
	t.Fatal("animation did not settle within three seconds")
	return m
}

func assertAnimationGrid(t *testing.T, m Model) {
	t.Helper()
	lines := strings.Split(m.View(), "\n")
	if len(lines) != m.height {
		t.Fatalf("animation rendered %d rows, want %d", len(lines), m.height)
	}
	for _, line := range lines {
		if width := runewidth.StringWidth(stripANSIForTest(line)); width != m.width {
			t.Fatalf("animation line is %d columns, want %d: %q", width, m.width, line)
		}
	}
}

func TestSpringCursorKeepsActionsImmediateAndOneTickPending(t *testing.T) {
	m := testModel(t, 5)
	model, cmd := m.Update(tea.KeyMsg{Type: tea.KeyDown})
	m = model.(Model)
	if cmd == nil || !m.animations.pending || !m.animations.cursor.active || m.selectedIndex != 1 || m.animations.cursor.value != 0 {
		t.Fatal("navigation did not start animation with immediate logical selection")
	}
	id := m.animations.frameID
	model, cmd = m.Update(tea.KeyMsg{Type: tea.KeyDown})
	m = model.(Model)
	if cmd != nil || m.animations.frameID != id || m.selectedIndex != 2 {
		t.Fatal("rapid navigation duplicated a tick or delayed selection")
	}
	m = updateModel(m, animationFrameMsg{id: id})
	if m.animations.cursor.value <= 0 || m.animations.cursor.value >= 2 {
		t.Fatal("spring did not interpolate toward the latest row")
	}
	if !strings.Contains(stripANSIForTest(m.renderApplication(0, 40)), "▎") || !strings.Contains(stripANSIForTest(m.renderApplication(1, 40)), "▎") {
		t.Fatal("fractional cursor position did not crossfade adjacent row markers")
	}
	_, launch := m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	if launch == nil || launch().(launchMsg).app.Name != "Application 02" {
		t.Fatal("Enter targeted the visual cursor instead of the selected application")
	}
	m = settleAnimations(t, m)
	if m.animations.cursor.value != 2 || m.animationTick() != nil {
		t.Fatal("cursor did not settle or animation continued while idle")
	}
	// Stale frames must not restart the animation.
	model, cmd = m.Update(animationFrameMsg{id: id})
	if cmd != nil || model.(Model).animations.pending {
		t.Fatal("stale animation frame restarted the scheduler")
	}
}

func TestListEntranceAndRefreshCancelOnNavigation(t *testing.T) {
	m := testModel(t, 8)
	m = updateModel(m, initMsg{})
	if !m.animations.list.active || m.entranceRows() != 3 {
		t.Fatal("initial list did not slide upward into place")
	}
	assertAnimationGrid(t, m)
	if !strings.Contains(stripANSIForTest(m.View()), "Application 00") {
		t.Fatal("entrance hid the selected application")
	}
	m = updateModel(m, tea.KeyMsg{Type: tea.KeyDown})
	if m.animations.list.active || m.entranceRows() != 0 || m.selectedIndex != 1 {
		t.Fatal("navigation waited for the entrance animation")
	}
	m = settleAnimations(t, m)
	m = updateModel(m, refreshCompleteMsg{apps: m.applications})
	if !m.animations.list.active {
		t.Fatal("successful refresh did not replay list entrance")
	}
	m = settleAnimations(t, m)
	if m.entranceRows() != 0 {
		t.Fatal("list did not settle into its normal layout")
	}
}

func TestHelpSlidesAndReversesWithoutChangingGrid(t *testing.T) {
	profile := lipgloss.ColorProfile()
	lipgloss.SetColorProfile(termenv.TrueColor)
	defer lipgloss.SetColorProfile(profile)
	for _, width := range []int{40, 80, 100} {
		m := testModel(t, 6)
		m.width, m.height = width, 32
		m = updateModel(m, tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'?'}})
		if !m.showHelp || !m.animations.help.active {
			t.Fatal("help did not start its slide-in transition")
		}
		for frame := 0; frame < 4; frame++ {
			m = updateModel(m, animationFrameMsg{id: m.animations.frameID})
			assertAnimationGrid(t, m)
		}
		if m.animations.help.value <= 0 || m.animations.help.value >= 1 {
			t.Fatal("help transition did not interpolate")
		}
		m = updateModel(m, tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'?'}})
		if m.showHelp || m.animations.help.target != 0 {
			t.Fatal("repeated help toggle did not reverse the current slide")
		}
		m = settleAnimations(t, m)
		assertAnimationGrid(t, m)
		m = updateModel(m, tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'?'}})
		m = settleAnimations(t, m)
		if !strings.Contains(m.View(), "Keyboard Shortcuts:") {
			t.Fatal("settled help view lost its content")
		}
		assertAnimationGrid(t, m)
	}
}

func TestErrorEntranceAndSearchPulse(t *testing.T) {
	profile := lipgloss.ColorProfile()
	lipgloss.SetColorProfile(termenv.TrueColor)
	defer lipgloss.SetColorProfile(profile)
	m := testModel(t, 8)
	m = updateModel(m, launchErrorMsg{err: fmt.Errorf("test launch error")})
	if !m.animations.error.active || m.animations.error.value != 6 {
		t.Fatal("recoverable error did not animate into place")
	}
	assertAnimationGrid(t, m)
	m = settleAnimations(t, m)
	if !strings.Contains(m.View(), "Error:") || m.animations.error.value != 0 {
		t.Fatal("error animation lost its content or failed to settle")
	}
	m = updateModel(m, tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'A'}})
	if !m.animations.pulse.active || m.searchQuery != "A" || len(m.filteredApps) == 0 {
		t.Fatal("search pulse did not start while filtering immediately")
	}
	if m.searchCursor() == cursorStyle.Render("█") {
		t.Fatal("search pulse did not change the cursor color")
	}
	m = settleAnimations(t, m)
	if m.searchCursor() != cursorStyle.Render("█") {
		t.Fatal("search pulse did not return to muted blue")
	}
	m = updateModel(m, tea.KeyMsg{Type: tea.KeyCtrlL})
	m = updateModel(m, tea.KeyMsg{Type: tea.KeyCtrlU})
	if m.errorMessage != "" || m.searchQuery != "" || m.animations.error.active || m.animations.pulse.active {
		t.Fatal("clearing search/error left stale animation state")
	}
}

func TestAnimationBoundsResizeEmptyResultsAndQuit(t *testing.T) {
	m := testModel(t, 30)
	for i := 0; i < 25; i++ {
		m = updateModel(m, tea.KeyMsg{Type: tea.KeyDown})
	}
	m = updateModel(m, tea.WindowSizeMsg{Width: 40, Height: 10})
	if m.selectedIndex != 25 || m.animations.cursor.active {
		t.Fatal("resize delayed or changed the selected application")
	}
	assertAnimationGrid(t, m)
	m = updateModel(m, tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'z'}})
	if len(m.filteredApps) != 0 || math.IsNaN(m.animations.cursor.value) {
		t.Fatal("empty results left an invalid animation target")
	}
	assertAnimationGrid(t, m)
	m = updateModel(m, tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'?'}})
	id := m.animations.frameID
	model, quit := m.Update(tea.KeyMsg{Type: tea.KeyEsc})
	m = model.(Model)
	if quit == nil || m.animations.active() || m.animations.pending {
		t.Fatal("quit did not stop animation scheduling")
	}
	model, cmd := m.Update(animationFrameMsg{id: id})
	if cmd != nil || model.(Model).animations.active() {
		t.Fatal("in-flight frame reactivated animations after quit")
	}
}
