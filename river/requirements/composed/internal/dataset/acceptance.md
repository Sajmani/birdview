package dataset

# dataset — acceptance criteria

**Composed, not authored.** Composed with [dataset.md](dataset.md); do not edit by hand, recompose.

How each requirement in [dataset.md](dataset.md) is decided. Each criterion names what it covers in
`Covers:`; mutable guardrail bindings and `Watched failing:` observations live in
[conformance.md](../../../../conformance.md).

---

## legibleDistinctColors

**Generated colors are legible and mutually distinguishable**

Covers: [product.generatedBadgeColors](../../../local/product/dataset.md#generatedbadgecolors)

**Method:** pure functions
over the color generator

**Assertions:** WCAG 2.2 contrast ≥ 7:1 against each badge's **own** text color; CIE76 ΔE ≥ 25
between families adjacent in the selector; both halves of the gamut are used; and the preset cycle
strictly alternates light and dark, which requires an even count

## scopeMatchesStatusTable

**Scope is what the status table says recurs**

Covers: [product.scopeIsRecurringAvifauna](../../../local/product/dataset.md#scopeisrecurringavifauna), [product.declaredExceptions](../../../local/product/dataset.md#declaredexceptions)

**Method:** unit test over ingest output

**Assertion:** the dataset contains every species `data/status.tsv` marks in scope (`established` or
`rare`, plus the declared exceptions of [product.declaredExceptions](../../../local/product/dataset.md#declaredexceptions)), less those AviList has no concept for and the four
species with no AVONET measurement — **1,082** species today — derived from the inputs rather than
hardcoded from a run.

## noRarityCodes

**Rarity codes never reach the dataset**

Covers: [product.scopeIsRecurringAvifauna](../../../local/product/dataset.md#scopeisrecurringavifauna)

**Method:** inspect the serialized record's field names
for anything carrying an ABA rarity code

## allFiveTraits

**Every record carries all five AVONET traits**

Covers: [product.traitsFromAvonet](../../../local/product/dataset.md#traitsfromavonet), [product.traitsPerSpecies](../../../local/product/dataset.md#traitsperspecies)

**Method:** property test over every record

**Assertion:** culmen, tail, wing chord, secondary, and mass are all present and positive.

## namesNeverDefaulted

**Names and families are populated from AviList, never defaulted**

Covers: [product.namesFromAvilist](../../../local/product/dataset.md#namesfromavilist)

**Method:** property test over the built dataset

**Assertions, names:** the **world** name is non-empty for every species; the **North American**
name is non-empty *after* the documented fallback, so a species with no Clements concept shows a
bird rather than a blank; family scientific and English names are non-empty; and **no record's
name equals its scientific name** — the tell of a silent fallback, which an earlier design
shipped for every one of its species.

**Assertions, the code:** every species carries an eBird code. AviList gives none to exactly
**three**, and they are **named, not counted** — *Oxyura jamaicensis*, *Pterodroma cervicalis*,
*Tyto furcata*, each a taxon AviList recognizes and Clements has no 1:1 counterpart for — and
`taxonomy.EBirdCodes` declares theirs as AviList's gaps and no other species' as one
([decisions.ebirdCodesComplete](../../../decisions.md#ebirdcodescomplete)). A fourth gap appearing means the join moved, and shows as a
species with no code at all, which a percentage threshold would have absorbed in silence.

## joinCoversEverySpecies

**The join covers every species, and the direct-match rate is monitored**

Covers: [product.joinOnBinomial](../../../local/product/dataset.md#joinonbinomial)

**Method:** unit test over ingest

**Assertions:** (a) after the mapping table is applied, every in-scope species (except the 4 named
in `unmeasurableNames`) resolves to a trait row; (b) `TestNearMatchesAreExactlyTheMappedOnes`
verifies that the 9 near-matches in the **1,082**-species dataset match `taxonomy.NearMatches`.

## unmatchedFailsBuild

**Unmatched species fail the build above a threshold**

Covers: [product.unmatchedSpeciesReported](../../../local/product/dataset.md#unmatchedspeciesreported)

**Method:** unit test with a fixture containing a bogus
binomial

**Assertion:** ingest exits non-zero and names the unmatched species.

## neckLengthSerialized

**The neck length is serialized with its model named**

Covers: [product.neckModelPerFamily](../../../local/product/diagram.md#neckmodelperfamily), [product.provenancePerMetric](../../../local/product/measurements.md#provenancepermetric)

**Method:** property test over every record

**Assertions:** every record carries a neck length; its provenance `kind` is `modeled`; its
`model` is non-empty and **names both the per-family ratio function and the numeric ratio applied
to that species' family**; and the ratio named matches the table entry for that family.

## provenanceWellFormed

**Every presented metric has a well-formed provenance record**

Covers: [product.provenancePerMetric](../../../local/product/measurements.md#provenancepermetric)

**Method:** schema validation plus a property test over
every record

**Assertions:** every metric carries `kind` in `{measured, near-match, modeled}`; `source` is
non-empty; `sourceTaxon` is non-empty for `measured` and `near-match`; `model` is non-empty for
`modeled`; `sourceTaxon` differs from the species' own name whenever `kind` is `near-match`.

## provenanceMatchesSources

**Provenance matches the sources actually used**

Covers: [product.lengthAndWingspanSources](../../../local/product/measurements.md#lengthandwingspansources), [product.provenancePerMetric](../../../local/product/measurements.md#provenancepermetric)

**Method:** unit test over ingest output

**Assertions:** every length has `kind = modeled`; exactly the species with an empirical Fox value
have `kind = measured` for wingspan; the six species named in
[decisions.abaSpeciesWithoutAvonet](../../../decisions.md#abaspecieswithoutavonet) have `kind = near-match` for their AVONET-derived traits.

## totalIsSumOfSegments

**Total length is reported as the sum of its segments**

Covers: [product.lengthAndWingspanSources](../../../local/product/measurements.md#lengthandwingspansources), [product.provenancePerMetric](../../../local/product/measurements.md#provenancepermetric)

**Method:** property test

**Assertions:** total length's `kind` is `modeled`; its `model` states that it is the sum of the
five segments; and the value equals that sum exactly (the same invariant segmentsSumExactly asserts from the
geometry side).

## wingspanModelReconstructsFox

**The fitted wingspan model reconstructs Fox's published one**

Covers: [product.lengthAndWingspanSources](../../../local/product/measurements.md#lengthandwingspansources), [product.provenancePerMetric](../../../local/product/measurements.md#provenancepermetric)

**Method:** fit the model, then compare its
estimate against `wingspan_predicted_*` for every species Fox predicted

**Assertions:** mean disagreement ≤ 2%; no more than 2% of species disagree by over 10%.

## outboundRequestsIdentified

**Outbound requests identify the client**

Covers: [tech.identifiedClient](../../../local/tech/delivery.md#identifiedclient), [gbifOccurrenceAndApiTerms](dataset.md#gbifoccurrenceandapiterms)

**Method:** unit test against an `httptest` server

**Assertion:** the `User-Agent` header is present, non-empty, and is not Go's default.

## craniumVaries

**The cranium varies with the bird**

Covers: [product.craniumFromMass](../../../local/product/diagram.md#craniumfrommass), [product.provenancePerMetric](../../../local/product/measurements.md#provenancepermetric)

**Method:** unit test on the model plus a
distribution check over the built dataset

**Assertions:** head length matches `7.5 × mass^(1/3)` at stated points, grows strictly across the
real mass range with no plateau, no single value is shared by more than 5% of species, and the
heaviest bird has the largest head.

## trunkScalesLikeSkeletons

**The trunk scales the way measured skeletons do, and is anchored where it was**

Covers: [product.trunkFromMass](../../../local/product/diagram.md#trunkfrommass)

**Method:** recover the model's own exponent by regression
over the real mass range, then check it against an interval taken from published measurement

**Assertions, the exponent:** fitting log trunk on log mass across the dataset's mass range
recovers an exponent inside **0.338 – 0.401** — the intersection of the 95% confidence intervals
[`bohmer-neck`](../../../imports.md#bohmer-neck) reports for neck length and total leg length, the two
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
claim [product.trunkFromMass](../../../local/product/diagram.md#trunkfrommass)'s rationale actually makes.

## habitatReviewComplete

**Every drawn species has a habitat review, current with AVONET, and only reviewed additions count**

Covers: [product.habitatAndDietFilters](../../../local/product/months.md#habitatanddietfilters)

**Method:** the review table parsed, then held against the built dataset

**Assertions:** every species BirdView draws has a row in `data/habitat-review.tsv` and no row names
a species it does not draw; each row's AVONET habitat is AVONET's today; a reviewed row's additions
reach the species as BirdView's, with their reasons, and a proposed or pending row's do not; and a
row with an unknown status, a reviewed row without reviewer or date, an addition outside AVONET's
vocabulary, or one without a reason is refused.

## gbifDownloadsRecorded

**A GBIF download is fetched and recorded only under the query it was requested from**

Covers: [tech.gbifDownloadsRecorded](../../../local/tech/process.md#gbifdownloadsrecorded), [gbifOccurrenceAndApiTerms](dataset.md#gbifoccurrenceandapiterms)

**Method:** the download client against a fake GBIF

**Assertions:** a query is validated without credentials and GBIF's formatting of it returned; a
rejected query fails validation; submitting sends the credentials and returns the key; waiting polls
until the download succeeds; fetching returns the file and its SHA-256; the record carries the key,
DOI, creation date, license, query file, and hash, and never the password; and a download is not
recorded if it has not succeeded or was requested with SQL other than the committed query, in either
its committed form or GBIF's formatting of it.

## habitatDietFilterPerSpecies

**Habitat and diet filter per species, and say where each value came from**

Covers: [product.habitatAndDietFilters](../../../local/product/months.md#habitatanddietfilters)

**Method:** unit tests over the trait table against the
compiled dataset, plus rendering tests over the controls

**Assertions, the vocabulary:** every value assigned to a species is one the control offers, so
a bird cannot carry a category nobody can select; every option has a label and a quoted
definition; and the jargon terms — `Invertivore`, `Vertivore`, `Granivore` and the rest — are
never shown untranslated.

**Assertions, origin:** every assignment carries an origin. A published one names its source,
a `birdview` one gives its reason. An assignment with no origin fails: that is the
opinion-as-finding failure [product.categoryProvenance](../../../local/product/measurements.md#categoryprovenance) exists to prevent.

**Assertions, the supplementary table:** **every key is a species in the dataset.** This is the
assertion the table most needs — a key AviList does not use misses silently, adds nothing, and
leaves the table looking correct. A second assertion counts the assignments that actually reach
the dataset against the number the table defines, so a miss surfaces as a number rather than as
nothing at all.

**Assertions, the control:** the Species panel offered Habitat and Diet before the taxonomic
family list, each option carrying its count and its definition, and the retired birder grouping
was not rendered. That check went with the classic view on 2026-10-27; the home page's Habitat
and Diet settings are decided through [app.nextViewDrawsOneSet](../../../composed/internal/app/acceptance.md#nextviewdrawsoneset) (`TestNext*`).

**Assertions, the additions page:** `/added` lists **every** assignment that is not published,
counted against the dataset rather than eyeballed — a page that quietly lists a subset is worse
than no page. It lists no published assignment, those being AVONET's rather than BirdView's to
claim, and every row names its origin and its reason.

**Assertions, coined terms:** a term BirdView added where AVONET has none says so in its own
definition and is marked in the control; it can only be assigned with origin `birdview`, since
a category this project invented cannot be published by anyone; and the number of them is
**capped**, because the difference between filling a gap and acquiring a habit is a count.

## duckAndGrebeProportions

**A diving duck draws the shorter neck and larger head beside a grebe, as in photographs**

Covers: [product.neckModelPerFamily](../../../local/product/diagram.md#neckmodelperfamily), [product.craniumFromMass](../../../local/product/diagram.md#craniumfrommass)

**Method:** the segment models run on each bird's own mass, bill, and tail

**Assertions:** Lesser Scaup's neck is a smaller share of its length than Horned Grebe's, and its
head a larger one; Western Grebe's neck is a larger share than Horned Grebe's; Canada Goose's neck
is a larger share than the scaup's; the scaup's neck model names its genus ratio and the grebe's
head model its family factor. Not checked by the test: that the values match photographs, which
they were set by judgment to approach ([todo.photoProportions](../../../../todo.md#photoproportions)).

## proportionsReviewTable

**Neck ratios and head, body, and wingspan factors come from a reviewed table, which the build applies**

Covers: [product.neckModelPerFamily](../../../local/product/diagram.md#neckmodelperfamily), [product.craniumFromMass](../../../local/product/diagram.md#craniumfrommass)

**Method:** the parser on fixtures and on `data/proportions-review.tsv`, then the built dataset

**Assertions:** a genus row wins over a family row and a family row over the code's family table;
an empty value inherits; the provenance names the row and says "proposed, not yet reviewed" or
"reviewed by" with the reviewer and date; an unknown status, a reviewed row without reviewer or date,
a proposed row without a reason, a value out of range, a duplicate taxon, or a short row is refused;
the table writes back exactly as read; every taxon in the committed table is a genus or family
BirdView draws; the built Lesser Scaup's neck names the *Aythya* row and the Horned Grebe's head
the Podicipedidae factor; a body factor scales the modeled trunk and is named; a wingspan factor
scales Clark's Grebe's modeled wingspan, recording the factor, and leaves Western Grebe's measured one
of the same genus alone.

## wingspanErrorStated

**Every modeled wingspan states a measured out-of-sample error**

Covers: [product.lengthAndWingspanSources](../../../local/product/measurements.md#lengthandwingspansources)

**Method:** cross-validation on a synthetic small group, then every fit of the model built from the Fox workbook

**Assertions:** seven species give a finite cross-validated error; every order group's fit states an
error; and the woodpeckers' and the loons' and grebes' modeled wingspans name a leave-one-out error
and never read "NaN".

## seaDistanceTable

**Records on water are banded by distance from shore, land cells dropped**

Covers: [tech.seaDistanceTable](../../../local/tech/delivery.md#seadistancetable)

**Method:** the geometry on a synthetic island with a lake, and the generator's measure on synthetic rows in the North Atlantic box

**Assertions:** a point inside land is on land and one in a lake or the sea is not; the distance to
the nearest shore is the local distance to the nearest edge, and reads the cap past it; a polygon
wholly east of the west limit is left out. The generator bands eBird rows by their cell center's
distance (under 10 km, 10 to 30, 30 or more), drops a cell whose center is on land, leaves out
iNaturalist rows and cells outside every sea box, counts unknown names apart, writes the table with
the download's DOI, and refuses a download without the offshore-cells columns. Not run by the
tests: the committed table, which reproduced the 2026-11-08 research exactly (869 species, 98,295
records unmatched).

## seaRefinementsDerived

**Every bird with Marine is refined as open ocean, sea near shore, or both, by the decision's rule**

Covers: [product.seaDistanceDerived](../../../local/product/dataset.md#seadistancederived)

**Method:** the built dataset

**Assertions:** every bird with Marine carries at least one refinement and no other bird carries
one; each is derived, or BirdView's by review; Black-footed Albatross is open ocean only,
White-tailed Tropicbird both, Common Eider sea near shore only, American Robin neither; some bird is
unmeasured, and every unmeasured bird is both; the albatross's reason gives its evidence; a
refinement the review set stands alone. The groups hold AVONET's habitats once each, Marine as its
two refinements, and the review accepts a refinement as an addition.

## neckRatiosReviewed

**Every family's neck ratio has been looked at**

Covers: [product.neckModelPerFamily](../../../local/product/diagram.md#neckmodelperfamily)

**Method:** test over the compiled
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

## regionJoinReportsUnmatched

**The region join normalizes names, and reports what it could not match**

Covers: [product.joinsReportDrops](../../../local/product/dataset.md#joinsreportdrops)

**Assertions:**

1. `taxonomy.Occurrence` maps **every** entry of `AviListRenames` and `CountedAs` onto its
   AviList name, from the lowercased form the occurrence data actually supplies.
2. A name in neither table passes through unchanged.
3. A **shared** name ([decisions.sharedCountShownOnBoth](../../../decisions.md#sharedcountshownonboth)) is *not* folded — those are disclosed,
   not reassigned, and folding one would assert a division nobody measured.
4. `birdview-regions` lists every name matching no drawn species with its record count, and
   **refuses to write** above `maxOrphanFraction`.

**Why the bar is 0.1% and not the measured residue.** A bar set at the measured value fails
the next time iNaturalist gains a subspecies, and a check that cries wolf gets deleted. The
residue was 0.0431% when this was written and is **0.0113%** since the rank fold of
[occurrenceReadersAgree](acceptance.md#occurrencereadersagree), so 0.1% is now 8.8× it — room to grow, far below the 0.81% this exists to
catch.

## disagreementsDeclared

**Every species the two sources disagree about is declared**

Covers: [product.sourceNamingMeasured](../../../local/product/dataset.md#sourcenamingmeasured)

**Method:** offline test over the
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

## occurrenceReadersAgree

**Both readers of the occurrence data land on the same species**

Covers: [product.joinsReportDrops](../../../local/product/dataset.md#joinsreportdrops)

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

**Why the property and not the three bugs.** [decisions.regionJoinNormalizes](../../../decisions.md#regionjoinnormalizes) was three
asymmetries found one at a time over two days, each invisible from inside the command that had
it right: `birdview-regions` applied neither `AviListRenames` nor `CountedAs`, and once it did,
still did not reduce a trinomial. Nothing compared the two readers, so every fix left the next
gap standing. Asserting that they agree closes the shape rather than the instances.

**Why it spans both namespaces end to end.** `birdview-status` resolves a genus move when it
*reads* its table and `birdview-regions` when it *writes* one. That is a real difference and
not a defect, so the criterion compares the name each finally joins on — a check pinned to an
intermediate value would report the difference as a bug every run.

**Why a hybrid is kept rather than dropped.** It is a bird somebody saw, so it belongs in the
denominator of [product.shareOfRecords](../../../local/product/presentation.md#shareofrecords)'s figure; it is no species BirdView draws, so it
belongs in nobody's numerator. Kept verbatim it is both, and it appears in the orphan report
under a name a reader recognizes as the deliberate loss it is.

## ebirdCountedInatFillsGaps

**eBird is counted; iNaturalist fills only the gaps eBird declares**

Covers: [product.scopeIsRecurringAvifauna](../../../local/product/dataset.md#scopeisrecurringavifauna), [product.regionAndMonthFilter](../../../local/product/filtering.md#regionandmonthfilter)

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

## whiteTernFromBirdlife

**Blue-billed White-Tern borrows from BirdLife's White Tern**

Covers: [product.lengthAndWingspanSources](../../../local/product/measurements.md#lengthandwingspansources), [product.provenancePerMetric](../../../local/product/measurements.md#provenancepermetric)

**Method:** unit test over ingest output

**Assertions:** *Gygis candida*'s bill and tail are `near-match`, with `sourceTaxon` *Gygis alba*
and `source` naming `AVONET1_BirdLife`, and carry that row's values, 49.7 mm and 103.8 mm, read
from the workbook rather than written into the test. Vega Gull borrows `AVONET1_BirdLife`'s
*Larus smithsonianus* ([decisions.vegaGullBorrows](../../../decisions.md#vegagullborrows)). Every other near-match species still
borrows from `AVONET2_eBird`. A sheet named in `taxonomy.NearMatchSheets` that does not hold the parent
fails the build instead of falling back to another sheet.

## herringGullFromEuropeanRow

**European Herring Gull is read from BirdLife's European-only row**

Covers: [product.lengthAndWingspanSources](../../../local/product/measurements.md#lengthandwingspansources), [product.provenancePerMetric](../../../local/product/measurements.md#provenancepermetric)

**Method:** unit test over ingest output

**Assertions:** *Larus argentatus*'s bill, tail, and mass are `measured`, with `source` naming
`AVONET1_BirdLife` and `sourceTaxon` *Larus argentatus*, and carry that row's values, read from
the workbook. The workbook's eBird and BirdLife rows must differ, or the check cannot tell them
apart. American Herring Gull is unchanged: `AVONET1_BirdLife` *Larus smithsonianus*, `measured`.
A sheet named in `taxonomy.OwnRowSheets` that does not hold the species fails the build instead
of falling back to the eBird sheet.
