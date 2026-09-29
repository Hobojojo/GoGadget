package search

import (
	"testing"
	"tui-app-launcher/internal/interfaces"
)

func TestSearchSettingsAndCache(t *testing.T) {
	apps := []interfaces.Application{{Name: "Café Browser", Exec: "hidden-command"}}
	f := NewFuzzySearcher()
	f.SetItems(apps)
	if f.lower[apps[0].Name] != "café browser" {
		t.Fatal("names not normalized at indexing")
	}
	f.ConfigureSearch([]string{"name"}, true)
	if len(f.Search("café", apps)) != 0 || len(f.Search("Café", apps)) != 1 || len(f.Search("hidden", apps)) != 0 {
		t.Fatal("settings ignored")
	}
	f.ConfigureSearch([]string{"exec"}, false)
	if len(f.Search("HIDDEN", apps)) != 1 || len(f.GetMatchPositions("Browser", apps[0].Name)) != 0 {
		t.Fatal("exec-only settings ignored")
	}
	f.ConfigureSearch(nil, false)
	// Scattered fuzzy matches must not be rejected by the optimization.
	if len(f.Search("cb", apps)) != 1 {
		t.Fatal("scattered match lost")
	}
	apps[0].Name = "Updated Name"
	f.SetItems(apps)
	if len(f.Search("updated", apps)) != 1 || len(f.Search("café", apps)) != 0 {
		t.Fatal("stale cache after reindex")
	}
}

func BenchmarkIndexedSearch(b *testing.B) {
	apps := make([]interfaces.Application, 500)
	for i := range apps {
		apps[i] = interfaces.Application{Name: "Firefox Developer Edition", Exec: "firefox-dev", Comment: "Browse the web"}
	}
	f := NewFuzzySearcher()
	f.SetItems(apps)
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		f.Search("fire dev", apps)
	}
}
