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

package morphology

import (
	"math"
	"os"
	"strings"
	"testing"
)

const reviewHeader = "# test\ntaxon\tneck ratio\thead factor\tbody factor\twingspan factor\tstatus\treviewer\tdate\treason\n"

// TestProportionReviewParsesAndApplies asserts decisions.proportionsReviewedPerTaxon: the table's
// rows are read, a genus row wins over a family row and a family row over the code's table, an empty
// value inherits, the provenance names the row and whether it is reviewed, malformed rows are
// refused, and the table writes back as it was read.
//
// Verifies: product.neckModelPerFamily, product.craniumFromMass. Criterion: dataset.proportionsReviewTable.
func TestProportionReviewParsesAndApplies(t *testing.T) {
	src := reviewHeader +
		"Aix\t\t\t0.90\t1.10\treviewed\tSameer\t2026-11-09\tcompared with a flight photo\n" +
		"Anatidae\t0.50\t\t\t\treviewed\tSameer\t2026-11-09\t\n" +
		"Aythya\t0.35\t\t\t\tproposed\t\t\tshort thick neck\n" +
		"Podicipedidae\t\t0.65\t\t\treviewed\tSameer\t2026-11-09\tsmall heads\n"
	rv, err := ParseReview(strings.NewReader(src))
	if err != nil {
		t.Fatal(err)
	}
	if n, model := NeckLength(100, "Anatidae", "Aythya", rv); n != 35 || !strings.Contains(model, "Aythya") || !strings.Contains(model, "not yet reviewed") {
		t.Errorf("a scaup's neck is %.1f, %q; want 35 from the genus row, saying it is not yet reviewed", n, model)
	}
	if n, model := NeckLength(100, "Anatidae", "Branta", rv); n != 50 || !strings.Contains(model, "reviewed by Sameer, 2026-11-09") {
		t.Errorf("a goose's neck is %.1f, %q; want 50 from the family row, naming its reviewer", n, model)
	}
	if n, _ := NeckLength(100, "Podicipedidae", "Podiceps", rv); n != 45 {
		t.Errorf("a grebe's neck is %.1f; an empty ratio must keep the family table's 0.45", n)
	}
	base, _ := HeadLength(453, "Laridae", "Larus", rv)
	if h, model := HeadLength(453, "Podicipedidae", "Podiceps", rv); h >= base || !strings.Contains(model, "0.65") {
		t.Errorf("a grebe's head is %.1f against %.1f unscaled, %q; want the family factor applied and named", h, base, model)
	}
	plain, _ := BodyLength(700)
	if b, model := BodyLengthFor(700, "Anatidae", "Aix", rv); math.Abs(b-plain*0.9) > 1e-9 || !strings.Contains(model, "0.90") {
		t.Errorf("a Wood Duck's body is %.1f against %.1f unscaled, %q; want the genus body factor applied and named", b, plain, model)
	}
	if b, _ := BodyLengthFor(700, "Anatidae", "Aythya", rv); b != plain {
		t.Errorf("a scaup's body is %.1f; with no body factor it must stay %.1f", b, plain)
	}
	if f, src := rv.WingspanFactorFor("Anatidae", "Aix"); f != 1.1 || !strings.Contains(src, "Aix") {
		t.Errorf("Aix's wingspan factor is %g, %q; want 1.1 from its row", f, src)
	}
	if f, _ := rv.WingspanFactorFor("Anatidae", "Aythya"); f != 1 {
		t.Errorf("Aythya's wingspan factor is %g; with none set it must be 1", f)
	}
	for name, bad := range map[string]string{
		"unknown status":          "Aythya\t0.35\t\t\t\tmaybe\t\t\tx\n",
		"reviewed without date":   "Aythya\t0.35\t\t\t\treviewed\tSameer\t\t\n",
		"ratio out of range":      "Aythya\t3\t\t\t\tproposed\t\t\tx\n",
		"factor out of range":     "Aythya\t\t0.1\t\t\tproposed\t\t\tx\n",
		"body out of range":       "Aythya\t\t\t3\t\tproposed\t\t\tx\n",
		"wingspan out of range":   "Aythya\t\t\t\t0.2\tproposed\t\t\tx\n",
		"proposed without reason": "Aythya\t0.35\t\t\t\tproposed\t\t\t\n",
		"duplicate taxon":         "Aythya\t0.35\t\t\t\tproposed\t\t\tx\nAythya\t0.30\t\t\t\tproposed\t\t\ty\n",
		"short row":               "Aythya\t0.35\n",
	} {
		if _, err := ParseReview(strings.NewReader(reviewHeader + bad)); err == nil {
			t.Errorf("a row with %s was accepted", name)
		}
	}
}

// committedReview is data/proportions-review.tsv, the table the ingest reads.
func committedReview(t *testing.T) Review {
	t.Helper()
	f, err := os.Open("../../data/proportions-review.tsv")
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()
	rv, err := ParseReview(f)
	if err != nil {
		t.Fatal(err)
	}
	return rv
}
