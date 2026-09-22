package view

import (
	"math"
	"math/rand"
	"testing"

	"birdview/internal/dataset"
)

// bird builds a species with segments that sum to total, as the ingest
// guarantees (P-041).
func bird(name string, bill, head, neck, body, tail, span, mass float64) *dataset.Species {
	return &dataset.Species{
		ScientificName: name,
		TotalLength:    dataset.Metric{Value: bill + head + neck + body + tail},
		Wingspan:       dataset.Metric{Value: span},
		Mass:           dataset.Metric{Value: mass},
		Segments: dataset.Segments{
			Bill: dataset.Metric{Value: bill}, Head: dataset.Metric{Value: head},
			Neck: dataset.Metric{Value: neck}, Body: dataset.Metric{Value: body},
			Tail: dataset.Metric{Value: tail},
		},
	}
}

func sample() []*dataset.Species {
	return []*dataset.Species{
		bird("Calypte anna", 18, 12, 6, 40, 26, 120, 4),
		bird("Passer domesticus", 12, 14, 9, 55, 55, 240, 28),
		bird("Turdus migratorius", 20, 18, 12, 90, 85, 390, 77),
		bird("Corvus brachyrhynchos", 50, 30, 22, 170, 180, 1000, 450),
		bird("Buteo jamaicensis", 26, 32, 25, 240, 210, 1215, 1100),
		bird("Ardea herodias", 140, 40, 200, 300, 180, 1825, 2400),
		bird("Cygnus buccinator", 105, 50, 400, 420, 180, 2400, 11000),
	}
}

// Verifies: P-021. Criterion: AC-020.
func TestSortedAscending(t *testing.T) {
	for _, m := range []Metric{ByLength, ByWingspan, ByMass} {
		secs := Sections(sample(), m, DefaultStepPercent)
		var prev float64
		for _, sec := range secs {
			for _, s := range sec.Species {
				if v := m.Value(s); v < prev {
					t.Errorf("%s: %s out of order (%v after %v)", m, s.ScientificName, v, prev)
				} else {
					prev = v
				}
			}
		}
	}
}

// TestSectionRule asserts P-022 as written: a section breaks at the first bird
// more than N% larger than the MINIMUM of the previous section.
//
// Both halves matter. The first says no section spans more than N%. The second
// says the break happens at the earliest such bird, which is what stops a
// section drifting wide one bird at a time -- the failure an implementation
// keyed on the previous maximum would produce while passing casual inspection.
//
// Verifies: P-022, P-023. Criterion: AC-021.
func TestSectionRule(t *testing.T) {
	r := rand.New(rand.NewSource(7))
	for _, step := range []float64{5, 10, DefaultStepPercent, 25, 50} {
		factor := 1 + step/100
		for trial := 0; trial < 300; trial++ {
			var in []*dataset.Species
			for i := 0; i < 1+r.Intn(60); i++ {
				v := 40 + r.Float64()*1200
				in = append(in, bird("s", v/5, v/5, v/5, v/5, v/5, v*1.5, v))
			}
			secs := Sections(in, ByLength, step)

			var prevMin float64
			for i, sec := range secs {
				if sec.Max >= sec.Min*factor {
					t.Fatalf("step %.0f%%: section %d spans %.3f..%.3f, wider than %.0f%%",
						step, i, sec.Min, sec.Max, step)
				}
				if i > 0 && sec.Min < prevMin*factor {
					t.Fatalf("step %.0f%%: section %d starts at %.3f, below %.0f%% above the "+
						"previous minimum %.3f -- it should have joined that section",
						step, i, sec.Min, step, prevMin)
				}
				prevMin = sec.Min
			}
		}
	}
}

// TestSectionsPartition checks every species lands in exactly one section.
//
// Verifies: P-022. Criterion: AC-021.
func TestSectionsPartition(t *testing.T) {
	in := sample()
	var n int
	seen := map[*dataset.Species]bool{}
	for _, sec := range Sections(in, ByLength, DefaultStepPercent) {
		for _, s := range sec.Species {
			if seen[s] {
				t.Errorf("%s appears twice", s.ScientificName)
			}
			seen[s] = true
			n++
		}
	}
	if n != len(in) {
		t.Errorf("sections hold %d species, input had %d", n, len(in))
	}
}

// TestSectionRangeMatchesContents verifies the header can state a range that is
// actually true of its members.
//
// Verifies: P-024. Criterion: AC-022.
func TestSectionRangeMatchesContents(t *testing.T) {
	for _, sec := range Sections(sample(), ByMass, DefaultStepPercent) {
		lo, hi := math.Inf(1), math.Inf(-1)
		for _, s := range sec.Species {
			v := ByMass.Value(s)
			lo, hi = math.Min(lo, v), math.Max(hi, v)
		}
		if sec.Min != lo || sec.Max != hi {
			t.Errorf("section reports %.1f..%.1f, contents are %.1f..%.1f", sec.Min, sec.Max, lo, hi)
		}
	}
}

// TestScaleModes asserts P-042: one scale per section, or one for everything.
//
// The second assertion is the one with teeth. "Identical within a section" is
// satisfied trivially by a per-tile implementation whenever a section holds one
// bird, so the test also requires that sections genuinely differ from each
// other -- which a per-tile scale could not produce.
//
// Verifies: P-042. Criterion: AC-031.
func TestScaleModes(t *testing.T) {
	secs := Sections(sample(), ByLength, DefaultStepPercent)
	if len(secs) < 2 {
		t.Fatal("sample should produce several sections")
	}

	perSection := Scales(secs, PerSection)
	distinct := map[float64]bool{}
	for i, ref := range perSection {
		var maxLen float64
		for _, s := range secs[i].Species {
			maxLen = math.Max(maxLen, s.TotalLength.Value)
		}
		if ref != maxLen {
			t.Errorf("section %d reference %.1f, want its largest bird %.1f", i, ref, maxLen)
		}
		distinct[ref] = true
	}
	if len(distinct) < 2 {
		t.Error("per-section references are all equal; that is a global scale, not a per-section one")
	}

	global := Scales(secs, Global)
	for i, ref := range global {
		if ref != global[0] {
			t.Errorf("global reference differs between sections %d and 0", i)
		}
	}
	if global[0] != math.Max(perSection[len(perSection)-1], global[0]) {
		t.Error("global reference should be the largest bird overall")
	}
}

// TestCrossbarIsToScale asserts P-043: the crossbar represents wingspan at the
// same scale as the vertical axis, so its length relative to the bird's drawn
// height equals wingspan relative to total length.
//
// An earlier design's formula carried an extra factor that inverted this with
// size -- larger birds drew shorter crossbars, and the 300 largest species
// floored out. This assertion catches that class exactly.
//
// Verifies: P-043. Criterion: AC-032.
func TestCrossbarIsToScale(t *testing.T) {
	species := sample()
	box := Box{W: 420, H: 160}
	var refLen, refSpan float64
	for _, s := range species {
		refLen = math.Max(refLen, s.TotalLength.Value)
		refSpan = math.Max(refSpan, s.Wingspan.Value)
	}
	for _, s := range species {
		g := Layout(s, refLen, refSpan, box)
		want := s.Wingspan.Value / s.TotalLength.Value
		got := (g.CrossbarHalfWidth * 2) / g.TotalHeight()
		if math.Abs(got-want) > 1e-9 {
			t.Errorf("%s: crossbar/height %.6f, wingspan/length %.6f", s.ScientificName, got, want)
		}
	}
}

// TestLargerBirdsDrawLarger is the property G-001 depends on: at a shared
// scale, a bigger bird is bigger on screen.
//
// Verifies: P-042. Criterion: AC-031.
func TestLargerBirdsDrawLarger(t *testing.T) {
	species := sample()
	box := Box{W: 420, H: 160}
	var refLen, refSpan float64
	for _, s := range species {
		refLen = math.Max(refLen, s.TotalLength.Value)
		refSpan = math.Max(refSpan, s.Wingspan.Value)
	}
	for i := 1; i < len(species); i++ {
		a, b := species[i-1], species[i]
		ga := Layout(a, refLen, refSpan, box)
		gb := Layout(b, refLen, refSpan, box)
		if a.TotalLength.Value < b.TotalLength.Value && ga.TotalHeight() >= gb.TotalHeight() {
			t.Errorf("%s (%.0f mm) draws %.1f px, %s (%.0f mm) draws %.1f px",
				a.ScientificName, a.TotalLength.Value, ga.TotalHeight(),
				b.ScientificName, b.TotalLength.Value, gb.TotalHeight())
		}
	}
}

// TestSegmentsAbutInPixels checks the drawn parts meet without gap or overlap,
// which is P-041 expressed in the renderer rather than in the data.
//
// Verifies: P-040, P-041.
func TestSegmentsAbutInPixels(t *testing.T) {
	s := sample()[5] // heron: a long neck makes ordering errors obvious
	g := Layout(s, s.TotalLength.Value, s.Wingspan.Value, Box{W: 420, H: 160})
	ys := []float64{g.YBillTip, g.YCraniumTop, g.YNeckTop, g.YShoulder, g.YBodyEnd, g.YTailTip}
	for i := 1; i < len(ys); i++ {
		if ys[i] < ys[i-1] {
			t.Errorf("coordinate %d (%.2f) is above its predecessor (%.2f)", i, ys[i], ys[i-1])
		}
	}
	if math.Abs(g.YHeadCenter-(g.YCraniumTop+g.HeadRy)) > 1e-9 {
		t.Error("head oval is not centered between its tangents")
	}
	if math.Abs((g.YNeckTop-g.YCraniumTop)-2*g.HeadRy) > 1e-9 {
		t.Error("head oval height does not match the gap between its tangents")
	}
}

// TestSharedScaleWithinSection asserts P-042 directly: birds given the same
// references are drawn at the same pixels-per-millimeter.
//
// This exists because the weaker assertions did not bite. Breaking Layout to
// scale each bird to its own length -- the exact failure P-042 forbids -- still
// passed TestLargerBirdsDrawLarger and TestCrossbarIsToScale, because the
// wingspan constraint happened to preserve size ordering. Comparing the scale
// factor itself is the only check that cannot be satisfied by accident.
//
// Verifies: P-042. Criterion: AC-031.
func TestSharedScaleWithinSection(t *testing.T) {
	species := sample()
	box := Box{W: 420, H: 160}
	var refLen, refSpan float64
	for _, s := range species {
		refLen = math.Max(refLen, s.TotalLength.Value)
		refSpan = math.Max(refSpan, s.Wingspan.Value)
	}
	want := Layout(species[0], refLen, refSpan, box).PxPerMm
	for _, s := range species {
		if got := Layout(s, refLen, refSpan, box).PxPerMm; math.Abs(got-want) > 1e-12 {
			t.Errorf("%s drawn at %.6f px/mm, others at %.6f -- the scale is not shared",
				s.ScientificName, got, want)
		}
	}
	// And the shared scale must actually fit the largest bird in the box.
	big := Layout(species[len(species)-1], refLen, refSpan, box)
	if big.TotalHeight() > box.H+1e-9 || big.CrossbarHalfWidth*2 > box.W+1e-9 {
		t.Errorf("largest bird overflows: %.1f x %.1f in a %.0f x %.0f box",
			big.CrossbarHalfWidth*2, big.TotalHeight(), box.W, box.H)
	}
}

// TestDrawnHeightCoversTheFeet asserts the invariant that replaced "the feet
// never fall below the tail tip".
//
// That older invariant was true of the measured tarsus and is false of the
// modeled leg, deliberately: 11 of 841 species now stand lower on their feet
// than their tail reaches, an American Flamingo by 30% of its own length. So
// the renderer can no longer center on bill-to-tail, and two quantities have to
// agree about how far a bird actually reaches -- VerticalExtent, which sizes a
// section's scale in millimeters before any pixels exist, and DrawnHeight,
// which the template centers on. Let them drift apart and the bird that gets
// clipped is the long-legged one that set the scale.
//
// So the assertion is that they are one quantity in two units. No change to the
// rake, the attachment point or the leg model can satisfy that by accident --
// which is the failure the first version of this test could not see, because it
// only ever compared pixels to other pixels.
//
// Verifies: P-040. Criterion: AC-066.
func TestDrawnHeightCoversTheFeet(t *testing.T) {
	// A stilt and a flamingo: feet well below the tail tip. A swift: nowhere
	// near it, so both branches of DrawnHeight are exercised.
	stilt := bird("Himantopus mexicanus", 63, 26, 23, 88, 76, 720, 166)
	stilt.Tarsus = dataset.Metric{Value: 106.7}
	stilt.Leg = dataset.Metric{Value: 106.7 * 1.7}
	flamingo := bird("Phoenicopterus ruber", 120, 55, 300, 390, 135, 1500, 2800)
	flamingo.Tarsus = dataset.Metric{Value: 310.0}
	flamingo.Leg = dataset.Metric{Value: 310.0 * 2.0}
	swift := bird("Chaetura pelagica", 10, 12, 8, 60, 40, 300, 24)
	swift.Tarsus = dataset.Metric{Value: 11.6}
	swift.Leg = dataset.Metric{Value: 11.6}

	box := Box{W: 420, H: 160}
	longLegged := []*dataset.Species{stilt, flamingo}
	all := append(append(sample(), longLegged...), swift)

	// The scale-free extent and the drawn extent are the same measurement.
	for _, s := range all {
		g := Layout(s, VerticalExtent(s), s.Wingspan.Value, box)
		want := VerticalExtent(s) * g.PxPerMm
		if math.Abs(g.DrawnHeight()-want) > 1e-9*want {
			t.Errorf("%s: drawn height is %.6f px but the extent that sized the box is %.6f px",
				s.ScientificName, g.DrawnHeight(), want)
		}
	}

	// The long-legged birds must really reach past their tails, or the check
	// above proves only that two expressions for total length agree.
	for _, s := range longLegged {
		g := Layout(s, VerticalExtent(s), s.Wingspan.Value, box)
		if g.DrawnHeight() <= g.TotalHeight()+1e-9 {
			t.Errorf("%s: drawn height %.1f does not exceed bill-to-tail %.1f -- this bird is "+
				"here to exercise the feet-below-tail case and no longer does",
				s.ScientificName, g.DrawnHeight(), g.TotalHeight())
		}
	}

	// And the extra room must not be unconditional: a swift's feet stop well
	// above its tail, so it needs exactly its bill-to-tail extent and no more.
	g := Layout(swift, VerticalExtent(swift), swift.Wingspan.Value, box)
	if math.Abs(g.DrawnHeight()-g.TotalHeight()) > 1e-9 {
		t.Errorf("swift: drawn height %.3f differs from bill-to-tail %.3f -- a bird whose feet "+
			"stop above its tail should ask for no extra room", g.DrawnHeight(), g.TotalHeight())
	}

	// Finally, end to end: at the reference Scales actually picks, every bird
	// fits the box feet and all.
	secs := Sections(all, ByLength, DefaultStepPercent)
	refs := Scales(secs, Global)
	for i, sec := range secs {
		for _, s := range sec.Species {
			g := Layout(s, refs[i], s.Wingspan.Value, box)
			if g.DrawnHeight() > box.H+1e-9 {
				t.Errorf("%s: draws %.1f px tall in a %.0f px box -- the section reference does "+
					"not cover its feet", s.ScientificName, g.DrawnHeight(), box.H)
			}
		}
	}
}

// TestLegsScaleWithTarsus checks the drawn leg is exactly the modeled leg long
// and rakes back more steeply than 45 degrees.
//
// Verifies: P-040. Criterion: AC-066.
func TestLegsScaleWithTarsus(t *testing.T) {
	short := bird("swift", 10, 12, 8, 60, 50, 300, 24)
	short.Tarsus = dataset.Metric{Value: 11.6}
	short.Leg = dataset.Metric{Value: 11.6}
	long := bird("crane", 120, 60, 200, 260, 200, 2100, 5000)
	long.Tarsus = dataset.Metric{Value: 288.8}
	long.Leg = dataset.Metric{Value: 288.8 * 1.9}

	box := Box{W: 420, H: 160}
	for _, s := range []*dataset.Species{short, long} {
		g := Layout(s, s.TotalLength.Value, s.Wingspan.Value, box)
		// Whatever the rake, the leg line must be exactly the tarsus long --
		// the angle may only redistribute that length between the axes.
		got := math.Hypot(g.LegDX, g.LegDY)
		want := s.Leg.Value * g.PxPerMm
		if math.Abs(got-want) > 1e-9 {
			t.Errorf("%s: leg line is %.3f px, want the scaled leg %.3f", s.ScientificName, got, want)
		}
		// And it must rake back more steeply than 45 degrees, or a
		// correct-length leg reads short.
		if g.LegDY <= g.LegDX {
			t.Errorf("%s: legs rake at or below 45 degrees (dx %.2f, dy %.2f); the vertical "+
				"share is what makes a leg look like a leg", s.ScientificName, g.LegDX, g.LegDY)
		}
		// And they attach at the body's midpoint.
		if mid := (g.YShoulder + g.YBodyEnd) / 2; math.Abs(g.YLegTop-mid) > 1e-9 {
			t.Errorf("%s: legs attach at %.2f, not the body midpoint %.2f", s.ScientificName, g.YLegTop, mid)
		}
	}
	// A crane's legs must dwarf a swift's, which is the whole point.
	gs := Layout(short, short.TotalLength.Value, short.Wingspan.Value, box)
	gl := Layout(long, long.TotalLength.Value, long.Wingspan.Value, box)
	if gl.LegDX/gl.PxPerMm <= 4*gs.LegDX/gs.PxPerMm {
		t.Error("crane legs are not dramatically longer than a swift's in real terms")
	}
}
