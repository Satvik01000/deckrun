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

func TestLocalImagesAndAtImgTags(t *testing.T) {
	md := `# Slide 1
@img(~/Desktop/my screenshot.png | right | opacity:0.85 | System Diagram)
@img(/Users/test/inline pic.png)
@img(~/Desktop/Screenshot 2026-11-14 at 11.31.58 AM)
@img(~/Desktop/bg.png | bg | 0.65)
![Alt](~/Desktop/file with spaces.png "left")
`
	slides := parser.ParseSlides(md)
	if len(slides) != 1 {
		t.Fatalf("expected 1 slide, got %d", len(slides))
	}
	s := slides[0]
	if s.RightImage == nil || s.RightImage.Alt != "System Diagram" || s.RightImage.Opacity != 0.85 || !strings.HasPrefix(s.RightImage.Src, "/__local?path=") {
		t.Errorf("RightImage from @img mismatch: %+v", s.RightImage)
	}
	if s.BGImage == nil || s.BGImage.Opacity != 0.65 || s.BGImage.Alt != "bg" {
		t.Errorf("BGImage bare opacity mismatch: %+v", s.BGImage)
	}
	if s.LeftImage == nil || !strings.HasPrefix(s.LeftImage.Src, "/__local?path=") {
		t.Errorf("LeftImage from Markdown space path mismatch: %+v", s.LeftImage)
	}
	if !strings.Contains(s.HTML, `src="/__local?path=%2FUsers%2Ftest%2Finline+pic.png"`) {
		t.Errorf("inline @img tag not rendered cleanly: %s", s.HTML)
	}
	if !strings.Contains(s.HTML, `alt="Screenshot 2026-11-14 at 11.31.58 AM"`) {
		t.Errorf("extensionless timestamp alt should not be truncated: %s", s.HTML)
	}
}

func TestRewriteLocalAssetURLsInHTML(t *testing.T) {
	raw := `<div>@img(~/Pictures/hero banner.png | opacity:0.9 | Hero)
<img src="@local:~/Desktop/shot.png" alt="Shot">
<img src="/Users/alice/diagram.png">
<img src="https://example.com/remote.png"></div>`

	out := parser.RewriteLocalAssetURLsInHTML(raw)
	if !strings.Contains(out, `src="/__local?path=~%2FPictures%2Fhero+banner.png"`) || !strings.Contains(out, `alt="Hero"`) || !strings.Contains(out, `opacity:0.9`) {
		t.Errorf("@img in HTML was not rewritten cleanly: %s", out)
	}
	if !strings.Contains(out, `src="/__local?path=~%2FDesktop%2Fshot.png"`) {
		t.Errorf("@local: img in HTML was not rewritten: %s", out)
	}
	if !strings.Contains(out, `src="/__local?path=%2FUsers%2Falice%2Fdiagram.png"`) {
		t.Errorf("absolute local img path in HTML was not rewritten: %s", out)
	}
	if !strings.Contains(out, `src="https://example.com/remote.png"`) {
		t.Errorf("remote URL should remain untouched: %s", out)
	}
}

func TestMultipleImagesPerSlideAndSizeScaling(t *testing.T) {
	md := `# Multi & Scaled
@img(~/College/Notes Images/Scaling example | right | 70%)
@img(~/College/Notes Images/12 | right | h:220px)
@img(~/Desktop/inline-one.png | 50%)
![Chart](~/Desktop/chart.png "w:320px")
`
	slides := parser.ParseSlides(md)
	if len(slides) != 1 {
		t.Fatalf("expected 1 slide, got %d", len(slides))
	}
	s := slides[0]
	if len(s.RightImages) != 2 {
		t.Fatalf("expected 2 right images, got %d: %+v", len(s.RightImages), s.RightImages)
	}
	if !strings.Contains(s.RightImages[0].Src, "Scaling+example") || !strings.Contains(s.RightImages[0].Size, "width:70%") {
		t.Errorf("first right image mismatch: %+v", s.RightImages[0])
	}
	if !strings.Contains(s.RightImages[1].Src, "12") || !strings.Contains(s.RightImages[1].Size, "height:220px") {
		t.Errorf("second right image mismatch: %+v", s.RightImages[1])
	}
	if !strings.Contains(s.HTML, `style="width:50%`) {
		t.Errorf("inline @img 50%% width missing in HTML: %s", s.HTML)
	}
	if !strings.Contains(s.HTML, `style="width:320px`) {
		t.Errorf("markdown image w:320px width missing in HTML: %s", s.HTML)
	}

	htmlOut := parser.RewriteLocalAssetURLsInHTML(`<p>@img(~/Pictures/hero.png | 45%)</p>`)
	if !strings.Contains(htmlOut, `width:45%`) {
		t.Errorf("RewriteLocalAssetURLsInHTML missing size style: %s", htmlOut)
	}
}

func TestAtImgCaptions(t *testing.T) {
	md := `@img(assets/image-3.png | h:56vh | left | caption:SDLC)
@img(assets/image-4.png | h:56vh | right | label:MLLC)
@img(assets/inline.png | cap:Pipeline)
`
	slides := parser.ParseSlides(md)
	if len(slides) != 1 {
		t.Fatalf("expected 1 slide, got %d", len(slides))
	}
	s := slides[0]
	if s.LeftImage == nil || s.LeftImage.Caption != "SDLC" {
		t.Errorf("LeftImage caption mismatch: %+v", s.LeftImage)
	}
	if s.RightImage == nil || s.RightImage.Caption != "MLLC" {
		t.Errorf("RightImage caption mismatch: %+v", s.RightImage)
	}
	if !strings.Contains(s.HTML, `<figure class="slide__figure">`) || !strings.Contains(s.HTML, `<figcaption>Pipeline</figcaption>`) {
		t.Errorf("inline caption figure missing: %s", s.HTML)
	}
}



