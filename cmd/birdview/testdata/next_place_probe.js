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

// Drives the Place search box for TestPlaceSearchInTheBrowser. The queries
// come in the URL fragment as JSON. It opens Place, waits for /places.json to
// load, types each query, and records the status text, close link, and each
// match link (class, text, href). It also opens the Browse disclosure at 360px
// width and reports whether the popover fits the viewport.
window.addEventListener("load", function () {
  var queries = JSON.parse(new URLSearchParams(location.hash.slice(1)).get("queries") || "[]");
  var ctl = document.querySelector('details.ctl[data-ctl="Place"]');
  var box = document.querySelector('form.placesearch input[name="place"]');
  var out = { typed: [] };
  function done() { document.documentElement.setAttribute("data-probe", JSON.stringify(out)); }
  if (!ctl || !box) { done(); return; }
  ctl.open = true;
  ctl.dispatchEvent(new Event("toggle"));
  box.focus();
  box.dispatchEvent(new Event("focus"));
  setTimeout(function () {
    queries.forEach(function (q) {
      box.value = q;
      box.dispatchEvent(new Event("input", { bubbles: true }));
      var f = document.querySelector(".placefinds .found");
      var x = f && f.querySelector("a.findclose");
      var r = {
        q: q,
        found: f ? f.textContent.replace(x ? x.textContent : "", "") : "",
        close: x ? x.getAttribute("href") : "",
        items: []
      };
      document.querySelectorAll(".placefinds a.btn").forEach(function (a) {
        r.items.push({ cls: a.className, text: a.textContent.trim(), href: a.getAttribute("href") });
      });
      out.typed.push(r);
    });
    // Check collapsed browse disclosure: starts closed, opens inside viewport.
    var browse = ctl.querySelector("details.browse");
    out.browseInitiallyOpen = browse ? browse.open : null;
    if (browse) {
      browse.open = true;
      browse.dispatchEvent(new Event("toggle"));
      var pop = ctl.querySelector(".pop");
      var rect = pop.getBoundingClientRect();
      out.popLeft = rect.left;
      out.popRight = rect.right;
      out.viewport = document.documentElement.clientWidth;
    }
    done();
  }, 1000);
});
