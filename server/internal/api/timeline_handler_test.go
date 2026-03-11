package api

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"wayback/internal/models"
)

func setupTimelineRouter(handler *Handler) *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	api := r.Group("/api")
	api.POST("/archive", handler.ArchivePage)
	api.GET("/pages/timeline", handler.GetPageTimeline)
	return r
}

func TestGetPageTimeline_MultipleSnapshots(t *testing.T) {
	handler, cleanup := setupTestHandler(t)
	defer cleanup()
	router := setupTimelineRouter(handler)

	testURL := "http://test-timeline-handler.example.com/page"

	// Create first snapshot
	req1 := models.CaptureRequest{
		URL:   testURL,
		Title: "Version 1",
		HTML:  "<html><body>Content Version 1</body></html>",
	}
	body1, _ := json.Marshal(req1)
	w1 := httptest.NewRecorder()
	httpReq1, _ := http.NewRequest("POST", "/api/archive", bytes.NewReader(body1))
	httpReq1.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w1, httpReq1)

	var resp1 models.ArchiveResponse
	json.Unmarshal(w1.Body.Bytes(), &resp1)
	if resp1.PageID <= 0 {
		t.Fatalf("failed to create first snapshot: %s", w1.Body.String())
	}
	defer handler.db.DeletePage(resp1.PageID)

	// Create second snapshot (different content)
	req2 := models.CaptureRequest{
		URL:   testURL,
		Title: "Version 2",
		HTML:  "<html><body>Content Version 2 - Updated</body></html>",
	}
	body2, _ := json.Marshal(req2)
	w2 := httptest.NewRecorder()
	httpReq2, _ := http.NewRequest("POST", "/api/archive", bytes.NewReader(body2))
	httpReq2.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w2, httpReq2)

	var resp2 models.ArchiveResponse
	json.Unmarshal(w2.Body.Bytes(), &resp2)
	if resp2.PageID <= 0 {
		t.Fatalf("failed to create second snapshot: %s", w2.Body.String())
	}
	defer handler.db.DeletePage(resp2.PageID)

	// Query timeline
	w := httptest.NewRecorder()
	httpReq, _ := http.NewRequest("GET", "/api/pages/timeline?url="+testURL, nil)
	router.ServeHTTP(w, httpReq)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
	}

	var result struct {
		URL       string        `json:"url"`
		Snapshots []models.Page `json:"snapshots"`
		Total     int           `json:"total"`
	}
	if err := json.Unmarshal(w.Body.Bytes(), &result); err != nil {
		t.Fatalf("failed to parse response: %v", err)
	}

	if result.URL != testURL {
		t.Errorf("url = %q, want %q", result.URL, testURL)
	}
	if result.Total != 2 {
		t.Errorf("total = %d, want 2", result.Total)
	}
	if len(result.Snapshots) != 2 {
		t.Fatalf("snapshots count = %d, want 2", len(result.Snapshots))
	}

	// Newest first
	if result.Snapshots[0].Title != "Version 2" {
		t.Errorf("first snapshot title = %q, want %q", result.Snapshots[0].Title, "Version 2")
	}
	if result.Snapshots[1].Title != "Version 1" {
		t.Errorf("second snapshot title = %q, want %q", result.Snapshots[1].Title, "Version 1")
	}
}

func TestGetPageTimeline_MissingURL(t *testing.T) {
	handler, cleanup := setupTestHandler(t)
	defer cleanup()
	router := setupTimelineRouter(handler)

	w := httptest.NewRecorder()
	httpReq, _ := http.NewRequest("GET", "/api/pages/timeline", nil)
	router.ServeHTTP(w, httpReq)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400 for missing url, got %d", w.Code)
	}
}

func TestGetPageTimeline_NoSnapshots(t *testing.T) {
	handler, cleanup := setupTestHandler(t)
	defer cleanup()
	router := setupTimelineRouter(handler)

	w := httptest.NewRecorder()
	httpReq, _ := http.NewRequest("GET", "/api/pages/timeline?url=http://nonexistent-timeline-test.example.com", nil)
	router.ServeHTTP(w, httpReq)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}

	var result struct {
		Total int `json:"total"`
	}
	json.Unmarshal(w.Body.Bytes(), &result)
	if result.Total != 0 {
		t.Errorf("total = %d, want 0", result.Total)
	}
}
