package tui

import (
	"strings"
	"testing"
	"time"

	tea "github.com/charmbracelet/bubbletea"
)

func mousePress(x, y int, button tea.MouseButton) tea.MouseMsg {
	return tea.MouseMsg{X: x, Y: y, Button: button, Action: tea.MouseActionPress}
}

func TestMouseSelectMatchesRenderedCells(t *testing.T) {
	for _, size := range []tea.WindowSizeMsg{{Width: 60, Height: 12}, {Width: 100, Height: 24}, {Width: 160, Height: 24}} {
		m := testModel(t, 50)
		m = updateModel(m, size)
		m.selectedIndex = 25
		m.ensureSelectionVisible()
		m.animations.stop()
		x, y := m.listOrigin()
		width := m.listWidth(m.width - 2)
		cellWidth := (width - (m.columns() - 1)) / m.columns()
		for column := 0; column < m.columns(); column++ {
			index := m.scrollOffset + m.columns() + column
			clickX, clickY := x+column*(cellWidth+1), y+1
			line := strings.Split(stripANSIForTest(m.View()), "\n")[clickY]
			if !strings.HasPrefix(string([]rune(line)[clickX:]), stripANSIForTest(m.renderApplication(index, cellWidth))) {
				t.Fatalf("click cell does not match rendered application: %q", line)
			}
			m = updateModel(m, mousePress(clickX, clickY, tea.MouseButtonLeft))
			if m.selectedIndex != index {
				t.Fatalf("size %+v column %d selected %d, want %d", size, column, m.selectedIndex, index)
			}
			assertVisible(t, m)
		}
	}
}

func TestMouseDoubleClickLaunchesOnlySameApp(t *testing.T) {
	m := testModel(t, 10)
	x, y := m.listOrigin()
	now := time.Now()
	click := mousePress(x, y+1, tea.MouseButtonLeft)
	if cmd := m.handleMouse(click, now); cmd != nil || m.selectedIndex != 1 {
		t.Fatal("first click must only select")
	}
	release := click
	release.Action = tea.MouseActionRelease
	if cmd := m.handleMouse(release, now.Add(10*time.Millisecond)); cmd != nil {
		t.Fatal("release launched application")
	}
	cmd := m.handleMouse(click, now.Add(100*time.Millisecond))
	if cmd == nil || cmd().(launchMsg).app.Name != m.filteredApps[1].Name {
		t.Fatal("double click did not launch selected application")
	}
	if cmd := m.handleMouse(click, now.Add(200*time.Millisecond)); cmd != nil {
		t.Fatal("third click launched again")
	}
	if cmd := m.handleMouse(click, now.Add(time.Second)); cmd != nil {
		t.Fatal("slow second click launched")
	}
	if cmd := m.handleMouse(mousePress(x, y+2, tea.MouseButtonLeft), now.Add(time.Second+10*time.Millisecond)); cmd != nil {
		t.Fatal("clicking different app launched")
	}
	m = updateModel(m, tea.KeyMsg{Type: tea.KeyDown})
	if !m.mouseClick.at.IsZero() {
		t.Fatal("keyboard input did not reset double-click sequence")
	}
}

func TestMouseWheelKeepsSelectionVisibleAndDoesNotWrap(t *testing.T) {
	for _, width := range []int{60, 160} {
		m := testModel(t, 50)
		m = updateModel(m, tea.WindowSizeMsg{Width: width, Height: 24})
		x, y := m.listOrigin()
		m = updateModel(m, mousePress(x, y, tea.MouseButtonWheelDown))
		if m.selectedIndex != m.columns() {
			t.Fatal("wheel did not navigate one row")
		}
		for i := 0; i < 60; i++ {
			m = updateModel(m, mousePress(x, y, tea.MouseButtonWheelDown))
			assertVisible(t, m)
		}
		if m.selectedIndex != 49 || m.scrollOffset == 0 {
			t.Fatal("wheel wrapped or failed to scroll")
		}
		for i := 0; i < 60; i++ {
			m = updateModel(m, mousePress(x, y, tea.MouseButtonWheelUp))
			assertVisible(t, m)
		}
		if m.selectedIndex != 0 || m.scrollOffset != 0 {
			t.Fatal("wheel did not clamp at first application")
		}
	}
}

func TestMouseIgnoresNonApplicationTargets(t *testing.T) {
	m := testModel(t, 3)
	m = updateModel(m, tea.WindowSizeMsg{Width: 160, Height: 24})
	x, y := m.listOrigin()
	width := m.listWidth(m.width - 2)
	for _, msg := range []tea.MouseMsg{
		mousePress(x-1, y, tea.MouseButtonLeft),               // border
		mousePress(x, y-1, tea.MouseButtonLeft),               // search/panel border
		mousePress(x+width+3, y, tea.MouseButtonLeft),         // details
		mousePress(x+(width-1)/2, y, tea.MouseButtonLeft),     // column gap
		mousePress(x+(width-1)/2+1, y+1, tea.MouseButtonLeft), // missing right-hand app
		mousePress(x, y+3, tea.MouseButtonLeft),               // blank row
		mousePress(x, y, tea.MouseButtonRight),
		{X: x, Y: y, Button: tea.MouseButtonLeft, Action: tea.MouseActionMotion},
	} {
		if cmd := m.handleMouse(msg, time.Now()); cmd != nil || m.selectedIndex != 0 {
			t.Fatalf("non-application target changed selection: %+v", msg)
		}
	}
	for _, state := range []string{"help", "transition", "empty"} {
		next := m
		switch state {
		case "help":
			next.showHelp = true
		case "transition":
			next.animations.help.active = true
		case "empty":
			next.filteredApps = nil
		}
		for _, button := range []tea.MouseButton{tea.MouseButtonLeft, tea.MouseButtonWheelDown} {
			if cmd := next.handleMouse(mousePress(x, y+1, button), time.Now()); cmd != nil || next.selectedIndex != 0 {
				t.Fatalf("mouse input changed %s", state)
			}
		}
	}
}

func TestMouseAccountsForEntranceAndErrors(t *testing.T) {
	m := testModel(t, 10)
	m = updateModel(m, tea.WindowSizeMsg{Width: 60, Height: 24})
	m.animations.list.enter(3, 0)
	x, y := m.listOrigin()
	if cmd := m.handleMouse(mousePress(x, y, tea.MouseButtonLeft), time.Now()); cmd != nil || !m.mouseClick.at.IsZero() {
		t.Fatal("entrance padding counted as application")
	}
	shift := m.entranceRows()
	m = updateModel(m, mousePress(x, y+shift+1, tea.MouseButtonLeft))
	if m.selectedIndex != 1 || m.entranceRows() != 0 {
		t.Fatal("entrance click selected wrong app or did not settle animation")
	}
	m.errorMessage = "test error"
	m.ensureSelectionVisible()
	m = updateModel(m, mousePress(x, y+m.visibleRows(), tea.MouseButtonLeft))
	if m.selectedIndex != 1 {
		t.Fatal("bottom border/error area selected an application")
	}
}
