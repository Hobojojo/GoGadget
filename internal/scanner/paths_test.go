package scanner

import (
	"os"
	"path/filepath"
	"reflect"
	"testing"
)

func TestDefaultScanPaths(t *testing.T) {
	for _, tt := range []struct {
		name string
		env  string
		want []string
	}{
		{"empty uses XDG defaults", "", []string{"/usr/local/share/applications", "/usr/share/applications"}},
		{"custom replaces defaults in order", "/opt/flatpak/exports/share:/var/lib/snapd/desktop", []string{"/opt/flatpak/exports/share/applications", "/var/lib/snapd/desktop/applications"}},
		{"ignores relative and empty entries and removes duplicates", ":relative:/opt/share/:/opt/share:/var/share", []string{"/opt/share/applications", "/var/share/applications"}},
	} {
		t.Run(tt.name, func(t *testing.T) {
			t.Setenv("XDG_DATA_DIRS", tt.env)
			if got := NewScanner().scanPaths; !reflect.DeepEqual(got, tt.want) {
				t.Fatalf("scan paths = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestScanApplicationsFromXDGDataDirs(t *testing.T) {
	root := t.TempDir()
	home := filepath.Join(root, "home")
	first := filepath.Join(root, "flatpak", "exports", "share")
	second := filepath.Join(root, "snap", "desktop")
	t.Setenv("HOME", home)
	t.Setenv("XDG_DATA_DIRS", first+":"+second+":"+first)

	writeEntry := func(dataDir, name, exec string) string {
		t.Helper()
		dir := filepath.Join(dataDir, "applications")
		if err := os.MkdirAll(dir, 0755); err != nil {
			t.Fatal(err)
		}
		path := filepath.Join(dir, name+".desktop")
		content := "[Desktop Entry]\nType=Application\nName=" + name + "\nExec=" + exec + "\n"
		if err := os.WriteFile(path, []byte(content), 0644); err != nil {
			t.Fatal(err)
		}
		return path
	}

	preferred := writeEntry(first, "Shared", "shared")
	writeEntry(second, "Shared", "shared")
	userOverride := writeEntry(filepath.Join(home, ".local", "share"), "User", "user")
	writeEntry(first, "User", "user")
	snapOnly := writeEntry(second, "SnapOnly", "snap-only")

	apps, err := NewScanner().ScanApplications()
	if err != nil {
		t.Fatal(err)
	}
	got := make(map[string]string)
	for _, app := range apps {
		got[app.Name] = app.DesktopFile
	}
	want := map[string]string{"Shared": preferred, "User": userOverride, "SnapOnly": snapOnly}
	if len(apps) != len(want) || !reflect.DeepEqual(got, want) {
		t.Fatalf("applications = %+v, want paths %v", apps, want)
	}
}
