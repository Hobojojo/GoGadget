# TUI App Launcher

A Terminal User Interface (TUI) application launcher for Linux systems that provides fast, keyboard-driven access to installed applications with fuzzy search and favorites management.

## Features

- **Fast Application Discovery**: Automatically scans system and user application directories
- **Fuzzy Search**: Find applications quickly with partial name matching
- **Favorites Management**: Mark frequently used applications as favorites
- **Keyboard-Driven Interface**: Efficient navigation without mouse dependency
- **Responsive Design**: Adapts to different terminal sizes
- **Error Handling**: Graceful handling of missing applications and configuration issues

## Installation

### Using Make (Recommended)

```bash
# Build and install to /usr/local/bin (requires sudo)
make install

# Or install to ~/.local/bin (no sudo required)
make install-user
```

### Manual Installation

```bash
# Build the application
go build -o tui-launcher

# Copy to your preferred location
sudo cp tui-launcher /usr/local/bin/
# or
cp tui-launcher ~/.local/bin/
```

### From Source

```bash
git clone <repository-url>
cd tui-app-launcher
make build
```

## Usage

### Starting the Application

```bash
tui-launcher
```

### Keyboard Shortcuts

| Key | Action |
|-----|--------|
| `↑/↓` | Navigate up/down through applications |
| `Enter` | Launch selected application |
| `Alt+1`–`Alt+9` | Launch numbered application in the visible window |
| `Alt+Enter` | Force launch in a terminal emulator |
| `Ctrl+D` or `Ctrl+Space` | Toggle favorite status of selected application |
| `Tab` / `Shift+Tab` | Cycle category forward/backward (including all categories) |
| `Home` / `End` | Select first/last result |
| `Page Up` / `Page Down` | Navigate by a visible page |
| `Left` / `Right`, `Ctrl+A` / `Ctrl+E` | Move search cursor; jump to start/end |
| `Ctrl+W` | Delete word before search cursor |
| `Delete` | Delete character after search cursor |
| `Ctrl+\` | Toggle case-sensitive search |
| `Backspace` | Delete character from search query |
| `Ctrl+U` | Clear search query |
| `Ctrl+L` | Clear error messages |
| `Ctrl+R` | Refresh application list |
| `?` | Toggle help screen |
| `Esc` or `Ctrl+C` | Exit application |

### Mouse Controls

In terminals that support mouse reporting (including the browser preview), click an application to select it, scroll the wheel over the list to navigate one row per notch, or double-click the same application within 400 ms to launch it. Both list columns are supported; clicks on borders, details, blank rows, or help do nothing. Keyboard controls are unchanged. Hold Shift to use your terminal's native text selection instead.

### Search

- Type any characters to search for applications using fuzzy matching
- Space-separated tokens must all match, across enabled names, executable commands, comments, generic names, keywords, and categories
- Use `/Development query` to filter by a desktop category and search within it
- Terminals wider than 100 columns show two row-major list columns beside the details pane; narrower terminals retain one column
- Empty and scanning states explain what is happening; `Ctrl+R` rescans
- `Alt+Enter` is the portable force-terminal shortcut: traditional terminals cannot distinguish `Shift+Enter` from `Enter`. `Ctrl+Tab` is also commonly indistinguishable from Tab, so plain Tab cycles categories.
- Clear the search query to view your favorites list
- Favorites appear first in search results

### Favorites Management

- Use `Ctrl+D` or `Ctrl+Space` to add or remove applications from favorites
- Favorites are displayed by default when no search query is entered
- Favorites are persisted across application restarts
- Favorite applications are marked with a ★ symbol

## Configuration

The application stores its configuration in `~/.config/tui-launcher/`:

- `config.json`: Contains favorites list and application settings
- `launcher.log`: Application logs (with automatic rotation)

### Configuration File Format

```json
{
  "favorites": [
    "Firefox Web Browser",
    "Visual Studio Code",
    "Terminal"
  ],
  "settings": {
    "search_fields": ["name", "exec", "comment", "generic_name", "categories", "keywords"],
    "ranking": "frecency",
    "multi_column": true,
    "case_sensitive": false
  }
}
```

Settings are optional and loaded on startup; old favorites-only files still work. `ranking` accepts `frecency` (default), `frequency`, `recency`, or `none`. Omitted/empty `search_fields` searches all supported fields. Launch history is stored automatically under `launch_history` after successful starts. Case-sensitivity toggles are session-only unless set in the file.

## Application Discovery

The launcher scans the `applications/` subdirectory of each absolute path in the colon-separated `XDG_DATA_DIRS` environment variable, in order. If unset or empty, it defaults to `/usr/local/share:/usr/share`. Relative and empty entries are ignored, and repeated directories are scanned only once.

It also scans `~/.local/share/applications/` for user applications, preserving user overrides. Flatpak, Snap, and custom installations are discovered when their data directories are included in `XDG_DATA_DIRS`.

Applications with `NoDisplay=true` are automatically filtered out.

## Development

### Building

```bash
# Build for current platform
make build

# Build with development flags (race detection)
make build-dev

# Cross-compile for multiple platforms
make build-all
```

### Testing

```bash
# Run all tests
make test

# Run tests with coverage
make test-coverage

# Run integration tests only
make test-integration
```

### Code Quality

```bash
# Format code
make fmt

# Lint code
make lint

# Check dependencies
make deps
```

### Project Structure

```
├── main.go                     # Entry point and application initialization
├── Makefile                    # Build and development tasks
├── internal/
│   ├── interfaces/             # Core interfaces and data structures
│   ├── scanner/                # Application discovery and desktop file parsing
│   ├── search/                 # Fuzzy search functionality
│   ├── config/                 # Configuration and favorites management
│   ├── launcher/               # Application launching functionality
│   ├── logging/                # Logging system with rotation
│   ├── errors/                 # Error handling and recovery
│   ├── tui/                    # Terminal user interface (Bubble Tea)
│   └── integration/            # Integration tests
└── go.mod                      # Go module definition
```

## Requirements

- Go 1.21 or later
- Linux operating system
- Terminal with ANSI color support (most modern terminals)

## Dependencies

- [Bubble Tea](https://github.com/charmbracelet/bubbletea) - TUI framework

## Troubleshooting

### Application Won't Start

1. Check that you have the required Go version: `go version`
2. Ensure all dependencies are installed: `make deps`
3. Check the log file: `~/.config/tui-launcher/launcher.log`

### Applications Not Found

1. Refresh the application list with `Ctrl+R`
2. Check that `.desktop` files exist in standard directories
3. Verify applications are not marked with `NoDisplay=true`

### Favorites Not Persisting

1. Check write permissions for `~/.config/tui-launcher/`
2. Verify the configuration file format is valid JSON
3. Check the log file for configuration errors

## Contributing

1. Fork the repository
2. Create a feature branch
3. Make your changes
4. Add tests for new functionality
5. Run `make test` and `make lint`
6. Submit a pull request

## License

[Add your license information here]