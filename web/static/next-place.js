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

// The home page's Place search box (product.nextView,
// decisions.placeControlUnifiesRegionsAndCounties), inlined into the page.
//
// As the reader types, it ranks states, provinces, islands, and seas ahead of
// counties and lists the top matches (always showing them even when more match,
// with a note naming the total). Clicking a match toggles it in region= with
// state/county replacement, matching Go's placeMatches and placeToggle.
// TestPlaceSearchInTheBrowser holds this to the server.
(function () {
  var box = document.querySelector('form.placesearch input[name="place"]');
  var out = document.querySelector(".placefinds");
  if (!box || !out) return;
  var limit = parseInt(out.getAttribute("data-limit"), 10) || 12;
  var placesURL = out.getAttribute("data-places") || "/places.json";
  var baseQuery = out.getAttribute("data-base") || "";
  var closeURL = out.getAttribute("data-close") || "/";
  var selected = (out.getAttribute("data-selected") || "").split(" ").filter(Boolean).map(function (s) {
    var i = s.indexOf(":");
    return { slug: i >= 0 ? s.slice(0, i) : s, parent: i >= 0 ? s.slice(i + 1) : "" };
  });
  var selMap = {};
  selected.forEach(function (s) { selMap[s.slug] = true; });
  var places = null, loading = null;

  function normalize(s) {
    return s.toLowerCase().normalize("NFD").replace(/[\u0300-\u036f]/g, "")
      .replace(/[^a-z0-9]+/g, " ").trim();
  }
  function load() {
    if (!loading) {
      loading = fetch(placesURL)
        .then(function (r) { return r.json(); })
        .then(function (p) { places = p; render(); });
    }
    return loading;
  }
  // Builds the exact same sorted query string as Go's url.Values.Encode().
  function toggleURL(slug, parent) {
    var next = [];
    var found = false;
    selected.forEach(function (item) {
      if (item.slug === slug) { found = true; return; }
      next.push(item);
    });
    if (!found) {
      next = next.filter(function (item) {
        if (parent && item.slug === parent) return false;
        if (!parent && item.parent === slug) return false;
        return true;
      });
      next.push({ slug: slug, parent: parent });
    }
    var params = {};
    if (baseQuery) {
      new URLSearchParams(baseQuery).forEach(function (v, k) {
        (params[k] = params[k] || []).push(v);
      });
    }
    if (next.length > 0) {
      params.region = next.map(function (item) { return item.slug; });
    }
    var keys = Object.keys(params).sort();
    if (keys.length === 0) return "/";
    var parts = [];
    keys.forEach(function (k) {
      params[k].forEach(function (v) {
        var sp = new URLSearchParams();
        sp.set(k, v);
        parts.push(sp.toString());
      });
    });
    return "/?" + parts.join("&");
  }
  function rank(e, q, toks) {
    var code = e[3], name = e[4], label = e[5], full = e[6];
    if (name === q || (code && code === q) || label === q) return 0;
    if (name.indexOf(q) === 0 || label.indexOf(q) === 0) return 1;
    var padded = " " + full;
    var allWords = true;
    for (var i = 0; i < toks.length; i++) {
      if (padded.indexOf(" " + toks[i]) < 0) { allWords = false; break; }
    }
    if (allWords) return 2;
    if (full.indexOf(q) >= 0) return 3;
    return -1;
  }
  function el(tag, cls, text) {
    var e = document.createElement(tag);
    if (cls) e.className = cls;
    if (text) e.textContent = text;
    return e;
  }
  function render() {
    var typed = box.value.trim(), q = normalize(typed);
    out.textContent = "";
    if (!q) return;
    if (!places) { load(); return; }
    var toks = q.split(" ");
    var hits = [];
    for (var i = 0; i < places.length; i++) {
      var r = rank(places[i], q, toks);
      if (r >= 0) hits.push({ e: places[i], r: r });
    }
    hits.sort(function (a, b) {
      var ac = a.e[2] ? 1 : 0, bc = b.e[2] ? 1 : 0;
      if (ac !== bc) return ac - bc;
      if (a.r !== b.r) return a.r - b.r;
      return a.e[1] < b.e[1] ? -1 : a.e[1] > b.e[1] ? 1 : 0;
    });
    var n = hits.length;
    var p = el("p", "found");
    var x = el("a", "findclose", "\u00d7");
    x.href = closeURL;
    x.title = "Clear place search";
    x.setAttribute("aria-label", x.title);
    x.addEventListener("click", function (ev) {
      ev.preventDefault();
      box.value = "";
      out.textContent = "";
      box.focus();
    });
    p.appendChild(x);
    if (n === 0) {
      p.appendChild(document.createTextNode("No place matches \u201c" + typed + "\u201d."));
      out.appendChild(p);
      return;
    }
    var msg = n === 1 ? "1 place matches."
      : n > limit ? "Showing " + limit + " of " + n + " matches \u2014 add a state or province to narrow."
      : n + " places match.";
    p.appendChild(document.createTextNode(msg));
    out.appendChild(p);
    var list = el("div", "popgroup");
    var shown = Math.min(n, limit);
    for (var j = 0; j < shown; j++) {
      var item = hits[j].e;
      var on = !!selMap[item[0]];
      var a = el("a", on ? "btn on" : "btn", item[1]);
      a.href = toggleURL(item[0], item[2]);
      a.setAttribute("rel", "nofollow");
      list.appendChild(a);
    }
    out.appendChild(list);
  }
  var ctl = box.closest && box.closest("details.ctl");
  if (ctl) ctl.addEventListener("toggle", function () { if (ctl.open) load(); });
  box.addEventListener("focus", load);
  box.addEventListener("input", render);
})();
