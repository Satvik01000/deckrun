package render

import (
	"fmt"

	"deckrun/internal/theme"
)

const (
	PreviewWidth  = 1600
	PreviewHeight = 900
)

func GeneratePreviewHTML(
	initialTheme string,
	fonts FontOverrides,
	initialTemplate string,
	initialTransition string,
) string {
	th := theme.ResolveThemeName(initialTheme)
	tm := theme.ResolveTemplateName(initialTemplate)
	tr := theme.ResolveTransitionName(initialTransition)

	headFont, _ := theme.FindFont(fonts.Head)
	bodyFont, _ := theme.FindFont(fonts.Body)
	fontAttrs := ""
	if headFont != "" {
		fontAttrs += fmt.Sprintf(` data-head="%s"`, headFont)
	}
	if bodyFont != "" {
		fontAttrs += fmt.Sprintf(` data-body="%s"`, bodyFont)
	}

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
  <title>preview · deckrun</title>
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
  </style>
</head>
<body>
<div id="backdrop" aria-hidden="true"></div>
<div id="presentation"></div>
<div id="pv-empty">nothing to preview yet</div>
<script>
%s

%s
</script>
<script>
%s
</script>
</body>
</html>`,
		th, theme.DecorOf(th), tm, tr, fontAttrs,
		theme.GoogleFontsHref(theme.ThemeIDs, extraFonts),
		theme.HLJSHref(th),
		RichContentHead(RichFeatures{Math: true, Mermaid: true}, "local"),
		ResetCSS,
		theme.ThemeSwitchableCSS(),
		theme.FontOverrideCSS(),
		SlideCSS,
		theme.TemplateCSS,
		theme.TransitionCSS,
		FragmentCSS,
		RichContentCSS,
		theme.DecorCSS,
		PreviewCSS,
		FragmentRuntime,
		RichContentRuntime,
		PreviewScript,
	)
}
