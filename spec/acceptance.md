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

### AC-002 — Generated colors are legible and mutually distinguishable
**Verifies:** T-010, P-005, `wcag/1.4.6` · **Level:** spec and code · **Method:** pure functions
over the color generator
**Assertions:** WCAG 2.2 contrast ≥ 7:1 against each badge's **own** text color; CIE76 ΔE ≥ 25
between families adjacent in the selector; both halves of the gamut are used; and the preset cycle
strictly alternates light and dark, which requires an even count
**Run:** `go test ./internal/palette`
**Status: verified 2026-09-22 — passing, and watched failing** on both halves: lowering the
contrast floor to 3.0 produced four violations, all blues and violets, where relative luminance is
lowest at a given HSL lightness; and restoring even hue stepping dropped adjacent families to
ΔE 2–6.

The distinctness half was added after the fact, and the gap it closed is worth recording. The
original check asserted only that no two families shared a hex value. All 91 differed, so it
passed — while every pair adjacent in the selector was visually identical. Exact inequality is not
distinguishability, and the entries a reader compares are the ones next to each other.

Two further assertions came from bugs found while fixing that. **Both halves of the gamut**: the
capacity argument behind P-005 depends on white text being used where it helps, and a generator
that quietly reverted to black-only would give that up invisibly. **Alternation**: an odd-length
light/dark cycle wraps onto itself, so one pair of hue-adjacent slots shares a lightness — which
is what left two pale Passeriformes families at ΔE 4.4. Both have been watched failing.
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
**Status: verified 2026-09-22 — the ingest produces exactly 841 species.**
on 2026-09-21, rarity distribution 490 + 258 + 93 — **not** from a run of our code. If the
checklist version changes, this number changes with it and the test should name the version.

### AC-011 — Rarity codes never reach the dataset
**Verifies:** P-001 · **Level:** code · **Method:** inspect the serialized record's field names
for anything carrying an ABA rarity code
**Status: verified 2026-09-22 — no rarity field appears in the serialized record.**
editorial judgment, and the project uses them only as a filter.

### AC-012 — Every record carries all five AVONET traits
**Verifies:** P-002, P-003 · **Level:** code · **Method:** property test over every record
**Assertion:** culmen, tail, wing chord, secondary and mass are all present and positive.
**Status: planned.** Trait coverage on `AVONET2_eBird` was measured complete for the target
families, so any failure means a join or parse bug rather than missing data.

### AC-013 — Names and families are populated from ABA, never defaulted
**Verifies:** P-004 · **Level:** code · **Method:** property test
**Assertion:** common name, family scientific name and family common name are all non-empty, and
**no record's common name equals its scientific name** — the tell of a silent fallback, which an
earlier design shipped for every one of its species.
**Status: verified 2026-09-22 — no species falls back to its scientific name as a common name.**

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
**Status: verified 2026-09-22 — ascending for all three metrics.**

### AC-021 — Section boundaries obey the N% rule
**Verifies:** P-022, P-023 · **Level:** code · **Method:** property test over generated inputs
**Assertions:** for every section, `max / min < 1 + N/100`; and the first member of each section
is `>= (1 + N/100) x` the **minimum** of the previous section. Both hold for N across its full
adjustable range, and for N = 15 by default.
**Status: verified 2026-09-22 — property test over 1,500 random datasets at five step values, and **watched failing**: keying the break on the previous *maximum* instead of the minimum let a section span 7% at a 5% step.**
rule is stated against the section *minimum*, not a running anchor — a plausible off-by-one
implementation using the previous *maximum* passes casual inspection and fails this.

### AC-022 — Section headers match their contents
**Verifies:** P-024 · **Level:** code · **Method:** unit test
**Assertion:** each header's stated range equals the min and max of its members in the selected
metric, and its count equals the number of members.
**Status: verified 2026-09-22 — header ranges match their contents.**

---

## Diagram

### AC-030 — Segments sum exactly to total length
**Verifies:** P-041 · **Level:** code · **Method:** property test over every record
**Assertion:** the five **serialized** segment values sum **exactly** to the serialized total
length — not within a tolerance.
**Status: verified 2026-09-22 — passing, and watched failing: rounding each segment independently instead of balancing broke the sum for both the property test and the real dataset, by 0.1 mm.**
rather than round each segment independently. An earlier design rounded independently and broke
this for 27.6% of its species by up to 0.1 mm; the drift was invisible on screen and made the
stated invariant false.

### AC-031 — Scale is shared within a section, and globally in global mode
**Verifies:** P-042 · **Level:** code · **Method:** unit test over computed geometry
**Assertions:** in per-section mode every tile in a section has an identical scale factor, and the
factor differs between at least two sections in a representative dataset; in global mode one
factor applies to every tile in the filtered set; **per-section is the default**, global is
reachable by `?scale=global`, and per-section is listed first in the controls.
**Status: verified 2026-09-22 — passing, and **watched failing**: scaling each bird to its own length was caught by the direct px/mm comparison. Worth noting that it was *not* caught by the first two assertions written for this criterion, because the wingspan constraint happened to preserve size ordering; comparing the scale factor itself is the only form that cannot pass by accident.**
implementation satisfies "identical within a section" trivially when sections hold one bird.

The default moved from global to per-section in [CR-017](decisions.md), and the tests asserting the
old default were inverted rather than deleted. Flipping it exposed a real bug: `parseOptions`
recognised only the **non**-default value, so the toggle stopped working entirely the moment the
default changed. Written that way the parser is correct exactly once.

### AC-032 — The wing crossbar is to scale with the vertical axis
**Verifies:** P-043 · **Level:** code · **Method:** property test over computed geometry
**Assertion:** `crossbarLength / bodyLength == wingspan / totalLength` for every tile, within
floating-point tolerance.
**Status: verified 2026-09-22 — passing, and **watched failing**: restoring the earlier design's formula, with its extra `availableH × scaleY` factor, broke the ratio for every species.**
factor, inverting the relationship so that larger birds drew *shorter* crossbars — 300 of 3,213
species floored out, and they were the 300 largest. This assertion catches that class exactly.

### AC-033 — The diagram has the specified parts in the specified order
**Verifies:** P-040 · **Level:** code · **Method:** unit test over rendered output
**Assertion:** five vertical segments top to bottom — bill, head, neck, body, tail — plus one
crossbar; the tail is rendered with a stroke style distinct from the body.
**Status: verified 2026-09-22 — coordinates are monotonic and the head oval is centered between its tangents.**

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
**Status: verified 2026-09-22 — every neck model names its family and a ratio.**
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
**Status: verified 2026-09-22 — passing, and watched failing: blanking the neck model string produced a violation for every species.**
something rather than being a label anyone can attach.

### AC-041 — Provenance matches the sources actually used
**Verifies:** P-060, P-062 · **Level:** code · **Method:** unit test over ingest output
**Assertions:** every length has `kind = modeled`; exactly the species with an empirical Fox value
have `kind = measured` for wingspan; the six species named in
[CR-006](decisions.md#cr-006) have `kind = near-match` for their AVONET-derived traits.
**Status: verified 2026-09-22 — 301 measured wingspans, matching the count derived from the sources before this code existed. Watched failing: removing the near-match inheritance reported 308, i.e. seven borrowed values mislabeled as measured.**
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
**Status: verified 2026-09-22 — all eight metrics appear with provenance, and **watched failing**: removing the neck disclosure was caught.**

### AC-044 — Total length is reported as the sum of its segments
**Verifies:** P-060, P-062, P-041 · **Level:** code · **Method:** property test
**Assertions:** total length's `kind` is `modeled`; its `model` states that it is the sum of the
five segments; and the value equals that sum exactly (the same invariant AC-030 asserts from the
geometry side).
**Status: verified 2026-09-22 — total length names itself a sum of measured and modeled parts.**
visible that a *measured* bill sits inside a *modeled* total; this criterion makes the output
explain that rather than leaving the reader to infer it.

### AC-045 — The fitted wingspan model reconstructs Fox's published one
**Verifies:** P-060, P-062 · **Level:** code · **Method:** fit the model, then compare its
estimate against `wingspan_predicted_*` for every species Fox predicted
**Assertions:** mean disagreement ≤ 2%; no more than 2% of species disagree by over 10%.
**Status: verified 2026-09-22 — passing at 1.32% mean, 1.4% over 10% across 9,194 species.**
The thresholds come from the pre-implementation comparison (mean 1.01%, p99 4.15%), not from
running this code. Its value is as a tie to an external published reference: if the fitting
regresses, the dataset does not quietly get worse, the check goes red.

---

## Technical

### AC-050 — Nothing restricted is tracked by version control
**Verifies:** T-004, T-005, G-004 · **Level:** code · **Method:** `git ls-files` assertion in CI
**Assertion:** no tracked path lies under `fetched/`, and the built dataset is not tracked.
**Status: planned, and it is the highest-value check in this document.** It is the only mechanical
guard on the project's licensing position; everything else protecting T-004 is a convention
someone can forget. Watch it fail by `git add -f` on a fetched file.

### AC-051 — The server makes no network request
**Verifies:** T-002 · **Level:** code · **Method:** rendering test scanning for external
**subresources** — any `src`, and `href` on a `<link>`
**Status: verified 2026-09-22 — passing, and **watched failing**: adding the old design's `cdn.tailwindcss.com` script tag was caught. The check scans rendered HTML for off-host `src`/`href`, which is where T-002 actually bites — a self-contained binary serving a page that fetches a CDN stylesheet is not self-contained.**

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
**Status: verified 2026-09-22 — every route serves, unknown species 404s.**

### AC-055 — The tile carries the header, diagram and main metrics
**Verifies:** P-031 · **Level:** code · **Method:** rendering test
**Assertions:** every tile shows the common name with its banding code in parentheses, the
scientific name, a family badge naming both the scientific and common family name, an SVG diagram,
and the three main metrics.
**Status: verified 2026-09-22.**

### AC-056 — The family filter selects and clears
**Verifies:** P-010 · **Level:** code · **Method:** rendering test over `?family=`
**Assertions:** filtering to one family shows its species and excludes others; no filter shows
all; every entry is labeled with its scientific name, its common name and its species count; the
selector is grouped by taxonomic order; groups appear in checklist sequence rather than
alphabetically; each family sits inside its own order's group; each order is a working toggle that
selects or clears **every** family in it, labeled with its common name; and **All** toggles the
same way over the whole set, clearing the selection when untoggled.
**Status: verified 2026-09-22 — passing, and watched failing** by alphabetizing the family list,
and separately by making the order toggle add only its first family.

Two lessons are recorded in the test itself. An earlier version asserted that order labels were
*inert text*; when the requirement inverted to make them toggles, that check kept passing, because
it searched for `<a class="ohdr` and the markup had become `<a class="btn ohdr`. And the
select-the-whole-group behavior could not be tested at all until the fixture gained an order
holding **two** families — with one family per order, "select the group" and "select the first
family" are the same operation.

The All control needed the same care for a different reason: "nothing selected" and "everything
selected" both display all 841 species, so no assertion on the species count can tell them apart.
The check reads the control's own state and the orders' states instead.

### AC-057 — The unit toggle switches imperial and metric
**Verifies:** P-032 · **Level:** code · **Method:** rendering test over the tile footer, plus a
unit test pinning the conversions
**Assertions:** the tile footer carries two length values and one mass value in the selected
system's units, and differs between the two systems; mass converts as `1 oz = 28.349523125 g`.
**Status: verified 2026-09-22 — passing, and watched failing** on both a hardcoded-inches footer
and a kilograms-under-imperial conversion.

The first version of this criterion asserted only that a centimeter value appeared *somewhere* on
the page, and passed for weeks while every tile was hardcoded to inches: the section headers alone
satisfied it. Counting units inside the footer is what makes it bite — a footer rendering `3.9"`
contains no `" in"` token, so an exclusion list misses it entirely.

### AC-058 — The detail view draws its bird full scale
**Verifies:** P-034 · **Level:** code · **Method:** rendering test over the smallest fixture
species, with and without an explicit `?scale=`
**Assertion:** the diagram fills the detail view's height regardless of the index's scale mode.
**Status: verified 2026-09-22 — passing, and watched failing** by making the detail view adopt the
dataset-wide reference, which inset the hummingbird by 128 px.

### AC-059 — The page explains every control and the diagram
**Verifies:** P-035 · **Level:** code · **Method:** rendering test over the help block
**Assertions:** the **Controls** panel explains every control it offers (order, scale, step, units,
family) and both scale modes; the **Presentation** panel's two sections appear in the stated order —
the diagram, then where the numbers come from; its **rendered legend** names all five segments plus
the wingspan and the wing chord; the estimate marker is explained; both panels start expanded.

Each explanation is asserted in its own scope. A whole-page search would let the controls panel
satisfy a claim about the Presentation panel, and vice versa.
**Status: verified 2026-09-27 — passing, and watched failing** by removing the scale explanation,
by removing the tail from the legend, by suppressing the legend diagram, and by swapping two
sections.

The legend is a **rendered diagram of a real species** rather than a list of colour swatches
([CR-017](decisions.md)), which closes a rot the old one had: it described the tail as "dashed, to
tell it from the body", and that stopped being true when CR-015 gave dash a different meaning.
Nothing caught it, because prose and stylesheet had no connection. A rendered diagram cannot drift
from the renderer.

This criterion is tied to the controls that exist, so adding a control without documenting it
fails here. That is deliberate: help text rots silently otherwise.

### AC-060 — Text on a generated background inherits that background's text color
**Verifies:** P-005, T-010 · **Level:** code · **Method:** structural check over `styles.css`
**Assertions:** no rule targeting a descendant of `.chip` or `.badge` sets its own `color`, and no
such rule dims text with `opacity`.

**Status: verified 2026-09-22 — passing, and watched failing** by restoring the fixed grey.

Chips and badges carry a generated background paired with a black-or-white text color chosen so
the pair clears 7:1. The stylesheet set the family common name to a fixed grey, which ignores
which half of the palette a chip landed on and went unreadable on both — the common name vanished
on roughly a dozen chips. Hierarchy has to come from weight and size; dimming undoes the very
ratio the background was selected for. Checking CSS is crude, but the defect is invisible to a
rendering test and this catches it exactly.

### AC-061 — The empty state guides rather than blames
**Verifies:** P-037, P-010 · **Level:** code · **Method:** rendering test after clearing every
selection
**Assertions:** no species are displayed; the page says how to get them back; it does **not** say
that nothing matched; the count reports zero; and All is no longer marked selected.
**Status: verified 2026-09-22 — passing, and watched failing** both by treating an emptied
selection as "no filter" (the previous behavior, which showed every bird) and by reverting the
message to "No species match that filter".

### AC-062 — The scroll script is embedded and scoped to the index
**Verifies:** P-038, T-002 · **Level:** code · **Method:** rendering test
**Assertions:** the index carries the script inline, never via `src`; it is scoped to `/` so a
detail page cannot consume the saved position; the detail view ships no script; the controls carry
an `id` so `#controls` works as a fallback target; both panels carry ids so their collapsed state
can be keyed on; both still default to open; and panel state is restored **before** the scroll.
**Status: verified 2026-09-22 — passing, and watched failing** by externalizing the script, by
removing its path guard, by dropping a panel's id, and by reordering the restores so the scroll
runs first.

**This does not check that scrolling works.** That needs a browser, which this project does not
have. The criterion covers presence, self-containment and scoping only, and the behavior itself is
a declared gap.

### AC-063 — The controls offer what the server accepts, default first
**Verifies:** P-023, P-042 · **Level:** code · **Method:** rendering test scoped to the controls
panel, plus round-tripping each offered step through `parseOptions`
**Assertions:** Global is listed before Per section; every offered section step (5, 10, 15, 20,
30) is parsed as itself rather than falling back to the default.
**Status: verified 2026-09-22 — passing, and watched failing** by reversing the scale pair and by
offering a step the server rejects.

The scoping matters and was got wrong first. Searching the whole page for the scale labels matched
the **Presentation panel**, which named both modes in the same order, so the check passed with the
buttons reversed. It now reads only the controls panel.

### AC-064 — Attribution is complete, and claims nothing unearned
**Verifies:** T-011, G-004 · **Level:** code · **Method:** rendering test over the footer
**Assertions:** each CC BY source is credited with creator, a link to its deposit and a link to the
license; the ABA checklist is credited as the species list; the footer discloses that lengths are
modeled; the detail view carries credit too, being a shareable URL of its own; and **no page claims
a permission the project does not hold**.
**Status: verified 2026-09-22 — passing, and watched failing** by stripping the license links and
by reinstating a "used by permission" claim.

The second half exists because the first version of this footer said the ABA checklist was "used by
permission". It is not — that request has not been sent. Claiming a right you have not been granted
is worse than claiming none, because it misleads the rights holder's own readers. Remove that
assertion only when a written grant exists and is recorded in [sources.md](sources.md).

### AC-065 — The cranium varies with the bird
**Verifies:** P-046, P-062 · **Level:** code · **Method:** unit test on the model plus a
distribution check over the built dataset
**Assertions:** head length matches `7.5 × mass^(1/3)` at stated points, grows strictly across the
real mass range with no plateau, no single value is shared by more than 5% of species, and the
heaviest bird has the largest head.
**Status: verified 2026-09-22 — passing, and watched failing** by reinstating a floor, which
pinned 809 of 841 species at one value.

The distribution check is the one that matters. A unit test on the formula would have passed
happily while the *old* model collapsed 442 species onto 12 mm, because that model also satisfied
its own definition. What was wrong was the shape of the result across the dataset, which only a
distribution assertion can see.

### AC-066 — *(withdrawn)* Legs are modeled from the tarsus
**Withdrawn 2026-09-23 by [CR-010](decisions.md#cr-010).** The requirement it verified no longer
exists: the diagram is a bird in flight and has no legs.

Kept as a record rather than deleted, because the criterion is why the removal was cheap. It
asserted that some species' feet fall below the tail tip **and** that only wader families may do
it, and it checked the leg model against the whole dataset rather than a fixture. Those checks are
what made it obvious that the modeled leg underpredicts the flight projection — 14 of 110 species
in the trailing-leg families clearing the tail against a true figure far higher, and a White Ibis
coming out negative. A criterion that only confirmed the code did what the code did would have
said nothing when the premise changed.

### AC-067 — A folded neck is the same stroke in less height
**Verifies:** P-040, P-041, P-034 · **Level:** code · **Method:** unit tests over the geometry plus
assertions on the rendered page

**Assertions, geometry:** for a folding family the neck is an S and for every other family it is a
stick; the S's arc length `2πr` equals the scaled neck exactly; its vertical span is `2/π` of that
length, checked against the ratio rather than against whatever the code produces; the two arcs join
halfway; and the whole drawing is shorter by exactly what the neck saved and by nothing else, so no
other segment may shift.

**Assertions, length as data:** a folding bird and a non-folding bird of identical total length
share a section, and the scale reference is the full total length in both scale modes — checked
against a section holding **only** the folding bird as well as a mixed one.

**Assertions, rendering:** the heron's neck reaches the page as `<path class="neck">` and every
other bird's as `<line class="neck">`, never both and never neither; and the Neck row of the detail
view still reports the full modeled length, read out of that row rather than searched for in the
whole page.

**Status: verified 2026-09-23 — passing, and watched failing** by: emptying the folding family
table; setting the arc radius to `neck/4` so the stroke is rescaled; leaking the fold into
`Scales`; forcing the template to always draw a line; and reporting the drawn neck instead of the
modeled one in the detail view.

Two vacuity notes, both found by running those breaks. The identical-twin construction that makes
the geometry assertions clean **defeats** the scale assertion: `Scales` takes a maximum, so the
extended twin masks any reduction applied to the heron, and the leak break passed until a
heron-only section was added. And the ordering test ends by asserting the two birds really do draw
to different heights — without it, a change that stopped folding anything would satisfy every
other line in that test.

### AC-068 — The iNaturalist link appears only for a confident match
**Verifies:** P-064, T-022 · **Level:** code · **Method:** unit tests over the matcher and the
mapping table, plus assertions on the rendered detail page; the resolver is exercised against an
`httptest` server, never the live API

**Assertions, matching:** a taxon whose species epithet differs is rejected; a taxon whose common
name disagrees with ABA's is rejected even when the scientific name matches exactly; and each of
the five forms of agreement is accepted — a Hawaiian okina against its absence, an ABA parenthetical
alternate, British against American spelling, a hyphen against a space, and a name iNaturalist holds
but does not prefer. A **different genus** is accepted only alongside an agreeing common name, and
refused without one ([CR-016](decisions.md)). A **gendered epithet** is matched across a change of
genus — *Coccothraustes vespertinus* against *Hesperiphona vespertina* — since a specific epithet
agrees in grammatical gender with its genus, but only where the shared stem is long enough to be
evidence.

**Assertions, coverage:** every one of the 841 in-scope species has a taxon ID. Total coverage is
required rather than reported: under these rules a species that fails to resolve is a fault to
diagnose, not an absence to accept.

**Assertions, the committed table:** every entry parses; every taxon ID is a positive integer and
unique; every scientific name in it is one BirdView actually has, so a stale row cannot sit
unnoticed after a taxonomic change.

**Assertions, rendering:** a species with a mapping shows a link whose href is exactly
`https://www.inaturalist.org/taxa/<id>`; a species **without** one shows no link and no
placeholder, checked by scoping to the row rather than searching the page; and the page embeds no
iNaturalist image — the existing AC-051 subresource scan already covers the mechanism, and this
adds the intent.

**Assertions, courtesy:** the resolver sends a descriptive `User-Agent` (T-006, `inat-api/R4`) and
waits at least the configured interval between requests (T-022, `inat-api/R1`), both observed
against the test server.

**Status: verified 2026-09-27 — passing, and watched failing** by: ignoring the common name;
emitting a search URL instead of nothing; removing the pacing; dropping the `User-Agent`; bypassing
the zero-ID guard; deleting a row from the committed table; and reverting the gender stemming.

Coverage is **841 of 841**, reached by widening the evidence rather than lowering the bar
([CR-016](decisions.md)).

The confidence gate is the part worth testing hardest, because the failure it prevents is silent.
Matching on scientific name alone would find the right page nearly always, and the rare miss would
be a *plausible wrong bird* rather than a visible error — which is exactly the class of bug this
project keeps finding in its own tests.

### AC-069 — A family that folds its neck has a neck to fold
**Verifies:** P-040, P-045 · **Level:** code · **Method:** unit test cross-checking the two neck
tables
**Assertion:** every family in the folded-neck set has an **explicit** entry in
`morphology.NeckRatios`, rather than falling back to the default ratio.

**Status: verified 2026-09-23 — passing, and watched failing against a real bug**, which is how it
was written: Pelecanidae was in the folded set with no ratio entry, so a pelican drew a
5%-of-length stub tucked into an S. The check failed on the live tree before either table changed.

The two tables are independently plausible and jointly wrong, which is why neither review caught
it. A family is in the folded set *because* its neck is the defining feature of its silhouette; a
family on the 0.18 default has no such neck. Listing one without the other is always a mistake.

This does **not** catch the other half of the same bug. Aramidae was missing from the ratio table
while not being in the folded set, so nothing was inconsistent — the Limpkin simply drew a short
neck, and a reader noticed. No automatic check can tell that a family's neck is too short when the
ratios are invented in the first place; only a review of the table against the family list will.
Recorded here so the limit is not mistaken for coverage.

### AC-070 — No text or link falls below its contrast floor
**Verifies:** T-023, `wcag/1.4.3`, `wcag/1.4.6` · **Level:** code · **Method:** two checks parsing
the stylesheet **as embedded in the binary**, not off disk

**Assertions, measured:** every `color` a rule declares — with custom properties resolved — clears
**7:1** against both `--bg` and `--panel`. Rules that also set a `background` are excluded, being
self-contained and governed by T-010.

**Assertions, structural:** the stylesheet carries top-level `a` and `a:visited` rules, so no link
can fall back to the browser default.

**Status: verified 2026-09-23 — passing, and watched failing** by removing the global link rules
(which reproduces the original defect exactly), by setting the visited color to the browser's
`#551A8B`, by picking a link color that clears AA at 4.90:1 but misses 7:1, and — after the floor
rose to 7:1 for everything ([CR-014](decisions.md)) — by restoring the previous `--dim` and
`--accent`, which fail at 5.71:1 and 6.83:1 on a panel.

The structural assertion is the one that would have caught the bug. A measured check sees only
colors that were declared, and the defect was a color nobody declared — the link inherited
`#0000EE` at 1.90:1. Measuring what is there cannot find what is missing.

Writing the check found a bug in the check: the selector capture was swallowing the comment
preceding each rule, so any rule documented with a sentence containing the word "a" was read as a
link selector and held to 7:1. `.sn` failed that way on the first run. Comments are stripped before
parsing, and selectors are split into tokens rather than substring-matched.

### AC-071 — The favicon is BirdView's own, and is actually renderable
**Verifies:** T-002 · **Level:** code · **Method:** assertions over the served asset and both pages

**Assertions:** `/static/favicon.svg` is served with status 200; it **parses as XML** and its root
element is `svg` with a `viewBox`; it is drawn in the diagram's own palette; both the index and a
detail page reference it; and `/favicon.ico` returns 404 rather than the index page, which the
catch-all route makes worth pinning.

**Status: verified 2026-09-23 — passing, and watched failing** by removing the asset while leaving
the reference, by removing the reference while leaving the asset, by pointing the reference at a
CDN (which trips AC-051's existing scan, as intended), and by reintroducing the XML defect below.

No new requirement was added. T-002 already forbids fetching a subresource from another host and
AC-051 already scans `<link href>` for it — confirmed here by watching a CDN favicon fail that
check. What AC-051 cannot see is an icon referenced but not served, or served but referenced by
nothing, so this criterion joins the two ends.

**The format assertion is the part that earned its place.** Its first version searched the body for
the substring `<svg` and passed against a file no browser could read: the SVG's comment contained a
double hyphen, which XML forbids, so the icon would simply never have appeared. It now runs the
parser. A format check that does not parse is not a format check — the same lesson as AC-051's
subresource scan, one layer down.

### AC-072 — The diagram distinguishes measured from modeled, and says so
**Verifies:** P-065, P-040 · **Level:** code · **Method:** geometry unit tests, a structural check
over the stylesheet, and assertions on both rendered pages

**Assertions, geometry:** `ChordTickX` equals the scaled half-span minus the scaled chord exactly;
it is zero when no chord is available or when the chord would reach the axis, so the renderer falls
back to a whole crossbar rather than drawing a tick outside the wing.

**Assertions, stroke style:** in the stylesheet, every modeled element — neck, body, and the
modeled crossbar — carries a `stroke-dasharray`; every measured one — bill, tail, and the plain
crossbar — does not; and the head is drawn with `fill: none`. The **tail must not be dashed**, which
is the assertion that catches a revert to the old meaning of dash.

**Assertions, the neck exception:** the neck's dash pattern differs from the body's. Stated as a
requirement of the check, not merely tolerated by it, because the two look like an inconsistency
and tidying them into agreement makes short necks render solid (CR-015).

**Assertions, rendering:** a species with a **measured** wingspan emits no dashed crossbar segment;
one with a **modeled** wingspan emits exactly one, flanked by two solid segments and marked by two
ticks; and the index and detail pages each carry text explaining what solid, dashed and the tick
mean.

**Status: verified 2026-09-24 — passing, and watched failing** by: restoring the tail's dash;
tidying the neck's pattern to match the body's; giving the head a fill back; placing the tick by
half the chord instead of the chord; removing the overrun fallback so a tick is drawn outside the
wing; and deleting the explanatory text from the index.

Found while writing it: the first version of the rendered-page check was **vacuous twice over**.
The server fixture carried no wing chord, so no crossbar ever split, and no fixture species had a
modeled wingspan, so the dashed case could not arise. Both were fixed in the fixture rather than
worked around in the assertion — every fixture bird now has a chord shorter than its own half-span,
as all 841 real species do, and the crow's wingspan is modeled on purpose.

The tail assertion is the one most likely to look redundant and is the most valuable. Dash used to
mean "this is the tail"; anyone restoring that on aesthetic grounds reverses the meaning of the
whole scheme, and every other assertion here would still pass.

### AC-073 — Every segment is labeled on the detail diagram, and no two labels collide
**Verifies:** P-066 · **Level:** code · **Method:** unit tests over the label layout plus
assertions on the rendered detail page

**Assertions, layout:** all five segments get a label; each label's anchor is the midpoint of its
own segment; labels stay in the order the segments do; no two are closer than the line height; and
the last one stays inside the box.

**Assertions, paint order:** in the rendered diagram every line element is emitted before every
text element, since SVG has no z-index and document order is paint order. The check also requires
the labels to exist, so it cannot be satisfied by a diagram with no text to cover.

**Assertions, geometry:** across every species, the wingspan label placed beyond the wingtip clears
the segment label column, and the reserved gutter is not so wide that the labels have more room
than the bird. Checked over the whole fixture rather than one bird, because the binding case is the
widest-winged species and which one that is changes with the data.

**Assertions, rendering:** the detail page emits five segment labels plus **wing chord** and
**wingspan** labels on the crossbar, each naming what it measures and carrying the same value the provenance table gives
for it — read out of the table rather than recomputed, so the two cannot drift apart. The segment
count is matched on the segment labels alone, so a missing segment cannot hide behind the chord. A
tile carries none.

**Status: verified 2026-09-27 — passing, and watched failing** by anchoring every label at the same
y, by removing the minimum-gap pass, by dropping a segment from the list, by removing the reserved
gutter and by inflating it, and by moving the wing labels back inside the centering group — which
reproduces the bug where the neck's leader struck through "Wing chord" and "Wingspan".

**The collision resolver does not fire on real data: 0 of 841 species need it** at the current box
size and font. It is kept anyway, and tested with a synthetic geometry rather than deleted, because
unlike the dead branch removed in CR-010 this one is not dead *by construction* — it is dead by the
current numbers. A larger font, a shorter box or a new segment would all wake it. The dataset-wide
figure is recorded here so that if someone later finds it untriggered, they find out why before
removing it.

---

## Traceability

| Requirement | Criteria | |
| --- | --- | --- |
| P-001 | AC-010, AC-011, AC-053 | |
| P-002, P-003 | AC-012 | |
| P-004 | AC-013 | |
| P-005 | AC-002, AC-060 | |
| P-006 | AC-014 | |
| P-007 | AC-015 | |
| P-010 | AC-056 | |
| P-011, P-012 | — | withdrawn |
| P-020, P-021 | AC-020 | |
| P-022, P-023 | AC-021 | |
| P-024 | AC-022 | |
| P-030 | AC-054 | |
| P-031 | AC-055 | |
| P-032 | AC-057 | |
| P-033 | AC-056 | |
| P-034 | AC-043, AC-054, AC-058 | |
| P-035 | AC-059 | |
| P-036 | AC-062 | |
| P-037 | AC-061 | |
| P-038 | AC-062 | partial — behavior unverifiable without a browser |
| P-040 | AC-033, AC-067, AC-069, AC-072 | AC-066 withdrawn (CR-010) |
| P-041 | AC-030, AC-067 | |
| P-042 | AC-031, AC-063 | |
| P-043 | AC-032 | |
| P-044 | AC-034 | |
| P-045 | AC-035, AC-069 | |
| P-046 | AC-065 | |
| P-060 (composite) | AC-044 | |
| P-060 (wingspan model) | AC-045 | |
| P-050 – P-053 | — | withdrawn |
| P-060 | AC-041 | |
| P-061 | AC-042 | |
| P-062 | AC-040 | |
| P-063 | AC-043 | |
| P-064 | AC-068 | |
| P-065 | AC-072 | |
| P-066 | AC-073 | |
| T-001, T-003 | AC-054 | |
| T-002 | AC-051, AC-071 | |
| T-004 | AC-050, AC-053 | |
| T-005 | AC-050 | |
| T-006 | AC-052 | |
| T-010 | AC-002 | |
| T-023 | AC-070 | |
| T-011 | AC-064 | partial — the non-commercial posture itself is still untestable |
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
