// Command birdview serves the size explorer on localhost.
//
// Rendering happens on the server, in html/template. That keeps the ordering,
// sectioning and geometry rules in Go where they are property-tested, rather
// than duplicating them in a language the acceptance criteria cannot reach.
// Filters and toggles are query parameters, and a tile's detail view is a URL.
//
// Verifies: T-001, T-002, T-003.
package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"html/template"
	"io/fs"
	"log"
	"math"
	"net/http"
	"os"
	"sort"
	"strconv"
	"strings"

	"birdview/internal/dataset"
	"birdview/internal/inat"
	"birdview/internal/view"
	"birdview/web"
)

func main() {
	addr := flag.String("addr", "localhost:8080", "listen address")
	data := flag.String("data", "web/data/species.json", "compiled dataset (a build artifact)")
	flag.Parse()

	ds, err := load(*data)
	if err != nil {
		log.Fatalf("birdview: %v\n\nRun birdview-ingest first; the dataset is a build artifact "+
			"and is not committed (T-005).", err)
	}
	srv, err := newServer(ds)
	if err != nil {
		log.Fatalf("birdview: %v", err)
	}
	fmt.Printf("BirdView: %d species, serving http://%s\n", len(ds.Species), *addr)
	if err := http.ListenAndServe(*addr, srv); err != nil {
		log.Fatalf("birdview: %v", err)
	}
}

func load(path string) (*dataset.Dataset, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	var ds dataset.Dataset
	if err := json.NewDecoder(f).Decode(&ds); err != nil {
		return nil, fmt.Errorf("decoding %s: %w", path, err)
	}
	if len(ds.Species) == 0 {
		return nil, fmt.Errorf("%s contains no species", path)
	}
	return &ds, nil
}

type server struct {
	mux  *http.ServeMux
	ds   *dataset.Dataset
	tmpl *template.Template
	byID map[string]*dataset.Species
	fams []familyOption
}

type familyOption struct {
	Name, Common     string
	Color, TextColor string
	Order            string
	OrderCommon      string
	Count            int
}

// orderGroup is one block of the family selector. Grouping by order in the
// checklist's taxonomic sequence matches how every field guide is arranged,
// and it is what lets family colors repeat across orders (P-010, P-005).
type orderGroup struct {
	Order    string
	Common   string
	Count    int
	Selected bool         // every family in the order is filtered in
	URL      template.URL // toggles the whole group
	Families []familyRow
}

func (s *server) ServeHTTP(w http.ResponseWriter, r *http.Request) { s.mux.ServeHTTP(w, r) }

func newServer(ds *dataset.Dataset) (*server, error) {
	tmpl, err := template.New("").Funcs(funcs).ParseFS(web.Static, "static/*.html")
	if err != nil {
		return nil, fmt.Errorf("parsing templates: %w", err)
	}
	s := &server{mux: http.NewServeMux(), ds: ds, tmpl: tmpl, byID: map[string]*dataset.Species{}}

	counts := map[string]int{}
	for i := range ds.Species {
		sp := &ds.Species[i]
		s.byID[sp.ID] = sp
		counts[sp.Family]++
	}
	// ds.Families is already in ABA taxonomic sequence; preserve it rather than
	// alphabetizing, which is what the checklist and every field guide use.
	for _, f := range ds.Families {
		s.fams = append(s.fams, familyOption{
			Name: f.Name, Common: f.Common, Color: f.Color, TextColor: f.TextColor,
			Order: f.Order, OrderCommon: f.OrderCommon, Count: counts[f.Name],
		})
	}

	assets, err := fs.Sub(web.Static, "static")
	if err != nil {
		return nil, err
	}
	s.mux.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.FS(assets))))
	s.mux.HandleFunc("/compare", s.compare)
	s.mux.HandleFunc("/species/", s.detail)
	s.mux.HandleFunc("/", s.index)
	return s, nil
}

// options are the user's current selections, all carried in the query string so
// that any view is a shareable URL.
type options struct {
	Metric view.Metric
	Step   float64
	Scale  view.ScaleMode
	// Families holds the explicitly named families. HasSelection distinguishes
	// a fresh page, where nothing is named and everything is selected, from an
	// explicitly emptied selection, where nothing is named and nothing is
	// selected. The two look identical in the query string otherwise, and they
	// mean opposite things.
	Families     map[string]bool
	HasSelection bool
	Metric3      bool // imperial when false
}

func parseOptions(r *http.Request) options {
	q := r.URL.Query()
	// Per-section is the default (P-042, CR-017). Global makes the cross-family
	// argument better and is one click away, but at global scale a median bird
	// is a couple of pixels tall, and none of the diagram's vocabulary --
	// dashed for modeled, the chord ticks, the labels -- is legible there. The
	// default that teaches the reader to read a bird wins over the one that
	// delivers the shock before they can.
	o := options{Metric: view.ByLength, Step: view.DefaultStepPercent, Scale: view.PerSection,
		Families: map[string]bool{}}
	switch view.Metric(q.Get("by")) {
	case view.ByWingspan:
		o.Metric = view.ByWingspan
	case view.ByMass:
		o.Metric = view.ByMass
	}
	// Recognise the NON-default explicitly. Written the other way round this
	// parses only the value that is already the default, so flipping the
	// default silently disables the toggle -- which is exactly what happened
	// when CR-017 moved it.
	if view.ScaleMode(q.Get("scale")) == view.Global {
		o.Scale = view.Global
	}
	if v, err := strconv.ParseFloat(q.Get("step"), 64); err == nil && v >= 2 && v <= 100 {
		o.Step = v
	}
	o.Metric3 = q.Get("units") == "metric"
	_, o.HasSelection = q["family"]
	for _, f := range q["family"] {
		if f != "" {
			o.Families[f] = true
		}
	}
	return o
}

// tile is one rendered species. The three footer metrics are formatted here
// rather than in the template: a template that does its own arithmetic cannot
// see the unit toggle, which is exactly how the footer came to ignore it.
type tile struct {
	S       *dataset.Species
	Geom    view.Geometry
	Box     view.Box
	Length  string
	Wing    string
	MassStr string

	// WingModeled drives the crossbar split: the inner span is dashed only when
	// the wingspan is modeled (P-065).
	WingModeled bool

	// Labels and ChordLabel are empty on an index tile: the annotations of P-066
	// need room, so they appear only in the detail view and in the help legend,
	// both of which build this struct through labelledTile.
	Labels     []diagramLabel
	TopOffset  float64
	ChordLabel string
	// WingspanLabel sits at the right wingtip, mirroring the chord label at the
	// left, so the crossbar names the whole span as well as its measured part.
	WingspanLabel string
	WingspanEst   bool
	// TotalLabel names the whole bird, centred on the axis below the tail tip
	// (P-066). It is the bird's reported total length, never the drawn height:
	// a folded neck makes those differ by 6-9%.
	//
	// TotalLabelY is its baseline, computed here rather than in the template so
	// it cannot drift from the strip reserved for it, and so it is testable.
	TotalLabel  string
	TotalEst    bool
	TotalLabelY float64
}

type sectionView struct {
	Index    int
	Min, Max string
	Count    int
	Tiles    []tile
}

// familyRow is one filter chip, with the URL that toggles it already built:
// a multi-valued query parameter is awkward to assemble in a template.
type familyRow struct {
	familyOption
	Selected bool
	URL      template.URL
}

type pageData struct {
	Sections []sectionView
	Groups   []orderGroup
	Opt      options
	Total    int
	Shown    int
	Query    func(...string) template.URL
	// AllSelected is true only when every family is explicitly filtered in --
	// not when nothing is selected. Both states display the whole dataset, but
	// they are different selections, and the control reflects the selection.
	AllSelected bool
	AllURL      template.URL
	// NothingSelected drives the empty state, which has to say something
	// different from "no matches": the user turned everything off.
	NothingSelected bool

	// Legend is the labeled example diagram the help opens with (P-035). Nil
	// only if no suitable species exists, in which case the help degrades to
	// prose rather than breaking.
	Legend *tile
}

// tileBox is the drawing area for one diagram. Roughly 2.6:1, because a
// to-scale wingspan runs up to ~2.5x body length and P-043 draws it at the
// vertical scale rather than shrinking it to fit.
var tileBox = view.Box{W: 380, H: 148}

func (s *server) index(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		http.NotFound(w, r)
		return
	}
	opt := parseOptions(r)

	// A fresh page selects everything; an emptied selection selects nothing.
	selected := opt.Families
	if !opt.HasSelection {
		selected = make(map[string]bool, len(s.fams))
		for _, f := range s.fams {
			selected[f.Name] = true
		}
	}
	var filtered []*dataset.Species
	for i := range s.ds.Species {
		sp := &s.ds.Species[i]
		if selected[sp.Family] {
			filtered = append(filtered, sp)
		}
	}
	secs := view.Sections(filtered, opt.Metric, opt.Step)
	refs := view.Scales(secs, opt.Scale)

	// The width reference must span the whole scale group too, or the crossbar
	// would be clipped for wide-winged birds in a section whose longest bird is
	// short -- a gull beside a rail.
	spanRefs := make([]float64, len(secs))
	var globalSpan float64
	for i, sec := range secs {
		for _, sp := range sec.Species {
			if w := sp.Wingspan.Value; w > spanRefs[i] {
				spanRefs[i] = w
			}
		}
		if spanRefs[i] > globalSpan {
			globalSpan = spanRefs[i]
		}
	}
	if opt.Scale == view.Global {
		for i := range spanRefs {
			spanRefs[i] = globalSpan
		}
	}

	page := pageData{Opt: opt, Total: len(s.ds.Species), Shown: len(filtered),
		Query: queryFor(r)}
	if sp := s.legendSpecies(); sp != nil {
		lt := labelledTile(sp, legendBox, opt.Metric3)
		page.Legend = &lt
	}

	// "All" toggles like an order does, over every family. Untoggling it clears
	// the selection outright rather than leaving 91 redundant parameters behind.
	page.AllSelected = len(s.fams) > 0
	for _, f := range s.fams {
		if !selected[f.Name] {
			page.AllSelected = false
			break
		}
	}
	every := make(map[string]bool, len(s.fams))
	for _, f := range s.fams {
		every[f.Name] = true
	}
	if page.AllSelected {
		page.AllURL = s.familyURL(r, nil)
	} else {
		page.AllURL = s.familyURL(r, every)
	}
	page.NothingSelected = len(selected) == 0
	byOrder := map[string]int{}
	for _, f := range s.fams {
		next := map[string]bool{}
		for k := range selected {
			next[k] = true
		}
		if next[f.Name] {
			delete(next, f.Name)
		} else {
			next[f.Name] = true
		}
		row := familyRow{familyOption: f, Selected: selected[f.Name], URL: s.familyURL(r, next)}
		gi, ok := byOrder[f.Order]
		if !ok {
			page.Groups = append(page.Groups, orderGroup{Order: f.Order, Common: f.OrderCommon})
			gi = len(page.Groups) - 1
			byOrder[f.Order] = gi
		}
		page.Groups[gi].Families = append(page.Groups[gi].Families, row)
		page.Groups[gi].Count += f.Count
	}
	// An order toggle selects its whole group, or clears it when the group is
	// already fully selected. Deriving "selected" from the families rather than
	// tracking it separately means the two can never disagree.
	for i := range page.Groups {
		g := &page.Groups[i]
		g.Selected = true
		for _, f := range g.Families {
			if !selected[f.Name] {
				g.Selected = false
				break
			}
		}
		next := map[string]bool{}
		for k := range selected {
			next[k] = true
		}
		for _, f := range g.Families {
			if g.Selected {
				delete(next, f.Name)
			} else {
				next[f.Name] = true
			}
		}
		g.URL = s.familyURL(r, next)
	}
	for i, sec := range secs {
		sv := sectionView{Index: i + 1, Count: len(sec.Species),
			Min: format(sec.Min, opt.Metric, opt.Metric3), Max: format(sec.Max, opt.Metric, opt.Metric3)}
		for _, sp := range sec.Species {
			g := view.Layout(sp, refs[i], spanRefs[i], tileBox)
			sv.Tiles = append(sv.Tiles, tile{
				S: sp, Geom: g, Box: tileBox,
				// Set here rather than recomputed in the template: the template
				// used to derive the centering from Box.H itself, which silently
				// disagreed with this value the moment Box stopped being the
				// drawing area (P-066's bottom strip).
				TopOffset:   (tileBox.H - g.TotalHeight()) / 2,
				WingModeled: sp.Wingspan.Prov.Kind == dataset.Modeled,
				Length:      format(sp.TotalLength.Value, view.ByLength, opt.Metric3),
				Wing:        format(sp.Wingspan.Value, view.ByWingspan, opt.Metric3),
				MassStr:     format(sp.Mass.Value, view.ByMass, opt.Metric3),
			})
		}
		page.Sections = append(page.Sections, sv)
	}
	s.render(w, "index.html", page)
}

type detailData struct {
	S       *dataset.Species
	Geom    view.Geometry
	Box     view.Box
	Metrics []metricRow
	Opt     options
	Query   func(...string) template.URL

	// INatURL is empty when no confident iNaturalist match exists, and the
	// template then renders nothing at all rather than a placeholder (P-064).
	INatURL string

	// WingModeled drives the crossbar split: the inner span is dashed only when
	// the wingspan is modeled (P-065).
	WingModeled bool

	// Labels annotate the detail diagram (P-066); empty on a tile.
	Labels        []diagramLabel
	TopOffset     float64
	ChordLabel    string
	WingspanLabel string
	WingspanEst   bool
	TotalLabel    string
	TotalEst      bool
	TotalLabelY   float64
}

// diagramLabel is one annotation: where view put it, what it says, and whether
// the value it reports is estimated.
type diagramLabel struct {
	view.Label
	Text string
	Est  bool
}

type metricRow struct {
	Label, Value, Kind, Detail string
}

func (s *server) detail(w http.ResponseWriter, r *http.Request) {
	sp, ok := s.byID[strings.TrimPrefix(r.URL.Path, "/species/")]
	if !ok {
		http.NotFound(w, r)
		return
	}
	opt := parseOptions(r)
	// Wider than the tile's 2.6:1 so the labels of P-066 have margin outside the
	// wingtips. The extra width costs nothing: the scale is already limited by
	// height for almost every bird, so the bird does not grow, the margin does.
	box := view.Box{W: tileBox.W * 2.6, H: tileBox.H * 1.9}
	// The detail view always draws the bird as large as the box allows, whatever
	// scale the index is using. Under a global scale a small bird is a few
	// pixels tall by design, and carrying that into the one view devoted to
	// that single bird would leave nothing to look at (P-034).

	rows := []metricRow{
		row("Total length", sp.TotalLength, view.ByLength, opt.Metric3),
		row("Wingspan", sp.Wingspan, view.ByWingspan, opt.Metric3),
		row("Wing chord", sp.WingChord, view.ByLength, opt.Metric3),
		row("Mass", sp.Mass, view.ByMass, opt.Metric3),
		row("Bill", sp.Segments.Bill, view.ByLength, opt.Metric3),
		row("Head", sp.Segments.Head, view.ByLength, opt.Metric3),
		row("Neck", sp.Segments.Neck, view.ByLength, opt.Metric3),
		row("Body", sp.Segments.Body, view.ByLength, opt.Metric3),
		row("Tail", sp.Segments.Tail, view.ByLength, opt.Metric3),
	}
	lt := labelledTile(sp, box, opt.Metric3)

	// Only a confident match gets a link; zero means none was found (P-064).
	var inatURL string
	if sp.INatTaxonID > 0 {
		inatURL = inat.TaxonURL(sp.INatTaxonID)
	}
	s.render(w, "detail.html", detailData{
		S: sp, Geom: lt.Geom, Box: lt.Box, Labels: lt.Labels, TopOffset: lt.TopOffset,
		ChordLabel: lt.ChordLabel, WingspanLabel: lt.WingspanLabel, WingspanEst: lt.WingspanEst,
		TotalLabel: lt.TotalLabel, TotalEst: lt.TotalEst, TotalLabelY: lt.TotalLabelY,
		Metrics: rows, Opt: opt, Query: queryFor(r), INatURL: inatURL,
		WingModeled: sp.Wingspan.Prov.Kind == dataset.Modeled,
	})
}

// comparePanel is one bird on the comparison page: the same annotated diagram
// the detail view draws, plus the two links that make the panel navigable.
//
// tile is embedded so the shared "diagram" template, which reads .Geom, .Box,
// .Labels and the rest off whatever it is handed, works here unchanged.
type comparePanel struct {
	tile
	// DetailURL is where this panel's numbers are explained. P-067 trades the
	// provenance table for this link, so it is not decoration -- it is the
	// whole of what makes omitting the table honest.
	DetailURL template.URL
	// RemoveURL is the same comparison without this bird.
	RemoveURL template.URL
}

type compareData struct {
	Panels []comparePanel
	Opt    options
	Query  func(...string) template.URL
}

// compare renders several birds to one shared scale (P-067).
func (s *server) compare(w http.ResponseWriter, r *http.Request) {
	opt := parseOptions(r)
	ids := r.URL.Query()["species"]

	// Named order is the reader's choice and is preserved; a repeat is dropped
	// rather than drawn twice, and an unknown id is an error rather than a
	// silently shorter page -- a shared link with a typo in it should say so.
	var chosen []*dataset.Species
	seen := map[string]bool{}
	for _, id := range ids {
		if id == "" || seen[id] {
			continue
		}
		sp, ok := s.byID[id]
		if !ok {
			http.Error(w, "unknown species: "+id, http.StatusNotFound)
			return
		}
		seen[id] = true
		chosen = append(chosen, sp)
	}

	// One scale for the whole set, taken from the largest bird in it, exactly
	// as a section does (P-042). Laying each panel out against its own length
	// would fill every panel equally and compare nothing.
	var refLength, refWingspan float64
	for _, sp := range chosen {
		refLength = math.Max(refLength, sp.TotalLength.Value)
		refWingspan = math.Max(refWingspan, sp.Wingspan.Value)
	}

	box := view.Box{W: tileBox.W * 2.6, H: tileBox.H * 1.9}
	panels := make([]comparePanel, 0, len(chosen))
	for _, sp := range chosen {
		panels = append(panels, comparePanel{
			tile:      labelledTileAt(sp, box, opt.Metric3, refLength, refWingspan),
			DetailURL: detailURL(r, sp.ID),
			RemoveURL: compareWithout(r, sp.ID),
		})
	}
	s.render(w, "compare.html", compareData{Panels: panels, Opt: opt, Query: queryFor(r)})
}

// detailURL points a panel at its own species page, carrying the display
// options but not the comparison's species list.
func detailURL(r *http.Request, id string) template.URL {
	q := r.URL.Query()
	q.Del("species")
	if len(q) == 0 {
		return template.URL("/species/" + id)
	}
	return template.URL("/species/" + id + "?" + q.Encode())
}

// compareWithout is the same comparison minus one bird. Every other parameter
// survives, including a repeat of some other species.
func compareWithout(r *http.Request, id string) template.URL {
	q := r.URL.Query()
	kept := q["species"][:0:0]
	for _, v := range q["species"] {
		if v != id {
			kept = append(kept, v)
		}
	}
	q["species"] = kept
	return template.URL("/compare?" + q.Encode())
}

// legendBox is the help diagram's drawing area: smaller than the detail view's
// but the same shape, since it renders through the same template.
var legendBox = view.Box{W: 900, H: 148 * 1.2}

// labelGutter is the width reserved at the right of an annotated diagram: the
// segment label column, plus room for the wingspan label that sits just beyond
// the wingtip.
//
// The bird is laid out in the remaining width, so a wide-winged bird is drawn
// slightly smaller rather than running under its own labels. Before this, a
// Black-footed Albatross's wingtip reached x=988 in a 988-wide box while the
// label column started at 838, so the crossbar crossed every label on that
// side.
// 320 rather than the 290 that merely fits: at 290 the widest-winged species
// clear the column by 2px, which is not clearance, it is luck.
const labelGutter = 320

// totalLabelStrip is the height reserved at the bottom of an annotated diagram
// for the total length label (P-066), the vertical counterpart of labelGutter.
//
// Without it the label is clipped rather than merely cramped: the detail view
// scales the bird to fill the box, so any bird whose height binds has its tail
// tip exactly on the bottom edge and nothing at all below it.
//
// totalLabelGap is where the baseline sits below the tail tip. The strip is the
// gap plus room for the 11px text's descender, so the glyphs stay inside the
// viewBox and not merely their baseline.
const (
	totalLabelGap   = 15
	totalLabelStrip = totalLabelGap + 9
)

// legendPreferred are the species the help diagram would like to use, best
// first. Each has a MODELED wingspan, so the dashed inner crossbar of P-065
// appears in the legend rather than only in the prose beside it, and none is a
// neck-folding family, so the neck reads as a plain stick.
var legendPreferred = []string{"cyanocitta_cristata", "corvus_brachyrhynchos", "cardinalis_cardinalis"}

// legendSpecies picks the bird the help draws. Preference first, then any
// species with a modeled wingspan, then nothing -- the help degrades to prose
// rather than breaking if the checklist ever drops all of them.
func (s *server) legendSpecies() *dataset.Species {
	for _, id := range legendPreferred {
		if sp := s.byID[id]; sp != nil {
			return sp
		}
	}
	for i := range s.ds.Species {
		if sp := &s.ds.Species[i]; sp.Wingspan.Prov.Kind == dataset.Modeled {
			return sp
		}
	}
	return nil
}

// labelledTile builds the annotated diagram for a bird drawn at its own scale:
// the detail view and the help legend, each of which shows one bird with
// nothing to measure it against (P-034).
func labelledTile(sp *dataset.Species, box view.Box, metric bool) tile {
	return labelledTileAt(sp, box, metric, sp.TotalLength.Value, sp.Wingspan.Value)
}

// labelledTileAt is the same diagram laid out against an EXTERNAL reference, so
// several birds can share one scale (P-067).
//
// That reference is the whole of the difference between the two views: the
// bird's own length makes a detail page fill its box, and the set maximum makes
// a comparison mean something. Getting the argument wrong yields a page that
// looks entirely correct and compares nothing, which is why AC-076 asserts
// drawn heights against the length ratio rather than asserting on the markup.
func labelledTileAt(sp *dataset.Species, box view.Box, metric bool, refLength, refWingspan float64) tile {
	// draw is the area the bird gets; the emitted box is that plus the strip the
	// total length label sits in, so the label costs the bird nothing. Going the
	// other way -- keeping the box and shrinking the bird into it -- would take
	// 8.5% of the detail view's height and contradict P-034, which requires the
	// bird drawn as large as the view allows. Width has no such freedom, which
	// is why labelGutter does shrink the bird and this does not.
	draw := box
	box.H += totalLabelStrip

	// The bird occupies everything left of the gutter; draw stays the full
	// height, because the label column is positioned against the emitted box.
	inner := view.Box{W: draw.W - labelGutter, H: draw.H}
	g := view.Layout(sp, refLength, refWingspan, inner)
	top := (draw.H - g.TotalHeight()) / 2
	segOf := map[string]dataset.Metric{
		"bill": sp.Segments.Bill, "head": sp.Segments.Head, "neck": sp.Segments.Neck,
		"body": sp.Segments.Body, "tail": sp.Segments.Tail,
	}
	t := tile{S: sp, Geom: g, Box: box, TopOffset: top,
		WingModeled: sp.Wingspan.Prov.Kind == dataset.Modeled}
	// draw, not box: the segment label column belongs in the drawing area, and
	// must not spill into the strip the total length label owns.
	for _, l := range g.DetailLabels(draw, top) {
		m := segOf[l.Key]
		t.Labels = append(t.Labels, diagramLabel{
			Label: l,
			Text:  strings.ToUpper(l.Key[:1]) + l.Key[1:] + "  " + format(m.Value, view.ByLength, metric),
			Est:   m.Prov.Kind == dataset.Modeled,
		})
	}
	if sp.WingChord.Value > 0 {
		t.ChordLabel = "Wing chord  " + format(sp.WingChord.Value, view.ByLength, metric)
	}
	t.WingspanLabel = "Wingspan  " + format(sp.Wingspan.Value, view.ByWingspan, metric)
	t.WingspanEst = sp.Wingspan.Prov.Kind == dataset.Modeled
	// The reported total length, not g.TotalHeight(): for a folded neck the
	// drawing is shorter than the bird (CR-011), and the label names the bird.
	t.TotalLabel = "Total length  " + format(sp.TotalLength.Value, view.ByLength, metric)
	t.TotalEst = sp.TotalLength.Prov.Kind == dataset.Modeled
	t.TotalLabelY = top + g.YTailTip + totalLabelGap
	return t
}

// row turns a metric into a display line, stating its provenance in plain
// language (P-063).
func row(label string, m dataset.Metric, as view.Metric, metric bool) metricRow {
	r := metricRow{Label: label, Value: format(m.Value, as, metric), Kind: string(m.Prov.Kind)}
	switch m.Prov.Kind {
	case dataset.Measured:
		r.Detail = "measured, from " + m.Prov.Source
		if m.Prov.SourceTaxon != "" {
			r.Detail += " (" + m.Prov.SourceTaxon + ")"
		}
	case dataset.NearMatch:
		r.Detail = "measured on " + m.Prov.SourceTaxon +
			", a different species this one was split from \u2014 no measurement exists for it. Source: " +
			m.Prov.Source
	default:
		r.Detail = "estimated: " + m.Prov.Model
	}
	if m.Min > 0 && m.Max > 0 {
		r.Detail += fmt.Sprintf(" Reported range %s\u2013%s.",
			format(m.Min, as, metric), format(m.Max, as, metric))
	}
	return r
}

func (s *server) render(w http.ResponseWriter, name string, data any) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	if err := s.tmpl.ExecuteTemplate(w, name, data); err != nil {
		log.Printf("birdview: rendering %s: %v", name, err)
	}
}

// queryFor returns a helper that rewrites the current query string, so a
// control can toggle one setting without discarding the others.
func queryFor(r *http.Request) func(...string) template.URL {
	return func(kv ...string) template.URL {
		q := r.URL.Query()
		for i := 0; i+1 < len(kv); i += 2 {
			if kv[i+1] == "" {
				q.Del(kv[i])
			} else {
				q.Set(kv[i], kv[i+1])
			}
		}
		return template.URL("/?" + q.Encode())
	}
}

// format renders a value in the selected unit system.
//
// Mass in imperial means ounces and pounds. An earlier version returned
// kilograms above a kilogram regardless of the toggle, so "imperial" showed a
// metric unit for exactly the birds where it mattered most.
func format(v float64, as view.Metric, metric bool) string {
	if as == view.ByMass {
		if metric {
			if v >= 1000 {
				return fmt.Sprintf("%.2f kg", v/1000)
			}
			return fmt.Sprintf("%.0f g", v)
		}
		oz := v / 28.349523125
		switch {
		case oz >= 16:
			return fmt.Sprintf("%.1f lb", oz/16)
		case oz >= 1:
			return fmt.Sprintf("%.1f oz", oz)
		default:
			// A 3 g hummingbird is 0.11 oz; one decimal would erase the
			// difference between the smallest birds entirely.
			return fmt.Sprintf("%.2f oz", oz)
		}
	}
	if metric {
		return fmt.Sprintf("%.1f cm", v/10)
	}
	return fmt.Sprintf("%.1f in", v/25.4)
}

// familyURL rebuilds the current URL with exactly the given families selected.
//
// Two states get a short encoding. Everything selected is the default, so it
// needs no family parameter at all; nothing selected is a single empty one.
// Without that distinction the two would be indistinguishable in the query
// string while meaning opposite things.
func (s *server) familyURL(r *http.Request, fams map[string]bool) template.URL {
	q := r.URL.Query()
	q.Del("family")
	switch {
	case len(fams) == len(s.fams):
		// default: no parameter
	case len(fams) == 0:
		q.Set("family", "")
	default:
		names := make([]string, 0, len(fams))
		for f := range fams {
			names = append(names, f)
		}
		sort.Strings(names)
		for _, f := range names {
			q.Add("family", f)
		}
	}
	return template.URL("/?" + q.Encode())
}

var funcs = template.FuncMap{
	"f1":   func(v float64) string { return strconv.FormatFloat(v, 'f', 1, 64) },
	"add":  func(a, b float64) float64 { return a + b },
	"sub":  func(a, b float64) float64 { return a - b },
	"div":  func(a, b float64) float64 { return a / b },
	"list": func(v ...string) []string { return v },
}
