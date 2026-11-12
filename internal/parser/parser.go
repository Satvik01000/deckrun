package parser

import (
	"bytes"
	"html"
	"regexp"
	"strconv"
	"strings"

	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/extension"
	gmhtml "github.com/yuin/goldmark/renderer/html"
)

type PositionedImage struct {
	Src     string  `json:"src"`
	Alt     string  `json:"alt"`
	Opacity float64 `json:"opacity"`
}

type Slide struct {
	HTML       string           `json:"html"`
	BGImage    *PositionedImage `json:"bgImage,omitempty"`
	RightImage *PositionedImage `json:"rightImage,omitempty"`
	LeftImage  *PositionedImage `json:"leftImage,omitempty"`
	Notes      string           `json:"notes,omitempty"`
}

var mdRenderer = goldmark.New(
	goldmark.WithExtensions(
		extension.GFM,
	),
	goldmark.WithRendererOptions(
		gmhtml.WithUnsafe(),
	),
)

var notesRegex = regexp.MustCompile(`(?i)<!--\s*notes?:\s*([\s\S]*?)\s*-->`)
var imgRegex = regexp.MustCompile(`!\[([^\]]*)\]\(([^\s)]+)(?:\s+"([^"]*)")?\)`)
var opacityRegex = regexp.MustCompile(`opacity[=:]?\s*([0-9]*\.?[0-9]+)`)

func parseImageDirective(title string) (position string, opacity float64) {
	position = "inline"
	opacity = 1.0

	t := strings.ToLower(strings.TrimSpace(title))
	if t == "" {
		return position, opacity
	}

	if strings.Contains(t, "right") {
		position = "right"
	} else if strings.Contains(t, "left") {
		position = "left"
	} else if strings.Contains(t, "bg") {
		position = "bg"
	}

	if m := opacityRegex.FindStringSubmatch(t); len(m) > 1 {
		if val, err := strconv.ParseFloat(m[1], 64); err == nil {
			if val < 0 {
				val = 0
			} else if val > 1 {
				val = 1
			}
			opacity = val
		}
	}

	return position, opacity
}

// SplitSlides splits markdown into individual slides by "---" on its own line,
// ignoring any "---" inside fenced code blocks (Bug 1 fix).
func SplitSlides(markdown string) []string {
	normalized := strings.ReplaceAll(markdown, "\r\n", "\n")
	normalized = strings.ReplaceAll(normalized, "\r", "\n")

	lines := strings.Split(normalized, "\n")
	var slides []string
	var curSlide []string
	inFence := false
	fenceMarker := ""

	for _, line := range lines {
		trimmed := strings.TrimSpace(line)

		// Check fence toggles (``` or ~~~)
		if strings.HasPrefix(trimmed, "```") || strings.HasPrefix(trimmed, "~~~") {
			prefix := trimmed[:3]
			if !inFence {
				inFence = true
				fenceMarker = prefix
			} else if strings.HasPrefix(trimmed, fenceMarker) {
				inFence = false
				fenceMarker = ""
			}
			curSlide = append(curSlide, line)
			continue
		}

		// If outside fence and matches "---", it's a slide divider
		if !inFence && (trimmed == "---" || isDashesDivider(trimmed)) {
			slideText := strings.TrimSpace(strings.Join(curSlide, "\n"))
			if len(slideText) > 0 {
				slides = append(slides, slideText)
			}
			curSlide = nil
			continue
		}

		curSlide = append(curSlide, line)
	}

	if len(curSlide) > 0 {
		slideText := strings.TrimSpace(strings.Join(curSlide, "\n"))
		if len(slideText) > 0 {
			slides = append(slides, slideText)
		}
	}

	return slides
}

func isDashesDivider(line string) bool {
	if len(line) < 3 {
		return false
	}
	for i := 0; i < len(line); i++ {
		if line[i] != '-' && line[i] != ' ' && line[i] != '\t' {
			return false
		}
	}
	return strings.Count(line, "-") >= 3
}

// TokenizeMathAndReveal handles block math, inline math, and reveal markers
// while preserving code fences so math/reveal inside code is not touched.
func TokenizeMathAndReveal(src string) string {
	lines := strings.Split(src, "\n")
	var out []string
	inFence := false
	fenceMarker := ""

	for i := 0; i < len(lines); i++ {
		line := lines[i]
		trimmed := strings.TrimSpace(line)

		if strings.HasPrefix(trimmed, "```") || strings.HasPrefix(trimmed, "~~~") {
			prefix := trimmed[:3]
			if !inFence {
				inFence = true
				fenceMarker = prefix
			} else if strings.HasPrefix(trimmed, fenceMarker) {
				inFence = false
				fenceMarker = ""
			}
			out = append(out, line)
			continue
		}

		if inFence {
			out = append(out, line)
			continue
		}

		// Handle multiline or single-line display math $$...$$ outside fence
		if strings.HasPrefix(trimmed, "$$") {
			// Check if closed on same line: $$formula$$
			rest := strings.TrimSpace(trimmed[2:])
			if strings.HasSuffix(rest, "$$") && len(rest) >= 2 {
				formula := strings.TrimSuffix(rest, "$$")
				out = append(out, `<div class="math-source" data-display="true">`+html.EscapeString(strings.TrimSpace(formula))+`</div>`, "")
				continue
			}

			// Multiline $$ block
			var mathLines []string
			if len(rest) > 0 {
				mathLines = append(mathLines, rest)
			}
			closed := false
			for j := i + 1; j < len(lines); j++ {
				nextTrim := strings.TrimSpace(lines[j])
				if strings.HasSuffix(nextTrim, "$$") {
					content := strings.TrimSuffix(nextTrim, "$$")
					if len(content) > 0 {
						mathLines = append(mathLines, content)
					}
					i = j
					closed = true
					break
				}
				mathLines = append(mathLines, lines[j])
			}
			if closed {
				formula := strings.Join(mathLines, "\n")
				out = append(out, `<div class="math-source" data-display="true">`+html.EscapeString(strings.TrimSpace(formula))+`</div>`, "")
				continue
			}
		}

		// Handle \[ ... \]
		if strings.HasPrefix(trimmed, `\[`) {
			rest := strings.TrimSpace(trimmed[2:])
			if strings.HasSuffix(rest, `\]`) && len(rest) >= 2 {
				formula := strings.TrimSuffix(rest, `\]`)
				out = append(out, `<div class="math-source" data-display="true">`+html.EscapeString(strings.TrimSpace(formula))+`</div>`, "")
				continue
			}
			var mathLines []string
			if len(rest) > 0 {
				mathLines = append(mathLines, rest)
			}
			closed := false
			for j := i + 1; j < len(lines); j++ {
				nextTrim := strings.TrimSpace(lines[j])
				if strings.HasSuffix(nextTrim, `\]`) {
					content := strings.TrimSuffix(nextTrim, `\]`)
					if len(content) > 0 {
						mathLines = append(mathLines, content)
					}
					i = j
					closed = true
					break
				}
				mathLines = append(mathLines, lines[j])
			}
			if closed {
				formula := strings.Join(mathLines, "\n")
				out = append(out, `<div class="math-source" data-display="true">`+html.EscapeString(strings.TrimSpace(formula))+`</div>`, "")
				continue
			}
		}

		// Process inline reveal markers: {reveal}
		line = strings.ReplaceAll(line, "{reveal}", `<span class="deckrun-fragment-marker" aria-hidden="true"></span>`)

		// Process inline math: $...$ or \(...\)
		line = replaceInlineMath(line)

		out = append(out, line)
	}

	return strings.Join(out, "\n")
}

var inlineParenMath = regexp.MustCompile(`\\\(((?:\\.|[^\\\n])*?)\\\)`)

func replaceInlineMath(line string) string {
	// First handle \( ... \)
	line = inlineParenMath.ReplaceAllStringFunc(line, func(m string) string {
		sub := inlineParenMath.FindStringSubmatch(m)
		if len(sub) > 1 {
			return `<span class="math-source" data-display="false">` + html.EscapeString(sub[1]) + `</span>`
		}
		return m
	})

	// Scan for $...$
	var sb strings.Builder
	n := len(line)
	i := 0

	for i < n {
		if line[i] == '$' {
			// Escaped \$
			if i > 0 && line[i-1] == '\\' {
				sb.WriteByte('$')
				i++
				continue
			}

			// $$ is display math, skip
			if i+1 < n && line[i+1] == '$' {
				sb.WriteString("$$")
				i += 2
				continue
			}

			// Opening $ must not be followed by space or newline
			if i+1 >= n || line[i+1] == ' ' || line[i+1] == '\t' || line[i+1] == '$' {
				sb.WriteByte('$')
				i++
				continue
			}

			// Search for matching closing $
			end := -1
			for j := i + 1; j < n; j++ {
				if line[j] == '$' && line[j-1] != '\\' {
					// Closing $ must not be preceded by space
					if line[j-1] != ' ' && line[j-1] != '\t' {
						// Closing $ must not be followed by a digit (currency heuristic: $5 and $10)
						if j+1 < n && line[j+1] >= '0' && line[j+1] <= '9' {
							// This $ is followed by a digit (like $10). It cannot close math,
							// AND since an unescaped $ cannot be inside math, the search must abort.
							break
						}
						// Must not be followed by $ (e.g. $$)
						if j+1 < n && line[j+1] == '$' {
							break
						}
						end = j
						break
					} else {
						// Found a $, but it was preceded by a space. Unescaped $ cannot be inside math.
						break
					}
				}
			}

			if end != -1 {
				formula := line[i+1 : end]
				sb.WriteString(`<span class="math-source" data-display="false">` + html.EscapeString(formula) + `</span>`)
				i = end + 1
				continue
			}
		}

		sb.WriteByte(line[i])
		i++
	}

	return sb.String()
}

// ParseSlides parses a markdown string into slides with positioned images, notes,
// and rendered HTML.
func ParseSlides(markdown string) []Slide {
	rawSlides := SplitSlides(markdown)
	slides := make([]Slide, 0, len(rawSlides))

	for _, raw := range rawSlides {
		slide := Slide{}

		// Extract speaker notes
		if m := notesRegex.FindStringSubmatch(raw); len(m) > 1 {
			slide.Notes = strings.TrimSpace(m[1])
		}
		processedMd := notesRegex.ReplaceAllString(raw, "")

		// Find positioned images via title attribute: ![alt](src "right opacity:0.7")
		matches := imgRegex.FindAllStringSubmatch(raw, -1)
		for _, m := range matches {
			full := m[0]
			alt := m[1]
			src := m[2]
			titleAttr := ""
			if len(m) > 3 {
				titleAttr = m[3]
			}

			pos, op := parseImageDirective(titleAttr)
			img := &PositionedImage{
				Src:     src,
				Alt:     alt,
				Opacity: op,
			}

			switch pos {
			case "bg":
				slide.BGImage = img
				processedMd = strings.Replace(processedMd, full, "", 1)
			case "right":
				slide.RightImage = img
				processedMd = strings.Replace(processedMd, full, "", 1)
			case "left":
				slide.LeftImage = img
				processedMd = strings.Replace(processedMd, full, "", 1)
			}
		}

		// Tokenize math & reveal markers before Markdown conversion
		tokenized := TokenizeMathAndReveal(strings.TrimSpace(processedMd))

		// Render Markdown with Goldmark
		var buf bytes.Buffer
		if err := mdRenderer.Convert([]byte(tokenized), &buf); err != nil {
			slide.HTML = html.EscapeString(tokenized)
		} else {
			slide.HTML = strings.TrimSpace(buf.String())
		}

		slides = append(slides, slide)
	}

	return slides
}

// DeckTitle returns the text of the first heading (h1-h6) with inline tags stripped,
// falling back to fallback if none is found.
var headingRegex = regexp.MustCompile(`(?i)<h[1-6][^>]*>([\s\S]*?)</h[1-6]>`)
var htmlTagRegex = regexp.MustCompile(`<[^>]+>`)

func DeckTitle(slides []Slide, fallback string) string {
	if len(slides) > 0 {
		if m := headingRegex.FindStringSubmatch(slides[0].HTML); len(m) > 1 {
			text := strings.TrimSpace(htmlTagRegex.ReplaceAllString(m[1], ""))
			if text != "" {
				return text
			}
		}
	}
	return fallback
}

// DocTitle returns an HTML document's <title> text, falling back to fallback.
var titleTagRegex = regexp.MustCompile(`(?i)<title[^>]*>([\s\S]*?)</title>`)

func DocTitle(htmlStr string, fallback string) string {
	if m := titleTagRegex.FindStringSubmatch(htmlStr); len(m) > 1 {
		text := strings.TrimSpace(htmlTagRegex.ReplaceAllString(m[1], ""))
		if text != "" {
			return text
		}
	}
	return fallback
}
