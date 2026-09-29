package search

import (
	"sort"
	"strings"
	"unicode"

	"tui-app-launcher/internal/interfaces"
)

// FuzzySearcher implements fuzzy search functionality
type FuzzySearcher struct {
	items         []interfaces.Application
	lower         map[string]string
	fields        []string
	caseSensitive bool
}

// NewFuzzySearcher creates a new fuzzy searcher instance
func NewFuzzySearcher() *FuzzySearcher {
	return &FuzzySearcher{
		items: make([]interfaces.Application, 0),
	}
}

// Search performs fuzzy search on applications
func (f *FuzzySearcher) Search(query string, items []interfaces.Application) []interfaces.SearchResult {
	if len(strings.Fields(query)) == 0 {
		// Return all items with zero score when query is empty
		results := make([]interfaces.SearchResult, len(items))
		for i, item := range items {
			results[i] = interfaces.SearchResult{
				Application: item,
				Score:       0,
				Matches:     []int{},
			}
		}
		return results
	}

	results := make([]interfaces.SearchResult, 0)
	tokens := strings.Fields(f.normalize(query))
	for _, item := range items {
		name := f.normalize(item.Name)
		metadata := make([]string, 0, 3+len(item.Categories)+len(item.Keywords))
		if f.fieldEnabled("exec") {
			metadata = append(metadata, item.Exec)
		}
		if f.fieldEnabled("comment") {
			metadata = append(metadata, item.Comment)
		}
		if f.fieldEnabled("generic_name") {
			metadata = append(metadata, item.GenericName)
		}
		if f.fieldEnabled("categories") {
			metadata = append(metadata, item.Categories...)
		}
		if f.fieldEnabled("keywords") {
			metadata = append(metadata, item.Keywords...)
		}
		total, complete := 0, true
		var positions []int
		for _, token := range tokens {
			score, matches := 0, []int(nil)
			if f.fieldEnabled("name") {
				score, matches = f.calculateScore(token, name)
			}
			if len(matches) > 0 {
				total += score
				positions = append(positions, matches...)
				continue
			}
			found := false
			for _, text := range metadata {
				candidate, hits := f.calculateScore(token, f.normalize(text))
				if len(hits) > 0 && (!found || candidate > score) {
					score, found = candidate, true
				}
			}
			if !found {
				complete = false
				break
			}
			total += score - 1 // Prefer name matches over metadata matches.
		}
		if complete {
			results = append(results, interfaces.SearchResult{Application: item, Score: total, Matches: uniquePositions(positions)})
		}
	}

	// Sort results by score (descending), then by name (ascending)
	sort.Slice(results, func(i, j int) bool {
		if results[i].Score == results[j].Score {
			return results[i].Application.Name < results[j].Application.Name
		}
		return results[i].Score > results[j].Score
	})

	return results
}

// SetItems updates the searchable items
func (f *FuzzySearcher) SetItems(items []interfaces.Application) {
	f.items = items
	f.lower = make(map[string]string, len(items)*4)
	for _, item := range items {
		texts := append([]string{item.Name, item.Exec, item.Comment, item.GenericName}, item.Categories...)
		texts = append(texts, item.Keywords...)
		for _, text := range texts {
			f.lower[text] = strings.ToLower(text)
		}
	}
}

// calculateScore computes the fuzzy match score and character positions
func (f *FuzzySearcher) calculateScore(query, target string) (int, []int) {
	if len(query) == 0 {
		return 0, []int{}
	}
	if len(target) == 0 {
		return 0, []int{}
	}

	queryRunes := []rune(query)
	// A cheap subsequence pass rejects impossible matches without allocating
	// positions. Unlike a substring filter, it preserves scattered fuzzy matches.
	matched := 0
	targetLength := 0
	for _, r := range target {
		targetLength++
		if matched < len(queryRunes) && r == queryRunes[matched] {
			matched++
		}
	}
	if matched != len(queryRunes) {
		return 0, nil
	}
	matches := make([]int, 0, len(queryRunes))
	score, queryIdx := 0, 0
	previous := rune(0)
	targetIdx := 0
	for _, targetChar := range target {
		if queryIdx >= len(queryRunes) {
			break
		}
		if queryRunes[queryIdx] == targetChar {
			charScore := 2
			if targetIdx == 0 {
				charScore += 3
			} else if isWordBoundary(previous) {
				charScore += 2
			}
			if len(matches) > 0 && targetIdx == matches[len(matches)-1]+1 {
				charScore++
			}
			matches = append(matches, targetIdx)
			score += charScore
			queryIdx++
		}
		previous = targetChar
		targetIdx++
	}

	// Only return results if all query characters were matched
	if queryIdx != len(queryRunes) {
		return 0, []int{}
	}

	// Apply length penalty for longer strings to prefer shorter matches
	lengthPenalty := targetLength - len(queryRunes)
	if lengthPenalty > 0 {
		score -= lengthPenalty / 4 // Mild penalty
	}

	return score, matches
}

// isWordBoundary checks if a character represents a word boundary
func isWordBoundary(r rune) bool {
	return unicode.IsSpace(r) || r == '-' || r == '_' || r == '.' || r == '/'
}

// HighlightMatches returns the text with highlighted matching characters
// Uses ANSI escape codes for terminal highlighting
func (f *FuzzySearcher) HighlightMatches(text string, matches []int) string {
	if len(matches) == 0 {
		return text
	}

	runes := []rune(text)
	if len(runes) == 0 {
		return text
	}

	// Create a map for quick lookup of match positions
	matchMap := make(map[int]bool)
	for _, pos := range matches {
		if pos >= 0 && pos < len(runes) {
			matchMap[pos] = true
		}
	}

	var result strings.Builder
	for i, r := range runes {
		if matchMap[i] {
			// Add highlighting for matched character
			result.WriteString("\033[1;33m") // Bold yellow
			result.WriteRune(r)
			result.WriteString("\033[0m") // Reset
		} else {
			result.WriteRune(r)
		}
	}

	return result.String()
}

// GetMatchPositions returns the character positions that matched for a given search
// This is a convenience method that re-runs the scoring to get match positions
func (f *FuzzySearcher) GetMatchPositions(query, target string) []int {
	if query == "" || target == "" || !f.fieldEnabled("name") {
		return []int{}
	}

	var positions []int
	for _, token := range strings.Fields(f.normalize(query)) {
		_, matches := f.calculateScore(token, f.normalize(target))
		positions = append(positions, matches...)
	}
	return uniquePositions(positions)
}

func uniquePositions(positions []int) []int {
	if len(positions) == 0 {
		return []int{}
	}
	sort.Ints(positions)
	unique := positions[:0]
	for _, pos := range positions {
		if len(unique) == 0 || unique[len(unique)-1] != pos {
			unique = append(unique, pos)
		}
	}
	return unique
}
