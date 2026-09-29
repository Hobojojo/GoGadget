package config

import (
	"slices"
	"tui-app-launcher/internal/interfaces"
)

// Settings returns a detached snapshot. Missing settings retain legacy defaults.
func (m *Manager) Settings() interfaces.Settings {
	m.mu.Lock()
	defer m.mu.Unlock()
	settings := m.config.Settings
	settings.SearchFields = slices.Clone(settings.SearchFields)
	if settings.MultiColumn != nil {
		value := *settings.MultiColumn
		settings.MultiColumn = &value
	}
	return settings
}
