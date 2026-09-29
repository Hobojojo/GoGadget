package search

import (
	"testing"

	"tui-app-launcher/internal/interfaces"
)

func TestSearchGenericNameAndKeywords(t *testing.T) {
	items := []interfaces.Application{{Name: "Writer", GenericName: "Text Editor", Keywords: []string{"markdown", "notes"}}}
	f := NewFuzzySearcher()
	for _, query := range []string{"Editor", "markdown"} {
		got := f.Search(query, items)
		if len(got) != 1 || got[0].Application.Name != "Writer" {
			t.Fatalf("searching %q did not match desktop metadata: %+v", query, got)
		}
	}
}
