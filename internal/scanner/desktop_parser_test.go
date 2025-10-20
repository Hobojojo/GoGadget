package scanner

import (
	"os"
	"path/filepath"
	"testing"
)

func TestParseDesktopFile(t *testing.T) {
	// Create temporary directory for test fixtures
	tempDir, err := os.MkdirTemp("", "desktop_parser_test")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}
	defer os.RemoveAll(tempDir)

	tests := []struct {
		name               string
		content            string
		expectedName       string
		expectedExec       string
		expectedIcon       string
		expectedComment    string
		expectedCategories []string
		expectedNoDisplay  bool
		expectedType       string
		shouldError        bool
	}{
		{
			name: "valid_application",
			content: `[Desktop Entry]
Name=Firefox
Exec=firefox %u
Icon=firefox
Comment=Web Browser
Categories=Network;WebBrowser;
Type=Application
NoDisplay=false`,
			expectedName:       "Firefox",
			expectedExec:       "firefox %u",
			expectedIcon:       "firefox",
			expectedComment:    "Web Browser",
			expectedCategories: []string{"Network", "WebBrowser"},
			expectedNoDisplay:  false,
			expectedType:       "Application",
			shouldError:        false,
		},
		{
			name: "minimal_valid_application",
			content: `[Desktop Entry]
Name=Simple App
Exec=simple-app`,
			expectedName:       "Simple App",
			expectedExec:       "simple-app",
			expectedIcon:       "",
			expectedComment:    "",
			expectedCategories: nil,
			expectedNoDisplay:  false,
			expectedType:       "",
			shouldError:        false,
		},
		{
			name: "application_with_comments_and_empty_lines",
			content: `# This is a comment
[Desktop Entry]
# Another comment
Name=Test App

Exec=test-app
# Comment in middle
Icon=test-icon

Comment=Test application
Categories=Utility;`,
			expectedName:       "Test App",
			expectedExec:       "test-app",
			expectedIcon:       "test-icon",
			expectedComment:    "Test application",
			expectedCategories: []string{"Utility"},
			expectedNoDisplay:  false,
			expectedType:       "",
			shouldError:        false,
		},
		{
			name: "application_with_multiple_sections",
			content: `[Desktop Entry]
Name=Multi Section App
Exec=multi-app
Icon=multi-icon

[Desktop Action New]
Name=New Window
Exec=multi-app --new-window

[Some Other Section]
Key=Value`,
			expectedName:       "Multi Section App",
			expectedExec:       "multi-app",
			expectedIcon:       "multi-icon",
			expectedComment:    "",
			expectedCategories: nil,
			expectedNoDisplay:  false,
			expectedType:       "",
			shouldError:        false,
		},
		{
			name: "application_with_nodisplay_true",
			content: `[Desktop Entry]
Name=Hidden App
Exec=hidden-app
NoDisplay=true`,
			expectedName:       "Hidden App",
			expectedExec:       "hidden-app",
			expectedIcon:       "",
			expectedComment:    "",
			expectedCategories: nil,
			expectedNoDisplay:  true,
			expectedType:       "",
			shouldError:        false,
		},
		{
			name: "malformed_key_value_pairs",
			content: `[Desktop Entry]
Name=Malformed App
Exec=malformed-app
InvalidLine
Key=
=Value
Key=Value=Extra`,
			expectedName:       "Malformed App",
			expectedExec:       "malformed-app",
			expectedIcon:       "",
			expectedComment:    "",
			expectedCategories: nil,
			expectedNoDisplay:  false,
			expectedType:       "",
			shouldError:        false,
		},
		{
			name: "empty_categories",
			content: `[Desktop Entry]
Name=Empty Categories App
Exec=empty-cat-app
Categories=;;Utility;;Development;`,
			expectedName:       "Empty Categories App",
			expectedExec:       "empty-cat-app",
			expectedIcon:       "",
			expectedComment:    "",
			expectedCategories: []string{"Utility", "Development"},
			expectedNoDisplay:  false,
			expectedType:       "",
			shouldError:        false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create test file
			testFile := filepath.Join(tempDir, tt.name+".desktop")
			err := os.WriteFile(testFile, []byte(tt.content), 0644)
			if err != nil {
				t.Fatalf("Failed to create test file: %v", err)
			}

			// Parse the file
			entry, err := ParseDesktopFile(testFile)

			if tt.shouldError {
				if err == nil {
					t.Errorf("Expected error but got none")
				}
				return
			}

			if err != nil {
				t.Fatalf("Unexpected error: %v", err)
			}

			// Verify parsed values
			if entry.Name != tt.expectedName {
				t.Errorf("Name: expected %q, got %q", tt.expectedName, entry.Name)
			}
			if entry.Exec != tt.expectedExec {
				t.Errorf("Exec: expected %q, got %q", tt.expectedExec, entry.Exec)
			}
			if entry.Icon != tt.expectedIcon {
				t.Errorf("Icon: expected %q, got %q", tt.expectedIcon, entry.Icon)
			}
			if entry.Comment != tt.expectedComment {
				t.Errorf("Comment: expected %q, got %q", tt.expectedComment, entry.Comment)
			}
			if entry.NoDisplay != tt.expectedNoDisplay {
				t.Errorf("NoDisplay: expected %v, got %v", tt.expectedNoDisplay, entry.NoDisplay)
			}
			if entry.Type != tt.expectedType {
				t.Errorf("Type: expected %q, got %q", tt.expectedType, entry.Type)
			}

			// Check categories
			if len(entry.Categories) != len(tt.expectedCategories) {
				t.Errorf("Categories length: expected %d, got %d", len(tt.expectedCategories), len(entry.Categories))
			} else {
				for i, cat := range tt.expectedCategories {
					if entry.Categories[i] != cat {
						t.Errorf("Categories[%d]: expected %q, got %q", i, cat, entry.Categories[i])
					}
				}
			}
		})
	}
}
func TestParseDesktopFile_FileErrors(t *testing.T) {
	tests := []struct {
		name        string
		filePath    string
		shouldError bool
	}{
		{
			name:        "nonexistent_file",
			filePath:    "/nonexistent/path/file.desktop",
			shouldError: true,
		},
		{
			name:        "empty_path",
			filePath:    "",
			shouldError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := ParseDesktopFile(tt.filePath)
			if tt.shouldError && err == nil {
				t.Errorf("Expected error for file %q but got none", tt.filePath)
			}
			if !tt.shouldError && err != nil {
				t.Errorf("Unexpected error for file %q: %v", tt.filePath, err)
			}
		})
	}
}

func TestDesktopEntry_IsValidApplication(t *testing.T) {
	tests := []struct {
		name     string
		entry    DesktopEntry
		expected bool
	}{
		{
			name: "valid_application",
			entry: DesktopEntry{
				Name: "Firefox",
				Exec: "firefox",
				Type: "Application",
			},
			expected: true,
		},
		{
			name: "valid_application_no_type",
			entry: DesktopEntry{
				Name: "Firefox",
				Exec: "firefox",
			},
			expected: true,
		},
		{
			name: "missing_name",
			entry: DesktopEntry{
				Exec: "firefox",
				Type: "Application",
			},
			expected: false,
		},
		{
			name: "missing_exec",
			entry: DesktopEntry{
				Name: "Firefox",
				Type: "Application",
			},
			expected: false,
		},
		{
			name: "nodisplay_true",
			entry: DesktopEntry{
				Name:      "Firefox",
				Exec:      "firefox",
				NoDisplay: true,
			},
			expected: false,
		},
		{
			name: "wrong_type",
			entry: DesktopEntry{
				Name: "Firefox",
				Exec: "firefox",
				Type: "Link",
			},
			expected: false,
		},
		{
			name: "empty_name_and_exec",
			entry: DesktopEntry{
				Name: "",
				Exec: "",
			},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.entry.IsValidApplication()
			if result != tt.expected {
				t.Errorf("IsValidApplication(): expected %v, got %v", tt.expected, result)
			}
		})
	}
}

func TestDesktopEntry_ToApplication(t *testing.T) {
	entry := DesktopEntry{
		Name:       "Test App",
		Exec:       "test-app --flag",
		Icon:       "test-icon",
		Comment:    "Test application",
		Categories: []string{"Utility", "Development"},
	}

	desktopFilePath := "/path/to/test.desktop"
	app := entry.ToApplication(desktopFilePath)

	if app.Name != entry.Name {
		t.Errorf("Name: expected %q, got %q", entry.Name, app.Name)
	}
	if app.Exec != entry.Exec {
		t.Errorf("Exec: expected %q, got %q", entry.Exec, app.Exec)
	}
	if app.Icon != entry.Icon {
		t.Errorf("Icon: expected %q, got %q", entry.Icon, app.Icon)
	}
	if app.Comment != entry.Comment {
		t.Errorf("Comment: expected %q, got %q", entry.Comment, app.Comment)
	}
	if app.DesktopFile != desktopFilePath {
		t.Errorf("DesktopFile: expected %q, got %q", desktopFilePath, app.DesktopFile)
	}
	if app.IsFavorite != false {
		t.Errorf("IsFavorite: expected false, got %v", app.IsFavorite)
	}

	// Check categories
	if len(app.Categories) != len(entry.Categories) {
		t.Errorf("Categories length: expected %d, got %d", len(entry.Categories), len(app.Categories))
	} else {
		for i, cat := range entry.Categories {
			if app.Categories[i] != cat {
				t.Errorf("Categories[%d]: expected %q, got %q", i, cat, app.Categories[i])
			}
		}
	}
}
