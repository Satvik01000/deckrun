package render_test

import (
	"strings"
	"testing"

	"deckrun/internal/parser"
	"deckrun/internal/render"
	"deckrun/internal/theme"
)

func TestGenerateHTML(t *testing.T) {
	slides := parser.ParseSlides("# Slide 1\nHello\n<!-- notes: Test note -->")
	html := render.GenerateHTML(
		slides,
		"My Test Deck",
		false,
		"nord",
		render.FontOverrides{},
		render.PresentationOptions{Template: "editorial", Transition: "fade"},
	)

	if !strings.Contains(html, `data-theme="nord"`) {
		t.Errorf("missing data-theme=\"nord\"")
	}
	if !strings.Contains(html, `data-template="editorial"`) {
		t.Errorf("missing data-template=\"editorial\"")
	}
	if !strings.Contains(html, `data-transition="fade"`) {
		t.Errorf("missing data-transition=\"fade\"")
	}
	if !strings.Contains(html, "Test note") {
		t.Errorf("missing notes in deck JSON")
	}
}

func TestGenerateDocHTML(t *testing.T) {
	docHTML := render.GenerateDocHTML("/__remote-doc", "My Remote Doc", false, "tokyo")
	if !strings.Contains(docHTML, `data-theme="tokyo"`) {
		t.Errorf("missing data-theme=\"tokyo\"")
	}
	if !strings.Contains(docHTML, `src="/__remote-doc"`) {
		t.Errorf("missing src=\"/__remote-doc\"")
	}
	if !strings.Contains(docHTML, "My Remote Doc") {
		t.Errorf("missing title")
	}
}

func TestEditorBootstrap(t *testing.T) {
	plain := render.GenerateEditorHTML("nord", render.FontOverrides{}, "classic", "slide", nil)
	if !strings.Contains(plain, `"file":null`) {
		t.Errorf("expected file:null in plain editor bootstrap")
	}

	backed := render.GenerateEditorHTML("nord", render.FontOverrides{}, "classic", "slide", &render.EditorFileInfo{
		Name:     "slides.md",
		Kind:     "markdown",
		Writable: true,
		Watched:  true,
	})
	if !strings.Contains(backed, `"name":"slides.md"`) {
		t.Errorf("missing name in backed editor")
	}
	if !strings.Contains(backed, `"writable":true`) {
		t.Errorf("missing writable:true")
	}
	if !strings.Contains(backed, `"watched":true`) {
		t.Errorf("missing watched:true")
	}
	if !strings.Contains(backed, "/__file") {
		t.Errorf("missing /__file endpoint")
	}
	if !strings.Contains(backed, "EventSource('/__events')") {
		t.Errorf("missing EventSource('/__events')")
	}
}

func TestGeneratePreviewHTML(t *testing.T) {
	html := render.GeneratePreviewHTML("gruvbox", render.FontOverrides{}, "minimal", "zoom")
	if !strings.Contains(html, `data-theme="gruvbox"`) {
		t.Errorf("missing data-theme=\"gruvbox\"")
	}
	if !strings.Contains(html, `data-template="minimal"`) {
		t.Errorf("missing data-template=\"minimal\"")
	}
	if !strings.Contains(html, `data-transition="zoom"`) {
		t.Errorf("missing data-transition=\"zoom\"")
	}
	if !strings.Contains(html, `id="presentation"`) {
		t.Errorf("missing id=\"presentation\"")
	}
}

func TestGenerateEditorHTML(t *testing.T) {
	html := render.GenerateEditorHTML("dracula", render.FontOverrides{Head: "inter", Body: "lora"}, "spotlight", "lift", nil)
	if !strings.Contains(html, `data-theme="dracula"`) {
		t.Errorf("missing data-theme=\"dracula\"")
	}
	if !strings.Contains(html, `data-template="spotlight"`) {
		t.Errorf("missing data-template=\"spotlight\"")
	}
	if !strings.Contains(html, `data-transition="lift"`) {
		t.Errorf("missing data-transition=\"lift\"")
	}
	if !strings.Contains(html, "deckrun · editor") {
		t.Errorf("missing editor title")
	}
}

func TestRichContentFeaturesAndHead(t *testing.T) {
	slidesWithoutRich := parser.ParseSlides("# Simple\nHello")
	feat1 := render.RichContentFeatures(slidesWithoutRich)
	if feat1.Math || feat1.Mermaid {
		t.Errorf("expected no math or mermaid")
	}
	if render.RichContentHead(feat1, "local") != "" {
		t.Errorf("expected empty head for no rich content")
	}

	slidesWithRich := parser.ParseSlides("# Math\n$$\nx = y\n$$\n```mermaid\ngraph TD; A-->B;\n```")
	feat2 := render.RichContentFeatures(slidesWithRich)
	if !feat2.Math || !feat2.Mermaid {
		t.Errorf("expected math and mermaid to be true")
	}

	localHead := render.RichContentHead(feat2, "local")
	if !strings.Contains(localHead, "/__vendor/katex.min.css") || !strings.Contains(localHead, "/__vendor/mermaid.min.js") {
		t.Errorf("missing local vendor in head: %s", localHead)
	}

	cdnHead := render.RichContentHead(feat2, "cdn")
	if !strings.Contains(cdnHead, "cdn.jsdelivr.net/npm/katex") || !strings.Contains(cdnHead, "cdn.jsdelivr.net/npm/mermaid") {
		t.Errorf("missing CDN vendor in head: %s", cdnHead)
	}
}

func TestSessionHighlightsShipOnSurfaces(t *testing.T) {
	slides := parser.ParseSlides("# Slide 1\nHighlight me")
	deck := render.GenerateHTML(slides, "Deck", false, "nord", render.FontOverrides{}, render.PresentationOptions{})

	if !strings.Contains(deck, "window.deckrunHighlights") {
		t.Errorf("deck missing window.deckrunHighlights")
	}
	if !strings.Contains(deck, "scopes: 'slides'") {
		t.Errorf("deck missing scopes: 'slides'")
	}
	if !strings.Contains(deck, "get('hl')") {
		t.Errorf("deck missing get('hl')")
	}

	doc := render.GenerateDocHTML("/?deck=1", "Doc", false, "nord")
	if !strings.Contains(doc, "window.deckrunHighlights") {
		t.Errorf("doc missing window.deckrunHighlights")
	}
	if !strings.Contains(doc, "scopes: 'doc'") {
		t.Errorf("doc missing scopes: 'doc'")
	}

	editor := render.GenerateEditorHTML("nord", render.FontOverrides{}, "classic", "slide", nil)
	if !strings.Contains(editor, "frame: frame") || !strings.Contains(editor, "frame: frameHtml") {
		t.Errorf("editor missing frame highlight links")
	}
	if !strings.Contains(editor, "'&hl=' + encodeURIComponent(highlightKey())") {
		t.Errorf("editor missing highlightKey query param")
	}
}

func TestHighlightColorsAreFixed(t *testing.T) {
	for _, id := range theme.ThemeIDs {
		if strings.Contains(theme.ThemeRootCSS(id), "--mark") {
			t.Errorf("%s defines a highlight color --mark", id)
		}
	}
	if strings.Contains(render.HighlightRuntime, "var(--mark") {
		t.Errorf("HighlightRuntime reaches for --mark")
	}
	if !strings.Contains(render.HighlightRuntime, "background: #ffe75e") {
		t.Errorf("missing yellow pen color")
	}
	if !strings.Contains(render.HighlightRuntime, "mark.dr-hl * { color: #15161a !important; }") {
		t.Errorf("missing ink override")
	}
	if !strings.Contains(render.HighlightRuntime, "background: #7ae6ff") {
		t.Errorf("missing cyan comment pen color")
	}
}

func TestTypeSizeOptionIsGone(t *testing.T) {
	slides := parser.ParseSlides("# One\nBody")
	deck := render.GenerateHTML(slides, "Deck", false, "nord", render.FontOverrides{}, render.PresentationOptions{})
	preview := render.GeneratePreviewHTML("nord", render.FontOverrides{}, "classic", "slide")
	editor := render.GenerateEditorHTML("nord", render.FontOverrides{}, "classic", "slide", nil)

	for name, h := range map[string]string{"deck": deck, "preview": preview, "editor": editor} {
		if strings.Contains(h, "data-size") {
			t.Errorf("%s contains data-size", name)
		}
		if strings.Contains(h, "--type-display") {
			t.Errorf("%s contains --type-display", name)
		}
		if strings.Contains(h, "--type-body") {
			t.Errorf("%s contains --type-body", name)
		}
	}

	if strings.Contains(editor, "sz-btn") {
		t.Errorf("editor contains sz-btn")
	}
	if strings.Contains(editor, "th-size") {
		t.Errorf("editor contains th-size")
	}
	if strings.Contains(editor, `"sizes"`) {
		t.Errorf("editor contains sizes")
	}
}

func TestNoUninterpolatedTemplateLiteralsOrCorruptedRegexes(t *testing.T) {
	slides := parser.ParseSlides("# One\nBody")
	surfaces := map[string]string{
		"deck":    render.GenerateHTML(slides, "Deck", false, "nord", render.FontOverrides{}, render.PresentationOptions{}),
		"doc":     render.GenerateDocHTML("/?deck=1", "Doc", false, "nord"),
		"preview": render.GeneratePreviewHTML("nord", render.FontOverrides{}, "classic", "slide"),
		"editor":  render.GenerateEditorHTML("nord", render.FontOverrides{}, "classic", "slide", nil),
	}

	for name, htmlStr := range surfaces {
		for _, forbidden := range []string{"${PREVIEW_WIDTH}", "${PREVIEW_HEIGHT}", "${theme}", "${hljsMapJson()}", "${decorMapJson()}", "${JSON.stringify(FONT_IDS)}"} {
			if strings.Contains(htmlStr, forbidden) {
				t.Errorf("%s contains un-interpolated JS template literal '%s'", name, forbidden)
			}
		}
		if strings.Contains(htmlStr, `'\\201C'`) {
			t.Errorf("%s contains double-escaped CSS unicode quote '\\\\201C'", name)
		}
		if strings.Contains(htmlStr, `/!\\[`) {
			t.Errorf("%s contains double-escaped JS regex '/!\\\\['", name)
		}
	}
}

