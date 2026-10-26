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

// Drives the search box for TestSearchInTheBrowser. The queries come in the
// URL fragment as JSON. For each, it types the query, records which tiles are
// visible, the count line and the sections left showing; then, with the last
// query still in the box, it clicks a few links with navigation cancelled and
// records where each would have gone. The result is data-probe on the root.
window.addEventListener("load", function () {
  var params = new URLSearchParams(location.hash.slice(1));
  var queries = JSON.parse(params.get("queries") || "[]");
  var box = document.querySelector('form.search input[name="q"]');
  var out = { found: !!box, typed: [], links: {} };
  if (!box) {
    document.documentElement.setAttribute("data-probe", JSON.stringify(out));
    return;
  }
  function visible(el) { return !el.closest("[hidden]"); }
  queries.forEach(function (q) {
    box.value = q;
    box.dispatchEvent(new Event("input", { bubbles: true }));
    var r = { q: q, ids: [], sections: [], bars: {} };
    document.querySelectorAll("[data-search]").forEach(function (t) {
      if (visible(t)) r.ids.push(t.id);
    });
    document.querySelectorAll("section[id^='tier-']").forEach(function (s) {
      if (!s.hidden) r.sections.push(s.id + ":" + s.querySelector("h2 .n").textContent);
    });
    document.querySelectorAll(".hist a[href^='#tier-']").forEach(function (a) {
      r.bars[a.getAttribute("href").slice(1)] = a.style.height;
    });
    r.count = document.querySelector(".count .shown").textContent;
    var f = document.querySelector(".fsearch");
    r.named = f && !f.hidden ? f.querySelector(".fsq").textContent : "";
    r.empty = !document.querySelector(".searchempty").hidden;
    out.typed.push(r);
  });
  // Where each link goes once clicked, without leaving the page.
  document.addEventListener("click", function (e) { e.preventDefault(); });
  function clicked(name, el) {
    if (!el) { out.links[name] = null; return; }
    el.dispatchEvent(new MouseEvent("click", { bubbles: true, cancelable: true }));
    out.links[name] = el.getAttribute("href") || el.getAttribute("data-template");
  }
  clicked("control", document.querySelector(".row.ordering a.btn:not(.on)"));
  clicked("tile", document.querySelector("a.tile"));
  clicked("welcome", document.querySelector(".welcome a"));
  clicked("reset", document.querySelector('.welcome a[href="/classic"]'));
  clicked("here", document.querySelector("button.here[data-template]"));
  clicked("docs", document.querySelector('a[href^="/docs"]'));
  document.documentElement.setAttribute("data-probe", JSON.stringify(out));
});
