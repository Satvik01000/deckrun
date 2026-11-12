package parser_test

import (
	"strings"
	"testing"

	"deckrun/internal/parser"
)

func TestParserHandlesSlidesNotesImagesMathAndReveals(t *testing.T) {
	md := `# Slide 1
Welcome to deckrun

<!-- notes: Introduction slide notes -->
---
# Slide 2
Here is a list:
- Point A
- Point B {reveal}

{reveal}
> Important quote

$$
E = mc^2
$$

Inline math $a^2 + b^2 = c^2$ here.
`

	slides := parser.ParseSlides(md)
	if len(slides) != 2 {
		t.Fatalf("expected 2 slides, got %d", len(slides))
	}

	if slides[0].Notes != "Introduction slide notes" {
		t.Errorf("expected notes 'Introduction slide notes', got '%s'", slides[0].Notes)
	}

	if !strings.Contains(slides[1].HTML, "deckrun-fragment-marker") {
		t.Errorf("expected fragment marker in slide 2")
	}

	if !strings.Contains(slides[1].HTML, `class="math-source"`) {
		t.Errorf("expected class=\"math-source\" in slide 2")
	}

	if !strings.Contains(slides[1].HTML, `data-display="true"`) {
		t.Errorf("expected data-display=\"true\" in slide 2")
	}

	if !strings.Contains(slides[1].HTML, `data-display="false"`) {
		t.Errorf("expected data-display=\"false\" in slide 2")
	}
}

func TestParserFenceAwareSlideSplit(t *testing.T) {
	md := `# Slide with Mermaid
` + "```mermaid" + `
stateDiagram-v2
  [*] --> Active
  --
  Active --> [*]
  ---
  state Concurrent {
    [*] --> Sub1
    --
    Sub1 --> [*]
  }
` + "```" + `

---
# Second Slide
Final point
`

	slides := parser.ParseSlides(md)
	if len(slides) != 2 {
		t.Fatalf("expected exactly 2 slides, but got %d (Mermaid --- split the slide incorrectly!)", len(slides))
	}

	if !strings.Contains(slides[0].HTML, "Concurrent") {
		t.Errorf("first slide should contain Mermaid with Concurrent state")
	}
	if !strings.Contains(slides[1].HTML, "Second Slide") {
		t.Errorf("second slide title mismatch")
	}
}

func TestPositionedImages(t *testing.T) {
	md := `# Slide With Images
![Right Pic](pic1.png "right opacity:0.7")
![Left Pic](pic2.png "left opacity=0.5")
![Background Pic](pic3.png "bg opacity:0.3")
Body text
`

	slides := parser.ParseSlides(md)
	if len(slides) != 1 {
		t.Fatalf("expected 1 slide, got %d", len(slides))
	}

	s := slides[0]
	if s.RightImage == nil || s.RightImage.Src != "pic1.png" || s.RightImage.Opacity != 0.7 {
		t.Errorf("RightImage mismatch: %+v", s.RightImage)
	}
	if s.LeftImage == nil || s.LeftImage.Src != "pic2.png" || s.LeftImage.Opacity != 0.5 {
		t.Errorf("LeftImage mismatch: %+v", s.LeftImage)
	}
	if s.BGImage == nil || s.BGImage.Src != "pic3.png" || s.BGImage.Opacity != 0.3 {
		t.Errorf("BGImage mismatch: %+v", s.BGImage)
	}
	if strings.Contains(s.HTML, "pic1.png") || strings.Contains(s.HTML, "pic2.png") || strings.Contains(s.HTML, "pic3.png") {
		t.Errorf("positioned images should be extracted from slide HTML")
	}
	if !strings.Contains(s.HTML, "Body text") {
		t.Errorf("body text missing")
	}
}

func TestDeckTitle(t *testing.T) {
	slides := parser.ParseSlides("# My Great Presentation\nContent")
	title := parser.DeckTitle(slides, "Fallback")
	if title != "My Great Presentation" {
		t.Errorf("expected 'My Great Presentation', got '%s'", title)
	}
}

func TestCurrencyVersusInlineMath(t *testing.T) {
	slides := parser.ParseSlides("Costs $5 and $10 here. Formula $x + y = z$.")
	if len(slides) != 1 {
		t.Fatalf("expected 1 slide")
	}
	if strings.Contains(slides[0].HTML, `class="math-source">5 and`) {
		t.Errorf("currency was converted to math")
	}
	if !strings.Contains(slides[0].HTML, `<span class="math-source" data-display="false">x + y = z</span>`) {
		t.Errorf("formula not converted to inline math: %s", slides[0].HTML)
	}
}

func TestCodeSpansAndFencedImagesAndComplexMath(t *testing.T) {
	md := "# Edge Cases\n" +
		"Inline code `$x$` and `{reveal}` vs $2 *x* y$ and $\\text{foo}_1 + \\text{bar}_2$\n\n" +
		"```markdown\n" +
		"![Diagram](diagram.png \"right\")\n" +
		"![Cover](cover.png \"bg opacity:0.25\")\n" +
		"```\n"

	slides := parser.ParseSlides(md)
	if len(slides) != 1 {
		t.Fatalf("expected 1 slide, got %d", len(slides))
	}
	s := slides[0]
	if s.RightImage != nil || s.BGImage != nil {
		t.Errorf("images inside fenced code block should not be extracted as slide positioned images")
	}
	if !strings.Contains(s.HTML, "<code>$x$</code>") {
		t.Errorf("inline code $x$ should stay literal: %s", s.HTML)
	}
	if !strings.Contains(s.HTML, "<code>{reveal}</code>") {
		t.Errorf("inline code {reveal} should stay literal: %s", s.HTML)
	}
	if !strings.Contains(s.HTML, `<span class="math-source" data-display="false">2 *x* y</span>`) {
		t.Errorf("math with * should not be turned into <em>: %s", s.HTML)
	}
	if !strings.Contains(s.HTML, `<span class="math-source" data-display="false">\text{foo}_1 + \text{bar}_2</span>`) {
		t.Errorf("math with _ should not be turned into <em>: %s", s.HTML)
	}
}

