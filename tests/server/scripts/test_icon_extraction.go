package main

import (
	"fmt"
	"log"
	"os"
	"wayback/internal/storage"
)

func main() {
	// 读取测试HTML
	htmlContent, err := os.ReadFile("test/test_icons.html")
	if err != nil {
		log.Fatal(err)
	}

	// 创建提取器
	extractor := storage.NewHTMLResourceExtractor()

	// 提取资源
	pageURL := "https://example.com/"
	resources := extractor.ExtractResources(string(htmlContent), pageURL)

	fmt.Printf("Total resources extracted: %d\n\n", len(resources))
	fmt.Println("Extracted resources:")
	for i, res := range resources {
		fmt.Printf("  %d. %s (type: %s)\n", i+1, res.URL, res.Type)
	}
}
