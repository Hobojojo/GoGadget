package tui

// visibleRows accounts for the outer frame, banner, aurora rule, search,
// bordered list panel and footer. Compact terminals use single-row banners.
func (m *Model) visibleRows() int {
	reserved := 12
	if m.height < 14 {
		reserved = 8
	}
	if m.showHelp {
		reserved-- // Help has no search field.
	} else if m.errorMessage != "" {
		reserved++
	}
	return max(0, m.height-reserved)
}

// ensureSelectionVisible maintains a bounded window without changing selection
// on resize. Navigation wraps at either end of the full list, not the viewport.
func (m *Model) ensureSelectionVisible() {
	if len(m.filteredApps) == 0 {
		m.selectedIndex = 0
		m.scrollOffset = 0
		return
	}
	m.selectedIndex = max(0, min(m.selectedIndex, len(m.filteredApps)-1))
	rows := max(1, m.visibleRows())
	if m.selectedIndex < m.scrollOffset {
		m.scrollOffset = m.selectedIndex
	} else if m.selectedIndex >= m.scrollOffset+rows {
		m.scrollOffset = m.selectedIndex - rows + 1
	}
	m.scrollOffset = max(0, min(m.scrollOffset, len(m.filteredApps)-rows))
}
