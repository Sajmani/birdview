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

// Package month holds the twelve months BirdView filters by, and the one
// rule that decides whether a species is present in one (product.regionAndMonthFilter, decisions.monthPresence).
//
// A package of its own, small as it is, because the rule has three callers
// that must agree — the index's species list, the tile's figure and the
// detail page's — and the last time one measure had two implementations the
// index and the detail page disagreed about where a bird was commonest
// (decisions.honestSeaDenominator).
package month

import "strings"

// Count is how many months there are, which is not a number worth writing
// out twice.
const Count = 12

// Abbrev is the three-letter form a URL and a chip both use, January first.
//
// English and not localized, like the region slugs beside it (decisions.regionUrlIsPostalCode): a URL
// that changed with the reader's language could not be shared.
var Abbrev = [Count]string{
	"Jan", "Feb", "Mar", "Apr", "May", "Jun",
	"Jul", "Aug", "Sep", "Oct", "Nov", "Dec",
}

// Full is the name a caption shows, where there is room for it.
var Full = [Count]string{
	"January", "February", "March", "April", "May", "June",
	"July", "August", "September", "October", "November", "December",
}

// Index returns the 0-based month for an abbreviation, case-insensitively.
// An unknown string is dropped rather than guessed at, exactly as an
// unknown region slug is.
func Index(abbrev string) (int, bool) {
	for i, a := range Abbrev {
		if strings.EqualFold(a, abbrev) {
			return i, true
		}
	}
	return 0, false
}

// MinRecords is the evidence floor below which the month filter abstains:
// a species with fewer records than this in a region is shown in EVERY
// selected month rather than hidden from any (decisions.monthPresence).
//
// Millerbird is why. It holds 28 records in Hawaii, all in five months, and
// lives on Nihoa the year round — so those months report when anyone last
// got to Nihoa, not when the bird is there. Hiding it from June would make
// a statement about the bird out of a fact about the survey.
//
// **200, and the first guess of 100 was wrong.** Measured 2026-10-07 over
// all 27,830 species-region cells and the real monthly effort of New York
// and Hawaii, where the risk is that a species present in every month is
// called absent in at least one by chance alone:
//
//	floor   cells abstaining   records abstaining   false absence NY / HI
//	 100         19.0%              0.0145%            10.2% / 7.9%
//	 150         22.4%              0.0240%             2.4% / 0.6%
//	 200         25.1%              0.0345%             0.5% / 0.2%
//
// A uniform month share makes 100 look safe at 0.3%. Effort is not uniform
// — New York records 20.9% of its birding in May and 4.8% in December — and
// against the real profile 100 misfires on a tenth of year-round species.
//
// A quarter of the cells and three hundredths of one percent of the records
// is what makes the floor cheap: abstention concentrates where there is
// nothing to abstain from.
const MinRecords = 200

// MinRatio is how much of its usual share a species must hold in a month to
// count as present then (decisions.monthPresence).
//
// Not load-bearing: measured across New York and Hawaii the passing ratios
// cluster at 0.31 and above and the failing ones at 0.17 and below, so any
// cut between gives the same answer. 0.2 rather than 0.3 because a stricter
// cut is easier to fail, which would raise the false-absence risk above for
// nothing.
const MinRatio = 0.2

// Present reports whether a species is present in a region in each month,
// and whether there was enough evidence to say at all.
//
// The measure is the species' share of that month against its share of the
// year — product.shareOfRecords's own logic at month granularity. Dividing by the month's
// own total is not optional: a raw share would call every bird a May bird,
// because May is when people are out.
//
// When ok is false the caller must show the species in every selected month
// and say why. An all-true result is NOT the same answer: that means the
// bird really is there all year.
//
// Verifies: product.regionAndMonthFilter, decisions.monthPresence.
func Present(speciesMonths, regionMonths [Count]int64) (in [Count]bool, ok bool) {
	var spYear, regYear int64
	for i := 0; i < Count; i++ {
		spYear += speciesMonths[i]
		regYear += regionMonths[i]
	}
	if spYear < MinRecords || regYear == 0 {
		return in, false
	}
	annual := float64(spYear) / float64(regYear)
	for i := 0; i < Count; i++ {
		if regionMonths[i] == 0 {
			// No birding at all that month, so nothing is known about any
			// species in it. Counted as present rather than absent: the
			// species is not what is missing.
			in[i] = true
			continue
		}
		in[i] = (float64(speciesMonths[i])/float64(regionMonths[i]))/annual >= MinRatio
	}
	return in, true
}
