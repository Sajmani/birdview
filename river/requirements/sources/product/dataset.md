package product

# Dataset and taxonomy

## scopeIsRecurringAvifauna

**Scope is the recurring avifauna of the regions BirdView covers**

Subject: `dataset.scope`
Value: `BirdView's regions, status established or rare`
Goals: goals.sizeIntuition, goals.comparisonAcrossFamilies

The dataset is the species whose observations **recur** in the regions BirdView covers: a
species is in scope when, in at least one of them, its records fall in repeated episodes across
multiple years.

**The records are eBird's**, and iNaturalist's only in a region and month where eBird is
confirmed to withhold the species ([decisions.countEbirdFirst](../../decisions.md#countebirdfirst)). Recurrence used to be measured
on both sources' dates together, so a species iNaturalist alone photographed in repeated years —
feral peafowl, escaped parakeets — recurred on evidence eBird's reviewers never saw.

**Those regions are BirdView's own list, not the ABA Area**, and the two are not the same.
[`region`](../../../../internal/region/region.go) holds the 50 US states and the District of Columbia, 13
Canadian provinces and territories, 5 US Minor Outlying Islands, and 5 stretches of open sea.
The ABA Area additionally includes **St-Pierre et Miquelon**, for which BirdView has no region;
BirdView additionally covers **Johnston Atoll, Palmyra Atoll, Wake Island and Navassa Island**,
which lie well outside it.

Calling the union "the ABA Area" was therefore wrong in both directions, and wrong in a second
way: it attributes BirdView's geography to the authority [decisions.scopeFromRecurrence](../../decisions.md#scopefromrecurrence) removed
it from. Nothing in the product names that boundary now — the page says *(all regions)*,
which describes itself. Status is derived by
[`birdview-status`](../../../../cmd/birdview-status) and committed to
[`data/status.tsv`](../../../../data/status.tsv); `established` and `rare` are in scope, `vagrant`,
`historical` and `absent` are not. A species the table classifies but leaves out still has a
page saying so ([undrawnSpeciesAnswer](presentation.md#undrawnspeciesanswer)). See [decisions.scopeFromRecurrence](../../decisions.md#scopefromrecurrence).

Rationale: a bird a reader will actually meet is one that keeps turning up, and **recurrence
measures that where a record count cannot**. Occurrence counts measure observer effort, so they
under-report pelagic species and remote-island endemics — nobody is there — and over-report
vagrants, because a twitchable rarity draws a crowd. Those biases pull in opposite directions
and no single threshold corrects both. Recurrence is insensitive to both: Akikiki has 376
records across 45 years and is a resident; the 2023 American Flamingo in Pennsylvania has 931
records across 13 days and is not.

**This replaces the ABA rarity codes**, which supplied scope until 2026-10-04 and could not be
redistributed. The codes were never a scope rule in their own right — they are a status
vocabulary, of which scopeIsRecurringAvifauna used three values as a cutoff — so what replaces them is BirdView's
own status vocabulary, published rather than hidden ([statusIsPublished](#statusispublished)).

**Two limits are declared rather than measured**, because occurrence data cannot express them.
Species with no open record at all are listed by hand, and so are species whose presence is
maintained by people rather than self-sustaining; see [declaredExceptions](#declaredexceptions). Nothing in either dataset
separates a feral population from a steady trickle of escapees.

## statusIsPublished

**Status is published, not just used**

Subject: `dataset.status`
Value: `established | rare | vagrant | absent, plus withheld and exotic flags`
Goals: goals.sizeIntuition, goals.comparisonAcrossFamilies

Each species carries the status [scopeIsRecurringAvifauna](#scopeisrecurringavifauna) derives, and it is shown and filterable
alongside Habitat and Diet ([decisions.habitatDietSpeciesLevel](../../decisions.md#habitatdietspecieslevel)) rather than consumed silently at
ingest. The flags `withheld` — eBird suppresses the species, so the evidence is iNaturalist's —
and `exotic` are orthogonal to the status, because a bird can be withheld *and* established.
`exotic` is a declared list, not a measurement, and what it does is [escapedBirdsListed](measurements.md#escapedbirdslisted).

Rationale: the ABA codes were read at ingest and discarded, which was right while they were
someone else's data under a license that forbade republishing them. BirdView's own status is
neither, and hiding it would waste the most useful thing the derivation produces: the reader
who wants to exclude vagrants, or to see only them, can say so. It also makes the judgement
auditable — a status shown next to the evidence that produced it invites correction in a way an
invisible ingest filter does not.

This is the first BirdView output that is a dataset in its own right rather than an input to a
drawing.

## declaredExceptions

**Where the data cannot speak, the exception is declared and named**

Subject: `dataset.status.exceptions`
Value: `declared table, each entry with a written reason`
Goals: goals.comparisonAcrossFamilies, goals.goodDataCitizen

Species whose status cannot be derived carry a declared one instead, from a committed table in
which every entry names its reason. Two classes qualify, and no others:

- **No open record exists.** Laysan Duck, Laysan Finch, Bermuda Petrel and Desertas Petrel have
  zero records in both eBird and iNaturalist; eBird withholds some sensitive species outright
  and iNaturalist obscures dates for others, so Akohekohe has 7 records and no parseable date.
- **Presence is maintained by people.** Indian Peafowl in New York recurs on every measure — 47
  episodes across 33 years — because ornamental birds are continuously replenished. Recurrence
  is robust to observer bias and blind to origin. Escaped cage, aviary and farm birds with no
  self-sustaining population anywhere BirdView covers are declared `exotic` and left off the
  index ([escapedBirdsListed](measurements.md#escapedbirdslisted)); their status is still the derived one.

Rationale: an exception that is declared and visible can be argued with. The failure this
guards against is the one [decisions.scopeFromRecurrence](../../decisions.md#scopefromrecurrence) rejected for the model as a whole —
quietly tuning a coefficient until the output looks right, which hides the judgement inside a
number. A short table of named exceptions is a larger admission and a smaller lie.

## traitsFromAvonet

**Morphometric traits come from AVONET**

Subject: `dataset.traits.source`
Value: `AVONET, preferring AVONET2_eBird`
Goals: goals.sizeIntuition, goals.shapeNotOnlySize

Bill, wing, tail and mass measurements come from [AVONET](../../requirements.md#avonet) — its `AVONET2_eBird`
sheet first, then `AVONET1_BirdLife` or `AVONET3_BirdTree` where the eBird sheet lacks a row for
the species ([decisions.abaSpeciesWithoutAvonet](../../decisions.md#abaspecieswithoutavonet)), or where another sheet holds a narrower concept of the
species ([decisions.herringGullEuropeanRow](../../decisions.md#herringgulleuropeanrow)) or of its near-match donor ([decisions.whiteTernBorrows](../../decisions.md#whiteternborrows)).
See [lengthAndWingspanSources](measurements.md#lengthandwingspansources).

## traitsPerSpecies

**Traits carried per species**

Goals: goals.shapeNotOnlySize

Each species record carries scientific name, family, and these AVONET traits: `Beak.Length_Culmen`,
`Tail.Length`, `Wing.Length`, `Secondary1`, `Mass`.

Rationale: this is the set the diagram and its derivations actually consume. AVONET's other
columns — `Beak.Width`, `Beak.Depth`, `Hand-Wing.Index` — are deliberately not ingested; carrying
unused data invites someone to invent a use for it.

The set moves both ways, and a column that stops earning its place leaves. `Beak.Depth` was
ingested until the cranium model left bill depth for mass. `Tarsus.Length` made the whole round
trip — excluded as unused, ingested when the diagram gained legs, then dropped again when the
flight premise removed them ([decisions.inFlightNoLegs](../../decisions.md#inflightnolegs)).

## namesFromAvilist

**Names come from AviList**

Subject: `dataset.names.source`
Value: `AviList v2025b`
Goals: goals.comparisonAcrossFamilies

English common names in **both locales** ([englishNameChoice](diagram.md#englishnamechoice)), the **eBird species code**, family
scientific and English names, the order, and the family **sequence** all come from
[AviList](../../requirements.md#avilist). Scientific names come from the scope list, and are the join key.

Rationale: **licensing, and nothing else.** The ABA checklist supplied all of this and supplied it
well, but it is all rights reserved — serving BirdView reproduces whatever names it, so the
checklist blocked public deployment outright ([decisions.avilistReplacesAba](../../decisions.md#avilistreplacesaba)). AviList is CC BY 4.0
and carries the same four things, so it can be served.

That it is *also* the consensus successor to IOC, Clements and BirdLife is a bonus, not the
reason.

**AviList recognizes 7 of BirdView's species under a different name and 2 not at all.** The seven
are one-to-one renamings and resolve through a synonym table. The two — Gray-headed Swamphen and
Western Warbling-Vireo, both ABA splits AviList lumps — are **dropped from the dataset**. That
follows from the decision rather than qualifying it: if AviList supplies the names it supplies the
species concepts, and keeping a bird AviList does not recognize would mean keeping the one name
the project may not redistribute.

Both tables are recorded in the code with this requirement cited beside them, and the ingest
**asserts the drop list** rather than discovering it: a species losing its name source quietly
would change its URL under a reader ([englishNameChoice](diagram.md#englishnamechoice)).

## generatedBadgeColors

**Family badge colors are generated**

Subject: `badge.color.derivation`
Value: `generated`
Goals: goals.comparisonAcrossFamilies

Colors are generated as **two palettes**: one per taxonomic order, and one per family *within* its
order. Both satisfy tech.badgeContrast, and families are visibly different from those listed next to them.

Two families in different orders may share a color. Their order badge distinguishes them.

Rationale: the ABA area spans 91 families. A hand-curated palette does not scale to that and
breaks on every taxonomic revision, so generation makes the palette a function of the taxonomy
rather than a table to maintain.

The two-palette split is what makes the problem solvable at all. A legible badge bounds the usable
gamut at roughly 43 mutually distinguishable colors with black text, or 69 if white text is also
allowed — against 91 families. No assignment fits 91 into 43. Scoping family colors to their order
reduces the requirement to the largest order, Passeriformes at 44 families, which does fit.

Distinctness must be measured perceptually, and between the entries a reader actually compares.
Evenly spaced hues satisfy "distinct colors" on paper and failed completely in practice: at 91
families that is a four-degree step, and the selector listed families in that same order, so every
neighbor was the same shade.

**Known residual:** within Passeriformes, 9 of 946 family pairs remain below ΔE 15. The dark half
of the gamut has less perceptual room than the light half, and 44 families is close to what the
space holds. Those pairs are never adjacent in the selector, and every badge carries its family
name.

## joinOnBinomial

**Sources join on normalized scientific binomial**

Subject: `dataset.join.key`
Value: `scientific binomial, normalized`
Goals: goals.sizeIntuition

Species are joined across sources on scientific binomial, after stripping `(...)` parentheticals
and collapsing internal whitespace, with a committed mapping table covering the taxonomic
divergence between AOS/NACC (which ABA follows) and the 2022-era eBird/Clements taxonomy AVONET
carries.

Rationale: the measured join rate on binomial alone, over the scopeIsRecurringAvifauna scope, is **97.38%** — 819 of
841 — leaving 22 species needing a mapping entry. English name is the worse key: hyphenation and
Eastern/Western prefixes churn on every taxonomy update.

The 22 are not all renames, and that distinction is the subject of [decisions.abaSpeciesWithoutAvonet](../../decisions.md#abaspecieswithoutavonet).
A rename may be mapped freely. A **split** cannot be mapped without borrowing the parent species'
measurements, which makes the value no longer a measurement *of that species* and therefore
engages lengthAndWingspanSources and tileMarksModeled.

## unmatchedSpeciesReported

**Unmatched species are reported, not silently dropped**

Goals: goals.sizeIntuition

Ingest reports every species that fails to join, by name and count, and fails the build if the
unmatched fraction exceeds a stated threshold.

Rationale: ABA, AOS and eBird/Clements revise on independent annual cycles. Without a loud
failure, a taxonomy update silently shrinks the dataset and nobody notices.

## joinsReportDrops

**Every join reports what it dropped**

Subject: `build.join_loss`
Value: `named and counted on every run; a declared bar above which the run refuses`
Goals: goals.goodDataCitizen, goals.nothingDisappearsQuietly

Every join between two datasets reports the records it could not match: how many, under what
names, and whether the loss was intended. Where a loss can be bounded, the bar is declared and
the run **refuses to write** rather than exceed it.

Rationale: [goals.nothingDisappearsQuietly](../../../goals/goals.md#nothingdisappearsquietly) at the point where data actually disappears. A join is the
one place in this pipeline where a record can cease to exist without anyone deciding it should,
and the failure is invisible on both sides — the input still has the record, the output simply
never mentions it.

**The ingest's joins already did this and the region pass did not**, which is exactly where the
loss was. `birdview-ingest` prints what the AviList join dropped and asserts it against a
declared list, prints what AVONET could not match and fails above a threshold, and requires
iNaturalist coverage to be total. `birdview-regions` matched raw facet names against AviList
names and discarded the rest in silence: **37 names, 10,015,688 records, of which 9,954,362
belonged to ten species BirdView draws** ([decisions.regionJoinNormalizes](../../decisions.md#regionjoinnormalizes)).

**A discarded record is worse than absent, because it still counts.** Every one of those names
reached the denominator of [product.shareOfRecords](../product/presentation.md#shareofrecords)'s share and no species' numerator — so
Hairy Woodpecker's 7,310,342 eBird records were counted as *some other bird* on every tile
including its own, and the bird showed 0.7% of its own record.

**The bar is a fraction of records, not a count of names.** A new vagrant adds a name and six
records; a genus move adds one name and seven million.

**Normalizing is part of the join, not a step before it.** Two commands read the occurrence
data, and the loss above was not a missing table entry — every entry already existed — but a
second reader that never consulted them. So every reader normalizes through **one function**,
and a criterion asserts the two agree on the species rather than asserting either in isolation
([acceptance.occurrenceReadersAgree](../../acceptance.md#occurrencereadersagree)). The three gaps [decisions.regionJoinNormalizes](../../decisions.md#regionjoinnormalizes) found surfaced
one at a time over two days precisely because nothing compared the readers, so each fix left
the next one standing: no rename, then no lump, then no reduction of rank.

**Rank is reduced before any name is looked up**, the naming tables being keyed on binomials
that `dryobates villosus harrisi` matches none of. The ordering that is load-bearing the other
way is the hybrid test, which runs *before* the truncation: cut to two tokens, `larus
smithsonianus × hyperboreus` is an ordinary species name that joins cleanly and credits a real
bird with records of something that is not it.

Measured 2026-10-06, the region pass's unmatched residue fell from **0.0431% under 1,754
names to 0.0113% under 727**, moving **387,515 records into 430 species**. Region totals came
back byte-identical, so this reattributes records rather than changing what they are divided
by, and **114 species-region presences were added and none removed** — the only shape the
change can have, since a trinomial either joins a drawn species or stays where it was. The
largest single beneficiaries are the two declared domestic forms, feral Rock Pigeon
(+109,401) and Muscovy Duck (+37,470), and one genuine subspecies recovery, Yellow-rumped
Warbler (+30,274) from `coronata` and `auduboni`.

## sourceNamingMeasured

**How each source names a species is measured, not assumed**

Subject: `build.name_correspondence`
Value: `committed per-source record counts; a declared ratio band`
Goals: goals.goodDataCitizen, goals.dataAsItIs, goals.nothingDisappearsQuietly

A committed table records how many records each occurrence source holds for every species
BirdView draws, under the name BirdView serves. A species far outside the usual ratio between
the two sources is reported, and every exception is declared with a reason.

Rationale: a name can fail to join in two ways and only one is visible. A name that **misses**
leaves a species at zero, which somebody eventually notices — Bullock's Oriole in no region at
all. A name that **matches the wrong concept** joins cleanly and stays wrong: eBird's
*Setophaga petechia* is AviList's Mangrove Yellow Warbler by string and the undivided Yellow
Warbler by meaning, and **no name-matching check can see it** —
[`data/gbif-taxa.tsv`](../../../../data/gbif-taxa.tsv) files that pair as `exact`, correctly.

What separates them is asking each database how many records it holds. The ratio is
unexpectedly stable — **median 48×, 5th percentile 10.7×, 95th 148×** — because iNaturalist is
uniformly shallower than eBird rather than differently distributed. Every kind of disagreement
shows as an outlier, and in opposite directions:

| shape | signature |
| --- | --- |
| a rename or lump | eBird ≈ 0; its records are under another name |
| a shared name | ratio enormous — *Setophaga petechia* is **117,834×** |
| a withheld species | eBird ≈ 0 under *every* name, iNaturalist normal |
| a domestic form | ratio large — Rock Pigeon 11,809×, iNaturalist filing under a trinomial |

**Only 7 of 1,019 species fall outside 1×–1000×, and every one is a real finding.** Measured
2026-10-05, the band immediately caught two species nothing had recorded as withheld —
**Great Gray Owl** (eBird 1, iNaturalist 2,859) and **Black Rail** (16 against 101) — and
exposed an error in [`sources.md`](../../requirements.md), which listed Blue-billed White-Tern among the
withheld when eBird holds 16,008 records of it under *Gygis alba*.

**The declared exceptions are the point, not an escape hatch.** Each names which of the four
shapes it is, so the table reads as a statement about taxonomy rather than a list of numbers
to ignore.
