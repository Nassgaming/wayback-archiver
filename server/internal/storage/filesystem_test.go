package storage

import (
	"os"
	"path/filepath"
	"testing"
)

func TestDeleteHTML(t *testing.T) {
	// Create a temp directory as base
	baseDir := t.TempDir()
	fs := &FileStorage{baseDir: baseDir}

	// Create a file to delete
	relPath := filepath.Join("html", "test", "page.html")
	fullPath := filepath.Join(baseDir, relPath)
	if err := os.MkdirAll(filepath.Dir(fullPath), 0755); err != nil {
		t.Fatalf("MkdirAll failed: %v", err)
	}
	if err := os.WriteFile(fullPath, []byte("<html>test</html>"), 0644); err != nil {
		t.Fatalf("WriteFile failed: %v", err)
	}

	// Verify file exists
	if _, err := os.Stat(fullPath); err != nil {
		t.Fatalf("file should exist before delete: %v", err)
	}

	// Delete
	err := fs.DeleteHTML(relPath)
	if err != nil {
		t.Fatalf("DeleteHTML failed: %v", err)
	}

	// Verify file is gone
	if _, err := os.Stat(fullPath); !os.IsNotExist(err) {
		t.Errorf("file should not exist after delete, got err: %v", err)
	}
}

func TestDeleteHTML_NonExistent(t *testing.T) {
	baseDir := t.TempDir()
	fs := &FileStorage{baseDir: baseDir}

	// Deleting a non-existent file should not error
	err := fs.DeleteHTML("html/does/not/exist.html")
	if err != nil {
		t.Fatalf("DeleteHTML on non-existent file should not error, got: %v", err)
	}
}
