package pdf

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"sync"
	"syscall"
	"time"
)

var candidates = map[string][]string{
	"darwin": {
		"/Applications/Google Chrome.app/Contents/MacOS/Google Chrome",
		"/Applications/Chromium.app/Contents/MacOS/Chromium",
		"/Applications/Microsoft Edge.app/Contents/MacOS/Microsoft Edge",
		"/Applications/Brave Browser.app/Contents/MacOS/Brave Browser",
		"/Applications/Arc.app/Contents/MacOS/Arc",
		"/Applications/Google Chrome Canary.app/Contents/MacOS/Google Chrome Canary",
	},
	"linux": {
		"/usr/bin/google-chrome",
		"/usr/bin/google-chrome-stable",
		"/usr/bin/chromium",
		"/usr/bin/chromium-browser",
		"/usr/bin/microsoft-edge",
		"/usr/bin/brave-browser",
		"/snap/bin/chromium",
		"/opt/google/chrome/chrome",
	},
	"windows": {
		`C:\Program Files\Google\Chrome\Application\chrome.exe`,
		`C:\Program Files (x86)\Google\Chrome\Application\chrome.exe`,
		`C:\Program Files\Microsoft\Edge\Application\msedge.exe`,
		`C:\Program Files (x86)\Microsoft\Edge\Application\msedge.exe`,
		`C:\Program Files\Chromium\Application\chrome.exe`,
	},
}

var envKeys = []string{
	"DECKRUN_BROWSER",
	"CHROME_PATH",
	"PUPPETEER_EXECUTABLE_PATH",
}

var (
	cachedBrowser     string
	cachedBrowserOnce sync.Once
)

func isExecutable(path string) bool {
	info, err := os.Stat(path)
	if err != nil {
		return false
	}
	if runtime.GOOS == "windows" {
		return !info.IsDir()
	}
	return !info.IsDir() && (info.Mode()&0111 != 0)
}

func FindBrowser() string {
	cachedBrowserOnce.Do(func() {
		for _, key := range envKeys {
			val := os.Getenv(key)
			if val != "" && isExecutable(val) {
				cachedBrowser = val
				return
			}
		}

		list := candidates[runtime.GOOS]
		for _, path := range list {
			if isExecutable(path) {
				cachedBrowser = path
				return
			}
		}
	})

	return cachedBrowser
}

const (
	renderTimeout = 45 * time.Second
	pollInterval  = 150 * time.Millisecond
)

func killTree(cmd *exec.Cmd) {
	if cmd == nil || cmd.Process == nil {
		return
	}
	if runtime.GOOS != "windows" && cmd.Process.Pid > 0 {
		// Kill the entire process group
		_ = syscall.Kill(-cmd.Process.Pid, syscall.SIGKILL)
	} else {
		_ = cmd.Process.Kill()
	}
}

func waitForPDF(out string, hasExited func() bool, spawnErr func() error) ([]byte, error) {
	deadline := time.Now().Add(renderTimeout)
	lastSize := int64(-1)

	for time.Now().Before(deadline) {
		if err := spawnErr(); err != nil {
			return nil, err
		}

		info, err := os.Stat(out)
		size := int64(-1)
		if err == nil {
			size = info.Size()
		}

		if size > 0 && size == lastSize {
			return os.ReadFile(out)
		}
		lastSize = size

		if hasExited() && size <= 0 {
			// One last look in case write landed as process left
			if info, err := os.Stat(out); err == nil && info.Size() > 0 {
				return os.ReadFile(out)
			}
			return nil, errors.New("the browser exited without producing a PDF")
		}

		time.Sleep(pollInterval)
	}

	return nil, errors.New("the browser took too long to render the deck")
}

func RenderPDF(urlStr string, browser string) ([]byte, error) {
	dir, err := os.MkdirTemp("", "deckrun-pdf-")
	if err != nil {
		return nil, fmt.Errorf("failed to create temp dir: %w", err)
	}
	defer os.RemoveAll(dir)

	out := filepath.Join(dir, "deck.pdf")
	profileDir := filepath.Join(dir, "profile")

	args := []string{
		"--headless",
		"--disable-gpu",
		"--hide-scrollbars",
		"--no-first-run",
		"--no-default-browser-check",
		"--disable-extensions",
		"--disable-sync",
		"--disable-default-apps",
		"--disable-component-update",
		"--no-service-autorun",
		"--mute-audio",
		"--user-data-dir=" + profileDir,
		"--virtual-time-budget=10000",
		"--run-all-compositor-stages-before-draw",
		"--no-pdf-header-footer",
		"--print-to-pdf-no-header",
		"--print-to-pdf=" + out,
		urlStr,
	}

	cmd := exec.Command(browser, args...)
	if runtime.GOOS != "windows" {
		cmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}
	}

	exited := false
	var exitMu sync.Mutex
	var runErr error

	if err := cmd.Start(); err != nil {
		return nil, fmt.Errorf("could not run %s: %w", browser, err)
	}

	go func() {
		err := cmd.Wait()
		exitMu.Lock()
		exited = true
		runErr = err
		exitMu.Unlock()
	}()

	isExited := func() bool {
		exitMu.Lock()
		defer exitMu.Unlock()
		return exited
	}

	getErr := func() error {
		exitMu.Lock()
		defer exitMu.Unlock()
		return runErr
	}

	defer killTree(cmd)

	return waitForPDF(out, isExited, getErr)
}

var pdfMu sync.Mutex

func RenderPDFSerial(urlStr string, browser string) ([]byte, error) {
	pdfMu.Lock()
	defer pdfMu.Unlock()
	return RenderPDF(urlStr, browser)
}

