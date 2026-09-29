package tui

import (
	"fmt"
	"strings"
	"testing"

	"tui-app-launcher/internal/config"
	"tui-app-launcher/internal/interfaces"
	"tui-app-launcher/internal/search"

	tea "github.com/charmbracelet/bubbletea"
)

func testModel(t *testing.T, count int) Model {
	t.Helper()
	t.Setenv("HOME", t.TempDir())
	m := NewModel(search.NewFuzzySearcher(), config.NewManager(), nil, nil)
	apps := make([]interfaces.Application, count)
	for i := range apps {
		apps[i] = interfaces.Application{Name: fmt.Sprintf("Application %02d", i), Exec: "echo"}
	}
	m.SetApplications(apps)
	return updateModel(m, tea.WindowSizeMsg{Width: 100, Height: 12})
}

func updateModel(m Model, msg tea.Msg) Model {
	next, _ := m.Update(msg)
	return next.(Model)
}

func assertVisible(t *testing.T, m Model) {
	t.Helper()
	if len(m.filteredApps) > 0 {
		name := m.filteredApps[m.selectedIndex].Name
		if !strings.Contains(m.View(), "> "+name) {
			t.Fatalf("selected %q is not visibly highlighted (offset %d)", name, m.scrollOffset)
		}
	}
	if rows := len(strings.Split(m.View(), "\n")); rows != m.height {
		t.Fatalf("rendered %d rows in a %d-row terminal", rows, m.height)
	}
}

func TestNavigationScrollsAndWraps(t *testing.T) {
	m := testModel(t, 30)
	for i := 1; i <= 30; i++ {
		m = updateModel(m, tea.KeyMsg{Type: tea.KeyDown})
		if m.selectedIndex != i%30 {
			t.Fatalf("wrong selected index: %d", m.selectedIndex)
		}
		assertVisible(t, m)
	}
	for i := 29; i >= 0; i-- {
		m = updateModel(m, tea.KeyMsg{Type: tea.KeyUp})
		if m.selectedIndex != i {
			t.Fatalf("wrong selected index: %d", m.selectedIndex)
		}
		assertVisible(t, m)
	}
}

func TestResizeKeepsSelectedApplication(t *testing.T) {
	m := testModel(t, 30)
	for i := 0; i < 20; i++ {
		m = updateModel(m, tea.KeyMsg{Type: tea.KeyDown})
	}
	for _, height := range []int{10, 20, 40, 12} {
		m = updateModel(m, tea.WindowSizeMsg{Width: 100, Height: height})
		if m.selectedIndex != 20 {
			t.Fatalf("resize changed selection to %d", m.selectedIndex)
		}
		assertVisible(t, m)
	}
}

func TestFilteringAndRefreshClampViewport(t *testing.T) {
	m := testModel(t, 30)
	m = updateModel(m, tea.KeyMsg{Type: tea.KeyUp})
	for _, char := range "29" {
		m = updateModel(m, tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{char}})
	}
	if len(m.filteredApps) != 1 || m.scrollOffset != 0 || m.selectedIndex != 0 {
		t.Fatal("filtering did not reset the viewport for a single result")
	}
	m = updateModel(m, tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'z'}})
	if len(m.filteredApps) != 0 || m.scrollOffset != 0 {
		t.Fatal("empty results left a stale viewport")
	}
	assertVisible(t, m)
	m = updateModel(m, tea.KeyMsg{Type: tea.KeyCtrlU})
	assertVisible(t, m)
	m = updateModel(m, tea.KeyMsg{Type: tea.KeyUp})
	m = updateModel(m, refreshCompleteMsg{apps: []interfaces.Application{{Name: "Replacement"}}})
	assertVisible(t, m)
}

func TestErrorRowsKeepSelectionVisible(t *testing.T) {
	m := testModel(t, 30)
	m = updateModel(m, tea.KeyMsg{Type: tea.KeyUp})
	m = updateModel(m, launchErrorMsg{err: fmt.Errorf("test launch error")})
	assertVisible(t, m)
	m = updateModel(m, tea.KeyMsg{Type: tea.KeyCtrlL})
	assertVisible(t, m)
}

func TestNavigationWithNoApplications(t *testing.T) {
	m := testModel(t, 0)
	for _, key := range []tea.KeyType{tea.KeyUp, tea.KeyDown} {
		m = updateModel(m, tea.KeyMsg{Type: key})
		if m.selectedIndex != 0 || m.scrollOffset != 0 {
			t.Fatal("empty list navigation changed selection")
		}
		assertVisible(t, m)
	}
}
