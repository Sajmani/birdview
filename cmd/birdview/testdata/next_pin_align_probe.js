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

// Drives TestNextPinAlignsWithCube. For every tile on the page, it measures
// how far the foot of the tile's pin is from the foot of its mass cube, and
// whether the tile is wider than its box or its drawing off center, and
// records the tiles measured and those more than a pixel off as data-probe
// on the root.
window.addEventListener("load", function () {
  setTimeout(function () {
    var out = { tiles: 0, notes: 0, off: [], uncentered: [] };
    document.querySelectorAll(".tilewrap").forEach(function (t) {
      var cube = t.querySelector("svg .masscube");
      var pin = t.querySelector(":scope > a.pin");
      if (!cube || !pin) return;
      out.tiles++;
      if (t.querySelector(".share .seam")) out.notes++;
      var d = Math.round(pin.getBoundingClientRect().bottom - cube.getBoundingClientRect().bottom);
      if (Math.abs(d) > 1) out.off.push(t.id + " " + d + "px");
      // A long name truncates; it must not widen the tile and push the
      // drawing off center (#4).
      var tile = t.querySelector("a.tile"), dia = tile.querySelector("svg.dia");
      var r = tile.getBoundingClientRect(), g = dia.getBoundingClientRect();
      var wide = tile.scrollWidth - tile.clientWidth;
      var shift = Math.round(g.left + g.width / 2 - (r.left + r.width / 2));
      if (wide > 1 || Math.abs(shift) > 1) out.uncentered.push(t.id + " wider by " + wide + "px, drawing off by " + shift + "px");
    });
    document.documentElement.setAttribute("data-probe", JSON.stringify(out));
  }, 100);
});
