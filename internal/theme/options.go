package theme

import (
	"fmt"
	"strings"
)

type TemplateName string
type TransitionName string

const (
	DefaultTemplate   TemplateName   = "classic"
	DefaultTransition TransitionName = "slide"
)

var TemplateIDs = []TemplateName{
	"classic",
	"minimal",
	"editorial",
	"spotlight",
}

var TransitionIDs = []TransitionName{
	"slide",
	"fade",
	"zoom",
	"lift",
	"none",
}

type OptionSummary struct {
	ID    string `json:"id"`
	Label string `json:"label"`
	Blurb string `json:"blurb"`
}

var TemplateSpecs = map[TemplateName]OptionSummary{
	"classic": {
		ID:    "classic",
		Label: "Classic",
		Blurb: "The original balanced deckrun layout",
	},
	"minimal": {
		ID:    "minimal",
		Label: "Minimal",
		Blurb: "Quiet surfaces, wider margins, fewer decorative treatments",
	},
	"editorial": {
		ID:    "editorial",
		Label: "Editorial",
		Blurb: "Strong rules and magazine-like reading rhythm",
	},
	"spotlight": {
		ID:    "spotlight",
		Label: "Spotlight",
		Blurb: "Centered, high-impact composition for concise keynote slides",
	},
}

var TransitionSpecs = map[TransitionName]OptionSummary{
	"slide": {
		ID:    "slide",
		Label: "Slide",
		Blurb: "Horizontal sliding transition between slides",
	},
	"fade": {
		ID:    "fade",
		Label: "Fade",
		Blurb: "Cross-fade between slides",
	},
	"zoom": {
		ID:    "zoom",
		Label: "Zoom",
		Blurb: "Scale up and down between slides",
	},
	"lift": {
		ID:    "lift",
		Label: "Lift",
		Blurb: "Vertical rising transition between slides",
	},
	"none": {
		ID:    "none",
		Label: "None",
		Blurb: "Instant cut between slides",
	},
}

func FindTemplate(input string) (TemplateName, bool) {
	key := strings.ToLower(strings.TrimSpace(input))
	if key == "" {
		return "", false
	}
	for _, id := range TemplateIDs {
		if string(id) == key {
			return id, true
		}
		if strings.ToLower(TemplateSpecs[id].Label) == key {
			return id, true
		}
	}
	return "", false
}

func ResolveTemplateName(input string) TemplateName {
	if t, ok := FindTemplate(input); ok {
		return t
	}
	return DefaultTemplate
}

func TemplateSummaries() []OptionSummary {
	res := make([]OptionSummary, len(TemplateIDs))
	for i, id := range TemplateIDs {
		res[i] = TemplateSpecs[id]
	}
	return res
}

func TemplateListing() []string {
	maxPad := 0
	for _, id := range TemplateIDs {
		if len(id) > maxPad {
			maxPad = len(id)
		}
	}
	lines := make([]string, len(TemplateIDs))
	for i, id := range TemplateIDs {
		s := TemplateSpecs[id]
		lines[i] = fmt.Sprintf("%-*s  %-10s  %s", maxPad, id, s.Label, s.Blurb)
	}
	return lines
}

func FindTransition(input string) (TransitionName, bool) {
	key := strings.ToLower(strings.TrimSpace(input))
	if key == "" {
		return "", false
	}
	for _, id := range TransitionIDs {
		if string(id) == key {
			return id, true
		}
		if strings.ToLower(TransitionSpecs[id].Label) == key {
			return id, true
		}
	}
	return "", false
}

func ResolveTransitionName(input string) TransitionName {
	if t, ok := FindTransition(input); ok {
		return t
	}
	return DefaultTransition
}

func TransitionSummaries() []OptionSummary {
	res := make([]OptionSummary, len(TransitionIDs))
	for i, id := range TransitionIDs {
		res[i] = TransitionSpecs[id]
	}
	return res
}

func TransitionListing() []string {
	maxPad := 0
	for _, id := range TransitionIDs {
		if len(id) > maxPad {
			maxPad = len(id)
		}
	}
	lines := make([]string, len(TransitionIDs))
	for i, id := range TransitionIDs {
		s := TransitionSpecs[id]
		lines[i] = fmt.Sprintf("%-*s  %-8s  %s", maxPad, id, s.Label, s.Blurb)
	}
	return lines
}

const TemplateCSS = `/* ── Templates ──────────────────────────────────────────────────────────── */

/* Minimal */
:root[data-template="minimal"] {
  --slide-pad-x: 10vw;
  --slide-pad-y: 8vh;
}
:root[data-template="minimal"] #backdrop {
  opacity: 0.15;
}
:root[data-template="minimal"] .slide__content h1 {
  letter-spacing: -0.02em;
}
:root[data-template="minimal"] blockquote {
  border-left-width: 2px;
  background: transparent;
}
:root[data-template="minimal"] pre {
  border: 1px solid var(--surface0);
}

/* Editorial */
:root[data-template="editorial"] {
  --slide-pad-x: 7vw;
}
:root[data-template="editorial"] .slide__content h1 {
  border-bottom: 3px solid var(--accent);
  padding-bottom: 0.3em;
  margin-bottom: 0.6em;
}
:root[data-template="editorial"] .slide__content h2 {
  border-bottom: 1px solid var(--surface1);
  padding-bottom: 0.2em;
}
:root[data-template="editorial"] blockquote {
  border-left: 4px solid var(--accent);
  font-style: italic;
  background: var(--surface0);
}
:root[data-template="editorial"] table th {
  border-bottom: 2px solid var(--accent);
}
:root[data-template="editorial"] hr {
  border: none;
  border-top: 2px solid var(--surface2);
  margin: 2em 0;
}

/* Spotlight */
:root[data-template="spotlight"] .slide {
  justify-content: center;
  align-items: center;
  text-align: center;
}
:root[data-template="spotlight"] .slide__content {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  text-align: center;
}
:root[data-template="spotlight"] .slide__content h1 {
  font-size: calc(clamp(2.1rem, 4.6vw, 3.6rem) * 1.15);
  text-align: center;
}
:root[data-template="spotlight"] .slide__content p {
  max-width: 80%;
  margin-left: auto;
  margin-right: auto;
}
:root[data-template="spotlight"] .slide__content ul,
:root[data-template="spotlight"] .slide__content ol {
  text-align: left;
  display: inline-block;
  margin-left: auto;
  margin-right: auto;
}
:root[data-template="spotlight"] .slide__content blockquote {
  text-align: center;
  border-left: none;
  border-top: 2px solid var(--accent);
  border-bottom: 2px solid var(--accent);
  padding: 1em 2em;
  background: transparent;
}
:root[data-template="spotlight"] .slide__content pre {
  text-align: left;
}
`

const TransitionCSS = `/* ── Transitions ────────────────────────────────────────────────────────── */

/* Fade */
:root[data-transition="fade"] .slide {
  transition: opacity 0.3s ease;
  transform: none !important;
}
:root[data-transition="fade"] .slide.exit-left,
:root[data-transition="fade"] .slide.exit-right,
:root[data-transition="fade"] .slide.enter-from-left,
:root[data-transition="fade"] .slide.enter-from-right {
  transform: none !important;
  opacity: 0;
}
:root[data-transition="fade"] .slide.is-active {
  opacity: 1;
  transform: none !important;
}

/* Zoom */
:root[data-transition="zoom"] .slide {
  transform: scale(0.92);
  transition: opacity 0.35s ease, transform 0.35s cubic-bezier(0.4, 0, 0.2, 1);
}
:root[data-transition="zoom"] .slide.is-active {
  transform: scale(1);
  opacity: 1;
}
:root[data-transition="zoom"] .slide.exit-left,
:root[data-transition="zoom"] .slide.exit-right {
  transform: scale(1.08);
  opacity: 0;
}
:root[data-transition="zoom"] .slide.enter-from-left,
:root[data-transition="zoom"] .slide.enter-from-right {
  transform: scale(0.92);
  opacity: 0;
}

/* Lift */
:root[data-transition="lift"] .slide {
  transform: translateY(48px);
  transition: opacity 0.35s ease, transform 0.35s cubic-bezier(0.4, 0, 0.2, 1);
}
:root[data-transition="lift"] .slide.is-active {
  transform: translateY(0);
  opacity: 1;
}
:root[data-transition="lift"] .slide.exit-left,
:root[data-transition="lift"] .slide.exit-right {
  transform: translateY(-48px);
  opacity: 0;
}
:root[data-transition="lift"] .slide.enter-from-left,
:root[data-transition="lift"] .slide.enter-from-right {
  transform: translateY(48px);
  opacity: 0;
}

/* None */
:root[data-transition="none"] .slide,
:root[data-transition="none"] .slide.is-active,
:root[data-transition="none"] .slide.exit-left,
:root[data-transition="none"] .slide.exit-right,
:root[data-transition="none"] .slide.enter-from-left,
:root[data-transition="none"] .slide.enter-from-right {
  transition: none !important;
  transform: none !important;
}

@media (prefers-reduced-motion: reduce) {
  :root[data-transition] .slide {
    transition: opacity 0.2s linear !important;
    transform: none !important;
  }
}

@media print {
  :root[data-transition] .slide {
    transition: none !important;
    transform: none !important;
    opacity: 1 !important;
  }
}
`

