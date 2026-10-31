package product

# Dataset and taxonomy

## scopeIsRecurringAvifauna

**Scope is the recurring avifauna of the regions BirdView covers**

Subject: `dataset.scope`
Value: `BirdView's regions, status established or rare`
Goals: [goals.localRelevance](../../../goals/goals.md#localrelevance), [goals.flexibleComparison](../../../goals/goals.md#flexiblecomparison)

The dataset is the species whose observations **recur** in the regions BirdView covers: a
species is in scope when, in at least one of them, its records fall in repeated episodes across
multiple years.

**The records are eBird's**, and iNaturalist's only in a region and month where eBird is
confirmed to withhold the species ([decisions.countEbirdFirst](../../decisions.md#countebirdfirst)). Otherwise a species
iNaturalist alone photographed in repeated years — feral peafowl, escaped parakeets — would
recur on evidence eBird's reviewers never saw.

**Those regions are BirdView's own list, not the ABA Area.**
[`region`](../../../../internal/region/region.go) holds the 50 US states and the District of Columbia, 13
Canadian provinces and territories, 5 US Minor Outlying Islands, and 5 stretches of open sea.
The ABA Area additionally includes **St-Pierre et Miquelon**, for which BirdView has no region;
BirdView additionally covers **Johnston Atoll, Palmyra Atoll, Wake Island, and Navassa Island**,
which lie well outside it. The product does not name that boundary: the page says *(all
regions)* ([decisions.scopeFromRecurrence](../../decisions.md#scopefromrecurrence)).

Status is derived by [`birdview-status`](../../../../cmd/birdview-status) and committed to
[`data/status.tsv`](../../../../data/status.tsv); `established` and `rare` are in scope, `vagrant`,
`historical`, and `absent` are not. A species the table classifies but leaves out still has a
page saying so ([undrawnSpeciesAnswer](presentation.md#undrawnspeciesanswer)). See [decisions.scopeFromRecurrence](../../decisions.md#scopefromrecurrence).

Rationale: a bird a reader will actually meet is one that keeps turning up, and **recurrence
measures that where a record count cannot**. Record counts measure observer effort: they
under-report pelagic species and remote-island endemics and over-report vagrants that draw a
crowd, and no single threshold corrects both. Akikiki has 376 records across 45 years and is a
resident; the 2023 American Flamingo in Pennsylvania has 931 records across 13 days and is not.

This replaces the ABA rarity codes, which supplied scope until 2026-10-04 and could not be
redistributed. Their replacement is BirdView's own status vocabulary, published
([statusIsPublished](#statusispublished)).

**Two limits are declared rather than measured**, because occurrence data cannot express them:
species with no open record, and species whose presence is maintained by people. Nothing in
either dataset separates a feral population from a steady trickle of escapees. See
[declaredExceptions](#declaredexceptions).

## statusIsPublished

**Status is published, not just used**

Subject: `dataset.status`
Value: `established | rare | vagrant | absent, plus withheld and exotic flags`
Goals: [goals.localRelevance](../../../goals/goals.md#localrelevance), [goals.honestData](../../../goals/goals.md#honestdata), [goals.giveBack](../../../goals/goals.md#giveback)

Each species carries the status [scopeIsRecurringAvifauna](#scopeisrecurringavifauna) derives, and it is shown and filterable
alongside Habitat and Diet ([decisions.habitatDietSpeciesLevel](../../decisions.md#habitatdietspecieslevel)) rather than consumed silently at
ingest. The flags `withheld` — eBird suppresses the species, so the evidence is iNaturalist's —
and `exotic` are orthogonal to the status, because a bird can be withheld *and* established.
`exotic` is a declared list, not a measurement, and what it does is [escapedBirdsListed](measurements.md#escapedbirdslisted).

Rationale: unlike the ABA codes, BirdView's status is its own to republish. Showing it lets a
reader exclude vagrants, or see only them, and makes the judgment auditable: a status shown
next to its evidence invites correction where a hidden ingest filter does not.

## declaredExceptions

**Where the data cannot speak, the exception is declared and named**

Subject: `dataset.status.exceptions`
Value: `declared table, each entry with a written reason`
Goals: [goals.honestData](../../../goals/goals.md#honestdata)

Species whose status cannot be derived carry a declared one instead, from a committed table in
which every entry names its reason. Two classes qualify, and no others:

- **No open record exists.** Laysan Duck, Laysan Finch, Bermuda Petrel, and Desertas Petrel have
  zero records in both eBird and iNaturalist; eBird withholds some sensitive species outright
  and iNaturalist obscures dates for others, so Akohekohe has 7 records and no parseable date.
- **Presence is maintained by people.** Indian Peafowl in New York recurs on every measure — 47
  episodes across 33 years — because ornamental birds are continuously replenished; recurrence
  is blind to origin. Escaped cage, aviary, and farm birds with no
  self-sustaining population anywhere BirdView covers are declared `exotic` and left off the
  home page ([escapedBirdsListed](measurements.md#escapedbirdslisted)); their status is still the derived one.

Rationale: a declared, visible exception can be argued with. The alternative is what
[decisions.scopeFromRecurrence](../../decisions.md#scopefromrecurrence) rejected for the whole model: tuning a coefficient until the
output looks right, which hides the judgment inside a number.

## traitsFromAvonet

**Morphometric traits come from AVONET**

Subject: `dataset.traits.source`
Value: `AVONET, preferring AVONET2_eBird`
Goals: [goals.sizeIntuition](../../../goals/goals.md#sizeintuition), [goals.shapeNotOnlySize](../../../goals/goals.md#shapenotonlysize)

Bill, wing, tail, and mass measurements come from [AVONET](../../imports.md#avonet) — its `AVONET2_eBird`
sheet first, then `AVONET1_BirdLife` or `AVONET3_BirdTree` where the eBird sheet lacks a row for
the species ([decisions.abaSpeciesWithoutAvonet](../../decisions.md#abaspecieswithoutavonet)), or where another sheet holds a narrower concept of the
species ([decisions.herringGullEuropeanRow](../../decisions.md#herringgulleuropeanrow)) or of its near-match donor ([decisions.whiteTernBorrows](../../decisions.md#whiteternborrows)).
See [lengthAndWingspanSources](measurements.md#lengthandwingspansources).

## traitsPerSpecies

**Traits carried per species**

Goals: [goals.sizeIntuition](../../../goals/goals.md#sizeintuition), [goals.shapeNotOnlySize](../../../goals/goals.md#shapenotonlysize)

Each species record carries scientific name, family, and these AVONET traits: `Beak.Length_Culmen`,
`Tail.Length`, `Wing.Length`, `Secondary1`, `Mass`.

Rationale: this is the set the diagram and its derivations consume. AVONET's other columns —
`Beak.Width`, `Beak.Depth`, `Hand-Wing.Index` — are deliberately not ingested; carrying unused
data invites someone to invent a use for it. A column that stops being used leaves, as
`Tarsus.Length` did when the flight premise removed the legs ([decisions.inFlightNoLegs](../../decisions.md#inflightnolegs)).

## namesFromAvilist

**Names come from AviList**

Subject: `dataset.names.source`
Value: `AviList v2025b`
Goals: [goals.goodDataCitizen](../../../goals/goals.md#gooddatacitizen), [goals.flexibleComparison](../../../goals/goals.md#flexiblecomparison)

English common names in **both locales** ([englishNameChoice](diagram.md#englishnamechoice)), the **eBird species code**, family
scientific and English names, the order, and the family **sequence** all come from
[AviList](../../imports.md#avilist). Scientific names come from the scope list, and are the join key.

Rationale: **licensing.** The ABA checklist supplied all of this but is all rights reserved, so
it blocked public deployment ([decisions.avilistReplacesAba](../../decisions.md#avilistreplacesaba)). AviList is CC BY 4.0 and carries
the same four things.

**AviList recognizes 7 of BirdView's species under a different name and 2 not at all.** The seven
are one-to-one renamings and resolve through a synonym table. The two — Gray-headed Swamphen and
Western Warbling-Vireo, both ABA splits AviList lumps — are **dropped from the dataset**: if
AviList supplies the names it supplies the species concepts, and keeping a bird AviList does not
recognize would mean keeping a name the project may not redistribute.

Both tables are recorded in the code with this requirement cited beside them, and the ingest
**asserts the drop list** rather than discovering it: a species losing its name source quietly
would change its URL under a reader ([englishNameChoice](diagram.md#englishnamechoice)).

## bandingCodes

**Each bird carries its four-letter banding code, where the Bird Banding Laboratory gives one**

Subject: `dataset.names.bandingCode`
Value: `USGS Bird Banding Laboratory species list, 2026-10-27`
Goals: [goals.flexibleComparison](../../../goals/goals.md#flexiblecomparison), [goals.goodDataCitizen](../../../goals/goals.md#gooddatacitizen)

Each species carries the four-letter alpha code the USGS Bird Banding Laboratory assigns it
(`AMRO` for American Robin), from the Laboratory's species list ([bbl](../../imports.md#bbl)),
joined by scientific name: where two rows share a scientific name, the row named as the species;
with no scientific match, the one row with the species' North American English name; and a code
two species would share goes to neither. A species the Laboratory does not list, as for most
introduced birds and gamebirds it does not band with federal bands, has none. The code names the
species whatever English names are shown. Every tile, the species page, and the find box's list
show it with the eBird code in parentheses: beside the name on the species page and in the list
("American Robin AMRO (amerob)"), and after the scientific name, under the drawing, on a tile
([tileContents](presentation.md#tilecontents)); the find box finds a bird by either code. Addresses stay the eBird code, which every bird has,
including any added from outside the Laboratory's area (Sameer, 2026-10-27).

Rationale: birders know birds by these codes, and the owner asked for them (2026-10-27). The ABA's
and IBP's lists were not usable ([decisions.avilistReplacesAba](../../decisions.md#avilistreplacesaba)); the Laboratory's is a U.S.
Government work in the public domain ([decisions.bandingCodesFromBbl](../../decisions.md#bandingcodesfrombbl)).

## generatedBadgeColors

**Family badge colors are generated**

Subject: `badge.color.derivation`
Value: `generated`
Goals: [goals.flexibleComparison](../../../goals/goals.md#flexiblecomparison), [goals.openToEveryone](../../../goals/goals.md#opentoeveryone)

Colors are generated as **two palettes**: one per taxonomic order, and one per family *within* its
order. Both satisfy [tech.badgeContrast](../tech/accessibility.md#badgecontrast), and families are visibly different from those listed next to them.

Two families in different orders may share a color. Their order badge distinguishes them.

Rationale: BirdView's scope spans 99 families. A hand-curated palette does not scale to that and
breaks on every taxonomic revision.

A legible badge allows roughly 43 mutually distinguishable colors with black text, or 69 with
white text too — fewer than 99. Scoping family colors to their order reduces the requirement to
the largest order, Passeriformes at 47 families, which fits.

Distinctness is measured perceptually, between the entries a reader compares. Evenly spaced hues
at 99 families are under four degrees apart, and in selector order every neighbor was the same shade.

**Known residual:** within Passeriformes, 12 of 1,081 family pairs remain below ΔE 15 (measured
2026-10-20), because the dark half of the gamut has less perceptual room. Those pairs are never
adjacent in the selector, and every badge carries its family name.

## joinOnBinomial

**Sources join on normalized scientific binomial**

Subject: `dataset.join.key`
Value: `scientific binomial, normalized`
Goals: [goals.honestData](../../../goals/goals.md#honestdata)

Species are joined across sources on scientific binomial, after stripping `(...)` parentheticals
and collapsing internal whitespace, with a committed mapping table covering the taxonomic
divergence between AOS/NACC (which ABA follows) and the 2022-era eBird/Clements taxonomy AVONET
carries.

Rationale: binomial alone joins **97.38%** of the scopeIsRecurringAvifauna scope — 819 of 841 — leaving 22
species needing a mapping entry. English name is the worse key: hyphenation and Eastern/Western
prefixes churn on every taxonomy update.

A rename may be mapped freely. A **split** cannot be mapped without borrowing the parent species'
measurements, which makes the value no longer a measurement *of that species* and engages
lengthAndWingspanSources and tileMarksModeled; see [decisions.abaSpeciesWithoutAvonet](../../decisions.md#abaspecieswithoutavonet).

## unmatchedSpeciesReported

**Unmatched species are reported, not silently dropped**

Goals: [goals.honestData](../../../goals/goals.md#honestdata)

Ingest reports every species that fails to join, by name and count, and fails the build if the
unmatched fraction exceeds a stated threshold.

Rationale: ABA, AOS, and eBird/Clements revise on independent annual cycles. Without a loud
failure, a taxonomy update silently shrinks the dataset and nobody notices.

## joinsReportDrops

**Every join reports what it dropped**

Subject: `build.join_loss`
Value: `named and counted on every run; a declared bar above which the run refuses`
Goals: [goals.honestData](../../../goals/goals.md#honestdata)

Every join between two datasets reports the records it could not match: how many, under what
names, and whether the loss was intended. Where a loss can be bounded, the bar is declared and
the run **refuses to write** rather than exceed it.

Rationale: [goals.honestData](../../../goals/goals.md#honestdata) at the point where data disappears. A join is where a record
can cease to exist without anyone deciding it should, and the loss is invisible on both sides.
`birdview-regions` once discarded **37 names and 10,015,688 records** in silence, 9,954,362 of
them belonging to ten species BirdView draws ([decisions.regionJoinNormalizes](../../decisions.md#regionjoinnormalizes)).

**A discarded record is worse than absent, because it still counts.** It reaches the
denominator of [product.shareOfRecords](presentation.md#shareofrecords)'s share and no species' numerator: Hairy Woodpecker
showed 0.7% of its own record.

**The bar is a fraction of records, not a count of names.** A new vagrant adds a name and six
records; a genus move adds one name and seven million.

**Normalizing is part of the join.** Every reader of the occurrence data normalizes through
**one function**, and a criterion asserts the readers agree on the species rather than
asserting either in isolation ([acceptance.occurrenceReadersAgree](../../acceptance.md#occurrencereadersagree)).

**Rank is reduced before any name is looked up**, because the naming tables are keyed on
binomials that `dryobates villosus harrisi` matches none of. The hybrid test runs *before* that
truncation: cut to two tokens, `larus smithsonianus × hyperboreus` joins cleanly and credits a
real bird with records of something that is not it.

## sourceNamingMeasured

**How each source names a species is measured, not assumed**

Subject: `build.name_correspondence`
Value: `committed per-source record counts; a declared ratio band`
Goals: [goals.honestData](../../../goals/goals.md#honestdata)

A committed table records how many records each occurrence source holds for every species
BirdView draws, under the name BirdView serves. A species far outside the usual ratio between
the two sources is reported, and every exception is declared with a reason.

Rationale: a name that **misses** leaves a species at zero, which somebody eventually notices. A
name that **matches the wrong concept** joins cleanly and stays wrong: eBird's *Setophaga
petechia* is AviList's Mangrove Yellow Warbler by string and the undivided Yellow Warbler by
meaning, and **no name-matching check can see it** —
[`data/gbif-taxa.tsv`](../../../../data/gbif-taxa.tsv) files that pair as `exact`, correctly.

Record counts separate them. The eBird-to-iNaturalist ratio is stable — **median 48×, 5th
percentile 10.7×, 95th 148×** — so every kind of disagreement shows as an outlier:

| shape | signature |
| --- | --- |
| a rename or lump | eBird ≈ 0; its records are under another name |
| a shared name | ratio enormous — *Setophaga petechia* is **117,834×** |
| a withheld species | eBird ≈ 0 under *every* name, iNaturalist normal |
| a domestic form | ratio large — Rock Pigeon 11,809×, iNaturalist filing under a trinomial |

**Only 7 of 1,019 species fall outside 1×–1000×, and every one is a real finding**: the band
caught two unrecorded withheld species, Great Gray Owl and Black Rail, and an error in
[`imports.md`](../../imports.md). Each declared exception names which of the four shapes it is, so
the table reads as a statement about taxonomy rather than a list of numbers to ignore.
