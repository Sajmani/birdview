# Building, running and changing BirdView

A single Go module. Two commands do the everyday work — an ingest that compiles the dataset from
published sources, and a server that embeds and serves the web UI — and five generators build
the committed lookup tables those two read. No database, no runtime data fetch, no JavaScript
framework.

---

## Prerequisites

- **Go.** No minimum version is maintained — whatever the dependencies require
  ([tech.goVersionFollowsDeps](../river/requirements/sources/tech/delivery.md#goversionfollowsdeps)). Developed on 1.25.
- **Chrome or Chromium**, optional: the tests that render pages in a browser skip without it
  (set `BIRDVIEW_CHROME` if it is not in a standard place).

That is all. The dependencies are `excelize`, for reading the source workbooks, and `goldmark`, for
rendering these guides.

---

## The data and its licenses

Everything the build needs is in the repository; nothing is fetched to build or to serve.

| Source | License | What is in this repository | Used for |
| --- | --- | --- | --- |
| [AviList](https://doi.org/10.2173/avilist.v2025b) v2025b | CC BY 4.0 | the workbook, in [`third_party/avilist/`](../third_party/avilist) | names, families, taxonomic sequence, species codes |
| [AVONET](https://doi.org/10.6084/m9.figshare.16586228.v7) (Tobias et al. 2022) | CC BY 4.0 | the workbook, in [`third_party/avonet/`](../third_party/avonet) | bill, wing, tail, mass, habitat, diet |
| [Fox et al. 2026 wingspan database](https://doi.org/10.6084/m9.figshare.32089083) | CC BY 4.0 | the workbook, in [`third_party/foxwingspan/`](../third_party/foxwingspan) | measured wingspans |
| [GBIF](https://www.gbif.org): the [eBird Observation Dataset](https://doi.org/10.15468/aomfnb) and iNaturalist research-grade observations | CC BY 4.0 | only tables BirdView derives, under `data/`; never a record | where each bird recurs, the regions, how common it is |
| [iNaturalist](https://www.inaturalist.org) | taxon IDs are facts; the terms are © iNaturalist | one taxon ID per species in `data/inat-taxa.tsv`, and the terms in [`third_party/inatterms/`](../third_party/inatterms) and [`third_party/inatapi/`](../third_party/inatapi) | the link from a species page to its photographs |
| [Natural Earth](https://www.naturalearthdata.com/) 1:10m states and provinces | public domain | the derived `data/region-boundaries.json`; the source is fetched by hash only to regenerate it | the Here button |
| [WCAG 2.2](https://www.w3.org/TR/2024/REC-WCAG22-20241212/) | W3C Document License | the Recommendation, in [`third_party/wcag/`](../third_party/wcag) | the accessibility rules BirdView meets, quoted |

**Nothing under `third_party/` is covered by BirdView's Apache 2.0 license.** Each directory carries
its own license and a README naming the source and how to cite it. The iNaturalist terms are
openly licensed by no one; they and the WCAG text are kept so every quotation of them in
[`river/`](../river) can be checked against the exact words. Retrieval details and hashes are
recorded once, under [`river/requirements/sources/`](../river/requirements/sources), and the
[source register](../river/requirements/requirements.md) records every source considered, including
the ones rejected and why.

**BirdView links to iNaturalist and copies nothing from it**: no photograph is embedded, cached or
hotlinked, so no contributor's license is engaged.

**Every CC BY source is credited on the page that uses it**, as its license asks: the index footer
names each with its citation.

**The built dataset, `web/data/species.json`, is not committed.** It is reproducible from the
committed inputs in one command, and a stale copy in the tree would be worse than none.

---

## Build and run

```
go run ./cmd/birdview-ingest     # compiles web/data/species.json
go run ./cmd/birdview            # serves http://localhost:8080
```

**Use `-dev web` while editing the UI.** Templates and static files are embedded with `go:embed`,
so a running server serves them as they were when it *started*: edit `index.html` or `styles.css`,
reload, and you see the old page with nothing to say it is stale. `-dev` reads both from the
working tree and re-parses the templates per request, so an edit shows on reload.

```
go run ./cmd/birdview -dev web
```

It is off by default and a deployed binary never uses it; `TestEmbedIsTheDefault` asserts that.
This flag exists because the embed caught three wrong observations in one afternoon — a
recalibrated dataset, a recoloured favicon and a footer credit, each reported from a page that no
longer matched the repository.

The ingest prints a coverage summary — species in scope, how many traits and wingspans are
measured, near-match or modeled. Expect **1082 species in 99 families across 27 orders**.

Flags, should you need them:

| Command | Flag | Default |
| --- | --- | --- |
| `birdview-ingest` | `-status` | `data/status.tsv` |
| | `-avilist` | `third_party/avilist/AviList-v2025b-10Jun2026-extended.xlsx` |
| | `-avonet` | `third_party/avonet/AVONET Supplementary dataset 1.xlsx` |
| | `-wingspan` | `third_party/foxwingspan/wingspan_fox2026_AppendixS1.xlsx` |
| | `-inat` | `data/inat-taxa.tsv` |
| | `-regions` | `data/region-presence.tsv` |
| | `-out` | `web/data/species.json` |
| `birdview` | `-addr` | `localhost:8080` |
| | `-data` | `web/data/species.json` |
| | `-dev` | off |

> **Gotcha that has bitten twice:** the server reads `species.json` **once, at startup**. After
> re-running the ingest you must restart the server, or you will spend a while debugging stale
> data that is already fixed on disk.

The built dataset is a build artifact and is gitignored ([tech.datasetNeverCommitted](../river/requirements/sources/tech/delivery.md#datasetnevercommitted)) — not for
licensing reasons any more, but because it is reproducible from the committed inputs in one
command and a stale copy in the tree is worse than none.

---

## The generators, and why they are separate commands

Four commands build committed lookup tables under `data/`. None of them runs as part of the
ingest, and that is deliberate (tech.pacedInaturalistMapping): between them a full run is thousands of requests to
someone else's server, **tests make no network request**, and a build that needs the network is
a build that fails on a plane. Each is run by hand when its source moves, its output is read as
a diff, and the diff is committed.
[How BirdView's data is built](pipeline.md) draws all of them, and the ingest they feed, as one
diagram.

| Command | Writes | Cost |
| --- | --- | --- |
| `birdview-inatmap` | `data/inat-taxa.tsv` — one iNaturalist taxon ID per species | ~15 min, 1 request/second |
| `birdview-gbifmap` | `data/gbif-taxa.tsv` — GBIF backbone name → AviList name | ~5 min, ~2,800 requests |
| `birdview-regions` | `data/region-presence.tsv` — which regions each species occurs in | one facet per region |
| `birdview-status` | `data/status.tsv` — BirdView's own occurrence status per species | two passes, resumable |
| `birdview-boundaries` | `data/region-boundaries.json` — region outlines for the Here button, from Natural Earth | one 15 MB download, hash-checked |

The iNaturalist pace is what that API asks for (`inatapi.RequestRate`) and is not negotiable downward.
The GBIF pace is a courtesy BirdView sets for itself, and the client widens it further whenever
GBIF answers 429 — a run that ignores the throttling spends its time in backoff rather than
working.

`birdview-status` caches raw fetches and skips what it already has, so an interrupted run
resumes instead of starting over. It refuses to write a partial table: a species whose dates
never arrived would otherwise be filed as a vagrant on no evidence.

A species that does not resolve confidently is left out on purpose, and its detail page then
shows no iNaturalist link.

**When a command's inputs or outputs change, update that diagram.** It is Mermaid in
`docs/pipeline.md`, which GitHub draws directly; BirdView serves an SVG rendered from it, because
drawing Mermaid in the browser would need a script from the network (tech.selfContainedServer). After editing the
Markdown, run `python3 docs/render-diagrams.py` (it needs Node and Google Chrome) and commit the
SVG it writes beside the page. `TestDiagramsAreCurrent` fails until you do, and
`TestPipelineDiagramNamesRealColumns` fails if the diagram names a source column no reader asks
for.

## Standing checks

All four, in this order, before any change is called done:

```
go build ./...
go vet ./...
gofmt -l .        # must print nothing
go test ./...
```

Every spec-level check is a Go test in `river/check/` and runs with the tests above: River's
checker and its shared cases, BirdView's stricter goal rule, the vendored-file hashes, and every
reference and anchor into the bundle. There is no separate command to remember.

These check **structure**, not substance. A green run means the bundle is internally consistent,
not that it is right.

---

## Repository layout

```
cmd/birdview-ingest/   builds species.json from the sources and the data/ tables
cmd/birdview/          the server; embeds web/ and serves it
cmd/birdview-inatmap/  generator: iNaturalist taxon IDs
cmd/birdview-gbifmap/  generator: GBIF backbone -> AviList name bridge
cmd/birdview-regions/  generator: per-region presence
cmd/birdview-status/   generator: BirdView's own occurrence status
cmd/birdview-boundaries/ generator: region boundaries for the Here button
internal/avilist/      the AviList workbook reader — names, families, codes, sequence
internal/avonet/       the trait workbook reader, including habitat and trophic niche
internal/wingspan/     the Fox dataset + the fitted wingspan regression
internal/taxonomy/     the synonym tables, in both directions, and the order labels
internal/trait/        the habitat and diet vocabularies and BirdView's additions
internal/morphology/   the segment models (head, body, neck) and rounding
internal/region/       the regions, their two groupings, and their ids
internal/boundary/     the region boundary file and the Go reference lookup over it
internal/gbif/         the paced GBIF client and the name bridge
internal/inat/         the iNaturalist taxon lookup
internal/status/       recurrence: episodes, and the status they imply
internal/view/         ordering, sectioning, diagram geometry — pure, no HTML
internal/dataset/      the serialized types and the provenance kinds
internal/palette/      family and order colors, with contrast guarantees
data/                  the committed generated tables the ingest reads
web/static/            templates and CSS, embedded into the binary
docs/                  this documentation, also served at /docs
river/                  the requirements bundle — normative; not served, read it here
```

`internal/view` is deliberately free of HTML and HTTP so the layout rules can be property-tested
directly. If you find yourself wanting to format a string in there, you are in the wrong package.

---

## The spec-driven workflow

**The requirements are normative and live in [`river/`](../river).** Code implements them; it does
not define them. Where this doc and `river/` disagree, `river/` wins.

[`river/README.md`](../river/README.md) is the method itself, copied into the repo so the project
does not depend on an external document. It is long. You do not need to read all of it to make a
change — but you do need the loop.

### The bundle

| File | Answers |
| --- | --- |
| [`README.md`](../river/README.md) | the method — River, copied in unchanged |
| [`goals/`](../river/goals/goals.md) | why the project exists — six goals, `goals.*` |
| [`requirements/sources/product/`](../river/requirements/sources/product/product.md) | what it must do, one file per group — `product.*` |
| [`requirements/sources/tech/`](../river/requirements/sources/tech/tech.md) | technical constraints, one file per group — `tech.*` |
| [`requirements/requirements.md`](../river/requirements/requirements.md) | every body of rules or data the project is subject to |
| [`requirements/decisions.md`](../river/requirements/decisions.md) | conflict resolutions — `decisions.*` |
| [`requirements/acceptance.md`](../river/requirements/acceptance.md) | how each requirement is checked — `acceptance.*` |
| [`todo.md`](../river/todo.md) | everything still open — work, questions, unverified criteria, stale text |
| [`arch.md`](../river/arch.md) | how the code is actually organized |

Cite IDs in code comments using the literal token, so one search covers the tree:

```go
// Verifies: product.scaleToggle, tech.badgeContrast.
```

### The loop

```
context2spec  →  GATE 1: human approves the requirements
spec2test     →  GATE 2: human approves the criteria
spec2code     →  code, tests, arch.md
```

Both gates have been passed. Ongoing changes re-enter at the top, but not every change needs all
three phases — classify first:

| Change | What it needs |
| --- | --- |
| New feature or changed behavior | Full loop: spec edit, then criterion, then code |
| Bug fix | The spec already forbids it. Add a criterion under the existing requirement, then fix. No requirement covers it? That is a spec gap — add the requirement first |
| Refactor | No spec change. The suite stays green. If you must edit a check's *assertions*, it is not a refactor |
| Adopting a new source | A `requirements.md` entry, then a full compose-and-resolve pass. Expect conflicts |
| Docs only | No spec change — unless writing the doc reveals the spec was wrong |

### Three rules that carry most of the weight

**Never resolve a conflict in code.** If two requirements disagree, or a change reverses something
the spec wrote down, stop and record it in [`decisions.md`](../river/requirements/decisions.md) as a new entry.
Picking one silently is the failure this process exists to prevent. Every interesting decision in
this project — rejecting Ricklefs, surfacing the neck model, fitting the wingspan regression,
deleting the legs — is a CR, and each one is more useful than the diff it produced.

**Never edit a requirement to match code that violates it** without saying so and getting
agreement. That converts a bug into a feature by fiat.

**A criterion you have not watched fail is not a criterion.** Deliberately break the code, confirm
the check goes red, then record *how you broke it* in the criterion's Status line. This has caught
around nine checks in this repo that passed while testing nothing — whole-page string searches that
matched the help text instead of the control, a fixture too small to exercise the behavior under
test, a test that survived the reversal of its own requirement, and a scale check whose identical
twins masked the very bug it was written for. Every one looked fine in review.

---

## Hard rules

Violating any of these is a licensing problem, not a style problem.

- **Never commit the built dataset.** It is reproducible from the committed inputs, and a stale
  copy in the tree is worse than none.
- **Never read or process Wiley Online Library material**, including Ricklefs (2017). Its terms
  prohibit AI tools ingesting it and prohibit automated access, and the source is formally rejected
  ([decisions.ricklefsProhibited](../river/requirements/decisions.md#ricklefsprohibited)). Do not try to revive it as a length source.
- **Only AVONET, the Fox wingspan data and AviList may be committed as source data.** All three
  are CC BY 4.0. Before adding a fourth, read [`requirements.md`](../river/requirements/requirements.md) — most candidates
  this project evaluated could not be redistributed. The generated tables under `data/` are not
  source data: they are the project's own reductions, and they carry their provenance in their
  own headers.
- **Never present a modeled value as a measurement** ([product.lengthAndWingspanSources](../river/requirements/sources/product/measurements.md#lengthandwingspansources), product.tileMarksModeled). This is
  the decision the whole measurement design rests on ([decisions.noOpenLengthAndWingspan](../river/requirements/decisions.md#noopenlengthandwingspan)).

`AGENTS.md` carries the same rules in the form AI coding agents read.

---

## Deployment

**BirdView is not yet deployed publicly.**

**The licensing blocker is gone as of 2026-10-05.** Serving BirdView used to reproduce the ABA
checklist to every visitor, and that checklist is all-rights-reserved with no license permitting
it. Permission was requested on 2026-09-27 and never answered, so the project stopped waiting:
names moved to AviList ([decisions.avilistReplacesAba](../river/requirements/decisions.md#avilistreplacesaba)) and scope to BirdView's own
derivation ([decisions.scopeFromRecurrence](../river/requirements/decisions.md#scopefromrecurrence)). Nothing all-rights-reserved is read at build
time, so nothing of anyone else's is served. When a regenerated scope drops a species under a
reader's bookmark, `/species/<code>` answers 404 naming the bird and its measured reason
([decisions.speciesUrlOutlivesSpecies](../river/requirements/decisions.md#speciesurloutlivesspecies)).

See [`river/todo.md`](../river/todo.md) for what remains before a public deployment.

### Container image

A `Dockerfile` at the repository root builds the dataset and the server in separate stages and
packages them into a minimal static image (`gcr.io/distroless/static-debian12:nonroot`):

1. **Stage `ingest`** compiles and runs `birdview-ingest`, producing `/out/species.json`.
2. **Stage `builder`** compiles the static `birdview` server binary.
3. **Final image** copies only the binary and `/data/species.json`.

```
docker build -t birdview .
docker run --rm -p 8080:8080 birdview
```

The server binds `0.0.0.0:8080` inside the container and reads `/data/species.json`. To override
either, or to pass a feedback URL ([product.noPersonalAddress](../river/requirements/sources/product/presentation.md#nopersonaladdress)):

```
docker run --rm -p 8080:8080 birdview \
  -addr 0.0.0.0:8080 \
  -data /data/species.json \
  -feedback https://github.com/google/birdview/issues
```

### Running without Docker

The server binds `localhost:8080` by default ([tech.servesOnLocalhost](../river/requirements/sources/tech/delivery.md#servesonlocalhost)). For a deployment
without a container, point `-addr` at the interface and port the host expects:

```
./birdview -addr 0.0.0.0:$PORT
```

---

## Contact

Through the project's issue tracker.
