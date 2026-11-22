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

/* Here (product.hereSelectsLocalRegion).
 *
 * Find the region the reader is in, here in the browser, and ask for the
 * page with that region selected -- the same request choosing it by hand
 * would make. Exact coordinates are never sent anywhere, and the state/province
 * boundary file is fetched on every page load, pressed or not
 * (product.locationStaysInBrowser).
 *
 * Under -counties (decisions.herePinpointsCountyTwoStage), once stage 1
 * resolves a US state or Canadian province locally, Here fetches that one
 * state's immutable county boundary file and resolves the county locally in
 * the browser, falling back to the state if the point lies outside county
 * polygons or the file fails to load.
 *
 * The lookup mirrors internal/boundary's Lookup: land rings by the even-odd
 * rule, in file order, then the sea boxes. TestHereInTheBrowser holds the
 * two to the same answers.
 *
 * Month's Now links to month=now; this script points it at the browser's
 * own month instead, and fills the same month into "the birds around you
 * now". Without the script, the server redirects month=now to its month. */
(function () {
  "use strict";
  var MONTHS = ["Jan", "Feb", "Mar", "Apr", "May", "Jun", "Jul", "Aug", "Sep", "Oct", "Nov", "Dec"];
  var thisMonth = MONTHS[new Date().getMonth()];
  document.querySelectorAll("a[data-now]").forEach(function (a) {
    var u = new URL(a.href);
    u.searchParams.set("month", thisMonth);
    a.href = u.pathname + u.search;
    a.title = "This month only: " + thisMonth;
  });

  var here = document.querySelector("button.here[data-template]");
  if (!here) return;
  var status = document.querySelector(".herestatus");
  var land = null, sea = null;

  // Decoded once: each ring to a flat [lon, lat, lon, lat, ...] array in
  // degrees, with the region's bounding box to skip most of them quickly.
  function decodeRings(list) {
    return (list || []).map(function (l) {
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
  }

  function decode(f) {
    land = decodeRings(f.land);
    sea = f.sea || [];
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

  function lookupLand(list, lat, lon) {
    for (var k = 0; k < list.length; k++) {
      var l = list[k], b = l.box;
      if (lat < b[0] || lat > b[1] || lon < b[2] || lon > b[3]) continue;
      var inside = false;
      for (var r = 0; r < l.rings.length; r++) {
        if (inRing(l.rings[r], lat, lon)) inside = !inside;
      }
      if (inside) return l.region;
    }
    return "";
  }

  function lookup(lat, lon) {
    var hit = lookupLand(land, lat, lon);
    if (hit) return hit;
    for (var s = 0; s < sea.length; s++) {
      var x = sea[s].box;
      if (lat >= x[0] && lat <= x[1] && lon >= x[2] && lon <= x[3]) return sea[s].region;
    }
    return "";
  }

  var countyTemplate = here.getAttribute("data-county-boundaries");
  var countyCache = {};
  function loadCounties(state) {
    if (!countyCache[state]) {
      var url = countyTemplate.replace("__STATE__", encodeURIComponent(state));
      countyCache[state] = fetch(url).then(function (r) {
        if (!r.ok) return null;
        return r.json().then(function (f) { return decodeRings(f.land); });
      }).catch(function () { return null; });
    }
    return countyCache[state];
  }

  function resolve(lat, lon) {
    return ready.then(function () {
      var landSlug = lookupLand(land, lat, lon);
      if (landSlug) {
        if (!countyTemplate || !(landSlug.indexOf("USA.") === 0 || landSlug.indexOf("CAN.") === 0)) {
          return landSlug;
        }
        return loadCounties(landSlug).then(function (counties) {
          return (counties && lookupLand(counties, lat, lon)) || landSlug;
        });
      }
      for (var s = 0; s < sea.length; s++) {
        var x = sea[s].box;
        if (lat >= x[0] && lat <= x[1] && lon >= x[2] && lon <= x[3]) return sea[s].region;
      }
      return "";
    });
  }

  var ready = fetch(here.getAttribute("data-boundaries"))
    .then(function (r) { if (!r.ok) throw new Error(r.status); return r.json(); })
    .then(decode);
  ready.catch(function () {});

  window.BirdViewHere = {
    ready: ready,
    lookup: function (lat, lon) { return lookup(lat, lon); },
    resolve: resolve
  };

  if (!("geolocation" in navigator)) {
    here.title = "Here needs a browser that can share your location";
    return;
  }
  here.removeAttribute("aria-disabled");
  here.title = "Select only the region you are in. Your location stays in this browser.";

  function say(text) { status.textContent = text; }

  // Arriving by a link ending in #here-and-now presses Here, and the
  // template sets the month to Now, the browser's month. On load, so a test can
  // stand in for the browser's location first, and when the fragment changes
  // to it, since the home page's link is its own address with the fragment
  // added, which a browser follows without loading the page. The fragment is
  // dropped either way, so a reload or the Back button does not ask again.
  var welcome = document.querySelector(".welcomestatus[data-template]");
  function asked() { return new URLSearchParams(location.hash.slice(1)).has("here-and-now"); }
  if (welcome) {
    // The page names where a region is chosen by hand, its Place setting;
    // a failure leaves the page as it was.
    var where = welcome.getAttribute("data-where") || "Place";
    var arrive = function () {
      var tell = function (text) {
        welcome.textContent = text;
        history.replaceState(null, "", location.pathname + location.search);
      };
      tell("Finding where you are\u2026");
      navigator.geolocation.getCurrentPosition(function (pos) {
        resolve(pos.coords.latitude, pos.coords.longitude).then(function (slug) {
          if (!slug) {
            tell("You are outside every region BirdView covers, so the page is unchanged.");
            return;
          }
          window.location.replace(welcome.getAttribute("data-template")
            .replace("__REGION__", slug).replace("__MONTH__", thisMonth));
        }, function () {
          tell("The region boundaries did not load, so the page is unchanged.");
        });
      }, function (err) {
        tell(err.code === 1
          ? "Permission to use your location was refused, so the page is unchanged. Choose yours under " + where + "."
          : "Your location could not be found, so the page is unchanged. Choose yours under " + where + ".");
      }, { timeout: 15000, maximumAge: 600000 });
    };
    if (asked()) window.addEventListener("load", arrive);
    window.addEventListener("hashchange", function () { if (asked()) arrive(); });
  }

  here.addEventListener("click", function () {
    say("Finding where you are\u2026");
    navigator.geolocation.getCurrentPosition(function (pos) {
      resolve(pos.coords.latitude, pos.coords.longitude).then(function (slug) {
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
