package main

import (
	"fmt"
	"log"
	"wayback/internal/database"
)

func main() {
	db, err := database.NewDB()
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	// Check for the 3 specific resources
	urls := []string{
		"https://newsnow.busiyi.world/assets/index-BcW9H-ZC.js",
		"https://newsnow.busiyi.world/assets/index-D0I4jTJu.css",
		"https://newsnow.busiyi.world/Baloo2-Bold.subset.ttf",
	}

	fmt.Println("Checking resources in database:")
	for _, url := range urls {
		resource, err := db.GetResourceByURLAndPageID(url, 99)
		if err != nil {
			fmt.Printf("✗ %s: ERROR - %v\n", url, err)
			continue
		}
		if resource == nil {
			fmt.Printf("✗ %s: NOT FOUND\n", url)
		} else {
			fmt.Printf("✓ %s: FOUND (ID: %d, Type: %s, Path: %s)\n",
				url, resource.ID, resource.ResourceType, resource.FilePath)
		}
	}
}
