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

// Drives TestNextSettingStaysOpen. In the top page it opens Month and clicks
// May with the navigation cancelled, so the page's own click handler runs
// and nothing leaves. It then loads May's address in a frame, which shares
// the session's storage, records which settings open there, loads the same
// address in a second frame, as a reload after no choice, and records again.
// The result is data-probe on the root. In a frame it does nothing.
window.addEventListener("load", function () {
  if (window !== window.top) return;
  function opened(doc) {
    return Array.prototype.map.call(doc.querySelectorAll("details.ctl[open]"),
      function (d) { return d.getAttribute("data-ctl"); });
  }
  function frame(src, then) {
    var f = document.createElement("iframe");
    f.addEventListener("load", function () {
      setTimeout(function () { then(f.contentDocument, f.contentWindow); }, 100);
    });
    f.src = src;
    document.body.appendChild(f);
  }
  var out = {};
  setTimeout(function () {
    var d = document.querySelector('details[data-ctl="Month"]');
    d.querySelector("summary").click();
    var a = d.querySelector('.pop a[title="May"]');
    document.addEventListener("click", function (e) { e.preventDefault(); });
    out.chosen = a.textContent;
    a.click();
    var href = a.getAttribute("href");
    frame(href, function (doc, win) {
      out.url = win.location.pathname + win.location.search;
      out.open = opened(doc);
      frame(href, function (doc2) {
        out.reload = opened(doc2);
        document.documentElement.setAttribute("data-probe", JSON.stringify(out));
      });
    });
  }, 100);
});
