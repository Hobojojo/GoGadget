package scanner

import (
	"os"
	"path/filepath"
	"strings"

	"tui-app-launcher/internal/errors"
	"tui-app-launcher/internal/interfaces"
	"tui-app-launcher/internal/logging"
)

// Scanner implements the ApplicationScanner interface
type Scanner struct {
	applications []interfaces.Application
	scanPaths    []string
	logger       *logging.Logger
}

// NewScanner creates a new Scanner with default scan paths
func NewScanner() *Scanner {
	return &Scanner{
		applications: make([]interfaces.Application, 0),
		scanPaths: []string{
			"/usr/share/applications",
			"/usr/local/share/applications",
		},
		logger: logging.GetGlobalLogger(),
	}
}

// NewScannerWithPaths creates a new Scanner with custom scan paths
func NewScannerWithPaths(paths []string) *Scanner {
	return &Scanner{
		applications: make([]interfaces.Application, 0),
		scanPaths:    paths,
		logger:       logging.GetGlobalLogger(),
	}
}

// ScanApplications discovers applications from desktop files
func (s *Scanner) ScanApplications() ([]interfaces.Application, error) {
	var applications []interfaces.Application

	// Create a copy of scan paths to avoid modifying the original slice
	scanPaths := make([]string, len(s.scanPaths))
	copy(scanPaths, s.scanPaths)

	// Add user-specific applications directory if not already present
	homeDir, err := os.UserHomeDir()
	if err == nil {
		userAppsDir := filepath.Join(homeDir, ".local", "share", "applications")
		// Check if user apps dir is already in the scan paths
		found := false
		for _, path := range scanPaths {
			if path == userAppsDir {
				found = true
				break
			}
		}
		if !found {
			scanPaths = append(scanPaths, userAppsDir)
		}
	}

	// Scan each directory
	var scanErrors []error
	totalFound := 0

	for _, scanPath := range scanPaths {
		apps, err := s.scanDirectory(scanPath)
		if err != nil {
			// Collect scan errors but continue with other directories
			scanErrors = append(scanErrors, errors.NewScanError("failed to scan directory", err).
				WithContext("directory", scanPath))
			if s.logger != nil {
				s.logger.LogApplicationScan(scanPath, 0, 1)
			}
			continue
		}

		applications = append(applications, apps...)
		totalFound += len(apps)

		if s.logger != nil {
			s.logger.LogApplicationScan(scanPath, len(apps), 0)
		}
	}

	// Log overall scan results
	if s.logger != nil {
		s.logger.Info("Application scan completed: total_found=%d, failed_directories=%d", totalFound, len(scanErrors))
	}

	// If we have scan errors but found some applications, return partial results
	if len(scanErrors) > 0 && len(applications) > 0 {
		// Return applications with a warning about partial scan
		s.applications = applications
		return applications, errors.NewScanError("partial scan completed", nil).
			WithContext("failed_directories", len(scanErrors)).
			WithContext("found_applications", len(applications))
	}

	// If we have scan errors and no applications, return the first error
	if len(scanErrors) > 0 && len(applications) == 0 {
		return nil, scanErrors[0]
	}

	// Remove duplicates (prefer user applications over system ones)
	applications = s.removeDuplicates(applications)

	s.applications = applications
	return applications, nil
}

// RefreshApplications rescans for newly installed applications
func (s *Scanner) RefreshApplications() error {
	_, err := s.ScanApplications()
	return err
}

// scanDirectory scans a single directory for .desktop files
func (s *Scanner) scanDirectory(dirPath string) ([]interfaces.Application, error) {
	var applications []interfaces.Application

	// Check if directory exists
	if _, err := os.Stat(dirPath); os.IsNotExist(err) {
		return applications, nil // Return empty slice, not an error
	}

	err := filepath.Walk(dirPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			// Skip files we can't access
			return nil
		}

		// Only process .desktop files
		if !strings.HasSuffix(strings.ToLower(info.Name()), ".desktop") {
			return nil
		}

		// Parse the desktop file
		entry, err := ParseDesktopFile(path)
		if err != nil {
			// Skip invalid desktop files but don't treat as fatal error
			return nil
		}

		// Check if it's a valid application
		if !entry.IsValidApplication() {
			return nil
		}

		// Convert to Application struct
		app := entry.ToApplication(path)
		applications = append(applications, app)

		return nil
	})

	return applications, err
}

// SetScanPaths sets custom scan paths for testing
func (s *Scanner) SetScanPaths(paths []string) {
	s.scanPaths = paths
}

// removeDuplicates removes duplicate applications, preferring user applications over system ones
func (s *Scanner) removeDuplicates(applications []interfaces.Application) []interfaces.Application {
	seen := make(map[string]interfaces.Application)

	for _, app := range applications {
		key := app.Name + "|" + app.Exec

		existing, exists := seen[key]
		if !exists {
			seen[key] = app
		} else {
			// Prefer user applications (in ~/.local/share/applications)
			if strings.Contains(app.DesktopFile, ".local/share/applications") &&
				!strings.Contains(existing.DesktopFile, ".local/share/applications") {
				seen[key] = app
			}
		}
	}

	// Convert map back to slice
	result := make([]interfaces.Application, 0, len(seen))
	for _, app := range seen {
		result = append(result, app)
	}

	return result
}
