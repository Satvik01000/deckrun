package render

import (
	"encoding/json"
	"fmt"
	"strings"

	"deckrun/internal/theme"
)

type EditorFileInfo struct {
	Name     string `json:"name"`
	Kind     string `json:"kind"` // "markdown" | "html"
	Writable bool   `json:"writable"`
	Watched  bool   `json:"watched"`
}

func BootstrapJSON(
	th string,
	fonts FontOverrides,
	tm string,
	tr string,
	file *EditorFileInfo,
) string {
	headFont, _ := theme.FindFont(fonts.Head)
	bodyFont, _ := theme.FindFont(fonts.Body)

	var headVal any
	if headFont != "" {
		headVal = headFont
	}
	var bodyVal any
	if bodyFont != "" {
		bodyVal = bodyFont
	}

	payload := map[string]any{
		"file":        file,
		"theme":       th,
		"themes":      theme.ThemeSummaries(),
		"fonts":       map[string]any{"head": headVal, "body": bodyVal},
		"faces":       theme.FontSummaries(),
		"template":    tm,
		"templates":   theme.TemplateSummaries(),
		"transition":  tr,
		"transitions": theme.TransitionSummaries(),
		"width":       PreviewWidth,
		"height":      PreviewHeight,
		"groups":      json.RawMessage(SnippetGroupsJSON),
		"snippets":    json.RawMessage(SnippetsJSON),
		"tips":        json.RawMessage(TipsJSON),
		"welcome":     WelcomeDeck,
	}

	b, _ := json.Marshal(payload)
	return strings.ReplaceAll(string(b), "<", `\u003c`)
}

func GenerateEditorHTML(
	themeInput string,
	fonts FontOverrides,
	templateInput string,
	transitionInput string,
	file *EditorFileInfo,
) string {
	th := theme.ResolveThemeName(themeInput)
	tm := theme.ResolveTemplateName(templateInput)
	tr := theme.ResolveTransitionName(transitionInput)

	return fmt.Sprintf(`<!DOCTYPE html>
<html lang="en" data-theme="%s" data-decor="%s" data-template="%s" data-transition="%s">
<head>
  <meta charset="UTF-8">
  <meta name="viewport" content="width=device-width, initial-scale=1.0">
  <title>deckrun · editor</title>
  <link rel="preconnect" href="https://fonts.googleapis.com">
  <link rel="preconnect" href="https://fonts.gstatic.com" crossorigin>
  <link href="%s" rel="stylesheet">
  <style>
%s

%s

%s
  </style>
</head>
<body>
%s

<script type="application/json" id="bootstrap">%s</script>
<script>
%s
</script>
<script>
%s
</script>
</body>
</html>`,
		th, theme.DecorOf(th), tm, tr,
		theme.GoogleFontsHref(theme.ThemeIDs, nil),
		ResetCSS,
		theme.ThemeSwitchableCSS(),
		EditorCSS,
		EditorBodyHTML,
		BootstrapJSON(th, fonts, string(tm), string(tr), file),
		HighlightRuntime,
		EditorScript,
	)
}

