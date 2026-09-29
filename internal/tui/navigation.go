package tui

import (
	tea "github.com/charmbracelet/bubbletea"
	"tui-app-launcher/internal/interfaces"
	"unicode"
)

func (m Model) columns() int {
	if m.width > 100 && (m.settings.MultiColumn == nil || *m.settings.MultiColumn) {
		return 2
	}
	return 1
}

func (m Model) visibleCapacity() int { return max(1, m.visibleRows()) * m.columns() }

// handleShortcut centralizes editing/navigation without changing launch delivery.
func (m *Model) handleShortcut(key tea.KeyMsg) (bool, tea.Cmd) {
	name := key.String()
	runes := []rune(m.searchQuery)
	m.searchCursorIndex = min(max(0, m.searchCursorIndex), len(runes))
	cursor := m.searchCursorIndex
	changed := false
	switch name {
	case "home":
		m.selectedIndex = 0
	case "end":
		m.selectedIndex = max(0, len(m.filteredApps)-1)
	case "pgup":
		m.selectedIndex = max(0, m.selectedIndex-m.visibleCapacity())
	case "pgdown":
		m.selectedIndex = min(max(0, len(m.filteredApps)-1), m.selectedIndex+m.visibleCapacity())
	case "ctrl+a":
		m.searchCursorIndex = 0
	case "ctrl+e":
		m.searchCursorIndex = len(runes)
	case "left":
		m.searchCursorIndex = max(0, cursor-1)
	case "right":
		m.searchCursorIndex = min(len(runes), cursor+1)
	case "ctrl+w":
		start := cursor
		for start > 0 && unicode.IsSpace(runes[start-1]) {
			start--
		}
		for start > 0 && !unicode.IsSpace(runes[start-1]) {
			start--
		}
		runes = append(runes[:start], runes[cursor:]...)
		m.searchCursorIndex, changed = start, true
	case "backspace":
		if cursor > 0 {
			runes = append(runes[:cursor-1], runes[cursor:]...)
			m.searchCursorIndex--
			changed = true
		}
	case "delete":
		if cursor < len(runes) {
			runes = append(runes[:cursor], runes[cursor+1:]...)
			changed = true
		}
	case "ctrl+u":
		runes, m.searchCursorIndex, changed = nil, 0, true
	case "tab", "ctrl+tab":
		m.cycleCategory(false)
	case "shift+tab":
		m.cycleCategory(true)
	case "ctrl+\\":
		m.caseSensitive = !m.caseSensitive
		m.configureSearch()
		m.updateFilteredApps()
	case "ctrl+d", "ctrl+@":
		if len(m.filteredApps) > 0 {
			return true, m.toggleFavorite(m.filteredApps[m.selectedIndex])
		}
	case "alt+enter", "shift+enter":
		if len(m.filteredApps) > 0 {
			app := m.filteredApps[m.selectedIndex]
			app.Terminal = true
			return true, m.launchApplication(app)
		}
	default:
		if key.Alt && key.Type == tea.KeyRunes && len(key.Runes) == 1 && key.Runes[0] >= '1' && key.Runes[0] <= '9' {
			index := m.scrollOffset + int(key.Runes[0]-'1')
			if index < len(m.filteredApps) && index < m.scrollOffset+m.visibleCapacity() {
				return true, m.launchApplication(m.filteredApps[index])
			}
			return true, nil
		}
		return false, nil
	}
	if changed {
		m.searchQuery = string(runes)
		m.updateFilteredApps()
	}
	m.ensureSelectionVisible()
	return true, nil
}

func (m *Model) configureSearch() {
	if searcher, ok := m.fuzzySearcher.(interfaces.SearchConfigurer); ok {
		searcher.ConfigureSearch(m.settings.SearchFields, m.caseSensitive)
	}
}
