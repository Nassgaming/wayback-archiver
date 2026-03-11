package database

import (
	"fmt"
	"testing"
	"time"
)

// skipIfNoDB connects to the test database or skips the test.
func skipIfNoDB(t *testing.T) *DB {
	t.Helper()
	db, err := New("localhost", "5432", "postgres", "", "wayback")
	if err != nil {
		t.Skipf("Skipping DB test (cannot connect): %v", err)
	}
	return db
}

func idStr(id int64) string {
	return fmt.Sprintf("%d", id)
}

func TestUpdatePageContent(t *testing.T) {
	db := skipIfNoDB(t)
	defer db.Close()

	now := time.Now()
	origHash := "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"
	newHash := "bbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb"

	pageID, err := db.CreatePage("http://test-update-content.example.com", "Original Title", "html/test/original.html", origHash, now)
	if err != nil {
		t.Fatalf("CreatePage failed: %v", err)
	}
	defer db.DeletePage(pageID)

	// Update content
	err = db.UpdatePageContent(pageID, "html/test/updated.html", newHash, "Updated Title")
	if err != nil {
		t.Fatalf("UpdatePageContent failed: %v", err)
	}

	// Verify
	page, err := db.GetPageByID(idStr(pageID))
	if err != nil {
		t.Fatalf("GetPageByID failed: %v", err)
	}
	if page == nil {
		t.Fatal("page should exist after update")
	}
	if page.HTMLPath != "html/test/updated.html" {
		t.Errorf("HTMLPath = %q, want %q", page.HTMLPath, "html/test/updated.html")
	}
	if page.ContentHash != newHash {
		t.Errorf("ContentHash = %q, want %q", page.ContentHash, newHash)
	}
	if page.Title != "Updated Title" {
		t.Errorf("Title = %q, want %q", page.Title, "Updated Title")
	}
	if !page.LastVisited.After(now) {
		t.Errorf("LastVisited should be updated to after creation time")
	}
}

func TestDeletePageResources(t *testing.T) {
	db := skipIfNoDB(t)
	defer db.Close()

	now := time.Now()
	pageID, err := db.CreatePage("http://test-delete-resources.example.com", "Test", "html/test/del.html", "cccccccccccccccccccccccccccccccccccccccccccccccccccccccccccccccc", now)
	if err != nil {
		t.Fatalf("CreatePage failed: %v", err)
	}
	defer db.DeletePage(pageID)

	// Create a resource and link it
	resID, err := db.CreateResource("http://test-delete-resources.example.com/style.css", "dddddddddddddddddddddddddddddddddddddddddddddddddddddddddddddddd", "css", "resources/dd/dd/dddd.css", 100)
	if err != nil {
		t.Fatalf("CreateResource failed: %v", err)
	}

	err = db.LinkPageResource(pageID, resID)
	if err != nil {
		t.Fatalf("LinkPageResource failed: %v", err)
	}

	// Verify link exists via count
	var count int
	err = db.conn.QueryRow("SELECT COUNT(*) FROM page_resources WHERE page_id = $1", pageID).Scan(&count)
	if err != nil {
		t.Fatalf("count query failed: %v", err)
	}
	if count != 1 {
		t.Fatalf("expected 1 linked resource, got %d", count)
	}

	// Delete page resources
	err = db.DeletePageResources(pageID)
	if err != nil {
		t.Fatalf("DeletePageResources failed: %v", err)
	}

	// Verify links are gone
	err = db.conn.QueryRow("SELECT COUNT(*) FROM page_resources WHERE page_id = $1", pageID).Scan(&count)
	if err != nil {
		t.Fatalf("count query after delete failed: %v", err)
	}
	if count != 0 {
		t.Errorf("expected 0 linked resources after delete, got %d", count)
	}

	// Verify the resource record itself still exists
	res, err := db.GetResourceByID(resID)
	if err != nil {
		t.Fatalf("GetResourceByID failed: %v", err)
	}
	if res == nil {
		t.Error("resource record should still exist after DeletePageResources")
	}

	// Cleanup resource
	db.conn.Exec("DELETE FROM resources WHERE id = $1", resID)
}

func TestUpdatePageContent_NonExistentPage(t *testing.T) {
	db := skipIfNoDB(t)
	defer db.Close()

	// UPDATE on non-existent row affects 0 rows — should not error
	err := db.UpdatePageContent(999999999, "html/test/nope.html", "zzzz", "Nope")
	if err != nil {
		t.Fatalf("UpdatePageContent on non-existent page should not error, got: %v", err)
	}
}

func TestDeletePageResources_NoLinks(t *testing.T) {
	db := skipIfNoDB(t)
	defer db.Close()

	err := db.DeletePageResources(999999999)
	if err != nil {
		t.Fatalf("DeletePageResources on page with no links should not error, got: %v", err)
	}
}
