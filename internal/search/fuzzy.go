package search

import (
	"sort"
	"strings"
	"unicode"

	"tui-app-launcher/internal/interfaces"
)

// FuzzySearcher implements fuzzy search functionality
type FuzzySearcher struct {
	items []interfaces.Application
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
	tokens := strings.Fields(strings.ToLower(query))
	for _, item := range items {
		name := strings.ToLower(item.Name)
		metadata := append([]string{item.Exec, item.Comment, item.GenericName}, item.Categories...)
		metadata = append(metadata, item.Keywords...)
		total, complete := 0, true
		var positions []int
		for _, token := range tokens {
			score, matches := f.calculateScore(token, name)
			if len(matches) > 0 {
				total += score
				positions = append(positions, matches...)
				continue
			}
			found := false
			for _, text := range metadata {
				candidate, hits := f.calculateScore(token, strings.ToLower(text))
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
	targetRunes := []rune(target)

	// Track matched positions
	matches := make([]int, 0, len(queryRunes))
	score := 0
	queryIdx := 0

	for targetIdx, targetChar := range targetRunes {
		if queryIdx >= len(queryRunes) {
			break
		}

		if queryRunes[queryIdx] == targetChar {
			matches = append(matches, targetIdx)

			// Base score for character match
			charScore := 1

			// Position weighting bonuses
			if targetIdx == 0 {
				// Bonus for matching at start of string
				charScore += 3
			} else if targetIdx > 0 && isWordBoundary(targetRunes[targetIdx-1]) {
				// Bonus for matching at word boundary
				charScore += 2
			}

			// Consecutive character bonus
			if queryIdx > 0 && len(matches) > 1 {
				prevMatchIdx := matches[len(matches)-2]
				if targetIdx == prevMatchIdx+1 {
					charScore += 1
				}
			}

			// Case match bonus (if original characters match case)
			if len(query) > queryIdx && len(target) > targetIdx {
				if rune(query[queryIdx]) == rune(target[targetIdx]) {
					charScore += 1
				}
			}

			score += charScore
			queryIdx++
		}
	}

	// Only return results if all query characters were matched
	if queryIdx != len(queryRunes) {
		return 0, []int{}
	}

	// Apply length penalty for longer strings to prefer shorter matches
	lengthPenalty := len(targetRunes) - len(queryRunes)
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
	if query == "" || target == "" {
		return []int{}
	}

	var positions []int
	for _, token := range strings.Fields(strings.ToLower(query)) {
		_, matches := f.calculateScore(token, strings.ToLower(target))
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
