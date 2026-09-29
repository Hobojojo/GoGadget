package search

import (
	"reflect"
	"strings"
	"testing"

	"tui-app-launcher/internal/interfaces"
)

func TestSearchKeepsNonpositiveMatches(t *testing.T) {
	searcher := NewFuzzySearcher()
	for _, suffixLength := range []int{20, 80} {
		name := "A" + strings.Repeat("b", suffixLength)
		t.Run(name, func(t *testing.T) {
			apps := []interfaces.Application{{Name: name}, {Name: "A"}, {Name: "Unrelated"}}
			results := searcher.Search("a", apps[:2])
			if len(results) != 2 {
				t.Fatalf("expected both complete matches, got %d", len(results))
			}
			if results[0].Application.Name != "A" || results[1].Application.Name != name {
				t.Fatal("length penalty should rank the shorter match first, not remove the longer one")
			}
			if results[1].Score > 0 {
				t.Fatalf("regression fixture must exercise a nonpositive score, got %d", results[1].Score)
			}
			if !reflect.DeepEqual(results[1].Matches, []int{0}) {
				t.Fatalf("match positions were lost: %v", results[1].Matches)
			}
			if got := searcher.Search("az", apps); len(got) != 0 {
				t.Fatal("incomplete subsequences must still be rejected")
			}
		})
	}
}

func TestSearchEmptyNameDoesNotMatch(t *testing.T) {
	if got := NewFuzzySearcher().Search("test", []interfaces.Application{{Name: ""}}); len(got) != 0 {
		t.Fatalf("empty name matched a nonempty query: %v", got)
	}
}
