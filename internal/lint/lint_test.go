package lint_test

import (
	"testing"

	"deckrun/internal/lint"
)

func TestLintCatchesErrorsAndWarnings(t *testing.T) {
	// Empty deck
	emptyRes := lint.LintMarkdown("   ")
	if emptyRes.Errors != 1 {
		t.Errorf("expected 1 error for empty deck, got %d", emptyRes.Errors)
	}
	if len(emptyRes.Issues) == 0 || emptyRes.Issues[0].Rule != "empty-deck" {
		t.Errorf("expected empty-deck rule, got %+v", emptyRes.Issues)
	}

	// Valid deck
	validMd := `# Clean Slide
- Item 1
- Item 2
`
	validRes := lint.LintMarkdown(validMd)
	if validRes.Errors != 0 || validRes.Warnings != 0 {
		t.Errorf("expected 0 errors and warnings, got e:%d w:%d", validRes.Errors, validRes.Warnings)
	}

	// Deck with untagged fence and unclosed fence
	unclosedMd := `# Slide
` + "```" + `
unclosed code
`
	unclosedRes := lint.LintMarkdown(unclosedMd)
	hasUntagged := false
	hasUnclosed := false
	for _, i := range unclosedRes.Issues {
		if i.Rule == "untagged-code-fence" {
			hasUntagged = true
		}
		if i.Rule == "unclosed-code-fence" {
			hasUnclosed = true
		}
	}
	if !hasUntagged || !hasUnclosed {
		t.Errorf("expected untagged and unclosed fence rules")
	}
	if unclosedRes.Errors != 1 {
		t.Errorf("expected 1 error, got %d", unclosedRes.Errors)
	}

	// Deck with empty alt image and invalid opacity
	badImgMd := `# Slide
![](pic.png "opacity=2.5")
`
	badImgRes := lint.LintMarkdown(badImgMd)
	hasMissingAlt := false
	hasInvalidOpacity := false
	for _, i := range badImgRes.Issues {
		if i.Rule == "missing-image-alt" {
			hasMissingAlt = true
		}
		if i.Rule == "invalid-image-opacity" {
			hasInvalidOpacity = true
		}
	}
	if !hasMissingAlt || !hasInvalidOpacity {
		t.Errorf("expected missing-image-alt and invalid-image-opacity rules")
	}

	// Complex bad deck
	complexBadMd := `# Slide with an extraordinarily long heading that exceeds eighty characters in length by a substantial margin
$$
x + y = z
---
# Dense slide
- A
- B
- C
- D
- E
- F
- G
- H
- I
{reveal} 1 {reveal} 2 {reveal} 3 {reveal} 4 {reveal} 5 {reveal} 6 {reveal} 7 {reveal} 8 {reveal} 9 {reveal} 10 {reveal} 11
`
	complexRes := lint.LintMarkdown(complexBadMd)
	rulesFound := make(map[string]bool)
	for _, i := range complexRes.Issues {
		rulesFound[i.Rule] = true
	}

	for _, expectedRule := range []string{"unclosed-math", "long-heading", "dense-slide", "reveal-excessive"} {
		if !rulesFound[expectedRule] {
			t.Errorf("missing expected lint rule %s, found: %+v", expectedRule, rulesFound)
		}
	}
}

