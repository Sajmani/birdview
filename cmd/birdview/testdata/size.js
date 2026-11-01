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


// Drives TestRenderedTextSize (tech.textSizeFloor). Opens every <details>, then measures every
// visible run of text at the size it is drawn on screen: its font size,
// times the scale of the SVG it sits in, if any. Records the page's own font
// size, how many runs it measured, and every run below the floor, as
// data-size on the root.
window.addEventListener("load", function () {
  setTimeout(function () {
    var floor = 13, out = { width: window.innerWidth, body: 0, checked: 0, failures: [] };
    document.querySelectorAll("details").forEach(function (d) { d.open = true; });
    out.body = parseFloat(getComputedStyle(document.body).fontSize);
    var walker = document.createTreeWalker(document.body, NodeFilter.SHOW_TEXT);
    while (walker.nextNode()) {
      var n = walker.currentNode, el = n.parentElement;
      if (!n.textContent.trim() || !el || el.closest("script,style,template,title")) continue;
      var r = el.getBoundingClientRect();
      if (r.width === 0 || r.height === 0) continue;
      var px = parseFloat(getComputedStyle(el).fontSize);
      if (el.closest("svg") && el.getScreenCTM) {
        var m = el.getScreenCTM();
        px *= Math.sqrt(m.a * m.a + m.b * m.b);
      }
      out.checked++;
      if (px < floor - 0.05 && out.failures.length < 40) {
        out.failures.push({ text: n.textContent.trim().slice(0, 40), element: el.tagName.toLowerCase() +
          (el.getAttribute("class") ? "." + el.getAttribute("class").split(" ").join(".") : ""), px: Math.round(px * 10) / 10 });
      }
    }
    document.documentElement.setAttribute("data-size", JSON.stringify(out));
  }, 300);
});
