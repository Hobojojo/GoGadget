package interfaces

import "context"

// Application represents an installed application with its metadata
type Application struct {
	Name          string
	Exec          string
	Icon          string
	Comment       string
	GenericName   string
	Keywords      []string
	Categories    []string
	Terminal      bool
	Path          string
	StartupNotify bool
	DesktopFile   string
	IsFavorite    bool
}

// SearchResult represents a fuzzy search result with scoring information
type SearchResult struct {
	Application Application
	Score       int
	Matches     []int // Character positions that matched
}

// ApplicationScanner interface for discovering and managing applications
type ApplicationScanner interface {
	ScanApplications() ([]Application, error)
	ScanApplicationsContext(ctx context.Context) ([]Application, error)
	RefreshApplications() error
	RefreshApplicationsContext(ctx context.Context) error
}

// FuzzySearcher interface for fuzzy search functionality
type FuzzySearcher interface {
	Search(query string, items []Application) []SearchResult
	SetItems(items []Application)
	GetMatchPositions(query string, text string) []int
}

// ConfigManager interface for managing user configuration and favorites
type ConfigManager interface {
	LoadFavorites() ([]string, error)
	SaveFavorites(favorites []string) error
	AddFavorite(appName string) error
	RemoveFavorite(appName string) error
	IsFavorite(appName string) bool
}

// ApplicationLauncher interface for launching applications
type ApplicationLauncher interface {
	LaunchApplication(app Application) error
	ValidateApplication(app Application) error
	GetLaunchCommand(app Application) (string, []string, error)
}
