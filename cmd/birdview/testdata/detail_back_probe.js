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

// Drives TestSpeciesBackUsesHistory. On a species page it stands in for
// history.back, clicks Back, and records whether the page went back by
// history or followed the link (which it then cancels). Arrived at from the
// home page, it runs in a frame the home page opened on the first tile's
// species page, so the species page has a same-site referrer, and posts its
// result to the home page; loaded directly, it has none, and records its
// own result. The result is data-probe on the top page's root.
window.addEventListener("load", function () {
  var root = document.documentElement;
  if (location.pathname === "/" && window === window.top) {
    window.addEventListener("message", function (e) {
      root.setAttribute("data-probe", JSON.stringify(e.data));
    });
    var f = document.createElement("iframe");
    f.src = document.querySelector("a.tile").getAttribute("href");
    document.body.appendChild(f);
    return;
  }
  if (location.pathname.indexOf("/species/") !== 0) return;
  setTimeout(function () {
    var out = { back: 0, followed: false, href: "", text: "" };
    history.back = function () { out.back++; };
    var a = document.querySelector("a.backlink");
    out.href = a.getAttribute("href");
    out.text = a.textContent.trim();
    document.addEventListener("click", function (e) {
      out.followed = !e.defaultPrevented;
      e.preventDefault();
    });
    a.click();
    if (window !== window.top) window.top.postMessage(out, "*");
    else root.setAttribute("data-probe", JSON.stringify(out));
  }, 50);
});
