# Building, running and changing BirdView

A single Go module. Two commands do the everyday work — an ingest that compiles the dataset from
published sources, and a server that embeds and serves the web UI — and four generators build
the committed lookup tables those two read. No database, no runtime data fetch, no JavaScript
framework.

---

## Prerequisites

- **Go.** No minimum version is maintained — whatever the dependencies require
  ([T-007](../spec/tech.md)). Developed on 1.25.
- **Python 3**, for the three spec-level check scripts.

That is all. The one dependency is `excelize`, for reading the source workbooks.

---

## Getting the data

Everything the build needs is in the repository.

### Already here (CC BY 4.0, redistributable)

- [`third_party/avilist/`](../third_party/avilist) — names, families and taxonomic sequence
- [`third_party/avonet/`](../third_party/avonet) — morphology traits
- [`third_party/fox-wingspan/`](../third_party/fox-wingspan) — empirical wingspans

Each directory carries the data's own license and a README naming its source and citation. They
are **not** covered by BirdView's Apache 2.0 license. Provenance and hashes are recorded under
[`spec/sources/`](../spec/sources).

### Nothing to fetch

There used to be a step here. The **ABA Checklist** supplied which species were in scope, was
all-rights-reserved, and had to be downloaded by hand into `fetched/` and never committed — and
because serving BirdView reproduced that list to every visitor, it was also what blocked public
deployment.

It is gone. Names, families and sequence come from AviList
([CR-022](../spec/decisions.md#cr-022)) and scope is now BirdView's own, derived from open
occurrence records ([CR-028](../spec/decisions.md#cr-028)) and committed as `data/status.tsv`.
Every input the build reads is either CC BY 4.0 or the project's own work.

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
| | `-wingspan` | `third_party/fox-wingspan/wingspan_fox2026_AppendixS1.xlsx` |
| | `-inat` | `data/inat-taxa.tsv` |
| | `-regions` | `data/region-presence.tsv` |
| | `-out` | `web/data/species.json` |
| `birdview` | `-addr` | `localhost:8080` |
| | `-data` | `web/data/species.json` |
| | `-dev` | off |

> **Gotcha that has bitten twice:** the server reads `species.json` **once, at startup**. After
> re-running the ingest you must restart the server, or you will spend a while debugging stale
> data that is already fixed on disk.

The built dataset is a build artifact and is gitignored ([T-005](../spec/tech.md)) — not for
licensing reasons any more, but because it is reproducible from the committed inputs in one
command and a stale copy in the tree is worse than none.

---

## The generators, and why they are separate commands

Four commands build committed lookup tables under `data/`. None of them runs as part of the
ingest, and that is deliberate (T-022): between them a full run is thousands of requests to
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

The iNaturalist pace is what that API asks for (`inat-api/R1`) and is not negotiable downward.
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
drawing Mermaid in the browser would need a script from the network (T-002). After editing the
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

Plus the spec-level checks, after any edit under `spec/`:

```
python3 spec/check/goals.py    # every requirement cites a goal; every goal is cited
python3 spec/check/pins.py     # vendored sources match their recorded hashes
python3 spec/check/refs.py     # every P-/T-/AC-/CR- cross-reference resolves
```

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
internal/avilist/      the AviList workbook reader — names, families, codes, sequence
internal/avonet/       the trait workbook reader, including habitat and trophic niche
internal/wingspan/     the Fox dataset + the fitted wingspan regression
internal/taxonomy/     the synonym tables, in both directions, and the order labels
internal/trait/        the habitat and diet vocabularies and BirdView's additions
internal/morphology/   the segment models (head, body, neck) and rounding
internal/region/       the regions, their two groupings, and their ids
internal/gbif/         the paced GBIF client and the name bridge
internal/inat/         the iNaturalist taxon lookup
internal/status/       recurrence: episodes, and the status they imply
internal/view/         ordering, sectioning, diagram geometry — pure, no HTML
internal/dataset/      the serialized types and the provenance kinds
internal/palette/      family and order colors, with contrast guarantees
data/                  the committed generated tables the ingest reads
web/static/            templates and CSS, embedded into the binary
docs/                  this documentation, also served at /docs
spec/                  the requirements bundle — normative, also served at /spec
```

`internal/view` is deliberately free of HTML and HTTP so the layout rules can be property-tested
directly. If you find yourself wanting to format a string in there, you are in the wrong package.

---

## The spec-driven workflow

**The requirements are normative and live in [`spec/`](../spec).** Code implements them; it does
not define them. Where this doc and `spec/` disagree, `spec/` wins.

[`spec/process.md`](../spec/process.md) is the method itself, copied into the repo so the project
does not depend on an external document. It is long. You do not need to read all of it to make a
change — but you do need the loop.

### The bundle

| File | Answers |
| --- | --- |
| [`goals.md`](../spec/goals.md) | why the project exists — four goals, G-001…G-004 |
| [`product.md`](../spec/product.md) | what it must do — `P-###` |
| [`tech.md`](../spec/tech.md) | technical constraints — `T-###` |
| [`sources.md`](../spec/sources.md) | every body of rules or data the project is subject to |
| [`decisions.md`](../spec/decisions.md) | conflict resolutions — `CR-###` |
| [`todo.md`](../spec/todo.md) | everything still open — work, questions, unverified criteria, stale text |
| [`acceptance.md`](../spec/acceptance.md) | how each requirement is checked — `AC-###` |
| [`arch.md`](../spec/arch.md) | how the code is actually organized |

Cite IDs in code comments using the literal token, so one search covers the tree:

```go
// Verifies: P-042, T-010.
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
| Adopting a new source | A `sources.md` entry, then a full compose-and-resolve pass. Expect conflicts |
| Docs only | No spec change — unless writing the doc reveals the spec was wrong |

### Three rules that carry most of the weight

**Never resolve a conflict in code.** If two requirements disagree, or a change reverses something
the spec wrote down, stop and record it in [`decisions.md`](../spec/decisions.md) as a `CR-###`.
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
  ([CR-002](../spec/decisions.md)). Do not try to revive it as a length source.
- **Only AVONET, the Fox wingspan data and AviList may be committed as source data.** All three
  are CC BY 4.0. Before adding a fourth, read [`sources.md`](../spec/sources.md) — most candidates
  this project evaluated could not be redistributed. The generated tables under `data/` are not
  source data: they are the project's own reductions, and they carry their provenance in their
  own headers.
- **Never present a modeled value as a measurement** ([P-060](../spec/product.md), P-061). This is
  the decision the whole measurement design rests on ([CR-001](../spec/decisions.md)).

`AGENTS.md` carries the same rules in the form AI coding agents read.

---

## Deployment

**BirdView is not deployed publicly, and cannot be yet.**

**The licensing blocker is gone as of 2026-10-05.** Serving BirdView used to reproduce the ABA
checklist to every visitor, and that checklist is all-rights-reserved with no license permitting
it. Permission was requested on 2026-09-27 and never answered, so the project stopped waiting:
names moved to AviList ([CR-022](../spec/decisions.md#cr-022)) and scope to BirdView's own
derivation ([CR-028](../spec/decisions.md#cr-028)). Nothing all-rights-reserved is read at build
time, so nothing of anyone else's is served.

What remains before a public deployment is ordinary engineering rather than permission: there is
no deployment plumbing, no host, and no answer yet to what happens when the derived scope
changes under a reader's bookmark. See [`todo.md`](../spec/todo.md).

Nothing else blocks it. There is no deployment tooling yet — no container image, and the server
binds `localhost` by default, so a real deployment needs `-addr` pointed at `0.0.0.0:$PORT`.

---

## Contact

Through the project's issue tracker.
