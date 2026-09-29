package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"slices"
	"sync"

	"tui-app-launcher/internal/errors"
	"tui-app-launcher/internal/logging"
)

// Config represents the application configuration structure
type Config struct {
	Favorites []string `json:"favorites"`
}

// Manager implements the ConfigManager interface
type Manager struct {
	mu         sync.Mutex
	configPath string
	configFile string
	config     *Config
	logger     *logging.Logger
}

// NewManager creates a new configuration manager
func NewManager() *Manager {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		// Fallback to current directory if home directory is not accessible
		homeDir = "."
	}

	configPath := filepath.Join(homeDir, ".config", "tui-launcher")
	configFile := filepath.Join(configPath, "config.json")

	manager := &Manager{
		configPath: configPath,
		configFile: configFile,
		config: &Config{
			Favorites: make([]string, 0),
		},
		logger: logging.GetGlobalLogger(),
	}

	// Try to load existing configuration
	if err := manager.loadConfig(); err != nil {
		if manager.logger != nil {
			manager.logger.LogConfigOperation("load_config", false, err)
		}
	} else {
		if manager.logger != nil {
			manager.logger.LogConfigOperation("load_config", true, nil)
		}
	}

	return manager
}

// loadConfig loads configuration from file or creates default if not exists
func (m *Manager) loadConfig() error {
	// Create config directory if it doesn't exist
	if err := os.MkdirAll(m.configPath, 0755); err != nil {
		return errors.NewConfigError("failed to create config directory", err).
			WithContext("path", m.configPath)
	}

	// Check if config file exists
	if _, err := os.Stat(m.configFile); os.IsNotExist(err) {
		// Create default config file
		return m.saveConfig()
	}

	// Read existing config file
	data, err := os.ReadFile(m.configFile)
	if err != nil {
		return errors.NewConfigError("failed to read config file", err).
			WithContext("file", m.configFile)
	}

	// Parse JSON
	if err := json.Unmarshal(data, m.config); err != nil {
		// If config is corrupted, create a new default config and continue
		m.config = &Config{
			Favorites: make([]string, 0),
		}
		// Try to save the default config, but don't fail if we can't
		_ = m.saveConfig()
		return errors.NewConfigError("config file corrupted, using defaults", err).
			WithContext("file", m.configFile)
	}

	return nil
}

// saveConfig saves current configuration to file
func (m *Manager) saveConfig() error {
	data, err := json.MarshalIndent(m.config, "", "  ")
	if err != nil {
		return errors.NewConfigError("failed to marshal config", err)
	}

	if err := os.WriteFile(m.configFile, data, 0644); err != nil {
		return errors.NewConfigError("failed to write config file", err).
			WithContext("file", m.configFile)
	}

	return nil
}

// LoadFavorites loads the user's favorite applications
func (m *Manager) LoadFavorites() ([]string, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if err := m.loadConfig(); err != nil {
		// Check if this is a recoverable error
		if launcherErr, ok := err.(*errors.LauncherError); ok && launcherErr.IsRecoverable() {
			// Return empty favorites list but log the error
			return make([]string, 0), nil
		}
		return nil, err
	}
	return slices.Clone(m.config.Favorites), nil
}

// SaveFavorites saves the user's favorite applications
func (m *Manager) SaveFavorites(favorites []string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.config.Favorites = slices.Clone(favorites)
	if err := m.saveConfig(); err != nil {
		return errors.NewConfigError("failed to save favorites", err).
			WithContext("favorites_count", len(favorites))
	}
	return nil
}

// AddFavorite adds an application to favorites
func (m *Manager) AddFavorite(appName string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	// Check if already in favorites
	for _, fav := range m.config.Favorites {
		if fav == appName {
			return nil // Already in favorites
		}
	}

	// Add to favorites
	m.config.Favorites = append(m.config.Favorites, appName)

	// Save configuration
	if err := m.saveConfig(); err != nil {
		if m.logger != nil {
			m.logger.LogConfigOperation(fmt.Sprintf("add_favorite_%s", appName), false, err)
		}
		return errors.NewConfigError("failed to save favorites after adding application", err).
			WithContext("app_name", appName)
	}

	if m.logger != nil {
		m.logger.LogConfigOperation(fmt.Sprintf("add_favorite_%s", appName), true, nil)
	}

	return nil
}

// RemoveFavorite removes an application from favorites
func (m *Manager) RemoveFavorite(appName string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	// Find and remove the favorite
	for i, fav := range m.config.Favorites {
		if fav == appName {
			// Remove by slicing
			m.config.Favorites = append(m.config.Favorites[:i], m.config.Favorites[i+1:]...)

			// Save configuration
			if err := m.saveConfig(); err != nil {
				if m.logger != nil {
					m.logger.LogConfigOperation(fmt.Sprintf("remove_favorite_%s", appName), false, err)
				}
				return errors.NewConfigError("failed to save favorites after removing application", err).
					WithContext("app_name", appName)
			}

			if m.logger != nil {
				m.logger.LogConfigOperation(fmt.Sprintf("remove_favorite_%s", appName), true, nil)
			}

			return nil
		}
	}

	// Not found, but that's not an error
	return nil
}

// IsFavorite checks if an application is in favorites
func (m *Manager) IsFavorite(appName string) bool {
	m.mu.Lock()
	defer m.mu.Unlock()

	for _, fav := range m.config.Favorites {
		if fav == appName {
			return true
		}
	}
	return false
}

// SetConfigPath sets a custom config path for testing
func (m *Manager) SetConfigPath(path string) {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.configPath = path
	m.configFile = filepath.Join(path, "config.json")
}
