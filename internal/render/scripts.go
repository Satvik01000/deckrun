package render

import (
	_ "embed"
	"encoding/json"
	"strings"

	"deckrun/internal/theme"
)

//go:embed static/preview.css
var PreviewCSS string

//go:embed static/preview.js
var previewScriptTemplate string

func PreviewScript() string {
	fontsJSON, _ := json.Marshal(theme.FontIDs)
	ps := strings.Replace(previewScriptTemplate, "__HLJS__", theme.HLJSMapJSON(), 1)
	ps = strings.Replace(ps, "__DECOR__", theme.DecorMapJSON(), 1)
	ps = strings.Replace(ps, "__FONTS__", string(fontsJSON), 1)
	return ps
}

//go:embed static/editor.css
var EditorCSS string

//go:embed static/editor.js
var EditorScript string


const PresenterScript = `(function () {
  'use strict';

  const slides = Array.from(document.querySelectorAll('.slide'));
  const total  = slides.length;
  let cur      = 0;
  let inOverview = false;

  const elCur      = document.getElementById('cur');
  const elFill     = document.getElementById('progress-fill');
  const elBtnPrev  = document.getElementById('btn-prev');
  const elBtnNext  = document.getElementById('btn-next');
  const elOverview = document.getElementById('overview');
  const elHint     = document.getElementById('kbd-hint');
  const elBoard    = document.getElementById('board');
  const elLaser    = document.getElementById('laser');
  const elBlack    = document.getElementById('blackout');
  const elHelp     = document.getElementById('help');
  const elThemes   = document.getElementById('themes');
  const elPenBar   = document.getElementById('pen-bar');
  const elPenWidth = document.getElementById('pen-width');

  if (window.deckrunPrepareFragments) {
    window.deckrunPrepareFragments(document.getElementById('presentation'), false);
  }

  const fragmentSteps = slides.map(() => 0);

  function fragmentsAt(index) {
    return Array.from(slides[index].querySelectorAll('.fragment'));
  }

  function applyFragmentStep(index) {
    const fragments = fragmentsAt(index);
    const step = Math.max(0, Math.min(fragmentSteps[index] || 0, fragments.length));
    fragmentSteps[index] = step;
    fragments.forEach(function (fragment, i) {
      const shown = i < step;
      fragment.classList.toggle('is-revealed', shown);
      fragment.setAttribute('aria-hidden', shown ? 'false' : 'true');
    });
  }

  function revealNext() {
    const count = fragmentsAt(cur).length;
    if ((fragmentSteps[cur] || 0) >= count) return false;
    fragmentSteps[cur] = (fragmentSteps[cur] || 0) + 1;
    applyFragmentStep(cur);
    updateHud();
    return true;
  }

  function concealPrevious() {
    if ((fragmentSteps[cur] || 0) <= 0) return false;
    fragmentSteps[cur] -= 1;
    applyFragmentStep(cur);
    updateHud();
    return true;
  }

  // ── Syntax highlighting ──────────────────────────────────────────────
  // Guarded: the highlighter comes off a CDN, and a deck presented offline
  // should still navigate rather than die on a missing global.
  if (window.hljs) {
    document.querySelectorAll('pre code:not(.language-mermaid):not(.lang-mermaid)').forEach(function (block) {
      try { window.hljs.highlightElement(block); } catch (e) {}
    });
  }

  const richReady = window.deckrunRenderRichContent
    ? window.deckrunRenderRichContent(document.getElementById('presentation'))
    : Promise.resolve();

  // ── Session highlights ───────────────────────────────────────────────
  // Whatever was highlighted in the editor's preview is already in this
  // browser session; hl=... names the document it belongs to, so the deck
  // opens with the same marks and comments already on the slides. Read-only
  // here: selecting text while presenting should not pop up "highlight" /
  // "highlight + comment" — that flow belongs to the editor's preview.
  if (window.deckrunHighlights) {
    const hlParam = new URLSearchParams(location.search).get('hl');
    window.deckrunHighlights.mount({
      doc: document,
      docKey: hlParam || 'default',
      scopes: 'slides',
      readOnly: true
    });
  }

  // ── Slide navigation ─────────────────────────────────────────────────
  function showSlide(next, direction, fragmentMode) {
    const prev = cur;
    if (next < 0 || next >= total) return;

    if (next === prev && !fragmentMode) return;

    const nextFragments = fragmentsAt(next).length;
    fragmentSteps[next] = fragmentMode === 'start' || (fragmentMode !== 'end' && direction === 'forward')
      ? 0
      : nextFragments;
    applyFragmentStep(next);

    if (next === prev) {
      updateHud();
      return;
    }

    const slideOut = slides[prev];
    const slideIn  = slides[next];

    // Set up entering slide position
    const enterClass = direction === 'forward' ? 'enter-from-right' : 'enter-from-left';
    const exitClass  = direction === 'forward' ? 'exit-left'        : 'exit-right';

    slideIn.classList.add(enterClass);
    slideIn.style.transition = 'none';

    // Force reflow so the initial position is painted
    void slideIn.offsetWidth;

    slideIn.style.transition = '';
    slideIn.classList.remove(enterClass);
    slideIn.classList.add('is-active');

    slideOut.classList.remove('is-active');
    slideOut.classList.add(exitClass);

    // Clean up exit class after transition
    var cleaned = false;
    function cleanup() {
      if (cleaned) return;
      cleaned = true;
      slideOut.classList.remove(exitClass, 'exit-left', 'exit-right');
      slideOut.removeEventListener('transitionend', cleanup);
    }
    slideOut.addEventListener('transitionend', cleanup);
    // The no-motion preset has no transitionend event. The timeout also guards interrupted
    // transitions, so no slide can retain an exit class indefinitely.
    if (document.documentElement.dataset.transition === 'none') cleanup();
    else setTimeout(cleanup, 650);

    cur = next;
    updateHud();
  }

  function updateHud() {
    redrawBoard();
    elCur.textContent = String(cur + 1);
    const pct = total > 1 ? (cur / (total - 1)) * 100 : 100;
    elFill.style.width = pct + '%';
    const step = fragmentSteps[cur] || 0;
    const count = fragmentsAt(cur).length;
    elBtnPrev.style.opacity = cur === 0 && step === 0 ? '0.2' : '1';
    elBtnNext.style.opacity = cur === total - 1 && step >= count ? '0.2' : '1';
    post({ type: 'state', index: cur, total: total });
  }

  function next() {
    if (!revealNext()) showSlide(cur + 1, 'forward');
  }

  function prev() {
    if (!concealPrevious()) showSlide(cur - 1, 'backward');
  }

  // ── Editor mirror ─────────────────────────────────────────────────────
  // The editor that presented this tab is a peer of it: this tab feeds it,
  // over BroadcastChannel, the slides on screen plus the speaker notes. The
  // session id comes from the URL the editor opened (ps=...), so only a deck
  // an editor actually presented listens; one opened straight from a file
  // skips the machinery entirely.
  const psParam = new URLSearchParams(location.search).get('ps');
  const sid = (psParam && /^[A-Za-z0-9-]{1,64}$/.test(psParam)) ? psParam : null;
  const channel = (sid && typeof BroadcastChannel !== 'undefined')
    ? new BroadcastChannel('deckrun:' + sid)
    : null;

  function post(msg) {
    if (!channel) return;
    msg.id = sid;
    channel.postMessage(msg);
  }

  if (channel) {
    const NOTES = (function readNotes() {
      try {
        return JSON.parse(document.getElementById('deck-notes').textContent);
      } catch (e) {
        return [];
      }
    })();

    channel.onmessage = function (e) {
      const m = e.data || {};
      if (m.id !== sid) return;
      if (m.type === 'ready') {
        post({ type: 'init', notes: NOTES });
        post({ type: 'state', index: cur, total: total });
      } else if (m.type === 'goto') {
        const i = parseInt(m.index, 10);
        if (!isNaN(i) && i >= 0 && i < total && i !== cur) {
          showSlide(i, i > cur ? 'forward' : 'backward');
        }
      }
    };
  }

  // ── Overview mode ────────────────────────────────────────────────────
  function buildOverview() {
    elOverview.innerHTML = '';
    slides.forEach((slide, i) => {
      const thumb = document.createElement('div');
      thumb.className = 'overview-thumb' + (i === cur ? ' is-current' : '');

      const num = document.createElement('span');
      num.className = 'overview-thumb__number';
      num.textContent = String(i + 1);

      // Clone slide content into thumbnail
      const inner = document.createElement('div');
      inner.className = 'overview-thumb__inner';
      inner.style.width  = window.innerWidth  + 'px';
      inner.style.height = window.innerHeight + 'px';
      const clone = slide.cloneNode(true);
      clone.classList.add('is-active');
      clone.style.transition = 'none';
      clone.querySelectorAll('.fragment').forEach(function (fragment) {
        fragment.classList.add('is-revealed');
        fragment.setAttribute('aria-hidden', 'false');
      });
      inner.appendChild(clone);

      thumb.appendChild(num);
      thumb.appendChild(inner);

      thumb.addEventListener('click', () => {
        const direction = i >= cur ? 'forward' : 'backward';
        toggleOverview(false);
        showSlide(i, direction);
      });

      elOverview.appendChild(thumb);
    });
  }

  function toggleOverview(force) {
    inOverview = force !== undefined ? force : !inOverview;
    if (inOverview) {
      buildOverview();
      elOverview.classList.remove('hidden');
    } else {
      elOverview.classList.add('hidden');
    }
  }

  // ── Presenter tools ───────────────────────────────────────────────────
  // One canvas serves both drawing modes: the pen annotates over the live
  // slide, and the blank canvas paints the same board opaque so the slide
  // disappears behind it. Strokes are kept per slide in normalised
  // coordinates, so a resize or a jump into fullscreen keeps them in place.
  const ctx = elBoard.getContext('2d');
  const rootStyle = getComputedStyle(document.documentElement);

  function themeColor(name, fallback) {
    const v = rootStyle.getPropertyValue('--' + name).trim();
    return v || fallback;
  }

  const PEN_COLORS = [
    themeColor('red',    '#f38ba8'),
    themeColor('yellow', '#f9e2af'),
    themeColor('green',  '#a6e3a1'),
    themeColor('blue',   '#89b4fa'),
    themeColor('text',   '#cdd6f4'),
  ];
  const PEN_WIDTHS = [2, 3, 4, 6, 9, 14];
  const ERASER_SCALE = 5;

  const strokes = [];           // strokes[slideIndex] = [{ color, width, erase, pts }]
  let colorIdx = 0;
  let widthIdx = 2;
  let stroke = null;            // the stroke being drawn right now

  let penOn   = false;
  let blankOn = false;
  let eraseOn = false;
  let laserOn = false;
  let blackOn = false;
  let helpOn  = false;

  const tools = {
    laser: document.getElementById('btn-laser'),
    pen:   document.getElementById('btn-pen'),
    blank: document.getElementById('btn-blank'),
    black: document.getElementById('btn-black'),
    theme: document.getElementById('btn-theme'),
    erase: document.getElementById('btn-erase'),
  };

  /** Click handler that drops focus, so Space keeps meaning "next slide". */
  function onClick(el, fn) {
    if (!el) return;
    el.addEventListener('click', function (e) {
      el.blur();
      fn(e);
    });
  }

  // ── Canvas sizing and painting ───────────────────────────────────────
  function sizeBoard() {
    const dpr = window.devicePixelRatio || 1;
    elBoard.width  = Math.round(window.innerWidth  * dpr);
    elBoard.height = Math.round(window.innerHeight * dpr);
    elBoard.style.width  = window.innerWidth  + 'px';
    elBoard.style.height = window.innerHeight + 'px';
    ctx.setTransform(dpr, 0, 0, dpr, 0, 0);
    redrawBoard();
  }

  function strokeStyle(s) {
    ctx.globalCompositeOperation = s.erase ? 'destination-out' : 'source-over';
    ctx.strokeStyle = s.color;
    ctx.lineWidth = s.erase ? s.width * ERASER_SCALE : s.width;
    ctx.lineJoin = 'round';
    ctx.lineCap = 'round';
  }

  function paintStroke(s) {
    const w = window.innerWidth, h = window.innerHeight;
    if (!s.pts.length) return;
    strokeStyle(s);
    ctx.beginPath();
    ctx.moveTo(s.pts[0][0] * w, s.pts[0][1] * h);
    if (s.pts.length === 1) {
      // A tap still deserves a dot.
      ctx.lineTo(s.pts[0][0] * w + 0.01, s.pts[0][1] * h);
    } else {
      for (let i = 1; i < s.pts.length; i++) ctx.lineTo(s.pts[i][0] * w, s.pts[i][1] * h);
    }
    ctx.stroke();
  }

  /** Draw only the newest segment — repainting everything on every move is
      wasteful once a slide carries a few dozen strokes. */
  function paintTip(s) {
    const w = window.innerWidth, h = window.innerHeight;
    const n = s.pts.length;
    if (n < 2) { paintStroke(s); return; }
    strokeStyle(s);
    ctx.beginPath();
    ctx.moveTo(s.pts[n - 2][0] * w, s.pts[n - 2][1] * h);
    ctx.lineTo(s.pts[n - 1][0] * w, s.pts[n - 1][1] * h);
    ctx.stroke();
  }

  function redrawBoard() {
    ctx.globalCompositeOperation = 'source-over';
    ctx.clearRect(0, 0, window.innerWidth, window.innerHeight);
    const list = strokes[cur] || [];
    for (let i = 0; i < list.length; i++) paintStroke(list[i]);
    ctx.globalCompositeOperation = 'source-over';
  }

  function hasInk() { return !!(strokes[cur] && strokes[cur].length); }

  // ── Drawing ──────────────────────────────────────────────────────────
  function pointOf(e) {
    return [e.clientX / window.innerWidth, e.clientY / window.innerHeight];
  }

  elBoard.addEventListener('pointerdown', function (e) {
    if (!penOn) return;
    e.preventDefault();
    try { elBoard.setPointerCapture(e.pointerId); } catch (err) {}
    stroke = {
      color: PEN_COLORS[colorIdx],
      width: PEN_WIDTHS[widthIdx],
      erase: eraseOn,
      pts: [pointOf(e)],
    };
    if (!strokes[cur]) strokes[cur] = [];
    strokes[cur].push(stroke);
    paintStroke(stroke);
  });

  elBoard.addEventListener('pointermove', function (e) {
    if (!stroke) return;
    e.preventDefault();
    stroke.pts.push(pointOf(e));
    paintTip(stroke);
  });

  function endStroke() {
    if (!stroke) return;
    stroke = null;
    ctx.globalCompositeOperation = 'source-over';
    syncTools();
  }

  elBoard.addEventListener('pointerup', endStroke);
  elBoard.addEventListener('pointercancel', endStroke);
  elBoard.addEventListener('pointerleave', endStroke);

  function undoStroke() {
    const list = strokes[cur];
    if (!list || !list.length) return;
    list.pop();
    redrawBoard();
    syncTools();
  }

  function clearSlide() {
    strokes[cur] = [];
    redrawBoard();
    syncTools();
  }

  // ── Tool state ───────────────────────────────────────────────────────
  function setPen(on) {
    penOn = !!on;
    if (!penOn) {
      endStroke();
      // The blank canvas has no meaning without a pen to use on it.
      blankOn = false;
      eraseOn = false;
    }
    syncTools();
  }

  function setBlank(on) {
    blankOn = !!on;
    // Opening the blank canvas arms the pen; closing it leaves the pen alone.
    if (blankOn) penOn = true;
    syncTools();
  }

  function setEraser(on) {
    eraseOn = !!on;
    if (eraseOn) penOn = true;
    syncTools();
  }

  function setLaser(on) {
    laserOn = !!on;
    syncTools();
  }

  function setBlack(on) {
    blackOn = !!on;
    syncTools();
  }

  function setColor(i) {
    colorIdx = Math.max(0, Math.min(PEN_COLORS.length - 1, i));
    eraseOn = false;
    penOn = true;
    syncTools();
  }

  function nudgeWidth(delta) {
    widthIdx = Math.max(0, Math.min(PEN_WIDTHS.length - 1, widthIdx + delta));
    syncTools();
  }

  const swatches = [];
  (function buildSwatches() {
    const host = document.getElementById('pen-swatches');
    PEN_COLORS.forEach(function (color, i) {
      const b = document.createElement('button');
      b.className = 'swatch';
      b.style.background = color;
      b.title = 'Pen color ' + (i + 1);
      onClick(b, function () { setColor(i); });
      host.appendChild(b);
      swatches.push(b);
    });
  })();

  function syncTools() {
    tools.laser.classList.toggle('is-on', laserOn);
    tools.pen.classList.toggle('is-on', penOn);
    tools.blank.classList.toggle('is-on', blankOn);
    tools.black.classList.toggle('is-on', blackOn);
    if (tools.theme) tools.theme.classList.toggle('is-on', themesOpen());
    tools.erase.classList.toggle('is-on', eraseOn);

    elBoard.classList.toggle('is-drawing', penOn);
    elBoard.classList.toggle('is-erasing', penOn && eraseOn);
    elBoard.classList.toggle('is-blank', blankOn);

    elPenBar.classList.toggle('is-on', penOn);
    elPenWidth.textContent = PEN_WIDTHS[widthIdx] + 'px';
    swatches.forEach(function (b, i) {
      b.classList.toggle('is-on', !eraseOn && i === colorIdx);
    });

    elLaser.classList.toggle('is-on', laserOn);
    document.body.classList.toggle('laser-on', laserOn);
    elBlack.classList.toggle('is-on', blackOn);
    elHelp.classList.toggle('is-on', helpOn);
  }

  // ── Theme picker ─────────────────────────────────────────────────────
  let themeBootstrap = { themes: [], hljsMap: {}, decorMap: {} };
  try {
    const elBt = document.getElementById('deck-themes');
    if (elBt) themeBootstrap = JSON.parse(elBt.textContent || '{}');
  } catch (e) {}
  const themeList = themeBootstrap.themes || [];
  const hljsMap = themeBootstrap.hljsMap || {};
  const decorMap = themeBootstrap.decorMap || {};
  const THEME_DATA = Object.fromEntries(themeList.map(function (t) { return [t.id, t]; }));
  let activeTheme = document.documentElement.dataset.theme || 'nord';
  let themeCommitted = activeTheme;
  let themeSel = 0;
  let themeCards = [];

  function themeThumb(t) {
    const c = t.colors;
    const wrap = document.createElement('div');
    wrap.className = 'th-thumb';
    wrap.style.background = c.crust;
    wrap.style.color = c.text;

    const glow = document.createElement('div');
    glow.className = 'th-thumb__glow';
    glow.style.background = 'radial-gradient(circle, ' + c.accent + ' 0%, transparent 70%)';
    glow.style.opacity = t.mood === 'dark' ? '0.22' : '0.12';
    wrap.appendChild(glow);

    const title = document.createElement('div');
    title.className = 'th-thumb__title';
    title.textContent = t.label;
    title.style.fontFamily = t.fonts.display;
    title.style.color = c.accent;
    wrap.appendChild(title);

    const rule = document.createElement('div');
    rule.className = 'th-thumb__rule';
    rule.style.background = c.accent2;
    wrap.appendChild(rule);

    [75, 55].forEach(function (w) {
      const line = document.createElement('div');
      line.className = 'th-thumb__line';
      line.style.width = w + '%';
      line.style.background = c.surface0;
      wrap.appendChild(line);
    });

    const code = document.createElement('div');
    code.className = 'th-thumb__code';
    code.textContent = 'const deck = md';
    code.style.fontFamily = t.fonts.mono;
    code.style.background = c.mantle;
    code.style.color = c.accent3;
    code.style.border = '1px solid ' + c.surface0;
    wrap.appendChild(code);

    return wrap;
  }

  function themeCard(t) {
    const card = document.createElement('button');
    card.className = 'th-card' + (t.id === activeTheme ? ' is-current' : '');
    card.dataset.theme = t.id;
    card.appendChild(themeThumb(t));

    const meta = document.createElement('div');
    meta.className = 'th-meta';

    const top = document.createElement('div');
    top.className = 'th-meta__top';
    const name = document.createElement('span');
    name.className = 'th-meta__name';
    name.textContent = t.label;
    const mood = document.createElement('span');
    mood.className = 'th-meta__mood';
    mood.textContent = t.mood;
    top.appendChild(name);
    top.appendChild(mood);

    const blurb = document.createElement('div');
    blurb.className = 'th-meta__blurb';
    blurb.textContent = t.blurb;

    meta.appendChild(top);
    meta.appendChild(blurb);
    card.appendChild(meta);
    return card;
  }

  function buildThemePicker() {
    const host = document.getElementById('themes__list');
    if (!host) return;
    host.innerHTML = '';
    themeCards = [];

    [['dark', 'dark palettes'], ['light', 'light palettes']].forEach(function (pair) {
      const list = themeList.filter(function (t) { return t.mood === pair[0]; });
      if (!list.length) return;

      const label = document.createElement('div');
      label.className = 'th-group';
      label.textContent = pair[1];
      host.appendChild(label);

      const grid = document.createElement('div');
      grid.className = 'th-grid';
      list.forEach(function (t) {
        const card = themeCard(t);
        card.addEventListener('mouseenter', function () { selectTheme(themeCards.indexOf(card)); });
        card.addEventListener('click', function () { selectTheme(themeCards.indexOf(card)); commitTheme(); });
        grid.appendChild(card);
        themeCards.push(card);
      });
      host.appendChild(grid);
    });
  }

  function updateThemePenColors() {
    const rs = getComputedStyle(document.documentElement);
    function tc(name, fb) {
      const v = rs.getPropertyValue('--' + name).trim();
      return v || fb;
    }
    PEN_COLORS[0] = tc('red',    '#f38ba8');
    PEN_COLORS[1] = tc('yellow', '#f9e2af');
    PEN_COLORS[2] = tc('green',  '#a6e3a1');
    PEN_COLORS[3] = tc('blue',   '#89b4fa');
    PEN_COLORS[4] = tc('text',   '#cdd6f4');
    swatches.forEach(function (sw, i) {
      if (sw && PEN_COLORS[i]) sw.style.background = PEN_COLORS[i];
    });
  }

  function setTheme(id, remember) {
    if (!id || !THEME_DATA[id]) return;
    activeTheme = id;
    document.documentElement.dataset.theme = id;
    document.documentElement.dataset.decor = decorMap[id] || 'orbs';
    const hljsLink = document.getElementById('hljs-theme');
    if (hljsLink && hljsMap[id]) {
      hljsLink.href = hljsMap[id];
    }
    if (window.deckrunRenderRichContent) {
      window.deckrunRenderRichContent(document.getElementById('presentation'));
    }
    updateThemePenColors();
    if (remember !== false) {
      try { localStorage.setItem('deckrun.theme.v1', id); } catch (e) {}
    }
    if (themeCards.length) {
      themeCards.forEach(function (card) {
        card.classList.toggle('is-current', card.dataset.theme === id);
      });
    }
    post({ type: 'theme', theme: id });
  }

  function selectTheme(i) {
    if (i < 0 || i >= themeCards.length) return;
    themeSel = i;
    themeCards.forEach(function (c, n) { c.classList.toggle('is-sel', n === i); });
    setTheme(themeCards[i].dataset.theme, false);
  }

  function openThemes() {
    if (helpOn) setHelp(false);
    if (inOverview) toggleOverview(false);
    themeCommitted = activeTheme;
    buildThemePicker();
    let at = 0;
    themeCards.forEach(function (c, n) { if (c.dataset.theme === activeTheme) at = n; });
    elThemes.classList.add('is-on');
    syncTools();
    selectTheme(at);
    if (themeCards[themeSel]) themeCards[themeSel].scrollIntoView({ block: 'nearest' });
  }

  function commitTheme() {
    themeCommitted = activeTheme;
    setTheme(activeTheme, true);
    closeThemes(false);
  }

  function closeThemes(restore) {
    if (restore && themeCommitted && themeCommitted !== activeTheme) {
      setTheme(themeCommitted, true);
    }
    elThemes.classList.remove('is-on');
    syncTools();
    themeCards.forEach(function (c) { c.classList.remove('is-sel'); });
  }

  function themesOpen() {
    return elThemes ? elThemes.classList.contains('is-on') : false;
  }

  // ── Laser pointer ────────────────────────────────────────────────────
  document.addEventListener('pointermove', function (e) {
    if (!laserOn) return;
    elLaser.style.transform = 'translate(' + e.clientX + 'px, ' + e.clientY + 'px)';
  }, { passive: true });

  // ── Controls overlay ─────────────────────────────────────────────────
  const HELP_GROUPS = [
    { title: 'navigate', rows: [
      { keys: ['→', '↓', 'Space'],        desc: 'Next reveal or slide' },
      { keys: ['←', '↑', 'Backspace'],    desc: 'Previous reveal or slide' },
      { keys: ['Home'],                   desc: 'First slide' },
      { keys: ['End'],                    desc: 'Last slide' },
      { keys: ['O'],                      desc: 'Overview grid' },
      { keys: ['Esc'],                    desc: 'Close what is open' },
    ]},
    { title: 'screen', rows: [
      { keys: ['F'],                      desc: 'Fullscreen' },
      { keys: ['T'],                      desc: 'Theme picker' },
      { keys: ['B'],                      desc: 'Black out the screen' },
      { keys: ['?'],                      desc: 'These controls' },
    ]},
    { title: 'point', rows: [
      { keys: ['L'],                      desc: 'Laser pointer' },
    ]},
    { title: 'mark up', rows: [
      { keys: ['select'],                 desc: 'Highlight, with or without a comment' },
      { keys: ['click'],                  desc: 'Edit or remove a highlight' },
    ]},
    { title: 'draw', rows: [
      { keys: ['D'],                      desc: 'Pen, over the slide' },
      { keys: ['C'],                      desc: 'Blank canvas' },
      { keys: ['1', '2', '3', '4', '5'],  desc: 'Pen color' },
      { keys: ['E'],                      desc: 'Eraser' },
      { keys: ['['], desc: 'Thinner' },
      { keys: [']'], desc: 'Thicker' },
      { keys: ['Ctrl', 'Z'],              desc: 'Undo last stroke' },
      { keys: ['X'],                      desc: 'Clear this slide' },
    ]},
  ];

  (function buildHelp() {
    const grid = document.getElementById('help__grid');
    HELP_GROUPS.forEach(function (group) {
      const box = document.createElement('div');
      box.className = 'help-group';

      const title = document.createElement('div');
      title.className = 'help-group__title';
      title.textContent = group.title;
      box.appendChild(title);

      group.rows.forEach(function (row) {
        const line = document.createElement('div');
        line.className = 'help-row';

        const desc = document.createElement('span');
        desc.textContent = row.desc;

        const keys = document.createElement('span');
        keys.className = 'help-row__keys';
        row.keys.forEach(function (k) {
          const kbd = document.createElement('kbd');
          kbd.textContent = k;
          keys.appendChild(kbd);
        });

        line.appendChild(desc);
        line.appendChild(keys);
        box.appendChild(line);
      });

      grid.appendChild(box);
    });
  })();

  function setHelp(on) {
    helpOn = !!on;
    syncTools();
  }

  function toggleFullscreen() {
    if (!document.fullscreenElement) document.documentElement.requestFullscreen().catch(() => {});
    else document.exitFullscreen().catch(() => {});
  }

  // ── Tool wiring ──────────────────────────────────────────────────────
  onClick(tools.laser, function () { setLaser(!laserOn); });
  onClick(tools.pen,   function () { setPen(!penOn); });
  onClick(tools.blank, function () { setBlank(!blankOn); });
  onClick(tools.black, function () { setBlack(true); });
  onClick(tools.erase, function () { setEraser(!eraseOn); });
  onClick(tools.theme, function () { themesOpen() ? closeThemes(false) : openThemes(); });
  onClick(document.getElementById('btn-thin'),  function () { nudgeWidth(-1); });
  onClick(document.getElementById('btn-thick'), function () { nudgeWidth(1); });
  onClick(document.getElementById('btn-clear'), function () { clearSlide(); });
  onClick(document.getElementById('btn-help'),  function () { setHelp(!helpOn); });
  onClick(elBlack, function () { setBlack(false); });

  Array.prototype.forEach.call(elThemes.querySelectorAll('[data-close="themes"]'), function (el) {
    onClick(el, function () { closeThemes(false); });
  });

  Array.prototype.forEach.call(elHelp.querySelectorAll('[data-close="help"]'), function (el) {
    onClick(el, function () { setHelp(false); });
  });

  window.addEventListener('resize', sizeBoard);

  // ── Keyboard ─────────────────────────────────────────────────────────
  document.addEventListener('keydown', function (e) {
    const k = e.key;

    // Undo and PDF/Print modifier combos
    if (e.metaKey || e.ctrlKey || e.altKey) {
      if ((k === 'z' || k === 'Z') && hasInk()) {
        e.preventDefault();
        undoStroke();
        return;
      }
      if (k === 'p' || k === 'P') {
        e.preventDefault();
        triggerPdfExport();
        return;
      }
      return;
    }

    // Each overlay swallows keys until it is dismissed, outermost first.
    if (themesOpen()) {
      if (k === 'Escape') {
        e.preventDefault();
        closeThemes(true);
        return;
      }
      if (k === 'Enter') {
        e.preventDefault();
        commitTheme();
        return;
      }
      if (k === 'ArrowRight' || k === 'ArrowDown') {
        e.preventDefault();
        const nextIdx = (themeSel + 1) % themeCards.length;
        selectTheme(nextIdx);
        if (themeCards[nextIdx]) themeCards[nextIdx].scrollIntoView({ block: 'nearest' });
        return;
      }
      if (k === 'ArrowLeft' || k === 'ArrowUp') {
        e.preventDefault();
        const prevIdx = (themeSel - 1 + themeCards.length) % themeCards.length;
        selectTheme(prevIdx);
        if (themeCards[prevIdx]) themeCards[prevIdx].scrollIntoView({ block: 'nearest' });
        return;
      }
      if (k === 't' || k === 'T') {
        e.preventDefault();
        closeThemes(false);
        return;
      }
      return;
    }

    if (helpOn) {
      if (k === 'Escape' || k === '?' || k === 'h' || k === 'H') {
        e.preventDefault();
        setHelp(false);
      }
      return;
    }
    if (k === '?' || k === 'h' || k === 'H') {
      e.preventDefault();
      setHelp(true);
      return;
    }

    if (blackOn) {
      // A stray key should not advance the deck behind a black screen.
      e.preventDefault();
      if (k === 'Escape' || k === 'b' || k === 'B' || k === ' ' || k === 'Enter') setBlack(false);
      return;
    }
    if (k === 'b' || k === 'B') {
      e.preventDefault();
      setBlack(true);
      return;
    }

    if (inOverview) {
      if (k === 'Escape' || k === 'o' || k === 'O') {
        e.preventDefault();
        toggleOverview(false);
      }
      return;
    }

    // Pen sub-controls only bind while the pen is down, so the letters stay
    // free for everything else the rest of the time.
    if (penOn) {
      if (k >= '1' && k <= String(PEN_COLORS.length)) { e.preventDefault(); setColor(parseInt(k, 10) - 1); return; }
      if (k === 'e' || k === 'E') { e.preventDefault(); setEraser(!eraseOn); return; }
      if (k === '[') { e.preventDefault(); nudgeWidth(-1); return; }
      if (k === ']') { e.preventDefault(); nudgeWidth(1); return; }
      if (k === 'x' || k === 'X') { e.preventDefault(); clearSlide(); return; }
    }

    switch (k) {
      case 'ArrowRight':
      case 'ArrowDown':
      case ' ':
      case 'PageDown':
        e.preventDefault();
        next();
        break;
      case 'ArrowLeft':
      case 'ArrowUp':
      case 'Backspace':
      case 'PageUp':
        e.preventDefault();
        prev();
        break;
      case 'Home':
        e.preventDefault();
        showSlide(0, 'backward', 'start');
        break;
      case 'End':
        e.preventDefault();
        showSlide(total - 1, 'forward', 'end');
        break;
      case 'f':
      case 'F':
        e.preventDefault();
        toggleFullscreen();
        break;
      case 't':
      case 'T':
        e.preventDefault();
        openThemes();
        break;
      case 'l':
      case 'L':
        e.preventDefault();
        setLaser(!laserOn);
        break;
      case 'd':
      case 'D':
        e.preventDefault();
        setPen(!penOn);
        break;
      case 'c':
      case 'C':
        e.preventDefault();
        setBlank(!blankOn);
        break;
      case 'o':
      case 'O':
        e.preventDefault();
        toggleOverview();
        break;
      case 'Escape':
        e.preventDefault();
        // Peel one layer at a time: themes, canvas, pen, laser, then the overview.
        if (themesOpen()) closeThemes(true);
        else if (blankOn) setBlank(false);
        else if (penOn) setPen(false);
        else if (laserOn) setLaser(false);
        else toggleOverview();
        break;
    }
  });

  // ── Mouse/touch ───────────────────────────────────────────────────────
  onClick(elBtnPrev, prev);
  onClick(elBtnNext, next);

  let wheelLock = false;
  window.addEventListener('wheel', function (e) {
    if (inOverview || themesOpen() || helpOn || blackOn) return;
    const delta = Math.abs(e.deltaY) >= Math.abs(e.deltaX) ? e.deltaY : e.deltaX;
    if (Math.abs(delta) < 15) return;
    if (wheelLock) return;
    wheelLock = true;
    setTimeout(function () { wheelLock = false; }, 250);
    if (delta > 0) {
      next();
    } else if (delta < 0) {
      prev();
    }
  }, { passive: true });

  let touchStartX = 0;
  document.addEventListener('touchstart', (e) => { touchStartX = e.touches[0].clientX; }, { passive: true });
  document.addEventListener('touchend', (e) => {
    // A stroke is not a swipe. Touch events fire alongside the pointer events
    // the canvas draws with, so an unguarded swipe would change slides
    // underneath every horizontal line drawn on a touchscreen.
    if (penOn) return;
    const dx = e.changedTouches[0].clientX - touchStartX;
    if (Math.abs(dx) > 50) dx < 0 ? next() : prev();
  }, { passive: true });

  // ── Hint auto-hide ────────────────────────────────────────────────────
  setTimeout(() => { elHint.classList.add('hidden'); }, 4000);

  // ── Auto-fullscreen ───────────────────────────────────────────────────
  const fsHint = document.getElementById('fs-hint');
  if (fsHint) {
    function enterFullscreen() {
      fsHint.classList.add('hidden');
      document.documentElement.requestFullscreen().catch(() => {});
    }
    fsHint.addEventListener('click', enterFullscreen, { once: true });
    document.addEventListener('keydown', function fsKey(e) {
      // Let the click handler own the 'f' key if hint is still visible
      document.removeEventListener('keydown', fsKey);
      enterFullscreen();
    }, { once: true });
  }

  // ── PDF export / Print intercept ───────────────────────────────────────────
  function triggerPdfExport() {
    fetch('/__pdf', { method: 'POST', headers: { 'Content-Type': 'application/json' } })
      .then(function (r) {
        if (r.ok) return r.blob();
        throw new Error('Server PDF unavailable');
      })
      .then(function (blob) {
        var url = URL.createObjectURL(blob);
        var a = document.createElement('a');
        a.href = url;
        var title = (document.title || 'deck').replace(/[\s/\\?%*:|"<>]+/g, '-').toLowerCase();
        a.download = (title.endsWith('.pdf') ? title : title + '.pdf');
        document.body.appendChild(a);
        a.click();
        document.body.removeChild(a);
        setTimeout(function () { URL.revokeObjectURL(url); }, 5000);
      })
      .catch(function () {
        var fontsReady = document.fonts && document.fonts.ready ? document.fonts.ready : Promise.resolve();
        Promise.all([richReady, fontsReady]).then(function () {
          setTimeout(function () { window.print(); }, 150);
        });
      });
  }

  // ── Print export ─────────────────────────────────────────────────────
  // Loading the deck with ?print=1 opens the print dialog once fonts and
  // highlighting have settled. The editor's PDF export uses this.
  var wantsPrint = false;
  try { wantsPrint = new URLSearchParams(location.search).has('print'); } catch (e) {}
  if (wantsPrint) {
    var openPrint = function () { setTimeout(function () { window.print(); }, 350); };
    var fontsReady = document.fonts && document.fonts.ready ? document.fonts.ready : Promise.resolve();
    Promise.all([richReady, fontsReady]).then(openPrint, openPrint);
  }

  // ── Init ─────────────────────────────────────────────────────────────
  slides[0].classList.add('is-active');
  sizeBoard();
  syncTools();
  updateHud();
})();
`

const DocScript = `(function () {
  'use strict';

  const elBoard    = document.getElementById('board');
  const elLaser    = document.getElementById('laser');
  const elBlack    = document.getElementById('blackout');
  const elHelp     = document.getElementById('help');
  const elThemes   = document.getElementById('themes');
  const elPenBar   = document.getElementById('pen-bar');
  const elPenWidth = document.getElementById('pen-width');
  const elFrame    = document.getElementById('doc-frame');

  // ── Session highlights ────────────────────────────────────────────────
  // The doc itself is same-origin, so marks made in the editor's preview
  // still render here, on the presenter chrome's palette. Read-only: this is
  // the presenting surface, not the editor, so selecting text here should
  // not offer to create a highlight. hl=... names whose highlights these are.
  if (window.deckrunHighlights) {
    window.deckrunHighlights.mount({
      frame: elFrame,
      docKey: new URLSearchParams(location.search).get('hl') || 'default',
      scopes: 'doc',
      readOnly: true
    });
  }

  // ── Presenter tools ───────────────────────────────────────────────────
  // One flat stroke list — unlike the slide deck, there is only ever one
  // "page" here, so there is nothing to index annotations by.
  const ctx = elBoard.getContext('2d');
  const rootStyle = getComputedStyle(document.documentElement);

  function themeColor(name, fallback) {
    const v = rootStyle.getPropertyValue('--' + name).trim();
    return v || fallback;
  }

  const PEN_COLORS = [
    themeColor('red',    '#f38ba8'),
    themeColor('yellow', '#f9e2af'),
    themeColor('green',  '#a6e3a1'),
    themeColor('blue',   '#89b4fa'),
    themeColor('text',   '#cdd6f4'),
  ];
  const PEN_WIDTHS = [2, 3, 4, 6, 9, 14];
  const ERASER_SCALE = 5;

  let strokes = [];             // flat list: [{ color, width, erase, pts }]
  let colorIdx = 0;
  let widthIdx = 2;
  let stroke = null;            // the stroke being drawn right now

  let penOn   = false;
  let blankOn = false;
  let eraseOn = false;
  let laserOn = false;
  let blackOn = false;
  let helpOn  = false;

  const tools = {
    laser: document.getElementById('btn-laser'),
    pen:   document.getElementById('btn-pen'),
    blank: document.getElementById('btn-blank'),
    black: document.getElementById('btn-black'),
    theme: document.getElementById('btn-theme'),
    erase: document.getElementById('btn-erase'),
  };

  /** Click handler that drops focus, so Space does not re-trigger it. */
  function onClick(el, fn) {
    if (!el) return;
    el.addEventListener('click', function (e) {
      el.blur();
      fn(e);
    });
  }

  // ── Canvas sizing and painting ───────────────────────────────────────
  function sizeBoard() {
    const dpr = window.devicePixelRatio || 1;
    elBoard.width  = Math.round(window.innerWidth  * dpr);
    elBoard.height = Math.round(window.innerHeight * dpr);
    elBoard.style.width  = window.innerWidth  + 'px';
    elBoard.style.height = window.innerHeight + 'px';
    ctx.setTransform(dpr, 0, 0, dpr, 0, 0);
    redrawBoard();
  }

  function strokeStyle(s) {
    ctx.globalCompositeOperation = s.erase ? 'destination-out' : 'source-over';
    ctx.strokeStyle = s.color;
    ctx.lineWidth = s.erase ? s.width * ERASER_SCALE : s.width;
    ctx.lineJoin = 'round';
    ctx.lineCap = 'round';
  }

  function paintStroke(s) {
    const w = window.innerWidth, h = window.innerHeight;
    if (!s.pts.length) return;
    strokeStyle(s);
    ctx.beginPath();
    ctx.moveTo(s.pts[0][0] * w, s.pts[0][1] * h);
    if (s.pts.length === 1) {
      // A tap still deserves a dot.
      ctx.lineTo(s.pts[0][0] * w + 0.01, s.pts[0][1] * h);
    } else {
      for (let i = 1; i < s.pts.length; i++) ctx.lineTo(s.pts[i][0] * w, s.pts[i][1] * h);
    }
    ctx.stroke();
  }

  /** Draw only the newest segment — repainting everything on every move is
      wasteful once the board carries a few dozen strokes. */
  function paintTip(s) {
    const w = window.innerWidth, h = window.innerHeight;
    const n = s.pts.length;
    if (n < 2) { paintStroke(s); return; }
    strokeStyle(s);
    ctx.beginPath();
    ctx.moveTo(s.pts[n - 2][0] * w, s.pts[n - 2][1] * h);
    ctx.lineTo(s.pts[n - 1][0] * w, s.pts[n - 1][1] * h);
    ctx.stroke();
  }

  function redrawBoard() {
    ctx.globalCompositeOperation = 'source-over';
    ctx.clearRect(0, 0, window.innerWidth, window.innerHeight);
    for (let i = 0; i < strokes.length; i++) paintStroke(strokes[i]);
    ctx.globalCompositeOperation = 'source-over';
  }

  function hasInk() { return strokes.length > 0; }

  // ── Drawing ──────────────────────────────────────────────────────────
  function pointOf(e) {
    return [e.clientX / window.innerWidth, e.clientY / window.innerHeight];
  }

  elBoard.addEventListener('pointerdown', function (e) {
    if (!penOn) return;
    e.preventDefault();
    try { elBoard.setPointerCapture(e.pointerId); } catch (err) {}
    stroke = {
      color: PEN_COLORS[colorIdx],
      width: PEN_WIDTHS[widthIdx],
      erase: eraseOn,
      pts: [pointOf(e)],
    };
    strokes.push(stroke);
    paintStroke(stroke);
  });

  elBoard.addEventListener('pointermove', function (e) {
    if (!stroke) return;
    e.preventDefault();
    stroke.pts.push(pointOf(e));
    paintTip(stroke);
  });

  function endStroke() {
    if (!stroke) return;
    stroke = null;
    ctx.globalCompositeOperation = 'source-over';
    syncTools();
  }

  elBoard.addEventListener('pointerup', endStroke);
  elBoard.addEventListener('pointercancel', endStroke);
  elBoard.addEventListener('pointerleave', endStroke);

  function undoStroke() {
    if (!strokes.length) return;
    strokes.pop();
    redrawBoard();
    syncTools();
  }

  function clearBoard() {
    strokes = [];
    redrawBoard();
    syncTools();
  }

  // ── Tool state ───────────────────────────────────────────────────────
  function setPen(on) {
    penOn = !!on;
    if (!penOn) {
      endStroke();
      // The blank canvas has no meaning without a pen to use on it.
      blankOn = false;
      eraseOn = false;
    }
    syncTools();
  }

  function setBlank(on) {
    blankOn = !!on;
    // Opening the blank canvas arms the pen; closing it leaves the pen alone.
    if (blankOn) penOn = true;
    syncTools();
  }

  function setEraser(on) {
    eraseOn = !!on;
    if (eraseOn) penOn = true;
    syncTools();
  }

  function setLaser(on) {
    laserOn = !!on;
    syncTools();
  }

  function setBlack(on) {
    blackOn = !!on;
    syncTools();
  }

  function setColor(i) {
    colorIdx = Math.max(0, Math.min(PEN_COLORS.length - 1, i));
    eraseOn = false;
    penOn = true;
    syncTools();
  }

  function nudgeWidth(delta) {
    widthIdx = Math.max(0, Math.min(PEN_WIDTHS.length - 1, widthIdx + delta));
    syncTools();
  }

  const swatches = [];
  (function buildSwatches() {
    const host = document.getElementById('pen-swatches');
    PEN_COLORS.forEach(function (color, i) {
      const b = document.createElement('button');
      b.className = 'swatch';
      b.style.background = color;
      b.title = 'Pen color ' + (i + 1);
      onClick(b, function () { setColor(i); });
      host.appendChild(b);
      swatches.push(b);
    });
  })();

  function syncTools() {
    tools.laser.classList.toggle('is-on', laserOn);
    tools.pen.classList.toggle('is-on', penOn);
    tools.blank.classList.toggle('is-on', blankOn);
    tools.black.classList.toggle('is-on', blackOn);
    if (tools.theme) tools.theme.classList.toggle('is-on', themesOpen());
    tools.erase.classList.toggle('is-on', eraseOn);

    elBoard.classList.toggle('is-drawing', penOn);
    elBoard.classList.toggle('is-erasing', penOn && eraseOn);
    elBoard.classList.toggle('is-blank', blankOn);

    elPenBar.classList.toggle('is-on', penOn);
    elPenWidth.textContent = PEN_WIDTHS[widthIdx] + 'px';
    swatches.forEach(function (b, i) {
      b.classList.toggle('is-on', !eraseOn && i === colorIdx);
    });

    elLaser.classList.toggle('is-on', laserOn);
    document.body.classList.toggle('laser-on', laserOn);
    elBlack.classList.toggle('is-on', blackOn);
    elHelp.classList.toggle('is-on', helpOn);
  }

  // ── Theme picker ─────────────────────────────────────────────────────
  let themeBootstrap = { themes: [], hljsMap: {}, decorMap: {} };
  try {
    const elBt = document.getElementById('deck-themes');
    if (elBt) themeBootstrap = JSON.parse(elBt.textContent || '{}');
  } catch (e) {}
  const themeList = themeBootstrap.themes || [];
  const hljsMap = themeBootstrap.hljsMap || {};
  const decorMap = themeBootstrap.decorMap || {};
  const THEME_DATA = Object.fromEntries(themeList.map(function (t) { return [t.id, t]; }));
  let activeTheme = document.documentElement.dataset.theme || 'nord';
  let themeCommitted = activeTheme;
  let themeSel = 0;
  let themeCards = [];

  function themeThumb(t) {
    const c = t.colors;
    const wrap = document.createElement('div');
    wrap.className = 'th-thumb';
    wrap.style.background = c.crust;
    wrap.style.color = c.text;

    const glow = document.createElement('div');
    glow.className = 'th-thumb__glow';
    glow.style.background = 'radial-gradient(circle, ' + c.accent + ' 0%, transparent 70%)';
    glow.style.opacity = t.mood === 'dark' ? '0.22' : '0.12';
    wrap.appendChild(glow);

    const title = document.createElement('div');
    title.className = 'th-thumb__title';
    title.textContent = t.label;
    title.style.fontFamily = t.fonts.display;
    title.style.color = c.accent;
    wrap.appendChild(title);

    const rule = document.createElement('div');
    rule.className = 'th-thumb__rule';
    rule.style.background = c.accent2;
    wrap.appendChild(rule);

    [75, 55].forEach(function (w) {
      const line = document.createElement('div');
      line.className = 'th-thumb__line';
      line.style.width = w + '%';
      line.style.background = c.surface0;
      wrap.appendChild(line);
    });

    const code = document.createElement('div');
    code.className = 'th-thumb__code';
    code.textContent = 'const deck = md';
    code.style.fontFamily = t.fonts.mono;
    code.style.background = c.mantle;
    code.style.color = c.accent3;
    code.style.border = '1px solid ' + c.surface0;
    wrap.appendChild(code);

    return wrap;
  }

  function themeCard(t) {
    const card = document.createElement('button');
    card.className = 'th-card' + (t.id === activeTheme ? ' is-current' : '');
    card.dataset.theme = t.id;
    card.appendChild(themeThumb(t));

    const meta = document.createElement('div');
    meta.className = 'th-meta';

    const top = document.createElement('div');
    top.className = 'th-meta__top';
    const name = document.createElement('span');
    name.className = 'th-meta__name';
    name.textContent = t.label;
    const mood = document.createElement('span');
    mood.className = 'th-meta__mood';
    mood.textContent = t.mood;
    top.appendChild(name);
    top.appendChild(mood);

    const blurb = document.createElement('div');
    blurb.className = 'th-meta__blurb';
    blurb.textContent = t.blurb;

    meta.appendChild(top);
    meta.appendChild(blurb);
    card.appendChild(meta);
    return card;
  }

  function buildThemePicker() {
    const host = document.getElementById('themes__list');
    if (!host) return;
    host.innerHTML = '';
    themeCards = [];

    [['dark', 'dark palettes'], ['light', 'light palettes']].forEach(function (pair) {
      const list = themeList.filter(function (t) { return t.mood === pair[0]; });
      if (!list.length) return;

      const label = document.createElement('div');
      label.className = 'th-group';
      label.textContent = pair[1];
      host.appendChild(label);

      const grid = document.createElement('div');
      grid.className = 'th-grid';
      list.forEach(function (t) {
        const card = themeCard(t);
        card.addEventListener('mouseenter', function () { selectTheme(themeCards.indexOf(card)); });
        card.addEventListener('click', function () { selectTheme(themeCards.indexOf(card)); commitTheme(); });
        grid.appendChild(card);
        themeCards.push(card);
      });
      host.appendChild(grid);
    });
  }

  function updateThemePenColors() {
    const rs = getComputedStyle(document.documentElement);
    function tc(name, fb) {
      const v = rs.getPropertyValue('--' + name).trim();
      return v || fb;
    }
    PEN_COLORS[0] = tc('red',    '#f38ba8');
    PEN_COLORS[1] = tc('yellow', '#f9e2af');
    PEN_COLORS[2] = tc('green',  '#a6e3a1');
    PEN_COLORS[3] = tc('blue',   '#89b4fa');
    PEN_COLORS[4] = tc('text',   '#cdd6f4');
    swatches.forEach(function (sw, i) {
      if (sw && PEN_COLORS[i]) sw.style.background = PEN_COLORS[i];
    });
  }

  function setTheme(id, remember) {
    if (!id || !THEME_DATA[id]) return;
    activeTheme = id;
    document.documentElement.dataset.theme = id;
    document.documentElement.dataset.decor = decorMap[id] || 'orbs';
    const hljsLink = document.getElementById('hljs-theme');
    if (hljsLink && hljsMap[id]) {
      hljsLink.href = hljsMap[id];
    }
    try {
      if (elFrame && elFrame.contentDocument && elFrame.contentDocument.documentElement) {
        elFrame.contentDocument.documentElement.dataset.theme = id;
      }
    } catch (e) {}
    try {
      if (elFrame && elFrame.contentWindow) {
        elFrame.contentWindow.postMessage({ type: 'theme', theme: id }, '*');
      }
    } catch (e) {}
    updateThemePenColors();
    if (remember !== false) {
      try { localStorage.setItem('deckrun.theme.v1', id); } catch (e) {}
    }
    if (themeCards.length) {
      themeCards.forEach(function (card) {
        card.classList.toggle('is-current', card.dataset.theme === id);
      });
    }
  }

  function selectTheme(i) {
    if (i < 0 || i >= themeCards.length) return;
    themeSel = i;
    themeCards.forEach(function (c, n) { c.classList.toggle('is-sel', n === i); });
    setTheme(themeCards[i].dataset.theme, false);
  }

  function openThemes() {
    if (helpOn) setHelp(false);
    themeCommitted = activeTheme;
    buildThemePicker();
    let at = 0;
    themeCards.forEach(function (c, n) { if (c.dataset.theme === activeTheme) at = n; });
    elThemes.classList.add('is-on');
    syncTools();
    selectTheme(at);
    if (themeCards[themeSel]) themeCards[themeSel].scrollIntoView({ block: 'nearest' });
  }

  function commitTheme() {
    themeCommitted = activeTheme;
    setTheme(activeTheme, true);
    closeThemes(false);
  }

  function closeThemes(restore) {
    if (restore && themeCommitted && themeCommitted !== activeTheme) {
      setTheme(themeCommitted, true);
    }
    elThemes.classList.remove('is-on');
    syncTools();
    themeCards.forEach(function (c) { c.classList.remove('is-sel'); });
  }

  function themesOpen() {
    return elThemes ? elThemes.classList.contains('is-on') : false;
  }

  // ── Laser pointer ────────────────────────────────────────────────────
  // Same-origin doc, no border on the iframe: client coordinates line up
  // with the outer viewport, so no translation is needed either way.
  function onPointerMove(e) {
    if (!laserOn) return;
    elLaser.style.transform = 'translate(' + e.clientX + 'px, ' + e.clientY + 'px)';
  }
  document.addEventListener('pointermove', onPointerMove, { passive: true });

  // ── Controls overlay ─────────────────────────────────────────────────
  const HELP_GROUPS = [
    { title: 'screen', rows: [
      { keys: ['F'],                      desc: 'Fullscreen' },
      { keys: ['T'],                      desc: 'Theme picker' },
      { keys: ['B'],                      desc: 'Black out the screen' },
      { keys: ['Esc'],                    desc: 'Close what is open' },
      { keys: ['?'],                      desc: 'These controls' },
    ]},
    { title: 'point', rows: [
      { keys: ['L'],                      desc: 'Laser pointer' },
    ]},
    { title: 'mark up', rows: [
      { keys: ['select'],                 desc: 'Highlight, with or without a comment' },
      { keys: ['click'],                  desc: 'Edit or remove a highlight' },
    ]},
    { title: 'draw', rows: [
      { keys: ['D'],                      desc: 'Pen, over the doc' },
      { keys: ['C'],                      desc: 'Blank canvas' },
      { keys: ['1', '2', '3', '4', '5'],  desc: 'Pen color' },
      { keys: ['E'],                      desc: 'Eraser' },
      { keys: ['['], desc: 'Thinner' },
      { keys: [']'], desc: 'Thicker' },
      { keys: ['Ctrl', 'Z'],              desc: 'Undo last stroke' },
      { keys: ['X'],                      desc: 'Clear' },
    ]},
  ];

  (function buildHelp() {
    const grid = document.getElementById('help__grid');
    HELP_GROUPS.forEach(function (group) {
      const box = document.createElement('div');
      box.className = 'help-group';

      const title = document.createElement('div');
      title.className = 'help-group__title';
      title.textContent = group.title;
      box.appendChild(title);

      group.rows.forEach(function (row) {
        const line = document.createElement('div');
        line.className = 'help-row';

        const desc = document.createElement('span');
        desc.textContent = row.desc;

        const keys = document.createElement('span');
        keys.className = 'help-row__keys';
        row.keys.forEach(function (k) {
          const kbd = document.createElement('kbd');
          kbd.textContent = k;
          keys.appendChild(kbd);
        });

        line.appendChild(desc);
        line.appendChild(keys);
        box.appendChild(line);
      });

      grid.appendChild(box);
    });
  })();

  function setHelp(on) {
    helpOn = !!on;
    syncTools();
  }

  function toggleFullscreen() {
    if (!document.fullscreenElement) document.documentElement.requestFullscreen().catch(() => {});
    else document.exitFullscreen().catch(() => {});
  }

  // ── Tool wiring ──────────────────────────────────────────────────────
  onClick(tools.laser, function () { setLaser(!laserOn); });
  onClick(tools.pen,   function () { setPen(!penOn); });
  onClick(tools.blank, function () { setBlank(!blankOn); });
  onClick(tools.black, function () { setBlack(true); });
  onClick(tools.erase, function () { setEraser(!eraseOn); });
  onClick(tools.theme, function () { themesOpen() ? closeThemes(false) : openThemes(); });
  onClick(document.getElementById('btn-thin'),  function () { nudgeWidth(-1); });
  onClick(document.getElementById('btn-thick'), function () { nudgeWidth(1); });
  onClick(document.getElementById('btn-clear'), function () { clearBoard(); });
  onClick(document.getElementById('btn-help'),  function () { setHelp(!helpOn); });
  onClick(elBlack, function () { setBlack(false); });

  Array.prototype.forEach.call(elThemes.querySelectorAll('[data-close="themes"]'), function (el) {
    onClick(el, function () { closeThemes(false); });
  });

  Array.prototype.forEach.call(elHelp.querySelectorAll('[data-close="help"]'), function (el) {
    onClick(el, function () { setHelp(false); });
  });

  window.addEventListener('resize', sizeBoard);

  // ── Keyboard ─────────────────────────────────────────────────────────
  function onKeydown(e) {
    const k = e.key;

    // Undo and PDF/Print modifier combos
    if (e.metaKey || e.ctrlKey || e.altKey) {
      if ((k === 'z' || k === 'Z') && hasInk()) {
        e.preventDefault();
        undoStroke();
        return;
      }
      if (k === 'p' || k === 'P') {
        e.preventDefault();
        fetch('/__pdf-doc', {
          method: 'POST',
          headers: { 'Content-Type': 'application/json' },
          body: JSON.stringify({ title: document.title })
        })
          .then(function (r) { if (r.ok) return r.blob(); throw new Error('PDF unavailable'); })
          .then(function (blob) {
            var url = URL.createObjectURL(blob);
            var a = document.createElement('a');
            a.href = url;
            a.download = 'document.pdf';
            document.body.appendChild(a);
            a.click();
            document.body.removeChild(a);
            setTimeout(function () { URL.revokeObjectURL(url); }, 5000);
          })
          .catch(function () {
            window.print();
          });
        return;
      }
      return;
    }

    // Each overlay swallows keys until it is dismissed, outermost first.
    if (themesOpen()) {
      if (k === 'Escape') {
        e.preventDefault();
        closeThemes(true);
        return;
      }
      if (k === 'Enter') {
        e.preventDefault();
        commitTheme();
        return;
      }
      if (k === 'ArrowRight' || k === 'ArrowDown') {
        e.preventDefault();
        const nextIdx = (themeSel + 1) % themeCards.length;
        selectTheme(nextIdx);
        if (themeCards[nextIdx]) themeCards[nextIdx].scrollIntoView({ block: 'nearest' });
        return;
      }
      if (k === 'ArrowLeft' || k === 'ArrowUp') {
        e.preventDefault();
        const prevIdx = (themeSel - 1 + themeCards.length) % themeCards.length;
        selectTheme(prevIdx);
        if (themeCards[prevIdx]) themeCards[prevIdx].scrollIntoView({ block: 'nearest' });
        return;
      }
      if (k === 't' || k === 'T') {
        e.preventDefault();
        closeThemes(false);
        return;
      }
      return;
    }

    if (helpOn) {
      if (k === 'Escape' || k === '?' || k === 'h' || k === 'H') {
        e.preventDefault();
        setHelp(false);
      }
      return;
    }
    if (k === '?' || k === 'h' || k === 'H') {
      e.preventDefault();
      setHelp(true);
      return;
    }

    if (blackOn) {
      // A stray key should not do anything behind a black screen.
      e.preventDefault();
      if (k === 'Escape' || k === 'b' || k === 'B' || k === ' ' || k === 'Enter') setBlack(false);
      return;
    }
    if (k === 'b' || k === 'B') {
      e.preventDefault();
      setBlack(true);
      return;
    }

    // Pen sub-controls only bind while the pen is down, so the letters stay
    // free for everything else the rest of the time.
    if (penOn) {
      if (k >= '1' && k <= String(PEN_COLORS.length)) { e.preventDefault(); setColor(parseInt(k, 10) - 1); return; }
      if (k === 'e' || k === 'E') { e.preventDefault(); setEraser(!eraseOn); return; }
      if (k === '[') { e.preventDefault(); nudgeWidth(-1); return; }
      if (k === ']') { e.preventDefault(); nudgeWidth(1); return; }
      if (k === 'x' || k === 'X') { e.preventDefault(); clearBoard(); return; }
    }

    switch (k) {
      case 'f':
      case 'F':
        e.preventDefault();
        toggleFullscreen();
        break;
      case 't':
      case 'T':
        e.preventDefault();
        openThemes();
        break;
      case 'l':
      case 'L':
        e.preventDefault();
        setLaser(!laserOn);
        break;
      case 'd':
      case 'D':
        e.preventDefault();
        setPen(!penOn);
        break;
      case 'c':
      case 'C':
        e.preventDefault();
        setBlank(!blankOn);
        break;
      case 'Escape':
        e.preventDefault();
        // Peel one layer at a time: themes, canvas, pen, laser.
        if (themesOpen()) closeThemes(true);
        else if (blankOn) setBlank(false);
        else if (penOn) setPen(false);
        else if (laserOn) setLaser(false);
        break;
    }
  }
  document.addEventListener('keydown', onKeydown);

  // A same-origin, unsandboxed iframe still owns its own keyboard/pointer
  // focus, so the parent's listeners never fire for events that start
  // inside the doc unless they are attached there too.
  function attachToFrame() {
    try {
      if (elFrame.contentDocument && elFrame.contentDocument.documentElement) {
        elFrame.contentDocument.documentElement.dataset.theme = document.documentElement.dataset.theme || 'nord';
      }
    } catch (err) {}
    try {
      if (elFrame.contentWindow) {
        elFrame.contentWindow.postMessage({ type: 'theme', theme: document.documentElement.dataset.theme || 'nord' }, '*');
      }
    } catch (err) {}
    try {
      elFrame.contentWindow.addEventListener('keydown', onKeydown);
      elFrame.contentWindow.addEventListener('pointermove', onPointerMove, { passive: true });
    } catch (err) {}
  }
  elFrame.addEventListener('load', attachToFrame);
  attachToFrame();

  // ── Print export ─────────────────────────────────────────────────────
  // Doc-mode PDF/print targets the raw doc URL directly (no chrome, see
  // index.ts /__pdf-doc), so this wrapper page never needs to print itself.

  // ── Init ─────────────────────────────────────────────────────────────
  sizeBoard();
  syncTools();
})();
`

const EditorBodyHTML = `<div id="app">
  <header id="topbar">
    <span id="brand">deckrun<span class="caret"></span></span>
    <input id="docname" value="deck" spellcheck="false" title="Deck name, also the download filename">
    <button class="btn" id="btn-decks" title="Switch between the decks in this browser">decks <span id="deck-count">1</span> <kbd>Cmd O</kbd></button>
    <span class="spacer"></span>
    <span id="topbar-right">
    <button class="btn" id="btn-guide" title="Everything you can put on a slide">guide <kbd>Cmd /</kbd></button>
    <button class="btn" id="btn-palette" title="Insert anything">insert <kbd>Cmd K</kbd></button>
    <button class="btn" id="btn-new" title="Start a new Markdown deck or HTML doc">new</button>
    <span class="menu">
      <button class="btn" id="btn-export" aria-haspopup="true" aria-expanded="false">export <span class="menu__chev">&#9662;</span></button>
      <div class="menu__pop" id="export-menu">
        <button data-export="md">
          <span class="menu__name">Markdown</span>
          <span class="menu__hint">a plain .md file</span>
          <kbd>Cmd S</kbd>
        </button>
        <button data-export="pdf">
          <span class="menu__name">PDF</span>
          <span class="menu__hint">16:9 pages, styling intact</span>
          <kbd>Cmd Shift S</kbd>
        </button>
        <button data-export="html">
          <span class="menu__name">HTML</span>
          <span class="menu__hint">one self-contained page</span>
          <kbd></kbd>
        </button>
      </div>
    </span>
    <span class="menu" id="template-wrap">
      <button class="btn" id="btn-template" aria-haspopup="true" aria-expanded="false" title="Composition template and slide transition">
        <span id="template-label">template</span> <span class="menu__chev">&#9662;</span>
      </button>
      <div class="menu__pop menu__pop--template" id="template-menu">
        <div class="menu-section">
          <div class="menu-section__title">template</div>
          <div id="template-list"></div>
        </div>
        <div class="menu-section">
          <div class="menu-section__title">transition</div>
          <div id="transition-list"></div>
        </div>
      </div>
    </span>
    <button class="btn" id="btn-theme" title="Pick a theme">
      <span class="th-dot" id="theme-dot"></span>
      <span id="theme-label">theme</span>
      <kbd>Cmd Shift L</kbd>
    </button>
    <span class="menu">
      <button class="btn" id="btn-font" aria-haspopup="true" aria-expanded="false" title="Heading and body faces">
        <span id="font-label">font</span> <span class="menu__chev">&#9662;</span>
      </button>
      <div class="menu__pop menu__pop--font" id="font-menu">
        <div id="ff-cols">
          <div class="ff-col">
            <div class="ff-col__title">heading &amp; title</div>
            <div class="ff-list" id="ff-head"></div>
          </div>
          <div class="ff-col">
            <div class="ff-col__title">body</div>
            <div class="ff-list" id="ff-body"></div>
          </div>
        </div>
        <div class="ff-foot">Code keeps the theme's monospace face.</div>
      </div>
    </span>
    <button class="btn btn--primary" id="btn-present" title="Open the real deck in a new tab">present <kbd>Cmd Enter</kbd></button>
    </span>
  </header>

  <main id="panes">
    <section id="pane-edit">
      <div id="edit-wrap">
        <pre id="hl" aria-hidden="true"></pre>
        <pre id="measure" aria-hidden="true"></pre>
        <textarea id="src" spellcheck="false" autocomplete="off" autocapitalize="off" wrap="soft" aria-label="Markdown source"></textarea>
      </div>
      <textarea id="src-html" spellcheck="false" autocomplete="off" autocapitalize="off" wrap="off" aria-label="HTML source"></textarea>
    </section>

    <div id="divider" title="Drag to resize, double-click to reset"></div>

    <section id="pane-prev">
      <div id="prev-head">
        <button class="step" id="btn-prev" title="Previous slide">&#8592;</button>
        <span id="prev-count">slide 0 / 0</span>
        <button class="step" id="btn-next" title="Next slide">&#8594;</button>
        <span id="overflow-badge" title="This slide is clipped. Shorten it or split it with three dashes.">overflow</span>
        <span class="spacer"></span>
        <span id="prev-scale"></span>
        <div class="seg">
          <button id="seg-single" class="is-on">single</button>
          <button id="seg-grid">grid</button>
        </div>
      </div>
      <div id="stage">
        <div id="frame-box" class="is-single">
          <iframe id="frame" src="/__preview" title="Slide preview"></iframe>
        </div>
      </div>
      <div id="notes">
        <div id="notes__label">speaker notes</div>
        <div id="notes__text"></div>
      </div>
      <iframe id="frame-html" title="HTML doc preview"></iframe>
    </section>
  </main>

  <footer id="statusbar">
    <span id="pos">Ln 1, Col 1</span>
    <span id="words">0 words</span>
    <span id="chip-slides">0 slides</span>
    <span id="save-state"></span>
    <span id="follow-chip" class="hidden" title="The editor is mirroring the presented deck. Click to stop following.">following deck &middot; <b id="follow-slide">1</b></span>
    <span id="tipbar">
      <span class="tag">tip</span>
      <span id="tip-text"></span>
      <button id="tip-prev" title="Previous tip">&#8249;</button>
      <button id="tip-next" title="Next tip">&#8250;</button>
    </span>
    <span>local only, nothing is uploaded</span>
  </footer>
</div>

<div id="palette">
  <div class="backdrop" data-close="palette"></div>
  <div id="pal-box">
    <input id="pal-input" placeholder="Insert a style, layout, embed, or action" spellcheck="false" autocomplete="off">
    <div id="pal-list"></div>
    <div id="pal-foot">
      <span>&#8593;&#8595; move</span><span>enter inserts</span><span>esc closes</span>
    </div>
  </div>
</div>

<div id="guide">
  <div class="backdrop" data-close="guide"></div>
  <div id="guide-panel">
    <div id="guide-head">
      <div style="flex:1 1 auto">
        <h2>Everything you can put on a slide</h2>
        <p>Click insert on any card to drop it at your caret.</p>
      </div>
      <button class="nudge-btn nudge-btn--x" data-close="guide" title="Close">&times;</button>
    </div>
    <div id="guide-body"></div>
  </div>
</div>

<div id="library">
  <div class="backdrop" data-close="library"></div>
  <div id="lib-box">
    <div id="lib-head">
      <span>Decks in this browser</span>
      <button id="lib-new">new deck</button>
      <button data-close="library">close</button>
    </div>
    <div id="lib-list"></div>
    <div id="pal-foot">
      <span>&#8593;&#8595; move</span><span>enter opens</span><span>esc closes</span>
    </div>
  </div>
</div>

<div id="themes">
  <div class="backdrop" data-close="themes"></div>
  <div id="th-box">
    <div id="th-head">
      <span>Themes</span>
      <span class="th-head__sub">Arrow keys preview a theme live &nbsp;·&nbsp; enter keeps it &nbsp;·&nbsp; esc puts it back</span>
      <button data-close="themes">close</button>
    </div>
    <div id="th-list"></div>
  </div>
</div>

<div id="screen-start">
  <div class="backdrop"></div>
  <div id="start-box">
    <div id="start-head">
      <span id="brand">deckrun<span class="caret"></span></span>
      <p>What are you making?</p>
    </div>
    <div id="start-cards">
      <div class="start-card" id="start-md-card">
        <span class="start-card__title">Markdown deck</span>
        <span class="start-card__desc">Slides written in Markdown, presented one at a time.</span>
        <span class="start-card__acts">
          <button id="start-md-blank">start blank</button>
          <button id="start-md-upload">upload .md</button>
        </span>
        <div class="start-options">
          <div class="start-options__row">
            <span class="start-options__label">template</span>
            <span class="start-options__choices" id="start-template-list"></span>
          </div>
          <div class="start-options__row">
            <span class="start-options__label">transition</span>
            <span class="start-options__choices" id="start-transition-list"></span>
          </div>
        </div>
      </div>
      <div class="start-card" id="start-html-card">
        <span class="start-card__title">HTML document</span>
        <span class="start-card__desc">One self-contained page, presented with the laser, pen, and canvas &mdash; no slide boundaries.</span>
        <span class="start-card__acts">
          <button id="start-html-upload">upload .html</button>
          <span class="start-card__url">
            <input type="url" id="start-html-url" placeholder="https://a-public-page.html" inputmode="url" autocomplete="off" spellcheck="false">
            <button id="start-html-url-go">load URL</button>
          </span>
        </span>
      </div>
      <button class="start-card" id="start-lib">
        <span class="start-card__title">Open from library</span>
        <span class="start-card__desc" id="start-lib-desc">Decks and docs already in this browser.</span>
      </button>
    </div>
  </div>
</div>

<div id="toasts"></div>
<input type="file" id="file-any" accept=".md,.markdown,text/markdown,text/plain,.html,.htm,text/html" hidden>

`
