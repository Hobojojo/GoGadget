package tui

import (
	tea "github.com/charmbracelet/bubbletea"
	"strings"
	"testing"
	"tui-app-launcher/internal/config"
	"tui-app-launcher/internal/interfaces"
)

func TestSearchEditingKeys(t *testing.T) {
	m := testModel(t, 0)
	key := func(k tea.KeyType) { m = updateModel(m, tea.KeyMsg{Type: k}) }
	typeText := func(text string) { m = updateModel(m, tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune(text)}) }
	typeText("café editor")
	key(tea.KeyCtrlW)
	if m.searchQuery != "café " {
		t.Fatal(m.searchQuery)
	}
	key(tea.KeyCtrlA)
	typeText("new ")
	if m.searchQuery != "new café " {
		t.Fatal(m.searchQuery)
	}
	key(tea.KeyCtrlE)
	key(tea.KeyBackspace)
	key(tea.KeyLeft)
	key(tea.KeyDelete)
	if m.searchQuery != "new caf" {
		t.Fatal(m.searchQuery)
	}
	key(tea.KeyCtrlU)
	if m.searchQuery != "" || m.searchCursorIndex != 0 {
		t.Fatal("clear left stale cursor")
	}
}

func TestPageNavigationAndTwoColumnViewport(t *testing.T) {
	m := testModel(t, 60)
	m = updateModel(m, tea.WindowSizeMsg{Width: 160, Height: 24})
	if m.columns() != 2 {
		t.Fatal("wide layout not enabled")
	}
	m = updateModel(m, tea.KeyMsg{Type: tea.KeyPgDown})
	if m.selectedIndex != m.visibleCapacity() {
		t.Fatal("wrong page step")
	}
	m = updateModel(m, tea.KeyMsg{Type: tea.KeyEnd})
	if m.selectedIndex != 59 || m.selectedIndex >= m.scrollOffset+m.visibleCapacity() {
		t.Fatal("end not visible")
	}
	m = updateModel(m, tea.KeyMsg{Type: tea.KeyHome})
	if m.selectedIndex != 0 || m.scrollOffset != 0 {
		t.Fatal("home not visible")
	}
	for i := 0; i < 60; i++ {
		m = updateModel(m, tea.KeyMsg{Type: tea.KeyDown})
		assertVisible(t, m)
	}
	disabled := false
	m.settings.MultiColumn = &disabled
	if m.columns() != 1 {
		t.Fatal("single column preference ignored")
	}
}

func TestCategoryAndCaseShortcuts(t *testing.T) {
	m := testModel(t, 0)
	m.SetApplications([]interfaces.Application{
		{Name: "Firefox", Exec: "firefox", Categories: []string{"Network"}},
		{Name: "Editor", Exec: "edit", Categories: []string{"Development"}},
	})
	m = updateModel(m, tea.KeyMsg{Type: tea.KeyTab})
	if m.category != "Development" || len(m.filteredApps) != 1 || m.filteredApps[0].Name != "Editor" {
		t.Fatal("category cycle")
	}
	m = updateModel(m, tea.KeyMsg{Type: tea.KeyShiftTab})
	if m.category != "" || len(m.filteredApps) != 2 {
		t.Fatal("category reverse")
	}
	m = updateModel(m, tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("/Network fire")})
	if len(m.filteredApps) != 1 || m.filteredApps[0].Name != "Firefox" {
		t.Fatal("category prefix")
	}
	m = updateModel(m, tea.KeyMsg{Type: tea.KeyCtrlBackslash})
	// Exec remains lowercase; use a name-only setting to isolate case behavior.
	m.settings.SearchFields = []string{"name"}
	m.configureSearch()
	m.updateFilteredApps()
	if len(m.filteredApps) != 0 {
		t.Fatal("case mode ignored")
	}
	m = updateModel(m, tea.KeyMsg{Type: tea.KeyCtrlBackslash})
	if len(m.filteredApps) != 1 {
		t.Fatal("case mode did not reverse")
	}
}

func TestQuickLaunchForceTerminalAndFavoritePersistence(t *testing.T) {
	m := testModel(t, 30)
	m.selectedIndex = 15
	m.ensureSelectionVisible()
	next, cmd := m.update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'2'}, Alt: true})
	if cmd == nil || cmd().(launchMsg).app.Name != m.filteredApps[m.scrollOffset+1].Name {
		t.Fatal("quick launch not viewport-relative")
	}
	if next.(Model).searchQuery != "" {
		t.Fatal("alt digit inserted text")
	}
	_, cmd = m.update(tea.KeyMsg{Type: tea.KeyEnter, Alt: true})
	if cmd == nil || !cmd().(launchMsg).app.Terminal || m.filteredApps[m.selectedIndex].Terminal {
		t.Fatal("terminal override mutated app")
	}
	_, cmd = m.update(tea.KeyMsg{Type: tea.KeyCtrlD})
	if cmd == nil {
		t.Fatal("favorite shortcut ignored")
	}
	app := cmd().(toggleFavoriteMsg).app
	m = updateModel(m, toggleFavoriteMsg{app: app})
	if !config.NewManager().IsFavorite(app.Name) {
		t.Fatal("favorite did not persist")
	}
}

func TestEmptyAndScanningStates(t *testing.T) {
	m := testModel(t, 0)
	if !strings.Contains(stripANSIForTest(m.View()), "No applications") {
		t.Fatal("empty message missing")
	}
	m.isRefreshing = true
	if !strings.Contains(stripANSIForTest(m.View()), "Scanning for applications") {
		t.Fatal("scan message missing")
	}
	m.isRefreshing = false
	m = updateModel(m, tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("missing")})
	if !strings.Contains(stripANSIForTest(m.View()), "No matches") {
		t.Fatal("no-result message missing")
	}
}
