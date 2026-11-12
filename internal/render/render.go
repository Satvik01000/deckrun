package render

import (
	"encoding/json"
	"fmt"
	"html"
	"strings"

	"deckrun/internal/parser"
	"deckrun/internal/theme"
)

type FontOverrides struct {
	Head string `json:"head,omitempty"`
	Body string `json:"body,omitempty"`
}

type PresentationOptions struct {
	Template   string `json:"template,omitempty"`
	Transition string `json:"transition,omitempty"`
	Standalone bool   `json:"standalone,omitempty"`
}

func escAttr(str string) string {
	return html.EscapeString(str)
}

func RenderSlide(slide parser.Slide, index int) string {
	bgStyle := ""
	if slide.BGImage != nil {
		bgStyle = fmt.Sprintf(` style="--slide-bg-url: url('%s'); --slide-bg-opacity: %v;"`, escAttr(slide.BGImage.Src), slide.BGImage.Opacity)
	}

	bgLayer := ""
	if slide.BGImage != nil {
		bgLayer = `<div class="slide__bg" aria-hidden="true"></div>`
	}

	var innerHTML string
	if slide.RightImage != nil {
		innerHTML = fmt.Sprintf(`
      <div class="slide__split">
        <div class="slide__content">%s</div>
        <div class="slide__image-panel" style="opacity:%v">
          <img src="%s" alt="%s" />
        </div>
      </div>`, slide.HTML, slide.RightImage.Opacity, escAttr(slide.RightImage.Src), escAttr(slide.RightImage.Alt))
	} else if slide.LeftImage != nil {
		innerHTML = fmt.Sprintf(`
      <div class="slide__split slide__split--left-image">
        <div class="slide__image-panel" style="opacity:%v">
          <img src="%s" alt="%s" />
        </div>
        <div class="slide__content">%s</div>
      </div>`, slide.LeftImage.Opacity, escAttr(slide.LeftImage.Src), escAttr(slide.LeftImage.Alt), slide.HTML)
	} else {
		innerHTML = fmt.Sprintf(`<div class="slide__content">%s</div>`, slide.HTML)
	}

	hasBG := ""
	if slide.BGImage != nil {
		hasBG = " slide--has-bg"
	}

	return fmt.Sprintf(`<div class="slide%s" data-index="%d"%s>
  %s
  %s
</div>`, hasBG, index, bgStyle, bgLayer, innerHTML)
}

func GenerateHTML(
	slides []parser.Slide,
	title string,
	autoFullscreen bool,
	themeInput string,
	fonts FontOverrides,
	presentation PresentationOptions,
) string {
	th := theme.ResolveThemeName(themeInput)
	template := theme.ResolveTemplateName(presentation.Template)
	transition := theme.ResolveTransitionName(presentation.Transition)

	headFont, _ := theme.FindFont(fonts.Head)
	bodyFont, _ := theme.FindFont(fonts.Body)
	fontAttrs := ""
	if headFont != "" {
		fontAttrs += fmt.Sprintf(` data-head="%s"`, headFont)
	}
	if bodyFont != "" {
		fontAttrs += fmt.Sprintf(` data-body="%s"`, bodyFont)
	}

	slideHTMLs := make([]string, len(slides))
	for i, s := range slides {
		slideHTMLs[i] = RenderSlide(s, i)
	}
	slideHTML := strings.Join(slideHTMLs, "\n")
	total := len(slides)
	rich := RichContentFeatures(slides)

	notesList := make([]string, len(slides))
	for i, s := range slides {
		notesList[i] = s.Notes
	}
	notesBytes, _ := json.Marshal(notesList)
	notesJSON := strings.ReplaceAll(string(notesBytes), "<", `\u003c`)

	pageTitle := "deckrun"
	if title != "" {
		if strings.Contains(strings.ToLower(title), "deckrun") {
			pageTitle = title
		} else {
			pageTitle = fmt.Sprintf("%s · deckrun", title)
		}
	}

	source := "local"
	if presentation.Standalone {
		source = "cdn"
	}

	fsHint := ""
	if autoFullscreen {
		fsHint = `<div id="fs-hint">
  <div id="fs-hint__inner">Press any key or click to enter fullscreen</div>
</div>`
	}

	themeSummariesJSON, _ := json.Marshal(map[string]any{
		"themes":   theme.ThemeSummaries(),
		"hljsMap":  json.RawMessage(theme.HLJSMapJSON()),
		"decorMap": json.RawMessage(theme.DecorMapJSON()),
	})

	extraFonts := []string{}
	if headFont != "" {
		extraFonts = append(extraFonts, headFont)
	}
	if bodyFont != "" {
		extraFonts = append(extraFonts, bodyFont)
	}

	return fmt.Sprintf(`<!DOCTYPE html>
<html lang="en" data-theme="%s" data-decor="%s" data-template="%s" data-transition="%s"%s>
<head>
  <meta charset="UTF-8">
  <meta name="viewport" content="width=device-width, initial-scale=1.0">
  <title>%s</title>
  <link rel="preconnect" href="https://fonts.googleapis.com">
  <link rel="preconnect" href="https://fonts.gstatic.com" crossorigin>
  <link href="%s" rel="stylesheet">
  <link rel="stylesheet" id="hljs-theme" href="%s">
  <script src="https://cdnjs.cloudflare.com/ajax/libs/highlight.js/11.9.0/highlight.min.js"></script>
  %s
  <style>
%s

%s

%s

%s

%s

%s

%s

%s

%s

%s

%s
  </style>
</head>
<body>

<div id="backdrop" aria-hidden="true"></div>

<div id="presentation">
%s
</div>

<div id="hud">
  <div id="progress-bar"><div id="progress-fill"></div></div>
  <div id="hud-row">
    <div id="hud-tools">
      <button class="hud-btn" id="btn-laser" title="Laser pointer (L)">laser <kbd>L</kbd></button>
      <button class="hud-btn" id="btn-pen" title="Draw on the slide (D)">pen <kbd>D</kbd></button>
      <button class="hud-btn" id="btn-blank" title="Blank canvas over the slide (C)">canvas <kbd>C</kbd></button>
      <button class="hud-btn" id="btn-black" title="Black out the screen (B)">black <kbd>B</kbd></button>
      <button class="hud-btn" id="btn-theme" title="Change theme (T)">theme <kbd>T</kbd></button>
      <div id="pen-bar">
        <span id="hud-sep"></span>
        <span id="pen-swatches"></span>
        <button class="hud-btn" id="btn-erase" title="Eraser (E)">erase <kbd>E</kbd></button>
        <button class="hud-btn" id="btn-thin" title="Thinner ([)">&minus;</button>
        <span id="pen-width">4px</span>
        <button class="hud-btn" id="btn-thick" title="Thicker (])">+</button>
        <button class="hud-btn" id="btn-clear" title="Clear this slide (X)">clear <kbd>X</kbd></button>
      </div>
      <button class="hud-btn" id="btn-help" title="Show every control (?)">? controls</button>
    </div>
    <div id="hud-right">
      <a href="https://github.com/arpitbbhayani/deckrun" target="_blank" rel="noopener noreferrer" class="hud-brand" id="hud-brand" title="deckrun — Markdown presentations">powered by <span>deckrun</span></a>
      <div id="slide-counter"><span id="cur">1</span>&nbsp;/&nbsp;<span id="tot">%d</span></div>
    </div>
  </div>
</div>

<canvas id="board"></canvas>
<div id="laser" aria-hidden="true"></div>
<div id="blackout" title="Click or press B to come back"></div>

<div id="themes" role="dialog" aria-modal="true" aria-label="Theme picker">
  <div id="themes__backdrop" data-close="themes"></div>
  <div id="themes__box">
    <div id="themes__head">
      <span class="th-head__title">themes</span>
      <span class="th-head__sub">Arrow keys preview live &nbsp;·&nbsp; enter selects &nbsp;·&nbsp; esc closes</span>
      <a href="https://github.com/arpitbbhayani/deckrun" target="_blank" rel="noopener noreferrer" class="th-head__brand" title="deckrun">deckrun</a>
      <button id="themes__close" data-close="themes" title="Close (Esc)">&times;</button>
    </div>
    <div id="themes__list"></div>
    <div id="themes__foot">
      <span class="th-foot__brand">powered by <a href="https://github.com/arpitbbhayani/deckrun" target="_blank" rel="noopener noreferrer">deckrun</a></span>
      <span class="th-foot__hint">Switch themes on the fly</span>
    </div>
  </div>
</div>

<div id="help" role="dialog" aria-modal="true" aria-label="Presenter controls">
  <div id="help__backdrop" data-close="help"></div>
  <div id="help__panel">
    <button id="help__close" data-close="help" title="Close (Esc)">&times;</button>
    <div id="help__head">
      <h2>controls</h2>
      <p>press <kbd>?</kbd> any time</p>
    </div>
    <div id="help__grid"></div>
    <div id="help__foot">
      <span>Annotations live per slide and survive navigation.</span>
      <span>powered by <a href="https://github.com/arpitbbhayani/deckrun" target="_blank" rel="noopener noreferrer">deckrun</a></span>
    </div>
  </div>
</div>

<button class="nav-arrow nav-arrow--prev" id="btn-prev" title="Previous (←)">&#8592;</button>
<button class="nav-arrow nav-arrow--next" id="btn-next" title="Next (→)">&#8594;</button>

<div id="cursor"></div>

<div id="overview" class="hidden"></div>

<div id="kbd-hint">← → reveal / navigate &nbsp;·&nbsp; O overview &nbsp;·&nbsp; F fullscreen &nbsp;·&nbsp; T theme &nbsp;·&nbsp; L laser &nbsp;·&nbsp; D draw &nbsp;·&nbsp; ? controls</div>

%s

<script id="deck-notes" type="application/json">%s</script>
<script id="deck-themes" type="application/json">%s</script>

<script>
%s

%s

%s
</script>
<script>
%s
</script>
</body>
</html>`,
		th, theme.DecorOf(th), template, transition, fontAttrs,
		escAttr(pageTitle),
		theme.GoogleFontsHref(theme.ThemeIDs, extraFonts),
		theme.HLJSHref(th),
		RichContentHead(rich, source),
		ResetCSS,
		theme.ThemeSwitchableCSS(),
		theme.FontOverrideCSS(),
		SlideCSS,
		theme.TemplateCSS,
		theme.TransitionCSS,
		FragmentCSS,
		RichContentCSS,
		theme.DecorCSS,
		ChromeCSS,
		PresenterCSS,
		slideHTML,
		total,
		fsHint,
		notesJSON,
		string(themeSummariesJSON),
		FragmentRuntime,
		RichContentRuntime,
		HighlightRuntime,
		PresenterScript,
	)
}

func GenerateDocHTML(docURL string, title string, autoFullscreen bool, themeInput string) string {
	th := theme.ResolveThemeName(themeInput)

	pageTitle := "deckrun"
	if title != "" {
		if strings.Contains(strings.ToLower(title), "deckrun") {
			pageTitle = title
		} else {
			pageTitle = fmt.Sprintf("%s · deckrun", title)
		}
	}

	fsHint := ""
	if autoFullscreen {
		fsHint = `<div id="fs-hint">
  <div id="fs-hint__inner">Press any key or click to enter fullscreen</div>
</div>`
	}

	themeSummariesJSON, _ := json.Marshal(map[string]any{
		"themes":   theme.ThemeSummaries(),
		"hljsMap":  json.RawMessage(theme.HLJSMapJSON()),
		"decorMap": json.RawMessage(theme.DecorMapJSON()),
	})

	return fmt.Sprintf(`<!DOCTYPE html>
<html lang="en" data-theme="%s" data-decor="%s">
<head>
  <meta charset="UTF-8">
  <meta name="viewport" content="width=device-width, initial-scale=1.0">
  <title>%s</title>
  <link rel="preconnect" href="https://fonts.googleapis.com">
  <link rel="preconnect" href="https://fonts.gstatic.com" crossorigin>
  <link href="%s" rel="stylesheet">
  <link rel="stylesheet" id="hljs-theme" href="%s">
  <style>
%s

%s

%s

%s

%s
  </style>
</head>
<body>

<iframe id="doc-frame" src="%s" title="%s"></iframe>

<div id="hud">
  <div id="hud-row">
    <div id="hud-tools">
      <button class="hud-btn" id="btn-laser" title="Laser pointer (L)">laser <kbd>L</kbd></button>
      <button class="hud-btn" id="btn-pen" title="Draw on the doc (D)">pen <kbd>D</kbd></button>
      <button class="hud-btn" id="btn-blank" title="Blank canvas over the doc (C)">canvas <kbd>C</kbd></button>
      <button class="hud-btn" id="btn-black" title="Black out the screen (B)">black <kbd>B</kbd></button>
      <button class="hud-btn" id="btn-theme" title="Change theme (T)">theme <kbd>T</kbd></button>
      <div id="pen-bar">
        <span id="hud-sep"></span>
        <span id="pen-swatches"></span>
        <button class="hud-btn" id="btn-erase" title="Eraser (E)">erase <kbd>E</kbd></button>
        <button class="hud-btn" id="btn-thin" title="Thinner ([)">&minus;</button>
        <span id="pen-width">4px</span>
        <button class="hud-btn" id="btn-thick" title="Thicker (])">+</button>
        <button class="hud-btn" id="btn-clear" title="Clear (X)">clear <kbd>X</kbd></button>
      </div>
      <button class="hud-btn" id="btn-help" title="Show every control (?)">? controls</button>
    </div>
    <div id="hud-right">
      <a href="https://github.com/arpitbbhayani/deckrun" target="_blank" rel="noopener noreferrer" class="hud-brand" id="hud-brand" title="deckrun — Markdown presentations">powered by <span>deckrun</span></a>
    </div>
  </div>
</div>

<canvas id="board"></canvas>
<div id="laser" aria-hidden="true"></div>
<div id="blackout" title="Click or press B to come back"></div>

<div id="themes" role="dialog" aria-modal="true" aria-label="Theme picker">
  <div id="themes__backdrop" data-close="themes"></div>
  <div id="themes__box">
    <div id="themes__head">
      <span class="th-head__title">themes</span>
      <span class="th-head__sub">Arrow keys preview live &nbsp;·&nbsp; enter selects &nbsp;·&nbsp; esc closes</span>
      <a href="https://github.com/arpitbbhayani/deckrun" target="_blank" rel="noopener noreferrer" class="th-head__brand" title="deckrun">deckrun</a>
      <button id="themes__close" data-close="themes" title="Close (Esc)">&times;</button>
    </div>
    <div id="themes__list"></div>
    <div id="themes__foot">
      <span class="th-foot__brand">powered by <a href="https://github.com/arpitbbhayani/deckrun" target="_blank" rel="noopener noreferrer">deckrun</a></span>
      <span class="th-foot__hint">Switch themes on the fly</span>
    </div>
  </div>
</div>

<div id="help" role="dialog" aria-modal="true" aria-label="Presenter controls">
  <div id="help__backdrop" data-close="help"></div>
  <div id="help__panel">
    <button id="help__close" data-close="help" title="Close (Esc)">&times;</button>
    <div id="help__head">
      <h2>controls</h2>
      <p>press <kbd>?</kbd> any time</p>
    </div>
    <div id="help__grid"></div>
    <div id="help__foot">
      <span>Annotations are not saved to disk, and reset if the page reloads.</span>
      <span>powered by <a href="https://github.com/arpitbbhayani/deckrun" target="_blank" rel="noopener noreferrer">deckrun</a></span>
    </div>
  </div>
</div>

%s

<script id="deck-themes" type="application/json">%s</script>

<script>
%s
</script>
<script>
%s
</script>
</body>
</html>`,
		th, theme.DecorOf(th),
		escAttr(pageTitle),
		theme.GoogleFontsHref(theme.ThemeIDs, nil),
		theme.HLJSHref(th),
		ResetCSS,
		theme.ThemeSwitchableCSS(),
		DocCSS,
		ChromeCSS,
		PresenterCSS,
		escAttr(docURL), escAttr(title),
		fsHint,
		string(themeSummariesJSON),
		HighlightRuntime,
		DocScript,
	)
}
