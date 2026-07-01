package main

import (
	"crypto/sha256"
	"flag"
	"fmt"
	"log"
	"net/url"
	"os"
	"path/filepath"
	"strings"
)

func main() {
	inputFile := flag.String("i", "", "Input Burp Suite XML export file")
	outputDir := flag.String("o", "output", "Output directory for extracted scripts")
	analyze := flag.Bool("analyze", true, "Perform pattern analysis on extracted scripts")
	flag.Parse()

	if *inputFile == "" {
		fmt.Println("Usage: traffic-analyzer -i <burp_export.xml> [-o <output_dir>]")
		flag.PrintDefaults()
		os.Exit(1)
	}

	// Ensure output directory exists
	if err := os.MkdirAll(*outputDir, 0755); err != nil {
		log.Fatalf("Error creating output directory: %v", err)
	}

	// Parse XML
	items, err := ParseBurpXML(*inputFile)
	if err != nil {
		log.Fatalf("Error parsing XML: %v", err)
	}

	fmt.Printf("[*] Parsed %d items from %s\n", len(items.Items), *inputFile)

	patterns := DefaultPatterns()
	allEndpoints := make(map[string]bool)
	processedScripts := make(map[string]bool)

	// Map to store all items by URL for recursive lookup
	urlMap := make(map[string]Item)
	for _, item := range items.Items {
		urlMap[item.URL] = item
	}

	// Queue for processing
	type Task struct {
		URL      string
		MimeType string
		Body     []byte
	}
	var queue []Task

	// Seed queue with all items
	for _, item := range items.Items {
		rawResponse, err := DecodeData(item.Response)
		if err != nil {
			continue
		}
		_, body := SplitHTTPResponse(rawResponse)
		if body == nil {
			continue
		}
		queue = append(queue, Task{URL: item.URL, MimeType: item.MimeType, Body: body})
	}

	// Keep track of what we've processed from the queue
	processedURLs := make(map[string]bool)

	for len(queue) > 0 {
		task := queue[0]
		queue = queue[1:]

		if processedURLs[task.URL] {
			continue
		}
		processedURLs[task.URL] = true

		// Extract Endpoints
		endpoints := ExtractEndpoints(string(task.Body))
		for _, e := range endpoints {
			allEndpoints[e] = true
		}

		// Extract Scripts
		scripts := ExtractJSFromResponse(task.URL, task.MimeType, task.Body)
		for i, script := range scripts {
			// Save script with a name that avoids collisions
			scriptName := generateSafeFilename(script.SourceURL, script.IsInline, i)

			scriptKey := script.SourceURL + script.Content
			if processedScripts[scriptKey] {
				continue
			}
			processedScripts[scriptKey] = true

			// Save script
			savePath := filepath.Join(*outputDir, scriptName)
			if err := os.WriteFile(savePath, []byte(script.Content), 0644); err != nil {
				fmt.Printf("[!] Error saving script %s: %v\n", savePath, err)
			}

			// Analyze script
			if *analyze {
				findings := AnalyzeScript(script, patterns)
				PrintFindings(script.SourceURL, findings)
			}

			// Extract more endpoints from script
			moreEndpoints := ExtractEndpoints(script.Content)
			for _, e := range moreEndpoints {
				allEndpoints[e] = true
			}

			// Recursive discovery: look for imports
			imports := GetJSImports(script.Content)
			for _, imp := range imports {
				absURL := resolveURL(script.SourceURL, imp)
				if item, ok := urlMap[absURL]; ok {
					rawResponse, err := DecodeData(item.Response)
					if err == nil {
						_, body := SplitHTTPResponse(rawResponse)
						if body != nil {
							queue = append(queue, Task{URL: item.URL, MimeType: item.MimeType, Body: body})
						}
					}
				}
			}
		}

		// Look for script references in HTML
		if strings.Contains(strings.ToLower(task.MimeType), "html") {
			refs := GetScriptReferences(task.Body)
			for _, ref := range refs {
				absURL := resolveURL(task.URL, ref)
				if item, ok := urlMap[absURL]; ok {
					rawResponse, err := DecodeData(item.Response)
					if err == nil {
						_, body := SplitHTTPResponse(rawResponse)
						if body != nil {
							queue = append(queue, Task{URL: item.URL, MimeType: item.MimeType, Body: body})
						}
					}
				}
			}
		}
	}

	fmt.Printf("\n[*] Found %d unique endpoints/URLs:\n", len(allEndpoints))
	for e := range allEndpoints {
		fmt.Printf("  - %s\n", e)
	}
}

func generateSafeFilename(sourceURL string, isInline bool, index int) string {
	u, err := url.Parse(sourceURL)
	if err != nil {
		return fmt.Sprintf("script_%x.js", sha256.Sum256([]byte(sourceURL)))
	}

	base := filepath.Base(u.Path)
	if base == "." || base == "/" || isInline {
		base = "index"
	} else {
		base = strings.TrimSuffix(base, filepath.Ext(base))
	}

	hashFull := sha256.Sum256([]byte(sourceURL))
	if isInline {
		return fmt.Sprintf("%s_inline_%d_%x.js", base, index, hashFull[:4])
	}

	// Add a hash of the URL to ensure uniqueness
	hash := fmt.Sprintf("%x", hashFull)[:8]
	return fmt.Sprintf("%s_%s.js", base, hash)
}

func resolveURL(base, ref string) string {
	baseURL, err := url.Parse(base)
	if err != nil {
		return ref
	}
	refURL, err := url.Parse(ref)
	if err != nil {
		return ref
	}
	return baseURL.ResolveReference(refURL).String()
}
