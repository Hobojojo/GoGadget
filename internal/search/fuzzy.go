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
	if query == "" {
		// Return all items with zero score when query is empty
		results := make([]interfaces.SearchResult, 0, len(items))
		for _, item := range items {
			if item.Name == "" {
				continue
			}
			results = append(results, interfaces.SearchResult{
				Application: item,
				Score:       0,
				Matches:     []int{},
			})
		}
		return results
	}

	var results []interfaces.SearchResult
	queryLower := strings.ToLower(query)
	queryRunes := []rune(queryLower)

	for _, item := range items {
		if item.Name == "" {
			continue
		}

		// Try matching Name (highest priority)
		score, matches := f.calculateScore(queryRunes, strings.ToLower(item.Name))

		// Try matching Comment if no match in name or to potentially increase score
		commentScore, _ := f.calculateScore(queryRunes, strings.ToLower(item.Comment))
		if commentScore > 0 {
			// If we matched both, take the better score (usually name)
			// For simplicity and since we only return one set of matches (for the name),
			// we prioritize name matches but allow finding by comment.
			if score == 0 {
				score = commentScore / 2 // Comment matches have lower weight
				matches = []int{}        // Don't highlight name if we matched comment
			}
		}

		// Try matching Categories
		for _, cat := range item.Categories {
			catScore, _ := f.calculateScore(queryRunes, strings.ToLower(cat))
			if catScore > 0 {
				if score == 0 {
					score = catScore / 3 // Category matches have lowest weight
					matches = []int{}
				}
				break
			}
		}

		if score > 0 {
			results = append(results, interfaces.SearchResult{
				Application: item,
				Score:       score,
				Matches:     matches,
			})
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
func (f *FuzzySearcher) calculateScore(queryRunes []rune, target string) (int, []int) {
	if len(queryRunes) == 0 {
		return 0, []int{}
	}
	if len(target) == 0 {
		return 0, []int{}
	}

	targetRunes := []rune(target)

	// Track matched positions
	matches := make([]int, 0, len(queryRunes))
	score := 0
	queryIdx := 0

SEARCH_LOOP:
	for targetIdx := 0; targetIdx < len(targetRunes); targetIdx++ {
		targetChar := targetRunes[targetIdx]
		if queryIdx >= len(queryRunes) {
			break
		}

		if queryRunes[queryIdx] == targetChar {
			// Check if there is a better match later (e.g. at a word boundary)
			// for the CURRENT query character, but only if we are not at a word boundary now.
			if !isWordBoundaryOrStart(targetRunes, targetIdx) {
				// Quick check: can we still match everything if we skip this 'targetChar'?
				// Actually, we specifically want to know if queryRunes[queryIdx] appears later at a better position.
				foundBetter := false
				for nextIdx := targetIdx + 1; nextIdx < len(targetRunes); nextIdx++ {
					if targetRunes[nextIdx] == queryRunes[queryIdx] && isWordBoundaryOrStart(targetRunes, nextIdx) {
						foundBetter = true
						break
					}
				}

				if foundBetter {
					// Also must ensure we can match the REST of the query from that better position or later.
					qIdx := queryIdx
					for tIdx := targetIdx + 1; tIdx < len(targetRunes); tIdx++ {
						if qIdx < len(queryRunes) && targetRunes[tIdx] == queryRunes[qIdx] {
							qIdx++
						}
					}
					if qIdx == len(queryRunes) {
						continue SEARCH_LOOP
					}
				}
			}

			matches = append(matches, targetIdx)

			// Base score for character match
			charScore := 10 // Increase base score to allow better granularity

			// Position weighting bonuses
			if targetIdx == 0 {
				// Bonus for matching at start of string
				charScore += 30
			} else if targetIdx > 0 && isWordBoundary(targetRunes[targetIdx-1]) {
				// Bonus for matching at word boundary
				charScore += 20
			}

			// Consecutive character bonus
			if queryIdx > 0 && len(matches) > 1 {
				prevMatchIdx := matches[len(matches)-2]
				if targetIdx == prevMatchIdx+1 {
					charScore += 15
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
		score -= lengthPenalty
	}

	return score, matches
}

// isWordBoundary checks if a character represents a word boundary
func isWordBoundary(r rune) bool {
	return unicode.IsSpace(r) || r == '-' || r == '_' || r == '.' || r == '/'
}

// isWordBoundaryOrStart checks if a character is at a word boundary or start of string
func isWordBoundaryOrStart(runes []rune, idx int) bool {
	if idx == 0 {
		return true
	}
	return isWordBoundary(runes[idx-1])
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

	queryLower := strings.ToLower(query)
	targetLower := strings.ToLower(target)
	queryRunes := []rune(queryLower)

	_, matches := f.calculateScore(queryRunes, targetLower)
	return matches
}
