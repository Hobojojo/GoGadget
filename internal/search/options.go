package search

import "strings"

var defaultFields = []string{"name", "exec", "comment", "generic_name", "categories", "keywords"}

func (f *FuzzySearcher) ConfigureSearch(fields []string, caseSensitive bool) {
	f.fields = append([]string(nil), fields...)
	f.caseSensitive = caseSensitive
}

func (f *FuzzySearcher) fieldEnabled(field string) bool {
	fields := f.fields
	if len(fields) == 0 {
		fields = defaultFields
	}
	for _, value := range fields {
		if value == field {
			return true
		}
	}
	return false
}

func (f *FuzzySearcher) normalize(text string) string {
	if f.caseSensitive {
		return text
	}
	if lower, ok := f.lower[text]; ok {
		return lower
	}
	return strings.ToLower(text)
}
