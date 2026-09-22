// Command birdview-ingest compiles the BirdView dataset from its sources.
//
// The output is a build artifact and is never committed: it carries values
// derived from the ABA checklist, which is all rights reserved (T-004, T-005).
package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"birdview/internal/aba"
	"birdview/internal/avonet"
	"birdview/internal/dataset"
	"birdview/internal/morphology"
	"birdview/internal/palette"
	"birdview/internal/taxonomy"
	"birdview/internal/wingspan"
)

// maxUnmatchedFraction fails the build when too many species cannot be joined.
// Without a loud failure a taxonomy update silently shrinks the dataset and
// nobody notices (P-007).
const maxUnmatchedFraction = 0.02

func main() {
	var (
		abaPath = flag.String("aba", "fetched/ABA_Checklist-8.19.csv", "ABA checklist CSV (fetched, never committed)")
		avoPath = flag.String("avonet", "AVONET Supplementary dataset 1.xlsx", "AVONET workbook")
		foxPath = flag.String("wingspan", "wingspan_fox2026_AppendixS1.xlsx", "Fox et al. wingspan workbook")
		outPath = flag.String("out", "web/data/species.json", "output path (a build artifact)")
	)
	flag.Parse()

	ds, report, err := build(*abaPath, *avoPath, *foxPath)
	if err != nil {
		log.Fatalf("birdview-ingest: %v", err)
	}
	fmt.Print(report)

	if err := os.MkdirAll(filepath.Dir(*outPath), 0o755); err != nil {
		log.Fatalf("birdview-ingest: %v", err)
	}
	f, err := os.Create(*outPath)
	if err != nil {
		log.Fatalf("birdview-ingest: %v", err)
	}
	defer f.Close()
	enc := json.NewEncoder(f)
	enc.SetIndent("", "  ")
	if err := enc.Encode(ds); err != nil {
		log.Fatalf("birdview-ingest: encoding %s: %v", *outPath, err)
	}
	fmt.Printf("wrote %d species to %s\n", len(ds.Species), *outPath)
}

// resolution records which source row a species' traits came from, and how
// well it corresponds to the species itself.
type resolution struct {
	traits avonet.Traits
	source string
	taxon  string
	kind   dataset.Kind
}

// resolve finds trait data for one species, preferring an exact taxon match in
// any AVONET taxonomy over borrowing a relative's measurements (CR-006).
func resolve(sp aba.Species, wb *avonet.Workbook) (resolution, bool) {
	try := func(name, sheet string, tbl map[string]avonet.Traits, kind dataset.Kind) (resolution, bool) {
		if t, ok := tbl[name]; ok && t.Culmen > 0 {
			return resolution{t, sheet, name, kind}, true
		}
		return resolution{}, false
	}

	// 1. The species' own name, in the eBird spine.
	if r, ok := try(sp.ScientificName, avonet.SheetEBird, wb.EBird, dataset.Measured); ok {
		return r, true
	}
	// 2. The AOS/Clements alternative the checklist itself supplies.
	if sp.AltScientificName != "" {
		if r, ok := try(sp.AltScientificName, avonet.SheetEBird, wb.EBird, dataset.Measured); ok {
			return r, true
		}
	}
	// 3. Its own row in another AVONET taxonomy, or under a known pre-revision
	// label -- either way still this species, so still measured.
	sheets := []struct {
		name string
		m    map[string]avonet.Traits
	}{{avonet.SheetEBird, wb.EBird}, {avonet.SheetBirdLife, wb.BirdLife}, {avonet.SheetBirdTree, wb.BirdTree}}
	names := append([]string{sp.ScientificName}, taxonomy.Renames[sp.ScientificName]...)
	for _, name := range names {
		for _, sh := range sheets {
			if r, ok := try(name, sh.name, sh.m, dataset.Measured); ok {
				return r, true
			}
		}
	}
	// 4. A pre-split parent: real data, different bird.
	if parent, ok := taxonomy.NearMatches[sp.ScientificName]; ok {
		for _, sh := range sheets {
			if r, ok := try(parent, sh.name, sh.m, dataset.NearMatch); ok {
				return r, true
			}
		}
	}
	return resolution{}, false
}

func build(abaPath, avoPath, foxPath string) (*dataset.Dataset, string, error) {
	raw, err := os.ReadFile(abaPath)
	if err != nil {
		return nil, "", fmt.Errorf("reading the ABA checklist: %w (fetch it to fetched/; it is never committed)", err)
	}
	species, err := aba.Parse(strings.NewReader(string(raw)))
	if err != nil {
		return nil, "", err
	}
	wb, err := avonet.Open(avoPath)
	if err != nil {
		return nil, "", err
	}
	fox, spanModel, err := wingspan.Open(foxPath)
	if err != nil {
		return nil, "", err
	}

	// Families in the checklist's own taxonomic sequence, with the order each
	// belongs to. Order is resolved through the same taxon lookup as traits:
	// taking it from the ABA name alone leaves Tytonidae unplaced, because
	// AVONET files that bird as Tyto alba (P-006).
	var famSeq []string
	famCommon := map[string]string{}
	famOrder := map[string]string{}
	seenFam := map[string]bool{}
	for _, sp := range species {
		if !seenFam[sp.Family] {
			seenFam[sp.Family] = true
			famSeq = append(famSeq, sp.Family)
			famCommon[sp.Family] = sp.FamilyCommon
		}
		if famOrder[sp.Family] == "" {
			if r, ok := resolve(sp, wb); ok {
				famOrder[sp.Family] = wb.OrderOf[r.taxon]
			}
		}
	}
	for _, f := range famSeq {
		if famOrder[f] == "" {
			return nil, "", fmt.Errorf("family %s could not be placed in an order; "+
				"the selector groups by order and has nowhere to put it", f)
		}
	}

	// One palette for orders, and one per order for its families. Family colors
	// need only be distinct within their order, which is what makes 91 families
	// fit a gamut that holds far fewer (P-005).
	var orderSeq []string
	byOrder := map[string][]string{}
	for _, f := range famSeq {
		o := famOrder[f]
		if len(byOrder[o]) == 0 {
			orderSeq = append(orderSeq, o)
		}
		byOrder[o] = append(byOrder[o], f)
	}
	orderColors := palette.Generate(orderSeq, 0)
	famColors := map[string]palette.Swatch{}
	for i, o := range orderSeq {
		// A different seed per order keeps families in different orders from
		// being handed the same color for no reason, though they are allowed to.
		for f, sw := range palette.Generate(byOrder[o], float64(i)*47.0) {
			famColors[f] = sw
		}
	}

	var (
		out       []dataset.Species
		unmatched []string
		counts    = map[string]int{}
	)

	for _, sp := range species {
		res, ok := resolve(sp, wb)
		if !ok {
			unmatched = append(unmatched, fmt.Sprintf("%s (%s)", sp.CommonName, sp.ScientificName))
			continue
		}
		counts[string(res.kind)]++

		traitProv := func() dataset.Provenance {
			return dataset.Provenance{
				Kind:        res.kind,
				Source:      "AVONET v7, sheet " + res.source,
				SourceTaxon: res.taxon,
			}
		}

		// Segments: bill and tail measured, head/neck/body modeled.
		headV, headModel := morphology.HeadLength(res.traits.Mass)
		bodyV, bodyModel := morphology.BodyLength(res.traits.Mass)
		neckV, neckModel := morphology.NeckLength(bodyV, sp.Family)
		bill, head, neck, body, tail, total := morphology.BalanceToSum(
			res.traits.Culmen, headV, neckV, bodyV, res.traits.Tail)

		modeled := func(model string) dataset.Provenance {
			return dataset.Provenance{Kind: dataset.Modeled, Model: model}
		}

		rec := dataset.Species{
			ID:             id(sp.ScientificName),
			CommonName:     sp.CommonName,
			BandingCode:    sp.Code,
			ScientificName: sp.ScientificName,
			Family:         sp.Family,
			FamilyCommon:   sp.FamilyCommon,
			FamilyColor:    famColors[sp.Family].Hex,
			FamilyText:     famColors[sp.Family].Text,
			Order:          famOrder[sp.Family],
			Segments: dataset.Segments{
				Bill: dataset.Metric{Value: bill, Prov: traitProv()},
				Head: dataset.Metric{Value: head, Prov: modeled(headModel)},
				Neck: dataset.Metric{Value: neck, Prov: modeled(neckModel)},
				Body: dataset.Metric{Value: body, Prov: modeled(bodyModel)},
				Tail: dataset.Metric{Value: tail, Prov: traitProv()},
			},
			Mass:   dataset.Metric{Value: morphology.Round1(res.traits.Mass), Prov: traitProv()},
			Tarsus: dataset.Metric{Value: morphology.Round1(res.traits.Tarsus), Prov: traitProv()},
			TotalLength: dataset.Metric{
				Value: total,
				Prov: modeled("sum of the five diagram segments; bill and tail measured, " +
					"head, neck and trunk modeled"),
			},
		}

		// Wingspan: empirical where Fox has one for a BirdLife name this
		// species maps to, modeled otherwise.
		//
		// The kind is inherited from the trait resolution. If traits came from
		// a pre-split parent, the crosswalk leads to that parent's BirdLife
		// name, so any wingspan found through it is the parent's measurement
		// too -- real data, wrong bird. Labeling it measured because the source
		// is empirical would be exactly the error P-062 exists to catch.
		var found bool
		for _, bl := range append([]string{sp.ScientificName}, wb.EBirdToBirdLife[res.taxon]...) {
			r, ok := fox[bl]
			if !ok {
				continue
			}
			kind := dataset.Measured
			if res.kind == dataset.NearMatch {
				kind = dataset.NearMatch
			}
			rec.Wingspan = dataset.Metric{
				Value: morphology.Round1(r.Mid * 10), // Fox publishes centimeters
				Min:   morphology.Round1(r.Min * 10),
				Max:   morphology.Round1(r.Max * 10),
				Prov: dataset.Provenance{
					Kind: kind, Source: wingspan.Source, SourceTaxon: r.Taxon,
				},
			}
			found = true
			counts["wingspan "+string(kind)]++
			break
		}
		if !found {
			v, model := spanModel.Predict(wb.OrderOf[res.taxon], res.traits.WingChord)
			if v <= 0 {
				return nil, "", fmt.Errorf("%s: no wingspan available and no wing chord to model from",
					sp.ScientificName)
			}
			rec.Wingspan = dataset.Metric{Value: morphology.Round1(v), Prov: modeled(model)}
			counts["wingspan modeled"]++
		}

		out = append(out, rec)
	}

	sort.Slice(out, func(i, j int) bool { return out[i].ScientificName < out[j].ScientificName })

	if frac := float64(len(unmatched)) / float64(len(species)); frac > maxUnmatchedFraction {
		return nil, "", fmt.Errorf(
			"%d of %d species (%.1f%%) could not be joined to trait data, above the %.0f%% threshold.\n"+
				"A taxonomy has probably moved; review internal/taxonomy/mapping.go.\nUnmatched:\n  %s",
			len(unmatched), len(species), frac*100, maxUnmatchedFraction*100,
			strings.Join(unmatched, "\n  "))
	}

	var b strings.Builder
	fmt.Fprintf(&b, "ABA codes 1-%d: %d species in %d families across %d orders\n",
		aba.MaxRarityCode, len(species), len(famSeq), len(orderSeq))
	fmt.Fprintf(&b, "  traits measured   : %d\n", counts[string(dataset.Measured)])
	fmt.Fprintf(&b, "  traits near-match : %d\n", counts[string(dataset.NearMatch)])
	fmt.Fprintf(&b, "  wingspan measured : %d\n", counts["wingspan measured"])
	fmt.Fprintf(&b, "  wingspan near-match: %d\n", counts["wingspan near-match"])
	fmt.Fprintf(&b, "  wingspan modeled  : %d\n", counts["wingspan modeled"])
	fmt.Fprintf(&b, "  length            : %d modeled (no usable measured source -- CR-002)\n", len(out))
	if len(unmatched) > 0 {
		fmt.Fprintf(&b, "  unmatched (%d, below threshold):\n    %s\n",
			len(unmatched), strings.Join(unmatched, "\n    "))
	}
	fmt.Fprintf(&b, "  awaiting taxonomic check (CR-006): %s\n",
		strings.Join(taxonomy.AwaitingTaxonomicCheck, ", "))

	fams := make([]dataset.Family, 0, len(famSeq))
	for _, f := range famSeq {
		o := famOrder[f]
		common, ok := taxonomy.OrderCommonNames[o]
		if !ok {
			return nil, "", fmt.Errorf("order %s has no common name; add one to "+
				"internal/taxonomy so the selector can label its group", o)
		}
		fams = append(fams, dataset.Family{
			Name: f, Common: famCommon[f], Order: o, OrderCommon: common,
			Color: famColors[f].Hex, TextColor: famColors[f].Text,
			OrderColor: orderColors[o].Hex, OrderText: orderColors[o].Text,
		})
	}

	return &dataset.Dataset{
		Families: fams,
		SourceVersions: map[string]string{
			"aba":          "ABA Checklist 8.19 (January 2026)",
			"avonet":       "AVONET v7, doi:10.6084/m9.figshare.16586228",
			"fox-wingspan": "Fox et al. 2026 v1, doi:10.6084/m9.figshare.32089083",
		},
		Species: out,
	}, b.String(), nil
}

func id(sci string) string {
	return strings.ToLower(strings.ReplaceAll(sci, " ", "_"))
}
