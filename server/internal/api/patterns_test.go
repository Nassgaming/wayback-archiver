package api

import (
	"testing"
)

func TestPatterns(t *testing.T) {
	tests := []struct {
		name    string
		pattern string
		input   string
		want    bool
	}{
		{
			name:    "bodyTagRe matches body tag",
			pattern: "bodyTagRe",
			input:   "<body>",
			want:    true,
		},
		{
			name:    "bodyTagRe matches body with attributes",
			pattern: "bodyTagRe",
			input:   `<body class="main">`,
			want:    true,
		},
		{
			name:    "bodyTagRe case insensitive",
			pattern: "bodyTagRe",
			input:   "<BODY>",
			want:    true,
		},
		{
			name:    "headTagRe matches head tag",
			pattern: "headTagRe",
			input:   "<head>",
			want:    true,
		},
		{
			name:    "htmlTagRe matches html tag",
			pattern: "htmlTagRe",
			input:   "<html>",
			want:    true,
		},
		{
			name:    "srcAttrRe matches src attribute",
			pattern: "srcAttrRe",
			input:   ` src="image.jpg"`,
			want:    true,
		},
		{
			name:    "hrefAttrRe matches href attribute",
			pattern: "hrefAttrRe",
			input:   ` href="style.css"`,
			want:    true,
		},
		{
			name:    "archivePathRe matches archive path",
			pattern: "archivePathRe",
			input:   "/archive/resources/",
			want:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var matches bool
			switch tt.pattern {
			case "bodyTagRe":
				matches = bodyTagRe.MatchString(tt.input)
			case "headTagRe":
				matches = headTagRe.MatchString(tt.input)
			case "htmlTagRe":
				matches = htmlTagRe.MatchString(tt.input)
			case "srcAttrRe":
				matches = srcAttrRe.MatchString(tt.input)
			case "hrefAttrRe":
				matches = hrefAttrRe.MatchString(tt.input)
			case "archivePathRe":
				matches = archivePathRe.MatchString(tt.input)
			}

			if matches != tt.want {
				t.Errorf("Pattern %s: got %v, want %v for input %q", tt.pattern, matches, tt.want, tt.input)
			}
		})
	}
}

func TestArchivePathRewrite(t *testing.T) {
	input := `<img src="/archive/resources/image.jpg">`
	expected := `<img src="/archive/123/20240310150405mp_/resources/image.jpg">`

	result := archivePathRe.ReplaceAllString(input, "/archive/123/20240310150405mp_/resources/")

	if result != expected {
		t.Errorf("archivePathRe replacement failed:\ngot:  %s\nwant: %s", result, expected)
	}
}
