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

// Drives TestSpeciesPinAsksTheHomePage. The home page, loaded on top, opens
// two species pages in frames, which share its session: amecro, unpinned,
// and annhum, pinned. Each reports its Pin's state; amecro's is clicked with
// history.back stood in for. Then the home page loads itself again in a
// third frame, as going back would, and that frame asks the server to toggle
// the bird, which the test reads from the server's log. The frames' reports
// are data-probe on the top page's root.
window.addEventListener("load", function () {
  var root = document.documentElement;
  if (window === window.top && location.pathname === "/") {
    var out = {};
    window.addEventListener("message", function (e) {
      out[e.data.frame] = e.data;
      if (out.speciesA && out.speciesB && !out.sent) {
        out.sent = true;
        var f = document.createElement("iframe");
        f.name = "home2";
        f.src = location.pathname + location.search;
        document.body.appendChild(f);
        setTimeout(function () { root.setAttribute("data-probe", JSON.stringify(out)); }, 1500);
      }
    });
    [["speciesA", "amecro"], ["speciesB", "annhum"]].forEach(function (p) {
      var f = document.createElement("iframe");
      f.name = p[0];
      f.src = "/species/" + p[1];
      document.body.appendChild(f);
    });
    return;
  }
  if (window.name !== "speciesA" && window.name !== "speciesB") return;
  setTimeout(function () {
    var a = document.querySelector("a[data-pin]");
    var r = { frame: window.name, label: a.querySelector(".pinlabel").textContent,
      pressed: a.getAttribute("aria-pressed"), back: 0, toggle: "" };
    if (window.name === "speciesA") {
      history.back = function () { r.back++; };
      a.click();
      r.toggle = sessionStorage.getItem("birdview.toggle") || "";
    }
    window.top.postMessage(r, "*");
  }, 50);
});
