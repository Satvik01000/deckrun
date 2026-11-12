(function () {
  'use strict';

  var D = JSON.parse(document.getElementById('bootstrap').textContent);
  var MAC = /Mac|iPhone|iPad/.test(navigator.platform || navigator.userAgent);
  var CMD = MAC ? 'Cmd' : 'Ctrl';

  // The document deckrun was launched with (a file on disk or a fetched
  // URL). When set, the editor is backed by it instead of the browser
  // library: content comes from /__file, edits save back through it, and
  // external changes on disk arrive over /__events.
  var FILE = D.file || null;

  var K = {
    index:   'deckrun.decks.v1',
    deck:    'deckrun.deck.',
    current: 'deckrun.current.v1',
    theme:   'deckrun.theme.v1',
    head:    'deckrun.font.head.v1',
    body:    'deckrun.font.body.v1',
    template:'deckrun.template.v1',
    transition:'deckrun.transition.v1',
    split:   'deckrun.split.v1',
    mode:    'deckrun.mode.v1',
    nudge:   'deckrun.nudges.v1',
    // Superseded by the deck library, read once to migrate.
    oldDoc:  'presentmd.doc.v1',
    oldName: 'presentmd.name.v1'
  };

  // present-md was renamed to deckrun — the whole storage namespace moves
  // with it, so this copies it over once rather than orphaning every deck
  // already saved under the old prefix.
  var OLD_NS = {
    index: 'presentmd.decks.v1', deck: 'presentmd.deck.', current: 'presentmd.current.v1',
    theme: 'presentmd.theme.v1', head: 'presentmd.font.head.v1',
    body: 'presentmd.font.body.v1', split: 'presentmd.split.v1', mode: 'presentmd.mode.v1',
    nudge: 'presentmd.nudges.v1'
  };
  (function migrateNamespace() {
    if (lsGet(K.index, null) !== null) return;
    var oldIndexRaw = lsGet(OLD_NS.index, null);
    if (oldIndexRaw === null) return;
    lsSet(K.index, oldIndexRaw);
    lsDel(OLD_NS.index);
    try {
      var oldIndex = JSON.parse(oldIndexRaw);
      if (Object.prototype.toString.call(oldIndex) === '[object Array]') {
        oldIndex.forEach(function (entry) {
          if (!entry || !entry.id) return;
          var content = lsGet(OLD_NS.deck + entry.id, null);
          if (content !== null) {
            lsSet(K.deck + entry.id, content);
            lsDel(OLD_NS.deck + entry.id);
          }
        });
      }
    } catch (e) {}
    ['current', 'theme', 'head', 'body', 'split', 'mode', 'nudge'].forEach(function (slot) {
      var v = lsGet(OLD_NS[slot], null);
      if (v !== null) { lsSet(K[slot], v); lsDel(OLD_NS[slot]); }
    });
  })();

  function lsGet(k, fallback) {
    try { var v = localStorage.getItem(k); return v === null ? fallback : v; }
    catch (e) { return fallback; }
  }

  function lsSet(k, v) {
    try { localStorage.setItem(k, v); return true; }
    catch (e) { return false; }
  }

  function lsDel(k) {
    try { localStorage.removeItem(k); } catch (e) {}
  }

  // ── Deck library ───────────────────────────────────────────────────────
  // The index holds metadata only, so listing decks never reads their text.
  // Each deck's Markdown lives under its own key.

  function loadIndex() {
    try {
      var list = JSON.parse(lsGet(K.index, '[]'));
      return Object.prototype.toString.call(list) === '[object Array]' ? list : [];
    } catch (e) { return []; }
  }

  function saveIndex(list) {
    return lsSet(K.index, JSON.stringify(list));
  }

  function newDeckId() {
    return 'd' + Date.now().toString(36) + Math.floor(Math.random() * 1e6).toString(36);
  }

  function readDeck(id) { return lsGet(K.deck + id, ''); }

  function findDeck(id) {
    var list = loadIndex();
    for (var i = 0; i < list.length; i++) if (list[i].id === id) return list[i];
    return null;
  }

  /** Returns the new deck's id, or null if this browser refused to store it. */
  function createDeck(name, content, kind, template, transition) {
    var id = newDeckId();
    if (!lsSet(K.deck + id, content)) return null;
    var list = loadIndex();
    list.unshift({
      id: id,
      name: name || 'untitled',
      kind: kind || 'markdown',
      slides: 0,
      chars: content.length,
      template: template || (typeof state !== 'undefined' && state.template ? state.template : D.template),
      transition: transition || (typeof state !== 'undefined' && state.transition ? state.transition : D.transition),
      at: Date.now()
    });
    if (!saveIndex(list)) { lsDel(K.deck + id); return null; }
    return id;
  }

  function uniqueName(base) {
    var list = loadIndex();
    var taken = {};
    list.forEach(function (d) { taken[d.name] = true; });
    if (!taken[base]) return base;
    for (var n = 2; n < 500; n++) if (!taken[base + ' ' + n]) return base + ' ' + n;
    return base;
  }

  function timeAgo(ts) {
    var mins = Math.floor((Date.now() - ts) / 60000);
    if (mins < 1) return 'just now';
    if (mins < 60) return mins + 'm ago';
    var hours = Math.floor(mins / 60);
    if (hours < 24) return hours + 'h ago';
    var days = Math.floor(hours / 24);
    return days + 'd ago';
  }

  function updateDocTitle() {
    if ($('screen-start') && $('screen-start').classList.contains('is-on')) {
      document.title = 'deckrun · start';
      return;
    }
    if ($('library') && $('library').classList.contains('is-on')) {
      document.title = 'deckrun · library';
      return;
    }
    var name = ($('docname').value || '').trim();
    document.title = name ? name + ' · deckrun' : 'deckrun · editor';
  }

  /** One deck's metadata is refreshed from the editor on every save. */
  function touchCurrent() {
    var list = loadIndex();
    for (var i = 0; i < list.length; i++) {
      if (list[i].id === state.deckId) {
        list[i].name = $('docname').value.trim() || 'untitled';
        list[i].kind = state.kind;
        list[i].template = state.template;
        list[i].transition = state.transition;
        if (state.kind === 'html') {
          list[i].slides = 0;
          list[i].chars = srcHtml.value.length;
        } else {
          list[i].slides = state.slides.length;
          list[i].chars = src.value.length;
        }
        list[i].at = Date.now();
        break;
      }
    }
    updateDocTitle();
    return saveIndex(list);
  }

  function updateDeckCount() {
    $('deck-count').textContent = String(loadIndex().length);
  }

  /** The open document's own text, whichever kind it is. */
  function curValue() { return state.kind === 'html' ? srcHtml.value : src.value; }

  /** Relabels the export menu, since "Markdown" makes no sense for a doc. */
  function paintExportMenu() {
    var mdName = document.querySelector('#export-menu [data-export="md"] .menu__name');
    var mdHint = document.querySelector('#export-menu [data-export="md"] .menu__hint');
    var htmlName = document.querySelector('#export-menu [data-export="html"] .menu__name');
    var htmlHint = document.querySelector('#export-menu [data-export="html"] .menu__hint');
    if (state.kind === 'html') {
      mdName.textContent = 'Source';
      mdHint.textContent = 'the raw .html file';
      htmlName.textContent = 'Presenter Page';
      htmlHint.textContent = 'standalone, with the tool belt built in';
    } else {
      mdName.textContent = 'Markdown';
      mdHint.textContent = 'a plain .md file';
      htmlName.textContent = 'HTML';
      htmlHint.textContent = 'one self-contained page';
    }
  }

  /** Switches which document kind the chrome is dressed for. */
  function setDocKind(kind) {
    state.kind = kind === 'html' ? 'html' : 'markdown';
    document.documentElement.dataset.docKind = state.kind;
    paintExportMenu();
  }

  // ── Themes ─────────────────────────────────────────────────────────────
  var THEMES = D.themes;
  var THEME_BY_ID = {};
  THEMES.forEach(function (t) { THEME_BY_ID[t.id] = t; });

  /** A theme id we can actually apply, whatever localStorage happens to hold. */
  function pickTheme(id) {
    return THEME_BY_ID[id] ? id : (THEME_BY_ID[D.theme] ? D.theme : THEMES[0].id);
  }

  var TEMPLATES = D.templates;
  var TEMPLATE_BY_ID = {};
  TEMPLATES.forEach(function (item) { TEMPLATE_BY_ID[item.id] = item; });

  function pickTemplate(id) {
    return TEMPLATE_BY_ID[id] ? id : (TEMPLATE_BY_ID[D.template] ? D.template : 'classic');
  }

  var TRANSITIONS = D.transitions;
  var TRANSITION_BY_ID = {};
  TRANSITIONS.forEach(function (item) { TRANSITION_BY_ID[item.id] = item; });

  function pickTransition(id) {
    return TRANSITION_BY_ID[id] ? id : (TRANSITION_BY_ID[D.transition] ? D.transition : 'slide');
  }

  var FACES = D.faces;
  var FACE_BY_ID = {};
  FACES.forEach(function (f) { FACE_BY_ID[f.id] = f; });

  /** null means "whatever the theme says", which is not a face id. */
  function pickFont(id) {
    return FACE_BY_ID[id] ? id : null;
  }

  // ── Elements ───────────────────────────────────────────────────────────
  var $ = function (id) { return document.getElementById(id); };
  var src = $('src'), hl = $('hl'), measure = $('measure');
  var frame = $('frame'), frameBox = $('frame-box'), stage = $('stage');
  var paneEdit = $('pane-edit'), panePrev = $('pane-prev'), divider = $('divider');
  var elChipSlides = $('chip-slides'), elSave = $('save-state'), elPos = $('pos'), elWords = $('words');
  var elCount = $('prev-count'), elScale = $('prev-scale');
  var elNotes = $('notes'), elNotesText = $('notes__text');
  var palette = $('palette'), palInput = $('pal-input'), palList = $('pal-list');
  var guide = $('guide'), guideBody = $('guide-body');
  var srcHtml = $('src-html'), frameHtml = $('frame-html');

  // ── State ──────────────────────────────────────────────────────────────
  var state = {
    deckId: null,
    kind: 'markdown',
    slides: [],
    notes: [],
    index: 0,
    mode: lsGet(K.mode, 'single') === 'grid' ? 'grid' : 'single',
    theme: pickTheme(lsGet(K.theme, D.theme)),
    head: pickFont(lsGet(K.head, D.fonts.head)),
    body: pickFont(lsGet(K.body, D.fonts.body)),
    template: pickTemplate(lsGet(K.template, D.template)),
    transition: pickTransition(lsGet(K.transition, D.transition)),
    frameReady: false,
    overflow: {},
    dismissed: {},
    activeNudge: null,
    nudgeShownAt: 0,
    seq: 0,
    inflight: null,
    tip: Math.floor(Math.random() * D.tips.length)
  };

  try {
    var saved = JSON.parse(lsGet(K.nudge, '[]'));
    if (saved && saved.length) saved.forEach(function (id) { state.dismissed[id] = true; });
  } catch (e) {}

  // ── Following the presented deck ───────────────────────────────────────
  // Presenting hands the deck tab this editor's session id (ps=... in the
  // URL), and the editor listens on that same channel: the preview and the
  // notes panel mirror the deck's current slide, so the editor window is the
  // presenter's screen and the deck tab is the projector. No second window
  // is involved.
  var FOLLOW_KEY = 'deckrun.presenter.session';
  var followSid = null;
  try { followSid = sessionStorage.getItem(FOLLOW_KEY); } catch (e) {}
  if (!followSid) {
    followSid = (window.crypto && crypto.randomUUID)
      ? crypto.randomUUID()
      : 'e' + Date.now().toString(36) + Math.random().toString(36).slice(2);
    try { sessionStorage.setItem(FOLLOW_KEY, followSid); } catch (e) {}
  }
  var followChan = (typeof BroadcastChannel !== 'undefined')
    ? new BroadcastChannel('deckrun:' + followSid)
    : null;

  var following = false;   // the editor is mirroring the deck right now
  var expectingDeck = false; // a present just went out: follow the deck that answers
  var deckNotes = [];      // the deck's own notes, authoritative while following
  var followChipEl = $('follow-chip');
  var followSlideEl = $('follow-slide');

  /** The editor drives the deck as well as the preview: wherever the
      presenter lands in the editor, the deck goes there too. The deck echoes
      its state back over the same channel, which is what keeps both in step. */
  function announceIndex(i) {
    if (followChan) followChan.postMessage({ id: followSid, type: 'goto', index: i });
  }

  function followDeck(i) {
    var n = state.slides.length;
    if (!n) return;
    i = Math.max(0, Math.min(n - 1, i));
    var changed = i !== state.index;
    state.index = i;
    state.notes = deckNotes;
    followSlideEl.textContent = String(i + 1);
    if (changed) pushFrame();
    renderCounts();
  }

  if (followChan) {
    followChan.onmessage = function (e) {
      var m = e.data || {};
      if (m.id !== followSid) return;
      if (m.type === 'init') {
        deckNotes = m.notes || [];
        if (expectingDeck) {
          expectingDeck = false;
          following = true;
        }
        followChipEl.classList.toggle('hidden', !following);
      } else if (m.type === 'state') {
        if (following) followDeck(m.index);
      }
    };
    setInterval(function () {
      followChan.postMessage({ id: followSid, type: 'ready' });
    }, 2000);
  }

  followChipEl.addEventListener('click', function () {
    following = false;
    followChipEl.classList.add('hidden');
  });

  // ── Session highlights ─────────────────────────────────────────────────
  // Selecting text in either preview paints a highlight and, optionally,
  // hangs a comment off it. The store is the browser session's, shared with
  // the tab present opens, so the marks are already there on the projector.
  // Nothing is written to disk, and closing the browser clears the lot.
  function highlightKey() {
    if (FILE) return 'file:' + FILE.name;
    return state.deckId ? 'deck:' + state.deckId : 'deck:unsaved';
  }

  var hlSlides = null, hlDoc = null;
  if (window.deckrunHighlights) {
    var onHighlightWarn = function (message) { toast(message, 'warn'); };
    hlSlides = window.deckrunHighlights.mount({
      frame: frame,
      docKey: highlightKey(),
      scopes: 'slides',
      onWarn: onHighlightWarn
    });
    hlDoc = window.deckrunHighlights.mount({
      frame: frameHtml,
      docKey: highlightKey(),
      scopes: 'doc',
      onWarn: onHighlightWarn
    });
  }

  /** Points both previews at whichever document is now open. */
  function syncHighlightDoc() {
    var key = highlightKey();
    if (hlSlides) hlSlides.setDocKey(key);
    if (hlDoc) hlDoc.setDocKey(key);
  }

  // ── Toasts ─────────────────────────────────────────────────────────────
  function toast(message, kind, actionLabel, action) {
    var el = document.createElement('div');
    el.className = 'toast' + (kind ? ' toast--' + kind : '');
    var text = document.createElement('span');
    text.style.flex = '1 1 auto';
    text.textContent = message;
    el.appendChild(text);
    if (actionLabel) {
      var btn = document.createElement('button');
      btn.textContent = actionLabel;
      btn.addEventListener('click', function () { action(); el.remove(); });
      el.appendChild(btn);
    }
    $('toasts').appendChild(el);
    setTimeout(function () {
      el.style.transition = 'opacity .3s ease';
      el.style.opacity = '0';
      setTimeout(function () { el.remove(); }, 320);
    }, actionLabel ? 8000 : 4200);
  }

  // ── Markdown highlighting for the editor surface ───────────────────────
  function esc(s) {
    return s.replace(/&/g, '&amp;').replace(/</g, '&lt;').replace(/>/g, '&gt;');
  }

  var HOLD = '\u0000';

  function inlineTokens(line) {
    var held = [];
    function hold(html) { held.push(html); return HOLD + (held.length - 1) + HOLD; }

    var out = esc(line);

    // Images first: the title attribute carries the layout directive.
    out = out.replace(/!\[([^\]\n]*)\]\(([^\s)]*)(\s+"[^"\n]*")?\)/g, function (m, alt, url, title) {
      var html = '<span class="t-img">![' + alt + ']</span><span class="t-url">(' + url;
      if (title) html += '</span><span class="t-dir">' + title + '</span><span class="t-url">';
      return hold(html + ')</span>');
    });

    // Links.
    out = out.replace(/\[([^\]\n]*)\]\(([^\s)]*)(\s+"[^"\n]*")?\)/g, function (m, label, url, title) {
      return hold('<span class="t-link">[' + label + ']</span><span class="t-url">(' + url +
        (title ? '</span><span class="t-dir">' + title + '</span><span class="t-url">' : '') + ')</span>');
    });

    // Inline code.
    out = out.replace(/`[^`\n]+`/g, function (m) { return hold('<span class="t-code">' + m + '</span>'); });

    // Raw HTML tags.
    out = out.replace(/&lt;\/?[a-zA-Z][^&]{0,200}?&gt;/g, function (m) {
      return hold('<span class="t-html">' + m + '</span>');
    });

    out = out.replace(/\*\*[^*\n]+\*\*/g, function (m) { return '<span class="t-strong">' + m + '</span>'; });
    out = out.replace(/(^|[^*\w])(\*[^*\n]+\*)/g, function (m, pre, body) {
      return pre + '<span class="t-em">' + body + '</span>';
    });
    out = out.replace(/(^|[^_\w])(_[^_\n]+_)/g, function (m, pre, body) {
      return pre + '<span class="t-em">' + body + '</span>';
    });

    return out.replace(new RegExp(HOLD + '(\\d+)' + HOLD, 'g'), function (m, i) { return held[+i]; });
  }

  function highlightMarkdown(text) {
    var lines = text.split('\n');
    var out = [];
    var inFence = false;

    for (var i = 0; i < lines.length; i++) {
      var line = lines[i];

      // A pasted data URI is one enormous line. Colour it in a single pass so
      // typing stays responsive.
      if (line.length > 4000) {
        out.push('<span class="t-url">' + esc(line) + '</span>');
        continue;
      }

      if (/^\s*```/.test(line)) {
        out.push('<span class="t-fence">' + esc(line) + '</span>');
        inFence = !inFence;
        continue;
      }
      if (inFence) { out.push('<span class="t-codeline">' + esc(line) + '</span>'); continue; }
      if (/^[ \t]*---[ \t]*$/.test(line)) {
        out.push('<span class="t-sep">' + esc(line) + '</span>');
        continue;
      }
      if (/^\s*<!--/.test(line)) {
        out.push('<span class="t-note">' + esc(line) + '</span>');
        continue;
      }

      var head = line.match(/^(#{1,6})\s/);
      if (head) {
        var lvl = Math.min(head[1].length, 4);
        out.push('<span class="t-h' + lvl + '">' + inlineTokens(line) + '</span>');
        continue;
      }
      if (/^\s*>/.test(line)) {
        out.push('<span class="t-quote">' + inlineTokens(line) + '</span>');
        continue;
      }
      if (/^\s*\|.*\|\s*$/.test(line)) {
        out.push('<span class="t-table">' + inlineTokens(line) + '</span>');
        continue;
      }
      var marker = line.match(/^(\s*)([-*+]|\d+[.)])(\s+)/);
      if (marker) {
        out.push(marker[1] + '<span class="t-marker">' + esc(marker[2]) + '</span>' + marker[3] +
          inlineTokens(line.slice(marker[0].length)));
        continue;
      }
      out.push(inlineTokens(line));
    }

    return out.join('\n') + '\n';
  }

  function paint() { hl.innerHTML = highlightMarkdown(src.value); }

  function syncScroll() { hl.scrollTop = src.scrollTop; hl.scrollLeft = src.scrollLeft; }

  // ── Slide map, mirroring the server-side split ─────────────────────────
  var SEP = /\r?\n[ \t]*---[ \t]*\r?\n/g;

  function slideMap() {
    var text = src.value, spans = [], last = 0, m;
    SEP.lastIndex = 0;
    while ((m = SEP.exec(text)) !== null) {
      var prefix = text.slice(0, m.index);
      var backticks = (prefix.match(/```/g) || []).length;
      var tildes = (prefix.match(/~~~/g) || []).length;
      if (backticks % 2 === 1 || tildes % 2 === 1) {
        continue;
      }
      spans.push([last, m.index]);
      last = m.index + m[0].length;
      SEP.lastIndex = last;
    }
    spans.push([last, text.length]);
    return spans.filter(function (s) { return text.slice(s[0], s[1]).trim().length > 0; });
  }

  function caretSlide() {
    var pos = src.selectionStart, spans = slideMap();
    for (var i = 0; i < spans.length; i++) {
      if (pos <= spans[i][1]) return i;
    }
    return Math.max(0, spans.length - 1);
  }

  function offsetTop(pos) {
    measure.textContent = src.value.slice(0, pos);
    var marker = document.createElement('span');
    marker.textContent = '\u200b';
    measure.appendChild(marker);
    return marker.offsetTop;
  }

  function scrollCaretIntoView() {
    var top = offsetTop(src.selectionStart);
    var view = src.clientHeight;
    if (top < src.scrollTop + 40 || top > src.scrollTop + view - 80) {
      src.scrollTop = Math.max(0, top - view / 3);
      syncScroll();
    }
  }

  // ── Preview plumbing ───────────────────────────────────────────────────
  /** Dropped rather than queued until the frame is up: see the ready handler. */
  function post(msg) {
    if (!state.frameReady) return;
    frame.contentWindow.postMessage(msg, '*');
  }

  function pushFrame() {
    post({ type: 'render', slides: state.slides, index: state.index, mode: state.mode });
  }

  function layout() {
    var padding = 32;
    var w = Math.max(120, stage.clientWidth - padding);
    var h = Math.max(90, stage.clientHeight - padding);
    var scale;
    if (state.mode === 'single') {
      scale = Math.min(w / D.width, h / D.height);
      frame.style.height = D.height + 'px';
      frameBox.classList.add('is-single');
      frameBox.style.width = Math.round(D.width * scale) + 'px';
      frameBox.style.height = Math.round(D.height * scale) + 'px';
    } else {
      scale = w / D.width;
      var virtualH = Math.round(h / scale);
      frame.style.height = virtualH + 'px';
      frameBox.classList.remove('is-single');
      frameBox.style.width = Math.round(w) + 'px';
      frameBox.style.height = Math.round(h) + 'px';
    }
    frame.style.transform = 'scale(' + scale + ')';
    elScale.textContent = Math.round(scale * 100) + '%';
  }

  function renderCounts() {
    var n = state.slides.length;
    elChipSlides.textContent = n + (n === 1 ? ' slide' : ' slides');
    elCount.textContent = 'slide ' + (n ? state.index + 1 : 0) + ' / ' + n;
    $('btn-prev').disabled = state.index <= 0;
    $('btn-next').disabled = state.index >= n - 1;
    $('overflow-badge').classList.toggle(
      'is-on',
      state.mode === 'single' && !!state.overflow[state.index]
    );

    var note = state.notes[state.index];
    if (note) { elNotesText.textContent = note; elNotes.classList.add('is-on'); }
    else { elNotes.classList.remove('is-on'); }

    var words = src.value.trim() ? src.value.trim().split(/\s+/).length : 0;
    elWords.textContent = words + (words === 1 ? ' word' : ' words');
  }

  function setIndex(i, moveCaret) {
    var n = state.slides.length;
    if (!n) return;
    i = Math.max(0, Math.min(n - 1, i));
    if (i === state.index && !moveCaret) return;
    state.index = i;
    if (moveCaret) {
      var spans = slideMap();
      if (spans[i]) {
        var target = spans[i][0];
        while (/\s/.test(src.value.charAt(target)) && target < spans[i][1]) target++;
        src.focus();
        src.setSelectionRange(target, target);
        scrollCaretIntoView();
        updateCaretUi();
      }
    }
    post({ type: 'index', index: state.index });
    announceIndex(state.index);
    renderCounts();
  }

  function updateCaretUi() {
    var before = src.value.slice(0, src.selectionStart);
    var lines = before.split('\n');
    elPos.textContent = 'Ln ' + lines.length + ', Col ' + (lines[lines.length - 1].length + 1);
  }

  // ── Parse round-trip: the server owns the Markdown, so what you see here
  //    is byte-for-byte what deckrun file.md renders. ───────────────
  function refresh() {
    var mine = ++state.seq;
    if (state.inflight) state.inflight.abort();
    state.inflight = new AbortController();

    fetch('/__parse', {
      method: 'POST',
      headers: { 'Content-Type': 'text/plain; charset=utf-8' },
      body: src.value,
      signal: state.inflight.signal
    })
      .then(function (r) { return r.json(); })
      .then(function (data) {
        if (mine !== state.seq) return;
        state.slides = data.slides || [];
        state.notes = data.notes || [];
        if (state.index >= state.slides.length) state.index = Math.max(0, state.slides.length - 1);
        pushFrame();
        renderCounts();
        evalNudges();
      })
      .catch(function (err) {
        if (err && err.name === 'AbortError') return;
        toast('Preview failed: ' + err.message + '. Is the server still running?', 'err');
      });
  }

  // ── Autosave ───────────────────────────────────────────────────────────
  var saveTimer = null;

  function saveNow() {
    if (FILE) {
      if (!FILE.writable) return;
      fetch('/__file', {
        method: 'POST',
        headers: { 'Content-Type': 'text/plain; charset=utf-8' },
        body: curValue()
      })
        .then(function (r) {
          if (!r.ok) throw new Error('HTTP ' + r.status);
          elSave.className = '';
          elSave.textContent = '';
        })
        .catch(function () {
          elSave.className = 'err';
          elSave.textContent = 'not saved: write to ' + FILE.name + ' failed';
        });
      return;
    }
    var ok = state.deckId && lsSet(K.deck + state.deckId, curValue()) && touchCurrent();
    if (ok) {
      elSave.className = '';
      elSave.textContent = '';
    } else if (state.deckId) {
      elSave.className = 'err';
      elSave.textContent = 'not saved: storage full';
      toast('This browser is out of room. Download this deck, or delete one you no longer need.', 'err', 'download', download);
    }
  }

  function scheduleSave() {
    if (saveTimer) clearTimeout(saveTimer);
    // Writing a megabyte of content costs real time. Back off on big docs.
    saveTimer = setTimeout(saveNow, curValue().length > 1024 * 1024 ? 1500 : 500);
  }

  // ── Text insertion, undo-safe ──────────────────────────────────────────
  function typeText(text) {
    src.focus();
    var inserted = false;
    try { inserted = document.execCommand('insertText', false, text); } catch (e) {}
    if (!inserted) {
      var s = src.selectionStart, e2 = src.selectionEnd;
      src.setRangeText(text, s, e2, 'end');
    }
  }

  var MARK = '\u0001';

  function insertSnippet(sn) {
    if (!sn.insert) return;
    var s = src.selectionStart, e = src.selectionEnd;
    var sel = src.value.slice(s, e);

    var text = sn.insert.split('{caret}').join(MARK);
    if (text.indexOf(MARK) === -1 && !sel) text = text.split('{sel}').join(MARK);
    text = text.split('{sel}').join(sel);

    var prefix = '', suffix = '';
    if (sn.block) {
      var before = src.value.slice(0, s);
      if (before && !/\n[ \t]*$/.test(before)) prefix = '\n\n';
      else if (before && !/\n[ \t]*\n[ \t]*$/.test(before)) prefix = '\n';
      var after = src.value.slice(e);
      if (after && !/^[ \t]*\n/.test(after)) suffix = '\n';
    }

    var caretAt = text.indexOf(MARK);
    text = text.split(MARK).join('');
    var full = prefix + text + suffix;

    typeText(full);
    var pos = s + prefix.length + (caretAt === -1 ? text.length : caretAt);
    src.setSelectionRange(pos, pos);
    onInput();
    scrollCaretIntoView();
  }

  // ── Nudges: rules evaluated against the live document ──────────────────
  var NUDGE_RULES = [
    {
      id: 'overflow',
      sticky: true,
      test: function (ctx) {
        var i = state.index;
        return state.overflow[i] ? 'Slide ' + (i + 1) + ' overflows and is being clipped. Trim it, or split it with <code>---</code>.' : null;
      },
      snippet: 'slide-break'
    },
    {
      id: 'no-split',
      test: function (ctx) {
        return ctx.count === 1 && ctx.text.length > 420
          ? 'This is one long slide. Three dashes on their own line start the next one.'
          : null;
      },
      snippet: 'slide-break'
    },
    {
      id: 'untagged-fence',
      test: function (ctx) {
        return ctx.untaggedFence
          ? 'That code block has no language tag. Write <code>&#96;&#96;&#96;go</code> and it gets highlighted.'
          : null;
      },
      snippet: 'code-go'
    },
    {
      id: 'img-plain',
      test: function (ctx) {
        return ctx.plainImage
          ? 'Your image is inline. Add <code>"right"</code> after the URL and the slide splits: text left, image right.'
          : null;
      },
      snippet: 'img-right'
    },
    {
      id: 'no-notes',
      test: function (ctx) {
        return ctx.count >= 3 && !ctx.hasNotes
          ? 'Speaker notes live in <code>&#60;!-- notes: ... --&#62;</code>. They show under the preview and never reach the deck.'
          : null;
      },
      snippet: 'slide-notes'
    },
    {
      id: 'no-image',
      test: function (ctx) {
        return ctx.count >= 4 && !ctx.hasImage
          ? 'All text so far. An image path resolves against the folder you launched in: <code>![alt](diagram.png "right")</code>.'
          : null;
      },
      snippet: 'img-right'
    },
    {
      id: 'no-table',
      test: function (ctx) {
        return ctx.count >= 5 && !ctx.hasTable
          ? 'Numbers land harder in a table than in bullets. Zebra rows and an accented header come for free.'
          : null;
      },
      snippet: 'table'
    },
    {
      id: 'no-embed',
      test: function (ctx) {
        return ctx.count >= 6 && !ctx.hasEmbed
          ? 'Raw HTML works, so a YouTube or dashboard <code>&#60;iframe&#62;</code> can live on a slide, sized to 16:9.'
          : null;
      },
      snippet: 'embed-youtube'
    },
    {
      id: 'dense',
      test: function (ctx) {
        return ctx.denseSlide !== -1
          ? 'Slide ' + (ctx.denseSlide + 1) + ' is dense. Around seven bullets is the ceiling from the back row.'
          : null;
      },
      snippet: 'slide-break'
    },
    {
      id: 'huge',
      sticky: true,
      test: function (ctx) {
        return ctx.text.length > 3.5 * 1024 * 1024
          ? 'This deck is over 3.5 MB, near what a browser will autosave. Download a copy.'
          : null;
      },
      action: 'download'
    }
  ];

  function docContext() {
    var text = src.value;
    var lines = text.split('\n');
    var untagged = false, inFence = false;
    for (var i = 0; i < lines.length; i++) {
      var m = lines[i].match(/^\s*```(.*)$/);
      if (!m) continue;
      if (!inFence && !m[1].trim()) untagged = true;
      inFence = !inFence;
    }

    var dense = -1;
    var spans = slideMap();
    for (var j = 0; j < spans.length; j++) {
      var body = text.slice(spans[j][0], spans[j][1]);
      if (body.indexOf('```') !== -1) continue;
      var bullets = body.split('\n').filter(function (l) { return /^\s*([-*+]|\d+[.)])\s/.test(l); });
      if (bullets.length > 9) { dense = j; break; }
    }

    return {
      text: text,
      count: state.slides.length,
      untaggedFence: untagged,
      plainImage: /!\[[^\]]*\]\([^\s)]+\)(?!\s*\{)/.test(text) && !/!\[[^\]]*\]\([^\s)]+\s+"[^"]*"\)/.test(text),
      hasNotes: /<!--\s*notes?:/i.test(text),
      hasImage: /!\[[^\]]*\]\(/.test(text),
      hasTable: /^\s*\|.*\|\s*$/m.test(text),
      hasEmbed: /<(iframe|video)\b/i.test(text),
      denseSlide: dense
    };
  }

  var bySnippetId = {};
  D.snippets.forEach(function (s) { bySnippetId[s.id] = s; });

  function evalNudges() {}
  function hideNudge() {}

  // ── Tips carousel ──────────────────────────────────────────────────────
  var tipTimer = null;

  function showTip(step) {
    state.tip = (state.tip + step + D.tips.length) % D.tips.length;
    $('tip-text').textContent = D.tips[state.tip];
  }

  function startTips() {
    showTip(0);
    if (tipTimer) clearInterval(tipTimer);
    tipTimer = setInterval(function () { showTip(1); }, 13000);
  }

  $('tip-prev').addEventListener('click', function () { showTip(-1); startTips(); });
  $('tip-next').addEventListener('click', function () { showTip(1); startTips(); });
  $('tipbar').addEventListener('mouseenter', function () { if (tipTimer) clearInterval(tipTimer); });
  $('tipbar').addEventListener('mouseleave', startTips);

  // ── Command palette ────────────────────────────────────────────────────
  var palRows = [], palSel = 0;

  function palFilter() {
    var q = palInput.value.trim().toLowerCase();
    var terms = q ? q.split(/\s+/) : [];
    return D.snippets.filter(function (s) {
      if (!terms.length) return true;
      var hay = (s.label + ' ' + s.group + ' ' + s.hint + ' ' + s.syntax).toLowerCase();
      return terms.every(function (t) { return hay.indexOf(t) !== -1; });
    });
  }

  function renderPalette() {
    var items = palFilter();
    palRows = items;
    if (palSel >= items.length) palSel = Math.max(0, items.length - 1);
    palList.innerHTML = '';

    if (!items.length) {
      var none = document.createElement('div');
      none.className = 'pal-group';
      none.textContent = 'nothing matches';
      palList.appendChild(none);
      return;
    }

    var group = null;
    items.forEach(function (item, i) {
      if (item.group !== group) {
        group = item.group;
        var head = document.createElement('div');
        head.className = 'pal-group';
        head.textContent = group;
        palList.appendChild(head);
      }
      var row = document.createElement('div');
      row.className = 'pal-row' + (i === palSel ? ' is-sel' : '');
      row.dataset.i = String(i);

      var main = document.createElement('div');
      main.className = 'pal-row__main';
      var label = document.createElement('div');
      label.className = 'pal-row__label';
      label.textContent = item.label;
      var hint = document.createElement('div');
      hint.className = 'pal-row__hint';
      hint.textContent = item.hint;
      main.appendChild(label);
      main.appendChild(hint);
      row.appendChild(main);

      if (item.syntax) {
        var syn = document.createElement('code');
        syn.className = 'pal-row__syntax';
        syn.textContent = item.syntax.split('\\n').join(' ');
        row.appendChild(syn);
      }
      if (item.keys) {
        var keys = document.createElement('span');
        keys.className = 'pal-row__keys';
        keys.textContent = item.keys.replace('Cmd', CMD);
        row.appendChild(keys);
      }

      row.addEventListener('mousemove', function () {
        if (palSel === i) return;
        palSel = i;
        var sel = palList.querySelector('.pal-row.is-sel');
        if (sel) sel.classList.remove('is-sel');
        row.classList.add('is-sel');
      });
      row.addEventListener('click', function () { palRun(i); });
      palList.appendChild(row);
    });
  }

  function palRun(i) {
    var item = palRows[i];
    if (!item) return;
    closeOverlays();
    if (item.action) runAction(item.action);
    else insertSnippet(item);
  }

  function palMove(step) {
    if (!palRows.length) return;
    palSel = (palSel + step + palRows.length) % palRows.length;
    renderPalette();
    var sel = palList.querySelector('.pal-row.is-sel');
    if (sel) sel.scrollIntoView({ block: 'nearest' });
  }

  function openPalette() {
    palette.classList.add('is-on');
    palInput.value = '';
    palSel = 0;
    renderPalette();
    palInput.focus();
  }

  palInput.addEventListener('input', function () { palSel = 0; renderPalette(); });
  palInput.addEventListener('keydown', function (e) {
    if (e.key === 'ArrowDown') { e.preventDefault(); palMove(1); }
    else if (e.key === 'ArrowUp') { e.preventDefault(); palMove(-1); }
    else if (e.key === 'Enter') { e.preventDefault(); palRun(palSel); }
    else if (e.key === 'Escape') { e.preventDefault(); closeOverlays(); }
  });

  // ── Guide drawer ───────────────────────────────────────────────────────
  function buildGuide() {
    guideBody.innerHTML = '';
    D.groups.forEach(function (group) {
      var items = D.snippets.filter(function (s) { return s.group === group; });
      if (!items.length) return;

      var sec = document.createElement('section');
      sec.className = 'gsec';
      var head = document.createElement('h3');
      head.textContent = group;
      sec.appendChild(head);

      items.forEach(function (item) {
        var card = document.createElement('div');
        card.className = 'gcard';

        var top = document.createElement('div');
        top.className = 'gcard__top';
        var label = document.createElement('span');
        label.className = 'gcard__label';
        label.textContent = item.label;
        top.appendChild(label);
        if (item.keys) {
          var keys = document.createElement('span');
          keys.className = 'gcard__keys';
          keys.textContent = item.keys.replace('Cmd', CMD);
          top.appendChild(keys);
        }
        card.appendChild(top);

        var hint = document.createElement('div');
        hint.className = 'gcard__hint';
        hint.textContent = item.hint;
        card.appendChild(hint);

        if (item.syntax) {
          var syn = document.createElement('pre');
          syn.className = 'gcard__syntax';
          syn.textContent = item.syntax.split('\\n').join('\n');
          card.appendChild(syn);
        }

        var btn = document.createElement('button');
        btn.className = 'gcard__ins';
        btn.textContent = item.action ? 'run' : 'insert at caret';
        btn.addEventListener('click', function () {
          closeOverlays();
          if (item.action) runAction(item.action);
          else insertSnippet(item);
        });
        card.appendChild(btn);

        sec.appendChild(card);
      });

      guideBody.appendChild(sec);
    });
  }

  function closeOverlays(keepFocus) {
    // A dropdown outranks the picker in the stacking order, so it has to go
    // down too or it floats over the modal that just opened.
    closeMenu();
    palette.classList.remove('is-on');
    guide.classList.remove('is-on');
    $('library').classList.remove('is-on');
    if ($('themes').classList.contains('is-on')) closeThemes(true);
    updateDocTitle();
    if (!keepFocus) src.focus();
  }

  Array.prototype.forEach.call(document.querySelectorAll('[data-close]'), function (el) {
    el.addEventListener('click', function () { closeOverlays(); });
  });

  // ── Actions ────────────────────────────────────────────────────────────
  function slugify(s) {
    return (s || 'deck').trim().toLowerCase().replace(/[^a-z0-9._-]+/g, '-').replace(/^-|-$/g, '') || 'deck';
  }

  function download() {
    var name = slugify($('docname').value);
    if (state.kind === 'html') {
      if (!/\.html?$/.test(name)) name += '.html';
      saveBlob(name, new Blob([srcHtml.value], { type: 'text/html;charset=utf-8' }));
      toast('Downloaded ' + name);
      return;
    }
    if (!/\.(md|markdown)$/.test(name)) name += '.md';
    saveBlob(name, new Blob([src.value], { type: 'text/markdown;charset=utf-8' }));
    toast('Downloaded ' + name);
  }

  function saveBlob(name, blob) {
    var a = document.createElement('a');
    a.href = URL.createObjectURL(blob);
    a.download = name;
    document.body.appendChild(a);
    a.click();
    setTimeout(function () { URL.revokeObjectURL(a.href); a.remove(); }, 1000);
  }

  function isEmpty() {
    return state.kind === 'html' ? !srcHtml.value.trim() : !state.slides.length;
  }

  /** Downloads the built deck or doc as one standalone page. */
  function exportHtml() {
    if (isEmpty()) { toast('Nothing to export yet.', 'warn'); return; }
    var builder = state.kind === 'html' ? buildHtmlDoc : buildDeck;
    builder(false, true)
      .then(function (data) { return fetch(data.path); })
      .then(function (r) {
        if (!r.ok) throw new Error('server said ' + r.status);
        return r.text();
      })
      .then(function (html) {
        var name = slugify($('docname').value).replace(/\.(md|markdown|html?)$/, '') + '.html';
        saveBlob(name, new Blob([html], { type: 'text/html;charset=utf-8' }));
        toast('Downloaded ' + name + '. Fonts and highlighting load from a CDN, so it needs a connection.');
      })
      .catch(function (err) { toast('Could not export HTML: ' + err.message, 'err'); });
  }

  /**
   * Exports a real PDF file. The server drives a headless browser, so nobody
   * has to find the landscape and background-graphics settings in a dialog.
   * Falls back to the dialog only when the machine has no browser to drive.
   */
  var pdfInFlight = false;

  function exportPdf() {
    if (isEmpty()) { toast('Nothing to export yet.', 'warn'); return; }
    if (pdfInFlight) { toast('Already building a PDF.'); return; }
    pdfInFlight = true;
    elSave.className = 'warn';
    elSave.textContent = 'building the PDF';

    var pdfUrl = state.kind === 'html' ? '/__pdf-doc' : '/__pdf';
    var pdfBody = state.kind === 'html'
      ? { html: srcHtml.value, title: $('docname').value }
      : {
          markdown: src.value,
          theme: state.theme,
          head: state.head,
          body: state.body,
          template: state.template,
          transition: state.transition,
          title: $('docname').value
        };

    fetch(pdfUrl, {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify(pdfBody)
    })
      .then(function (r) {
        if (r.status === 501) {
          return r.json().then(function (info) {
            printFallback(info && info.detail);
            return null;
          });
        }
        if (!r.ok) {
          return r.json().then(
            function (info) { throw new Error((info && info.detail) || 'server said ' + r.status); },
            function () { throw new Error('server said ' + r.status); }
          );
        }
        return r.blob();
      })
      .then(function (blob) {
        if (!blob) return;
        var name = slugify($('docname').value).replace(/\.(md|markdown|html?)$/, '') + '.pdf';
        saveBlob(name, blob);
        toast('Downloaded ' + name + (state.kind === 'html' ? '.' : ', one 16:9 page per slide.'));
      })
      .catch(function (err) { toast('Could not export the PDF: ' + err.message, 'err'); })
      .then(function () {
        pdfInFlight = false;
        elSave.className = '';
        elSave.textContent = '';
        scheduleSave();
      });
  }

  /** No local browser to drive: hand the deck or doc to the print dialog instead. */
  function printFallback(detail) {
    var tab = window.open('about:blank', '_blank');
    var builder = state.kind === 'html' ? buildHtmlDoc : buildDeck;
    builder(true)
      .then(function (data) {
        var url = location.origin + data.path;
        if (!tab) {
          toast('Allow pop-ups, or press Cmd P, to save a PDF.', 'warn', 'open here', function () { location.href = url; });
          return;
        }
        tab.location.replace(url);
        toast((detail || 'No local browser to render with.') + ' Using the print dialog instead, which is already set to landscape.', 'warn');
      })
      .catch(function (err) {
        if (tab) tab.close();
        toast('Could not build the page: ' + err.message, 'err');
      });
  }

  function buildDeck(forPrint, standalone) {
    return fetch('/__present', {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({
        markdown: src.value,
        theme: state.theme,
        head: state.head,
        body: state.body,
        template: state.template,
        transition: state.transition,
        title: $('docname').value,
        print: !!forPrint,
        standalone: !!standalone
      })
    }).then(function (r) {
      if (!r.ok) throw new Error('server said ' + r.status);
      return r.json();
    });
  }

  function buildHtmlDoc(forPrint) {
    return fetch('/__present-doc', {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({
        html: srcHtml.value,
        theme: state.theme,
        title: $('docname').value,
        print: !!forPrint
      })
    }).then(function (r) {
      if (!r.ok) throw new Error('server said ' + r.status);
      return r.json();
    });
  }

  function present() {
    if (isEmpty()) { toast('Nothing to present yet.', 'warn'); return; }
    expectingDeck = true;
    var builder = state.kind === 'html' ? buildHtmlDoc : buildDeck;
    var tab = window.open('about:blank', '_blank');
    builder(false)
      .then(function (data) {
        var url = location.origin + data.path +
          (data.path.indexOf('?') >= 0 ? '&' : '?') + 'ps=' + followSid +
          '&hl=' + encodeURIComponent(highlightKey());
        if (tab) tab.location.replace(url);
        else toast('Allow pop-ups to present in a new tab.', 'warn', 'present here', function () { location.href = url; });
      })
      .catch(function (err) {
        if (tab) tab.close();
        toast('Could not build the page: ' + err.message, 'err');
      });
  }

  /**
   * Applies a theme everywhere at once: the editor chrome, the preview iframe,
   * and the decor attribute the backdrop patterns key off. A remember of false
   * means the picker is only previewing, so arrowing past a theme never
   * overwrites the one the user actually chose.
   */
  function setTheme(next, remember) {
    var id = pickTheme(next);
    state.theme = id;
    document.documentElement.dataset.theme = id;
    document.documentElement.dataset.decor = THEME_BY_ID[id].decor;
    if (remember !== false) lsSet(K.theme, id);
    pushLook();
  }

  function setTemplate(next, remember) {
    var id = pickTemplate(next);
    state.template = id;
    document.documentElement.dataset.template = id;
    if (remember !== false) {
      lsSet(K.template, id);
      scheduleSave();
    }
    pushLook();
    paintTemplateMenu();
  }

  function setTransition(next, remember) {
    var id = pickTransition(next);
    state.transition = id;
    document.documentElement.dataset.transition = id;
    if (remember !== false) {
      lsSet(K.transition, id);
      scheduleSave();
    }
    pushLook();
    paintTemplateMenu();
  }

  /**
   * Heading and body faces are chosen separately, and either can be handed
   * back to the theme by picking nothing. Passing null is what clears it.
   */
  function setFont(slot, id, remember) {
    if (slot !== 'head' && slot !== 'body') return;
    var next = pickFont(id);
    state[slot] = next;
    if (remember !== false) {
      // '' rather than a missing key, so a cleared override survives a reload
      // instead of falling back to whatever the CLI started with.
      lsSet(slot === 'head' ? K.head : K.body, next || '');
    }
    pushLook();
    paintFontMenu();
  }

  function syncHtmlFrameTheme() {
    if (!frameHtml) return;
    try {
      var root = frameHtml.contentDocument && frameHtml.contentDocument.documentElement;
      if (root) {
        root.dataset.theme = state.theme;
        root.dataset.decor = decorOf(state.theme);
        root.dataset.template = state.template;
        root.dataset.transition = state.transition;
        if (state.head) root.dataset.head = state.head; else delete root.dataset.head;
        if (state.body) root.dataset.body = state.body; else delete root.dataset.body;
      }
    } catch (e) {}
    try {
      if (frameHtml.contentWindow) {
        frameHtml.contentWindow.postMessage({
          type: 'theme',
          theme: state.theme,
          head: state.head,
          body: state.body,
          template: state.template,
          transition: state.transition
        }, '*');
      }
    } catch (e) {}
  }

  /** One message for all four, since the preview applies them to one root. */
  function pushLook() {
    post({
      type: 'theme',
      theme: state.theme,
      head: state.head,
      body: state.body,
      template: state.template,
      transition: state.transition
    });
    syncHtmlFrameTheme();
    paintThemeButton();
    paintTemplateMenu();
  }

  function faceLabel(id) {
    return id && FACE_BY_ID[id] ? FACE_BY_ID[id].name : 'theme';
  }

  function buildFontMenu() {
    [['head', 'ff-head'], ['body', 'ff-body']].forEach(function (pair) {
      var slot = pair[0];
      var host = $(pair[1]);
      host.innerHTML = '';

      host.appendChild(fontRow(slot, null, 'theme default', null));

      var lastKind = null;
      FACES.forEach(function (f) {
        if (f.kind !== lastKind) {
          lastKind = f.kind;
          var head = document.createElement('div');
          head.className = 'ff-col__title';
          head.textContent = f.kind;
          host.appendChild(head);
        }
        host.appendChild(fontRow(slot, f.id, f.name, f));
      });
    });
    paintFontMenu();
  }

  function fontRow(slot, id, label, face) {
    var b = document.createElement('button');
    b.className = 'ff-row' + (id ? '' : ' ff-row--auto');
    b.dataset.slot = slot;
    if (id) b.dataset.font = id;

    var name = document.createElement('span');
    name.className = 'ff-row__name';
    name.textContent = label;
    if (face) name.style.fontFamily = face.stack;
    b.appendChild(name);

    return b;
  }

  function paintFontMenu() {
    Array.prototype.forEach.call($('font-menu').querySelectorAll('.ff-row'), function (b) {
      var slot = b.dataset.slot;
      b.classList.toggle('is-on', (b.dataset.font || null) === state[slot]);
    });
    var custom = state.head || state.body;
    $('font-label').textContent = custom
      ? faceLabel(state.head) + ' / ' + faceLabel(state.body)
      : 'font';
    $('btn-font').title = 'Heading ' + faceLabel(state.head) +
      ' · body ' + faceLabel(state.body);
  }

  function paintThemeButton() {
    var t = THEME_BY_ID[state.theme];
    $('theme-label').textContent = t ? t.label : 'theme';
    $('btn-theme').title = t ? t.label + ' — ' + t.blurb : 'Pick a theme';
  }

  function optionRow(kind, item) {
    var b = document.createElement('button');
    b.dataset[kind] = item.id;
    var name = document.createElement('span');
    name.className = 'menu__name';
    name.textContent = item.label;
    var hint = document.createElement('span');
    hint.className = 'menu__hint';
    hint.textContent = item.blurb;
    b.appendChild(name);
    b.appendChild(hint);
    return b;
  }

  function buildTemplateMenu() {
    var templateHost = $('template-list');
    var transitionHost = $('transition-list');
    templateHost.innerHTML = '';
    transitionHost.innerHTML = '';
    TEMPLATES.forEach(function (item) { templateHost.appendChild(optionRow('template', item)); });
    TRANSITIONS.forEach(function (item) { transitionHost.appendChild(optionRow('transition', item)); });
    paintTemplateMenu();
  }

  function paintTemplateMenu() {
    var menu = $('template-menu');
    if (!menu) return;
    Array.prototype.forEach.call(menu.querySelectorAll('[data-template]'), function (b) {
      b.classList.toggle('is-on', b.dataset.template === state.template);
    });
    Array.prototype.forEach.call(menu.querySelectorAll('[data-transition]'), function (b) {
      b.classList.toggle('is-on', b.dataset.transition === state.transition);
    });
    var current = TEMPLATE_BY_ID[state.template];
    $('template-label').textContent = current ? current.label : 'template';
    $('btn-template').title = (current ? current.blurb : 'Composition template') +
      ' · transition ' + state.transition;
  }

  // ── Theme picker ───────────────────────────────────────────────────────
  var themeCards = [];
  var themeSel = 0;
  var themeCommitted = null;

  /** A slide in miniature, painted with the theme's own palette and faces. */
  function themeThumb(t) {
    var c = t.colors;
    var wrap = document.createElement('div');
    wrap.className = 'th-thumb';
    wrap.style.background = c.crust;

    var glow = document.createElement('div');
    glow.className = 'th-thumb__glow';
    glow.style.background = 'radial-gradient(circle, ' + c.accent + '2e 0%, transparent 68%)';
    wrap.appendChild(glow);

    var title = document.createElement('div');
    title.className = 'th-thumb__title';
    title.textContent = 'Aa';
    title.style.fontFamily = t.fonts.display;
    title.style.fontWeight = '700';
    title.style.color = c.accent;
    wrap.appendChild(title);

    var rule = document.createElement('div');
    rule.className = 'th-thumb__rule';
    rule.style.background = 'linear-gradient(90deg, ' + c.accent + ', ' + c.accent2 + ')';
    wrap.appendChild(rule);

    [88, 70, 52].forEach(function (pct, i) {
      var line = document.createElement('div');
      line.className = 'th-thumb__line';
      line.style.width = pct + '%';
      line.style.background = i === 0 ? c.subtext1 : c.overlay1;
      line.style.opacity = i === 0 ? '0.85' : '0.5';
      wrap.appendChild(line);
    });

    var code = document.createElement('div');
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
    var card = document.createElement('button');
    card.className = 'th-card' + (t.id === state.theme ? ' is-current' : '');
    card.dataset.theme = t.id;
    card.appendChild(themeThumb(t));

    var meta = document.createElement('div');
    meta.className = 'th-meta';

    var top = document.createElement('div');
    top.className = 'th-meta__top';
    var name = document.createElement('span');
    name.className = 'th-meta__name';
    name.textContent = t.label;
    var mood = document.createElement('span');
    mood.className = 'th-meta__mood';
    mood.textContent = t.mood;
    top.appendChild(name);
    top.appendChild(mood);

    var blurb = document.createElement('div');
    blurb.className = 'th-meta__blurb';
    blurb.textContent = t.blurb;

    meta.appendChild(top);
    meta.appendChild(blurb);
    card.appendChild(meta);
    return card;
  }

  function buildThemePicker() {
    var host = $('th-list');
    host.innerHTML = '';
    themeCards = [];

    [['dark', 'dark'], ['light', 'light']].forEach(function (pair) {
      var list = THEMES.filter(function (t) { return t.mood === pair[0]; });
      if (!list.length) return;

      var label = document.createElement('div');
      label.className = 'th-group';
      label.textContent = pair[1];
      host.appendChild(label);

      var grid = document.createElement('div');
      grid.className = 'th-grid';
      list.forEach(function (t) {
        var card = themeCard(t);
        // Hovering previews too, so the mouse gets the same instant feedback
        // as the arrow keys.
        card.addEventListener('mouseenter', function () { selectTheme(themeCards.indexOf(card)); });
        card.addEventListener('click', function () { selectTheme(themeCards.indexOf(card)); commitTheme(); });
        grid.appendChild(card);
        themeCards.push(card);
      });
      host.appendChild(grid);
    });
  }

  function selectTheme(i) {
    if (i < 0 || i >= themeCards.length) return;
    themeSel = i;
    themeCards.forEach(function (c, n) { c.classList.toggle('is-sel', n === i); });
    setTheme(themeCards[i].dataset.theme, false);
  }

  function openThemes() {
    closeOverlays(true);
    themeCommitted = state.theme;
    buildThemePicker();
    var at = 0;
    themeCards.forEach(function (c, n) { if (c.dataset.theme === state.theme) at = n; });
    $('themes').classList.add('is-on');
    selectTheme(at);
    themeCards[themeSel].scrollIntoView({ block: 'nearest' });
    src.blur();
  }

  function commitTheme() {
    themeCommitted = state.theme;
    lsSet(K.theme, state.theme);
    closeThemes();
    var t = THEME_BY_ID[state.theme];
    toast('Theme set to ' + (t ? t.label : state.theme) + '.');
  }

  /** Leaving without choosing puts back whatever was on when the picker opened. */
  function closeThemes(restore) {
    if (restore && themeCommitted && themeCommitted !== state.theme) {
      setTheme(themeCommitted, true);
    }
    $('themes').classList.remove('is-on');
    themeCards.forEach(function (c) { c.classList.remove('is-sel'); });
  }

  function themesOpen() { return $('themes').classList.contains('is-on'); }

  function setMode(next) {
    state.mode = next;
    lsSet(K.mode, next);
    $('seg-single').classList.toggle('is-on', next === 'single');
    $('seg-grid').classList.toggle('is-on', next === 'grid');
    layout();
    pushFrame();
  }

  /** The status-bar counters' HTML-doc analog: chars and caret position. */
  function renderHtmlCounts() {
    elChipSlides.textContent = srcHtml.value.length + ' chars';
    var before = srcHtml.value.slice(0, srcHtml.selectionStart || 0);
    var lines = before.split('\n');
    elPos.textContent = 'Ln ' + lines.length + ', Col ' + (lines[lines.length - 1].length + 1);
    elWords.textContent = '';
  }

  /** Load a deck or doc into the editor, saving whatever is open first. */
  /** File mode owns the session: the browser library stays out of the way. */
  function fileModeBlocks() {
    if (!FILE) return false;
    toast('Editing ' + FILE.name + ' — the browser library is available when deckrun runs without a file.', 'warn');
    return true;
  }

  function openDeck(id, announce) {
    if (fileModeBlocks()) return;
    var meta = findDeck(id);
    if (!meta) { toast('That deck is gone.', 'warn'); return; }
    if (state.deckId && id !== state.deckId) saveNow();
    state.deckId = id;
    lsSet(K.current, id);
    syncHighlightDoc();
    setDocKind(meta.kind || 'markdown');
    setTemplate(meta.template || state.template, true);
    setTransition(meta.transition || state.transition, true);
    $('docname').value = meta.name;
    updateDocTitle();
    updateDeckCount();
    elSave.className = '';
    elSave.textContent = '';

    if (state.kind === 'html') {
      srcHtml.value = readDeck(id);
      frameHtml.srcdoc = srcHtml.value;
      renderHtmlCounts();
      if (announce) toast('Opened ' + meta.name);
      srcHtml.focus();
      return;
    }

    src.value = readDeck(id);
    src.setSelectionRange(0, 0);
    src.scrollTop = 0;
    state.index = 0;
    state.overflow = {};
    paint();
    syncScroll();
    updateCaretUi();
    refresh();
    if (announce) toast('Opened ' + meta.name);
    src.focus();
  }

  function noRoom() {
    toast('This browser has no room for another deck. Delete one, or download this one first.', 'err');
  }

  function newDeck() {
    if (fileModeBlocks()) return;
    saveNow();
    var id = createDeck(uniqueName('untitled'), '', 'markdown');
    if (!id) { noRoom(); return; }
    openDeck(id);
    toast('New deck. Press ' + CMD + ' K to see what you can add.');
  }

  function duplicateDeck() {
    if (fileModeBlocks()) return;
    saveNow();
    var meta = findDeck(state.deckId);
    var id = createDeck(uniqueName((meta ? meta.name : 'deck') + ' copy'), curValue(), state.kind);
    if (!id) { noRoom(); return; }
    openDeck(id);
    toast('Duplicated');
  }

  function deleteDeck(id) {
    if (fileModeBlocks()) return;
    var meta = findDeck(id);
    if (!meta) return;
    if (!confirm('Delete "' + meta.name + '"? This cannot be undone, and it is only in this browser.')) return;

    var list = loadIndex().filter(function (d) { return d.id !== id; });
    saveIndex(list);
    lsDel(K.deck + id);
    if (window.deckrunHighlights) window.deckrunHighlights.forget('deck:' + id);

    if (id === state.deckId) {
      if (list.length) {
        openDeck(list[0].id);
      } else {
        var fresh = createDeck('deck', '', 'markdown');
        if (fresh) openDeck(fresh);
      }
    }
    updateDeckCount();
    renderLibrary();
    toast('Deleted ' + meta.name);
  }

  // ── Library panel ──────────────────────────────────────────────────────
  var libRows = [], libSel = 0;

  /** Most recently touched first, the useful order for a switcher. */
  function libraryList() {
    return loadIndex().slice().sort(function (a, b) { return (b.at || 0) - (a.at || 0); });
  }

  function renderLibrary() {
    var list = libraryList();
    libRows = list;
    if (libSel >= list.length) libSel = Math.max(0, list.length - 1);
    var host = $('lib-list');
    host.innerHTML = '';

    list.forEach(function (deck, i) {
      var row = document.createElement('div');
      row.className = 'lib-row' + (i === libSel ? ' is-sel' : '') +
        (deck.id === state.deckId ? ' is-current' : '');

      var main = document.createElement('div');
      main.className = 'lib-row__main';
      var name = document.createElement('div');
      name.className = 'lib-row__name';
      name.textContent = deck.name + (deck.id === state.deckId ? '  (open)' : '');
      var meta = document.createElement('div');
      meta.className = 'lib-row__meta';
      var kind = deck.kind || 'markdown';
      if (kind === 'html') {
        var chars = deck.id === state.deckId ? srcHtml.value.length : deck.chars;
        meta.textContent = 'html' +
          '  ·  ' + Math.max(1, Math.round((chars || 0) / 1024)) + ' KB' +
          '  ·  ' + timeAgo(deck.at || Date.now());
      } else {
        var slides = deck.id === state.deckId ? state.slides.length : deck.slides;
        meta.textContent = (slides || 0) + (slides === 1 ? ' slide' : ' slides') +
          '  ·  ' + Math.max(1, Math.round((deck.chars || 0) / 1024)) + ' KB' +
          '  ·  ' + timeAgo(deck.at || Date.now());
      }
      main.appendChild(name);
      main.appendChild(meta);
      row.appendChild(main);

      var acts = document.createElement('div');
      acts.className = 'lib-row__acts';

      var dupe = document.createElement('button');
      dupe.textContent = 'duplicate';
      dupe.addEventListener('click', function (e) {
        e.stopPropagation();
        if (deck.id === state.deckId) { duplicateDeck(); renderLibrary(); return; }
        var id = createDeck(
          uniqueName(deck.name + ' copy'),
          readDeck(deck.id),
          deck.kind || 'markdown',
          deck.template || state.template,
          deck.transition || state.transition
        );
        if (!id) { noRoom(); return; }
        updateDeckCount();
        renderLibrary();
        toast('Duplicated ' + deck.name);
      });
      acts.appendChild(dupe);

      var del = document.createElement('button');
      del.className = 'danger';
      del.textContent = 'delete';
      del.addEventListener('click', function (e) { e.stopPropagation(); deleteDeck(deck.id); });
      acts.appendChild(del);

      row.appendChild(acts);
      row.addEventListener('mousemove', function () {
        if (libSel === i) return;
        libSel = i;
        var sel = host.querySelector('.lib-row.is-sel');
        if (sel) sel.classList.remove('is-sel');
        row.classList.add('is-sel');
      });
      row.addEventListener('click', function () { closeOverlays(); openDeck(deck.id, true); });
      host.appendChild(row);
    });
  }

  function openLibrary() {
    if (fileModeBlocks()) return;
    var list = libraryList();
    libSel = 0;
    for (var i = 0; i < list.length; i++) if (list[i].id === state.deckId) libSel = i;
    renderLibrary();
    $('library').classList.add('is-on');
    updateDocTitle();
  }

  $('lib-new').addEventListener('click', function () { closeOverlays(); showStartScreen(); });

  $('btn-decks').addEventListener('click', openLibrary);

  // ── Start screen ─────────────────────────────────────────────────────
  // Shown on a true first run, and whenever "new" is chosen explicitly from
  // the library — never on an ordinary launch that has a deck to resume.
  var startTemplate = state.template;
  var startTransition = state.transition;

  function paintStartOptions() {
    Array.prototype.forEach.call($('start-template-list').querySelectorAll('[data-template]'), function (b) {
      b.classList.toggle('is-on', b.dataset.template === startTemplate);
    });
    Array.prototype.forEach.call($('start-transition-list').querySelectorAll('[data-transition]'), function (b) {
      b.classList.toggle('is-on', b.dataset.transition === startTransition);
    });
  }

  function buildStartOptions() {
    var templateHost = $('start-template-list');
    var transitionHost = $('start-transition-list');
    templateHost.innerHTML = '';
    transitionHost.innerHTML = '';
    TEMPLATES.forEach(function (item) {
      var b = document.createElement('button');
      b.className = 'start-choice';
      b.dataset.template = item.id;
      b.textContent = item.label;
      b.title = item.blurb;
      b.addEventListener('click', function () { startTemplate = item.id; paintStartOptions(); });
      templateHost.appendChild(b);
    });
    TRANSITIONS.forEach(function (item) {
      var b = document.createElement('button');
      b.className = 'start-choice';
      b.dataset.transition = item.id;
      b.textContent = item.label;
      b.title = item.blurb;
      b.addEventListener('click', function () { startTransition = item.id; paintStartOptions(); });
      transitionHost.appendChild(b);
    });
    paintStartOptions();
  }

  function showStartScreen() {
    startTemplate = state.template;
    startTransition = state.transition;
    paintStartOptions();
    $('start-lib').classList.toggle('is-disabled', !loadIndex().length);
    $('screen-start').classList.add('is-on');
    updateDocTitle();
  }

  function hideStartScreenSilently() {
    $('screen-start').classList.remove('is-on');
    updateDocTitle();
  }

  /** Escape/backdrop close. With nothing open yet, that would be a dead end,
      so the one specified fallback is a blank Markdown deck. */
  function hideStartScreen(fallbackIfEmpty) {
    hideStartScreenSilently();
    if (fallbackIfEmpty && !state.deckId) startNewMarkdown();
  }

  function startNewMarkdown() {
    setTemplate(startTemplate, true);
    setTransition(startTransition, true);
    var id = createDeck(uniqueName('untitled'), D.welcome, 'markdown', startTemplate, startTransition);
    hideStartScreenSilently();
    if (!id) {
      // Storage refuses even a first write: use the deck unsaved rather than
      // leaving the editor with nothing open.
      state.deckId = newDeckId();
      setDocKind('markdown');
      src.value = D.welcome;
      $('docname').value = 'deck';
      src.setSelectionRange(0, 0);
      src.scrollTop = 0;
      paint();
      syncScroll();
      updateCaretUi();
      refresh();
      updateDeckCount();
      elSave.className = 'err';
      elSave.textContent = 'this browser blocks local storage';
      setTimeout(function () {
        toast('This browser will not let the editor save anything. Download your deck to keep it.', 'err', 'download', download);
      }, 700);
      src.focus();
      return;
    }
    openDeck(id);
    setTimeout(function () {
      toast('This deck is yours to overwrite. Press ' + CMD + ' K to see every layout and style.', null, 'open the guide', function () { runAction('guide'); });
    }, 900);
  }

  function startOpenLibrary() {
    if (!loadIndex().length) return;
    hideStartScreenSilently();
    openLibrary();
  }

  /** Fetches a public URL server-side (sidesteps CORS) and opens it as a new Markdown deck or HTML doc. */
  function loadHtmlUrl() {
    var input = $('start-html-url');
    var raw = input.value.trim();
    if (!raw) { input.focus(); return; }
    var url;
    try { url = new URL(raw); } catch (err) { toast('Not a valid URL.', 'err'); return; }
    if (url.protocol !== 'http:' && url.protocol !== 'https:') {
      toast('URL must be http or https.', 'err');
      return;
    }
    var btn = $('start-html-url-go');
    btn.disabled = true;
    fetch('/__fetch-doc', {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ url: url.toString() }),
    })
      .then(function (res) {
        return res.json().then(function (body) {
          if (!res.ok) throw new Error(body.detail || body.error || 'could not load that URL');
          return body;
        });
      })
      .then(function (body) {
        saveNow();
        var name = uniqueName(body.title || url.hostname);
        var kind = body.kind || (body.markdown ? 'markdown' : 'html');
        var content = kind === 'markdown' ? (body.markdown || body.content || '') : (body.html || body.content || '');
        var id = createDeck(
          name,
          content,
          kind,
          kind === 'markdown' ? startTemplate : state.template,
          kind === 'markdown' ? startTransition : state.transition
        );
        if (!id) { noRoom(); return; }
        input.value = '';
        hideStartScreenSilently();
        openDeck(id);
        toast('Loaded ' + (kind === 'markdown' ? 'Markdown deck' : 'HTML doc') + ' from ' + url.hostname);
      })
      .catch(function (err) { toast('Could not load URL: ' + err.message, 'err'); })
      .then(function () { btn.disabled = false; });
  }

  $('start-md-blank').addEventListener('click', startNewMarkdown);
  $('start-md-upload').addEventListener('click', function () { $('file-any').click(); });
  $('start-html-upload').addEventListener('click', function () { $('file-any').click(); });
  $('start-html-url-go').addEventListener('click', loadHtmlUrl);
  $('start-html-url').addEventListener('keydown', function (e) {
    if (e.key === 'Enter') { e.preventDefault(); loadHtmlUrl(); }
  });
  $('start-lib').addEventListener('click', startOpenLibrary);
  $('screen-start').querySelector('.backdrop').addEventListener('click', function () { hideStartScreen(true); });

  function runAction(name) {
    if (name === 'present') present();
    else if (name === 'download') download();
    else if (name === 'pdf') exportPdf();
    else if (name === 'html') exportHtml();
    else if (name === 'open') $('file-any').click();
    else if (name === 'grid') setMode(state.mode === 'grid' ? 'single' : 'grid');
    else if (name === 'theme') openThemes();
    else if (name === 'guide') { guide.classList.add('is-on'); }
    else if (name === 'palette') openPalette();
    else if (name === 'decks') openLibrary();
    else if (name === 'new') newDeck();
    else if (name === 'duplicate') duplicateDeck();
    else if (name === 'delete') deleteDeck(state.deckId);
  }

  // ── Wiring ─────────────────────────────────────────────────────────────
  var parseTimer = null;

  function onInput() {
    // Typing takes the editor back from the deck.
    if (following) { following = false; followChipEl.classList.add('hidden'); }
    paint();
    syncScroll();
    updateCaretUi();
    scheduleSave();
    if (parseTimer) clearTimeout(parseTimer);
    parseTimer = setTimeout(refresh, 140);
  }

  src.addEventListener('input', onInput);
  src.addEventListener('scroll', syncScroll);
  src.addEventListener('click', updateCaretUi);
  src.addEventListener('keyup', function (e) {
    if (e.key.indexOf('Arrow') === 0 || e.key === 'Home' || e.key === 'End' ||
        e.key === 'PageUp' || e.key === 'PageDown') updateCaretUi();
  });

  src.addEventListener('keydown', function (e) {
    if (e.key === 'Tab') {
      e.preventDefault();
      typeText('  ');
      onInput();
    }
  });

  // ── HTML doc source ──────────────────────────────────────────────────
  var htmlPreviewTimer = null;

  function onHtmlInput() {
    scheduleSave();
    renderHtmlCounts();
    if (htmlPreviewTimer) clearTimeout(htmlPreviewTimer);
    htmlPreviewTimer = setTimeout(function () { frameHtml.srcdoc = srcHtml.value; }, 140);
  }

  srcHtml.addEventListener('input', onHtmlInput);
  frameHtml.addEventListener('load', syncHtmlFrameTheme);
  srcHtml.addEventListener('click', renderHtmlCounts);
  srcHtml.addEventListener('keyup', function (e) {
    if (e.key.indexOf('Arrow') === 0 || e.key === 'Home' || e.key === 'End' ||
        e.key === 'PageUp' || e.key === 'PageDown') renderHtmlCounts();
  });
  srcHtml.addEventListener('keydown', function (e) {
    if (e.key !== 'Tab') return;
    e.preventDefault();
    srcHtml.focus();
    var inserted = false;
    try { inserted = document.execCommand('insertText', false, '  '); } catch (err) {}
    if (!inserted) {
      var s = srcHtml.selectionStart, e2 = srcHtml.selectionEnd;
      srcHtml.setRangeText('  ', s, e2, 'end');
    }
    onHtmlInput();
  });

  // Drag and drop a .md file onto the editor to open it.
  ['dragenter', 'dragover'].forEach(function (type) {
    paneEdit.addEventListener(type, function (e) {
      if (!e.dataTransfer) return;
      e.preventDefault();
      paneEdit.classList.add('is-dropping');
    });
  });

  ['dragleave', 'dragend'].forEach(function (type) {
    paneEdit.addEventListener(type, function (e) {
      if (e.target === paneEdit || type === 'dragend') paneEdit.classList.remove('is-dropping');
    });
  });

  paneEdit.addEventListener('drop', function (e) {
    e.preventDefault();
    paneEdit.classList.remove('is-dropping');
    var files = e.dataTransfer ? e.dataTransfer.files : null;
    if (!files || !files.length) return;
    for (var i = 0; i < files.length; i++) {
      var f = files[i];
      if (/\.(md|markdown|txt)$/i.test(f.name)) loadMarkdownFile(f);
      else if (/\.html?$/i.test(f.name)) loadHtmlFile(f);
      else toast('Skipped ' + f.name + '. Drop a Markdown or HTML file to open it.', 'warn');
    }
  });

  /** An opened file lands as a new deck, so nothing in the library is lost. */
  function loadMarkdownFile(file, chosenTemplate, chosenTransition) {
    if (fileModeBlocks()) return;
    var fr = new FileReader();
    fr.onload = function () {
      saveNow();
      var markdown = String(fr.result).replace(/\r\n/g, '\n').replace(/\r/g, '\n');
      var name = uniqueName(file.name.replace(/\.(md|markdown|txt)$/i, '') || 'untitled');
      var id = createDeck(
        name,
        markdown,
        'markdown',
        chosenTemplate || state.template,
        chosenTransition || state.transition
      );
      if (!id) { noRoom(); return; }
      hideStartScreenSilently();
      openDeck(id);
      toast('Loaded ' + file.name + ' as a new deck');
    };
    fr.onerror = function () { toast('Could not read ' + file.name, 'err'); };
    fr.readAsText(file);
  }

  /** An uploaded HTML doc lands as a new library entry, same as a .md file. */
  function loadHtmlFile(file) {
    if (fileModeBlocks()) return;
    var fr = new FileReader();
    fr.onload = function () {
      saveNow();
      var html = String(fr.result);
      var name = uniqueName(file.name.replace(/\.html?$/i, '') || 'untitled');
      var id = createDeck(name, html, 'html');
      if (!id) { noRoom(); return; }
      hideStartScreenSilently();
      openDeck(id);
      toast('Loaded ' + file.name + ' as a new HTML doc');
    };
    fr.onerror = function () { toast('Could not read ' + file.name, 'err'); };
    fr.readAsText(file);
  }

  $('file-any').addEventListener('change', function (e) {
    var f = e.target.files[0];
    if (f) {
      var fromStart = $('screen-start').classList.contains('is-on');
      if (/\.(md|markdown|txt)$/i.test(f.name)) {
        loadMarkdownFile(
          f,
          fromStart ? startTemplate : state.template,
          fromStart ? startTransition : state.transition
        );
      }
      else if (/\.html?$/i.test(f.name)) loadHtmlFile(f);
      else toast('Unrecognized file type.', 'warn');
    }
    e.target.value = '';
  });

  $('docname').addEventListener('input', function () {
    updateDocTitle();
    scheduleSave();
  });

  $('btn-guide').addEventListener('click', function () { runAction('guide'); });
  $('btn-palette').addEventListener('click', openPalette);
  $('btn-new').addEventListener('click', function () { closeOverlays(); showStartScreen(); });
  // ── Dropdown menus ─────────────────────────────────────────────────────
  // Both the export and font popovers behave the same way, so one open menu is
  // tracked here rather than each growing its own copy of this.
  var openMenu = null;

  function menuIsOpen() { return !!openMenu; }

  function closeMenu() {
    if (!openMenu) return;
    openMenu.classList.remove('is-open');
    var btn = openMenu.querySelector('.btn');
    if (btn) btn.setAttribute('aria-expanded', 'false');
    openMenu = null;
  }

  function toggleMenu(el) {
    var wasOpen = openMenu === el;
    closeMenu();
    if (wasOpen) return;
    openMenu = el;
    el.classList.add('is-open');
    var btn = el.querySelector('.btn');
    if (btn) btn.setAttribute('aria-expanded', 'true');
    var first = el.querySelector('.menu__pop button');
    if (first) first.focus();
  }

  /** Wires one dropdown: its trigger, its arrow keys, and its own Escape. */
  function bindMenu(triggerId, popId, onPick, pickSelector) {
    var wrap = $(triggerId).parentNode;
    var pop = $(popId);

    $(triggerId).addEventListener('click', function (e) {
      e.stopPropagation();
      toggleMenu(wrap);
    });

    pop.addEventListener('click', function (e) {
      var btn = e.target.closest ? e.target.closest(pickSelector) : null;
      if (!btn) return;
      onPick(btn);
    });

    pop.addEventListener('keydown', function (e) {
      var items = Array.prototype.slice.call(pop.querySelectorAll('button'));
      var at = items.indexOf(document.activeElement);
      if (e.key === 'ArrowDown' || e.key === 'ArrowUp') {
        e.preventDefault();
        var next = (at + (e.key === 'ArrowDown' ? 1 : -1) + items.length) % items.length;
        items[next].focus();
      } else if (e.key === 'Escape') {
        e.preventDefault();
        closeMenu();
        src.focus();
      }
    });

    return wrap;
  }

  bindMenu('btn-export', 'export-menu', function (btn) {
    closeMenu();
    runExport(btn.dataset.export);
  }, 'button[data-export]');

  // The font menu stays open across picks: choosing a heading face and then a
  // body face is one decision, and closing between them would fight the user.
  bindMenu('btn-font', 'font-menu', function (btn) {
    setFont(btn.dataset.slot, btn.dataset.font || null, true);
  }, 'button[data-slot]');

  // Template and transition compose, so the menu remains open while either
  // choice is changed and the existing Markdown reflows immediately.
  bindMenu('btn-template', 'template-menu', function (btn) {
    if (btn.dataset.template) setTemplate(btn.dataset.template, true);
    if (btn.dataset.transition) setTransition(btn.dataset.transition, true);
  }, 'button[data-template], button[data-transition]');

  document.addEventListener('click', function (e) {
    if (openMenu && !openMenu.contains(e.target)) closeMenu();
  });

  function runExport(kind) {
    if (kind === 'md') download();
    else if (kind === 'pdf') exportPdf();
    else if (kind === 'html') exportHtml();
  }
  $('btn-theme').addEventListener('click', function () { runAction('theme'); });

  // The picker owns the arrow keys while it is up, and Enter keeps the preview.
  document.addEventListener('keydown', function (e) {
    if (!themesOpen() || e.metaKey || e.ctrlKey || e.altKey) return;
    var cols = 0;
    if (themeCards.length) {
      var grid = themeCards[themeSel].parentNode;
      cols = Math.max(1, Math.round(grid.clientWidth / themeCards[themeSel].offsetWidth));
    }
    var step = 0;
    if (e.key === 'ArrowRight') step = 1;
    else if (e.key === 'ArrowLeft') step = -1;
    else if (e.key === 'ArrowDown') step = cols;
    else if (e.key === 'ArrowUp') step = -cols;
    else if (e.key === 'Enter') { e.preventDefault(); commitTheme(); return; }
    else if (e.key === 'Escape') { e.preventDefault(); closeOverlays(); return; }
    else return;
    e.preventDefault();
    var next = (themeSel + step + themeCards.length) % themeCards.length;
    selectTheme(next);
    themeCards[next].scrollIntoView({ block: 'nearest' });
  }, true);
  $('btn-present').addEventListener('click', present);
  $('btn-prev').addEventListener('click', function () { setIndex(state.index - 1, false); });
  $('btn-next').addEventListener('click', function () { setIndex(state.index + 1, false); });
  $('seg-single').addEventListener('click', function () { setMode('single'); });
  $('seg-grid').addEventListener('click', function () { setMode('grid'); });

  $('pane-prev').addEventListener('mousedown', function () {
    if (document.activeElement && isEditingText()) {
      document.activeElement.blur();
    }
  });

  var panePrevWheelLock = false;
  $('pane-prev').addEventListener('wheel', function (e) {
    if (state.mode === 'grid') return;
    var delta = Math.abs(e.deltaY) >= Math.abs(e.deltaX) ? e.deltaY : e.deltaX;
    if (Math.abs(delta) < 15) return;
    if (panePrevWheelLock) return;
    panePrevWheelLock = true;
    setTimeout(function () { panePrevWheelLock = false; }, 250);
    if (delta > 0) {
      setIndex(state.index + 1, false);
    } else if (delta < 0) {
      setIndex(state.index - 1, false);
    }
  }, { passive: true });

  // ── Global keys ────────────────────────────────────────────────────────
  function isEditingText() {
    var el = document.activeElement;
    if (!el) return false;
    var tag = el.tagName ? el.tagName.toLowerCase() : '';
    return tag === 'textarea' || tag === 'input' || el.isContentEditable;
  }

  document.addEventListener('keydown', function (e) {
    var mod = e.metaKey || e.ctrlKey;

    if (themesOpen()) return;

    if ($('screen-start').classList.contains('is-on')) {
      if (e.key === 'Escape') { e.preventDefault(); hideStartScreen(true); }
      return;
    }

    if ($('library').classList.contains('is-on') && !mod) {
      if (e.key === 'ArrowDown' || e.key === 'ArrowUp') {
        e.preventDefault();
        if (libRows.length) {
          libSel = (libSel + (e.key === 'ArrowDown' ? 1 : -1) + libRows.length) % libRows.length;
          renderLibrary();
          var sel = $('lib-list').querySelector('.lib-row.is-sel');
          if (sel) sel.scrollIntoView({ block: 'nearest' });
        }
        return;
      }
      if (e.key === 'Enter') {
        e.preventDefault();
        if (libRows[libSel]) { closeOverlays(); openDeck(libRows[libSel].id, true); }
        return;
      }
    }

    if (e.key === 'Escape') {
      if (menuIsOpen()) { e.preventDefault(); closeMenu(); src.focus(); return; }
      if (palette.classList.contains('is-on') || guide.classList.contains('is-on') ||
          $('library').classList.contains('is-on')) {
        e.preventDefault();
        closeOverlays();
      }
      return;
    }

    // Option + Control (or Option + Cmd) + arrows/page keys hops between slides (works anywhere, including inside editor)
    if (e.altKey && (e.ctrlKey || e.metaKey)) {
      if (e.key === 'ArrowDown' || e.key === 'ArrowRight' || e.key === 'PageDown') {
        e.preventDefault();
        setIndex(state.index + 1, false);
        return;
      } else if (e.key === 'ArrowUp' || e.key === 'ArrowLeft' || e.key === 'PageUp') {
        e.preventDefault();
        setIndex(state.index - 1, false);
        return;
      }
    }

    if (mod) {
      var key = e.key.toLowerCase();
      var md = state.kind === 'markdown';

      if (key === 'k' && !e.shiftKey && md) { e.preventDefault(); openPalette(); }
      else if (key === 'o') { e.preventDefault(); openLibrary(); }
      else if (key === '/' && md) { e.preventDefault(); runAction('guide'); }
      else if (key === 'p' || (key === 's' && e.shiftKey)) { e.preventDefault(); exportPdf(); }
      else if (key === 's') { e.preventDefault(); saveNow(); download(); }
      else if (key === 'enter') { e.preventDefault(); present(); }
      else if (key === 'd' && md) { e.preventDefault(); insertSnippet(bySnippetId['slide-break']); }
      else if (key === 'b' && md) { e.preventDefault(); insertSnippet(bySnippetId.bold); }
      else if (key === 'i' && md) { e.preventDefault(); insertSnippet(bySnippetId.italic); }
      else if (key === 'e' && md) { e.preventDefault(); insertSnippet(bySnippetId['inline-code']); }
      else if (key === 'g' && md) { e.preventDefault(); runAction('grid'); }
      else if (key === 'l' && e.shiftKey) { e.preventDefault(); runAction('theme'); }
      return;
    }

    if (palette.classList.contains('is-on') || guide.classList.contains('is-on') ||
        $('library').classList.contains('is-on') || menuIsOpen()) {
      return;
    }

    // Direct arrow keys / page keys when not typing into a text field.
    if (!isEditingText()) {
      if (e.key === 'ArrowRight' || e.key === 'ArrowDown' || e.key === ' ' || e.key === 'PageDown') {
        e.preventDefault();
        setIndex(state.index + 1, false);
      } else if (e.key === 'ArrowLeft' || e.key === 'ArrowUp' || e.key === 'PageUp') {
        e.preventDefault();
        setIndex(state.index - 1, false);
      } else if (e.key === 'Home') {
        e.preventDefault();
        setIndex(0, false);
      } else if (e.key === 'End') {
        e.preventDefault();
        setIndex(state.slides.length - 1, false);
      }
    }
  });

  // ── Split pane ─────────────────────────────────────────────────────────
  var splitPct = parseFloat(lsGet(K.split, '48'));
  if (!(splitPct > 15 && splitPct < 85)) splitPct = 48;

  function applySplit() {
    paneEdit.style.flex = '0 0 ' + splitPct + '%';
    layout();
  }

  divider.addEventListener('mousedown', function (e) {
    e.preventDefault();
    divider.classList.add('is-dragging');
    document.body.style.cursor = 'col-resize';

    function move(ev) {
      var box = $('panes').getBoundingClientRect();
      var pct = ((ev.clientX - box.left) / box.width) * 100;
      splitPct = Math.max(18, Math.min(82, pct));
      applySplit();
    }

    function up() {
      divider.classList.remove('is-dragging');
      document.body.style.cursor = '';
      lsSet(K.split, String(Math.round(splitPct)));
      document.removeEventListener('mousemove', move);
      document.removeEventListener('mouseup', up);
    }

    document.addEventListener('mousemove', move);
    document.addEventListener('mouseup', up);
  });

  divider.addEventListener('dblclick', function () {
    splitPct = 48;
    lsSet(K.split, '48');
    applySplit();
  });

  window.addEventListener('resize', layout);

  // ── Frame messages ─────────────────────────────────────────────────────
  window.addEventListener('message', function (e) {
    var m = e.data || {};
    if (m.type === 'ready') {
      // pushLook and pushFrame between them send the whole of the current
      // state, so a message dropped while the frame was loading is already
      // superseded by the time it is up.
      state.frameReady = true;
      pushLook();
      pushFrame();
    } else if (m.type === 'goto') {
      setMode('single');
      setIndex(m.index, true);
    } else if (m.type === 'nav') {
      setIndex(state.index + m.delta, false);
    } else if (m.type === 'index-select') {
      setIndex(m.index, false);
    } else if (m.type === 'action') {
      runAction(m.action);
    } else if (m.type === 'overflow') {
      var had = !!state.overflow[m.index];
      state.overflow[m.index] = m.overflow;
      renderCounts();
      if (had !== m.overflow) evalNudges();
    }
  });

  // ── Boot ───────────────────────────────────────────────────────────────
  state.bootAt = Date.now();

  // Carry over a deck saved before the library existed.
  var legacy = lsGet(K.oldDoc, null);
  if (legacy !== null && !loadIndex().length) {
    createDeck(lsGet(K.oldName, 'deck'), legacy, 'markdown');
    lsDel(K.oldDoc);
    lsDel(K.oldName);
  }

  document.documentElement.dataset.theme = state.theme;
  document.documentElement.dataset.decor = THEME_BY_ID[state.theme].decor;
  document.documentElement.dataset.template = state.template;
  document.documentElement.dataset.transition = state.transition;
  buildFontMenu();
  buildTemplateMenu();
  buildStartOptions();
  paintThemeButton();
  $('seg-single').classList.toggle('is-on', state.mode === 'single');
  $('seg-grid').classList.toggle('is-on', state.mode === 'grid');

  Array.prototype.forEach.call(document.querySelectorAll('.btn kbd, .menu__pop kbd'), function (el) {
    el.textContent = el.textContent.replace('Cmd', CMD);
  });

  buildGuide();
  startTips();
  applySplit();
  paintExportMenu();
  updateDeckCount();

  // ── File mode ──────────────────────────────────────────────────────────
  /** Replaces the open document with what the server just read. */
  function setFileContent(content) {
    if (state.kind === 'html') {
      srcHtml.value = content;
      frameHtml.srcdoc = content;
      renderHtmlCounts();
      return;
    }
    // Keep the caret and scroll roughly in place across an external reload.
    var s = src.selectionStart, e = src.selectionEnd, top = src.scrollTop;
    src.value = content;
    src.setSelectionRange(Math.min(s, content.length), Math.min(e, content.length));
    src.scrollTop = top;
    state.overflow = {};
    paint();
    syncScroll();
    updateCaretUi();
    refresh();
  }

  function loadFromDisk(announce) {
    fetch('/__file')
      .then(function (r) {
        if (!r.ok) throw new Error('HTTP ' + r.status);
        return r.text();
      })
      .then(function (content) {
        if (content === curValue()) return;
        setFileContent(content);
        if (announce) toast('Reloaded ' + FILE.name + ' from disk');
      })
      .catch(function (err) {
        toast('Cannot read ' + FILE.name + ': ' + err.message, 'err');
      });
  }

  function bootFromFile() {
    syncHighlightDoc();
    setDocKind(FILE.kind);
    $('docname').value = FILE.name;
    // The name is the file's; renaming a disk file is the shell's job.
    $('docname').readOnly = true;
    $('btn-decks').style.display = 'none';
    updateDocTitle();
    loadFromDisk(false);
    if (FILE.watched && typeof EventSource !== 'undefined') {
      var es = new EventSource('/__events');
      es.onmessage = function (e) {
        if (e.data === 'reload') loadFromDisk(true);
      };
    }
    (state.kind === 'html' ? srcHtml : src).focus();
  }

  // A launch with a file opens straight into it; otherwise a deck to resume
  // opens with zero clicks, exactly as before, and only a true first run —
  // nothing in the library at all — shows the start screen.
  if (FILE) {
    bootFromFile();
  } else {
    var index = loadIndex();
    if (index.length === 0) {
      showStartScreen();
    } else {
      var wanted = lsGet(K.current, null);
      openDeck(findDeck(wanted) ? wanted : index[0].id);
    }
  }
})();