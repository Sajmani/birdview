# Building, running and changing BirdView

BirdView is a single Go module. Two commands do the everyday work: an ingest that compiles the
dataset from published sources, and a server that embeds and serves the web UI. Five generators
build the committed tables under `data/` that the ingest and server read. There is no database,
no data fetch at run time and no JavaScript framework.

---

## Prerequisites

- **Go**, at the version `go.mod` requires. No separate minimum is maintained; the version follows
  what the dependencies need.
- **Chrome or Chromium**, optional. Tests that render pages in a browser skip without it; set
  `BIRDVIEW_CHROME` if it is not in a standard place. A skip is not a pass, so check the test
  output for skip lines when your change affects rendered pages.
- **Python 3, Node and Google Chrome**, only to re-render the pipeline diagram (see below).

The one Go dependency is `goldmark`, for rendering these guides. The source workbooks are read by
BirdView's own `internal/xlsx`, using only the standard library.

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

From the repository root:

```
go run ./cmd/birdview-ingest     # compiles web/data/species.json
go run ./cmd/birdview            # serves http://localhost:8080
```

The ingest prints a coverage summary: species in scope, and how many traits and wingspans are
measured, near-match or modeled. Expect **1082 species in 99 families across 27 orders**.

**The server reads `species.json` once, at startup.** After re-running the ingest, restart the
server, or it keeps serving the old data.

**Use `-dev web` while editing the UI.** Templates and static files are embedded with `go:embed`,
so a normal server serves them as they were when it was built: edit `index.html` or `styles.css`,
reload, and you see the old page with nothing to say it is stale. `-dev web` reads them from
`web/static/` in the working tree and re-parses the templates on every request, so an edit shows
on reload.

```
go run ./cmd/birdview -dev web
```

It is off by default, and `TestEmbedIsTheDefault` checks that. Go code is still compiled in, so
after changing a Go file (for example, adding a field a template uses), restart the server.

Flags:

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
| | `-feedback` | none; the issue tracker URL pages point readers to, omitted when unset |

The defaults are relative paths, so run both commands from the repository root.

---

## The generators

Five commands build committed tables under `data/`. None of them runs as part of the ingest,
because between them a full run is thousands of requests to other people's servers, tests make no
network requests, and the build must work offline. Run a generator by hand when its source
changes, review its output as a diff, and commit the diff.
[How BirdView's data is built](pipeline.md) shows all of them, and the ingest they feed, in one
diagram.

| Command | Writes | Cost |
| --- | --- | --- |
| `birdview-inatmap` | `data/inat-taxa.tsv`: one iNaturalist taxon ID per species | ~20 min, 1 request/second |
| `birdview-gbifmap` | `data/gbif-taxa.tsv`: GBIF backbone name → AviList name | ~5 min, ~2,800 requests |
| `birdview-regions` | `data/region-presence.tsv`: which regions each species occurs in | one facet query per region, resumable |
| `birdview-status` | `data/status.tsv`: BirdView's own occurrence status per species | two passes, resumable |
| `birdview-boundaries` | `data/region-boundaries.json`: region outlines for the Here button, from Natural Earth | one ~15 MB download, hash-checked |

`birdview-inatmap` and `birdview-regions` read the built dataset, so run the ingest first.

The iNaturalist rate of one request per second is what iNaturalist's API terms ask for; do not
lower the interval. The GBIF rate is one BirdView sets for itself, and the client slows down
further whenever GBIF answers 429.

`birdview-inatmap` leaves out a species that does not resolve confidently, and that species' page
then shows no iNaturalist link.

`birdview-status` caches raw fetches under `/tmp` and skips what it already has, so an interrupted
run resumes instead of starting over. It refuses to write a partial table, because a species whose
dates never arrived would otherwise be classified as a vagrant on no evidence.

A sixth command, `birdview-names`, writes `data/source-names.tsv`: how many records eBird and
iNaturalist each hold under each species' name, used to find names the two sources disagree about.
Nothing in the build reads that table; a test does.

**When a command's inputs or outputs change, update the diagram.** It is Mermaid in
`docs/pipeline.md`, which GitHub draws directly. BirdView serves an SVG rendered from it, because
drawing Mermaid in the browser would need a script from the network. After editing the Markdown,
run `python3 docs/render-diagrams.py` and commit the SVG it writes beside the page.
`TestDiagramsAreCurrent` fails until you do, and `TestPipelineDiagramNamesRealColumns` fails if the
diagram names a source column no reader asks for.

---

## Standing checks

Run all four, in this order, before calling a change done:

```
go build ./...
go vet ./...
gofmt -l .        # must print nothing
go test ./...
```

Every spec check is a Go test in `river/check/` and runs with `go test ./...`: River's checker and
its shared cases, BirdView's stricter goal rule, the vendored-file hashes, every quotation against
its vendored source, and every reference, anchor and relative link into the bundle.

These checks cover the bundle's **structure**, not its substance. A green run means the bundle is
internally consistent, not that it is right.

---

## Repository layout

```
cmd/birdview-ingest/     builds species.json from the sources and the data/ tables
cmd/birdview/            the server; embeds web/ and docs/ and serves them
cmd/birdview-inatmap/    generator: iNaturalist taxon IDs
cmd/birdview-gbifmap/    generator: GBIF backbone -> AviList name bridge
cmd/birdview-regions/    generator: per-region presence
cmd/birdview-status/     generator: BirdView's own occurrence status
cmd/birdview-boundaries/ generator: region boundaries for the Here button
cmd/birdview-names/      measures how eBird and iNaturalist name each species
internal/avilist/        the AviList workbook reader: names, families, codes, sequence
internal/avonet/         the trait workbook reader, including habitat and trophic niche
internal/wingspan/       the Fox dataset and the fitted wingspan regression
internal/taxonomy/       the synonym tables, in both directions, and the order labels
internal/trait/          the habitat and diet vocabularies and BirdView's additions
internal/morphology/     the segment models (head, body, neck) and rounding
internal/region/         the regions, their two groupings, and their IDs
internal/boundary/       the region boundary file and the reference lookup over it
internal/month/          the twelve months and the rule for presence in one
internal/gbif/           the paced GBIF client and the name bridge
internal/inat/           the iNaturalist taxon lookup
internal/status/         recurrence: episodes, and the status they imply
internal/view/           ordering, sectioning, diagram geometry; pure, no HTML
internal/dataset/        the serialized types and the provenance kinds
internal/palette/        family and order colors, with contrast guarantees
data/                    the committed generated tables
web/static/              templates, CSS, the Here script and the favicon, embedded
web/data/                the built species.json (gitignored)
docs/                    this documentation, also served at /docs
third_party/             vendored sources, each under its own license
river/                   the requirements bundle; normative, not served
river/check/             the spec checks, as Go tests
```

`internal/view` has no HTML or HTTP so the layout rules can be property-tested directly. If you
want to format a string there, the code belongs in another package.

---

## The spec-driven workflow

**The requirements are normative and live in [`river/`](../river).** Code implements them; it does
not define them. Where this guide and `river/` disagree, `river/` wins.

[`river/README.md`](../river/README.md) describes the method, River, and is kept in the repository
so the project does not depend on an external document. You do not need to read all of it to make
a change, but you do need the loop below.

### The bundle

| File | Contents |
| --- | --- |
| [`README.md`](../river/README.md) | the method, River, copied in unchanged |
| [`bindings.md`](../river/bindings.md) | how BirdView applies the method, and where it departs from it |
| [`goals/`](../river/goals/goals.md) | why the project exists: fourteen goals, `goals.*` |
| [`requirements/sources/product/`](../river/requirements/sources/product/product.md) | what it must do, one file per group: `product.*` |
| [`requirements/sources/tech/`](../river/requirements/sources/tech/tech.md) | technical constraints, one file per group: `tech.*` |
| [`requirements/app/`](../river/requirements/app/app.md) | every requirement the web application must meet, and their criteria |
| [`requirements/requirements.md`](../river/requirements/requirements.md) | every body of rules or data the project is subject to |
| [`requirements/decisions.md`](../river/requirements/decisions.md) | how each conflict was resolved: `decisions.*` |
| [`requirements/acceptance.md`](../river/requirements/acceptance.md) | how each remaining requirement is checked: `acceptance.*` |
| [`todo.md`](../river/todo.md) | everything still open: work, questions, unverified criteria, stale text |
| [`arch.md`](../river/arch.md) | how the code is organized |
| [`check/`](../river/check) | the spec checks |

Each requirement has an ID such as `product.scaleToggle`. A requirement's first paragraph is the
rule and the rest is rationale. Cite IDs in code comments using the literal token, so one search
finds every use:

```go
// Verifies: product.scaleToggle, tech.badgeContrast.
```

### The loop

A change goes requirement first, then acceptance criterion, then code. An agent or contributor may
draft requirements and criteria; the project owner approves them. Not every change needs all
three steps:

| Change | What it needs |
| --- | --- |
| New feature or changed behavior | Edit the requirement, then the criterion, then the code |
| Bug fix | If a requirement already forbids the bug, add a criterion under it, then fix. If none does, that is a gap in the spec: add the requirement first |
| Refactor | No spec change, and the suite stays green. If you must change a check's *assertions*, it is not a refactor |
| Adopting a new source | An entry in `requirements.md`, then a pass to compose its requirements and resolve conflicts with the existing ones. Expect conflicts |
| Docs only | No spec change, unless writing the doc shows the spec is wrong |

Open work is tracked in [`todo.md`](../river/todo.md), not in `TODO` comments. Read it before
starting something new, and add an item for anything you find or leave unfinished.

### Three rules

**Never resolve a conflict in code.** If two requirements disagree, or a change reverses something
the spec states, stop and record it as a new entry in
[`decisions.md`](../river/requirements/decisions.md). Choosing one silently is the failure this
process exists to prevent.

**Never edit a requirement to match code that violates it** without saying so and getting
agreement. That turns a bug into a feature without anyone deciding it should be one.

**A criterion you have not watched fail is not a criterion.** Break the code on purpose, confirm the
check fails, and read the failure message to confirm it fails for the right reason. Then record
how you broke it in the criterion's Status line. Make the deliberate break fail loudly if it does
not apply, or a no-op edit will look like a passing check. Checks that pass while testing nothing
are common: a whole-page string search that matches the help text instead of the control, a
fixture too small to exercise the behavior, a test that still passes after its requirement is
reversed. They look fine in review.

---

## Hard rules

Most of these are licensing rules; breaking them is not a matter of style.

- **Never commit the built dataset.** It is reproducible from the committed inputs, and a stale
  copy in the tree is worse than none.
- **Never read or process Wiley Online Library material**, including Ricklefs (2017). Its terms
  prohibit AI tools from ingesting it and prohibit automated access, and the source is
  [rejected](../river/requirements/decisions.md#ricklefsprohibited). Do not try to revive it as a
  length source.
- **Never fetch, download or programmatically process anything from BirdLife's DataZone**
  (`datazone.birdlife.org`). Its terms forbid automated access and text and data mining; see
  [the register entry](../river/requirements/requirements.md#birdlife-checklist). AVONET's BirdLife
  sheet is part of AVONET, under CC BY 4.0, and is not affected.
- **Only AVONET, the Fox wingspan data and AviList may be committed as source data.** All three
  are CC BY 4.0. Before adding a fourth, read
  [`requirements.md`](../river/requirements/requirements.md): most candidates this project evaluated
  could not be redistributed. The generated tables under `data/` are not source data; they are the
  project's own reductions and carry their provenance in their headers. The vendored terms and
  WCAG text are kept for quotation, not as data.
- **Every third-party file goes under `third_party/<name>/`**, with its own license and README,
  and its provenance under `river/requirements/sources/<name>/`.
- **Never present a modeled value as a measurement.** Every length and wingspan is marked measured
  or modeled, and the page shows which. The whole measurement design rests on
  [this decision](../river/requirements/decisions.md#noopenlengthandwingspan).

`AGENTS.md` carries the same rules in the form AI coding agents read.

---

## Deployment

Every file the build reads is in the repository, so deploying needs only the container image
below, or the server binary and a built dataset.

### Container image

The `Dockerfile` at the repository root builds in three stages and produces a minimal static image
based on `gcr.io/distroless/static-debian12:nonroot`:

1. **`ingest`** builds and runs `birdview-ingest`, producing `/out/species.json`.
2. **`builder`** builds the static `birdview` server binary.
3. **The final image** copies in only the binary and the dataset, as `/data/species.json`.

```
docker build -t birdview .
docker run --rm -p 8080:8080 birdview
```

By default the container runs `birdview -addr 0.0.0.0:8080 -data /data/species.json`. Arguments
after the image name replace those defaults entirely, so to add a flag such as `-feedback`, repeat
`-addr` and `-data` as well:

```
docker run --rm -p 8080:8080 birdview \
  -addr 0.0.0.0:8080 \
  -data /data/species.json \
  -feedback <issue-tracker-url>
```

Without `-feedback`, pages say nothing about feedback.

### Running without Docker

The server binds `localhost:8080` by default. On a host, build the binary, build the dataset, and
point `-addr` at the interface and port the host expects:

```
go build -o birdview ./cmd/birdview
go run ./cmd/birdview-ingest
./birdview -addr 0.0.0.0:$PORT -data web/data/species.json
```

The templates, static files and these guides are embedded in the binary; the dataset is the only
file it reads at run time.

When a regenerated dataset drops a species, `/species/<code>` answers 404 with a page naming the
bird and why it is no longer drawn, so old links still explain themselves.

---

## Contact

Through the project's issue tracker.
