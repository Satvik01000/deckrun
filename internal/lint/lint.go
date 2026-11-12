package lint

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
)

type Severity string

const (
	SeverityError   Severity = "error"
	SeverityWarning Severity = "warning"
)

type Issue struct {
	Rule     string   `json:"rule"`
	Severity Severity `json:"severity"`
	Message  string   `json:"message"`
	Line     int      `json:"line"`
	Column   int      `json:"column"`
	Slide    int      `json:"slide,omitempty"`
}

type Result struct {
	File     string  `json:"file,omitempty"`
	Slides   int     `json:"slides"`
	Errors   int     `json:"errors"`
	Warnings int     `json:"warnings"`
	Issues   []Issue `json:"issues"`
}

var (
	fenceRegex   = regexp.MustCompile(`^(\s*)(` + "`" + `{3,}|~{3,})(.*)$`)
	headingRegex = regexp.MustCompile(`^(\s*#{1,6}\s+)(.*)$`)
	imgRegex     = regexp.MustCompile(`!\[([^\]]*)\]\(([^\s)]+)(?:\s+"([^"]*)")?\)`)
	opacityRegex = regexp.MustCompile(`(?i)opacity[=:]?\s*([^\s"]+)`)
	bulletRegex  = regexp.MustCompile(`^\s*([-*+]|\d+[.)])\s+`)
)

type slideBoundary struct {
	slideIndex int
	startLine  int
	endLine    int
	lines      []string
}

func LintMarkdown(markdown string) Result {
	trimmed := strings.TrimSpace(markdown)
	if trimmed == "" {
		return Result{
			Slides:   0,
			Errors:   1,
			Warnings: 0,
			Issues: []Issue{
				{
					Rule:     "empty-deck",
					Severity: SeverityError,
					Message:  "The deck is empty.",
					Line:     1,
					Column:   1,
				},
			},
		}
	}

	normalized := strings.ReplaceAll(markdown, "\r\n", "\n")
	normalized = strings.ReplaceAll(normalized, "\r", "\n")
	lines := strings.Split(normalized, "\n")

	// Track slide boundaries (fence-aware)
	var slides []slideBoundary
	var curSlideLines []string
	curStartLine := 1
	slideIndex := 1
	inFenceForSplit := false
	fenceMarkerForSplit := ""

	for i, line := range lines {
		trimmedLine := strings.TrimSpace(line)

		if strings.HasPrefix(trimmedLine, "```") || strings.HasPrefix(trimmedLine, "~~~") {
			prefix := trimmedLine[:3]
			if !inFenceForSplit {
				inFenceForSplit = true
				fenceMarkerForSplit = prefix
			} else if strings.HasPrefix(trimmedLine, fenceMarkerForSplit) {
				inFenceForSplit = false
				fenceMarkerForSplit = ""
			}
		}

		if !inFenceForSplit && isDashesLine(trimmedLine) {
			slides = append(slides, slideBoundary{
				slideIndex: slideIndex,
				startLine:  curStartLine,
				endLine:    i, // 0-based index is line number i (inclusive of lines before divider)
				lines:      curSlideLines,
			})
			slideIndex++
			curStartLine = i + 2
			curSlideLines = nil
		} else {
			curSlideLines = append(curSlideLines, line)
		}
	}
	slides = append(slides, slideBoundary{
		slideIndex: slideIndex,
		startLine:  curStartLine,
		endLine:    len(lines),
		lines:      curSlideLines,
	})

	issues := []Issue{}

	// Global & Slide checks
	inCodeFence := false
	fenceStartLine := 1
	fenceStartCol := 1
	fenceMarker := ""

	inDisplayMath := false
	mathStartLine := 1
	mathStartCol := 1

	for i, line := range lines {
		lineNum := i + 1

		// Find current slide number
		currentSlide := 1
		for _, s := range slides {
			if lineNum >= s.startLine && lineNum <= s.endLine {
				currentSlide = s.slideIndex
				break
			}
		}

		// Check code fences
		fenceMatch := fenceRegex.FindStringSubmatch(line)
		if fenceMatch != nil {
			prefix := fenceMatch[2][:3]
			if !inCodeFence {
				inCodeFence = true
				fenceMarker = prefix
				fenceStartLine = lineNum
				fenceStartCol = len(fenceMatch[1]) + 1
				tag := strings.TrimSpace(fenceMatch[3])
				if tag == "" {
					issues = append(issues, Issue{
						Rule:     "untagged-code-fence",
						Severity: SeverityWarning,
						Message:  "Code fence has no language tag for syntax highlighting.",
						Line:     lineNum,
						Column:   fenceStartCol,
						Slide:    currentSlide,
					})
				}
			} else if strings.HasPrefix(fenceMatch[2], fenceMarker) {
				inCodeFence = false
				fenceMarker = ""
			}
		}

		// Checks outside code fence
		if !inCodeFence {
			trimmedLine := strings.TrimSpace(line)
			if trimmedLine == "$$" || trimmedLine == `\[` {
				if !inDisplayMath {
					inDisplayMath = true
					mathStartLine = lineNum
					mathStartCol = 1
				} else {
					inDisplayMath = false
				}
			} else if strings.Contains(line, "$$") {
				count := strings.Count(line, "$$")
				if count%2 != 0 {
					inDisplayMath = !inDisplayMath
					if inDisplayMath {
						mathStartLine = lineNum
						mathStartCol = strings.Index(line, "$$") + 1
					}
				}
			}

			// Check headings
			if hm := headingRegex.FindStringSubmatch(line); hm != nil {
				headingText := hm[2]
				if len(headingText) > 80 {
					issues = append(issues, Issue{
						Rule:     "long-heading",
						Severity: SeverityWarning,
						Message:  fmt.Sprintf("Heading is %d characters long; consider shortening for presentation readability.", len(headingText)),
						Line:     lineNum,
						Column:   len(hm[1]) + 1,
						Slide:    currentSlide,
					})
				}
			}

			// Check image directives
			matches := imgRegex.FindAllStringSubmatchIndex(line, -1)
			for _, loc := range matches {
				alt := line[loc[2]:loc[3]]
				title := ""
				if loc[6] >= 0 && loc[7] >= 0 {
					title = line[loc[6]:loc[7]]
				}
				col := loc[0] + 1

				if strings.TrimSpace(alt) == "" {
					issues = append(issues, Issue{
						Rule:     "missing-image-alt",
						Severity: SeverityWarning,
						Message:  "Image is missing alt text.",
						Line:     lineNum,
						Column:   col,
						Slide:    currentSlide,
					})
				}

				if title != "" {
					if om := opacityRegex.FindStringSubmatch(title); len(om) > 1 {
						val, err := strconv.ParseFloat(om[1], 64)
						if err != nil || val < 0 || val > 1 {
							issues = append(issues, Issue{
								Rule:     "invalid-image-opacity",
								Severity: SeverityWarning,
								Message:  fmt.Sprintf("Invalid image opacity '%s'; expected a number between 0 and 1.", om[1]),
								Line:     lineNum,
								Column:   col,
								Slide:    currentSlide,
							})
						}
					}
				}
			}
		}
	}

	if inCodeFence {
		issues = append(issues, Issue{
			Rule:     "unclosed-code-fence",
			Severity: SeverityError,
			Message:  "Code fence was opened but never closed.",
			Line:     fenceStartLine,
			Column:   fenceStartCol,
		})
	}

	if inDisplayMath {
		issues = append(issues, Issue{
			Rule:     "unclosed-math",
			Severity: SeverityError,
			Message:  "Display math block was opened but never closed.",
			Line:     mathStartLine,
			Column:   mathStartCol,
		})
	}

	// Per-slide checks
	for _, s := range slides {
		slideContent := strings.TrimSpace(strings.Join(s.lines, "\n"))
		if slideContent == "" {
			issues = append(issues, Issue{
				Rule:     "empty-slide",
				Severity: SeverityWarning,
				Message:  fmt.Sprintf("Slide %d is empty.", s.slideIndex),
				Line:     s.startLine,
				Column:   1,
				Slide:    s.slideIndex,
			})
			continue
		}

		// Bullet density
		bulletCount := 0
		for _, l := range s.lines {
			if bulletRegex.MatchString(l) {
				bulletCount++
			}
		}
		if bulletCount > 8 {
			issues = append(issues, Issue{
				Rule:     "dense-slide",
				Severity: SeverityWarning,
				Message:  fmt.Sprintf("Slide %d has %d bullets (recommended maximum is 8).", s.slideIndex, bulletCount),
				Line:     s.startLine,
				Column:   1,
				Slide:    s.slideIndex,
			})
		}

		// Reveal count
		revealCount := strings.Count(slideContent, "{reveal}")
		if revealCount > 10 {
			issues = append(issues, Issue{
				Rule:     "reveal-excessive",
				Severity: SeverityWarning,
				Message:  fmt.Sprintf("Slide %d has %d reveal markers (recommended maximum is 10).", s.slideIndex, revealCount),
				Line:     s.startLine,
				Column:   1,
				Slide:    s.slideIndex,
			})
		}
	}

	errCount := 0
	warnCount := 0
	for _, i := range issues {
		if i.Severity == SeverityError {
			errCount++
		} else {
			warnCount++
		}
	}

	return Result{
		Slides:   len(slides),
		Errors:   errCount,
		Warnings: warnCount,
		Issues:   issues,
	}
}

func isDashesLine(s string) bool {
	if len(s) < 3 {
		return false
	}
	for i := 0; i < len(s); i++ {
		if s[i] != '-' && s[i] != ' ' && s[i] != '\t' {
			return false
		}
	}
	return strings.Count(s, "-") >= 3
}

func FormatStylish(reports []Result) string {
	var sb strings.Builder
	totalErrors := 0
	totalWarnings := 0
	totalSlides := 0

	for _, r := range reports {
		totalErrors += r.Errors
		totalWarnings += r.Warnings
		totalSlides += r.Slides

		if len(r.Issues) == 0 {
			continue
		}

		sb.WriteString("\n" + r.File + "\n")
		for _, item := range r.Issues {
			pos := fmt.Sprintf("%d:%d", item.Line, item.Column)
			slideStr := ""
			if item.Slide > 0 {
				slideStr = fmt.Sprintf("slide %d · ", item.Slide)
			}
			sb.WriteString(fmt.Sprintf("  %-9s %-7s %s%s  %s\n", pos, string(item.Severity), slideStr, item.Message, item.Rule))
		}
	}

	totalIssues := totalErrors + totalWarnings
	if totalIssues == 0 {
		filePlural := "files"
		if len(reports) == 1 {
			filePlural = "file"
		}
		slidePlural := "slides"
		if totalSlides == 1 {
			slidePlural = "slide"
		}
		sb.WriteString(fmt.Sprintf("✓ %d %s, %d %s, no problems\n", len(reports), filePlural, totalSlides, slidePlural))
	} else {
		errPlural := "errors"
		if totalErrors == 1 {
			errPlural = "error"
		}
		warnPlural := "warnings"
		if totalWarnings == 1 {
			warnPlural = "warning"
		}
		probPlural := "problems"
		if totalIssues == 1 {
			probPlural = "problem"
		}
		sb.WriteString(fmt.Sprintf("\n✖ %d %s (%d %s, %d %s)\n", totalIssues, probPlural, totalErrors, errPlural, totalWarnings, warnPlural))
	}

	return sb.String()
}
