package main

import (
	"fmt"
	"net/url"
)

func resolveURL(rawURL, baseURL string) string {
	// 如果已经是完整URL，直接返回
	if rawURL[:4] == "http" {
		return rawURL
	}

	// 解析基础URL
	base, err := url.Parse(baseURL)
	if err != nil {
		fmt.Printf("Failed to parse base URL %s: %v\n", baseURL, err)
		return rawURL
	}

	// 解析相对URL
	ref, err := url.Parse(rawURL)
	if err != nil {
		fmt.Printf("Failed to parse relative URL %s: %v\n", rawURL, err)
		return rawURL
	}

	// 合并URL
	resolved := base.ResolveReference(ref)
	return resolved.String()
}

func main() {
	baseURL := "https://newsnow.busiyi.world/"

	testCases := []string{
		"/icon.svg",
		"/icons/zhihu.png",
		"./style.css",
		"script.js",
		"https://example.com/test.js",
	}

	fmt.Println("Testing URL resolution:")
	for _, testURL := range testCases {
		resolved := resolveURL(testURL, baseURL)
		fmt.Printf("  %s -> %s\n", testURL, resolved)
	}
}
