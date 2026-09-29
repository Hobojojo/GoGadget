package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLegacyAndSettingsRoundTrip(t *testing.T) {
	t.Setenv("HOME", t.TempDir())
	dir := filepath.Join(os.Getenv("HOME"), ".config", "tui-launcher")
	if err := os.MkdirAll(dir, 0755); err != nil {
		t.Fatal(err)
	}
	file := filepath.Join(dir, "config.json")
	for _, data := range []string{`{"favorites":["Editor"]}`, `{"favorites":["Editor"],"settings":{"search_fields":["name"],"ranking":"frequency","multi_column":false,"case_sensitive":true}}`} {
		if err := os.WriteFile(file, []byte(data), 0600); err != nil {
			t.Fatal(err)
		}
		manager := NewManager()
		if !manager.IsFavorite("Editor") {
			t.Fatal("legacy favorites lost")
		}
		before := manager.Settings()
		if err := manager.AddFavorite("Browser"); err != nil {
			t.Fatal(err)
		}
		after := NewManager().Settings()
		if before.Ranking != after.Ranking || before.CaseSensitive != after.CaseSensitive {
			t.Fatal("settings lost on save")
		}
		if len(before.SearchFields) > 0 {
			before.SearchFields[0] = "exec"
			*before.MultiColumn = true
			snapshot := manager.Settings()
			if snapshot.SearchFields[0] != "name" || *snapshot.MultiColumn {
				t.Fatal("settings snapshot aliases config")
			}
		}
	}
}
