package main

import (
	"testing"
)

func TestParseBurpXML(t *testing.T) {
	items, err := ParseBurpXML("../../sample_burp.xml")
	if err != nil {
		t.Fatalf("Failed to parse XML: %v", err)
	}

	if len(items.Items) != 2 {
		t.Errorf("Expected 2 items, got %d", len(items.Items))
	}

	if items.Items[0].URL != "https://example.com/index.html" {
		t.Errorf("Unexpected URL: %s", items.Items[0].URL)
	}
}

func TestDecodeData(t *testing.T) {
	data := Data{Base64: true, Value: "R0VUIC8gSFRUUC8xLjENCkhvc3Q6IGV4YW1wbGUuY29tDQoNCg=="}
	decoded, err := DecodeData(data)
	if err != nil {
		t.Fatalf("Failed to decode data: %v", err)
	}

	expected := "GET / HTTP/1.1\r\nHost: example.com\r\n\r\n"
	if string(decoded) != expected {
		t.Errorf("Expected %q, got %q", expected, string(decoded))
	}
}

func TestExtractJSFromResponse(t *testing.T) {
	html := []byte(`<html><script>console.log("hello")</script></html>`)
	scripts := ExtractJSFromResponse("http://test.com", "text/html", html)

	if len(scripts) != 1 {
		t.Fatalf("Expected 1 script, got %d", len(scripts))
	}

	if scripts[0].Content != `console.log("hello")` {
		t.Errorf("Unexpected script content: %s", scripts[0].Content)
	}
}

func TestGetScriptReferences(t *testing.T) {
	html := []byte(`<html><script src="/js/app.js"></script><script src="https://cdn.com/lib.js"></script></html>`)
	refs := GetScriptReferences(html)

	if len(refs) != 2 {
		t.Fatalf("Expected 2 refs, got %d", len(refs))
	}

	if refs[0] != "/js/app.js" || refs[1] != "https://cdn.com/lib.js" {
		t.Errorf("Unexpected refs: %v", refs)
	}
}

func TestExtractEndpoints(t *testing.T) {
	content := `fetch("https://api.example.com/v1/data"); var path = "/internal/config";`
	endpoints := ExtractEndpoints(content)

	foundURL := false
	foundPath := false
	for _, e := range endpoints {
		if e == "https://api.example.com/v1/data" {
			foundURL = true
		}
		if e == "/internal/config" {
			foundPath = true
		}
	}

	if !foundURL {
		t.Error("Failed to find URL endpoint")
	}
	if !foundPath {
		t.Error("Failed to find path endpoint")
	}
}

func TestAnalyzeScript(t *testing.T) {
	script := ExtractedScript{
		Content: "var x = eval('alert(1)'); document.innerHTML = x;",
	}
	patterns := DefaultPatterns()
	findings := AnalyzeScript(script, patterns)

	if len(findings) < 2 {
		t.Errorf("Expected at least 2 findings, got %d", len(findings))
	}
}

func TestResolveURL(t *testing.T) {
	base := "https://example.com/pages/index.html"
	tests := []struct {
		ref      string
		expected string
	}{
		{"/js/app.js", "https://example.com/js/app.js"},
		{"app.js", "https://example.com/pages/app.js"},
		{"https://cdn.com/lib.js", "https://cdn.com/lib.js"},
	}

	for _, tt := range tests {
		res := resolveURL(base, tt.ref)
		if res != tt.expected {
			t.Errorf("resolveURL(%q, %q) = %q, expected %q", base, tt.ref, res, tt.expected)
		}
	}
}
