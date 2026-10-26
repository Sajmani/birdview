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

// Drives Here and Now for TestHereInTheBrowser. Instructions come in the URL
// fragment, which the browser never sends, so the server sees exactly what a
// reader's browser would send:
//
//   probe=1          run, and record the result as data-probe on the root
//   points=[[lat,lon],...]  look each up with the page's own lookup
//   click=lat,lon | denied | unavailable
//                    stub the browser's location that way, then press Here
//   geo=lat,lon | denied | unavailable
//                    stub the location as the page loads, for #here-and-now
//   follow=1         click the welcome line's #here-and-now link, as a reader
//                    would, instead of arriving with the fragment
//
// The fragment is read as soon as this runs, because the page may drop it on
// load before this script's own load handler sees it.
// Every page also records its own path and query as data-url, so a press
// that navigates can be seen to have landed.
function stubLocation(how) {
  navigator.geolocation.getCurrentPosition = function (ok, fail) {
    if (how === "denied") fail({ code: 1 });
    else if (how === "unavailable") fail({ code: 2 });
    else {
      var c = how.split(",");
      ok({ coords: { latitude: +c[0], longitude: +c[1] } });
    }
  };
}
var params = new URLSearchParams(location.hash.slice(1));
if (params.has("geo")) stubLocation(params.get("geo"));
window.addEventListener("load", function () {
  var root = document.documentElement;
  root.setAttribute("data-url", location.pathname + location.search);
  if (!params.has("probe")) return;
  var out = {};
  var now = document.querySelector("a.now");
  out.now = now ? now.getAttribute("href") : null;
  var here = document.querySelector("button.here");
  out.enabled = here ? !here.hasAttribute("aria-disabled") : false;
  window.BirdViewHere.ready.then(function () {
    out.lookups = JSON.parse(params.get("points") || "[]").map(function (p) {
      return window.BirdViewHere.lookup(p[0], p[1]);
    });
    if (params.has("follow")) {
      document.querySelector('.welcome a[href$="#here-and-now"]').click();
    }
    var click = params.get("click");
    if (click) {
      stubLocation(click);
      here.click();
    }
    setTimeout(function () {
      var status = document.querySelector(".herestatus");
      out.status = status ? status.textContent : null;
      var welcome = document.querySelector(".welcomestatus");
      out.welcome = welcome ? welcome.textContent : null;
      root.setAttribute("data-probe", JSON.stringify(out));
    }, 50);
  }, function (err) {
    root.setAttribute("data-probe", JSON.stringify({ error: String(err) }));
  });
});
