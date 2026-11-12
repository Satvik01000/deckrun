package server_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"deckrun/internal/server"
)

func TestServerEndpoints(t *testing.T) {
	tempDir := t.TempDir()
	testFilePath := filepath.Join(tempDir, "slides.md")
	_ = os.WriteFile(testFilePath, []byte("# Hello\nSlide 1"), 0644)

	watcher := server.NewWatcher(testFilePath)
	srv := server.NewServer(server.Mode{
		Theme:      "nord",
		Template:   "classic",
		Transition: "slide",
		File: &server.EditorFile{
			Name:    "slides.md",
			Kind:    "markdown",
			Path:    testFilePath,
			Watched: true,
		},
		Origin: "http://127.0.0.1:7890",
	}, tempDir, watcher)

	// 1. GET /
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	w := httptest.NewRecorder()
	srv.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Errorf("GET / returned %d", w.Code)
	}
	if !strings.Contains(w.Body.String(), "deckrun · editor") {
		t.Errorf("GET / missing editor title")
	}

	// 2. GET /__preview
	req = httptest.NewRequest(http.MethodGet, "/__preview", nil)
	w = httptest.NewRecorder()
	srv.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Errorf("GET /__preview returned %d", w.Code)
	}
	if !strings.Contains(w.Body.String(), "preview · deckrun") {
		t.Errorf("GET /__preview missing preview title")
	}

	// 3. GET /__vendor/katex.min.css
	req = httptest.NewRequest(http.MethodGet, "/__vendor/katex.min.css", nil)
	w = httptest.NewRecorder()
	srv.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Errorf("GET /__vendor/katex.min.css returned %d", w.Code)
	}
	if w.Header().Get("Cache-Control") == "" {
		t.Errorf("missing Cache-Control header for vendor asset")
	}

	// 4. POST /__parse
	parseBody := []byte("# Slide 1\nHello\n---\n# Slide 2\nWorld")
	req = httptest.NewRequest(http.MethodPost, "/__parse", bytes.NewReader(parseBody))
	w = httptest.NewRecorder()
	srv.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Errorf("POST /__parse returned %d", w.Code)
	}
	var parseResp struct {
		Slides []string `json:"slides"`
		Title  string   `json:"title"`
	}
	if err := json.Unmarshal(w.Body.Bytes(), &parseResp); err != nil {
		t.Fatalf("failed to unmarshal parse response: %v", err)
	}
	if len(parseResp.Slides) != 2 {
		t.Errorf("expected 2 slides in parse response, got %d", len(parseResp.Slides))
	}
	if parseResp.Title != "Slide 1" {
		t.Errorf("expected title 'Slide 1', got '%s'", parseResp.Title)
	}

	// 5. POST /__present
	presentBody, _ := json.Marshal(map[string]any{
		"markdown":   "# My Presentation\nContent",
		"theme":      "tokyo",
		"template":   "editorial",
		"transition": "zoom",
	})
	req = httptest.NewRequest(http.MethodPost, "/__present", bytes.NewReader(presentBody))
	w = httptest.NewRecorder()
	srv.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Errorf("POST /__present returned %d", w.Code)
	}
	var presentResp struct {
		Path string `json:"path"`
	}
	_ = json.Unmarshal(w.Body.Bytes(), &presentResp)
	if !strings.HasPrefix(presentResp.Path, "/?deck=") {
		t.Errorf("expected path to start with /?deck=, got '%s'", presentResp.Path)
	}

	// Fetch stashed deck
	req = httptest.NewRequest(http.MethodGet, presentResp.Path, nil)
	w = httptest.NewRecorder()
	srv.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Errorf("fetching stashed deck %s returned %d", presentResp.Path, w.Code)
	}
	if !strings.Contains(w.Body.String(), `data-theme="tokyo"`) {
		t.Errorf("stashed deck missing tokyo theme")
	}

	// 6. File read & write
	newContent := "# Updated Content\nSave back to disk"
	req = httptest.NewRequest(http.MethodPost, "/__file", strings.NewReader(newContent))
	w = httptest.NewRecorder()
	srv.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Errorf("POST /__file returned %d", w.Code)
	}

	readOnDisk, _ := os.ReadFile(testFilePath)
	if string(readOnDisk) != newContent {
		t.Errorf("file on disk did not match saved content: %s", string(readOnDisk))
	}

	req = httptest.NewRequest(http.MethodGet, "/__file", nil)
	w = httptest.NewRecorder()
	srv.ServeHTTP(w, req)
	if w.Code != http.StatusOK || w.Body.String() != newContent {
		t.Errorf("GET /__file returned %d, content %s", w.Code, w.Body.String())
	}

	// 7. Path traversal forbidden
	req = httptest.NewRequest(http.MethodGet, "/../../../../etc/passwd", nil)
	w = httptest.NewRecorder()
	srv.ServeHTTP(w, req)
	if w.Code != http.StatusForbidden {
		t.Errorf("path traversal returned %d instead of 403", w.Code)
	}
}

func TestFindFreePort(t *testing.T) {
	p := server.FindFreePort(7890)
	if p <= 0 {
		t.Errorf("expected valid port, got %d", p)
	}
}
