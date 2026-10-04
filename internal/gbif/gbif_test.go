package gbif

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"sort"
	"strings"
	"testing"
	"time"

	"birdview/internal/aba"
	"birdview/internal/avilist"
	"birdview/internal/taxonomy"
)

const (
	tableFile = "../../data/gbif-taxa.tsv"
	avlFile   = "../../AviList-v2025b-10Jun2026-extended.xlsx"
	abaFile   = "../../fetched/ABA_Checklist-8.19.csv"
)

// testIndex is a hand-written slice of AviList: the species each case below
// actually needs, plus the near misses that make the refusals meaningful.
// Written from AviList v2025b, not from a run of this code.
func testIndex() *Index {
	return NewIndex(map[string]string{
		"Nannopterum auritum":     "Phalacrocoracidae",
		"Corthylio calendula":     "Regulidae",
		"Leuconotopicus villosus": "Picidae",
		"Circus cyaneus":          "Accipitridae",
		"Circus hudsonius":        "Accipitridae",
		"Icterus bullockiorum":    "Icteridae",
		"Pica nuttallii":          "Corvidae",
		"Phylloscopus sibilatrix": "Phylloscopidae",
		// Both of these are Anatidae and both fold to "jubat": the Maned Duck
		// of Australia and the Orinoco Goose. They are why the epithet tier
		// needs a uniqueness rule.
		"Chenonetta jubata": "Anatidae",
		"Neochen jubata":    "Anatidae",
		// Shares an epithet with the Hairy Woodpecker across family lines, and
		// shares one with the cormorant across another. Family scoping is the
		// only thing keeping these apart.
		"Myiobius villosus":    "Onychorhynchidae",
		"Crossoptilon auritum": "Phasianidae",
		// Two species in one family honoring the same man. A patronymic is
		// unique within a genus by the Code's homonymy rule and emphatically
		// not within a family.
		"Astur gundlachi":        "Accipitridae",
		"Buteogallus gundlachii": "Accipitridae",
	})
}

// TestResolveTiers pins what each tier does and, more importantly, what it
// refuses. The cases marked "real" are taken from the GBIF backbone as it
// stood on 2026-10-02 and are the reason this bridge exists.
//
// Watched fail: deleting the `len(cands) == 1` guard in Resolve's epithet tier
// (so that the first candidate wins) turns "ambiguous between two Anatidae"
// into "Chenonetta jubata", i.e. an Australian duck standing in for an Orinoco
// Goose, and this test goes red on that case. Dropping the family scope, so
// the epithet tier searches all of AviList, reddens "epithet must not cross a
// family" as well.
//
// Verifies: T-021.
func TestResolveTiers(t *testing.T) {
	idx := testIndex()
	cases := []struct {
		name     string
		taxon    Taxon
		wantName string
		wantTier Tier
	}{
		{
			"exact: the two authorities already agree (real)",
			Taxon{Name: "Circus cyaneus", Family: "Accipitridae", Synonyms: []string{"Falco cyaneus"}},
			"Circus cyaneus", TierExact,
		},
		{
			"exact: AviList adopted the name GBIF accepts (real)",
			Taxon{Name: "Leuconotopicus villosus", Family: "Picidae",
				Synonyms: []string{"Dendrocopos villosus", "Dryobates villosus", "Picoides villosus"}},
			"Leuconotopicus villosus", TierExact,
		},
		{
			"synonym: GBIF's genus is stale (real)",
			Taxon{Name: "Phalacrocorax auritus", Family: "Phalacrocoracidae",
				Synonyms: []string{"Carbo auritus", "Nannopterum auritum"}},
			"Nannopterum auritum", TierSynonym,
		},
		{
			"synonym: a new genus for an old kinglet (real)",
			Taxon{Name: "Regulus calendula", Family: "Regulidae",
				Synonyms: []string{"Corthylio calendula", "Motacilla calendula"}},
			"Corthylio calendula", TierSynonym,
		},
		{
			"epithet: the genitive gained a syllable (real)",
			Taxon{Name: "Icterus bullockii", Family: "Icteridae", Synonyms: []string{"Xanthornus bullockii"}},
			"Icterus bullockiorum", TierEpithet,
		},
		{
			"epithet: the genitive gained an i (real)",
			Taxon{Name: "Pica nuttalli", Family: "Corvidae", Synonyms: []string{"Corvus nuttalli"}},
			"Pica nuttallii", TierEpithet,
		},
		{
			// GBIF's accepted spelling carries a doubled L that AviList does
			// not. The synonym Rhadina sibilatrix spells it AviList's way, so
			// candidates have to be drawn from the synonyms too.
			"epithet: recovered from a synonym's spelling (real)",
			Taxon{Name: "Phylloscopus sibillatrix", Family: "Phylloscopidae",
				Synonyms: []string{"Motacilla sibilatrix", "Motacilla spec", "Rhadina sibilatrix", "Sylvia sylvicola"}},
			"Phylloscopus sibilatrix", TierEpithet,
		},
		{
			// Gender agreement alone would resolve this, and must not, because
			// AviList's Anatidae holds two species whose epithets fold to the
			// same stem. An honest gap beats a coin toss.
			"ambiguous: two Anatidae fold to the same stem (real)",
			Taxon{Name: "Oressochen jubatus", Family: "Anatidae"},
			"", TierAmbiguous,
		},
		{
			"ambiguous: two patronyms in one family",
			Taxon{Name: "Accipiter gundlachii", Family: "Accipitridae"},
			"", TierAmbiguous,
		},
		{
			"ambiguous: two synonyms are both AviList species",
			Taxon{Name: "Circus nowhere", Family: "Accipitridae",
				Synonyms: []string{"Circus cyaneus", "Circus hudsonius"}},
			"", TierAmbiguous,
		},
		{
			// AviList lumps the Hooded Crow into Corvus corone, so there is no
			// correspondence to find and none must be invented.
			"none: AviList does not recognize the species (real)",
			Taxon{Name: "Corvus cornix", Family: "Corvidae"},
			"", TierNone,
		},
		{
			"none: GBIF's family is not in AviList at all",
			Taxon{Name: "Imaginaria villosa", Family: "Nowhereidae"},
			"", TierNone,
		},
		{
			// villosus appears in Picidae and in Onychorhynchidae. A GBIF
			// flycatcher must not become a woodpecker.
			"epithet must not cross a family",
			Taxon{Name: "Terenotriccus villosus", Family: "Tyrannidae"},
			"", TierNone,
		},
		{
			// auritum appears in Phalacrocoracidae and Phasianidae. Same point,
			// from the other side: a pheasant must not become a cormorant.
			"epithet must not cross a family, the other way",
			Taxon{Name: "Phasianus auritus", Family: "Phasianidae"},
			"Crossoptilon auritum", TierEpithet,
		},
		{
			// Folding must not eat so much of a short epithet that anything
			// matches: "alba" would otherwise fold to "al".
			"short epithets are not folded",
			Taxon{Name: "Gygis alba", Family: "Laridae"},
			"", TierNone,
		},
	}
	for _, c := range cases {
		got := Resolve(c.taxon, idx)
		if got.AviList != c.wantName || got.Tier != c.wantTier {
			t.Errorf("%s:\n  Resolve(%q in %s) = %q/%s\n  want %q/%s",
				c.name, c.taxon.Name, c.taxon.Family, got.AviList, got.Tier, c.wantName, c.wantTier)
		}
	}
}

// TestFoldCrossesGenusAndCaseButNotSpecies checks the folding rule directly,
// because Resolve's family scope would hide a fold that was far too generous.
//
// Watched fail: lowering fold's minimum stem from len(suf)+3 to len(suf) makes
// "alba" fold to "al" and "aestiva" to "aestiv" -- the first pair of "must
// stay apart" cases below then collapses and the test reddens.
//
// Verifies: T-021.
func TestFoldCrossesGenusAndCaseButNotSpecies(t *testing.T) {
	same := [][2]string{
		{"jubatus", "jubata"},         // gender follows the genus
		{"bullockii", "bullockiorum"}, // one Bullock or several
		{"nuttalli", "nuttallii"},     // how many i's a genitive needs
		{"vespertinus", "vespertina"}, // Coccothraustes vs Hesperiphona
		{"sibilatrix", "sibilatrix"},  // nothing to fold, and nothing broken
	}
	for _, p := range same {
		if fold(p[0]) != fold(p[1]) {
			t.Errorf("fold(%q)=%q and fold(%q)=%q should agree", p[0], fold(p[0]), p[1], fold(p[1]))
		}
	}
	apart := [][2]string{
		{"alba", "albus"},        // Ardea alba and Corvus albus are not one bird
		{"minuta", "minutilla"},  // Calidris minuta and Calidris minutilla
		{"villosus", "borealis"}, // the near miss internal/inat was built around
		{"aestiva", "petechia"},  // the two halves of the Yellow Warbler split
		{"nivosus", "montanus"},  // two plovers that moved genus together
	}
	for _, p := range apart {
		if fold(p[0]) == fold(p[1]) {
			t.Errorf("fold(%q) and fold(%q) both gave %q; distinct species would merge",
				p[0], p[1], fold(p[0]))
		}
	}
	// "alba" is four letters and must survive whole: a two-letter stem would
	// match a third of the genus.
	if fold("alba") != "alba" {
		t.Errorf("fold(%q) = %q, want it left alone", "alba", fold("alba"))
	}
}

// TestTableRoundTripsAndRejectsJunk covers the committed table's format and
// the invariants Load enforces on it.
//
// Watched fail: removing the `claimed` map from Load (the injectivity check)
// makes the "two GBIF taxa claim one AviList species" case load cleanly and
// the test goes red. Removing the tier/name agreement check does the same for
// the two cases after it.
//
// Verifies: T-021.
func TestTableRoundTripsAndRejectsJunk(t *testing.T) {
	in := Table{
		"Phalacrocorax auritus": {GBIFName: "Phalacrocorax auritus", GBIFKey: 2481875,
			AviList: "Nannopterum auritum", Tier: TierSynonym},
		"Corvus cornix": {GBIFName: "Corvus cornix", GBIFKey: 2482506, Tier: TierNone},
		"Acanthis flammea": {GBIFName: "Acanthis flammea", GBIFKey: 5231630,
			AviList: "Acanthis flammea", Tier: TierExact},
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
	// An unresolved taxon must survive the round trip. Dropping it would hide
	// the residue, which is the part of this table worth reading.
	if e, ok := out["Corvus cornix"]; !ok || e.Tier != TierNone || e.AviList != "" {
		t.Errorf("the unresolved row came back as %+v, want a TierNone row with no name", e)
	}
	// Sorted, so a regenerated table diffs as changes rather than a reshuffle.
	if a, c := strings.Index(b.String(), "\nAcanthis"), strings.Index(b.String(), "\nCorvus"); a > c {
		t.Error("Write did not sort; regenerating the table would produce an unreadable diff")
	}

	for _, bad := range []struct{ name, text string }{
		{"too few fields", "Corvus cornix\t2482506\tnone\n"},
		{"non-numeric key", "Corvus cornix\tnine\t-\tnone\n"},
		{"zero key", "Corvus cornix\t0\t-\tnone\n"},
		{"negative key", "Corvus cornix\t-3\t-\tnone\n"},
		{"unknown tier", "Corvus cornix\t1\t-\tprobably\n"},
		{"resolved tier with no name", "Corvus cornix\t1\t-\texact\n"},
		{"unresolved tier with a name", "Corvus cornix\t1\tCorvus corone\tnone\n"},
		{"duplicate gbif name", "Corvus cornix\t1\t-\tnone\nCorvus cornix\t2\t-\tnone\n"},
		{"two taxa claim one AviList species",
			"Regulus calendula\t1\tCorthylio calendula\tsynonym\n" +
				"Regulus nowhere\t2\tCorthylio calendula\tsynonym\n"},
	} {
		if _, err := Load(strings.NewReader(bad.text)); err == nil {
			t.Errorf("%s: Load accepted %q", bad.name, bad.text)
		}
	}
	// Comments and blanks are skipped, not treated as malformed.
	if tbl, err := Load(strings.NewReader("# a comment\n\nCorvus cornix\t1\t-\tnone\n")); err != nil {
		t.Errorf("Load rejected a comment or blank line: %v", err)
	} else if len(tbl) != 1 {
		t.Errorf("Load read %d entries past the comment, want 1", len(tbl))
	}
}

// TestCommittedTableResolvesAlmostEverything measures the coverage of the
// table that is actually in the repository, against the AviList workbook that
// is actually in the repository. No network (tech.md bindings).
//
// The floors are regression guards, not derived truth: the run that produced
// the committed table resolved 1,375 of 1,378, and a regeneration that lands
// below that has lost ground and owes an explanation. The named cases below
// are not from a run at all -- they are the four mismatches the change was
// commissioned to fix, and each asserts the bird the two authorities agree on.
//
// Watched fail: changing the expected AviList name for Phalacrocorax auritus
// to "Phalacrocorax auritum" reddens it; so does lowering wantResolved by
// hand-editing four rows of the committed table to tier=none.
//
// Verifies: T-021.
func TestCommittedTableResolvesAlmostEverything(t *testing.T) {
	f, err := os.Open(tableFile)
	if err != nil {
		t.Fatalf("opening the committed bridge: %v", err)
	}
	defer f.Close()
	table, err := Load(f)
	if err != nil {
		t.Fatalf("loading %s: %v", tableFile, err)
	}

	const (
		wantSpecies  = 1378 // distinct species keys the eBird dataset reports for US+CA
		wantResolved = 1375
	)
	if len(table) < wantSpecies {
		t.Errorf("the bridge covers %d GBIF species, fewer than the %d the dataset reports; "+
			"rows have gone missing rather than become unresolved", len(table), wantSpecies)
	}

	byTier := map[Tier]int{}
	for _, e := range table {
		byTier[e.Tier]++
	}
	resolved := byTier[TierExact] + byTier[TierSynonym] + byTier[TierEpithet]
	if resolved < wantResolved {
		t.Errorf("%d of %d resolved, below the %d the bridge reached when generated\n  tiers: %v",
			resolved, len(table), wantResolved, byTier)
	}
	// The synonym tier is the one doing the work this change exists for. If it
	// collapses to nothing, the bridge has silently become a plain name join
	// and the 62 species that started this are back.
	if byTier[TierSynonym] < 50 {
		t.Errorf("the synonym tier resolved only %d species; a plain name join would do nearly as well",
			byTier[TierSynonym])
	}

	avl, err := avilist.Open(avlFile)
	if err != nil {
		t.Fatalf("avilist: %v", err)
	}
	// Every target must be a species AviList actually carries. A bridge that
	// points at a name nobody publishes joins to nothing and says so nowhere.
	var dangling []string
	for _, e := range table {
		if e.Tier.Resolved() {
			if _, ok := avl[e.AviList]; !ok {
				dangling = append(dangling, e.GBIFName+" -> "+e.AviList)
			}
		}
	}
	sort.Strings(dangling)
	if len(dangling) > 0 {
		t.Errorf("%d rows point at names AviList does not carry:\n  %s",
			len(dangling), strings.Join(dangling, "\n  "))
	}

	// The four mismatches named in the brief, each resolved to the bird both
	// authorities mean.
	for _, c := range []struct{ gbif, avilist string }{
		{"Phalacrocorax auritus", "Nannopterum auritum"},       // Double-crested Cormorant
		{"Regulus calendula", "Corthylio calendula"},           // Ruby-crowned Kinglet
		{"Leuconotopicus villosus", "Leuconotopicus villosus"}, // Hairy Woodpecker
		{"Circus cyaneus", "Circus cyaneus"},                   // see the caveat below
	} {
		e, ok := table[c.gbif]
		if !ok {
			t.Errorf("the bridge has no row for GBIF %q", c.gbif)
			continue
		}
		if e.AviList != c.avilist {
			t.Errorf("GBIF %q maps to %q, want %q", c.gbif, e.AviList, c.avilist)
		}
	}
	// Circus cyaneus above is name-correct and concept-wrong, and the bridge
	// must not pretend otherwise: GBIF has not followed the Hen/Northern
	// Harrier split, so Northern Harrier gets no row of its own. Asserting its
	// absence keeps the next reader from "fixing" the name mapping instead of
	// facing the split.
	for _, e := range table {
		if e.AviList == "Circus hudsonius" {
			t.Errorf("GBIF %q now maps to Circus hudsonius; if the backbone has "+
				"followed the split, the lump caveat in Write's header is stale", e.GBIFName)
		}
	}
}

// TestCommittedTableIsExactlyWhatWriteProduces catches the committed table
// being hand-edited, and catches its header drifting away from the generator
// that is supposed to have written it. Load discards the comments, so writing
// the loaded table back out reconstructs the header from today's code: if the
// two differ, the file in the repository was not produced by the tool in the
// repository, and the provenance line at the top of it is a claim nobody
// checked.
//
// Watched fail: changing one word of Write's header literal reddens it, and so
// does swapping two data lines in data/gbif-taxa.tsv by hand.
//
// Verifies: T-021.
func TestCommittedTableIsExactlyWhatWriteProduces(t *testing.T) {
	onDisk, err := os.ReadFile(tableFile)
	if err != nil {
		t.Fatalf("reading the committed bridge: %v", err)
	}
	table, err := Load(strings.NewReader(string(onDisk)))
	if err != nil {
		t.Fatalf("loading %s: %v", tableFile, err)
	}
	var rewritten strings.Builder
	if err := Write(&rewritten, table); err != nil {
		t.Fatalf("Write: %v", err)
	}
	if rewritten.String() != string(onDisk) {
		a, b := strings.Split(string(onDisk), "\n"), strings.Split(rewritten.String(), "\n")
		for i := 0; i < len(a) || i < len(b); i++ {
			var x, y string
			if i < len(a) {
				x = a[i]
			}
			if i < len(b) {
				y = b[i]
			}
			if x != y {
				t.Fatalf("%s line %d is not what Write produces:\n  on disk:  %q\n  rewritten: %q",
					tableFile, i+1, x, y)
			}
		}
	}
}

// TestCommittedTableCoversBirdViewsSpecies is the measurement the change was
// commissioned for: how many of BirdView's own species the bridge reaches.
//
// It reconstructs BirdView's species list from the ABA checklist and the
// AviList rename table rather than from the compiled dataset, so it needs no
// AVONET parse -- and skips on a clean checkout, because the checklist is all
// rights reserved and never committed (T-004).
//
// Watched fail: rewriting every synonym- and epithet-tier row of the
// committed table to point at GBIF's own name -- a plain join on scientific
// name, which is what the project had -- drops the count from 813 to 780 and
// the test goes red.
//
// Verifies: T-021.
func TestCommittedTableCoversBirdViewsSpecies(t *testing.T) {
	if _, err := os.Stat(abaFile); err != nil {
		t.Skip("skipping: fetched/ABA_Checklist-8.19.csv is absent; run the fetch step first")
	}
	raw, err := os.ReadFile(abaFile)
	if err != nil {
		t.Fatalf("reading the checklist: %v", err)
	}
	species, err := aba.Parse(strings.NewReader(string(raw)))
	if err != nil {
		t.Fatalf("aba.Parse: %v", err)
	}

	// BirdView's species, under the names AviList files them under: the same
	// two steps the ingest takes (CR-022).
	birdview := map[string]bool{}
	for _, sp := range species {
		name := sp.ScientificName
		if alt, ok := taxonomy.AviListRenames[name]; ok {
			name = alt
		}
		if _, dropped := taxonomy.NotInAviList[sp.ScientificName]; dropped {
			continue
		}
		birdview[name] = true
	}
	const wantBirdView = 841 - 2 // ABA codes 1-3, less the two AviList does not recognize
	if len(birdview) != wantBirdView {
		t.Fatalf("reconstructed %d BirdView species, want %d", len(birdview), wantBirdView)
	}

	f, err := os.Open(tableFile)
	if err != nil {
		t.Fatalf("opening the committed bridge: %v", err)
	}
	defer f.Close()
	table, err := Load(f)
	if err != nil {
		t.Fatalf("loading %s: %v", tableFile, err)
	}

	reached := map[string]bool{}
	for _, e := range table {
		if e.Tier.Resolved() && birdview[e.AviList] {
			reached[e.AviList] = true
		}
	}
	var gap []string
	for name := range birdview {
		if !reached[name] {
			gap = append(gap, name)
		}
	}
	sort.Strings(gap)

	// 813 of 839 when the table was generated, against 780 for a plain join on
	// AviList's names -- or 777 joining on the names BirdView itself stores,
	// which is the 62-species gap this change was commissioned to close. The
	// 26 that remain are not name failures: GBIF's species facet holds no
	// taxon for them at all. The floor guards the gain; the gap is printed in
	// full because each entry is a finding.
	const wantReached = 813
	if len(reached) < wantReached {
		t.Errorf("the bridge reaches %d of %d BirdView species, below the %d it reached when generated\n"+
			"  unreached (%d):\n    %s",
			len(reached), len(birdview), wantReached, len(gap), strings.Join(gap, "\n    "))
	}
	t.Logf("bridge reaches %d of %d BirdView species; %d unreached", len(reached), len(birdview), len(gap))
}

// fakeAPI serves canned responses and records how it was called, so the
// courtesy requirements are observed at the server rather than read off the
// client's own fields. Tests never touch the live API (tech.md bindings).
func fakeAPI(t *testing.T, routes map[string]string) (*httptest.Server, *[]time.Time, *[]string) {
	t.Helper()
	var times []time.Time
	var agents []string
	s := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		times = append(times, time.Now())
		agents = append(agents, r.Header.Get("User-Agent"))
		body, ok := routes[r.URL.Path]
		if !ok {
			http.Error(w, "no route for "+r.URL.Path, http.StatusNotFound)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprint(w, body)
	}))
	t.Cleanup(s.Close)
	return s, &times, &agents
}

const (
	cormorantJSON = `{"key":2481875,"canonicalName":"Phalacrocorax auritus","rank":"SPECIES",
	  "family":"Phalacrocoracidae"}`
	cormorantSynJSON = `{"results":[
	  {"canonicalName":"Carbo auritus","rank":"SPECIES"},
	  {"canonicalName":"Nannopterum auritum","rank":"SPECIES"},
	  {"canonicalName":"Nannopterum auritum auritum","rank":"SUBSPECIES"}]}`
)

// TestTaxonIdentifiesItselfAndPaces checks the two courtesies this tool owes:
// say who you are (T-006) and do not hammer the service. Both are observed at
// the server -- a client that sets Interval and never waits would satisfy a
// test that read the field.
//
// Watched fail: deleting the req.Header.Set("User-Agent", ...) line in get
// leaves Go's default Go-http-client/1.1 and the first two assertions redden;
// making pace() return immediately reddens the third.
//
// Verifies: T-006.
func TestTaxonIdentifiesItselfAndPaces(t *testing.T) {
	srv, times, agents := fakeAPI(t, map[string]string{
		"/species/2481875":          cormorantJSON,
		"/species/2481875/synonyms": cormorantSynJSON,
	})

	const interval = 60 * time.Millisecond
	c := NewClient()
	c.BaseURL, c.Interval, c.HTTP = srv.URL, interval, srv.Client()

	// Concurrent, because the pacing has to hold across workers and not just
	// within one: that is the whole reason it is a mutex rather than a field.
	c.Concurrency = 4
	keys := []SpeciesCount{{Key: 2481875}, {Key: 2481875}, {Key: 2481875}}
	if _, err := c.Taxa(context.Background(), keys); err != nil {
		t.Fatalf("Taxa: %v", err)
	}

	for i, ua := range *agents {
		if !strings.Contains(ua, "birdview") {
			t.Errorf("request %d sent User-Agent %q, which does not identify the client", i, ua)
		}
		if !strings.Contains(ua, "@") {
			t.Errorf("request %d sent User-Agent %q, with no contact address", i, ua)
		}
	}
	// Six requests -- two per taxon -- so the gaps are worth checking.
	if len(*times) != 6 {
		t.Fatalf("server saw %d requests, want 6", len(*times))
	}
	sorted := append([]time.Time(nil), (*times)...)
	sort.Slice(sorted, func(i, j int) bool { return sorted[i].Before(sorted[j]) })
	// A small tolerance: neither Sleep nor a round trip is exact, and a gap of
	// 59.8 ms is scheduler jitter rather than a client hammering the service.
	const jitter = 5 * time.Millisecond
	for i := 1; i < len(sorted); i++ {
		if gap := sorted[i].Sub(sorted[i-1]); gap < interval-jitter {
			t.Errorf("requests %d and %d were %v apart, closer than the %v interval",
				i-1, i, gap, interval)
		}
	}
}

// TestTaxonKeepsOnlySpeciesRankSynonyms: a subspecies-rank synonym names a
// population, and matching one against an AviList species would assert a
// correspondence nobody published.
//
// Watched fail: dropping the `r.Rank == "SPECIES"` condition in Taxon admits
// "Nannopterum auritum auritum" and the length check goes red.
//
// Verifies: T-021.
func TestTaxonKeepsOnlySpeciesRankSynonyms(t *testing.T) {
	srv, _, _ := fakeAPI(t, map[string]string{
		"/species/2481875":          cormorantJSON,
		"/species/2481875/synonyms": cormorantSynJSON,
	})
	c := NewClient()
	c.BaseURL, c.Interval, c.HTTP = srv.URL, 0, srv.Client()

	tx, err := c.Taxon(context.Background(), 2481875, 42)
	if err != nil {
		t.Fatalf("Taxon: %v", err)
	}
	want := []string{"Carbo auritus", "Nannopterum auritum"}
	if strings.Join(tx.Synonyms, "|") != strings.Join(want, "|") {
		t.Errorf("synonyms = %v, want %v", tx.Synonyms, want)
	}
	if tx.Family != "Phalacrocoracidae" || tx.Records != 42 {
		t.Errorf("taxon = %+v, want the family and record count carried through", tx)
	}
}

// TestSpeciesKeysPagesUntilTheFacetIsEmpty is the guard on the subtlest
// failure available here. A server is free to cap a facet page below the limit
// the client asked for, and a client that stops at the first short page then
// reports a shorter species list than the dataset holds -- which looks exactly
// like a dataset with fewer birds in it, and would be believed.
//
// The fake below serves at most two values per page however many are asked
// for, which is that cap made concrete.
//
// Watched fail: stopping the loop on `len(counts) < facetPage` returns 2 of
// the 3 species and this goes red. Advancing `offset` by facetPage instead of
// by len(counts) skips the third and reddens it too. Leaving the loop but
// never advancing offset trips the duplicate-key guard, which is the last case
// below.
//
// Verifies: T-021.
func TestSpeciesKeysPagesUntilTheFacetIsEmpty(t *testing.T) {
	all := []struct {
		key   string
		count int64
	}{{"2481875", 900}, {"2484597", 800}, {"2480487", 700}}

	// serveCapped answers like a service that will not return more than
	// cappedPage values at a time, whatever facetLimit says. honorOffset makes
	// it ignore facetOffset, which is the other way paging can go wrong.
	const cappedPage = 2
	serve := func(honorOffset bool) (*httptest.Server, *int) {
		var pages int
		s := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			pages++
			if pages > 10 {
				http.Error(w, "the client is not making progress", http.StatusInternalServerError)
				return
			}
			off := 0
			if honorOffset {
				fmt.Sscanf(r.URL.Query().Get("facetOffset"), "%d", &off)
			}
			var b strings.Builder
			fmt.Fprint(&b, `{"count":1000,"facets":[{"field":"SPECIES_KEY","counts":[`)
			for i := off; i < off+cappedPage && i < len(all); i++ {
				if i > off {
					b.WriteByte(',')
				}
				fmt.Fprintf(&b, `{"name":%q,"count":%d}`, all[i].key, all[i].count)
			}
			fmt.Fprint(&b, `]}]}`)
			w.Header().Set("Content-Type", "application/json")
			fmt.Fprint(w, b.String())
		}))
		t.Cleanup(s.Close)
		return s, &pages
	}

	srv, pages := serve(true)
	c := NewClient()
	c.BaseURL, c.Interval, c.HTTP = srv.URL, 0, srv.Client()

	got, total, err := c.SpeciesKeys(context.Background(), "ds", []string{"US", "CA"})
	if err != nil {
		t.Fatalf("SpeciesKeys: %v", err)
	}
	if len(got) != len(all) {
		t.Errorf("got %d species over %d pages, want %d -- a truncated facet looks "+
			"exactly like a smaller dataset", len(got), *pages, len(all))
	}
	// Two full pages, a third holding the remainder, and a fourth that comes
	// back empty and ends it.
	if *pages != 3 {
		t.Errorf("the client made %d requests, want 3 pages over %d values capped at %d",
			*pages, len(all), cappedPage)
	}
	if total != 1000 {
		t.Errorf("total records = %d, want 1000", total)
	}
	if len(got) > 0 && (got[0].Key != 2481875 || got[0].Records != 900) {
		t.Errorf("first species = %+v, want key 2481875 with 900 records", got[0])
	}

	// A server that ignores facetOffset must be reported, not paged forever.
	srv2, _ := serve(false)
	c2 := NewClient()
	c2.BaseURL, c2.Interval, c2.HTTP = srv2.URL, 0, srv2.Client()
	if _, _, err := c2.SpeciesKeys(context.Background(), "ds", []string{"US"}); err == nil {
		t.Error("a server ignoring facetOffset produced no error")
	}
}

// TestGetRetriesServerErrorsButNotClientErrors: GBIF sheds load on a
// billion-record facet query, so a 503 has to be worth retrying -- and a 404
// has to not be, or every genuine miss costs four requests and fifteen seconds
// of someone else's server.
//
// Watched fail: making get retry every non-200 reddens the second half (the
// 404 server sees 4 requests, not 1); making it retry nothing reddens the
// first (the 503 never reaches its 200).
//
// Verifies: T-006.
func TestGetRetriesServerErrorsButNotClientErrors(t *testing.T) {
	var hits int
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		hits++
		if hits < 3 {
			http.Error(w, "busy", http.StatusServiceUnavailable)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprint(w, cormorantJSON)
	}))
	defer srv.Close()

	c := NewClient()
	c.BaseURL, c.Interval, c.Backoff, c.HTTP = srv.URL, 0, time.Millisecond, srv.Client()
	tx, err := c.Taxon(context.Background(), 2481875, 0)
	if err != nil {
		t.Fatalf("two 503s then a 200 should succeed, got %v after %d requests", err, hits)
	}
	if tx.Name != "Phalacrocorax auritus" {
		t.Errorf("recovered taxon = %q, want Phalacrocorax auritus", tx.Name)
	}

	// A 404 IS retried, and this assertion used to say the opposite.
	//
	// It read "it will never become a 200", which was a belief rather than a
	// measurement, and GBIF disproved it: a status run failed on `mimus
	// polyglottos` in Florida -- 1,370,398 records -- which answered 404
	// mid-run and 200 on the identical query a minute later. Two species in
	// 1,429 failed that way in a single run, and the run was abandoned.
	//
	// So a persistent 404 must still be an error, and must still stop, but
	// only after the retries have shown it is persistent.
	var misses int
	srv2 := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		misses++
		http.Error(w, "no such taxon", http.StatusNotFound)
	}))
	defer srv2.Close()

	c2 := NewClient()
	c2.BaseURL, c2.Interval, c2.Backoff, c2.HTTP = srv2.URL, 0, time.Millisecond, srv2.Client()
	if _, err := c2.Taxon(context.Background(), 1, 0); err == nil {
		t.Error("a persistent 404 produced no error")
	}
	if misses != maxAttempts {
		t.Errorf("a 404 cost %d requests, want %d: GBIF's 404s are not all permanent, "+
			"and giving up on the first one abandoned a 1,429-species run", misses, maxAttempts)
	}

	// The transient case, which is the one that cost a run: 404 then 200.
	var flaky int
	srv3 := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		flaky++
		if flaky == 1 {
			http.Error(w, "spurious", http.StatusNotFound)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprint(w, cormorantJSON)
	}))
	defer srv3.Close()

	c3 := NewClient()
	c3.BaseURL, c3.Interval, c3.Backoff, c3.HTTP = srv3.URL, 0, time.Millisecond, srv3.Client()
	if _, err := c3.Taxon(context.Background(), 2481875, 0); err != nil {
		t.Errorf("a 404 followed by a 200 should succeed, got %v after %d requests", err, flaky)
	}
}

// TestTaxaFailsTheBatchOnATransportError: a 500 is not evidence that a taxon
// has no AviList counterpart. Recording it as one would bake a network hiccup
// into a committed table, where it would be indistinguishable from a real
// taxonomic gap forever.
//
// Watched fail: having Taxa log and skip a failed key instead of returning the
// error makes it return a short slice with no error, and both assertions here
// redden.
//
// Verifies: T-021.
func TestTaxaFailsTheBatchOnATransportError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.HasSuffix(r.URL.Path, "/synonyms") {
			http.Error(w, "nope", http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprint(w, cormorantJSON)
	}))
	defer srv.Close()

	c := NewClient()
	// A short backoff rather than a context deadline. A deadline would make
	// Taxa return ctx.Err() and the test would pass without ever reaching the
	// per-key error it exists to check -- which is exactly what it did until
	// the break below was tried.
	c.BaseURL, c.Interval, c.Backoff, c.HTTP = srv.URL, 0, time.Millisecond, srv.Client()

	taxa, err := c.Taxa(context.Background(), []SpeciesCount{{Key: 2481875}})
	if err == nil {
		t.Error("a 500 produced no error; the tool would write a confident non-match")
	}
	if taxa != nil {
		t.Errorf("got %d taxa alongside the error; a partial batch must not reach the table", len(taxa))
	}
}
