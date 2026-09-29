package config

import (
	"maps"
	"time"

	"tui-app-launcher/internal/interfaces"
)

// LaunchHistory returns a snapshot; callers cannot mutate the stored history.
func (m *Manager) LaunchHistory() map[string]interfaces.LaunchRecord {
	m.mu.Lock()
	defer m.mu.Unlock()
	return maps.Clone(m.config.LaunchHistory)
}

// RecordLaunch commits successful launches atomically and keeps memory unchanged
// if persistence fails. Old favorites-only configs need no migration.
func (m *Manager) RecordLaunch(app interfaces.Application) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	previous := m.config.LaunchHistory
	history := maps.Clone(previous)
	if history == nil {
		history = make(map[string]interfaces.LaunchRecord)
	}
	id := app.HistoryID()
	record := history[id]
	record.Count++
	record.LastUsed = time.Now().UTC()
	history[id] = record
	m.config.LaunchHistory = history
	if err := m.saveConfig(); err != nil {
		m.config.LaunchHistory = previous
		return err
	}
	return nil
}
