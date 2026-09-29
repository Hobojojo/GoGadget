package scanner

import (
	"reflect"
	"testing"

	"tui-app-launcher/internal/interfaces"
)

func TestRemoveDuplicatesPreservesOrderAndPrefersUserApp(t *testing.T) {
	s := NewScannerWithPaths(nil)
	apps := []interfaces.Application{
		{Name: "Alpha", Exec: "a", DesktopFile: "/usr/share/applications/a.desktop"},
		{Name: "Beta", Exec: "b", DesktopFile: "/usr/share/applications/b.desktop"},
		{Name: "Alpha", Exec: "a", DesktopFile: "/home/user/.local/share/applications/a.desktop"},
		{Name: "Gamma", Exec: "g", DesktopFile: "/usr/share/applications/g.desktop"},
		{Name: "Beta", Exec: "b", DesktopFile: "/usr/local/share/applications/b.desktop"},
	}
	want := []interfaces.Application{apps[2], apps[1], apps[3]}
	for i := 0; i < 20; i++ {
		if got := s.removeDuplicates(apps); !reflect.DeepEqual(got, want) {
			t.Fatalf("deduplicated applications changed order or preference: %+v", got)
		}
	}
}
