package render

const ResetCSS = `/* ── Reset & base ─────────────────────────────────────────────────────── */
*, *::before, *::after { box-sizing: border-box; margin: 0; padding: 0; }
::selection {
  background: var(--selection-bg, var(--accent-line, rgba(56, 139, 253, 0.22)));
  color: var(--selection-text, inherit);
}
::-moz-selection {
  background: var(--selection-bg, var(--accent-line, rgba(56, 139, 253, 0.22)));
  color: var(--selection-text, inherit);
}`

const SlideCSS = `/* The gutter a slide keeps around its content. A variable rather than a
   literal because the composition templates each reset it. */
:root {
  --slide-pad-y: 4.4rem;
  --slide-pad-x: 6rem;
}

html, body {
  height: 100%;
  overflow: hidden;
  background: var(--crust);
  color: var(--text);
  font-family: var(--font-body);
  letter-spacing: var(--body-tracking);
  -webkit-font-smoothing: antialiased;
  -moz-osx-font-smoothing: grayscale;
  text-rendering: optimizeLegibility;
  font-variant-ligatures: common-ligatures;
}

/* ── Presentation shell ───────────────────────────────────────────────── */
#presentation {
  position: relative;
  z-index: 1;
  width: 100vw;
  height: 100vh;
  overflow: hidden;
}

/* ── Slide base ───────────────────────────────────────────────────────── */
.slide {
  position: absolute;
  inset: 0;
  display: flex;
  flex-direction: column;
  align-items: stretch;
  justify-content: flex-start;
  padding: var(--slide-pad-y) var(--slide-pad-x);
  opacity: 0;
  pointer-events: none;
  /* forward: enter from right */
  transform: translateX(48px);
  transition:
    opacity 0.38s cubic-bezier(0.4, 0, 0.2, 1),
    transform 0.38s cubic-bezier(0.4, 0, 0.2, 1);
}

.slide.is-active {
  opacity: 1;
  transform: translateX(0);
  pointer-events: all;
}

/* Exiting slide direction classes — set by JS before transition */
.slide.exit-left  { opacity: 0; transform: translateX(-48px); }
.slide.exit-right { opacity: 0; transform: translateX(48px); }
.slide.enter-from-left  { transform: translateX(-48px); opacity: 0; }
.slide.enter-from-right { transform: translateX(48px);  opacity: 0; }

/* ── Background image layer ───────────────────────────────────────────── */
.slide--has-bg {
  background: var(--base);
}

.slide__bg {
  position: absolute;
  inset: 0;
  background-image: var(--slide-bg-url);
  background-size: cover;
  background-position: center;
  opacity: var(--slide-bg-opacity, 0.5);
  z-index: 0;
}

/* ── Content area ─────────────────────────────────────────────────────── */
.slide__content {
  position: relative;
  z-index: 1;
  width: 100%;
  max-width: 100%;
  max-height: calc(100vh - var(--slide-pad-y) * 2);
  overflow: hidden;
}

/* Each block lands a beat after the one above it, so a slide assembles
   itself instead of appearing all at once. */
.slide.is-active .slide__content > * {
  animation: slide-rise 0.52s cubic-bezier(0.22, 1, 0.36, 1) both;
}

.slide.is-active .slide__content > *:nth-child(1) { animation-delay: 0.05s; }
.slide.is-active .slide__content > *:nth-child(2) { animation-delay: 0.11s; }
.slide.is-active .slide__content > *:nth-child(3) { animation-delay: 0.17s; }
.slide.is-active .slide__content > *:nth-child(4) { animation-delay: 0.23s; }
.slide.is-active .slide__content > *:nth-child(5) { animation-delay: 0.29s; }
.slide.is-active .slide__content > *:nth-child(6) { animation-delay: 0.34s; }
.slide.is-active .slide__content > *:nth-child(n+7) { animation-delay: 0.39s; }

@keyframes slide-rise {
  from { opacity: 0; transform: translateY(15px); }
  to   { opacity: 1; transform: none; }
}

@media (prefers-reduced-motion: reduce) {
  .slide { transition: opacity 0.2s linear; transform: none !important; }
  .slide.is-active .slide__content > * { animation: none !important; }
}

/* ── Split layouts ────────────────────────────────────────────────────── */
.slide__split {
  position: relative;
  z-index: 1;
  display: flex;
  width: 100%;
  max-width: 100%;
  height: calc(100vh - var(--slide-pad-y) * 2);
  align-items: center;
  gap: 3.2rem;
}

.slide__split .slide__content {
  flex: 1;
  max-width: none;
}

.slide__image-panel {
  flex: 1;
  display: flex;
  align-items: center;
  justify-content: center;
  max-height: 80vh;
}

.slide__image-panel img {
  max-width: 100%;
  max-height: 78vh;
  object-fit: contain;
  border-radius: 12px;
  border: 1px solid var(--hairline);
  box-shadow: var(--shadow-lg);
}

/* ── Typography ───────────────────────────────────────────────────────── */
.slide__content h1,
.slide__content h2,
.slide__content h3,
.slide__content h4 {
  font-family: var(--font-display);
  letter-spacing: var(--display-tracking);
  text-wrap: balance;
}

.slide__content h1 {
  position: relative;
  font-size: clamp(2.1rem, 4.6vw, 3.6rem);
  font-weight: var(--display-weight);
  text-transform: var(--display-case);
  color: var(--accent);
  margin-bottom: 1.5rem;
  padding-bottom: 0.55rem;
  line-height: 1.1;
}

/* A gradient rule under the title, fading out rather than stopping dead. */
.slide__content h1::after {
  content: '';
  position: absolute;
  left: 0;
  bottom: 0;
  width: 2.4em;
  max-width: 55%;
  height: 3px;
  border-radius: 3px;
  background: var(--accent-fade);
}

.slide__content h2 {
  font-size: clamp(1.55rem, 3vw, 2.5rem);
  font-weight: 600;
  color: var(--accent-2);
  margin-bottom: 1.05rem;
  line-height: 1.2;
}

.slide__content h3 {
  font-size: clamp(1.15rem, 2vw, 1.8rem);
  font-weight: 600;
  color: var(--accent-3);
  margin-bottom: 0.75rem;
  line-height: 1.28;
}

.slide__content h4 {
  font-size: 1.3rem;
  font-weight: 600;
  color: var(--subtext1);
  letter-spacing: 0.08em;
  text-transform: uppercase;
  margin-bottom: 0.55rem;
}

.slide__content p {
  font-size: clamp(1rem, 1.6vw, 1.35rem);
  line-height: 1.72;
  margin-bottom: 1rem;
  color: var(--subtext1);
}

.slide__content strong {
  color: var(--text);
  font-weight: 700;
  /* A tinted underlay instead of a second color, so emphasis reads without
     turning the sentence into a swatch. */
  background: linear-gradient(transparent 62%, var(--accent-soft) 62%);
}

.slide__content em {
  color: var(--accent-3);
  font-style: italic;
}

/* ── Lists ────────────────────────────────────────────────────────────── */
.slide__content ul,
.slide__content ol {
  font-size: clamp(0.95rem, 1.5vw, 1.28rem);
  line-height: 1.72;
  margin-bottom: 1rem;
  color: var(--subtext1);
}

.slide__content ul { list-style: none; padding-left: 1.35em; }

/* Numbers live in the marker box, outside the content box, so an ordered list
   needs room for "10." before its text starts. */
.slide__content ol { padding-left: 2.05em; }

.slide__content li {
  margin-bottom: 0.42em;
  padding-left: 0.15em;
}

.slide__content ol li::marker {
  color: var(--accent);
  font-family: var(--font-mono);
  font-size: 0.86em;
  font-weight: 600;
}

/* A custom bullet: small, accent-colored, and vertically centred on the
   first line rather than sitting on the baseline like a period. */
.slide__content ul > li {
  position: relative;
  padding-left: 0.9em;
}

.slide__content ul > li::before {
  content: '';
  position: absolute;
  left: 0;
  top: 0.62em;
  width: 0.33em;
  height: 0.33em;
  border-radius: 0.1em;
  background: var(--accent);
  transform: rotate(45deg);
}

.slide__content ul ul > li::before {
  background: transparent;
  border: 1.5px solid var(--accent-3);
}

.slide__content ul ul,
.slide__content ol ol,
.slide__content ul ol,
.slide__content ol ul {
  margin-top: 0.34em;
  margin-bottom: 0;
  padding-left: 1em;
  font-size: 0.94em;
  color: var(--subtext0);
}

/* ── Code ─────────────────────────────────────────────────────────────── */
.slide__content pre {
  position: relative;
  margin: 1.2rem 0;
  border-radius: 12px;
  border: 1px solid var(--hairline);
  overflow-x: auto;
  font-size: clamp(0.75rem, 1.1vw, 1rem);
  box-shadow: var(--code-shadow);
  background: var(--mantle);
}

/* A hairline of accent along the top edge, so a code block reads as a panel
   rather than as a hole in the slide. */
.slide__content pre::before {
  content: '';
  position: absolute;
  inset: 0 0 auto 0;
  height: 2px;
  border-radius: 12px 12px 0 0;
  background: var(--accent-fade);
  opacity: 0.8;
}

/* Override hljs background to match our theme */
.slide__content pre code.hljs {
  background: transparent !important;
  border-radius: 12px;
  padding: 1.5rem 1.7rem;
  font-family: var(--font-mono);
  font-size: inherit;
  line-height: 1.66;
  letter-spacing: 0;
}

/* Inline code */
.slide__content :not(pre) > code {
  font-family: var(--font-mono);
  font-size: 0.86em;
  background: var(--surface-soft);
  color: var(--accent-3);
  border-radius: 5px;
  padding: 0.16em 0.42em;
  border: 1px solid var(--hairline);
  letter-spacing: 0;
}

/* ── Blockquotes ──────────────────────────────────────────────────────── */
.slide__content blockquote {
  position: relative;
  padding: 0.78em 1.25em 0.78em 1.7em;
  margin: 1.2rem 0;
  background: var(--accent-soft);
  border-left: 3px solid var(--accent);
  border-radius: 0 10px 10px 0;
  color: var(--subtext1);
  font-size: clamp(1rem, 1.5vw, 1.3rem);
}

.slide__content blockquote::before {
  content: '\\201C';
  position: absolute;
  /* em here is the glyph's own 2.6em, not the quote's. */
  left: 0.16em;
  top: 0.04em;
  font-family: var(--font-display);
  font-size: 2.6em;
  line-height: 1;
  color: var(--accent-line);
  pointer-events: none;
}

.slide__content blockquote p {
  font-size: inherit;
  margin-bottom: 0;
  color: inherit;
  max-width: none;
}

/* ── Tables ───────────────────────────────────────────────────────────── */
.slide__content table {
  border-collapse: collapse;
  width: 100%;
  margin: 1.2rem 0;
  font-size: clamp(0.85rem, 1.2vw, 1.05rem);
  /* Rules, not boxes: a grid of borders fights the text for attention. */
  border-bottom: 1px solid var(--hairline);
}

.slide__content th {
  background: transparent;
  color: var(--accent);
  font-family: var(--font-display);
  font-weight: 600;
  font-size: 0.86em;
  letter-spacing: 0.1em;
  text-transform: uppercase;
  padding: 0.5rem 1rem;
  text-align: left;
  border: none;
  border-bottom: 2px solid var(--accent-line);
}

.slide__content td {
  padding: 0.6rem 1rem;
  border: none;
  border-bottom: 1px solid var(--hairline);
  color: var(--subtext1);
}

.slide__content tr:nth-child(even) td {
  background: var(--surface-soft);
}

/* ── Links ────────────────────────────────────────────────────────────── */
.slide__content a {
  color: var(--accent-2);
  text-decoration: none;
  background: linear-gradient(var(--accent-2), var(--accent-2)) 0 100% / 100% 1px no-repeat;
  padding-bottom: 2px;
  transition: color 0.18s ease, background-size 0.18s ease;
}

.slide__content a:hover {
  color: var(--accent);
  background-image: linear-gradient(var(--accent), var(--accent));
  background-size: 100% 2px;
}

/* ── Inline images (no positioning) ──────────────────────────────────── */
.slide__content img {
  max-width: 100%;
  max-height: 55vh;
  border-radius: 10px;
  display: block;
  margin: 1rem auto;
  box-shadow: var(--shadow-md);
}

/* ── Embeds: raw HTML iframe / video ─────────────────────────────────── */
.slide__content iframe {
  display: block;
  width: 100%;
  max-width: 100%;
  aspect-ratio: 16 / 9;
  height: auto;
  margin: 1.1rem auto;
  border: 1px solid var(--hairline);
  border-radius: 12px;
  background: var(--mantle);
  box-shadow: var(--shadow-md);
}

.slide__content video {
  display: block;
  max-width: 100%;
  max-height: 60vh;
  margin: 1.1rem auto;
  border-radius: 12px;
  background: var(--crust);
  object-fit: contain;
  box-shadow: var(--shadow-md);
}

/* ── Inline HTML accents ──────────────────────────────────────────────── */
.slide__content kbd {
  display: inline-block;
  font-family: var(--font-mono);
  font-size: 0.8em;
  background: var(--surface0);
  border: 1px solid var(--surface2);
  border-bottom-width: 2px;
  border-radius: 6px;
  padding: 0.12em 0.45em;
  color: var(--text);
  white-space: nowrap;
}

.slide__content mark {
  background: var(--accent-soft);
  color: var(--accent);
  box-shadow: inset 0 -0.42em 0 var(--accent-soft);
  border-radius: 3px;
  padding: 0.05em 0.24em;
  font-weight: 600;
}

/* ── Horizontal rule ──────────────────────────────────────────────────── */
.slide__content hr {
  border: none;
  height: 1px;
  margin: 1.8rem 0;
  background: linear-gradient(90deg, var(--accent-line), var(--hairline) 40%, transparent);
}`

const ChromeCSS = `/* ── HUD (progress + counter) ────────────────────────────────────────── */
#hud {
  font-family: var(--font-mono);
  position: fixed;
  bottom: 0;
  left: 0;
  right: 0;
  z-index: 200;
  pointer-events: none;
}

#progress-bar {
  height: 2px;
  background: var(--surface0);
}

#progress-fill {
  height: 100%;
  background: var(--gradient);
  transition: width 0.3s ease;
  width: 0%;
}

#slide-counter {
  flex: 0 0 auto;
  text-align: right;
  font-size: 0.72rem;
  color: var(--overlay1);
  letter-spacing: 0.08em;
  white-space: nowrap;
}

/* ── Nav arrows ───────────────────────────────────────────────────────── */
.nav-arrow {
  position: fixed;
  top: 50%;
  transform: translateY(-50%);
  background: transparent;
  border: none;
  color: var(--surface1);
  cursor: pointer;
  padding: 1.2rem 0.8rem;
  z-index: 200;
  transition: color 0.2s ease;
  line-height: 1;
  font-size: 1.4rem;
  pointer-events: all;
}

.nav-arrow:hover { color: var(--text); }
.nav-arrow--prev { left: 0.5rem; }
.nav-arrow--next { right: 0.5rem; }

/* ── Overview mode ────────────────────────────────────────────────────── */
#overview {
  font-family: var(--font-mono);
  position: fixed;
  inset: 0;
  background: var(--crust);
  z-index: 300;
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(280px, 1fr));
  gap: 1.5rem;
  padding: 2rem;
  overflow-y: auto;
}

#overview.hidden { display: none; }

.overview-thumb {
  background: var(--base);
  border: 1px solid var(--surface0);
  border-radius: 12px;
  cursor: pointer;
  overflow: hidden;
  aspect-ratio: 16/9;
  display: flex;
  align-items: center;
  justify-content: center;
  position: relative;
  transition: border-color 0.2s ease, transform 0.2s ease;
}

.overview-thumb:hover { border-color: var(--accent); transform: translateY(-3px) scale(1.02); box-shadow: var(--shadow-md); }
.overview-thumb.is-current { border-color: var(--accent-2); box-shadow: 0 0 0 1px var(--accent-2); }

.overview-thumb__number {
  position: absolute;
  top: 0.4rem;
  left: 0.5rem;
  font-size: 0.65rem;
  color: var(--overlay0);
  z-index: 1;
}

.overview-thumb__inner {
  width: 100%;
  height: 100%;
  transform: scale(0.28);
  transform-origin: top left;
  pointer-events: none;
  position: absolute;
  top: 0;
  left: 0;
}

/* ── Kbd hint ─────────────────────────────────────────────────────────── */
#kbd-hint {
  font-family: var(--font-mono);
  position: fixed;
  bottom: 2.2rem;
  left: 50%;
  transform: translateX(-50%);
  font-size: 0.65rem;
  color: var(--overlay0);
  letter-spacing: 0.06em;
  pointer-events: none;
  opacity: 1;
  transition: opacity 0.6s ease;
  z-index: 150;
}

#kbd-hint.hidden { opacity: 0; }

/* ── Blinking cursor ──────────────────────────────────────────────────── */
/* Parked where an h1's cap height sits, so it reads as the title's caret. */
#cursor {
  position: fixed;
  top: var(--slide-pad-y);
  right: var(--slide-pad-x);
  width: 12px;
  height: clamp(2rem, 4.5vw, 3.4rem);
  background: var(--accent);
  box-shadow: 0 0 22px var(--glow);
  z-index: 100;
  pointer-events: none;
  animation: cursor-blink 1.1s step-start infinite;
}

@keyframes cursor-blink {
  0%, 100% { opacity: 1; }
  50%       { opacity: 0; }
}

@media print {
  #cursor { display: none !important; }
}

/* ── Fullscreen hint ──────────────────────────────────────────────────── */
#fs-hint {
  position: fixed;
  inset: 0;
  z-index: 500;
  display: flex;
  align-items: center;
  justify-content: center;
  background: var(--crust-overlay);
  cursor: pointer;
  transition: opacity 0.4s ease;
}

#fs-hint.hidden { opacity: 0; pointer-events: none; }

#fs-hint__inner {
  font-family: var(--font-mono);
  text-align: center;
  color: var(--subtext1);
  font-size: 0.9rem;
  letter-spacing: 0.06em;
  border: 1px solid var(--surface1);
  border-radius: 8px;
  padding: 1.6rem 2.8rem;
  background: var(--base);
}

#fs-hint__inner kbd {
  display: inline-block;
  background: var(--surface0);
  border: 1px solid var(--surface1);
  border-radius: 5px;
  padding: 0.1em 0.5em;
  font-family: var(--font-mono);
  color: var(--accent);
}

/* ── Scrollbar ────────────────────────────────────────────────────────── */
::-webkit-scrollbar { width: 6px; height: 6px; }
::-webkit-scrollbar-track { background: transparent; }
::-webkit-scrollbar-thumb { background: var(--surface1); border-radius: 3px; }

/* ── Print / PDF export ───────────────────────────────────────────────── */
/* One 16:9 page per slide, edge to edge. 13.333in x 7.5in is the standard
   widescreen slide size, so the page box needs no orientation choice. */
@page {
  size: 13.333in 7.5in;
  margin: 0;
}

@media print {
  /* Without this, printing drops every background: the theme, the code block
     surfaces, and the background images all vanish behind white paper. */
  *, *::before, *::after {
    -webkit-print-color-adjust: exact !important;
    print-color-adjust: exact !important;
  }

  html, body {
    width: 13.333in !important;
    height: auto !important;
    overflow: visible !important;
    background: var(--crust) !important;
  }

  #presentation {
    position: static !important;
    width: 13.333in !important;
    height: auto !important;
    overflow: visible !important;
  }

  .slide {
    position: relative !important;
    inset: auto !important;
    opacity: 1 !important;
    transform: none !important;
    pointer-events: all !important;
    transition: none !important;
    /* Absolute units, not vw/vh: viewport units in paged media resolve
       against the page box, which is not what the deck was laid out for. */
    width: 13.333in !important;
    height: 7.5in !important;
    page-break-after: always;
    break-after: page;
    break-inside: avoid;
    overflow: hidden !important;
  }

  .slide:last-of-type {
    page-break-after: avoid;
    break-after: avoid;
  }

  .slide__content {
    max-height: calc(7.5in - var(--slide-pad-y) * 2) !important;
  }

  /* The staggered entrance would freeze mid-flight on paper. */
  .slide__content > * { animation: none !important; opacity: 1 !important; transform: none !important; }

  .slide__split {
    height: calc(7.5in - var(--slide-pad-y) * 2) !important;
  }

  .slide__image-panel { max-height: calc(7.5in - var(--slide-pad-y) * 2) !important; }
  .slide__image-panel img { max-height: calc(7.5in - var(--slide-pad-y) * 2 - 1rem) !important; }
  .slide__content img { max-height: 4in !important; }
  .slide__content iframe, .slide__content video { max-height: 4in !important; }

  #hud, .nav-arrow, #overview, #kbd-hint, #cursor, #fs-hint,
  #board, #laser, #blackout, #help, #themes {
    display: none !important;
  }
}`

const PresenterCSS = `/* ── HUD tool strip ───────────────────────────────────────────────────── */
#hud-row {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 1rem;
  padding: 0.26rem 1.1rem 0.34rem;
}

#hud-tools {
  display: flex;
  align-items: center;
  gap: 0.28rem;
  min-width: 0;
  flex-wrap: wrap;
}

.hud-btn {
  pointer-events: all;
  display: inline-flex;
  align-items: center;
  gap: 0.34rem;
  background: transparent;
  border: 1px solid transparent;
  border-radius: 6px;
  padding: 0.14rem 0.44rem;
  font: inherit;
  font-size: 0.66rem;
  letter-spacing: 0.07em;
  color: var(--overlay0);
  cursor: pointer;
  transition: color 0.15s ease, border-color 0.15s ease, background 0.15s ease;
}

.hud-btn:hover { color: var(--text); border-color: var(--surface1); }
.hud-btn.is-on {
  color: var(--accent);
  border-color: var(--accent-line);
  background: var(--accent-soft);
}

.hud-btn kbd {
  font: inherit;
  font-size: 0.58rem;
  color: inherit;
  opacity: 0.7;
  border: 1px solid currentColor;
  border-radius: 3px;
  padding: 0 0.28em;
}

#hud-sep {
  width: 1px;
  height: 12px;
  background: var(--surface1);
  margin: 0 0.2rem;
}

/* ── Pen strip (only while the pen is down) ───────────────────────────── */
#pen-bar {
  display: none;
  align-items: center;
  gap: 0.28rem;
}

#pen-bar.is-on { display: inline-flex; }

#pen-swatches { display: inline-flex; align-items: center; gap: 0.3rem; }

.swatch {
  pointer-events: all;
  width: 13px;
  height: 13px;
  padding: 0;
  border-radius: 50%;
  border: 1px solid var(--surface2);
  cursor: pointer;
  transition: transform 0.12s ease, border-color 0.12s ease;
}

.swatch:hover { transform: scale(1.2); }
.swatch.is-on { transform: scale(1.35); border-color: var(--text); }

#pen-width {
  font-size: 0.6rem;
  color: var(--overlay0);
  letter-spacing: 0.06em;
  min-width: 2.2em;
  text-align: center;
}

#hud-right {
  display: flex;
  align-items: center;
  gap: 1.1rem;
  flex: 0 0 auto;
}

.hud-brand {
  pointer-events: all;
  font-size: 0.64rem;
  color: var(--overlay0);
  text-decoration: none;
  letter-spacing: 0.06em;
  opacity: 0.72;
  transition: opacity 0.15s ease, color 0.15s ease;
  white-space: nowrap;
}

.hud-brand:hover {
  opacity: 1;
  color: var(--accent);
}

.hud-brand span {
  font-weight: 600;
  color: var(--subtext1);
}

.hud-brand:hover span {
  color: var(--accent);
}

/* ── Annotation canvas ────────────────────────────────────────────────── */
#board {
  position: fixed;
  inset: 0;
  z-index: 180;
  pointer-events: none;
  touch-action: none;
  background: transparent;
  transition: background 0.18s ease;
}

/* Only the pen makes the canvas swallow clicks, so navigation keeps working
   whenever annotations are merely on display. */
#board.is-drawing { pointer-events: all; cursor: crosshair; }
#board.is-erasing { cursor: cell; }
#board.is-blank   { background: var(--crust); }

/* ── Laser pointer ────────────────────────────────────────────────────── */
#laser {
  position: fixed;
  left: 0;
  top: 0;
  width: 20px;
  height: 20px;
  margin: -10px 0 0 -10px;
  border-radius: 50%;
  background: radial-gradient(circle,
    rgba(255,255,255,0.95) 0%,
    var(--red) 38%,
    var(--glow) 62%,
    transparent 74%);
  box-shadow: 0 0 20px 8px var(--glow);
  z-index: 380;
  pointer-events: none;
  display: none;
  will-change: transform;
}

#laser.is-on { display: block; }

/* The dot replaces the cursor, so the real one gets out of the way. */
body.laser-on, body.laser-on #board.is-drawing { cursor: none; }

/* ── Blackout ─────────────────────────────────────────────────────────── */
#blackout {
  position: fixed;
  inset: 0;
  background: #000;
  z-index: 420;
  display: none;
  cursor: pointer;
}

#blackout.is-on { display: block; }

/* ── Theme picker overlay ─────────────────────────────────────────────── */
#themes {
  position: fixed;
  inset: 0;
  z-index: 450;
  display: none;
}

#themes.is-on { display: block; }

#themes__backdrop {
  position: absolute;
  inset: 0;
  background: var(--crust-overlay);
  backdrop-filter: blur(4px);
}

#themes__box {
  font-family: var(--font-mono);
  position: relative;
  width: min(980px, 94vw);
  max-height: 84vh;
  margin: 6vh auto 0;
  background: var(--mantle);
  border: 1px solid var(--surface1);
  border-radius: 14px;
  box-shadow: var(--shadow-lg);
  overflow: hidden;
  display: flex;
  flex-direction: column;
}

#themes__head {
  display: flex;
  align-items: baseline;
  justify-content: space-between;
  gap: 0.8rem;
  padding: 1rem 1.4rem;
  border-bottom: 1px solid var(--surface0);
  background: var(--mantle);
}

.th-head__title {
  font-size: 0.88rem;
  font-weight: 600;
  letter-spacing: 0.1em;
  text-transform: uppercase;
  color: var(--text);
}

.th-head__sub {
  flex: 1 1 auto;
  font-size: 0.68rem;
  color: var(--overlay1);
}

.th-head__sub kbd {
  display: inline-block;
  font: inherit;
  font-size: 0.6rem;
  background: var(--surface0);
  border: 1px solid var(--surface1);
  border-radius: 3px;
  padding: 0.05em 0.35em;
  color: var(--lavender);
}

.th-head__brand {
  font-size: 0.65rem;
  color: var(--overlay0);
  text-decoration: none;
  padding: 0.2rem 0.5rem;
  border-radius: 4px;
  background: var(--surface0);
  border: 1px solid var(--surface1);
  transition: color 0.15s ease, border-color 0.15s ease;
}
.th-head__brand:hover { color: var(--accent); border-color: var(--accent-line); }

#themes__close {
  background: transparent;
  border: none;
  color: var(--overlay0);
  font-size: 1.25rem;
  line-height: 1;
  cursor: pointer;
  padding: 0 0.2rem;
}
#themes__close:hover { color: var(--text); }

#themes__list {
  flex: 1 1 auto;
  overflow-y: auto;
  padding: 1.1rem 1.4rem 1.4rem;
}

.th-group {
  padding: 0.4rem 0.2rem 0.6rem;
  font-size: 0.62rem;
  letter-spacing: 0.16em;
  text-transform: uppercase;
  color: var(--accent);
}

.th-grid {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(195px, 1fr));
  gap: 0.85rem;
  margin-bottom: 1.2rem;
}

.th-card {
  display: flex;
  flex-direction: column;
  padding: 0;
  overflow: hidden;
  border: 1px solid var(--surface1);
  border-radius: 10px;
  background: var(--base);
  cursor: pointer;
  text-align: left;
  font-family: inherit;
  transition: border-color 0.14s ease, transform 0.14s ease, box-shadow 0.14s ease;
}

.th-card:hover, .th-card.is-sel {
  border-color: var(--accent);
  transform: translateY(-2px);
  box-shadow: var(--shadow-md);
}

.th-card.is-current { box-shadow: inset 0 0 0 1px var(--accent); }

.th-thumb {
  position: relative;
  aspect-ratio: 16 / 9;
  padding: 11px 12px;
  overflow: hidden;
  border-bottom: 1px solid var(--surface0);
}

.th-thumb__glow {
  position: absolute;
  width: 150%;
  height: 150%;
  right: -55%;
  top: -60%;
  border-radius: 50%;
  pointer-events: none;
}

.th-thumb__title {
  position: relative;
  font-size: 13px;
  line-height: 1.1;
  margin-bottom: 5px;
}

.th-thumb__rule { position: relative; width: 30px; height: 2px; border-radius: 2px; margin-bottom: 7px; }
.th-thumb__line { position: relative; height: 2.5px; border-radius: 2px; margin-bottom: 4px; }
.th-thumb__code {
  position: relative;
  margin-top: 6px;
  border-radius: 4px;
  padding: 3px 5px;
  font-size: 7.5px;
  letter-spacing: 0;
  white-space: nowrap;
  overflow: hidden;
}

.th-meta { padding: 8px 10px 10px; }
.th-meta__top { display: flex; align-items: baseline; gap: 6px; }
.th-meta__name { flex: 1 1 auto; font-size: 11.5px; color: var(--text); }
.th-meta__mood {
  flex: 0 0 auto;
  font-size: 8px;
  letter-spacing: 0.12em;
  text-transform: uppercase;
  color: var(--overlay0);
}
.th-meta__blurb {
  font-size: 9.5px;
  line-height: 1.5;
  color: var(--overlay1);
  margin-top: 3px;
}

#themes__foot {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 1rem;
  padding: 0.8rem 1.4rem;
  border-top: 1px solid var(--surface0);
  background: var(--mantle);
  font-size: 0.65rem;
  color: var(--overlay1);
}

.th-foot__brand a, #help__foot a {
  color: var(--accent);
  text-decoration: none;
  font-weight: 600;
}
.th-foot__brand a:hover, #help__foot a:hover {
  text-decoration: underline;
}

/* ── Controls overlay ─────────────────────────────────────────────────── */
#help {
  position: fixed;
  inset: 0;
  z-index: 460;
  display: none;
}

#help.is-on { display: block; }

#help__backdrop {
  position: absolute;
  inset: 0;
  background: var(--crust-overlay);
  backdrop-filter: blur(4px);
}

#help__panel {
  font-family: var(--font-mono);
  position: relative;
  width: min(780px, 92vw);
  max-height: 84vh;
  overflow-y: auto;
  margin: 7vh auto 0;
  padding: 1.4rem 1.7rem 1.7rem;
  background: var(--mantle);
  border: 1px solid var(--surface1);
  border-radius: 12px;
  box-shadow: var(--shadow-lg);
}

#help__head {
  display: flex;
  align-items: baseline;
  justify-content: space-between;
  gap: 1rem;
  margin-bottom: 1.1rem;
}

#help__head h2 {
  font-size: 0.92rem;
  font-weight: 600;
  letter-spacing: 0.1em;
  text-transform: uppercase;
  color: var(--text);
}

#help__head p { font-size: 0.68rem; color: var(--overlay1); }

#help__head kbd,
.help-row__keys kbd {
  display: inline-block;
  font: inherit;
  font-size: 0.63rem;
  background: var(--surface0);
  border: 1px solid var(--surface1);
  border-bottom-width: 2px;
  border-radius: 4px;
  padding: 0.05em 0.4em;
  color: var(--lavender);
  white-space: nowrap;
}

#help__grid {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(240px, 1fr));
  gap: 1.2rem 2rem;
}

.help-group__title {
  font-size: 0.6rem;
  text-transform: uppercase;
  letter-spacing: 0.16em;
  color: var(--accent);
  margin-bottom: 0.45rem;
}

.help-row {
  display: flex;
  align-items: baseline;
  justify-content: space-between;
  gap: 0.9rem;
  padding: 0.17rem 0;
  font-size: 0.72rem;
  color: var(--subtext0);
}

.help-row__keys { flex: 0 0 auto; display: flex; gap: 0.22rem; }

#help__close {
  position: absolute;
  top: 0.6rem;
  right: 0.8rem;
  background: transparent;
  border: none;
  color: var(--overlay0);
  font-size: 1.15rem;
  line-height: 1;
  cursor: pointer;
}

#help__close:hover { color: var(--text); }

#help__foot {
  margin-top: 1.2rem;
  padding-top: 0.8rem;
  border-top: 1px solid var(--surface0);
  font-size: 0.66rem;
  color: var(--overlay1);
  line-height: 1.7;
  display: flex;
  justify-content: space-between;
  align-items: center;
  flex-wrap: wrap;
  gap: 0.5rem;
}`

const DocCSS = `html, body {
  height: 100%;
  overflow: hidden;
  background: var(--crust);
}

#doc-frame {
  position: fixed;
  inset: 0;
  width: 100vw;
  height: 100vh;
  border: none;
  z-index: 1;
  background: var(--crust);
}`
