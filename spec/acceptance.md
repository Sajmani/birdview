# Acceptance criteria

For each requirement, a check that would fail if the requirement were violated.

> **No code-level criterion here has been watched failing, because no code exists yet.** Every one
> is `planned`. Per [process.md](process.md#guidance), a criterion you have not seen go red is not
> yet a criterion — it is an intention. Before any status becomes `verified`, break the thing it
> checks and confirm it fails.
>
> The four spec-level criteria are the exception. They run today, they pass, and three of them have
> each been **deliberately broken and watched failing** — the fourth, AC-002, has no code to run
> against yet.

Each criterion records its ID, the requirements it verifies, its level (`spec` or `code`), its
method, how to run it, and its status.

---

## Standing checks

Every change passes these, in order, regardless of what it touched:

```
go build ./...
go vet ./...
gofmt -l .            # must print nothing
go test ./...
```

---

## Spec-level criteria

These have the requirement set as their subject and can run before a line of code exists. They
re-run whenever a requirement or a source pin changes.

### AC-001 — Every requirement cites a declared goal, and every goal is cited
**Verifies:** the structural rules in [process.md](process.md) · **Level:** spec · **Method:**
script over `goals.md`, `product.md`, `tech.md`
**Run:** `python3 spec/check/goals.py`
**Status: verified — run 2026-09-21, passing, and watched failing.** Three rules: every active
requirement has a `Goals:` field, every cited goal is declared, every declared goal is cited. It
found four uncited requirements on first run (T-003, T-006, T-011, T-020), which is what prompted
G-004. Exercised by deleting T-003's `Goals:` line: exit 1, `requirements with no Goals: field ->
T-003`.

**This check cannot tell whether a requirement genuinely serves the goal it names.** Goals are not
measurable by construction. That judgment belongs to Gate 1 and to review, and a green run here
looks exactly like a validated one.

### AC-002 — Generated family colors meet the contrast requirement
**Verifies:** T-010, P-005, `wcag/1.4.6` · **Level:** spec and code · **Method:** pure function
over the color generator, asserting WCAG 2.2 relative-luminance contrast ≥ 7:1 against `#000000`
for every family in the dataset
**Run:** `go test ./internal/palette -run TestContrast`
**Status: planned.** The expected value comes from the requirement (7:1), not from any run. Watch
it fail by lowering a generated color's luminance.

### AC-003 — Every vendored source matches its recorded hash
**Verifies:** T-021, G-004 · **Level:** spec · **Method:** hash each committed input and compare
against `spec/sources/*/PROVENANCE.md`
**Run:** `python3 spec/check/pins.py`
**Status: verified — run 2026-09-21, passing** for `avonet` and `fox-wingspan`, **and watched
failing**: zeroing AVONET's recorded hash produced exit 1 and a recorded-vs-actual diff.
**Scope exception:** `wiley` is quoted, not vendored, so it has no hash to check — see
[sources.md](sources.md#wiley). `aba` is fetched at build and checked by AC-020 instead.

### AC-004 — Decision records reference requirements that exist
**Verifies:** process integrity · **Level:** spec · **Method:** extract every `P-###` / `T-###`
cited in `decisions.md` and confirm each resolves
**Run:** `python3 spec/check/refs.py`
**Status: verified — run 2026-09-21, passing, and watched failing** — appending a reference to a
requirement id in the P-9xx range, which nothing defines, produced exit 1.

*Known limitation, found by this check catching its own documentation:* the scanner cannot
distinguish a citation from a mention, so writing an example id in prose trips it. Worth
tolerating — the failure is loud, immediate, and safer than a checker that guesses intent.

---

## Dataset and ingest

### AC-010 — Scope is exactly the ABA code 1–3 species
**Verifies:** P-001 · **Level:** code · **Method:** unit test over ingest output
**Assertion:** the dataset contains **841** species for ABA checklist 8.19, and no species whose
rarity code is 4, 5 or 6 appears.
**Status: planned.** 841 is derived from the source file — counted from `ABA_Checklist-8.19.csv`
on 2026-09-21, rarity distribution 490 + 258 + 93 — **not** from a run of our code. If the
checklist version changes, this number changes with it and the test should name the version.

### AC-011 — Rarity codes never reach the dataset
**Verifies:** P-001 · **Level:** code · **Method:** schema assertion — no field of the output
record carries a rarity code, under any name
**Status: planned.** Cheap, and it protects a licensing-adjacent decision: the codes are ABA's
editorial judgment, and the project uses them only as a filter.

### AC-012 — Every record carries all six AVONET traits
**Verifies:** P-002, P-003 · **Level:** code · **Method:** property test over every record
**Assertion:** culmen, bill depth, tail, wing chord, secondary and mass are all present and
positive.
**Status: planned.** Trait coverage on `AVONET2_eBird` was measured complete for the target
families, so any failure means a join or parse bug rather than missing data.

### AC-013 — Names and families are populated from ABA, never defaulted
**Verifies:** P-004 · **Level:** code · **Method:** property test
**Assertion:** common name, family scientific name and family common name are all non-empty, and
**no record's common name equals its scientific name** — the tell of a silent fallback, which an
earlier design shipped for every one of its species.
**Status: planned.**

### AC-014 — The join covers every species, and the direct-match rate is monitored
**Verifies:** P-006 · **Level:** code · **Method:** unit test over ingest
**Assertions:** (a) after the mapping table is applied, 100% of in-scope species resolve to a
trait row; (b) the *direct* binomial match rate is reported, and is **≥ 97.38%** for the current
source versions.
**Status: planned.** 97.38% (819 of 841) was measured against `AVONET2_eBird` on 2026-09-21. It is
a regression guard, not a target: a drop means a taxonomy moved and the mapping table needs review.

### AC-015 — Unmatched species fail the build above a threshold
**Verifies:** P-007 · **Level:** code · **Method:** unit test with a fixture containing a bogus
binomial
**Assertion:** ingest exits non-zero and names the unmatched species.
**Status: planned.** Watch it fail by raising the threshold until the bogus fixture passes.

---

## Ordering and sectioning

### AC-020 — Sorting is ascending by the selected metric
**Verifies:** P-020, P-021 · **Level:** code · **Method:** property test for each of the three
metrics
**Status: planned.**

### AC-021 — Section boundaries obey the N% rule
**Verifies:** P-022, P-023 · **Level:** code · **Method:** property test over generated inputs
**Assertions:** for every section, `max / min < 1 + N/100`; and the first member of each section
is `>= (1 + N/100) x` the **minimum** of the previous section. Both hold for N across its full
adjustable range, and for N = 15 by default.
**Status: planned.** Derived from the requirement's wording, not from observed output. Note the
rule is stated against the section *minimum*, not a running anchor — a plausible off-by-one
implementation using the previous *maximum* passes casual inspection and fails this.

### AC-022 — Section headers match their contents
**Verifies:** P-024 · **Level:** code · **Method:** unit test
**Assertion:** each header's stated range equals the min and max of its members in the selected
metric, and its count equals the number of members.
**Status: planned.**

---

## Diagram

### AC-030 — Segments sum exactly to total length
**Verifies:** P-041 · **Level:** code · **Method:** property test over every record
**Assertion:** the five **serialized** segment values sum **exactly** to the serialized total
length — not within a tolerance.
**Status: planned.** Exactness is deliberate and forces the implementation to balance rounding
rather than round each segment independently. An earlier design rounded independently and broke
this for 27.6% of its species by up to 0.1 mm; the drift was invisible on screen and made the
stated invariant false.

### AC-031 — Scale is shared within a section, and globally in global mode
**Verifies:** P-042 · **Level:** code · **Method:** unit test over computed geometry
**Assertions:** in per-section mode every tile in a section has an identical scale factor, and the
factor differs between at least two sections in a representative dataset; in global mode one
factor applies to every tile in the filtered set.
**Status: planned.** The second half of the per-section assertion matters — a per-tile
implementation satisfies "identical within a section" trivially when sections hold one bird.

### AC-032 — The wing crossbar is to scale with the vertical axis
**Verifies:** P-043 · **Level:** code · **Method:** property test over computed geometry
**Assertion:** `crossbarLength / bodyLength == wingspan / totalLength` for every tile, within
floating-point tolerance.
**Status: planned.** Derived from the requirement. An earlier design's formula carried an extra
factor, inverting the relationship so that larger birds drew *shorter* crossbars — 300 of 3,213
species floored out, and they were the 300 largest. This assertion catches that class exactly.

### AC-033 — The diagram has the specified parts in the specified order
**Verifies:** P-040 · **Level:** code · **Method:** unit test over rendered output
**Assertion:** five vertical segments top to bottom — bill, head, neck, body, tail — plus one
crossbar; the tail is rendered with a stroke style distinct from the body.
**Status: planned.**

### AC-034 — Proportions differentiate species as the requirement claims
**Verifies:** P-044 · **Level:** code · **Method:** ordering assertions over a small set of
species pairs
**Assertion:** a dowitcher's culmen fraction exceeds a robin's; a heron's neck fraction exceeds a
sparrow's; an accipiter's tail fraction exceeds a duck's.
**Status: planned.** The expected orderings come from ornithology, not from running the code —
which is what makes this a check rather than a characterization. If an ordering fails, the model
is wrong, not the test.

### AC-035 — The neck length is serialized with its model named
**Verifies:** P-045, P-062 · **Level:** code · **Method:** property test over every record
**Assertions:** every record carries a neck length; its provenance `kind` is `modeled`; its
`model` is non-empty and **names both the per-family ratio function and the numeric ratio applied
to that species' family**; and the ratio named matches the table entry for that family.
**Status: planned.** Inverted by [CR-007](decisions.md#cr-007) — this criterion previously
asserted the opposite, that no neck length was ever serialized. The last assertion is the one with
teeth: a `model` string that names a function but not the family's actual value would satisfy a
looser check while telling the reader nothing they could verify.

---

## Provenance

### AC-040 — Every presented metric has a well-formed provenance record
**Verifies:** P-062 · **Level:** code · **Method:** schema validation plus a property test over
every record
**Assertions:** every metric carries `kind` in `{measured, near-match, modeled}`; `source` is
non-empty; `sourceTaxon` is non-empty for `measured` and `near-match`; `model` is non-empty for
`modeled`; `sourceTaxon` differs from the species' own name whenever `kind` is `near-match`.
**Status: planned.** The last assertion is the one with teeth — it makes "near-match" mean
something rather than being a label anyone can attach.

### AC-041 — Provenance matches the sources actually used
**Verifies:** P-060, P-062 · **Level:** code · **Method:** unit test over ingest output
**Assertions:** every length has `kind = modeled`; exactly the species with an empirical Fox value
have `kind = measured` for wingspan; the six species named in
[CR-006](decisions.md#cr-006) have `kind = near-match` for their AVONET-derived traits.
**Status: planned.** Measured 2026-09-21 from the source data: **301 of 841** in-scope species
have an empirical wingspan (`needs_prediction = 0`), 510 have only a predicted value which this
project excludes, and 30 have no row at all — every one of them a Galliform, an order the Fox
dataset omits entirely. These counts come from the sources, not from our output.

### AC-042 — The tile marks every non-measured metric
**Verifies:** P-061 · **Level:** code · **Method:** rendering test asserting a marker is present
for each metric whose `kind` is not `measured`, and absent otherwise
**Status: planned.**

### AC-043 — The detail view shows every metric and segment with provenance
**Verifies:** P-063, P-034 · **Level:** code · **Method:** rendering test
**Assertions:** the detail view shows all three main metrics **and all five diagram segments**
with their values; for each it names the source; for a `near-match` it names the taxon actually
measured; for a `modeled` value it names the model. The detail view reproduces the tile's header,
diagram and main metrics — it is the same layout enlarged, not a different screen.
**Status: planned.**

### AC-044 — Total length is reported as the sum of its segments
**Verifies:** P-060, P-062, P-041 · **Level:** code · **Method:** property test
**Assertions:** total length's `kind` is `modeled`; its `model` states that it is the sum of the
five segments; and the value equals that sum exactly (the same invariant AC-030 asserts from the
geometry side).
**Status: planned.** Added by [CR-007](decisions.md#cr-007). Surfacing the segments makes it
visible that a *measured* bill sits inside a *modeled* total; this criterion makes the output
explain that rather than leaving the reader to infer it.

---

## Technical

### AC-050 — Nothing restricted is tracked by version control
**Verifies:** T-004, T-005, G-004 · **Level:** code · **Method:** `git ls-files` assertion in CI
**Assertion:** no tracked path lies under `fetched/`, and the built dataset is not tracked.
**Status: planned, and it is the highest-value check in this document.** It is the only mechanical
guard on the project's licensing position; everything else protecting T-004 is a convention
someone can forget. Watch it fail by `git add -f` on a fetched file.

### AC-051 — The server makes no network request
**Verifies:** T-002 · **Level:** code · **Method:** integration test serving every route with
outbound network unavailable
**Status: planned.**

### AC-052 — Outbound requests identify the client
**Verifies:** T-006, G-004 · **Level:** code · **Method:** unit test against an `httptest` server
**Assertion:** the `User-Agent` header is present, non-empty, and is not Go's default.
**Status: planned.**

### AC-053 — The fetched ABA file matches its recorded hash
**Verifies:** T-004, P-001 · **Level:** code · **Method:** the ingest verifies the hash after
download and fails if it differs
**Assertion:** SHA-256 `9c9539b7…6331d3`, verified against the real file on 2026-09-21.
**Status: planned.**

### AC-054 — Both commands build and the server serves
**Verifies:** T-001, T-003 · **Level:** code · **Method:** covered by the standing checks plus an
integration test that starts the server and requests `/`
**Status: planned.**

---

## Traceability

| Requirement | Criteria | |
| --- | --- | --- |
| P-001 | AC-010, AC-011, AC-053 | |
| P-002, P-003 | AC-012 | |
| P-004 | AC-013 | |
| P-005 | AC-002 | |
| P-006 | AC-014 | |
| P-007 | AC-015 | |
| P-010 | — | **gap** |
| P-011, P-012 | — | withdrawn |
| P-020, P-021 | AC-020 | |
| P-022, P-023 | AC-021 | |
| P-024 | AC-022 | |
| P-030 | — | **gap** |
| P-031 | — | **gap** |
| P-032 | — | **gap** |
| P-033 | — | **gap** |
| P-034 | AC-043 | partial |
| P-040 | AC-033 | |
| P-041 | AC-030 | |
| P-042 | AC-031 | |
| P-043 | AC-032 | |
| P-044 | AC-034 | |
| P-045 | AC-035 | |
| P-060 (composite) | AC-044 | |
| P-050 – P-053 | — | withdrawn |
| P-060 | AC-041 | |
| P-061 | AC-042 | |
| P-062 | AC-040 | |
| P-063 | AC-043 | |
| T-001, T-003 | AC-054 | |
| T-002 | AC-051 | |
| T-004 | AC-050, AC-053 | |
| T-005 | AC-050 | |
| T-006 | AC-052 | |
| T-010 | AC-002 | |
| T-011 | — | **gap** — a posture, not a testable assertion |
| T-020, T-021 | AC-001, AC-003, AC-004 | |
| `wcag/1.4.6` | AC-002 | |

### Declared gaps

Honest gaps, rather than criteria that would not bite:

- **P-010 (family filter), P-030–P-033 (modes, tile contents, units, first load).** These are
  interaction behaviors. They are testable, but only with a browser harness this project does not
  have, and a unit test asserting that a template contains a string verifies the template rather
  than the behavior. Covered by human review until a harness exists; promote them when one does.
- **T-011 (never commercial).** A standing posture with no code to check. It is enforced by the
  human gate and by `sources.md`, not by a test.
- **P-034 (detail view)** is only partially covered: AC-043 checks that provenance appears there,
  not that the rest of the view is right.

### Criteria that cannot yet fail

All `planned` criteria. The project has no code, so none of them has been watched going red. As
each is implemented, break its subject deliberately and record the result here — a check believed
but never seen failing is worse than no check.
