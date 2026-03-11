package api

import (
	"strings"
	"testing"

	"wayback/internal/models"
)

func TestRewriteResourcePathsWithResources(t *testing.T) {
	resources := []models.Resource{
		{
			ID:           1,
			URL:          "https://example.com/style.css",
			ResourceType: "css",
			FilePath:     "resources/ab/cd/hash.css",
		},
		{
			ID:           2,
			URL:          "https://example.com/image.jpg",
			ResourceType: "image",
			FilePath:     "resources/ef/gh/hash.jpg",
		},
	}

	t.Run("rewrites absolute URL in src attribute", func(t *testing.T) {
		html := `<img src="https://example.com/image.jpg">`
		result := rewriteResourcePathsWithResources(html, 123, "20240310150405", resources, "https://example.com")

		if !strings.Contains(result, "/archive/123/20240310150405mp_/https://example.com/image.jpg") {
			t.Errorf("Expected proxy URL in result, got: %s", result)
		}
	})

	t.Run("rewrites absolute URL in href attribute", func(t *testing.T) {
		html := `<link href="https://example.com/style.css">`
		result := rewriteResourcePathsWithResources(html, 123, "20240310150405", resources, "https://example.com")

		if !strings.Contains(result, "/archive/123/20240310150405mp_/https://example.com/style.css") {
			t.Errorf("Expected proxy URL in result, got: %s", result)
		}
	})

	t.Run("rewrites archive/resources path", func(t *testing.T) {
		html := `<img src="/archive/resources/image.jpg">`
		result := rewriteResourcePathsWithResources(html, 123, "20240310150405", resources, "https://example.com")

		if !strings.Contains(result, "/archive/123/20240310150405mp_/resources/image.jpg") {
			t.Errorf("Expected rewritten archive path in result, got: %s", result)
		}
	})

	t.Run("rewrites protocol-relative URL", func(t *testing.T) {
		html := `<img src="//example.com/image.jpg">`
		result := rewriteResourcePathsWithResources(html, 123, "20240310150405", resources, "https://example.com")

		if !strings.Contains(result, "/archive/123/20240310150405mp_/https://example.com/image.jpg") {
			t.Errorf("Expected proxy URL for protocol-relative URL, got: %s", result)
		}
	})

	t.Run("rewrites absolute path", func(t *testing.T) {
		html := `<img src="/image.jpg">`
		result := rewriteResourcePathsWithResources(html, 123, "20240310150405", resources, "https://example.com")

		if !strings.Contains(result, "/archive/123/20240310150405mp_/https://example.com/image.jpg") {
			t.Errorf("Expected proxy URL for absolute path, got: %s", result)
		}
	})

	t.Run("handles empty resources", func(t *testing.T) {
		html := `<img src="https://example.com/image.jpg">`
		result := rewriteResourcePathsWithResources(html, 123, "20240310150405", []models.Resource{}, "https://example.com")

		// Should not modify the HTML when no resources
		if result != html {
			t.Errorf("Expected unchanged HTML with no resources, got: %s", result)
		}
	})
}
