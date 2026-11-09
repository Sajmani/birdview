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

package main

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"testing"
	"time"

	"birdview/internal/dataset"
)

// reviewServer is the test server with review mode on, over a copy of a small table.
func reviewServer(t *testing.T) (*server, string) {
	t.Helper()
	path := filepath.Join(t.TempDir(), "proportions-review.tsv")
	table := "# a test table\ntaxon\tneck ratio\thead factor\tbody factor\twingspan factor\tstatus\treviewer\tdate\treason\n" +
		"Ardea\t0.80\t\t\t\tproposed\t\t\tlong-necked heron\n"
	if err := os.WriteFile(path, []byte(table), 0o644); err != nil {
		t.Fatal(err)
	}
	s := newTestServer(t)
	// The heron's wingspan stands for a modeled one, so its page offers the wingspan slider.
	for i := range s.ds.Species {
		if s.ds.Species[i].ID == "greblu" {
			s.ds.Species[i].Wingspan.Prov = dataset.Provenance{Kind: dataset.Modeled, Model: "a fitted model"}
		}
	}
	if err := s.enableReview(path, "Tester"); err != nil {
		t.Fatal(err)
	}
	return s, path
}

// local sends a request from the machine itself, or from elsewhere.
func local(s *server, method, target string, form url.Values, from string) *httptest.ResponseRecorder {
	var r *http.Request
	if form != nil {
		r = httptest.NewRequest(method, target, strings.NewReader(form.Encode()))
		r.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	} else {
		r = httptest.NewRequest(method, target, nil)
	}
	r.RemoteAddr = from + ":5555"
	w := httptest.NewRecorder()
	s.ServeHTTP(w, r)
	return w
}

// TestReviewModeIsLocalAndOptIn asserts tech.proportionsReviewTool's boundary: without -review there
// is no review page; with it, only the machine itself is answered; and the server refuses review
// mode beside a canonical host, or without a reviewer.
//
// Verifies: tech.proportionsReviewTool. Criterion: acceptance.proportionsReviewTool.
func TestReviewModeIsLocalAndOptIn(t *testing.T) {
	if w := local(newTestServer(t), "GET", "/review", nil, "127.0.0.1"); w.Code != http.StatusNotFound {
		t.Errorf("without -review, /review answered %d; want 404", w.Code)
	}
	s, _ := reviewServer(t)
	for _, from := range []string{"192.0.2.1", "203.0.113.9"} {
		if w := local(s, "GET", "/review", nil, from); w.Code != http.StatusForbidden {
			t.Errorf("/review from %s answered %d; want 403", from, w.Code)
		}
		if w := local(s, "POST", "/review/save", url.Values{"genus": {"Ardea"}, "choice": {"current"}}, from); w.Code != http.StatusForbidden {
			t.Errorf("a save from %s answered %d; want 403", from, w.Code)
		}
	}
	for _, from := range []string{"127.0.0.1", "[::1]"} {
		w := local(s, "GET", "/review", nil, from)
		if w.Code != http.StatusOK || !strings.Contains(w.Body.String(), "Ardea") {
			t.Errorf("/review from %s answered %d; want the queue, naming Ardea", from, w.Code)
		}
	}
	for c, ok := range map[[3]string]bool{
		{"", "", ""}:              true,
		{"", "", "birdview.site"}: true,
		{"data/proportions-review.tsv", "Sameer", ""}:              true,
		{"data/proportions-review.tsv", "", ""}:                    false,
		{"data/proportions-review.tsv", "Sameer", "birdview.site"}: false,
	} {
		if err := checkReviewFlags(c[0], c[1], c[2]); (err == nil) != ok {
			t.Errorf("checkReviewFlags(%q, %q, %q) = %v; want accepted %v", c[0], c[1], c[2], err, ok)
		}
	}
}

// TestReviewPageAndSave asserts the review tool's page and its writes: a genus's page draws its
// representative bird as it is now and as adjusted, with a slider for every modeled measurement (the
// wingspan only where it is modeled) set to the current values, an overlay that turns to any angle, a
// Commons search for its scientific name, and its eBird page; the figure endpoint redraws at the
// values given, the body slider changing its length; choosing adjusted, current, or the whole family
// writes a reviewed row, a wingspan factor only for a modeled wingspan, with the
// reviewer and today's date, and the next page reads it.
//
// Verifies: tech.proportionsReviewTool, product.neckModelPerFamily. Criterion: acceptance.proportionsReviewTool.
func TestReviewPageAndSave(t *testing.T) {
	s, path := reviewServer(t)
	w := local(s, "GET", "/review?genus=Ardea", nil, "127.0.0.1")
	body := w.Body.String()
	for _, want := range []string{`data-sci="Ardea herodias"`, `name="neck"`, `value="0.80"`, `name="head"`,
		`value="1.00"`, `name="body"`, `name="wing"`, `id="rotate"`, "Great Blue Heron", "ebird.org/species/",
		`id="current"`, `id="adjusted"`, "proposed"} {
		if !strings.Contains(body, want) {
			t.Errorf("the Ardea review page lacks %q", want)
		}
	}
	if body := local(s, "GET", "/review?genus=Uria", nil, "127.0.0.1").Body.String(); strings.Contains(body, `name="wing"`) || !strings.Contains(body, `name="body"`) {
		t.Error("the murre's page, whose wingspan is measured, must offer a body slider and no wingspan slider")
	}
	length := func(q string) float64 {
		t.Helper()
		w := local(s, "GET", "/review/figure?genus=Ardea&"+q, nil, "127.0.0.1")
		_, adjusted, _ := strings.Cut(w.Body.String(), `id="adjusted"`)
		m := regexp.MustCompile(`length (\d+) mm`).FindStringSubmatch(adjusted)
		if m == nil {
			t.Fatalf("the figure at %s gives no length:\n%s", q, w.Body.String())
		}
		v, _ := strconv.ParseFloat(m[1], 64)
		return v
	}
	if short, long := length("neck=0.8&head=1&body=0.8&wing=1"), length("neck=0.8&head=1&body=1.2&wing=1"); long <= short {
		t.Errorf("the figure is %.0f mm long at body 0.8 and %.0f mm at 1.2; the body slider does not redraw it", short, long)
	}
	neck := func(ratio string) float64 {
		t.Helper()
		w := local(s, "GET", "/review/figure?genus=Ardea&neck="+ratio+"&head=1&body=1&wing=1", nil, "127.0.0.1")
		_, adjusted, _ := strings.Cut(w.Body.String(), `id="adjusted"`)
		m := regexp.MustCompile(`neck (\d+(?:\.\d)?) mm`).FindStringSubmatch(adjusted)
		if m == nil {
			t.Fatalf("the figure at neck %s gives no neck length:\n%s", ratio, w.Body.String())
		}
		v, _ := strconv.ParseFloat(m[1], 64)
		return v
	}
	if short, long := neck("0.40"), neck("0.90"); long <= short {
		t.Errorf("the figure's neck is %.1f mm at 0.40 and %.1f mm at 0.90; the slider does not redraw it", short, long)
	}
	today := time.Now().Format("2006-01-02")
	for _, c := range []struct {
		form url.Values
		row  string
	}{
		{url.Values{"genus": {"Ardea"}, "choice": {"adjusted"}, "neck": {"0.90"}, "head": {"0.90"}, "body": {"0.95"}, "wing": {"1.05"}, "reason": {"checked against photos"}},
			"Ardea\t0.90\t0.90\t0.95\t1.05\treviewed\tTester\t" + today + "\tchecked against photos"},
		{url.Values{"genus": {"Uria"}, "choice": {"current"}, "neck": {"0.30"}, "head": {"1.00"}, "body": {"1.00"}},
			"Uria\t0.10\t\t\t\treviewed\tTester\t" + today + "\t"},
		{url.Values{"genus": {"Himantopus"}, "choice": {"adjusted"}, "family": {"on"}, "neck": {"0.50"}, "head": {"1.00"}, "body": {"1.00"}, "wing": {"1.20"}},
			"Recurvirostridae\t0.50\t\t\t\treviewed\tTester\t" + today + "\t"},
	} {
		w := local(s, "POST", "/review/save", c.form, "127.0.0.1")
		if w.Code != http.StatusSeeOther || !strings.HasPrefix(w.Header().Get("Location"), "/review") {
			t.Errorf("saving %v answered %d to %q; want 303 to the next review", c.form, w.Code, w.Header().Get("Location"))
		}
		b, _ := os.ReadFile(path)
		if !strings.Contains(string(b), c.row+"\n") {
			t.Errorf("after saving %v the table reads\n%s\nwant the row %q", c.form, b, c.row)
		}
		if !strings.HasPrefix(string(b), "# a test table\n") {
			t.Errorf("saving dropped the table's header:\n%s", b)
		}
	}
	if body := local(s, "GET", "/review?genus=Ardea", nil, "127.0.0.1").Body.String(); !strings.Contains(body, `value="0.90"`) || !strings.Contains(body, "reviewed by Tester") {
		t.Error("after saving, the Ardea page does not read the new row")
	}
}
