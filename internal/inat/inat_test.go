package inat

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

// TestAcceptRequiresBothNames asserts the confidence gate of P-064.
//
// The failure this prevents is silent. Matching on scientific name alone would
// find the right page nearly always, and the rare miss would be a plausible
// *wrong bird* rather than a visible error -- on a page whose whole purpose is
// being careful about what it claims. So the common name is a required second
// opinion, and the test checks it actually rejects.
//
// Verifies: P-064. Criterion: AC-068.
func TestAcceptRequiresBothNames(t *testing.T) {
	cases := []struct {
		name                               string
		wantSci, wantCommon, gotSci, gotCo string
		accept                             bool
	}{
		{"both agree", "Ardea herodias", "Great Blue Heron", "Ardea herodias", "Great Blue Heron", true},
		// The five buckets that left 30 species unlinked, each now a case.
		{"hawaiian okina", "Himatione sanguinea", "Apapane", "Himatione sanguinea", "\u02bbApapane", true},
		{"okina mid-word", "Chlorodrepanis virens", "Hawaii Amakihi", "Chlorodrepanis virens", "Hawai\u02bbi \u02bbAmakihi", true},
		{"ABA parenthetical", "Garrulax canorus", "Chinese Hwamei (Hwamei)", "Garrulax canorus", "Chinese Hwamei", true},
		{"parenthetical mid-name", "Porphyrio poliocephalus", "Gray-headed (Purple) Swamphen", "Porphyrio poliocephalus", "Gray-headed Swamphen", true},
		{"british spelling", "Tringa brevipes", "Gray-tailed Tattler", "Tringa brevipes", "Grey-tailed Tattler", true},
		{"genus reassigned", "Dryobates villosus", "Hairy Woodpecker", "Leuconotopicus villosus", "Hairy Woodpecker", true},
		// A change of genus can rewrite the epithet's gender: Coccothraustes is
		// masculine, Hesperiphona feminine, and it is the same grosbeak.
		{"gender agreement", "Coccothraustes vespertinus", "Evening Grosbeak", "Hesperiphona vespertina", "Evening Grosbeak", true},
		{"gender, other direction", "Hesperiphona vespertina", "Evening Grosbeak", "Coccothraustes vespertinus", "Evening Grosbeak", true},
		// Stems must still be substantial: a two-letter trim off a short
		// epithet is not evidence of anything.
		{"short epithets not stemmed", "Genus alba", "Some Bird", "Other albus", "Some Bird", false},
		// A different genus is accepted ONLY alongside an agreeing common name.
		{"genus differs, common differs", "Dryobates villosus", "Hairy Woodpecker", "Leuconotopicus arizonae", "Arizona Woodpecker", false},
		{"epithet differs too", "Dryobates villosus", "Hairy Woodpecker", "Leuconotopicus borealis", "Hairy Woodpecker", false},
		{"scientific name differs", "Ardea herodias", "Great Blue Heron", "Ardea alba", "Great Egret", false},
		{"common name disagrees", "Ardea herodias", "Great Blue Heron", "Ardea herodias", "Grey Heron", false},
		// The real case this tolerance exists for: ABA hyphenates, iNat does not.
		{"hyphen only", "Vireo swainsoni", "Western Warbling-Vireo", "Vireo swainsoni", "Western Warbling Vireo", true},
		{"case only", "Ardea alba", "Great Egret", "Ardea alba", "GREAT EGRET", true},
		{"extra whitespace", "Ardea alba", "Great Egret", "Ardea alba", "  Great   Egret ", true},
		// A blank common name must not sail through by normalizing to equality.
		{"both common names empty", "Ardea alba", "", "Ardea alba", "", false},
		{"iNat has no common name", "Ardea alba", "Great Egret", "Ardea alba", "", false},
	}
	for _, c := range cases {
		if got := Accept(c.wantSci, c.wantCommon, c.gotSci, []string{c.gotCo}); got != c.accept {
			t.Errorf("%s: Accept(%q,%q,%q,%q) = %v, want %v",
				c.name, c.wantSci, c.wantCommon, c.gotSci, c.gotCo, got, c.accept)
		}
	}
}

// TestAcceptUsesEveryEnglishName covers the case that made 100% coverage
// possible without lowering the bar: iNaturalist prefers "Siberian Tit" for
// Poecile cinctus and ALSO carries ABA's "Gray-headed Chickadee". The
// authorities agree; they disagree only about which name leads.
//
// Verifies: P-064. Criterion: AC-068.
func TestAcceptUsesEveryEnglishName(t *testing.T) {
	all := []string{"Siberian Tit", "Gray-headed Chickadee", "Alaska Chickadee"}
	if !Accept("Poecile cinctus", "Gray-headed Chickadee", "Poecile cinctus", all) {
		t.Error("a name iNaturalist holds but does not prefer was not accepted")
	}
	// Still a real bar: a name it does not hold at all is refused.
	if Accept("Poecile cinctus", "Boreal Chickadee", "Poecile cinctus", all) {
		t.Error("accepted a common name iNaturalist does not carry for this taxon")
	}
	// And the preferred name alone is not special-cased away.
	if !Accept("Poecile cinctus", "Siberian Tit", "Poecile cinctus", all) {
		t.Error("the preferred name stopped matching")
	}
}

// TestTableRoundTripsAndRejectsJunk covers the committed table's format.
//
// Verifies: T-022. Criterion: AC-068.
func TestTableRoundTripsAndRejectsJunk(t *testing.T) {
	in := Table{
		"Ardea herodias": {ScientificName: "Ardea herodias", TaxonID: 4956, CommonName: "Great Blue Heron"},
		"Calypte anna":   {ScientificName: "Calypte anna", TaxonID: 5305, CommonName: "Anna's Hummingbird"},
	}
	var b strings.Builder
	if err := Write(&b, in); err != nil {
		t.Fatalf("Write: %v", err)
	}
	out, err := Load(strings.NewReader(b.String()))
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if len(out) != len(in) {
		t.Fatalf("round trip produced %d entries, want %d", len(out), len(in))
	}
	for k, want := range in {
		if got := out[k]; got != want {
			t.Errorf("%s round-tripped to %+v, want %+v", k, got, want)
		}
	}
	// Sorted output, so a regenerated table diffs as changes rather than a reshuffle.
	if a, c := strings.Index(b.String(), "Ardea"), strings.Index(b.String(), "Calypte"); a > c {
		t.Error("Write did not sort; regenerating the table would produce an unreadable diff")
	}

	for _, bad := range []struct{ name, text string }{
		{"too few fields", "Ardea herodias\t4956\n"},
		{"non-numeric id", "Ardea herodias\tfour\tGreat Blue Heron\n"},
		{"zero id", "Ardea herodias\t0\tGreat Blue Heron\n"},
		{"negative id", "Ardea herodias\t-3\tGreat Blue Heron\n"},
		{"duplicate name", "Ardea herodias\t1\tA\nArdea herodias\t2\tB\n"},
	} {
		if _, err := Load(strings.NewReader(bad.text)); err == nil {
			t.Errorf("%s: Load accepted %q", bad.name, bad.text)
		}
	}
	// Comments and blanks are skipped, not treated as malformed.
	if tbl, err := Load(strings.NewReader("# a comment\n\nArdea alba\t7\tGreat Egret\n")); err != nil {
		t.Errorf("Load rejected a comment or blank line: %v", err)
	} else if len(tbl) != 1 {
		t.Errorf("Load read %d entries past the comment, want 1", len(tbl))
	}
}

// TestTaxonURL pins the link format P-064 specifies.
//
// Verifies: P-064. Criterion: AC-068.
func TestTaxonURL(t *testing.T) {
	if got, want := TaxonURL(4956), "https://www.inaturalist.org/taxa/4956"; got != want {
		t.Errorf("TaxonURL(4956) = %q, want %q", got, want)
	}
}

// fakeAPI serves one canned /v1/taxa response and records how it was called, so
// the courtesy requirements can be observed rather than assumed. Tests never
// touch the live API (tech.md bindings).
func fakeAPI(t *testing.T, body string) (*httptest.Server, *[]time.Time, *[]string) {
	t.Helper()
	var times []time.Time
	var agents []string
	s := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		times = append(times, time.Now())
		agents = append(agents, r.Header.Get("User-Agent"))
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprint(w, body)
	}))
	t.Cleanup(s.Close)
	return s, &times, &agents
}

const heronJSON = `{"results":[
  {"id":4956,"name":"Ardea herodias","rank":"species","preferred_common_name":"Great Blue Heron"}]}`

// TestLookupHonorsTheSourceRequirements checks the two obligations the adopted
// source actually imposes: identify yourself (inat-api/R4, T-006) and pace the
// requests (inat-api/R1, T-022).
//
// Both are observed at the server, not read off the client's own fields -- a
// client that sets Interval and never waits would satisfy the latter.
//
// Verifies: T-022, T-006. Criterion: AC-068.
func TestLookupHonorsTheSourceRequirements(t *testing.T) {
	srv, times, agents := fakeAPI(t, heronJSON)

	const interval = 120 * time.Millisecond
	c := NewClient()
	c.BaseURL, c.Interval, c.HTTP = srv.URL, interval, srv.Client()

	for i := 0; i < 3; i++ {
		if _, ok, err := c.Lookup(context.Background(), "Ardea herodias", "Great Blue Heron"); err != nil || !ok {
			t.Fatalf("lookup %d: ok=%v err=%v", i, ok, err)
		}
	}

	for i, ua := range *agents {
		if !strings.Contains(ua, "birdview") {
			t.Errorf("request %d sent User-Agent %q, which does not identify the client", i, ua)
		}
		if !strings.Contains(ua, "@") {
			t.Errorf("request %d sent User-Agent %q, with no contact address", i, ua)
		}
	}
	// A small tolerance, because the assertion is "paced at about the interval"
	// and neither Sleep nor a round trip is exact to the microsecond. Without
	// it this fails a few times in a hundred on a gap like 119.8ms, which is
	// scheduler jitter rather than a client hammering the service.
	const jitter = 5 * time.Millisecond
	for i := 1; i < len(*times); i++ {
		if gap := (*times)[i].Sub((*times)[i-1]); gap < interval-jitter {
			t.Errorf("requests %d and %d were %v apart, closer than the %v the source asks for",
				i-1, i, gap, interval)
		}
	}
}

// TestLookupRejectsAWrongBird is the gate at the network boundary: iNaturalist
// answering with a different species must produce no entry, not a link to it.
//
// Verifies: P-064. Criterion: AC-068.
func TestLookupRejectsAWrongBird(t *testing.T) {
	// A plausible near miss -- right genus, wrong species, and a common name a
	// careless matcher might wave through.
	srv, _, _ := fakeAPI(t, `{"results":[
	  {"id":999,"name":"Ardea alba","rank":"species","preferred_common_name":"Great Egret"}]}`)
	c := NewClient()
	c.BaseURL, c.Interval, c.HTTP = srv.URL, 0, srv.Client()

	if e, ok, err := c.Lookup(context.Background(), "Ardea herodias", "Great Blue Heron"); err != nil {
		t.Fatalf("Lookup: %v", err)
	} else if ok {
		t.Errorf("accepted taxon %d (%s) for Ardea herodias -- a wrong bird is worse than no link",
			e.TaxonID, e.CommonName)
	}

	// A genus-rank hit must not be accepted either.
	srv2, _, _ := fakeAPI(t, `{"results":[
	  {"id":3,"name":"Ardea herodias","rank":"genus","preferred_common_name":"Great Blue Heron"}]}`)
	c2 := NewClient()
	c2.BaseURL, c2.Interval, c2.HTTP = srv2.URL, 0, srv2.Client()
	if _, ok, err := c2.Lookup(context.Background(), "Ardea herodias", "Great Blue Heron"); err != nil {
		t.Fatalf("Lookup: %v", err)
	} else if ok {
		t.Error("accepted a genus-rank taxon; only a species may be linked")
	}
}

// TestLookupReportsTransportErrorsSeparately: a 500 is not a confident
// non-match, and must not be recorded as one. The mapping tool needs the
// difference so a rerun can fill the gap instead of baking in a missing link.
//
// Verifies: T-022. Criterion: AC-068.
func TestLookupReportsTransportErrorsSeparately(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "nope", http.StatusInternalServerError)
	}))
	defer srv.Close()

	c := NewClient()
	c.BaseURL, c.Interval, c.HTTP = srv.URL, 0, srv.Client()
	_, ok, err := c.Lookup(context.Background(), "Ardea herodias", "Great Blue Heron")
	if err == nil {
		t.Error("a 500 returned no error; the tool would record a confident non-match")
	}
	if ok {
		t.Error("a 500 returned ok")
	}
}
