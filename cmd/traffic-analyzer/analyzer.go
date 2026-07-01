package main

import (
	"fmt"
	"strings"
)

// Finding represents a pattern match found in a script.
type Finding struct {
	PatternName string
	Description string
	Snippet     string
	LineNumber  int
}

// AnalyzeScript scans a script for patterns and returns a list of findings.
func AnalyzeScript(script ExtractedScript, patterns []Pattern) []Finding {
	var findings []Finding
	lines := strings.Split(script.Content, "\n")

	for i, line := range lines {
		for _, p := range patterns {
			if p.Regex.MatchString(line) {
				findings = append(findings, Finding{
					PatternName: p.Name,
					Description: p.Description,
					Snippet:     strings.TrimSpace(line),
					LineNumber:  i + 1,
				})
			}
		}
	}

	return findings
}

// PrintFindings displays the findings in a readable format.
func PrintFindings(source string, findings []Finding) {
	if len(findings) == 0 {
		return
	}

	fmt.Printf("\n[+] Findings in %s:\n", source)
	for _, f := range findings {
		fmt.Printf("  - [%s] (Line %d): %s\n", f.PatternName, f.LineNumber, f.Description)
		fmt.Printf("    Snippet: %s\n", f.Snippet)
	}
}
