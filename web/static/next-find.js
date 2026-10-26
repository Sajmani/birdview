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

// /next's find box (product.nextView), inlined into the page.
//
// As the reader types, it says how many birds match and lists them once
// there are few enough, each with its pin. The rule is Go's, as in search.js: the names file carries each bird's
// normalized search text, and this only checks that what is typed,
// normalized as normalizeSearch does, occurs in it. The pins go to the
// server's bird= toggle, which decides what pinning means.
// TestNextFindInTheBrowser holds this to the Go page.
(function () {
  var box = document.querySelector('form.search input[name="find"]');
  var out = document.querySelector(".finds");
  if (!box || !out) return;
  var limit = parseInt(out.getAttribute("data-limit"), 10) || 8;
  var toggle = out.getAttribute("data-toggle");
  var pinned = {};
  (out.getAttribute("data-pinned") || "").split(" ").forEach(function (id) {
    if (id) pinned[id] = true;
  });
  var icon = document.getElementById("pinicon");
  var birds = null, loading = null;

  function normalize(s) {
    return s.toLowerCase().replace(/[^a-z0-9]+/g, " ").trim();
  }
  function load() {
    if (!loading) {
      loading = fetch(out.getAttribute("data-birds"))
        .then(function (r) { return r.json(); })
        .then(function (b) { birds = b; render(); });
    }
    return loading;
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
    if (!birds) { load(); return; }
    var found = birds.filter(function (b) { return b[3].indexOf(q) >= 0; });
    var n = found.length;
    var msg = n === 0 ? "No bird\u2019s name contains \u201c" + typed + "\u201d."
      : n === 1 ? "1 bird matches." : n + " birds match.";
    if (n > limit) msg += " Type more to list them.";
    out.appendChild(el("p", "found", msg));
    if (n === 0 || n > limit) return;
    var ul = el("ul");
    found.forEach(function (b) {
      var li = el("li");
      var on = !!pinned[b[0]];
      var a = el("a", on ? "pin on" : "pin");
      a.href = toggle + encodeURIComponent(b[0]);
      a.setAttribute("aria-pressed", on ? "true" : "false");
      a.title = (on ? "Unpin " : "Pin ") + b[1];
      a.setAttribute("aria-label", a.title);
      if (icon) a.appendChild(icon.content.cloneNode(true));
      li.appendChild(a);
      li.appendChild(document.createTextNode(" " + b[1] + " "));
      li.appendChild(el("i", "", b[2]));
      ul.appendChild(li);
    });
    out.appendChild(ul);
  }
  box.addEventListener("focus", load);
  box.addEventListener("input", render);
})();
