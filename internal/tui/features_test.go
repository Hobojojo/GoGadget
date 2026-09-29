package tui

import (
	"errors"
	"strings"
	"testing"
	"time"

	"tui-app-launcher/internal/config"
	"tui-app-launcher/internal/interfaces"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/mattn/go-runewidth"
)

type historyLauncher struct{ fail bool }

func (l historyLauncher) ValidateApplication(interfaces.Application) error { return nil }
func (l historyLauncher) LaunchApplication(interfaces.Application) error {
	if l.fail {
		return errors.New("failed start")
	}
	return nil
}
func (l historyLauncher) GetLaunchCommand(interfaces.Application) (string, []string, error) {
	return "echo", nil, nil
}

func TestLaunchGestureRecordsOnlySuccess(t *testing.T) {
	for _, fail := range []bool{false, true} {
		m := testModel(t, 1)
		m.launcher = historyLauncher{fail: fail}
		app := m.filteredApps[0]
		// Exercise Enter and follow the actual command/messages, not just RecordLaunch.
		next, cmd := m.update(tea.KeyMsg{Type: tea.KeyEnter})
		if cmd == nil {
			t.Fatal("Enter did not schedule a launch")
		}
		_, result := next.(Model).update(cmd())
		if result == nil {
			t.Fatal("launch has no outcome")
		}
		msg := result()
		history := m.configManager.(interfaces.LaunchHistoryManager).LaunchHistory()
		if fail {
			if _, ok := msg.(launchErrorMsg); !ok || history[app.HistoryID()].Count != 0 {
				t.Fatal("failed launch recorded as success")
			}
		} else {
			if _, ok := msg.(launchSuccessMsg); !ok {
				t.Fatal("successful launch was not reported")
			}
			if config.NewManager().LaunchHistory()[app.HistoryID()].Count != 1 {
				t.Fatal("Enter launch history did not persist")
			}
		}
	}
}

func TestHistoryRankingAndDecay(t *testing.T) {
	m := testModel(t, 3)
	original := m.applications[2]
	manager := m.configManager.(interfaces.LaunchHistoryManager)
	if err := manager.RecordLaunch(original); err != nil {
		t.Fatal(err)
	}
	m.updateFilteredApps()
	if m.filteredApps[0].Name != original.Name {
		t.Fatal("recent app not ranked first")
	}
	if m.applications[0].Name != "Application 00" {
		t.Fatal("ranking mutated source ordering")
	}
	m.searchQuery = "Application"
	m.updateFilteredApps()
	if m.filteredApps[0].Name != original.Name {
		t.Fatal("usage should break equal search score ties")
	}
	now := time.Now()
	if frecency(interfaces.LaunchRecord{Count: 2, LastUsed: now.Add(-7 * 24 * time.Hour)}, now) != 1 {
		t.Fatal("seven-day decay")
	}
	if frecency(interfaces.LaunchRecord{Count: 2, LastUsed: now.Add(time.Hour)}, now) != 2 {
		t.Fatal("future dates must be clamped")
	}
}

func TestSearchGestureAndResponsiveDetails(t *testing.T) {
	m := testModel(t, 0)
	m.SetApplications([]interfaces.Application{
		{Name: "Firefox Developer Edition", Exec: "/opt/browser-bin", Comment: "Browse the web", Terminal: true, Categories: []string{"Internet"}, Keywords: []string{"web"}, DesktopFile: "/apps/browser.desktop"},
		{Name: "Editor", Exec: "edit", Comment: "Edit documents"},
	})
	m = updateModel(m, tea.WindowSizeMsg{Width: 120, Height: 32})
	m = updateModel(m, tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("dev browser-bin")})
	if len(m.filteredApps) != 1 || m.filteredApps[0].Name != "Firefox Developer Edition" {
		t.Fatal("typed tokens did not match name plus executable")
	}
	view := stripANSIForTest(m.View())
	for _, text := range []string{"App details", "Exec: /opt/browser-bin", "Terminal: true", "Categories: Internet", "Keywords: web", "Desktop file: /apps/browser.desktop"} {
		if !strings.Contains(view, text) {
			t.Fatalf("missing details %q", text)
		}
	}
	m = updateModel(m, tea.KeyMsg{Type: tea.KeyCtrlU})
	m = updateModel(m, tea.KeyMsg{Type: tea.KeyDown})
	if !strings.Contains(stripANSIForTest(m.View()), "Edit documents") {
		t.Fatal("details did not follow navigation")
	}
	for _, width := range []int{40, 80, 81, 100, 120} {
		for _, height := range []int{10, 12, 24, 32} {
			m = updateModel(m, tea.WindowSizeMsg{Width: width, Height: height})
			view := stripANSIForTest(m.View())
			lines := strings.Split(view, "\n")
			if len(lines) != height {
				t.Fatalf("%dx%d: rendered %d rows", width, height, len(lines))
			}
			for _, line := range lines {
				if got := runewidth.StringWidth(line); got != width {
					t.Fatalf("%dx%d: line width %d", width, height, got)
				}
			}
			if width <= 80 && strings.Contains(view, "App details") {
				t.Fatal("details should collapse on narrow terminals")
			}
		}
	}
}
