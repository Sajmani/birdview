package status

import (
	"testing"
	"time"
)

// days expands an ISO start date and a list of day offsets into dates. The
// offsets state the shape being tested; writing the dates out hides it.
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
		got := Classify(Measure(tt.records, tt.dates, false))
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
	if got := Classify(s); got != Established {
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
	if got := Classify(s); got != Established {
		t.Errorf("got %q, want %q -- a truncated facet means more dates than a page holds",
			got, Established)
	}
}

// Watched fail: returning Vagrant instead of Absent for an empty date list
// turns this red. The distinction matters because Absent means "no open
// record exists", which P-071 handles by declaration, while Vagrant is a
// measurement.
func TestNoDatesIsAbsent(t *testing.T) {
	if got := Classify(Measure(0, nil, false)); got != Absent {
		t.Errorf("got %q, want %q", got, Absent)
	}
	// Records with no usable date are still Absent for this purpose: that is
	// Akohekohe, 7 iNaturalist records and not one parseable date.
	if got := Classify(Measure(7, nil, false)); got != Absent {
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
