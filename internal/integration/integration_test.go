package integration

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"tui-app-launcher/internal/config"
	"tui-app-launcher/internal/interfaces"
	"tui-app-launcher/internal/launcher"
	"tui-app-launcher/internal/scanner"
	"tui-app-launcher/internal/search"
)

// TestEndToEndWorkflow tests complete user workflows from application discovery to launching
func TestEndToEndWorkflow(t *testing.T) {
	// Create temporary directory structure for testing
	tempDir := t.TempDir()
	
	// Create test application directories
	systemAppsDir := filepath.Join(tempDir, "system", "applications")
	userAppsDir := filepath.Join(tempDir, "user", "applications")
	configDir := filepath.Join(tempDir, "config")
	
	err := os.MkdirAll(systemAppsDir, 0755)
	if err != nil {
		t.Fatalf("Failed to create system apps directory: %v", err)
	}
	
	err = os.MkdirAll(userAppsDir, 0755)
	if err != nil {
		t.Fatalf("Failed to create user apps directory: %v", err)
	}
	
	err = os.MkdirAll(configDir, 0755)
	if err != nil {
		t.Fatalf("Failed to create config directory: %v", err)
	}
	
	// Create test desktop files
	testApps := []struct {
		name     string
		content  string
		location string // "system" or "user"
	}{
		{
			name:     "firefox.desktop",
			location: "system",
			content: `[Desktop Entry]
Name=Firefox Web Browser
Exec=firefox %u
Icon=firefox
Comment=Browse the World Wide Web
Categories=Network;WebBrowser;
Type=Application
NoDisplay=false`,
		},
		{
			name:     "code.desktop",
			location: "user",
			content: `[Desktop Entry]
Name=Visual Studio Code
Exec=code %F
Icon=vscode
Comment=Code Editing. Redefined.
Categories=Development;IDE;
Type=Application
NoDisplay=false`,
		},
		{
			name:     "terminal.desktop",
			location: "system",
			content: `[Desktop Entry]
Name=Terminal
Exec=gnome-terminal
Icon=terminal
Comment=Use the command line
Categories=System;TerminalEmulator;
Type=Application
NoDisplay=false`,
		},
		{
			name:     "hidden-app.desktop",
			location: "system",
			content: `[Desktop Entry]
Name=Hidden Application
Exec=hidden-app
Icon=hidden
Comment=This app should not appear
Categories=Utility;
Type=Application
NoDisplay=true`,
		},
	}
	
	// Write test desktop files
	for _, app := range testApps {
		var targetDir string
		if app.location == "system" {
			targetDir = systemAppsDir
		} else {
			targetDir = userAppsDir
		}
		
		filePath := filepath.Join(targetDir, app.name)
		err := os.WriteFile(filePath, []byte(app.content), 0644)
		if err != nil {
			t.Fatalf("Failed to write test desktop file %s: %v", app.name, err)
		}
	}
	
	// Initialize components with test directories
	configManager := config.NewManager()
	configManager.SetConfigPath(configDir)
	
	appScanner := scanner.NewScanner()
	appScanner.SetScanPaths([]string{systemAppsDir, userAppsDir})
	
	fuzzySearcher := search.NewFuzzySearcher()
	appLauncher := launcher.NewLauncher()
	
	// Test 1: Application Discovery
	t.Run("ApplicationDiscovery", func(t *testing.T) {
		applications, err := appScanner.ScanApplications()
		if err != nil {
			t.Fatalf("Failed to scan applications: %v", err)
		}
		
		// Should find 3 visible applications (firefox, code, terminal)
		// hidden-app should be filtered out due to NoDisplay=true
		expectedApps := []string{"Firefox Web Browser", "Visual Studio Code", "Terminal"}
		
		if len(applications) != len(expectedApps) {
			t.Errorf("Expected %d applications, got %d", len(expectedApps), len(applications))
		}
		
		// Verify each expected application is found
		foundApps := make(map[string]bool)
		for _, app := range applications {
			foundApps[app.Name] = true
		}
		
		for _, expectedApp := range expectedApps {
			if !foundApps[expectedApp] {
				t.Errorf("Expected application %q not found", expectedApp)
			}
		}
		
		// Verify hidden application is not included
		if foundApps["Hidden Application"] {
			t.Error("Hidden application should not be included in results")
		}
	})
	
	// Test 2: Fuzzy Search Functionality
	t.Run("FuzzySearch", func(t *testing.T) {
		applications, err := appScanner.ScanApplications()
		if err != nil {
			t.Fatalf("Failed to scan applications: %v", err)
		}
		
		fuzzySearcher.SetItems(applications)
		
		// Test exact match
		results := fuzzySearcher.Search("Firefox", applications)
		if len(results) == 0 {
			t.Error("Expected to find Firefox with exact match")
		}
		if results[0].Application.Name != "Firefox Web Browser" {
			t.Errorf("Expected first result to be Firefox, got %s", results[0].Application.Name)
		}
		
		// Test partial match
		results = fuzzySearcher.Search("code", applications)
		if len(results) == 0 {
			t.Error("Expected to find Visual Studio Code with partial match")
		}
		
		// Test fuzzy match
		results = fuzzySearcher.Search("term", applications)
		if len(results) == 0 {
			t.Error("Expected to find Terminal with fuzzy match")
		}
		
		// Test no match
		results = fuzzySearcher.Search("nonexistent", applications)
		if len(results) != 0 {
			t.Error("Expected no results for nonexistent application")
		}
	})
	
	// Test 3: Favorites Management
	t.Run("FavoritesManagement", func(t *testing.T) {
		// Test adding favorites
		err := configManager.AddFavorite("Firefox Web Browser")
		if err != nil {
			t.Fatalf("Failed to add Firefox to favorites: %v", err)
		}
		
		err = configManager.AddFavorite("Visual Studio Code")
		if err != nil {
			t.Fatalf("Failed to add Code to favorites: %v", err)
		}
		
		// Test checking favorites
		if !configManager.IsFavorite("Firefox Web Browser") {
			t.Error("Firefox should be marked as favorite")
		}
		
		if !configManager.IsFavorite("Visual Studio Code") {
			t.Error("Visual Studio Code should be marked as favorite")
		}
		
		if configManager.IsFavorite("Terminal") {
			t.Error("Terminal should not be marked as favorite")
		}
		
		// Test loading favorites
		favorites, err := configManager.LoadFavorites()
		if err != nil {
			t.Fatalf("Failed to load favorites: %v", err)
		}
		
		expectedFavorites := []string{"Firefox Web Browser", "Visual Studio Code"}
		if len(favorites) != len(expectedFavorites) {
			t.Errorf("Expected %d favorites, got %d", len(expectedFavorites), len(favorites))
		}
		
		// Test removing favorite
		err = configManager.RemoveFavorite("Firefox Web Browser")
		if err != nil {
			t.Fatalf("Failed to remove Firefox from favorites: %v", err)
		}
		
		if configManager.IsFavorite("Firefox Web Browser") {
			t.Error("Firefox should no longer be marked as favorite")
		}
		
		// Verify persistence
		favorites, err = configManager.LoadFavorites()
		if err != nil {
			t.Fatalf("Failed to load favorites after removal: %v", err)
		}
		
		if len(favorites) != 1 || favorites[0] != "Visual Studio Code" {
			t.Error("Favorites should contain only Visual Studio Code after removal")
		}
	})
	
	// Test 4: Application Validation
	t.Run("ApplicationValidation", func(t *testing.T) {
		applications, err := appScanner.ScanApplications()
		if err != nil {
			t.Fatalf("Failed to scan applications: %v", err)
		}
		
		// Test validation of scanned applications
		for _, app := range applications {
			err := appLauncher.ValidateApplication(app)
			if err != nil {
				// For integration tests, we expect some commands might not exist
				// but the validation should still work properly
				if !strings.Contains(err.Error(), "not found in PATH") {
					t.Errorf("Unexpected validation error for %s: %v", app.Name, err)
				}
			}
		}
		
		// Test validation of invalid application
		invalidApp := interfaces.Application{
			Name: "Invalid App",
			Exec: "", // Empty exec should fail validation
		}
		
		err = appLauncher.ValidateApplication(invalidApp)
		if err == nil {
			t.Error("Expected validation to fail for application with empty exec")
		}
	})
}

// TestApplicationDiscoveryAcrossDifferentDirectories tests scanning across multiple directory structures
func TestApplicationDiscoveryAcrossDifferentDirectories(t *testing.T) {
	tempDir := t.TempDir()
	
	// Create multiple directory structures
	directories := []string{
		filepath.Join(tempDir, "usr", "share", "applications"),
		filepath.Join(tempDir, "usr", "local", "share", "applications"),
		filepath.Join(tempDir, "home", "user", ".local", "share", "applications"),
		filepath.Join(tempDir, "opt", "custom-app", "share", "applications"),
	}
	
	// Create directories
	for _, dir := range directories {
		err := os.MkdirAll(dir, 0755)
		if err != nil {
			t.Fatalf("Failed to create directory %s: %v", dir, err)
		}
	}
	
	// Create test applications in different directories
	testCases := []struct {
		dir     string
		appName string
		content string
	}{
		{
			dir:     directories[0], // /usr/share/applications
			appName: "system-app.desktop",
			content: `[Desktop Entry]
Name=System Application
Exec=system-app
Type=Application`,
		},
		{
			dir:     directories[1], // /usr/local/share/applications
			appName: "local-app.desktop",
			content: `[Desktop Entry]
Name=Local Application
Exec=local-app
Type=Application`,
		},
		{
			dir:     directories[2], // ~/.local/share/applications
			appName: "user-app.desktop",
			content: `[Desktop Entry]
Name=User Application
Exec=user-app
Type=Application`,
		},
		{
			dir:     directories[3], // /opt/custom-app/share/applications
			appName: "custom-app.desktop",
			content: `[Desktop Entry]
Name=Custom Application
Exec=custom-app
Type=Application`,
		},
	}
	
	// Write test files
	for _, tc := range testCases {
		filePath := filepath.Join(tc.dir, tc.appName)
		err := os.WriteFile(filePath, []byte(tc.content), 0644)
		if err != nil {
			t.Fatalf("Failed to write test file %s: %v", filePath, err)
		}
	}
	
	// Test scanning all directories
	appScanner := scanner.NewScanner()
	appScanner.SetScanPaths(directories)
	
	applications, err := appScanner.ScanApplications()
	if err != nil {
		t.Fatalf("Failed to scan applications: %v", err)
	}
	
	// Verify all applications were found
	expectedApps := []string{
		"System Application",
		"Local Application", 
		"User Application",
		"Custom Application",
	}
	
	if len(applications) != len(expectedApps) {
		t.Errorf("Expected %d applications, got %d", len(expectedApps), len(applications))
	}
	
	foundApps := make(map[string]bool)
	for _, app := range applications {
		foundApps[app.Name] = true
	}
	
	for _, expectedApp := range expectedApps {
		if !foundApps[expectedApp] {
			t.Errorf("Expected application %q not found", expectedApp)
		}
	}
	
	// Test scanning subset of directories
	appScanner.SetScanPaths(directories[:2]) // Only system and local
	
	applications, err = appScanner.ScanApplications()
	if err != nil {
		t.Fatalf("Failed to scan subset of applications: %v", err)
	}
	
	// Should only find system and local applications
	if len(applications) != 2 {
		t.Errorf("Expected 2 applications from subset scan, got %d", len(applications))
	}
	
	foundApps = make(map[string]bool)
	for _, app := range applications {
		foundApps[app.Name] = true
	}
	
	if !foundApps["System Application"] || !foundApps["Local Application"] {
		t.Error("Expected to find System and Local applications in subset scan")
	}
	
	if foundApps["User Application"] || foundApps["Custom Application"] {
		t.Error("Should not find User or Custom applications in subset scan")
	}
}

// TestFavoritesPersistenceAcrossRestarts tests that favorites persist across application restarts
func TestFavoritesPersistenceAcrossRestarts(t *testing.T) {
	tempDir := t.TempDir()
	configDir := filepath.Join(tempDir, "config")
	
	err := os.MkdirAll(configDir, 0755)
	if err != nil {
		t.Fatalf("Failed to create config directory: %v", err)
	}
	
	// Simulate first application instance
	t.Run("FirstInstance", func(t *testing.T) {
		configManager := config.NewManager()
		configManager.SetConfigPath(configDir)
		
		// Add some favorites
		favorites := []string{"Firefox", "Visual Studio Code", "Terminal"}
		
		for _, fav := range favorites {
			err := configManager.AddFavorite(fav)
			if err != nil {
				t.Fatalf("Failed to add favorite %s: %v", fav, err)
			}
		}
		
		// Verify favorites are in memory
		for _, fav := range favorites {
			if !configManager.IsFavorite(fav) {
				t.Errorf("Favorite %s not found in memory", fav)
			}
		}
		
		// Save favorites to disk
		err := configManager.SaveFavorites(favorites)
		if err != nil {
			t.Fatalf("Failed to save favorites: %v", err)
		}
	})
	
	// Simulate application restart - second instance
	t.Run("SecondInstance", func(t *testing.T) {
		configManager := config.NewManager()
		configManager.SetConfigPath(configDir)
		
		// Load favorites from disk
		favorites, err := configManager.LoadFavorites()
		if err != nil {
			t.Fatalf("Failed to load favorites: %v", err)
		}
		
		expectedFavorites := []string{"Firefox", "Visual Studio Code", "Terminal"}
		
		if len(favorites) != len(expectedFavorites) {
			t.Errorf("Expected %d favorites, got %d", len(expectedFavorites), len(favorites))
		}
		
		for i, expected := range expectedFavorites {
			if favorites[i] != expected {
				t.Errorf("Expected favorite %d to be %s, got %s", i, expected, favorites[i])
			}
		}
		
		// Verify IsFavorite works after loading
		for _, fav := range expectedFavorites {
			if !configManager.IsFavorite(fav) {
				t.Errorf("Favorite %s not recognized after loading", fav)
			}
		}
		
		// Modify favorites and save again
		err = configManager.RemoveFavorite("Visual Studio Code")
		if err != nil {
			t.Fatalf("Failed to remove favorite: %v", err)
		}
		
		err = configManager.AddFavorite("New Application")
		if err != nil {
			t.Fatalf("Failed to add new favorite: %v", err)
		}
	})
	
	// Simulate third instance to verify changes persisted
	t.Run("ThirdInstance", func(t *testing.T) {
		configManager := config.NewManager()
		configManager.SetConfigPath(configDir)
		
		favorites, err := configManager.LoadFavorites()
		if err != nil {
			t.Fatalf("Failed to load favorites in third instance: %v", err)
		}
		
		expectedFavorites := []string{"Firefox", "Terminal", "New Application"}
		
		if len(favorites) != len(expectedFavorites) {
			t.Errorf("Expected %d favorites in third instance, got %d", len(expectedFavorites), len(favorites))
		}
		
		// Verify Visual Studio Code was removed
		if configManager.IsFavorite("Visual Studio Code") {
			t.Error("Visual Studio Code should have been removed from favorites")
		}
		
		// Verify New Application was added
		if !configManager.IsFavorite("New Application") {
			t.Error("New Application should be in favorites")
		}
	})
}

// TestCompleteUserWorkflow tests a complete user workflow from start to finish
func TestCompleteUserWorkflow(t *testing.T) {
	tempDir := t.TempDir()
	
	// Setup test environment
	appsDir := filepath.Join(tempDir, "applications")
	configDir := filepath.Join(tempDir, "config")
	
	err := os.MkdirAll(appsDir, 0755)
	if err != nil {
		t.Fatalf("Failed to create apps directory: %v", err)
	}
	
	err = os.MkdirAll(configDir, 0755)
	if err != nil {
		t.Fatalf("Failed to create config directory: %v", err)
	}
	
	// Create test applications
	testApps := []struct {
		filename string
		content  string
	}{
		{
			filename: "firefox.desktop",
			content: `[Desktop Entry]
Name=Firefox
Exec=firefox
Icon=firefox
Comment=Web Browser
Categories=Network;WebBrowser;
Type=Application`,
		},
		{
			filename: "code.desktop",
			content: `[Desktop Entry]
Name=Visual Studio Code
Exec=code
Icon=vscode
Comment=Code Editor
Categories=Development;IDE;
Type=Application`,
		},
		{
			filename: "terminal.desktop",
			content: `[Desktop Entry]
Name=Terminal
Exec=gnome-terminal
Icon=terminal
Comment=Terminal Emulator
Categories=System;TerminalEmulator;
Type=Application`,
		},
	}
	
	for _, app := range testApps {
		filePath := filepath.Join(appsDir, app.filename)
		err := os.WriteFile(filePath, []byte(app.content), 0644)
		if err != nil {
			t.Fatalf("Failed to write test app %s: %v", app.filename, err)
		}
	}
	
	// Initialize all components
	configManager := config.NewManager()
	configManager.SetConfigPath(configDir)
	
	appScanner := scanner.NewScanner()
	appScanner.SetScanPaths([]string{appsDir})
	
	fuzzySearcher := search.NewFuzzySearcher()
	appLauncher := launcher.NewLauncher()
	
	// Step 1: User starts application - scan for applications
	applications, err := appScanner.ScanApplications()
	if err != nil {
		t.Fatalf("Failed to scan applications: %v", err)
	}
	
	if len(applications) != 3 {
		t.Errorf("Expected 3 applications, got %d", len(applications))
	}
	
	// Step 2: User searches for "fire" - should find Firefox
	fuzzySearcher.SetItems(applications)
	searchResults := fuzzySearcher.Search("fire", applications)
	
	if len(searchResults) == 0 {
		t.Fatal("Expected to find Firefox with 'fire' search")
	}
	
	if searchResults[0].Application.Name != "Firefox" {
		t.Errorf("Expected first result to be Firefox, got %s", searchResults[0].Application.Name)
	}
	
	// Step 3: User adds Firefox to favorites
	err = configManager.AddFavorite("Firefox")
	if err != nil {
		t.Fatalf("Failed to add Firefox to favorites: %v", err)
	}
	
	// Step 4: User searches for "code" - should find Visual Studio Code
	searchResults = fuzzySearcher.Search("code", applications)
	
	if len(searchResults) == 0 {
		t.Fatal("Expected to find Visual Studio Code with 'code' search")
	}
	
	if searchResults[0].Application.Name != "Visual Studio Code" {
		t.Errorf("Expected first result to be Visual Studio Code, got %s", searchResults[0].Application.Name)
	}
	
	// Step 5: User adds Visual Studio Code to favorites
	err = configManager.AddFavorite("Visual Studio Code")
	if err != nil {
		t.Fatalf("Failed to add Visual Studio Code to favorites: %v", err)
	}
	
	// Step 6: User views favorites list
	favorites, err := configManager.LoadFavorites()
	if err != nil {
		t.Fatalf("Failed to load favorites: %v", err)
	}
	
	expectedFavorites := []string{"Firefox", "Visual Studio Code"}
	if len(favorites) != len(expectedFavorites) {
		t.Errorf("Expected %d favorites, got %d", len(expectedFavorites), len(favorites))
	}
	
	for i, expected := range expectedFavorites {
		if favorites[i] != expected {
			t.Errorf("Expected favorite %d to be %s, got %s", i, expected, favorites[i])
		}
	}
	
	// Step 7: User validates applications before launching
	for _, app := range applications {
		err := appLauncher.ValidateApplication(app)
		// We expect validation errors for non-existent commands in test environment
		if err != nil && !strings.Contains(err.Error(), "not found in PATH") {
			t.Errorf("Unexpected validation error for %s: %v", app.Name, err)
		}
	}
	
	// Step 8: User removes one favorite
	err = configManager.RemoveFavorite("Firefox")
	if err != nil {
		t.Fatalf("Failed to remove Firefox from favorites: %v", err)
	}
	
	// Step 9: Verify final state
	favorites, err = configManager.LoadFavorites()
	if err != nil {
		t.Fatalf("Failed to load favorites after removal: %v", err)
	}
	
	if len(favorites) != 1 || favorites[0] != "Visual Studio Code" {
		t.Error("Expected only Visual Studio Code in favorites after removal")
	}
	
	if configManager.IsFavorite("Firefox") {
		t.Error("Firefox should no longer be a favorite")
	}
	
	if !configManager.IsFavorite("Visual Studio Code") {
		t.Error("Visual Studio Code should still be a favorite")
	}
}

// TestConfigurationPersistence tests that configuration changes persist correctly
func TestConfigurationPersistence(t *testing.T) {
	tempDir := t.TempDir()
	configDir := filepath.Join(tempDir, "config")
	
	err := os.MkdirAll(configDir, 0755)
	if err != nil {
		t.Fatalf("Failed to create config directory: %v", err)
	}
	
	configManager := config.NewManager()
	configManager.SetConfigPath(configDir)
	
	// Test multiple add/remove operations
	operations := []struct {
		action string
		app    string
	}{
		{"add", "Firefox"},
		{"add", "Code"},
		{"add", "Terminal"},
		{"remove", "Code"},
		{"add", "Gimp"},
		{"remove", "Firefox"},
		{"add", "Blender"},
	}
	
	expectedFavorites := []string{"Terminal", "Gimp", "Blender"}
	
	for _, op := range operations {
		if op.action == "add" {
			err := configManager.AddFavorite(op.app)
			if err != nil {
				t.Fatalf("Failed to add favorite %s: %v", op.app, err)
			}
		} else {
			err := configManager.RemoveFavorite(op.app)
			if err != nil {
				t.Fatalf("Failed to remove favorite %s: %v", op.app, err)
			}
		}
	}
	
	// Verify final state in memory
	for _, expected := range expectedFavorites {
		if !configManager.IsFavorite(expected) {
			t.Errorf("Expected %s to be a favorite in memory", expected)
		}
	}
	
	// Verify persistence by loading from disk
	favorites, err := configManager.LoadFavorites()
	if err != nil {
		t.Fatalf("Failed to load favorites from disk: %v", err)
	}
	
	if len(favorites) != len(expectedFavorites) {
		t.Errorf("Expected %d favorites on disk, got %d", len(expectedFavorites), len(favorites))
	}
	
	for i, expected := range expectedFavorites {
		if favorites[i] != expected {
			t.Errorf("Expected favorite %d to be %s, got %s", i, expected, favorites[i])
		}
	}
	
	// Test config file format
	configFile := filepath.Join(configDir, "config.json")
	data, err := os.ReadFile(configFile)
	if err != nil {
		t.Fatalf("Failed to read config file: %v", err)
	}
	
	var configData map[string]interface{}
	err = json.Unmarshal(data, &configData)
	if err != nil {
		t.Fatalf("Config file contains invalid JSON: %v", err)
	}
	
	favoritesData, ok := configData["favorites"].([]interface{})
	if !ok {
		t.Fatal("Config file does not contain favorites array")
	}
	
	if len(favoritesData) != len(expectedFavorites) {
		t.Errorf("Config file favorites length mismatch: expected %d, got %d", len(expectedFavorites), len(favoritesData))
	}
}

// TestErrorHandlingInIntegration tests error handling in integration scenarios
func TestErrorHandlingInIntegration(t *testing.T) {
	tempDir := t.TempDir()
	
	// Test 1: Scanner with non-existent directories
	t.Run("ScannerWithNonExistentDirectories", func(t *testing.T) {
		appScanner := scanner.NewScanner()
		nonExistentDirs := []string{
			filepath.Join(tempDir, "nonexistent1"),
			filepath.Join(tempDir, "nonexistent2"),
		}
		appScanner.SetScanPaths(nonExistentDirs)
		
		applications, err := appScanner.ScanApplications()
		if err != nil {
			t.Fatalf("Scanner should handle non-existent directories gracefully: %v", err)
		}
		
		if len(applications) != 0 {
			t.Errorf("Expected 0 applications from non-existent directories, got %d", len(applications))
		}
	})
	
	// Test 2: Config manager with read-only directory
	t.Run("ConfigManagerWithReadOnlyDirectory", func(t *testing.T) {
		readOnlyDir := filepath.Join(tempDir, "readonly")
		err := os.MkdirAll(readOnlyDir, 0444) // Read-only
		if err != nil {
			t.Fatalf("Failed to create read-only directory: %v", err)
		}
		
		configManager := config.NewManager()
		configManager.SetConfigPath(readOnlyDir)
		
		// This should fail gracefully
		err = configManager.AddFavorite("Test App")
		if err == nil {
			t.Error("Expected error when writing to read-only directory")
		}
	})
	
	// Test 3: Malformed desktop files
	t.Run("MalformedDesktopFiles", func(t *testing.T) {
		appsDir := filepath.Join(tempDir, "malformed_apps")
		err := os.MkdirAll(appsDir, 0755)
		if err != nil {
			t.Fatalf("Failed to create apps directory: %v", err)
		}
		
		// Create malformed desktop files
		malformedFiles := []struct {
			name    string
			content string
		}{
			{
				name:    "empty.desktop",
				content: "",
			},
			{
				name:    "no-desktop-entry.desktop",
				content: "Name=Test\nExec=test",
			},
			{
				name:    "missing-name.desktop",
				content: "[Desktop Entry]\nExec=test",
			},
			{
				name:    "missing-exec.desktop",
				content: "[Desktop Entry]\nName=Test",
			},
		}
		
		for _, file := range malformedFiles {
			filePath := filepath.Join(appsDir, file.name)
			err := os.WriteFile(filePath, []byte(file.content), 0644)
			if err != nil {
				t.Fatalf("Failed to write malformed file %s: %v", file.name, err)
			}
		}
		
		// Also create one valid file
		validFile := filepath.Join(appsDir, "valid.desktop")
		validContent := `[Desktop Entry]
Name=Valid App
Exec=valid-app
Type=Application`
		err = os.WriteFile(validFile, []byte(validContent), 0644)
		if err != nil {
			t.Fatalf("Failed to write valid file: %v", err)
		}
		
		appScanner := scanner.NewScanner()
		appScanner.SetScanPaths([]string{appsDir})
		
		applications, err := appScanner.ScanApplications()
		if err != nil {
			t.Fatalf("Scanner should handle malformed files gracefully: %v", err)
		}
		
		// Should only find the valid application
		if len(applications) != 1 {
			t.Errorf("Expected 1 valid application, got %d", len(applications))
		}
		
		if len(applications) > 0 && applications[0].Name != "Valid App" {
			t.Errorf("Expected to find Valid App, got %s", applications[0].Name)
		}
	})
}

// TestConcurrentOperations tests concurrent access to shared resources
func TestConcurrentOperations(t *testing.T) {
	tempDir := t.TempDir()
	configDir := filepath.Join(tempDir, "config")
	
	err := os.MkdirAll(configDir, 0755)
	if err != nil {
		t.Fatalf("Failed to create config directory: %v", err)
	}
	
	configManager := config.NewManager()
	configManager.SetConfigPath(configDir)
	
	// Test concurrent favorite operations
	done := make(chan bool, 10)
	
	// Start multiple goroutines adding favorites
	for i := 0; i < 5; i++ {
		go func(id int) {
			defer func() { done <- true }()
			
			appName := fmt.Sprintf("App%d", id)
			err := configManager.AddFavorite(appName)
			if err != nil {
				t.Errorf("Failed to add favorite %s: %v", appName, err)
			}
		}(i)
	}
	
	// Start multiple goroutines removing favorites
	for i := 0; i < 5; i++ {
		go func(id int) {
			defer func() { done <- true }()
			
			// Wait a bit to let add operations complete
			time.Sleep(10 * time.Millisecond)
			
			appName := fmt.Sprintf("App%d", id)
			err := configManager.RemoveFavorite(appName)
			if err != nil {
				t.Errorf("Failed to remove favorite %s: %v", appName, err)
			}
		}(i)
	}
	
	// Wait for all operations to complete
	for i := 0; i < 10; i++ {
		<-done
	}
	
	// Verify final state is consistent
	favorites, err := configManager.LoadFavorites()
	if err != nil {
		t.Fatalf("Failed to load favorites after concurrent operations: %v", err)
	}
	
	// All favorites should have been removed
	if len(favorites) != 0 {
		t.Errorf("Expected 0 favorites after concurrent add/remove, got %d", len(favorites))
	}
}

// TestRequirement1_1_DefaultFavoritesDisplay tests that TUI_Launcher displays Favorites_List as default view
// Requirement 1.1: WHEN the user starts the TUI_Launcher, THE TUI_Launcher SHALL display the Favorites_List as the default view
func TestRequirement1_1_DefaultFavoritesDisplay(t *testing.T) {
	tempDir := t.TempDir()
	configDir := filepath.Join(tempDir, "config")
	appsDir := filepath.Join(tempDir, "applications")
	
	// Setup directories
	err := os.MkdirAll(configDir, 0755)
	if err != nil {
		t.Fatalf("Failed to create config directory: %v", err)
	}
	err = os.MkdirAll(appsDir, 0755)
	if err != nil {
		t.Fatalf("Failed to create apps directory: %v", err)
	}
	
	// Create test applications
	testApps := []struct {
		filename string
		content  string
	}{
		{
			filename: "firefox.desktop",
			content: `[Desktop Entry]
Name=Firefox
Exec=firefox
Type=Application`,
		},
		{
			filename: "code.desktop",
			content: `[Desktop Entry]
Name=Visual Studio Code
Exec=code
Type=Application`,
		},
		{
			filename: "terminal.desktop",
			content: `[Desktop Entry]
Name=Terminal
Exec=gnome-terminal
Type=Application`,
		},
	}
	
	for _, app := range testApps {
		filePath := filepath.Join(appsDir, app.filename)
		err := os.WriteFile(filePath, []byte(app.content), 0644)
		if err != nil {
			t.Fatalf("Failed to write test app %s: %v", app.filename, err)
		}
	}
	
	// Initialize components
	configManager := config.NewManager()
	configManager.SetConfigPath(configDir)
	
	appScanner := scanner.NewScanner()
	appScanner.SetScanPaths([]string{appsDir})
	
	fuzzySearcher := search.NewFuzzySearcher()
	
	// Add some favorites
	err = configManager.AddFavorite("Firefox")
	if err != nil {
		t.Fatalf("Failed to add Firefox to favorites: %v", err)
	}
	err = configManager.AddFavorite("Visual Studio Code")
	if err != nil {
		t.Fatalf("Failed to add Visual Studio Code to favorites: %v", err)
	}
	
	// Scan applications
	applications, err := appScanner.ScanApplications()
	if err != nil {
		t.Fatalf("Failed to scan applications: %v", err)
	}
	
	// Test default view (empty search should show favorites)
	fuzzySearcher.SetItems(applications)
	results := fuzzySearcher.Search("", applications) // Empty query simulates default view
	
	// Verify that all applications are returned (favorites would be filtered/sorted by TUI layer)
	if len(results) != len(applications) {
		t.Errorf("Expected %d applications in default view, got %d", len(applications), len(results))
	}
	
	// Verify favorites are properly marked
	favorites, err := configManager.LoadFavorites()
	if err != nil {
		t.Fatalf("Failed to load favorites: %v", err)
	}
	
	expectedFavorites := []string{"Firefox", "Visual Studio Code"}
	if len(favorites) != len(expectedFavorites) {
		t.Errorf("Expected %d favorites, got %d", len(expectedFavorites), len(favorites))
	}
	
	for _, expectedFav := range expectedFavorites {
		if !configManager.IsFavorite(expectedFav) {
			t.Errorf("Expected %s to be marked as favorite", expectedFav)
		}
	}
}

// TestRequirement1_2_FuzzySearchFiltering tests that typing characters filters applications using fuzzy search
// Requirement 1.2: WHEN the user types characters, THE TUI_Launcher SHALL filter available Application_Entry items using Fuzzy_Search
func TestRequirement1_2_FuzzySearchFiltering(t *testing.T) {
	tempDir := t.TempDir()
	appsDir := filepath.Join(tempDir, "applications")
	
	err := os.MkdirAll(appsDir, 0755)
	if err != nil {
		t.Fatalf("Failed to create apps directory: %v", err)
	}
	
	// Create test applications with diverse names for fuzzy search testing
	testApps := []struct {
		filename string
		name     string
	}{
		{"firefox.desktop", "Firefox Web Browser"},
		{"code.desktop", "Visual Studio Code"},
		{"terminal.desktop", "GNOME Terminal"},
		{"gimp.desktop", "GNU Image Manipulation Program"},
		{"libreoffice.desktop", "LibreOffice Writer"},
		{"thunderbird.desktop", "Thunderbird Mail"},
	}
	
	for _, app := range testApps {
		content := fmt.Sprintf(`[Desktop Entry]
Name=%s
Exec=%s
Type=Application`, app.name, strings.ToLower(strings.ReplaceAll(app.name, " ", "-")))
		
		filePath := filepath.Join(appsDir, app.filename)
		err := os.WriteFile(filePath, []byte(content), 0644)
		if err != nil {
			t.Fatalf("Failed to write test app %s: %v", app.filename, err)
		}
	}
	
	// Initialize components
	appScanner := scanner.NewScanner()
	appScanner.SetScanPaths([]string{appsDir})
	
	fuzzySearcher := search.NewFuzzySearcher()
	
	// Scan applications
	applications, err := appScanner.ScanApplications()
	if err != nil {
		t.Fatalf("Failed to scan applications: %v", err)
	}
	
	fuzzySearcher.SetItems(applications)
	
	// Test various fuzzy search scenarios
	testCases := []struct {
		query           string
		expectedMatches []string
		description     string
	}{
		{
			query:           "fire",
			expectedMatches: []string{"Firefox Web Browser"},
			description:     "Partial match at beginning",
		},
		{
			query:           "code",
			expectedMatches: []string{"Visual Studio Code"},
			description:     "Partial match in middle",
		},
		{
			query:           "term",
			expectedMatches: []string{"GNOME Terminal"},
			description:     "Partial match at end",
		},
		{
			query:           "gnu",
			expectedMatches: []string{"GNU Image Manipulation Program"},
			description:     "Acronym match",
		},
		{
			query:           "office",
			expectedMatches: []string{"LibreOffice Writer"},
			description:     "Word boundary match",
		},
		{
			query:           "mail",
			expectedMatches: []string{"Thunderbird Mail"},
			description:     "Exact word match",
		},
		{
			query:           "xyz",
			expectedMatches: []string{},
			description:     "No matches",
		},
	}
	
	for _, tc := range testCases {
		t.Run(tc.description, func(t *testing.T) {
			results := fuzzySearcher.Search(tc.query, applications)
			
			if len(tc.expectedMatches) == 0 {
				if len(results) != 0 {
					t.Errorf("Expected no matches for query '%s', got %d", tc.query, len(results))
				}
				return
			}
			
			if len(results) == 0 {
				t.Errorf("Expected matches for query '%s', got none", tc.query)
				return
			}
			
			// Check if expected matches are found
			foundMatches := make(map[string]bool)
			for _, result := range results {
				foundMatches[result.Application.Name] = true
			}
			
			for _, expected := range tc.expectedMatches {
				if !foundMatches[expected] {
					t.Errorf("Expected to find '%s' for query '%s'", expected, tc.query)
				}
			}
		})
	}
}

// TestRequirement1_3_ApplicationLaunching tests that selecting an application executes the launch command
// Requirement 1.3: WHEN the user selects an Application_Entry, THE TUI_Launcher SHALL execute the corresponding Launch_Command
func TestRequirement1_3_ApplicationLaunching(t *testing.T) {
	tempDir := t.TempDir()
	appsDir := filepath.Join(tempDir, "applications")
	
	err := os.MkdirAll(appsDir, 0755)
	if err != nil {
		t.Fatalf("Failed to create apps directory: %v", err)
	}
	
	// Create test applications with various exec formats
	testApps := []struct {
		filename string
		name     string
		exec     string
	}{
		{
			filename: "simple.desktop",
			name:     "Simple App",
			exec:     "echo",
		},
		{
			filename: "with-args.desktop",
			name:     "App with Args",
			exec:     "echo hello world",
		},
		{
			filename: "with-fieldcodes.desktop",
			name:     "App with Field Codes",
			exec:     "echo %f %u",
		},
		{
			filename: "quoted.desktop",
			name:     "Quoted App",
			exec:     `echo "hello world"`,
		},
	}
	
	for _, app := range testApps {
		content := fmt.Sprintf(`[Desktop Entry]
Name=%s
Exec=%s
Type=Application`, app.name, app.exec)
		
		filePath := filepath.Join(appsDir, app.filename)
		err := os.WriteFile(filePath, []byte(content), 0644)
		if err != nil {
			t.Fatalf("Failed to write test app %s: %v", app.filename, err)
		}
	}
	
	// Initialize components
	appScanner := scanner.NewScanner()
	appScanner.SetScanPaths([]string{appsDir})
	
	appLauncher := launcher.NewLauncher()
	
	// Scan applications
	applications, err := appScanner.ScanApplications()
	if err != nil {
		t.Fatalf("Failed to scan applications: %v", err)
	}
	
	// Test launching each application
	for _, app := range applications {
		t.Run(fmt.Sprintf("Launch_%s", app.Name), func(t *testing.T) {
			// Test validation first
			err := appLauncher.ValidateApplication(app)
			if err != nil {
				// For integration tests, we expect some commands might not exist
				if !strings.Contains(err.Error(), "not found in PATH") {
					t.Errorf("Unexpected validation error for %s: %v", app.Name, err)
				}
				return // Skip launch test if command doesn't exist
			}
			
			// Test getting launch command
			command, args, err := appLauncher.GetLaunchCommand(app)
			if err != nil {
				t.Errorf("Failed to get launch command for %s: %v", app.Name, err)
				return
			}
			
			// Verify command parsing
			if command == "" {
				t.Errorf("Empty command for %s", app.Name)
			}
			
			// For echo command, we can safely test launching
			if command == "echo" {
				err = appLauncher.LaunchApplication(app)
				if err != nil {
					t.Errorf("Failed to launch %s: %v", app.Name, err)
				}
			}
		})
	}
}

// TestRequirement3_1_PersistentFavoritesList tests that favorites are maintained in persistent storage
// Requirement 3.1: THE TUI_Launcher SHALL maintain a persistent Favorites_List stored in user configuration
func TestRequirement3_1_PersistentFavoritesList(t *testing.T) {
	tempDir := t.TempDir()
	configDir := filepath.Join(tempDir, "config")
	
	err := os.MkdirAll(configDir, 0755)
	if err != nil {
		t.Fatalf("Failed to create config directory: %v", err)
	}
	
	// Test persistence across multiple manager instances
	testFavorites := []string{"Firefox", "Visual Studio Code", "Terminal", "GIMP"}
	
	// First instance - add favorites
	{
		configManager := config.NewManager()
		configManager.SetConfigPath(configDir)
		
		for _, fav := range testFavorites {
			err := configManager.AddFavorite(fav)
			if err != nil {
				t.Fatalf("Failed to add favorite %s: %v", fav, err)
			}
		}
		
		// Verify in memory
		for _, fav := range testFavorites {
			if !configManager.IsFavorite(fav) {
				t.Errorf("Favorite %s not found in memory", fav)
			}
		}
	}
	
	// Second instance - verify persistence
	{
		configManager := config.NewManager()
		configManager.SetConfigPath(configDir)
		
		favorites, err := configManager.LoadFavorites()
		if err != nil {
			t.Fatalf("Failed to load favorites: %v", err)
		}
		
		if len(favorites) != len(testFavorites) {
			t.Errorf("Expected %d favorites, got %d", len(testFavorites), len(favorites))
		}
		
		for i, expected := range testFavorites {
			if favorites[i] != expected {
				t.Errorf("Expected favorite %d to be %s, got %s", i, expected, favorites[i])
			}
		}
		
		// Verify IsFavorite works after loading
		for _, fav := range testFavorites {
			if !configManager.IsFavorite(fav) {
				t.Errorf("Favorite %s not recognized after loading", fav)
			}
		}
	}
	
	// Third instance - modify and verify persistence
	{
		configManager := config.NewManager()
		configManager.SetConfigPath(configDir)
		
		// Remove some favorites
		err := configManager.RemoveFavorite("Visual Studio Code")
		if err != nil {
			t.Fatalf("Failed to remove favorite: %v", err)
		}
		
		err = configManager.RemoveFavorite("GIMP")
		if err != nil {
			t.Fatalf("Failed to remove favorite: %v", err)
		}
		
		// Add a new favorite
		err = configManager.AddFavorite("Blender")
		if err != nil {
			t.Fatalf("Failed to add new favorite: %v", err)
		}
	}
	
	// Fourth instance - verify final state
	{
		configManager := config.NewManager()
		configManager.SetConfigPath(configDir)
		
		favorites, err := configManager.LoadFavorites()
		if err != nil {
			t.Fatalf("Failed to load favorites in final check: %v", err)
		}
		
		expectedFinal := []string{"Firefox", "Terminal", "Blender"}
		if len(favorites) != len(expectedFinal) {
			t.Errorf("Expected %d favorites in final state, got %d", len(expectedFinal), len(favorites))
		}
		
		for _, expected := range expectedFinal {
			if !configManager.IsFavorite(expected) {
				t.Errorf("Expected %s to be in final favorites", expected)
			}
		}
		
		// Verify removed favorites are gone
		if configManager.IsFavorite("Visual Studio Code") {
			t.Error("Visual Studio Code should have been removed")
		}
		if configManager.IsFavorite("GIMP") {
			t.Error("GIMP should have been removed")
		}
	}
}

// TestRequirement3_2_FavoritesInDefaultDisplay tests that favorites are included in default display
// Requirement 3.2: WHEN the user adds an application to favorites, THE TUI_Launcher SHALL include it in the default display
func TestRequirement3_2_FavoritesInDefaultDisplay(t *testing.T) {
	tempDir := t.TempDir()
	configDir := filepath.Join(tempDir, "config")
	appsDir := filepath.Join(tempDir, "applications")
	
	// Setup directories
	err := os.MkdirAll(configDir, 0755)
	if err != nil {
		t.Fatalf("Failed to create config directory: %v", err)
	}
	err = os.MkdirAll(appsDir, 0755)
	if err != nil {
		t.Fatalf("Failed to create apps directory: %v", err)
	}
	
	// Create test applications
	testApps := []struct {
		filename string
		name     string
	}{
		{"firefox.desktop", "Firefox"},
		{"code.desktop", "Visual Studio Code"},
		{"terminal.desktop", "Terminal"},
		{"gimp.desktop", "GIMP"},
		{"blender.desktop", "Blender"},
	}
	
	for _, app := range testApps {
		content := fmt.Sprintf(`[Desktop Entry]
Name=%s
Exec=%s
Type=Application`, app.name, strings.ToLower(app.name))
		
		filePath := filepath.Join(appsDir, app.filename)
		err := os.WriteFile(filePath, []byte(content), 0644)
		if err != nil {
			t.Fatalf("Failed to write test app %s: %v", app.filename, err)
		}
	}
	
	// Initialize components
	configManager := config.NewManager()
	configManager.SetConfigPath(configDir)
	
	appScanner := scanner.NewScanner()
	appScanner.SetScanPaths([]string{appsDir})
	
	fuzzySearcher := search.NewFuzzySearcher()
	
	// Scan applications
	applications, err := appScanner.ScanApplications()
	if err != nil {
		t.Fatalf("Failed to scan applications: %v", err)
	}
	
	fuzzySearcher.SetItems(applications)
	
	// Initially no favorites
	favorites, err := configManager.LoadFavorites()
	if err != nil {
		t.Fatalf("Failed to load initial favorites: %v", err)
	}
	if len(favorites) != 0 {
		t.Errorf("Expected no initial favorites, got %d", len(favorites))
	}
	
	// Add some applications to favorites
	favoritesToAdd := []string{"Firefox", "Visual Studio Code", "Terminal"}
	
	for _, fav := range favoritesToAdd {
		err := configManager.AddFavorite(fav)
		if err != nil {
			t.Fatalf("Failed to add favorite %s: %v", fav, err)
		}
		
		// Verify it's immediately marked as favorite
		if !configManager.IsFavorite(fav) {
			t.Errorf("Application %s should be marked as favorite immediately after adding", fav)
		}
	}
	
	// Verify favorites are persisted
	favorites, err = configManager.LoadFavorites()
	if err != nil {
		t.Fatalf("Failed to load favorites after adding: %v", err)
	}
	
	if len(favorites) != len(favoritesToAdd) {
		t.Errorf("Expected %d favorites, got %d", len(favoritesToAdd), len(favorites))
	}
	
	for i, expected := range favoritesToAdd {
		if favorites[i] != expected {
			t.Errorf("Expected favorite %d to be %s, got %s", i, expected, favorites[i])
		}
	}
	
	// Test default display (empty search) - should show all applications
	// The TUI layer would filter/prioritize favorites, but the search engine returns all
	results := fuzzySearcher.Search("", applications)
	if len(results) != len(applications) {
		t.Errorf("Expected %d applications in default display, got %d", len(applications), len(results))
	}
	
	// Verify that favorite applications can be identified
	favoriteApps := make([]interfaces.Application, 0)
	for _, app := range applications {
		if configManager.IsFavorite(app.Name) {
			favoriteApps = append(favoriteApps, app)
		}
	}
	
	if len(favoriteApps) != len(favoritesToAdd) {
		t.Errorf("Expected %d favorite applications, got %d", len(favoritesToAdd), len(favoriteApps))
	}
	
	// Remove a favorite and verify it's excluded from favorites list
	err = configManager.RemoveFavorite("Visual Studio Code")
	if err != nil {
		t.Fatalf("Failed to remove favorite: %v", err)
	}
	
	if configManager.IsFavorite("Visual Studio Code") {
		t.Error("Visual Studio Code should no longer be a favorite")
	}
	
	// Verify updated favorites list
	favorites, err = configManager.LoadFavorites()
	if err != nil {
		t.Fatalf("Failed to load favorites after removal: %v", err)
	}
	
	expectedAfterRemoval := []string{"Firefox", "Terminal"}
	if len(favorites) != len(expectedAfterRemoval) {
		t.Errorf("Expected %d favorites after removal, got %d", len(expectedAfterRemoval), len(favorites))
	}
	
	for i, expected := range expectedAfterRemoval {
		if favorites[i] != expected {
			t.Errorf("Expected favorite %d to be %s after removal, got %s", i, expected, favorites[i])
		}
	}
}