package server

import (
	"log"
	"net/http"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/fsnotify/fsnotify"
)

type Watcher struct {
	mu              sync.Mutex
	absPath         string
	targetName      string
	lastEditorWrite string
	clients         map[chan string]struct{}
	stopCh          chan struct{}
}

func NewWatcher(absPath string) *Watcher {
	return &Watcher{
		absPath:    absPath,
		targetName: filepath.Base(absPath),
		clients:    make(map[chan string]struct{}),
		stopCh:     make(chan struct{}),
	}
}

func (w *Watcher) SetLastEditorWrite(content string) {
	w.mu.Lock()
	defer w.mu.Unlock()
	w.lastEditorWrite = content
}

func (w *Watcher) AddClient() chan string {
	w.mu.Lock()
	defer w.mu.Unlock()
	ch := make(chan string, 16)
	w.clients[ch] = struct{}{}
	return ch
}

func (w *Watcher) RemoveClient(ch chan string) {
	w.mu.Lock()
	defer w.mu.Unlock()
	delete(w.clients, ch)
	close(ch)
}

func (w *Watcher) NotifyAll() {
	w.mu.Lock()
	defer w.mu.Unlock()
	for ch := range w.clients {
		select {
		case ch <- "reload":
		default:
		}
	}
}

func (w *Watcher) Start() {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		log.Printf("deckrun: cannot watch '%s' for changes; live reload is off.", w.targetName)
		return
	}

	dir := filepath.Dir(w.absPath)
	if err := watcher.Add(dir); err != nil {
		log.Printf("deckrun: cannot watch directory '%s'; live reload is off.", dir)
		_ = watcher.Close()
		return
	}

	go func() {
		defer watcher.Close()
		var timer *time.Timer
		var timerMu sync.Mutex

		for {
			select {
			case <-w.stopCh:
				return
			case event, ok := <-watcher.Events:
				if !ok {
					return
				}

				// Only consider writes/creates/renames for our target file
				if filepath.Base(event.Name) != w.targetName {
					continue
				}

				timerMu.Lock()
				if timer != nil {
					timer.Stop()
				}
				timer = time.AfterFunc(150*time.Millisecond, func() {
					contentBytes, err := os.ReadFile(w.absPath)
					if err != nil {
						return // mid-save window of an atomic rename
					}
					content := string(contentBytes)

					w.mu.Lock()
					isOurWrite := (content == w.lastEditorWrite)
					w.mu.Unlock()

					if isOurWrite {
						return
					}

					w.NotifyAll()
				})
				timerMu.Unlock()

			case _, ok := <-watcher.Errors:
				if !ok {
					return
				}
			}
		}
	}()
}

func (w *Watcher) Stop() {
	close(w.stopCh)
}

func ServeEvents(w *Watcher, res http.ResponseWriter, req *http.Request) {
	flusher, ok := res.(http.Flusher)
	if !ok {
		http.Error(res, "Streaming unsupported", http.StatusInternalServerError)
		return
	}

	res.Header().Set("Content-Type", "text/event-stream")
	res.Header().Set("Cache-Control", "no-store")
	res.Header().Set("Connection", "keep-alive")

	_, _ = res.Write([]byte(": connected\n\n"))
	flusher.Flush()

	clientCh := w.AddClient()
	defer w.RemoveClient(clientCh)

	notify := req.Context().Done()

	for {
		select {
		case <-notify:
			return
		case msg, ok := <-clientCh:
			if !ok {
				return
			}
			_, err := res.Write([]byte("data: " + msg + "\n\n"))
			if err != nil {
				return
			}
			flusher.Flush()
		}
	}
}

