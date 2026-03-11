package storage

import (
	"strings"

	"golang.org/x/net/html"
)

// ExtractBodyText 从 HTML 中提取纯文本内容（用于全文搜索）
func ExtractBodyText(htmlContent string) string {
	doc, err := html.Parse(strings.NewReader(htmlContent))
	if err != nil {
		return ""
	}

	var b strings.Builder
	var inBody bool
	var extract func(*html.Node)
	extract = func(n *html.Node) {
		// 跳过 script/style/noscript 标签
		if n.Type == html.ElementNode {
			switch n.Data {
			case "script", "style", "noscript", "svg", "head":
				return
			case "body":
				inBody = true
			}
		}

		if n.Type == html.TextNode && inBody {
			text := strings.TrimSpace(n.Data)
			if text != "" {
				if b.Len() > 0 {
					b.WriteByte(' ')
				}
				b.WriteString(text)
			}
		}

		for c := n.FirstChild; c != nil; c = c.NextSibling {
			extract(c)
		}
	}
	extract(doc)

	return b.String()
}
