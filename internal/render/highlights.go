package render

const HighlightWarning = "Highlights and comments live only in this browser session. They travel to the deck you present, and disappear when the tab closes. Nothing is written to disk."

const HighlightRuntime = `(function () {
  'use strict';
  if (window.deckrunHighlights) return;

  var MARK_CSS = "mark.dr-hl {\n  background: #ffe75e;\n  border-radius: 2px;\n  padding: 0 0.06em;\n  box-shadow: inset 0 -0.14em 0 rgba(173, 129, 0, 0.55);\n  cursor: pointer;\n  transition: background 0.12s ease;\n}\n\n/* Body text, links, and every syntax-highlighting class set a colour of their\n   own, and most of them are unreadable on a solid pen stroke. Inside a mark\n   the ink wins, whatever the text was wearing before. */\nmark.dr-hl, mark.dr-hl * { color: #15161a !important; }\n\nmark.dr-hl:hover { background: #ffdb1f; }\n\n/* A highlight carrying a comment is a different pen, not a different shade of\n   the same one: told apart across a room, not by hovering. */\nmark.dr-hl--note {\n  background: #7ae6ff;\n  box-shadow: inset 0 -0.14em 0 rgba(0, 119, 156, 0.5);\n}\n\nmark.dr-hl--note:hover { background: #3ed7ff; }\n\n/* Only the closing fragment of a multi-element highlight carries the dot. */\nmark.dr-hl--note[data-dr-end]::after {\n  content: \"\";\n  display: inline-block;\n  width: 0.4em;\n  height: 0.4em;\n  margin-left: 0.2em;\n  border-radius: 50%;\n  vertical-align: super;\n  background: #0a5f75;\n}";
  var UI_CSS = ".dr-bar, .dr-note, .dr-warn {\n  position: fixed;\n  z-index: 2147483000;\n  font-family: var(--font-mono, ui-monospace, SFMono-Regular, Menlo, monospace);\n  color: var(--text, #cdd6f4);\n}\n\n.dr-bar {\n  display: flex;\n  align-items: center;\n  gap: 2px;\n  padding: 3px;\n  background: var(--mantle, #181825);\n  border: 1px solid var(--surface1, #45475a);\n  border-radius: 8px;\n  box-shadow: 0 10px 30px rgba(0, 0, 0, 0.34);\n}\n\n.dr-bar button, .dr-note button {\n  font: inherit;\n  font-size: 11px;\n  letter-spacing: 0.06em;\n  padding: 4px 9px;\n  border-radius: 5px;\n  border: 1px solid transparent;\n  background: transparent;\n  color: var(--subtext0, #a6adc8);\n  cursor: pointer;\n  white-space: nowrap;\n}\n\n.dr-bar button:hover, .dr-note button:hover {\n  color: var(--text, #cdd6f4);\n  border-color: var(--surface1, #45475a);\n  background: var(--surface0, #313244);\n}\n\n.dr-note button.dr-danger:hover {\n  color: var(--red, #f38ba8);\n  border-color: var(--red, #f38ba8);\n  background: transparent;\n}\n\n.dr-note {\n  width: 272px;\n  padding: 10px 11px 9px;\n  background: var(--mantle, #181825);\n  border: 1px solid var(--surface1, #45475a);\n  border-radius: 10px;\n  box-shadow: 0 14px 40px rgba(0, 0, 0, 0.4);\n}\n\n.dr-note__quote {\n  font-size: 10.5px;\n  line-height: 1.5;\n  color: var(--overlay1, #7f849c);\n  border-left: 2px solid #ffe75e;\n  padding-left: 7px;\n  margin-bottom: 8px;\n  max-height: 46px;\n  overflow: hidden;\n}\n\n.dr-note textarea {\n  display: block;\n  width: 100%;\n  height: 76px;\n  resize: vertical;\n  font: inherit;\n  font-size: 12px;\n  line-height: 1.5;\n  padding: 6px 7px;\n  border-radius: 6px;\n  border: 1px solid var(--surface1, #45475a);\n  background: var(--base, #1e1e2e);\n  color: var(--text, #cdd6f4);\n  outline: none;\n}\n\n.dr-note textarea:focus { border-color: var(--accent, #89b4fa); }\n\n.dr-note__row {\n  display: flex;\n  align-items: center;\n  gap: 3px;\n  margin-top: 7px;\n}\n\n.dr-note__row .dr-spacer { flex: 1 1 auto; }\n\n.dr-warn {\n  left: 50%;\n  bottom: 26px;\n  transform: translateX(-50%);\n  display: flex;\n  align-items: flex-start;\n  gap: 10px;\n  max-width: min(520px, 92vw);\n  padding: 10px 12px;\n  font-size: 11.5px;\n  line-height: 1.6;\n  background: var(--mantle, #181825);\n  border: 1px solid #ffe75e;\n  border-radius: 9px;\n  box-shadow: 0 14px 40px rgba(0, 0, 0, 0.4);\n  color: var(--subtext0, #a6adc8);\n}\n\n.dr-warn b { color: var(--text, #cdd6f4); font-weight: 600; }\n\n.dr-warn button {\n  flex: 0 0 auto;\n  font: inherit;\n  font-size: 14px;\n  line-height: 1;\n  padding: 0 2px;\n  background: transparent;\n  border: none;\n  color: var(--overlay0, #6c7086);\n  cursor: pointer;\n}\n\n.dr-warn button:hover { color: var(--text, #cdd6f4); }";
  var WARNING = "Highlights and comments live only in this browser session. They travel to the deck you present, and disappear when the tab closes. Nothing is written to disk.";
  var KEY = 'deckrun.highlights';
  var SLIDE_SELECTOR = '#presentation .slide[data-index]';
  var TAB = 'tab' + Math.random().toString(36).slice(2);

  // ── Store ────────────────────────────────────────────────────────────
  // One session-scoped blob for every document this browser session has
  // touched, mirrored to every other tab of this origin as it changes.
  var store = null;
  var listeners = [];

  function blank() { return { v: 1, warned: false, docs: {} }; }

  function load() {
    if (store) return store;
    try {
      var raw = window.sessionStorage.getItem(KEY);
      store = raw ? JSON.parse(raw) : null;
    } catch (e) { store = null; }
    if (!store || typeof store !== 'object' || store.v !== 1 || !store.docs) store = blank();
    return store;
  }

  function persist() {
    try { window.sessionStorage.setItem(KEY, JSON.stringify(load())); } catch (e) {}
  }

  var chan = null;
  try { chan = new BroadcastChannel(KEY); } catch (e) { chan = null; }

  if (chan) {
    chan.onmessage = function (e) {
      var m = e.data;
      if (!m || m.tab === TAB || !m.store || m.store.v !== 1) return;
      store = m.store;
      if (!store.docs) store.docs = {};
      persist();
      emit();
    };
  }

  function broadcast() {
    if (!chan) return;
    try { chan.postMessage({ tab: TAB, store: load() }); } catch (e) {}
  }

  function emit() {
    for (var i = 0; i < listeners.length; i++) {
      try { listeners[i](); } catch (e) {}
    }
  }

  /** A change worth telling the other tabs and every mount about. */
  function commit() { persist(); broadcast(); emit(); }

  function bucket(docKey, scopeKey, create) {
    var docs = load().docs;
    var doc = docs[docKey];
    if (!doc) { if (!create) return null; doc = docs[docKey] = {}; }
    var list = doc[scopeKey];
    if (!list) { if (!create) return null; list = doc[scopeKey] = []; }
    return list;
  }

  function newId() {
    return 'h' + Date.now().toString(36) + Math.random().toString(36).slice(2, 7);
  }

  // ── Text addressing ──────────────────────────────────────────────────
  // A highlight is remembered as a character offset into a scope's plain
  // text plus the text itself. Offsets alone would not survive an edit two
  // paragraphs up; the quote is what lets a highlight re-find its home.
  function skip(el) {
    var name = el.nodeName;
    if (name === 'SCRIPT' || name === 'STYLE' || name === 'NOSCRIPT' ||
        name === 'TEXTAREA' || name === 'svg') return true;
    if (el.hasAttribute && el.hasAttribute('data-dr-ui')) return true;
    // Rendered maths and diagrams keep shadow copies of their own source
    // text; splitting either one apart breaks the render.
    return !!(el.classList && (el.classList.contains('katex') || el.classList.contains('mermaid-container')));
  }

  function textNodes(root) {
    var out = [];
    var doc = root.ownerDocument;
    var walker = doc.createTreeWalker(root, NodeFilter.SHOW_TEXT, {
      acceptNode: function (node) {
        if (!node.nodeValue) return NodeFilter.FILTER_REJECT;
        var p = node.parentNode;
        while (p && p !== root) {
          if (p.nodeType === 1 && skip(p)) return NodeFilter.FILTER_REJECT;
          p = p.parentNode;
        }
        return NodeFilter.FILTER_ACCEPT;
      }
    });
    var n;
    while ((n = walker.nextNode())) out.push(n);
    return out;
  }

  function joinText(nodes) {
    var s = '';
    for (var i = 0; i < nodes.length; i++) s += nodes[i].nodeValue;
    return s;
  }

  /** Character offset of a DOM point, measured over a scope's visible text. */
  function pointOffset(root, nodes, container, offset) {
    var range;
    try {
      range = root.ownerDocument.createRange();
      range.selectNodeContents(root);
      range.setEnd(container, offset);
    } catch (e) { return -1; }

    var total = 0;
    for (var i = 0; i < nodes.length; i++) {
      var node = nodes[i];
      var len = node.nodeValue.length;
      var before, after;
      try {
        before = range.comparePoint(node, 0);
        after = range.comparePoint(node, len);
      } catch (e) { continue; }
      if (before > 0) break;
      if (after <= 0) { total += len; continue; }
      // The end lands inside this node, which only happens when the point
      // itself is a text position: an element boundary never splits a node.
      total += (container === node) ? offset : 0;
      break;
    }
    return total;
  }

  /** Where a remembered highlight sits now, or null if its text is gone. */
  function locate(text, item) {
    var quote = item.text || '';
    if (!quote) return null;
    if (text.substr(item.start, quote.length) === quote) {
      return { start: item.start, end: item.start + quote.length };
    }
    var best = -1, bestGap = Infinity, at = text.indexOf(quote);
    while (at !== -1) {
      var gap = Math.abs(at - item.start);
      if (gap < bestGap) { bestGap = gap; best = at; }
      at = text.indexOf(quote, at + 1);
    }
    if (best < 0) return null;
    return { start: best, end: best + quote.length };
  }

  function unwrap(root) {
    var marks = root.querySelectorAll('mark.dr-hl');
    for (var i = 0; i < marks.length; i++) {
      var mark = marks[i];
      var parent = mark.parentNode;
      if (!parent) continue;
      while (mark.firstChild) parent.insertBefore(mark.firstChild, mark);
      parent.removeChild(mark);
      parent.normalize();
    }
  }

  /** Paints one highlight, which may straddle any number of elements. */
  function wrapRange(root, start, end, item) {
    var doc = root.ownerDocument;
    var nodes = textNodes(root);
    var pieces = [];
    var pos = 0;
    for (var i = 0; i < nodes.length && pos < end; i++) {
      var node = nodes[i];
      var len = node.nodeValue.length;
      var from = Math.max(start, pos) - pos;
      var to = Math.min(end, pos + len) - pos;
      // Whitespace between blocks is invisible, and wrapping it can land a
      // mark somewhere the parser will not keep it, such as inside a table.
      if (from < to && node.nodeValue.slice(from, to).trim()) {
        pieces.push({ node: node, from: from, to: to });
      }
      pos += len;
    }

    for (var p = 0; p < pieces.length; p++) {
      var piece = pieces[p];
      var target = piece.node;
      if (piece.to < target.nodeValue.length) target.splitText(piece.to);
      if (piece.from > 0) target = target.splitText(piece.from);
      var mark = doc.createElement('mark');
      mark.className = 'dr-hl' + (item.note ? ' dr-hl--note' : '');
      mark.setAttribute('data-dr-hl', item.id);
      if (item.note) mark.setAttribute('title', item.note);
      if (p === pieces.length - 1) mark.setAttribute('data-dr-end', '1');
      var holder = target.parentNode;
      if (!holder) continue;
      holder.replaceChild(mark, target);
      mark.appendChild(target);
    }
    return pieces.length > 0;
  }

  // ── Mount ────────────────────────────────────────────────────────────
  /**
   * opts.doc      document holding the text (omit when opts.frame is given)
   * opts.frame    iframe whose document holds the text, re-resolved per render
   * opts.docKey   which document's highlights these are
   * opts.scopes   'slides' | 'doc' | function(doc) -> [{ key, el }]
   * opts.onWarn   host's own way of showing the first-use warning
   * opts.readOnly true to only paint highlights made elsewhere (editor's
   *                preview) and skip the create-on-select bar — for the deck
   *                someone is actually presenting, where selecting text to
   *                click through slides should not pop up "highlight".
   */
  function mount(opts) {
    opts = opts || {};
    var frame = opts.frame || null;
    var uiDoc = frame ? frame.ownerDocument : (opts.doc || document);
    var scopes = opts.scopes || 'slides';
    var docKey = opts.docKey || 'default';
    var readOnly = !!opts.readOnly;
    var bound = null;      // document the listeners and observer sit on
    var observer = null;
    var queued = false;
    var applying = false;
    var pending = null;    // the selection the bar is offering to highlight
    var openId = null;     // highlight whose comment card is open
    var bar = null, card = null;

    function target() {
      if (opts.doc) return opts.doc;
      if (!frame) return null;
      try { return frame.contentDocument; } catch (e) { return null; }
    }

    // ── Styling ──
    function styleInto(doc, id, css) {
      if (!doc || !doc.head || doc.getElementById(id)) return;
      var tag = doc.createElement('style');
      tag.id = id;
      tag.setAttribute('data-dr-ui', '1');
      tag.textContent = css;
      doc.head.appendChild(tag);
    }

    // ── Scopes ──
    function scopeList() {
      var doc = target();
      if (!doc || !doc.body) return [];
      if (typeof scopes === 'function') return scopes(doc) || [];
      if (scopes === 'doc') return [{ key: 'doc', el: doc.body }];
      var out = [];
      var els = doc.querySelectorAll(SLIDE_SELECTOR);
      for (var i = 0; i < els.length; i++) {
        out.push({ key: 's' + els[i].getAttribute('data-index'), el: els[i] });
      }
      return out;
    }

    function scopeOf(node) {
      var list = scopeList();
      var el = node && node.nodeType === 1 ? node : (node ? node.parentNode : null);
      for (var i = 0; i < list.length; i++) {
        if (el && list[i].el.contains(el)) return list[i];
      }
      return null;
    }

    // ── Render ──
    function paintScope(scope) {
      unwrap(scope.el);
      var items = bucket(docKey, scope.key, false);
      if (!items || !items.length) return false;
      var moved = false;
      for (var i = 0; i < items.length; i++) {
        var item = items[i];
        var nodes = textNodes(scope.el);
        var found = locate(joinText(nodes), item);
        if (!found) continue;
        if (found.start !== item.start) { item.start = found.start; moved = true; }
        wrapRange(scope.el, found.start, found.end, item);
      }
      return moved;
    }

    function render() {
      var doc = target();
      if (!doc || !doc.body) return;
      bind(doc);
      styleInto(doc, 'dr-hl-style', MARK_CSS);

      applying = true;
      var moved = false;
      try {
        var list = scopeList();
        for (var i = 0; i < list.length; i++) {
          if (paintScope(list[i])) moved = true;
        }
      } catch (e) {
      } finally {
        applying = false;
        // Our own wrapping queued records; drop them so painting the marks
        // does not read as a change that needs painting again.
        if (observer) observer.takeRecords();
      }
      if (moved) persist();
      placeBar();
      placeCard();
    }

    function schedule() {
      if (queued) return;
      queued = true;
      setTimeout(function () { queued = false; render(); }, 60);
    }

    // ── Binding ──
    function bind(doc) {
      if (bound === doc) return;
      bound = doc;
      if (observer) observer.disconnect();
      if (!readOnly) {
        doc.addEventListener('mouseup', onSelect, true);
        doc.addEventListener('keyup', onSelect, true);
      }
      doc.addEventListener('click', onClickMark, true);
      doc.addEventListener('scroll', reposition, true);
      if (doc.defaultView) doc.defaultView.addEventListener('resize', reposition);
      observer = new MutationObserver(function () { if (!applying) schedule(); });
      observer.observe(doc.documentElement, { childList: true, subtree: true, characterData: true });
    }

    function reposition() { placeBar(); placeCard(); }

    // ── Coordinates ──
    /** Rect in the host document's viewport, undoing any iframe scaling. */
    function hostRect(rect) {
      if (!frame) {
        return { left: rect.left, top: rect.top, right: rect.right, bottom: rect.bottom,
          width: rect.width, height: rect.height };
      }
      var box = frame.getBoundingClientRect();
      var kx = 1, ky = 1;
      try {
        var win = frame.contentWindow;
        if (win && win.innerWidth) kx = box.width / win.innerWidth;
        if (win && win.innerHeight) ky = box.height / win.innerHeight;
      } catch (e) {}
      return {
        left: box.left + rect.left * kx,
        top: box.top + rect.top * ky,
        right: box.left + rect.right * kx,
        bottom: box.top + rect.bottom * ky,
        width: rect.width * kx,
        height: rect.height * ky
      };
    }

    /** Anchors a popup above its rect, flipping and clamping to stay on screen. */
    function place(el, rect) {
      el.style.left = '0px';
      el.style.top = '0px';
      var size = el.getBoundingClientRect();
      var vw = uiDoc.documentElement.clientWidth;
      var vh = uiDoc.documentElement.clientHeight;
      var left = rect.left + rect.width / 2 - size.width / 2;
      var top = rect.top - size.height - 9;
      if (top < 8) top = Math.min(rect.bottom + 9, vh - size.height - 8);
      el.style.left = Math.round(Math.max(8, Math.min(left, vw - size.width - 8))) + 'px';
      el.style.top = Math.round(Math.max(8, top)) + 'px';
    }

    function marksFor(id) {
      var doc = target();
      if (!doc) return [];
      return doc.querySelectorAll('mark[data-dr-hl="' + id + '"]');
    }

    function rectOfId(id) {
      var found = marksFor(id);
      if (!found.length) return null;
      return hostRect(found[0].getBoundingClientRect());
    }

    // ── Selection bar ──
    function onSelect(e) {
      if (e && e.type === 'keyup' && !e.shiftKey && e.key !== 'Shift') return;
      setTimeout(readSelection, 0);
    }

    function readSelection() {
      var doc = target();
      if (!doc) return;
      var sel = doc.getSelection ? doc.getSelection() : null;
      if (!sel || sel.isCollapsed || !sel.rangeCount) { hideBar(); return; }
      var range = sel.getRangeAt(0);
      var text = range.toString();
      if (!text || !text.trim()) { hideBar(); return; }

      var scope = scopeOf(range.commonAncestorContainer);
      if (!scope) { hideBar(); return; }

      var nodes = textNodes(scope.el);
      var start = pointOffset(scope.el, nodes, range.startContainer, range.startOffset);
      var end = pointOffset(scope.el, nodes, range.endContainer, range.endOffset);
      if (start < 0 || end <= start) { hideBar(); return; }

      pending = {
        scope: scope.key,
        start: start,
        text: joinText(nodes).slice(start, end),
        rect: hostRect(range.getBoundingClientRect())
      };
      showBar();
    }

    function showBar() {
      if (!bar) {
        bar = uiDoc.createElement('div');
        bar.className = 'dr-bar';
        bar.setAttribute('data-dr-ui', '1');
        bar.appendChild(button('highlight', function () { create(false); }));
        bar.appendChild(button('highlight + comment', function () { create(true); }));
        bar.addEventListener('mousedown', function (e) { e.preventDefault(); });
        uiDoc.body.appendChild(bar);
      }
      bar.style.display = 'flex';
      placeBar();
    }

    function placeBar() {
      if (!bar || bar.style.display === 'none' || !pending) return;
      place(bar, pending.rect);
    }

    function hideBar() {
      pending = null;
      if (bar) bar.style.display = 'none';
    }

    function button(label, onClick, className) {
      var el = uiDoc.createElement('button');
      el.type = 'button';
      el.textContent = label;
      if (className) el.className = className;
      el.addEventListener('click', function (e) { e.preventDefault(); e.stopPropagation(); onClick(); });
      return el;
    }

    // ── Creating and editing ──
    function create(withComment) {
      if (!pending) return;
      var item = { id: newId(), start: pending.start, text: pending.text, note: '', at: Date.now() };
      var scopeKey = pending.scope;
      bucket(docKey, scopeKey, true).push(item);
      commit();

      var doc = target();
      try { if (doc && doc.getSelection) doc.getSelection().removeAllRanges(); } catch (e) {}
      hideBar();
      render();
      warnOnce();
      if (withComment) openCard(item.id);
    }

    function find(id) {
      var docs = load().docs[docKey];
      if (!docs) return null;
      for (var scopeKey in docs) {
        if (!Object.prototype.hasOwnProperty.call(docs, scopeKey)) continue;
        var list = docs[scopeKey];
        for (var i = 0; i < list.length; i++) {
          if (list[i].id === id) return { item: list[i], scope: scopeKey, list: list, at: i };
        }
      }
      return null;
    }

    function onClickMark(e) {
      var el = e.target;
      while (el && el.nodeType === 1 && !(el.classList && el.classList.contains('dr-hl'))) {
        el = el.parentNode;
      }
      if (!el || el.nodeType !== 1 || !el.classList.contains('dr-hl')) return;
      if (!scopeOf(el)) return;
      e.preventDefault();
      e.stopPropagation();
      hideBar();
      openCard(el.getAttribute('data-dr-hl'));
    }

    function openCard(id) {
      var hit = find(id);
      if (!hit) return;
      closeCard();
      openId = id;

      card = uiDoc.createElement('div');
      card.className = 'dr-note';
      card.setAttribute('data-dr-ui', '1');

      var quote = uiDoc.createElement('div');
      quote.className = 'dr-note__quote';
      quote.textContent = hit.item.text;
      card.appendChild(quote);

      var input = uiDoc.createElement('textarea');
      input.placeholder = 'Add a comment';
      input.value = hit.item.note || '';
      card.appendChild(input);

      var row = uiDoc.createElement('div');
      row.className = 'dr-note__row';
      row.appendChild(button('remove', function () { remove(id); }, 'dr-danger'));
      var spacer = uiDoc.createElement('span');
      spacer.className = 'dr-spacer';
      row.appendChild(spacer);
      row.appendChild(button('close', function () { closeCard(); }));
      row.appendChild(button('save', function () { saveNote(id, input.value); }));
      card.appendChild(row);

      // The deck and the editor both listen for bare keys on the document;
      // typing a comment must not black out the screen or open the palette.
      card.addEventListener('keydown', function (e) {
        e.stopPropagation();
        if (e.key === 'Escape') { e.preventDefault(); closeCard(); }
        else if (e.key === 'Enter' && (e.metaKey || e.ctrlKey)) { e.preventDefault(); saveNote(id, input.value); }
      });

      uiDoc.body.appendChild(card);
      placeCard();
      input.focus();
    }

    function placeCard() {
      if (!card || !openId) return;
      var rect = rectOfId(openId);
      if (!rect) { closeCard(); return; }
      place(card, rect);
    }

    function closeCard() {
      if (card && card.parentNode) card.parentNode.removeChild(card);
      card = null;
      openId = null;
    }

    function saveNote(id, value) {
      var hit = find(id);
      if (!hit) { closeCard(); return; }
      hit.item.note = (value || '').trim();
      commit();
      closeCard();
      render();
    }

    function remove(id) {
      var hit = find(id);
      if (hit) {
        hit.list.splice(hit.at, 1);
        commit();
      }
      closeCard();
      render();
    }

    // ── First-use warning ──
    function warnOnce() {
      var state = load();
      if (state.warned) return;
      state.warned = true;
      persist();
      broadcast();
      if (opts.onWarn) { opts.onWarn(WARNING); return; }

      var box = uiDoc.createElement('div');
      box.className = 'dr-warn';
      box.setAttribute('data-dr-ui', '1');
      var text = uiDoc.createElement('span');
      text.innerHTML = '<b>Nothing is saved.</b> ';
      text.appendChild(uiDoc.createTextNode(WARNING));
      box.appendChild(text);
      var close = uiDoc.createElement('button');
      close.type = 'button';
      close.textContent = '\u00d7';
      close.addEventListener('click', function () { if (box.parentNode) box.parentNode.removeChild(box); });
      box.appendChild(close);
      uiDoc.body.appendChild(box);
      setTimeout(function () { if (box.parentNode) box.parentNode.removeChild(box); }, 11000);
    }

    // ── Wiring ──
    styleInto(uiDoc, 'dr-hl-ui-style', UI_CSS);

    uiDoc.addEventListener('mousedown', function (e) {
      if (bar && bar.contains(e.target)) return;
      if (card && card.contains(e.target)) return;
      if (frame && frame.contains(e.target)) return;
      hideBar();
      closeCard();
    }, true);

    uiDoc.addEventListener('keydown', function (e) {
      if (e.key === 'Escape' && (bar || card)) { hideBar(); closeCard(); }
    });

    listeners.push(schedule);
    // A frame that has not loaded yet has no document to paint into, and a
    // reassigned srcdoc replaces the one that was there.
    if (frame) frame.addEventListener('load', function () { render(); });

    var api = {
      refresh: function () { render(); },
      setDocKey: function (key) {
        if (key === docKey) return;
        docKey = key || 'default';
        hideBar();
        closeCard();
        render();
      },
      docKey: function () { return docKey; }
    };

    render();
    return api;
  }

  window.deckrunHighlights = {
    mount: mount,
    warning: WARNING,
    /** Drops a document's highlights, for when the document itself is gone. */
    forget: function (docKey) {
      var docs = load().docs;
      if (!docs[docKey]) return;
      delete docs[docKey];
      commit();
    }
  };
})();`
