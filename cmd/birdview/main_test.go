package main

import (
	"net/http"
	"net/http/httptest"
	"regexp"
	"strings"
	"testing"

	"birdview/internal/dataset"
)

func testDataset() *dataset.Dataset {
	mk := func(id, common, sci, fam string, bill, head, neck, body, tail, span, mass float64, k dataset.Kind) dataset.Species {
		m := func(v float64) dataset.Metric {
			return dataset.Metric{Value: v, Prov: dataset.Provenance{
				Kind: dataset.Measured, Source: "AVONET v7, sheet AVONET2_eBird", SourceTaxon: sci}}
		}
		mod := func(v float64, s string) dataset.Metric {
			return dataset.Metric{Value: v, Prov: dataset.Provenance{Kind: dataset.Modeled, Model: s}}
		}
		sp := dataset.Species{
			ID: id, CommonName: common, ScientificName: sci, Family: fam,
			FamilyCommon: fam + " family", FamilyColor: "#86efac",
			TotalLength: mod(bill+head+neck+body+tail, "sum of the five diagram segments"),
			Wingspan:    dataset.Metric{Value: span, Prov: dataset.Provenance{Kind: k, Source: "Fox et al. 2026", SourceTaxon: "Other species"}},
			Mass:        m(mass),
			Segments: dataset.Segments{
				Bill: m(bill), Head: mod(head, "cranial vault model"), Neck: mod(neck, "neck ratio for "+fam),
				Body: mod(body, "trunk model"), Tail: m(tail),
			},
		}
		return sp
	}
	return &dataset.Dataset{Species: []dataset.Species{
		mk("calypte_anna", "Anna's Hummingbird", "Calypte anna", "Trochilidae", 18, 12, 6, 40, 26, 120, 4, dataset.Measured),
		mk("turdus_migratorius", "American Robin", "Turdus migratorius", "Turdidae", 20, 18, 12, 90, 85, 390, 77, dataset.Measured),
		mk("ardea_herodias", "Great Blue Heron", "Ardea herodias", "Ardeidae", 140, 40, 200, 300, 180, 1825, 2400, dataset.NearMatch),
	}}
}

func newTestServer(t *testing.T) *server {
	t.Helper()
	s, err := newServer(testDataset())
	if err != nil {
		t.Fatalf("newServer: %v", err)
	}
	return s
}

func get(t *testing.T, s *server, path string) (int, string) {
	t.Helper()
	w := httptest.NewRecorder()
	s.ServeHTTP(w, httptest.NewRequest(http.MethodGet, path, nil))
	return w.Code, w.Body.String()
}

// Verifies: T-001, T-003. Criterion: AC-054.
func TestRoutesServe(t *testing.T) {
	s := newTestServer(t)
	for _, path := range []string{
		"/", "/static/styles.css", "/species/ardea_herodias",
		"/?family=Ardeidae", "/?by=mass&scale=global&units=metric&step=30",
	} {
		if code, body := get(t, s, path); code != http.StatusOK {
			t.Errorf("%s: status %d", path, code)
		} else if len(body) == 0 {
			t.Errorf("%s: empty body", path)
		}
	}
	if code, _ := get(t, s, "/species/nonesuch"); code != http.StatusNotFound {
		t.Errorf("unknown species: status %d, want 404", code)
	}
}

// externalRef finds a resource the browser would fetch from another host.
var externalRef = regexp.MustCompile(`(?:src|href)\s*=\s*"(?:https?:)?//[^"]+`)

// TestNoExternalResources asserts T-002 where it actually bites: a page that
// pulls a stylesheet or script from a CDN is not self-contained, however
// self-contained the binary is. The previous design loaded Tailwind from
// cdn.tailwindcss.com and would have failed this.
//
// Verifies: T-002. Criterion: AC-051.
func TestNoExternalResources(t *testing.T) {
	s := newTestServer(t)
	for _, path := range []string{"/", "/species/ardea_herodias"} {
		_, body := get(t, s, path)
		if m := externalRef.FindAllString(body, -1); len(m) > 0 {
			t.Errorf("%s references external resources: %v", path, m)
		}
	}
}

// Verifies: P-031. Criterion: AC-055.
func TestTileShowsHeaderAndMetrics(t *testing.T) {
	_, body := get(t, newTestServer(t), "/")
	for _, want := range []string{
		"American Robin",     // common name
		"Turdus migratorius", // scientific name
		"Turdidae",           // family badge
		"<svg",               // the diagram
		"L ", "W ", "M ",     // the three main metrics
	} {
		if !strings.Contains(body, want) {
			t.Errorf("tile is missing %q", want)
		}
	}
}

// TestDetailShowsEverySegmentWithProvenance asserts P-034 and P-063: the detail
// view lists all five segments as well as the three main metrics, each with its
// provenance explained.
//
// Verifies: P-034, P-063. Criterion: AC-043.
func TestDetailShowsEverySegmentWithProvenance(t *testing.T) {
	_, body := get(t, newTestServer(t), "/species/ardea_herodias")
	for _, label := range []string{"Total length", "Wingspan", "Mass", "Bill", "Head", "Neck", "Body", "Tail"} {
		if !strings.Contains(body, ">"+label+"<") {
			t.Errorf("detail view omits %q", label)
		}
	}
	for _, want := range []string{
		"neck ratio for Ardeidae",   // the family's model is named (P-045)
		"measured on Other species", // near-match names the borrowed taxon
		"drawing convention",        // the neck disclosure (P-063)
	} {
		if !strings.Contains(body, want) {
			t.Errorf("detail view omits %q", want)
		}
	}
	// It is the same tile, enlarged: the header must be there too.
	for _, want := range []string{"Great Blue Heron", "Ardea herodias", "<svg"} {
		if !strings.Contains(body, want) {
			t.Errorf("detail view is not the tile enlarged; missing %q", want)
		}
	}
}

// Verifies: P-010. Criterion: AC-056.
func TestFamilyFilter(t *testing.T) {
	s := newTestServer(t)
	_, all := get(t, s, "/")
	if !strings.Contains(all, "American Robin") || !strings.Contains(all, "Great Blue Heron") {
		t.Fatal("unfiltered view should show every species")
	}
	_, one := get(t, s, "/?family=Ardeidae")
	if strings.Contains(one, "American Robin") {
		t.Error("filtering to Ardeidae still shows a Turdidae species")
	}
	if !strings.Contains(one, "Great Blue Heron") {
		t.Error("filtering to Ardeidae dropped the heron")
	}
}

// Verifies: P-032. Criterion: AC-057.
func TestUnitToggle(t *testing.T) {
	s := newTestServer(t)
	if _, b := get(t, s, "/"); !strings.Contains(b, " in ") && !strings.Contains(b, " in<") {
		t.Error("imperial is the default but no inch values appear")
	}
	if _, b := get(t, s, "/?units=metric"); !strings.Contains(b, " cm") {
		t.Error("metric toggle produced no centimeter values")
	}
}
