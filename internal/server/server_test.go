package server_test

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/url"
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

func TestLocalMediaUploadListAndServe(t *testing.T) {
	baseDir := t.TempDir()
	outsideDir := t.TempDir()
	outsidePNG := filepath.Join(outsideDir, "chart screenshot.png")
	_ = os.WriteFile(outsidePNG, []byte("PNGDATA"), 0644)
	_ = os.WriteFile(filepath.Join(baseDir, "secret.env"), []byte("SECRET=1"), 0644)

	srv := server.NewServer(server.Mode{
		Theme:      "nord",
		Template:   "classic",
		Transition: "slide",
		Origin:     "http://127.0.0.1:7890",
	}, baseDir, nil)

	// 1. GET /__local serves an allowed external media file
	req := httptest.NewRequest(http.MethodGet, "/__local?path="+url.QueryEscape(outsidePNG), nil)
	w := httptest.NewRecorder()
	srv.ServeHTTP(w, req)
	if w.Code != http.StatusOK || w.Body.String() != "PNGDATA" {
		t.Errorf("GET /__local external png returned %d (%q)", w.Code, w.Body.String())
	}

	// 1b. GET /__local resolves extensionless macOS screenshot with \u202f narrow no-break space
	macShotName := "Screenshot 2026-11-14 at 11.31.58\u202fAM.png"
	_ = os.WriteFile(filepath.Join(outsideDir, macShotName), []byte("MACSHOT"), 0644)
	macShotQuery := filepath.Join(outsideDir, "Screenshot 2026-11-14 at 11.31.58 AM")
	req = httptest.NewRequest(http.MethodGet, "/__local?path="+url.QueryEscape(macShotQuery), nil)
	w = httptest.NewRecorder()
	srv.ServeHTTP(w, req)
	if w.Code != http.StatusOK || w.Body.String() != "MACSHOT" || !strings.HasPrefix(w.Header().Get("Content-Type"), "image/png") {
		t.Errorf("GET /__local macOS extensionless screenshot returned %d CT=%q body=%q", w.Code, w.Header().Get("Content-Type"), w.Body.String())
	}

	// 2. GET /__local rejects non-media extensions with 403
	req = httptest.NewRequest(http.MethodGet, "/__local?path="+url.QueryEscape(filepath.Join(baseDir, "secret.env")), nil)
	w = httptest.NewRecorder()
	srv.ServeHTTP(w, req)
	if w.Code != http.StatusForbidden {
		t.Errorf("expected 403 for non-media file, got %d", w.Code)
	}

	// 3. POST /__upload-asset writes image to baseDir/assets/
	uploadBody, _ := json.Marshal(map[string]string{
		"name": "My Screenshot.png",
		"data": "data:image/png;base64," + base64.StdEncoding.EncodeToString([]byte("UPLOADED")),
	})
	req = httptest.NewRequest(http.MethodPost, "/__upload-asset", bytes.NewReader(uploadBody))
	w = httptest.NewRecorder()
	srv.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("POST /__upload-asset failed with %d: %s", w.Code, w.Body.String())
	}
	var upResp struct {
		Path string `json:"path"`
		URL  string `json:"url"`
	}
	if err := json.Unmarshal(w.Body.Bytes(), &upResp); err != nil {
		t.Fatalf("unmarshal upload resp: %v", err)
	}
	if upResp.Path != "assets/my-screenshot.png" {
		t.Errorf("expected assets/my-screenshot.png, got %q", upResp.Path)
	}
	diskData, err := os.ReadFile(filepath.Join(baseDir, upResp.Path))
	if err != nil || string(diskData) != "UPLOADED" {
		t.Errorf("uploaded file not written to disk cleanly: %v, %q", err, string(diskData))
	}

	// 4. GET /__assets lists the uploaded image
	req = httptest.NewRequest(http.MethodGet, "/__assets", nil)
	w = httptest.NewRecorder()
	srv.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("GET /__assets failed with %d", w.Code)
	}
	var listResp struct {
		Files []string `json:"files"`
	}
	if err := json.Unmarshal(w.Body.Bytes(), &listResp); err != nil {
		t.Fatalf("unmarshal assets resp: %v", err)
	}
	if len(listResp.Files) != 1 || listResp.Files[0] != "assets/my-screenshot.png" {
		t.Errorf("expected [assets/my-screenshot.png], got %+v", listResp.Files)
	}
}
