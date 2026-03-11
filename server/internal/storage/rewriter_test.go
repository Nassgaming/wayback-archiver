package storage

import (
	"strings"
	"testing"
)

func newTestRewriter(pageID int64, timestamp string, mappings map[string]string) *URLRewriter {
	r := NewURLRewriter()
	r.SetPageID(pageID)
	r.SetTimestamp(timestamp)
	for url, path := range mappings {
		r.AddMapping(url, path)
	}
	return r
}

func TestRewriteHTML_DataSrcPreserved(t *testing.T) {
	r := newTestRewriter(1, "20260310", map[string]string{
		"https://example.com/real.jpg": "resources/ab/cd/hash.img",
	})

	html := `<img data-src="https://example.com/real.jpg" src="https://example.com/real.jpg">`
	result := r.RewriteHTML(html)

	// data-src should NOT be rewritten
	if !strings.Contains(result, `data-src="https://example.com/real.jpg"`) {
		t.Errorf("data-src should be preserved, got: %s", result)
	}
	// src should be rewritten
	if !strings.Contains(result, `src="/archive/1/20260310mp_/https://example.com/real.jpg"`) {
		t.Errorf("src should be rewritten, got: %s", result)
	}
}

func TestRewriteHTML_PosterAttribute(t *testing.T) {
	r := newTestRewriter(1, "20260310", map[string]string{
		"https://example.com/thumb.jpg": "resources/ab/cd/hash.img",
	})

	html := `<video src="other.mp4" poster="https://example.com/thumb.jpg" autoplay>`
	result := r.RewriteHTML(html)

	if !strings.Contains(result, `poster="/archive/1/20260310mp_/https://example.com/thumb.jpg"`) {
		t.Errorf("poster should be rewritten, got: %s", result)
	}
}

func TestRewriteHTML_HTMLEntityAmp(t *testing.T) {
	r := newTestRewriter(1, "20260310", map[string]string{
		"https://example.com/img.jpg?a=1&b=2": "resources/ab/cd/hash.img",
	})

	// HTML contains &amp; encoded version
	html := `<img src="https://example.com/img.jpg?a=1&amp;b=2">`
	result := r.RewriteHTML(html)

	if !strings.Contains(result, `src="/archive/1/20260310mp_/https://example.com/img.jpg?a=1&b=2"`) {
		t.Errorf("&amp; encoded URL should be rewritten to local path, got: %s", result)
	}
}

func TestRewriteHTML_ProtocolRelativeWithAmp(t *testing.T) {
	r := newTestRewriter(1, "20260310", map[string]string{
		"https://f.video.com/v.mp4?a=1&b=2": "resources/ab/cd/hash.bin",
	})

	// Protocol-relative + &amp; combo
	html := `<video src="//f.video.com/v.mp4?a=1&amp;b=2">`
	result := r.RewriteHTML(html)

	if !strings.Contains(result, `src="/archive/1/20260310mp_/https://f.video.com/v.mp4?a=1&b=2"`) {
		t.Errorf("Protocol-relative + &amp; URL should be rewritten, got: %s", result)
	}
}

func TestRewriteHTML_URLQuotEncoding(t *testing.T) {
	r := newTestRewriter(1, "20260310", map[string]string{
		"https://example.com/bg.jpg": "resources/ab/cd/hash.img",
	})

	html := `<div style="background-image: url(&quot;https://example.com/bg.jpg&quot;);">`
	result := r.RewriteHTML(html)

	if !strings.Contains(result, `url(&quot;/archive/1/20260310mp_/https://example.com/bg.jpg&quot;)`) {
		t.Errorf("url(&quot;...&quot;) should be rewritten, got: %s", result)
	}
}

func TestRewriteHTML_NormalSrcHref(t *testing.T) {
	r := newTestRewriter(1, "20260310", map[string]string{
		"https://example.com/style.css": "resources/ab/cd/hash.css",
		"https://example.com/img.png":   "resources/ef/gh/hash.img",
	})

	html := `<link href="https://example.com/style.css" rel="stylesheet"><img src="https://example.com/img.png">`
	result := r.RewriteHTML(html)

	if !strings.Contains(result, `href="/archive/1/20260310mp_/https://example.com/style.css"`) {
		t.Errorf("href should be rewritten, got: %s", result)
	}
	if !strings.Contains(result, `src="/archive/1/20260310mp_/https://example.com/img.png"`) {
		t.Errorf("src should be rewritten, got: %s", result)
	}
}
