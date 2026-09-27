package view

import (
	"math"
	"math/rand"
	"testing"

	"birdview/internal/dataset"
	"birdview/internal/morphology"
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

// foldingPair builds two birds with identical measurements whose only
// difference is that one belongs to a family that folds its neck in flight.
// Every assertion below is then attributable to the fold and nothing else.
func foldingPair(neck float64) (folds, extends *dataset.Species) {
	folds = bird("Ardea herodias", 140, 40, neck, 300, 180, 1825, 2400)
	folds.Family = "Ardeidae"
	extends = bird("Antigone canadensis", 140, 40, neck, 300, 180, 1825, 2400)
	extends.Family = "Gruidae"
	return folds, extends
}

// TestFoldedNeckIsTheSameStrokeInLessHeight asserts P-040's folded neck: for the
// families that fly with the head drawn back onto the shoulders, the neck is an
// S of two semicircles rather than a stick.
//
// What makes the S defensible is that nothing is rescaled. Two semicircles of
// radius r run 2*pi*r of stroke through 4r of height, so setting r from the
// neck length means the ink you see is exactly the neck, at the same
// pixels-per-millimeter as the rest of the bird -- the same neck, curved. The
// compression that follows is 2/pi, a consequence of the shape rather than a
// number anyone picked, which is precisely what separates this from the
// per-family leg factors that CR-010 deleted.
//
// So the assertion is on the arc length, which no change to the radius can
// satisfy by accident, and the compression is checked against 2/pi rather than
// against whatever the code happens to produce.
//
// Verifies: P-040, P-041. Criterion: AC-067.
func TestFoldedNeckIsTheSameStrokeInLessHeight(t *testing.T) {
	const neck = 200.0
	heron, crane := foldingPair(neck)

	box := Box{W: 420, H: 160}
	gh := Layout(heron, heron.TotalLength.Value, heron.Wingspan.Value, box)
	gc := Layout(crane, crane.TotalLength.Value, crane.Wingspan.Value, box)

	if !gh.NeckFolded {
		t.Fatal("Ardeidae flies with the neck folded; Layout did not fold it")
	}
	if gc.NeckFolded || gc.NeckArcR != 0 {
		t.Fatalf("Gruidae flies neck-extended, but Layout folded it (r=%.3f)", gc.NeckArcR)
	}

	want := neck * gh.PxPerMm

	// The stroke is exactly the neck long.
	if got := 2 * math.Pi * gh.NeckArcR; math.Abs(got-want) > 1e-9*want {
		t.Errorf("the folded neck runs %.6f px of arc, but the scaled neck is %.6f px",
			got, want)
	}
	// ...through less height, by exactly 2/pi.
	drawn := gh.YShoulder - gh.YNeckTop
	if exp := neckFoldCompression * want; math.Abs(drawn-exp) > 1e-9*want {
		t.Errorf("the folded neck spans %.6f px, want %.6f -- 2/pi of its own length", drawn, exp)
	}
	if drawn >= want {
		t.Errorf("the folded neck spans %.2f px for a %.2f px neck; folding saved no height",
			drawn, want)
	}
	// The twin that does not fold keeps its full height.
	if straight := gc.YShoulder - gc.YNeckTop; math.Abs(straight-want) > 1e-9*want {
		t.Errorf("the straight neck spans %.6f px, want its full %.6f", straight, want)
	}
	// The two arcs join halfway down.
	if mid := gh.YNeckTop + 2*gh.NeckArcR; math.Abs(gh.NeckMid()-mid) > 1e-9 {
		t.Errorf("NeckMid is %.6f, but the arcs join at %.6f", gh.NeckMid(), mid)
	}
	// The whole drawing is shorter by exactly what the neck saved, and by
	// nothing else -- no other segment may shift.
	if saved, diff := want-drawn, gc.TotalHeight()-gh.TotalHeight(); math.Abs(diff-saved) > 1e-9*want {
		t.Errorf("the folded bird draws %.6f px shorter, but its neck saved %.6f -- something "+
			"other than the neck moved", diff, saved)
	}
}

// TestFoldedNeckDoesNotChangeOrdering asserts the other half of CR-011: the fold
// is a rendering change and nothing else.
//
// Two birds of identical total length, one folding and one not, must sort into
// the same section and be measured against the same reference even though one
// of them draws visibly shorter. If the fold ever leaks into Sections or Scales,
// a heron starts sorting as a smaller bird than it is -- which is the one thing
// this change was not allowed to do.
//
// Verifies: P-021, P-042, P-041. Criterion: AC-067.
func TestFoldedNeckDoesNotChangeOrdering(t *testing.T) {
	heron, crane := foldingPair(200)
	in := []*dataset.Species{heron, crane}

	secs := Sections(in, ByLength, DefaultStepPercent)
	if len(secs) != 1 || len(secs[0].Species) != 2 {
		t.Fatalf("two birds of equal length should share one section; got %d sections", len(secs))
	}
	// The reference must be the folding bird's OWN full length, so it is also
	// checked against a section holding nothing else. Asking only the mixed
	// section proves nothing: Scales takes a maximum, so the extended twin masks
	// any reduction applied to the heron -- which is exactly how the first
	// version of this check passed while the fold leaked into the scale.
	alone := Sections([]*dataset.Species{heron}, ByLength, DefaultStepPercent)
	for _, mode := range []ScaleMode{PerSection, Global} {
		for _, c := range []struct {
			label string
			ref   float64
		}{{"heron alone", Scales(alone, mode)[0]}, {"mixed section", Scales(secs, mode)[0]}} {
			if math.Abs(c.ref-heron.TotalLength.Value) > 1e-9 {
				t.Errorf("%s, %s: scale reference is %.3f, want the full total length %.3f -- "+
					"the folded neck has leaked into the scale", mode, c.label, c.ref,
					heron.TotalLength.Value)
			}
		}
	}

	// And the premise of all of the above: they really do draw differently.
	box := Box{W: 420, H: 160}
	gh := Layout(heron, heron.TotalLength.Value, heron.Wingspan.Value, box)
	gc := Layout(crane, crane.TotalLength.Value, crane.Wingspan.Value, box)
	if gh.TotalHeight() >= gc.TotalHeight() {
		t.Errorf("the folded bird draws %.2f px and the extended one %.2f -- they are identical, "+
			"so the checks above prove nothing", gh.TotalHeight(), gc.TotalHeight())
	}
}

// TestEveryFoldedFamilyHasALongNeck asserts the invariant that ties the two
// neck tables together: a family only appears in foldedNecks because its neck is
// long enough that folding it is the defining thing about its flight silhouette.
// A family drawing the default ratio has no such neck, so listing it in one
// table and not the other means one of them is wrong.
//
// This is not hypothetical. Pelecanidae was in foldedNecks while falling back to
// the 0.18 default, so a pelican -- a bird whose neck is its most obvious
// feature -- folded a 5%-of-length stub. The bug survived CR-011 review because
// each table looked reasonable on its own.
//
// Verifies: P-040, P-045. Criterion: AC-069.
func TestEveryFoldedFamilyHasALongNeck(t *testing.T) {
	for fam := range foldedNecks {
		ratio, explicit := morphology.NeckRatio(fam)
		if !explicit {
			t.Errorf("%s folds its neck but has no entry in morphology.NeckRatios, so it draws "+
				"the %.2f default -- one of the two tables is wrong", fam, ratio)
		}
	}
	if len(foldedNecks) == 0 {
		t.Fatal("no families fold, so this check asserts nothing")
	}
}

// TestChordTickMarksTheMeasuredPortion asserts the geometry behind P-065's
// split crossbar: the tick sits exactly one scaled chord inboard of the wingtip.
//
// That identity is the whole argument for splitting the bar. The outer segment
// is a measured length, so it stays true even when the span model is wrong, and
// the inner segment absorbs the entire error (CR-015). If the tick is placed by
// any other rule, the outer segment stops being the chord and the claim the
// drawing makes becomes false.
//
// Verifies: P-065. Criterion: AC-072.
func TestChordTickMarksTheMeasuredPortion(t *testing.T) {
	box := Box{W: 420, H: 160}
	s := bird("Ardea herodias", 140, 40, 200, 300, 180, 1825, 2400)
	s.WingChord = dataset.Metric{Value: 466.4}

	g := Layout(s, s.TotalLength.Value, s.Wingspan.Value, box)
	outer := g.CrossbarHalfWidth - g.ChordTickX
	if want := s.WingChord.Value * g.PxPerMm; math.Abs(outer-want) > 1e-9 {
		t.Errorf("outer crossbar segment is %.6f px, want the scaled chord %.6f", outer, want)
	}
	if g.ChordTickX <= 0 || g.ChordTickX >= g.CrossbarHalfWidth {
		t.Errorf("tick at %.3f is not strictly inside the half-span %.3f",
			g.ChordTickX, g.CrossbarHalfWidth)
	}

	// No chord, and a chord that would reach past the axis, both fall back to a
	// whole crossbar rather than a tick drawn somewhere impossible.
	for _, c := range []struct {
		name  string
		chord float64
	}{{"absent", 0}, {"equal to the half-span", 912.5}, {"longer than the half-span", 2000}} {
		s2 := bird("x", 140, 40, 200, 300, 180, 1825, 2400)
		s2.WingChord = dataset.Metric{Value: c.chord}
		if got := Layout(s2, s2.TotalLength.Value, s2.Wingspan.Value, box).ChordTickX; got != 0 {
			t.Errorf("chord %s: ChordTickX is %.3f, want 0 so the bar is drawn whole", c.name, got)
		}
	}
}

// TestDetailLabelsDoNotCollide asserts P-066's layout rules.
//
// The collision pass is exercised with a synthetic geometry, because no real
// species needs it: at the current box and font, 0 of 841 have segment
// midpoints closer than a line height. It is kept rather than deleted because
// unlike the dead branch CR-010 removed, this one is dead by the current
// numbers rather than by construction -- a larger font or a shorter box wakes
// it -- so it has to be tested by construction too.
//
// Verifies: P-066. Criterion: AC-073.
func TestDetailLabelsDoNotCollide(t *testing.T) {
	box := Box{W: 988, H: 281}

	// Segments a couple of px apart: every label wants nearly the same y.
	tight := Geometry{YBillTip: 100, YCraniumTop: 102, YNeckTop: 104,
		YShoulder: 106, YBodyEnd: 108, YTailTip: 110}
	got := tight.DetailLabels(box, 0)
	if len(got) != 5 {
		t.Fatalf("got %d labels, want one per segment", len(got))
	}
	for i, want := range []string{"bill", "head", "neck", "body", "tail"} {
		if got[i].Key != want {
			t.Errorf("label %d is %q, want %q -- labels must stay in segment order", i, got[i].Key, want)
		}
	}
	for i := 1; i < len(got); i++ {
		if d := got[i].TextY - got[i-1].TextY; d < labelMinGap-1e-9 {
			t.Errorf("labels %s and %s are %.2f px apart, closer than the %d px line height",
				got[i-1].Key, got[i].Key, d, labelMinGap)
		}
	}
	if last := got[len(got)-1].TextY; last > box.H {
		t.Errorf("last label at %.1f is below the %.0f px box", last, box.H)
	}

	// Anchors are untouched by the spreading: the leader is what ties a label
	// to its segment, so the anchor must stay on the segment's midpoint.
	for _, c := range []struct {
		key  string
		a, b float64
	}{{"bill", 100, 102}, {"head", 102, 104}, {"neck", 104, 106}, {"body", 106, 108}, {"tail", 108, 110}} {
		for _, l := range got {
			if l.Key == c.key {
				if want := (c.a + c.b) / 2; math.Abs(l.AnchorY-want) > 1e-9 {
					t.Errorf("%s anchor is %.3f, want its midpoint %.3f", c.key, l.AnchorY, want)
				}
			}
		}
	}

	// And on a roomy bird nothing moves at all: text sits exactly on the anchor.
	roomy := Geometry{YBillTip: 0, YCraniumTop: 40, YNeckTop: 80,
		YShoulder: 130, YBodyEnd: 200, YTailTip: 260}
	for _, l := range roomy.DetailLabels(box, 0) {
		if math.Abs(l.TextY-l.AnchorY) > 1e-9 {
			t.Errorf("%s moved to %.2f from %.2f with room to spare", l.Key, l.TextY, l.AnchorY)
		}
	}
}
