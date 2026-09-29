package scanner

import (
	"os"
	"path/filepath"
)

// defaultScanPaths follows XDG_DATA_DIRS order so the first data directory wins.
func defaultScanPaths() []string {
	dataDirs := os.Getenv("XDG_DATA_DIRS")
	if dataDirs == "" {
		dataDirs = "/usr/local/share:/usr/share"
	}

	var paths []string
	seen := make(map[string]bool)
	for _, dir := range filepath.SplitList(dataDirs) {
		// XDG base directories must be absolute; empty entries are not cwd.
		if !filepath.IsAbs(dir) {
			continue
		}
		path := filepath.Join(dir, "applications")
		if !seen[path] {
			seen[path] = true
			paths = append(paths, path)
		}
	}
	return paths
}
