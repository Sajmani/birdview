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

/* The proportions review tool (tech.proportionsReviewTool), local only.

   The sliders redraw the adjusted figure through /review/figure. The photographs come from
   Wikimedia Commons' public search, shown from Commons with their author and license; nothing
   is stored. The adjusted figure can be laid over a photograph, dragged, sized, turned, and
   flipped, to compare outlines. */
(function () {
  "use strict";
  var form = document.getElementById("form");
  if (!form) return;
  var genus = form.elements.genus.value;
  // One slider for every modeled measurement; the wingspan's only where it is modeled.
  var sliders = ["neck", "head", "body", "wing"].map(function (n) { return form.elements[n]; })
    .filter(Boolean);
  var start = sliders.map(function (el) { return el.value; });
  var pair = document.getElementById("pair");
  var overlay = document.getElementById("overlay");

  function show() {
    sliders.forEach(function (el) {
      form.querySelector('output[data-for="' + el.name + '"]').textContent = Number(el.value).toFixed(2);
    });
  }
  // The overlay is the adjusted figure without its mass cube, which would only clutter the photo.
  function copyOverlay() {
    var svg = pair.querySelector("#adjusted svg");
    overlay.textContent = "";
    if (!svg) return;
    var c = svg.cloneNode(true);
    Array.prototype.forEach.call(c.querySelectorAll(".masscube"), function (g) { g.remove(); });
    overlay.appendChild(c);
  }
  var timer = 0, asked = 0;
  function redraw() {
    show();
    clearTimeout(timer);
    timer = setTimeout(function () {
      var n = ++asked;
      var q = sliders.map(function (el) { return "&" + el.name + "=" + el.value; }).join("");
      fetch("/review/figure?genus=" + encodeURIComponent(genus) + q)
        .then(function (r) { return r.ok ? r.text() : Promise.reject(r.status); })
        .then(function (html) { if (n === asked) { pair.innerHTML = html; copyOverlay(); } });
    }, 80);
  }
  sliders.forEach(function (el) { el.addEventListener("input", redraw); });
  Array.prototype.forEach.call(document.querySelectorAll("[data-nudge]"), function (b) {
    b.addEventListener("click", function () {
      var el = form.elements[b.dataset.nudge];
      var v = Math.min(Number(el.max), Math.max(Number(el.min), Number(el.value) + Number(b.dataset.by)));
      el.value = v.toFixed(2);
      redraw();
    });
  });
  document.getElementById("reset").addEventListener("click", function () {
    sliders.forEach(function (el, i) { el.value = start[i]; });
    redraw();
  });
  // Keys: [ and ] for the neck, - and = for the head, ; and ' for the body.
  document.addEventListener("keydown", function (e) {
    if (e.target.tagName === "INPUT" && e.target.type === "text") return;
    var keys = { "[": ["neck", -0.01], "]": ["neck", 0.01], "-": ["head", -0.01], "=": ["head", 0.01],
      ";": ["body", -0.01], "'": ["body", 0.01] };
    var k = keys[e.key];
    if (!k) return;
    var el = form.elements[k[0]];
    el.value = (Number(el.value) + k[1]).toFixed(2);
    redraw();
    e.preventDefault();
  });

  // Photographs from Commons.
  var panel = document.getElementById("photos");
  var sci = panel.dataset.sci;
  var thumbs = document.getElementById("thumbs");
  var status = document.getElementById("status");
  var photo = document.getElementById("photo");
  var credit = document.getElementById("credit");
  function text(html) {
    return new DOMParser().parseFromString(html || "", "text/html").body.textContent.trim();
  }
  function search(q) {
    status.textContent = "Searching Commons for " + q + "…";
    thumbs.textContent = "";
    var url = "https://commons.wikimedia.org/w/api.php?action=query&format=json&origin=*" +
      "&generator=search&gsrnamespace=6&gsrlimit=24&gsrsearch=" + encodeURIComponent(q + " filetype:bitmap") +
      "&prop=imageinfo&iiprop=url|extmetadata|mime&iiurlwidth=900" +
      "&iiextmetadatafilter=Artist|LicenseShortName|LicenseUrl";
    fetch(url).then(function (r) { return r.json(); }).then(function (j) {
      var pages = j.query ? Object.values(j.query.pages) : [];
      pages.sort(function (a, b) { return a.index - b.index; });
      var shown = 0;
      pages.forEach(function (p) {
        var ii = p.imageinfo && p.imageinfo[0];
        if (!ii || !/^image\/(jpeg|png)$/.test(ii.mime)) return;
        var m = ii.extmetadata || {};
        var item = {
          src: ii.thumburl || ii.url, page: ii.descriptionurl, title: p.title.replace(/^File:/, ""),
          author: text(m.Artist && m.Artist.value), license: text(m.LicenseShortName && m.LicenseShortName.value),
          licenseUrl: m.LicenseUrl && m.LicenseUrl.value
        };
        var b = document.createElement("button");
        b.type = "button";
        b.title = item.title;
        var img = document.createElement("img");
        img.src = item.src; img.alt = item.title; img.loading = "lazy";
        b.appendChild(img);
        b.addEventListener("click", function () { choose(item, b); });
        thumbs.appendChild(b);
        if (shown++ === 0) choose(item, b);
      });
      status.textContent = shown ? shown + " photographs for " + q + " (pick one)" : "No photographs for " + q + ".";
    }).catch(function () { status.textContent = "Commons did not answer."; });
  }
  function choose(item, button) {
    Array.prototype.forEach.call(thumbs.children, function (c) { c.classList.toggle("on", c === button); });
    photo.src = item.src;
    photo.alt = item.title;
    credit.textContent = "";
    var a = document.createElement("a");
    a.href = item.page; a.target = "_blank"; a.rel = "noopener noreferrer"; a.textContent = item.title;
    credit.appendChild(a);
    credit.appendChild(document.createTextNode(" — " + (item.author || "author not given") + ", "));
    if (item.licenseUrl) {
      var l = document.createElement("a");
      l.href = item.licenseUrl; l.target = "_blank"; l.rel = "noopener noreferrer"; l.textContent = item.license || "license";
      credit.appendChild(l);
    } else {
      credit.appendChild(document.createTextNode(item.license || "license not given"));
    }
    credit.appendChild(document.createTextNode(", via Wikimedia Commons"));
    copyOverlay();
    overlay.hidden = !document.getElementById("ov").checked;
  }
  search('"' + sci + '" flying');
  document.getElementById("more").addEventListener("click", function () { search('"' + sci + '"'); });

  // The overlay, turned to lie along a bird flying across the frame, as most flight photographs show it.
  var o = { x: 20, y: 20, s: 1, r: 90, f: 1, op: 0.7 };
  function place() {
    overlay.style.transform = "translate(" + o.x + "px," + o.y + "px) rotate(" + o.r + "deg) scale(" + (o.s * o.f) + "," + o.s + ")";
    overlay.style.opacity = o.op;
  }
  document.getElementById("ov").addEventListener("change", function (e) { overlay.hidden = !e.target.checked; });
  document.getElementById("op").addEventListener("input", function (e) { o.op = Number(e.target.value); place(); });
  document.getElementById("sz").addEventListener("input", function (e) { o.s = Number(e.target.value); place(); });
  // Any angle: the slider, +90 in steps, or shift and the scroll wheel over the photograph.
  var rotate = document.getElementById("rotate");
  function turn(deg) {
    o.r = ((Math.round(deg) + 540) % 360) - 180;
    rotate.value = o.r;
    document.getElementById("rotout").textContent = o.r + "°";
    place();
  }
  rotate.addEventListener("input", function () { turn(Number(rotate.value)); });
  document.getElementById("rot").addEventListener("click", function () { turn(o.r + 90); });
  document.getElementById("flip").addEventListener("click", function () { o.f = -o.f; place(); });
  var drag = null;
  overlay.addEventListener("pointerdown", function (e) {
    drag = { x: e.clientX - o.x, y: e.clientY - o.y };
    overlay.setPointerCapture(e.pointerId);
  });
  overlay.addEventListener("pointermove", function (e) {
    if (!drag) return;
    o.x = e.clientX - drag.x; o.y = e.clientY - drag.y; place();
  });
  overlay.addEventListener("pointerup", function () { drag = null; });
  // Scroll over the photo to size the figure.
  document.getElementById("stage").addEventListener("wheel", function (e) {
    if (overlay.hidden) return;
    e.preventDefault();
    if (e.shiftKey) { turn(o.r + (e.deltaY < 0 ? 2 : -2)); return; }
    o.s = Math.min(4, Math.max(0.2, o.s * (e.deltaY < 0 ? 1.05 : 0.95)));
    document.getElementById("sz").value = o.s;
    place();
  }, { passive: false });
  place();
  show();
})();
