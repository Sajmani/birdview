# Architecture

How the code is actually put together. Descriptive: where this and the code disagree, this file
is wrong.

## Shape

Two commands over a shared set of internal packages.

```
cmd/birdview-ingest   compiles sources into web/data/species.json (a build artifact)
cmd/birdview          serves it on localhost, rendering HTML server-side
cmd/birdview-inatmap  builds data/inat-taxa.tsv; run by hand, never by the build
cmd/birdview-gbifmap  builds data/gbif-taxa.tsv; likewise
internal/aba          parses the ABA checklist; forward-fills family from section headers
internal/avonet       reads traits from all three AVONET taxonomy sheets + the crosswalk
internal/wingspan     empirical wingspans, and the regression fitted from them
internal/taxonomy     the name mapping: renames vs near-matches
internal/gbif         the GBIF -> AviList name bridge, and the paced API client
internal/morphology   the segment models, and the rounding that keeps them summing
internal/palette      generated family colors under the contrast rule
internal/dataset      the record and the provenance vocabulary
internal/view         ordering, sectioning, scale and diagram geometry
web/                  embedded templates and CSS
```

## Why rendering is server-side

The ordering, sectioning and geometry rules carry acceptance criteria, and property tests over
pure Go functions are the cheapest checker that would actually catch a violation. Putting those
rules in JavaScript would have meant either duplicating them in Go for the tests, or testing them
through a browser harness the project does not have.

So the server computes sections and geometry and renders HTML. Every control is a query parameter
and every detail view is a URL, which also makes any view shareable. The cost is a page load per
interaction, which is unnoticeable on localhost.

**One exception, added deliberately.** A page load per interaction means the browser lands at the
top each time, discarding your place in a list of 841 birds. A dozen inline lines save the scroll
position on leaving the index and restore it on return. It is the only script in the project and
holds no application logic — the rule it protects is that *logic* stays in Go, not that the client
is inert.

This is what let AC-056 and AC-057 be real checks rather than declared gaps.

## Data flow

```
ABA checklist ──┐
                ├─► resolve each species to a trait row ──► segments ──► species.json ──► server
AVONET ─────────┘        (eBird, then BirdLife, then           │
                          BirdTree, then a near-match)         │
Fox wingspans ─────────► empirical value, else the fitted ─────┘
                         order-level regression
```

Resolution order is the substance of CR-006: prefer any row that is genuinely *this* taxon —
whatever name or sheet it is filed under — over borrowing a relative's measurements. Only when no
such row exists anywhere does a near-match apply, and the label then says so.

## The invariant that shapes the code

Segments must sum **exactly** to total length, because the drawn height of a diagram is what makes
two birds comparable. `morphology.BalanceToSum` rounds the total and four segments, then derives
the trunk as the remainder, so the rounded parts always reconcile. Rounding each part
independently leaves up to 0.2 mm adrift; the previous design did that and broke the invariant for
27.6% of its species.

## Where the scale decisions live

`view.Sections` groups; `view.Scales` picks one reference length per group — the section's largest
bird, or the whole set's in global mode. `view.Layout` turns a reference into pixels. Keeping the
reference separate from the pixel conversion is what lets AC-031 assert that every bird in a
section shares a scale factor, which is the property P-042 actually requires.

The server also computes a *width* reference per group, because a to-scale crossbar on a
wide-winged bird in a section of short ones would otherwise overflow. Both references come from
the same group, so one scale still serves the whole section.

## The two mapping tools, and why they are not in the ingest

`birdview-inatmap` and `birdview-gbifmap` both resolve BirdView's species against somebody else's
taxonomy over the network, and both write a committed table rather than running at build time
(T-022). The reasons are the same in each case: tests make no network request, a build that needs
the network fails on a plane, and a quarter of an hour of someone else's server time does not
belong in a `go build`. The tables go stale; a stale row fails visibly and the refresh is one
command.

They differ in what a miss means. iNaturalist supplies a *link*, so a miss costs a link and
nothing else, and `inat.Table` simply omits the species. GBIF supplies a *name correspondence*,
where the set of things that did not resolve is the interesting output — so `gbif.Table` carries
every species GBIF reported, with a tier saying how it resolved or why it did not.

`gbif.Resolve` tries three tiers and refuses ambiguity at each: the accepted name, then the
synonyms GBIF files under it, then the epithet within GBIF's family after folding gender agreement
and Latin case. The last is the only fuzzy one, it contributes three rows out of 1,378, and it is
scoped and uniqueness-checked because AviList holds 369 (family, folded epithet) buckets with more
than one species in them.

**Nothing joins against `data/gbif-taxa.tsv` yet.** The bridge exists so that a later decision
about where species scope comes from can be made on measured ground; making that decision is not
this code's job.

## Known rough edges

- **The neck ratio table covers 22 families**; everything else takes the 0.18 default. The
  provenance string says which applied, so this is visible rather than hidden, but the table is
  thin for a 99-family dataset.
- **`headAspect` is an unexamined drawing constant.** Unlike the neck ratio it produces no metric,
  so it carries no provenance — but it is just as invented.
- **The detail view re-renders the tile at 1.9x** rather than sharing a sized component. If the
  tile layout changes, both places need it.
- **Hummingbirds and gamebirds get the global wingspan fit**, not an order-level one, because
  neither order has five empirical species — Galliformes has none at all. Their estimates carry
  roughly 7.3% median error against 4.6% for well-covered orders.
- **The ingest re-reads the 21 MB AVONET workbook three times**, once per taxonomy sheet. It takes
  about four seconds, which has not been worth optimizing.
- **A resolved row in the GBIF bridge is name-correct and may still be concept-wrong.** Where GBIF
  has not followed a split, its name maps to whichever half kept the name: GBIF's `Circus cyaneus`
  maps to the Eurasian Hen Harrier while 3.8 million of those records are Northern Harriers. Nine
  such lumps are known and are listed in [sources.md](sources.md#gbif); the table cannot express
  them, because it has one AviList column and a lump needs two. Nothing detects them
  automatically — they were found by hand, one query per unreached species.
