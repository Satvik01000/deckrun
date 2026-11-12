package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"time"

	"deckrun/internal/lint"
	"deckrun/internal/parser"
	"deckrun/internal/render"
	"deckrun/internal/server"
	"deckrun/internal/theme"
)

var version = "2.0.0"

const (
	colReset   = "\x1b[0m"
	colBold    = "\x1b[1m"
	colDim     = "\x1b[2m"
	colCyan    = "\x1b[36m"
	colGreen   = "\x1b[32m"
	colYellow  = "\x1b[33m"
	colMagenta = "\x1b[35m"
)

func openBrowser(targetURL string) {
	var cmd *exec.Cmd
	switch runtime.GOOS {
	case "darwin":
		cmd = exec.Command("open", targetURL)
	case "windows":
		cmd = exec.Command("rundll32", "url.dll,FileProtocolHandler", targetURL)
	default:
		cmd = exec.Command("xdg-open", targetURL)
	}
	_ = cmd.Start()
}

func printHelp() {
	fmt.Print(`Usage: deckrun [options] [command] [file]

Open a Markdown file, HTML file, or public URL in the built-in editor, and
present from there. Run without an argument for a blank editor.

Arguments:
  file                       Markdown file, HTML file, or public URL to open in
                             the editor. Omit it for a blank editor.

Options:
  -v, --version              Print the version number
  -p, --port <number>        Port to serve on (default: "7890")
  --no-open                  Do not automatically open the browser
  --no-watch                 Do not watch the opened file for changes on disk
  --fullscreen               Auto-enter fullscreen on first interaction
  --theme <name>             Color theme, by id (see --list-themes) (default: "nord")
  --head-font <name>         Override the theme's heading face (see --list-fonts)
  --body-font <name>         Override the theme's body face (see --list-fonts)
  --template <name>          Composition template (see --list-templates) (default: "classic")
  --transition <name>        Slide transition (see --list-transitions) (default: "slide")
  --list-themes              Print every theme and exit
  --list-fonts               Print every font face and exit
  --list-templates           Print every composition template and exit
  --list-transitions         Print every slide transition and exit
  -h, --help                 display help for command

Commands:
  lint [options] <files...>  Check Markdown decks for common authoring and
                             rendering problems
`)
}

func printLintHelp() {
	fmt.Print(`Usage: deckrun lint [options] <files...>

Check Markdown decks for common authoring and rendering problems

Arguments:
  files                      Markdown files to check; use - to read standard input

Options:
  --format <format>          Output format: stylish or json (default: "stylish")
  --max-warnings <number>    Warnings allowed before the command fails (default: "0")
  -h, --help                 display help for command
`)
}

func handleLint(args []string) {
	format := "stylish"
	maxWarnings := 0
	maxWarningsSet := false
	var files []string

	for i := 0; i < len(args); i++ {
		arg := args[i]
		switch {
		case arg == "-h" || arg == "--help":
			printLintHelp()
			os.Exit(0)
		case arg == "--format":
			if i+1 < len(args) {
				format = args[i+1]
				i++
			}
		case strings.HasPrefix(arg, "--format="):
			format = strings.TrimPrefix(arg, "--format=")
		case arg == "--max-warnings":
			if i+1 < len(args) {
				val, err := strconv.Atoi(args[i+1])
				if err != nil || (val < -1) {
					fmt.Fprintln(os.Stderr, "deckrun lint: --max-warnings must be -1 or a non-negative integer.")
					os.Exit(2)
				}
				maxWarnings = val
				maxWarningsSet = true
				i++
			}
		case strings.HasPrefix(arg, "--max-warnings="):
			val, err := strconv.Atoi(strings.TrimPrefix(arg, "--max-warnings="))
			if err != nil || (val < -1) {
				fmt.Fprintln(os.Stderr, "deckrun lint: --max-warnings must be -1 or a non-negative integer.")
				os.Exit(2)
			}
			maxWarnings = val
			maxWarningsSet = true
		default:
			if !strings.HasPrefix(arg, "-") {
				files = append(files, arg)
			}
		}
	}

	if format != "stylish" && format != "json" {
		fmt.Fprintln(os.Stderr, "deckrun lint: --format must be 'stylish' or 'json'.")
		os.Exit(2)
	}

	if !maxWarningsSet {
		maxWarnings = 0
	}

	if len(files) == 0 {
		printLintHelp()
		os.Exit(1)
	}

	var reports []lint.Result
	for _, file := range files {
		var mdContent string
		if file == "-" {
			b, err := io.ReadAll(os.Stdin)
			if err != nil {
				reports = append(reports, lint.Result{
					File:     file,
					Slides:   0,
					Errors:   1,
					Warnings: 0,
					Issues: []lint.Issue{
						{
							Rule:     "file-read",
							Severity: lint.SeverityError,
							Message:  "The file could not be read.",
							Line:     1,
							Column:   1,
						},
					},
				})
				continue
			}
			mdContent = string(b)
		} else {
			b, err := os.ReadFile(file)
			if err != nil {
				reports = append(reports, lint.Result{
					File:     file,
					Slides:   0,
					Errors:   1,
					Warnings: 0,
					Issues: []lint.Issue{
						{
							Rule:     "file-read",
							Severity: lint.SeverityError,
							Message:  "The file could not be read.",
							Line:     1,
							Column:   1,
						},
					},
				})
				continue
			}
			mdContent = string(b)
		}

		res := lint.LintMarkdown(mdContent)
		res.File = file
		reports = append(reports, res)
	}

	totalErrors := 0
	totalWarnings := 0
	for _, r := range reports {
		totalErrors += r.Errors
		totalWarnings += r.Warnings
	}

	if format == "json" {
		out := map[string]any{
			"files":    reports,
			"errors":   totalErrors,
			"warnings": totalWarnings,
		}
		b, _ := json.MarshalIndent(out, "", "  ")
		fmt.Println(string(b))
	} else {
		fmt.Print(lint.FormatStylish(reports))
	}

	if totalErrors > 0 || (maxWarnings != -1 && totalWarnings > maxWarnings) {
		os.Exit(1)
	}
	os.Exit(0)
}

func main() {
	args := os.Args[1:]

	if len(args) > 0 && args[0] == "lint" {
		handleLint(args[1:])
		return
	}

	port := 7890
	shouldOpen := true
	shouldWatch := true
	fullscreen := false
	themeName := theme.DefaultTheme
	headFont := ""
	bodyFont := ""
	templateName := string(theme.DefaultTemplate)
	transitionName := string(theme.DefaultTransition)

	var targetFile string

	for i := 0; i < len(args); i++ {
		arg := args[i]
		switch {
		case arg == "-v" || arg == "--version":
			fmt.Println(version)
			return
		case arg == "-h" || arg == "--help":
			printHelp()
			return
		case arg == "--list-themes":
			for _, line := range theme.ThemeListing() {
				fmt.Println(line)
			}
			return
		case arg == "--list-fonts":
			for _, line := range theme.FontListing() {
				fmt.Println(line)
			}
			return
		case arg == "--list-templates":
			for _, line := range theme.TemplateListing() {
				fmt.Println(line)
			}
			return
		case arg == "--list-transitions":
			for _, line := range theme.TransitionListing() {
				fmt.Println(line)
			}
			return
		case arg == "-p" || arg == "--port":
			if i+1 < len(args) {
				p, err := strconv.Atoi(args[i+1])
				if err == nil {
					port = p
				}
				i++
			}
		case strings.HasPrefix(arg, "--port="):
			p, err := strconv.Atoi(strings.TrimPrefix(arg, "--port="))
			if err == nil {
				port = p
			}
		case arg == "--no-open":
			shouldOpen = false
		case arg == "--no-watch":
			shouldWatch = false
		case arg == "--fullscreen":
			fullscreen = true
		case arg == "--theme":
			if i+1 < len(args) {
				themeName = args[i+1]
				i++
			}
		case strings.HasPrefix(arg, "--theme="):
			themeName = strings.TrimPrefix(arg, "--theme=")
		case arg == "--head-font":
			if i+1 < len(args) {
				headFont = args[i+1]
				i++
			}
		case strings.HasPrefix(arg, "--head-font="):
			headFont = strings.TrimPrefix(arg, "--head-font=")
		case arg == "--body-font":
			if i+1 < len(args) {
				bodyFont = args[i+1]
				i++
			}
		case strings.HasPrefix(arg, "--body-font="):
			bodyFont = strings.TrimPrefix(arg, "--body-font=")
		case arg == "--template":
			if i+1 < len(args) {
				templateName = args[i+1]
				i++
			}
		case strings.HasPrefix(arg, "--template="):
			templateName = strings.TrimPrefix(arg, "--template=")
		case arg == "--transition":
			if i+1 < len(args) {
				transitionName = args[i+1]
				i++
			}
		case strings.HasPrefix(arg, "--transition="):
			transitionName = strings.TrimPrefix(arg, "--transition=")
		default:
			if !strings.HasPrefix(arg, "-") && targetFile == "" {
				targetFile = arg
			}
		}
	}

	resolvedTheme, ok := theme.FindTheme(themeName)
	if !ok {
		fmt.Fprintf(os.Stderr, "deckrun: unknown theme '%s'. Run --list-themes to see them all.\n", themeName)
		os.Exit(1)
	}

	resolvedTemplate, ok := theme.FindTemplate(templateName)
	if !ok {
		fmt.Fprintf(os.Stderr, "deckrun: unknown template '%s'. Run --list-templates to see them all.\n", templateName)
		os.Exit(1)
	}

	resolvedTransition, ok := theme.FindTransition(transitionName)
	if !ok {
		fmt.Fprintf(os.Stderr, "deckrun: unknown transition '%s'. Run --list-transitions to see them all.\n", transitionName)
		os.Exit(1)
	}

	fonts := render.FontOverrides{}
	if headFont != "" {
		hf, ok := theme.FindFont(headFont)
		if !ok {
			fmt.Fprintf(os.Stderr, "deckrun: unknown font '%s' for --head-font. Run --list-fonts to see them all.\n", headFont)
			os.Exit(1)
		}
		fonts.Head = hf
	}
	if bodyFont != "" {
		bf, ok := theme.FindFont(bodyFont)
		if !ok {
			fmt.Fprintf(os.Stderr, "deckrun: unknown font '%s' for --body-font. Run --list-fonts to see them all.\n", bodyFont)
			os.Exit(1)
		}
		fonts.Body = bf
	}

	cwd, _ := os.Getwd()
	baseDir := cwd
	var editorFile *server.EditorFile

	if targetFile != "" {
		if strings.HasPrefix(strings.ToLower(targetFile), "http://") || strings.HasPrefix(strings.ToLower(targetFile), "https://") {
			target, err := url.Parse(targetFile)
			if err != nil {
				fmt.Fprintf(os.Stderr, "deckrun: invalid URL '%s'\n", targetFile)
				os.Exit(1)
			}

			client := &http.Client{Timeout: 15 * time.Second}
			req, _ := http.NewRequest(http.MethodGet, target.String(), nil)
			req.Header.Set("User-Agent", "deckrun")
			resp, err := client.Do(req)
			if err != nil {
				fmt.Fprintf(os.Stderr, "deckrun: cannot fetch '%s': %v\n", targetFile, err)
				os.Exit(1)
			}
			defer resp.Body.Close()

			if resp.StatusCode < 200 || resp.StatusCode >= 300 {
				fmt.Fprintf(os.Stderr, "deckrun: fetch failed for '%s' (HTTP %d)\n", targetFile, resp.StatusCode)
				os.Exit(1)
			}

			bodyBytes, err := io.ReadAll(resp.Body)
			if err != nil || len(strings.TrimSpace(string(bodyBytes))) == 0 {
				fmt.Fprintf(os.Stderr, "deckrun: empty document fetched from '%s'\n", targetFile)
				os.Exit(1)
			}

			rawContent := string(bodyBytes)
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

			if isHTML {
				docHTML := rawContent
				if !strings.Contains(strings.ToLower(docHTML), "<base ") {
					if strings.Contains(strings.ToLower(docHTML), "<head") {
						idx := strings.Index(strings.ToLower(docHTML), "<head")
						end := strings.Index(docHTML[idx:], ">")
						if end != -1 {
							insertPos := idx + end + 1
							docHTML = docHTML[:insertPos] + fmt.Sprintf("\n  <base href=\"%s\">", target.String()) + docHTML[insertPos:]
						}
					} else {
						docHTML = fmt.Sprintf("<base href=\"%s\">\n", target.String()) + docHTML
					}
				}
				editorFile = &server.EditorFile{
					Name:    parser.DocTitle(rawContent, defaultName),
					Kind:    "html",
					Content: docHTML,
					Watched: false,
				}
				fmt.Printf("%sopening %s in the editor as an HTML doc%s\n", colDim, targetFile, colReset)
			} else {
				slides := parser.ParseSlides(rawContent)
				editorFile = &server.EditorFile{
					Name:    defaultName,
					Kind:    "markdown",
					Content: rawContent,
					Watched: false,
				}
				slideWord := "slides"
				if len(slides) == 1 {
					slideWord = "slide"
				}
				fmt.Printf("%s%d %s from %s · opening in the editor%s\n", colDim, len(slides), slideWord, targetFile, colReset)
			}
		} else {
			absPath, err := filepath.Abs(targetFile)
			if err != nil {
				fmt.Fprintf(os.Stderr, "deckrun: cannot read file '%s'\n", targetFile)
				os.Exit(1)
			}

			b, err := os.ReadFile(absPath)
			if err != nil {
				fmt.Fprintf(os.Stderr, "deckrun: cannot read file '%s'\n", targetFile)
				os.Exit(1)
			}

			baseDir = filepath.Dir(absPath)
			ext := strings.ToLower(filepath.Ext(absPath))
			kind := "markdown"
			if ext == ".html" || ext == ".htm" {
				kind = "html"
			}

			editorFile = &server.EditorFile{
				Name:    filepath.Base(absPath),
				Kind:    kind,
				Path:    absPath,
				Watched: shouldWatch,
			}

			if kind == "markdown" {
				slides := parser.ParseSlides(string(b))
				slideWord := "slides"
				if len(slides) == 1 {
					slideWord = "slide"
				}
				fmt.Printf("%s%d %s from %s · opening in the editor%s\n", colDim, len(slides), slideWord, filepath.Base(absPath), colReset)
			} else {
				fmt.Printf("%sopening %s in the editor as an HTML doc%s\n", colDim, filepath.Base(absPath), colReset)
			}
		}
	}

	freePort := server.FindFreePort(port)
	origin := fmt.Sprintf("http://127.0.0.1:%d", freePort)

	var watcher *server.Watcher
	if editorFile != nil && editorFile.Watched && editorFile.Path != "" {
		watcher = server.NewWatcher(editorFile.Path)
		watcher.Start()
		defer watcher.Stop()
	}

	srv := server.NewServer(server.Mode{
		Theme:      resolvedTheme,
		Fonts:      fonts,
		Template:   string(resolvedTemplate),
		Transition: string(resolvedTransition),
		Fullscreen: fullscreen,
		File:       editorFile,
		Origin:     origin,
	}, baseDir, watcher)

	listener, err := net.Listen("tcp", fmt.Sprintf("127.0.0.1:%d", freePort))
	if err != nil {
		fmt.Fprintf(os.Stderr, "deckrun: cannot listen on port %d: %v\n", freePort, err)
		os.Exit(1)
	}

	httpServer := &http.Server{
		Handler: srv,
	}

	fmt.Printf("%s%seditor%s %s→%s %s%s%s%s  %s(Ctrl+C to stop)%s\n",
		colBold, colMagenta, colReset,
		colDim, colReset,
		colCyan, colBold, origin, colReset,
		colDim, colReset,
	)

	if editorFile != nil && editorFile.Path != "" {
		fmt.Printf("%swrite on the left, live deck on the right. saves back to %s.%s\n", colDim, editorFile.Name, colReset)
		if editorFile.Watched {
			fmt.Printf("%sedits to %s on disk reload the editor as well.%s\n", colDim, editorFile.Name, colReset)
		}
	} else {
		fmt.Printf("%swrite on the left, live deck on the right. autosaves to your browser.%s\n", colDim, colReset)
	}
	fmt.Printf("%sCmd/Ctrl+K inserts anything · template/theme controls recompose live · Cmd/Ctrl+Enter presents%s\n", colDim, colReset)

	if shouldOpen {
		go func() {
			time.Sleep(50 * time.Millisecond)
			openBrowser(origin)
		}()
	}

	if err := httpServer.Serve(listener); err != nil && err != http.ErrServerClosed {
		fmt.Fprintf(os.Stderr, "deckrun server error: %v\n", err)
		os.Exit(1)
	}
}
