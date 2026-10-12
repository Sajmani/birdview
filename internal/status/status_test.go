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

package status

import (
	"strings"
	"testing"
	"time"
)

// days expands an ISO start date and a list of day offsets into dates. The
// offsets state the shape being tested; writing the dates out hides it.
// since is the recency cutoff these cases are judged against. Fixed and early
// enough that every fixture below counts as recent, so the cases about
// episodes are about episodes; TestHistoricalRecordsAreNotPresence is where
// the cutoff itself is the subject.
var since = time.Date(2000, 1, 1, 0, 0, 0, 0, time.UTC)

func days(start string, offsets ...int) []time.Time {
	t0, err := time.Parse("2006-01-02", start)
	if err != nil {
		panic(err)
	}
	var out []time.Time
	for _, d := range offsets {
		out = append(out, t0.AddDate(0, 0, d))
	}
	return out
}

// seq is n consecutive day offsets, 0..n-1.
func seq(n int) []int {
	out := make([]int, n)
	for i := range out {
		out[i] = i
	}
	return out
}

// The two vagrants named in CR-028, as they appear in the data. Both must come
// out Vagrant, and the second is the one that matters: it straddles a New Year,
// so a rule counting distinct YEARS sees two and lets it through. Counting
// episodes sees one arrival.
//
// Watched fail: changing Classify's Vagrant default to Rare turns both red.
// Replacing the recurrence arm with a bare `s.Years >= 2` -- the shape of the
// four-year-window rule CR-028 rejected -- turns LucysWarbler red while the
// flamingo stays green, because 13 days in September sit inside one calendar
// year and Lucy's six days straddle a New Year. That asymmetry is the whole
// reason the second case is here.
func TestKnownVagrantsClassifyAsVagrant(t *testing.T) {
	tests := []struct {
		name    string
		records int64
		dates   []time.Time
	}{
		// American Flamingo, Pennsylvania, post-Idalia: 931 records on 13
		// consecutive days, one pond.
		{"AmericanFlamingo/PA", 931, days("2023-09-07", seq(13)...)},
		// Lucy's Warbler, Virginia: 53 records over 2016-12-31..2017-01-06.
		{"LucysWarbler/VA", 53, days("2016-12-31", 0, 1, 2, 3, 5, 6)},
	}
	for _, tt := range tests {
		got := Classify(Measure(tt.records, tt.dates, false), since)
		if got != Vagrant {
			t.Errorf("%s: got %q, want %q -- one arrival is not recurrence, "+
				"however many observers it drew", tt.name, got, Vagrant)
		}
	}
}

// A bird seen a few times a year for decades is the case an abundance rule
// kills and this one must keep: Akikiki has fewer records than the flamingo
// above and is a critically endangered resident.
//
// Watched fail: raising EstablishedEpisodes above 20 drops this to Rare and
// the test reports the status it got.
func TestScarceButRecurringIsEstablished(t *testing.T) {
	// Twenty years, one sighting each spring: far apart in time, few in number.
	var offsets []int
	for y := 0; y < 20; y++ {
		offsets = append(offsets, y*365)
	}
	s := Measure(376, days("2004-05-01", offsets...), false)
	if s.Episodes != 20 {
		t.Fatalf("episodes = %d, want 20 -- sightings a year apart are separate arrivals", s.Episodes)
	}
	if got := Classify(s, since); got != Established {
		t.Errorf("got %q, want %q -- 20 arrivals across 20 years is recurrence", got, Established)
	}
}

// Gap is a declared coefficient, so its boundary is pinned. Dates exactly Gap
// apart are one episode; one day more is two.
//
// Watched fail: changing the comparison in Measure from `> Gap` to `>= Gap`
// makes the first case report 2 episodes.
func TestGapBoundary(t *testing.T) {
	if got := Measure(2, days("2020-01-01", 0, Gap), false).Episodes; got != 1 {
		t.Errorf("dates %d days apart gave %d episodes, want 1", Gap, got)
	}
	if got := Measure(2, days("2020-01-01", 0, Gap+1), false).Episodes; got != 2 {
		t.Errorf("dates %d days apart gave %d episodes, want 2", Gap+1, got)
	}
}

// A truncated facet is a floor, not a count. It must classify as Established
// without ever being compared against the episode bars, because the floor
// cannot reach them.
//
// Watched fail: deleting the `case s.Truncated` arm drops this to Vagrant --
// one page of consecutive dates is one episode -- which is the exact failure
// that would have shipped a robin as a vagrant.
func TestTruncatedFacetIsEstablished(t *testing.T) {
	// One page of consecutive dates: one episode, one year, by the numbers.
	s := Measure(1498010, days("2024-01-01", seq(300)...), true)
	if s.Episodes != 1 {
		t.Fatalf("episodes = %d, want 1 -- the fixture must be one unbroken run "+
			"or it would pass for the wrong reason", s.Episodes)
	}
	if got := Classify(s, since); got != Established {
		t.Errorf("got %q, want %q -- a truncated facet means more dates than a page holds",
			got, Established)
	}
}

// Watched fail: returning Vagrant instead of Absent for an empty date list
// turns this red. The distinction matters because Absent means "no open
// record exists", which P-071 handles by declaration, while Vagrant is a
// measurement.
func TestNoDatesIsAbsent(t *testing.T) {
	if got := Classify(Measure(0, nil, false), since); got != Absent {
		t.Errorf("got %q, want %q", got, Absent)
	}
	// Records with no usable date are still Absent for this purpose: that is
	// Akohekohe, 7 iNaturalist records and not one parseable date.
	if got := Classify(Measure(7, nil, false), since); got != Absent {
		t.Errorf("records but no dates: got %q, want %q", got, Absent)
	}
}

// Scope follows the strongest region, not the nearest one. A flamingo is a
// vagrant in Pennsylvania and a resident in Florida and belongs in the data.
//
// Watched fail: making Best return the LAST status rather than the strongest
// turns this red.
func TestBestTakesTheStrongestRegion(t *testing.T) {
	got := Best([]Status{Vagrant, Established, Absent, Rare})
	if got != Established {
		t.Errorf("got %q, want %q", got, Established)
	}
	if !got.InScope() {
		t.Error("Established must be in scope")
	}
	if Best(nil) != Absent {
		t.Error("no regions at all must be Absent")
	}
	if Vagrant.InScope() || Absent.InScope() {
		t.Error("Vagrant and Absent are out of scope (P-001)")
	}
}

// Measure must not reorder or depend on the caller's ordering.
//
// Watched fail: dropping the defensive copy in Measure makes this report that
// the input was reordered.
func TestMeasureDoesNotDisturbItsInput(t *testing.T) {
	in := days("2020-06-01", 10, 0, 5)
	want := append([]time.Time(nil), in...)
	s := Measure(3, in, false)
	for i := range in {
		if !in[i].Equal(want[i]) {
			t.Fatalf("Measure reordered its input at %d", i)
		}
	}
	if s.Dates != 3 || s.SpanDays != 11 {
		t.Errorf("dates=%d span=%d, want 3 and 11", s.Dates, s.SpanDays)
	}
}

// Duplicate dates are one date: the facet reports counts per day, and a day
// seen by forty observers is still one day.
//
// Watched fail: removing the uniqueness check in Measure reports 4 dates.
func TestRepeatedDatesCollapse(t *testing.T) {
	s := Measure(40, days("2021-03-01", 0, 0, 1, 1), false)
	if s.Dates != 2 {
		t.Errorf("dates = %d, want 2", s.Dates)
	}
	if s.Episodes != 1 {
		t.Errorf("episodes = %d, want 1", s.Episodes)
	}
}

// A bird nobody has seen in a century recurs beautifully, and is not present.
//
// GBIF carries museum specimens under their collection dates, so recurrence
// alone cannot tell an extinct species from a resident. Passenger Pigeon has
// 237 records on 213 distinct dates across the regions BirdView covers,
// every one of them
// between 1802 and 1902 -- 89 episodes across 52 years, which clears every
// bar in this package.
//
// Watched fail: deleting the LastSeen arm from Classify returns established
// for the pigeon and this reports it.
func TestHistoricalRecordsAreNotPresence(t *testing.T) {
	// Fifty-two years of sightings, ending in 1902.
	var offsets []int
	for y := 0; y < 52; y++ {
		offsets = append(offsets, y*365)
	}
	pigeon := Measure(237, days("1850-05-18", offsets...), false)
	if pigeon.Episodes < EstablishedEpisodes || pigeon.Years < EstablishedYears {
		t.Fatalf("the fixture must clear the established bars or it proves nothing: "+
			"episodes=%d years=%d", pigeon.Episodes, pigeon.Years)
	}
	if got := Classify(pigeon, since); got != Historical {
		t.Errorf("got %q, want %q -- it recurs across half a century and the last "+
			"record is from 1902", got, Historical)
	}
	if Historical.InScope() {
		t.Error("historical must be out of scope: the bird is gone")
	}
	// The same shape, recent, is established -- so the cutoff is doing the
	// work and not the episode counts.
	live := Measure(237, days("1990-05-18", offsets...), false)
	if got := Classify(live, since); got != Established {
		t.Errorf("the same shape ending recently got %q, want %q", got, Established)
	}
}

// TestReadParsesTheLastSeenColumn asserts the tenth column round-trips, and
// that a row missing it is refused rather than read short.
//
// The refusal is the half that matters. Read split on tabs and indexed
// positionally, so a nine-field line left over from before the column
// existed would have parsed happily and left every LastSeen zero -- which
// reads as "never recorded" and is exactly the state an `absent` species is
// in. A stale table would then have turned every historical bird into an
// absent one silently, on a page whose whole job is to say which.
//
// Watched fail: relaxing the field count to `< 9` accepts the short row and
// reports a zero date for Passenger Pigeon; dropping the time.Parse call
// leaves the date zero on the full row too.
//
// Verifies: P-075.
func TestReadParsesTheLastSeenColumn(t *testing.T) {
	const full = "# a comment\n" +
		"ectopistes migratorius\thistorical\tCAN.9_1\t89\t52\t145\tCAN.6_1\t4\t303189\t1902-05-18\n" +
		"agapornis lilianae\tabsent\t-\t0\t0\t0\tUSA.44_1\t1\t70979374\t-\n"

	rows, err := Read(strings.NewReader(full))
	if err != nil {
		t.Fatalf("Read: %v", err)
	}
	if len(rows) != 2 {
		t.Fatalf("read %d rows, want 2", len(rows))
	}
	want := time.Date(1902, 5, 18, 0, 0, 0, 0, time.UTC)
	if !rows[0].LastSeen.Equal(want) {
		t.Errorf("Passenger Pigeon last seen %v, want %v; the date is the whole "+
			"content of a historical verdict", rows[0].LastSeen, want)
	}
	if !rows[1].LastSeen.IsZero() {
		t.Errorf("an absent species has a last-seen date of %v; `-` means the "+
			"derivation measured none", rows[1].LastSeen)
	}

	// A table from before the column existed must be refused, not read short.
	short := "ectopistes migratorius\thistorical\tCAN.9_1\t89\t52\t145\tCAN.6_1\t4\t303189\n"
	if _, err := Read(strings.NewReader(short)); err == nil {
		t.Error("a nine-field row was accepted; every date would silently read as " +
			"never-recorded, which is the state `absent` means")
	}

	// And a date that is not one is an error, not a zero.
	bad := "ectopistes migratorius\thistorical\tCAN.9_1\t89\t52\t145\tCAN.6_1\t4\t303189\tMay 1902\n"
	if _, err := Read(strings.NewReader(bad)); err == nil {
		t.Error("an unparseable date was accepted as no date at all")
	}
}

// TestScopeBringsBothHalvesOfASharedName asserts that a name covering two
// species brings both into scope (P-077, CR-038).
//
// Once scope is measured on eBird, Northern Yellow Warbler has no row of its
// own that recurs: eBird files every one of its records under Setophaga
// petechia. Without this the commonest warbler in North America would leave
// BirdView over a filing choice -- and nothing else would notice, because
// the dataset and the scope would shrink together.
//
// Verifies: P-077, CR-038. Criterion: AC-114.
func TestScopeBringsBothHalvesOfASharedName(t *testing.T) {
	rows := []Row{
		{Name: "setophaga petechia", Status: Established},
		{Name: "setophaga aestiva", Status: Absent},
		{Name: "trogon elegans", Status: Established},
		{Name: "trogon ambiguus", Status: Absent},
	}
	in := map[string]bool{}
	for _, n := range Scope(rows) {
		in[n] = true
	}
	for _, want := range []string{"Setophaga aestiva", "Setophaga petechia", "Trogon ambiguus", "Trogon elegans"} {
		if !in[want] {
			t.Errorf("%s is out of scope; its records are filed under the shared name, "+
				"which recurs", want)
		}
	}
}
