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

// TestSourceNamesAgree asserts tech.sourceNamingMeasured over the committed correspondence
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
// Verifies: tech.sourceNamingMeasured. Criterion: acceptance.disagreementsDeclared.
func TestSourceNamesAgree(t *testing.T) {
	// Two asymmetric rules, because the two databases are not peers:
	// iNaturalist is uniformly about 48x shallower than eBird, so what
	// counts as surprising differs by direction.
	//
	//   - iNaturalist OUT-recording eBird is always wrong, whichever name
	//     is at fault. A 3x margin is required because a 48x expectation
	//     inverted threefold is a 144x departure, where 27 against 26 is
	//     two vagrant sightings and means nothing. Lesser Frigatebird sits
	//     exactly there and is correctly ignored.
	//
	//   - eBird DWARFING iNaturalist is ordinary at every scale, so it only
	//     signals with volume behind it: 10,000 records, where iNaturalist
	//     should hold about 200.
	//
	// The floors are what make this a check rather than a noise generator.
	// Without them it fires on 45 species, 43 of which are vagrants with a
	// handful of records apiece where the shallower database happens to
	// hold none -- and a check that cries wolf gets deleted.
	const (
		minINat   = 20
		iNatMulti = 3
		minEBird  = 10000
		maxRatio  = 1000.0
	)

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
	for _, c := range CountedAs {
		to := c.Species
		expected[to] = "eBird files it under a name AviList does not carry"
	}
	for _, sn := range SharedNames {
		for _, c := range sn.Covers {
			expected[c] = "shares the name " + sn.Source + " with another species"
		}
	}
	for sci := range WithheldByEBird {
		expected[sci] = "eBird withholds it everywhere"
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

		var odd bool
		switch {
		case in >= minINat && eb*iNatMulti < in:
			odd = true
		case eb >= minEBird && (in == 0 || float64(eb)/float64(in) > maxRatio):
			odd = true
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
			"CountedAs), it shares a name with another species (SharedNames), eBird "+
			"withholds it (WithheldByEBird), or iNaturalist files a domestic form "+
			"separately (DomesticForms). tech.sourceNamingMeasured.", s)
	}
	t.Logf("%d species, %d flagged as a source disagreement, all declared", checked, flagged)
}
