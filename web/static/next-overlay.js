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
 * script supports 2D canvas drags and arrow-key nudges while syncing 2D offsets
 * in real millimeters (drag=<id>:<dx>,<dy>) via history.replaceState. */
(function () {
  "use strict";
  var wrap = document.querySelector(".ov-wrap");
  var svg = document.querySelector("svg.ov-dia");
  if (!wrap || !svg) return;

  var pxPerMm = parseFloat(wrap.getAttribute("data-px-per-mm")) || 0;
  var nudgeMm = parseInt(wrap.getAttribute("data-nudge-mm"), 10) || 10;
  var boxW = parseFloat(wrap.getAttribute("data-box-w")) || 640;
  var origBoxH = parseFloat(wrap.getAttribute("data-box-h")) || 468;
  var origViewX = parseFloat(wrap.getAttribute("data-view-x")) || 0;
  var pad = 24;
  var maxOffset = 5000;
  if (!(pxPerMm > 0)) return;

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
    var b = {
      id: id,
      el: el,
      cubeEl: svg.querySelector('.ov-cube[data-id="' + id + '"]'),
      baseX: parseFloat(el.getAttribute("data-base-x")) || 0,
      baseY: parseFloat(el.getAttribute("data-base-y")) || 0,
      dragX: parseInt(el.getAttribute("data-drag-x"), 10) || 0,
      dragY: parseInt(el.getAttribute("data-drag-y"), 10) || 0,
      cx: parseFloat(el.getAttribute("data-cx")) || 0,
      halfW: parseFloat(el.getAttribute("data-half-w")) || 0,
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
    if (!id) svg.classList.remove("dragging");
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

  function updateBounds() {
    var minX = 0;
    var maxX = boxW;
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
      if (y + b.span + pad > maxY) maxY = y + b.span + pad;
    });
    svg.setAttribute(
      "viewBox",
      minX.toFixed(2) + " " + minY.toFixed(2) + " " + (maxX - minX).toFixed(2) + " " + (maxY - minY).toFixed(2)
    );
    if (guideLine) {
      guideLine.setAttribute("x1", minX.toFixed(2));
      guideLine.setAttribute("x2", maxX.toFixed(2));
    }
    if (cubesLayer) {
      var dx = minX - origViewX;
      var dy = maxY - origBoxH;
      if (Math.abs(dx) > 0.01 || Math.abs(dy) > 0.01) {
        cubesLayer.setAttribute("transform", "translate(" + dx.toFixed(2) + "," + dy.toFixed(2) + ")");
      } else {
        cubesLayer.removeAttribute("transform");
      }
    }
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
})();
