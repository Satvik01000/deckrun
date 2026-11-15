package theme_test

import (
	"strings"
	"testing"

	"deckrun/internal/theme"
)

func TestThemeResolvers(t *testing.T) {
	if res := theme.ResolveThemeName("nord"); res != "nord" {
		t.Errorf("expected nord, got %s", res)
	}
	if res := theme.ResolveThemeName("Nord"); res != "nord" {
		t.Errorf("expected nord, got %s", res)
	}
	if res := theme.ResolveThemeName("invalid-theme"); res != "nord" {
		t.Errorf("expected nord for invalid, got %s", res)
	}
	if res := theme.ResolveThemeName("mocha"); res != "midnight" {
		t.Errorf("expected midnight for mocha alias, got %s", res)
	}

	if f, ok := theme.FindFont("inter"); !ok || f != "inter" {
		t.Errorf("expected inter font, got %s, %v", f, ok)
	}
	if f, ok := theme.FindFont("Inter"); !ok || f != "inter" {
		t.Errorf("expected inter font, got %s, %v", f, ok)
	}
	if f, ok := theme.FindFont("Space Grotesk"); !ok || f != "spaceGrotesk" {
		t.Errorf("expected spaceGrotesk font, got %s, %v", f, ok)
	}
	if _, ok := theme.FindFont("non-existent-font"); ok {
		t.Errorf("expected non-existent-font to fail")
	}
}

func TestPresentationOptions(t *testing.T) {
	if tm, ok := theme.FindTemplate("minimal"); !ok || tm != "minimal" {
		t.Errorf("expected minimal, got %s", tm)
	}
	if tm, ok := theme.FindTemplate("Classic"); !ok || tm != "classic" {
		t.Errorf("expected classic, got %s", tm)
	}
	if _, ok := theme.FindTemplate("non-existent"); ok {
		t.Errorf("expected non-existent to fail")
	}
	if tm := theme.ResolveTemplateName("spotlight"); tm != "spotlight" {
		t.Errorf("expected spotlight, got %s", tm)
	}
	if tm := theme.ResolveTemplateName("invalid"); tm != "classic" {
		t.Errorf("expected classic for invalid, got %s", tm)
	}

	if tr, ok := theme.FindTransition("fade"); !ok || tr != "fade" {
		t.Errorf("expected fade, got %s", tr)
	}
	if tr, ok := theme.FindTransition("Zoom"); !ok || tr != "zoom" {
		t.Errorf("expected zoom, got %s", tr)
	}
	if _, ok := theme.FindTransition("non-existent"); ok {
		t.Errorf("expected non-existent transition to fail")
	}
	if tr := theme.ResolveTransitionName("lift"); tr != "lift" {
		t.Errorf("expected lift, got %s", tr)
	}
	if tr := theme.ResolveTransitionName("invalid"); tr != "slide" {
		t.Errorf("expected slide for invalid, got %s", tr)
	}

	if len(theme.TemplateSummaries()) != 4 {
		t.Errorf("expected 4 template summaries")
	}
	if len(theme.TransitionSummaries()) != 5 {
		t.Errorf("expected 5 transition summaries")
	}
	if len(theme.TemplateListing()) != 4 {
		t.Errorf("expected 4 template listings")
	}
	if len(theme.TransitionListing()) != 5 {
		t.Errorf("expected 5 transition listings")
	}

	if !strings.Contains(theme.TemplateCSS, "spotlight") {
		t.Errorf("TemplateCSS missing spotlight")
	}
	if !strings.Contains(theme.TransitionCSS, "lift") {
		t.Errorf("TransitionCSS missing lift")
	}
}

func TestCatppuccinAndOneDarkPro(t *testing.T) {
	flavors := map[string]struct {
		mood  theme.Mood
		base  string
		text  string
		mauve string
	}{
		"catppuccin-latte":     {mood: theme.MoodLight, base: "#eff1f5", text: "#4c4f69", mauve: "#8839ef"},
		"catppuccin-frappe":    {mood: theme.MoodDark, base: "#303446", text: "#c6d0f5", mauve: "#ca9ee6"},
		"catppuccin-macchiato": {mood: theme.MoodDark, base: "#24273a", text: "#cad3f5", mauve: "#c6a0f6"},
		"catppuccin-mocha":     {mood: theme.MoodDark, base: "#1e1e2e", text: "#cdd6f4", mauve: "#cba6f7"},
	}

	for id, expected := range flavors {
		th, ok := theme.Themes[id]
		if !ok {
			t.Fatalf("theme %s not found", id)
		}
		if th.Mood != expected.mood {
			t.Errorf("%s mood mismatch: got %s, want %s", id, th.Mood, expected.mood)
		}
		if th.Neutrals.Base != expected.base {
			t.Errorf("%s base mismatch: got %s, want %s", id, th.Neutrals.Base, expected.base)
		}
		if th.Neutrals.Text != expected.text {
			t.Errorf("%s text mismatch: got %s, want %s", id, th.Neutrals.Text, expected.text)
		}
		if th.Accents.Mauve != expected.mauve {
			t.Errorf("%s mauve mismatch: got %s, want %s", id, th.Accents.Mauve, expected.mauve)
		}
		css := theme.ThemeRootCSS(id)
		if !strings.Contains(css, "--base:") || !strings.Contains(css, expected.base) {
			t.Errorf("%s missing base in ThemeRootCSS", id)
		}
	}

	odp, ok := theme.Themes["onedarkpro"]
	if !ok {
		t.Fatalf("onedarkpro not found")
	}
	if odp.Mood != theme.MoodDark || odp.Neutrals.Base != "#282c34" || odp.Accents.Mauve != "#c678dd" {
		t.Errorf("onedarkpro palette mismatch: %+v", odp)
	}
	if !strings.Contains(odp.HLJS, "atom-one-dark.min.css") {
		t.Errorf("onedarkpro HLJS mismatch")
	}
}

func TestMatteThemes(t *testing.T) {
	cases := []struct {
		id     string
		mood   theme.Mood
		base   string
		text   string
		accent string
		aliases []string
	}{
		{
			id:      "matte-black",
			mood:    theme.MoodDark,
			base:    "#161618",
			text:    "#f0f0f3",
			accent:  "#f0f0f3",
			aliases: []string{"matteblack", "matte black", "black"},
		},
		{
			id:      "matte-white",
			mood:    theme.MoodLight,
			base:    "#f7f7f5",
			text:    "#141413",
			accent:  "#141413",
			aliases: []string{"mattewhite", "matte white", "white"},
		},
	}

	for _, tc := range cases {
		th, ok := theme.Themes[tc.id]
		if !ok {
			t.Fatalf("theme %s not found", tc.id)
		}
		if th.Mood != tc.mood {
			t.Errorf("%s mood mismatch: got %s, want %s", tc.id, th.Mood, tc.mood)
		}
		if th.Neutrals.Base != tc.base {
			t.Errorf("%s base mismatch: got %s, want %s", tc.id, th.Neutrals.Base, tc.base)
		}
		if th.Neutrals.Text != tc.text {
			t.Errorf("%s text mismatch: got %s, want %s", tc.id, th.Neutrals.Text, tc.text)
		}
		if got := th.RoleColor("accent"); got != tc.accent {
			t.Errorf("%s RoleColor(accent) mismatch: got %s, want %s", tc.id, got, tc.accent)
		}
		for _, alias := range tc.aliases {
			if res := theme.ResolveThemeName(alias); res != tc.id {
				t.Errorf("alias %q resolved to %s, want %s", alias, res, tc.id)
			}
		}
	}
}

