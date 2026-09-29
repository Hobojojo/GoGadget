package scanner

import (
	"bufio"
	"context"
	"os"
	"strconv"
	"strings"

	"tui-app-launcher/internal/interfaces"
)

// DesktopEntry represents a parsed .desktop file
type DesktopEntry struct {
	Name          string
	Exec          string
	Icon          string
	Comment       string
	GenericName   string
	Keywords      []string
	Categories    []string
	NoDisplay     bool
	Hidden        bool
	Terminal      bool
	Path          string
	StartupNotify bool
	Type          string
}

// ParseDesktopFile parses a .desktop file and returns a DesktopEntry.
func ParseDesktopFile(filePath string) (*DesktopEntry, error) {
	return ParseDesktopFileContext(context.Background(), filePath)
}

// ParseDesktopFileContext checks for cancellation while reading a desktop file.
func ParseDesktopFileContext(ctx context.Context, filePath string) (*DesktopEntry, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	file, err := os.Open(filePath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	entry := &DesktopEntry{}
	scanner := bufio.NewScanner(file)
	inDesktopEntry := false

	for scanner.Scan() {
		if err := ctx.Err(); err != nil {
			return nil, err
		}
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
		case "GenericName":
			entry.GenericName = value
		case "Path":
			entry.Path = value
		case "Terminal":
			entry.Terminal, _ = strconv.ParseBool(value)
		case "Hidden":
			entry.Hidden, _ = strconv.ParseBool(value)
		case "StartupNotify":
			entry.StartupNotify, _ = strconv.ParseBool(value)
		case "Keywords":
			entry.Keywords = splitDesktopList(value)
		case "Categories":
			entry.Categories = splitDesktopList(value)
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

	if err := ctx.Err(); err != nil {
		return nil, err
	}
	return entry, nil
}

func splitDesktopList(value string) []string {
	var result []string
	for _, part := range strings.Split(value, ";") {
		if part = strings.TrimSpace(part); part != "" {
			result = append(result, part)
		}
	}
	return result
}

// ToApplication converts a DesktopEntry to an Application struct
func (de *DesktopEntry) ToApplication(desktopFilePath string) interfaces.Application {
	return interfaces.Application{
		Name:          de.Name,
		Exec:          de.Exec,
		Icon:          de.Icon,
		Comment:       de.Comment,
		GenericName:   de.GenericName,
		Keywords:      de.Keywords,
		Categories:    de.Categories,
		Terminal:      de.Terminal,
		Path:          de.Path,
		StartupNotify: de.StartupNotify,
		DesktopFile:   desktopFilePath,
		IsFavorite:    false, // Will be set by config manager
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
	if de.NoDisplay || de.Hidden {
		return false
	}

	return true
}
