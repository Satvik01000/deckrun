package parser

import (
	"bytes"
	"fmt"
	"html"
	"net/url"
	"path/filepath"
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
	Size    string  `json:"size,omitempty"`
	Caption string  `json:"caption,omitempty"`
}

type Slide struct {
	HTML        string             `json:"html"`
	BGImage     *PositionedImage   `json:"bgImage,omitempty"`
	RightImage  *PositionedImage   `json:"rightImage,omitempty"`
	RightImages []*PositionedImage `json:"rightImages,omitempty"`
	LeftImage   *PositionedImage   `json:"leftImage,omitempty"`
	LeftImages  []*PositionedImage `json:"leftImages,omitempty"`
	Notes       string             `json:"notes,omitempty"`
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
var imgRegex = regexp.MustCompile(`!\[([^\]]*)\]\((?:<([^>]+)>|([^")\n]+?))(?:\s+"([^"]*)")?\)`)
var atImgRegex = regexp.MustCompile(`(?i)@img\(([^)\n]+)\)`)
var opacityRegex = regexp.MustCompile(`opacity[=:]?\s*([0-9]*\.?[0-9]+)`)

func isRemoteOrSpecialURL(s string) bool {
	lower := strings.ToLower(s)
	return s == "" ||
		strings.HasPrefix(lower, "http://") ||
		strings.HasPrefix(lower, "https://") ||
		strings.HasPrefix(lower, "data:") ||
		strings.HasPrefix(lower, "blob:") ||
		strings.HasPrefix(s, "#") ||
		strings.HasPrefix(s, "/__")
}

func stripLocalScheme(s string) string {
	lower := strings.ToLower(s)
	if strings.HasPrefix(lower, "@local:") {
		return strings.TrimSpace(s[len("@local:"):])
	}
	if strings.HasPrefix(lower, "local://") {
		return strings.TrimSpace(s[len("local://"):])
	}
	if strings.HasPrefix(lower, "file://") {
		return strings.TrimSpace(s[len("file://"):])
	}
	return s
}

// NormalizeLocalAssetURL rewrites local filesystem paths, @local: refs, ~/ paths,
// parent-relative paths (../), and filenames with spaces to /__local?path=... while
// keeping simple workspace-relative paths without spaces unchanged.
func NormalizeLocalAssetURL(rawSrc string) string {
	s := strings.TrimSpace(rawSrc)
	if strings.HasPrefix(s, "<") && strings.HasSuffix(s, ">") && len(s) >= 2 {
		s = strings.TrimSpace(s[1 : len(s)-1])
	}
	if isRemoteOrSpecialURL(s) {
		return s
	}

	lower := strings.ToLower(s)
	needsLocal := strings.HasPrefix(lower, "@local:") ||
		strings.HasPrefix(lower, "local://") ||
		strings.HasPrefix(lower, "file://") ||
		strings.HasPrefix(s, "~/") ||
		strings.HasPrefix(s, "/") ||
		strings.HasPrefix(s, "../") ||
		strings.Contains(s, " ")

	if !needsLocal {
		return s
	}

	clean := stripLocalScheme(s)
	return "/__local?path=" + url.QueryEscape(clean)
}

func normalizeExplicitLocalURL(rawSrc string) string {
	s := strings.TrimSpace(rawSrc)
	if strings.HasPrefix(s, "<") && strings.HasSuffix(s, ">") && len(s) >= 2 {
		s = strings.TrimSpace(s[1 : len(s)-1])
	}
	if isRemoteOrSpecialURL(s) {
		return s
	}
	clean := stripLocalScheme(s)
	return "/__local?path=" + url.QueryEscape(clean)
}

func isLikelyFileExt(ext string) bool {
	if len(ext) < 2 || len(ext) > 6 || ext[0] != '.' || strings.Contains(ext, " ") {
		return false
	}
	first := ext[1]
	if !((first >= 'a' && first <= 'z') || (first >= 'A' && first <= 'Z')) {
		return false
	}
	for i := 2; i < len(ext); i++ {
		c := ext[i]
		isAlphaNum := (c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z') || (c >= '0' && c <= '9')
		if !isAlphaNum {
			return false
		}
	}
	return true
}

func defaultAltFromPath(p string) string {
	base := filepath.Base(stripLocalScheme(strings.TrimSpace(p)))
	ext := filepath.Ext(base)
	alt := base
	if isLikelyFileExt(ext) {
		alt = strings.TrimSuffix(base, ext)
	}
	alt = strings.NewReplacer("\u202f", " ", "\u00a0", " ").Replace(alt)
	if alt == "." || alt == "/" {
		return ""
	}
	return alt
}

var cssSizeValRegex = regexp.MustCompile(`^(?:\d+(?:\.\d+)?)(?:%|px|vh|vw|rem|em)$`)
var bareDigitsRegex = regexp.MustCompile(`^\d+(?:\.\d+)?$`)

func parseImageSizeToken(token string) string {
	t := strings.ToLower(strings.TrimSpace(token))
	if t == "" {
		return ""
	}
	switch t {
	case "sm", "small":
		return "width:35%;max-height:32vh;"
	case "md", "medium":
		return "width:55%;max-height:45vh;"
	case "lg", "large":
		return "width:80%;max-height:68vh;"
	}

	if strings.HasPrefix(t, "w:") || strings.HasPrefix(t, "w=") || strings.HasPrefix(t, "width:") || strings.HasPrefix(t, "width=") {
		sep := strings.IndexAny(t, ":=")
		val := strings.TrimSpace(t[sep+1:])
		if bareDigitsRegex.MatchString(val) {
			val += "px"
		}
		if cssSizeValRegex.MatchString(val) {
			return "width:" + val + ";max-height:none;"
		}
		return ""
	}

	if strings.HasPrefix(t, "h:") || strings.HasPrefix(t, "h=") || strings.HasPrefix(t, "height:") || strings.HasPrefix(t, "height=") {
		sep := strings.IndexAny(t, ":=")
		val := strings.TrimSpace(t[sep+1:])
		if bareDigitsRegex.MatchString(val) {
			val += "px"
		}
		if cssSizeValRegex.MatchString(val) {
			return "max-height:" + val + ";width:auto;"
		}
		return ""
	}

	if cssSizeValRegex.MatchString(t) {
		if strings.HasSuffix(t, "vh") {
			return "max-height:" + t + ";width:auto;"
		}
		return "width:" + t + ";max-height:none;"
	}
	return ""
}

func buildImgStyleAttr(sizeStyle string, op float64) string {
	var parts []string
	if sizeStyle != "" {
		parts = append(parts, strings.TrimSuffix(sizeStyle, ";"))
	}
	if op < 1.0 {
		parts = append(parts, fmt.Sprintf("opacity:%v", op))
	}
	if len(parts) == 0 {
		return ""
	}
	return fmt.Sprintf(` style="%s;"`, html.EscapeString(strings.Join(parts, ";")))
}

func parseAtImgCaptionToken(seg string) (string, bool) {
	s := strings.TrimSpace(seg)
	low := strings.ToLower(s)
	for _, prefix := range []string{"caption:", "caption=", "cap:", "cap=", "label:", "label="} {
		if strings.HasPrefix(low, prefix) {
			val := strings.TrimSpace(s[len(prefix):])
			return strings.Trim(val, `"'`), true
		}
	}
	return "", false
}

func parseAtImgInside(inside string) (src, alt, pos string, op float64, sizeStyle, caption string) {
	pos = "inline"
	op = 1.0

	if !strings.Contains(inside, "|") {
		// Check for optional trailing quoted directive: @img(path "right opacity:0.8")
		reQuoted := regexp.MustCompile(`^(.*?)\s+"([^"]*)"\s*$`)
		if m := reQuoted.FindStringSubmatch(strings.TrimSpace(inside)); len(m) > 2 {
			src = strings.TrimSpace(m[1])
			pos, op, sizeStyle = parseImageDirective(m[2])
		} else {
			src = strings.TrimSpace(inside)
		}
		src = strings.Trim(src, `"'`)
		alt = defaultAltFromPath(src)
		return src, alt, pos, op, sizeStyle, caption
	}

	parts := strings.Split(inside, "|")
	src = strings.Trim(strings.TrimSpace(parts[0]), `"'`)
	for _, rawSeg := range parts[1:] {
		seg := strings.TrimSpace(rawSeg)
		if seg == "" {
			continue
		}
		lower := strings.ToLower(seg)
		if lower == "right" || lower == "left" || lower == "bg" || lower == "inline" {
			pos = lower
			continue
		}
		if capVal, ok := parseAtImgCaptionToken(seg); ok {
			caption = capVal
			continue
		}
		if sz := parseImageSizeToken(seg); sz != "" {
			sizeStyle = sz
			continue
		}
		segPos, segOp, segSize := parseImageDirective(seg)
		isDirectiveOnly := false
		if segPos != "inline" {
			pos = segPos
			isDirectiveOnly = true
		}
		if segSize != "" {
			sizeStyle = segSize
			isDirectiveOnly = true
		}
		if opacityRegex.MatchString(lower) {
			op = segOp
			isDirectiveOnly = true
		} else if strings.Contains(lower, ".") || lower == "0" || lower == "1" {
			if val, err := strconv.ParseFloat(lower, 64); err == nil && val >= 0 && val <= 1 {
				op = val
				isDirectiveOnly = true
			}
		}
		if !isDirectiveOnly && alt == "" {
			alt = strings.Trim(seg, `"'`)
		}
	}
	if alt == "" {
		if caption != "" {
			alt = caption
		} else {
			alt = defaultAltFromPath(src)
		}
	}
	return src, alt, pos, op, sizeStyle, caption
}

func parseImageDirective(title string) (position string, opacity float64, sizeStyle string) {
	position = "inline"
	opacity = 1.0

	t := strings.ToLower(strings.TrimSpace(title))
	if t == "" {
		return position, opacity, sizeStyle
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

	for _, tok := range strings.Fields(t) {
		if sz := parseImageSizeToken(tok); sz != "" {
			sizeStyle = sz
		}
	}

	return position, opacity, sizeStyle
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

		// 1. Handle @img(...) shorthand tags
		atMatches := atImgRegex.FindAllStringSubmatch(chunk, -1)
		for _, m := range atMatches {
			full := m[0]
			inside := m[1]
			src, alt, pos, op, sizeStyle, caption := parseAtImgInside(inside)
			if src == "" {
				continue
			}
			normSrc := normalizeExplicitLocalURL(src)
			img := &PositionedImage{
				Src:     normSrc,
				Alt:     alt,
				Opacity: op,
				Size:    sizeStyle,
				Caption: caption,
			}

			switch pos {
			case "bg":
				slide.BGImage = img
				chunk = strings.Replace(chunk, full, "", 1)
			case "right":
				slide.RightImage = img
				slide.RightImages = append(slide.RightImages, img)
				chunk = strings.Replace(chunk, full, "", 1)
			case "left":
				slide.LeftImage = img
				slide.LeftImages = append(slide.LeftImages, img)
				chunk = strings.Replace(chunk, full, "", 1)
			default:
				styleAttr := buildImgStyleAttr(sizeStyle, op)
				imgTag := fmt.Sprintf(`<img src="%s" alt="%s"%s />`, html.EscapeString(normSrc), html.EscapeString(alt), styleAttr)
				if caption != "" {
					imgTag = fmt.Sprintf(`<figure class="slide__figure">%s<figcaption>%s</figcaption></figure>`, imgTag, html.EscapeString(caption))
				}
				chunk = strings.Replace(chunk, full, imgTag, 1)
			}
		}

		// 2. Handle standard Markdown ![alt](src "title") images
		matches := imgRegex.FindAllStringSubmatch(chunk, -1)
		for _, m := range matches {
			full := m[0]
			alt := m[1]
			rawSrc := strings.TrimSpace(m[2])
			if rawSrc == "" && len(m) > 3 {
				rawSrc = strings.TrimSpace(m[3])
			}
			titleAttr := ""
			if len(m) > 4 {
				titleAttr = m[4]
			}
			if rawSrc == "" {
				continue
			}

			normSrc := NormalizeLocalAssetURL(rawSrc)
			pos, op, sizeStyle := parseImageDirective(titleAttr)
			img := &PositionedImage{
				Src:     normSrc,
				Alt:     alt,
				Opacity: op,
				Size:    sizeStyle,
			}

			switch pos {
			case "bg":
				slide.BGImage = img
				chunk = strings.Replace(chunk, full, "", 1)
			case "right":
				slide.RightImage = img
				slide.RightImages = append(slide.RightImages, img)
				chunk = strings.Replace(chunk, full, "", 1)
			case "left":
				slide.LeftImage = img
				slide.LeftImages = append(slide.LeftImages, img)
				chunk = strings.Replace(chunk, full, "", 1)
			default:
				if sizeStyle != "" || op < 1.0 {
					styleAttr := buildImgStyleAttr(sizeStyle, op)
					repl := fmt.Sprintf(`<img src="%s" alt="%s"%s />`, html.EscapeString(normSrc), html.EscapeString(alt), styleAttr)
					chunk = strings.Replace(chunk, full, repl, 1)
				} else if normSrc != rawSrc || m[2] != "" {
					var repl string
					if titleAttr != "" {
						repl = fmt.Sprintf(`![%s](%s "%s")`, alt, normSrc, titleAttr)
					} else {
						repl = fmt.Sprintf(`![%s](%s)`, alt, normSrc)
					}
					chunk = strings.Replace(chunk, full, repl, 1)
				}
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

var htmlMediaTagRegex = regexp.MustCompile(`(?i)(<(?:img|video|source)\b[^>]*?\bsrc\s*=\s*)(?:"([^"]*)"|'([^']*)')`)

func isLocalHTMLAssetSrc(src string) bool {
	s := strings.TrimSpace(src)
	if isRemoteOrSpecialURL(s) {
		return false
	}
	lower := strings.ToLower(s)
	return strings.HasPrefix(lower, "@local:") ||
		strings.HasPrefix(lower, "local://") ||
		strings.HasPrefix(lower, "file://") ||
		strings.HasPrefix(s, "~/") ||
		strings.HasPrefix(s, "../") ||
		strings.HasPrefix(s, "/Users/") ||
		strings.HasPrefix(s, "/home/") ||
		strings.HasPrefix(s, "/tmp/") ||
		strings.HasPrefix(s, "/var/") ||
		strings.HasPrefix(s, "/opt/") ||
		strings.HasPrefix(s, "/Volumes/")
}

// RewriteLocalAssetURLsInHTML expands @img(...) tags and rewrites local asset
// URLs on <img>, <video>, and <source> tags to /__local?path=... for HTML docs.
func RewriteLocalAssetURLsInHTML(rawHTML string) string {
	out := atImgRegex.ReplaceAllStringFunc(rawHTML, func(match string) string {
		m := atImgRegex.FindStringSubmatch(match)
		if len(m) < 2 {
			return match
		}
		src, alt, _, op, sizeStyle, caption := parseAtImgInside(m[1])
		if src == "" {
			return match
		}
		normSrc := normalizeExplicitLocalURL(src)
		style := buildImgStyleAttr(sizeStyle, op)
		imgTag := fmt.Sprintf(`<img src="%s" alt="%s"%s />`, html.EscapeString(normSrc), html.EscapeString(alt), style)
		if caption != "" {
			return fmt.Sprintf(`<figure class="slide__figure">%s<figcaption>%s</figcaption></figure>`, imgTag, html.EscapeString(caption))
		}
		return imgTag
	})

	out = htmlMediaTagRegex.ReplaceAllStringFunc(out, func(match string) string {
		m := htmlMediaTagRegex.FindStringSubmatch(match)
		if len(m) < 4 {
			return match
		}
		prefix := m[1]
		quote := `"`
		srcVal := m[2]
		if srcVal == "" && m[3] != "" {
			srcVal = m[3]
			quote = `'`
		}
		if !isLocalHTMLAssetSrc(srcVal) {
			return match
		}
		norm := NormalizeLocalAssetURL(srcVal)
		return prefix + quote + html.EscapeString(norm) + quote
	})

	return out
}


