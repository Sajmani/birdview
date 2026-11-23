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

/* Overlay interaction (decisions.overlayMode).
 *
 * The server renders the full shared-scale SVG and legend so the page works
 * cleanly without JavaScript. This script adds:
 *   - selecting a bird from the legend (or directly on the canvas) to bring it
 *     forward and dim the others;
 *   - dragging vertically (directly on a bird, or anywhere on the canvas once a
 *     bird is selected) and nudging via ▲/▼ buttons or arrow keys;
 *   - syncing vertical shifts in real millimeters (dy=<id>:<mm>) into the URL
 *     via history.replaceState so shared links match across screen sizes. */
(function () {
  "use strict";
  var wrap = document.querySelector(".ov-wrap");
  var svg = document.querySelector("svg.ov-dia");
  if (!wrap || !svg) return;

  var pxPerMm = parseFloat(wrap.getAttribute("data-px-per-mm")) || 0;
  var nudgeMm = parseInt(wrap.getAttribute("data-nudge-mm"), 10) || 10;
  var boxW = parseFloat(wrap.getAttribute("data-box-w")) || 640;
  var origBoxH = parseFloat(wrap.getAttribute("data-box-h")) || 468;
  var metric = wrap.getAttribute("data-metric") === "1";
  var maxOffset = 5000;
  if (!(pxPerMm > 0)) return;

  var birdsLayer = svg.querySelector(".ov-birds");
  var cubesLayer = svg.querySelector(".ov-cubes");
  var resetBtn = wrap.querySelector(".ov-reset");
  var spreadBtn = wrap.querySelector(".ov-spread");
  var shiftBtn = wrap.querySelector(".ov-shift");

  var birds = {};
  var order = [];
  svg.querySelectorAll(".ov-bird[data-id]").forEach(function (el) {
    var id = el.getAttribute("data-id");
    var row = wrap.querySelector('.ov-row[data-id="' + id + '"]');
    var b = {
      id: id,
      el: el,
      cubeEl: svg.querySelector('.ov-cube[data-id="' + id + '"]'),
      baseY: parseFloat(el.getAttribute("data-base-y")) || 0,
      dx: parseFloat(el.getAttribute("data-dx")) || 0,
      dy: parseInt(el.getAttribute("data-dy"), 10) || 0,
      span: parseFloat(el.getAttribute("data-span")) || 0,
      row: row,
      pick: row ? row.querySelector(".ov-pick") : null,
      dyEl: row ? row.querySelector(".ov-dy") : null,
      zeroEl: row ? row.querySelector(".ov-zero") : null,
      upEl: row ? row.querySelector(".ov-up") : null,
      dnEl: row ? row.querySelector(".ov-dn") : null
    };
    birds[id] = b;
    order.push(id);
  });

  var selectedId = "";

  function select(id) {
    if (id && !birds[id]) id = "";
    selectedId = id;
    svg.classList.toggle("has-sel", !!id);
    order.forEach(function (k) {
      var b = birds[k];
      var on = k === id;
      b.el.classList.toggle("selected", on);
      if (b.cubeEl) b.cubeEl.classList.toggle("selected", on);
      if (b.row) b.row.classList.toggle("selected", on);
      if (b.pick) b.pick.setAttribute("aria-pressed", on ? "true" : "false");
    });
    if (id) {
      if (birdsLayer) birdsLayer.appendChild(birds[id].el);
      if (cubesLayer && birds[id].cubeEl) cubesLayer.appendChild(birds[id].cubeEl);
    }
  }

  function formatShift(mm) {
    if (!mm) return "";
    var sign = mm > 0 ? "+" : "";
    if (metric) return sign + (mm / 10).toFixed(1) + " cm";
    return sign + (mm / 25.4).toFixed(2) + " in";
  }

  function updateBounds() {
    var minY = 0;
    var maxY = origBoxH;
    order.forEach(function (id) {
      var b = birds[id];
      var y = b.baseY + b.dy * pxPerMm;
      if (y - 24 < minY) minY = y - 24;
      if (y + b.span + 24 > maxY) maxY = y + b.span + 24;
    });
    svg.setAttribute("viewBox", "0 " + minY.toFixed(2) + " " + boxW.toFixed(1) + " " + (maxY - minY).toFixed(2));
    if (cubesLayer) {
      var drop = maxY - origBoxH;
      if (drop > 0.01) cubesLayer.setAttribute("transform", "translate(0," + drop.toFixed(2) + ")");
      else cubesLayer.removeAttribute("transform");
    }
  }

  function syncURL() {
    var params = new URLSearchParams(location.search);
    params.delete("dy");
    var ids = order.slice().sort();
    ids.forEach(function (id) {
      if (birds[id].dy !== 0) {
        params.append("dy", id + ":" + birds[id].dy);
      }
    });
    var qs = params.toString();
    var target = location.pathname + (qs ? "?" + qs : "") + location.hash;
    history.replaceState(null, "", target);

    [spreadBtn, shiftBtn].forEach(function (btn) {
      if (!btn) return;
      var su = new URL(btn.href, location.origin);
      su.searchParams.delete("dy");
      ids.forEach(function (id) {
        if (birds[id].dy !== 0) su.searchParams.append("dy", id + ":" + birds[id].dy);
      });
      btn.href = su.pathname + su.search;
    });
  }

  function applyOffset(id, mm, save) {
    var b = birds[id];
    if (!b) return;
    mm = Math.round(mm);
    if (mm < -maxOffset) mm = -maxOffset;
    if (mm > maxOffset) mm = maxOffset;
    b.dy = mm;
    b.el.setAttribute("data-dy", String(mm));
    var y = b.baseY + mm * pxPerMm;
    b.el.setAttribute("transform", "translate(" + b.dx.toFixed(2) + "," + y.toFixed(2) + ")");
    if (b.dyEl) {
      b.dyEl.textContent = formatShift(mm);
      if (mm === 0) b.dyEl.setAttribute("hidden", "");
      else b.dyEl.removeAttribute("hidden");
    }
    if (b.zeroEl) {
      if (mm === 0) b.zeroEl.setAttribute("hidden", "");
      else b.zeroEl.removeAttribute("hidden");
    }
    var any = order.some(function (k) { return birds[k].dy !== 0; });
    if (resetBtn) {
      if (!any) resetBtn.setAttribute("hidden", "");
      else resetBtn.removeAttribute("hidden");
    }
    updateBounds();
    if (save !== false) syncURL();
  }

  order.forEach(function (id) {
    var b = birds[id];
    if (b.pick) {
      b.pick.addEventListener("click", function () {
        select(selectedId === id ? "" : id);
      });
    }
    if (b.upEl) {
      b.upEl.addEventListener("click", function (e) {
        e.preventDefault();
        select(id);
        applyOffset(id, b.dy - nudgeMm);
      });
    }
    if (b.dnEl) {
      b.dnEl.addEventListener("click", function (e) {
        e.preventDefault();
        select(id);
        applyOffset(id, b.dy + nudgeMm);
      });
    }
    if (b.zeroEl) {
      b.zeroEl.addEventListener("click", function (e) {
        e.preventDefault();
        applyOffset(id, 0);
      });
    }
  });

  if (resetBtn) {
    resetBtn.addEventListener("click", function (e) {
      e.preventDefault();
      order.forEach(function (id) { applyOffset(id, 0, false); });
      syncURL();
    });
  }

  var drag = null;

  svg.addEventListener("pointerdown", function (e) {
    if (e.button && e.button !== 0) return;
    var hitBird = e.target.closest && e.target.closest(".ov-bird[data-id]");
    var id = hitBird ? hitBird.getAttribute("data-id") : selectedId;
    if (!id || !birds[id]) return;
    if (hitBird && selectedId !== id) {
      select(id);
    }
    var rect = svg.getBoundingClientRect();
    var vb = svg.viewBox.baseVal;
    var scaleY = rect.height > 0 && vb && vb.height > 0 ? vb.height / rect.height : 1;
    drag = {
      id: id,
      pointerId: e.pointerId,
      startClientY: e.clientY,
      startDy: birds[id].dy,
      scaleY: scaleY,
      fromBird: !!hitBird,
      moved: false
    };
    if (svg.setPointerCapture) {
      try { svg.setPointerCapture(e.pointerId); } catch (err) {}
    }
  });

  svg.addEventListener("pointermove", function (e) {
    if (!drag) return;
    var clientDelta = e.clientY - drag.startClientY;
    if (Math.abs(clientDelta) >= 3) drag.moved = true;
    if (!drag.moved) return;
    var nextMm = Math.round(drag.startDy + (clientDelta * drag.scaleY) / pxPerMm);
    if (nextMm !== birds[drag.id].dy) {
      applyOffset(drag.id, nextMm);
    }
  });

  function endDrag(e) {
    if (!drag) return;
    var was = drag;
    drag = null;
    if (svg.releasePointerCapture) {
      try { svg.releasePointerCapture(e.pointerId); } catch (err) {}
    }
    if (!was.moved && !was.fromBird && selectedId) {
      select("");
    }
  }

  svg.addEventListener("pointerup", endDrag);
  svg.addEventListener("pointercancel", endDrag);

  document.addEventListener("keydown", function (e) {
    if (!selectedId) return;
    var tag = e.target && e.target.tagName;
    if (tag === "INPUT" || tag === "TEXTAREA" || tag === "SELECT") return;
    var fine = Math.max(1, Math.round(nudgeMm / 5));
    var step = e.shiftKey ? nudgeMm : fine;
    if (e.key === "ArrowUp") {
      e.preventDefault();
      applyOffset(selectedId, birds[selectedId].dy - step);
    } else if (e.key === "ArrowDown") {
      e.preventDefault();
      applyOffset(selectedId, birds[selectedId].dy + step);
    } else if (e.key === "Escape") {
      select("");
    }
  });
})();
