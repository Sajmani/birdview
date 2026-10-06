package taxonomy

import (
	"bufio"
	"fmt"
	"os"
	"sort"
	"strconv"
	"strings"
	"testing"
)

// TestSourceNamesAgree asserts T-028 over the committed correspondence
// table: for every species BirdView draws, the two occurrence sources hold
// comparable numbers of records under the name BirdView serves — and every
// species that breaks that is declared, with which of the four shapes it is.
//
// The band exists because a name can fail to join in two ways and only one
// is visible. A name that MISSES leaves a species at zero. A name that
// MATCHES THE WRONG CONCEPT joins cleanly and stays wrong forever, and no
// string comparison can see it: data/gbif-taxa.tsv files Setophaga petechia
// against itself as `exact`, which is true and useless. Asking each database
// how many records it holds is the only instrument that separates them.
//
// Verifies: T-028. Criterion: AC-105.
func TestSourceNamesAgree(t *testing.T) {
	// 1x to 1000x. The measured distribution is far tighter -- median 48x,
	// 95th percentile 148x -- so the band is deliberately loose: it is
	// looking for the two orders of magnitude that mean a taxonomy
	// disagreement, not for ordinary variation between a deep database and
	// a shallow one.
	const lo, hi = 1.0, 1000.0

	f, err := os.Open("../../data/source-names.tsv")
	if err != nil {
		t.Skipf("no correspondence table: %v (run cmd/birdview-names)", err)
	}
	defer f.Close()

	// Every mechanism that legitimately skews the ratio, from the tables
	// that already describe it. Derived rather than re-listed, so a new
	// rename does not need adding in two places.
	expected := map[string]string{}
	for _, to := range AviListRenames {
		expected[to] = "eBird files it under its pre-rename name"
	}
	for _, to := range LumpedInto {
		expected[to] = "eBird files it under a name AviList does not carry"
	}
	for _, sn := range SharedNames {
		for _, c := range sn.Covers {
			expected[c] = "shares the name " + sn.Source + " with another species"
		}
	}
	for sci, common := range WithheldByEBird {
		expected[sci] = "eBird withholds " + common
	}
	for sci, why := range DomesticForms {
		expected[sci] = why
	}

	var checked, flagged int
	var surprises []string
	sc := bufio.NewScanner(f)
	for sc.Scan() {
		line := sc.Text()
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		p := strings.Split(line, "\t")
		if len(p) != 4 {
			t.Fatalf("malformed row: %q", line)
		}
		sci := p[1]
		eb, _ := strconv.ParseInt(p[2], 10, 64)
		in, _ := strconv.ParseInt(p[3], 10, 64)
		checked++

		// A species absent from one source entirely, or wildly lopsided, is
		// a disagreement about what the name denotes.
		var odd bool
		switch {
		case eb == 0 || in == 0:
			odd = eb+in > 0 // both zero is a bird neither source has, which is fine
		default:
			r := float64(eb) / float64(in)
			odd = r < lo || r > hi
		}
		if !odd {
			continue
		}
		flagged++
		if _, ok := expected[sci]; !ok {
			surprises = append(surprises,
				fmt.Sprintf("%-28s eBird %-10d iNaturalist %-8d", sci, eb, in))
		}
	}
	if err := sc.Err(); err != nil {
		t.Fatal(err)
	}
	if checked < 1000 {
		t.Fatalf("only %d species in the table; it is stale or truncated", checked)
	}
	sort.Strings(surprises)
	for _, s := range surprises {
		t.Errorf("the two sources disagree about this name and nothing says why:\n    %s\n"+
			"    Either eBird files it under another name (add to AviListRenames or "+
			"LumpedInto), it shares a name with another species (SharedNames), eBird "+
			"withholds it (WithheldByEBird), or iNaturalist files a domestic form "+
			"separately (DomesticForms). T-028.", s)
	}
	t.Logf("%d species, %d outside the %.0fx-%.0fx band, all declared", checked, flagged, lo, hi)
}
