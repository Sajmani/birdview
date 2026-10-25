// Copyright 2026 Google LLC
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     https://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

// Measures whether the page fits the window it is drawn in, for
// TestPagesFitAPhone, and records the result on the root element as
// data-contrast (the name the harness reads).
//
// Two failures are reported:
//
//   - wide: an element that reaches past the right edge of the window, so the
//     page scrolls sideways. Only the outermost such element is reported; its
//     descendants are past the edge because it is.
//   - bleed: text drawn outside the box it sits in, where a box is the nearest
//     ancestor with a background or a border. Text inside an element that clips
//     or scrolls (overflow other than visible) is not bleeding: it is cut off
//     or scrolled, deliberately.
//
// Every <details> is opened first: a closed panel's contents are part of the
// page, drawn when the reader opens it.
window.addEventListener('load', function () {
  document.querySelectorAll('details').forEach(function (d) { d.open = true; });
  var W = document.documentElement.clientWidth;
  var out = { width: W, pageWidth: document.documentElement.scrollWidth, checked: 0, failures: [] };

  function name(el) {
    var cls = el.getAttribute('class');
    return el.localName + (cls ? '.' + cls.trim().split(/\s+/).join('.') : '');
  }
  function textOf(el) { return (el.textContent || '').replace(/\s+/g, ' ').trim().slice(0, 50); }

  var seen = {};
  // wide: outermost elements past the right edge, one of each kind.
  var all = document.body ? document.body.querySelectorAll('*') : [];
  for (var i = 0; i < all.length; i++) {
    var el = all[i];
    if (el.closest('svg') && el.localName !== 'svg') continue;
    var r = el.getBoundingClientRect();
    if (r.width < 1 || r.right <= W + 1) continue;
    var p = el.parentElement, clipped = false, parentWide = false;
    for (var a = p; a && a !== document.body; a = a.parentElement) {
      if (getComputedStyle(a).overflowX !== 'visible') { clipped = true; break; }
    }
    if (clipped) continue;
    if (p && p !== document.body && p.getBoundingClientRect().right > W + 1) parentWide = true;
    if (parentWide) continue;
    if (seen["wide|" + name(el)]) continue;
    seen["wide|" + name(el)] = 1;
    out.failures.push({ kind: 'wide', element: name(el), text: textOf(el),
      right: Math.round(r.right), width: W });
  }

  // bleed: text outside its box.
  function isBox(el) {
    var cs = getComputedStyle(el);
    var bg = /rgba?\(([^)]+)\)/.exec(cs.backgroundColor);
    if (bg) {
      var parts = bg[1].split(/[\s,\/]+/).filter(Boolean);
      if (parts.length < 4 || parseFloat(parts[3]) > 0) return true;
    }
    return parseFloat(cs.borderRightWidth) > 0 && cs.borderRightStyle !== 'none' &&
      parseFloat(cs.borderLeftWidth) > 0 && cs.borderLeftStyle !== 'none';
  }
  var skip = { script: 1, style: 1, title: 1, desc: 1, metadata: 1, noscript: 1 };
  var walker = document.createTreeWalker(document.body || document.documentElement, NodeFilter.SHOW_TEXT);
  var n;
  while ((n = walker.nextNode())) {
    var text = n.nodeValue.replace(/\s+/g, ' ').trim();
    var t = n.parentElement;
    if (!text || !t || skip[t.localName]) continue;
    if (!t.checkVisibility({ opacityProperty: true, visibilityProperty: true })) continue;
    var range = document.createRange();
    range.selectNodeContents(n);
    var tr = range.getBoundingClientRect();
    if (tr.width < 1 || tr.height < 1) continue;
    var box = null, cut = false;
    for (var b = t; b && b !== document.documentElement; b = b.parentElement) {
      if (getComputedStyle(b).overflowX !== 'visible') { cut = true; break; }
      if (isBox(b)) { box = b; break; }
    }
    if (cut || !box || box === document.body) continue;
    out.checked++;
    var br = box.getBoundingClientRect();
    var cs = getComputedStyle(box);
    var left = br.left + parseFloat(cs.borderLeftWidth), right = br.right - parseFloat(cs.borderRightWidth);
    if (tr.right > right + 1 || tr.left < left - 1) {
      var key = name(box) + '|' + name(t);
      if (seen[key]) continue;
      seen[key] = 1;
      out.failures.push({ kind: 'bleed', element: name(t), box: name(box), text: text.slice(0, 50),
        right: Math.round(tr.right), boxRight: Math.round(right) });
    }
  }
  document.documentElement.setAttribute('data-contrast', JSON.stringify(out));
});
