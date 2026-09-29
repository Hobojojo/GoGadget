package tui

// visibleRows reserves five header rows and three footer rows. Errors use two
// additional rows. A very short terminal may have no room for application rows.
func (m *Model) visibleRows() int {
	reserved := 8
	if m.errorMessage != "" {
		reserved += 2
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
