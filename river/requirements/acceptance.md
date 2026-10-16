# Acceptance criteria

For each requirement, a check that would fail if the requirement were violated.

> Per [README.md](../README.md#acceptance-criteria), a criterion you have not seen go red is not
> yet a criterion — it is an intention. Before any status becomes `verified`, break the thing it
> checks and confirm it fails.

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

### everyRequirementCitesGoal

**Every requirement cites a declared goal, and every goal is cited**

**Verifies:** the structural rules in [README.md](../README.md) · **Level:** spec · **Method:**
script over `goals.md`, `product.md`, `tech.md`
**Run:** `python3 river/check/goals.py`
**Status: verified — run 2026-09-21, passing, and watched failing.** Three rules: every active
requirement has a `Goals:` field, every cited goal is declared, every declared goal is cited. It
found four uncited requirements on first run (tech.servesOnLocalhost, tech.identifiedClient, tech.localNowPublicLater, tech.bundleIsNormative), which is what prompted
goals.goodDataCitizen. Exercised by deleting tech.servesOnLocalhost's `Goals:` line: exit 1, `requirements with no Goals: field ->
tech.servesOnLocalhost`.

**This check cannot tell whether a requirement genuinely serves the goal it names.** Goals are not
measurable by construction. That judgment belongs to Gate 1 and to review, and a green run here
looks exactly like a validated one.

### legibleDistinctColors

**Generated colors are legible and mutually distinguishable**

**Verifies:** tech.badgeContrast, product.generatedBadgeColors, `wcag/1.4.6` · **Level:** spec and code · **Method:** pure functions
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
capacity argument behind product.generatedBadgeColors depends on white text being used where it helps, and a generator
that quietly reverted to black-only would give that up invisibly. **Alternation**: an odd-length
light/dark cycle wraps onto itself, so one pair of hue-adjacent slots shares a lightness — which
is what left two pale Passeriformes families at ΔE 4.4. Both have been watched failing.
it fail by lowering a generated color's luminance.

### vendoredHashesMatch

**Every vendored source matches its recorded hash**

**Verifies:** tech.recordBeforeConsuming, goals.goodDataCitizen · **Level:** spec · **Method:** hash each committed input and compare
against `river/requirements/sources/*/PROVENANCE.md`
**Run:** `python3 river/check/pins.py`
**Status: verified — run 2026-09-21, passing** for `avonet` and `fox-wingspan`, **and watched
failing**: zeroing AVONET's recorded hash produced exit 1 and a recorded-vs-actual diff.
**Scope exception:** `wiley` is quoted, not vendored, so it has no hash to check — see
[requirements.md](requirements.md#wiley). `aba` is fetched at build and checked by sortsAscending instead.

### decisionReferencesExist

**Decision records reference requirements that exist**

**Verifies:** process integrity · **Level:** spec · **Method:** extract every qualified reference
(`product.x`, `tech.x`, `goals.x`, `decisions.x`, `acceptance.x`) and every anchored link, and
confirm each resolves
**Run:** `python3 river/check/refs.py`
**Status: verified — run 2026-09-21, passing, and watched failing** — appending a reference to a
requirement id in the P-9xx range, which nothing defines, produced exit 1.

*Known limitation, found by this check catching its own documentation:* the scanner cannot
distinguish a citation from a mention, so writing an example id in prose trips it. Worth
tolerating — the failure is loud, immediate, and safer than a checker that guesses intent.

---

## Dataset and ingest

### scopeMatchesStatusTable

**Scope is what the status table says recurs**

**Verifies:** product.scopeIsRecurringAvifauna · **Level:** code · **Method:** unit test over ingest output
**Assertion:** the dataset contains every species `data/status.tsv` marks in scope (`established` or
`rare`, plus the declared exceptions of product.declaredExceptions), less those AviList has no concept for and the four
species with no AVONET measurement — **1,082** species today — derived from the inputs rather than
hardcoded from a run.
**Status: verified 2026-10-05 — passing (`TestScopeIsWhatTheStatusTableSaysRecurs`).** Previously
verified 2026-09-22 against ABA checklist 8.19 (841 species); rewritten for
[decisions.scopeFromRecurrence](decisions.md#scopefromrecurrence).

### noRarityCodes

**Rarity codes never reach the dataset**

**Verifies:** product.scopeIsRecurringAvifauna · **Level:** code · **Method:** inspect the serialized record's field names
for anything carrying an ABA rarity code
**Status: verified 2026-09-22 — no rarity field appears in the serialized record.** Retained as a
regression guard after [decisions.scopeFromRecurrence](decisions.md#scopefromrecurrence) removed the ABA checklist.

### allFiveTraits

**Every record carries all five AVONET traits**

**Verifies:** product.traitsFromAvonet, product.traitsPerSpecies · **Level:** code · **Method:** property test over every record
**Assertion:** culmen, tail, wing chord, secondary and mass are all present and positive.
**Status: planned.** Trait coverage across AVONET's three sheets brings **1,082** of the 1,086
named in-scope species to a complete row (the 4 unmatched species are named in `unmeasurableNames`),
so any failure on a built record means a join or parse bug rather than missing data.

### namesNeverDefaulted

**Names and families are populated from AviList, never defaulted**

**Verifies:** product.namesFromAvilist · **Level:** code · **Method:** property test over the built dataset

**Assertions, names:** the **world** name is non-empty for every species; the **North American**
name is non-empty *after* the documented fallback, so a species with no Clements concept shows a
bird rather than a blank; family scientific and English names are non-empty; and **no record's
name equals its scientific name** — the tell of a silent fallback, which an earlier design
shipped for every one of its species.

**Assertions, the code:** exactly **three** species carry no eBird code, and they are **named,
not counted** — *Oxyura jamaicensis*, *Pterodroma cervicalis*, *Tyto furcata*, each a taxon
AviList recognizes and Clements has no 1:1 counterpart for. A fourth appearing means the join
moved, which a percentage threshold would have absorbed in silence.

**Status: verified 2026-10-02 — passing, and watched failing** by blanking the world name, which
leaves the fallback with nothing to fall back to; and by adding a fourth species to the uncoded
set, which the named list catches where a percentage threshold would have absorbed it.
Previously verified 2026-09-22 against the ABA checklist; rewritten for
[decisions.avilistReplacesAba](decisions.md#avilistreplacesaba).

### stableSpeciesId

**A species id is unique, stable and fit for a URL**

**Verifies:** product.englishNameChoice · **Level:** code · **Method:** property test over the built dataset, plus
assertions on the served routes

**Assertions, uniqueness:** no two species share an id. This is not a formality — AviList lumps
two BirdView species into taxa already in scope, and had they been lumped in place rather than
dropped, two birds would have shared one URL.

**Assertions, URL fitness:** every id survives `url.PathEscape` unchanged and none is empty, so
no species needs encoding to be linked.

**Assertions, the fallback:** the three species with no eBird code carry a scientific-name slug
instead, and are reachable at it.

**Assertions, the routes:** `/species/<code>` resolves, and the comparison view accepts codes in
`species=` and renders those birds in that order.

**Status: verified 2026-10-02 — passing, and watched failing** by mapping both lumped species
onto their AviList taxon instead of dropping them, which collides *Vireo swainsoni* onto
`warvir`, already held by the Eastern Warbling-Vireo; and by blanking the fallback so the three
uncoded species get an empty id.

**The ingest asserts its own drop list rather than reporting it.** `taxonomy.NotInAviList` names
the two species that may legitimately fail the AviList join, and the build fails if the set that
actually failed differs — in either direction. A third species going quiet would otherwise be a
silent deletion, and a species regaining a concept would leave a dead entry nobody noticed.

### nameLocaleToggle

**The name locale is a toggle, and only names move**

**Verifies:** product.englishNameChoice · **Level:** code · **Method:** assertions on the rendered pages under both
locales

**Assertions, both names are served:** with no parameter the page shows the **North American**
name; with `names=world` the same URL shows the **world** name. The fixture gives every species a
world name that differs from its North American one, so a toggle that silently serves the same
string either way cannot pass.

**Assertions, the id does not move:** the species code and the URL are **identical under both
locales**. This is the property the whole change rests on, and the one an alpha code could not
have had.

**Assertions, the control:** the index offers both locales, marks the active one, and the links
preserve the rest of the view state.

**Assertions, the fallback:** a species with no North American name renders its world name under
either locale, rather than an empty header.

**Status: verified 2026-10-02 — passing, and watched failing** by serving the North American name
for both locales; by dropping the `names` parameter from the link builder, so the toggle renders
and does nothing; by hardcoding the on-state; and by removing the fallback, which blanks five
headers under the default locale.

### joinCoversEverySpecies

**The join covers every species, and the direct-match rate is monitored**

**Verifies:** product.joinOnBinomial · **Level:** code · **Method:** unit test over ingest
**Assertions:** (a) after the mapping table is applied, every in-scope species (except the 4 named
in `unmeasurableNames`) resolves to a trait row; (b) `TestNearMatchesAreExactlyTheMappedOnes`
verifies that the 9 near-matches in the **1,082**-species dataset match `taxonomy.NearMatches`.
**Status: planned.** 1,073 measured + 9 near-match = 1,082 species (earlier 97.38%, 819 of 841, was
measured against `AVONET2_eBird` on 2026-09-21 before [decisions.scopeFromRecurrence](decisions.md#scopefromrecurrence)). It is a
regression guard, not a target: a drop means a taxonomy moved and the mapping table needs review.

### unmatchedFailsBuild

**Unmatched species fail the build above a threshold**

**Verifies:** product.unmatchedSpeciesReported · **Level:** code · **Method:** unit test with a fixture containing a bogus
binomial
**Assertion:** ingest exits non-zero and names the unmatched species.
**Status: planned.** Watch it fail by raising the threshold until the bogus fixture passes.

---

## Ordering and sectioning

### sortsAscending

**Sorting is ascending by the selected metric**

**Verifies:** product.oneSelectedMetric, product.ascendingOrder · **Level:** code · **Method:** property test for each of the three
metrics
**Status: verified 2026-09-22 — ascending for all three metrics.**

### sectionBoundariesObeyRule

**Section boundaries obey the N% rule**

**Verifies:** product.sectionBreakRule, product.sectionThresholdDefault · **Level:** code · **Method:** property test over generated inputs
**Assertions:** for every section, `max / min < 1 + N/100`; and the first member of each section
is `>= (1 + N/100) x` the **minimum** of the previous section. Both hold for N across its full
adjustable range, and for N = 15 by default.
**Status: verified 2026-09-22 — property test over 1,500 random datasets at five step values, and **watched failing**: keying the break on the previous *maximum* instead of the minimum let a section span 7% at a 5% step.**
rule is stated against the section *minimum*, not a running anchor — a plausible off-by-one
implementation using the previous *maximum* passes casual inspection and fails this.

### sectionHeadersMatch

**Section headers match their contents**

**Verifies:** product.sectionHeadersStateRange · **Level:** code · **Method:** unit test
**Assertion:** each header's stated range equals the min and max of its members in the selected
metric, and its count equals the number of members.
**Status: verified 2026-09-22 — header ranges match their contents.**

---

## Diagram

### segmentsSumExactly

**Segments sum exactly to total length**

**Verifies:** product.segmentsSumToLength · **Level:** code · **Method:** property test over every record
**Assertion:** the five **serialized** segment values sum **exactly** to the serialized total
length — not within a tolerance.
**Status: verified 2026-09-22 — passing, and watched failing: rounding each segment independently instead of balancing broke the sum for both the property test and the real dataset, by 0.1 mm.**
rather than round each segment independently. An earlier design rounded independently and broke
this for 27.6% of its species by up to 0.1 mm; the drift was invisible on screen and made the
stated invariant false.

### sharedScale

**Scale is shared within a section, and globally in global mode**

**Verifies:** product.scaleToggle · **Level:** code · **Method:** unit test over computed geometry
**Assertions:** in per-section mode every tile in a section has an identical scale factor, and the
factor differs between at least two sections in a representative dataset; in global mode one
factor applies to every tile in the filtered set; **per-section is the default**, global is
reachable by `?scale=global`, and per-section is listed first in the controls.
**Status: verified 2026-09-22 — passing, and **watched failing**: scaling each bird to its own length was caught by the direct px/mm comparison. Worth noting that it was *not* caught by the first two assertions written for this criterion, because the wingspan constraint happened to preserve size ordering; comparing the scale factor itself is the only form that cannot pass by accident.**
implementation satisfies "identical within a section" trivially when sections hold one bird.

The default moved from global to per-section in [decisions.perSectionDefault](decisions.md#persectiondefault), and the tests asserting the
old default were inverted rather than deleted. Flipping it exposed a real bug: `parseOptions`
recognised only the **non**-default value, so the toggle stopped working entirely the moment the
default changed. Written that way the parser is correct exactly once.

### crossbarToScale

**The wing crossbar is to scale with the vertical axis**

**Verifies:** product.crossbarAtVerticalScale · **Level:** code · **Method:** property test over computed geometry
**Assertion:** `crossbarLength / bodyLength == wingspan / totalLength` for every tile, within
floating-point tolerance.
**Status: verified 2026-09-22 — passing, and **watched failing**: restoring the earlier design's formula, with its extra `availableH × scaleY` factor, broke the ratio for every species.**
factor, inverting the relationship so that larger birds drew *shorter* crossbars — 300 of 3,213
species floored out, and they were the 300 largest. This assertion catches that class exactly.

### diagramPartsInOrder

**The diagram has the specified parts in the specified order**

**Verifies:** product.segmentComposition · **Level:** code · **Method:** unit test over rendered output
**Assertion:** five vertical segments top to bottom — bill, head, neck, body, tail — plus one
crossbar; the tail is rendered with a stroke style distinct from the body.
**Status: verified 2026-09-22 — coordinates are monotonic and the head oval is centered between its tangents.**

### proportionsDifferentiate

**Proportions differentiate species as the requirement claims**

**Verifies:** product.proportionDifferentiates · **Level:** code · **Method:** ordering assertions over a small set of
species pairs
**Assertion:** a dowitcher's culmen fraction exceeds a robin's; a heron's neck fraction exceeds a
sparrow's; an accipiter's tail fraction exceeds a duck's.
**Status: planned.** The expected orderings come from ornithology, not from running the code —
which is what makes this a check rather than a characterization. If an ordering fails, the model
is wrong, not the test.

### neckLengthSerialized

**The neck length is serialized with its model named**

**Verifies:** product.neckModelPerFamily, product.provenancePerMetric · **Level:** code · **Method:** property test over every record
**Assertions:** every record carries a neck length; its provenance `kind` is `modeled`; its
`model` is non-empty and **names both the per-family ratio function and the numeric ratio applied
to that species' family**; and the ratio named matches the table entry for that family.
**Status: verified 2026-09-22 — every neck model names its family and a ratio.**
asserted the opposite, that no neck length was ever serialized. The last assertion is the one with
teeth: a `model` string that names a function but not the family's actual value would satisfy a
looser check while telling the reader nothing they could verify.

---

## Provenance

### provenanceWellFormed

**Every presented metric has a well-formed provenance record**

**Verifies:** product.provenancePerMetric · **Level:** code · **Method:** schema validation plus a property test over
every record
**Assertions:** every metric carries `kind` in `{measured, near-match, modeled}`; `source` is
non-empty; `sourceTaxon` is non-empty for `measured` and `near-match`; `model` is non-empty for
`modeled`; `sourceTaxon` differs from the species' own name whenever `kind` is `near-match`.
**Status: verified 2026-09-22 — passing, and watched failing: blanking the neck model string produced a violation for every species.**
something rather than being a label anyone can attach.

### provenanceMatchesSources

**Provenance matches the sources actually used**

**Verifies:** product.lengthAndWingspanSources, product.provenancePerMetric · **Level:** code · **Method:** unit test over ingest output
**Assertions:** every length has `kind = modeled`; exactly the species with an empirical Fox value
have `kind = measured` for wingspan; the six species named in
[decisions.abaSpeciesWithoutAvonet](decisions.md#abaspecieswithoutavonet) have `kind = near-match` for their AVONET-derived traits.
**Status: verified 2026-09-22 — 301 measured wingspans, matching the count derived from the sources before this code existed. Watched failing: removing the near-match inheritance reported 308, i.e. seven borrowed values mislabeled as measured.**
have an empirical wingspan (`needs_prediction = 0`), 510 have only a predicted value which this
project excludes, and 30 have no row at all — every one of them a Galliform, an order the Fox
dataset omits entirely. These counts come from the sources, not from our output.

### tileMarksNonMeasured

**The tile marks every non-measured metric**

**Verifies:** product.tileMarksModeled · **Level:** code · **Method:** rendering test asserting a marker is present
for each metric whose `kind` is not `measured`, and absent otherwise
**Status: planned.**

### detailShowsProvenance

**The detail view shows every metric and segment with provenance**

**Verifies:** product.metricProvenanceExplained, product.tileOpensDetail · **Level:** code · **Method:** rendering test
**Assertions:** the detail view shows all three main metrics **and all five diagram segments**
with their values; for each it names the source; for a `near-match` it names the taxon actually
measured; for a `modeled` value it names the model. The detail view reproduces the tile's header,
diagram and main metrics — it is the same layout enlarged, not a different screen.
**Status: verified 2026-09-22 — all eight metrics appear with provenance, and **watched failing**: removing the neck disclosure was caught.**

### totalIsSumOfSegments

**Total length is reported as the sum of its segments**

**Verifies:** product.lengthAndWingspanSources, product.provenancePerMetric, product.segmentsSumToLength · **Level:** code · **Method:** property test
**Assertions:** total length's `kind` is `modeled`; its `model` states that it is the sum of the
five segments; and the value equals that sum exactly (the same invariant segmentsSumExactly asserts from the
geometry side).
**Status: verified 2026-09-22 — total length names itself a sum of measured and modeled parts.**
visible that a *measured* bill sits inside a *modeled* total; this criterion makes the output
explain that rather than leaving the reader to infer it.

### wingspanModelReconstructsFox

**The fitted wingspan model reconstructs Fox's published one**

**Verifies:** product.lengthAndWingspanSources, product.provenancePerMetric · **Level:** code · **Method:** fit the model, then compare its
estimate against `wingspan_predicted_*` for every species Fox predicted
**Assertions:** mean disagreement ≤ 2%; no more than 2% of species disagree by over 10%.
**Status: verified 2026-09-22 — passing at 1.32% mean, 1.4% over 10% across 9,194 species.**
The thresholds come from the pre-implementation comparison (mean 1.01%, p99 4.15%), not from
running this code. Its value is as a tie to an external published reference: if the fitting
regresses, the dataset does not quietly get worse, the check goes red.

---

## Technical

### nothingRestrictedTracked

**Nothing restricted is tracked by version control**

**Verifies:** tech.fetchRestrictedInputs, tech.datasetNeverCommitted, goals.goodDataCitizen · **Level:** code · **Method:** `git ls-files` assertion in CI
**Assertion:** no tracked path lies under `fetched/`, and the built dataset is not tracked.
**Status: planned, and it is the highest-value check in this document.** It is the only mechanical
guard on the project's licensing position; everything else protecting tech.fetchRestrictedInputs is a convention
someone can forget. Watch it fail by `git add -f` on a fetched file.

### noServerNetworkRequest

**The server makes no network request**

**Verifies:** tech.selfContainedServer · **Level:** code · **Method:** rendering test scanning for external
**subresources** — any `src`, and `href` on a `<link>`
**Status: verified 2026-09-22 — passing, and **watched failing**: adding the old design's `cdn.tailwindcss.com` script tag was caught. The check scans rendered HTML for off-host `src`/`href`, which is where tech.selfContainedServer actually bites — a self-contained binary serving a page that fetches a CDN stylesheet is not self-contained.**

### outboundRequestsIdentified

**Outbound requests identify the client**

**Verifies:** tech.identifiedClient, goals.goodDataCitizen · **Level:** code · **Method:** unit test against an `httptest` server
**Assertion:** the `User-Agent` header is present, non-empty, and is not Go's default.
**Status: planned.**

### abaFileHashMatches

**The fetched ABA file matches its recorded hash**

**Verifies:** tech.fetchRestrictedInputs, product.scopeIsRecurringAvifauna · **Level:** code · **Method:** the ingest verifies the hash after
download and fails if it differs
**Assertion:** SHA-256 `9c9539b7…6331d3`, verified against the real file on 2026-09-21.
**Status: planned.**

### bothCommandsBuild

**Both commands build and the server serves**

**Verifies:** tech.twoCommandsOneModule, tech.servesOnLocalhost · **Level:** code · **Method:** covered by the standing checks plus an
integration test that starts the server and requests `/`
**Status: verified 2026-09-22 — every route serves, unknown species 404s.**

### tileCarriesHeaderAndMetrics

**The tile carries the header, diagram and main metrics**

**Verifies:** product.tileContents · **Level:** code · **Method:** rendering test
**Assertions:** every tile shows the common name with its banding code in parentheses, the
scientific name, a family badge naming both the scientific and common family name, an SVG diagram,
and the three main metrics **named in full** — `Length`, `Wingspan`, `Mass`, not `L`, `W`, `M`
([decisions.massCubeOnTiles](decisions.md#masscubeontiles)).
**Status: verified 2026-10-01 — passing, and watched failing** by restoring the initials.
Spelling them out also makes the assertion worth something: `"M "` matched almost any sentence on
the page, so the old check passed whether the tile carried a mass or not.

### familyFilterSelectsAndClears

**The family filter selects and clears**

**Verifies:** product.familyFilter · **Level:** code · **Method:** rendering test over `?family=`
**Assertions:** filtering to one family shows its species and excludes others; no filter shows
all; every entry is labeled with its scientific name and its common name; **no entry carries a
species count**, and a single count sits above the list instead; the selector is grouped by
taxonomic order; groups appear in checklist sequence rather than
alphabetically; each family sits inside its own order's group; each order is a working toggle that
selects or clears **every** family in it, labeled with its common name; and **All** toggles the
same way over the whole set, clearing the selection when untoggled.
**Status: verified 2026-09-22 — passing, and watched failing** by alphabetizing the family list,
and separately by making the order toggle add only its first family. **Counts moved above the
list 2026-10-02** ([product.familyFilter](sources/product/filtering.md#familyfilter)), watched failing by putting a count back on the
chips — which is the direction the mistake would come from, since the count reads as helpful
right up until a second filter makes it false.

Two lessons are recorded in the test itself. An earlier version asserted that order labels were
*inert text*; when the requirement inverted to make them toggles, that check kept passing, because
it searched for `<a class="ohdr` and the markup had become `<a class="btn ohdr`. And the
select-the-whole-group behavior could not be tested at all until the fixture gained an order
holding **two** families — with one family per order, "select the group" and "select the first
family" are the same operation.

The All control needed the same care for a different reason: "nothing selected" and "everything
selected" both display all 841 species, so no assertion on the species count can tell them apart.
The check reads the control's own state and the orders' states instead.

### unitToggleSwitches

**The unit toggle switches imperial and metric**

**Verifies:** product.imperialByDefault · **Level:** code · **Method:** rendering test over the tile footer, plus a
unit test pinning the conversions
**Assertions:** the tile footer carries two length values and one mass value in the selected
system's units, and differs between the two systems; mass converts as `1 oz = 28.349523125 g`.
**Status: verified 2026-09-22 — passing, and watched failing** on both a hardcoded-inches footer
and a kilograms-under-imperial conversion.

The first version of this criterion asserted only that a centimeter value appeared *somewhere* on
the page, and passed for weeks while every tile was hardcoded to inches: the section headers alone
satisfied it. Counting units inside the footer is what makes it bite — a footer rendering `3.9"`
contains no `" in"` token, so an exclusion list misses it entirely.

### detailDrawsFullScale

**The detail view draws its bird full scale**

**Verifies:** product.tileOpensDetail · **Level:** code · **Method:** rendering test over the smallest fixture
species, with and without an explicit `?scale=`
**Assertion:** the diagram fills the detail view's height regardless of the index's scale mode.
**Status: verified 2026-09-22 — passing, and watched failing** by making the detail view adopt the
dataset-wide reference, which inset the hummingbird by 128 px.

### pageExplainsControls

**The page explains every control and the diagram**

**Verifies:** product.pageExplainsItself · **Level:** code · **Method:** rendering test over the help block
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
([decisions.perSectionDefault](decisions.md#persectiondefault)), which closes a rot the old one had: it described the tail as "dashed, to
tell it from the body", and that stopped being true when decisions.drawingCarriesProvenance gave dash a different meaning.
Nothing caught it, because prose and stylesheet had no connection. A rendered diagram cannot drift
from the renderer.

This criterion is tied to the controls that exist, so adding a control without documenting it
fails here. That is deliberate: help text rots silently otherwise.

### badgeTextInheritsColor

**Text on a generated background inherits that background's text color**

**Verifies:** product.generatedBadgeColors, tech.badgeContrast · **Level:** code · **Method:** structural check over `styles.css`
**Assertions:** no rule targeting a descendant of `.chip` or `.badge` sets its own `color`, and no
such rule dims text with `opacity`.

**Status: verified 2026-09-22 — passing, and watched failing** by restoring the fixed grey.

Chips and badges carry a generated background paired with a black-or-white text color chosen so
the pair clears 7:1. The stylesheet set the family common name to a fixed grey, which ignores
which half of the palette a chip landed on and went unreadable on both — the common name vanished
on roughly a dozen chips. Hierarchy has to come from weight and size; dimming undoes the very
ratio the background was selected for. Checking CSS is crude, but the defect is invisible to a
rendering test and this catches it exactly.

### emptyStateGuides

**The empty state guides rather than blames**

**Verifies:** product.helpfulEmptyState, product.familyFilter · **Level:** code · **Method:** rendering test after clearing every
selection
**Assertions, nothing selected:** no species are displayed; the page says how to get them back; it
does **not** say that nothing matched; the count reports zero; and All is no longer marked
selected.

**Assertions, filters named while the list is full:** the **count line names each active filter
and what it is set to**, whenever any is set — not only once nothing matches. A filter that is
not narrowing anything, every family selected being the arrival state, is **absent**, since
naming it would present the default as a choice the reader made.

**Assertions, nothing matched:** the message says the intersection is empty and offers a link
that clears the filters. It does **not** repeat the filter names, which the count line
immediately above already carries; a test pins that they appear once on the page, not twice.

**Status: the count-line half verified 2026-10-03 — passing, and watched failing** by removing
the filter list from the count line, and by leaving it in the empty state as well, which puts
four filter chips on a page that should carry two. The rendering gap recorded below is now
closed: the region filter made an empty intersection reachable through the UI, and
`TestFiltersAreNamedOnTheCountLine` exercises it end to end.

**Status: verified 2026-09-22 — passing, and watched failing** both by treating an emptied
selection as "no filter" (the previous behavior, which showed every bird) and by reverting the
message to "No species match that filter". **The filter description was added 2026-10-02**
([product.helpfulEmptyState](sources/product/presentation.md#helpfulemptystate)) and watched failing by describing a filter that narrows nothing, and
by making the nothing-selected branch claim a failed match.

**The description is tested through the function, not the page, and that is a gap.** With only a
family filter the empty-intersection branch is **unreachable through the UI** — one filter cannot
contradict itself. It becomes reachable when the region filter of [product.regionAndMonthFilter](sources/product/filtering.md#regionandmonthfilter)
lands, and this should become a rendering test then. Recorded rather than left for someone to
discover the branch was never exercised end to end.

### scrollScriptScoped

**The scroll script is embedded and scoped to the index**

**Verifies:** product.togglesKeepPlace, tech.selfContainedServer · **Level:** code · **Method:** rendering test
**Assertions:** the index carries the script inline, never via `src`; it is scoped to `/` so a
detail page cannot consume the saved position; the detail view ships no script; the controls carry
an `id` so `#controls` works as a fallback target; both panels carry ids so their collapsed state
can be keyed on; every panel defaults to **closed** and carries a one-line description in its
summary ([product.collapsibleControls](sources/product/presentation.md#collapsiblecontrols)); and panel state is restored **before** the scroll.
**Status: verified 2026-09-22, amended 2026-10-04 — passing, and watched failing** by
externalizing the script, by removing its path guard, by dropping a panel's id, and by reordering
the restores so the scroll runs first. The default was *open* until product.collapsibleControls changed it; the
amended check was watched failing by restoring `open` to the markup and by deleting a summary
description.

**This does not check that scrolling works.** That needs a browser, which this project does not
have. The criterion covers presence, self-containment and scoping only, and the behavior itself is
a declared gap.

### controlsMatchServer

**The controls offer what the server accepts, default first**

**Verifies:** product.sectionThresholdDefault, product.scaleToggle · **Level:** code · **Method:** rendering test scoped to the controls
panel, plus round-tripping each offered step through `parseOptions`
**Assertions:** Global is listed before Per section; every offered section step (5, 10, 15, 20,
30) is parsed as itself rather than falling back to the default.
**Status: verified 2026-09-22 — passing, and watched failing** by reversing the scale pair and by
offering a step the server rejects.

The scoping matters and was got wrong first. Searching the whole page for the scale labels matched
the **Presentation panel**, which named both modes in the same order, so the check passed with the
buttons reversed. It now reads only the controls panel.

### attributionComplete

**Attribution is complete, and claims nothing unearned**

**Verifies:** tech.localNowPublicLater, goals.goodDataCitizen · **Level:** code · **Method:** rendering test over the footer
**Assertions:** each CC BY source is credited with creator, a link to its deposit and a link to the
license — **AVONET, Fox and AviList by name and DOI**; the ABA checklist is credited for the
species list and *only* that; the footer discloses that lengths are modeled; the detail view
carries credit too, being a shareable URL of its own; and **no page claims a permission the
project does not hold**.
**Status: verified 2026-09-22 — passing, and watched failing** by stripping the license links and
by reinstating a "used by permission" claim. **AviList added 2026-10-02**
([decisions.avilistReplacesAba](decisions.md#avilistreplacesaba)) and watched failing by deleting its credit while leaving its names
on every tile — which is the shape the obligation actually has: CC BY permits the use *because*
of the attribution, so an unasserted credit is a licence breach waiting to happen.

The second half exists because the first version of this footer said the ABA checklist was "used by
permission". It is not — that request has not been sent. Claiming a right you have not been granted
is worse than claiming none, because it misleads the rights holder's own readers. Remove that
assertion only when a written grant exists and is recorded in [requirements.md](requirements.md).

### craniumVaries

**The cranium varies with the bird**

**Verifies:** product.craniumFromMass, product.provenancePerMetric · **Level:** code · **Method:** unit test on the model plus a
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

### legsFromTarsusWithdrawn

**Legs are modeled from the tarsus** *(withdrawn)*

**Withdrawn 2026-09-23 by [decisions.inFlightNoLegs](decisions.md#inflightnolegs).** The requirement it verified no longer
exists: the diagram is a bird in flight and has no legs.

Kept as a record rather than deleted, because the criterion is why the removal was cheap. It
asserted that some species' feet fall below the tail tip **and** that only wader families may do
it, and it checked the leg model against the whole dataset rather than a fixture. Those checks are
what made it obvious that the modeled leg underpredicts the flight projection — 14 of 110 species
in the trailing-leg families clearing the tail against a true figure far higher, and a White Ibis
coming out negative. A criterion that only confirmed the code did what the code did would have
said nothing when the premise changed.

### foldedNeckSameStroke

**A folded neck is the same stroke in less height**

**Verifies:** product.segmentComposition, product.segmentsSumToLength, product.tileOpensDetail · **Level:** code · **Method:** unit tests over the geometry plus
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

### inaturalistLinkConfidentOnly

**The iNaturalist link appears only for a confident match**

**Verifies:** product.inaturalistLink, tech.pacedInaturalistMapping · **Level:** code · **Method:** unit tests over the matcher and the
mapping table, plus assertions on the rendered detail page; the resolver is exercised against an
`httptest` server, never the live API

**Assertions, matching:** a taxon whose species epithet differs is rejected; a taxon whose common
name disagrees with ABA's is rejected even when the scientific name matches exactly; and each of
the five forms of agreement is accepted — a Hawaiian okina against its absence, an ABA parenthetical
alternate, British against American spelling, a hyphen against a space, and a name iNaturalist holds
but does not prefer. A **different genus** is accepted only alongside an agreeing common name, and
refused without one ([decisions.fullInaturalistCoverage](decisions.md#fullinaturalistcoverage)). A **gendered epithet** is matched across a change of
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
iNaturalist image — the existing noServerNetworkRequest subresource scan already covers the mechanism, and this
adds the intent.

**Assertions, courtesy:** the resolver sends a descriptive `User-Agent` (tech.identifiedClient, `inatapi.UserAgent`) and
waits at least the configured interval between requests (tech.pacedInaturalistMapping, `inatapi.RequestRate`), both observed
against the test server.

**Status: verified 2026-09-27 — passing, and watched failing** by: ignoring the common name;
emitting a search URL instead of nothing; removing the pacing; dropping the `User-Agent`; bypassing
the zero-ID guard; deleting a row from the committed table; and reverting the gender stemming.

Coverage is **841 of 841**, reached by widening the evidence rather than lowering the bar
([decisions.fullInaturalistCoverage](decisions.md#fullinaturalistcoverage)).

The confidence gate is the part worth testing hardest, because the failure it prevents is silent.
Matching on scientific name alone would find the right page nearly always, and the rare miss would
be a *plausible wrong bird* rather than a visible error — which is exactly the class of bug this
project keeps finding in its own tests.

### foldingFamiliesHaveNecks

**A family that folds its neck has a neck to fold**

**Verifies:** product.segmentComposition, product.neckModelPerFamily · **Level:** code · **Method:** unit test cross-checking the two neck
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

### noTextBelowContrastFloor

**No text or link falls below its contrast floor**

**Verifies:** tech.textContrastFloor, `wcag/1.4.3`, `wcag/1.4.6` · **Level:** code · **Method:** two checks parsing
the stylesheet **as embedded in the binary**, not off disk

**Assertions, measured:** every `color` a rule declares — with custom properties resolved — clears
**7:1** against both `--bg` and `--panel`. Rules that also set a `background` are excluded, being
self-contained and governed by tech.badgeContrast.

**Assertions, SVG text:** every `fill` a rule declares **on a text element** clears the same 7:1
against both surfaces. Which rules those are is derived rather than listed: a class token that any
rule gives a font property to is a text token, and any rule naming one is styling text. So
`.dlabel { font-size: 11px }` makes `dlabel` a text token, and the `fill` on `.dlabel.chord` is a
text color. A `fill` on a shape — the head oval, the mass cube — is untouched, being a graphical
object under WCAG 1.4.11's 3:1 instead.

Deriving it is the point. An explicit list of text selectors would need maintaining by the same
person who just added an unchecked one, which is the failure this assertion exists to fix.

**Assertions, structural:** the stylesheet carries top-level `a` and `a:visited` rules, so no link
can fall back to the browser default.

**Status: verified 2026-09-23 — passing, and watched failing** by removing the global link rules
(which reproduces the original defect exactly), by setting the visited color to the browser's
`#551A8B`, by picking a link color that clears AA at 4.90:1 but misses 7:1, and — after the floor
rose to 7:1 for everything ([decisions.oneContrastFloor](decisions.md#onecontrastfloor)) — by restoring the previous `--dim` and
`--accent`, which fail at 5.71:1 and 6.83:1 on a panel.

**SVG text assertion added 2026-10-02 — and it caught a live defect on its first run**, not a
planted one: `.dlabel.chord, .dlabel.span` was painting the wing chord and wingspan labels
`#0ea5e9`, 6.44:1 and 5.28:1. Watched failing three ways:

| breakage | what happened |
| --- | --- |
| restore `#0ea5e9` on the label fill | both surfaces flagged — this *is* the shipped bug |
| set a shape fill (`.mcface`) to near-black `#010203` | **not** flagged, correctly: a cube face is a graphical object, so the check is selective rather than indiscriminate |
| drop `font-size` from `.dlabel`, so `dlabel` stops being a text token | `checked == 0` fires and the test fails outright |

The third is the one worth having. Removing the font property makes the derivation cover nothing
at all, and without that guard the check would go quietly green while protecting zero labels —
the exact shape of failure this project keeps finding. A check that cannot report "I am no longer
checking anything" is not a check.

The structural assertion is the one that would have caught the bug. A measured check sees only
colors that were declared, and the defect was a color nobody declared — the link inherited
`#0000EE` at 1.90:1. Measuring what is there cannot find what is missing.

Writing the check found a bug in the check: the selector capture was swallowing the comment
preceding each rule, so any rule documented with a sentence containing the word "a" was read as a
link selector and held to 7:1. `.sn` failed that way on the first run. Comments are stripped before
parsing, and selectors are split into tokens rather than substring-matched.

### ownRenderableFavicon

**The favicon is BirdView's own, and is actually renderable**

**Verifies:** tech.selfContainedServer · **Level:** code · **Method:** assertions over the served asset and both pages

**Assertions:** `/static/favicon.svg` is served with status 200; it **parses as XML** and its root
element is `svg` with a `viewBox`; both the index and a detail page reference it; and
`/favicon.ico` returns 404 rather than the index page, which the catch-all route makes worth
pinning.

**The palette assertion moved to [faviconFromPalette](#faviconfrompalette)**, where it is derived from the stylesheet. It
lived here as a search for three hardcoded literals, one of which existed only in the favicon, so
it compared the icon against a copy of itself and could not fail. This criterion keeps what it can
actually see from the server: that the asset is served, parseable and referenced.

**Status: verified 2026-09-23 — passing, and watched failing** by removing the asset while leaving
the reference, by removing the reference while leaving the asset, by pointing the reference at a
CDN (which trips noServerNetworkRequest's existing scan, as intended), and by reintroducing the XML defect below.

No new requirement was added. tech.selfContainedServer already forbids fetching a subresource from another host and
noServerNetworkRequest already scans `<link href>` for it — confirmed here by watching a CDN favicon fail that
check. What noServerNetworkRequest cannot see is an icon referenced but not served, or served but referenced by
nothing, so this criterion joins the two ends.

**The format assertion is the part that earned its place.** Its first version searched the body for
the substring `<svg` and passed against a file no browser could read: the SVG's comment contained a
double hyphen, which XML forbids, so the icon would simply never have appeared. It now runs the
parser. A format check that does not parse is not a format check — the same lesson as noServerNetworkRequest's
subresource scan, one layer down.

### diagramDistinguishesModeled

**The diagram distinguishes measured from modeled, and says so**

**Verifies:** product.measuredVersusModeled, product.segmentComposition · **Level:** code · **Method:** geometry unit tests, a structural check
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
and tidying them into agreement makes short necks render solid (decisions.drawingCarriesProvenance).

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

### segmentLabelsDontCollide

**Every segment is labeled on the detail diagram, and no two labels collide**

**Verifies:** product.labeledDetailDiagram · **Level:** code · **Method:** unit tests over the label layout plus
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

**Assertions, completeness:** **every row of the provenance table is labelled in the diagram**,
with the same value, in both unit systems. Written against the table rather than a list of metric
names, so it maintains itself: adding a row without labelling it fails. Watched failing by
removing the mass label — which is the state this check was written to catch — and by labelling
mass with the wingspan's value.

**Assertions, the estimated marker:** a diagram label whose value is modeled carries the `~` **in
the emitted markup**, and one whose value is measured does not. Asserted on the text content, not
on the class attribute: the class is what the inert CSS rule keyed off, so a check that accepted it
would re-certify the original bug.

**Status: verified 2026-09-28 — passing, and watched failing** by anchoring every label at the same
y, by removing the minimum-gap pass, by dropping a segment from the list, by removing the reserved
gutter and by inflating it, by moving the wing labels back inside the centering group — which
reproduces the bug where the neck's leader struck through "Wing chord" and "Wingspan" — and, for
the marker, by reverting the `~` to the `.dlabel.est::after` rule it came from and by marking every
label estimated regardless of provenance.

**The marker assertion was added because the criterion claimed it and never checked it.** product.labeledDetailDiagram has
required the `~` since these labels were introduced in `a826b96`, which implemented it as
`.dlabel.est::after` — CSS generated content on an SVG `<text>`, which no browser renders. It was
found by screenshotting the same detail page twice, once with the rule and once without, and
getting byte-identical PNGs. So from `a826b96` until this criterion was extended, every modeled
value on the diagram was drawn exactly like a measured one, which is the one thing
[decisions.noOpenLengthAndWingspan](decisions.md#noopenlengthandwingspan) exists to prevent. The lesson is the one this file keeps relearning: the obvious
check would have asserted on `class="dlabel est"`, which was present, correct, and meaningless.

**The collision resolver fires on the comparison view, and until then fired nowhere.** It was
recorded here as dead on real data — 0 of 841 species needed it — because the detail view draws
every bird to fill its box, so no two segments are closer than the line height. That note
predicted "a larger font, a shorter box or a new segment would all wake it". What actually woke it
was [product.comparisonOneScale](sources/product/presentation.md#comparisononescale), which is none of the three: at a scale set by a larger bird, a
small one's five segments collapse into a few pixels and the resolver is the only reason its
labels can be read at all.

Left here as the worked example of why a branch that is dead *by the current numbers* is worth
keeping, where one dead *by construction* is not — the distinction decisions.inFlightNoLegs turned on. Deleting it
while it was unreachable would have meant rediscovering the need for it as a bug, the first time
two birds were drawn to one scale.

### totalLengthLabeled

**The detail diagram labels the whole bird's total length, under the tail**

**Verifies:** product.labeledDetailDiagram · **Level:** code · **Method:** geometry assertions over the whole fixture plus
assertions on the rendered detail page, the help legend and an index tile

**Assertions, placement:** the label is centred on the drawing axis (`text-anchor="middle"` at the
bird's centreline, not the box's — they differ, because the bird is laid out left of the label
gutter) and sits **below the tail tip**. It is horizontal: no `transform`/`rotate` on the element.
It has **no leader line**, so the leader count stays at one per segment.

**Assertions, room:** across **every species in the fixture**, the label's baseline plus its
descender stays inside the box. The binding case is a bird whose height fills the box — most of
them, since the detail view scales the bird to fit — so this would fail for the entire dataset if
the bottom strip were not reserved, and it is checked over all 841 rather than one bird because the
worst case is whichever species is tallest relative to its wingspan.

**Assertions, value:** the label states the **total length from the provenance table**, read out of
the rendered table rather than recomputed. The fixture includes a **folded-neck species**
(Ardeidae), where the drawn height is 6–9% short of the total length — so an implementation that
labeled the drawn extent instead of the datum produces a visibly different number and fails. This
is the assertion that would otherwise be vacuous: for every unfolded bird the two agree exactly.

**Assertions, provenance:** total length is modeled for all 841 species today
([decisions.ricklefsProhibited](decisions.md#ricklefsprohibited)), so the label carries the `~`. The check reads the provenance rather
than hardcoding it, so it keeps working — and starts asserting the opposite — if the Ricklefs ask
is ever granted and the value becomes measured.

**Assertions, scope:** the detail view and the help legend carry the label; an **index tile carries
none**, like every other annotation of product.labeledDetailDiagram.

**Status: verified 2026-10-01** — re-verified after mass joined total length under the tail.
The box-fit check now asserts the **last** label in the stack, since the strip must clear the
whole stack and not just its first line; watched failing by sizing the strip for one line, which
overran by 4px on every species. Originally **verified 2026-09-28 — passing, and watched
failing** by removing the label from the
template; by centring it on the box centre, which put it at x=494 against the bird's axis at x=334,
exactly half the gutter adrift; by placing it above the tail tip (y=254.3 against a tail tip at
269.3); by removing the reserved strip, which overran the box for **all five fixture species**, the
worst by 15 px; by labeling `TotalHeight()/PxPerMm` instead of the reported length, which the Great
Blue Heron caught at 78.7 cm against a table reading 86.0 cm; and by dropping the `~`.

**The folded-neck fixture was itself verified to be load-bearing, not decoration.** With the
`TotalHeight()` bug still in place, the species was swapped to the American Robin and the value
assertion **passed** — the drawn extent and the total length agree exactly for every unfolded bird,
so the check would have been vacuous on any of the other four fixtures. The test therefore carries
a guard that fails loudly if the fixture ever stops folding its neck, and that guard was watched
firing during this exercise.

### familyListWraps

**A long family list wraps inside its order's column**

**Verifies:** product.familyFilter · **Level:** code · **Method:** a structural check over the stylesheet and
assertions on the rendered markup

**Assertions, layout:** `.ogroup` is a **grid with exactly two tracks** — the order toggle, then
the families — and `.ofams`, the families' own box, **wraps**. One track would stack the two; a
third would leave a gap the chips could wrap into.

**Assertions, markup:** every order group contains **exactly one** `.ofams`, every family chip is
inside it, the order header is **outside** it, and the groups between them account for every
family chip on the page — so a chip cannot escape an order group entirely.

Both halves are required because either alone does nothing: the grid gives the families a column,
and the wrapper is what occupies it. A chip left as a direct child of `.ogroup` becomes a third
grid item and lands in a new implicit row back at the page margin, which is the original bug
reintroduced for one order.

**Status: verified 2026-09-28 — passing, and watched failing** by reverting `.ogroup` to the
wrapping flex row it was (the original bug); by collapsing it to one grid column; by removing
`flex-wrap` from `.ofams`; by deleting the `.ofams` rule; by removing the wrapper from the markup;
by leaving one chip outside it; and by moving the order header inside it.

**The first version of the markup check found the bug and reported the wrong thing.** It located
each group with a regex over balanced `</div>`s — but that nesting is precisely what the test
exists to verify, so removing the wrapper made the parse collapse two groups into one and the
failure read "only 1 order group in the fixture" rather than "this group has no family box". It
now splits on the group opener instead. A test whose *parser* depends on the structure under test
reports its own confusion instead of the defect.

### comparisonSharesScale

**The comparison view shares one scale, wraps, and explains itself once**

**Verifies:** product.comparisonOneScale · **Level:** code · **Method:** assertions on the rendered page, a geometry
check over the handler's own output, and a structural check over the stylesheet

**Assertions, shared scale:** every panel is laid out against **the same reference**, and that
reference is the **largest total length** in the set, not each bird's own. Checked by comparing
the drawn height of two birds of known different lengths: their heights are in the same ratio as
their total lengths, to within a rounding pixel. This is the assertion the view exists for, and
the one that fails if a panel is ever laid out against itself — which is what the detail view
does, and what this code is one argument away from doing by accident.

**Assertions, selection:** the species named in the URL all appear, **in the order given**; a
species named twice appears **once**; and a request naming none renders an explanatory page
rather than an empty one.

**Assertions, layout:** the panels sit in a container the stylesheet makes **wrap**, the diagram
key appears **exactly once** however many birds are shown, and **no provenance table** is emitted.

**Assertions, provenance:** every panel links to its own species detail view, and the estimated
marker still appears on modeled values — the compare view may omit the *explanation* (product.metricProvenanceExplained) but
not the *marking* (product.tileMarksModeled).

**Assertions, leaving:** the link back to the index carries **both the display options and the
species list** — the picks are live state and returning to add another bird must not discard them
([decisions.compareUrlContract](decisions.md#compareurlcontract)). **Clear** is the one link that drops them, and it must drop them
all. The two cannot be told apart by URL shape, since both point at `/?…`, so the markup has to
distinguish them and the check matches on that.

**Assertions, unrecognised ids:** an id matching no species is **dropped with the rest of the set
rendered**, is **named on the page**, and appears in **no link the page builds** — so the set
repairs itself on the next click. Asserted with a set mixing valid and invalid ids, because a set
of one invalid id cannot tell "dropped it" from "refused the page".

**Status: verified 2026-09-30 — passing, and watched failing** eleven ways: by laying each panel
out against its own length (every bird drew at 281.2px, filling its panel, and both scale
assertions fired); by sharing a scale set by **half** the largest, which the proportionality check
**passed** and only the largest-bird check caught, at 301.9px against 281.2px; by dropping the
dedupe; by sorting the panels by length; by silently skipping an unknown id; by rendering the key
once per bird (3 times for 3 birds); by dropping the detail links; by adding a provenance table;
by removing `flex-wrap`; by removing the panel's flex basis; and by gutting the empty state.

**The second breakage is why this criterion asserts on geometry at all.** A per-panel scale
produces a page that is correct in every particular a markup assertion can reach — right birds,
right order, right labels, right numbers, every panel neatly filled — and compares nothing. And a
*shared* scale set by the wrong bird passes a proportionality check on its own. Only the pair
together pins it: one constant across the set, and that constant equal to the scale the largest
bird gets on its own detail page.

**Both of those assertions reverse what this criterion said on 2026-09-30**, and were re-verified
on 2026-10-01 — **watched failing** by restoring the old `queryFor(r, "species")` so the back link
dropped the picks again, by restoring the 404 on a dead id, by dropping the dead id with no
notice, and by reverting the back link's wording. The reasoning for the reversal is in
[decisions.compareUrlContract](decisions.md#compareurlcontract). In short: the back link dropped the species list because the
index did not read it, and product.pickFromIndex makes the index read it; and the 404 was the honest option until
drop-and-name turned out to serve the same goal better. The original reasoning for each is kept
below rather than deleted, because both were right when written.

**The leaving assertion was added after the fact, from a user report.** The compare view reused
`queryFor` unchanged from the detail view, where it is correct because the species lives in the
path and never appears in the query. On a page that names its subjects *in* the query the same
call leaks them, and the leak is self-propagating rather than cosmetic. It was watched failing by
restoring the original call, and over-dropping to confirm the options half bites.

**The fixture is unfolded on purpose.** A folded neck draws shorter than the bird is long
(decisions.foldedNeckAsS), so its height is deliberately not proportional to its length and would make the scale
assertion wrong rather than strict. The test fails loudly if any of its three species ever starts
folding, so the constraint cannot be lost silently — the same guard totalLengthLabeled needed for the opposite
reason.

### pickModeSurvivesNavigation

**Pick mode is the species list, and the picks survive being navigated through**

**Verifies:** product.pickFromIndex, product.tileContents · **Level:** code · **Method:** assertions on the rendered index and
detail pages, plus a structural check over the stylesheet

**Assertions, the mode:** with no `species` parameter the index carries **no pick control on any
tile** and no bar. With one or more it carries a control on **every** tile and a bar. A **bare
`species=`** holds the mode open with nothing picked — the control and the bar are present and the
bar says nothing is picked yet.

**Assertions, the tile is not repurposed:** in **both** modes every tile's own link still points
at `/species/{id}`. This is product.tileOpensDetail, and it is the assertion that fails if the pick control is ever
implemented by changing the tile's href instead of sitting beside it.

**Assertions, picking:** a tile not yet picked offers a link that adds it; a tile already picked
renders as picked and offers a link that removes it; and the two are **inverses** — following add
then remove returns the original set, in the original order.

**Assertions, the round trip:** following a tile into the detail view and then its back link
returns an index whose picks are **unchanged**. This is the flow that does not work before this
change, and it is checked with **two or more picks** so that "preserved the set" cannot be
confused with "rebuilt a set of one".

**Assertions, href weight:** a tile href carries `species` and `units` and **not** `family`,
checked on an index that has an active family filter so the parameter would be there to leak. This
is a size bound — the regression is invisible until the page is slow — so it is asserted
directly rather than left to review.

**Assertions, the filter:** picking inside a filtered index neither resets nor ignores the filter,
and the bar lists a picked bird that the filter is **currently hiding**. A filter that appeared to
discard picks silently would be worse than one that refused to run.

**Assertions, the Controls entry point:** the toggle is a **labelled control set of its own** and
its button **states the mode** — `Off` while browsing, `On` while picking, matching its styling.
Both, because an unlabelled button beside the Section step percentages reads as a sixth step
option. The check finds the button by class rather than by text, since the name is now on the
label and matching the text would find that instead. The toggle is off while browsing and on
while picking; following it from browsing enters pick mode with nothing picked; and following
it while picking leaves by **the same URL Clear uses**, since they do the same thing and two
links that could drift would eventually disagree. Checked with another parameter in the query:
with only `species` present, Clear and a bare `/?` are the same string and the comparison cannot
tell a shared link from a coincidence.

**Assertions, not moving the reader:** every pick link ends at **its own tile's fragment**, and
the scroll script checks for a hash. Both halves, because either alone reintroduces the jump —
an anchor the script immediately overrides, or a stood-down script with nothing to stand down
for.

**Assertions, leaving:** **Clear** removes every `species` parameter and returns the page to
browsing mode, and **no other control does** — asserted by following every other link on the page
and confirming the picks survive it.

**Assertions, unrecognised ids:** a set mixing a good id with a dead one renders the good bird,
**names the dead one on the page**, and emits **no link containing it**. Asserted with a mixed set
because an all-invalid set cannot distinguish dropping from refusing.

**Status: verified 2026-10-01 — passing, and watched failing** ten ways: rendering the pick
control in browsing mode (5 controls on 5 tiles); implementing it by rewriting the tile's own
href, which made a tile link to `/?species=` instead of a detail view; prepending instead of
appending, so add-then-remove returned a different order; dropping `species` from the tile href,
which emptied the picks across the round trip; dropping `family` from the pick link, which reset
the filter the reader was picking from; filtering the bar's own list, which hid a pick the family
filter was hiding; building the index's control links with `species` stripped; stopping Clear
clearing; treating a bare `species=` as not-picking, which gave 0 controls for 5 tiles; and
keeping a dead id in the links the page builds.

**The href-weight assertion is in pickModeSurvivesNavigation's sibling test and was watched failing separately**, by
reverting the tile href to a bare path and by keeping `family` — the latter checked on an index
that has a filter active, because without one the parameter is not there to leak and the check
would pass against the bug.

**One of the ten breakages exposed a hole in the test rather than the code.** Rendering the pick
control in browsing mode produced a control with an *empty* href, and the matcher required
`href="([^"]+)"` — so it saw nothing and the test passed against a page that visibly had the
control on every tile. The pattern now accepts an empty href. A control with no destination is
still a control on the page, and a check that cannot see it is not checking for its absence.

**The Controls toggle and the anchor were added 2026-10-01 and watched failing five ways:** a
toggle that never reads as on; one that enters pick mode without the bare `species=`, so the mode
does not hold; one that leaves by a hardcoded URL rather than Clear's; pick links without their
fragment; and a scroll script that does not check for a hash.

**Following a pick link in a test needs the fragment stripped, as a browser strips it.**
`httptest.NewRequest` takes its target literally, so `#s-calypte_anna` lands inside `RawQuery`
and the final species value becomes `calypte_anna#s-calypte_anna` — unresolvable, and quietly
dropped as a dead id by the very rot handling above. The symptom was an unrelated criterion
failing; the cause was the harness exercising a request no browser can send. The shared `get`
helper now drops the fragment.

**The toggle-matches-Clear assertion was vacuous when first written.** Both were `/?` on a URL
whose only parameter was `species`, so a hardcoded `/?` passed as "the same link as Clear". The
fixture now carries `units` as well, which is the smallest change that makes the two
distinguishable.

**Self-healing was the half that silently failed first.** The notice naming a dropped id rendered
correctly while **970 links on the same page still carried it**, because the links were built from
the raw query and only the bar was built from the resolved set. The request's own query is now
rewritten before anything is built from it, which is why the criterion asserts on every `href` on
the page rather than on the bar.

**Vacuity traps already identified, recorded so the checks are not written into them:** searching
the whole page for the word "picked" matches the tile controls as well as the bar, so the bar must
be matched by its own container; the back link and Clear both point at `/?…` and differ only in
whether `species` survives, so they must be told apart by class and not by URL shape; and a
one-species fixture makes the round-trip assertion vacuous.

**The comparison view has two exits** (added 2026-10-09). Below *back to picking birds*, which
keeps the picks, *stop comparing* returns to the index with pick mode off and every other
carried setting kept. It is asserted to be **the bar's Clear URL exactly**, not merely a URL
without `species`: two exits from one mode, computed separately, will one day disagree about
what else they keep. Watched failing with a stop link that keeps the picks
(`still carries species, so it lands in pick mode`) and with one that resets to `/`
(`drops region=CA; leaving pick mode must not reset the view`).

### massDrawnAsWaterCube

**Mass is drawn as the cube of water that weighs the same**

**Verifies:** product.labeledDetailDiagram, product.pageExplainsItself · **Level:** code · **Method:** geometry assertions over every fixture
species, plus assertions on the rendered detail and comparison pages

**Assertions, arithmetic:** the square's side is `10 × ∛(mass in grams) × px/mm`, to within a
twentieth of a pixel, for every species. Derived from the definition of the gram — one gram of
water is one cubic centimetre — and never read off a run. This is the criterion that makes the
square's claim checkable: it is a real volume at the diagram's own scale, not a glyph on a scale
of its own, and only the arithmetic can tell those apart.

**Assertions, placement:** the front face's **base and left edge are identical in every
diagram** — that fixed corner is the whole point of the cube form, since it gives the eye a
common origin and lets two cubes be compared directly; a base that drifted with the bird would
cost that without changing any number. The projection has **non-zero depth**, or it is the square
this replaced. It stays inside the drawing area above and never reaches the bird's axis to the
right. Its label is centred under the face but **held clear of the left edge**: a Scissor-tailed
Flycatcher's 29 px cube renders a ~70 px label, which read "ass 1.4 oz" before it was clamped.

**Assertions, the strip below:** two things now hang under the drawing — the stack under the tail
and the mass label under the cube — and the reserved strip must clear **whichever reaches
lower**. Both are asserted; checking one leaves the other free to overrun.

**Assertions, provenance:** solid when the mass is measured, dashed when it is not, as product.measuredVersusModeled
requires of every other stroke.

**Assertions, rendering:** the square reaches both the detail page and the comparison page, the
diagram key appears **exactly once** on each however many birds are shown, and it **names the
square**. An unexplained shape beside a bird reads as part of the bird, so the key naming it is
not documentation of the feature, it is the feature.

**Assertions, the index tiles:** **every** tile carries a cube — counted against the number of
tiles, not merely found once — and **no tile carries a label of any kind**, the mass being named
in the footer instead ([decisions.massCubeOnTiles](decisions.md#masscubeontiles)). This reverses what this criterion asserted
when it was written, which was that tiles carry none; the reasoning is in decisions.massCubeOnTiles.

**Assertions, clearance:** across the real dataset the tile cube never reaches the crossbar, at
every scale and metric the index offers — 839 species × 6 combinations, asserted rather than
recorded. This was prose, **"0 overlaps in 841"**, and it rotted the moment lengths moved:
[decisions.trunkExponentCalibrated](decisions.md#trunkexponentcalibrated) invalidated the figure and nothing failed, because nothing ran. It
was caught by reading.

What is asserted is a **clearance greater than zero**, taken from the requirement; the tightest
margin measured is logged rather than asserted, so a change that halves it is visible without
pinning today's number as a rule. A boolean overlap check proved a poor guard — the cube sits
bottom-left and the crossbar crosses the upper middle, so floating the cube to the top of the
tile, sliding it 120px right, tripling the projection depth and scaling it by 1.6 all leave it
technically clear while moving the margin a long way.

**Assertions, the Presentation panel:** it explains the cube in prose, naming what it is and
warning that a cube root flattens. The diagram key alone is not enough — the Presentation panel
is where a reader learns to read the drawing (product.pageExplainsItself), and the cube is the one mark that is not a
length.

**Status: verified 2026-10-01 — passing, and watched failing** by removing the cube from the
index tiles, which left 0 across 6; by giving a tile the label too; by removing that prose; by
sizing the side on mass rather
than its cube root; by giving the cube a scale of its own instead of the diagram's px/mm; by
drawing a near-match mass as measured; by removing the key's description of it; by moving it to
the lower-right, where it reached x=980 against a bird axis at x=334; by standing the face on the
tail tip instead of the floor, so the base moved per bird; by zeroing the depth, which is the
square this replaced; and by centring the label without the clamp, which clipped it to
"ass 0.14 oz" on the hummingbird.

**The provenance assertion was vacuous when written, and breaking the code is what showed it.**
Every fixture mass was `measured`, so the dashed branch never rendered and drawing a near-match
square as solid changed nothing. The fixture now gives the stilt a near-match mass, as 10 of the
841 real ones are — the same gap the crow's modeled wingspan already fills for the crossbar.

**unitToggleSwitches's legend check was satisfied by a tooltip, and adding the cube is what exposed it.** It
asserted that each diagram part appears in the Presentation panel as `">Bill "`, `">Mass "` and
so on — a raw search over the whole panel. The cube carries a `<title>` reading *"Mass as water:
a cube of water weighing the same"*, which matched `">Mass "` whatever the label did, so blanking
the legend's mass label entirely left the check green. It now collects the legend's `<text>`
labels and requires the part to begin one of them. The looser form had been passing for the other
seven parts on luck: nothing else in the panel happened to contain `">Bill "`.

**The diagram key is one template now, not two.** It was duplicated verbatim in the detail and
comparison views, and this is the change that would have drifted them: the sentence naming the
square had to be added to both. A page carrying a square its key does not explain is the failure
this guards.

### comparisonSortRewritesUrl

**Sorting a comparison rewrites its URL, and nothing else**

**Verifies:** product.comparisonOneScale · **Level:** code · **Method:** assertions on the rendered comparison view,
following each sort link

**Assertions, the ordering lives in one place:** each sort link is an ordinary comparison URL
naming **the same species in a new order**, carries **no `sort` or `by` parameter**, and keeps
the display options. Following it draws the panels in exactly the order its URL names. This is
the design decision the criterion exists to pin: an ordering parameter would let the URL and the
panels disagree, and a sorted comparison would stop being shareable as itself.

**Assertions, the sort:** each metric orders the species **ascending**, as the index orders its
sections, and the option reads as **on** after its own link is followed. The on-state is a
statement about the arrangement on screen, so it is checked by sorting and re-reading rather than
by looking for stored state — there is none.

**Assertions, scope:** no sort control below two birds, there being nothing to sort.

**Status: verified 2026-10-01 — passing, and watched failing** by sorting descending; by sorting
every option by length regardless of its metric; by adding a `sort=` parameter instead of
reordering; by offering the control for a single bird; and by hardcoding the on-state true.

**The fixture could not tell the three metrics apart, and the "sort by length regardless"
breakage is what exposed it.** All five original species were monotonic in length, wingspan and
mass alike, so every sort produced the same order and a control that ignored its metric entirely
passed. The fixture now includes the Thick-billed Murre, which is heavier than the crow while
being shorter and narrower-winged — the one species for which the metrics disagree. The test
asserts up front that length and mass sort the set differently, so if that property is ever lost
the test says so instead of quietly passing.

### trunkScalesLikeSkeletons

**The trunk scales the way measured skeletons do, and is anchored where it was**

**Verifies:** product.trunkFromMass · **Level:** code · **Method:** recover the model's own exponent by regression
over the real mass range, then check it against an interval taken from published measurement

**Assertions, the exponent:** fitting log trunk on log mass across the dataset's mass range
recovers an exponent inside **0.338 – 0.401** — the intersection of the 95% confidence intervals
[`bohmer-neck`](requirements.md#bohmer-neck) reports for neck length and total leg length, the two
multi-element dimensions it measures.

**Assertions, the interval actually discriminates:** the superseded exponent **0.310 lies outside
that interval**, and so does an exaggerated 0.45. A band wide enough to admit the value it
replaced would be a check that cannot fail, which is the specific failure this project keeps
finding; the test asserts the band's sharpness rather than trusting it.

**Assertions, the anchor:** at **63.4 g** the trunk equals what `18.0 × mass^0.31` gave, within
0.01 mm. This pins the coefficient, which the exponent assertion alone cannot see — a model with
the right exponent and a doubled coefficient passes every scaling check while drawing every bird
wrong.

**Assertions, the direction:** relative to the superseded model, species lighter than the anchor
get shorter trunks and species heavier than it get longer ones, with the ratio increasing
monotonically in mass. This is what "a change of proportion, not of scale" means, and it is the
claim product.trunkFromMass's rationale actually makes.

**Status: verified 2026-10-02 — passing, and watched failing** four ways:

| breakage | what went red |
| --- | --- |
| restore the superseded model whole — `18.0 × mass^0.31` | exponent and direction. **The anchor assertion passed.** |
| keep 0.37, restore the 18.0 coefficient | anchor and direction. **The exponent assertion passed.** |
| overcorrect the exponent to 0.45 | exponent, anchor and direction together |
| widen the band to 0.30 – 0.45 | sharpness, before any model is evaluated |

**The first two rows are the point, and the first one corrected a wrong guess.** This criterion
was drafted claiming the restored old model would fail the exponent *and* anchor assertions. It
does not: at the anchor the superseded model trivially equals itself, so the anchor check is
silent and only the exponent catches it. Which means the anchor assertion's own value could only
be demonstrated by the second breakage — right exponent, wrong coefficient, which scales
perfectly while drawing every bird 28% too long. Each assertion is load-bearing for a failure the
other cannot see, and that is only known because both were tried.

**The anchor assertion nearly was not written at all.** The first draft checked the exponent
alone, which is the number the evidence is about. The evidence constrains a slope; the requirement
constrains a curve.

### faviconFromPalette

**The favicon is painted from the app's palette, not its own**

**Verifies:** product.oneDiagramVocabulary · **Level:** code · **Method:** parse both assets out of the embedded FS and
compare the favicon's colours against the stylesheet's

**Assertions, no invented colours:** every colour literal in `favicon.svg` is one the stylesheet
declares — as a diagram part's `stroke`/`fill`, or as a surface custom property. A shade that
exists only in the icon fails, which is the exact drift this criterion exists for.

**Assertions, the parts it does draw:** the favicon contains the **crossbar** colour and the
**bill** colour, both read from the stylesheet's `.wing` and `.bill` rules rather than written
here. Change either in the stylesheet alone and this goes red.

**Assertions, the comparison can still bite:** the stylesheet must yield a non-empty palette and
the favicon a non-empty set of colours, so a parser that silently matches nothing fails loudly
instead of passing.

**Status: verified 2026-10-02 — passing, and watched failing** three ways:

| breakage | what went red |
| --- | --- |
| restore the favicon's old `#5ccbfa` crossbar | **both** assertions — the colour paints no diagram part, *and* the wing colour is absent. The real, shipped drift, not a planted one |
| change `.wing` in the stylesheet alone, leaving the icon behind | the part assertion, naming the new colour. This is the direction the next drift will come from |
| rename the `{{define "diagram"}}` block the part set is derived from | `the part set cannot be derived` — a fatal, rather than an empty part set quietly matching nothing |

The third matters because the derivation runs in two steps: the parts come from the template, the
colours come from the stylesheet rules naming those parts. If step one yields nothing, step two
yields an empty palette and the check either flags every favicon colour or, just as easily, none.
It fails loudly instead of choosing one.

**What this does not check.** It verifies the favicon uses *colours from* the palette, not that it
maps each part to the right one. At 16px the icon has no neck and no separate tail, and it paints
the body in the head's `#cbd5e1` because the body's own `#64748b` is 3.75:1 and unreadable at that
size — a size concession product.oneDiagramVocabulary allows. A part-by-part mapping check would have to encode those
omissions and would break every time the icon was redrawn, so the weaker invariant is deliberate.
It still catches drift, which is the failure that actually occurred.

**This replaces an assertion that could not fail.** ownRenderableFavicon claimed the favicon "is drawn in the
diagram's palette" and implemented it as a search for three hardcoded literals — one of which,
`#5ccbfa`, appeared nowhere but the favicon. It compared the icon against a copy of itself, so it
would have passed through any change to the app's colours whatever. That is why the drift it was
written to prevent went unnoticed until someone looked.

### tileHeaderNeverWraps

**Nothing in a tile header can wrap to a second line**

**Verifies:** product.tileContents · **Level:** code · **Method:** a structural check over the stylesheet, plus
assertions on the rendered markup — the pattern [familyListWraps](#familylistwraps) uses, for the same reason

**Assertions, the name line cannot wrap:** `.nameline` declares `flex-wrap: nowrap`, and the
common name sits in its own `.cnname` element carrying `white-space: nowrap` with
`text-overflow: ellipsis`. Both halves are needed: without the element the ellipsis would eat the
banding code at the end of `.cn`, which is the one part that must survive.

**Assertions, the code never truncates:** `.cn .code` declares `flex: none`, so it keeps its
width while everything beside it gives way.

**Assertions, the badge cannot wrap:** `.badge` declares `white-space: nowrap` with
`text-overflow: ellipsis` and a `max-width`, a bare `nowrap` being worse than wrapping — it would
overflow the tile instead.

**Assertions, markup:** every tile header contains exactly one `.cnname`, and the banding code
when present is a sibling of it rather than inside it. A code nested in the truncating element is
the defect this criterion exists to prevent, and it looks identical until a name is long enough.

**Status: verified 2026-10-02 — passing, and watched failing** by removing `flex-wrap: nowrap`
from `.nameline`, which is the shipped bug; by dropping `white-space: nowrap` from `.cnname`; by
dropping `flex: none` from `.code`; by moving the code inside `.cnname` in the template, where
the markup assertion catches what no stylesheet check can see; and by reverting `.badge` to its
wrapping form.

**Why this is structural and not visual.** Whether text *actually* wraps depends on font metrics
and the viewport, which this project cannot measure without a browser harness — the gap already
declared for product.familyFilter and product.diagramOnlyMode–product.firstLoadShowsAll. What is checkable is that the rules making wrapping
impossible are present and that the markup gives them something to act on. That is weaker than
seeing it, and it is what caught the bug's shape: the original defect was not a bad measurement
but a missing declaration.

### everyControlDocumented

**Every control is documented, and nothing else is**

**Verifies:** product.pageExplainsItself · **Level:** code · **Method:** compare two sets extracted from the rendered
index — the control labels, and the terms the Controls panel defines

**Assertions:** the set of control labels **across every panel** and the set of terms the panels'
documentation defines are **equal**. Each direction is a separate failure with its own message: a
control with no entry is undocumented, and an entry with no control is documentation for
something that no longer exists.

Entries marked `ctrl note` are excluded from the comparison. They explain something that is not a
control — what a zone is, what pelagic means — and requiring a matching label would force every
explanation to pretend to be a button. The exclusion is narrow and marked in the markup, so an
entry cannot drift out of the check by accident.

**Assertions, the comparison can bite:** the documentation block is **located before it is
sliced**, both sets are non-empty, and the label set contains a control known to be present. A
selector that silently matches nothing would make two empty sets compare equal and the check pass
while verifying nothing.

The locate-before-slice step is not pedantry. The first version indexed straight into the body,
so breaking the selector panicked on `body[-1:]` before any assertion ran — reporting a crash
instead of "the selector is broken". That is the same confusion familyListWraps records, where a parser
that depended on the structure under test reported its own failure rather than the defect.

**Status: verified 2026-10-02 — passing, and watched failing** by the live defect it was written
for: **Names**, added with [product.englishNameChoice](sources/product/diagram.md#englishnamechoice), was a control with no documentation, and
this criterion is what caught it. **Re-verified after the panel split** ([product.collapsibleControls](sources/product/presentation.md#collapsiblecontrols)),
where it immediately caught the Family documentation left behind in the Controls box. Also watched failing by deleting the Units entry, by adding an
entry for a control that does not exist, and by breaking the extraction selector so both sets come
back empty.

**This replaces an enumeration in the requirement itself.** product.pageExplainsItself used to name the controls it
covered, and had been updated for neither Compare birds nor Names. A list cannot notice what it
omits — the requirement, the panel and the controls each looked complete on their own. Comparing
the two sets is the only form of this check that fails when someone adds a control and forgets.

### regionFilterBehaves

**The region filter selects, groups and composes**

**Verifies:** product.regionAndMonthFilter, product.filtersCompose · **Level:** code · **Method:** rendering tests over `?region=`

**Assertions, filtering:** filtering to one region shows species reported there and excludes
others; no parameter shows all; and an explicitly empty `region=` shows none, the same
absent-versus-empty distinction the family filter draws.

**Assertions, the selector:** the box offers **both groupings** of [product.regionsGroupedTwoWays](sources/product/months.md#regionsgroupedtwoways),
each labelled and each carrying a note saying whether it is measured or drawn, with
**Geographic first**; within each,
regions are grouped into zones, **Pelagic is the last zone**, and every region is present, each
grouping being a partition. Each zone header toggles its whole group; **All** toggles
everything; and each chip carries the standard postal abbreviation with the full name beside it.

Checking each grouping **separately** is the point. Reading the zone toggles across the whole
box and checking only the final one passed on the hand-drawn grouping alone and left the derived
grouping's ordering unchecked — one of the whole-page assertions this file keeps having to
scope.

**Assertions, composition:** region and family narrow together ([product.filtersCompose](sources/product/months.md#filterscompose)), and
an empty intersection renders the empty state naming **both** filters — which is now reachable
through the UI, where with one filter it was not.

**Assertions, the sea:** a pelagic species appears under its ocean, and a seabird that also comes
near shore appears under both its ocean and the states it is seen from. Species with no region at
all are not silently swallowed.

**Status: verified 2026-10-03 — passing, and watched failing** by ignoring the parameter; by
treating an empty `region=` as no filter, which is the bug the family filter had first; by moving
Pelagic out of last place; by filtering pelagic species out of every region, which hides 53
birds that have nowhere else to be; by rendering only the first grouping, which the per-block
split catches and the old whole-box read did not; and by blanking the hand-drawn grouping's
note, which leaves the two indistinguishable.

### zonesAnswerToData

**The derived zones answer to the data**

**Verifies:** product.regionsGroupedTwoWays · **Level:** code · **Method:** recomputation over `data/region-presence.tsv`

The "Common birds" grouping is a **claim about the records**, so it is checked against the
records — the same committed table the filter itself uses, not a fixture, because a fixture
would let the shipped grouping and the shipped data drift apart.

**It does not re-run the clustering.** Re-running the algorithm would compare it against itself
and ratify whatever it produced. What is checked is the property the grouping is *sold* on.

**Assertions, per region:** every region resembles its own zone-mates more than it resembles
regions at large. Per region, not per zone: a per-zone average hides an intruder, and the first
version of this check — comparing a zone's internal similarity against its surroundings —
**passed with Hawaii filed under Great Lakes**. The bar is zero rather than a fitted number,
because "resembles its zone-mates more than regions at large" is what belonging to a group
means; any positive threshold would record what this data happened to do.

**Assertions, against the baseline:** the derived grouping beats the hand-drawn one both on mean
within-zone similarity and on the **worst region's own-zone rank**. Both are comparative, so
neither carries a fitted threshold, and the hand-drawn grouping is a real baseline rather than
an invented one. Rank is the measure with teeth: mean cohesion barely moves when one region is
misfiled, but a misfiled region drags its new zone-mates down everyone's ranking.

**What is deliberately not asserted:** that every region's own zone is its best match. That is
false here, and recording it as a target would be dishonest. Nine regions sit closer to a
neighbouring zone than to their own, by margins as small as 0.001; reassigning them iteratively
does **not** converge but cycles, five regions moving every sweep. The mainland is a smooth
gradient and any cut is a choice of where to slice ([decisions.regionsTwoWays](decisions.md#regionstwoways)).

**Status: verified 2026-10-03 — passing, and watched failing** by filing Hawaii under Great
Lakes and Texas under New England (both caught on rank), and by filing Navassa under
Mid-Atlantic and Alaska under Gulf Coast (both caught per region, Alaska at 0.446 against 0.465
for regions at large). The earlier per-zone form of the check was watched **passing** under the
Hawaii break, which is why it was replaced.

### regionUrlRoundTrips

**A region's URL is readable, unique, and round-trips**

**Verifies:** product.regionAndMonthFilter · **Level:** code · **Method:** unit tests over the region table, plus a
rendering test over the selector

**Assertions, the slug:** every region's slug is its postal code where it has one and its
PascalCase name where it has none; slugs are **unique across all 74 regions**, and every slug is
URL-safe without escaping. Uniqueness is the load-bearing one — two regions sharing a slug makes
one of them unreachable, and the collision would be silent.

**Assertions, round-trip:** resolving a region's own slug returns that same region, for every
region. A slug that renders but does not resolve is a dead link on every chip.

**Assertions, no invented codes:** every non-empty code is two characters, and the ten regions
with no standard abbreviation carry an **empty** code rather than a made-up one
([decisions.regionUrlIsPostalCode](decisions.md#regionurlispostalcode)). This is the assertion that stops the rejected proposal being
reintroduced quietly.

**Assertions, rendering:** region links carry slugs and **no GADM id appears in any URL**; a
chip with a code shows code and name, and a chip without one shows the name alone — not an
empty code slot, which the stylesheet would precede with a stray separator.

**Status: verified 2026-10-03 — passing, and watched failing** by returning a stored slug
instead of deriving it, so a renamed region kept its old URL; by emitting GADM ids in links
again; by giving Palmyra the code `PM`, which the no-invented-code assertion rejects; and by
rendering the empty code, which puts a leading separator on ten chips.

Giving two regions the same slug does not reach an assertion: the index is built once at
startup and **panics** there, naming both regions. That is deliberate — a collision makes a
region silently unreachable, so it should stop the binary rather than wait for a test run — and
it is recorded here because "watched failing" meant watching a panic, not a red assertion.

### familyGroupingComplete

**The family grouping covers every family, and stays off the tile**

**Verifies:** product.familiesByOrder, product.tileContents · **Level:** code · **Method:** unit tests over the grouping against
the compiled dataset, plus a rendering test over the selector and a tile

**Assertions, coverage:** every family of the dataset sits in exactly one taxonomic order, and
every order holds at least one family. Checked **against the compiled dataset, not a fixture** —
the family list comes from AviList, which renamed 55 of 90 English family names in
[decisions.avilistReplacesAba](decisions.md#avilistreplacesaba), so a family that is added or renamed must fail a test rather
than silently vanish from the control.

**The hand-drawn birder grouping this criterion also covered is gone**, retired by
[decisions.habitatDietSpeciesLevel](decisions.md#habitatdietspecieslevel) in favour of the per-species filters of
[product.habitatAndDietFilters](sources/product/months.md#habitatanddietfilters). Its distinctness and ordering assertions went with it, and so did
[decisions.familyInTwoGroups](decisions.md#familyintwogroups)'s cap on duplicate placements — there is no hand-drawn grouping
left for a family to be placed in twice.

**Assertions, the tile:** a tile carries **exactly one badge**. This is the assertion that
keeps [product.tileContents](sources/product/presentation.md#tilecontents) true when a third grouping arrives: without a count, "the other
groupings need no badge" is a sentence in a decision record rather than a property of the code.
It is a count, not a list of forbidden badges, because a list invites one more entry.

**Status: verified 2026-10-03 — passing, and watched failing** by adding an order badge beside
the family badge, which makes 6 birds carry 12 badges.

Its earlier breaks — moving Corvidae out of the table, collapsing the two groupings into one,
reversing their order, orphaning Strigidae, tripping the duplicate cap — were all against the
hand-drawn grouping and went when [decisions.habitatDietSpeciesLevel](decisions.md#habitatdietspecieslevel) retired it. They are recorded
here rather than deleted, because a criterion that quietly loses most of its assertions should
say so.

The coverage check **skips** rather than fails when `web/data/species.json` is absent, since
the dataset is a build artifact and is not committed (tech.datasetNeverCommitted). That is a real gap: on a machine
that has never run the ingest, this criterion is silent.

### creditsComplete

**Credits are complete and the docs are not stale**

**Verifies:** product.selfDescriptionStaysTrue · **Level:** code · **Method:** tests over the rendered pages and over
`README.md` and `docs/`, against the compiled dataset

**Assertions, credits:** **every page that shows a CC BY source's output credits that source on
that page.** Checked per page rather than site-wide, because that is how the obligation works
and because the gap this criterion was written for was exactly a per-page one — the index
credited AviList and `detail.html` and `compare.html` did not, while showing AviList names,
families and species codes throughout.

**Assertions, counts:** every species or family count stated in `README.md` or `docs/` matches
the compiled dataset. `docs/data.md` said 841 species after [decisions.avilistReplacesAba](decisions.md#avilistreplacesaba) cut
the list to 839.

**Assertions, attribution claims:** no document credits a source for something it no longer
supplies. `README.md` listed the ABA Checklist as the source of "species list, names, families"
after names and families moved to AviList.

**What this cannot check, stated so nobody assumes otherwise:** whether the prose is *true*. A
test can compare a number against the dataset and a credit against a page; it cannot read
`docs/using.md` and notice that the control it describes was redesigned. That remains a reading
task before release, and [product.selfDescriptionStaysTrue](sources/product/presentation.md#selfdescriptionstaystrue) requires it as a release gate rather than
pretending a test covers it.

**Counts are checked only where a claim is about the whole dataset** — "N species across M
families", and a bolded "**N species**". Every number before the word "species" is deliberately
not checked: "533 species without an empirical wingspan" is a subset, "442 of 841 species"
records a measurement taken when the dataset was that size, and Böhmer's "68 families" is
someone else's study. A check that flagged those would be switched off within a week.

**Status: verified 2026-10-03 — passing, having first failed against the real defects**: the
missing AviList credit on `detail.html` and `compare.html`, and `README.md` and
`docs/building.md` crediting the ABA Checklist with names and families. Then watched failing
again after the fixes, by removing the AviList credit, by restoring the 841 count, and by
re-crediting the ABA.

Two of the checks were wrong before they were right, and both were caught by watching them
fail rather than by reading them. The count check first flagged three legitimate numbers. The
attribution check flagged **its own correction**, because "names come from AviList, not the ABA"
contains both the source and the word; scoping it to the sentence fixed that, and then the
sentence splitter broke on the "." in "ABA Checklist 8.19" and silently stopped catching
anything — visible only because the deliberate break failed to go red.

### categoryOriginShown

**Every category on the detail page says where it came from**

**Verifies:** product.categoryProvenance · **Level:** code · **Method:** rendering test over the detail view

**Assertions:** the detail page lists the bird's family, order, regions, habitat and niche, and
any group it is in; **each carries a provenance of published, derived or ours**; and a published
one names its source. A category rendered with no provenance fails, which is the whole point —
the failure mode is an opinion of ours sitting unlabelled beside a published fact.

**Assertions, regions:** the regions a bird is listed in are **one row**, not a row apiece.
They share a single rule, so sixty copies of one sentence would bury the habitat rows where
BirdView sometimes differs from AVONET. A bird in no region gets **no** Regions row rather than
an empty one — six species are at sea where no land polygon reaches ([product.regionAndMonthFilter](sources/product/filtering.md#regionandmonthfilter))
and must not be shown as placed nowhere by a rule.

**Assertions, links:** where provenance is an algorithm, the cell names the rule and links to a
document **served by BirdView**, not to an external host. Every `/docs/` link on the page is
fetched and must answer 200, so a fabricated citation cannot pass — **and where the link names
a section, that section must exist**. A link to the right document and a wrong anchor resolves
with a 200 and lands the reader nowhere near the explanation, which is how the regions row came
to cite a document that said nothing whatever about regions.

**Assertions, between documents:** every internal link in every served document resolves. The
guides cite the spec as `../river/requirements/sources/product/`, which is right in a checkout and a 404 in a
browser, and the per-species link check could not see it because these are links between
documents rather than on a species page. **Relative links are resolved as a browser resolves
them**, against the page they sit on; until 2026-10-15 the check followed only absolute links,
so it could not see an unrewritten link at all. A link a guide makes into the repository —
a requirement, source code, a data table — is rewritten to the published repository, and the
check confirms the file it names exists in the checkout.

**Assertions, titles:** a document is titled and indexed by its **first top-level heading**,
not by its first line. The case that forced it was the requirements' `process.md`, which opened
with an HTML comment and was listed in the index as `<!--`; the requirements are no longer
served, and the rule stands for the guides.

**Assertions, what is not served:** nothing under `river/` or `third_party/` is served.
`river/requirements/sources/` and `third_party/` hold other people's terms and must not be reachable
([tech.fetchRestrictedInputs](sources/tech/delivery.md#fetchrestrictedinputs)). Since [decisions.requirementsNotServed](decisions.md#requirementsnotserved) nothing under `river/` is embedded
at all; the probes remain, so that serving the requirements again cannot quietly bring the
vendored terms back with them.

**Re-verified 2026-10-16** when the requirements stopped being served
([decisions.requirementsNotServed](decisions.md#requirementsnotserved)). Watched failing by planting a misspelled repository link in a
guide — `river/decisons.md does not exist in the repository` — and, for the coverage
[noContactAddressPublished](#nocontactaddresspublished) would otherwise have lost, by planting an address in
`river/goals/goals.md`.

**Status: verified 2026-10-03 — passing, and watched failing** by giving each region its own
row; by rendering a category with an empty origin; by citing `/docs/regions`, a document the
server does not serve; by citing `#how-regions-work`, a section that does not exist; by
renaming the section the citation names; by sending the spec's sibling links to `/docs`, which 404s eight
requirement documents; by widening the embed to `*.md sources`, which pulls in the vendored
terms; and by reading the title from line one again, which puts `<!--` back in the index.

**Re-verified 2026-10-08** when the vendored iNaturalist pages moved from `spec/sources/` to
`third_party/` for the Apache 2.0 release. The check gained probes for the new tree, and was
watched failing by mounting `third_party/` on a file server:
`/third_party/inatterms/terms.html is served; vendored terms and non-documents must not be`.
A check that knew only the old home would have gone on passing after the files it guards had
moved somewhere it never looks.

### habitatDietFilterPerSpecies

**Habitat and diet filter per species, and say where each value came from**

**Verifies:** product.habitatAndDietFilters · **Level:** code · **Method:** unit tests over the trait table against the
compiled dataset, plus rendering tests over the controls

**Assertions, the vocabulary:** every value assigned to a species is one the control offers, so
a bird cannot carry a category nobody can select; every option has a label and a quoted
definition; and the jargon terms — `Invertivore`, `Vertivore`, `Granivore` and the rest — are
never shown untranslated.

**Assertions, origin:** every assignment carries an origin. A published one names its source,
a `birdview` one gives its reason. An assignment with no origin fails: that is the
opinion-as-finding failure [product.categoryProvenance](sources/product/measurements.md#categoryprovenance) exists to prevent.

**Assertions, the supplementary table:** **every key is a species in the dataset.** This is the
assertion the table most needs — a key AviList does not use misses silently, adds nothing, and
leaves the table looking correct. A second assertion counts the assignments that actually reach
the dataset against the number the table defines, so a miss surfaces as a number rather than as
nothing at all.

**Assertions, the control:** the Species panel offers Habitat and Diet before the taxonomic
family list, each option carrying its count and its definition, and the retired birder grouping
is not rendered.

**Assertions, the additions page:** `/added` lists **every** assignment that is not published,
counted against the dataset rather than eyeballed — a page that quietly lists a subset is worse
than no page. It lists no published assignment, those being AVONET's rather than BirdView's to
claim, and every row names its origin and its reason.

**Assertions, coined terms:** a term BirdView added where AVONET has none says so in its own
definition and is marked in the control; it can only be assigned with origin `birdview`, since
a category this project invented cannot be published by anyone; and the number of them is
**capped**, because the difference between filling a gap and acquiring a habit is a count.

**Status: verified 2026-10-03 — passing, and watched failing** by coining two further terms
for niches AVONET already covers, which trips the cap; by assigning the coined `Sap` category
with origin `published`; and by the two supplementary keys
that were dead on arrival: `Phalacrocorax auritus` and `Larus argentatus`, where AviList says
`Nannopterum auritum` and `Larus smithsonianus`. They produced no assignment and no error, and
the only symptom was the count being 3 where the table defines 5 — which is exactly why the
count assertion exists beside the key assertion.

---

## Traceability

| Requirement | Criteria | |
| --- | --- | --- |
| product.scopeIsRecurringAvifauna | scopeMatchesStatusTable, noRarityCodes, abaFileHashMatches | |
| product.traitsFromAvonet, product.traitsPerSpecies | allFiveTraits | |
| product.namesFromAvilist | namesNeverDefaulted | |
| product.englishNameChoice | nameLocaleToggle, stableSpeciesId | |
| product.generatedBadgeColors | legibleDistinctColors, badgeTextInheritsColor | |
| product.joinOnBinomial | joinCoversEverySpecies | |
| product.unmatchedSpeciesReported | unmatchedFailsBuild | |
| product.familyFilter | familyFilterSelectsAndClears, familyListWraps | |
| product.regionAndMonthFilter, product.filtersCompose | regionFilterBehaves, regionUrlRoundTrips, belowBarCountsKept | |
| product.regionsGroupedTwoWays | regionFilterBehaves, zonesAnswerToData | |
| product.familiesByOrder | familyGroupingComplete | |
| product.habitatAndDietFilters | habitatDietFilterPerSpecies | |
| product.categoryProvenance | categoryOriginShown | |
| product.oneSelectedMetric, product.ascendingOrder | sortsAscending | |
| product.sectionBreakRule, product.sectionThresholdDefault | sectionBoundariesObeyRule | |
| product.sectionHeadersStateRange | sectionHeadersMatch | |
| product.diagramOnlyMode | bothCommandsBuild | |
| product.tileContents | tileCarriesHeaderAndMetrics, pickModeSurvivesNavigation, tileHeaderNeverWraps, familyGroupingComplete | |
| product.imperialByDefault | unitToggleSwitches | |
| product.firstLoadShowsAll | familyFilterSelectsAndClears | |
| product.tileOpensDetail | detailShowsProvenance, bothCommandsBuild, detailDrawsFullScale | |
| product.pageExplainsItself | pageExplainsControls, massDrawnAsWaterCube, everyControlDocumented | |
| product.collapsibleControls | scrollScriptScoped | |
| product.helpfulEmptyState | emptyStateGuides | |
| product.togglesKeepPlace | scrollScriptScoped | partial — behavior unverifiable without a browser |
| product.segmentComposition | diagramPartsInOrder, foldedNeckSameStroke, foldingFamiliesHaveNecks, diagramDistinguishesModeled | legsFromTarsusWithdrawn withdrawn (decisions.inFlightNoLegs) |
| product.segmentsSumToLength | segmentsSumExactly, foldedNeckSameStroke | |
| product.scaleToggle | sharedScale, controlsMatchServer | |
| product.crossbarAtVerticalScale | crossbarToScale | |
| product.proportionDifferentiates | proportionsDifferentiate | |
| product.neckModelPerFamily | neckLengthSerialized, foldingFamiliesHaveNecks | |
| product.craniumFromMass | craniumVaries | |
| product.trunkFromMass | trunkScalesLikeSkeletons | |
| product.oneDiagramVocabulary | faviconFromPalette | |
| product.lengthAndWingspanSources (composite) | totalIsSumOfSegments | |
| product.lengthAndWingspanSources (wingspan model) | wingspanModelReconstructsFox | |
| product.speciesPhotographs – product.photoCaching | — | withdrawn |
| product.lengthAndWingspanSources | provenanceMatchesSources, whiteTernFromBirdlife, herringGullFromEuropeanRow | |
| product.borrowedMeasurementsSaySo | sharedMeasurementsDisclosed | |
| product.escapedBirdsListed | escapedBirdsListedAndServed | |
| product.tileMarksModeled | tileMarksNonMeasured | |
| product.provenancePerMetric | provenanceWellFormed | |
| product.metricProvenanceExplained | detailShowsProvenance | |
| product.inaturalistLink | inaturalistLinkConfidentOnly | |
| product.measuredVersusModeled | diagramDistinguishesModeled | |
| product.labeledDetailDiagram | segmentLabelsDontCollide, totalLengthLabeled, massDrawnAsWaterCube | |
| product.comparisonOneScale | comparisonSharesScale, comparisonSortRewritesUrl | |
| product.pickFromIndex | pickModeSurvivesNavigation | |
| tech.twoCommandsOneModule, tech.servesOnLocalhost | bothCommandsBuild | |
| tech.selfContainedServer | noServerNetworkRequest, ownRenderableFavicon | |
| tech.fetchRestrictedInputs | nothingRestrictedTracked, abaFileHashMatches | |
| tech.datasetNeverCommitted | nothingRestrictedTracked | |
| tech.identifiedClient | outboundRequestsIdentified | |
| tech.badgeContrast | legibleDistinctColors | |
| tech.textContrastFloor | noTextBelowContrastFloor | |
| product.selfDescriptionStaysTrue | creditsComplete | |
| tech.localNowPublicLater | attributionComplete | partial — the non-commercial posture itself is still untestable |
| tech.bundleIsNormative, tech.recordBeforeConsuming | everyRequirementCitesGoal, vendoredHashesMatch, decisionReferencesExist | |
| `wcag/1.4.6` | legibleDistinctColors | |

### Declared gaps

Honest gaps, rather than criteria that would not bite:

- **product.familyFilter (family filter), product.diagramOnlyMode–product.firstLoadShowsAll (modes, tile contents, units, first load).** These are
  interaction behaviors. They are testable, but only with a browser harness this project does not
  have, and a unit test asserting that a template contains a string verifies the template rather
  than the behavior. Covered by human review until a harness exists; promote them when one does.
  familyFilterSelectsAndClears and familyListWraps do not close this gap: they check the selector's structure and layout, not
  what clicking a family does.
- **tech.localNowPublicLater (never commercial).** A standing posture with no code to check. It is enforced by the
  human gate and by `requirements.md`, not by a test.
- **product.tileOpensDetail (detail view)** is only partially covered: detailShowsProvenance checks that provenance appears there,
  not that the rest of the view is right.

### Criteria that cannot yet fail

All `planned` criteria. As each is implemented, break its subject deliberately and record the
result here — a check believed but never seen failing is worse than no check.

### onlyNarrowsOneGroup

**"Only" narrows to one group and leaves other sections alone**

**Verifies:** [product.onlyPerGroup](sources/product/presentation.md#onlypergroup) · **Level:** code · **Method:** rendering tests over the
index, following each control's own URL

**Assertions:** every group selector in Regions, Diet and Habitat and Families is followed by an
**Only** link. Following one leaves exactly that group's members selected within its section and
nothing else: Only beside a region zone selects that zone's regions and drops every other
region; Only beside an order selects that order's families and drops every other family.

**Assertions, section independence:** Only on a diet value leaves the habitat selection
untouched, and the reverse. Only on a region touches neither families nor traits. This is the
assertion that matters — a narrowing control that quietly empties a neighbouring filter returns
an empty page and looks like a data bug.

**Assertions, independence from current state:** the target is built from the group, not from
what is selected, so the same Only link appears whatever else is on. Two different starting
selections produce the same Only target for the same group.

**Status: verified 2026-10-04 — passing, and watched failing** by building the target from the
current selection rather than from the group (the two starting selections then disagree), and by
pointing a trait's Only at both trait parameters instead of its own (the habitat selection is
then cleared by a diet click).

### massInOneUnit

**Mass is written in one unit across a comparison**

**Verifies:** [product.oneUnitInComparison](sources/product/presentation.md#oneunitincomparison) · **Level:** code · **Method:** unit tests over
`unitsFor`, plus rendering tests over the index and comparison pages

**Assertions, the rule:** the smallest mass in a scope chooses the unit, so a scope spanning the
switch point is written in the smaller one throughout. A scope entirely above it keeps the large
unit — the rule is "the smallest decides", not "always ounces". A scope with no recorded mass
does not crash and does not pick the large unit on no evidence.

**Assertions, the scopes:** on the index the scope follows the **scale** setting — per section
under a per-section scale, the whole shown set under a global one — and a section's own
`Min`–`Max` header is written in that section's unit, not a per-value one. On the comparison
page the scope is the set named in the address: adding a small bird to a set of large ones
rewrites the large ones.

**Assertions, no mixing:** over the real dataset, no section and no comparison renders two
different mass units at once. This is the assertion that would have caught the original defect,
which was found in a screenshot rather than by any check.

**Status: verified 2026-10-04 — passing, and watched failing** by choosing the unit from the
largest mass instead of the smallest — a 27 g House Sparrow beside a Trumpeter Swan then
renders `0.1 lb`, one significant figure for a whole bird — and by reverting the index to a
per-value unit, which makes sections 36, 37 and 38 render both `oz` and `lb`.

### neckRatiosReviewed

**Every family's neck ratio has been looked at**

**Verifies:** [product.neckModelPerFamily](sources/product/diagram.md#neckmodelperfamily) · **Level:** code · **Method:** test over the compiled
dataset against `internal/morphology`

**Assertions:** every family in the dataset that is **not** in Passeriformes appears either in
`NeckRatios`, with a ratio, or in `NeckDefaultReviewed`, with a written reason for leaving it on
the default. A family in both fails, since the two mean opposite things, and a reviewed entry
with an empty reason fails, since it records nothing.

**Why the order-level exemption is sound:** all 43 Passeriformes families here are small compact
songbirds with no exposed neck, so the default is a short stub and correct for every one. Writing
the same sentence 43 times would bury the families where the question is live. The exemption is
itself the claim: a new passerine family needs no review, a new family in any other order does.

**What this replaces:** the table's own comment conceded that "only a review of the table against
the full family list will catch the next one", and nothing made that happen. Absence from
`NeckRatios` meant two different things — considered and unremarkable, or never considered — and
nothing could tell them apart. It cost a Limpkin drawn with a stub neck, a Pomarine Jaeger drawn
with half a gull's, and a Cathartidae nobody revisited after AviList gave the New World vultures
their own order.

**Status: verified 2026-10-04 — passing, and watched failing** by removing Stercorariidae from
`NeckRatios`, by removing Procellariidae from `NeckDefaultReviewed`, and by listing Alcidae in
both.

### commonestShareIsShare

**The commonest-region share is a share, and its citation resolves**

**Verifies:** [product.shareOfRecords](sources/product/presentation.md#shareofrecords), [product.categoryProvenance](sources/product/measurements.md#categoryprovenance) · **Level:** code ·
**Method:** unit tests over the metric, rendering tests over the tile and detail page, and a
walk of every anchored citation the site emits

**Assertions, the metric:** the region chosen is the one where the species is the largest
**share** of that region's bird records, not the one holding the most of its records — the two
differ for 564 of 1,081 species. Both halves of the ratio reach the dataset, so the detail page
can show the evidence a thin share rests on.

**Assertions, the tile:** the share **follows the region filter** — the same bird reads
differently under `?region=AK` than across everything, and narrowing to where it is densest
raises its share. Region ids and their counts travel as parallel slices, so their lengths are
pinned, no count exceeds its region's total, and every region a species is placed in has a
total to divide by. The line renders on **every** tile, em dash and all, because a line that
appears on some tiles and not others makes those tiles shorter and takes the whole flex row out
of step. The footer stays one line ([product.tileContents](sources/product/presentation.md#tilecontents)) — a fourth cell there wrapped it,
which is why the share has its own line.

**Assertions, the citation:** every `#fragment` the site emits resolves to a heading in the
document it points at. This is the assertion that matters most, and it did not exist: the
link-walking check drops the fragment first — *"fragments are within the page"* — so it verified
`/docs/data` and never `#commonest`. Writing the row before the section left the suite green.

Walking it found **176 broken cross-references in the served spec**: every `(product.md#scopeisrecurringavifauna)`
in the bundle, which `river/check/refs.py` validates in the *markdown*, rendered as
`#scopeisrecurringavifauna--scope-is-the-aba-area-avifauna-that-recurs-by-birdviews-own-measure` and landed every
reader at the top of the document. Headings opening with a requirement id now anchor on the id.

**Status: verified 2026-10-05 — passing, and watched failing** by ignoring the region filter in
`shareOf`, which makes the Alaska and everywhere figures identical; by appending one extra count
in `applyRegions`, which desyncs the parallel slices; by removing the heading-id
transformer, which reports `#scopeisrecurringavifauna` and `#avilistreplacesaba` dangling, and by renaming the `Commonest`
heading in `docs/data.md`. The rename needed the fixture to carry a commonest region first:
without it no citation rendered and the check passed over nothing, which is the same vacuous
shape it was written to catch.

### noContactAddressPublished

**No served page publishes a contact address**

**Verifies:** [product.noPersonalAddress](sources/product/presentation.md#nopersonaladdress) · **Level:** code · **Method:** rendering test over every
served page, including every document the two indexes list

**Assertions:** no page matches a plain address, the spelled-out obfuscation, or a mail link.
The feedback link appears only when the tracker URL is configured, and an unset one renders
nothing rather than an empty href.

**Why it matches the obfuscation and not just the at-sign:** the address that shipped was
spelled out in words and contains no at-sign at all. A check for one would have passed over the
exact string it exists to catch.

**It is deliberately not limited to the owner's address.** Run the first time it failed twice.
Once on a source owner's mailbox in the served manifest, republished by BirdView because the
provenance table recorded it — a provenance record names who owns a source without handing out
their mail. And once on this bundle: the requirement and this criterion both quoted the address
as an example, and the spec is served, so they republished on `/spec/tech` and
`/spec/acceptance` precisely what the rule forbids. Neither quotes it now. Exempting the spec
was the wrong fix, because the spec is a page readers reach. Since [decisions.requirementsNotServed](decisions.md#requirementsnotserved)
it is no longer served, but it is still published with the repository, so the check reads
`river/*.md` from the checkout instead — the same files, with no loss of coverage.

**Status: verified 2026-10-05 — passing, and watched failing** by restoring the old header
sentence, which reports it on every page that renders the header.

### outOfScopeExplained

**A species left out of scope is named and explained, at 404**

**Verifies:** [product.undrawnSpeciesAnswer](sources/product/presentation.md#undrawnspeciesanswer), [decisions.speciesUrlOutlivesSpecies](decisions.md#speciesurloutlivesspecies) · **Level:** code ·
**Method:** handler test over the real dataset, plus an ingest test over the real status table

**Assertions:**

1. `GET /species/<id>` for an out-of-scope species returns **404** and a body naming the bird,
   its scientific name and its status. Not 200 and not 410.
2. The body carries the evidence the status rests on — episodes, years and records for
   `vagrant` and `historical` — and for `historical` the **date last recorded**.
3. For `absent` the body carries **no counts at all**, because none were measured.
4. An id matching nothing gets the bare `404 page not found`, with no bird named.
5. The ingest carries every classified-but-not-drawn species into the dataset, and **no id
   collides with a drawn species'**. A collision would hide a living page behind an epitaph.

**Why it is checked over the real dataset and not a fixture:** assertion 5 is a statement about
1,411 real ids produced by one AviList join, and a fixture asserts it over whichever handful the
fixture holds. The fixture tests cover the rendering; the dataset tests cover the claim.

**Three tests, because the assertions sit at different levels.**
`TestUndrawnSpeciesNamesItselfAt404` renders all three verdicts and the typo case against the
fixture; `TestNoIDIsBothDrawnAndNot` and `TestAbsentSpeciesCarryNoEvidence` run over the
compiled dataset. The second of those asserts **both halves** — absent species carry no
evidence, and every other verdict does — because a check written one way round is satisfied by
an ingest that publishes no evidence at all.

**Status: verified 2026-10-05 — passing, and watched failing** under all of:

| Mutation | What went red |
| --- | --- |
| `detail` restored to a bare `http.NotFound` | all three verdicts, "no verdict on the page" |
| the handler writes 200 | "the address serves no bird, and 200 would tell every cache otherwise" |
| the handler writes 410 | both the status assertion and the decisions.speciesUrlOutlivesSpecies one |
| the historical sentence drops `LastSeen` | `page does not say "18 May 1902"` |
| the `absent` branch is given the evidence sentence | four separate count-shaped strings on the page |
| an unmatched id is handed a synthesized record | "unknown id gets the named page" |
| the region prints as its GADM id | missing `Ontario` and `Alaska`, present `CAN.9_1` |
| `status.OutOfScope` filters on `InScope` rather than complementing `Scope` | **the ingest refuses the build**, naming `leastp5` as both drawn and not; with the ingest guard also removed, the dataset test names seven colliding ids, and the server refuses to start |
| the ingest publishes the `absent` species' unmeasured zeros | 167 species reported |
| the ingest withholds evidence from every verdict | the other half, on episodes and region |

**The eighth mutation is the one worth keeping.** It is not a hypothetical: deriving the
rejected set from `Status.InScope()` instead of from the complement of `Scope()` is the obvious
way to write it, and it silently republishes the nine [product.declaredExceptions](sources/product/dataset.md#declaredexceptions) exceptions —
birds that are in scope *with* an out-of-scope status — as epitaphs for birds the index is
drawing. Three guards caught it independently, which is the right number for an invariant that
decides what a URL means.

### commonestIsLargestShare

**The commonest region is the largest share, and the seas are out of the ratio**

**Verifies:** [product.shareOfRecords](sources/product/presentation.md#shareofrecords), [decisions.seaShareDiffers](decisions.md#seasharediffers) · **Level:** code ·
**Method:** dataset test over all 1,083 species, plus rendering tests for the wording

**Assertions:**

1. For every species carrying a `commonest`, **no land region it occupies has a larger share**.
   This is product.shareOfRecords's central claim and nothing asserted it before.
2. `RegionTotals` carries **no marine entry**, so nothing can divide by one.
3. A species' share is computed over regions the totals have an entry for, on **both** sides:
   no region contributes to the numerator without contributing to the denominator.
4. The detail page names the row **Commonest land region**, not "Commonest in".
5. A selection whose regions have no record total reads `No data for these regions yet` —
   never anything of the form *no records*, which would state a gap in BirdView's data as a
   fact about the bird.
7. **No page of the interface, and no guide under `/docs`, uses the word "share"** as the
   name of the metric. The normative documents under `river/` are exempt; they define it, and
they are not served.
6. A species with no records in the regions on screen reads **`0 in <denominator>`** in the
   same sentence every other tile uses — never an em-dash. The denominator is stated, because
   zero in a billion records and zero in nine hundred are different findings.

**Why assertion 1 is over the real dataset.** The server fixture hardcodes `commonestOf` and
every test renders it back, so the claim has only ever been checked against a value the test
itself supplied. That is how [decisions.seaShareDiffers](decisions.md#seasharediffers)(a) shipped: a value the fixture
provides is a value no fixture test can falsify.

**Why assertion 3 is stated as an invariant and not as a region list.** The numerator and the
denominator used to be two separate loops with two separate conditions — `sel[id]` over the
species' regions, `all || sel[id]` over the totals — which is exactly how they came to range
over different sets. Deriving both from "does this region have a denominator" makes the
mismatch unrepresentable rather than merely absent today.

**Status: verified 2026-10-05 — passing, and watched failing** under all of:

| Mutation | What went red |
| --- | --- |
| `shareIn` back to two independent loops | `Anous ceruleus: numerator 192 over every region, want 51 from land alone` |
| a sea-only selection falls back to the em-dash | "selecting only seas renders the em-dash, whose tooltip says *too few records here*" |
| the ingest carries the marine totals again — the shipped bug | **105 violations** of assertion 1, headed by `Ardenna bulleri`: page says Alaska at 1 in 11,171, North Pacific is 1 in 37. `TestRegionCountsLineUp` names all five seas |
| the zero branch collapses back to a nil cell | 29 tiles on the unfiltered index lose their figure; `TestShareStatesZeroRatherThanDashing` names them |
| the ratio is computed before the zero is checked | **a divide-by-zero panic** taking `TestRoutesServe` with it — which is how this was found, rather than by reading the diff |
| the no-denominator label worded as `No records in selected regions` | `Acanthis flammea has 234408 records in CAN.11_1` — a gap in BirdView stated as a fact about the bird |
| the same label left reading `No share for this selection` | the jargon check, `"share" is jargon on a one-line cell` |

| the tile tooltip reads `A share of what is recorded` again | `/ says "A share"` |
| `docs/data.md` reads `what share of the bird records` again | `/docs/data says "what share"` |

**The vocabulary check matches the noun, not the verb.** `Bookmark it, share it`, `need not
share a habitat` and `one shared scale` are ordinary English in a different sense and stay;
the metric always appears as *share of* or behind a determiner. Four such lines survive in
`docs/using.md` unflagged, which is the check working rather than the rule being loose.

**A mutation that does not apply is not a mutation.** The first attempt at the two above
replaced `Note:  "..."` with two spaces, which `gofmt` had already collapsed to one, so the
edit silently no-opped and both "mutations" passed. The anchor is now asserted before the
substitution. This is the same failure shape as a vacuous check, one level up: the test was
fine and the thing testing the test was broken.

The third is the one to keep: it restores exactly the state that shipped, and the check that
now catches it is the check nobody had written.

### peakAndShareOneCalculation

**The peak region and the index share are one calculation**

**Verifies:** [product.shareOfRecords](sources/product/presentation.md#shareofrecords), [decisions.honestSeaDenominator](decisions.md#honestseadenominator) · **Level:** code ·
**Method:** dataset test over every placed species, plus a client test for the measurement

**Assertions:**

1. For every species carrying a peak, **selecting that one region on the index reproduces the
   detail page's figure exactly.**
2. The peak is always **a region the species is listed in**. The old figure could name one the
   reader would not find in the filter at all.
3. At least one species peaks at sea, or the case decisions.honestSeaDenominator exists for is unexercised. (45 do.)
4. The row is labelled **Top region**, and none of its superseded names survives — not
   `Commonest in`, not `Commonest land region`. Rejecting the old forms as well as requiring
   the new one is what would have caught `Commonest land region` outliving
   [decisions.honestSeaDenominator](decisions.md#honestseadenominator), which gave the seas the denominator that qualifier existed
   for.
4. Every region has a record total, seas included.
5. A sea's total **exceeds the sum of the species placed there**. A denominator built out of
   its own numerators is the shape of the decisions.seaShareDiffers bug, and the only part of it a test can see
   without the network.
6. `CountOffLandInBox` subtracts the land, and **refuses** a facet whose rows sum past the box
   rather than returning a negative count.

**Why assertion 1 is the one that matters.** The bug the owner reported was not in either
page's logic — it was that nothing compared them. The peak came from `data/status.tsv` and the
share from `data/region-presence.tsv`, two tables with different denominators and different
membership rules, and each was tested against its own source. A check can only catch that by
asserting the pairing.

**Why assertion 5 names a margin and not a number.** A sea's total runs 8× to 200× the sum of
what BirdView places there, the tightest being the North Pacific at 11.9%. Land is exempt:
there the drawn species really are almost all the records — 98–100%, and Navassa is 6 of 6 —
so no margin exists to assert.

**Status: verified 2026-10-05 — passing, and watched failing** under all of:

| Mutation | What went red |
| --- | --- |
| `Commonest` taken from `status.tsv` again — the reported bug, exactly | **380 violations**, both kinds: peaks naming regions the bird is not listed in, and denominators disagreeing by small amounts (`Accipiter nisus`: detail page 1 in 263,247, index 1 in 263,266) |
| `CountOffLandInBox` returns the box count instead of the off-land count | `off land 1000, want 100` |
| the multi-GADM-level guard removed | `returned -220 and no error` |
| the ingest drops marine totals again | every sea reported as having no total |
| the row relabelled `Commonest land region` | `does not label the peak row "Top region"` **and** `still says "Commonest land region"` |

**The second mutation is the live one.** Faceting plain `gadmGid` rather than `gadmLevel0Gid`
returns every administrative level at once — a record in Virginia counted under `USA`,
`USA.47_1` and `USA.47_13` alike — and measured against the real API the Atlantic box's rows
summed to 1,175,397,927 against the 413,188,208 it holds. Without the guard that is a negative
off-land count, silently stored.

### responsesCompress

**Responses compress, and say that they did**

**Verifies:** [tech.compressedResponses](sources/tech/delivery.md#compressedresponses) · **Level:** code · **Method:** handler test over a page,
a detail view, a static file and a served document

**Assertions:**

1. With `Accept-Encoding: gzip`, the response carries `Content-Encoding: gzip` and the body is
   a valid gzip stream that decompresses **byte-for-byte** to the uncompressed response.
2. Without it, no `Content-Encoding` and the plain body.
3. `Vary: Accept-Encoding` on **both**.
4. No `Content-Length` describing the uncompressed body survives.
5. A [product.undrawnSpeciesAnswer](sources/product/presentation.md#undrawnspeciesanswer) 404 keeps its status **and** its `Content-Encoding`.
6. Each page saves at least 50%, with the real figure logged rather than asserted.

**The header is asserted before the body, and that order is the criterion.** Reading the body
with `gzip.NewReader` does not catch an unlabelled response — the stream is perfectly valid,
it is simply unannounced — so a check that looks at the body first passes over the exact bug
this was written during. It did: the first version of the middleware set `Content-Encoding`
from `Write`, which is too late for any handler that calls `WriteHeader` itself, and the
static files and all 328 undrawn-species pages went out as compressed bodies labelled as HTML.
Confirmed live before the fix: `HTTP/1.1 404`, no `Content-Encoding`, `file` reporting
`gzip compressed data`.

**Why 50% and not a measured figure.** What must not regress is that compression happens;
pinning 94% would fail the first time a template gained a paragraph. The measured savings are
logged on every run — 94% for the index, 70% for a detail page, 67% for the stylesheet, 62%
for a served document.

**Status: verified 2026-10-05 — passing, and watched failing** under all of:

| Mutation | What went red |
| --- | --- |
| label from `Write` rather than `WriteHeader` — the shipped bug | `Content-Encoding "", want gzip` on three of four pages |
| compress regardless of `Accept-Encoding` | `compressed for a client that did not ask` |
| drop the `Vary` header | `Vary is "", missing Accept-Encoding` |

### frequencyTiers

**Frequency orders commonest first, in tiers of powers of ten**

**Verifies:** [product.orderByFrequency](sources/product/presentation.md#orderbyfrequency), [decisions.frequencyMetric](decisions.md#frequencymetric) · **Level:** code ·
**Method:** rendering test over the real dataset, reading the section headers

**Assertions:**

1. `?by=frequency` produces about nine sections, ascending, every edge a power of ten.
2. **Each tier holds exactly the species whose frequency falls in the band its header claims**,
   recomputed independently in the test.
3. The birds with no records in the selection are kept in **exactly one** final section.
4. `step=100` is strictly coarser than `step=10`.
5. **Every tier edge under a region filter also appears unfiltered** — the property the whole
   design exists for.
6. No tier header contains `common`, `scarce`, `rare`, `abundant`, `uncommon` or `everyday`.
7. Scale defaults to **global** under Frequency and **per-section** under the size metrics
   ([decisions.frequencyGlobalScale](decisions.md#frequencyglobalscale)), and **both** values remain selectable under both.
8. The Order by row **aligns on the text baseline**, so the label, the size buttons and the
   Frequency button sit on one line with the captions hanging below. Both the declaration and
   the class that carries it are asserted: either alone matches nothing.
9. The Order by row **states its direction for every metric** — *smallest first* under the
   three size metrics, *most common first* under Frequency, captioned under each group. Both
   stand on every view: a caption that followed the selection would report what the reader had
   just chosen, where two standing captions say the controls differ before they are clicked.

**Assertion 2 is the one with teeth, and it was added after the first version passed a
mutation it should have caught.** The header is written by `frequencyBand` and the banding done
by `SectionsBy`, so a tier can be labelled `1 in 1,000 – 1 in 9,999` while holding whatever the
sectioning rule actually put in it. Replacing the decade rule with a proportional one left
every label a tidy power of ten; only the tier *count* gave it away, and only because
`step=100` then stopped changing anything. Checking membership turns that into five explicit
failures — `tier "1 in 10 – 1 in 99" holds 102 species; 23 have a frequency in that band`.

**Status: verified 2026-10-05 — passing, and watched failing** under all of:

| Mutation | What went red |
| --- | --- |
| band proportionally (`min × 10`) rather than by decade | five tiers with the wrong membership, plus `100x gave 9 sections against 10x's 9` |
| restore `Inf >= Inf` in `SectionsBy` | `24 sections for the birds with no records; want exactly one` |
| name a tier `Common — 1 in …` | `a tier header says "common"` |
| drop the metric-dependent scale default | `/?by=frequency: scale "section", want "global"`, and the control showing Per section active |
| move Frequency into the size metrics' group | `1 button groups, want 2 — the size metrics and Frequency` |
| one caption that follows the selection instead of two standing ones | `1 direction captions, want 2` |
| drop `align-items: baseline` from the ordering row | `does not align on the baseline; its two-line groups will drag ORDER BY down level with the rules` |
| drop the `ordering` class from the row | `no row carries the ordering class, so the alignment rule matches nothing` |
| recognise only the non-default scale value, as the parser did before | `/?by=frequency&scale=section: scale "global", want "section"` — *Per section* is a dead button |

**The last of those is why both scale values are now parsed explicitly.** The old code read
only the non-default, with a comment warning that this disables the toggle whenever the
default moves — which is what happened when [decisions.perSectionDefault](decisions.md#persectiondefault) moved it. A default
that varies by metric is the same bug by a different road.

### sharedCountDisclosed

**A count covering two species says so, on the tile and on the page**

**Verifies:** [product.namesRecordsFiledUnder](sources/product/presentation.md#namesrecordsfiledunder), [decisions.sharedCountShownOnBoth](decisions.md#sharedcountshownonboth) · **Level:** code ·
**Method:** rendering test over the real dataset, both members of every declared shared name

**Assertions:**

1. For **both** species a shared name covers, the detail page carries a *Recorded as* row
   naming the source binomial and the other species.
2. The **index tile** carries the marker too, naming the other bird.
3. A species eBird **withholds** says so, and is **not** reported as a naming disagreement.
4. An ordinary species carries none of it.

**Why the index and not just the detail page.** The bug that prompted this was visible as a
ranking — a Caribbean mangrove resident sitting 51st most-recorded in North America — and a
note on a page nobody opened would not have caught it. The marker goes where the misreading
happens.

**Why assertion 3 is separate.** Withholding and renaming are indistinguishable in the data:
eBird returns zero for both. They are opposite in meaning, and conflating them would tell a
reader that eBird calls the Spotted Owl something else, sending them after a name that does
not exist. Nine species are in that state.

**Status: verified 2026-10-05 — passing, and watched failing** under both of:

| Mutation | What went red |
| --- | --- |
| the tile marker is suppressed, leaving only the detail page | `the index tile for yelwar1 (Setophaga aestiva) carries no shared-count marker` |
| a withheld species is labelled as a shared name | `/species/gyrfal (Gyrfalcon) does not say eBird withholds it` **and** `is withheld, not shared, but is reported as shared` |

### regionJoinReportsUnmatched

**The region join normalizes names, and reports what it could not match**

**Verifies:** [product.joinsReportDrops](sources/product/dataset.md#joinsreportdrops), [decisions.regionJoinNormalizes](decisions.md#regionjoinnormalizes) · **Level:** code and
build · **Method:** unit test over the shared normalization, plus a bounded report on every run

**Assertions:**

1. `taxonomy.Occurrence` maps **every** entry of `AviListRenames` and `CountedAs` onto its
   AviList name, from the lowercased form the occurrence data actually supplies.
2. A name in neither table passes through unchanged.
3. A **shared** name ([decisions.sharedCountShownOnBoth](decisions.md#sharedcountshownonboth)) is *not* folded — those are disclosed,
   not reassigned, and folding one would assert a division nobody measured.
4. `birdview-regions` lists every name matching no drawn species with its record count, and
   **refuses to write** above `maxOrphanFraction`.

**Why the bar is 0.1% and not the measured residue.** A bar set at the measured value fails
the next time iNaturalist gains a subspecies, and a check that cries wolf gets deleted. The
residue was 0.0431% when this was written and is **0.0113%** since the rank fold of
[occurrenceReadersAgree](#occurrencereadersagree), so 0.1% is now 8.8× it — room to grow, far below the 0.81% this exists to
catch.

**Status: verified 2026-10-05 — passing, and watched failing** under both of:

| Mutation | What went red |
| --- | --- |
| `Occurrence` skips `AviListRenames` — the shipped bug | `Occurrence("Coccothraustes vespertinus") = "coccothraustes vespertinus", want "hesperiphona vespertina"` |
| `Occurrence` folds the shared name into *aestiva* | `folded a shared name to "setophaga aestiva"; decisions.sharedCountShownOnBoth discloses those, it does not reassign them` |

**And one it caught unprompted.** The full `birdview-regions` run reverted
[decisions.honestSeaDenominator](decisions.md#honestseadenominator)'s off-land marine denominators, because that fix had only been
written into the `-marine-only` path. `TestRegionCountsLineUp` failed on all five seas —
*"a sea's denominator is every bird on that water, not a sum of the birds we happen to draw"* —
before the table was committed. Both paths now compute it the same way. This is the clearest
return the project has had on a check written to assert an invariant rather than a value.

### disagreementsDeclared

**Every species the two sources disagree about is declared**

**Verifies:** [product.sourceNamingMeasured](sources/product/dataset.md#sourcenamingmeasured) · **Level:** code · **Method:** offline test over the
committed correspondence table

**Assertions:**

1. Every drawn species has a row; the table is refused below 1,000 of them.
2. A species the two sources disagree about **must be declared** — in `AviListRenames`,
   `CountedAs`, `SharedNames`, `WithheldByEBird` or `DomesticForms`.
3. The failure message names which of those five it probably belongs in.

**The rule is asymmetric, because the sources are not peers.** iNaturalist is uniformly about
48× shallower than eBird, so what counts as surprising differs by direction:

| direction | rule | why |
| --- | --- | --- |
| iNaturalist out-records eBird | iNat ≥ 20 **and** 3× eBird | always wrong whichever name is at fault; a 48× expectation inverted threefold is a 144× departure |
| eBird dwarfs iNaturalist | eBird ≥ 10,000 **and** ratio > 1000× | ordinary at every scale, so it signals only with volume behind it |

**The floors are what make this a check rather than a noise generator.** The first version used
a plain 1×–1000× band and fired on **45 species, 43 of them vagrants** with a handful of
records where the shallower database happened to hold none — Plum-headed Parakeet at eBird 5,
iNaturalist 0, which at a 48× ratio is exactly what you would predict. Lesser Frigatebird sat
at 26 against 27 and "out-recorded" eBird by a single sighting. With the floors it flags **21
species and every one is declared.**

**Status: verified 2026-10-05 — passing, and watched failing** under all of:

| Mutation | What went red |
| --- | --- |
| drop Great Gray Owl from `WithheldByEBird` | `the two sources disagree about this name and nothing says why` |
| drop *Dryobates villosus* from `AviListRenames` | the same, on *Leuconotopicus villosus* |
| drop the Trogon entry from `SharedNames` | both halves at once — `Trogon ambiguus` 0/1,180 and `Trogon elegans` 34,913/0 |

**The third mutation is the one worth recording.** Its first attempt silently failed to apply
and the test "passed" — the same no-op trap that caught me earlier in this session, two hours
after I wrote the rule into `AGENTS.md` telling myself to assert that the mutation landed.
Re-run with the anchor asserted, it fails as above.

### occurrenceReadersAgree

**Both readers of the occurrence data land on the same species**

**Verifies:** [product.joinsReportDrops](sources/product/dataset.md#joinsreportdrops), [decisions.regionJoinNormalizes](decisions.md#regionjoinnormalizes) · **Level:** code ·
**Method:** unit test over the two pipelines, from one verbatim name to the name each joins on

**Assertions:**

1. For a verbatim occurrence name of each shape the facet actually contains — plain binomial,
   genus move, subspecies, subspecies *of* a genus move, domestic form, lump — the AviList
   species `birdview-status` files it under and the one `birdview-regions` files it under are
   the same name.
2. `taxonomy.Binomial` reduces a trinomial to its species and reports a hybrid and a bare
   genus as identifying none.
3. A hybrid is **not** reduced to its first parent — the rejection runs before the truncation.
4. The two commands' one permitted difference holds: `birdview-status` drops a name that
   identifies no species, and `birdview-regions` keeps it verbatim.

**Why the property and not the three bugs.** [decisions.regionJoinNormalizes](decisions.md#regionjoinnormalizes) was three
asymmetries found one at a time over two days, each invisible from inside the command that had
it right: `birdview-regions` applied neither `AviListRenames` nor `CountedAs`, and once it did,
still did not reduce a trinomial. Nothing compared the two readers, so every fix left the next
gap standing. Asserting that they agree closes the shape rather than the instances.

**Why it spans both namespaces end to end.** `birdview-status` resolves a genus move when it
*reads* its table and `birdview-regions` when it *writes* one. That is a real difference and
not a defect, so the criterion compares the name each finally joins on — a check pinned to an
intermediate value would report the difference as a bug every run.

**Why a hybrid is kept rather than dropped.** It is a bird somebody saw, so it belongs in the
denominator of [product.shareOfRecords](sources/product/presentation.md#shareofrecords)'s figure; it is no species BirdView draws, so it
belongs in nobody's numerator. Kept verbatim it is both, and it appears in the orphan report
under a name a reader recognizes as the deliberate loss it is.

**Status: verified 2026-10-06 — passing, and watched failing** under all of:

| Mutation | What went red |
| --- | --- |
| `Occurrence` does not reduce rank — the state before this change | the two readers disagree about three of the seven names, `"columba livia domestica"` / `"dryobates villosus harrisi"` / `"setophaga coronata auduboni"` each filed under itself |
| `Occurrence` consults the tables *before* reducing rank | only the renamed subspecies, `birdview-regions files "dryobates villosus harrisi" under "dryobates villosus", want "leuconotopicus villosus"` |
| the hybrid guard is removed from `Binomial` | `the hybrid test runs after the truncation, so these records would be filed as larus smithsonianus's`, plus `TestToBinomialsReducesTheFacetToSpecies` on the pintail |
| `Occurrence` drops a hybrid instead of keeping it | `birdview-regions turned the hybrid into "", want it left verbatim … it leaves the region's denominator short of birds that were genuinely seen` |

**A fifth mutation never applied**, and the harness said so instead of reporting a pass — the
anchor had been written against comment text that wrapped differently. That is the third time
this session a mutation silently matched nothing, and the first time the rule in `AGENTS.md`
caught it before it cost anything.

**What the fold actually did**, from the full run of 2026-10-06. The unmatched residue fell
from **0.0431% under 1,754 names to 0.0113% under 727**, and every one of the twenty largest
survivors is a hybrid, which is what should be left. **387,515 records moved into 430
species**, and **114 species-region presences were added with none removed** — the only shape
the change can have, since a trinomial either joins a drawn species or stays exactly where it
was. Region totals came back byte-identical, which is the independent confirmation that this
reattributes records rather than altering what they are divided by.

**The beneficiaries divide in two, and both were already the declared rule.** Genuine
subspecies identifications rejoining their species — Yellow-rumped Warbler +30,274 from
`coronata` and `auduboni`, Mallard +43,031, Red-tailed Hawk +14,973 — and domestic forms
rejoining their wild species, feral Rock Pigeon +109,401 and Muscovy Duck +37,470.
`birdview-status` had folded both since it was written; this is `birdview-regions` catching
up, not a new decision about either.

### figureSaysWhatItCounts

**The figure says what it counts, and it is not individual birds**

**Verifies:** [product.shareOfRecords](sources/product/presentation.md#shareofrecords) · **Level:** code · **Method:** unit test over the
rendered tile cell and the tooltip beside it

**Assertions:**

1. The index tile's `1 in N` is followed, **in the same cell**, by the unit `species
   observations`.
2. The tooltip that spells the figure out counts in the same noun, and no longer in `bird
   observations`.

**The noun was wrong in the one way a reader would not catch.** The denominator counts species
on checklists and deliberately ignores how many individuals each entry saw — twelve
White-throated Sparrows is one observation, and so is a flock of fifty thousand Snow Geese.
"bird observations" names precisely the quantity the figure is *not*, so someone reading
`1 in 32` as one bird in thirty-two has misread nothing except the unit, and the unit is the
only thing on the tile that could have told them.

**The cell, not the page.** `species observations` also appears in the help text and in
[`docs/data.md`](../../docs/data.md), so a whole-page search would pass with the figure absent
from every tile. The assertion matches the number and its unit as one pattern inside the
robin's own tile — the vacuous-check shape this project has now paid for about a dozen times.

**Status: verified 2026-10-06 — passing, and watched failing** under both of:

| Mutation | What went red |
| --- | --- |
| the tile template reverts to `bird observations` | `the tile counts in "bird observations", want "species observations"` |
| the tooltip reverts while the tile does not | `the tooltip reads "912,542 of the 1,252,475,518 bird observations from every region BirdView covers …", which does not count in species observations as the cell beside it does` |

### detailSaysWhatTileSays

**The detail page says everything the tile says**

**Verifies:** [product.tileOpensDetail](sources/product/presentation.md#tileopensdetail), [product.shareOfRecords](sources/product/presentation.md#shareofrecords) · **Level:** code ·
**Method:** property test comparing rendered tiles against rendered detail pages, under two
region scopes

**Assertions:**

1. For a sample of species across the index, the figure printed on the tile appears verbatim
   on that species' detail page.
2. It appears **under the scope the reader selected** — `(all regions)` unfiltered, `in
   Alaska` under `?region=AK` — not quietly swapped for another.
3. The same bird's figure **differs** between the two scopes, or both assertions above could
   be satisfied by a pair of pages that ignore the filter entirely.
4. The index parse is floored per scope — 1,000 tiles unfiltered, 300 in Alaska — so a regex
   that silently matched nothing cannot make the rest vacuous.

**The gap this closes.** [product.tileOpensDetail](sources/product/presentation.md#tileopensdetail) described the detail view as a larger
version of the tile "plus" three additions, and the *plus* was read as permission to add to
the detail page rather than as a floor under it. So when [product.shareOfRecords](sources/product/presentation.md#shareofrecords)'s figure
went onto every tile it never reached the detail page: a reader saw `1 in 23,393 species
observations (all regions)` on the Northern Yellow Warbler's tile, opened the bird, and found
only `Top region — 1 in 2,986 there`. A different number, in the same unit, for a scope they
had not asked about.

**The scope half is the sharper of the two.** A figure that is absent is visibly absent. A
figure that silently changes what it ranges over between one page and the next is a wrong
answer with nothing on either page to reveal the substitution — which is why assertion 2
exists and why the row is built by `shareWithSeams`, the same function the tile calls, rather
than recomputed. Two call sites deriving one number from the same inputs is exactly how the
index and the detail page came to disagree about where a bird is commonest
([decisions.honestSeaDenominator](decisions.md#honestseadenominator)).

**Status: verified 2026-10-07 — passing, and watched failing** under all of:

| Mutation | What went red |
| --- | --- |
| the `Frequency` row is not emitted — the state before this change | `calhum: the tile says "1 in 6,707 species observations (all regions)" and the detail page does not` |
| the row is built from `options{}`, ignoring the reader's selection | only the filtered scope: `coshum?region=AK: the tile says "1 in 38,567 species observations in Alaska" and the detail page does not` |
| the value drops `cell.Where`, printing the number without its scope | `the tile says "1 in 6,707 species observations (all regions)" and the detail page does not` |

**A fourth attempt was abandoned as uninformative**: swapping `shareWithSeams` for `shareOf`
produces an identical `Ratio` and `Where`, so it cannot fail this criterion. That is a true
statement about the two functions — they differ only in the seam disclosure, which
[sharedCountDisclosed](#sharedcountdisclosed) covers — and not a hole here.

### disputedNamesShown

**Every name the occurrence data disputes is named on the page**

**Verifies:** [product.namesRecordsFiledUnder](sources/product/presentation.md#namesrecordsfiledunder) · **Level:** code · **Method:** property test over the
five naming tables and the rendered detail pages

**Assertions:**

1. Every drawn species in `AviListRenames`, `CountedAs`, `SharedNames`, `DomesticForms` or
   `WithheldByEBird` carries a `Recorded as` row.
2. The row names the specific thing: the alias for a rename, the published binomial for a
   lump, the covering name for a shared name, `domestic and feral` for a form,
   `withheld by eBird` for a suppression.
3. **No species outside those tables carries the row.** 26 of 1,083 are in the tables — 27
   until [decisions.countEbirdFirst](decisions.md#countebirdfirst) derived `WithheldByEBird` from eBird's list and Desertas
   Petrel, on no entry of it, left — plus 24 that eBird withholds only in some regions or
   seasons ([withheldMonthsExplained](#withheldmonthsexplained)). The other 1,033 must stay silent.
4. The expected set is floored at 25 species, so emptying the tables cannot make the rest
   vacuous.

**Two of the four kinds were never implemented**, though [product.namesRecordsFiledUnder](sources/product/presentation.md#namesrecordsfiledunder) has named
the alias case since it was written — *"the page names the alias so a reader can find the bird
in the source."* A visitor who knows the Hairy Woodpecker as *Dryobates villosus* and finds a
page headed *Leuconotopicus villosus* could not tell a taxonomic update from a mistake, nor go
back to eBird and check. That is the reader this requirement exists for.

**Assertion 3 is what makes the other three mean something.** A disclosure satisfied by
printing it everywhere says nothing and trains a reader to skip it, and a one-directional
check would accept exactly that.

**The all-rows rule is not reachable by test today.** No species is in two of the five tables,
so a regression to the previous `if/else-if` chain would still pass. It is recorded in
[product.namesRecordsFiledUnder](sources/product/presentation.md#namesrecordsfiledunder) as the required shape rather than asserted here, because the
alternative — inventing an overlap in the real tables to exercise it — would put a false entry
in a published taxonomy to satisfy a test.

**Status: verified 2026-10-07 — passing, and watched failing** under all of:

| Mutation | What went red |
| --- | --- |
| the alias row is not emitted — two of the four kinds, as shipped until now | `Pica nuttallii (yebmag): the occurrence data disputes this name and the page is silent` |
| the lump row is not emitted | `Numenius hudsonicus (whimbr3)`, `Gygis candida (whiter4)`, `Saxicola maurus (y00427)` |
| a row is emitted for every species | the other direction: `Acanthis flammea (redpol1) is in none of the naming tables and still carries a "Recorded as" row` |

**Still open: the record counts.** [product.namesRecordsFiledUnder](sources/product/presentation.md#namesrecordsfiledunder) asks for each source's binomial
**and how many records it holds under it**, and only the binomial is on the page.
[`data/source-names.tsv`](../../data/source-names.tsv) holds per-source counts and nothing reads
it — but it counts under the *AviList* name only, so for an alias it reports eBird 0 and not
the 7,310,342 filed under *Dryobates villosus*. Closing this needs `cmd/birdview-names` to
query the alias too, and a run. Recorded in [todo.md](../todo.md).

### honestRevalidation

**Every response can be revalidated, and the validator is honest**

**Verifies:** [tech.cacheLifetimeStated](sources/tech/delivery.md#cachelifetimestated) · **Level:** code · **Method:** unit test over rendered
responses, their headers and their conditional round trip

**Assertions:**

1. Every response — pages and static files alike — carries an `ETag` and
   `Cache-Control: no-cache`.
2. Two different URLs never share an `ETag`.
3. Re-requesting with the returned `ETag` yields **304 with an empty body**, so the validator
   saves the thing it costs a header to offer.
4. A page renders **byte-identically** three times running, which is the assumption the
   fingerprint rests on.
5. Changing the dataset changes the index's `ETag`.

**Assertions 4 and 5 are the ones that matter**, because the ETag is computed from the build
and the URL rather than from the body. That substitute is what makes the scheme affordable —
hashing the body means buffering 3.3 MB of index per concurrent request — and it is sound
only if rendering is deterministic. A clock anywhere in the render path would make every
cached page wrong in a way no other check here would notice, so determinism is asserted and
not assumed. Assertion 5 closes the other end: without it a constant would satisfy the rest.

**What this fixes.** BirdView sent no `Cache-Control`, no `ETag` and no `Last-Modified` on
anything, which does not mean "do not cache" — it means the browser may apply its own
heuristic and has nothing to revalidate against when it does. It surfaced as a review round
on `/species/yelwar1`, reported as rendering old state when the server was serving the new
one correctly.

**Status: verified 2026-10-07 — passing, and watched failing** under all of:

| Mutation | What went red |
| --- | --- |
| no `Cache-Control` is set — the state before this change | `/: Cache-Control is "", want "no-cache"` |
| the ETag is built from the build alone, ignoring the URL | `/species/yelwar1 and / share the ETag "13b5953…"; a cache would serve one for the other` |
| the fingerprint is a constant rather than a hash of the build | `the index keeps ETag "bce5905…" after the dataset changed; every reader with a cached copy would keep it` |

### nameDisclosureLegible

**The name disclosure is legible and its link lands on the explanation**

**Verifies:** [product.namesRecordsFiledUnder](sources/product/presentation.md#namesrecordsfiledunder) · **Level:** code · **Method:** unit test over every
renamed species' rendered row

**Assertions:**

1. Every `Recorded as` row links to `/docs/data#when-the-sources-disagree-about-a-name`, and
   that anchor exists in the rendered page.
2. The row gives eBird's name **in full** — the actionable half, since it is what a reader
   types back into eBird.
3. Where the genus is unchanged the text says **epithet** and must not say *moved*.
4. Where the genus changed the text says **genus**.
5. The row names eBird.

**Both halves were wrong and both were reported.** Every row linked to *how common a bird is*,
the section about the frequency figure, while its own label read "how names are joined". And
the text said *"AviList moved this species to Pica nuttallii and eBird has not followed, so
BirdView looks the records up under Pica nuttalli"* — which for a one-letter difference looks
like the same name printed twice, and says *moved* about a species whose genus did not change.
Five of the seven renames are genus moves; two are corrected epithets, and one sentence was
being used for both.

**Why the row says eBird and not "the occurrence data".** Naming the sources generically would
be wrong in both directions, and the committed counts say so: iNaturalist follows AviList for
all five genus moves — 53,705 Hairy Woodpecker records under *Leuconotopicus villosus* — and
for the two spelling corrections **neither** database uses AviList's form. eBird is the single
claim true of all seven, holding nothing under the AviList name, which is why the join needs
the old one at all.

**Assertion 2 is narrower than it first was.** It began as "both binomials appear in the row",
which failed on the genus-move rows: those name the genus and the epithet separately and never
the full AviList string, because the full string is the page heading two inches above. Writing
the stricter rule would have forced repetition to satisfy a check rather than a reader.

**A boundary disagreement names both databases, and only calls itself a lump when it is one.**
`CountedAs` holds three entries and they are not the same taxonomic event: AviList **split**
the Whimbrel and the white terns and eBird has not adopted either split, while AviList
**declines to split** the Siberian Stonechat where both databases do. One sentence covering all
three named the wrong event for two of them, and asserted that AviList "does not carry"
*Numenius phaeopus* — which it does, as the Eurasian bird. The reason is therefore carried per
entry, and the criterion asserts each row names eBird *and* iNaturalist, says `lump` only for
*Saxicola maurus*, and says `does not carry` only where that is true.

**Status: verified 2026-10-07 — passing, and watched failing** under all of:

| Mutation | What went red |
| --- | --- |
| a boundary entry reverts to one vague sentence calling itself a lump | three at once — `does not say what iNaturalist holds`, `calls itself a lump, but AviList SPLIT this species`, and `says AviList does not carry Numenius phaeopus, which it does` |
| the link points back at `#how-common-a-bird-is` | `Hesperiphona vespertina: the row links somewhere other than #when-the-sources-disagree-about-a-name; its own label says "how names are joined"` |
| the epithet case reverts to the single "moved" sentence | `Pica nuttallii: same genus (Pica) and the text says "moved"` **and** `differs from Pica nuttalli only in the epithet (nuttallii against nuttalli) and the text never says so, which reads as the same name printed twice` |

**A third mutation never applied** — `gofmt` had realigned the struct field the anchor was
written against, so the substitution matched nothing. The harness refused instead of
reporting a pass. That is the second time this session and the same cause as the first.

### monthsFilterAndAbstain

**Months filter the list, scope the figure, and abstain where they cannot tell**

**Verifies:** [product.regionAndMonthFilter](sources/product/filtering.md#regionandmonthfilter), [product.shareOfRecords](sources/product/presentation.md#shareofrecords),
[decisions.monthPresence](decisions.md#monthpresence) · **Level:** code and build · **Method:** unit tests over the
presence rule and the rendered index, plus a refusal in the build

**Assertions:**

1. `month.Present` recovers the known season of seven species measured in New York —
   resident, partial migrant, winter visitor, long-distance migrant, breeder, passage
   shorebird — month for month.
2. It **abstains** below 200 records, and abstaining is distinguishable from twelve trues: the
   first says nothing is known, the second says the bird is there all year.
3. A month nobody birded in reads **present**, not absent: the missing records are everyone's.
4. All twelve months selected reproduce the annual figure **exactly**, for every species and
   for the denominator every tile divides by. `birdview-regions` refuses to write a table
   where the twelve do not reconstruct the year.
5. Months narrow the **list**, not only the figure: June shows fewer species than the year and
   a different set from January, while all twelve shows exactly what the year shows.
6. The scope label collapses runs and **wraps the year end** — a winter selection reads
   `Nov-Feb`, not `Jan, Feb, Nov, Dec`.

**Assertion 1 is the calibration, not a spot check.** The rule is effort-corrected because a
raw share would call every bird a May bird: an American Robin takes 22.1% of its New York
records in May and 2.6% in December, and is present in both. The seven profiles are measured
counts, so a change to the measure shows up as a wrong month for a named bird rather than as a
number moving.

**Assertion 3 is the half that is easy to get backwards.** A region with no birding in a month
says nothing about any species in it, so calling the bird absent would report an absence of
measurement as a measurement of absence — the distinction [product.undrawnSpeciesAnswer](sources/product/presentation.md#undrawnspeciesanswer) exists for.

**Assertion 6 earns its own function.** Winter is the selection a birder is most likely to
make and the one a renderer that walks January to December mangles worst.

**Status: verified 2026-10-07 — passing, and watched failing** under both of:

| Mutation | What went red |
| --- | --- |
| the ratio drops its effort correction, becoming a raw share of the species' own records | `Dark-eyed Junco, winter visitor: got +++++++--+++, want +++++++-++++` — September, where New York's birding is 7.8% of the year |
| the 200-record floor is removed | `judged a species on 28 records; below 200 those months record when anyone visited, not when the bird was there` |
| the month filter stops narrowing the list | `June shows 1083 of 1083 species; a month that admits everything is not filtering` |
| the denominator ignores the month selection | `Acanthis flammea in January: divided by 1252561983, want 87426358 — that month's birding, not another period's` |

**The abstention marker says why on the tile itself.** It first read *shown in every month*,
which states the opposite of what is meant: it reads as a claim that the bird is present all
year, the exact misreading abstaining exists to prevent. It now reads *Shown in every month due
to too few records*. Checked under `?region=HI&month=Jun`: Millerbird, 28 Hawaii records,
carries it, and Warbling White-eye, 143,057, does not. Watched failing on the old wording
(`Millerbird is shown ... without saying why`) and with the marker forced onto every bird
(`Warbling White-eye ... carries the too-few-records marker`).

**The fourth mutation survived the first draft of this criterion**, and it is the one that
would have mattered most. A month-blind denominator divides June's records by the whole
year's birding, so every figure reads about twelve times scarcer with nothing on the page to
explain it — and it passes assertion 4, because with all twelve months selected the two
halves agree again. The criterion now checks each single month's denominator against that
month's own birding, and that the twelve sum to the year on **both** halves of the ratio.

**A third mutation corrupted the file instead of reverting it**, because two were chained and
the second overwrote the first's backup. The package is new and untracked, so `git` could not
undo it and the damage had to be read off the diff by hand. Mutations are applied one at a
time from a known-clean tree.

### withheldMonthsExplained

**A month eBird withholds is shown and explained, not read as absence**

**Verifies:** [product.regionAndMonthFilter](sources/product/filtering.md#regionandmonthfilter), [product.namesRecordsFiledUnder](sources/product/presentation.md#namesrecordsfiledunder),
[decisions.ebirdSilenceNotAbsence](decisions.md#ebirdsilencenotabsence) · **Level:** code and build · **Method:** unit tests over the
restated list, the confirmation rule and the table format, plus rendered pages on the real
dataset

**Assertions:**

1. `taxonomy.EBirdSensitive` is well formed: every window parses, every region is a GADM id or a
   country, and every species in `WithheldByEBird` is on the list — except Desertas Petrel,
   which is not, and whose missing records are a naming question.
2. A window marks only the months it covers **entirely**, and may wrap the year end. Painted
   Bunting's Florida window ends April 15, and April — which still holds 12,483 eBird records —
   is not marked.
3. **A listing is acted on only where the records confirm it.** Painted Bunting in Florida and
   Long-eared Owl in New York are confirmed; Greater Sage-Grouse in Utah, whose records run
   higher inside its window than outside, and Spruce Grouse in Vermont, with 1,303 eBird records,
   are not.
4. The table's withheld flags round-trip, and malformed flags are refused rather than read as
   "nothing withheld".
5. Under `?region=FL&month=Jan` Painted Bunting is **shown** and its tile says *Withheld by eBird
   in Florida, Jan*; under `?region=NY&month=Jun` the same holds for Short-eared Owl. Without a
   month filter the Florida caveat reads *Oct-Mar*, wrapping the year end.
6. Greater Sage-Grouse in Wyoming in March carries **no** caveat: its records are in the export.
7. **Controls:** Bobolink in New York in January and Snow Bunting in New York in June stay
   hidden. They are not withheld, just gone.
8. The species page lists the withheld regions and months (product.tileOpensDetail).

**Assertion 7 is what keeps the fix honest.** Abstaining is easy to get right by abstaining
everywhere, which would quietly switch the month filter off; the controls are the birds it must
still hide.

**What the build found.** 202 region cells across 34 species are marked withheld. No count
changed: the table was rebuilt from the same cached facets as the one it replaces, and every
cell's twelve monthly figures are identical. The flags are new information, not a new
measurement.

**Status: verified 2026-10-09 — passing, and watched failing** under all of:

| Mutation | What went red |
| --- | --- |
| the month filter judges withheld months again — the bug as shipped | `paibun is hidden under /?region=FL&month=Jan, in a month eBird withholds it: missing by policy is not absent`, and the same for `sheowl` in New York in June |
| the caveat ignores the month selection | `paibun under /?region=FL&month=Jan does not say "Withheld by eBird in Florida, Jan"` |
| confirmation is skipped and the list trusted | `Greater Sage-Grouse, Utah: withheld -++++-------, want ------------` and `Spruce Grouse, Vermont: withheld ++++++++++++` |
| a half-covered month counts as withheld | `"Jan 01"-"Apr 15" covers ++++--------, want +++---------` |
| the parser drops the flags | `withheld = [false false …], want [true true true …]`, and `flags "!11100" were accepted` |

### ebirdCountedInatFillsGaps

**eBird is counted; iNaturalist fills only the gaps eBird declares**

**Verifies:** [product.scopeIsRecurringAvifauna](sources/product/dataset.md#scopeisrecurringavifauna), [product.regionAndMonthFilter](sources/product/filtering.md#regionandmonthfilter), [product.shareOfRecords](sources/product/presentation.md#shareofrecords),
[product.namesRecordsFiledUnder](sources/product/presentation.md#namesrecordsfiledunder), [decisions.countEbirdFirst](decisions.md#countebirdfirst) · **Level:** code and build ·
**Method:** unit tests over the counting rule in `birdview-regions` and `birdview-status`, the
derived `WithheldByEBird`, and the rebuilt tables

**Assertions:**

1. **eBird alone, as a rule.** In a region the size of New York, iNaturalist's robins are not
   added to eBird's, and an escaped Diamond Dove — New York's real 2 eBird records and 1
   iNaturalist photo — is counted at 2 and is **not** present.
2. **iNaturalist where eBird withholds.** Long-eared Owl in New York, silent on eBird and 39 on
   iNaturalist, is counted at 39, present on iNaturalist's own total, and flagged withheld all
   year. Painted Bunting in Florida takes iNaturalist's records in its withheld months and only
   those.
3. **Shared names.** Northern Yellow Warbler and Mangrove Yellow Warbler both carry eBird's
   shared count, and the region total counts it **once**.
4. **Scope.** An escape iNaturalist alone photographed in three different years keeps no dates;
   Long-eared Owl in New York keeps iNaturalist's; Painted Bunting keeps iNaturalist's January
   date and not its June one.
5. **`WithheldByEBird` is derived:** the nine global species plus Black Rail; not Long-eared Owl,
   withheld year-round only in the US; not Desertas Petrel, on no entry of eBird's list.

**Assertion 1 needed its fixture rebuilt at New York's real size.** The presence bar is relative,
and at 1.2 million records two Diamond Doves clear 1 per million — so the first fixture failed
against correct code, and the first version of the escape case, with no eBird records at all,
passed with the old presence rule put back, because a name with nothing counted is never judged.
Both are recorded so the next fixture is not written into either.

**What the build found.** Regions were rebuilt from the same cached facets as before, so every
difference is the rule:

| | Before | After |
| --- | --- | --- |
| land presences | 27,831 | 26,848 |
| species placed in a region | 1,062 | 1,035 |
| records counted (all regions) | 1,252,561,983 | 1,237,352,304 (−1.21%) |
| Spotted Owl regions | 10 | 10 |
| Long-eared Owl regions | 51 | 50 |
| Northern / Mangrove Yellow Warbler | 1 in 23,244 / 1 in 203 | 1 in 200 / 1 in 200 |

The 26 species that leave every region are 24 escapes — Diamond Dove in 8, Silver Pheasant in 10,
Green-cheeked Parakeet in 7, Emu in 6, Lady Amherst's Pheasant in 5 — and two genuine vagrants
with very few records, Ruddy Quail-Dove (16) and Ferruginous Duck (7). Red Junglefowl falls from 59
regions to 17. None newly enters a region.

**Scope barely moved**, which was the open question: 1,083 drawn species to 1,082. **Mottled
Owl** leaves, rare to vagrant, on losing two iNaturalist dates (1,133 records to 1,131); Vireo
swainsoni leaves scope too, but AviList does not carry it and it was dropped from the dataset
already. The escapes do **not** leave scope: eBird records the feral populations itself, so they
recur on eBird's evidence, and only their region presence narrows. Measuring this took a
`birdview-status` run of some 8,600 requests that stalled for two days under GBIF throttling and
ended with 15 queries failed; the resume fetched the missing 15 in under a minute.

Two things the rebuild caught, both fixed: Northern Yellow Warbler and Coppery-tailed Trogon would
have **left scope**, since their records sit under the shared name and their own rows no longer
recur — `status.Scope` now brings both halves of a shared name in; and the ingest refused to run,
because its check that *Vireo swainsoni* is dropped by the AviList join fired on a species that
correctly never reached the join.

**Status: verified 2026-10-12 — passing, and watched failing** under all of:

| Mutation | What went red |
| --- | --- |
| iNaturalist counted everywhere | `robin counts 62412000, want 62,400,000 -- eBird's alone` |
| the old per-source presence rule | `an escaped Diamond Dove is present in New York on one iNaturalist photo` |
| no shared-name copy | `aestiva 0, petechia 6000; both should carry eBird's shared 6,000` |
| the shared count added to the total twice | `total 62412041, want 62406041` |
| scope reverts to the union of both sources' dates | `an iNaturalist-only escape kept 3 dates` and `Painted Bunting's June iNaturalist date was counted` |
| `status.Scope` ignores shared names | `Setophaga aestiva is out of scope` |
| `WithheldByEBird` from global entries only | `WithheldByEBird[Laterallus jamaicensis] = false, want true` |

### pipelineDiagramServed

**The pipeline diagram is served, current, self-contained and linked**

**Verifies:** [product.pipelineDiagram](sources/product/measurements.md#pipelinediagram), [product.categoryProvenance](sources/product/measurements.md#categoryprovenance), [tech.selfContainedServer](sources/tech/delivery.md#selfcontainedserver) ·
**Level:** code · **Method:** unit tests over the docs embed and the served pages

**Assertions:**

1. **Current.** Every page in `docs/` with a Mermaid block has an SVG of the same name, whose
   first line records the SHA-256 of that block.
2. **Self-contained.** The SVG carries no `<script>`, no `<foreignObject>`, no event handler and
   no URL other than XML namespace declarations.
3. **Served.** `/docs/pipeline` shows the picture, linked to itself full size, in place of the
   Mermaid source; `/docs/pipeline.svg` answers as `image/svg+xml`; the index lists the page and
   not the picture.
4. **Faithful to the readers.** Every column the diagram names for AviList, AVONET and the Fox
   wingspans is a column the reader code asks for, so a renamed or dropped column cannot leave
   the diagram describing a pipeline that no longer exists.
5. **Linked** from `docs/data.md`, `docs/building.md` and `docs/using.md` as served, from the
   source list at the foot of the index, and from the README.
6. **One field per line, one mark per field.** Every box with fields lists one field per line,
   followed by exactly one of ★ · ✗ and an optional note in italics; the lines under a box's name
   are italic. A box with fields written as a plain label, which cannot carry italics, fails.
7. **Tables read left-aligned.** Every box with fields, and the legend's key to the marks, is
   left-aligned in the rendered SVG; steps and commands stay centered. Mermaid
   centers every label, so the renderer moves them, and a render that skipped it would still
   match its source's hash.
8. **Every taxonomy table the ingest reads is in the diagram.** Each `taxonomy.X` that
   `cmd/birdview-ingest` refers to is named in `docs/pipeline.md`, except a listed few that are
   not data: the inverse of a table the diagram already names, and a list printed in the build
   report. A table added to the ingest without a line in the diagram fails.

**The column check found two mislabeled fields before it existed.** Writing the diagram from the
readers turned up AVONET's `Family1`–`3` and `Secondary1`, both read and neither used — `FamilyOf`
was kept to cross-check ABA's families, and ABA is gone — and a first draft that called
`Trophic.Level` a fallback for a missing niche, when it adds a second diet where AVONET's level says
Omnivore and its niche does not. The check itself first failed on its own diagram, on a field label
wrapped onto a second line; one field per line is now the rule it relies on.

**Assertion 6 came from review.** The first diagram used · both as a mark and as a separator
between fields on one line, so "species · region · date window" read as one field with two marks,
and the legend's "one field per line" held for some boxes and not others. Rewriting every box to
the rule also found what the diagram had left out: the server reads `internal/taxonomy`'s tables
directly for *Recorded as* and the withheld notes, three generators read the previous build's
`species.json` for their species list, and `birdview-gbifmap` reads AviList.

**Status: verified 2026-10-12 — passing, and watched failing** under all of:

| Mutation | What went red |
| --- | --- |
| a field label edited, the SVG not re-rendered | `docs/pipeline.svg was drawn from a different source than docs/pipeline.md holds` |
| the SVGs dropped from the embed | `has a diagram and no docs/pipeline.svg`, `no SVG is embedded`, and `/docs/pipeline.svg answers 404` |
| a `<script>` added to the SVG | `docs/pipeline.svg carries <script` |
| an `<image>` from another host added | `refers off the server, to https://example.com/x.png` |
| an `onload` handler added | `carries an event handler: onload=` |
| the Mermaid block left in place | `/docs/pipeline shows the Mermaid source instead of the picture` |
| the index lists every embedded file | `the docs index lists the picture as a document of its own` |
| the SVG served without its content type | `answers 200 as "text/html; charset=utf-8"` |
| AVONET's `Order3` renamed in the reader | `names AVO column "Order3", which ... avonet.go never asks for` |
| Fox's `wingspan_reported_min_all` renamed | `names FOX column "wingspan_reported_min_all"` |
| the link dropped from `docs/data.md` | `/docs/data does not link to the pipeline diagram` |
| the link dropped from the index's source list | `/ does not link to the pipeline diagram` |
| both README links pointed elsewhere | `README.md does not link to docs/pipeline.md` |
| three fields put back on one line, `species · region · date window` | `SENS: "species · region · date window" is not one field, one mark and an optional italic note` |
| a note's italics removed | `AVL: "Taxon_rank · keep species rows" is not one field, ...` |
| none: assertion 8 was written before the diagram caught up with [decisions.whiteTernBorrows](decisions.md#whiteternborrows) and [decisions.herringGullEuropeanRow](decisions.md#herringgulleuropeanrow) | `the ingest reads taxonomy.OwnRowSheets, which docs/pipeline.md does not name`, and the same for `NearMatchSheets` (2026-10-13) |
| a field's mark removed | `AVO: "Habitat" is not one field, ...` |
| the line under a box's name not italic | `SENS: "web page" under the box's name is not italic` |
| a box written as a plain label | `a box with fields is a plain label, so its notes cannot be italic` |
| the renderer's left-align step removed, and the SVG re-rendered | `docs/pipeline.svg left-aligns 0 tables of the 19 in docs/pipeline.md` |
| the legend's key left out of the renderer's rule, and the SVG re-rendered | `left-aligns 19 boxes of the 20 tables and keys` |

**Two mutations survived, and are recorded so they are not mistaken for coverage.** Removing the
bare-name guard on `.svg` paths changes nothing observable: `embed.FS` refuses `..` and finds no
nested file, so the guard repeats a refusal the filesystem already makes. Removing `Secondary1`
from AVONET's list of required columns survives assertion 4, which asks only that the reader name
the column, and the reader still names it where it reads the value.

### whiteTernFromBirdlife

**Blue-billed White-Tern borrows from BirdLife's White Tern**

**Verifies:** [product.lengthAndWingspanSources](sources/product/measurements.md#lengthandwingspansources), [product.provenancePerMetric](sources/product/measurements.md#provenancepermetric) · **Level:** code ·
**Method:** unit test over ingest output

**Assertions:** *Gygis candida*'s bill and tail are `near-match`, with `sourceTaxon` *Gygis alba*
and `source` naming `AVONET1_BirdLife`, and carry that row's values, 49.7 mm and 103.8 mm, read
from the workbook rather than written into the test. Vega Gull borrows `AVONET1_BirdLife`'s
*Larus smithsonianus* ([decisions.vegaGullBorrows](decisions.md#vegagullborrows)). Every other near-match species still
borrows from `AVONET2_eBird`. A sheet named in `taxonomy.NearMatchSheets` that does not hold the parent
fails the build instead of falling back to another sheet.

**Status: verified 2026-10-13 — passing, and watched failing** under all of:

| Mutation | What went red |
| --- | --- |
| the `Gygis candida` entry removed from `NearMatchSheets` | `Gygis candida has bill 44.4, tail 90.2; BirdLife Gygis alba has 49.7, 103.8` |
| `resolve` ignoring the declared sheet | `Gygis candida bill comes from "AVONET v7, sheet AVONET2_eBird", want sheet AVONET1_BirdLife` |
| the missing-row check in `checkNearMatchSheets` disabled | `a near-match sheet that holds no row for the parent was accepted` |
| `Larus vegae` also declared for `AVONET1_BirdLife` | `Larus vegae bill comes from "AVONET v7, sheet AVONET1_BirdLife", want sheet AVONET2_eBird` |

**The last mutation first survived.** The test took each species' expected sheet from
`NearMatchSheets`, the table under test, so adding a second entry moved the expectation along with
the code. The expected sheet is now written in the test.

### sharedMeasurementsDisclosed

**Species drawn from the same measurements say so**

**Verifies:** [product.borrowedMeasurementsSaySo](sources/product/measurements.md#borrowedmeasurementssayso) · **Level:** code · **Method:** unit test over ingest
output, and rendering tests over the index and the detail page

**Assertions:**

1. **The groups are derived.** Grouping the built dataset by the bill's `source` and
   `sourceTaxon` gives exactly the pairs the page notes, and includes Vega Gull with American
   Herring Gull, both at `AVONET1_BirdLife` *Larus smithsonianus*, with the same bill, tail, mass
   and wingspan.
2. **Every tile in a group says so.** The tile of each member carries *Same measurements as* and
   the English name of every other member, inside that tile's own markup. A tile outside every
   group carries no such line.
3. **Every detail page in a group says so,** naming the other members, the source sheet and the
   source taxon.
4. **No hand table.** A species given a near-match to a drawn relative in a test dataset gets
   the note on both tiles without any other change.

**Status: verified 2026-10-13 — passing, and watched failing** under all of:

| Mutation | What went red |
| --- | --- |
| `sharedRows` replaced by an empty map | `the index carries 0 shared-measurement notes, want 8`, and both fixture tiles unnamed |
| rows grouped by sheet alone, not sheet and taxon | `the index carries 1125004 shared-measurement notes, want 8`; `Corvus brachyrhynchos shares no row, but its tile says it does` |
| the tile's note deleted from `index.html` | `the tile for Sula leucogaster does not say it shares measurements with Cocos Booby` |
| the detail page's note deleted from `detail.html` | `/species/veggul1 does not name the shared row (Larus smithsonianus, sheet AVONET1_BirdLife)` |
| the detail note naming the taxon but not the sheet | the same, for all eight pages |
| Vega Gull mapped back to *Larus argentatus* | `Larus vegae is drawn from "AVONET v7, sheet AVONET2_eBird" "Larus argentatus"`; `Vega Gull's bill is 70.5 and American Herring Gull's 69.3` |

**The first template mutations failed for the wrong reason.** Renaming the field the template
ranges over, rather than deleting the markup, made template execution fail, so pages errored and
the tile checks reported tiles missing altogether. The table records the clean deletions.

### herringGullFromEuropeanRow

**European Herring Gull is read from BirdLife's European-only row**

**Verifies:** [product.lengthAndWingspanSources](sources/product/measurements.md#lengthandwingspansources), [product.provenancePerMetric](sources/product/measurements.md#provenancepermetric) · **Level:** code ·
**Method:** unit test over ingest output

**Assertions:** *Larus argentatus*'s bill, tail and mass are `measured`, with `source` naming
`AVONET1_BirdLife` and `sourceTaxon` *Larus argentatus*, and carry that row's values, read from
the workbook. The workbook's eBird and BirdLife rows must differ, or the check cannot tell them
apart. American Herring Gull is unchanged: `AVONET1_BirdLife` *Larus smithsonianus*, `measured`.
A sheet named in `taxonomy.OwnRowSheets` that does not hold the species fails the build instead
of falling back to the eBird sheet.

**Status: verified 2026-10-13 — passing, and watched failing** under all of:

| Mutation | What went red |
| --- | --- |
| the `Larus argentatus` entry removed from `OwnRowSheets` | `Larus argentatus bill is measured from "AVONET v7, sheet AVONET2_eBird"`; `European Herring Gull has bill 70.5, tail 168.5; BirdLife Larus argentatus has 74, 170.2` |
| `resolve` ignoring the declared sheet | the same |
| the missing-row check in `checkOwnRowSheets` disabled | `an own-row sheet that holds no row for the species was accepted` |
| the own row labeled `near-match` | `Larus argentatus bill is near-match from "AVONET v7, sheet AVONET1_BirdLife"` |

### belowBarCountsKept

**Records below the presence bar are counted**

**Verifies:** [product.regionAndMonthFilter](sources/product/filtering.md#regionandmonthfilter), [product.shareOfRecords](sources/product/presentation.md#shareofrecords) · **Level:** code ·
**Method:** unit tests on the table's reader and writer, and over the built dataset and the
rendered index

**Assertions:**

1. **The table keeps them.** A cell marked below the bar round-trips through the writer and
   `region.LoadPresence` with its counts, its months and its mark, and an unmarked cell reads as
   present.
2. **The ingest separates them.** A below-bar cell never appears in `regions`, so region and
   month filtering are unchanged, and its counts appear in the species' below-bar fields.
3. **The figure counts them.** With every region selected, a species' figure is its records in
   every region it has any, present or not. Emu's is not zero, and no drawn species with records
   reads zero.

**Status: verified 2026-10-13 — passing, and watched failing** under all of:

| Mutation | What went red |
| --- | --- |
| none: the real-data test run against the table built before this change | `no species has a region below the bar; the table was built without them` |
| the writer dropping the `~` mark | `the below-bar cell read back as {… Below:false}` |
| the reader ignoring the mark | the same, and `regions [USA.44_1 USA.5_1] counts [19 12], want only USA.5_1 with 12` |
| the ingest listing below-bar regions with the rest | `regions [USA.44_1 USA.5_1] counts [19 12]`; `below regions [] …, want USA.44_1 with 19` |
| the figure ignoring the below-bar counts | 717 species, e.g. `Acanthis flammea: the figure counts 902200 records, want 902219` |

The rebuilt table, from GBIF on 2026-10-13, holds 26,847 listed cells and 3,689 below the bar.
Emu has 19 records in Texas and 14 regions in all, every one below the bar, and is listed under
none. Of the 47 drawn species that read zero before, only Desertas Petrel still does: it has no
open record at all ([product.declaredExceptions](sources/product/dataset.md#declaredexceptions)).

### escapedBirdsListedAndServed

**Escaped and kept birds are off the index, listed and still served**

**Verifies:** [product.escapedBirdsListed](sources/product/measurements.md#escapedbirdslisted), [product.statusIsPublished](sources/product/dataset.md#statusispublished) · **Level:** code ·
**Method:** unit test over ingest output, and rendering tests over the index, `/excluded` and
the detail pages

**Assertions:**

1. **Flagged from the list.** Every species in `status.Exotic` is drawn and carries its reason as
   `exotic`; no other species carries one. A listed name that is not drawn fails the build.
2. **Off the index.** No flagged species has a tile on the index, and none is offered or counted
   by the family, habitat, diet or region controls. A family whose every species is flagged is
   not offered.
3. **Listed.** `/excluded` names every flagged species with its reason and links to its detail
   page, and the index links to `/excluded`.
4. **Still served.** Each flagged species' detail page answers 200, says it is not on the index
   with its reason, and links to `/excluded`.

**Status: verified 2026-10-14 — passing, and watched failing** under all of:

| Mutation | What went red |
| --- | --- |
| the index view keeping the exotics | `Amadina fasciata is flagged exotic but has a tile on the index`, and the reported count |
| families of flagged birds only still offered | `family Casuariidae holds only flagged birds but is still offered`, and Ramphastidae |
| `/excluded` reading the index view instead of the whole dataset | `/excluded does not link Amadina fasciata` |
| the detail page's note deleted | `/species/blhwea1 does not say it is off the index, why, and where the list is` |
| the index's link to `/excluded` removed | `the index does not link to /excluded` |
| the ingest not setting the flag | `Amadina fasciata carries exotic "", want "A cage bird from Africa; …"` |

The first try at the third mutation changed the slice indexed but not the loop over it, and
panicked rather than failing; the table records the corrected mutation. The list removes 41
birds: the index shows 1,041 of 1,082, and Casuariidae and Ramphastidae leave the family control.

### readersToldStickFigure

**Readers are told the drawing is a stick figure**

**Verifies:** [product.segmentComposition](sources/product/diagram.md#segmentcomposition), [decisions.stickFigure](decisions.md#stickfigure) · **Level:** code ·
**Method:** rendering test over the visible text of the index, a detail page, the comparison page
and the using guide

**Assertions:**

1. No visible text on the index, a detail page or the comparison page says *diagram* or
   *diagrams*, except the link to the data-pipeline diagram. Class names, template comments and
   scripts are not visible text and are not checked.
2. The index and `docs/using.md` call the drawing a *stick figure*.

**The check found one the edit had missed:** the detail page's provenance for total length read
*sum of the five diagram segments*, a model string written by the ingest rather than the
templates. It now reads *sum of the five drawn segments*.

**Status: verified 2026-10-15 — passing, and watched failing** under all of:

| Mutation | What went red |
| --- | --- |
| the index's heading back to *The diagram* | `/ says "what they do not The diagram Every bird is drawn as a"` |
| the figure key back to *lines in the diagram* | the same line, on `/species/amerob` and `/compare` |
| every *stick figure* in `docs/using.md` replaced | `the using guide never calls the drawing a stick figure` |
| the pipeline link no longer exempt | `/ says "... is set out in one diagram ."` |

**Two mutations survived, and were replaced by the ones above.** Removing one *stick figure* from
the using guide leaves another, so assertion 2 still holds; and making the pipeline-link
exemption greedy only widens it past the last *diagram* on the page, so nothing new is hidden.
