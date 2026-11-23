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
 * cleanly without JavaScript. Birds are not draggable on the canvas until one
 * is selected below, keeping page touch scrolling intact; once selected, this
 * script supports 2D canvas drags and arrow-key nudges, tracks and de-overlaps
 * wingtip and mass-cube badges, and syncs 2D offsets in real millimeters
 * (drag=<id>:<dx>,<dy>) via history.replaceState. */
(function () {
  "use strict";
  var wrap = document.querySelector(".ov-wrap");
  var svg = document.querySelector("svg.ov-dia");
  if (!wrap || !svg) return;

  var pxPerMm = parseFloat(wrap.getAttribute("data-px-per-mm")) || 0;
  var nudgeMm = parseInt(wrap.getAttribute("data-nudge-mm"), 10) || 10;
  var boxW = parseFloat(wrap.getAttribute("data-box-w")) || 640;
  var origBoxH = parseFloat(wrap.getAttribute("data-box-h")) || 468;
  var cubeMinX = parseFloat(wrap.getAttribute("data-cube-min-x"));
  if (isNaN(cubeMinX)) cubeMinX = 0;
  var cubeMaxX = parseFloat(wrap.getAttribute("data-cube-max-x"));
  if (isNaN(cubeMaxX)) cubeMaxX = boxW;
  var cubeFloorY = parseFloat(wrap.getAttribute("data-cube-floor-y")) || 0;
  var cubeReserve = parseFloat(wrap.getAttribute("data-cube-reserve")) || 80;
  var pad = 24;
  var maxOffset = 5000;
  if (!(pxPerMm > 0)) return;

  var canvas = wrap.querySelector(".ov-canvas");
  var birdsLayer = svg.querySelector(".ov-birds");
  var cubesLayer = svg.querySelector(".ov-cubes");
  var guideLine = svg.querySelector(".ov-guide");
  var resetBtn = wrap.querySelector(".ov-reset");
  var spreadBtn = wrap.querySelector(".ov-spread");
  var shiftBtn = wrap.querySelector(".ov-shift");
  var alignLinks = Array.prototype.slice.call(wrap.querySelectorAll(".ov-bar .seg a[href]"));

  var birds = {};
  var order = [];
  svg.querySelectorAll(".ov-bird[data-id]").forEach(function (el) {
    var id = el.getAttribute("data-id");
    var row = wrap.querySelector('.ov-row[data-id="' + id + '"]');
    var cubeTagEl = wrap.querySelector('.ov-tag.ov-ctag[data-id="' + id + '"]');
    var b = {
      id: id,
      el: el,
      cubeEl: svg.querySelector('.ov-cube[data-id="' + id + '"]'),
      tagEl: wrap.querySelector('.ov-tag.ov-wtag[data-id="' + id + '"]'),
      cubeTagEl: cubeTagEl,
      cubeX: cubeTagEl ? parseFloat(cubeTagEl.getAttribute("data-cube-x")) || 0 : 0,
      baseX: parseFloat(el.getAttribute("data-base-x")) || 0,
      baseY: parseFloat(el.getAttribute("data-base-y")) || 0,
      dragX: parseInt(el.getAttribute("data-drag-x"), 10) || 0,
      dragY: parseInt(el.getAttribute("data-drag-y"), 10) || 0,
      cx: parseFloat(el.getAttribute("data-cx")) || 0,
      halfW: parseFloat(el.getAttribute("data-half-w")) || 0,
      shoulderY: parseFloat(el.getAttribute("data-shoulder-y")) || 0,
      span: parseFloat(el.getAttribute("data-span")) || 0,
      row: row,
      pick: row ? row.querySelector(".ov-pick") : null,
      zeroEl: row ? row.querySelector(".ov-zero") : null
    };
    birds[id] = b;
    order.push(id);
  });

  var selectedId = "";

  function clampMm(v) {
    v = Math.round(v);
    if (v < -maxOffset) return -maxOffset;
    if (v > maxOffset) return maxOffset;
    return v;
  }

  function select(id) {
    if (id && !birds[id]) id = "";
    selectedId = id;
    svg.classList.toggle("has-sel", !!id);
    if (canvas) canvas.classList.toggle("has-sel", !!id);
    if (!id) svg.classList.remove("dragging");
    order.forEach(function (k) {
      var b = birds[k];
      var on = k === id;
      b.el.classList.toggle("selected", on);
      if (b.cubeEl) b.cubeEl.classList.toggle("selected", on);
      if (b.tagEl) b.tagEl.classList.toggle("selected", on);
      if (b.cubeTagEl) b.cubeTagEl.classList.toggle("selected", on);
      if (b.row) b.row.classList.toggle("selected", on);
      if (b.pick) b.pick.setAttribute("aria-pressed", on ? "true" : "false");
    });
    if (id) {
      if (birdsLayer) birdsLayer.appendChild(birds[id].el);
      if (cubesLayer && birds[id].cubeEl) cubesLayer.appendChild(birds[id].cubeEl);
    }
  }

  function updateTags(vx, vy, vw, vh, drop) {
    var rect = svg.getBoundingClientRect();
    if (!(rect.width > 0 && rect.height > 0 && vw > 0 && vh > 0)) return;
    var s = Math.min(rect.width / vw, rect.height / vh);
    var offX = (rect.width - vw * s) / 2;
    var offY = (rect.height - vh * s) / 2;
    var minX = 12;
    var maxX = Math.max(12, rect.width - 12);
    var minY = 12;
    var maxY = Math.max(12, rect.height - 12);
    var minDX = 22;
    var minDY = 22;

    var placed = [];
    function conflicts(px, py) {
      for (var i = 0; i < placed.length; i++) {
        if (Math.abs(px - placed[i].x) < minDX && Math.abs(py - placed[i].y) < minDY) {
          return true;
        }
      }
      return false;
    }

    order.forEach(function (id) {
      var b = birds[id];
      if (!b.tagEl) return;
      var bx = b.baseX + b.dragX * pxPerMm;
      var by = b.baseY + b.dragY * pxPerMm;
      var tipX = offX + (bx + b.cx + b.halfW - vx) * s;
      var tipY = offY + (by + b.shoulderY - vy) * s - 14;
      var px = Math.max(minX, Math.min(maxX, tipX));
      var basePy = Math.max(minY, Math.min(maxY, tipY));
      var py = basePy;
      if (conflicts(px, py)) {
        var found = false;
        for (var step = 1; step <= 12; step++) {
          var up = basePy - step * minDY;
          if (up >= minY && !conflicts(px, up)) {
            py = up;
            found = true;
            break;
          }
          var dn = basePy + step * minDY;
          if (dn <= maxY && !conflicts(px, dn)) {
            py = dn;
            found = true;
            break;
          }
        }
        if (!found) {
          var bestClear = -1;
          for (var cy = minY; cy <= maxY; cy += 2) {
            var nearest = 1e9;
            for (var i = 0; i < placed.length; i++) {
              if (Math.abs(px - placed[i].x) < minDX) {
                var d = Math.abs(cy - placed[i].y);
                if (d < nearest) nearest = d;
              }
            }
            if (nearest > bestClear) {
              bestClear = nearest;
              py = cy;
            }
          }
        }
      }
      placed.push({ x: px, y: py });
      b.tagEl.style.left = ((px / rect.width) * 100).toFixed(2) + "%";
      b.tagEl.style.top = ((py / rect.height) * 100).toFixed(2) + "%";
    });

    if (cubeFloorY > 0) {
      var cubePy = Math.max(minY, Math.min(maxY, offY + (cubeFloorY + drop - vy) * s + 14));
      order.forEach(function (id) {
        var b = birds[id];
        if (!b.cubeTagEl) return;
        var cpx = Math.max(minX, Math.min(maxX, offX + (b.cubeX - vx) * s));
        b.cubeTagEl.style.left = ((cpx / rect.width) * 100).toFixed(2) + "%";
        b.cubeTagEl.style.top = ((cubePy / rect.height) * 100).toFixed(2) + "%";
      });
    }
  }

  function updateBounds() {
    var minX = Math.min(0, cubeMinX);
    var maxX = Math.max(boxW, cubeMaxX);
    var minY = 0;
    var maxY = origBoxH;
    order.forEach(function (id) {
      var b = birds[id];
      var x = b.baseX + b.dragX * pxPerMm;
      var y = b.baseY + b.dragY * pxPerMm;
      var left = x + b.cx - b.halfW - pad;
      var right = x + b.cx + b.halfW + pad;
      if (left < minX) minX = left;
      if (right > maxX) maxX = right;
      if (y - pad < minY) minY = y - pad;
      if (y + b.span + cubeReserve > maxY) maxY = y + b.span + cubeReserve;
    });
    var vw = maxX - minX;
    var vh = maxY - minY;
    svg.setAttribute("viewBox", minX.toFixed(2) + " " + minY.toFixed(2) + " " + vw.toFixed(2) + " " + vh.toFixed(2));
    if (guideLine) {
      guideLine.setAttribute("x1", minX.toFixed(2));
      guideLine.setAttribute("x2", maxX.toFixed(2));
    }
    var drop = maxY - origBoxH;
    if (cubesLayer) {
      if (Math.abs(drop) > 0.01) {
        cubesLayer.setAttribute("transform", "translate(0," + drop.toFixed(2) + ")");
      } else {
        cubesLayer.removeAttribute("transform");
      }
    }
    updateTags(minX, minY, vw, vh, drop);
  }

  function appendDrags(params, ids, verticalToo) {
    params.delete("drag");
    params.delete("dy");
    ids.forEach(function (id) {
      var b = birds[id];
      var dy = verticalToo ? b.dragY : 0;
      if (b.dragX !== 0 || dy !== 0) {
        params.append("drag", id + ":" + b.dragX + "," + dy);
      }
    });
  }

  function syncURL() {
    var ids = order.slice().sort();
    var params = new URLSearchParams(location.search);
    appendDrags(params, ids, true);
    var qs = params.toString();
    history.replaceState(null, "", location.pathname + (qs ? "?" + qs : "") + location.hash);

    [spreadBtn, shiftBtn].forEach(function (btn) {
      if (!btn) return;
      var su = new URL(btn.href, location.origin);
      appendDrags(su.searchParams, ids, true);
      btn.href = su.pathname + su.search;
    });

    alignLinks.forEach(function (link) {
      var au = new URL(link.href, location.origin);
      appendDrags(au.searchParams, ids, false);
      link.href = au.pathname + au.search;
    });
  }

  function applyOffset(id, mmX, mmY, save) {
    var b = birds[id];
    if (!b) return;
    mmX = clampMm(mmX);
    mmY = clampMm(mmY);
    b.dragX = mmX;
    b.dragY = mmY;
    b.el.setAttribute("data-drag-x", String(mmX));
    b.el.setAttribute("data-drag-y", String(mmY));
    var x = b.baseX + mmX * pxPerMm;
    var y = b.baseY + mmY * pxPerMm;
    b.el.setAttribute("transform", "translate(" + x.toFixed(2) + "," + y.toFixed(2) + ")");
    if (b.zeroEl) {
      if (mmX === 0 && mmY === 0) b.zeroEl.setAttribute("hidden", "");
      else b.zeroEl.removeAttribute("hidden");
    }
    var any = order.some(function (k) { return birds[k].dragX !== 0 || birds[k].dragY !== 0; });
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
    if (b.zeroEl) {
      b.zeroEl.addEventListener("click", function (e) {
        e.preventDefault();
        applyOffset(id, 0, 0);
      });
    }
  });

  if (resetBtn) {
    resetBtn.addEventListener("click", function (e) {
      e.preventDefault();
      order.forEach(function (id) { applyOffset(id, 0, 0, false); });
      syncURL();
    });
  }

  var drag = null;

  svg.addEventListener("pointerdown", function (e) {
    if (e.button && e.button !== 0) return;
    if (!selectedId || !birds[selectedId]) return;
    var rect = svg.getBoundingClientRect();
    var vb = svg.viewBox.baseVal;
    var scale = 1;
    if (rect.width > 0 && rect.height > 0 && vb && vb.width > 0 && vb.height > 0) {
      scale = Math.max(vb.width / rect.width, vb.height / rect.height);
    }
    drag = {
      id: selectedId,
      pointerId: e.pointerId,
      startClientX: e.clientX,
      startClientY: e.clientY,
      startDragX: birds[selectedId].dragX,
      startDragY: birds[selectedId].dragY,
      scale: scale,
      moved: false
    };
    svg.classList.add("dragging");
    if (svg.setPointerCapture) {
      try { svg.setPointerCapture(e.pointerId); } catch (err) {}
    }
  });

  svg.addEventListener("pointermove", function (e) {
    if (!drag) return;
    var dxClient = e.clientX - drag.startClientX;
    var dyClient = e.clientY - drag.startClientY;
    if (Math.hypot(dxClient, dyClient) >= 3) drag.moved = true;
    if (!drag.moved) return;
    var nextX = clampMm(drag.startDragX + (dxClient * drag.scale) / pxPerMm);
    var nextY = clampMm(drag.startDragY + (dyClient * drag.scale) / pxPerMm);
    var b = birds[drag.id];
    if (nextX !== b.dragX || nextY !== b.dragY) {
      applyOffset(drag.id, nextX, nextY);
    }
  });

  function endDrag(e) {
    if (!drag) return;
    var was = drag;
    drag = null;
    svg.classList.remove("dragging");
    if (svg.releasePointerCapture) {
      try { svg.releasePointerCapture(e.pointerId); } catch (err) {}
    }
    if (!was.moved && selectedId) {
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
    var b = birds[selectedId];
    if (e.key === "ArrowLeft") {
      e.preventDefault();
      applyOffset(selectedId, b.dragX - step, b.dragY);
    } else if (e.key === "ArrowRight") {
      e.preventDefault();
      applyOffset(selectedId, b.dragX + step, b.dragY);
    } else if (e.key === "ArrowUp") {
      e.preventDefault();
      applyOffset(selectedId, b.dragX, b.dragY - step);
    } else if (e.key === "ArrowDown") {
      e.preventDefault();
      applyOffset(selectedId, b.dragX, b.dragY + step);
    } else if (e.key === "Escape") {
      select("");
    }
  });

  window.addEventListener("resize", updateBounds);
  updateBounds();
})();
