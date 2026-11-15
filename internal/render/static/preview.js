(function () {
  'use strict';

  var VW = 1600;
  var HLJS = __HLJS__;
  var DECOR = __DECOR__;
  var FONTS = __FONTS__;

  var stage = document.getElementById('presentation');
  var slides = [];
  var mode = 'single';
  var index = 0;

  function send(msg) {
    if (window.parent !== window) window.parent.postMessage(msg, '*');
  }

  var syncFlashTimer = null;
  function flashPreviewElement(el) {
    if (!el) return;
    if (syncFlashTimer) { clearTimeout(syncFlashTimer); syncFlashTimer = null; }
    var prev = stage.querySelectorAll('.dr-sync-flash');
    for (var i = 0; i < prev.length; i++) prev[i].classList.remove('dr-sync-flash');
    void el.offsetWidth;
    el.classList.add('dr-sync-flash');
    syncFlashTimer = setTimeout(function () {
      el.classList.remove('dr-sync-flash');
      syncFlashTimer = null;
    }, 1400);
  }

  function normSyncText(s) {
    return (s || '').replace(/\s+/g, ' ').trim();
  }

  function countBefore(haystack, needle) {
    if (!haystack || !needle) return 0;
    var h = haystack.toLowerCase();
    var n = needle.toLowerCase();
    var count = 0;
    var pos = 0;
    while ((pos = h.indexOf(n, pos)) !== -1) {
      count++;
      pos += Math.max(1, n.length);
    }
    return count;
  }

  function flashMatchingBlock(lineText, word) {
    var slideEl = stage.querySelector('.slide');
    if (!slideEl) return;
    var qLine = normSyncText(lineText).toLowerCase();
    var qWord = normSyncText(word).toLowerCase();
    var blocks = slideEl.querySelectorAll('h1, h2, h3, h4, h5, h6, li, p, td, th, blockquote, pre, img');
    var best = null;
    var bestScore = 0;
    for (var i = 0; i < blocks.length; i++) {
      var b = blocks[i];
      if (b.tagName === 'IMG') {
        var bSrc = (b.getAttribute('src') || '').toLowerCase();
        var lm = bSrc.match(/[?&]path=([^&]+)/);
        if (lm) {
          try { bSrc = decodeURIComponent(lm[1].replace(/\+/g, ' ')).toLowerCase(); } catch (err) {}
        }
        var fname = bSrc.split('/').pop().replace(/\.[^.]+$/, '');
        var bAlt = (b.getAttribute('alt') || '').toLowerCase();
        var imgScore = 0;
        if (fname && qLine.indexOf(fname) !== -1) imgScore = 100 + fname.length;
        else if (bAlt && qLine.indexOf(bAlt) !== -1) imgScore = 60 + bAlt.length;
        else if (qWord && (bSrc.indexOf(qWord) !== -1 || bAlt.indexOf(qWord) !== -1)) imgScore = 40;
        if (imgScore > bestScore) { bestScore = imgScore; best = b; }
        continue;
      }
      var txt = normSyncText(b.innerText || b.textContent).toLowerCase();
      if (!txt) continue;
      var score = 0;
      if (qLine && txt === qLine) score = 200;
      else if (qLine && qLine.length >= 4 && txt.indexOf(qLine) !== -1) score = 120 + qLine.length;
      else if (qLine && txt.length >= 4 && qLine.indexOf(txt) !== -1) score = 100 + txt.length;
      else if (qWord && txt.indexOf(qWord) !== -1) {
        score = 30;
        if (qLine) {
          var words = qLine.split(/\s+/);
          for (var w = 0; w < words.length; w++) {
            if (words[w].length >= 3 && txt.indexOf(words[w]) !== -1) score += 8;
          }
        }
      }
      // Prefer leaf-most block (e.g., li/p/pre over outer blockquote/wrapper) on tie.
      if (score > bestScore || (score === bestScore && score > 0 && best && best.contains(b))) {
        bestScore = score;
        best = b;
      }
    }
    if (best) flashPreviewElement(best);
  }

  function applyFont(slot, id) {
    if (FONTS.indexOf(id) !== -1) document.documentElement.dataset[slot] = id;
    else delete document.documentElement.dataset[slot];
  }

  function highlight(root) {
    if (!window.hljs) return;
    var blocks = root.querySelectorAll('pre code:not(.language-mermaid):not(.lang-mermaid)');
    for (var i = 0; i < blocks.length; i++) {
      if (!blocks[i].dataset.highlighted) {
        try { window.hljs.highlightElement(blocks[i]); } catch (e) {}
      }
    }
  }

  /** Report whether the visible slide clips its own content. */
  function reportOverflow() {
    if (mode !== 'single') return;
    var content = stage.querySelector('.slide__content');
    var over = false;
    if (content) {
      over = content.scrollHeight - content.clientHeight > 6 ||
        content.scrollWidth - content.clientWidth > 6;

      // KaTeX display equations and Mermaid hosts deliberately hide their own
      // overflow to keep a projected slide tidy. Inspect them separately so a
      // clipped formula or diagram still triggers the editor's overflow nudge.
      var rich = content.querySelectorAll('.katex-display, .mermaid-container');
      for (var i = 0; !over && i < rich.length; i++) {
        over = rich[i].scrollHeight - rich[i].clientHeight > 6 ||
          rich[i].scrollWidth - rich[i].clientWidth > 6;
      }
    }
    send({ type: 'overflow', index: index, overflow: over });
  }

  function renderSingle() {
    document.body.classList.remove('is-grid');
    stage.innerHTML = slides[index] || '';
    var el = stage.querySelector('.slide');
    if (el) el.classList.add('is-active');
    if (window.deckrunPrepareFragments) window.deckrunPrepareFragments(stage, true);
    highlight(stage);
    var rich = window.deckrunRenderRichContent ? window.deckrunRenderRichContent(stage) : Promise.resolve();
    rich.then(function () { requestAnimationFrame(reportOverflow); });
  }

  function renderGrid() {
    document.body.classList.add('is-grid');
    stage.innerHTML = '';
    var frag = document.createDocumentFragment();
    slides.forEach(function (html, i) {
      var thumb = document.createElement('div');
      thumb.className = 'pv-thumb' + (i === index ? ' is-current' : '');
      thumb.dataset.index = String(i);

      var inner = document.createElement('div');
      inner.className = 'pv-thumb__inner';
      inner.innerHTML = html;
      var el = inner.querySelector('.slide');
      if (el) el.classList.add('is-active');

      var num = document.createElement('span');
      num.className = 'pv-thumb__num';
      num.textContent = String(i + 1);

      thumb.appendChild(inner);
      thumb.appendChild(num);
      frag.appendChild(thumb);
    });
    stage.appendChild(frag);
    scaleThumbs();
    if (window.deckrunPrepareFragments) window.deckrunPrepareFragments(stage, true);
    highlight(stage);
    var rich = window.deckrunRenderRichContent ? window.deckrunRenderRichContent(stage) : Promise.resolve();
    rich.then(scaleThumbs);
  }

  function scaleThumbs() {
    var thumbs = stage.querySelectorAll('.pv-thumb');
    for (var i = 0; i < thumbs.length; i++) {
      var inner = thumbs[i].querySelector('.pv-thumb__inner');
      if (inner) inner.style.transform = 'scale(' + (thumbs[i].clientWidth / VW) + ')';
    }
  }

  function render() {
    document.body.classList.toggle('is-empty', slides.length === 0);
    if (slides.length === 0) { stage.innerHTML = ''; return; }
    if (index >= slides.length) index = slides.length - 1;
    if (index < 0) index = 0;
    if (mode === 'grid') renderGrid(); else renderSingle();
  }

  stage.addEventListener('click', function (e) {
    if (mode !== 'grid') return;
    var thumb = e.target.closest ? e.target.closest('.pv-thumb') : null;
    if (thumb) send({ type: 'goto', index: parseInt(thumb.dataset.index, 10) });
  });

  stage.addEventListener('dblclick', function (e) {
    var slideIdx = index;
    if (mode === 'grid') {
      var thumb = e.target.closest ? e.target.closest('.pv-thumb') : null;
      if (!thumb) return;
      slideIdx = parseInt(thumb.dataset.index, 10);
      var heading = thumb.querySelector('h1, h2, h3, p, li');
      send({
        type: 'sync-source',
        index: slideIdx,
        word: '',
        blockText: normSyncText(heading ? (heading.innerText || heading.textContent) : ''),
        contextBefore: '',
        contextAfter: '',
        tag: heading ? heading.tagName.toLowerCase() : '',
        occurrence: 0
      });
      return;
    }

    var slideEl = (e.target.closest && e.target.closest('.slide')) || stage.querySelector('.slide');
    if (!slideEl) return;

    var blockEl = (e.target.closest && e.target.closest('h1, h2, h3, h4, h5, h6, li, p, td, th, blockquote, pre, .katex-display, .katex, img')) || slideEl;
    if (blockEl.tagName === 'IMG') {
      var imgSrc = blockEl.getAttribute('src') || '';
      var lm = imgSrc.match(/[?&]path=([^&]+)/);
      if (lm) {
        try { imgSrc = decodeURIComponent(lm[1].replace(/\+/g, ' ')); } catch (err) {}
      }
      var imgAlt = blockEl.getAttribute('alt') || '';
      flashPreviewElement(blockEl);
      send({
        type: 'sync-source',
        index: slideIdx,
        tag: 'img',
        imgSrc: imgSrc,
        imgAlt: imgAlt
      });
      return;
    }

    if (blockEl.classList && (blockEl.classList.contains('katex') || blockEl.classList.contains('katex-display'))) {
      var ann = blockEl.querySelector('annotation[encoding="application/x-tex"]');
      if (ann && ann.textContent.trim()) {
        try { var kSel = window.getSelection(); if (kSel) kSel.removeAllRanges(); } catch (err) {}
        flashPreviewElement(blockEl);
        send({
          type: 'sync-source',
          index: slideIdx,
          word: ann.textContent.trim(),
          blockText: ann.textContent.trim(),
          contextBefore: '',
          contextAfter: '',
          tag: 'katex',
          occurrence: 0
        });
        return;
      }
    }

    var sel = window.getSelection ? window.getSelection() : null;
    var word = '';
    var contextBefore = '';
    var contextAfter = '';
    var occurrence = 0;

    if (sel && sel.rangeCount > 0) {
      var range = sel.getRangeAt(0);
      word = normSyncText(sel.toString());
      try {
        if (blockEl.contains(range.startContainer)) {
          var rBefore = document.createRange();
          rBefore.selectNodeContents(blockEl);
          rBefore.setEnd(range.startContainer, range.startOffset);
          var beforeStr = normSyncText(rBefore.toString());
          contextBefore = beforeStr.slice(Math.max(0, beforeStr.length - 36)).trim();
        }
        if (blockEl.contains(range.endContainer)) {
          var rAfter = document.createRange();
          rAfter.selectNodeContents(blockEl);
          rAfter.setStart(range.endContainer, range.endOffset);
          var afterStr = normSyncText(rAfter.toString());
          contextAfter = afterStr.slice(0, 36).trim();
        }
        if (word && slideEl.contains(range.startContainer)) {
          var rSlideBefore = document.createRange();
          rSlideBefore.selectNodeContents(slideEl);
          rSlideBefore.setEnd(range.startContainer, range.startOffset);
          occurrence = countBefore(normSyncText(rSlideBefore.toString()), word);
        }
      } catch (err) {}
    }

    var blockText = normSyncText(blockEl.innerText || blockEl.textContent);
    try { if (sel) sel.removeAllRanges(); } catch (err) {}
    flashPreviewElement(blockEl);
    send({
      type: 'sync-source',
      index: slideIdx,
      word: word,
      blockText: blockText,
      contextBefore: contextBefore,
      contextAfter: contextAfter,
      tag: (blockEl.tagName || '').toLowerCase(),
      occurrence: occurrence
    });
  });

  var wheelLock = false;
  window.addEventListener('wheel', function (e) {
    if (mode === 'grid') return;
    var delta = Math.abs(e.deltaY) >= Math.abs(e.deltaX) ? e.deltaY : e.deltaX;
    if (Math.abs(delta) < 15) return;
    if (wheelLock) return;
    wheelLock = true;
    setTimeout(function () { wheelLock = false; }, 250);
    if (delta > 0) {
      send({ type: 'nav', delta: 1 });
    } else if (delta < 0) {
      send({ type: 'nav', delta: -1 });
    }
  }, { passive: true });

  window.addEventListener('keydown', function (e) {
    // Option + Control (or Option + Cmd) + arrows/page keys hops between slides
    if (e.altKey && (e.ctrlKey || e.metaKey)) {
      if (e.key === 'ArrowDown' || e.key === 'ArrowRight' || e.key === 'PageDown') {
        e.preventDefault();
        send({ type: 'nav', delta: 1 });
        return;
      } else if (e.key === 'ArrowUp' || e.key === 'ArrowLeft' || e.key === 'PageUp') {
        e.preventDefault();
        send({ type: 'nav', delta: -1 });
        return;
      }
    }

    var mod = e.metaKey || e.ctrlKey;
    if (mod) {
      var k = e.key.toLowerCase();
      if (k === 'enter') { e.preventDefault(); send({ type: 'action', action: 'present' }); }
      else if (k === 'k' && !e.shiftKey) { e.preventDefault(); send({ type: 'action', action: 'palette' }); }
      else if (k === 'g') { e.preventDefault(); send({ type: 'action', action: 'grid' }); }
      else if (k === 'o') { e.preventDefault(); send({ type: 'action', action: 'decks' }); }
      else if (k === '/') { e.preventDefault(); send({ type: 'action', action: 'guide' }); }
      else if (k === 'p' || (k === 's' && e.shiftKey)) { e.preventDefault(); send({ type: 'action', action: 'pdf' }); }
      else if (k === 's') { e.preventDefault(); send({ type: 'action', action: 'download' }); }
      else if (k === 'l' && e.shiftKey) { e.preventDefault(); send({ type: 'action', action: 'theme' }); }
      return;
    }

    if (mode === 'grid') {
      var thumbs = stage.querySelectorAll('.pv-thumb');
      if (!thumbs.length) return;
      var cols = 1;
      if (thumbs.length > 1) {
        var firstTop = thumbs[0].offsetTop;
        for (var c = 1; c < thumbs.length; c++) {
          if (thumbs[c].offsetTop !== firstTop) { cols = c; break; }
        }
        if (cols === 1 && thumbs.length > 1 && thumbs[1].offsetTop === firstTop) {
          cols = thumbs.length;
        }
      }
      var step = 0;
      if (e.key === 'ArrowRight') step = 1;
      else if (e.key === 'ArrowLeft') step = -1;
      else if (e.key === 'ArrowDown') step = cols;
      else if (e.key === 'ArrowUp') step = -cols;
      else if (e.key === 'Home') { e.preventDefault(); send({ type: 'goto', index: 0 }); return; }
      else if (e.key === 'End') { e.preventDefault(); send({ type: 'goto', index: slides.length - 1 }); return; }
      else if (e.key === 'Enter') { e.preventDefault(); send({ type: 'goto', index: index }); return; }
      else if (e.key === 'Escape') { e.preventDefault(); send({ type: 'goto', index: index }); return; }
      else return;

      e.preventDefault();
      var next = Math.max(0, Math.min(slides.length - 1, index + step));
      index = next;
      var cur = stage.querySelector('.pv-thumb.is-current');
      if (cur) cur.classList.remove('is-current');
      var nextThumb = stage.querySelector('.pv-thumb[data-index="' + index + '"]');
      if (nextThumb) {
        nextThumb.classList.add('is-current');
        nextThumb.scrollIntoView({ block: 'nearest' });
      }
      send({ type: 'index-select', index: index });
      return;
    }

    if (e.key === 'ArrowRight' || e.key === 'ArrowDown' || e.key === ' ' || e.key === 'PageDown') {
      e.preventDefault();
      send({ type: 'nav', delta: 1 });
    } else if (e.key === 'ArrowLeft' || e.key === 'ArrowUp' || e.key === 'PageUp') {
      e.preventDefault();
      send({ type: 'nav', delta: -1 });
    } else if (e.key === 'Home') {
      e.preventDefault();
      send({ type: 'goto', index: 0 });
    } else if (e.key === 'End') {
      e.preventDefault();
      send({ type: 'goto', index: slides.length - 1 });
    }
  });

  window.addEventListener('resize', function () {
    if (mode === 'grid') scaleThumbs();
  });

  window.addEventListener('message', function (e) {
    var m = e.data || {};
    if (m.type === 'render') {
      var sameSet = m.slides && slides.length === m.slides.length &&
        m.slides.every(function (h, i) { return h === slides[i]; });
      slides = m.slides || [];
      var modeChanged = m.mode !== mode;
      var indexChanged = m.index !== index;
      mode = m.mode || 'single';
      index = typeof m.index === 'number' ? m.index : 0;
      if (sameSet && !modeChanged && !indexChanged) { reportOverflow(); return; }
      render();
    } else if (m.type === 'theme') {
      if (HLJS[m.theme]) {
        document.documentElement.dataset.theme = m.theme;
        document.documentElement.dataset.decor = DECOR[m.theme];
        var link = document.getElementById('hljs-theme');
        if (link) link.href = HLJS[m.theme];
      }
      // An empty string clears the override and hands the slot back to the
      // theme, which delete does and an assignment of '' would not.
      applyFont('head', m.head);
      applyFont('body', m.body);
      if (m.template) document.documentElement.dataset.template = m.template;
      if (m.transition) document.documentElement.dataset.transition = m.transition;
      // A different face changes how tall a slide's content runs, so the
      // editor's overflow warning has to be re-measured against it.
      render();
    } else if (m.type === 'index') {
      index = m.index;
      if (mode === 'grid') {
        var cur = stage.querySelector('.pv-thumb.is-current');
        if (cur) cur.classList.remove('is-current');
        var next = stage.querySelector('.pv-thumb[data-index="' + index + '"]');
        if (next) {
          next.classList.add('is-current');
          next.scrollIntoView({ block: 'nearest', behavior: 'smooth' });
        }
      } else {
        render();
      }
    } else if (m.type === 'flash-preview') {
      if (typeof m.index === 'number' && m.index !== index) {
        index = m.index;
        mode = 'single';
        render();
      }
      requestAnimationFrame(function () {
        flashMatchingBlock(m.lineText || '', m.word || '');
      });
    }
  });

  send({ type: 'ready' });
})();