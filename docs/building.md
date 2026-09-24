# Building, running and changing BirdView

A single Go module with two commands: an ingest that compiles the dataset from published sources,
and a server that embeds and serves the web UI. No database, no runtime data fetch, no JavaScript
framework.

---

## Prerequisites

- **Go.** No minimum version is maintained — whatever the dependencies require
  ([T-007](../spec/tech.md)). Developed on 1.25.
- **Python 3**, for the three spec-level check scripts.

That is all. The one dependency is `excelize`, for reading the source workbooks.

---

## Getting the data

Three inputs. Two are in the repo; one you must fetch yourself.

### Already here (CC BY 4.0, redistributable)

- `AVONET Supplementary dataset 1.xlsx` — morphology traits
- `wingspan_fox2026_AppendixS1.xlsx` — empirical wingspans

Both are vendored with provenance and hashes under [`spec/sources/`](../spec/sources).

### You must fetch (all rights reserved, never committed)

The **ABA Checklist 8.19** supplies the species list, common names and families.

```
mkdir -p fetched
curl -A 'your-name/1.0 (contact: you@example.com)' \
  -o fetched/ABA_Checklist-8.19.csv.zip \
  https://www.aba.org/wp-content/uploads/2026/03/ABA_Checklist-8.19.csv.zip
unzip -o -d fetched fetched/ABA_Checklist-8.19.csv.zip
```

Send a descriptive `User-Agent` naming who you are ([T-006](../spec/tech.md)) — a default
`Go-http-client/1.1` earns a 403 from some hosts and is discourteous everywhere.

> **This file is copyrighted and must never be committed or redistributed.** Its own first row says
> so. `/fetched/` is gitignored and that is not a bug to fix. If you plan to serve BirdView to
> anyone but yourself, read [Deployment](#deployment) first — it is genuinely blocked.

---

## Build and run

```
go run ./cmd/birdview-ingest     # compiles web/data/species.json
go run ./cmd/birdview            # serves http://localhost:8080
```

The ingest prints a coverage summary — species in scope, how many traits and wingspans are
measured, near-match or modeled. Expect **841 species in 91 families across 25 orders**.

Flags, should you need them:

| Command | Flag | Default |
| --- | --- | --- |
| `birdview-ingest` | `-aba` | `fetched/ABA_Checklist-8.19.csv` |
| | `-avonet` | `AVONET Supplementary dataset 1.xlsx` |
| | `-wingspan` | `wingspan_fox2026_AppendixS1.xlsx` |
| | `-out` | `web/data/species.json` |
| | `-inat` | `data/inat-taxa.tsv` |
| `birdview` | `-addr` | `localhost:8080` |
| | `-data` | `web/data/species.json` |

> **Gotcha that has bitten twice:** the server reads `species.json` **once, at startup**. After
> re-running the ingest you must restart the server, or you will spend a while debugging stale
> data that is already fixed on disk.

The built dataset is a build artifact and is gitignored ([T-005](../spec/tech.md)) — it carries
values derived from the ABA checklist.

---

## Refreshing the iNaturalist mapping

`data/inat-taxa.tsv` maps each species to an iNaturalist taxon ID, and is **committed**. The
ingest joins against it offline; nothing fetches it at build or run time (T-022).

```
go run ./cmd/birdview-inatmap        # ~15 minutes: 841 species at 1 request/second
```

The pace is what iNaturalist asks for (`inat-api/R1`) and is not negotiable downward. Run it when
the taxonomy moves — annually, alongside the source refresh — read the diff, and commit it. A
species that does not resolve confidently is left out on purpose, and its detail page then shows no
link.

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
cmd/birdview-ingest/   builds species.json from the three sources
cmd/birdview/          the server; embeds web/ and serves it
internal/aba/          the checklist parser (scope filter, names, families)
internal/avonet/       the trait workbook reader
internal/wingspan/     the Fox dataset + the fitted wingspan regression
internal/taxonomy/     ABA <-> BirdLife synonym table
internal/morphology/   the segment models (head, body, neck) and rounding
internal/view/         ordering, sectioning, diagram geometry — pure, no HTML
internal/dataset/      the serialized types and the provenance kinds
internal/palette/      family and order colors, with contrast guarantees
web/static/            templates and CSS, embedded into the binary
spec/                  the requirements bundle — normative
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
| [`decisions.md`](../spec/decisions.md) | conflict resolutions — `CR-###` — and Work arising |
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

- **Never commit anything under `fetched/`.** The ABA checklist is all-rights-reserved.
- **Never commit the built dataset.** It carries values derived from that checklist.
- **Never read or process Wiley Online Library material**, including Ricklefs (2017). Its terms
  prohibit AI tools ingesting it and prohibit automated access, and the source is formally rejected
  ([CR-002](../spec/decisions.md)). Do not try to revive it as a length source.
- **Only AVONET and the Fox wingspan data may be committed as inputs.** Both are CC BY 4.0. Before
  adding a third data file, read [`sources.md`](../spec/sources.md) — most candidates this project
  evaluated could not be redistributed.
- **Never present a modeled value as a measurement** ([P-060](../spec/product.md), P-061). This is
  the decision the whole measurement design rests on ([CR-001](../spec/decisions.md)).

`AGENTS.md` carries the same rules in the form AI coding agents read.

---

## Deployment

**BirdView is not deployed publicly, and cannot be yet.**

Serving it reproduces the ABA checklist to every visitor, and the checklist is all-rights-reserved
with no license permitting that. A written permission request to the ABA is drafted and unsent; the
app stays local until it is granted. See the Work arising table in
[`decisions.md`](../spec/decisions.md).

Nothing else blocks it. There is no deployment tooling yet — no container image, and the server
binds `localhost` by default, so a real deployment needs `-addr` pointed at `0.0.0.0:$PORT`.

---

## Contact

**birdview at ajmani dot net**
