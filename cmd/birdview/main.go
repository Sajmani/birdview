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
	"net/http"
	"os"
	"sort"
	"strconv"
	"strings"

	"birdview/internal/dataset"
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
	Name, Common, Color string
	Count               int
}

func (s *server) ServeHTTP(w http.ResponseWriter, r *http.Request) { s.mux.ServeHTTP(w, r) }

func newServer(ds *dataset.Dataset) (*server, error) {
	tmpl, err := template.New("").Funcs(funcs).ParseFS(web.Static, "static/*.html")
	if err != nil {
		return nil, fmt.Errorf("parsing templates: %w", err)
	}
	s := &server{mux: http.NewServeMux(), ds: ds, tmpl: tmpl, byID: map[string]*dataset.Species{}}

	counts := map[string]*familyOption{}
	for i := range ds.Species {
		sp := &ds.Species[i]
		s.byID[sp.ID] = sp
		f, ok := counts[sp.Family]
		if !ok {
			f = &familyOption{Name: sp.Family, Common: sp.FamilyCommon, Color: sp.FamilyColor}
			counts[sp.Family] = f
		}
		f.Count++
	}
	for _, f := range counts {
		s.fams = append(s.fams, *f)
	}
	sort.Slice(s.fams, func(i, j int) bool { return s.fams[i].Name < s.fams[j].Name })

	assets, err := fs.Sub(web.Static, "static")
	if err != nil {
		return nil, err
	}
	s.mux.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.FS(assets))))
	s.mux.HandleFunc("/species/", s.detail)
	s.mux.HandleFunc("/", s.index)
	return s, nil
}

// options are the user's current selections, all carried in the query string so
// that any view is a shareable URL.
type options struct {
	Metric   view.Metric
	Step     float64
	Scale    view.ScaleMode
	Families map[string]bool
	Metric3  bool // imperial when false
}

func parseOptions(r *http.Request) options {
	q := r.URL.Query()
	o := options{Metric: view.ByLength, Step: view.DefaultStepPercent, Scale: view.PerSection,
		Families: map[string]bool{}}
	switch view.Metric(q.Get("by")) {
	case view.ByWingspan:
		o.Metric = view.ByWingspan
	case view.ByMass:
		o.Metric = view.ByMass
	}
	if view.ScaleMode(q.Get("scale")) == view.Global {
		o.Scale = view.Global
	}
	if v, err := strconv.ParseFloat(q.Get("step"), 64); err == nil && v >= 2 && v <= 100 {
		o.Step = v
	}
	o.Metric3 = q.Get("units") == "metric"
	for _, f := range q["family"] {
		if f != "" {
			o.Families[f] = true
		}
	}
	return o
}

// tile is one rendered species.
type tile struct {
	S    *dataset.Species
	Geom view.Geometry
	Box  view.Box
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
	Families []familyRow
	Opt      options
	Total    int
	Shown    int
	Query    func(...string) template.URL
	ClearURL template.URL
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

	var filtered []*dataset.Species
	for i := range s.ds.Species {
		sp := &s.ds.Species[i]
		if len(opt.Families) == 0 || opt.Families[sp.Family] {
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
			if sp.Wingspan.Value > spanRefs[i] {
				spanRefs[i] = sp.Wingspan.Value
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
		Query: queryFor(r), ClearURL: familyURL(r, nil)}
	for _, f := range s.fams {
		next := map[string]bool{}
		for k := range opt.Families {
			next[k] = true
		}
		if next[f.Name] {
			delete(next, f.Name)
		} else {
			next[f.Name] = true
		}
		page.Families = append(page.Families, familyRow{
			familyOption: f, Selected: opt.Families[f.Name], URL: familyURL(r, next),
		})
	}
	for i, sec := range secs {
		sv := sectionView{Index: i + 1, Count: len(sec.Species),
			Min: format(sec.Min, opt.Metric, opt.Metric3), Max: format(sec.Max, opt.Metric, opt.Metric3)}
		for _, sp := range sec.Species {
			sv.Tiles = append(sv.Tiles, tile{S: sp, Geom: view.Layout(sp, refs[i], spanRefs[i], tileBox), Box: tileBox})
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
	box := view.Box{W: tileBox.W * 1.9, H: tileBox.H * 1.9}

	rows := []metricRow{
		row("Total length", sp.TotalLength, view.ByLength, opt.Metric3),
		row("Wingspan", sp.Wingspan, view.ByWingspan, opt.Metric3),
		row("Mass", sp.Mass, view.ByMass, opt.Metric3),
		row("Bill", sp.Segments.Bill, view.ByLength, opt.Metric3),
		row("Head", sp.Segments.Head, view.ByLength, opt.Metric3),
		row("Neck", sp.Segments.Neck, view.ByLength, opt.Metric3),
		row("Body", sp.Segments.Body, view.ByLength, opt.Metric3),
		row("Tail", sp.Segments.Tail, view.ByLength, opt.Metric3),
	}
	s.render(w, "detail.html", detailData{
		S: sp, Geom: view.Layout(sp, sp.TotalLength.Value, sp.Wingspan.Value, box), Box: box,
		Metrics: rows, Opt: opt, Query: queryFor(r),
	})
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

func format(v float64, as view.Metric, metric bool) string {
	if as == view.ByMass {
		if metric || v < 1000 {
			return fmt.Sprintf("%.0f g", v)
		}
		return fmt.Sprintf("%.2f kg", v/1000)
	}
	if metric {
		return fmt.Sprintf("%.1f cm", v/10)
	}
	return fmt.Sprintf("%.1f in", v/25.4)
}

// familyURL rebuilds the current URL with exactly the given families selected.
func familyURL(r *http.Request, fams map[string]bool) template.URL {
	q := r.URL.Query()
	q.Del("family")
	names := make([]string, 0, len(fams))
	for f := range fams {
		names = append(names, f)
	}
	sort.Strings(names)
	for _, f := range names {
		q.Add("family", f)
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
