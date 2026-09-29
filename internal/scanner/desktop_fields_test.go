package scanner

import (
	"os"
	"path/filepath"
	"reflect"
	"testing"
)

func TestAdditionalDesktopFields(t *testing.T) {
	file := filepath.Join(t.TempDir(), "test.desktop")
	contents := `[Desktop Entry]
Name=Editor
GenericName=Text Editor
Exec=editor
Keywords=write; code;;
Terminal=true
Path=/tmp
Hidden=false
StartupNotify=true
`
	if err := os.WriteFile(file, []byte(contents), 0600); err != nil {
		t.Fatal(err)
	}
	entry, err := ParseDesktopFile(file)
	if err != nil {
		t.Fatal(err)
	}
	if entry.GenericName != "Text Editor" || !reflect.DeepEqual(entry.Keywords, []string{"write", "code"}) || !entry.Terminal || entry.Path != "/tmp" || !entry.StartupNotify || entry.Hidden {
		t.Fatalf("missing parsed desktop fields: %+v", entry)
	}
	app := entry.ToApplication(file)
	if app.GenericName != entry.GenericName || !reflect.DeepEqual(app.Keywords, entry.Keywords) || !app.Terminal || app.Path != entry.Path || !app.StartupNotify {
		t.Fatalf("desktop metadata not propagated: %+v", app)
	}
	entry.Hidden = true
	if entry.IsValidApplication() {
		t.Fatal("Hidden=true entries must not appear in the launcher")
	}
}
