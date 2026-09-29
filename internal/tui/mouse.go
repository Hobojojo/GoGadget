package tui

import (
	"time"

	tea "github.com/charmbracelet/bubbletea"
)

const doubleClickInterval = 400 * time.Millisecond

type mouseClickState struct {
	index int
	at    time.Time
}

// listOrigin uses zero-based terminal cells, including both panel borders.
func (m Model) listOrigin() (int, int) {
	y := 7 // Outer border, three-row header, rule, search, panel border.
	if m.height < 14 {
		y = 5 // Compact header occupies one row.
	}
	return 2, y
}

func (m *Model) handleMouse(msg tea.MouseMsg, now time.Time) tea.Cmd {
	if msg.Action != tea.MouseActionPress {
		return nil // Releases and dragging must never count as another click.
	}
	if m.showHelp || m.animations.help.active || msg.Ctrl || msg.Alt || msg.Shift {
		m.mouseClick = mouseClickState{}
		return nil
	}
	x, y := m.listOrigin()
	width := m.listWidth(m.width - 2)
	if msg.X < x || msg.X >= x+width || msg.Y < y || msg.Y >= y+m.visibleRows() || len(m.filteredApps) == 0 {
		m.mouseClick = mouseClickState{}
		return nil
	}

	switch msg.Button {
	case tea.MouseButtonWheelUp, tea.MouseButtonWheelDown:
		m.mouseClick = mouseClickState{}
		step := m.columns()
		if msg.Button == tea.MouseButtonWheelUp {
			step = -step
		}
		m.selectedIndex = max(0, min(len(m.filteredApps)-1, m.selectedIndex+step))
		m.ensureSelectionVisible()
		m.animations.list.snap(0)
	case tea.MouseButtonLeft:
		columns := m.columns()
		column := 0
		if columns == 2 {
			cellWidth := (width - 1) / 2
			if msg.X-x == cellWidth {
				m.mouseClick = mouseClickState{}
				return nil // The gap between columns is not an application.
			}
			if msg.X-x > cellWidth {
				column = 1
			}
		}
		row := msg.Y - y - m.entranceRows()
		index := m.scrollOffset + row*columns + column
		if row < 0 || index >= len(m.filteredApps) {
			m.mouseClick = mouseClickState{}
			return nil
		}
		m.selectedIndex = index
		m.ensureSelectionVisible()
		m.animations.list.snap(0)
		elapsed := now.Sub(m.mouseClick.at)
		if !m.mouseClick.at.IsZero() && m.mouseClick.index == index && elapsed >= 0 && elapsed <= doubleClickInterval {
			m.mouseClick = mouseClickState{}
			return m.launchApplication(m.filteredApps[index])
		}
		m.mouseClick = mouseClickState{index: index, at: now}
	default:
		m.mouseClick = mouseClickState{}
	}
	return nil
}
