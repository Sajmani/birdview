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

// Drives TestNextFirstCommitsOnClick. It opens Show, types a count into the
// First box, and clicks three places in turn: the box, the First button, and
// the choice's own edge outside both. Each submission is recorded with the
// count it would send and then cancelled, so the page stays. The result is
// data-probe on the root.
window.addEventListener("load", function () {
  setTimeout(function () {
    var out = { box: 0, button: 0, edge: 0, sent: [] };
    var d = document.querySelector('details[data-ctl="Show"]');
    d.open = true;
    var f = d.querySelector("form.showfirst"), box = f.querySelector("input.cap");
    var where = "";
    f.addEventListener("submit", function (e) {
      e.preventDefault();
      out[where]++;
      out.sent.push(new FormData(f).get("show"));
    });
    box.value = "7";
    where = "box"; box.click();
    where = "button"; f.querySelector("button").click();
    var r = f.getBoundingClientRect();
    var at = document.elementFromPoint(r.right - 2, r.top + r.height / 2);
    where = "edge"; at.click();
    out.edgeHit = at === f;
    document.documentElement.setAttribute("data-probe", JSON.stringify(out));
  }, 100);
});
