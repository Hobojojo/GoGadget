# TUI App Launcher Design Document

## Overview

The TUI App Launcher is a Go-based terminal application that provides fast, keyboard-driven access to Linux applications. It uses a clean, responsive interface built with the Bubble Tea TUI framework, implements fuzzy search using the fzf algorithm, and maintains user preferences through a simple configuration system.

## Architecture

### High-Level Architecture

```
┌─────────────────┐    ┌─────────────────┐    ┌─────────────────┐
│   TUI Layer     │    │  Application    │    │  Configuration  │
│  (Bubble Tea)   │◄──►│    Scanner      │    │    Manager      │
│                 │    │                 │    │                 │
└─────────────────┘    └─────────────────┘    └─────────────────┘
         │                       │                       │
         ▼                       ▼                       ▼
┌─────────────────┐    ┌─────────────────┐    ┌─────────────────┐
│  Fuzzy Search   │    │  Desktop File   │    │   Favorites     │
│    Engine       │    │     Parser      │    │    Storage      │
│                 │    │                 │    │                 │
└─────────────────┘    └─────────────────┘    └─────────────────┘
```

### Core Components

1. **Main Application Controller**: Orchestrates the overall application flow
2. **TUI Interface**: Handles user interaction and display using Bubble Tea
3. **Application Scanner**: Discovers and parses .desktop files
4. **Fuzzy Search Engine**: Implements fast fuzzy matching
5. **Configuration Manager**: Handles favorites and user preferences
6. **Application Launcher**: Executes selected applications

## Components and Interfaces

### Application Model

```go
type Application struct {
    Name        string
    Exec        string
    Icon        string
    Comment     string
    Categories  []string
    DesktopFile string
    IsFavorite  bool
}
```

### Main Interfaces

#### ApplicationScanner Interface
```go
type ApplicationScanner interface {
    ScanApplications() ([]Application, error)
    RefreshApplications() error
}
```

#### FuzzySearcher Interface
```go
type FuzzySearcher interface {
    Search(query string, items []Application) []SearchResult
    SetItems(items []Application)
}

type SearchResult struct {
    Application Application
    Score      int
    Matches    []int // Character positions that matched
}
```

#### ConfigManager Interface
```go
type ConfigManager interface {
    LoadFavorites() ([]string, error)
    SaveFavorites(favorites []string) error
    AddFavorite(appName string) error
    RemoveFavorite(appName string) error
    IsFavorite(appName string) bool
}
```

### TUI Model Structure

```go
type Model struct {
    applications    []Application
    filteredApps    []Application
    searchQuery     string
    selectedIndex   int
    fuzzySearcher   FuzzySearcher
    configManager   ConfigManager
    showHelp        bool
    width           int
    height          int
}
```

## Data Models

### Configuration File Structure

The application will store configuration in `~/.config/tui-launcher/config.json`:

```json
{
    "favorites": [
        "firefox",
        "code",
        "terminal"
    ],
    "scan_paths": [
        "/usr/share/applications",
        "/usr/local/share/applications",
        "~/.local/share/applications"
    ]
}
```

### Desktop File Parsing

The scanner will parse standard Linux .desktop files, extracting:
- `Name`: Application display name
- `Exec`: Command to execute
- `Icon`: Icon name/path
- `Comment`: Application description
- `Categories`: Application categories
- `NoDisplay`: Whether to hide from menus

## Error Handling

### Error Categories

1. **Configuration Errors**: Invalid config files, permission issues
2. **Scanning Errors**: Inaccessible directories, malformed .desktop files
3. **Launch Errors**: Invalid executables, missing dependencies
4. **TUI Errors**: Terminal compatibility issues

### Error Handling Strategy

- **Graceful Degradation**: Continue operation with reduced functionality when possible
- **User Feedback**: Display clear error messages in the TUI
- **Logging**: Write detailed errors to `~/.config/tui-launcher/launcher.log`
- **Recovery**: Attempt to recover from transient errors automatically

### Specific Error Handling

```go
type LauncherError struct {
    Type    ErrorType
    Message string
    Cause   error
}

type ErrorType int

const (
    ConfigError ErrorType = iota
    ScanError
    LaunchError
    TUIError
)
```

## Testing Strategy

### Unit Testing Focus Areas

1. **Desktop File Parser**: Test parsing of various .desktop file formats
2. **Fuzzy Search Algorithm**: Verify search accuracy and performance
3. **Configuration Manager**: Test favorites management and persistence
4. **Application Scanner**: Test directory scanning and filtering

### Integration Testing

1. **End-to-End Workflow**: Test complete user interaction flows
2. **Configuration Persistence**: Verify favorites are saved and loaded correctly
3. **Application Discovery**: Test scanning across different Linux distributions

### Test Data Strategy

- **Mock .desktop Files**: Create test fixtures with various formats
- **Temporary Directories**: Use isolated test environments
- **Configuration Mocking**: Test with different configuration scenarios

## Implementation Details

### Fuzzy Search Algorithm

Using a simplified fzf-style algorithm:
1. **Character Matching**: Score based on consecutive character matches
2. **Position Weighting**: Prefer matches at word boundaries and start of strings
3. **Case Sensitivity**: Case-insensitive matching with bonus for case matches
4. **Ranking**: Sort results by score, then alphabetically

### TUI Layout

```
┌─────────────────────────────────────────────────────────────┐
│ TUI App Launcher                                    [?] Help │
├─────────────────────────────────────────────────────────────┤
│ Search: firefox_                                            │
├─────────────────────────────────────────────────────────────┤
│ > Firefox Web Browser                               ★       │
│   Firefox Developer Edition                                 │
│   Thunderbird                                               │
│   File Manager                                              │
│                                                             │
├─────────────────────────────────────────────────────────────┤
│ ↑/↓: Navigate  Enter: Launch  Tab: Toggle Favorite  Esc: Exit │
└─────────────────────────────────────────────────────────────┘
```

### Key Bindings

- **Arrow Keys / j,k**: Navigate up/down
- **Enter**: Launch selected application
- **Tab**: Toggle favorite status
- **Escape / Ctrl+C**: Exit
- **?**: Toggle help display
- **Ctrl+R**: Refresh application list

### Performance Considerations

1. **Lazy Loading**: Load applications on startup, cache results
2. **Incremental Search**: Update search results as user types
3. **Debouncing**: Prevent excessive search operations during rapid typing
4. **Memory Management**: Efficient string handling for large application lists

### Cross-Platform Considerations

While focused on Linux, the design allows for future extension:
- **Abstracted Scanner**: Interface allows different implementations
- **Configurable Paths**: Scan paths can be customized per platform
- **Executable Detection**: Handle different executable formats