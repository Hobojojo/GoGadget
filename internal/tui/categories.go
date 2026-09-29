package tui

import (
	"sort"
	"strings"
	"tui-app-launcher/internal/interfaces"
)

func (m Model) categories() []string {
	seen := make(map[string]bool)
	for _, app := range m.applications {
		for _, category := range app.Categories {
			if category != "" {
				seen[category] = true
			}
		}
	}
	values := make([]string, 0, len(seen))
	for category := range seen {
		values = append(values, category)
	}
	sort.Strings(values)
	return append([]string{""}, values...)
}

func (m *Model) cycleCategory(backward bool) {
	categories := m.categories()
	index := 0
	for i, category := range categories {
		if category == m.category {
			index = i
			break
		}
	}
	if backward {
		index = (index + len(categories) - 1) % len(categories)
	} else {
		index = (index + 1) % len(categories)
	}
	m.category = categories[index]
	m.selectedIndex, m.scrollOffset = 0, 0
	m.updateFilteredApps()
}

// /Development firefox combines a category prefix with ordinary fuzzy tokens.
func (m Model) searchScope() (string, string) {
	query, category := m.searchQuery, m.category
	if strings.HasPrefix(query, "/") {
		parts := strings.SplitN(query[1:], " ", 2)
		category, query = parts[0], ""
		if len(parts) == 2 {
			query = parts[1]
		}
	}
	return query, category
}

func inCategory(app interfaces.Application, category string) bool {
	if category == "" {
		return true
	}
	for _, value := range app.Categories {
		if strings.EqualFold(value, category) {
			return true
		}
	}
	return false
}
