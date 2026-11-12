package render

import (
	"strings"

	"deckrun/internal/parser"
)

type RichFeatures struct {
	Math    bool `json:"math"`
	Mermaid bool `json:"mermaid"`
}

func RichContentFeatures(slides []parser.Slide) RichFeatures {
	var math, mermaid bool
	for _, slide := range slides {
		if !math && (strings.Contains(slide.HTML, `class="math-source"`) || strings.Contains(slide.HTML, "math-source")) {
			math = true
		}
		if !mermaid && (strings.Contains(slide.HTML, "language-mermaid") ||
			strings.Contains(slide.HTML, "lang-mermaid") ||
			strings.Contains(slide.HTML, `class="mermaid"`) ||
			strings.Contains(slide.HTML, "mermaid-source")) {
			mermaid = true
		}
	}
	return RichFeatures{Math: math, Mermaid: mermaid}
}

func RichContentHead(features RichFeatures, source string) string {
	var parts []string
	if features.Math {
		if source == "cdn" {
			parts = append(parts, `<link rel="stylesheet" href="https://cdn.jsdelivr.net/npm/katex@0.18.4/dist/katex.min.css">`)
			parts = append(parts, `<script src="https://cdn.jsdelivr.net/npm/katex@0.18.4/dist/katex.min.js"></script>`)
		} else {
			parts = append(parts, `<link rel="stylesheet" href="/__vendor/katex.min.css">`)
			parts = append(parts, `<script src="/__vendor/katex.min.js"></script>`)
		}
	}
	if features.Mermaid {
		if source == "cdn" {
			parts = append(parts, `<script src="https://cdn.jsdelivr.net/npm/mermaid@11.17.2/dist/mermaid.min.js"></script>`)
		} else {
			parts = append(parts, `<script src="/__vendor/mermaid.min.js"></script>`)
		}
	}
	return strings.Join(parts, "\n  ")
}

const RichContentCSS = `/* ── Rich Content (Math & Diagrams) ─────────────────────────────────────── */

.math-source {
  font-family: inherit;
}

div.math-source {
  display: flex;
  justify-content: center;
  margin: 1.2em 0;
  overflow-x: auto;
  overflow-y: hidden;
}

span.math-source {
  display: inline;
}

.katex-display {
  margin: 0.8em 0;
  overflow-x: auto;
  overflow-y: hidden;
}

.katex {
  font-size: 1.15em;
  text-rendering: auto;
}

.math-error {
  color: var(--maroon, #f38ba8);
  background: var(--surface0, rgba(255, 0, 0, 0.1));
  padding: 2px 6px;
  border-radius: 4px;
  font-family: var(--font-mono, monospace);
  font-size: 0.85em;
}

/* ── Mermaid Diagrams (Bug 2 & Bug 3 Fixes) ───────────────────────────── */
.mermaid-container {
  display: flex;
  justify-content: center;
  align-items: center;
  text-align: center;
  margin: 1.2em auto;
  max-width: 100%;
  overflow: hidden;
}

.mermaid-container svg {
  max-width: 100%;
  max-height: 60vh;
  height: auto;
}

/* Isolate slide typography from leaking into Mermaid SVG foreignObject text */
.slide__content .mermaid-container p,
.slide__content .mermaid-container span,
.slide__content .mermaid-container div {
  font-size: 16px !important;
  line-height: 1.35 !important;
  margin: 0 !important;
  padding: 0 !important;
}

.mermaid-error {
  color: #f38ba8;
  background: rgba(255, 0, 0, 0.08);
  border: 1px solid #f38ba8;
  border-radius: 6px;
  padding: 12px 16px;
  font-family: monospace;
  font-size: 0.9em;
  white-space: pre-wrap;
  margin: 1em 0;
}
`

const RichContentRuntime = `(function () {
  window.deckrunRenderRichContent = function (root) {
    if (!root) return Promise.resolve();

    // 1. Render KaTeX math
    if (window.katex) {
      var mathNodes = root.querySelectorAll('.math-source:not([data-rendered])');
      for (var i = 0; i < mathNodes.length; i++) {
        var el = mathNodes[i];
        var tex = el.textContent || '';
        var isDisplay = el.dataset.display === 'true';
        try {
          window.katex.render(tex, el, {
            displayMode: isDisplay,
            throwOnError: false,
            output: 'htmlAndMathml'
          });
          el.setAttribute('data-rendered', 'true');
        } catch (err) {
          el.innerHTML = '<span class="math-error">' + (err && err.message ? err.message : 'Math rendering error') + '</span>';
          el.setAttribute('data-rendered', 'true');
        }
      }
    }

    // 2. Render Mermaid diagrams (Bug 4 & Bug 5 Fixes)
    var codeBlocks = Array.from(root.querySelectorAll('pre code.language-mermaid, pre code.lang-mermaid'));
    var existingContainers = Array.from(root.querySelectorAll('.mermaid-container[data-mermaid-code]'));
    if ((codeBlocks.length === 0 && existingContainers.length === 0) || !window.mermaid) {
      return Promise.resolve();
    }

    var lightThemes = ['daylight', 'paper', 'solarized', 'rosequartz', 'swiss', 'arctic', 'catppuccin-latte'];
    var curTheme = document.documentElement.dataset.theme || '';
    var isLight = lightThemes.indexOf(curTheme) !== -1;
    var targetMood = isLight ? 'light' : 'dark';

    var needReinit = false;
    var staleContainers = existingContainers.filter(function (c) {
      return c.dataset.mermaidMood !== targetMood;
    });
    if (codeBlocks.length > 0 || staleContainers.length > 0) {
      needReinit = true;
    }
    if (!needReinit) {
      return Promise.resolve();
    }

    try {
      window.mermaid.initialize({
        startOnLoad: false,
        theme: isLight ? 'default' : 'dark',
        securityLevel: 'loose',
        flowchart: {
          htmlLabels: false,
          curve: 'basis'
        }
      });
    } catch (e) {}

    var promises = [];

    function renderIntoWrapper(wrapper, code) {
      wrapper.dataset.mermaidCode = code;
      wrapper.dataset.mermaidMood = targetMood;
      var randId = 'mermaid-' + Math.random().toString(36).slice(2, 10);
      var p = window.mermaid.render(randId, code)
        .then(function (result) {
          wrapper.classList.remove('mermaid-error');
          if (!wrapper.classList.contains('mermaid-container')) {
            wrapper.classList.add('mermaid-container');
          }
          wrapper.innerHTML = result.svg;
        })
        .catch(function (err) {
          var stray = document.getElementById('d' + randId);
          if (stray) stray.remove();

          wrapper.className = 'mermaid-error';
          wrapper.textContent = 'Mermaid error: ' + (err && err.message ? err.message : String(err));
        });
      promises.push(p);
    }

    for (var j = 0; j < codeBlocks.length; j++) {
      (function (codeEl) {
        var preEl = codeEl.closest('pre');
        if (!preEl || preEl.dataset.mermaidDone) return;
        preEl.dataset.mermaidDone = 'true';

        var code = (codeEl.textContent || '').trim();
        if (!code) return;

        var wrapper = document.createElement('div');
        wrapper.className = 'mermaid-container';
        if (preEl.classList.contains('fragment')) wrapper.classList.add('fragment');
        if (preEl.classList.contains('is-visible')) wrapper.classList.add('is-visible');
        if (preEl.classList.contains('is-revealed')) wrapper.classList.add('is-revealed');
        if (preEl.hasAttribute('aria-hidden')) {
          wrapper.setAttribute('aria-hidden', preEl.getAttribute('aria-hidden'));
        }

        preEl.parentNode.insertBefore(wrapper, preEl);
        preEl.remove();
        renderIntoWrapper(wrapper, code);
      })(codeBlocks[j]);
    }

    for (var k = 0; k < staleContainers.length; k++) {
      renderIntoWrapper(staleContainers[k], staleContainers[k].dataset.mermaidCode);
    }

    return Promise.all(promises).then(function () {});
  };
})();
`

