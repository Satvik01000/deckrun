package server

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"mime"
	"net"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"sync"
	"time"

	"deckrun/internal/assets"
	"deckrun/internal/parser"
	"deckrun/internal/pdf"
	"deckrun/internal/render"
	"deckrun/internal/theme"
)

const maxBody = 32 * 1024 * 1024

type EditorFile struct {
	Name    string
	Kind    string // "markdown" | "html"
	Path    string // on disk
	Content string // in memory (if from URL)
	Watched bool
}

type Mode struct {
	Theme      string
	Fonts      render.FontOverrides
	Template   string
	Transition string
	Fullscreen bool
	File       *EditorFile
	Origin     string
}

type Server struct {
	Mode    Mode
	BaseDir string
	Watcher *Watcher

	decksMu sync.Mutex
	decks   map[int]string
	deckSeq int
}

func NewServer(mode Mode, baseDir string, watcher *Watcher) *Server {
	return &Server{
		Mode:    mode,
		BaseDir: baseDir,
		Watcher: watcher,
		decks:   make(map[int]string),
	}
}

func (s *Server) StashDeck(htmlStr string) string {
	s.decksMu.Lock()
	defer s.decksMu.Unlock()

	s.deckSeq++
	id := s.deckSeq
	s.decks[id] = htmlStr

	// Keep only the most recent 8 builds
	for len(s.decks) > 8 {
		minKey := -1
		for k := range s.decks {
			if minKey == -1 || k < minKey {
				minKey = k
			}
		}
		if minKey != -1 {
			delete(s.decks, minKey)
		}
	}

	return fmt.Sprintf("/?deck=%d", id)
}

func (s *Server) GetStashedDeck(id int) (string, bool) {
	s.decksMu.Lock()
	defer s.decksMu.Unlock()
	h, ok := s.decks[id]
	return h, ok
}

func readBody(r *http.Request) ([]byte, error) {
	limited := io.LimitReader(r.Body, maxBody+1)
	data, err := io.ReadAll(limited)
	if err != nil {
		return nil, err
	}
	if len(data) > maxBody {
		return nil, errors.New("request body too large")
	}
	return data, nil
}

func sendJSON(w http.ResponseWriter, payload any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.Header().Set("Cache-Control", "no-store")
	_ = json.NewEncoder(w).Encode(payload)
}

func sendHTML(w http.ResponseWriter, htmlStr string) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.Header().Set("Cache-Control", "no-store")
	_, _ = w.Write([]byte(htmlStr))
}

var safeFilenameRegex = regexp.MustCompile(`[^a-z0-9._-]+`)

func safeFilename(name string) string {
	slug := strings.TrimSpace(strings.ToLower(name))
	slug = strings.TrimSuffix(slug, ".md")
	slug = strings.TrimSuffix(slug, ".markdown")
	slug = safeFilenameRegex.ReplaceAllString(slug, "-")
	slug = strings.Trim(slug, "-")
	if slug == "" {
		return "deck"
	}
	return slug
}

func getMime(path string) string {
	ext := strings.ToLower(filepath.Ext(path))
	switch ext {
	case ".html", ".htm":
		return "text/html; charset=utf-8"
	case ".css":
		return "text/css"
	case ".js", ".mjs":
		return "application/javascript"
	case ".json":
		return "application/json"
	case ".md":
		return "text/markdown; charset=utf-8"
	case ".txt":
		return "text/plain; charset=utf-8"
	case ".png":
		return "image/png"
	case ".jpg", ".jpeg":
		return "image/jpeg"
	case ".gif":
		return "image/gif"
	case ".svg":
		return "image/svg+xml"
	case ".webp":
		return "image/webp"
	case ".ico":
		return "image/x-icon"
	case ".avif":
		return "image/avif"
	case ".bmp":
		return "image/bmp"
	case ".mp4":
		return "video/mp4"
	case ".webm":
		return "video/webm"
	case ".mov":
		return "video/quicktime"
	case ".ogg":
		return "video/ogg"
	case ".woff":
		return "font/woff"
	case ".woff2":
		return "font/woff2"
	case ".ttf":
		return "font/ttf"
	default:
		c := mime.TypeByExtension(ext)
		if c != "" {
			return c
		}
		return "application/octet-stream"
	}
}

var allowedMediaExts = map[string]bool{
	".png":  true,
	".jpg":  true,
	".jpeg": true,
	".gif":  true,
	".svg":  true,
	".webp": true,
	".avif": true,
	".ico":  true,
	".bmp":  true,
	".mp4":  true,
	".webm": true,
	".mov":  true,
	".ogg":  true,
}

var mediaExtCandidates = []string{
	".png", ".jpg", ".jpeg", ".webp", ".gif", ".svg", ".avif", ".bmp", ".mp4", ".webm", ".mov",
}

func normMediaName(s string) string {
	r := strings.NewReplacer(
		"\u202f", " ",
		"\u00a0", " ",
		"\u2009", " ",
		"\u2007", " ",
	)
	return strings.ToLower(r.Replace(s))
}

func isExplicitNonMediaExt(ext string) bool {
	if ext == "" || strings.Contains(ext, " ") || len(ext) < 2 {
		return false
	}
	first := ext[1]
	return !(first >= '0' && first <= '9')
}

// ResolveLocalMediaPath resolves a local media path (supporting @local:, file://,
// ~/, relative paths, and absolute paths), probes common media extensions when omitted,
// handles macOS Unicode space normalization, and restricts access to allowed media extensions.
func ResolveLocalMediaPath(baseDir, rawPath string) (resolved string, allowed bool, err error) {
	p := strings.TrimSpace(rawPath)
	lower := strings.ToLower(p)
	switch {
	case strings.HasPrefix(lower, "@local:"):
		p = strings.TrimSpace(p[len("@local:"):])
	case strings.HasPrefix(lower, "local://"):
		p = strings.TrimSpace(p[len("local://"):])
	case strings.HasPrefix(lower, "file://"):
		p = strings.TrimSpace(p[len("file://"):])
	}
	if p == "" {
		return "", false, nil
	}

	if strings.HasPrefix(p, "~/") || p == "~" {
		home, herr := os.UserHomeDir()
		if herr != nil {
			return "", false, nil
		}
		if p == "~" {
			p = home
		} else {
			p = filepath.Join(home, p[2:])
		}
	} else if !filepath.IsAbs(p) {
		p = filepath.Join(baseDir, p)
	}

	p = filepath.Clean(p)
	if abs, aerr := filepath.Abs(p); aerr == nil {
		p = abs
	}

	ext := strings.ToLower(filepath.Ext(p))
	dir := filepath.Dir(p)
	wantBase := normMediaName(filepath.Base(p))

	// Case A: explicit allowed media extension
	if allowedMediaExts[ext] {
		if info, statErr := os.Stat(p); statErr == nil && !info.IsDir() {
			return p, true, nil
		}
		if entries, rerr := os.ReadDir(dir); rerr == nil {
			for _, entry := range entries {
				if entry.IsDir() {
					continue
				}
				if normMediaName(entry.Name()) == wantBase {
					return filepath.Join(dir, entry.Name()), true, nil
				}
			}
		}
		return p, true, os.ErrNotExist
	}

	// Case B: missing or unrecognized extension (e.g. Finder-hidden .png or timestamp dots like ".58 AM")
	for _, candExt := range mediaExtCandidates {
		cand := p + candExt
		if info, statErr := os.Stat(cand); statErr == nil && !info.IsDir() {
			return cand, true, nil
		}
	}

	if entries, rerr := os.ReadDir(dir); rerr == nil {
		byNorm := make(map[string]string, len(entries))
		for _, entry := range entries {
			if entry.IsDir() {
				continue
			}
			entryExt := strings.ToLower(filepath.Ext(entry.Name()))
			if !allowedMediaExts[entryExt] {
				continue
			}
			normEntry := normMediaName(entry.Name())
			if _, exists := byNorm[normEntry]; !exists {
				byNorm[normEntry] = entry.Name()
			}
		}
		if name, ok := byNorm[wantBase]; ok {
			return filepath.Join(dir, name), true, nil
		}
		for _, candExt := range mediaExtCandidates {
			if name, ok := byNorm[wantBase+candExt]; ok {
				return filepath.Join(dir, name), true, nil
			}
		}
	}

	if info, statErr := os.Stat(p); statErr == nil && !info.IsDir() {
		return "", false, nil
	}
	if isExplicitNonMediaExt(ext) {
		return "", false, nil
	}
	return p, true, os.ErrNotExist
}

func safeAssetName(rawName string) (string, string, bool) {
	base := filepath.Base(strings.TrimSpace(rawName))
	if base == "" || base == "." || base == "/" {
		base = "image.png"
	}
	ext := strings.ToLower(filepath.Ext(base))
	stem := strings.TrimSuffix(base, filepath.Ext(base))
	if ext == "" {
		ext = ".png"
	}
	if !allowedMediaExts[ext] {
		return "", "", false
	}
	cleanStem := strings.ToLower(strings.TrimSpace(stem))
	cleanStem = safeFilenameRegex.ReplaceAllString(cleanStem, "-")
	cleanStem = strings.Trim(cleanStem, "-.")
	if cleanStem == "" {
		cleanStem = "image"
	}
	return cleanStem, ext, true
}

func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	rawPath := r.URL.Path
	pathname, err := url.PathUnescape(rawPath)
	if err != nil {
		pathname = rawPath
	}

	// 1. Root and Stashed Decks
	if pathname == "/" || pathname == "/index.html" {
		deckIDStr := r.URL.Query().Get("deck")
		if deckIDStr != "" {
			var id int
			_, _ = fmt.Sscanf(deckIDStr, "%d", &id)
			htmlStr, ok := s.GetStashedDeck(id)
			if !ok {
				http.Error(w, "This build has expired. Press present again in the editor.", http.StatusGone)
				return
			}
			sendHTML(w, htmlStr)
			return
		}

		var fileInfo *render.EditorFileInfo
		if s.Mode.File != nil {
			fileInfo = &render.EditorFileInfo{
				Name:     s.Mode.File.Name,
				Kind:     s.Mode.File.Kind,
				Writable: s.Mode.File.Path != "",
				Watched:  s.Mode.File.Watched,
			}
		}

		editorHTML := render.GenerateEditorHTML(
			s.Mode.Theme,
			s.Mode.Fonts,
			s.Mode.Template,
			s.Mode.Transition,
			fileInfo,
		)
		sendHTML(w, editorHTML)
		return
	}

	// 2. Embedded Vendor Assets (KaTeX & Mermaid)
	if strings.HasPrefix(pathname, "/__vendor/") {
		rel := strings.TrimPrefix(pathname, "/__vendor/")
		assetPath := "vendor/" + rel

		data, err := assets.VendorFS.ReadFile(assetPath)
		if err != nil {
			http.NotFound(w, r)
			return
		}

		w.Header().Set("Content-Type", getMime(assetPath))
		w.Header().Set("Cache-Control", "public, max-age=31536000, immutable")
		_, _ = w.Write(data)
		return
	}

	// 2b. Local Media Assets (GET /__local?path=..., GET /__assets, POST /__upload-asset)
	if pathname == "/__local" && r.Method == http.MethodGet {
		qPath := r.URL.Query().Get("path")
		resolved, allowed, resolveErr := ResolveLocalMediaPath(s.BaseDir, qPath)
		if !allowed {
			http.Error(w, "Forbidden", http.StatusForbidden)
			return
		}
		if resolveErr != nil {
			http.NotFound(w, r)
			return
		}
		data, err := os.ReadFile(resolved)
		if err != nil {
			http.NotFound(w, r)
			return
		}
		w.Header().Set("Content-Type", getMime(resolved))
		w.Header().Set("Cache-Control", "no-cache")
		_, _ = w.Write(data)
		return
	}

	if pathname == "/__assets" && r.Method == http.MethodGet {
		files := []string{}
		_ = filepath.WalkDir(s.BaseDir, func(path string, d fs.DirEntry, err error) error {
			if err != nil {
				return nil
			}
			rel, relErr := filepath.Rel(s.BaseDir, path)
			if relErr != nil || rel == "." {
				return nil
			}
			slashRel := filepath.ToSlash(rel)
			parts := strings.Split(slashRel, "/")
			if d.IsDir() {
				name := d.Name()
				if strings.HasPrefix(name, ".") || name == "node_modules" || name == "bin" || len(parts) > 2 {
					return filepath.SkipDir
				}
				return nil
			}
			if len(parts) > 3 {
				return nil
			}
			ext := strings.ToLower(filepath.Ext(d.Name()))
			if allowedMediaExts[ext] {
				files = append(files, slashRel)
				if len(files) >= 60 {
					return fs.SkipAll
				}
			}
			return nil
		})
		sendJSON(w, map[string]any{
			"baseDir": s.BaseDir,
			"files":   files,
		})
		return
	}

	if pathname == "/__upload-asset" && r.Method == http.MethodPost {
		body, err := readBody(r)
		if err != nil {
			http.Error(w, err.Error(), http.StatusRequestEntityTooLarge)
			return
		}
		var req struct {
			Name string `json:"name"`
			Data string `json:"data"`
		}
		if err := json.Unmarshal(body, &req); err != nil {
			http.Error(w, "invalid JSON", http.StatusBadRequest)
			return
		}

		rawData := strings.TrimSpace(req.Data)
		if strings.HasPrefix(rawData, "data:") {
			if comma := strings.Index(rawData, ","); comma != -1 {
				rawData = rawData[comma+1:]
			}
		}
		decoded, err := base64.StdEncoding.DecodeString(rawData)
		if err != nil {
			decoded, err = base64.RawStdEncoding.DecodeString(rawData)
			if err != nil {
				http.Error(w, "invalid base64 image data", http.StatusBadRequest)
				return
			}
		}

		stem, ext, ok := safeAssetName(req.Name)
		if !ok {
			http.Error(w, "unsupported media file type", http.StatusBadRequest)
			return
		}

		assetsDir := filepath.Join(s.BaseDir, "assets")
		if err := os.MkdirAll(assetsDir, 0755); err != nil {
			http.Error(w, "cannot create assets directory", http.StatusInternalServerError)
			return
		}

		finalName := stem + ext
		for idx := 2; idx < 1000; idx++ {
			candPath := filepath.Join(assetsDir, finalName)
			existing, statErr := os.ReadFile(candPath)
			if statErr != nil {
				break
			}
			if bytes.Equal(existing, decoded) {
				break
			}
			finalName = fmt.Sprintf("%s-%d%s", stem, idx, ext)
		}

		destPath := filepath.Join(assetsDir, finalName)
		if err := os.WriteFile(destPath, decoded, 0644); err != nil {
			http.Error(w, "failed to save asset", http.StatusInternalServerError)
			return
		}

		relPath := "assets/" + finalName
		sendJSON(w, map[string]string{
			"path": relPath,
			"url":  "/__local?path=" + url.QueryEscape(relPath),
		})
		return
	}

	// 3. Preview Frame
	if pathname == "/__preview" && r.Method == http.MethodGet {
		previewHTML := render.GeneratePreviewHTML(
			s.Mode.Theme,
			s.Mode.Fonts,
			s.Mode.Template,
			s.Mode.Transition,
		)
		sendHTML(w, previewHTML)
		return
	}

	// 4. File Read & Write
	if pathname == "/__file" {
		if r.Method == http.MethodGet && s.Mode.File != nil {
			content := s.Mode.File.Content
			if s.Mode.File.Path != "" {
				b, err := os.ReadFile(s.Mode.File.Path)
				if err != nil {
					http.Error(w, "Cannot read the file.", http.StatusInternalServerError)
					return
				}
				content = string(b)
			}
			w.Header().Set("Content-Type", "text/plain; charset=utf-8")
			w.Header().Set("Cache-Control", "no-store")
			_, _ = w.Write([]byte(content))
			return
		}

		if r.Method == http.MethodPost && s.Mode.File != nil {
			if s.Mode.File.Path == "" {
				w.WriteHeader(http.StatusMethodNotAllowed)
				sendJSON(w, map[string]string{
					"error":  "read-only",
					"detail": "This document came from a URL; download it to keep changes.",
				})
				return
			}

			body, err := readBody(r)
			if err != nil {
				http.Error(w, err.Error(), http.StatusRequestEntityTooLarge)
				return
			}

			content := string(body)
			if s.Watcher != nil {
				s.Watcher.SetLastEditorWrite(content)
			}

			if err := os.WriteFile(s.Mode.File.Path, body, 0644); err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				sendJSON(w, map[string]string{"error": "write failed"})
				return
			}

			sendJSON(w, map[string]bool{"ok": true})
			return
		}
	}

	// 5. SSE Live Events
	if pathname == "/__events" && r.Method == http.MethodGet && s.Watcher != nil && s.Mode.File != nil && s.Mode.File.Watched {
		ServeEvents(s.Watcher, w, r)
		return
	}

	// 6. Parse Markdown
	if pathname == "/__parse" && r.Method == http.MethodPost {
		body, err := readBody(r)
		if err != nil {
			http.Error(w, err.Error(), http.StatusRequestEntityTooLarge)
			return
		}

		slides := parser.ParseSlides(string(body))
		slideHTMLs := make([]string, len(slides))
		notes := make([]string, len(slides))
		for i, s := range slides {
			slideHTMLs[i] = render.RenderSlide(s, i)
			notes[i] = s.Notes
		}

		sendJSON(w, map[string]any{
			"slides": slideHTMLs,
			"notes":  notes,
			"title":  parser.DeckTitle(slides, ""),
		})
		return
	}

	// 7. Present Deck
	if pathname == "/__present" && r.Method == http.MethodPost {
		body, err := readBody(r)
		if err != nil {
			http.Error(w, err.Error(), http.StatusRequestEntityTooLarge)
			return
		}

		var req struct {
			Markdown   string  `json:"markdown"`
			Theme      string  `json:"theme"`
			Head       *string `json:"head"`
			Body       *string `json:"body"`
			Title      string  `json:"title"`
			Print      bool    `json:"print"`
			Template   string  `json:"template"`
			Transition string  `json:"transition"`
			Standalone bool    `json:"standalone"`
		}
		if err := json.Unmarshal(body, &req); err != nil {
			http.Error(w, "invalid JSON", http.StatusBadRequest)
			return
		}

		slides := parser.ParseSlides(req.Markdown)
		if len(slides) == 0 {
			w.WriteHeader(http.StatusUnprocessableEntity)
			sendJSON(w, map[string]string{"error": "no slides"})
			return
		}

		th := theme.ResolveThemeName(req.Theme)
		tm := theme.ResolveTemplateName(req.Template)
		tr := theme.ResolveTransitionName(req.Transition)
		fallbackTitle := strings.TrimSpace(req.Title)
		if fallbackTitle == "" {
			fallbackTitle = "deckrun"
		}
		t := parser.DeckTitle(slides, fallbackTitle)

		headFont := ""
		if req.Head != nil {
			headFont = *req.Head
		}
		bodyFont := ""
		if req.Body != nil {
			bodyFont = *req.Body
		}

		forPrint := req.Print
		autoFullscreen := s.Mode.Fullscreen
		if forPrint {
			autoFullscreen = false
		}

		htmlStr := render.GenerateHTML(
			slides,
			t,
			autoFullscreen,
			th,
			render.FontOverrides{Head: headFont, Body: bodyFont},
			render.PresentationOptions{Template: string(tm), Transition: string(tr), Standalone: req.Standalone},
		)

		path := s.StashDeck(htmlStr)
		if forPrint {
			path += "&print=1"
		}

		sendJSON(w, map[string]string{"path": path})
		return
	}

	// 8. Present HTML Doc
	if pathname == "/__present-doc" && r.Method == http.MethodPost {
		body, err := readBody(r)
		if err != nil {
			http.Error(w, err.Error(), http.StatusRequestEntityTooLarge)
			return
		}

		var req struct {
			HTML  string `json:"html"`
			Theme string `json:"theme"`
			Title string `json:"title"`
			Print bool   `json:"print"`
		}
		if err := json.Unmarshal(body, &req); err != nil {
			http.Error(w, "invalid JSON", http.StatusBadRequest)
			return
		}

		raw := strings.TrimSpace(req.HTML)
		if raw == "" {
			w.WriteHeader(http.StatusUnprocessableEntity)
			sendJSON(w, map[string]string{"error": "empty document"})
			return
		}
		raw = parser.RewriteLocalAssetURLsInHTML(raw)

		th := theme.ResolveThemeName(req.Theme)
		fallbackTitle := strings.TrimSpace(req.Title)
		if fallbackTitle == "" {
			fallbackTitle = "deckrun"
		}
		t := parser.DocTitle(raw, fallbackTitle)

		forPrint := req.Print
		autoFullscreen := s.Mode.Fullscreen
		if forPrint {
			autoFullscreen = false
		}

		docPath := s.StashDeck(raw)
		wrapperPath := s.StashDeck(render.GenerateDocHTML(docPath, t, autoFullscreen, th))
		if forPrint {
			wrapperPath += "&print=1"
		}

		sendJSON(w, map[string]string{
			"path":    wrapperPath,
			"docPath": docPath,
		})
		return
	}

	// 9. Export PDF
	if pathname == "/__pdf" && r.Method == http.MethodPost {
		body, err := readBody(r)
		if err != nil {
			http.Error(w, err.Error(), http.StatusRequestEntityTooLarge)
			return
		}

		var req struct {
			Markdown   string  `json:"markdown"`
			Theme      string  `json:"theme"`
			Head       *string `json:"head"`
			Body       *string `json:"body"`
			Title      string  `json:"title"`
			Template   string  `json:"template"`
			Transition string  `json:"transition"`
		}
		if err := json.Unmarshal(body, &req); err != nil {
			http.Error(w, "invalid JSON", http.StatusBadRequest)
			return
		}

		slides := parser.ParseSlides(req.Markdown)
		if len(slides) == 0 {
			w.WriteHeader(http.StatusUnprocessableEntity)
			sendJSON(w, map[string]string{"error": "no slides"})
			return
		}

		browser := pdf.FindBrowser()
		if browser == "" {
			w.WriteHeader(http.StatusNotImplemented)
			sendJSON(w, map[string]string{
				"error":  "no browser",
				"detail": "No Chrome, Chromium, Edge, or Brave found. Set DECKRUN_BROWSER to one to export PDFs directly.",
			})
			return
		}

		th := theme.ResolveThemeName(req.Theme)
		tm := theme.ResolveTemplateName(req.Template)
		tr := theme.ResolveTransitionName(req.Transition)
		fallbackTitle := strings.TrimSpace(req.Title)
		if fallbackTitle == "" {
			fallbackTitle = "deckrun"
		}
		t := parser.DeckTitle(slides, fallbackTitle)

		headFont := ""
		if req.Head != nil {
			headFont = *req.Head
		}
		bodyFont := ""
		if req.Body != nil {
			bodyFont = *req.Body
		}

		deckHTML := render.GenerateHTML(
			slides,
			t,
			false,
			th,
			render.FontOverrides{Head: headFont, Body: bodyFont},
			render.PresentationOptions{Template: string(tm), Transition: string(tr)},
		)

		path := s.StashDeck(deckHTML)
		pdfBytes, err := pdf.RenderPDFSerial(s.Mode.Origin+path, browser)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			sendJSON(w, map[string]string{
				"error":  "render failed",
				"detail": err.Error(),
			})
			return
		}

		filename := safeFilename(t) + ".pdf"
		w.Header().Set("Content-Type", "application/pdf")
		w.Header().Set("Content-Length", fmt.Sprintf("%d", len(pdfBytes)))
		w.Header().Set("Content-Disposition", fmt.Sprintf(`attachment; filename="%s"`, filename))
		w.Header().Set("Cache-Control", "no-store")
		_, _ = w.Write(pdfBytes)
		return
	}

	// 10. Export Doc PDF
	if pathname == "/__pdf-doc" && r.Method == http.MethodPost {
		body, err := readBody(r)
		if err != nil {
			http.Error(w, err.Error(), http.StatusRequestEntityTooLarge)
			return
		}

		var req struct {
			HTML  string `json:"html"`
			Title string `json:"title"`
		}
		if err := json.Unmarshal(body, &req); err != nil {
			http.Error(w, "invalid JSON", http.StatusBadRequest)
			return
		}

		raw := strings.TrimSpace(req.HTML)
		if raw == "" {
			w.WriteHeader(http.StatusUnprocessableEntity)
			sendJSON(w, map[string]string{"error": "empty document"})
			return
		}
		raw = parser.RewriteLocalAssetURLsInHTML(raw)

		browser := pdf.FindBrowser()
		if browser == "" {
			w.WriteHeader(http.StatusNotImplemented)
			sendJSON(w, map[string]string{
				"error":  "no browser",
				"detail": "No Chrome, Chromium, Edge, or Brave found. Set DECKRUN_BROWSER to one to export PDFs directly.",
			})
			return
		}

		fallbackTitle := strings.TrimSpace(req.Title)
		if fallbackTitle == "" {
			fallbackTitle = "deckrun"
		}
		t := parser.DocTitle(raw, fallbackTitle)

		docPath := s.StashDeck(raw)
		pdfBytes, err := pdf.RenderPDFSerial(s.Mode.Origin+docPath, browser)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			sendJSON(w, map[string]string{
				"error":  "render failed",
				"detail": err.Error(),
			})
			return
		}

		filename := safeFilename(t) + ".pdf"
		w.Header().Set("Content-Type", "application/pdf")
		w.Header().Set("Content-Length", fmt.Sprintf("%d", len(pdfBytes)))
		w.Header().Set("Content-Disposition", fmt.Sprintf(`attachment; filename="%s"`, filename))
		w.Header().Set("Cache-Control", "no-store")
		_, _ = w.Write(pdfBytes)
		return
	}

	// 11. Fetch Remote Doc
	if pathname == "/__fetch-doc" && r.Method == http.MethodPost {
		body, err := readBody(r)
		if err != nil {
			http.Error(w, err.Error(), http.StatusRequestEntityTooLarge)
			return
		}

		var req struct {
			URL string `json:"url"`
		}
		_ = json.Unmarshal(body, &req)
		rawURL := strings.TrimSpace(req.URL)

		target, err := url.Parse(rawURL)
		if err != nil || (target.Scheme != "http" && target.Scheme != "https") {
			w.WriteHeader(http.StatusUnprocessableEntity)
			sendJSON(w, map[string]string{"error": "url must be http or https"})
			return
		}

		client := &http.Client{
			Timeout: 15 * time.Second,
		}
		httpReq, _ := http.NewRequest(http.MethodGet, target.String(), nil)
		httpReq.Header.Set("User-Agent", "deckrun")

		resp, err := client.Do(httpReq)
		if err != nil {
			w.WriteHeader(http.StatusBadGateway)
			sendJSON(w, map[string]string{"error": "fetch failed", "detail": err.Error()})
			return
		}
		defer resp.Body.Close()

		if resp.StatusCode < 200 || resp.StatusCode >= 300 {
			w.WriteHeader(http.StatusBadGateway)
			sendJSON(w, map[string]string{"error": "fetch failed", "detail": fmt.Sprintf("upstream responded %d", resp.StatusCode)})
			return
		}

		limited := io.LimitReader(resp.Body, maxBody+1)
		contentBytes, err := io.ReadAll(limited)
		if err != nil {
			w.WriteHeader(http.StatusBadGateway)
			sendJSON(w, map[string]string{"error": "fetch failed", "detail": err.Error()})
			return
		}

		if len(contentBytes) > maxBody {
			w.WriteHeader(http.StatusRequestEntityTooLarge)
			sendJSON(w, map[string]string{"error": "too large", "detail": fmt.Sprintf("page is larger than %d MB", maxBody/1024/1024)})
			return
		}

		rawContent := string(contentBytes)
		if strings.TrimSpace(rawContent) == "" {
			w.WriteHeader(http.StatusUnprocessableEntity)
			sendJSON(w, map[string]string{"error": "empty document"})
			return
		}

		contentType := strings.ToLower(resp.Header.Get("Content-Type"))
		pathLower := strings.ToLower(target.Path)

		isHTML := false
		if strings.HasSuffix(pathLower, ".html") || strings.HasSuffix(pathLower, ".htm") {
			isHTML = true
		} else if strings.HasSuffix(pathLower, ".md") || strings.HasSuffix(pathLower, ".markdown") {
			isHTML = false
		} else if strings.Contains(contentType, "text/html") || strings.Contains(contentType, "application/xhtml+xml") {
			isHTML = true
		} else if strings.Contains(contentType, "text/markdown") || strings.Contains(contentType, "text/x-markdown") || strings.Contains(contentType, "text/plain") {
			isHTML = false
		} else if strings.Contains(strings.ToLower(rawContent), "<!doctype html") || strings.Contains(strings.ToLower(rawContent), "<html") {
			isHTML = true
		}

		parts := strings.Split(strings.Trim(target.Path, "/"), "/")
		defaultName := target.Hostname()
		if len(parts) > 0 && parts[len(parts)-1] != "" {
			defaultName = parts[len(parts)-1]
		}

		kind := "markdown"
		title := ""
		if isHTML {
			kind = "html"
			title = parser.DocTitle(rawContent, defaultName)
		} else {
			slides := parser.ParseSlides(rawContent)
			title = parser.DeckTitle(slides, defaultName)
		}

		sendJSON(w, map[string]any{
			"kind":     kind,
			"content":  rawContent,
			"html":     rawContent,
			"markdown": rawContent,
			"title":    title,
		})
		return
	}

	// 12. Static Files from Workspace
	cleanRel := filepath.Clean(strings.TrimPrefix(pathname, "/"))
	filePath := filepath.Join(s.BaseDir, cleanRel)

	rel, err := filepath.Rel(s.BaseDir, filePath)
	if err != nil || strings.HasPrefix(rel, "..") || filepath.IsAbs(rel) {
		http.Error(w, "Forbidden", http.StatusForbidden)
		return
	}

	info, err := os.Stat(filePath)
	if err != nil || info.IsDir() {
		if pathname == "/favicon.ico" {
			w.Header().Set("Content-Type", "image/svg+xml")
			w.Header().Set("Cache-Control", "public, max-age=86400")
			_, _ = w.Write([]byte(`<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 32 32"><rect width="32" height="32" rx="7" fill="#1e1e2e"/><rect x="8" y="7" width="6" height="18" rx="2" fill="#cba6f7"/><polygon points="17,7 26,16 17,25" fill="#89b4fa"/></svg>`))
			return
		}
		http.NotFound(w, r)
		return
	}

	data, err := os.ReadFile(filePath)
	if err != nil {
		http.NotFound(w, r)
		return
	}

	w.Header().Set("Content-Type", getMime(filePath))
	_, _ = w.Write(data)
}

func FindFreePort(preferred int) int {
	addr := fmt.Sprintf("127.0.0.1:%d", preferred)
	l, err := net.Listen("tcp", addr)
	if err == nil {
		defer l.Close()
		return preferred
	}

	l, err = net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		return preferred
	}
	defer l.Close()
	return l.Addr().(*net.TCPAddr).Port
}
