package main

import (
	"bytes"
	"regexp"
	"strings"

	"golang.org/x/net/html"
)

// ExtractedScript represents a JavaScript snippet or file found in the traffic.
type ExtractedScript struct {
	SourceURL string
	Content   string
	IsInline  bool
}

// ExtractJSFromResponse parses an HTTP response and extracts JavaScript content.
func ExtractJSFromResponse(url string, mimeType string, responseBody []byte) []ExtractedScript {
	var scripts []ExtractedScript

	// If the MIME type is script, the whole body is JS.
	if strings.Contains(strings.ToLower(mimeType), "script") || strings.Contains(strings.ToLower(mimeType), "javascript") {
		scripts = append(scripts, ExtractedScript{
			SourceURL: url,
			Content:   string(responseBody),
			IsInline:  false,
		})
		return scripts
	}

	// If it's HTML, use a proper parser to look for <script> tags.
	if strings.Contains(strings.ToLower(mimeType), "html") {
		doc, err := html.Parse(bytes.NewReader(responseBody))
		if err != nil {
			return scripts
		}

		var f func(*html.Node)
		f = func(n *html.Node) {
			if n.Type == html.ElementNode && n.Data == "script" {
				for _, a := range n.Attr {
					if a.Key == "src" {
						// This is an external script reference.
						// We'll handle pulling these in the main loop if they are present in Burp items.
						break
					}
				}
				// Inline script content
				if n.FirstChild != nil && n.FirstChild.Type == html.TextNode {
					scripts = append(scripts, ExtractedScript{
						SourceURL: url,
						Content:   n.FirstChild.Data,
						IsInline:  true,
					})
				}
			}
			for c := n.FirstChild; c != nil; c = c.NextSibling {
				f(c)
			}
		}
		f(doc)
	}

	return scripts
}

// GetScriptReferences extracts external script URLs from HTML content.
func GetScriptReferences(htmlContent []byte) []string {
	var refs []string
	doc, err := html.Parse(bytes.NewReader(htmlContent))
	if err != nil {
		return refs
	}

	var f func(*html.Node)
	f = func(n *html.Node) {
		if n.Type == html.ElementNode && n.Data == "script" {
			for _, a := range n.Attr {
				if a.Key == "src" && a.Val != "" {
					refs = append(refs, a.Val)
				}
			}
		}
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			f(c)
		}
	}
	f(doc)
	return refs
}

// GetJSImports extracts import references from JavaScript content.
func GetJSImports(jsContent string) []string {
	var refs []string
	// Match both 'import ... from "..."' and 'import "..."'
	importRegex := regexp.MustCompile(`(?m)import\s+(?:(?:\w+|{[^{}]+})\s+from\s+)?["'](.*?)["']`)
	matches := importRegex.FindAllStringSubmatch(jsContent, -1)
	for _, match := range matches {
		if len(match) > 1 {
			refs = append(refs, match[1])
		}
	}
	return refs
}

// ExtractEndpoints finds URLs and API-like paths in text content.
func ExtractEndpoints(content string) []string {
	endpoints := make(map[string]bool)

	// Regex for URLs
	urlRegex := regexp.MustCompile(`https?://[a-zA-Z0-9\-\.]+(?::\d+)?(?:/[a-zA-Z0-9\-\._\?\,\'/\\\+&amp;%\$#\=~]*)?`)
	for _, match := range urlRegex.FindAllString(content, -1) {
		endpoints[match] = true
	}

	// Regex for potential API paths (starting with /api/, /v1/, etc.)
	pathRegex := regexp.MustCompile(`(?:"|')(/[a-zA-Z0-9\-\._\?\,\'/\\\+&amp;%\$#\=~]+)(?:"|')`)
	for _, match := range pathRegex.FindAllStringSubmatch(content, -1) {
		if len(match) > 1 {
			path := match[1]
			// Filter out common false positives like file extensions we don't want or very short strings
			if len(path) > 2 && !strings.HasSuffix(path, ".css") && !strings.HasSuffix(path, ".png") && !strings.HasSuffix(path, ".jpg") {
				endpoints[path] = true
			}
		}
	}

	var result []string
	for k := range endpoints {
		result = append(result, k)
	}
	return result
}

// SplitHTTPResponse splits a raw HTTP response into headers and body.
func SplitHTTPResponse(raw []byte) ([]byte, []byte) {
	parts := bytes.SplitN(raw, []byte("\r\n\r\n"), 2)
	if len(parts) < 2 {
		return raw, nil
	}
	return parts[0], parts[1]
}
