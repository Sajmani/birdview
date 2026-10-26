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

// Drives TestNextPinKeepsPlace. In the top page it scrolls down, clicks the
// pin of a tile in view with the navigation cancelled, so the page's own
// click handler runs, then loads the pin's address in a frame as tall as the
// window, which shares the session's storage, and records where the frame
// is scrolled. The result is data-probe on the root. In a frame it does
// nothing.
window.addEventListener("load", function () {
  if (window !== window.top) return;
  var out = {};
  setTimeout(function () {
    window.scrollTo(0, 400);
    out.before = Math.round(window.scrollY);
    var pin = document.querySelectorAll(".tilewrap > a.pin")[3];
    document.addEventListener("click", function (e) { e.preventDefault(); });
    pin.click();
    out.href = pin.getAttribute("href");
    var f = document.createElement("iframe");
    f.style.cssText = "position:fixed;top:0;left:0;width:800px;height:600px;border:0";
    f.addEventListener("load", function () {
      setTimeout(function () {
        out.after = Math.round(f.contentWindow.scrollY);
        document.documentElement.setAttribute("data-probe", JSON.stringify(out));
      }, 100);
    });
    f.src = out.href;
    document.body.appendChild(f);
  }, 100);
});
