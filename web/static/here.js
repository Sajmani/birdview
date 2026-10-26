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

/* Here and Now (product.hereSelectsLocalRegion, product.nowSelectsCurrentMonth).
 *
 * Now: point the button at the reader's own month. Without this script it
 * goes through /now, which uses the server's.
 *
 * Here: find the region the reader is in, here in the browser, and ask for
 * the page with that region selected -- the same request choosing it by hand
 * would make. The position is never sent anywhere, and the boundary file is
 * fetched on every page load, pressed or not, so the server cannot even tell
 * that Here was used (product.locationStaysInBrowser).
 *
 * The lookup mirrors internal/boundary's Lookup: land rings by the even-odd
 * rule, in file order, then the sea boxes. TestHereInTheBrowser holds the
 * two to the same answers. */
(function () {
  "use strict";
  var MONTHS = ["Jan", "Feb", "Mar", "Apr", "May", "Jun",
    "Jul", "Aug", "Sep", "Oct", "Nov", "Dec"];

  var now = document.querySelector("a.now[data-template]");
  if (now) {
    now.href = now.getAttribute("data-template")
      .replace("__MONTH__", MONTHS[new Date().getMonth()]);
  }

  var here = document.querySelector("button.here[data-template]");
  if (!here) return;
  var status = document.querySelector(".herestatus");
  var land = null, sea = null;

  // Decoded once: each ring to a flat [lon, lat, lon, lat, ...] array in
  // degrees, with the region's bounding box to skip most of them quickly.
  function decode(f) {
    land = f.land.map(function (l) {
      var rings = [], box = [Infinity, -Infinity, Infinity, -Infinity];
      l.rings.forEach(function (enc) {
        var pts = new Float64Array(enc.length), x = 0, y = 0;
        for (var i = 0; i + 1 < enc.length; i += 2) {
          x += enc[i]; y += enc[i + 1];
          pts[i] = x / 1000; pts[i + 1] = y / 1000;
          if (pts[i + 1] < box[0]) box[0] = pts[i + 1];
          if (pts[i + 1] > box[1]) box[1] = pts[i + 1];
          if (pts[i] < box[2]) box[2] = pts[i];
          if (pts[i] > box[3]) box[3] = pts[i];
        }
        rings.push(pts);
      });
      return { region: l.region, rings: rings, box: box };
    });
    sea = f.sea;
  }

  function inRing(pts, lat, lon) {
    var inside = false, n = pts.length;
    for (var i = 0, j = n - 2; i < n; j = i, i += 2) {
      var xi = pts[i], yi = pts[i + 1], xj = pts[j], yj = pts[j + 1];
      if ((yi > lat) !== (yj > lat) && lon < (xj - xi) * (lat - yi) / (yj - yi) + xi) {
        inside = !inside;
      }
    }
    return inside;
  }

  function lookup(lat, lon) {
    for (var k = 0; k < land.length; k++) {
      var l = land[k], b = l.box;
      if (lat < b[0] || lat > b[1] || lon < b[2] || lon > b[3]) continue;
      var inside = false;
      for (var r = 0; r < l.rings.length; r++) {
        if (inRing(l.rings[r], lat, lon)) inside = !inside;
      }
      if (inside) return l.region;
    }
    for (var s = 0; s < sea.length; s++) {
      var x = sea[s].box;
      if (lat >= x[0] && lat <= x[1] && lon >= x[2] && lon <= x[3]) return sea[s].region;
    }
    return "";
  }

  var ready = fetch(here.getAttribute("data-boundaries"))
    .then(function (r) { if (!r.ok) throw new Error(r.status); return r.json(); })
    .then(decode);
  ready.catch(function () {});

  window.BirdViewHere = { ready: ready, lookup: function (lat, lon) { return lookup(lat, lon); } };

  if (!("geolocation" in navigator)) {
    here.title = "Here needs a browser that can share your location";
    return;
  }
  here.removeAttribute("aria-disabled");
  here.title = "Select only the region you are in. Your location stays in this browser.";

  function say(text) { status.textContent = text; }

  // Arriving by a link ending in #here-and-now presses Here and Now
  // together. On load, so a test can stand in for the browser's location
  // first. The fragment is dropped either way, so a reload or the Back button
  // does not ask again.
  var welcome = document.querySelector(".welcomestatus[data-template]");
  if (welcome && new URLSearchParams(location.hash.slice(1)).has("here-and-now")) {
    // The page names where a region is chosen by hand; the classic view's
    // Regions and months panel unless it says otherwise. The home page,
    // already in this month, names its Region setting and offers no Now.
    var where = welcome.getAttribute("data-where") || "Regions and months";
    var nowHint = welcome.hasAttribute("data-where") ? "" : " Press Now under Regions and months for this month.";
    var scope = welcome.hasAttribute("data-where") ? "every region" : "every region and month";
    window.addEventListener("load", function () {
      var tell = function (text) {
        welcome.textContent = text;
        history.replaceState(null, "", location.pathname + location.search);
      };
      tell("Finding where you are\u2026");
      navigator.geolocation.getCurrentPosition(function (pos) {
        ready.then(function () {
          var slug = lookup(pos.coords.latitude, pos.coords.longitude);
          if (!slug) {
            tell("You are outside every region BirdView covers, so this shows every region." + nowHint);
            return;
          }
          window.location.replace(welcome.getAttribute("data-template")
            .replace("__REGION__", slug).replace("__MONTH__", MONTHS[new Date().getMonth()]));
        }, function () {
          tell("The region boundaries did not load, so this shows every region.");
        });
      }, function (err) {
        tell(err.code === 1
          ? "Permission to use your location was refused, so this shows " + scope + ". Choose yours under " + where + "."
          : "Your location could not be found, so this shows " + scope + ". Choose yours under " + where + ".");
      }, { timeout: 15000, maximumAge: 600000 });
    });
  }

  here.addEventListener("click", function () {
    say("Finding where you are\u2026");
    navigator.geolocation.getCurrentPosition(function (pos) {
      ready.then(function () {
        var slug = lookup(pos.coords.latitude, pos.coords.longitude);
        if (!slug) {
          say("You are outside every region BirdView covers, so the regions are unchanged.");
          return;
        }
        window.location.assign(here.getAttribute("data-template").replace("__REGION__", slug));
      }, function () {
        say("The region boundaries did not load, so the regions are unchanged.");
      });
    }, function (err) {
      say(err.code === 1
        ? "Permission to use your location was refused, so the regions are unchanged."
        : "Your location could not be found, so the regions are unchanged.");
    }, { timeout: 15000, maximumAge: 600000 });
  });
})();
