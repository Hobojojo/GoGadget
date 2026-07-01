package main

import "regexp"

// Pattern represents a regex pattern of interest for analysis.
type Pattern struct {
	Name        string
	Regex       *regexp.Regexp
	Description string
}

// DefaultPatterns returns a list of common patterns to look for in JS code.
func DefaultPatterns() []Pattern {
	return []Pattern{
		{
			Name:        "eval",
			Regex:       regexp.MustCompile(`(?i)\beval\s*\(`),
			Description: "Use of eval() function",
		},
		{
			Name:        "innerHTML",
			Regex:       regexp.MustCompile(`(?i)\.innerHTML\s*=`),
			Description: "Assignment to innerHTML",
		},
		{
			Name:        "document.write",
			Regex:       regexp.MustCompile(`(?i)document\.write\s*\(`),
			Description: "Use of document.write()",
		},
		{
			Name:        "setTimeout_string",
			Regex:       regexp.MustCompile(`(?i)setTimeout\s*\(\s*["']`),
			Description: "setTimeout with string argument",
		},
		{
			Name:        "Sensitive_Data",
			Regex:       regexp.MustCompile(`(?i)(api[_-]?key|secret|token|auth|password|credentials)["']?\s*[:=]\s*["']([a-zA-Z0-9\-_.~]{16,})["']`),
			Description: "Potential hardcoded sensitive data",
		},
		{
			Name:        "PostMessage",
			Regex:       regexp.MustCompile(`\.postMessage\s*\(`),
			Description: "Use of postMessage",
		},
		{
			Name:        "LocalStorage",
			Regex:       regexp.MustCompile(`localStorage\.setItem\s*\(`),
			Description: "Storage of data in localStorage",
		},
	}
}
