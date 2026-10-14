# AGENTS.md

Guidance for AI coding agents working in this repository.

**The requirements are normative and live in [spec/](spec/).** This file summarizes the rules that
prevent damage and points at the rest; where it disagrees with `spec/`, `spec/` wins. Requirement
IDs below (`P-###`, `T-###`) are citations — search `spec/` for the full statement and rationale.

## What this is

BirdView is a tool for birders that shows how big birds actually are, relative to one another,
within and across families. It draws each species as a scaled schematic — bill, head, neck, body,
tail, and a wing crossbar — grouped into sections of similar size. See
[spec/goals.md](spec/goals.md).

It is a single Go module: an ingest command that builds the dataset from published sources, and a
server that embeds and serves the web UI (T-001).

## Status

**Gate 1 and Gate 2 both passed.** The requirements and the acceptance criteria are approved, and
the code is live: `birdview-ingest` builds 841 species across 91 families, and `birdview` serves
them. A behavior change still goes spec first, then the criterion, then the code — see
[spec/process.md](spec/process.md) and "Working on a change" below.

**The licensing blocker is gone as of 2026-10-05.** Scope was the last thing the ABA checklist
supplied, and [CR-028](spec/decisions.md#cr-028) replaced it with BirdView's own derivation from
open occurrence records. `internal/aba` is deleted and nothing reads `fetched/`. What remains
before a public deployment is plumbing, not permission.

**A criterion you have not watched fail is not a criterion.** Break the code deliberately, confirm
each new check goes red, and record in the criterion how you broke it before marking it verified.
This has caught roughly ten checks that passed vacuously — whole-page string searches matching
the help text rather than the control, fixtures too small to exercise the behavior under test, a
test that survived the reversal of its own requirement, one that read a response body before the
header that was wrong, and one that asserted a label written by a different function from the
code under test.

**Assert that the mutation applied.** Twice now a "mutation" was a no-op and the check "passed"
— once because `gofmt` had collapsed the whitespace the edit was anchored on. A substitution
that silently matches nothing proves the opposite of what it looks like. Make the edit fail
loudly when its anchor is missing.

**Watch it fail for the right reason.** Read the failure message, not just the red. A frequency
check failed under its mutation on a secondary assertion while the primary one — the thing the
test was for — passed over the bug untouched; only the message showed it.

## Open work

**Everything outstanding is in [spec/todo.md](spec/todo.md)**: unfinished work, open questions,
known rough edges, criteria not yet watched failing, and spec or doc text that has gone stale.
Read it before starting something new. When you find work, or leave some unfinished, add a row
there, citing the requirement, CR or AC it comes from. Do not leave a `TODO` in code or a "left
open" note in a decision record without a row. Close a row by striking it through and moving it
to *Closed*; never delete one.

## Hard rules

Violating any of these is a licensing problem, not a style problem.

- **Never reintroduce the ABA checklist.** It is all-rights-reserved and says so in the first row
  of its own CSV, and reproducing it to every visitor is what blocked deployment for weeks. Scope
  is now derived ([CR-028](spec/decisions.md#cr-028)); `fetched/` stays gitignored so a stray copy
  cannot creep back in.
- **Never read or process Wiley Online Library material**, including Ricklefs 2017. Its terms
  prohibit AI tools from ingesting it and prohibit automated access to the site, and the source is
  rejected ([CR-002](spec/decisions.md#cr-002)). Do not try to revive it as a length source.
- **Never commit the built dataset** (T-005). It carries values derived from those inputs.
- **Never fetch, download or programmatically process anything from BirdLife's DataZone**
  (`datazone.birdlife.org`), including its checklist archives. Its terms forbid automated access
  and text and data mining (clause 11), and one consultation already breached them; see
  [`birdlife-checklist`](spec/sources.md#birdlife-checklist). A person may read it on screen.
  AVONET's BirdLife sheet is AVONET's, under CC BY 4.0, and is not affected.
- **Only AVONET, the Fox wingspan data and AviList may be committed as source *data*.** All three
  are CC BY 4.0. If you are about to add a fourth data file, check
  [spec/sources.md](spec/sources.md) first — most candidates this project evaluated could not be
  redistributed. Two other committed files are
  deliberate and are not source data: the generated tables under `data/` — `inat-taxa.tsv` of
  iNaturalist taxon IDs (T-022) and `gbif-taxa.tsv` of GBIF-to-AviList name correspondences — and
  the vendored iNaturalist terms.
- **Every third-party file lives under `third_party/<name>/`, with its own license and README.**
  BirdView is Apache 2.0 and none of these files are, so they must not sit where the root
  `LICENSE` reads as covering them. Their provenance and hashes stay in
  `spec/sources/<name>/PROVENANCE.md` — see the location exception in
  [spec/tech.md](spec/tech.md#location-exception-the-third_party-directory). New
  first-party source files carry the Apache 2.0 header; files under `third_party/` never do.
- **BirdView links to iNaturalist and copies nothing from it** (P-064). Never embed, hotlink or
  cache a photograph: those are contributors' work under their own licenses, frequently CC BY-NC.
  T-002 forbids the mechanism anyway. Taxon IDs are fine — the terms say nothing about the
  taxonomy, which is checkable against the hashed copies in `third_party/inat-terms/`.
- **Never present a modeled value as a measurement** (P-060, P-061). Every length and wingspan is
  tagged measured or estimated, and the UI shows which. This is the decision the whole measurement
  design rests on — see [CR-001](spec/decisions.md#cr-001).
- **Never resolve a conflict in code.** If two requirements disagree, stop and take it to
  [spec/decisions.md](spec/decisions.md). Picking one silently is the failure mode the process
  exists to prevent.

## Commands

```
go build ./...
go vet ./...
gofmt -l .            # must print nothing
go test ./...
```

Run all four before declaring work finished (see Project bindings in
[spec/tech.md](spec/tech.md#project-bindings)). The three spec checks —
`python3 spec/check/goals.py`, `spec/check/pins.py`, `spec/check/refs.py` — go with them whenever
`spec/` is touched.

### Leave a server running, every turn

**End every turn with a freshly built server on <http://localhost:8080>**, whether or not
anything changed, so the owner can reload and see the current state without asking. Rebuild
first — the point is that a reload shows *this* turn's work.

```
go build -o /tmp/birdview-serve ./cmd/birdview && pkill -f /tmp/birdview-serve
/tmp/birdview-serve            # background it, FROM THE REPO ROOT
```

Three ways this goes wrong, all already paid for:

- **It must run from the repo root.** `-data` defaults to the relative
  `web/data/species.json`, so started anywhere else it exits with
  `open web/data/species.json: no such file or directory`.
- **`kill %1` does not reap `go run`'s child.** The parent exits, the server keeps the port, and
  the next start fails with `address already in use` while the old process still answers.
  Build to a path and `pkill -f` that path.
- **A stale `-dev` server is worse than no server.** `-dev` reads templates from disk per
  request but the Go binary is frozen at start, so a new template field reports
  `can't evaluate field X` against the old struct — a template bug that is really a stale
  process. Rebuild and restart before believing any dev-mode error.

## Working on a change

Changes follow the loop in [spec/process.md](spec/process.md): spec first, then acceptance
criteria, then code. A behavior change edits [spec/product.md](spec/product.md) in the same commit.

**A change to the data pipeline updates [docs/pipeline.md](docs/pipeline.md) in the same
change** (P-078): the diagram, the step descriptions under it, and the picture re-rendered with
`python3 docs/render-diagrams.py`. Tests catch a stale picture, a column the readers no longer ask
for, and a taxonomy table the ingest reads but the diagram omits; whether a step's description is
still true is for you to check.

Cite requirement IDs in comments using the literal token, so one search covers the tree:

```go
// Verifies: P-042, T-010.
```

## Gotchas

- **The neck/body ratio has no measured basis, and is surfaced anyway** (P-045). It is a
  per-family modeling function, and the neck length *is* serialized and shown in the detail view —
  with `kind: modeled` and a `model` naming the function and that family's ratio. Do not quietly
  drop the model string: naming it is the whole reason showing the number is defensible. See
  [CR-007](spec/decisions.md#cr-007), which superseded the opposite rule in CR-004.
- **Per-tile diagram scaling destroys the product** (P-042). Scaling every bird to fill its tile
  makes a hummingbird and a swan identical on screen, which is exactly what BirdView exists to
  prevent. Scale is per-section or global, never per-tile.
- **The wing crossbar is to scale with the vertical axis** (P-043), which is why tiles are wide.
  A crossbar on its own convenient scale looks authoritative and says something false.
- **AVONET has no common names and no body length or wingspan.** Three separate designs have
  foundered on assuming otherwise. Common names come from AviList (P-004, CR-022); length and
  wingspan are handled by P-060.
- **Scope is recurrence, not abundance** (P-001, CR-028). A species is in when its records fall
  in repeated episodes across years in some region — not when it has enough of them. A record
  count measures observer effort, which under-reports pelagics and island endemics and
  over-reports vagrants, and no threshold on it corrects both at once.
- **Identify the client on outbound requests** (T-006). Go's default `Go-http-client/1.1` earns a
  403 from Wikimedia and is discourteous everywhere else.

## Conventions

**American spellings** in prose, comments, and commit messages — but **never edit a quotation or a
vendored document** to match. Quoted terms of service are verbatim citations with recorded hashes;
"fixing" a publisher's spelling falsifies the quote and breaks the record.

Keep the repository's formatter away from `third_party/`, `spec/sources/` and `spec/process.md`
for the same reason.
