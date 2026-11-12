package parser

import (
	"bytes"
	"fmt"
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
// while preserving code fences and inline code spans so math/reveal inside code
// is not touched, and holding math HTML in placeholders so Goldmark does not
// mangle *, _, or \ inside LaTeX formulas.
func TokenizeMathAndReveal(src string) (string, []string) {
	lines := strings.Split(src, "\n")
	var out []string
	var held []string

	hold := func(tag, htmlStr string) string {
		idx := len(held)
		held = append(held, htmlStr)
		return fmt.Sprintf(`<%s data-dr-hold="%d"></%s>`, tag, idx, tag)
	}

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
			rest := strings.TrimSpace(trimmed[2:])
			if strings.HasSuffix(rest, "$$") && len(rest) >= 2 {
				formula := strings.TrimSuffix(rest, "$$")
				blockHTML := `<div class="math-source" data-display="true">` + html.EscapeString(strings.TrimSpace(formula)) + `</div>`
				out = append(out, hold("div", blockHTML), "")
				continue
			}

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
				blockHTML := `<div class="math-source" data-display="true">` + html.EscapeString(strings.TrimSpace(formula)) + `</div>`
				out = append(out, hold("div", blockHTML), "")
				continue
			}
		}

		// Handle \[ ... \]
		if strings.HasPrefix(trimmed, `\[`) {
			rest := strings.TrimSpace(trimmed[2:])
			if strings.HasSuffix(rest, `\]`) && len(rest) >= 2 {
				formula := strings.TrimSuffix(rest, `\]`)
				blockHTML := `<div class="math-source" data-display="true">` + html.EscapeString(strings.TrimSpace(formula)) + `</div>`
				out = append(out, hold("div", blockHTML), "")
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
				blockHTML := `<div class="math-source" data-display="true">` + html.EscapeString(strings.TrimSpace(formula)) + `</div>`
				out = append(out, hold("div", blockHTML), "")
				continue
			}
		}

		// Process inline code spans, {reveal} markers, and inline math ($...$ and \(...\))
		line = tokenizeInlineLine(line, func(htmlStr string) string {
			return hold("span", htmlStr)
		})

		out = append(out, line)
	}

	return strings.Join(out, "\n"), held
}

func restoreHeldTokens(rendered string, held []string) string {
	for idx, h := range held {
		spanPlaceholder := fmt.Sprintf(`<span data-dr-hold="%d"></span>`, idx)
		divPlaceholder := fmt.Sprintf(`<div data-dr-hold="%d"></div>`, idx)
		rendered = strings.ReplaceAll(rendered, spanPlaceholder, h)
		rendered = strings.ReplaceAll(rendered, divPlaceholder, h)
	}
	return rendered
}

func tokenizeInlineLine(line string, holdInline func(string) string) string {
	var sb strings.Builder
	n := len(line)
	i := 0

	for i < n {
		// 1. Skip inline code spans (`...` or ``...``) so math/reveal inside backticks stay literal
		if line[i] == '`' {
			tickCount := 0
			for i+tickCount < n && line[i+tickCount] == '`' {
				tickCount++
			}
			ticks := line[i : i+tickCount]
			closeIdx := -1
			searchFrom := i + tickCount
			for searchFrom < n {
				rel := strings.Index(line[searchFrom:], ticks)
				if rel < 0 {
					break
				}
				cand := searchFrom + rel
				// Ensure exact tickCount match (not part of a longer run of backticks)
				afterCand := cand + tickCount
				if afterCand < n && line[afterCand] == '`' {
					for afterCand < n && line[afterCand] == '`' {
						afterCand++
					}
					searchFrom = afterCand
					continue
				}
				closeIdx = cand
				break
			}
			if closeIdx != -1 {
				sb.WriteString(line[i : closeIdx+tickCount])
				i = closeIdx + tickCount
				continue
			}
			sb.WriteString(ticks)
			i += tickCount
			continue
		}

		// 2. Check for {reveal}
		if strings.HasPrefix(line[i:], "{reveal}") {
			sb.WriteString(holdInline(`<span class="deckrun-fragment-marker" aria-hidden="true"></span>`))
			i += len("{reveal}")
			continue
		}

		// 3. Check for \( ... \) inline math
		if strings.HasPrefix(line[i:], `\(`) {
			end := -1
			for j := i + 2; j+1 < n; j++ {
				if line[j] == '\\' && line[j+1] == ')' {
					end = j
					break
				}
			}
			if end != -1 {
				formula := line[i+2 : end]
				sb.WriteString(holdInline(`<span class="math-source" data-display="false">` + html.EscapeString(formula) + `</span>`))
				i = end + 2
				continue
			}
		}

		// 4. Check for $ ... $ inline math
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
							break
						}
						// Must not be followed by $ (e.g. $$)
						if j+1 < n && line[j+1] == '$' {
							break
						}
						end = j
						break
					} else {
						break
					}
				}
			}

			if end != -1 {
				formula := line[i+1 : end]
				sb.WriteString(holdInline(`<span class="math-source" data-display="false">` + html.EscapeString(formula) + `</span>`))
				i = end + 1
				continue
			}
		}

		sb.WriteByte(line[i])
		i++
	}

	return sb.String()
}

// stripNotesAndPositionedImages extracts speaker notes and positioned images
// only from content outside fenced code blocks, so code examples in slides
// (like the Welcome Deck's markdown image syntax example) are never stripped.
func extractNotesAndPositionedImages(raw string, slide *Slide) string {
	lines := strings.Split(raw, "\n")
	var outLines []string
	var nonFenceBuf []string

	flushNonFence := func() {
		if len(nonFenceBuf) == 0 {
			return
		}
		chunk := strings.Join(nonFenceBuf, "\n")
		nonFenceBuf = nil

		if m := notesRegex.FindStringSubmatch(chunk); len(m) > 1 {
			slide.Notes = strings.TrimSpace(m[1])
		}
		chunk = notesRegex.ReplaceAllString(chunk, "")

		matches := imgRegex.FindAllStringSubmatch(chunk, -1)
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
				chunk = strings.Replace(chunk, full, "", 1)
			case "right":
				slide.RightImage = img
				chunk = strings.Replace(chunk, full, "", 1)
			case "left":
				slide.LeftImage = img
				chunk = strings.Replace(chunk, full, "", 1)
			}
		}
		outLines = append(outLines, chunk)
	}

	inFence := false
	fenceMarker := ""

	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if strings.HasPrefix(trimmed, "```") || strings.HasPrefix(trimmed, "~~~") {
			prefix := trimmed[:3]
			if !inFence {
				flushNonFence()
				inFence = true
				fenceMarker = prefix
				outLines = append(outLines, line)
				continue
			} else if strings.HasPrefix(trimmed, fenceMarker) {
				inFence = false
				fenceMarker = ""
				outLines = append(outLines, line)
				continue
			}
		}
		if inFence {
			outLines = append(outLines, line)
		} else {
			nonFenceBuf = append(nonFenceBuf, line)
		}
	}
	flushNonFence()

	return strings.Join(outLines, "\n")
}

// ParseSlides parses a markdown string into slides with positioned images, notes,
// and rendered HTML.
func ParseSlides(markdown string) []Slide {
	rawSlides := SplitSlides(markdown)
	slides := make([]Slide, 0, len(rawSlides))

	for _, raw := range rawSlides {
		slide := Slide{}

		processedMd := extractNotesAndPositionedImages(raw, &slide)

		// Tokenize math & reveal markers before Markdown conversion
		tokenized, held := TokenizeMathAndReveal(strings.TrimSpace(processedMd))

		// Render Markdown with Goldmark
		var buf bytes.Buffer
		if err := mdRenderer.Convert([]byte(tokenized), &buf); err != nil {
			slide.HTML = html.EscapeString(tokenized)
		} else {
			slide.HTML = restoreHeldTokens(strings.TrimSpace(buf.String()), held)
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
				return html.UnescapeString(text)
			}
		}
	}
	return fallback
}

var docTitleRegex = regexp.MustCompile(`(?i)<title[^>]*>([\s\S]*?)</title>`)

func DocTitle(rawHTML string, fallback string) string {
	if m := docTitleRegex.FindStringSubmatch(rawHTML); len(m) > 1 {
		text := strings.TrimSpace(htmlTagRegex.ReplaceAllString(m[1], ""))
		if text != "" {
			return html.UnescapeString(text)
		}
	}
	if m := headingRegex.FindStringSubmatch(rawHTML); len(m) > 1 {
		text := strings.TrimSpace(htmlTagRegex.ReplaceAllString(m[1], ""))
		if text != "" {
			return html.UnescapeString(text)
		}
	}
	return fallback
}

