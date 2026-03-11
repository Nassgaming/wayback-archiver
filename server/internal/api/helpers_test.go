package api

import (
	"testing"
)

func TestEscapeHTML(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string
	}{
		{
			name:  "escape ampersand",
			input: "Tom & Jerry",
			want:  "Tom &amp; Jerry",
		},
		{
			name:  "escape less than",
			input: "1 < 2",
			want:  "1 &lt; 2",
		},
		{
			name:  "escape greater than",
			input: "2 > 1",
			want:  "2 &gt; 1",
		},
		{
			name:  "escape double quote",
			input: `He said "hello"`,
			want:  "He said &quot;hello&quot;",
		},
		{
			name:  "escape single quote",
			input: "It's working",
			want:  "It&#39;s working",
		},
		{
			name:  "escape multiple characters",
			input: `<script>alert("XSS & injection")</script>`,
			want:  "&lt;script&gt;alert(&quot;XSS &amp; injection&quot;)&lt;/script&gt;",
		},
		{
			name:  "no escaping needed",
			input: "Hello World",
			want:  "Hello World",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := escapeHTML(tt.input)
			if got != tt.want {
				t.Errorf("escapeHTML() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestDetectFontType(t *testing.T) {
	tests := []struct {
		name     string
		filePath string
		want     string
	}{
		{
			name:     "woff font",
			filePath: "/path/to/font.woff",
			want:     "font/woff",
		},
		{
			name:     "woff2 font",
			filePath: "/path/to/font.woff2",
			want:     "font/woff2",
		},
		{
			name:     "ttf font",
			filePath: "/path/to/font.ttf",
			want:     "font/ttf",
		},
		{
			name:     "otf font",
			filePath: "/path/to/font.otf",
			want:     "font/otf",
		},
		{
			name:     "eot font",
			filePath: "/path/to/font.eot",
			want:     "application/vnd.ms-fontobject",
		},
		{
			name:     "unknown extension defaults to woff2",
			filePath: "/path/to/font.unknown",
			want:     "font/woff2",
		},
		{
			name:     "uppercase extension",
			filePath: "/path/to/font.WOFF",
			want:     "font/woff",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := detectFontType(tt.filePath)
			if got != tt.want {
				t.Errorf("detectFontType() = %q, want %q", got, tt.want)
			}
		})
	}
}
