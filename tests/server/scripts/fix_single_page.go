package main

import (
	"flag"
	"log"
	"os"
	"path/filepath"
	"strings"

	"wayback/internal/database"
	"wayback/internal/models"
	"wayback/internal/storage"
)

func main() {
	pageID := flag.Int("page", 0, "Page ID to fix")
	flag.Parse()

	if *pageID == 0 {
		log.Fatal("Please specify a page ID with -page flag")
	}

	// 连接数据库
	db, err := database.New("localhost", "5432", "postgres", "", "wayback")
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	// 获取所有页面并找到指定的页面
	pages, err := db.ListPages(10000, 0)
	if err != nil {
		log.Fatalf("Failed to list pages: %v", err)
	}

	var page *models.Page
	for i := range pages {
		if pages[i].ID == int64(*pageID) {
			page = &pages[i]
			break
		}
	}

	if page == nil {
		log.Fatalf("Page %d not found", *pageID)
	}

	log.Printf("Processing page %d: %s", page.ID, page.URL)

	dataDir := "data"
	fileStorage := storage.NewFileStorage(dataDir)
	dedup := storage.NewDeduplicator(db, fileStorage)

	// 读取 HTML 文件
	htmlPath := filepath.Join(dataDir, page.HTMLPath)
	htmlContent, err := os.ReadFile(htmlPath)
	if err != nil {
		log.Fatalf("Failed to read HTML: %v", err)
	}

	html := string(htmlContent)

	// 提取 HTML 中的外部资源
	extractor := storage.NewHTMLResourceExtractor()
	externalURLs := extractor.ExtractResources(html, page.URL)
	log.Printf("  Found %d external URLs in HTML", len(externalURLs))

	// 获取当前页面已有的资源
	existingResources, err := db.GetResourcesByPageID(page.ID)
	if err != nil {
		log.Printf("  Failed to get existing resources: %v", err)
		return
	}

	existingURLs := make(map[string]bool)
	for _, res := range existingResources {
		existingURLs[res.URL] = true
	}

	// 找出缺失的资源
	var missingResources []storage.ResourceRef
	for _, res := range externalURLs {
		if !existingURLs[res.URL] {
			missingResources = append(missingResources, res)
		}
	}

	if len(missingResources) == 0 {
		log.Printf("  No missing resources found")
		return
	}

	log.Printf("  Found %d missing resources, downloading...", len(missingResources))

	// 下载并保存缺失的资源
	succeeded := 0
	failed := 0

	for _, res := range missingResources {
		// 使用 deduplicator 下载并保存资源
		resourceID, _, err := dedup.ProcessResource(res.URL, res.Type, "")
		if err != nil {
			log.Printf("    ✗ %s: %v", res.URL, err)
			failed++
			continue
		}

		// 关联到页面
		if err := db.LinkPageResource(page.ID, resourceID); err != nil {
			log.Printf("    ✗ Failed to link %s: %v", res.URL, err)
			failed++
			continue
		}

		succeeded++
	}

	log.Printf("  Result: %d succeeded, %d failed", succeeded, failed)
}

func guessResourceType(url string) string {
	url = strings.ToLower(url)
	if strings.Contains(url, ".js") {
		return "js"
	} else if strings.Contains(url, ".css") {
		return "css"
	} else if strings.Contains(url, ".png") || strings.Contains(url, ".jpg") ||
		strings.Contains(url, ".jpeg") || strings.Contains(url, ".gif") ||
		strings.Contains(url, ".svg") || strings.Contains(url, ".webp") {
		return "image"
	} else if strings.Contains(url, ".woff") || strings.Contains(url, ".ttf") ||
		strings.Contains(url, ".otf") || strings.Contains(url, ".eot") {
		return "font"
	}
	return "other"
}
