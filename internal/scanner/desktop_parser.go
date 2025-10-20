package scanner

import (
	"bufio"
	"os"
	"strconv"
	"strings"

	"tui-app-launcher/internal/interfaces"
)

// DesktopEntry represents a parsed .desktop file
type DesktopEntry struct {
	Name       string
	Exec       string
	Icon       string
	Comment    string
	Categories []string
	NoDisplay  bool
	Type       string
}

// ParseDesktopFile parses a .desktop file and returns a DesktopEntry
func ParseDesktopFile(filePath string) (*DesktopEntry, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	entry := &DesktopEntry{}
	scanner := bufio.NewScanner(file)
	inDesktopEntry := false

	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())

		// Skip empty lines and comments
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		// Check for [Desktop Entry] section
		if line == "[Desktop Entry]" {
			inDesktopEntry = true
			continue
		}

		// Check for other sections (stop parsing Desktop Entry)
		if strings.HasPrefix(line, "[") && strings.HasSuffix(line, "]") {
			inDesktopEntry = false
			continue
		}

		// Only parse lines within [Desktop Entry] section
		if !inDesktopEntry {
			continue
		}

		// Parse key=value pairs
		parts := strings.SplitN(line, "=", 2)
		if len(parts) != 2 {
			continue
		}

		key := strings.TrimSpace(parts[0])
		value := strings.TrimSpace(parts[1])

		switch key {
		case "Name":
			entry.Name = value
		case "Exec":
			entry.Exec = value
		case "Icon":
			entry.Icon = value
		case "Comment":
			entry.Comment = value
		case "Categories":
			if value != "" {
				entry.Categories = strings.Split(value, ";")
				// Remove empty strings from categories
				filtered := make([]string, 0, len(entry.Categories))
				for _, cat := range entry.Categories {
					if strings.TrimSpace(cat) != "" {
						filtered = append(filtered, strings.TrimSpace(cat))
					}
				}
				entry.Categories = filtered
			}
		case "NoDisplay":
			if noDisplay, err := strconv.ParseBool(value); err == nil {
				entry.NoDisplay = noDisplay
			}
		case "Type":
			entry.Type = value
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, err
	}

	return entry, nil
}

// ToApplication converts a DesktopEntry to an Application struct
func (de *DesktopEntry) ToApplication(desktopFilePath string) interfaces.Application {
	return interfaces.Application{
		Name:        de.Name,
		Exec:        de.Exec,
		Icon:        de.Icon,
		Comment:     de.Comment,
		Categories:  de.Categories,
		DesktopFile: desktopFilePath,
		IsFavorite:  false, // Will be set by config manager
	}
}

// IsValidApplication checks if the desktop entry represents a valid application
func (de *DesktopEntry) IsValidApplication() bool {
	// Must have a name and exec command
	if de.Name == "" || de.Exec == "" {
		return false
	}

	// Must be an Application type (default if not specified)
	if de.Type != "" && de.Type != "Application" {
		return false
	}

	// Skip if NoDisplay is true
	if de.NoDisplay {
		return false
	}

	return true
}
