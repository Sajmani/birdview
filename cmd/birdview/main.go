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
	"compress/gzip"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"flag"
	"fmt"
	"html/template"
	"io"
	"io/fs"
	"log"
	"math"
	"net/http"
	"net/url"
	"os"
	"regexp"
	"slices"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"

	"bytes"

	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/extension"
	goldmarkhtml "github.com/yuin/goldmark/renderer/html"

	"embed"

	goldmarkast "github.com/yuin/goldmark/ast"
	goldmarkparser "github.com/yuin/goldmark/parser"
	goldmarktext "github.com/yuin/goldmark/text"
	goldmarkutil "github.com/yuin/goldmark/util"

	"birdview/docs"
	"birdview/internal/dataset"
	"birdview/internal/inat"
	"birdview/internal/month"
	"birdview/internal/region"
	"birdview/internal/status"
	"birdview/internal/taxonomy"
	"birdview/internal/trait"
	"birdview/internal/view"
	"birdview/river"
	"birdview/web"
)

func main() {
	addr := flag.String("addr", "localhost:8080", "listen address")
	data := flag.String("data", "web/data/species.json", "compiled dataset (a build artifact)")
	dev := flag.String("dev", "", "serve templates and static files from this directory instead of the embed, so edits show on reload")
	// Where readers are asked to send feedback. A URL, not an address: a
	// mailto on a public page is a mailbox harvested within the week, and
	// the owner's personal address is not the project's contact point
	// (T-025).
	//
	// Empty by default, and an empty value publishes NOTHING rather than a
	// dead link -- a site can go up before its tracker exists, but it must
	// not invite a reader to a page that does not.
	feedback := flag.String("feedback", "", "issue tracker URL for reader feedback; omitted from the page when unset")
	flag.Parse()

	if *dev != "" {
		if err := devAssets(*dev); err != nil {
			log.Fatalf("birdview: %v", err)
		}
		log.Printf("birdview: DEV MODE, assets from %s — the embed is not being used", *dev)
	}

	ds, err := load(*data)
	if err != nil {
		log.Fatalf("birdview: %v\n\nRun birdview-ingest first; the dataset is a build artifact "+
			"and is not committed (T-005).", err)
	}
	srv, err := newServer(ds)
	if err == nil {
		srv.feedback = *feedback
	}
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
	mux *http.ServeMux
	// ds is what the index draws: every species except the exotics
	// (P-080). all is the whole dataset, for the detail pages and the lists
	// that must name every bird.
	ds   *dataset.Dataset
	all  *dataset.Dataset
	tmpl *template.Template
	byID map[string]*dataset.Species
	// idOf maps an AviList scientific name to the served id, so a page can
	// name another species without searching the slice (P-077).
	idOf map[string]string
	// sameAs maps a scientific name to the other drawn species whose traits
	// come from the same AVONET row, sorted (P-079).
	sameAs map[string][]string
	// undrawnByID holds the species the derivation classified and rejected.
	// Consulted only after byID misses, so a URL that stopped working can say
	// why instead of 404ing into silence (P-075).
	undrawnByID map[string]*dataset.Undrawn
	fams        []familyOption
	// devDir is empty for a real binary. When set, render re-parses the
	// templates on every request so an edit shows on reload.
	devDir string
	// feedback is the issue tracker readers are pointed at, or empty, in
	// which case no page says anything about feedback at all (T-025).
	feedback string
	// build fingerprints everything a response is rendered from: the
	// dataset, the templates and the static files. It is the first half of
	// every ETag (T-029).
	build string
}

type familyOption struct {
	Name, Common     string
	Color, TextColor string
	Order            string
	OrderCommon      string
}

// orderGroup is one group of families inside a block of the selector --
// historically an order, now either an order or a birder group (P-014).
// Grouping by order in the checklist's taxonomic sequence matches how every
// field guide is arranged, and it is what lets family colors repeat across
// orders (P-010, P-005).
type orderGroup struct {
	Order    string       // the order's name
	Common   string       // the order's English name; empty for a birder group
	Selected bool         // every family in the group is filtered in
	URL      template.URL // toggles the whole group
	// OnlyURL selects this group's families and nothing else.
	OnlyURL  template.URL
	Families []familyRow
}

// familyGroupingBlock is one way of carving the families up, rendered as its
// own labelled block (P-014). The same shape as the region selector's
// groupingBlock, and it reuses that CSS.
type familyGroupingBlock struct {
	Name   string
	Note   string
	Groups []orderGroup
}

// groupKey names one group a family belongs to, and the group's subtitle.
type groupKey struct{ Name, Common string }

// groupFamilies gathers family rows into one block's groups.
//
// `keyOf` names **every** group a family belongs to: a birder group may hold
// a family that another birder group also holds, since an owl is both a bird
// of prey and a bird of the night (P-014, CR-026). `order` fixes the sequence
// groups are offered in; passing nil keeps the order families first appear
// in, which for the taxonomic block is checklist sequence.
//
// A family the table does not know -- one AviList added that the birder table
// has not been told about -- yields no groups and is gathered under the empty
// name rather than dropped, so it stays reachable and AC-089 reports it.
func (s *server) groupFamilies(r *http.Request, selected map[string]bool, rows []familyRow,
	order []string, keyOf func(familyOption) []groupKey) []orderGroup {

	var out []orderGroup
	at := map[string]int{}
	ensure := func(name, common string) int {
		if i, ok := at[name]; ok {
			return i
		}
		out = append(out, orderGroup{Order: name, Common: common})
		at[name] = len(out) - 1
		return len(out) - 1
	}
	// Seed in the requested order, so an empty group still holds its place
	// and the sequence does not depend on which families happen to exist.
	for _, name := range order {
		ensure(name, "")
	}
	for _, row := range rows {
		keys := keyOf(row.familyOption)
		if len(keys) == 0 {
			keys = []groupKey{{}} // unknown family: visible under the empty name
		}
		for _, k := range keys {
			i := ensure(k.Name, k.Common)
			out[i].Families = append(out[i].Families, row)
		}
	}
	// Drop groups no family landed in, which only happens for a seeded name
	// the dataset has no families for.
	kept := out[:0]
	for _, g := range out {
		if len(g.Families) > 0 {
			kept = append(kept, g)
		}
	}
	out = kept

	// A group toggle selects its whole group, or clears it when the group is
	// already fully selected. Deriving "selected" from the families rather
	// than tracking it separately means the two can never disagree.
	for i := range out {
		g := &out[i]
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

		// Only: exactly this group's families. Built from the group alone,
		// not from the current selection, so it lands on the same page
		// whatever was on before.
		only := make(map[string]bool, len(g.Families))
		for _, f := range g.Families {
			only[f.Name] = true
		}
		g.OnlyURL = s.familyURL(r, only)
	}
	return out
}

func (s *server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	compressed(s.revalidated(s.mux)).ServeHTTP(w, r)
}

// revalidated gives every response a validator and tells caches to check it
// before reusing it (T-029).
//
// BirdView sent no Cache-Control, no ETag and no Last-Modified at all, which
// does not mean "do not cache" -- it means the browser may apply its own
// heuristic and has nothing to revalidate against when it does. The symptom
// is a reader reloading a page after a rebuild and being served the old one,
// with no way to tell and no way to force the issue. It cost a review round
// on a page that was correct on the server.
//
// `no-cache` is the counter-intuitive half: it means *store this, but ask
// before reusing it*, which is exactly right here. The page is cheap to
// validate and must never be stale, since the whole point of a figure like
// P-074's is that it is current.
//
// The ETag is computed from the build and the URL rather than by hashing the
// body, because hashing the body means buffering it, and the index is 3.3 MB
// before compression. That is sound only because a page is a pure function
// of (dataset, templates, static files, request URI) -- there is no clock in
// the render path -- and AC-110 asserts that rather than trusting it.
func (s *server) revalidated(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// -dev re-reads templates per request, so no fingerprint taken at
		// startup stays true. Sending no validator is correct there: the
		// whole point of dev mode is that an edit shows on reload.
		if s.build == "" {
			w.Header().Set("Cache-Control", "no-store")
			h.ServeHTTP(w, r)
			return
		}
		sum := sha256.Sum256([]byte(s.build + "\x00" + r.Method + "\x00" + r.URL.RequestURI()))
		etag := `"` + hex.EncodeToString(sum[:16]) + `"`
		w.Header().Set("ETag", etag)
		w.Header().Set("Cache-Control", "no-cache")
		// Compared against the bare tag: the client echoes whatever we sent,
		// and the gzip layer may have suffixed it to distinguish the encoded
		// representation from the identity one.
		if m := r.Header.Get("If-None-Match"); m != "" {
			for _, c := range strings.Split(m, ",") {
				c = strings.TrimSpace(c)
				c = strings.TrimPrefix(c, "W/")
				if c == etag || c == etag[:len(etag)-1]+`-gz"` {
					w.WriteHeader(http.StatusNotModified)
					return
				}
			}
		}
		h.ServeHTTP(w, r)
	})
}

// renamed says how two binomials for one taxon differ: a short noun for the
// row's own line, and a sentence naming both halves next to each other.
//
// Two shapes in AviListRenames and they are not the same event. Five are
// genus moves, where the epithet is untouched and the eye catches the
// difference at once. Two are the same genus with a corrected epithet, and
// there the two names can differ by a single letter -- *Pica nuttalli* and
// *Pica nuttallii* -- so a sentence that names both without saying what
// changed reads as though it has printed the same name twice. It was
// reported as a bug in the page, which is what it looked like.
//
// Both halves are therefore put side by side in one clause, because the
// only way to show a one-letter difference in running text is to make the
// reader's eye compare two adjacent words.
//
// It says eBird and not "the occurrence data", which would be wrong in both
// directions. iNaturalist follows AviList for all five woodpeckers and the
// grosbeak -- 53,705 records under *Leuconotopicus villosus* -- and follows
// neither spelling for *Pica nuttallii*, where both databases hold nothing.
// eBird is the one thing true of every row: it holds zero under the AviList
// name in all seven, which is why the join needs the old one at all.
func renamed(old, current string) (what, how string) {
	og, oe, _ := strings.Cut(old, " ")
	ng, ne, _ := strings.Cut(current, " ")
	switch {
	case og != ng && oe != ne:
		return "name", "AviList places this species in the genus " + ng +
			" and spells the epithet " + ne + "; eBird files it as " + old + "."
	case og != ng:
		return "genus", "AviList places this species in the genus " + ng +
			" and eBird files it under " + og + "."
	case oe != ne:
		return "spelling", "AviList spells the epithet " + ne +
			" and eBird spells it " + oe + "; the genus is the same in both."
	}
	return "name", "AviList and eBird differ over this name."
}

// fingerprint hashes everything a rendered page depends on besides the URL:
// the dataset, and every template and static file the binary serves.
//
// All of them together rather than one per asset, so a CSS edit invalidates
// the HTML that loads it. Over-invalidating by a little is the safe
// direction -- the cost is one refetch after a deploy, and the alternative
// is a page styled by a stylesheet it was not written for, which is the
// failure that is hard to recognize as a caching problem.
func fingerprint(ds *dataset.Dataset) (string, error) {
	h := sha256.New()
	paths, err := fs.Glob(assetFS, "static/*")
	if err != nil {
		return "", err
	}
	sort.Strings(paths)
	for _, p := range paths {
		b, err := fs.ReadFile(assetFS, p)
		if err != nil {
			return "", err
		}
		fmt.Fprintf(h, "%s\x00%d\x00", p, len(b))
		h.Write(b)
	}
	// Marshal rather than a file hash: the server is handed a parsed
	// dataset and never sees the bytes it came from. Go sorts map keys when
	// marshalling, so this is stable across runs.
	b, err := json.Marshal(ds)
	if err != nil {
		return "", err
	}
	h.Write(b)
	return hex.EncodeToString(h.Sum(nil)[:16]), nil
}

// gzipPool reuses compressors. A gzip.Writer carries a ~64 KiB window and a
// hash table, so allocating one per request is the kind of garbage a server
// should not make.
var gzipPool = sync.Pool{New: func() any { return gzip.NewWriter(io.Discard) }}

// gzipResponse compresses what a handler writes.
//
// The headers are decided in WriteHeader, not at the first Write, and that
// ordering is the whole of the type. Setting Content-Encoding from Write is
// too late for any handler that calls WriteHeader itself -- the header map is
// already committed -- so the body goes out compressed and unlabelled, and a
// browser renders binary. Both of BirdView's static files and every page of
// P-075's 404 did exactly that in the first version of this.
type gzipResponse struct {
	http.ResponseWriter
	gz *gzip.Writer
	// on is false for a status that may not carry a body, where emitting a
	// gzip stream would be a protocol error rather than a saving.
	on      bool
	started bool
}

func (g *gzipResponse) start(code int) {
	if g.started {
		return
	}
	g.started = true
	// 204 and 304 carry no body. Compressing nothing would produce the
	// twenty-byte header of an empty gzip stream where the protocol says
	// there must be no bytes at all.
	if code == http.StatusNoContent || code == http.StatusNotModified {
		g.ResponseWriter.WriteHeader(code)
		return
	}
	g.on = true
	// The gzipped body is a different representation of the resource and
	// may not carry the identity representation's validator: a cache that
	// mixed them would eventually hand gzip bytes to a client that asked
	// for none. Suffixed rather than dropped, so revalidation still works
	// for the clients that do send Accept-Encoding (T-029).
	if et := g.Header().Get("ETag"); strings.HasSuffix(et, `"`) {
		g.Header().Set("ETag", et[:len(et)-1]+`-gz"`)
	}
	// Content-Length describes the uncompressed body and is now a lie.
	// Deleted rather than corrected: the compressed length is not known
	// until the stream closes, and Go will chunk happily without it.
	g.Header().Del("Content-Length")
	g.Header().Set("Content-Encoding", "gzip")
	g.ResponseWriter.WriteHeader(code)
}

func (g *gzipResponse) WriteHeader(code int) { g.start(code) }

func (g *gzipResponse) Write(b []byte) (int, error) {
	g.start(http.StatusOK)
	if !g.on {
		return g.ResponseWriter.Write(b)
	}
	return g.gz.Write(b)
}

func (g *gzipResponse) Flush() {
	if g.on {
		g.gz.Flush()
	}
	if f, ok := g.ResponseWriter.(http.Flusher); ok {
		f.Flush()
	}
}

// compressed gzips a response when the client asks for it (T-026).
//
// Worth doing here and not left to a CDN, because the pages are enormous and
// the saving is on the one axis that costs money to serve. The index inlines
// an SVG diagram per species -- 3.3 MB of markup for 1,083 birds -- and
// compresses to 270 KB, a 92% saving. Against Cloud Run's 1 GiB of free
// monthly egress that is the difference between about 320 index loads and
// about 3,980.
//
// Vary is set on every response, not only compressed ones: a cache that
// stored the uncompressed reply without it would go on serving that reply to
// clients that asked for gzip, and worse, the reverse.
func compressed(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Add("Vary", "Accept-Encoding")
		if !strings.Contains(r.Header.Get("Accept-Encoding"), "gzip") {
			h.ServeHTTP(w, r)
			return
		}
		gz := gzipPool.Get().(*gzip.Writer)
		defer gzipPool.Put(gz)
		gz.Reset(w)
		g := &gzipResponse{ResponseWriter: w, gz: gz}
		h.ServeHTTP(g, r)
		if g.on {
			// Close only a stream that was opened. Closing an unused writer
			// emits a gzip header for a body nobody wrote.
			gz.Close()
		}
	})
}

// assetFS is where templates and static files are read from.
//
// Embedded for a real binary, and overridable from disk for development. That
// exists because the embed has an edge nobody notices until it bites: a
// running server holds the templates and CSS as they were when it STARTED, so
// editing one and reloading shows the old page with no sign anything is stale.
// It cost three wrong observations in one afternoon -- a recalibrated dataset,
// a recoloured favicon and a footer credit, each reported from a page that no
// longer matched the repository.
var assetFS fs.FS = web.Static

// devAssets points the server at the working tree, so an edit shows up on
// reload. Never used by a deployed binary: the flag defaults off and the
// embed is what ships.
func devAssets(dir string) error {
	sub := os.DirFS(dir)
	if _, err := fs.Stat(sub, "static/index.html"); err != nil {
		return fmt.Errorf("-dev %s: no static/index.html below it: %w", dir, err)
	}
	assetFS = sub
	devDir = dir
	return nil
}

// devDir is non-empty in dev mode, and is what makes render re-parse.
var devDir string

func newServer(ds *dataset.Dataset) (*server, error) {
	tmpl, err := template.New("").Funcs(funcs).ParseFS(assetFS, "static/*.html")
	if err != nil {
		return nil, fmt.Errorf("parsing templates: %w", err)
	}
	s := &server{mux: http.NewServeMux(), ds: ds, tmpl: tmpl, devDir: devDir,
		byID: map[string]*dataset.Species{}, undrawnByID: map[string]*dataset.Undrawn{},
		idOf: map[string]string{}}

	// Once, at startup: every response's ETag is built from this (T-029).
	// In dev mode the templates are re-read per request, so a fingerprint
	// taken now would go stale on the first edit -- which is the precise
	// bug -dev exists to avoid. Left empty there, and revalidated skips.
	if devDir == "" {
		s.build, err = fingerprint(ds)
		if err != nil {
			return nil, fmt.Errorf("fingerprinting the build: %w", err)
		}
	}

	counts := map[string]int{}
	for i := range ds.Species {
		sp := &ds.Species[i]
		s.byID[sp.ID] = sp
		s.idOf[sp.ScientificName] = sp.ID
		counts[sp.Family]++
	}
	s.sameAs = sharedRows(ds)
	// The index view: the same dataset without the exotics, which keep
	// their pages through byID (P-080). A shallow copy, so the totals and
	// the undrawn records are shared rather than duplicated.
	s.all = ds
	shown := *ds
	shown.Species = nil
	for i := range ds.Species {
		if ds.Species[i].Exotic == "" {
			shown.Species = append(shown.Species, ds.Species[i])
		}
	}
	s.ds = &shown
	for i := range ds.Undrawn {
		u := &ds.Undrawn[i]
		// The ingest refuses to emit a collision, so one here means the
		// dataset on disk was not built by it. Dropping the undrawn record is
		// the safe direction: a drawn bird keeps its page.
		if _, drawn := s.byID[u.ID]; drawn {
			return nil, fmt.Errorf("id %q is both drawn and not drawn; "+
				"one URL cannot mean two birds (P-075)", u.ID)
		}
		s.undrawnByID[u.ID] = u
	}
	// ds.Families is already in AviList's taxonomic sequence; preserve it
	// rather than alphabetizing, which is the order every field guide uses.
	onIndex := map[string]bool{}
	for i := range s.ds.Species {
		onIndex[s.ds.Species[i].Family] = true
	}
	for _, f := range ds.Families {
		// A family every one of whose birds is off the index is not
		// offered: a control that selects nothing is noise (P-080).
		if !onIndex[f.Name] {
			continue
		}
		s.fams = append(s.fams, familyOption{
			Name: f.Name, Common: f.Common, Color: f.Color, TextColor: f.TextColor,
			Order: f.Order, OrderCommon: f.OrderCommon,
		})
	}

	assets, err := fs.Sub(assetFS, "static")
	if err != nil {
		return nil, err
	}
	s.mux.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.FS(assets))))
	s.mux.HandleFunc("/added", s.added)
	s.mux.HandleFunc("/excluded", s.excluded)
	for _, set := range []string{"/docs", "/spec"} {
		// Both the exact path and the subtree, so the bare name serves the
		// index instead of a 301 to it.
		s.mux.HandleFunc(set, s.document)
		s.mux.HandleFunc(set+"/", s.document)
	}
	s.mux.HandleFunc("/compare", s.compare)
	s.mux.HandleFunc("/species/", s.detail)
	s.mux.HandleFunc("/", s.index)
	return s, nil
}

// options are the user's current selections, all carried in the query string so
// that any view is a shareable URL.
type options struct {
	Metric view.Metric
	// Step is P-023's section-step percentage, used by the three size
	// metrics. Frequency ignores it and bands by Decades instead.
	Step float64
	// Decades is how many orders of magnitude a frequency section spans
	// (P-076): 1 for "1 in 100 to 1 in 999", 2 for a band four times wider.
	// A separate field rather than an overloaded Step, because one float
	// meaning "15 percent" in three views and "one decade" in a fourth is
	// the kind of thing that reads fine and divides wrong.
	Decades int
	Scale   view.ScaleMode
	// Families holds the explicitly named families. HasSelection distinguishes
	// a fresh page, where nothing is named and everything is selected, from an
	// explicitly emptied selection, where nothing is named and nothing is
	// selected. The two look identical in the query string otherwise, and they
	// mean opposite things.
	Families     map[string]bool
	HasSelection bool
	// Regions and HasRegions draw the same distinction for P-011's filter:
	// absent means everywhere, explicitly empty means nowhere.
	Regions    map[string]bool
	HasRegions bool
	// Months and HasMonths draw the same distinction over the twelve
	// (P-011, CR-036). Indexed 0 to 11, January first.
	Months    map[int]bool
	HasMonths bool
	// Habitats and Niches are P-015's categorical filters, and draw the same
	// absent-versus-empty distinction. A species matches when ANY of its
	// assignments is selected: a bird of two habitats is in both, and
	// requiring all of them would mean selecting two habitats returned only
	// birds that use both, which is not what picking two places asks for.
	Habitats    map[string]bool
	HasHabitats bool
	Niches      map[string]bool
	HasNiches   bool
	Metric3     bool // imperial when false

	// Locale selects which English name is served (P-049). North American is
	// the default: this is an ABA-area dataset and Dovekie is what its reader
	// calls that bird.
	Locale dataset.Locale

	// Picking is pick mode (P-068): the `species` parameter is present, even
	// if it is empty. Picks are the ids named in it, deduplicated, in the
	// order given -- that order is the reader's and no page reorders it.
	Picking bool
	Picks   []string
}

func parseOptions(r *http.Request) options {
	q := r.URL.Query()
	// Per-section is the default (P-042, CR-017). Global makes the cross-family
	// argument better and is one click away, but at global scale a median bird
	// is a couple of pixels tall, and none of the diagram's vocabulary --
	// dashed for modeled, the chord ticks, the labels -- is legible there. The
	// default that teaches the reader to read a bird wins over the one that
	// delivers the shock before they can.
	o := options{Metric: view.ByLength, Step: view.DefaultStepPercent, Decades: 1,
		Scale:  view.PerSection,
		Locale: dataset.LocaleNA, Families: map[string]bool{}, Regions: map[string]bool{},
		Months:   map[int]bool{},
		Habitats: map[string]bool{}, Niches: map[string]bool{}}
	switch view.Metric(q.Get("by")) {
	case view.ByWingspan:
		o.Metric = view.ByWingspan
	case view.ByMass:
		o.Metric = view.ByMass
	case view.ByFrequency:
		o.Metric = view.ByFrequency
	}
	// Frequency defaults to global (CR-033). Per-section scaling rests on
	// P-022's guarantee that a section is at most N% wide, and a frequency
	// tier has no such bound: it would draw the one 140 mm bird in the last
	// tier at the height of a 1,201 mm swan six tiers up.
	if o.Metric == view.ByFrequency {
		o.Scale = view.Global
	}
	// BOTH values recognised, not just the non-default. The old comment here
	// warned that reading only the non-default silently disables the toggle
	// when the default moves -- which is what happened when CR-017 moved it.
	// With a default that now varies by metric, reading only Global would
	// leave "Per section" dead under Frequency, which is the same bug
	// arriving by a different road.
	switch view.ScaleMode(q.Get("scale")) {
	case view.Global:
		o.Scale = view.Global
	case view.PerSection:
		o.Scale = view.PerSection
	}
	// One `step` parameter, read differently by the two sectioning rules
	// (CR-032). Under Frequency the control offers band widths in orders of
	// magnitude, so the value is 10 or 100 rather than a percentage -- and
	// 10 is a legal percentage too, which is why the metric decides first
	// instead of the range of the number.
	if o.Metric == view.ByFrequency {
		if v := q.Get("step"); v == "100" {
			o.Decades = 2
		}
	} else if v, err := strconv.ParseFloat(q.Get("step"), 64); err == nil && v >= 2 && v <= 100 {
		o.Step = v
	}
	o.Metric3 = q.Get("units") == "metric"
	// Recognise the non-default, as with scale above: writing this the other
	// way round disables the toggle the moment the default moves.
	if dataset.Locale(q.Get("names")) == dataset.LocaleWorld {
		o.Locale = dataset.LocaleWorld
	}
	// Pick mode is the presence of the parameter, not a value in it: a bare
	// `species=` holds the mode open with nothing picked (P-068). Exactly the
	// distinction drawn just below for `family`.
	_, o.Picking = q["species"]
	picked := map[string]bool{}
	for _, id := range q["species"] {
		if id == "" || picked[id] {
			continue
		}
		picked[id] = true
		o.Picks = append(o.Picks, id)
	}
	_, o.HasSelection = q["family"]
	for _, f := range q["family"] {
		if f != "" {
			o.Families[f] = true
		}
	}
	_, o.HasHabitats = q["habitat"]
	for _, v := range q["habitat"] {
		if v != "" {
			o.Habitats[v] = true
		}
	}
	_, o.HasNiches = q["niche"]
	for _, v := range q["niche"] {
		if v != "" {
			o.Niches[v] = true
		}
	}
	_, o.HasRegions = q["region"]
	for _, g := range q["region"] {
		// A URL carries slugs ("CA", "GulfOfMexico"); everything downstream
		// works in GADM ids, so translate here and nowhere else (CR-024).
		// An unknown slug is dropped rather than guessed at.
		if rg, ok := region.BySlug(g); ok {
			o.Regions[rg.ID] = true
		}
	}
	_, o.HasMonths = q["month"]
	for _, m := range q["month"] {
		// "Jun", not 6: a month number in a URL is off by one as often as
		// not, and an unknown abbreviation is dropped as a region slug is.
		if i, ok := month.Index(m); ok {
			o.Months[i] = true
		}
	}
	return o
}

// tile is one rendered species. The three footer metrics are formatted here
// rather than in the template: a template that does its own arithmetic cannot
// see the unit toggle, which is exactly how the footer came to ignore it.
type tile struct {
	S *dataset.Species
	// Name is the species' English name in the reader's chosen locale,
	// resolved once here so no template has to know about locales (P-049).
	Name    string
	Geom    view.Geometry
	Box     view.Box
	Length  string
	Wing    string
	MassStr string
	// Commonest is the share line: where the bird is the biggest part of
	// what gets recorded, and how big a part (P-074). Nil when unknown.
	Commonest *shareCell
	// SameAs names the other drawn species with the same measurements
	// (P-079). Outside Commonest, because it is about the drawing and must
	// show whether or not there is a count.
	SameAs []seam

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
	// FootLabels name the whole bird rather than any part of it, stacked on
	// the drawing axis below the tail tip (P-066): its total length and its
	// mass. A slice rather than named fields because P-066 requires every
	// metric the detail view reports to appear in the drawing, so the next
	// whole-bird metric should be one more entry and not a new layout.
	//
	// Total length is the bird's reported length, never the drawn height: a
	// folded neck makes those differ by 6-9%.
	FootLabels []footLabel

	// MassCube is the bird's mass drawn as water: one face of the cube of
	// water that weighs the same (P-066, CR-019). Side zero means none.
	MassCube massCube

	// DetailURL is the tile's own link. Built in Go rather than in the
	// template so it carries the reader's view state (P-068); empty on the
	// legend tile, which is an illustration and links nowhere.
	DetailURL template.URL

	// PickURL is the pick control, a SIBLING of the tile link and never a
	// replacement for it -- the tile opens the detail view in both modes, so
	// P-034 holds without exception. Empty outside pick mode, which is what
	// keeps the tile unchanged during ordinary browsing (P-031).
	PickURL template.URL
	Picked  bool
	// Anchor is this tile's fragment id. Picking navigates to it, so the
	// browser restores the reading position itself during load instead of the
	// scroll script jumping there after the page has already painted -- which
	// on a 1.7 MB index is a visible flash (P-038).
	Anchor string
}

// tileAnchor is the fragment id of one tile. Declared once because the href
// and the element's id must agree, and they are written in different files.
func tileAnchor(id string) string { return "s-" + id }

type sectionView struct {
	Index    int
	Min, Max string
	Count    int
	Tiles    []tile
	// Note replaces the range where there is no range to state: the birds
	// with no records in the regions selected have no ratio and so no tier
	// (P-076).
	Note string
}

// frequencyBand labels a decade section (P-076).
//
// The label states the BAND, not the extremes of the birds in it: a tier
// reading "1 in 1,000 - 1 in 9,999" means the same thing on every page,
// where "1 in 1,184 - 1 in 9,781" would be a different pair of numbers under
// every region filter and would invite a reader to compare two tiers that
// are not the same tier. P-024 asks for the range the section covers, and
// under this rule the band IS that range.
func frequencyBand(min, max float64, decades int) (lo, hi, note string) {
	if math.IsInf(min, 1) {
		return "", "", "No records in the regions selected"
	}
	k := int(math.Floor(math.Log10(min)))
	if decades > 1 {
		k -= ((k % decades) + decades) % decades
	}
	start := math.Pow(10, float64(k))
	end := view.Decade(min, decades) - 1
	// The last band runs off the end of the data rather than to a boundary
	// nothing reaches, so it says so instead of naming an empty ceiling.
	if math.IsInf(max, 1) || end > 1e12 {
		return "1 in " + comma(int64(start)), "", ""
	}
	return "1 in " + comma(int64(start)), "1 in " + comma(int64(end)), ""
}

// familyRow is one filter chip, with the URL that toggles it already built:
// a multi-valued query parameter is awkward to assemble in a template.
type familyRow struct {
	familyOption
	Selected bool
	URL      template.URL
}

type pageData struct {
	// Feedback is the issue tracker URL, or empty. Empty publishes nothing
	// (T-025): a site may go up before its tracker exists, but must not
	// invite a reader somewhere that does not.
	Feedback string

	Sections []sectionView
	// FamilyGroupings is the family selector (P-014): families grouped by
	// taxonomic order.
	FamilyGroupings []familyGroupingBlock
	// Traits are the categorical controls of P-015: habitat and niche,
	// assigned per species rather than per family, which is why they can be
	// right where a family-level grouping could not (CR-027).
	Traits []traitBlock
	// RegionGroupings is the region selector (P-011): two groupings of the
	// same regions (P-013), each a labelled block of zones.
	RegionGroupings []groupingBlock
	AllRegions      bool
	AllRegionsURL   template.URL
	NoRegion        bool
	// Months is the twelve-month filter beside the regions (P-011).
	Months       []monthChip
	AllMonths    bool
	AllMonthsURL template.URL
	NoMonth      bool
	Opt          options
	Total        int
	Shown        int
	Query        func(...string) template.URL
	// AllSelected is true only when every family is explicitly filtered in --
	// not when nothing is selected. Both states display the whole dataset, but
	// they are different selections, and the control reflects the selection.
	AllSelected bool
	AllURL      template.URL
	// NothingSelected drives the empty state, which has to say something
	// different from "no matches": the user turned everything off.
	NothingSelected bool

	// Filters describes what is currently narrowing the list, for the empty
	// state to name (P-037). Empty when nothing is narrowing it.
	Filters []activeFilter
	// ClearURL drops every filter, so the dead end has an exit.
	ClearURL template.URL

	// Legend is the labeled example diagram the help opens with (P-035). Nil
	// only if no suitable species exists, in which case the help degrades to
	// prose rather than breaking.
	Legend *tile

	// Bar is the pick-mode furniture of P-068, shared with the detail view.
	Bar pickBar
	// CompareToggleURL enters pick mode with nothing picked, or leaves it.
	// The entry point for a reader who wants to compare before they have a
	// particular bird in mind; the detail view's control is the one for a
	// reader who already does.
	CompareToggleURL template.URL
}

// pickBar is everything the pick bar needs, built identically for the index
// and the detail view so the two cannot drift apart (P-068).
//
// Picks are resolved against the WHOLE dataset rather than the filtered list,
// so a bird the family filter is currently hiding is still named here. A
// filter that appeared to discard picks would be worse than one that refused
// to run.
// activeFilter is one filter and what it is set to, in words.
//
// It exists so the empty state can say WHICH filter emptied the page rather
// than that one did (P-037). The same description belongs in the collapsed
// summary of a filter control, which is why it is built once here rather than
// composed in the template: two places phrasing it separately is two places to
// drift.
type activeFilter struct {
	Label string // "Family"
	Value string // "Parulidae, Icteridae" or "3 of 90 selected"
}

// describeFilters lists what is narrowing the view. A filter that is not
// narrowing anything -- every family selected, which is the arrival state --
// is deliberately absent: naming it would describe the default as a choice.
//
// Verifies: P-037.
func describeFilters(fams []familyOption, opt options) []activeFilter {
	// Past a handful the names stop being readable and the count is the
	// useful fact. The threshold is a presentation judgement, not a rule.
	const nameLimit = 6
	describe := func(picked []string, total int) string {
		if len(picked) > nameLimit {
			return fmt.Sprintf("%d of %d selected", len(picked), total)
		}
		return strings.Join(picked, ", ")
	}

	var out []activeFilter
	// Region first, matching the order the boxes appear in.
	if opt.HasRegions && len(opt.Regions) > 0 && len(opt.Regions) != region.Len() {
		var picked []string
		// Any grouping enumerates every region exactly once, being a
		// partition; the geographic one is used because its order is the one
		// a reader scans a list of place names in.
		for _, z := range region.ZonesOf(region.ByGeography) {
			for _, rg := range region.InZone(region.ByGeography, z) {
				if opt.Regions[rg.ID] {
					picked = append(picked, rg.Name)
				}
			}
		}
		out = append(out, activeFilter{Label: "Region", Value: describe(picked, region.Len())})
	}
	if opt.HasSelection && len(opt.Families) > 0 && len(opt.Families) != len(fams) {
		var picked []string
		for _, f := range fams {
			if opt.Families[f.Name] {
				picked = append(picked, f.Name)
			}
		}
		out = append(out, activeFilter{Label: "Family", Value: describe(picked, len(fams))})
	}
	// Habitat and diet, in the order the controls appear (P-015). Labels
	// rather than AVONET's terms: the reader picked "Eats insects" and
	// should be told "Eats insects", not "Invertivore".
	cat := func(label string, terms []trait.Term, sel map[string]bool, active bool) {
		if !active || len(sel) == 0 || len(sel) == len(terms) {
			return
		}
		var picked []string
		for _, t := range terms {
			if sel[t.Value] {
				picked = append(picked, t.Label)
			}
		}
		out = append(out, activeFilter{Label: label, Value: describe(picked, len(terms))})
	}
	cat("Habitat", trait.Habitats, opt.Habitats, opt.HasHabitats)
	cat("Diet", trait.Niches, opt.Niches, opt.HasNiches)
	return out
}

type pickBar struct {
	Picking    bool
	Picks      []pickChip
	PicksMore  int
	Missing    []string
	CompareURL template.URL
	ClearURL   template.URL
}

// pickChip is one bird in the pick bar, with the link that unpicks it.
type pickChip struct {
	Name      string
	RemoveURL template.URL
}

// healPicks drops picks that no longer resolve and rewrites the request's own
// query to the survivors, returning the dead ids so the page can name them.
//
// Called before anything is built from the request, which is the whole point:
// every link the page goes on to build then omits the dead id, and the set
// repairs itself on the reader's next click (P-068, CR-018). Healing it later
// produced a correct notice beside 970 links that still carried the dead id.
func (s *server) healPicks(r *http.Request, opt *options) []string {
	if !opt.Picking {
		return nil
	}
	found, missing := s.resolvePicks(opt.Picks)
	if len(missing) == 0 {
		return nil
	}
	live := make([]string, 0, len(found))
	for _, sp := range found {
		live = append(live, sp.ID)
	}
	opt.Picks = live
	q := r.URL.Query()
	q.Del("species")
	if len(live) == 0 {
		q.Set("species", "")
	}
	for _, id := range live {
		q.Add("species", id)
	}
	r.URL.RawQuery = q.Encode()
	return missing
}

// buildPickBar assembles the bar, and returns the live picks so the caller can
// mark its own tiles. opt.Picks must already have been healed.
func (s *server) buildPickBar(r *http.Request, opt options, missing []string) (pickBar, []string) {
	b := pickBar{Picking: opt.Picking, Missing: missing}
	if !opt.Picking {
		return b, nil
	}
	found, _ := s.resolvePicks(opt.Picks)
	live := make([]string, 0, len(found))
	for _, sp := range found {
		live = append(live, sp.ID)
	}
	for i, sp := range found {
		if i == pickBarNames {
			b.PicksMore = len(found) - pickBarNames
			break
		}
		b.Picks = append(b.Picks, pickChip{
			Name: sp.Name(opt.Locale), RemoveURL: pickURL(r, toggled(live, sp.ID))})
	}
	b.CompareURL = compareURL(r, live)
	b.ClearURL = queryFor(r, "species")()
	return b, live
}

// pickBarNames caps how many picks the bar names before collapsing to a count.
// There is no cap on the picks themselves (the owner's call), so the bar needs
// one of its own: it is sticky, and an uncapped list of chips would eat the
// viewport it is pinned to.
const pickBarNames = 8

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

	// Heal the request before anything is built from it. A pick that no longer
	// resolves is dropped here and reported in the bar, and the request's own
	// query is rewritten to the surviving set -- so every link the page goes on
	// to build, tile links and control links alike, omits the dead id and the
	// set repairs itself on the reader's next click (P-068, CR-018).
	//
	// Done here rather than where the bar is assembled because by then the
	// links have already been built from the raw query, which is the bug this
	// prevents: the notice appeared while 970 links still carried the dead id.
	missingPicks := s.healPicks(r, &opt)

	// A fresh page selects everything; an emptied selection selects nothing.
	selected := opt.Families
	if !opt.HasSelection {
		selected = make(map[string]bool, len(s.fams))
		for _, f := range s.fams {
			selected[f.Name] = true
		}
	}
	// Region works the same way, and composes with family (P-012): a bird is
	// shown when every active filter admits it.
	inRegion := func(sp *dataset.Species) bool {
		if !opt.HasRegions {
			return true
		}
		for _, id := range sp.Regions {
			if opt.Regions[id] {
				return true
			}
		}
		return false
	}
	// Months narrow the list as every other filter does (P-011, CR-036).
	// A species is shown when SOME selected region admits it in SOME
	// selected month, which is the same "any" the region and habitat
	// filters use: picking two places asks for birds in either.
	allMonths := !opt.HasMonths || len(opt.Months) == month.Count
	inMonth := func(sp *dataset.Species) bool {
		if allMonths {
			return true
		}
		for i, id := range sp.Regions {
			if opt.HasRegions && !opt.Regions[id] {
				continue
			}
			if i >= len(sp.RegionMonths) {
				continue
			}
			// A month eBird withholds this species here is missing by
			// policy, so it is not judged: the species is shown, and its
			// tile says why (CR-037). Painted Bunting was hidden from
			// Florida in January this way -- eBird publishes none of its
			// Florida records from October to March.
			withheld := sp.WithheldIn(id)
			for m := range withheld {
				if withheld[m] && opt.Months[m] {
					return true
				}
			}
			in, ok := month.Present(sp.RegionMonths[i], s.ds.RegionMonthlyTotals[id])
			if !ok {
				// Too thin to say when. Shown in every month rather than
				// hidden from any, and the tile says so (CR-036).
				return true
			}
			for m := range in {
				if in[m] && opt.Months[m] {
					return true
				}
			}
		}
		return false
	}
	// Habitat and niche compose the same way (P-015, P-012). A species
	// matches when ANY of its assignments is selected, and the assignment's
	// origin is irrelevant here: a habitat BirdView added is still a habitat
	// the bird uses, and filtering it out would make the addition pointless.
	anyAssigned := func(as []dataset.Assignment, want map[string]bool, active bool) bool {
		if !active {
			return true
		}
		for _, a := range as {
			if want[a.Value] {
				return true
			}
		}
		return false
	}
	var filtered []*dataset.Species
	for i := range s.ds.Species {
		sp := &s.ds.Species[i]
		if selected[sp.Family] && inRegion(sp) && inMonth(sp) &&
			anyAssigned(sp.Habitats, opt.Habitats, opt.HasHabitats) &&
			anyAssigned(sp.Niches, opt.Niches, opt.HasNiches) {
			filtered = append(filtered, sp)
		}
	}
	// Frequency is not a property of a bird, it is a ratio over the regions
	// the reader has selected (P-074), so it is computed once here for the
	// current selection rather than read off the species. Once per species,
	// not once per comparison: the sort would otherwise walk every region
	// 1,083 log 1,083 times.
	var secs []view.Section
	freq := map[*dataset.Species]float64{}
	if opt.Metric == view.ByFrequency {
		for _, sp := range filtered {
			num, den := shareIn(sp, s.ds, opt)
			if num == 0 || den == 0 {
				// No ratio at all. +Inf sorts them last and keeps them in
				// one final section rather than one section each (P-076).
				freq[sp] = math.Inf(1)
				continue
			}
			freq[sp] = float64(den) / float64(num)
		}
		secs = view.SectionsBy(filtered,
			func(sp *dataset.Species) float64 { return freq[sp] },
			func(min float64) float64 { return view.Decade(min, opt.Decades) })
	} else {
		secs = view.Sections(filtered, opt.Metric, opt.Step)
	}
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

	page := pageData{Feedback: s.feedback, Opt: opt, Total: len(s.ds.Species), Shown: len(filtered),
		Query:    queryFor(r),
		Filters:  describeFilters(s.fams, opt),
		ClearURL: queryFor(r)("family", "region"),
	}
	if sp := s.legendSpecies(); sp != nil {
		lt := labelledTile(sp, legendBox, unitsFor(opt.Metric3, sp.Mass.Value), opt.Locale)
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
	// Either filter emptied to nothing means nothing is selected: the reader
	// turned something off, and no bird can satisfy an empty set (P-037).
	page.NothingSelected = len(selected) == 0 || (opt.HasRegions && len(opt.Regions) == 0)

	// The region selector, and its All control, mirroring the family one.
	page.RegionGroupings = s.regionGroups(r, opt)
	page.AllRegions = !opt.HasRegions || len(opt.Regions) == region.Len()
	everyRegion := map[string]bool{}
	for _, id := range append(region.LandIDs(), region.MarineIDs()...) {
		everyRegion[id] = true
	}
	if page.AllRegions {
		page.AllRegionsURL = s.regionURL(r, nil)
	} else {
		page.AllRegionsURL = s.regionURL(r, everyRegion)
	}
	page.NoRegion = opt.HasRegions && len(opt.Regions) == 0
	page.Months = s.monthChips(r, opt)
	page.AllMonths = !opt.HasMonths || len(opt.Months) == month.Count
	page.NoMonth = opt.HasMonths && len(opt.Months) == 0
	if page.AllMonths {
		// All is a toggle, as it is for regions: pressed when everything is
		// on, it clears everything. That is what makes one month two clicks
		// away and a dedicated control unnecessary (CR-036).
		page.AllMonthsURL = s.monthURL(r, nil)
	} else {
		every := map[int]bool{}
		for i := 0; i < month.Count; i++ {
			every[i] = true
		}
		page.AllMonthsURL = s.monthURL(r, every)
	}

	bar, live := s.buildPickBar(r, opt, missingPicks)
	page.Bar = bar
	// Leaving is exactly what Clear does, so it is the same URL rather than a
	// second one that could drift from it.
	page.CompareToggleURL = bar.ClearURL
	if !opt.Picking {
		page.CompareToggleURL = pickURL(r, nil)
	}
	pickedSet := map[string]bool{}
	for _, id := range live {
		pickedSet[id] = true
	}
	pickOf := func(string) template.URL { return "" }
	if opt.Picking {
		pickOf = func(id string) template.URL {
			// Anchored at the tile it came from: the browser scrolls there as
			// part of loading, rather than painting at the top and then being
			// moved by the script.
			return template.URL(string(pickURL(r, toggled(live, id))) + "#" + tileAnchor(id))
		}
	}
	// One row per family, independent of grouping: a family's own toggle does
	// the same thing whichever block it is clicked in. The rows are then
	// grouped twice (P-014).
	rows := make([]familyRow, 0, len(s.fams))
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
		rows = append(rows, familyRow{familyOption: f, Selected: selected[f.Name], URL: s.familyURL(r, next)})
	}
	page.Traits = []traitBlock{
		s.traitControl(r, "Diet", "What the bird mostly eats. AVONET's classification, quoted — except Drinks sap, which is BirdView's own: AVONET has no category for birds that live off tree sap.",
			"niche", trait.Niches, opt.Niches, opt.HasNiches,
			func(sp *dataset.Species) []dataset.Assignment { return sp.Niches }),
		s.traitControl(r, "Habitat", "Where the bird mostly lives. AVONET's classification, quoted; BirdView adds a second habitat for a few birds and marks those as its own.",
			"habitat", trait.Habitats, opt.Habitats, opt.HasHabitats,
			func(sp *dataset.Species) []dataset.Assignment { return sp.Habitats }),
	}
	page.FamilyGroupings = []familyGroupingBlock{
		{Name: "Taxonomic order", Note: "AviList's taxonomy, in checklist sequence — the order a field guide runs in.",
			Groups: s.groupFamilies(r, selected, rows, nil,
				func(f familyOption) []groupKey {
					return []groupKey{{Name: f.Order, Common: f.OrderCommon}}
				})},
	}
	// The unit scope follows the SCALE setting, because that is what the page
	// is claiming to compare (P-073). Per section, each section shares a unit;
	// globally, so does the whole set.
	var globalUnits units
	if opt.Scale == view.Global {
		all := make([]float64, 0, len(filtered))
		for _, sp := range filtered {
			all = append(all, sp.Mass.Value)
		}
		globalUnits = unitsFor(opt.Metric3, all...)
	}
	for i, sec := range secs {
		u := globalUnits
		if opt.Scale != view.Global {
			ms := make([]float64, 0, len(sec.Species))
			for _, sp := range sec.Species {
				ms = append(ms, sp.Mass.Value)
			}
			u = unitsFor(opt.Metric3, ms...)
		}
		sv := sectionView{Index: i + 1, Count: len(sec.Species),
			Min: u.format(sec.Min, opt.Metric), Max: u.format(sec.Max, opt.Metric)}
		if opt.Metric == view.ByFrequency {
			sv.Min, sv.Max, sv.Note = frequencyBand(sec.Min, sec.Max, opt.Decades)
		}
		for _, sp := range sec.Species {
			g := view.Layout(sp, refs[i], spanRefs[i], tileBox)
			sv.Tiles = append(sv.Tiles, tile{
				S:    sp,
				Name: sp.Name(opt.Locale), Geom: g, Box: tileBox,
				// Set here rather than recomputed in the template: the template
				// used to derive the centering from Box.H itself, which silently
				// disagreed with this value the moment Box stopped being the
				// drawing area (P-066's bottom strip).
				TopOffset: (tileBox.H - g.TotalHeight()) / 2,
				DetailURL: detailLink(r, sp.ID),
				PickURL:   pickOf(sp.ID),
				Picked:    pickedSet[sp.ID],
				Anchor:    tileAnchor(sp.ID),
				// Unlabelled: the tile carries no labels and its footer names
				// the mass already (P-031, CR-020).
				MassCube:    newMassCube(sp, tileBox.H, g.PxPerMm, ""),
				WingModeled: sp.Wingspan.Prov.Kind == dataset.Modeled,
				Length:      u.format(sp.TotalLength.Value, view.ByLength),
				Wing:        u.format(sp.Wingspan.Value, view.ByWingspan),
				MassStr:     u.format(sp.Mass.Value, view.ByMass),
				Commonest:   s.shareWithSeams(sp, opt),
				SameAs:      s.sameMeasurements(sp, opt.Locale),
			})
		}
		page.Sections = append(page.Sections, sv)
	}
	s.render(w, "index.html", page)
}

type detailData struct {
	S *dataset.Species
	// Name is the species' English name in the reader's locale (P-049).
	Name    string
	Geom    view.Geometry
	Box     view.Box
	Metrics []metricRow
	// Categories is every group the bird is sorted into, with who decided
	// each (P-069).
	Categories []categoryRow
	// SameAs names the other drawn species with the same measurements
	// (P-079).
	SameAs []seam
	Opt    options
	Query  func(...string) template.URL

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
	FootLabels    []footLabel
	MassCube      massCube

	// Bar is the pick-mode furniture (P-068), shared with the index.
	Bar pickBar
	// PickURL toggles THIS bird in the comparison, and PickLabel says which of
	// the three states the control is in: not picking, picking but this bird
	// is not in the set, or picking and it already is. Offering "add" for a
	// bird already picked would be a click that silently does nothing, since
	// the comparison view deduplicates.
	PickURL   template.URL
	PickLabel string
	Picked    bool
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
	wanted := strings.TrimPrefix(r.URL.Path, "/species/")
	sp, ok := s.byID[wanted]
	if !ok {
		// A bird the derivation classified and left out of scope gets a page
		// naming it and the measured reason, at 404 (P-075, CR-029). Anything
		// else gets the bare 404: this page is for a bird BirdView knows of
		// and chose not to draw, and it would be a fabrication for a typo.
		if u, known := s.undrawnByID[wanted]; known {
			s.undrawn(w, r, u)
			return
		}
		http.NotFound(w, r)
		return
	}
	opt := parseOptions(r)
	missingPicks := s.healPicks(r, &opt)
	// Wider than the tile's 2.6:1 so the labels of P-066 have margin outside the
	// wingtips. The extra width costs nothing: the scale is already limited by
	// height for almost every bird, so the bird does not grow, the margin does.
	box := view.Box{W: tileBox.W * 2.6, H: tileBox.H * 1.9}
	// The detail view always draws the bird as large as the box allows, whatever
	// scale the index is using. Under a global scale a small bird is a few
	// pixels tall by design, and carrying that into the one view devoted to
	// that single bird would leave nothing to look at (P-034).

	// A detail page compares nothing, so the bird is its own scope and the
	// unit is whichever suits it.
	one := unitsFor(opt.Metric3, sp.Mass.Value)
	rows := []metricRow{
		row("Total length", sp.TotalLength, view.ByLength, one),
		row("Wingspan", sp.Wingspan, view.ByWingspan, one),
		row("Wing chord", sp.WingChord, view.ByLength, one),
		row("Mass", sp.Mass, view.ByMass, one),
		row("Bill", sp.Segments.Bill, view.ByLength, one),
		row("Head", sp.Segments.Head, view.ByLength, one),
		row("Neck", sp.Segments.Neck, view.ByLength, one),
		row("Body", sp.Segments.Body, view.ByLength, one),
		row("Tail", sp.Segments.Tail, view.ByLength, one),
	}
	lt := labelledTile(sp, box, one, opt.Locale)

	// Only a confident match gets a link; zero means none was found (P-064).
	var inatURL string
	if sp.INatTaxonID > 0 {
		inatURL = inat.TaxonURL(sp.INatTaxonID)
	}
	bar, live := s.buildPickBar(r, opt, missingPicks)
	picked := slices.Contains(live, sp.ID)
	// Three states, not two (P-068).
	pickLabel := "Compare with another bird\u2026"
	switch {
	case picked:
		pickLabel = "\u2713 Picked for comparison \u2014 remove"
	case opt.Picking:
		pickLabel = "Pick for comparison"
	}

	s.render(w, "detail.html", detailData{
		S: sp, Name: lt.Name, Geom: lt.Geom, Box: lt.Box, Labels: lt.Labels, TopOffset: lt.TopOffset,
		Bar: bar, Picked: picked, PickLabel: pickLabel,
		PickURL:    pickURL(r, toggled(live, sp.ID)),
		ChordLabel: lt.ChordLabel, WingspanLabel: lt.WingspanLabel, WingspanEst: lt.WingspanEst,
		FootLabels: lt.FootLabels, MassCube: lt.MassCube,
		Metrics: rows, Categories: s.categories(sp, opt), Opt: opt, Query: queryFor(r), INatURL: inatURL,
		SameAs:      s.sameMeasurements(sp, opt.Locale),
		WingModeled: sp.Wingspan.Prov.Kind == dataset.Modeled,
	})
}

// undrawnData is the page for a bird the derivation rejected (P-075).
//
// The prose is assembled in Go rather than branched in the template, because
// the three verdicts say genuinely different things and a template that tries
// to interleave them ends up asserting the intersection — which for `absent`
// would be counts that were never measured.
type undrawnData struct {
	Feedback string
	Query    func(...string) template.URL
	Opt      options

	Name string
	// AlsoCalled is the other locale's English name, present only when it
	// differs. A reader who followed a stale link may have it under the other
	// one (P-049).
	AlsoCalled string
	U          *dataset.Undrawn
	// Verdict is the one-line answer, and Why the paragraph behind it.
	Verdict string
	Why     string
	// Evidence is the measured sentence, empty for `absent` — where there is
	// no measurement, and printing the zeros would present the absence of a
	// measurement as a measurement of absence.
	Evidence string
}

// undrawn renders the page for a species the status table classified and
// scope rejected, at 404 (P-075, CR-029).
//
// 404 and not 410: scope is regenerated from live occurrence data, so a bird
// that misses the recurrence bar on one run can clear it on the next, and
// "gone permanently" is a claim this derivation does not support.
//
// Verifies: P-075.
func (s *server) undrawn(w http.ResponseWriter, r *http.Request, u *dataset.Undrawn) {
	opt := parseOptions(r)
	name := u.Name(opt.Locale)
	other := u.Names[dataset.LocaleNA]
	if opt.Locale == dataset.LocaleNA {
		other = u.Names[dataset.LocaleWorld]
	}
	if other == name {
		other = ""
	}

	where := u.Region
	if rg, ok := region.Get(u.Region); ok {
		where = rg.Name
	}
	// Both thresholds are read from the derivation rather than written out
	// here. A page quoting "three episodes across three years" beside a rule
	// that had moved to four would be a confident, checkable lie.
	bar := fmt.Sprintf("%d episodes across %d years", status.MinEpisodes, status.MinYears)
	recurs := fmt.Sprintf("%s episodes across %s years in %s, from %s records",
		comma(int64(u.Episodes)), comma(int64(u.Years)), where, comma(u.Records))

	d := undrawnData{
		Feedback: s.feedback, Query: queryFor(r), Opt: opt,
		Name: name, AlsoCalled: other, U: u,
	}
	switch u.Status {
	case string(status.Vagrant):
		d.Verdict = "it has turned up, but it does not come back"
		d.Evidence = "Recorded in " + recurs + "."
		d.Why = "BirdView draws the birds whose records recur: repeated arrivals, across " +
			"several years, somewhere in the regions it covers. The bar is " + bar +
			", and this bird has not reached it. A crowd of records is not evidence " +
			"against that — a twitchable rarity draws observers, so a single arrival can " +
			"out-record a resident."
	case string(status.Historical):
		d.Verdict = "its records recur, but they stopped"
		d.Evidence = "Recorded in " + recurs + ", the last of them on " + prettyDate(u.LastSeen) + "."
		d.Why = fmt.Sprintf("Recurrence alone cannot tell a resident from an extinct bird. "+
			"GBIF carries museum specimens and old sight records under their original "+
			"collection dates, so a species can satisfy every test for coming back and "+
			"still not have been seen in a lifetime. A bird last recorded more than %d "+
			"years before the occurrence data ends is filed as historical rather than "+
			"present.", status.RecentYears)
	default:
		d.Verdict = "the open record holds nothing that can be measured"
		d.Why = "BirdView decides scope from the dates on which a species was recorded, " +
			"and for this bird there are too few of them anywhere it covers to ask the " +
			"question. No episode count is shown because none was computed: there are " +
			"no zeros here that were measured. The name is known at all only because " +
			"the occurrence data mentions it in passing."
	}

	// The body says what the status code says, and the status code is set
	// before a byte of it is written.
	w.WriteHeader(http.StatusNotFound)
	s.render(w, "undrawn.html", d)
}

// prettyDate turns an ISO date from the dataset into one a reader reads.
// Falls back to the input, which is already legible, rather than to a blank.
func prettyDate(iso string) string {
	t, err := time.Parse("2006-01-02", iso)
	if err != nil {
		return iso
	}
	return t.Format("2 January 2006")
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

// sortOption is one choice in the comparison view's Sort by control.
//
// On is a fact about the panels as they stand, not a stored mode: the sort
// rewrites the species order in the URL, so there is nowhere for "sorted by
// mass" to be remembered, and nothing to go stale when a bird is added.
type sortOption struct {
	Label string
	URL   template.URL
	On    bool
}

type compareData struct {
	Panels []comparePanel
	Opt    options
	Query  func(...string) template.URL
	// StopURL leaves pick mode: the index with the picks dropped and every
	// other carried setting kept. The same URL as the pick bar's Clear, so
	// the two exits cannot drift apart (P-068).
	StopURL template.URL
	// Sorts is empty below two birds, there being nothing to sort.
	Sorts []sortOption
	// Missing are picked ids that no longer resolve, named on the page so a
	// rotted link says what it lost instead of silently shrinking (CR-018).
	Missing []string
}

// compare renders several birds to one shared scale (P-067).
func (s *server) compare(w http.ResponseWriter, r *http.Request) {
	opt := parseOptions(r)
	// Named order is the reader's and is preserved; a repeat is drawn once,
	// both handled when the parameter is parsed.
	//
	// An id that no longer resolves is dropped and named rather than refused
	// (CR-018). This page used to 404, which was the honest option until
	// drop-and-name existed; the case that decides it is not a typo but rot,
	// since a comparison link is made to be kept and ids follow scientific
	// names that get reassigned.
	missing := s.healPicks(r, &opt)
	chosen, _ := s.resolvePicks(opt.Picks)

	// One scale for the whole set, taken from the largest bird in it, exactly
	// as a section does (P-042). Laying each panel out against its own length
	// would fill every panel equally and compare nothing.
	var refLength, refWingspan float64
	masses := make([]float64, 0, len(chosen))
	for _, sp := range chosen {
		refLength = math.Max(refLength, sp.TotalLength.Value)
		refWingspan = math.Max(refWingspan, sp.Wingspan.Value)
		masses = append(masses, sp.Mass.Value)
	}
	// One unit for the whole set, for the same reason there is one scale
	// (P-073). The set named in the address IS the scope here -- comparing a
	// swan with a sparrow must not print "4.2 lb" beside "0.9 oz", which is
	// two scales in the one place the page exists to put birds on one.
	u := unitsFor(opt.Metric3, masses...)

	box := view.Box{W: tileBox.W * 2.6, H: tileBox.H * 1.9}
	panels := make([]comparePanel, 0, len(chosen))
	for _, sp := range chosen {
		panels = append(panels, comparePanel{
			tile:      labelledTileAt(sp, box, u, refLength, refWingspan, opt.Locale),
			DetailURL: detailLink(r, sp.ID),
			RemoveURL: compareWithout(r, sp.ID),
		})
	}
	// The species list goes back with the reader: the index is the picker, and
	// "back" here means "add another bird" (CR-018). Clear, on the index, is
	// the one control that drops it.
	s.render(w, "compare.html", compareData{Panels: panels, Opt: opt,
		Sorts: sortOptions(r, chosen), Missing: missing, Query: queryFor(r),
		StopURL: queryFor(r, "species")()})
}

// resolvePicks turns requested ids into species, separating the ones that no
// longer exist (P-068).
//
// Missing ids are dropped and reported rather than refused, because the case
// that matters is not a typo but rot: ids follow scientific names and this
// checklist renames them -- the Sandhill Crane is antigone_canadensis and was
// grus_canadensis. A comparison link is made to be kept, so it outlives some of
// its own ids, and refusing the whole page would kill a five-bird comparison
// over one reassigned genus. See CR-018.
func (s *server) resolvePicks(ids []string) (found []*dataset.Species, missing []string) {
	for _, id := range ids {
		if sp, ok := s.byID[id]; ok {
			found = append(found, sp)
		} else {
			missing = append(missing, id)
		}
	}
	return found, missing
}

// carried is the reader's view state minus the one parameter too large to copy
// across 841 links. See detailLink for why `family` is the exception.
func carried(r *http.Request) url.Values {
	q := r.URL.Query()
	q.Del("family")
	return q
}

// pickURL is the current page with the picks replaced. It keeps every other
// parameter INCLUDING `family`, because picking a bird must not silently drop
// the filter the reader is picking from (P-068).
//
// An empty set writes a bare `species=`, which holds pick mode open with
// nothing picked rather than dropping the reader back into browsing mid-task.
func pickURL(r *http.Request, ids []string) template.URL {
	q := r.URL.Query()
	q.Del("species")
	if len(ids) == 0 {
		q.Set("species", "")
	}
	for _, id := range ids {
		q.Add("species", id)
	}
	return template.URL("/?" + q.Encode())
}

// sortedIDs is the species reordered by one metric, ascending and stably, as
// the index orders its sections (P-021).
func sortedIDs(sp []*dataset.Species, m view.Metric) []string {
	s := append([]*dataset.Species(nil), sp...)
	sort.SliceStable(s, func(i, j int) bool { return m.Value(s[i]) < m.Value(s[j]) })
	out := make([]string, len(s))
	for i, x := range s {
		out[i] = x.ID
	}
	return out
}

// sortOptions builds the Sort by control. Each link is an ordinary comparison
// URL naming the same birds in a new order (P-067), so a sorted comparison is
// indistinguishable from one assembled by hand and there is no second place an
// ordering can live.
func sortOptions(r *http.Request, chosen []*dataset.Species) []sortOption {
	if len(chosen) < 2 {
		return nil
	}
	now := make([]string, len(chosen))
	for i, sp := range chosen {
		now[i] = sp.ID
	}
	var out []sortOption
	for _, m := range []struct {
		label  string
		metric view.Metric
	}{{"Length", view.ByLength}, {"Wingspan", view.ByWingspan}, {"Mass", view.ByMass}} {
		ids := sortedIDs(chosen, m.metric)
		out = append(out, sortOption{
			Label: m.label, URL: compareURL(r, ids), On: slices.Equal(ids, now)})
	}
	return out
}

// compareURL sends the current picks to the comparison view.
func compareURL(r *http.Request, ids []string) template.URL {
	q := carried(r)
	q.Del("species")
	for _, id := range ids {
		q.Add("species", id)
	}
	return template.URL("/compare?" + q.Encode())
}

// toggled returns the picks with one id added or removed. Adding appends, so
// the reader's order is the order they picked in.
func toggled(picks []string, id string) []string {
	out := make([]string, 0, len(picks)+1)
	found := false
	for _, p := range picks {
		if p == id {
			found = true
			continue
		}
		out = append(out, p)
	}
	if !found {
		out = append(out, id)
	}
	return out
}

// detailLink builds a link from any page into a species detail view, carrying
// the view state the reader is in -- minus the one parameter too large to copy
// 841 times (P-068).
//
// The picks ride along because the detail view cannot otherwise tell whether to
// offer to pick that bird, nor build a link to the set plus it; `units` because
// the page formats every value with it, and a link without it is why opening a
// bird from a metric index used to land on an imperial page; `by`, `scale` and
// `step` so the back link rebuilds the index the reader left.
//
// `family` is deliberately dropped. It is the only parameter with no bound on
// its length: 90 families is about 1.7 KB, which repeated across 841 tile
// hrefs would add roughly 1.4 MB to one page. The cost, accepted in P-068, is
// that the in-page back link cannot restore the filter. The browser's own Back
// button does, because it replays the previous URL.
func detailLink(r *http.Request, id string) template.URL {
	q := r.URL.Query()
	q.Del("family")
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

// The strip reserved at the bottom of an annotated diagram for the whole-bird
// labels (P-066), the vertical counterpart of labelGutter.
//
// Without it they are clipped rather than merely cramped: the detail view
// scales the bird to fill the box, so any bird whose height binds has its tail
// tip exactly on the bottom edge and nothing at all below it.
//
// footGap is where the first baseline sits below the tail tip and footLine is
// the spacing between them; footDescender keeps the glyphs of the last line
// inside the viewBox, not merely its baseline.
const (
	footGap       = 15
	footLine      = 13
	footDescender = 9
)

// footLabel is one whole-bird annotation under the tail.
type footLabel struct {
	Key  string // total, mass -- also the CSS hook
	Text string
	Y    float64
	Est  bool
}

// massCube is the cube of water that weighs what the bird weighs, drawn in the
// lower-left in oblique projection (P-066, CR-019).
//
// X,Y is the top-left of the FRONT FACE, and the face's base sits on the
// bottom of the drawing area in every diagram. That is what makes two cubes
// comparable at a glance: the projection grows up and to the right from a
// corner that never moves, so only the cube's size differs between birds.
type massCube struct {
	X, Y, Side, Depth float64
	// The mass label sits centred under the face rather than in the stack
	// under the tail, so the number is beside the shape it describes.
	Label          string
	LabelX, LabelY float64
	Est            bool // mass not measured -- dashed, per P-065
	Title          string
}

// massCubeDepth is the oblique projection's depth, as a fraction of the side,
// offset at 45 degrees.
//
// Cabinet projection's conventional 0.5 also fits -- measured over all 841
// species, neither value lets the cube reach the crossbar -- but it leaves
// only 14px of clearance on Western Grebe against this one's 21px, and the
// unit tests cannot see that margin because the fixture holds five birds and
// no grebe. The looser value is the one that survives a data change.
const massCubeDepth = 0.4

// massLabelGap is how far the mass label's baseline sits below the face.
//
// Wider than the footGap the stack under the tail uses, and deliberately: the
// tail is a hairline and the cube is a solid block, so the same gap that looks
// airy under one looks cramped under the other.
const massLabelGap = 18

// massLabelCharWidth approximates one character of the 11px label, used only
// to stop a small cube's centred label running off the left edge of the
// drawing -- a Scissor-tailed Flycatcher's cube is 29px wide and its label
// about 70px, which rendered as "ass 1.4 oz".
//
// Deliberately generous: overestimating only nudges the label right, while
// underestimating clips it again.
const massLabelCharWidth = 6.2

// massCubeSide is the side of that cube, in millimetres, for a mass in grams.
//
// One gram of water is one cubic centimetre, by the historical definition of
// the gram, so this is a unit conversion and not a model: it assumes nothing
// about the density of birds. That is what lets the square share the diagram's
// own scale instead of needing one of its own, which P-043 would refuse.
func massCubeSide(massG float64) float64 {
	if massG <= 0 {
		return 0
	}
	return 10 * math.Cbrt(massG) // cm -> mm
}

// massCubeMargin insets the cube from the corner of the drawing area.
const massCubeMargin = 8

// newMassCube places the cube of water in the lower-left of a drawing area
// drawH tall, at pxPerMm (P-066, CR-019).
//
// label is empty on an index tile, which carries no labels at all and names
// the mass in its footer instead (P-031, CR-020). Everything else about the
// cube is identical there, including the face standing on the floor, so the
// two surfaces cannot drift.
func newMassCube(sp *dataset.Species, drawH, pxPerMm float64, label string) massCube {
	side := massCubeSide(sp.Mass.Value) * pxPerMm
	if side <= 0 {
		return massCube{}
	}
	base := drawH - massCubeMargin
	c := massCube{
		X: massCubeMargin, Y: base - side, Side: side, Depth: side * massCubeDepth,
		Est:   sp.Mass.Prov.Kind != dataset.Measured,
		Title: "Mass as water: a cube of water weighing the same, to scale",
	}
	if label != "" {
		c.Label = label
		// Centred under the face, but never far enough left to be clipped.
		c.LabelX = math.Max(massCubeMargin+side/2, float64(len(label))*massLabelCharWidth/2)
		c.LabelY = base + massLabelGap
	}
	return c
}

// footStrip is the height n stacked whole-bird labels need.
func footStrip(n int) float64 {
	if n == 0 {
		return 0
	}
	return footGap + float64(n-1)*footLine + footDescender
}

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
func labelledTile(sp *dataset.Species, box view.Box, u units, loc dataset.Locale) tile {
	return labelledTileAt(sp, box, u, sp.TotalLength.Value, sp.Wingspan.Value, loc)
}

// labelledTileAt is the same diagram laid out against an EXTERNAL reference, so
// several birds can share one scale (P-067).
//
// That reference is the whole of the difference between the two views: the
// bird's own length makes a detail page fill its box, and the set maximum makes
// a comparison mean something. Getting the argument wrong yields a page that
// looks entirely correct and compares nothing, which is why AC-076 asserts
// drawn heights against the length ratio rather than asserting on the markup.
func labelledTileAt(sp *dataset.Species, box view.Box, u units, refLength, refWingspan float64, loc dataset.Locale) tile {
	// draw is the area the bird gets; the emitted box is that plus the strip the
	// total length label sits in, so the label costs the bird nothing. Going the
	// other way -- keeping the box and shrinking the bird into it -- would take
	// 8.5% of the detail view's height and contradict P-034, which requires the
	// bird drawn as large as the view allows. Width has no such freedom, which
	// is why labelGutter does shrink the bird and this does not.
	// Every metric the detail view reports has to appear in the drawing
	// (P-066). The two that name the whole bird rather than a part of it go
	// under the tail; mass is here and not in the stack because it is not a
	// length and has no part to point at.
	//
	// The reported total length, never g.TotalHeight(): for a folded neck the
	// drawing is shorter than the bird (CR-011), and the label names the bird.
	foot := []footLabel{{
		Key:  "total",
		Text: "Total length  " + u.format(sp.TotalLength.Value, view.ByLength),
		Est:  sp.TotalLength.Prov.Kind != dataset.Measured,
	}}

	draw := box
	// The strip has to clear whichever reaches lower: the stack under the tail
	// or the mass label under the cube.
	box.H += math.Max(footStrip(len(foot)), massLabelGap+footDescender)

	// The bird occupies everything left of the gutter; draw stays the full
	// height, because the label column is positioned against the emitted box.
	inner := view.Box{W: draw.W - labelGutter, H: draw.H}
	g := view.Layout(sp, refLength, refWingspan, inner)
	top := (draw.H - g.TotalHeight()) / 2
	segOf := map[string]dataset.Metric{
		"bill": sp.Segments.Bill, "head": sp.Segments.Head, "neck": sp.Segments.Neck,
		"body": sp.Segments.Body, "tail": sp.Segments.Tail,
	}
	t := tile{S: sp, Name: sp.Name(loc), Geom: g, Box: box, TopOffset: top,
		WingModeled: sp.Wingspan.Prov.Kind == dataset.Modeled}
	// draw, not box: the segment label column belongs in the drawing area, and
	// must not spill into the strip the total length label owns.
	for _, l := range g.DetailLabels(draw, top) {
		m := segOf[l.Key]
		t.Labels = append(t.Labels, diagramLabel{
			Label: l,
			Text:  strings.ToUpper(l.Key[:1]) + l.Key[1:] + "  " + u.format(m.Value, view.ByLength),
			Est:   m.Prov.Kind == dataset.Modeled,
		})
	}
	if sp.WingChord.Value > 0 {
		t.ChordLabel = "Wing chord  " + u.format(sp.WingChord.Value, view.ByLength)
	}
	t.WingspanLabel = "Wingspan  " + u.format(sp.Wingspan.Value, view.ByWingspan)
	t.WingspanEst = sp.Wingspan.Prov.Kind == dataset.Modeled
	for i := range foot {
		foot[i].Y = top + g.YTailTip + footGap + float64(i)*footLine
	}
	t.FootLabels = foot

	// Lower-left of the DRAWING area, not the emitted box: the box is taller by
	// the foot strip and wider by the label gutter, and the square belongs
	// inside the bird's own area. Measured over the whole dataset before this
	// was built -- the crossbar reaches into that corner for 0 of 841 species,
	// the tightest clearance being 50px on Western Grebe.
	t.MassCube = newMassCube(sp, draw.H, g.PxPerMm,
		"Mass  "+u.format(sp.Mass.Value, view.ByMass))
	return t
}

// row turns a metric into a display line, stating its provenance in plain
// language (P-063).
func row(label string, m dataset.Metric, as view.Metric, u units) metricRow {
	r := metricRow{Label: label, Value: u.format(m.Value, as), Kind: string(m.Prov.Kind)}
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
			u.format(m.Min, as), u.format(m.Max, as))
	}
	return r
}

func (s *server) render(w http.ResponseWriter, name string, data any) {
	tmpl := s.tmpl
	// In dev mode the templates are re-parsed per request, which is the whole
	// point: parsing once at startup is exactly how an edited template goes on
	// serving its old self until someone restarts. Serving from a directory
	// without this fixes the CSS and leaves the HTML stale, which is worse
	// than not having the flag -- it looks like it works.
	if s.devDir != "" {
		t, err := template.New("").Funcs(funcs).ParseFS(assetFS, "static/*.html")
		if err != nil {
			log.Printf("birdview: dev re-parse failed, serving the last good templates: %v", err)
		} else {
			tmpl = t
		}
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	if err := tmpl.ExecuteTemplate(w, name, data); err != nil {
		log.Printf("birdview: rendering %s: %v", name, err)
	}
}

// queryFor returns a helper that rewrites the current query string, so a
// control can toggle one setting without discarding the others.
//
// drop names parameters the current page owns and the index does not
// understand. Every link built here points at the index, and the index
// preserves whatever it is handed -- so a stray parameter does not just look
// untidy on one link, it attaches itself to every control the reader touches
// afterwards. The comparison view drops its own species list for that reason.
func queryFor(r *http.Request, drop ...string) func(...string) template.URL {
	return func(kv ...string) template.URL {
		q := r.URL.Query()
		for _, k := range drop {
			q.Del(k)
		}
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
// gramsPerOunce converts to the avoirdupois ounce.
const gramsPerOunce = 28.349523125

// units fixes how every number in one comparison is written.
//
// The mass unit is a property of the SCOPE, not of the value. Chosen per
// value, a section could print "14.5 oz" beside "1.2 lb" for two birds of
// the same length -- arithmetic the reader has to do, on the page whose
// whole purpose is comparison. Length never switches, so only mass needs
// this.
//
// The scope is what the page is comparing: a section or the whole set on the
// index, depending on the scale setting, and the birds named in the address
// on the comparison page. A detail page compares nothing, so its scope is
// the one bird.
type units struct {
	// metric selects centimetres and grams over inches and ounces.
	metric bool
	// big selects kilograms or pounds over grams or ounces.
	big bool
}

// unitsFor picks the unit a scope shares.
//
// The SMALLEST mass decides, so a scope holding one small bird is written in
// the small unit throughout: a swan beside a sparrow reads as "4.2 lb" and
// "0.9 oz" otherwise, and the two cannot be compared by eye at all. Writing
// the swan as 67 oz keeps one scale across the set, which is the point.
// Taking the largest instead would round every small bird to 0.0 lb.
func unitsFor(metric bool, masses ...float64) units {
	u := units{metric: metric, big: true}
	seen := false
	for _, m := range masses {
		if m <= 0 {
			continue // no mass recorded; it cannot decide the unit
		}
		seen = true
		if metric && m < 1000 {
			u.big = false
		}
		if !metric && m/gramsPerOunce < 16 {
			u.big = false
		}
	}
	if !seen {
		u.big = false
	}
	return u
}

func (u units) format(v float64, as view.Metric) string {
	if as == view.ByMass {
		if u.metric {
			if u.big {
				return fmt.Sprintf("%.2f kg", v/1000)
			}
			return fmt.Sprintf("%.0f g", v)
		}
		oz := v / gramsPerOunce
		switch {
		case u.big:
			return fmt.Sprintf("%.1f lb", oz/16)
		case oz >= 1:
			return fmt.Sprintf("%.1f oz", oz)
		default:
			// A 3 g hummingbird is 0.11 oz; one decimal would erase the
			// difference between the smallest birds entirely. This is a
			// precision choice within one unit, not a second unit, so it
			// stays per value.
			return fmt.Sprintf("%.2f oz", oz)
		}
	}
	if u.metric {
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

// regionRow is one region chip, and zoneGroup one block of them. They mirror
// familyRow and orderGroup exactly: zones play the part orders play, so the
// two selectors share markup, CSS and the wrap criterion (P-011, AC-075).
type regionRow struct {
	ID   string
	Code string
	Name string
	// Title is the tooltip: the name, plus the gloss where a name needs one
	// to say which stretch of ocean it means.
	Title    string
	Selected bool
	URL      template.URL
}

type zoneGroup struct {
	Zone     string
	Selected bool
	URL      template.URL
	// OnlyURL selects this zone's regions and nothing else. The plain URL
	// toggles the zone against whatever is already on, which is the right
	// default and a slow way to get to one zone from everywhere -- the
	// reader would have to clear all, then pick. Only is that in one click.
	OnlyURL template.URL
	Regions []regionRow
}

// groupingBlock is one way of carving the regions up, rendered as its own
// labelled block of chips (P-013). The Note is not decoration: it is where
// the derived grouping and the hand-drawn one are told apart.
type groupingBlock struct {
	Name  string
	Note  string
	Zones []zoneGroup
}

// regionURL builds the link that results from toggling regions, using the
// same absent / explicitly-empty distinction as familyURL: no parameter means
// everywhere, `region=` means nowhere.
func (s *server) regionURL(r *http.Request, want map[string]bool) template.URL {
	q := r.URL.Query()
	q.Del("region")
	switch {
	case len(want) == region.Len():
		// default: no parameter
	case len(want) == 0:
		q.Set("region", "")
	default:
		// `want` is keyed by GADM id; the URL carries slugs (CR-024).
		slugs := make([]string, 0, len(want))
		for id := range want {
			rg, ok := region.Get(id)
			if !ok {
				continue
			}
			slugs = append(slugs, rg.Slug())
		}
		sort.Strings(slugs)
		for _, s := range slugs {
			q.Add("region", s)
		}
	}
	if len(q) == 0 {
		return "/"
	}
	return template.URL("/?" + q.Encode())
}

// monthURL is regionURL for the twelve (P-011). Same three cases: all of
// them is the default and carries no parameter, none of them is an explicit
// empty, and anything between is listed.
func (s *server) monthURL(r *http.Request, want map[int]bool) template.URL {
	q := r.URL.Query()
	q.Del("month")
	switch {
	case len(want) == month.Count:
		// default: no parameter
	case len(want) == 0:
		q.Set("month", "")
	default:
		// In calendar order, not the map's: a URL that reshuffled its own
		// months between loads would make two identical selections look
		// like two different pages.
		for i := 0; i < month.Count; i++ {
			if want[i] {
				q.Add("month", month.Abbrev[i])
			}
		}
	}
	if len(q) == 0 {
		return "/"
	}
	return template.URL("/?" + q.Encode())
}

// monthChip is one of the twelve in the control.
type monthChip struct {
	Abbrev   string
	Name     string
	Selected bool
	URL      template.URL
	Title    string
}

// monthChips builds the twelve toggles. No groups and no season names:
// P-076 forbids naming a band where the range can be stated, and a season
// inverts between hemispheres and means nothing in the tropics, two of
// which BirdView covers (CR-036).
func (s *server) monthChips(r *http.Request, opt options) []monthChip {
	out := make([]monthChip, 0, month.Count)
	for i := 0; i < month.Count; i++ {
		on := !opt.HasMonths || opt.Months[i]
		next := map[int]bool{}
		for j := 0; j < month.Count; j++ {
			if !opt.HasMonths || opt.Months[j] {
				next[j] = true
			}
		}
		if on {
			delete(next, i)
		} else {
			next[i] = true
		}
		title := "Add " + month.Full[i]
		if on {
			title = "Remove " + month.Full[i]
		}
		out = append(out, monthChip{
			Abbrev: month.Abbrev[i], Name: month.Full[i], Selected: on,
			URL: s.monthURL(r, next), Title: title,
		})
	}
	return out
}

// regionGroups builds the region selector: each grouping, its zones, their
// regions, and the link each toggle produces. A region appears once per
// grouping, so the same chip is rendered twice with two different URLs —
// which is correct, since toggling it from either place does the same thing.
func (s *server) regionGroups(r *http.Request, opt options) []groupingBlock {
	sel := opt.Regions
	if !opt.HasRegions {
		sel = make(map[string]bool, region.Len())
		for _, id := range append(region.LandIDs(), region.MarineIDs()...) {
			sel[id] = true
		}
	}
	var out []groupingBlock
	for _, grouping := range region.Groupings {
		block := groupingBlock{Name: grouping.Name, Note: grouping.Note}
		out = append(out, block)
	}
	for bi, grouping := range region.Groupings {
		for _, z := range grouping.Zones {
			g := zoneGroup{Zone: z, Selected: true}
			inZone := region.InZone(grouping.Name, z)
			for _, rg := range inZone {
				if !sel[rg.ID] {
					g.Selected = false
				}
			}
			// The group toggle clears its zone when the whole zone is on, and
			// selects it otherwise -- an order header's behaviour exactly.
			next := map[string]bool{}
			for k, v := range sel {
				if v {
					next[k] = true
				}
			}
			for _, rg := range inZone {
				if g.Selected {
					delete(next, rg.ID)
				} else {
					next[rg.ID] = true
				}
			}
			g.URL = s.regionURL(r, next)

			// Only: exactly this zone's regions, whatever was on before.
			// Built from inZone alone rather than from sel, so it does not
			// depend on the current selection and one zone is reachable from
			// any state of the page in a single click.
			only := make(map[string]bool, len(inZone))
			for _, rg := range inZone {
				only[rg.ID] = true
			}
			g.OnlyURL = s.regionURL(r, only)

			for _, rg := range inZone {
				one := map[string]bool{}
				for k, v := range sel {
					if v {
						one[k] = true
					}
				}
				if sel[rg.ID] {
					delete(one, rg.ID)
				} else {
					one[rg.ID] = true
				}
				title := rg.Name
				if rg.Gloss != "" {
					title += ", " + rg.Gloss
				}
				g.Regions = append(g.Regions, regionRow{
					ID: rg.ID, Code: rg.Code, Name: rg.Name, Title: title,
					Selected: sel[rg.ID], URL: s.regionURL(r, one),
				})
			}
			out[bi].Zones = append(out[bi].Zones, g)
		}
	}
	return out
}

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
	// whynot explains, in a tooltip, why a value is not this bird's own
	// measurement. The three footer cells used to say "estimated" for
	// anything not measured, which is false for a near-match and said so
	// right beside a stroke P-065 draws SOLID because it IS a measurement.
	"whynot": whyNotMeasured,
}

// whyNotMeasured turns a provenance into a short tooltip.
//
// It names the donor for a near-match, because that is the fact a reader can
// act on: two tiles showing an identical mass are not a coincidence when one
// of them borrowed it from the other, and nothing else on the index says so.
//
// Verifies: P-062, P-065.
func whyNotMeasured(p dataset.Provenance) string {
	switch p.Kind {
	case dataset.NearMatch:
		if p.SourceTaxon != "" {
			return "measured, but on " + p.SourceTaxon + " \u2014 not this species"
		}
		return "measured on a related species, not this one"
	case dataset.Modeled:
		if p.Model != "" {
			return "estimated: " + p.Model
		}
		return "estimated"
	}
	return "estimated"
}

// traitRow is one option of a categorical control: its label, AVONET's
// definition, how many species carry it, and the link that toggles it
// (P-015).
type traitRow struct {
	Value       string
	Label       string
	Description string
	// Coined marks a category BirdView added where AVONET has none. The
	// control says so: every other row leads with AVONET's term, and a
	// term of ours must not borrow that authority (P-015).
	Coined   bool
	Count    int
	Selected bool
	URL      template.URL
	// OnlyURL selects this value alone within its axis. Habitat and diet are
	// separate axes and separate parameters, so this never touches the other.
	OnlyURL template.URL
}

// traitBlock is one categorical control -- habitat or niche.
type traitBlock struct {
	Name   string
	Note   string
	Param  string
	All    bool
	AllURL template.URL
	Rows   []traitRow
}

// traitURL builds the link for a categorical filter, with the same
// absent-versus-empty distinction every other filter uses: no parameter
// means all, `habitat=` means none.
func (s *server) traitURL(r *http.Request, param string, want map[string]bool, total int) template.URL {
	q := r.URL.Query()
	q.Del(param)
	switch {
	case len(want) == total:
		// default: no parameter
	case len(want) == 0:
		q.Set(param, "")
	default:
		vals := make([]string, 0, len(want))
		for v := range want {
			vals = append(vals, v)
		}
		sort.Strings(vals)
		for _, v := range vals {
			q.Add(param, v)
		}
	}
	if len(q) == 0 {
		return "/"
	}
	return template.URL("/?" + q.Encode())
}

// traitControl builds one categorical block.
//
// **Counts are over the whole dataset, not the current view.** A count that
// moved with the other filters would be telling the reader how many birds
// they would still see, which the species count above the list already
// says; what this answers is "how many birds are there of this kind", which
// is a fact about the data and should not shift underfoot.
func (s *server) traitControl(r *http.Request, name, note, param string,
	terms []trait.Term, sel map[string]bool, active bool,
	get func(*dataset.Species) []dataset.Assignment) traitBlock {

	count := map[string]int{}
	for i := range s.ds.Species {
		seen := map[string]bool{}
		for _, a := range get(&s.ds.Species[i]) {
			if !seen[a.Value] {
				seen[a.Value] = true
				count[a.Value]++
			}
		}
	}

	on := func(v string) bool {
		if !active {
			return true // arrival: everything is selected
		}
		return sel[v]
	}

	// An option no bird carries is a dead control: it renders a row, offers
	// a click, and returns an empty page. AVONET's vocabulary is worldwide
	// and this dataset is one continent's worth, so Desert has no ABA-area
	// species at all. Offer only what is there.
	offered := terms[:0:0]
	for _, t := range terms {
		if count[t.Value] > 0 {
			offered = append(offered, t)
		}
	}
	terms = offered

	cur := map[string]bool{}
	for _, t := range terms {
		if on(t.Value) {
			cur[t.Value] = true
		}
	}

	b := traitBlock{Name: name, Note: note, Param: param, All: len(cur) == len(terms)}
	if b.All {
		b.AllURL = s.traitURL(r, param, nil, len(terms))
	} else {
		every := map[string]bool{}
		for _, t := range terms {
			every[t.Value] = true
		}
		b.AllURL = s.traitURL(r, param, every, len(terms))
	}
	for _, t := range terms {
		next := map[string]bool{}
		for k := range cur {
			next[k] = true
		}
		if next[t.Value] {
			delete(next, t.Value)
		} else {
			next[t.Value] = true
		}
		b.Rows = append(b.Rows, traitRow{
			Value: t.Value, Label: t.Label, Description: t.Description,
			Count: count[t.Value], Selected: on(t.Value),
			URL: s.traitURL(r, param, next, len(terms)),
			// Only: this value alone within its axis. Habitat and diet carry
			// separate parameters, so narrowing one leaves the other whole.
			OnlyURL: s.traitURL(r, param, map[string]bool{t.Value: true}, len(terms)),
		})
	}
	return b
}

// categoryRow is one line of the detail view's category table (P-069): what
// the bird is sorted as, and who decided.
type categoryRow struct {
	// Label is blank on a continuation row, so two habitats read as one
	// entry with two values rather than as two unrelated facts.
	Label  string
	Value  string
	Origin string
	Detail string
	// Doc, when set, is a document BirdView serves that explains the rule.
	Doc, DocLabel string
}

// categories lists every category a species is in, with its provenance.
//
// **Regions are one row, not sixty.** Every region a bird is listed in was
// decided by the same rule over the same data, so a row apiece would repeat
// one sentence sixty times and bury the two facts that differ -- habitat and
// diet, where BirdView sometimes disagrees with AVONET. The regions
// themselves are listed inline by their postal codes.
func (s *server) categories(sp *dataset.Species, opt options) []categoryRow {
	loc := opt.Locale
	out := []categoryRow{
		{Label: "Family", Value: sp.Family + " — " + sp.FamilyCommon,
			Origin: string(dataset.Published), Detail: "AviList v2025b, CC BY 4.0"},
		{Label: "Order", Value: sp.Order,
			Origin: string(dataset.Published), Detail: "AviList v2025b, CC BY 4.0"},
	}

	add := func(label string, as []dataset.Assignment, terms []trait.Term) {
		for i, a := range as {
			row := categoryRow{Value: trait.LabelOf(terms, a.Value), Origin: string(a.Origin)}
			if i == 0 {
				row.Label = label
			}
			switch a.Origin {
			case dataset.Published:
				row.Detail = a.Source + ", as " + a.Value
			default:
				row.Detail = a.Why
			}
			if a.Doc != "" {
				row.Doc, row.DocLabel = "/docs/"+a.Doc, "how BirdView adds one"
			}
			out = append(out, row)
		}
	}
	add("Habitat", sp.Habitats, trait.Habitats)
	add("Diet", sp.Niches, trait.Niches)

	// One row for every region, sharing the one rule that put them all there.
	if len(sp.Regions) > 0 {
		codes := make([]string, 0, len(sp.Regions))
		for _, id := range sp.Regions {
			if rg, ok := region.Get(id); ok {
				codes = append(codes, rg.Code)
				if rg.Code == "" {
					codes[len(codes)-1] = rg.Name
				}
			}
		}
		sort.Strings(codes)
		out = append(out, categoryRow{
			Label:  fmt.Sprintf("Regions (%d)", len(codes)),
			Value:  strings.Join(codes, " "),
			Origin: string(dataset.Derived),
			Detail: "reported there in at least 1 record per million, from eBird and " +
				"iNaturalist records via GBIF",
			Doc: "/docs/data#regions", DocLabel: "how regions are decided",
		})
	}

	// Every way this bird's name or record differs between AviList, which
	// supplies the names, and the occurrence data, which supplies the
	// evidence (P-077). Placed before the count rows deliberately: all four
	// change how the numbers below should be read.
	//
	// All of them, not the first that matches. This was an if/else-if over
	// two of the four tables, which would have shown a shared name OR a
	// withheld note had any species ever been both. None is today, and the
	// chain was still the wrong shape: a disclosure silently suppressed by
	// another disclosure is exactly what G-006 forbids.
	//
	// Why a reader wants these and not just the number: a visitor who knows
	// the bird as Dryobates villosus and finds a page headed Leuconotopicus
	// has no way to tell a taxonomic update from a mistake, and no way to go
	// back to eBird and check. Naming the alias answers both.
	var seams []categoryRow
	seam := func(value, detail string) {
		seams = append(seams, categoryRow{
			Value: value, Origin: string(dataset.Derived), Detail: detail,
			Doc:      "/docs/data#when-the-sources-disagree-about-a-name",
			DocLabel: "how names are joined",
		})
	}

	// An alias: one taxon, two labels. No count is affected, which is why
	// this says so outright -- a row that merely noted a discrepancy would
	// leave a reader wondering what it cost.
	if old, ok := taxonomy.AviListRenamedFrom[sp.ScientificName]; ok {
		what, how := renamed(old, sp.ScientificName)
		seam(old+" — the "+what+" eBird still uses",
			how+" BirdView publishes AviList's name and joins on eBird's, so one taxon "+
				"carries two labels here: every record on this page is this bird's, and "+
				"the counts are unaffected.")
	}

	// A name whose records are this species' under the taxonomy BirdView
	// serves, because the sources draw the boundary somewhere else.
	//
	// The reason is carried per entry and not written here, because it is
	// not the same reason twice: AviList split the Whimbrel and eBird did
	// not follow, while AviList declined to split the Siberian Stonechat
	// and both databases did. One sentence covering both named the wrong
	// taxonomic event for two of the three, and claimed AviList does not
	// carry Numenius phaeopus, which it does -- as the Eurasian bird.
	for _, c := range taxonomy.CountedInto(sp.ScientificName) {
		seam(c.Published+" — the name eBird files these birds under",
			c.Why+" BirdView follows AviList and counts them as this species, rather "+
				"than choosing between the two databases or publishing the bird twice.")
	}

	// A shared name: the one case where the count is not this bird's alone.
	if other, src, why, ok := taxonomy.SharedWith(sp.ScientificName); ok {
		name := other
		if o := s.byID[s.idOf[other]]; o != nil {
			name = fmt.Sprintf("%s (%s)", o.Name(loc), other)
		}
		seam(src+" — a name covering two species",
			why+" The counts on this page are therefore shared with "+name+
				", and BirdView does not divide them: no source says how they divide.")
	}

	// A form filed below the species. Newly true of the counts rather than
	// merely of the data: until the rank fold these trinomials matched no
	// species at all and were discarded, and now they are counted here.
	if why, ok := taxonomy.DomesticForms[sp.ScientificName]; ok {
		seam("domestic and feral birds, counted here",
			why+", below the species rank. BirdView folds a form into the species it is a "+
				"form of, so the counts on this page include those birds. They are not "+
				"separated, because no source separates them.")
	}

	// Withheld is NOT a name disagreement and must not read as one: telling
	// a reader to look for another name would send them after one that does
	// not exist (P-077).
	if taxonomy.WithheldByEBird[sp.ScientificName] {
		seam("withheld by eBird",
			"eBird suppresses "+sp.Name(loc)+" from its public export, as it does for "+
				"species vulnerable to disturbance. The counts here are iNaturalist's alone, "+
				"so they are far lower than the bird's actual record. This is a gap in what "+
				"BirdView can see, not a statement about the bird.")
	}

	// Withheld in particular regions or seasons, under eBird's Sensitive
	// Species list and confirmed against the records (CR-037). Every region,
	// not just the ones on screen: this page is the complete record, and the
	// tile -- which names only what is on screen -- must never say more
	// (P-034). A species eBird withholds everywhere already has its row above.
	if !taxonomy.WithheldByEBird[sp.ScientificName] && len(sp.RegionWithheld) > 0 {
		type where struct{ name, months string }
		var ws []where
		for id := range sp.RegionWithheld {
			name := id
			if rg, ok := region.Get(id); ok {
				name = rg.Name
			}
			ws = append(ws, where{name, maskScope(sp.WithheldIn(id))})
		}
		sort.Slice(ws, func(i, j int) bool { return ws[i].name < ws[j].name })
		value := fmt.Sprintf("withheld by eBird in %d regions", len(ws))
		if len(ws) == 1 {
			value = "withheld by eBird in " + ws[0].name + ", " + ws[0].months
		}
		parts := make([]string, len(ws))
		for i, w := range ws {
			parts[i] = w.name + " (" + w.months + ")"
		}
		seam(value, "eBird keeps this species out of its public records in "+
			strings.Join(parts, ", ")+", to protect it from trapping, disturbance or targeted "+
			"hunting. In those months the counts here are iNaturalist's alone and far below "+
			"the bird's real record, and the month filter does not judge them: a month with "+
			"no eBird records by policy is not a month without the bird.")
	}

	// One label over the group, continuation rows blank, as the trait rows
	// above do it.
	if len(seams) > 0 {
		seams[0].Label = "Recorded as"
		out = append(out, seams...)
	}

	// The month filter declined to judge this bird, which the tile marks
	// and so this must too (P-034, CR-036).
	if where, n, yes := s.abstains(sp, opt); yes {
		out = append(out, categoryRow{
			Label:  "Months",
			Value:  "shown in every month — too few records to say",
			Origin: string(dataset.Derived),
			Detail: fmt.Sprintf("BirdView holds %s records of this bird in %s, under the %d "+
				"it takes to tell the months a bird is present from the months nobody was "+
				"looking. Hiding it from a month on evidence that thin would state "+
				"something about the bird that is really something about the survey.",
				comma(n), where, month.MinRecords),
			Doc: "/docs/data#how-common-a-bird-is", DocLabel: "what this measures",
		})
	}

	// The figure the tile showed, for the regions the reader had selected
	// when they clicked it (P-034).
	//
	// Built by shareWithSeams, the same function the tile calls, rather
	// than recomputed here. Two call sites deriving one number from the
	// same inputs is how the index and the detail page came to disagree
	// about where a bird is commonest (CR-031); one function cannot.
	//
	// It is NOT the row below. "Top region" is where the bird peaks, which
	// is a different scope and usually a smaller N, and a reader who saw
	// 1 in 23,393 on the tile and found only 1 in 2,986 here had no way to
	// tell that the page had changed the question.
	if cell := s.shareWithSeams(sp, opt); cell != nil {
		value, detail := cell.Note, cell.Title
		if cell.Ratio != "" {
			value = cell.Ratio + " species observations " + cell.Where
			detail = cell.Title
			for _, sm := range cell.Seams {
				detail += " " + sm.Title
			}
		}
		out = append(out, categoryRow{
			Label:  "Frequency",
			Value:  value,
			Origin: string(dataset.Derived),
			Detail: detail,
			Doc:    "/docs/data#how-common-a-bird-is", DocLabel: "what this measures",
		})
	}

	// Where the bird is the biggest share of what gets recorded (P-074).
	//
	// Stated as "1 in N species observations there" rather than as a percentage: the
	// share runs from 1 in 2 to 1 in 117 million across the dataset, and a
	// percentage is 0.000001% at that end and unreadable well before it.
	// The records it rests on are shown so a thin number discloses itself
	// instead of needing a minimum-evidence cutoff.
	if c := sp.Commonest; c != nil && c.OneIn() > 0 {
		where := c.Region
		if rg, ok := region.Get(c.Region); ok {
			where = rg.Name
		}
		out = append(out, categoryRow{
			// "Top region". It began as the bare "Commonest in", which said
			// nothing about what it ranged over, and spent a while as
			// "Commonest land region" while the seas had no honest
			// denominator (CR-030). They have one now (CR-031), so the row
			// ranges over everything again and the qualifier went -- and
			// "top" says plainly that this is the maximum of the same
			// figure the index shows, which selecting this one region there
			// reproduces exactly (AC-100).
			Label:  "Top region",
			Value:  fmt.Sprintf("%s — 1 in %s species observations there", where, comma(int64(c.OneIn()))),
			Origin: string(dataset.Derived),
			// Both halves of the contrast, because the first on its own only
			// says what the number is NOT. A share is where the bird is.
			Detail: fmt.Sprintf("%s of the %s species observations from %s, in eBird via GBIF -- and "+
				"iNaturalist where eBird withholds the bird. That is how much of what people record there is this bird, not "+
				"how many of them there are: a plain count says where the birdwatchers are. "+
				"A sea's records are those on open water inside its box, so a bird watched "+
				"from the shore counts to the shore.",
				comma(c.Records), comma(c.RegionRecords), where),
			Doc: "/docs/data#how-common-a-bird-is", DocLabel: "what this measures",
		})
	}
	return out
}

// shareCell is the tile's share line, pre-rendered: the region's short code,
// the share as one-in-N, and the full sentence as a tooltip.
type shareCell struct {
	// Where names the scope the share is over, already prefixed with "in":
	// the one region when exactly one is on, "every region" when all are,
	// and "the selected regions" for any set in between.
	//
	// Not "the ABA Area", which this is not. That area includes St-Pierre
	// et Miquelon, which BirdView has no region for, and excludes Johnston,
	// Palmyra, Wake and Navassa, which it has regions for. Naming someone
	// else's boundary would also re-attribute BirdView's scope to the
	// authority CR-028 just removed it from. The selection describes itself.
	Where string
	// Ratio is the whole left-hand phrase, "1 in 883,328" or "0 in
	// 1,225,176,035". Both forms are spelled out rather than the second
	// collapsing to an em-dash: a dash makes a reader wonder whether the
	// number is missing or the bird is, where "0 in 1,225,176,035 bird
	// records across every land region" says which, and says how hard anyone
	// looked before concluding it.
	Ratio string
	Title string
	// Note replaces the ratio when there is nothing to divide by at all --
	// the reader has selected only seas, which carry presence and no share
	// (P-074, CR-030). Empty whenever Ratio is set, and never both: a cell
	// either divides or explains.
	Note string
	// Seams are caveats on the figure, one line each under it: a count
	// shared with another species (P-077, CR-034), a bird the month filter
	// declined to judge (CR-036), a species eBird withholds in the regions
	// on screen (CR-037). They accompany the Ratio rather than replacing it:
	// suppressing a count would leave a reader unable to tell it from a bird
	// genuinely absent (G-006).
	//
	// A list, and every caveat that applies is on it. The first version held
	// one, and the abstention marker replaced the shared-name marker when
	// both applied -- a disclosure silently suppressed by another, which is
	// the failure G-006 names.
	Seams []seam
}

// seam is one caveat line on a tile: the short text, and the tooltip that
// explains it.
type seam struct{ Text, Title string }

// shareIn gives a species' share of the species observations held by the regions
// currently on screen: its records there, over every bird observation from the
// same regions.
//
// **Computed per request, and no subaggregate is precomputed.** The obvious
// worry is combinatorial -- 69 regions is 2^69 possible selections -- but it
// never arises, because the metric is a ratio of two SUMS over the selection
// and sums compose from their parts. The per-region counts are already the
// finest-grained subaggregate there is; any selection is one pass over them.
// Measured at 2.6 ms for the worst case in a Python sketch, which is 1,081
// species against all 69 regions; in Go it does not register.
//
// A region the reader selected where the species was not reported adds to
// the denominator and nothing to the numerator, which is correct: widening
// the selection to places a bird is absent from should lower its share.
//
// **A region with no entry in totals contributes to NEITHER side** (CR-030).
// That is the whole guard, and it replaces a pair of conditions that drifted:
// the numerator ranged over the species' regions and the denominator over the
// totals, so when the marine regions turned out to have counts on one side
// and no honest denominator on the other, every pelagic's share was computed
// as sea records over land records. Deriving both sides from "is there a
// denominator for this region" makes that mismatch unrepresentable rather
// than merely fixed, which matters because the two loops read as a matching
// pair and were not one.
func shareIn(sp *dataset.Species, ds *dataset.Dataset, opt options) (num, den int64) {
	totals := ds.RegionTotals
	all := !opt.HasRegions || len(opt.Regions) == region.Len()
	counted := func(id string) bool {
		if _, ok := totals[id]; !ok {
			return false
		}
		return all || opt.Regions[id]
	}
	// Every month unless the reader said otherwise. Both halves of the
	// ratio are narrowed together, which is the whole point: a June figure
	// divided by June's birding is comparable with a December one, where
	// June's records over the year's birding would not be (P-011).
	allMonths := !opt.HasMonths || len(opt.Months) == month.Count
	inMonth := func(m int) bool { return allMonths || opt.Months[m] }

	// The species' regions and the regions where it stays under the bar
	// both count: the bar decides where it is listed, not what it counts
	// (P-011, CR-042).
	add := func(ids []string, counts []int64, months [][12]int64) {
		for i, id := range ids {
			if !counted(id) {
				continue
			}
			if allMonths {
				if i < len(counts) {
					num += counts[i]
				}
				continue
			}
			if i < len(months) {
				for m, n := range months[i] {
					if inMonth(m) {
						num += n
					}
				}
			}
		}
	}
	add(sp.Regions, sp.RegionCounts, sp.RegionMonths)
	add(sp.BelowRegions, sp.BelowCounts, sp.BelowMonths)
	for id, n := range totals {
		if !counted(id) {
			continue
		}
		if allMonths {
			den += n
			continue
		}
		for m, v := range ds.RegionMonthlyTotals[id] {
			if inMonth(m) {
				den += v
			}
		}
	}
	return num, den
}

// shareOf builds the tile's "1 in N" line for the regions currently on
// or nil when there is nothing to divide.
//
// It deliberately does NOT reuse the detail page's "commonest in" figure.
// That one answers "where in North America is this bird most of what people
// see", which is trivia a reader looking at a filtered index did not ask:
// filtering to Arizona and being told about Yukon is a non sequitur. This
// answers the question the page is actually posing.
func shareOf(sp *dataset.Species, ds *dataset.Dataset, opt options) *shareCell {
	all := !opt.HasRegions || len(opt.Regions) == region.Len()
	if opt.HasRegions && len(opt.Regions) == 0 {
		return nil // nothing selected; the page shows no birds anyway
	}
	if opt.HasMonths && len(opt.Months) == 0 {
		return nil // likewise: no month is selected, so nothing is in view
	}
	num, den := shareIn(sp, ds, opt)
	// No denominator at all means nothing on screen is a land region --
	// in practice, the reader has selected only seas, which carry presence
	// and no share (CR-030). Named for the selection rather than for the
	// sea, because that is the thing the reader can act on.
	// No denominator for anything selected. Unreachable from the UI today,
	// every region having a total -- but it is how the marine regions
	// behaved for months before CR-031 measured one, so a region added to
	// internal/region ahead of its data lands here rather than dividing by
	// zero.
	if den == 0 {
		return &shareCell{
			// NOT "no records here". The bird may have plenty: this fires
			// when BirdView lacks the REGION's total, and the numerator is
			// zero only because a region with no denominator is skipped on
			// both sides. Buller's Shearwater has 39,143 records in the
			// North Pacific, and would land here reading "no records" if
			// that sea's total ever went missing again -- which is a
			// statement about the bird, and false.
			Note: "No data for these regions yet",
			Title: "BirdView does not know how many birds have been recorded in the regions " +
				"selected, so it cannot say what part of them this one is. That is a gap in " +
				"BirdView's data, not a statement about the bird.",
		}
	}
	where, scope := "in the selected regions", "the selected regions"
	if !all && len(opt.Regions) == 1 {
		for id := range opt.Regions {
			if rg, ok := region.Get(id); ok {
				where, scope = "in "+rg.Name, rg.Name
			}
		}
	} else if all {
		where, scope = "(all regions)", "every region BirdView covers"
	}
	// The months the figure ranges over, where that is not the whole year.
	// Appended to both halves rather than replacing either: a reader needs
	// to know the figure narrowed, and narrowing in two dimensions at once
	// is exactly when a scope label stops being obvious (P-011).
	if ms := monthScope(opt); ms != "" {
		if all {
			where = "(all regions, " + ms + ")"
		} else {
			where += ", " + ms
		}
		scope += " in " + ms
	}
	// Zero is stated, not dashed. "0 in 1,225,176,035 species observations across
	// every land region" carries the effort behind the absence; an em-dash
	// left a reader unable to tell a missing figure from a missing bird.
	ratio, detail := "0 in "+comma(den), "None"
	if num > 0 {
		ratio, detail = "1 in "+comma(den/num), comma(num)
	}
	return &shareCell{
		Where: where,
		Ratio: ratio,
		Title: fmt.Sprintf("%s of the %s species observations from %s are this species. "+
			"That is how much of what people record is this bird, not how many of them "+
			"are out there.",
			detail, comma(den), scope),
	}
}

// monthScope names the months a figure ranges over, or "" for the whole
// year, where the scope is the default and saying so is noise.
//
// Runs of three or more collapse to "Jun-Aug", and a run is allowed to WRAP
// the year end, so a winter selection reads "Nov-Feb" rather than
// "Jan, Feb, Nov, Dec". Winter is the one selection a birder is most likely
// to make and the one a naive renderer mangles worst.
func monthScope(opt options) string {
	if !opt.HasMonths || len(opt.Months) == 0 || len(opt.Months) == month.Count {
		return ""
	}
	var mask [month.Count]bool
	for m := range opt.Months {
		mask[m] = true
	}
	return maskScope(mask)
}

// maskScope names a set of months the way monthScope does -- runs of three
// collapsed, the year end allowed to wrap -- with two ends of its own: no
// months is "", and all twelve is "all year", because a withheld-by-eBird
// caveat covering the whole year should say so rather than say nothing.
func maskScope(mask [month.Count]bool) string {
	n := 0
	for _, on := range mask {
		if on {
			n++
		}
	}
	switch n {
	case 0:
		return ""
	case month.Count:
		return "all year"
	}
	// Find a start: a set month whose predecessor is not set. With fewer
	// than twelve set one always exists.
	start := -1
	for i := 0; i < month.Count; i++ {
		if mask[i] && !mask[(i+month.Count-1)%month.Count] {
			start = i
			break
		}
	}
	var runs [][]int
	var cur []int
	for k := 0; k < month.Count; k++ {
		i := (start + k) % month.Count
		if mask[i] {
			cur = append(cur, i)
			continue
		}
		if len(cur) > 0 {
			runs = append(runs, cur)
			cur = nil
		}
	}
	if len(cur) > 0 {
		runs = append(runs, cur)
	}
	parts := make([]string, 0, len(runs))
	for _, r := range runs {
		if len(r) >= 3 {
			parts = append(parts, month.Abbrev[r[0]]+"-"+month.Abbrev[r[len(r)-1]])
			continue
		}
		for _, i := range r {
			parts = append(parts, month.Abbrev[i])
		}
	}
	return strings.Join(parts, ", ")
}

// shareWithSeams is shareOf plus the disclosure of P-077: where an
// occurrence source files two BirdView species under one name, the tile says
// so rather than presenting the count as this bird's alone.
//
// On the tile and not only on the detail page, because the index is where
// the misreading happens. A Caribbean mangrove resident ranked 51st
// most-recorded in North America looks like a fact until you open it.
func (s *server) shareWithSeams(sp *dataset.Species, opt options) *shareCell {
	cell := shareOf(sp, s.ds, opt)
	if cell == nil || cell.Ratio == "" {
		return cell
	}
	if text, title, yes := s.withheldSeam(sp, opt); yes {
		cell.Seams = append(cell.Seams, seam{text, title})
	}
	if where, n, yes := s.abstains(sp, opt); yes {
		// The reason is on the tile and not only in the tooltip: "shown in
		// every month" alone reads as a fact about the bird, which is the
		// one misreading the abstention exists to prevent (CR-036).
		cell.Seams = append(cell.Seams, seam{"Shown in every month due to too few records",
			fmt.Sprintf("BirdView holds %s records of this bird in %s, which "+
				"is too few to tell the months it is here from the months nobody looked. It is "+
				"shown whatever months you select, rather than hidden from some on evidence "+
				"that thin.", comma(n), where)})
	}
	if other, src, _, ok := taxonomy.SharedWith(sp.ScientificName); ok {
		name := other
		if o := s.byID[s.idOf[other]]; o != nil {
			name = o.Name(opt.Locale)
		}
		// Sentence case, like the other caveats: each sits on its own line
		// under the figure and reads as a sentence, not as a fragment of it.
		cell.Seams = append(cell.Seams, seam{"Shared with " + name,
			fmt.Sprintf("This count is not this bird's alone. The occurrence data "+
				"files both it and %s under %s, and does not separate them, so BirdView does not "+
				"divide the records either.", name, src)})
	}
	return cell
}

// sharedRows groups the drawn species whose traits come from one AVONET row:
// the same sheet and the same filed taxon. Such species are drawn alike
// because their measurements are the same, not because the birds are (P-079).
//
// Derived from the provenance rather than from taxonomy.NearMatches, so a
// borrowing reached any other way still gets its note.
func sharedRows(ds *dataset.Dataset) map[string][]string {
	byRow := map[string][]string{}
	for i := range ds.Species {
		p := ds.Species[i].Segments.Bill.Prov
		if p.SourceTaxon == "" {
			continue
		}
		key := p.Source + "\x00" + p.SourceTaxon
		byRow[key] = append(byRow[key], ds.Species[i].ScientificName)
	}
	out := map[string][]string{}
	for _, names := range byRow {
		if len(names) < 2 {
			continue
		}
		sort.Strings(names)
		for _, n := range names {
			for _, o := range names {
				if o != n {
					out[n] = append(out[n], o)
				}
			}
		}
	}
	return out
}

// sameMeasurements gives one caveat per other drawn species that shares this
// one's AVONET row, for the tile and the detail page (P-079). The tooltip
// says which way the borrowing runs, because the two species are not in the
// same position: one has no measurements of its own.
func (s *server) sameMeasurements(sp *dataset.Species, loc dataset.Locale) []seam {
	var out []seam
	row := sp.Segments.Bill.Prov
	where := fmt.Sprintf("%s, %s", row.SourceTaxon, strings.TrimPrefix(row.Source, "AVONET v7, "))
	for _, other := range s.sameAs[sp.ScientificName] {
		o := s.byID[s.idOf[other]]
		if o == nil {
			continue
		}
		name := o.Name(loc)
		var why string
		switch {
		case sp.Segments.Bill.Prov.Kind == dataset.NearMatch && o.Segments.Bill.Prov.Kind != dataset.NearMatch:
			why = fmt.Sprintf("AVONET has no measurements of this bird, so it is drawn from %s's (%s). "+
				"The two drawings are alike because the measurements are the same.", name, where)
		case o.Segments.Bill.Prov.Kind == dataset.NearMatch && sp.Segments.Bill.Prov.Kind != dataset.NearMatch:
			why = fmt.Sprintf("AVONET has no measurements of %s, so it is drawn from this bird's (%s). "+
				"The two drawings are alike because the measurements are the same.", name, where)
		default:
			why = fmt.Sprintf("This bird and %s are both drawn from one AVONET row (%s). "+
				"The two drawings are alike because the measurements are the same.", name, where)
		}
		out = append(out, seam{"Same measurements as " + name, why})
	}
	return out
}

// withheldIn returns, for the regions and months on screen, the months eBird
// withholds a species in each region that has any (CR-037). Months outside
// the selection are dropped, so a June filter does not report Florida's
// winter.
func (s *server) withheldIn(sp *dataset.Species, opt options) map[string][12]bool {
	allMonths := !opt.HasMonths || len(opt.Months) == month.Count
	out := map[string][12]bool{}
	for id := range sp.RegionWithheld {
		if opt.HasRegions && !opt.Regions[id] {
			continue
		}
		w := sp.WithheldIn(id)
		set := false
		for m := range w {
			if !allMonths && !opt.Months[m] {
				w[m] = false
			}
			set = set || w[m]
		}
		if set {
			out[id] = w
		}
	}
	return out
}

// withheldSeam says, on the tile, that eBird withholds this species in the
// regions and months on screen, so its figure is short by however many
// records eBird kept back (CR-037).
//
// One region is named with its months, because that is the case a reader
// can act on: "Withheld by eBird in Florida, Oct-Mar". Several are counted,
// because thirty-eight state names do not fit on a tile, and the detail page
// lists them.
func (s *server) withheldSeam(sp *dataset.Species, opt options) (text, title string, yes bool) {
	w := s.withheldIn(sp, opt)
	if len(w) == 0 {
		return "", "", false
	}
	title = "eBird keeps this species out of its public records here, to protect it -- " +
		"from trapping, disturbance or targeted hunting -- so the figure counts only what " +
		"iNaturalist holds in those months and is far below the bird's real record. That " +
		"is a gap in what BirdView can see, not a statement about the bird."
	if len(w) == 1 {
		for id, months := range w {
			name := id
			if rg, ok := region.Get(id); ok {
				name = rg.Name
			}
			text = "Withheld by eBird in " + name
			if sc := maskScope(months); sc != "" {
				text += ", " + sc
			}
		}
		return text, title, true
	}
	return fmt.Sprintf("Withheld by eBird in %d regions", len(w)), title, true
}

// abstains reports whether the month filter declined to judge this species,
// and in which region it has the most to go on (CR-036).
//
// A species under 200 records in a region cannot have its months told apart
// from its surveys: Millerbird holds 28 in Hawaii, all in five months, and
// lives on Nihoa the year round. It is shown in every selected month, and
// saying so is the whole of the honesty here — the alternative is a page
// that hides a bird from June on the strength of nobody having sailed to
// Nihoa in June.
//
// Reported only when a month filter is actually narrowing something. With
// the whole year selected nothing was judged, so there is nothing to
// disclose and a marker would be noise on every tile.
func (s *server) abstains(sp *dataset.Species, opt options) (where string, most int64, yes bool) {
	if !opt.HasMonths || len(opt.Months) == 0 || len(opt.Months) == month.Count {
		return "", 0, false
	}
	var best string
	var bestN int64 = -1
	for i, id := range sp.Regions {
		if opt.HasRegions && !opt.Regions[id] {
			continue
		}
		if i >= len(sp.RegionMonths) {
			continue
		}
		if _, ok := month.Present(sp.RegionMonths[i], s.ds.RegionMonthlyTotals[id]); ok {
			// Some region in view had enough to judge on, so the filter did
			// its job and the species is here on its merits.
			return "", 0, false
		}
		if n := sp.RegionCounts[i]; n > bestN {
			best, bestN = id, n
		}
	}
	if bestN < 0 {
		return "", 0, false
	}
	if rg, ok := region.Get(best); ok {
		best = rg.Name
	}
	return best, bestN, true
}

// comma groups an integer with thousands separators. The shares reach the
// hundred millions, where an ungrouped run of digits cannot be read at all.
func comma(n int64) string {
	s := strconv.FormatInt(n, 10)
	if len(s) <= 3 {
		return s
	}
	var b strings.Builder
	for i, c := range s {
		if i > 0 && (len(s)-i)%3 == 0 {
			b.WriteByte(',')
		}
		b.WriteRune(c)
	}
	return b.String()
}

// served names the two document sets BirdView publishes: the user-facing
// guides and the normative spec. Both are embedded, so a provenance
// citation resolves without the network (P-069, T-002).
var served = map[string]struct {
	fs   embed.FS
	kind string
}{
	"docs": {docs.Files, "documentation"},
	"spec": {river.Files, "requirement"},
}

// document serves one embedded Markdown file, rendered.
func (s *server) document(w http.ResponseWriter, r *http.Request) {
	set, rest, _ := strings.Cut(strings.TrimPrefix(r.URL.Path, "/"), "/")
	src, ok := served[set]
	if !ok {
		http.NotFound(w, r)
		return
	}
	if rest == "" {
		s.documentIndex(w, set)
		return
	}
	// A diagram, pre-rendered beside the page it belongs to.
	if name, ok := strings.CutSuffix(rest, ".svg"); ok && name != "" && !strings.ContainsAny(name, "/.") {
		svg, err := src.fs.ReadFile(name + ".svg")
		if err != nil {
			http.NotFound(w, r)
			return
		}
		w.Header().Set("Content-Type", "image/svg+xml")
		// Opened full size, an SVG is a document of its own; it carries no
		// script, and this says so to the browser as well (T-002).
		w.Header().Set("Content-Security-Policy", "default-src 'none'; style-src 'unsafe-inline'")
		w.Write(svg)
		return
	}
	// Bare names only, so there is no traversal to reason about.
	if strings.ContainsAny(rest, "/.") {
		http.NotFound(w, r)
		return
	}
	body, err := src.fs.ReadFile(rest + ".md")
	if err != nil {
		http.NotFound(w, r)
		return
	}
	title := markdownTitle(body, rest)
	s.renderMarkdown(w, set, title, withDiagram(src.fs, set, rest, title, body))
}

// mermaidBlock is a page's diagram source. render-diagrams.py reads the same
// pattern, one block per page.
var mermaidBlock = regexp.MustCompile("(?s)```mermaid\n.*?\n```")

// withDiagram replaces a page's Mermaid source with the picture rendered
// from it.
//
// GitHub draws the source itself; BirdView would need Mermaid's script to,
// which T-002 rules out. A page without its rendered SVG keeps the source,
// shown as code, and TestDiagramsAreCurrent fails.
//
// The picture links to itself because it is several times the page's width:
// shrunk to fit, its field names are too small to read.
func withDiagram(fsys embed.FS, set, name, title string, body []byte) []byte {
	if _, err := fsys.ReadFile(name + ".svg"); err != nil {
		return body
	}
	src := "/" + set + "/" + name + ".svg"
	img := fmt.Sprintf("[![Diagram: %s](%s)](%s)\n\n*Select the diagram to open it full size.*",
		title, src, src)
	return mermaidBlock.ReplaceAllLiteral(body, []byte(img))
}

// documentIndex lists a set, so /spec and /docs are not dead ends.
func (s *server) documentIndex(w http.ResponseWriter, set string) {
	src := served[set]
	entries, _ := src.fs.ReadDir(".")
	var b bytes.Buffer
	fmt.Fprintf(&b, "# BirdView %s documents\n\n", src.kind)
	for _, e := range entries {
		name, ok := strings.CutSuffix(e.Name(), ".md")
		if !ok {
			continue // a diagram, listed by the page it belongs to
		}
		body, _ := src.fs.ReadFile(e.Name())
		fmt.Fprintf(&b, "- [%s](/%s/%s)\n", markdownTitle(body, name), set, name)
	}
	s.renderMarkdown(w, set, "BirdView "+src.kind+" documents", b.Bytes())
}

// markdownTitle is a document's first top-level heading, falling back to
// its file name.
//
// **The first heading, not the first line.** river/process.md opens with an
// HTML comment recording where the file came from, and reading line one
// listed it in the index as "<!--".
func markdownTitle(src []byte, fallback string) string {
	for _, line := range strings.Split(string(src), "\n") {
		if t, ok := strings.CutPrefix(line, "# "); ok {
			return strings.TrimSpace(t)
		}
	}
	return fallback
}

func (s *server) renderMarkdown(w http.ResponseWriter, set, title string, src []byte) {
	var buf bytes.Buffer
	if err := markdown[set].Convert(src, &buf); err != nil {
		log.Printf("birdview: rendering %q: %v", title, err)
		http.Error(w, "could not render that document", http.StatusInternalServerError)
		return
	}
	s.render(w, "docs.html", struct {
		Title string
		Body  template.HTML
	}{title, template.HTML(buf.String())})
}

// repoLinks rewrites the repository-relative links these documents carry
// into the paths BirdView serves them at.
//
// **Rewritten at render time, not in the files.** `../river/product.md` is
// the correct link in a repository checkout and `/spec/product` is the
// correct one in a browser; editing the files would fix the second by
// breaking the first, and these are served precisely because they are the
// files the repository carries.
//
// The `base` is the set the document being rendered belongs to, because a
// bare sibling link means different things in each: `product.md` inside
// river/ is a requirement document, and the same link inside docs/ is a
// guide. Rewriting both to /docs/ was the first attempt and sent half the
// spec to a 404.
//
// **The directory and the URL are named separately.** The requirements live
// in river/, the directory River's method gives them, and are served at
// /spec, which is what a reader of the site understands; servedAt maps one
// to the other so that neither name has to bend to the other.
type repoLinks struct{ base string }

var repoLink = regexp.MustCompile(`^(?:\.\./)?(river|docs)(?:/([a-z-]+?)(?:\.md)?)?(#.*)?$`)

// servedAt maps a repository directory to the URL its documents are served at.
var servedAt = map[string]string{"river": "spec", "docs": "docs"}

func (t repoLinks) Transform(doc *goldmarkast.Document, reader goldmarktext.Reader, _ goldmarkparser.Context) {
	_ = goldmarkast.Walk(doc, func(n goldmarkast.Node, entering bool) (goldmarkast.WalkStatus, error) {
		link, ok := n.(*goldmarkast.Link)
		if !entering || !ok {
			return goldmarkast.WalkContinue, nil
		}
		m := repoLink.FindStringSubmatch(string(link.Destination))
		if m == nil {
			// A bare sibling document, as docs/ links to its neighbours.
			if sib := regexp.MustCompile(`^([a-z-]+)\.md(#.*)?$`).
				FindStringSubmatch(string(link.Destination)); sib != nil {
				link.Destination = []byte("/" + t.base + "/" + sib[1] + sib[2])
			}
			return goldmarkast.WalkContinue, nil
		}
		dest := "/" + servedAt[m[1]] + "/"
		if m[2] != "" {
			dest += m[2]
		}
		link.Destination = []byte(dest + m[3])
		return goldmarkast.WalkContinue, nil
	})
}

// markdown renders the served documents. GFM tables are enabled because the
// documents use them heavily -- the source tables and the work-arising list
// are tables -- and without the extension they render as runs of pipes.
//
// Raw HTML in the source is NOT passed through: these files are ours, but a
// renderer that emits whatever HTML it is given is a habit worth not having.
// idAnchors gives a heading whose text begins with a requirement id -- "P-001
// — Scope is ..." -- the id alone as its anchor.
//
// Every cross-reference in the spec is written `(product.md#p-001)`, and
// river/check/refs.py checks those resolve. It checks the MARKDOWN. Served,
// goldmark's auto id turned that heading into
// `p-001--scope-is-the-aba-area-avifauna-that-recurs-by-birdviews-own-measure`,
// so all 176 anchored cross-references in the rendered spec landed at the top
// of the document instead of on the requirement they cite. Nothing noticed,
// because the link-walking test drops the fragment before checking anything.
//
// Runs after WithAutoHeadingID, replacing the generated id where a heading
// opens with an id token and leaving every other heading alone.
//
// Verifies: P-069.
type idAnchors struct{}

// idToken matches a leading requirement id: P-001, T-024, AC-091, CR-028, G-004.
var idToken = regexp.MustCompile(`^(?:P|T|AC|CR|G)-\d+\b`)

func (idAnchors) Transform(doc *goldmarkast.Document, reader goldmarktext.Reader, _ goldmarkparser.Context) {
	src := reader.Source()
	_ = goldmarkast.Walk(doc, func(n goldmarkast.Node, entering bool) (goldmarkast.WalkStatus, error) {
		h, ok := n.(*goldmarkast.Heading)
		if !entering || !ok {
			return goldmarkast.WalkContinue, nil
		}
		if m := idToken.Find(h.Text(src)); m != nil {
			h.SetAttribute([]byte("id"), []byte(strings.ToLower(string(m))))
		}
		return goldmarkast.WalkContinue, nil
	})
}

var markdown = map[string]goldmark.Markdown{
	"docs": newMarkdown("docs"),
	"spec": newMarkdown("spec"),
}

func newMarkdown(base string) goldmark.Markdown {
	return goldmark.New(
		goldmark.WithExtensions(extension.Table, extension.Strikethrough, extension.Linkify),
		goldmark.WithParserOptions(
			// Heading ids, so a provenance citation can point at the
			// section that answers it rather than at the whole document
			// and a hope (P-069).
			goldmarkparser.WithAutoHeadingID(),
			goldmarkparser.WithASTTransformers(
				goldmarkutil.Prioritized(repoLinks{base}, 100),
				goldmarkutil.Prioritized(idAnchors{}, 200))),
		goldmark.WithRendererOptions(goldmarkhtml.WithHardWraps()),
	)
}

// added lists every category BirdView assigned itself, in one place.
//
// The detail pages already say it bird by bird, which answers "why is this
// robin a frugivore" and not "what has this project decided on its own
// account". Those are different questions, and the second is the one
// somebody auditing the project would ask. Published assignments are
// excluded: there are 1,678 of them and they are AVONET's, not news.
func (s *server) added(w http.ResponseWriter, r *http.Request) {
	type row struct{ kind, species, value, origin, why string }
	var rows []row
	for i := range s.all.Species {
		sp := &s.all.Species[i]
		for _, a := range sp.Habitats {
			if a.Origin != dataset.Published {
				rows = append(rows, row{"Habitat", sp.Name(dataset.LocaleNA),
					trait.LabelOf(trait.Habitats, a.Value), string(a.Origin), a.Why})
			}
		}
		for _, a := range sp.Niches {
			if a.Origin != dataset.Published {
				rows = append(rows, row{"Diet", sp.Name(dataset.LocaleNA),
					trait.LabelOf(trait.Niches, a.Value), string(a.Origin), a.Why})
			}
		}
	}
	sort.Slice(rows, func(i, j int) bool {
		if rows[i].kind != rows[j].kind {
			return rows[i].kind > rows[j].kind // Habitat before Diet
		}
		if rows[i].origin != rows[j].origin {
			return rows[i].origin < rows[j].origin
		}
		return rows[i].species < rows[j].species
	})

	var b bytes.Buffer
	b.WriteString("# What BirdView adds to its sources\n\n")
	fmt.Fprintf(&b, "Every category on this site comes from a published source except the %d below. "+
		"These are BirdView's own, listed here so they can be read in one place rather than "+
		"found one bird at a time. Each also appears on that bird's own page.\n\n", len(rows))
	b.WriteString("`derived` means BirdView computed it from published data by a stated rule. " +
		"`birdview` means BirdView decided it and no source says it.\n\n")
	b.WriteString("Everything else — habitat, diet, family, order — is AVONET's or AviList's, " +
		"unchanged. The regions each bird is listed in are derived in bulk from GBIF records " +
		"and are described in [the data document](/docs/data) rather than repeated here.\n\n")
	b.WriteString("| | Species | Category | Origin | Why |\n|---|---|---|---|---|\n")
	for _, r := range rows {
		fmt.Fprintf(&b, "| %s | %s | %s | `%s` | %s |\n",
			r.kind, r.species, r.value, r.origin, r.why)
	}
	s.renderMarkdown(w, "docs", "What BirdView adds to its sources", b.Bytes())
}

// excluded lists every bird left off the index, with its reason and its
// records, each linked to its detail page (P-080). A bird a reader has seen
// should not vanish without a word; this is the word.
func (s *server) excluded(w http.ResponseWriter, r *http.Request) {
	var rows []*dataset.Species
	for i := range s.all.Species {
		if sp := &s.all.Species[i]; sp.Exotic != "" {
			rows = append(rows, sp)
		}
	}
	sort.Slice(rows, func(i, j int) bool {
		return rows[i].Name(dataset.LocaleNA) < rows[j].Name(dataset.LocaleNA)
	})
	var b bytes.Buffer
	b.WriteString("# Birds left off the index\n\n")
	fmt.Fprintf(&b, "These %d birds recur in the records BirdView reads, so they pass its scope "+
		"rule, but their records here are escapes, releases or birds kept by people: none has a "+
		"population of its own anywhere BirdView covers. They are left off the index so they do "+
		"not crowd out the birds a reader can expect to meet, and listed here so that leaving "+
		"them off is not a silent edit. Each keeps its own page.\n\n", len(rows))
	b.WriteString("The list is BirdView's own judgement, not a measurement: the records cannot say " +
		"where a bird came from. Long-feral populations, such as Indian Peafowl in Florida, are " +
		"not on it, and neither is any wild vagrant. See [the specification](/spec/product#p-080).\n\n")
	b.WriteString("| Bird | Scientific name | Records | Why it is left off |\n|---|---|---:|---|\n")
	for _, sp := range rows {
		var n int64
		for _, c := range sp.RegionCounts {
			n += c
		}
		for _, c := range sp.BelowCounts {
			n += c
		}
		fmt.Fprintf(&b, "| [%s](/species/%s) | *%s* | %s | %s |\n",
			sp.Name(dataset.LocaleNA), sp.ID, sp.ScientificName, comma(n), sp.Exotic)
	}
	s.renderMarkdown(w, "docs", "Birds left off the index", b.Bytes())
}
