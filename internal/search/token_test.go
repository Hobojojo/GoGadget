package search

import (
	"reflect"
	"testing"

	"tui-app-launcher/internal/interfaces"
)

func TestTokensAcrossNamesAndMetadata(t *testing.T) {
	f := NewFuzzySearcher()
	apps := []interfaces.Application{
		{Name: "Firefox Developer Edition", Exec: "/opt/browser-bin", Keywords: []string{"web"}},
		{Name: "Firefox", Exec: "firefox"},
	}
	for _, query := range []string{"fire dev", "dev fire", "  FIRE\tdev ", "fire browser-bin", "web dev"} {
		got := f.Search(query, apps)
		if len(got) != 1 || got[0].Application.Name != apps[0].Name {
			t.Fatalf("%q: %+v", query, got)
		}
	}
	if got := f.Search("fire impossible", apps); len(got) != 0 {
		t.Fatal("all tokens must match")
	}
	if got := f.Search(" \t ", apps); len(got) != 2 {
		t.Fatal("whitespace-only query should return all apps")
	}
	want := []int{0, 1, 2, 3, 8, 9, 10}
	if got := f.GetMatchPositions("dev fire fire", apps[0].Name); !reflect.DeepEqual(got, want) {
		t.Fatalf("deduplicated highlights: %v", got)
	}
	got := f.Search("browser-bin", apps)
	if len(got) != 1 || len(got[0].Matches) != 0 {
		t.Fatalf("exec-only hit should not highlight the name: %+v", got)
	}
}
