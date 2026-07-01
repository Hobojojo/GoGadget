package search

import (
	"reflect"
	"testing"

	"tui-app-launcher/internal/interfaces"
)

func TestFuzzySearcher_Search(t *testing.T) {
	searcher := NewFuzzySearcher()

	// Test applications
	apps := []interfaces.Application{
		{Name: "Firefox", Exec: "firefox", Comment: "Web browser"},
		{Name: "File Manager", Exec: "nautilus", Comment: "File manager"},
		{Name: "Terminal", Exec: "gnome-terminal", Comment: "Terminal emulator"},
		{Name: "Text Editor", Exec: "gedit", Comment: "Text editor"},
		{Name: "Calculator", Exec: "gnome-calculator", Comment: "Calculator app"},
		{Name: "Firefox Developer Edition", Exec: "firefox-dev", Comment: "Developer browser"},
	}

	tests := []struct {
		name          string
		query         string
		expectedCount int
		expectedFirst string // Name of first result
		description   string
	}{
		{
			name:          "Empty query returns all items",
			query:         "",
			expectedCount: 6,
			expectedFirst: "Firefox", // First in original order
			description:   "Empty query should return all applications with zero scores",
		},
		{
			name:          "Exact match",
			query:         "Firefox",
			expectedCount: 2,
			expectedFirst: "Firefox",
			description:   "Exact match should score highest",
		},
		{
			name:          "Partial match",
			query:         "fire",
			expectedCount: 2,
			expectedFirst: "Firefox",
			description:   "Partial matches should work",
		},
		{
			name:          "Case insensitive",
			query:         "FIREFOX",
			expectedCount: 2,
			expectedFirst: "Firefox",
			description:   "Search should be case insensitive",
		},
		{
			name:          "Character sequence match",
			query:         "ff",
			expectedCount: 2,
			expectedFirst: "Firefox",
			description:   "Should match character sequences",
		},
		{
			name:          "Word boundary preference",
			query:         "fm",
			expectedCount: 1,
			expectedFirst: "File Manager",
			description:   "Should prefer word boundary matches",
		},
		{
			name:          "No matches",
			query:         "xyz",
			expectedCount: 0,
			expectedFirst: "",
			description:   "Non-matching query should return no results",
		},
		{
			name:          "Single character",
			query:         "t",
			expectedCount: 3, // Terminal, Text Editor, Calculator (app)
			expectedFirst: "Terminal", // Should score higher due to start position
			description:   "Single character should match multiple apps",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			results := searcher.Search(tt.query, apps)

			if len(results) != tt.expectedCount {
				t.Errorf("Expected %d results, got %d", tt.expectedCount, len(results))
			}

			if tt.expectedCount > 0 && results[0].Application.Name != tt.expectedFirst {
				t.Errorf("Expected first result to be '%s', got '%s'", tt.expectedFirst, results[0].Application.Name)
			}

			// Verify results are sorted by score (descending)
			for i := 1; i < len(results); i++ {
				if results[i-1].Score < results[i].Score {
					t.Errorf("Results not properly sorted by score: %d < %d at positions %d, %d",
						results[i-1].Score, results[i].Score, i-1, i)
				}
			}
		})
	}
}

func TestFuzzySearcher_ScoreRanking(t *testing.T) {
	searcher := NewFuzzySearcher()

	apps := []interfaces.Application{
		{Name: "Firefox", Exec: "firefox"},
		{Name: "File Firefox", Exec: "file-firefox"},
		{Name: "My Firefox Browser", Exec: "my-firefox"},
		{Name: "Firefox Developer Edition", Exec: "firefox-dev"},
	}

	results := searcher.Search("firefox", apps)

	if len(results) != 4 {
		t.Fatalf("Expected 4 results, got %d", len(results))
	}

	// "Firefox" should score highest (exact match at start)
	if results[0].Application.Name != "Firefox" {
		t.Errorf("Expected 'Firefox' to score highest, got '%s'", results[0].Application.Name)
	}

	// Verify scores are in descending order
	for i := 1; i < len(results); i++ {
		if results[i-1].Score < results[i].Score {
			t.Errorf("Scores not in descending order: %d < %d", results[i-1].Score, results[i].Score)
		}
	}

	// Test that start-of-string matches score higher than word boundary matches
	wordBoundaryApps := []interfaces.Application{
		{Name: "Terminal", Exec: "terminal"},
		{Name: "My Terminal", Exec: "my-terminal"},
	}

	termResults := searcher.Search("terminal", wordBoundaryApps)
	if len(termResults) != 2 {
		t.Fatalf("Expected 2 results for terminal search, got %d", len(termResults))
	}

	if termResults[0].Application.Name != "Terminal" {
		t.Errorf("Expected 'Terminal' to score higher than 'My Terminal', got '%s' first",
			termResults[0].Application.Name)
	}
}

func TestFuzzySearcher_MatchPositions(t *testing.T) {
	searcher := NewFuzzySearcher()

	apps := []interfaces.Application{
		{Name: "Firefox", Exec: "firefox"},
		{Name: "File Manager", Exec: "nautilus"},
	}

	tests := []struct {
		name            string
		query           string
		appName         string
		expectedMatches []int
	}{
		{
			name:            "Consecutive characters",
			query:           "fire",
			appName:         "Firefox",
			expectedMatches: []int{0, 1, 2, 3}, // F-i-r-e
		},
		{
			name:            "Word boundary matches",
			query:           "fm",
			appName:         "File Manager",
			expectedMatches: []int{0, 5}, // F-ile M-anager
		},
		{
			name:            "Scattered matches",
			query:           "fx",
			appName:         "Firefox",
			expectedMatches: []int{0, 6}, // F-irefo-x (index 6 is 'x')
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			results := searcher.Search(tt.query, apps)

			var targetResult *interfaces.SearchResult
			for _, result := range results {
				if result.Application.Name == tt.appName {
					targetResult = &result
					break
				}
			}

			if targetResult == nil {
				t.Fatalf("Expected to find result for '%s'", tt.appName)
			}

			if !reflect.DeepEqual(targetResult.Matches, tt.expectedMatches) {
				t.Errorf("Expected matches %v, got %v", tt.expectedMatches, targetResult.Matches)
			}
		})
	}
}

func TestFuzzySearcher_EdgeCases(t *testing.T) {
	searcher := NewFuzzySearcher()

	apps := []interfaces.Application{
		{Name: "", Exec: "empty"},
		{Name: "A", Exec: "single"},
		{Name: "Special-Chars.App", Exec: "special"},
	}

	tests := []struct {
		name        string
		query       string
		expectCount int
		description string
	}{
		{
			name:        "Empty app name",
			query:       "chars",
			expectCount: 1, // Matches "Special-Chars.App", but should skip the empty name one
			description: "Should handle empty app names gracefully",
		},
		{
			name:        "Single character app",
			query:       "a",
			expectCount: 2, // "A" (name) and "Special-Chars_Test.App" (contains 'a')
			description: "Should match single character apps",
		},
		{
			name:        "Special characters",
			query:       "special",
			expectCount: 1,
			description: "Should handle special characters in app names",
		},
		{
			name:        "Query longer than app name",
			query:       "verylongquery",
			expectCount: 0,
			description: "Should handle queries longer than any app name",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			results := searcher.Search(tt.query, apps)
			if len(results) != tt.expectCount {
				t.Errorf("Expected %d results, got %d", tt.expectCount, len(results))
			}
		})
	}
}

func TestFuzzySearcher_HighlightMatches(t *testing.T) {
	searcher := NewFuzzySearcher()

	tests := []struct {
		name     string
		text     string
		matches  []int
		expected string
	}{
		{
			name:     "No matches",
			text:     "Firefox",
			matches:  []int{},
			expected: "Firefox",
		},
		{
			name:     "Single match",
			text:     "Firefox",
			matches:  []int{0},
			expected: "\033[1;33mF\033[0mirefox",
		},
		{
			name:     "Multiple matches",
			text:     "Firefox",
			matches:  []int{0, 1, 2},
			expected: "\033[1;33mF\033[0m\033[1;33mi\033[0m\033[1;33mr\033[0mefox",
		},
		{
			name:     "Empty text",
			text:     "",
			matches:  []int{0},
			expected: "",
		},
		{
			name:     "Out of bounds matches",
			text:     "Test",
			matches:  []int{0, 10},
			expected: "\033[1;33mT\033[0mest",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := searcher.HighlightMatches(tt.text, tt.matches)
			if result != tt.expected {
				t.Errorf("Expected '%s', got '%s'", tt.expected, result)
			}
		})
	}
}

func TestFuzzySearcher_GetMatchPositions(t *testing.T) {
	searcher := NewFuzzySearcher()

	tests := []struct {
		name     string
		query    string
		target   string
		expected []int
	}{
		{
			name:     "Basic match",
			query:    "fire",
			target:   "Firefox",
			expected: []int{0, 1, 2, 3},
		},
		{
			name:     "Case insensitive",
			query:    "FIRE",
			target:   "firefox",
			expected: []int{0, 1, 2, 3},
		},
		{
			name:     "No match",
			query:    "xyz",
			target:   "Firefox",
			expected: []int{},
		},
		{
			name:     "Empty query",
			query:    "",
			target:   "Firefox",
			expected: []int{},
		},
		{
			name:     "Empty target",
			query:    "test",
			target:   "",
			expected: []int{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := searcher.GetMatchPositions(tt.query, tt.target)
			if !reflect.DeepEqual(result, tt.expected) {
				t.Errorf("Expected %v, got %v", tt.expected, result)
			}
		})
	}
}

func TestFuzzySearcher_SetItems(t *testing.T) {
	searcher := NewFuzzySearcher()

	apps := []interfaces.Application{
		{Name: "Test App", Exec: "test"},
	}

	// Test that SetItems doesn't panic and stores items
	searcher.SetItems(apps)

	// Verify items are accessible (indirectly through search)
	results := searcher.Search("test", apps)
	if len(results) != 1 {
		t.Errorf("Expected 1 result after SetItems, got %d", len(results))
	}
}

// Benchmark tests for performance verification
func BenchmarkFuzzySearcher_Search(b *testing.B) {
	searcher := NewFuzzySearcher()

	// Create a larger set of test applications
	apps := make([]interfaces.Application, 100)
	for i := 0; i < 100; i++ {
		apps[i] = interfaces.Application{
			Name: "Application " + string(rune('A'+i%26)) + string(rune('a'+i%26)),
			Exec: "app" + string(rune('0'+i%10)),
		}
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		searcher.Search("app", apps)
	}
}

func BenchmarkFuzzySearcher_CalculateScore(b *testing.B) {
	searcher := NewFuzzySearcher()
	queryRunes := []rune("firefox")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		searcher.calculateScore(queryRunes, "firefox web browser")
	}
}
