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
	"strings"
	"time"
)

// Withholding is one entry of eBird's Sensitive Species list: a species
// eBird keeps out of its public output in a region, in a window of the year
// (decisions.ebirdSilenceNotAbsence).
type Withholding struct {
	// Species is the AviList name BirdView draws.
	Species string
	// Region is a GADM id, a country prefix ("USA", "CAN") meaning every
	// region of that country, or empty meaning everywhere.
	Region string
	// Where is the region as eBird's list names it, kept so an entry can be
	// checked against the page it came from.
	Where string
	// From and To bound the window, as eBird writes them ("Oct 01"); both
	// empty means the whole year. A window may wrap the year end.
	From, To string
}

// Covers reports whether the entry applies to a GADM region.
func (w Withholding) Covers(gid string) bool {
	switch {
	case w.Region == "":
		return true
	case w.Region == gid:
		return true
	case !strings.Contains(w.Region, "."):
		return strings.HasPrefix(gid, w.Region+".")
	}
	return false
}

// FullMonths reports the months the window covers from their first day to
// their last.
//
// Only whole months. A window ending April 15 leaves half of April's
// records in the export -- Painted Bunting has 12,483 of them in Florida --
// and a month that still holds half its evidence is judged on it rather
// than declared withheld (decisions.ebirdSilenceNotAbsence).
func (w Withholding) FullMonths() [12]bool {
	var out [12]bool
	if w.From == "" {
		for i := range out {
			out[i] = true
		}
		return out
	}
	from, err1 := time.Parse("Jan 02 2006", w.From+" 2001")
	to, err2 := time.Parse("Jan 02 2006", w.To+" 2001")
	if err1 != nil || err2 != nil {
		return out // malformed: covers nothing, and TestEBirdSensitiveIsWellFormed says so
	}
	in := func(d time.Time) bool {
		if !to.Before(from) {
			return !d.Before(from) && !d.After(to)
		}
		return !d.Before(from) || !d.After(to) // wraps the year end
	}
	for m := 0; m < 12; m++ {
		first := time.Date(2001, time.Month(m+1), 1, 0, 0, 0, 0, time.UTC)
		last := first.AddDate(0, 1, -1)
		out[m] = in(first) && in(last)
	}
	return out
}

// EBirdSensitive is every entry of eBird's Sensitive Species list for a
// species BirdView draws, in the US and Canada (decisions.ebirdSilenceNotAbsence). Retrieved
// 2026-10-09 from the page named in the source manifest.
//
// A listing says where eBird withholds a bird, not where the bird is, and
// not even that the withholding reached the export: 26 of the 230 region
// cells these entries name hold eBird records anyway. So nothing reads this
// table on its own -- birdview-regions confirms each entry against the
// records before marking a month withheld.
//
// The entries were generated from the page and every one is checked by
// TestEBirdSensitiveIsWellFormed. A refresh should regenerate the table
// rather than edit it, and re-run birdview-regions.
//
// Verifies: decisions.ebirdSilenceNotAbsence.
var EBirdSensitive = []Withholding{
	// Global: withheld everywhere, all year.
	{Species: "Centrocercus minimus"},
	{Species: "Falco rusticolus"},
	{Species: "Myadestes palmeri"},
	{Species: "Palmeria dolei"},
	{Species: "Pseudonestor xanthophrys"},
	{Species: "Strix nebulosa"},
	{Species: "Strix occidentalis"},
	{Species: "Surnia ulula"},
	{Species: "Tympanuchus pallidicinctus"},

	// Northern Saw-whet Owl
	{Species: "Aegolius acadicus", Region: "USA.21_1", Where: "Maryland-US", From: "Apr 01", To: "Sep 30"},

	// Boreal Owl
	{Species: "Aegolius funereus", Region: "USA.24_1", Where: "Minnesota-US"},
	{Species: "Aegolius funereus", Region: "USA.32_1", Where: "New Mexico-US"},
	{Species: "Aegolius funereus", Region: "USA.50_1", Where: "Wisconsin-US"},
	{Species: "Aegolius funereus", Region: "USA.51_1", Where: "Wyoming-US", From: "Feb 01", To: "Aug 31"},

	// Golden Eagle
	{Species: "Aquila chrysaetos", Region: "USA.20_1", Where: "Maine-US", From: "Apr 10", To: "Sep 25"},
	{Species: "Aquila chrysaetos", Region: "USA.33_1", Where: "New York-US", From: "Apr 15", To: "Sep 15"},

	// Short-eared Owl
	{Species: "Asio flammeus", Region: "USA.21_1", Where: "Maryland-US", From: "May 01", To: "Sep 30"},
	{Species: "Asio flammeus", Region: "USA.33_1", Where: "New York-US", From: "Apr 15", To: "Jul 15"},
	{Species: "Asio flammeus", Region: "USA.39_1", Where: "Pennsylvania-US", From: "Apr 20", To: "Jun 30"},

	// Long-eared Owl
	{Species: "Asio otus", Region: "CAN.2_1", Where: "British Columbia-CA", From: "Jan 01", To: "Apr 30"},
	{Species: "Asio otus", Region: "CAN.2_1", Where: "British Columbia-CA", From: "Oct 01", To: "Dec 31"},
	{Species: "Asio otus", Region: "USA", Where: "United States"},

	// American Goshawk
	{Species: "Astur atricapillus", Region: "CAN.9_1", Where: "Ontario-CA", From: "Mar 01", To: "Jul 31"},
	{Species: "Astur atricapillus", Region: "USA.21_1", Where: "Maryland-US", From: "Apr 01", To: "Aug 31"},
	{Species: "Astur atricapillus", Region: "USA.33_1", Where: "New York-US", From: "Mar 15", To: "Aug 15"},
	{Species: "Astur atricapillus", Region: "USA.39_1", Where: "Pennsylvania-US", From: "Mar 01", To: "Aug 31"},
	{Species: "Astur atricapillus", Region: "USA.50_1", Where: "Wisconsin-US", From: "Mar 01", To: "Aug 01"},

	// Burrowing Owl
	{Species: "Athene cunicularia", Region: "CAN.12_1", Where: "Saskatchewan-CA"},
	{Species: "Athene cunicularia", Region: "CAN.1_1", Where: "Alberta-CA"},
	{Species: "Athene cunicularia", Region: "CAN.2_1", Where: "British Columbia-CA"},
	{Species: "Athene cunicularia", Region: "CAN.3_1", Where: "Manitoba-CA"},
	{Species: "Athene cunicularia", Region: "USA.24_1", Where: "Minnesota-US", From: "Apr 15", To: "Aug 31"},

	// Spruce Grouse
	{Species: "Canachites canadensis", Region: "USA.33_1", Where: "New York-US"},
	{Species: "Canachites canadensis", Region: "USA.46_1", Where: "Vermont-US"},

	// Greater Sage-Grouse
	{Species: "Centrocercus urophasianus", Region: "CAN", Where: "Canada"},
	{Species: "Centrocercus urophasianus", Region: "USA", Where: "United States", From: "Feb 01", To: "May 31"},

	// Henslow's Sparrow
	{Species: "Centronyx henslowii", Region: "CAN.9_1", Where: "Ontario-CA", From: "May 15", To: "Aug 15"},

	// Piping Plover
	{Species: "Charadrius melodus", Region: "USA.50_1", Where: "Wisconsin-US"},

	// Sedge Wren
	{Species: "Cistothorus stellaris", Region: "USA.39_1", Where: "Pennsylvania-US", From: "Jun 01", To: "Aug 20"},

	// Yellow-billed Cuckoo
	{Species: "Coccyzus americanus", Region: "USA.13_1", Where: "Idaho-US"},

	// Common Loon
	{Species: "Gavia immer", Region: "USA.51_1", Where: "Wyoming-US", From: "May 01", To: "Sep 30"},

	// Ferruginous Pygmy-Owl
	{Species: "Glaucidium brasilianum", Region: "USA.3_1", Where: "Arizona-US"},
	{Species: "Glaucidium brasilianum", Region: "USA.44_1", Where: "Texas-US"},
	{Species: "Glaucidium brasilianum", Region: "USA.5_1", Where: "California-US"},

	// Northern Pygmy-Owl
	{Species: "Glaucidium gnoma", Region: "USA.51_1", Where: "Wyoming-US", From: "Feb 01", To: "Aug 31"},

	// Whooping Crane
	{Species: "Grus americana", Region: "USA.11_1", Where: "Georgia-US"},
	{Species: "Grus americana", Region: "USA.15_1", Where: "Indiana-US"},
	{Species: "Grus americana", Region: "USA.18_1", Where: "Kentucky-US"},
	{Species: "Grus americana", Region: "USA.19_1", Where: "Louisiana-US"},
	{Species: "Grus americana", Region: "USA.28_1", Where: "Nebraska-US"},
	{Species: "Grus americana", Region: "USA.44_1", Where: "Texas-US", From: "Apr 01", To: "Sep 01"},
	{Species: "Grus americana", Region: "USA.50_1", Where: "Wisconsin-US"},

	// Loggerhead Shrike
	{Species: "Lanius ludovicianus", Region: "USA.33_1", Where: "New York-US", From: "Apr 15", To: "Aug 01"},
	{Species: "Lanius ludovicianus", Region: "USA.39_1", Where: "Pennsylvania-US", From: "May 10", To: "Jul 31"},
	{Species: "Lanius ludovicianus", Region: "USA.50_1", Where: "Wisconsin-US", From: "Apr 15", To: "Aug 15"},

	// Black Rail
	{Species: "Laterallus jamaicensis", Region: "CAN", Where: "Canada"},
	{Species: "Laterallus jamaicensis", Region: "USA", Where: "United States"},

	// Swainson's Warbler
	{Species: "Limnothlypis swainsonii", Region: "USA.21_1", Where: "Maryland-US", From: "Apr 10", To: "Aug 20"},

	// Elf Owl
	{Species: "Micrathene whitneyi", Region: "USA.5_1", Where: "California-US"},

	// Painted Bunting
	{Species: "Passerina ciris", Region: "USA.10_1", Where: "Florida-US", From: "Jan 01", To: "Apr 15"},
	{Species: "Passerina ciris", Region: "USA.10_1", Where: "Florida-US", From: "Oct 01", To: "Dec 31"},

	// King Rail
	{Species: "Rallus elegans", Region: "CAN.9_1", Where: "Ontario-CA"},
	{Species: "Rallus elegans", Region: "USA.14_1", Where: "Illinois-US", From: "May 15", To: "Aug 01"},
	{Species: "Rallus elegans", Region: "USA.33_1", Where: "New York-US", From: "Mar 15", To: "Aug 15"},
	{Species: "Rallus elegans", Region: "USA.39_1", Where: "Pennsylvania-US", From: "May 15", To: "Aug 20"},
	{Species: "Rallus elegans", Region: "USA.50_1", Where: "Wisconsin-US", From: "May 15", To: "Aug 01"},

	// Kirtland's Warbler
	{Species: "Setophaga kirtlandii", Region: "CAN.9_1", Where: "Ontario-CA", From: "May 20", To: "Aug 15"},
	{Species: "Setophaga kirtlandii", Region: "USA.50_1", Where: "Wisconsin-US"},

	// Greater Prairie-Chicken
	{Species: "Tympanuchus cupido", Region: "USA.14_1", Where: "Illinois-US"},

	// Sharp-tailed Grouse
	{Species: "Tympanuchus phasianellus", Region: "CAN.2_1", Where: "British Columbia-CA"},
	{Species: "Tympanuchus phasianellus", Region: "USA.48_1", Where: "Washington-US", From: "Feb 01", To: "May 31"},
	{Species: "Tympanuchus phasianellus", Region: "USA.50_1", Where: "Wisconsin-US"},

	// American Barn Owl
	{Species: "Tyto furcata", Region: "CAN", Where: "Canada"},
	{Species: "Tyto furcata", Region: "USA.21_1", Where: "Maryland-US"},
	{Species: "Tyto furcata", Region: "USA.33_1", Where: "New York-US"},
	{Species: "Tyto furcata", Region: "USA.39_1", Where: "Pennsylvania-US"},
	{Species: "Tyto furcata", Region: "USA.50_1", Where: "Wisconsin-US"},
}

// Withholdings returns the entries for a species that apply to a region.
func Withholdings(species, gid string) []Withholding {
	var out []Withholding
	for _, w := range EBirdSensitive {
		if w.Species == species && w.Covers(gid) {
			out = append(out, w)
		}
	}
	return out
}

// Confirmation thresholds (decisions.ebirdSilenceNotAbsence). Neither is load-bearing for the cases
// that motivated them -- Painted Bunting in Florida runs at zero inside its
// window against thousands a month outside -- and both are recorded so the
// borderline cases have a stated rule rather than a judgement call.
const (
	// withheldRate is how far eBird's monthly rate inside a seasonal window
	// must fall, as a share of its rate in the same region's other months.
	withheldRate = 0.05
	// outRecords is how many times eBird's count iNaturalist must hold for
	// a whole-year listing to count as applied: acceptance.disagreementsDeclared's existing test for
	// a source that out-records eBird, applied to one region.
	outRecords = 3
)

// ConfirmedMonths reports the months in which eBird is withholding a species
// in a region: named by an entry of its Sensitive Species list AND borne out
// by the records (decisions.ebirdSilenceNotAbsence).
//
// The list says where eBird intends to withhold; the export says whether it
// did. They disagree for 26 of the 230 region cells the US and Canada
// entries name -- Greater Sage-Grouse in Utah from February to May holds 422
// eBird records a month inside the window and 61 outside -- and marking
// those would put "withheld by eBird" on records eBird published.
//
// ebird and inat are the species' monthly counts in the region from each
// source, January first.
func ConfirmedMonths(entries []Withholding, ebird, inat [12]int64) [12]bool {
	var out [12]bool
	var eTotal, iTotal int64
	for m := 0; m < 12; m++ {
		eTotal += ebird[m]
		iTotal += inat[m]
	}
	for _, w := range entries {
		in := w.FullMonths()
		var inside, outside int64
		var nIn, nOut int
		for m := 0; m < 12; m++ {
			if in[m] {
				inside += ebird[m]
				nIn++
			} else {
				outside += ebird[m]
				nOut++
			}
		}
		if nIn == 0 {
			continue
		}
		var ok bool
		if w.From != "" && outside > 0 {
			// Seasonal, with eBird's own other months to compare against:
			// the comparison needs no second source and no effort model,
			// because the same observers in the same region are on both
			// sides of it.
			ok = float64(inside)/float64(nIn) <= withheldRate*float64(outside)/float64(nOut)
		} else {
			// Whole year, or a window with nothing outside it to compare:
			// only the other source can show what eBird is missing.
			ok = iTotal > 0 && outRecords*eTotal <= iTotal
		}
		if ok {
			for m := 0; m < 12; m++ {
				out[m] = out[m] || in[m]
			}
		}
	}
	return out
}

// WithheldByEBird is the set of species eBird withholds everywhere BirdView
// looks: a global entry on its Sensitive Species list, or year-round entries
// for both the United States and Canada (decisions.sharedCountShownOnBoth, decisions.countEbirdFirst).
//
// Derived from the list rather than kept by hand. The hand-kept version
// held eleven species, found by querying every plausible alias and finding
// nothing, plus Great Gray Owl and Black Rail, which the ratio band of product.sourceNamingMeasured
// caught after the zero-record sweep missed them -- eBird holds one record of
// the owl and sixteen of the rail. It also held **Desertas Petrel**, which is
// on no entry of eBird's list: its missing records are a naming question,
// and saying "withheld by eBird" about it was false.
//
// Kept apart from the naming tables because the two are indistinguishable in
// the data -- eBird returns zero for both -- and opposite in meaning: a page
// telling a reader to look for another name would send them after one that
// does not exist.
//
// Verifies: product.namesRecordsFiledUnder.
var WithheldByEBird = func() map[string]bool {
	allYear := map[string]map[string]bool{}
	out := map[string]bool{}
	for _, w := range EBirdSensitive {
		switch {
		case w.Region == "":
			out[w.Species] = true
		case w.From == "" && (w.Region == "USA" || w.Region == "CAN"):
			if allYear[w.Species] == nil {
				allYear[w.Species] = map[string]bool{}
			}
			allYear[w.Species][w.Region] = true
		}
	}
	for sp, c := range allYear {
		if c["USA"] && c["CAN"] {
			out[sp] = true
		}
	}
	return out
}()
