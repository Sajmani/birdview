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


// Drives TestNextTileCornerAndRows. For every tile on the page, it measures
// whether the pin sits in the tile's top-right corner with the common name
// short of it, and the iNat link in the lower-right corner with the
// scientific name short of it; whether
// the tile is wider than its box or its drawing off center; and, row by row,
// whether every drawing starts at the same height. It records the tiles
// measured and every one more than a pixel off as data-probe on the root.
window.addEventListener("load", function () {
  setTimeout(function () {
    var out = { tiles: 0, notes: 0, inats: 0, rows: 0, long: 0, corner: [], uncentered: [], ragged: [] };
    var rows = {};
    document.querySelectorAll(".tilewrap").forEach(function (t) {
      var pin = t.querySelector(":scope > a.pin");
      var tile = t.querySelector("a.tile"), dia = tile && tile.querySelector("svg.dia");
      if (!pin || !dia) return;
      out.tiles++;
      if (t.id == "s-migwar" || t.id == "s-gnlthr") out.long++;
      if (t.querySelector(".seam")) out.notes++;
      var r = tile.getBoundingClientRect(), g = dia.getBoundingClientRect();
      var p = pin.getBoundingClientRect();
      var name = tile.querySelector(".cnname").getBoundingClientRect();
      // In the corner: inside the tile, within its 10px padding and border
      // of the top and right edges.
      var top = Math.round(p.top - r.top), right = Math.round(r.right - p.right);
      if (top < 0 || top > 12 || right < 0 || right > 12) out.corner.push(t.id + " pin " + top + "px from the top, " + right + "px from the right");
      if (name.right > p.left + 0.5) out.corner.push(t.id + " name runs " + Math.round(name.right - p.left) + "px under the pin");
      // The iNat link: in the lower-right corner, its foot on the bottom
      // line's and its top below the rule, with the scientific name short of it.
      var inat = t.querySelector(":scope > a.inat");
      if (inat) {
        out.inats++;
        var n = inat.getBoundingClientRect(), line = tile.querySelector(".idline").getBoundingClientRect();
        var rule = tile.querySelector(".lo").getBoundingClientRect(), sn = tile.querySelector(".sn").getBoundingClientRect();
        var ir = Math.round(r.right - n.right);
        if (ir < 0 || ir > 12) out.corner.push(t.id + " iNat " + ir + "px from the right");
        if (Math.abs(n.bottom - line.bottom) > 1) out.corner.push(t.id + " iNat's foot " + Math.round(n.bottom - line.bottom) + "px off the bottom line's");
        if (n.top < rule.top) out.corner.push(t.id + " iNat reaches " + Math.round(rule.top - n.top) + "px above the rule");
        if (sn.right > n.left + 0.5) out.corner.push(t.id + " scientific name runs " + Math.round(sn.right - n.left) + "px under iNat");
      }
      // A long name truncates; it must not widen the tile and push the
      // drawing off center (#4).
      var wide = tile.scrollWidth - tile.clientWidth;
      var shift = Math.round(g.left + g.width / 2 - (r.left + r.width / 2));
      if (wide > 1 || Math.abs(shift) > 1) out.uncentered.push(t.id + " wider by " + wide + "px, drawing off by " + shift + "px");
      var key = Math.round(r.top);
      (rows[key] = rows[key] || []).push({ id: t.id, top: g.top });
    });
    // Tiles side by side share a top; their drawings must too, so notes,
    // which only some tiles carry, may not push one drawing down.
    Object.keys(rows).forEach(function (k) {
      var row = rows[k];
      if (row.length < 2) return;
      out.rows++;
      row.forEach(function (x) {
        if (Math.abs(x.top - row[0].top) > 1) out.ragged.push(x.id + " drawing " + Math.round(x.top - row[0].top) + "px below " + row[0].id + "'s");
      });
    });
    document.documentElement.setAttribute("data-probe", JSON.stringify(out));
  }, 100);
});
