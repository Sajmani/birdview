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

// Measures the contrast of every piece of visible text on the page against
// what is drawn behind it, for TestRenderedTextContrast, and records the result
// on the root element as data-contrast.
//
// The background is found by looking, not by inheritance: everything under the
// middle of the text, top to bottom, until an opaque layer, with translucent
// layers composited over it. An HTML element contributes its background color
// and an SVG shape its fill. Where nothing opaque is found, white is assumed,
// as wcag.ContrastRatio says. Relative luminance is WCAG 2.2's.
//
// Every <details> is opened first, and text covered by something else at its
// own middle is counted as unmeasured rather than measured against the cover.
//
// Known limits: a background image or gradient counts as nothing, and text
// whose middle lands on a thin line drawn behind it is measured against the
// line.
window.addEventListener('load', function () {
  var NS = 'http://www.w3.org/1999/xhtml';
  // elementsFromPoint skips anything with pointer-events: none, which is how
  // diagram labels and decorations are often styled.
  var st = document.createElementNS(NS, 'style');
  st.textContent = '* { pointer-events: auto !important; }';
  (document.head || document.documentElement).appendChild(st);
  // A closed panel's contents are part of the page (wcag.FullPages), drawn
  // when the reader opens it, so every panel is measured open.
  document.querySelectorAll('details').forEach(function (d) { d.open = true; });

  function parse(c) {
    var m = /rgba?\(([^)]+)\)/.exec(c || '');
    if (!m) return null;
    var p = m[1].split(/[\s,\/]+/).filter(Boolean).map(parseFloat);
    return { r: p[0], g: p[1], b: p[2], a: p.length > 3 ? p[3] : 1 };
  }
  function lin(v) {
    v /= 255;
    return v <= 0.04045 ? v / 12.92 : Math.pow((v + 0.055) / 1.055, 2.4);
  }
  function lum(c) { return 0.2126 * lin(c.r) + 0.7152 * lin(c.g) + 0.0722 * lin(c.b); }
  function over(top, under) {
    var a = top.a;
    return { r: top.r * a + under.r * (1 - a), g: top.g * a + under.g * (1 - a),
      b: top.b * a + under.b * (1 - a), a: 1 };
  }
  function show(c) { return 'rgb(' + [c.r, c.g, c.b].map(Math.round).join(', ') + ')'; }

  var shapes = { rect: 1, circle: 1, ellipse: 1, polygon: 1, path: 1 };
  function layerOf(el) {
    var cs = getComputedStyle(el), c;
    if (el instanceof SVGElement) {
      if (!shapes[el.localName]) return null;
      c = parse(cs.fill);
      if (!c) return null;
      c.a *= parseFloat(cs.fillOpacity || '1') * parseFloat(cs.opacity || '1');
    } else {
      c = parse(cs.backgroundColor);
      if (!c) return null;
    }
    return c.a > 0 ? c : null;
  }
  function background(x, y) {
    var stack = document.elementsFromPoint(x, y), layers = [];
    for (var i = 0; i < stack.length; i++) {
      var l = layerOf(stack[i]);
      if (l) {
        layers.push(l);
        if (l.a >= 1) break;
      }
    }
    var bg = { r: 255, g: 255, b: 255, a: 1 };
    for (var j = layers.length - 1; j >= 0; j--) bg = over(layers[j], bg);
    return bg;
  }

  var skip = { script: 1, style: 1, title: 1, desc: 1, metadata: 1, noscript: 1 };
  var out = { checked: 0, unmeasured: 0, failures: [] };
  var walker = document.createTreeWalker(document.documentElement, NodeFilter.SHOW_TEXT);
  var n;
  while ((n = walker.nextNode())) {
    var text = n.nodeValue.replace(/\s+/g, ' ').trim();
    var el = n.parentElement;
    if (!text || !el || skip[el.localName]) continue;
    var cs = getComputedStyle(el);
    if (!el.checkVisibility({ opacityProperty: true, visibilityProperty: true })) continue;
    var range = document.createRange();
    range.selectNodeContents(n);
    var r = range.getBoundingClientRect();
    if (r.width < 1 || r.height < 1) continue; // not rendered: display none, a closed details
    window.scrollBy(r.left + r.width / 2 - innerWidth / 2, r.top + r.height / 2 - innerHeight / 2);
    r = range.getBoundingClientRect();
    var x = r.left + r.width / 2, y = r.top + r.height / 2;
    // The text has to be what is drawn at its own middle. Anything else on top
    // -- a sticky bar, a clipping container -- means the reader cannot see it
    // there either, and measuring it against the cover would be wrong.
    var top = x >= 0 && y >= 0 && x < innerWidth && y < innerHeight && document.elementFromPoint(x, y);
    if (!top || !(top === el || el.contains(top) || top.contains(el))) {
      out.unmeasured++;
      continue;
    }
    var svg = el instanceof SVGElement;
    var fg = parse(svg ? cs.fill : cs.color);
    if (!fg) continue;
    if (svg) fg.a *= parseFloat(cs.fillOpacity || '1');
    var bg = background(x, y);
    var ink = over(fg, bg);
    var l1 = lum(ink), l2 = lum(bg);
    var ratio = (Math.max(l1, l2) + 0.05) / (Math.min(l1, l2) + 0.05);
    out.checked++;
    if (ratio < 7) {
      var cls = el.getAttribute('class');
      out.failures.push({ text: text.slice(0, 50), element: el.localName + (cls ? '.' + cls : ''),
        color: show(ink), background: show(bg), ratio: Math.round(ratio * 100) / 100 });
    }
  }
  document.documentElement.setAttribute('data-contrast', JSON.stringify(out));
});
