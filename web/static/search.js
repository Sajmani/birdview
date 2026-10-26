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

// The search box (product.searchBox), inlined into the index.
//
// The rule is Go's: the server writes each tile's normalized fields into
// data-search, joined by "|", and this only checks that what is typed,
// normalized as normalizeSearch does, occurs there. It can never contain "|",
// so it cannot match across two fields. TestSearchInTheBrowser
// holds the two together.
//
// It hides and shows tiles, keeps the counts and histogram in step, and puts
// the search into the next same-site link the reader follows, except those in
// the welcome line and the empty state, which return to a fixed view.
(function () {
  var box = document.querySelector('form.search input[name="q"]');
  if (!box) return;

  function normalize(s) {
    return s.toLowerCase().replace(/[^a-z0-9]+/g, " ").trim();
  }
  function matches(text, q) {
    return text.indexOf(q) >= 0;
  }

  var shownEl = document.querySelector(".count .shown");
  var fsum = document.querySelector(".fsearch");
  var fsq = fsum && fsum.querySelector(".fsq");
  var empty = document.querySelector(".searchempty");
  var sections = Array.prototype.slice.call(document.querySelectorAll("section[id^='tier-']"));

  function apply() {
    var q = normalize(box.value);
    var counts = {};
    var total = 0, largest = 1;
    sections.forEach(function (sec) {
      var n = 0;
      sec.querySelectorAll("[data-search]").forEach(function (t) {
        var ok = matches(t.getAttribute("data-search") || "", q);
        t.hidden = !ok;
        if (ok) n++;
      });
      sec.hidden = n === 0;
      var label = sec.querySelector("h2 .n");
      if (label) label.textContent = n + " species";
      counts[sec.id] = n;
      total += n;
      if (n > largest) largest = n;
    });
    document.querySelectorAll(".hist a[href^='#tier-']").forEach(function (a) {
      var n = counts[a.getAttribute("href").slice(1)] || 0;
      a.style.height = Math.round(100 * n / largest) + "%";
      var l = (a.getAttribute("aria-label") || "").replace(/\d+ species$/, n + " species");
      a.setAttribute("aria-label", l);
      a.setAttribute("title", l);
    });
    if (shownEl) shownEl.textContent = total + " of " + shownEl.getAttribute("data-total") + " species";
    var typed = box.value.trim();
    if (fsum) {
      fsum.hidden = !typed;
      fsq.textContent = typed;
    }
    if (empty) empty.hidden = !(total === 0 && sections.length > 0);
  }

  box.addEventListener("input", apply);

  // The search reaches the address on the next click that changes it.
  var carries = /^\/(classic|compare|now|species\/.*)$/;
  function withQuery(href) {
    var u = new URL(href, location.href);
    if (u.origin !== location.origin || !carries.test(u.pathname)) return null;
    var q = box.value.trim();
    if (q) u.searchParams.set("q", q);
    else u.searchParams.delete("q");
    return u.pathname + u.search + u.hash;
  }
  document.addEventListener("click", function (e) {
    if (!e.target.closest) return;
    // Here navigates later, from its template, once the browser has a position.
    var here = e.target.closest("button.here[data-template]");
    if (here) {
      var t = withQuery(here.getAttribute("data-template"));
      if (t) here.setAttribute("data-template", t);
      return;
    }
    var a = e.target.closest("a[href]");
    if (!a || a.closest(".welcome, .empty") || a.getAttribute("href").charAt(0) === "#") return;
    var h = withQuery(a.href);
    if (h) a.href = h;
  }, true);
})();
