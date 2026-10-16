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

// Package inat resolves BirdView species to iNaturalist taxon IDs and turns
// them into links.
//
// BirdView links to iNaturalist and copies nothing from it: no photograph is
// embedded, hotlinked, cached or reproduced, so no contributor's CC BY-NC
// license is engaged (inat-terms/R1, decisions.inaturalistLinkAdopted). tech.selfContainedServer forbids the embedding
// independently.
//
// The mapping itself is a committed table rather than a build-time fetch --
// see cmd/birdview-inatmap and tech.pacedInaturalistMapping.
//
// Verifies: product.inaturalistLink, tech.pacedInaturalistMapping.
package inat

import (
	"bufio"
	"fmt"
	"io"
	"regexp"
	"sort"
	"strconv"
	"strings"
)

// TaxonURL is where a taxon ID resolves on the public site.
func TaxonURL(id int) string {
	return "https://www.inaturalist.org/taxa/" + strconv.Itoa(id)
}

// Entry is one resolved species.
type Entry struct {
	ScientificName string
	TaxonID        int
	CommonName     string // as iNaturalist returned it, for the audit trail
}

// Table maps a BirdView scientific name to its iNaturalist taxon.
type Table map[string]Entry

// parenthetical matches an ABA alternate name, e.g. the "(Hwamei)" in
// "Chinese Hwamei (Hwamei)". The checklist's own header explains these: "where
// discrepancies in names occur between Clements/eBird and the AOS, both are
// listed, that of AOS in parentheses."
var parenthetical = regexp.MustCompile(`\s*\([^)]*\)`)

// normalize reduces a common name to the form the two authorities can be
// compared in: lowercase, British spelling folded, and ASCII letters and digits
// only, with every separator DROPPED rather than collapsed to a space.
//
// Two decisions here, both learned from real failures.
//
// ASCII-only, rather than unicode.IsLetter: the Hawaiian okina is U+02BB, whose
// Unicode category is Lm -- a LETTER -- so IsLetter keeps it and "ʻApapane"
// never matches "Apapane". Fifteen Hawaiian species went unlinked for that.
//
// Dropped rather than collapsed: an okina sits INSIDE a word, so turning it
// into a space splits "Hawaiʻi" into "hawai i" and the mismatch survives in a
// new form. Dropping every separator makes word boundaries irrelevant on both
// sides, which also absorbs the hyphen in "Western Warbling-Vireo". The cost is
// that "Graybird" and "Gray bird" would compare equal; the benefit is that no
// punctuation convention can break a match, and the scientific name is checked
// independently anyway.
func normalize(s string) string {
	s = strings.ToLower(s)
	s = strings.ReplaceAll(s, "grey", "gray") // Grey-tailed Tattler vs Gray-tailed

	var b strings.Builder
	for _, r := range s {
		if (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9') {
			b.WriteRune(r)
		}
	}
	return b.String()
}

// commonVariants returns the forms of a common name worth comparing: the whole
// thing, and the same with any ABA alternate in parentheses removed.
func commonVariants(s string) []string {
	out := []string{normalize(s)}
	if stripped := parenthetical.ReplaceAllString(s, ""); stripped != s {
		if n := normalize(stripped); n != "" && n != out[0] {
			out = append(out, n)
		}
	}
	return out
}

// gendered are the Latin adjectival endings that change when a species moves
// genus, longest first so that "rum" is tried before "um".
//
// A specific epithet agrees in grammatical gender with its genus, so a generic
// reassignment can rewrite it: Coccothraustes is masculine and takes
// vespertinus, Hesperiphona is feminine and takes vespertina. Same bird, same
// epithet, different ending.
var gendered = []string{"rum", "um", "us", "is", "ra", "er", "a", "e"}

// stem strips a gendered ending, so epithets can be compared across a change of
// genus. Short results are left alone: trimming two letters off a five-letter
// epithet leaves too little to be evidence of anything.
func stem(epithet string) string {
	for _, suf := range gendered {
		if len(epithet) > len(suf)+3 && strings.HasSuffix(epithet, suf) {
			return epithet[:len(epithet)-len(suf)]
		}
	}
	return epithet
}

// sameSpecies reports whether two scientific names denote the same species:
// identical, or differing only in genus while sharing a species epithet -- up
// to the gender agreement that a change of genus forces.
//
// On its own an epithet match would be far too loose; it is only ever accepted
// alongside an agreeing common name, which is what makes it safe. Dryobates
// villosus links to Leuconotopicus villosus because both are the Hairy
// Woodpecker, and does not link to Leuconotopicus borealis.
func sameSpecies(want, got string) bool {
	want, got = strings.TrimSpace(want), strings.TrimSpace(got)
	if strings.EqualFold(want, got) {
		return true
	}
	w, g := strings.Fields(want), strings.Fields(got)
	if len(w) != 2 || len(g) != 2 {
		return false
	}
	we, ge := strings.ToLower(w[1]), strings.ToLower(g[1])
	return we == ge || stem(we) == stem(ge)
}

// Accept reports whether an iNaturalist result is a confident enough match for
// a BirdView species to link to (product.inaturalistLink).
//
// Both names must agree. The scientific name identifies the species, exactly or
// through a generic reassignment; the common name is the independent second
// opinion that stops a near miss becoming a link to the wrong bird. Neither
// alone is sufficient, and that is the whole design: a plausible wrong bird is
// worse than no link on a page about being careful what you claim.
//
// gotCommons is EVERY English name iNaturalist holds for the taxon, not just
// its preferred one. That is more evidence at the same bar, not a lower bar:
// iNaturalist prefers "Siberian Tit" for Poecile cinctus and also carries ABA's
// "Gray-headed Chickadee", and the two authorities do agree -- they just
// disagree about which name leads.
func Accept(wantSci, wantCommon, gotSci string, gotCommons []string) bool {
	if !sameSpecies(wantSci, gotSci) {
		return false
	}
	want := commonVariants(wantCommon)
	if len(want) == 0 || want[0] == "" {
		return false
	}
	for _, gc := range gotCommons {
		for _, g := range commonVariants(gc) {
			if g == "" {
				continue
			}
			for _, w := range want {
				if w == g {
					return true
				}
			}
		}
	}
	return false
}

// Load reads a mapping table. The format is one tab-separated record per line,
// `scientific name<TAB>taxon id<TAB>common name as iNaturalist returned it`,
// with `#` comments and blank lines ignored. Plain text so a human can read the
// diff when the table is refreshed.
func Load(r io.Reader) (Table, error) {
	t := Table{}
	sc := bufio.NewScanner(r)
	for line := 1; sc.Scan(); line++ {
		s := strings.TrimSpace(sc.Text())
		if s == "" || strings.HasPrefix(s, "#") {
			continue
		}
		f := strings.Split(s, "\t")
		if len(f) != 3 {
			return nil, fmt.Errorf("inat table line %d: got %d fields, want 3", line, len(f))
		}
		id, err := strconv.Atoi(f[1])
		if err != nil {
			return nil, fmt.Errorf("inat table line %d: taxon id %q: %w", line, f[1], err)
		}
		if id <= 0 {
			return nil, fmt.Errorf("inat table line %d: taxon id %d is not positive", line, id)
		}
		if _, dup := t[f[0]]; dup {
			return nil, fmt.Errorf("inat table line %d: %q appears twice", line, f[0])
		}
		t[f[0]] = Entry{ScientificName: f[0], TaxonID: id, CommonName: f[2]}
	}
	return t, sc.Err()
}

// Write serializes a table in Load's format, sorted, so that regenerating it
// produces a reviewable diff rather than a reshuffle.
func Write(w io.Writer, t Table) error {
	names := make([]string, 0, len(t))
	for k := range t {
		names = append(names, k)
	}
	sort.Strings(names)

	if _, err := fmt.Fprint(w,
		"# BirdView -> iNaturalist taxon mapping. Generated by cmd/birdview-inatmap (tech.pacedInaturalistMapping).\n"+
			"# scientific name<TAB>taxon id<TAB>common name as iNaturalist returned it\n"+
			"#\n"+
			"# A species appears here only when its scientific name matched exactly AND the\n"+
			"# common name agreed with ABA's once punctuation was normalized (product.inaturalistLink, decisions.inaturalistLinkAdopted).\n"+
			"# Absent means no confident match: the detail page then shows no link.\n"); err != nil {
		return err
	}
	for _, n := range names {
		e := t[n]
		if _, err := fmt.Fprintf(w, "%s\t%d\t%s\n", n, e.TaxonID, e.CommonName); err != nil {
			return err
		}
	}
	return nil
}
