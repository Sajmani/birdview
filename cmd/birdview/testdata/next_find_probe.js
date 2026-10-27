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

// Drives /next's find box for TestNextFindInTheBrowser. The queries come in
// the URL fragment as JSON. It focuses the box, which loads the names, then
// types each query and records the count line and each listed bird's link
// class and address. The result is data-probe on the root.
window.addEventListener("load", function () {
  var queries = JSON.parse(new URLSearchParams(location.hash.slice(1)).get("queries") || "[]");
  var box = document.querySelector('form.search input[name="find"]');
  var out = { typed: [] };
  function done() { document.documentElement.setAttribute("data-probe", JSON.stringify(out)); }
  if (!box) { done(); return; }
  box.focus();
  box.dispatchEvent(new Event("focus"));
  setTimeout(function () {
    queries.forEach(function (q) {
      box.value = q;
      box.dispatchEvent(new Event("input", { bubbles: true }));
      var f = document.querySelector(".finds .found");
      var x = f && f.querySelector("a.findclose");
      var r = { q: q, found: f ? f.textContent.replace(x ? x.textContent : "", "") : "",
        close: x ? x.getAttribute("href") : "", items: [] };
      document.querySelectorAll(".finds li a").forEach(function (a) {
        r.items.push(a.className + " " + a.getAttribute("href"));
      });
      r.texts = [];
      document.querySelectorAll(".finds li").forEach(function (li) {
        r.texts.push(li.textContent.replace(/\s+/g, " ").trim());
      });
      out.typed.push(r);
    });
    // A tile's pin, clicked with text in the box, carries the text.
    box.value = "crow";
    var tp = document.querySelector(".tilewrap > a.pin");
    document.addEventListener("click", function (e) { e.preventDefault(); });
    tp.click();
    out.tilePin = tp.getAttribute("href");
    done();
  }, 1000);
});
