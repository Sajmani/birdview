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

**Gate 1 passed; Gate 2 not started.** The requirements are approved. `spec/acceptance.md` does
not exist yet, and neither does any code. Per [spec/process.md](spec/process.md), criteria come
before implementation — do not start writing the ingest or the server until `acceptance.md` exists
and has been approved.

## Hard rules

Violating any of these is a licensing problem, not a style problem.

- **Never commit anything under `fetched/`** (T-004). The ABA checklist is all-rights-reserved and
  says so in the first row of its own CSV; Ricklefs 2017 is not openly licensed. They are fetched
  onto the building machine, used, and never redistributed. `.gitignore` covers this — do not
  "fix" it.
- **Never commit the built dataset** (T-005). It carries values derived from those inputs.
- **Only AVONET and the Fox wingspan data may be committed** as inputs. Both are CC BY 4.0. If you
  are about to add a third data file to the repository, check [spec/sources.md](spec/sources.md)
  first — most candidates this project evaluated could not be redistributed.
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
[spec/tech.md](spec/tech.md#project-bindings)).

## Working on a change

Changes follow the loop in [spec/process.md](spec/process.md): spec first, then acceptance
criteria, then code. A behavior change edits [spec/product.md](spec/product.md) in the same commit.

Cite requirement IDs in comments using the literal token, so one search covers the tree:

```go
// Verifies: P-042, T-010.
```

## Gotchas

- **The neck/body ratio is invented, deliberately** (P-045). It is a visual design parameter with
  no measured basis, and it is sound *only* while it stays out of the dataset and out of any
  number shown to the user. If you find yourself serializing a neck length, read
  [CR-004](spec/decisions.md#cr-004) — that decision becomes void.
- **Per-card diagram scaling destroys the product** (P-042). Scaling every bird to fill its card
  makes a hummingbird and a swan identical on screen, which is exactly what BirdView exists to
  prevent. Scale is per-section or global, never per-card.
- **The wing crossbar is to scale with the vertical axis** (P-043), which is why cards are wide.
  A crossbar on its own convenient scale looks authoritative and says something false.
- **AVONET has no common names and no body length or wingspan.** Three separate designs have
  foundered on assuming otherwise. Common names come from the ABA checklist (P-004); length and
  wingspan are handled by P-060.
- **Rarity codes are an ingest-time filter only** (P-001). They are read to select codes 1–3 and
  then discarded — they never enter the dataset.
- **Identify the client on outbound requests** (T-006). Go's default `Go-http-client/1.1` earns a
  403 from Wikimedia and is discourteous everywhere else.

## Conventions

**American spellings** in prose, comments, and commit messages — but **never edit a quotation or a
vendored document** to match. Quoted terms of service are verbatim citations with recorded hashes;
"fixing" a publisher's spelling falsifies the quote and breaks the record.

Keep the repository's formatter away from `spec/sources/` and `spec/process.md` for the same
reason.
