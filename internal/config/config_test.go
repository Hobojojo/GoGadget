package config

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"tui-app-launcher/internal/errors"
)

func TestNewManager(t *testing.T) {
	t.Setenv("HOME", t.TempDir())
	manager := NewManager()

	if manager == nil {
		t.Fatal("NewManager() returned nil")
	}

	if manager.config == nil {
		t.Fatal("Manager config is nil")
	}

	if manager.config.Favorites == nil {
		t.Fatal("Manager favorites slice is nil")
	}
}

func TestLoadFavorites(t *testing.T) {
	t.Setenv("HOME", t.TempDir())
	// Create temporary directory for testing
	tempDir := t.TempDir()

	// Create manager with custom config path
	manager := &Manager{
		configPath: tempDir,
		configFile: filepath.Join(tempDir, "config.json"),
		config: &Config{
			Favorites: make([]string, 0),
		},
	}

	// Test loading favorites when no config file exists
	favorites, err := manager.LoadFavorites()
	if err != nil {
		t.Fatalf("LoadFavorites() failed: %v", err)
	}

	if len(favorites) != 0 {
		t.Errorf("Expected empty favorites list, got %d items", len(favorites))
	}

	// Add some favorites and save
	testFavorites := []string{"firefox", "code", "terminal"}
	err = manager.SaveFavorites(testFavorites)
	if err != nil {
		t.Fatalf("SaveFavorites() failed: %v", err)
	}

	// Load favorites again
	favorites, err = manager.LoadFavorites()
	if err != nil {
		t.Fatalf("LoadFavorites() failed after save: %v", err)
	}

	if len(favorites) != len(testFavorites) {
		t.Errorf("Expected %d favorites, got %d", len(testFavorites), len(favorites))
	}

	for i, expected := range testFavorites {
		if favorites[i] != expected {
			t.Errorf("Expected favorite %d to be %s, got %s", i, expected, favorites[i])
		}
	}
}

func TestSaveFavorites(t *testing.T) {
	t.Setenv("HOME", t.TempDir())
	// Create temporary directory for testing
	tempDir := t.TempDir()

	manager := &Manager{
		configPath: tempDir,
		configFile: filepath.Join(tempDir, "config.json"),
		config: &Config{
			Favorites: make([]string, 0),
		},
	}

	testFavorites := []string{"firefox", "code", "terminal"}

	// Test saving favorites
	err := manager.SaveFavorites(testFavorites)
	if err != nil {
		t.Fatalf("SaveFavorites() failed: %v", err)
	}

	// Verify config file was created
	if _, err := os.Stat(manager.configFile); os.IsNotExist(err) {
		t.Fatal("Config file was not created")
	}

	// Verify file contents
	data, err := os.ReadFile(manager.configFile)
	if err != nil {
		t.Fatalf("Failed to read config file: %v", err)
	}

	var config Config
	err = json.Unmarshal(data, &config)
	if err != nil {
		t.Fatalf("Failed to parse config file: %v", err)
	}

	if len(config.Favorites) != len(testFavorites) {
		t.Errorf("Expected %d favorites in file, got %d", len(testFavorites), len(config.Favorites))
	}

	for i, expected := range testFavorites {
		if config.Favorites[i] != expected {
			t.Errorf("Expected favorite %d to be %s, got %s", i, expected, config.Favorites[i])
		}
	}
}

func TestAddFavorite(t *testing.T) {
	t.Setenv("HOME", t.TempDir())
	// Create temporary directory for testing
	tempDir := t.TempDir()

	manager := &Manager{
		configPath: tempDir,
		configFile: filepath.Join(tempDir, "config.json"),
		config: &Config{
			Favorites: make([]string, 0),
		},
	}

	// Test adding a favorite
	err := manager.AddFavorite("firefox")
	if err != nil {
		t.Fatalf("AddFavorite() failed: %v", err)
	}

	if len(manager.config.Favorites) != 1 {
		t.Errorf("Expected 1 favorite, got %d", len(manager.config.Favorites))
	}

	if manager.config.Favorites[0] != "firefox" {
		t.Errorf("Expected favorite to be 'firefox', got '%s'", manager.config.Favorites[0])
	}

	// Test adding duplicate favorite (should not add again)
	err = manager.AddFavorite("firefox")
	if err != nil {
		t.Fatalf("AddFavorite() failed on duplicate: %v", err)
	}

	if len(manager.config.Favorites) != 1 {
		t.Errorf("Expected 1 favorite after duplicate add, got %d", len(manager.config.Favorites))
	}

	// Test adding another favorite
	err = manager.AddFavorite("code")
	if err != nil {
		t.Fatalf("AddFavorite() failed for second item: %v", err)
	}

	if len(manager.config.Favorites) != 2 {
		t.Errorf("Expected 2 favorites, got %d", len(manager.config.Favorites))
	}
}

func TestRemoveFavorite(t *testing.T) {
	t.Setenv("HOME", t.TempDir())
	// Create temporary directory for testing
	tempDir := t.TempDir()

	manager := &Manager{
		configPath: tempDir,
		configFile: filepath.Join(tempDir, "config.json"),
		config: &Config{
			Favorites: []string{"firefox", "code", "terminal"},
		},
	}

	// Test removing existing favorite
	err := manager.RemoveFavorite("code")
	if err != nil {
		t.Fatalf("RemoveFavorite() failed: %v", err)
	}

	if len(manager.config.Favorites) != 2 {
		t.Errorf("Expected 2 favorites after removal, got %d", len(manager.config.Favorites))
	}

	// Verify "code" was removed and order is preserved
	expected := []string{"firefox", "terminal"}
	for i, fav := range manager.config.Favorites {
		if fav != expected[i] {
			t.Errorf("Expected favorite %d to be %s, got %s", i, expected[i], fav)
		}
	}

	// Test removing non-existent favorite (should not error)
	err = manager.RemoveFavorite("nonexistent")
	if err != nil {
		t.Fatalf("RemoveFavorite() failed for non-existent item: %v", err)
	}

	if len(manager.config.Favorites) != 2 {
		t.Errorf("Expected 2 favorites after removing non-existent, got %d", len(manager.config.Favorites))
	}
}

func TestIsFavorite(t *testing.T) {
	t.Setenv("HOME", t.TempDir())
	manager := &Manager{
		config: &Config{
			Favorites: []string{"firefox", "code", "terminal"},
		},
	}

	// Test existing favorites
	if !manager.IsFavorite("firefox") {
		t.Error("Expected 'firefox' to be a favorite")
	}

	if !manager.IsFavorite("code") {
		t.Error("Expected 'code' to be a favorite")
	}

	if !manager.IsFavorite("terminal") {
		t.Error("Expected 'terminal' to be a favorite")
	}

	// Test non-existing favorite
	if manager.IsFavorite("nonexistent") {
		t.Error("Expected 'nonexistent' to not be a favorite")
	}

	// Test empty string
	if manager.IsFavorite("") {
		t.Error("Expected empty string to not be a favorite")
	}
}

func TestConfigFileCreation(t *testing.T) {
	t.Setenv("HOME", t.TempDir())
	// Create temporary directory for testing
	tempDir := t.TempDir()

	manager := &Manager{
		configPath: tempDir,
		configFile: filepath.Join(tempDir, "config.json"),
		config: &Config{
			Favorites: make([]string, 0),
		},
	}

	// Test that config directory is created
	err := manager.loadConfig()
	if err != nil {
		t.Fatalf("loadConfig() failed: %v", err)
	}

	// Verify config directory exists
	if _, err := os.Stat(tempDir); os.IsNotExist(err) {
		t.Fatal("Config directory was not created")
	}

	// Verify config file exists
	if _, err := os.Stat(manager.configFile); os.IsNotExist(err) {
		t.Fatal("Config file was not created")
	}

	// Verify config file has valid JSON
	data, err := os.ReadFile(manager.configFile)
	if err != nil {
		t.Fatalf("Failed to read config file: %v", err)
	}

	var config Config
	err = json.Unmarshal(data, &config)
	if err != nil {
		t.Fatalf("Config file contains invalid JSON: %v", err)
	}
}

func TestCorruptedConfigFile(t *testing.T) {
	t.Setenv("HOME", t.TempDir())
	// Create temporary directory for testing
	tempDir := t.TempDir()

	configFile := filepath.Join(tempDir, "config.json")

	// Create corrupted config file
	err := os.WriteFile(configFile, []byte("invalid json content"), 0644)
	if err != nil {
		t.Fatalf("Failed to create corrupted config file: %v", err)
	}

	manager := &Manager{
		configPath: tempDir,
		configFile: configFile,
		config: &Config{
			Favorites: make([]string, 0),
		},
	}

	// Test loading corrupted config (should create new default config)
	err = manager.loadConfig()
	launcherErr, ok := err.(*errors.LauncherError)
	if !ok || launcherErr.Type != errors.ConfigError || !launcherErr.IsRecoverable() {
		t.Fatalf("expected a recoverable corruption warning, got: %v", err)
	}

	// Verify that config was reset to default
	if len(manager.config.Favorites) != 0 {
		t.Errorf("Expected empty favorites after corrupted config, got %d items", len(manager.config.Favorites))
	}

	// Verify that config file was overwritten with valid JSON
	data, err := os.ReadFile(configFile)
	if err != nil {
		t.Fatalf("Failed to read config file after corruption recovery: %v", err)
	}

	var config Config
	err = json.Unmarshal(data, &config)
	if err != nil {
		t.Fatalf("Config file still contains invalid JSON after recovery: %v", err)
	}
}

func TestFavoritesPersistenceAcrossInstances(t *testing.T) {
	t.Setenv("HOME", t.TempDir())
	// Create temporary directory for testing
	tempDir := t.TempDir()

	configFile := filepath.Join(tempDir, "config.json")

	// Create first manager instance and add favorites
	manager1 := &Manager{
		configPath: tempDir,
		configFile: configFile,
		config: &Config{
			Favorites: make([]string, 0),
		},
	}

	testFavorites := []string{"firefox", "code", "terminal"}
	err := manager1.SaveFavorites(testFavorites)
	if err != nil {
		t.Fatalf("SaveFavorites() failed: %v", err)
	}

	// Create second manager instance (simulating app restart)
	manager2 := &Manager{
		configPath: tempDir,
		configFile: configFile,
		config: &Config{
			Favorites: make([]string, 0),
		},
	}

	// Load favorites in second instance
	favorites, err := manager2.LoadFavorites()
	if err != nil {
		t.Fatalf("LoadFavorites() failed in second instance: %v", err)
	}

	// Verify favorites persisted
	if len(favorites) != len(testFavorites) {
		t.Errorf("Expected %d favorites in second instance, got %d", len(testFavorites), len(favorites))
	}

	for i, expected := range testFavorites {
		if favorites[i] != expected {
			t.Errorf("Expected favorite %d to be %s, got %s", i, expected, favorites[i])
		}
	}
}

func TestAddRemoveFavoritePersistence(t *testing.T) {
	t.Setenv("HOME", t.TempDir())
	// Create temporary directory for testing
	tempDir := t.TempDir()

	manager := &Manager{
		configPath: tempDir,
		configFile: filepath.Join(tempDir, "config.json"),
		config: &Config{
			Favorites: make([]string, 0),
		},
	}

	// Add favorites one by one and verify persistence
	favorites := []string{"firefox", "code", "terminal"}

	for i, fav := range favorites {
		err := manager.AddFavorite(fav)
		if err != nil {
			t.Fatalf("AddFavorite(%s) failed: %v", fav, err)
		}

		// Verify it was added to memory
		if !manager.IsFavorite(fav) {
			t.Errorf("Favorite %s not found in memory after adding", fav)
		}

		// Verify it was persisted to file
		loadedFavorites, err := manager.LoadFavorites()
		if err != nil {
			t.Fatalf("LoadFavorites() failed after adding %s: %v", fav, err)
		}

		if len(loadedFavorites) != i+1 {
			t.Errorf("Expected %d favorites after adding %s, got %d", i+1, fav, len(loadedFavorites))
		}
	}

	// Remove favorites one by one and verify persistence
	for i, fav := range favorites {
		err := manager.RemoveFavorite(fav)
		if err != nil {
			t.Fatalf("RemoveFavorite(%s) failed: %v", fav, err)
		}

		// Verify it was removed from memory
		if manager.IsFavorite(fav) {
			t.Errorf("Favorite %s still found in memory after removing", fav)
		}

		// Verify it was persisted to file
		loadedFavorites, err := manager.LoadFavorites()
		if err != nil {
			t.Fatalf("LoadFavorites() failed after removing %s: %v", fav, err)
		}

		expectedCount := len(favorites) - i - 1
		if len(loadedFavorites) != expectedCount {
			t.Errorf("Expected %d favorites after removing %s, got %d", expectedCount, fav, len(loadedFavorites))
		}
	}
}
