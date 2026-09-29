package config

import (
	"os"
	"path/filepath"
	"testing"

	"tui-app-launcher/internal/interfaces"
)

func TestHistoryPersistsAlongsideLegacyFavorites(t *testing.T) {
	t.Setenv("HOME", t.TempDir())
	m := NewManager()
	if err := m.AddFavorite("Browser"); err != nil {
		t.Fatal(err)
	}
	app := interfaces.Application{Name: "Browser", Exec: "browser", DesktopFile: "/apps/browser.desktop"}
	for i := 0; i < 2; i++ {
		if err := m.RecordLaunch(app); err != nil {
			t.Fatal(err)
		}
	}
	loaded := NewManager()
	record := loaded.LaunchHistory()[app.HistoryID()]
	if record.Count != 2 || record.LastUsed.IsZero() || !loaded.IsFavorite("Browser") {
		t.Fatalf("history/favorites did not survive reload: %+v", record)
	}
	copy := loaded.LaunchHistory()
	delete(copy, app.HistoryID())
	if loaded.LaunchHistory()[app.HistoryID()].Count != 2 {
		t.Fatal("history snapshot aliases config")
	}
	other := app
	other.DesktopFile = "/apps/other.desktop"
	if other.HistoryID() == app.HistoryID() {
		t.Fatal("equal names must have independent history")
	}
	if err := loaded.SaveFavorites([]string{"Other"}); err != nil {
		t.Fatal(err)
	}
	if NewManager().LaunchHistory()[app.HistoryID()].Count != 2 {
		t.Fatal("saving favorites erased history")
	}
}

func TestHistoryRollbackOnWriteFailure(t *testing.T) {
	t.Setenv("HOME", t.TempDir())
	m := NewManager()
	app := interfaces.Application{Name: "Browser", Exec: "browser"}
	if err := m.RecordLaunch(app); err != nil {
		t.Fatal(err)
	}
	bad := filepath.Join(t.TempDir(), "not-a-directory")
	if err := os.WriteFile(bad, []byte("x"), 0600); err != nil {
		t.Fatal(err)
	}
	m.SetConfigPath(bad)
	if err := m.RecordLaunch(app); err == nil {
		t.Fatal("expected a write failure")
	}
	if m.LaunchHistory()[app.HistoryID()].Count != 1 {
		t.Fatal("failed write changed in-memory history")
	}
}
