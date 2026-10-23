package decisions

# Decisions

How each conflict was settled, each judgment made and each gap knowingly accepted, by whom and
why. Oldest first. The format and the rules for adding an entry are in River's README
([Keeping the bundle's files](../README.md#keeping-the-bundles-files)); the checker enforces the
fields. Older entries also carry `Subject:` and `Kind:`, and their bodies are as first written.

```
grep -n -A2 '^## ' decisions.md                                       # every decision: id and title
awk '/^## /{id=$2} /^Affects:.*\.regionAndMonthFilter(,|$)/{print id}' decisions.md  # what touches a requirement
awk '/^## /{id=$2} /^Status: (amended|superseded)/{print id": "$0}' decisions.md     # no longer in force as written
awk '/^## monthPresence$/,/^---$/' decisions.md                       # one entry, nothing else
```

Append-only, except `Status:`: a decision that turns out badly is superseded or amended by a new
entry, and the old entry's `Status:` line alone changes, to point at it.

---

## noOpenLengthAndWingspan

**No open source supplies both body length and wingspan**

Date: 2026-09-20
Decided by: Sameer
Status: accepted
Affects: [product.lengthAndWingspanSources](local/product/measurements.md#lengthandwingspansources), [product.tileMarksModeled](local/product/measurements.md#tilemarksmodeled), [product.scopeIsRecurringAvifauna](local/product/dataset.md#scopeisrecurringavifauna), [tech.fetchRestrictedInputs](local/tech/delivery.md#fetchrestrictedinputs), [tech.recordBeforeConsuming](local/tech/process.md#recordbeforeconsuming)
Subject: `measurement.provenance`
Kind: Over-constraint — no value satisfies every applicable constraint

### The conflict

Three requirements could not hold at once:

1. Lengths and wingspans shown to the user are real published measurements, not model output
   (owner's stated position, after rejecting the earlier design's invented allometry).
2. The dataset covers the regularly occurring ABA avifauna — 841 species ([product.scopeIsRecurringAvifauna](local/product/dataset.md#scopeisrecurringavifauna)).
3. Inputs are freely licensed and machine-readable ([tech.fetchRestrictedInputs](local/tech/delivery.md#fetchrestrictedinputs), [tech.recordBeforeConsuming](local/tech/process.md#recordbeforeconsuming)).

Research against actual sources established that no combination satisfies all three. AVONET
records neither dimension. Of the candidates that do: Fox et al. 2026 is CC BY 4.0 but carries
empirical wingspan for only ~1,424 species globally, and 6 of 20 common North American landbirds
sampled; Ricklefs 2017 has genuine ruler-measured total length but only for passerines and is not
openly licensed; Wikidata is CC0 with 1,110 wingspans and **nine** lengths; Audubon and All About
Birds have both dimensions and prohibit or block reuse. EltonTraits, the Amniote database,
BIRDBASE, Lislevand and Dunning were each checked and carry neither.

### Options considered

| Option | Satisfies | Violates | Cost |
| --- | --- | --- | --- |
| Real data only, accept gaps | 1, 3 | 2 | ~30% wingspan coverage, passerine-only length. Most tiles incomplete or absent. |
| Return to a modeled source | 2, 3 | 1 | Complete coverage; every number synthetic, which is the position already rejected. |
| Per-value provenance | 2, 3, and 1 *conditionally* | — | Full coverage with each value labeled measured or estimated. Requires the labeling to be real. |
| Pursue licenses first | possibly all | — | Indefinite delay, uncertain outcome. |

### Decision

**Per-value provenance.** Real measurements where they exist, modeled estimates elsewhere, each
tagged, with the interface distinguishing them.

The conditional matters: this only satisfies (1) because **[product.tileMarksModeled](local/product/measurements.md#tilemarksmodeled) makes the labeling
user-visible**. Without [product.tileMarksModeled](local/product/measurements.md#tilemarksmodeled) the decision degenerates into presenting modeled numbers as data,
which is the thing that was rejected. [product.tileMarksModeled](local/product/measurements.md#tilemarksmodeled) is therefore not a nice-to-have and must not be
dropped as a simplification.

### Resulting status

- **[product.lengthAndWingspanSources](local/product/measurements.md#lengthandwingspansources)** Active, as amended.
- **[product.tileMarksModeled](local/product/measurements.md#tilemarksmodeled)** Active, and load-bearing for this decision.
- Honest expectation: most v1 lengths will be estimates. Ricklefs covers passerines only, so
  roughly half of the 841 species have no measured length available at any price.

---

## ricklefsProhibited

**Ricklefs 2017 cannot be used; its terms prohibit it** *(resolved)*

Date: 2026-09-21
Decided by: Sameer
Status: accepted
Affects: [product.lengthAndWingspanSources](local/product/measurements.md#lengthandwingspansources)
Subject: `dataset.length.source`
Kind: Cross-source — a mandatory source's terms versus a local requirement
Date raised: 2026-09-20
Resolved: 2026-09-21

### Resolution

**The source is withdrawn. BirdView will not use Ricklefs 2017.** Every total length in v1 is
`modeled`, tagged as such under [product.provenancePerMetric](local/product/measurements.md#provenancepermetric) and marked on the tile under [product.tileMarksModeled](local/product/measurements.md#tilemarksmodeled).

The terms were finally read — the maintainer printed the page from a browser, since the site 403s
automated requests — and four clauses each independently prohibit what BirdView would do. They are
transcribed with quotations in
[sources/wiley/requirements.md](imported/onlinelibrary.wiley.com/wiley/wiley.md). In brief:

| Clause | Effect |
| --- | --- |
| Rules of Use B | No storing material "in a retrieval system", no "systematic distribution" — and expressly "whether or not for commercial or non-profit use or for a fee or free of charge" |
| Rules of Use C | No integrating the material "with other material or otherwise creat[ing] derivative works" — which is exactly the AVONET join |
| Rules of Use G | No "automated means of any kind" to access the site, which forecloses the fetch-at-build mechanism of [tech.fetchRestrictedInputs](local/tech/delivery.md#fetchrestrictedinputs) |
| Rules of Use F | No incorporation "into any institutional or other repository under any circumstances" |

The decisive detail is in B. This project's non-commercial posture ([tech.localNowPublicLater](local/tech/posture.md#localnowpubliclater)) has unlocked other
sources — Macaulay embedding, iNaturalist's CC BY-NC corpus — and it was the natural hope here
too. The clause forecloses it in as many words.

### Consequence

Roughly half the dataset would have had a measured length; now none does. That is a real quality
loss and it is worth naming rather than absorbing quietly. What makes it survivable is [product.provenancePerMetric](local/product/measurements.md#provenancepermetric):
before per-metric provenance existed, an all-modeled length column meant presenting synthetic
numbers as data. Now it means a length column honestly labeled `modeled` throughout, with the
model named in the detail view.

### Still worth doing

**Ask Ricklefs or ESA.** These are platform terms accepted by using Wiley Online Library; they are
not the last word on what the author or the society may permit for the data itself. A direct grant
would come with none of the above restrictions.

**Deferred until BirdView runs**, by decision of 2026-09-21: a demo that shows the app crediting
its sources and labeling every modeled value as modeled makes a far better case than a description
of an intention. See [notes.md](../notes.md#why-the-permission-asks-wait-for-a-demo). This is an upside
path, not a blocker — v1 ships with modeled lengths either way.

**Not relied upon**, and recorded so nobody revives them as shortcuts: that `RS-eco/traitdata`
redistributes the data under GPL-3 (a package license does not establish the data's license), and
that measurements may be uncopyrightable facts (a legal judgment nobody here is qualified to make,
and one that would not touch the contract in any case).

### Note for agents

Rules of Use G also prohibits AI tools from ingesting or processing Wiley material. No agent on
this project should read the Ricklefs article or its data. Reading the Terms of Use themselves,
from a copy a human supplied, to establish compliance is the brief-quotation-for-comment case that
Rules of Use C permits — and should be the last contact this project has with Wiley material.

---

### Original analysis

Ricklefs 2017 is the only source of real total-length measurements for North American birds that
this project found. Crossref reports only Wiley version-of-record terms and no CC license, and the
Wiley terms page returns HTTP 403 to automated retrieval — so **the terms have not been read**.

Per [README.md](../README.md#sources), the gap is left unfilled rather
than guessed at. Specifically *not* relied upon: that the `RS-eco/traitdata` R package
redistributes the data under GPL-3 (a package license does not establish the data's license), and
the argument that measurements are uncopyrightable facts (a legal judgment nobody here is
qualified to make).

**Three exits, none yet taken:**

1. A human reads the Wiley terms and records what they permit.
2. Ask Ricklefs or ESA for permission for non-commercial open-source use.
3. Drop the measured-length half of [product.lengthAndWingspanSources](local/product/measurements.md#lengthandwingspansources), making every length an estimate.

Until one is taken, the measured-length branch of [product.lengthAndWingspanSources](local/product/measurements.md#lengthandwingspansources) is **Deferred (owner: Sameer)**. The
modeled branch is unaffected, so this does not block v1 — it bounds how good v1 can be.

### Checked and ruled out: the CC0 Dryad deposit *(2026-09-21)*

Crossref reports no related deposits for `10.1002/ecy.1783`, but a separate CC0 dataset exists
with Ricklefs as co-author, and it looked like it might dissolve this conflict:

> Jønsson, Lessard & Ricklefs 2015, "The evolution of morphological diversity in continental
> assemblages of Passerine birds" — `doi:10.5061/dryad.qk461`, CC0-1.0, primary article
> `10.1111/evo.12622`. One file, `data_morphospace_FINAL.xls`.

**It does not contain total length.** Columns, read from the file by the owner — Dryad blocks
automated download and its `robots.txt` disallows `/downloads`:

```
Species name · Clade name · wing length · tail length · tarsus length · middle toe length ·
bill length · bill width · bill depth · Neotropics · Afrotropics · Orient · Australasia ·
other region
```

Two reasons this closes the lead rather than merely deferring it. The measurement suite is
**exactly what AVONET already provides**, so the deposit adds nothing even ignoring length. And
its region columns are tropical, with no Nearctic among them, so North American coverage would
need checking even if the field existed.

The negative result sharpens the original problem: the `Length` column in Ricklefs 2017 is what
distinguishes that dataset, and it is absent even from the same author's other public deposits.
Nothing found so far substitutes for it.

### Reframed by product.provenancePerMetric *(2026-09-21)*

This conflict was raised when provenance was binary — a value was either measured or the app was
misrepresenting it. [product.provenancePerMetric](local/product/measurements.md#provenancepermetric) removed that premise. If no measured length is ever obtained, every
length is labeled `modeled`, and the product is honest and complete.

ricklefsProhibited is therefore **not a blocker** and should not be treated as one. It governs how good v1's
length data is, not whether v1 can ship.

---

## noLocationData

**Location filtering has no redistributable data source**

Date: 2026-09-20
Decided by: Sameer
Status: superseded by regionFilterReinstated
Affects: [product.regionAndMonthFilter](local/product/filtering.md#regionandmonthfilter), [product.filtersCompose](local/product/months.md#filterscompose)
Subject: `filter.location`
Kind: Over-constraint

The owner asked for filtering by location. The species-list source chosen for [product.scopeIsRecurringAvifauna](local/product/dataset.md#scopeisrecurringavifauna) carries no
geography, and no freely redistributable per-region occurrence data was found at useful
granularity. eBird's region species-list endpoint would serve, but requires an API key, and
`api.ebird.org/robots.txt` is `Disallow: /` with the API terms page returning 403 — so what it
permits is unestablished.

**Decision:** withdraw [product.regionAndMonthFilter](local/product/filtering.md#regionandmonthfilter) and [product.filtersCompose](local/product/months.md#filterscompose) from v1 rather than defer them, and revisit with a range
source. Withdrawing rather than deferring is deliberate: a `Deferred` entry needs an owner and a
date, and there is no date until a source exists.

**Resulting status:** [product.regionAndMonthFilter](local/product/filtering.md#regionandmonthfilter) Withdrawn · [product.filtersCompose](local/product/months.md#filterscompose) Withdrawn · family is the only filter.

---

## neckSplitUnmeasured

**The neck/body split has no measured basis**

Date: 2026-09-20
Decided by: Sameer
Status: amended by neckLengthAsMetric
Affects: [product.neckModelPerFamily](local/product/diagram.md#neckmodelperfamily)
Subject: `diagram.neck_ratio`
Kind: Over-constraint, and a consistency question against noOpenLengthAndWingspan

With a total length in hand, AVONET supplies culmen and tail directly and head follows from bill
depth — but nothing measures where the neck ends and the body begins. The earlier design used
hand-assigned per-family neck ratios, which is the same class of invented number that noOpenLengthAndWingspan
rejected for lengths.

| Option | Consequence |
| --- | --- |
| Per-family ratio | Herons look like herons; the split is a guess with no source |
| One fixed ratio | No invented numbers; every bird the same shape, largely defeating [goals.shapeNotOnlySize](../goals/goals.md#shapenotonlysize) |
| Find a source | No per-species neck-length data is believed to exist at this coverage |
| Merge neck into body | Drops a segment the owner explicitly asked to keep |

**Decision:** keep the per-family ratio, **declared as a visual design parameter** carrying no
measurement claim and never surfaced as data ([product.neckModelPerFamily](local/product/diagram.md#neckmodelperfamily)).

This is consistent with noOpenLengthAndWingspan rather than in tension with it, on one condition: noOpenLengthAndWingspan governs
*numbers presented to the user as measurements*, and this governs *pixels in a schematic*. The
distinction holds only while the ratio stays out of the dataset and out of the UI's numbers. If a
neck length ever appears as a value, this decision is void and must be reopened.

---

## photographsWithdrawn

**Photographs withdrawn from v1**

Date: 2026-09-20
Decided by: Sameer
Status: accepted
Affects: [product.speciesPhotographs](local/product/diagram.md#speciesphotographs), [product.perchedOrInFlight](local/product/diagram.md#perchedorinflight), [product.photoAttribution](local/product/diagram.md#photoattribution), [product.photoCaching](local/product/diagram.md#photocaching), [product.diagramOnlyMode](local/product/presentation.md#diagramonlymode)
Subject: `presentation.mode`
Kind: Scope

Photographs were researched in full before being scoped out, and the findings are recorded here so
the work is not repeated:

- **Wikimedia Commons** is the recommended primary. Its licensing policy guarantees commercial-use
  compatibility, the Action API returns structured license metadata (`extmetadata`) sufficient to
  build a credit line automatically, no API key is needed, caching is permitted and preferred over
  hot-linking, and per-species `in flight` categories exist. There is **no "perched" equivalent**
  — 30 category hits in total, mostly cats — so in-flight is a positive filter and perched is the
  residual default.
- **iNaturalist** is a viable fallback *because BirdView is non-commercial* ([tech.localNowPublicLater](local/tech/posture.md#localnowpubliclater)), which makes
  its default CC BY-NC corpus usable. Its API supplies a ready-made attribution string per photo.
  It has a **verified absence** of behavior annotation: the controlled-terms endpoint returns
  seven terms, none of them about flight or perching.
- **Macaulay Library** permits embedding for non-commercial use, quoted from its own page: *"The
  appropriate data and credit will automatically display along with the image… embedding/sharing
  media should be used for non-commercial purposes only."* Raw CDN hot-linking is a different act
  and is not covered by that permission. The blocker is not licensing but discovery — there is no
  public API to find a good photo of a given species, and the media-request path prunes species
  lists over 100.
- **Flickr** is legally usable but operationally hostile: caching is forbidden, galleries are
  capped at 30 photos per page, and there is a 24-hour takedown-propagation duty.

**Decision:** ship v1 with diagrams only. [product.speciesPhotographs](local/product/diagram.md#speciesphotographs) through [product.photoCaching](local/product/diagram.md#photocaching) are withdrawn, not abandoned.

**Resulting status:** [product.diagramOnlyMode](local/product/presentation.md#diagramonlymode) reduced to a single mode · [product.speciesPhotographs](local/product/diagram.md#speciesphotographs)–[product.photoCaching](local/product/diagram.md#photocaching) Withdrawn.

---

## abaSpeciesWithoutAvonet

**ABA species with no AVONET row of their own** *(resolved)*

Date: 2026-09-21
Decided by: Sameer
Status: amended by whiteTernBorrows, vegaGullBorrows
Affects: [product.joinOnBinomial](local/product/dataset.md#joinonbinomial), [product.provenancePerMetric](local/product/measurements.md#provenancepermetric)
Subject: `dataset.join.key`
Kind: Derived — no pair of requirements contradicts; the conflict appears only when [product.joinOnBinomial](local/product/dataset.md#joinonbinomial)'s mapping meets the provenance rule
Date raised: 2026-09-21 · **Resolved:** 2026-09-21

### Resolution

Two decisions, together:

1. **Read each species from whichever AVONET taxonomy carries it as its own taxon**, recording the
   sheet and the filed name in its provenance record. This reduces the problem from 9 species to
   6.
2. **For the remaining 6, borrow the parent taxon's traits and label them `near-match`** under
   [product.provenancePerMetric](local/product/measurements.md#provenancepermetric). All 841 species stay in the dataset, and every borrowed value says whose measurement it
   actually is.

A value read from another sheet, or found under a pre-revision name, is **`measured`** — the
specimens are the right birds and only the filing changed. `near-match` is reserved for a value
measured on a genuinely different taxon.

This mirrors noOpenLengthAndWingspan exactly: mixed authority is acceptable when, and only when, each value is
labeled. [product.provenancePerMetric](local/product/measurements.md#provenancepermetric) now provides the vocabulary both decisions needed, and the detail view of [product.tileOpensDetail](local/product/presentation.md#tileopensdetail) /
[product.metricProvenanceExplained](local/product/measurements.md#metricprovenanceexplained) is where it reaches the user.

**Resulting status:** [product.joinOnBinomial](local/product/dataset.md#joinonbinomial) Active as amended · [product.provenancePerMetric](local/product/measurements.md#provenancepermetric) Active · [product.metricProvenanceExplained](local/product/measurements.md#metricprovenanceexplained) Active · the six near-match
species are `Larus vegae`, `Numenius hudsonicus`, `Sturnella lilianae`, `Vireo swainsoni`,
`Sula brewsteri`, `Anthus japonicus`.

**Carried forward, not closed by this:** the mass caveat below, and the four renames-in-disguise
at the end, still need a human taxonomic check before the mapping table is written.

---

### Original analysis

### What was measured

Joining ABA 8.19 codes 1–3 against `AVONET2_eBird` on normalized binomial: **819 of 841 = 97.38%**.
AVONET carries the 2022-era eBird taxonomy, so four years of revision separate the two sources.
Full classification is in [sources/aba/PROVENANCE.md](imported/aba.org/aba/PROVENANCE.md#taxonomy).

Twelve misses look like renames and can be mapped. **Nine are splits**, where the pre-split parent
is still a separate ABA species in its own right:

| ABA species | Would borrow traits from |
| --- | --- |
| American Herring Gull *Larus smithsonianus* | *Larus argentatus* |
| Vega Gull *Larus vegae* | *Larus argentatus* |
| Hudsonian Whimbrel *Numenius hudsonicus* | *Numenius phaeopus* |
| Cory's Shearwater *Calonectris borealis* | *Calonectris diomedea* |
| Northern Yellow Warbler *Setophaga aestiva* | *Setophaga petechia* |
| Chihuahuan Meadowlark *Sturnella lilianae* | *Sturnella magna* |
| Western Warbling-Vireo *Vireo swainsoni* | *Vireo gilvus* |
| Cocos Booby *Sula brewsteri* | *Sula leucogaster* |
| Siberian Pipit *Anthus japonicus* | *Anthus rubescens* |

### Why this is a conflict rather than a chore

Mapping a split onto its parent silently attributes one species' measurements to another. Two of
the rows above resolve to the *same* parent, so American Herring Gull and Vega Gull would render
as pixel-identical birds in an application whose entire purpose is showing that birds differ in
size and shape.

It also collides with [noOpenLengthAndWingspan](#noopenlengthandwingspan). [product.lengthAndWingspanSources](local/product/measurements.md#lengthandwingspansources) divides every value into *measured* or *estimated*;
a trait borrowed from a pre-split parent is neither. It is a real measurement of a different
taxon, which is a third category the provenance model does not currently have.

### Options

| Option | Effect |
| --- | --- |
| Borrow parent traits, tag as a third provenance class | All 841 species present; honest labeling; product.lengthAndWingspanSources/product.tileMarksModeled grow a category |
| Borrow parent traits silently | Complete dataset, false precision, and identical gull diagrams |
| Exclude the 9 | 832 species, every one genuinely its own measurement; some common birds missing |
| Source the 9 elsewhere | Correct and slow; no identified source |

### Amended 2026-09-21 — most of them are in AVONET after all

The analysis above searched only `AVONET2_eBird`. AVONET ships **three parallel taxonomies**, and
searching all of them changes the numbers materially:

| Resolution | Count | Species |
| --- | --- | --- |
| Present in `AVONET2_eBird` under an older name | 11 | the four *Anarhynchus* (ex-*Charadrius*), *Astur* ×2 (ex-*Accipiter*), *Ardea ibis* (ex-*Bubulcus*), *Botaurus exilis* (ex-*Ixobrychus*), *Gygis candida* (ex-*G. alba*), *Tyto furcata* (ex-*T. alba*), *Ramosomyia violiceps* (ex-***Leucolia violiceps***) |
| Own row in `AVONET1_BirdLife` | 4 | *Larus smithsonianus*, *Calonectris borealis*, *Pterodroma deserta*, *Spilopelia chinensis* |
| Own row in `AVONET3_BirdTree` | 1 | *Setophaga aestiva*, as *Dendroica aestiva* |
| **Absent from every AVONET taxonomy** | **6** | *Larus vegae*, *Numenius hudsonicus*, *Sturnella lilianae*, *Vireo swainsoni*, *Sula brewsteri*, *Anthus japonicus* |

*Ramosomyia violiceps* is resolved: the earlier lookup guessed *Amazilia violiceps*, which is the
BirdLife and BirdTree name. `AVONET2_eBird` calls it *Leucolia violiceps*.

**A new option therefore exists:** read trait rows from whichever AVONET taxonomy sheet carries
the species as its own taxon, recording which sheet each row came from. For the five species
above this yields genuine own-taxon measurements instead of a borrowed parent's.

**And a new caveat.** Where BirdLife splits a pair, the *linear* traits differ but **mass is
identical**:

```
Larus smithsonianus   culmen 69.3  depth 18.3  tail 166.0  wing 429.5  mass 1090.99
Larus argentatus      culmen 74.0  depth 19.6  tail 170.2  wing 430.8  mass 1090.99
```

The same holds for *Calonectris borealis*/*diomedea* (both 534.99 g) and *Pterodroma
deserta*/*feae* (both 311 g). AVONET's mass column is sourced separately from its morphometrics
and was evidently not split. So diagrams would differentiate correctly while **ordering by mass
([product.oneSelectedMetric](local/product/ordering.md#oneselectedmetric)) would still show the pair as identical** — a smaller version of the original problem,
confined to one of the three sort metrics.

### Options considered

| Option | Effect | |
| --- | --- | --- |
| Read across all three AVONET sheets, recording the source sheet per species | Reduces the problem from 9 to 6; mixes taxonomies, which must be recorded per row | **chosen** |
| For the remaining 6, borrow parent traits as `near-match` | All 841 present, honestly labeled | **chosen** |
| For the remaining 6, exclude them | 835 species, every trait its own taxon's | rejected — loses common birds for a labeling problem |
| Borrow silently | Complete dataset, false precision, identical gull diagrams | rejected |

### Still unresolved

The rename/split classification remains provisional. *Astur atricapillus*, *Ardea ibis*, *Gygis
candida* and *Tyto furcata* are listed as renames because the parent is absent from ABA, but each
is really a split whose `AVONET2_eBird` row represents the combined taxon. Mapping them is
therefore the same borrowing problem in a quieter form, and each needs a human taxonomic check
before entering the mapping table.

---

## neckLengthAsMetric

**The neck length is surfaced as a metric, superseding neckSplitUnmeasured's condition**

Date: 2026-09-21
Decided by: Sameer
Status: accepted
Amends: neckSplitUnmeasured
Affects: [product.neckModelPerFamily](local/product/diagram.md#neckmodelperfamily), [product.metricProvenanceExplained](local/product/measurements.md#metricprovenanceexplained), [product.tileOpensDetail](local/product/presentation.md#tileopensdetail), [product.tileContents](local/product/presentation.md#tilecontents), [product.lengthAndWingspanSources](local/product/measurements.md#lengthandwingspansources)
Subject: `diagram.neck_ratio`
Kind: Direct — a later decision reverses a condition an earlier one set

### What changed

[neckSplitUnmeasured](#necksplitunmeasured) kept the per-family neck ratio on one condition, stated plainly: *"If a neck
length ever appears as a value, this decision is void and must be reopened."* The owner has now
decided it should appear — in the detail view, alongside every other segment, with provenance
naming the specific modeling function and the ratio used for that bird's family.

This record reopens it, as neckSplitUnmeasured required.

### Why it resolves the other way now

neckSplitUnmeasured was decided before [product.provenancePerMetric](local/product/measurements.md#provenancepermetric) existed. At the time the choice lay between hiding the number and
presenting it as though it were measured, because there was no way to say "drawing convention" in
the data. Confining it to pixels was the least-bad option available.

Per-metric provenance removes the dilemma. A neck length labeled `kind: modeled` with
`model: "per-family neck ratio, Ardeidae = 0.35"` is not passing itself off as a measurement; it
states exactly what it is and what produced it.

And the concealment was never worth much. **The user sees the neck either way** — it is drawn, at
a length, in every diagram. Withholding the number withheld the explanation, not the claim. Naming
the function is strictly more honest than drawing an unexplained proportion.

This is the third time [product.provenancePerMetric](local/product/measurements.md#provenancepermetric) has dissolved a constraint that existed only because provenance was
binary, after [noOpenLengthAndWingspan](#noopenlengthandwingspan) on modeled lengths and [abaSpeciesWithoutAvonet](#abaspecieswithoutavonet) on split taxa.

### Resulting status

- **[product.neckModelPerFamily](local/product/diagram.md#neckmodelperfamily)** amended: the ratio is a named per-family model, and the neck length is presented with
  its function named. No longer "never surfaced as data".
- **[product.metricProvenanceExplained](local/product/measurements.md#metricprovenanceexplained)** amended: every metric, including all five segments, is shown with its provenance.
- **[product.tileOpensDetail](local/product/presentation.md#tileopensdetail)** amended: the detail view is a larger version of the same tile, plus per-segment
  measurements and provenance.
- **[product.tileContents](local/product/presentation.md#tilecontents)** amended: the tile header carries scientific name as well as common name and badge.
- **[acceptance.neckLengthSerialized](acceptance.md#necklengthserialized)** inverted: it asserted the neck length never leaves the drawing layer; it now asserts
  the neck length *is* serialized, with a `model` naming the family ratio.
- **neckSplitUnmeasured** stands as history. Its reasoning was right for the constraints of its time; only its
  condition is superseded.

### Consequence worth watching

Surfacing the segments makes the composition of total length visible. A reader will see a
**measured** bill inside a **modeled** total and may reasonably ask how both are true. The answer
is that total length is the sum of five segments, two measured and three modeled — now stated in
[product.lengthAndWingspanSources](local/product/measurements.md#lengthandwingspansources) so the detail view can say it rather than leave the reader to infer it.

---

## fittedWingspanModel

**Fit the wingspan model rather than invent one**

Date: 2026-09-22
Decided by: Sameer
Status: accepted
Affects: [product.lengthAndWingspanSources](local/product/measurements.md#lengthandwingspansources), [product.provenancePerMetric](local/product/measurements.md#provenancepermetric)
Subject: `measurement.wingspan.model`
Kind: Over-constraint, reopened by evidence

### What prompted it

The ingest shipped with an invented wingspan model — `2 x chord x 1.15 + 10.5 x mass^0.33` — used
for 533 of 841 species. Measured against the 1,424 empirical wingspans in the Fox dataset, its
**mean absolute error is 18.7%**, median 18.2%, worst 49.6%. For an application whose purpose is
judging relative size, applying that to nearly two thirds of the dataset is not a rounding
concern.

Five-fold cross-validation on the same empirical data, comparing candidate forms:

| Model | mean | median | p90 |
| --- | --- | --- | --- |
| The invented constants | 18.7% | 18.2% | 29.9% |
| Global log-log on wing chord | 9.0% | 7.3% | 19.0% |
| **Order-level log-log** | **6.0%** | **4.6%** | **12.6%** |

### The obstacle, and why it was not one

Fox et al. publish exactly this: an order-level regression, with its fitted values shipped in
`wingspan_predicted_*`. But [requirements.md](imports.md#fox-wingspan) excludes those columns, so
ingesting them would have needed an amendment — and spec2code does not get to reinterpret an
exclusion.

The resolution avoids the question. BirdView fits **its own** regression on the empirical column
it already adopts, which is modeling from data the project may use rather than ingesting someone
else's estimates as data. The exclusion stands unchanged.

### Validation, which is what made the choice safe

Fitted this way and asked for the same 9,194 species Fox predicted, BirdView's model agrees with
theirs to a **mean of 1.32%** and a median under 1%, with 1.4% of species differing by more than
10% — mostly extreme-aspect seabirds. It is, in effect, their model reconstructed.

That comparison is now a standing check ([acceptance.wingspanModelReconstructsFox](acceptance.md#wingspanmodelreconstructsfox)), so a regression in the fitting code shows up as
disagreement with an external published reference rather than as a silently worse dataset.

Note what could **not** be tested: Fox report no predicted value for species that have an
empirical one, so there is no held-out set on which to compare their published predictions
directly against ours. The agreement above is between two models, and the error figures are our
own cross-validation.

### Also decided: modeled values state their error

A provenance string now reads *"log-log regression of wingspan on wing chord, fitted on 360
empirical species in Passeriformes from the Fox et al. 2026 dataset; cross-validated median error
4.7%"*. Knowing a number is an estimate is much less useful than knowing it is an estimate good to
about 5%, and the figure falls out of the fitting already being done.

### Resulting status

- **[product.lengthAndWingspanSources](local/product/measurements.md#lengthandwingspansources)** amended: the modeled-wingspan branch names the fitted regression.
- **[product.provenancePerMetric](local/product/measurements.md#provenancepermetric)** amended: a `model` states the computation *and its measured error* where one exists.
- **[acceptance.wingspanModelReconstructsFox](acceptance.md#wingspanmodelreconstructsfox)** added: agreement with Fox's published predictions stays within 2% mean.
- `morphology.Wingspan`, the invented model, is deleted rather than left for someone to reach for.
- Coverage is unchanged at 301 measured, 7 near-match, 533 modeled; only the modeled values move.

---

## legFromTarsus

**The exposed leg is modeled from the tarsus, reversing [product.segmentComposition](local/product/diagram.md#segmentcomposition)**

Date: 2026-09-22
Decided by: Sameer
Status: superseded by inFlightNoLegs
Affects: [product.segmentComposition](local/product/diagram.md#segmentcomposition)
Subject: `diagram.legs`
Kind: Direct — a later decision reverses a rejection an earlier one recorded

### What changed

[product.segmentComposition](local/product/diagram.md#segmentcomposition) drew the legs as the measured tarsus and explicitly rejected the
alternative: *"scaling up by a per-family factor to approximate the exposed leg would replace the
diagram's one new measurement with another invented table, so it is not done."* That is now done.
`morphology.LegFactors` scales the tarsus by a per-family factor between 1.0 and 2.0, and the
diagram draws the result.

### Why it resolves the other way now

The rejected option was rejected for purity, and the purity was costing accuracy in exactly the
birds the feature exists for. The tarsus is the tarsometatarsus alone; the tibiotarsus above it is
feathered over in passerines but largely bare in waders and cranes. So drawing the tarsus
understates the leg, and **understates it most in the long-legged birds** — a Reddish Egret, a
flamingo and a crane all drew up as unremarkable, which is the one impression the diagram must not
give.

The purity was also smaller than it looked. A modeled leg tagged `kind: modeled` with
`model: "exposed leg modeled as 1.9 x the measured tarsus of 167.6 mm (family-specific factor for
Ardeidae)"` is not passing itself off as a measurement, and **the measured tarsus is still carried
and still shown**, on its own row in the detail view. Nothing was concealed to buy the accuracy; a
second row was added.

This is the same shape of argument as [neckLengthAsMetric](#necklengthasmetric), and the fourth time [product.provenancePerMetric](local/product/measurements.md#provenancepermetric) has dissolved a
constraint that existed only because there was no way to say "this is a model" in the data.

### What it costs, stated plainly

The diagram used to have three parts that were not modeled — bill, tail and tarsus. It now has
two. The factors are **invented**: no source measures exposed leg length, and unlike the wingspan
model of [fittedWingspanModel](#fittedwingspanmodel) there is no empirical column to fit against or validate on. They are a
drawing parameter of the same kind as the neck ratio of [product.neckModelPerFamily](local/product/diagram.md#neckmodelperfamily), and carry the same obligation —
named in the provenance of every value they produce, never presented as measurement.

### Consequence the renderer had to absorb

[acceptance.legsFromTarsusWithdrawn](acceptance.md#legsfromtarsuswithdrawn) recorded an invariant: feet never fall below the tail tip, which is what let the renderer
center a diagram on its bill-to-tail extent. **The modeled leg makes that false** for 11 of the
841 species — an American Flamingo by 30% of its own length, then Whooping Crane, Reddish Egret
and Black-necked Stilt. That is correct anatomy, so the renderer accommodates it: `VerticalExtent`
sizes each section's scale on the drawn extent rather than on total length, and the template
centers on `DrawnHeight`. The invariant is not merely dropped — its replacement is asserted in
both directions, so a bird standing below its own tail is still confined to the wader families.

Drawing order changed with it. The body stick is painted **after** the legs, so the legs pass
behind the body rather than crossing it and the junction mark.

### Resulting status

- **[product.segmentComposition](local/product/diagram.md#segmentcomposition)** amended: legs are the modeled exposed leg, not the measured tarsus; the rejection
  above is superseded; the draw order is stated.
- **[acceptance.legsFromTarsusWithdrawn](acceptance.md#legsfromtarsuswithdrawn)** rewritten: the feet-above-tail assertion is inverted into a drawn-extent invariant,
  and the leg model gains dataset-wide checks.
- **[product.provenancePerMetric](local/product/measurements.md#provenancepermetric)** unchanged but newly load-bearing: `Species.Leg` is modeled, `Species.Tarsus` measured,
  and both are serialized and shown.
- The tile width reference stays the wingspan alone. A pair of legs never spreads wider than the
  wings — the widest case in the dataset reaches 35% of it — so that is asserted over the whole
  dataset rather than guarded by a branch nothing would ever exercise.

**Superseded in full by [inFlightNoLegs](#inflightnolegs).** The diagram is a bird in flight; legs are gone.

---

## inFlightNoLegs

**The diagram depicts a bird in flight, and legs are removed entirely**

Date: 2026-09-23
Decided by: Sameer
Status: amended by foldedNeckAsS
Supersedes: legFromTarsus
Affects: [product.segmentComposition](local/product/diagram.md#segmentcomposition), [product.scaleToggle](local/product/diagram.md#scaletoggle), [product.traitsFromAvonet](local/product/dataset.md#traitsfromavonet), [product.traitsPerSpecies](local/product/dataset.md#traitsperspecies)
Subject: `diagram.posture`, `diagram.legs`
Kind: Premise settled, and a scope cut following from it — supersedes [legFromTarsus](#legfromtarsus)

### The premise, which was never written down

The diagram has always been ambiguous about what it depicts. A vertical body axis reads as a
standing bird; a spread wing crossbar reads as a flying one. Nothing in the spec said which, and
every leg decision so far — the rake angle, the Z-shape proposal, legFromTarsus itself — was really an
argument about that unstated premise.

It is now stated: **the diagram is a top-down view of a bird in flight.** That reading is what the
crossbar already assumed, and it makes [product.crossbarAtVerticalScale](local/product/diagram.md#crossbaratverticalscale) better founded than it was — seen from above, wingspan
and body length genuinely lie in one plane, so drawing them at one scale is not a convention but a
fact.

### Why legs cannot survive that premise

In flight a bird does one of two things with its legs. Most tuck them up into the flank and belly
feathers, where they are **invisible**. A minority — herons, cranes, storks, ibises, flamingos,
stilts, rails, loons, grebes — extend them straight back, where what you see is not the leg but its
**projection past the tail tip**:

```
projection = (femur + full tibiotarsus + tarsometatarsus + toes) − (hip → tail tip)
```

BirdView has exactly one of those terms. `Tarsus.Length` is the tarsometatarsus. The femur is not
measured by any source; the toes are not measured; the hip's position along the body axis is not
measured and was simply assumed to be the body midpoint. Worst of all, `LegFactors` models the leg
*visible on a standing bird*, which deliberately **excludes** the feathered proximal tibiotarsus —
and that is precisely the part that straightens out and becomes visible in flight. The model
excludes the term the new premise needs most.

So the quantity is a small residual between two large modeled numbers, with most of one number
missing.

### The evidence that it does not work

Counting species in the trailing-leg families whose leg clears the tail tip:

| Using | Clears the tail |
| --- | --- |
| The measured tarsus | 1 of 110 |
| The modeled leg (legFromTarsus) | 14 of 110 |

The true figure is far higher — every large heron, egret, crane, stork and ibis does it. Checking
the textbook field marks against the modeled leg: Whooping Crane +208 mm, Sandhill Crane +115,
Wood Stork +70, Black-necked Stilt +63, Great Egret +55 — but **White Ibis −14 mm**, a bird whose
trailing legs are unmistakable in flight. The sign comes out right for the extremes and wrong in
the middle, which is what systematic underprediction looks like. The plausible numbers were luck,
not model.

### What is lost, stated plainly

The tarsus was a **measured** trait spanning 115×, from a Rufous Hummingbird's 2.5 mm to a Whooping
Crane's 289 mm, and legFromTarsus's rationale for drawing it at all still reads well. That range is real
and it is now not shown anywhere. The honest accounting is that it was only ever *visible* because
the diagram tacitly depicted a standing bird — the posture the project has now decided against. A
measurement that can only be displayed by implying a false posture is not a measurement the
diagram can carry.

Considered and rejected: a fixed-length, not-to-scale marker past the tail for the trailing-leg
families. It would capture a genuine field mark, but it puts an unscaled element into a diagram
whose entire premise is that everything in it is to scale. That cost is higher than the gain.

### Resulting status

- **[product.segmentComposition](local/product/diagram.md#segmentcomposition)** amended: five segments and a crossbar. No legs. The flight premise is stated there.
- **[product.scaleToggle](local/product/diagram.md#scaletoggle)** reverted: the scale reference is the largest total length again, not the drawn extent.
- **product.traitsFromAvonet/product.traitsPerSpecies** amended: `Tarsus.Length` is no longer ingested, leaving five AVONET traits.
- **[acceptance.legsFromTarsusWithdrawn](acceptance.md#legsfromtarsuswithdrawn)** withdrawn — the requirement it verified no longer exists.
- **[acceptance.allFiveTraits](acceptance.md#allfivetraits)** amended: five traits, not six.
- Deleted from the code: `morphology.LegFactors`, `DefaultLegFactor`, `LegLength`;
  `dataset.Species.Leg` and `.Tarsus`; `avonet.Traits.Tarsus`; `view.VerticalExtent`,
  `Geometry.DrawnHeight`, `LegCos`, `legRake`, `legSin` and the leg coordinates; the leg strokes,
  legend entry and CSS.
- legFromTarsus's renderer machinery goes with it. That work was correct for its premise and survived
  about a day; the premise, not the code, was the thing that was wrong.

### Left open

The **neck** is still modeled on a standing posture, and under the flight premise it is wrong for
the families that fold it — Ardeidae and Pelecanidae fly with the head drawn back onto the
shoulders, while our Ardeidae neck ratio of 0.85 is the second-highest in the table. Fixing it
collides with [product.segmentsSumToLength](local/product/diagram.md#segmentssumtolength), since a folded-neck heron's drawn extent is genuinely shorter than its
catalogued total length. **Resolved by [foldedNeckAsS](#foldedneckass).**

---

## foldedNeckAsS

**A folded neck is drawn as an S; length as data is untouched**

Date: 2026-09-23
Decided by: Sameer
Status: accepted
Amends: inFlightNoLegs
Affects: [product.segmentComposition](local/product/diagram.md#segmentcomposition), [product.segmentsSumToLength](local/product/diagram.md#segmentssumtolength)
Subject: `diagram.neck.posture`
Kind: Open item from [inFlightNoLegs](#inflightnolegs), resolved

### What changed

For the families that fold the neck in flight — **Ardeidae** and **Pelecanidae** — the neck is
drawn as an **S of two semicircular arcs** rather than a straight stick. Everything else about the
neck is unchanged.

### The property that makes it cheap

The S is **two semicircles of radius r**. Its path length is `2πr` and its vertical extent is `4r`,
so a neck of length L is drawn with `r = L/2π`, the stroke is **exactly L long at the diagram's
scale**, and it occupies `2L/π ≈ 63.7%` of L vertically.

That ratio is a **consequence of the shape, not a parameter**. No new number enters the system —
which is exactly what separates this from the leg factors of legFromTarsus, and why this change needs no
provenance record: like `headAspect`, it produces no metric. Nothing is rescaled, nothing is
hidden, and the ink you see is still the neck at the same pixels-per-millimeter as every other
part of the bird. It is the same neck, curved.

Tightening it further would mean arcs of more than 180°, and *that* would introduce a chosen
number. If 63.7% ever proves too loose to read as a fold, the cost of the next step is a parameter.

### What it costs

For these 14 species the **drawn vertical extent is no longer the total length**. [product.segmentsSumToLength](local/product/diagram.md#segmentssumtolength)'s rationale
— "this is what lets the vertical extent of the drawing *be* the bird's length" — now holds for
827 of 841 species rather than all of them.

That is acceptable, and arguably a correction rather than a concession. A catalogued total length
is a measurement convention taken from a stretched specimen; a heron in flight genuinely is shorter
bill-to-tail than that figure. Drawing it at its catalogued extent is the falsehood, not drawing it
short.

**Length as data is untouched.** Sorting, sectioning, the scale reference and the detail view all
continue to use the full total length, including the full modeled neck. A heron still sorts and
sections exactly where its length puts it, and the detail view still reports the neck at its full
value. Only the silhouette compresses.

### Scope, and the families deliberately left alone

Pelecanidae currently draws the default neck ratio of 0.18, so folding it changes little on screen.
The flag is still right: it records the posture, and if the ratio is ever corrected — pelicans do
have long necks — the folding is already in place.

Left as straight sticks, all correct as drawn: cranes, storks, ibises and spoonbills, flamingos,
waterfowl, loons, grebes, cormorants and anhingas all fly with the neck **extended**. Gulls and
shorebirds retract the head somewhat but do not fold, and their modeled necks are short enough that
a stick is not misleading.

### Resulting status

- **[product.segmentComposition](local/product/diagram.md#segmentcomposition)** amended: the neck is a stick or an S depending on the family's flight posture.
- **[product.segmentsSumToLength](local/product/diagram.md#segmentssumtolength)** amended: the rationale is qualified — drawn extent equals total length except where the
  neck is folded, and the sum over the *metrics* is unaffected.
- **[acceptance.foldedNeckSameStroke](acceptance.md#foldednecksamestroke)** added.
- No change to [product.neckModelPerFamily](local/product/diagram.md#neckmodelperfamily), the neck-ratio model, or to any serialized value.

---

## inaturalistLinkAdopted

**Adopting iNaturalist for a species link, and what it is allowed to be**

Date: 2026-09-23
Decided by: Sameer
Status: amended by fullInaturalistCoverage
Affects: [product.inaturalistLink](local/product/presentation.md#inaturalistlink), [tech.pacedInaturalistMapping](local/tech/process.md#pacedinaturalistmapping)
Subject: `presentation.detail_view.external_link`
Kind: Source adoption, with three sub-decisions

### What prompted it

BirdView shows abstract diagrams. A reader who has just learned that a Merlin is smaller than they
thought will immediately want to see one, and BirdView has no photographs and will never have any.
A link out is the whole answer to that.

### Why not All About Birds, which is the better destination

Cornell's All About Birds has better photography and better species accounts than anything else
for North American birds, and it was the first choice. It was rejected on two grounds, the second
decisive.

**Its terms could not be read.** `birds.cornell.edu/home/terms-of-use/` and
`allaboutbirds.org/robots.txt` both answer an automated request with a Cloudflare HTTP 403. By
this project's own standard — terms read and quoted before a source is adopted — that is not a
position from which "permitted" can be written down. Working around the block was rejected on the
same reasoning that rejected it for the iNaturalist pages, and for Wiley in
[ricklefsProhibited](#ricklefsprohibited).

**Its URLs cannot be constructed.** All About Birds keys on Clements/eBird common names; BirdView
carries ABA/AOS names, and the ABA checklist's own header warns that the two authorities disagree.
The disagreement is worst exactly where BirdView is already weakest — the ten recent splits that
carry near-match provenance. A slug transform would emit links that 404 at an unknown rate, and
shipping unverified URLs is the same failure as presenting a modeled value as measured.

### Why iNaturalist

It keys on **scientific name**, which BirdView holds and which is far more stable across
authorities than any common name. Measured on a 55-species stratified sample containing all ten
near-match splits:

| | |
| --- | --- |
| resolved EXACT on scientific name | **55 / 55** |
| common name *also* agreed exactly | **54 / 55** |

The single disagreement is a hyphen — iNat's "Western Warbling Vireo" against ABA's "Western
Warbling-Vireo". Every one of the ten splits resolved, with BirdView's own common names returned:
`Astur atricapillus` → American Goshawk, `Gygis candida` → Blue-billed White-Tern, `Tyto furcata`
→ American Barn Owl.

Wikimedia Commons was also tested, because its URLs are constructible from the scientific name
with no API call at all. It failed on **three of six** of the hard cases: no `Astur atricapillus`,
no `Gygis candida`, no `Tyto furcata`. It lags exactly where the splits are.

### Sub-decision 1: the common name is a required cross-check, not a bonus

A link is emitted only when the scientific name matches exactly **and** the returned common name
agrees with ABA's, compared with punctuation and case normalized so the Warbling-Vireo hyphen does
not cost a link.

Scientific-name-exact alone would be enough to find the right page almost always. Requiring both
means a taxonomic revision on either side shows up as a **dropped link** rather than a link that
quietly points at the wrong bird — and a wrong bird is worse than no bird on a page whose entire
purpose is being careful about what it claims.

### Sub-decision 2: a species with no confident match gets no link

Not a search URL, not a guess, not a link to a genus. Nothing. The detail page simply does not
show the row.

A search link never 404s, which is precisely what makes it a bad idea: it always looks like it
worked. An absent link is legible — the reader learns that BirdView could not confidently identify
this bird on iNaturalist, which is true and worth knowing.

### Sub-decision 3: the mapping is a committed table, not a build-time fetch

`birdview-ingest` currently makes **no network request at all**, and [tech.selfContainedServer](local/tech/delivery.md#selfcontainedserver) keeps the
server offline at runtime while `tech.md` binds tests to make none either. Resolving 841 names at
the one-request-per-second that [`inatapi.RequestRate`](imported/inaturalist.org/inatapi/inatapi.md#requestrate) asks for would add roughly fifteen minutes to every
build and break it entirely offline.

So resolution is a separate, rarely run tool — `cmd/birdview-inatmap` — writing a table that is
**committed to the repository**, exactly as `internal/taxonomy`'s synonym table already works. The
ingest joins against the table and never touches the network. The cost is that the table goes
stale between runs, which is the correct trade: a stale identifier fails visibly as a dropped or
wrong link, and the refresh is one command.

### What is licensed, and what is not

Nothing in the iNaturalist terms or community guidelines restricts reusing a taxon identifier or
linking to a taxon page; the vendored copies contain no occurrence of *taxon*, *taxonomy*,
*database* or *derivative*. The CC BY-NC clause governs Content a contributor posts — the
photographs a reader will go and look at — not the taxonomy.

**BirdView therefore links and never copies.** No photograph is embedded, hotlinked, cached or
reproduced, so no contributor's license is engaged. [tech.selfContainedServer](local/tech/delivery.md#selfcontainedserver) forbids the embedding
independently, which is a useful accident: two unrelated rules point the same way.

### Resulting status

- **`sources.md`**: `inat-api` (governing) and `inat-terms` (mandatory) adopted, both vendored by
  copy from the maintainer's `birdsync` repository and verified byte-for-byte against the SHA-256
  hashes recorded there. Neither could have been fetched — `inaturalist.org/pages/*` answers an
  agent with a 403 and a JavaScript challenge.
- **[product.inaturalistLink](local/product/presentation.md#inaturalistlink)** added: the detail view links to iNaturalist when, and only when, the match is
  confident.
- **[tech.pacedInaturalistMapping](local/tech/process.md#pacedinaturalistmapping)** added: the mapping tool paces its requests and identifies itself.
- **[acceptance.inaturalistLinkConfidentOnly](acceptance.md#inaturalistlinkconfidentonly)** added.
- The existing **[tech.identifiedClient](local/tech/delivery.md#identifiedclient)** (identify the client) and **[acceptance.outboundRequestsIdentified](acceptance.md#outboundrequestsidentified)** already discharge [`inatapi.UserAgent`](imported/inaturalist.org/inatapi/inatapi.md#useragent),
  which is why that requirement needed no new work.

---

## twoContrastFloors

**Contrast is checked for all text, at two floors rather than one**

Date: 2026-09-23
Decided by: Sameer
Status: amended by oneContrastFloor
Affects: [tech.textContrastFloor](local/tech/accessibility.md#textcontrastfloor), [wcag.ContrastMinimum](imported/w3.org/wcag/wcag.md#contrastminimum), [wcag.ContrastEnhanced](imported/w3.org/wcag/wcag.md#contrastenhanced)
Subject: `text.contrast.ratio`
Kind: Source scope widened, after a defect the existing scope could not see

### What prompted it

The iNaturalist link of [inaturalistLinkAdopted](#inaturalistlinkadopted) shipped with no `color` rule, so it fell back to the
browser default: `#0000EE` on a `#0f172a` background, which is **1.90:1**, and `#551A8B` once
visited, which is **1.62:1**. Effectively invisible. A reader spotted it immediately.

Nothing caught it, and nothing could have. [`wcag`](imports.md) was adopted with its scope set to
*"Family badge colors only"*, on the reasoning that badge colors are **generated** and an unchecked
generator is obviously dangerous. That reasoning was right and incomplete: hand-written CSS turned
out to be just as dangerous, and had no check at all.

### The decision that was not obvious

Extending the badge floor of 7:1 to everything would have been the tidy answer, and it fails
immediately. `--dim`, the secondary text color used for scientific names, provenance detail and
counts, is **6.96:1** on the page background and **5.71:1** on a panel. So would `--accent` on a
panel, at 6.83:1.

Raising the floor to 7:1 therefore means repainting the secondary text of the entire interface —
a much larger change than a broken link warrants, and one nobody asked for.

So **two floors**, and the gap between them stated rather than rounded up:

| | Floor | Why |
| --- | --- | --- |
| Badges and chips | 7:1 against their own text ([tech.badgeContrast](local/tech/accessibility.md#badgecontrast), unchanged) | the colors are generated |
| **Links** | **7:1** against the page background | interactive, and the palette clears it easily at 8.33:1 |
| **All other text** | **4.5:1** against both surfaces | AA; `--dim` clears this comfortably |

Links get the strict floor because nothing makes it hard: `--accent` is already 8.33:1 on the page
background, and the new visited color is 9.67:1. The 7:1 link check is applied against the page
background specifically, having verified that no bare `<a>` sits inside `.help` or `.controls`,
whose surface is `--panel`. If one ever does, the 4.5:1 rule still governs it and the check still
runs.

### The second assertion, which is the one that matters

A contrast check over *declared* colors is blind to the colour nobody declared — which is precisely
how this bug shipped. So [acceptance.noTextBelowContrastFloor](acceptance.md#notextbelowcontrastfloor) also asserts the stylesheet carries top-level `a` and `a:visited`
rules, making "someone forgot" impossible rather than merely unlikely.

Colors are now set by a global rule rather than per component, so a new link inherits a good color
by default instead of needing to be remembered.

### A visited state now exists at all

There was none before. Every link in the interface rendered identically whether or not it had been
followed, because the components that styled links set `color` without a `:visited` companion,
which suppresses the browser's own visited color as a side effect. That was invisible rather than
wrong, until a link appeared with no rule and the default purple arrived at 1.62:1.
`--accent-visited: #c4b5fd` is 9.67:1 on the page background and clearly distinct from the sky-blue
accent.

### Resulting status

- **`sources.md`**: the `wcag` scope widens from badges to badges *plus* link and text colors. Both
  already-adopted criteria are now used, at the two floors above.
- **[tech.textContrastFloor](local/tech/accessibility.md#textcontrastfloor)** added.
- **[acceptance.noTextBelowContrastFloor](acceptance.md#notextbelowcontrastfloor)** added, watched failing three ways, including against the original defect.
- `--accent-visited` added; link colors moved to global `a` / `a:visited` rules; `.tile`, `.btn` and
  `.chip` given explicit `:visited` rules so the global rule cannot recolor them.
- Still excluded, and still claimed nowhere: keyboard access, focus order, timing, reading level.
  **BirdView claims no WCAG conformance level.**

**The two-floor scheme was superseded the next day by [oneContrastFloor](#onecontrastfloor).**

---

## oneContrastFloor

**One contrast floor, superseding twoContrastFloors's two**

Date: 2026-09-23
Decided by: Sameer
Status: accepted
Amends: twoContrastFloors
Affects: [tech.textContrastFloor](local/tech/accessibility.md#textcontrastfloor)
Subject: `text.contrast.ratio`
Kind: Direct — a later decision reverses a tradeoff an earlier one accepted

### What changed

[twoContrastFloors](#twocontrastfloors) set two floors: 7:1 for links and badges, 4.5:1 for everything else. There is now
one, **7:1**, and `--dim` and `--accent` moved to clear it.

| | was | now | on page bg | on panel |
| --- | --- | --- | --- | --- |
| `--dim` | `#94a3b8` | `#adbbcc` | 6.96 → **9.14** | 5.71 → **7.49** |
| `--accent` | `#38bdf8` | `#5ccbfa` | 8.33 → **9.67** | 6.83 → **7.93** |

### Why it resolves the other way now

twoContrastFloors's reasoning was explicit and, at the time, sound: raising the floor meant repainting the
interface's secondary text, which looked disproportionate to the one broken link that prompted the
check. So it recorded the gap instead of closing it.

What that reasoning was missing is that the gap was not hypothetical. **The owner independently
reported the dim text as hard to read** — scientific names, provenance detail, section counts, all
at 5.71:1 on a panel. A measured shortfall that a user has actually noticed is no longer a tradeoff
worth defending; it is the same defect as the link, one tier less severe.

Cheap, too, in the end. Two variables moved and nothing else did. `--accent` as the `.btn.on` fill
still clears its own dark text at 8.25:1, and `--accent-visited` was already at 7.92:1 on a panel.
The repaint twoContrastFloors was avoiding turned out to be two hex values.

### What this says about recording a known gap

twoContrastFloors did the right thing by writing the shortfall down with its numbers rather than rounding up
to a claim BirdView could not support. That is what made this decision a one-line lookup rather
than a rediscovery — the exact figures were already on the page.

But a recorded gap is a deferral, not a resolution, and this one lasted a day. Worth remembering
the next time an exception looks cheaper than a fix.

### Resulting status

- **[tech.textContrastFloor](local/tech/accessibility.md#textcontrastfloor)** amended: one floor, 7:1, links and text alike, against both surfaces.
- **[acceptance.noTextBelowContrastFloor](acceptance.md#notextbelowcontrastfloor)** amended: the measured assertion is now a single floor, and the watched-failing list
  gains the two restored colors.
- The test's two-tier machinery — a separate link floor, and the selector tokenizer that decided
  which rules were links — is deleted rather than left switched off. It is recoverable from history
  if the floors ever diverge again.
- **`sources.md`** unchanged: both WCAG criteria were already adopted, and only which one governs
  ordinary text has moved.
- Still claimed nowhere: any WCAG conformance level.

---

## drawingCarriesProvenance

**The drawing carries provenance, and the crossbar splits to do it**

Date: 2026-09-24
Decided by: Sameer
Status: accepted
Affects: [product.measuredVersusModeled](local/product/measurements.md#measuredversusmodeled), [product.segmentComposition](local/product/diagram.md#segmentcomposition)
Subject: `diagram.provenance`
Kind: New capability, with one reversal and two exceptions worth defending

### What changed

Stroke style now encodes provenance: solid measured, dashed modeled, head hollow. The wing crossbar
is split at the wing chord, so its measured outer portions stay solid even when the span is
modeled. Both pages document the code. [product.measuredVersusModeled](local/product/measurements.md#measuredversusmodeled) states it.

### The reversal: what dash used to mean

Dash previously meant *"this is the tail, not the body"*. It now means *modeled*, and the tail —
being measured — becomes **solid**. The tail is still distinguishable: it keeps its own colour and
the junction mark, which was always doing most of the work. This is a strictly better use of the
channel, since tail-versus-body was a distinction the drawing already made twice.

### The crossbar split, and an argument that was wrong on the first pass

The initial objection to splitting was this: `span = a × chord^b` is multiplicative
([fittedWingspanModel](#fittedwingspanmodel)), so the span does not decompose into a measured part and a modeled remainder,
and drawing it as though it did would assert a structure the model lacks.

**That objection was aimed at the wrong claim.** It refutes "the span is built from a measured piece
plus a modeled piece", which nobody proposed. What is actually drawn is a *measured length overlaid
on* the span: the chord is a real measurement of a real anatomical stretch — carpal joint to primary
tip — which genuinely lies at the outer end of a spread wing. Overlaying a measured length is not
the same operation as decomposing a modeled one.

The split is also **more** honest than dashing the whole crossbar, which is what the wrong argument
would have produced. If the span model is off by 20%, the outer segment is still exactly the length
it claims; the entire error lands in the inner segment. Measure inward from the wingtip and you get
truth, outward from the body and you get model. Dashing the whole bar would have spread the
uncertainty over a stretch that does not carry any.

Checked against all 841 species before adopting, because the argument only holds if the data
cooperates:

| | |
| --- | --- |
| chord provenance | measured 831, near-match 10, **modeled 0** |
| chord ≥ half-span (which would put the tick outside the wing) | **0 species** |
| widest chord relative to half-span | 0.845, Calliope Hummingbird |

There is no species for which the outer segment is not a measured length.

Residual approximations, recorded rather than waved away. The chord is measured on a *folded* wing;
on a spread wing the hand sweeps back slightly, so its projection on the span axis is marginally
shorter and the tick sits a touch too far inboard — the same order of approximation as the head
oval's aspect ratio. And the tick's *position* is derived from a modeled span even though its
distance from the wingtip is not, which is why nothing calls it "the carpal joint".

### Exception 1: the neck's dash pattern differs, deliberately

The body takes the crossbar's exact treatment — 2.5px, 6/3, round caps. The neck keeps a finer 3/2
on butt caps.

This looks like an inconsistency and is not. Round caps extend every dash by half the stroke width
at each end, so 6/3 at 2.5px paints 8.5px and leaves 0.5px. A body is tens of pixels long and
carries that happily. A median neck is about 8px at its own scale and a hummingbird's about 6, so a
single lozenge covers the whole segment and the neck reads **solid** — the exact opposite of what it
is there to say. Caught by the owner on sight, after the matched treatment was tried and shipped to
the prototype.

Recorded because the next person to read the stylesheet will want to tidy the two into agreement.

### Exception 2: three of the five segments never vary

Head, neck and body are modeled for all 841 species, so that part of the signal is constant — worth
learning once, not worth shouting on every tile. Only the wingspan's provenance differs bird to
bird, which is why the crossbar gets the loud treatment and the vertical axis is quieter. The head
is hollow rather than dashed for the same reason as well as the geometric one.

A constant signal is still worth drawing. It is the answer to "which parts of this shape are real",
which a reader can only ask once but must be able to answer.

### Resulting status

- **[product.measuredVersusModeled](local/product/measurements.md#measuredversusmodeled)** added; **[product.segmentComposition](local/product/diagram.md#segmentcomposition)** amended, since the tail's stroke style now means something else.
- **[acceptance.diagramDistinguishesModeled](acceptance.md#diagramdistinguishesmodeled)** added.
- `dataset.Species.WingChord` added: AVONET's `Wing.Length`, previously consumed by the wingspan
  model and discarded. Note this is the **ornithological** chord, carpal joint to primary tip, not
  the aerodynamic chord of leading edge to trailing edge — the two share a name, and only the first
  can be drawn along a crossbar.
- The `?prov=1` prototype flag and its head/body variant switches are **removed**; the rendering is
  unconditional.

---

## fullInaturalistCoverage

**Total iNaturalist coverage, reached by widening the evidence rather than lowering the bar**

Date: 2026-09-27
Decided by: Sameer
Status: accepted
Amends: inaturalistLinkAdopted
Affects: [product.inaturalistLink](local/product/presentation.md#inaturalistlink)
Subject: `presentation.detail_view.external_link`
Kind: Requirement tightened, after diagnosis showed the shortfall was artifactual

### What prompted it

[inaturalistLinkAdopted](#inaturalistlinkadopted) shipped with 811 of 841 species linked and treated the other 30 as correct
absences: species iNaturalist had not confidently matched, mostly assumed to be splits it had not
followed. The owner asked for **100% coverage**.

That request sits directly against inaturalistLinkAdopted's central rule, that a wrong bird is worse than no link.
Reaching totality by relaxing the match would have traded a visible absence for an invisible error
— the worse failure. So the 30 were diagnosed one at a time first.

### What the 30 actually were

Not one was a genuine disagreement. Every one was a formatting or synonym artifact:

| | n | cause |
| --- | --- | --- |
| Hawaiian okina | 15 | `ʻ` is U+02BB, Unicode category **Lm** — a *letter* — so `unicode.IsLetter` kept it and `ʻApapane` never matched `Apapane` |
| ABA parenthetical alternates | 6 | "Chinese Hwamei (Hwamei)": the checklist lists both authorities' names, as its own header says |
| British spelling | 2 | "Grey-tailed Tattler" against "Gray-tailed Tattler" |
| Generic reassignment | 5 | iNaturalist uses *Leuconotopicus* and *Hesperiphona* where ABA uses *Dryobates* and *Coccothraustes* |
| Search failure | 1 | `Alle alle` — the plain search returns a mustard plant and a domestic cat |
| Preferred-name mismatch | 1 | *Poecile cinctus* |

### The one that looked genuine, and was not

*Poecile cinctus* appeared to be a real vernacular disagreement: iNaturalist prefers **Siberian
Tit**, ABA says **Gray-headed Chickadee**. Linking it looked like it would need either a
hand-curated override or scientific-name-only evidence.

Asking for every vernacular rather than the preferred one settled it. iNaturalist holds
`["Siberian Tit", "Gray-headed Chickadee", "Grey-headed Chickadee", "Alaska Chickadee"]`. **The
authorities agree**; they disagree only about which name leads.

So comparing against all English names is *more* evidence at the same bar, not a lower bar. Every
other rule below is the same shape: a name that was always the same name, previously compared in a
form that could not see it.

### The one rule that genuinely loosens, and why it is safe

Accepting a different genus is the only change that admits a pair the old rule refused on the
merits. It is bounded two ways: the species epithet must match, **and** a common name must still
agree. An epithet match alone would be far too loose, and is never accepted alone. *Dryobates
villosus* links to *Leuconotopicus villosus* because both are the Hairy Woodpecker; it does not link
to *Leuconotopicus borealis*, and the criterion asserts that.

### Also found

`exact_query` is not a parameter the iNaturalist API recognises. Passed anyway, it is silently
ignored and the endpoint returns its default popular taxa — Mallard, Western Honey Bee. A resolver
trusting it would have linked every species to a Mallard while appearing to work. Recorded because
the failure mode is silent and the parameter name is plausible.

The search endpoint was also replaced with `/v1/taxa/autocomplete`, which finds `Alle alle` where
the other cannot.

### Resulting status

- **[product.inaturalistLink](local/product/presentation.md#inaturalistlink)** amended: both names must still agree, but each is compared more thoroughly.
- **[acceptance.inaturalistLinkConfidentOnly](acceptance.md#inaturalistlinkconfidentonly)** extended: a case per bucket, plus the assertion that a differing genus is refused when
  the common name does not agree.
- Coverage is now a **required** 841 of 841, asserted over the built dataset rather than reported.
  A species that stops resolving is a fault, not an absence.
- The normalizer drops separators rather than collapsing them, because an okina sits *inside* a
  word and collapsing it to a space merely moves the mismatch: `Hawaiʻi` became `hawai i`.

---

## perSectionDefault

**Per-section becomes the default, and the legend becomes a diagram**

Date: 2026-09-27
Decided by: Sameer
Status: amended by frequencyGlobalScale
Affects: [product.scaleToggle](local/product/diagram.md#scaletoggle), [product.pageExplainsItself](local/product/presentation.md#pageexplainsitself)
Subject: `diagram.scale.mode`, `presentation.help`
Kind: Direct — a later decision reverses a default an earlier one argued for

### What changed

The scale default returns to **per-section**, listed first. The help's diagram section becomes a
real labeled diagram instead of a list of colour swatches, and the three sections are reordered:
the diagram, where the numbers come from, then the controls.

### Why the default moves, against product.scaleToggle's own argument

[product.scaleToggle](local/product/diagram.md#scaletoggle) argued the case for global plainly and it has not become wrong: *"the cross-family surprise
**is** the product… a reader could scroll the whole list without ever seeing that a Trumpeter Swan
dwarfs a kinglet."* That is still true, and global is still one click away and still named first
in the help text.

What changed is the diagram. Since that default was chosen it has acquired provenance in the stroke
style ([drawingCarriesProvenance](#drawingcarriesprovenance)), a split crossbar with chord ticks, and labels ([product.labeledDetailDiagram](local/product/measurements.md#labeleddetaildiagram)). At
global scale a median bird is a **few pixels** tall — measured earlier at 1.7px for a robin's neck
and 0.4px for a hummingbird's — so none of that vocabulary is visible on the landing page. The
default was hiding the drawing's own language from every first-time reader.

The trade is real and goes both ways, which is why it is recorded rather than quietly flipped. The
new default costs the immediate cross-family shock; it buys a reader who can actually read a bird,
and who is better placed to be surprised when they do switch to global. The surprise survives one
click; the vocabulary did not survive the default.

### Why the legend becomes a diagram

A swatch list asks the reader to map six colour chips onto a drawing. A labeled example **is** the
drawing with the answers written on it — and BirdView now has exactly that machinery, built for the
detail view, so the help can show the real thing rather than describe it.

It also removes a class of rot the old legend was prone to: the swatch list described the tail as
"dashed, to tell it from the body", which stopped being true when drawingCarriesProvenance gave dash a different
meaning, and nothing caught it because the text and the stylesheet had no connection. A rendered
diagram cannot drift from the renderer.

### Resulting status

- **[product.scaleToggle](local/product/diagram.md#scaletoggle)** amended: per-section is the default and is listed first.
- **[product.pageExplainsItself](local/product/presentation.md#pageexplainsitself)** amended: three sections in a stated order, and the diagram section is a rendered,
  labeled diagram.
- **[acceptance.sharedScale](acceptance.md#sharedscale)** and **[acceptance.pageExplainsControls](acceptance.md#pageexplainscontrols)** amended, and the tests that asserted the global default inverted.
- The help diagram is drawn from a real species with a **modeled** wingspan, so the dashed inner
  crossbar of [product.measuredVersusModeled](local/product/measurements.md#measuredversusmodeled) is visible in the legend rather than only in the prose beside it.

---

## compareUrlContract

**The compare URL contract changes when the index becomes the picker**

Date: 2026-10-01
Decided by: Sameer
Status: amended by speciesUrlOutlivesSpecies
Affects: [product.pickFromIndex](local/product/presentation.md#pickfromindex), [product.comparisonOneScale](local/product/presentation.md#comparisononescale)
Subject: `presentation.compare_view`, `presentation.pick_mode`
Kind: Direct — a later decision reverses two conditions an earlier one set

### What changed

Two behaviors settled days ago now resolve the other way, both because the same premise
disappeared. [product.pickFromIndex](local/product/presentation.md#pickfromindex) gives the index a picker, which means the index now reads
the `species` parameter that the compare view writes.

**a. Leaving the compare view keeps the species list.** Commit `10e938d` removed it, and
[acceptance.comparisonSharesScale](acceptance.md#comparisonsharesscale) asserted that no link back to the index may carry `species`.

**b. An unrecognised id is dropped and named, not a 404.** `/compare` refused the whole page; it
now renders the species it recognises and says which it could not.

### Why (a) reverses

The fix in `10e938d` was correct and its reasoning was sound: *"the index does not read `species`,
so the parameter is copied into every control link on that page and follows the reader through
every toggle afterwards."* A parameter that means nothing where it lands is litter, and litter
that propagates is worse.

That argument rests entirely on the index not reading it. [product.pickFromIndex](local/product/presentation.md#pickfromindex) makes the index read it — the
species list is now the live state of a pick in progress. Carrying it back is no longer litter; it
is the whole point. Dropping it would break the plainest flow the picker has: open a bird to look
at it, go back, add it.

What still drops the list is **Clear**, which is how a reader leaves pick mode. So the mechanism
added in `10e938d` survives; only the control it hangs on moves.

### Why (b) reverses

The 404 was justified as *"a shared link with a typo should say so rather than quietly render a
shorter page."* The goal was honesty about loss, and that goal has not changed. What changed is
that a third option exists, which was not considered at the time: **render what is recognised and
name what is not.** It satisfies the original reasoning better than refusing the page does — the
reader is told exactly what was lost *and* still gets the comparison.

The case that decides it is not a typo but **rot**. Species ids are derived from scientific names,
and this checklist renames them: the Sandhill Crane is `antigone_canadensis`, and was
`grus_canadensis`. A comparison URL is made to be shared and kept, so it will outlive some of the
ids in it. Under the 404 rule a five-bird comparison dies completely when one genus is reassigned.
Under drop-and-name it loses one bird and says so, and every link the page then builds omits the
dead id, so the set repairs itself on the next click.

The index does the same thing for the same reason, so one parameter has one meaning everywhere.

### What this costs

A reader who mistypes an id gets a page rather than an error, and must read the notice to learn
that something was dropped. That is a real regression in loudness, accepted because the rotted
link is both likelier and more harmful than the typo: a typo is made and corrected in seconds by
the person who made it, while a rotted link fails for someone who did not write it and cannot
tell what was supposed to be there.

## massAsWaterCube

**Mass gets a shape after all: the cube of water**

Date: 2026-10-01
Decided by: Sameer
Status: amended by massCubeOnTiles
Affects: [product.labeledDetailDiagram](local/product/measurements.md#labeleddetaildiagram)
Subject: `presentation.detail_view.diagram_labels`
Kind: Direct — a later decision reverses one an earlier one argued for, on finding one of its two reasons did not hold

### What changed

[product.labeledDetailDiagram](local/product/measurements.md#labeleddetaildiagram) said mass is drawn as a number **and not as a shape**, one day before
this. It now also carries a square in the lower-left corner: one face of the cube of water that
weighs what the bird weighs, drawn at the diagram's own scale.

### Why it reverses

[product.labeledDetailDiagram](local/product/measurements.md#labeleddetaildiagram) gave two reasons and they are not equally good. The first — that a mass glyph puts a
second, arbitrary scale on a figure whose every other dimension shares one, which
[product.crossbarAtVerticalScale](local/product/diagram.md#crossbaratverticalscale) refuses — **does not apply to this glyph.** One gram of water is one
cubic centimetre by the historical definition of the gram, so ∛(mass) is a unit conversion and
not a model: it assumes nothing about the density of birds, and it is a real length that draws at
the diagram's own px/mm. The objection was to arbitrary scales, and this is not one. The earlier
text conflated the general case with the specific proposal and rejected both.

The second reason stands and is recorded in [product.labeledDetailDiagram](local/product/measurements.md#labeleddetaildiagram) as a limitation rather than deleted: **the cube
compresses.** A Bushtit and a Common Raven differ 175× in mass and their cubes differ 5.6×. That
is correct for equal-volume cubes and understates the gap. What changed is that it no longer has
to carry the magnitude — the mass label was added in the same breath, so the number states the
175× and the cube says what 928 g of bird is *like*.

### What the cube earns its place with

It shows the one thing a length-and-wingspan diagram cannot: **bulk against elongation.** Read
against the bird's own length, the largest cubes in the set belong to alcids — Least Auklet,
Marbled Murrelet, Whiskered Auklet, Parakeet Auklet, dense diving birds at about 27% of the
drawing height — and the smallest to long-tailed, light-bodied species: Fork-tailed Flycatcher,
Scissor-tailed Flycatcher and Black-billed Magpie, at 10–13%. Nothing else on the figure
distinguishes a bird that is long from a bird that is heavy.

*(Re-measured after [trunkExponentCalibrated](#trunkexponentcalibrated): the alcid end read 28–29% before heavy birds lengthened.
The comparison and its direction are unchanged.)*

Measured before it was built, at the detail view's geometry: 28 px smallest, 52 px median, 80 px
largest, in a drawing area 668 × 281. In the lower-left corner at an 8 px margin the crossbar
intrudes for **0 of 841 species**, the tightest clearance being 50 px on Western Grebe.

### The risk being accepted

A reader who takes the cubes as *the* mass comparison reads them wrong, by a factor of the cube
root. The mitigation is that the number is right beside it and the diagram key names what the
square is. If it proves misleading in use, the thing to remove is the square and not the number —
the number is what [product.labeledDetailDiagram](local/product/measurements.md#labeleddetaildiagram) requires.

## massCubeOnTiles

**The mass cube goes on the index tiles too**

Date: 2026-10-01
Decided by: Sameer
Status: accepted
Amends: massAsWaterCube
Affects: [product.tileContents](local/product/presentation.md#tilecontents)
Subject: `presentation.tile`
Kind: Direct — a later decision reverses a scope an earlier one set

### What changed

[massAsWaterCube](#massaswatercube) put the cube of water on the annotated diagrams and kept it off the index
tiles, and [acceptance.massDrawnAsWaterCube](acceptance.md#massdrawnaswatercube) asserted that exclusion. The tiles now carry it as
well, in the same lower-left corner, **without the mass label** — a tile carries no labels at all
([product.tileContents](local/product/presentation.md#tilecontents)), and the mass is already in its footer.

### Why it reverses

The exclusion rested on two claims. One was crowding: *"the tile is dense by design and this is
annotated-diagram furniture, as [product.labeledDetailDiagram](local/product/measurements.md#labeleddetaildiagram)'s labels already are."* That reasoning holds for **labels**,
which need a gutter a third of the width and a leader line to the part they name. It does not
transfer to the cube, which needs neither: it sits in a corner the drawing does not use, and
measured across the whole dataset it never reaches the crossbar — **0 overlaps in 841**, which is
the claim that matters and which [trunkExponentCalibrated](#trunkexponentcalibrated) re-verified after every length moved.

The accompanying figure was "the tightest clearance is 7.5px on Western Grebe". Re-measuring at
per-section scale as the vertical gap between the cube's top vertex and the crossbar, over the
tiles where the two share any horizontal extent, gives **23.6px on Least Grebe** — and gives
16.8px on Eared Grebe for the *old* dataset, where the original method gave 7.5px on a different
bird. The two methods therefore do not measure the same gap, and the numbers should not be read
as a before and after. The invariant both support is the one stated above.

The other was that an unexplained shape reads as part of the bird. That was true when the cube was
introduced and is no longer: the Presentation panel now explains it, and the tile footer names
mass in full rather than as `M`.

What the exclusion cost is the comparison the index is *for*. Every tile is drawn to a shared
scale ([product.scaleToggle](local/product/diagram.md#scaletoggle)), so cubes on tiles are directly comparable down a whole section — and the thing the
cube shows, bulk against elongation, is exactly what a reader scanning a section of similar-length
birds cannot otherwise see. Confining it to the detail view meant it could only ever be read one
bird at a time, or two or three on a comparison page assembled by hand.

### What this does not change

The cube stays off the tile in no other way: no label, no leader, no footer row of its own. The
tile's contract is unchanged apart from the shape itself.

## trunkExponentCalibrated

**The trunk's mass exponent is calibrated against skeletons, not invented**

Date: 2026-10-02
Decided by: Sameer
Status: accepted
Affects: [product.trunkFromMass](local/product/diagram.md#trunkfrommass)
Subject: `diagram.trunk_model`
Kind: Direct — a measurement contradicts a constant the project chose by eye

### What changed

`BodyLength` modeled the trunk as `18.0 x mass^0.31`, an unvalidated constant carried since the
first ingest. It is now **`14.0329 x mass^0.37`**, and the exponent is pinned to published skeletal
measurement by [product.trunkFromMass](local/product/diagram.md#trunkfrommass) and [acceptance.trunkScalesLikeSkeletons](acceptance.md#trunkscaleslikeskeletons).

### Why

The question was raised by an Ostrich. Measuring what a worldwide scope would do to the models
produced a bird with a 360 mm head on a 119 mm neck and a total length of 1,539 mm against a real
2.1–2.75 m, which prompted the right follow-up: *is the model also wrong for the 841 we already
have?*

It is, and the evidence converged from three independent directions. Regressing log dimension on
log mass:

| dimension | n | exponent | 95% CI |
| --- | ---: | ---: | :---: |
| measured wingspan ([`fox-wingspan`](imports.md#fox-wingspan)) | 301 | 0.368 | — |
| neck length ([`bohmer-neck`](imports.md#bohmer-neck)) | 103 | **0.380** | 0.338 – 0.421 |
| total leg length ([`bohmer-neck`](imports.md#bohmer-neck)) | 103 | **0.364** | 0.328 – 0.401 |
| | | | |
| *BirdView trunk, as shipped* | — | *0.310* | — |
| *BirdView composite total length* | — | *0.302* | — |

**The wingspan result alone was not enough, and was nearly discarded.** Large birds are
disproportionately long-winged — the heaviest species with measured spans are albatrosses, condors
and pelicans — so 0.368 looked inflated by an aerodynamic confound, and the first verdict was
"inconclusive, sits at the bottom of the plausible range". Necks and legs settle it: neither is an
airfoil, neither is under selection for span, and both land in the same band. Three dimensions
agreeing at 0.364–0.380 is a biological signal, not an artifact of flight.

0.310 falls **below** the 95% interval of neck, total leg and tibiotarsus alike. It sits inside
the interval only for the two single-element measures — femur 0.325, central cervical vertebra
0.314 — which is exactly where isometry is expected, and exactly what a total length is not.

Within the current scope's 4,258× mass range, the gap understated the large end by **1.65×**.

### Why 0.37, and why the coefficient moved with it

0.37 is the round value nearest the mean of the two non-aerodynamic estimates (0.372), and the
wingspan estimate at 0.368 corroborates it. It sits comfortably inside the band [product.trunkFromMass](local/product/diagram.md#trunkfrommass) requires.

Raising the exponent alone would have lengthened every bird, which is not what the evidence says
— the evidence is about *relative* size across the range. The coefficient is therefore re-anchored
so the model is unchanged at the median mass of 63.4 g:

| species | mass g | before | after | change |
| --- | ---: | ---: | ---: | ---: |
| Calliope Hummingbird | 3 | 24.2 | 20.0 | 0.83× |
| Redpoll | 13 | 39.9 | 36.2 | 0.91× |
| American Robin | 78 | 69.6 | 70.5 | **1.01×** |
| Mallard | 843 | 145.3 | 169.7 | 1.17× |
| California Condor | 8,443 | 296.8 | 398.1 | 1.34× |
| Trumpeter Swan | 11,071 | 322.8 | 440.0 | 1.36× |

### What this does *not* establish

Recorded at length, because the temptation to over-read it is strong.

**Böhmer never measured a trunk.** The inference runs neck + leg → trunk, and a trunk is not
obviously like either: it houses the flight apparatus and may be more conserved. This is the
weakest link in the argument, and the thing a better source would most repay.

**The direct species-level comparison is not evidence, in either direction.** Joining Böhmer's
taxa to BirdView gives 19 exact matches, on which BirdView's modeled neck is a median 29% of
Böhmer's measured neck. That is expected: Böhmer sums cervical centrum lengths, including the part
inside the body contour, where [product.neckModelPerFamily](local/product/diagram.md#neckmodelperfamily) models the *exposed* neck of a
silhouette. A raven's cervical column is 106 mm; its visible neck is not. The apparent mass trend
in that join (+0.115, CI +0.001 to +0.229) points the opposite way from the exponent argument and
is an artifact of which families were sampled — its large taxa sit in long-neck families (Anatidae
0.70, Phoenicopteridae 0.95, Gaviidae 0.50), its small ones in short-neck families (Alcedinidae
0.10). The family table drives it, not biology.

The decision rests only on the pure scaling fits, which involve no BirdView model at all.

**The head model is untouched** ([product.craniumFromMass](local/product/diagram.md#craniumfrommass)). Single-element data supports isometry
for a cranium, so the cube root stays — along with its own unresolved problem, recorded there.

**Total length is still `modeled`.** A better exponent is not a measurement. [noOpenLengthAndWingspan](#noopenlengthandwingspan) and
[product.provenancePerMetric](local/product/measurements.md#provenancepermetric) are unaffected, and every length stays dashed and marked.

### What it costs

Every length in the dataset moves, so section boundaries, the global scale anchor and the ordering
within sections all shift. That churn is why this lands as its own commit, ahead of the AviList
names work rather than beside it: two large dataset changes at once would be impossible to tell
apart in review or in a screenshot.

## avilistReplacesAba

**AviList replaces the ABA checklist for names, and the eBird code becomes the URL**

Date: 2026-10-02
Decided by: Sameer
Status: amended by scopeFromRecurrence
Affects: [product.namesFromAvilist](local/product/dataset.md#namesfromavilist), [product.englishNameChoice](local/product/diagram.md#englishnamechoice), [product.familiesByOrder](local/product/months.md#familiesbyorder)
Subject: `dataset.names.source`
Kind: Direct — a licensing constraint forces a source change, which then enables a feature

### What changed

Every name BirdView serves now comes from [AviList](imports.md#avilist) (CC BY 4.0) instead of
the ABA checklist (all rights reserved): both English names, the family scientific and English
names, the order, and the family sequence. The four-letter alpha code is replaced by the **eBird
species code**, which also becomes the **species id and URL**.

### Why

**To unblock deployment.** Serving BirdView reproduces whatever supplies its names, so an
all-rights-reserved checklist made a public instance a redistribution the project has no grant
for. The ABA was asked on 2026-09-27 and has not replied. This removes the dependency rather than
waiting on it.

The objection that defeated AviList a day earlier was that its English names are IOC usage —
Little Auk for Dovekie — and wrong for this audience. The extended release answers it: it carries
the **Clements** name as well, which is this audience's usage. Measured over BirdView's scope, the
Clements column reproduces the ABA name for **827 of 827** species where both exist, the only
five differences being ABA's own `X (alternate)` notation. One file, both vocabularies — which is
what turned a loss into [product.englishNameChoice](local/product/diagram.md#englishnamechoice).

### Why the code changed too, which was not the original plan

The alpha code was going to be dropped: it comes from the ABA checklist, and the IBP list it
otherwise lives in publishes **no licence statement at all**, which is the same reason AOS was
rejected. The owner asked whether eBird's code could replace it, and the measurement made it
close to forced:

| | |
| --- | --- |
| coverage over scope | **829 of 832** matched species, and 836 of 839 after the synonym table |
| collisions | **zero** |
| licence | the same AviList file — no new source, no new permission |
| worldwide | **11,118 of 11,131 (99.9%)**, where IBP alpha codes do not exist outside the AOS area |

The decisive argument is none of those. **An alpha code is derived from the English name**, so
under the world locale `DOVE` would have to become `LIAU`. Adding localization is precisely what
breaks a name-derived code and leaves an assigned one intact. The feature and the identifier
could not both be had the old way.

Making it the URL follows: `/species/amerob` beats `/species/turdus_migratorius` for typing and
for stability, since a genus move no longer moves the URL — and this dataset contains seven genus
moves.

### What it costs

**Two species leave the dataset**, 841 → 839. AviList lumps Gray-headed Swamphen into *Porphyrio
porphyrio* and Western Warbling-Vireo into *Vireo gilvus*, and both lumps land on taxa already in
scope. Keeping them would mean keeping an ABA name — the one thing this change exists to remove —
and lumping them in place would give two birds one URL, because *Vireo gilvus* is already
BirdView's Eastern Warbling-Vireo. Dropping them is the only option consistent with the decision:
if AviList supplies the names, it supplies the species concepts.

**Every species URL changes.** Nothing links to BirdView yet, so this is free now and would not
be later — which is an argument for doing it now rather than after a deployment.

**55 of 90 family English names change**, accepted by the owner in advance. Two family scientific
names change (Sylviidae → Paradoxornithidae, Icteriidae → Icteridae) and four orders, including
three vultures moving to **Cathartiformes** — which is why Accipitriformes is relabelled "Hawks
and Eagles": it no longer holds a vulture.

**The family sequence is now AviList's**, not the ABA's, so the family selector and the section
order both reshuffle. Taking the names from one source and the ordering from another would have
left the ordering as the last thing the ABA still supplied.

### What this does *not* do

**It does not unblock deployment on its own.** The ABA checklist is still read at build time for
**scope** — which 841 species are in — and scope is the remaining dependency. [tech.fetchRestrictedInputs](local/tech/delivery.md#fetchrestrictedinputs) already
permits fetching it at build; what it does not permit is serving it. The next step is a list
BirdView derives and owns.

**It does not touch any measurement.** AVONET and Fox are unaffected, every provenance record is
unchanged, and no number moved.

**It is not a route into Cornell's data.** [requirements.md](imports.md#avilist) records why the
Clements and Cornell columns are usable here when eBird/Clements itself is rejected: AviList is a
separate CC BY 4.0 publication, and nothing is taken from Cornell's own distribution.

## regionsTwoWays

**Regions are grouped two ways: by shared birds, and by geography**

Date: 2026-10-03
Decided by: Sameer
Status: accepted
Affects: [product.regionAndMonthFilter](local/product/filtering.md#regionandmonthfilter), [product.regionsGroupedTwoWays](local/product/months.md#regionsgroupedtwoways)
Subject: `presentation.region_filter.zones`
Kind: Reversal — [product.regionAndMonthFilter](local/product/filtering.md#regionandmonthfilter) said zones "claim nothing about birds". They now claim something about birds, because a corrected measurement says they can.

### What changed

The region selector offers **two groupings** instead of one. **Common birds** is derived by
clustering regions on how much their species lists overlap. **Geographic** is the existing
hand-made grouping, kept as a convenience. A region sits in exactly one zone of each, so it
appears twice in the selector.

### Why the old claim was wrong

[product.regionAndMonthFilter](local/product/filtering.md#regionandmonthfilter) said there was "no data-driven way to subdivide the east", citing a 42-member block
covering everything from Maine to Georgia. **That measurement was an artifact.** It came from
faceting GBIF's `/occurrence/search` on `speciesKey`, which silently drops records whose names
the backbone cannot resolve — the same trap recorded in [requirements.md](imports.md#gbif), which
reports **zero** US and Canadian records for Cooper's Hawk.

Redone against `data/region-presence.tsv` — built from **verbatim** names, which is the table the
app already serves — the east subdivides cleanly into four groups. The old claim was a statement
about a broken query, not about birds, and is withdrawn.

### How the bird zones are derived

Average-linkage agglomerative clustering on Jaccard distance over each region's species set, cut
at k=26 while the groups are still sharp; every cluster of two or more becomes a seed, and each
remaining region joins the seed it most resembles.

**Cutting high and then assigning, rather than simply cutting lower, is forced by chaining.** At
k=10 average linkage has collapsed the Pacific into the Southwest; by the time Texas has company,
the distinctions worth keeping are gone. Cutting at k=26 and assigning afterwards keeps both.

Accepted placements run **0.652 to 0.825** similarity — a tight band, so none is marginal. Four
regions fall outside it and are left alone rather than forced into a group:

| Region | Best match | Why it stands apart |
| --- | --- | --- |
| Hawaii | 0.239 | Nothing on the mainland is close; the next-lowest accepted is 0.440 |
| Nunavut | 0.440 | High arctic, below the floor |
| Navassa | 0.010 | 4 species, sharing **zero** with any other island |
| the atolls | ≤0.077 | Johnston, Palmyra and Wake match each other at 0.44–0.55, the mainland at noise |

Midway joins Hawaii at 0.378, which is right: Midway is in the Hawaiian chain. So the old
"Outlying Islands" zone splits three ways under this grouping — Hawaiian Islands, Tropical
Pacific Atolls, Caribbean — which the owner authorized in advance.

### Why keep the geographic grouping at all

Because the two answer different questions, and the earlier framing wrongly treated them as
rivals. A reader looking for their state wants the grouping they already carry in their head; a
reader asking which areas share birds wants the measured one. The clustering scatters the Arctic
— Alaska to the Pacific Northwest, Yukon and the Northwest Territories to the Prairies, Nunavut
alone — which is defensible on birds and useless for finding Alaska. Offering both costs one
extra block of chips and settles the question without overriding the data.

### A zone name must not promise a boundary the birds do not keep

Reported 2026-10-04: selecting the geographic *Northeast* showed Connecticut and Rhode Island
in the derived *Mid-Atlantic* group rather than with the rest of New England, which reads as an
error. It is not one, and the investigation is worth keeping because the conclusion was the
opposite of the obvious fix.

**Rhode Island is a genuine finding.** Its single closest region anywhere is **Delaware**
(0.866), ahead of Connecticut (0.862) and New Hampshire (0.861). Southern New England shares
its avifauna with the coastal plain, not with the boreal north.

**Connecticut is marginal and lands there for a structural reason.** Its three nearest
neighbours are New Hampshire, Massachusetts and Rhode Island — all New England — but assignment
compares a region against a group's **mean**, and the northern group reaches Newfoundland (0.800
from Connecticut) and Prince Edward Island (0.742). Those drag its average below the compact
coastal-plain group's. Same artifact class as the islands earlier in this CR: a group's spread
distorting what joins it.

**Both repairs were measured and both lose.** Moving the two drops mean within-zone cohesion
over the pair from **0.8257 to 0.8239**. Splitting the northern group leaves five regions nearer
some other group than their own — Québec, New Brunswick and Nova Scotia all closer to New
England than to each other, Vermont closer to the Maritimes.

So the data was right and the **names** were wrong: "Mid-Atlantic" is a geographic label on what
is really the mild Atlantic coastal plain from Virginia to Rhode Island. They are now
**Mid-Atlantic & Southern New England** and **Northern New England & the Maritimes**, which
describe what they contain. A derived group earns its name from its members, not from the map.

### What it costs

**A region is no longer in one zone.** `Region.Zone` becomes a zone per grouping, and the
invariant weakens from "exactly one zone" to "exactly one zone **per grouping**" — still a
partition, still testable, but it is a real weakening and anything relying on the old shape
breaks loudly.

**The bird zones are only as good as the presence table.** They are derived from a generated
file, so a regeneration can move them. [acceptance.zonesAnswerToData](acceptance.md#zonesanswertodata) recomputes cohesion from
that table rather than pinning the names, so a drift that makes the grouping dishonest fails a
test instead of going unnoticed.

## regionUrlIsPostalCode

**A region's URL is its postal code, or its name, never an invented code**

Date: 2026-10-03
Decided by: Sameer
Status: accepted
Affects: [product.regionAndMonthFilter](local/product/filtering.md#regionandmonthfilter)
Subject: `presentation.region_filter.url`
Kind: Direct — the URL carried an internal database key, and the obvious fix would have required inventing identifiers the project has no authority to mint

### What changed

A region in a URL was its GADM id — `?region=USA.5_1`. It is now its **postal abbreviation**
where one exists (`?region=CA`), and its **name in PascalCase** where none does
(`?region=JohnstonAtoll`, `?region=GulfOfMexico`). The ten regions with no postal code now show
**no code at all** on their chip, just their name.

GADM ids remain the internal key: they key `data/region-presence.tsv` and they are the `gadmGid`
value sent to GBIF. Only the URL and the chip changed.

### Why not keep the GADM id

It is a database key doing a user's job. `?region=USA.5_1` cannot be typed, guessed or read, and
a single zone toggle enumerates seventy of them — the index page carries **198 KB** of region
hrefs, with a mean URL of 1,125 characters.

The argument for keeping it was real and was tested rather than assumed. GADM's own metadata
says *"The GID codes are persistent after version 3.6… If an area changes, for example if it
splits into two new areas, two new codes will be assigned, and the old code will not be used any
more."* So the ids are stable, and the stability case for changing them collapsed. What did not
survive is the readability case: the id is still a key, and a shareable `?region=TX` is worth
more than a stable string nobody can read.

### Why not invent two-letter codes, which was the first proposal

Because **the four obvious abbreviations are already taken**: Midway cannot be `MI` (Michigan),
Navassa cannot be `NV` (Nevada), Palmyra cannot be `PA` (Pennsylvania), Wake cannot be `WA`
(Washington). A complete two-letter scheme was found — `JA MW NI PM WK` and `NP CP NA GM AO`,
all ten clear of the 64 codes in use and of the eleven USPS codes outside the set — but every
one of the four forced codes is deliberately *not* what a reader would guess, chosen only to
dodge a collision.

That is the project minting identifiers in a namespace it does not own, and putting them beside
64 codes that are USPS and Canada Post standards, where nothing on screen distinguishes the
invented from the official. The owner rejected it. **A name is longer and invents nothing.**

### What it costs

**The ten chips lose their code.** An ocean chip read `N Pacific`, which was self-evidently not
a state; it now reads `North Pacific`. The only remaining cues that it is not a state are the
name itself and the Pelagic zone it sits in. That is accepted.

**Three marine names were shortened.** `North Pacific, off the west coast` would have slugged to
`NorthPacificOffTheWestCoast`, so the clarifier moved to a separate gloss shown in the chip's
tooltip, leaving the name proper. The slug is **derived** from the name rather than stored, so a
renamed region cannot keep a stale URL ([acceptance.regionUrlRoundTrips](acceptance.md#regionurlroundtrips)).

**Old URLs break, and no alias is kept.** Nothing links to BirdView yet, so this is free now and
would not be later — the same argument [avilistReplacesAba](#avilistreplacesaba) made for species URLs. Accepting both
forms would mean two identifiers for one region and a second thing to keep in step.

## familiesTwoWays

**Families are grouped two ways, and habitat is not one of them**

Date: 2026-10-03
Decided by: Sameer
Status: amended by familyInTwoGroups, habitatDietSpeciesLevel
Affects: [product.familiesByOrder](local/product/months.md#familiesbyorder)
Subject: `presentation.family_filter.groups`
Kind: Direct — the taxonomic grouping is unusable for navigation, and the obvious replacement turned out to be unsupportable

### What changed

The family selector offers **two groupings**. **Birder groups** is hand-drawn and offered
first — Shorebirds, Birds of Prey, Owls & Nightbirds. **Taxonomic order** is the existing
grouping from AviList. A family sits in exactly one group of each, so it appears twice in the
selector, exactly as a region now does ([regionsTwoWays](#regionstwoways)).

### Why the taxonomic grouping needed company

It is not wrong, it is lopsided. **43 of 90 families sit in Passeriformes**, rendered as one
chip block labelled "Perching Birds", while **15 orders hold exactly one family**. For 25 of
26 orders the taxonomic name is perfectly intuitive — Waterfowl, Owls, Tubenoses, Grebes — so
the whole navigation problem is one order, and no relabelling fixes it.

### Why not habitat, which was the first idea

AVONET carries a `Habitat` field per species, is already vendored under CC BY 4.0, and is
already read by the ingest — so this was free, and it failed anyway. **Habitat is a property of
a species, not of a family.** Assigning each family the habitat most of its species use
misfiles:

| Buckets | Species misfiled |
| --- | --- |
| AVONET's 12 habitats | 289/818 (35%) |
| collapsed to 5 | 202 (25%) |
| collapsed to 4 | 191 (23%) |
| collapsed to 3 — Land / Fresh Water / Coast & Ocean | 73 (9%) |

The breakage is worst in the families a reader cares most about: Gulls and Terns split Marine
18 / Coastal 18, a dead tie; New World Sparrows spread over five habitats; Sandpipers Wetland
27, Grassland 12, Coastal 7. Even Ducks, Swans and Geese is only 72% Wetland.

Nor is this a quirk of habitat. Across every AVONET trait, families large enough to matter are
almost never homogeneous — **only 12% of families with ten or more species share one habitat**
(16% for migration, 24% for lifestyle, 36% for trophic niche), and **81% of species sit in a
family that is mixed**. The families that look pure are mostly the 23 that hold a single
species. Three buckets is the only honest cut and its shape is Land 60 / Fresh Water 15 /
Coast & Ocean 14, which is true and not worth a row of chips.

**A hand-drawn grouping escapes this entirely, because it claims nothing about the birds.**
"Shorebirds" says Scolopacidae is a family birders call a shorebird. It does not say every
sandpiper lives on a shore, so there is nothing to misfile.

### Why this is not merely taxonomy relabelled

The two groupings disagree for **76 of 90 families (84%)**; only 14 are filed the same way by
both. Perching Birds splits across twelve birder groups; Birds of Prey draws from three orders
(Hawks and Eagles, Falcons, New World Vultures); Herons, Ibises & Storks from three. Selecting
in one grouping and watching the selection scatter in the other is the point, as it is for
regions.

### What it costs

**Every family chip renders twice**, once per grouping, as region chips already do.

**The grouping is this project's editorial opinion**, and the UI says so where it says the
Geographic region grouping is hand-drawn. It is not a measurement and must never be presented
as one.

**The table is keyed on the scientific family name.** [avilistReplacesAba](#avilistreplacesaba) renamed **55 of 90**
English family names and only **2** scientific ones, so the English name is the fragile key.

### What review changed, and the rule that came out of it

The first draft had 27 groups and the owner's review cut it to **25** and fixed three groups.
One failure mode produced all three, and it is worth naming because it recurred after being
pointed out twice:

> **A group must be something a birder says, not what is left over.**

- *Pigeons, Cuckoos & Nightbirds* bundled three unrelated families because they were mid-sized
  landbirds that fit nowhere else. Split: pigeons and cuckoos stand alone, and nightjars join
  **Owls & Nightbirds**, which is a real category — what you look for after dark.
- *Old World Warblers & Babblers* was to be folded into the thrushes and mimids on the reasoning
  that both are Old World passerines. That is a taxonomic observation, not a birder's thought.
  Kept separate as **Old World Songbirds** — nine birds of Eurasian lineage, which in this scope
  means Hawaii's introductions and a couple of Alaskan strays — and the **Wrentit** moved out to
  Chickadees, Nuthatches & Wrens, since a west-coast chaparral resident is not something anyone
  looks for under "Old World".
- *Swifts & Hummingbirds* paired the two because they share an order. **Swifts & Swallows** is
  the pairing that helps: "swift or swallow?" is the question a bird overhead provokes, and
  nobody has had to tell a hummingbird from a swift. Hummingbirds stand alone.

Two merges were approved on the same principle: **Seabirds** (tubenoses with boobies,
cormorants and pelicans) and **Waterfowl, Loons & Grebes**. Larks left the swallows for
**Sparrows, Buntings & Larks**, joining the longspurs and Snow Bunting they flock with in winter
stubble rather than the swallows they merely share open country with.

The lesson generalizes past this grouping: where a hand-drawn category is doing the work, the
test is whether a reader would say the name out loud, not whether the members have something in
common.

### What it does not do

**It adds no badge to a tile.** The tile carries the family badge alone
([product.tileContents](local/product/presentation.md#tilecontents)); groupings beyond family are navigation. Every tile is one height
and a second badge costs a line on all of them.

## familyInTwoGroups

**A family may be in two birder groups, and that is not a route to habitat**

Date: 2026-10-03
Decided by: Sameer
Status: amended by habitatDietSpeciesLevel
Amends: familiesTwoWays
Affects: [product.familiesByOrder](local/product/months.md#familiesbyorder)
Subject: `presentation.family_filter.groups`
Kind: Direct — [product.familiesByOrder](local/product/months.md#familiesbyorder) required a partition, and one family has two honest homes

### What changed

The birder grouping **covers** the families instead of partitioning them. The owls
(Strigidae, Tytonidae) are now under **Birds of Prey** as well as **Owls & Nightbirds**.

### Why

Owls sit on the seam in the whole scheme: some groups name what a bird *is* — Birds of Prey,
Waterfowl — and some name when or where you *meet* it — Shorebirds, Seabirds. Owls are
unambiguously birds of prey, and owling is unambiguously its own activity at its own hour. Both
sentences are true, and picking one made the grouping half wrong whichever way it went.

### The limit, which is the important half of this decision

Overlap looks like it reopens habitat: tag a family with *every* habitat its species use and
nothing is excluded. **Measured, that is worse, not better.** It does not remove the error, it
inverts it — from excluding birds that belong to including birds that do not:

| Select | Species returned | Actually that habitat | Precision |
| --- | --- | --- | --- |
| Land | 641 | 531 | 83% |
| Fresh Water | 378 | 145 | **38%** |
| Salt Water | 321 | 142 | **44%** |

Asking for freshwater birds would return every New World Warbler, because one warbler is a
wetland species. The single-assignment version misfiled 9% of species; this misfires on nearly
two-thirds of what it returns.

**So the rule is not "overlap is allowed" but "a tag must be true of the whole family."** Every
owl is a bird of prey, so Strigidae carries both tags honestly. Only 27 of 50 sandpipers are
freshwater birds, so no family-level habitat tag is honest in either direction. A property that
varies within a family is a **species** trait, and the place for it is a species filter — which
remains Work arising and is unaffected by this change.

*Waders versus floaters* was raised as a middle case and is a genuinely better axis than
habitat — 12 of 18 waterbird families are uniform on swims-versus-walks — but it fails where it
costs most. **Anatidae is the largest family in the dataset at 61 species and is 67/33 split**,
because geese walk and graze while ducks float; Rallidae splits the same way, coots against
rails. An axis that is wrong about ducks is not worth a row of chips.

### What it costs

**A group toggle now moves another group's header.** Clearing Birds of Prey leaves Owls &
Nightbirds showing as not fully selected, because the owls it holds have gone. That is correct
and it is also surprising, and it is the price of the overlap. With one overlapping family it
is explicable; with several axes it would be unpredictable, which is why
[acceptance.familyGroupingComplete](acceptance.md#familygroupingcomplete) caps the number of duplicated placements rather than trusting
restraint.

## habitatDietSpeciesLevel

**Habitat and diet become species-level filters, and the birder grouping is retired**

Date: 2026-10-03
Decided by: Sameer
Status: accepted
Amends: familiesTwoWays, familyInTwoGroups
Affects: [product.familiesByOrder](local/product/months.md#familiesbyorder), [product.habitatAndDietFilters](local/product/months.md#habitatanddietfilters)
Subject: `presentation.trait_filter`
Kind: Reversal — [product.familiesByOrder](local/product/months.md#familiesbyorder)'s hand-drawn family grouping is withdrawn and replaced by published per-species classifiers

### What changed

Two new filters, **Habitat** and **Diet**, taken from AVONET and assigned per species. The
hand-drawn **Birder groups** are deleted, and `internal/family` with them. Taxonomic order
stays; the panel is renamed **Species**, families being groups of species already.

### Why the birder grouping had to go

It was a family-level answer to a species-level question, and it failed in exactly that shape
every time. The errors found by review, in order: pigeons bundled with cuckoos; swifts with
hummingbirds rather than swallows; Old World warblers folded into the thrushes; the Anhinga — a
freshwater swamp bird — under *Seabirds*; the Phainopepla, a frugivore, among insect-eating
flycatchers; Fringillidae called seed-eaters when **only 13 of its 37 species eat seeds**,
the rest being Hawaiian honeycreepers that sip nectar and probe bark.

Patching each in turn was the wrong move, and the owner named why: **the grouping should sort
species, not families.** At species level the same AVONET data gets every one of those cases
right, because the information was never at the family level to begin with.

### Why AVONET's classification rather than our own

Measured earlier and recorded in familiesTwoWays: **only 12% of families with ten or more species share
one habitat, and 81% of species sit in a family that is mixed.** That killed habitat as a
family grouping. It says nothing against habitat as a *species* trait, which is what AVONET
publishes and what is adopted here.

Coverage is **839 of 839** through the ingest's existing name resolution — an earlier figure of
97.5% was an artifact of matching AviList names verbatim instead of using that path.

Three axes were considered. Habitat (2.83 bits) and Trophic Niche (2.49) are adopted;
**Primary Lifestyle is held back** — 5 options, a third of all birds in `Insessorial`, and a
`Generalist` bucket of 121 species that tells a reader nothing. Mutual information between the
axes is 0.66–0.83 bits, so none is redundant; lifestyle is simply the weakest and the most
jargon-laden.

### What BirdView adds, and the limit on it

AVONET records **one** primary value per species. Five supplementary assignments are added where
that hides something obvious — American White Pelican on the coast, Double-crested Cormorant
inland, Great Blue Heron on tidal flats, Osprey over salt water, American Herring Gull at
landfill. Each is marked **`birdview`** with a reason, never merged into AVONET's.

**Two of the first five did nothing at all**, keyed on `Phalacrocorax auritus` and
`Larus argentatus` where AviList says `Nannopterum auritum` and `Larus smithsonianus`. A map
lookup that misses is silent, so [acceptance.habitatDietFilterPerSpecies](acceptance.md#habitatdietfilterperspecies) now checks every key against the
dataset.

### What it costs

**The familiar birding vocabulary is gone from the controls.** "Shorebirds", "Sparrows" and
"Finches" are not habitats or diets, and AVONET has no vocabulary for them. A reader wanting
the shorebirds selects four families in the taxonomic block. That is the real price of this
change, and it is accepted because a grouping that is *wrong* about the Anhinga is worse than
one that is merely less convenient.

**One value per species is still a simplification**, and the control says so rather than
implying completeness.

## scopeFromRecurrence

**Scope becomes a consequence of a derived status label, and recurrence replaces abundance**

Date: 2026-10-04
Decided by: Sameer
Status: amended by countEbirdFirst
Amends: avilistReplacesAba
Affects: [product.scopeIsRecurringAvifauna](local/product/dataset.md#scopeisrecurringavifauna), [product.statusIsPublished](local/product/dataset.md#statusispublished), [product.declaredExceptions](local/product/dataset.md#declaredexceptions)
Subject: `dataset.scope`
Kind: Replacement — [product.scopeIsRecurringAvifauna](local/product/dataset.md#scopeisrecurringavifauna)'s ABA rarity-code filter is withdrawn in favor of a status vocabulary BirdView derives from open occurrence data and owns

### What changes

Two things, and the second is the point.

**Scope stops being a threshold.** Every attempt to derive scope from a *volume* of
observations has failed, and the next one would fail too. What replaces it is **recurrence**:
a species is in scope when its records in at least one region recur across years, however few
of them there are.

**Status becomes a product.** The ABA codes 1–6 that [product.scopeIsRecurringAvifauna](local/product/dataset.md#scopeisrecurringavifauna) filters on are not a scope rule at
all — they are a *status vocabulary*, and [product.scopeIsRecurringAvifauna](local/product/dataset.md#scopeisrecurringavifauna) uses three of its values as a cutoff. BirdView
deriving its own status labels makes scope a consequence of labeling rather than a knife-edge,
and the labels themselves become a navigation facet alongside Habitat and Diet
([habitatDietSpeciesLevel](#habitatdietspecieslevel)). A vagrant is no longer *excluded*; it is *labeled*, and the default view
does not show it. That turns the hardest judgement in the project into a filter the reader
controls.

This is the first piece of BirdView that is a dataset in its own right rather than an input to
the drawing: a derived table of species, status and the evidence behind it.

### Why abundance cannot be rescued

Occurrence counts measure **detection effort × detectability × observer behavior**, not
presence. The three failures recorded in Work arising are one failure:

- Pelagic species, remote-island endemics and cryptic or nocturnal birds are **under**-recorded,
  because observers are not where they are.
- Vagrants are **over**-recorded, because observers go to where vagrants are.

Those pull in opposite directions, so no single threshold can correct both — which is why the
four-year window, the distinct-places threshold, the local-against-continental ratio and the
two-dataset ppm rule each fixed one end and broke the other. At 2 ppm the best abundance rule
drops 43 species structurally: 32 pelagic seabirds and 11 remote-island endemics.

Two further biases are visible in BirdView's own measurements and were not in the original list:
19,214 iNaturalist records are identified only to genus, concentrated in the hard-to-identify
groups (*Empidonax*, peeps, gulls); and iNaturalist is photograph-driven, so nocturnal and
skulking birds are under-counted independently of how many there are.

### What recurrence measures instead

`GET /occurrence/search?facet=eventDate` returns **exact calendar dates with counts, over all
records**. That is what makes this measurable at all: the Work arising caveat warned that a page
of ≤300 records is a contiguous slice of GBIF's ordering rather than a sample, and this is the
facet it said any real rule would need.

Per species per region, dates are grouped into **episodes** — runs separated by more than 30
days, so one arrival is one episode however many people came to see it. Measured 2026-10-04:

| species / region | records | dates | span | years | reads as |
| --- | ---: | ---: | ---: | ---: | --- |
| American Flamingo / PA | 931 | 13 | 13 days | 1 | vagrant |
| Lucy's Warbler / VA | 53 | 6 | 7 days | 2 | vagrant |
| Wood Stork / NS | 21 | 9 | 17 days | 1 | vagrant |
| Akikiki / HI | 376 | 268 | 50 years | 45 | established |
| Whiskered Auklet / AK | 1,123 | 478 | 93 years | 49 | established |
| American Robin / PA | 1,498,010 | 1,200+ | 13 years | 13 | established |

**Recurrence is effort-independent in the way abundance is not.** A bird that is really there
gets recorded occasionally across decades however few observers pass by; a bird that is not gets
recorded in one burst. Akikiki has *fewer* records than the Pennsylvania flamingo and is
unmistakably the resident of the two. Lucy's Warbler in Virginia spans 2016-12-31 to 2017-01-06
and so defeated the four-year-window rule by straddling a boundary; its 7-day span does not care.

### What the probe corrected

**The signal only works per region.** Over the whole ABA Area it collapses: Fork-tailed
Flycatcher, a genuine accidental, shows 1,182 dates in 150 episodes across 65 years, because
accidentals accumulate independent events continent-wide. Only a species seen literally once
separates — Slender-billed Curlew, 1 record, 1925-10-01. So the generator must run per region,
which is the cost the caveat predicted: roughly 1,378 species × 70 regions.

**A truncated facet is a floor, not a measurement.** Facets cap at 1,200 values, so any abundant
pair reports exactly 1,200 dates. Truncation must be read as *established* and never as a date
count — the same shape as the vacuous checks this project has caught before.

**Pelagic species need the `SEA.*` regions**, which [`region`](../../internal/region/region.go)
already defines. A query for Black-capped Petrel against North Carolina's land polygon returns
6 records; the species has 13,601 in the ABA Area, nearly all offshore.

### Why the weights are declared and never fitted

The obvious way to build this is to fit coefficients so the output reproduces the ABA's 839.
**That is rejected.** It would re-target exactly what [`requirements.md`](imports.md#why-occurrence-data-cannot-reproduce-the-aba-scope)
disclaimed when the owner chose to own the list, and it would launder the ABA's editorial
judgement into model weights — a list that is theirs in substance and ours in format is not a
list BirdView owns. There is no other labeled training signal available, so there is no
supervised version of this model that is not that one.

Everything here is therefore either **measured** (episodes, dates, spans) or **declared** — a
stated coefficient with a written reason, as [`OrderCommonNames`](../../internal/taxonomy/mapping.go)
and the neck ratios already are. A declared weight can be defended in prose; a fitted one cannot.

### The bias-weighting model may not be needed, and that is the finding

This change was proposed as a per-habitat and per-family correction for under-reporting. The
probe suggests **the correction is unnecessary**, because recurrence is already insensitive to
the bias it would correct: the pelagics and endemics that an abundance rule drops score as
established on episodes without any adjustment for how rarely they are seen.

That is the better outcome. Every declared weight is an editorial judgement that has to be
defended, so a model that needs none is stronger than one that needs twelve. The weighting is
held back rather than abandoned, pending the full per-region run.

### What the data cannot do, and what must be declared instead

- **Four species have no open record at all** — Laysan Duck, Laysan Finch, Bermuda Petrel,
  Desertas Petrel. No statistic recovers them.
- **iNaturalist obscures dates for sensitive taxa**, which is where it was adopted to help:
  Akohekohe has 7 iNaturalist records and *zero* parseable dates. Recurrence is unmeasurable for
  precisely the species eBird withholds. Spotted Owl is better but thin — 517 records on 22
  dates.
- **Origin is not an occurrence property, and recurrence does not rescue it.** This is the one
  bias recurrence does *not* fix, and the example is already shipped: **Indian Peafowl in New
  York** — 255 records on 215 distinct dates, 47 episodes across 33 years. Every recurrence
  measure reads *established*, because ornamental birds on estates and farms are continuously
  replenished by people, and a maintained population is indistinguishable from a wild one in
  occurrence data. Compare the genuine feral population in Florida: 23,029 records, 11 years.
  Recurrence is robust to **observer** bias and blind to **origin** bias. If BirdView labels a
  species `exotic`, that label is a declared editorial list, not a measurement, and must say so.

These three need a short declared exceptions table, owned and documented. That is a far smaller
editorial surface than the weighted model this CR was asked for.

### Measured 2026-10-04, and one defect the run exposed

The first complete run classified **1,424 species**: 956 established, 131 rare, 131 vagrant,
38 historical, 173 absent. Against BirdView's current 839 it **keeps 830 and drops 9**, where
the best abundance rule dropped 59. The nine are the declared-exception cases of
[product.declaredExceptions](local/product/dataset.md#declaredexceptions) and nothing else: three with no open record at all (Maui Parrotbill,
Akohekohe, Townsend's Storm-Petrel), two outside the US/CA filter (Bermuda and Desertas
Petrel), and four single-episode seabirds and Hawaiian endemics.

**`historical` earns its place.** It caught 38, headed by Passenger Pigeon — 89 episodes across
52 years, every record between 1802 and 1902 — with Carolina Parakeet, Ivory-billed Woodpecker,
Eskimo Curlew, Bachman's Warbler and three extinct Hawaiian endemics behind it.

**The scope bar is not sensitive where it matters.** From 2/2 to 6/6 the number of currently
listed species kept stays between 826 and 831. What the bar moves is how many *new* species come
in — 300 at 2/2, 183 at 6/6 — so it is a dial on inclusiveness, not on whether real birds
survive. There is no natural break to anchor it to: 61% of species truncate the date facet and
are established by inspection, and the remaining 491 shade smoothly from 1 to 70 episodes.

**The defect: the lumps are not corrected, so three species are counted twice.** The generator
does not apply the correction this CR records above, and the run admitted both halves of all
three pairs as separate species:

| eBird publishes | records | BirdView carries | records |
| --- | ---: | --- | ---: |
| `Numenius phaeopus` | 25,946 | `Numenius hudsonicus` | 169 |
| `Gygis alba` | 2,481 | `Gygis candida` | 22 |
| `Setophaga petechia` | 243,939 | `Setophaga aestiva` | 116 |

A table written from this would list Whimbrel twice. **Resolved 2026-10-04** on two stated
principles: extra species beat omission, and eBird is the more authoritative source.

Those settle it once separated. eBird is authoritative for **evidence** — it is eighty times
deeper and better curated — while AviList is already the authority for **names** (avilistReplacesAba). So
eBird's records are credited to the AviList name whose birds they are, which is what
`AviListRenames` does for genus moves; `taxonomy.CountedAs` now does it where the sources draw the boundary elsewhere.

"Extra beats omission" holds everywhere except where the extra is the **same drawing twice**.
`Numenius phaeopus` and `Gygis alba` are the AVONET donors for the halves BirdView carries, so
admitting them adds two entries with byte-identical length and mass — no information, and it
reads as a bug. They are merged away. `Setophaga petechia` is **not**: Mangrove Yellow Warbler
is carried with an AVONET row of its own, and both halves clear the bar on their own records, so
the misattribution changes no outcome and is recorded rather than guessed at. Trogon is the same
shape.

**Merging the counts was not enough**, and the first attempt did only that. It changed which
regions were queried and left pass 2 asking GBIF for `numenius hudsonicus`, a name eBird holds
**zero** records under, so the merged species came back with iNaturalist's 169 and none of
eBird's half-million. The dates exist only under the lump and must be fetched under it. With
both names queried, Whimbrel reads 82 episodes across 40 years on 25,961 records where it read
37 and 169.

### Settled on approval

**The vocabulary** is `established`, `rare`, `vagrant` and `absent`, with `withheld` and
`exotic` as orthogonal flags rather than statuses — a bird can be withheld *and* established,
and the peafowl is exotic *and* recurring, so folding either into the status would lose one of
the two facts.

**The run is ~1,400 queries, not ~96,000.** Querying every species against every region was the
wrong shape. Scope asks whether a species recurs *somewhere*, so the cheap form is two passes:

1. One `verbatimScientificName` facet per region — 64 queries — which yields the record count
   for every (species, region) pair at once. This is what
   [`birdview-regions`](../../cmd/birdview-regions) already does.
2. An `eventDate` facet for each species in its **top three regions by record count** — about
   4,100 queries. If a species does not recur where it is seen most, it does not recur anywhere.

**The established/rare boundary is set after the run, not before.** Picking it now would be
inventing a coefficient and then finding data to justify it; the honest order is to measure the
distribution first and then declare the cut with a written reason. This is the one piece of the
model still owed a number.

## speciesUrlOutlivesSpecies

**A species URL outlives the species, but does not keep its 200**

Date: 2026-10-05
Decided by: Sameer
Status: accepted
Amends: compareUrlContract
Affects: [product.undrawnSpeciesAnswer](local/product/presentation.md#undrawnspeciesanswer), [product.scopeIsRecurringAvifauna](local/product/dataset.md#scopeisrecurringavifauna)
Subject: `presentation.unknown_species_id`
Kind: Conditional — [compareUrlContract](#compareurlcontract)(b) set this subject for `/compare`, and the detail view resolves it the same way in substance and the opposite way in status code

### The question compareUrlContract left open

compareUrlContract(b) decided that an unrecognised species id in `/compare` is **dropped and named**, not a
404: *"render what is recognised and name what is not."* `/species/<id>` was never revisited, and
it still calls `http.NotFound`, which writes `404 page not found` and nothing else.

That was harmless while scope came from a checklist nobody intended to change. It stopped being
harmless with [scopeFromRecurrence](#scopefromrecurrence): scope is now **derived from live occurrence data and
regenerated**, so the set of species URLs that work is a function of a GBIF query, and a future
run can take `/species/<code>` away from a bird that has one today.

### Why it is not simply compareUrlContract again

compareUrlContract's answer has two halves — *render what you can* and *name what you cannot* — and only
the second half survives the move. A comparison of four birds minus one is still a comparison
of three; a detail page minus its one bird is nothing. There is no partial page to render.

That difference decides the status code, which is the only point on which this record departs
from compareUrlContract. `/compare` returns **200** because it delivered a comparison and merely delivered
a smaller one than was asked for. `/species/<id>` has delivered **no bird at all**, and
returning 200 would tell every cache and crawler that an address which serves nothing is fine.

### Options considered

| | Status | What the reader sees | Why not |
| --- | --- | --- | --- |
| Keep `http.NotFound` | 404 | `404 page not found` | Throws away a measured verdict the derivation already holds. This is the status quo being changed |
| Redirect to the index | 302 | The index, no explanation | Loses the question. The reader asked about one bird and is handed 1,083 |
| **Named 404** | **404** | The bird, its status, its evidence | **Chosen** |
| Named 410 Gone | 410 | The same page | Overclaims. 410 asserts the resource is permanently gone and should be de-indexed; scope is regenerated, and a bird that fails the recurrence bar on one run can clear it on the next. Three of the nine [product.declaredExceptions](local/product/dataset.md#declaredexceptions) exceptions exist precisely because sparse evidence moves |
| Serve the bird anyway, marked out of scope | 200 | A drawing | Rejected outright. [product.scopeIsRecurringAvifauna](local/product/dataset.md#scopeisrecurringavifauna) decides what BirdView draws, and a page that draws a bird the dataset excludes makes the scope rule a suggestion |

### Decision

[product.undrawnSpeciesAnswer](local/product/presentation.md#undrawnspeciesanswer): **404 with a page that names the bird and gives the measured reason.**
`410` is rejected for overclaiming, `200` for lying to caches, and a redirect for losing the
question.

**An id matching no classified species keeps the bare 404.** The named page requires a row in
the status table; without one BirdView has nothing to say, and inventing a bird to apologise
for would be worse than the terse answer.

### What it costs

The out-of-scope species have to reach the server, which means the ingest carries them into
`species.json` — 328 entries of id, names, status and evidence, 58 KB against a 4.3 MB file.
They are not
`Species` records: they have no measurements, no family colors and no diagram, and giving them
a half-filled `Species` would put 328 birds with zero mass inside the type every drawing
function consumes.

It also costs a **tenth column in `data/status.tsv`**, the date each species was last recorded.
The table carried episodes, years and records but not *when*, so "historical" could be stated
and not shown. For the 37 historical species that date is the whole explanation, and a verdict
whose evidence is withheld is the shape this project keeps rejecting elsewhere.

### What regenerating the table found

**`data/status.tsv` was not a function of its inputs.** Rebuilding it from the cached fetches —
same records, same dates, no network — produced a table that differed from the committed one,
and differed again on the next run. `bestRegion` ranked candidate regions on status and episode
count and then walked a Go map, so any tie was broken by iteration order. **Fifteen rows named a
different region on each run** and carried their episode, year, record and now last-seen
figures with them: Cooper's Hawk recurs in 39 episodes in both Alaska and Arizona, and the
table reported 24 years and 216,391 records one run and 18 and 119,194 the next.

Harmless while nothing read the region column; not harmless once [product.undrawnSpeciesAnswer](local/product/presentation.md#undrawnspeciesanswer)
prints the date on a page as the reason a bird is not drawn. A reason that changes when nothing
changed is not a reason.

The tiebreak is now declared rather than incidental: status, then episodes, then **years, then
records, then the region id**. Years before records because the model is recurrence and not
abundance ([scopeFromRecurrence](#scopefromrecurrence)) — between two regions a bird recurs in equally often, the one it
came back to across more years is the better evidence, whatever the crowds did. The id settles
what nothing else can: arbitrary, but the same arbitrary answer every run. Fifteen rows moved,
most of them onto strictly more evidence; **no species changed status, and scope is unchanged at
1,083.**

### Status of requirements touched

- [product.scopeIsRecurringAvifauna](local/product/dataset.md#scopeisrecurringavifauna) — Active; amended to name `historical` in the out-of-scope list,
  which it had omitted, and to point at [product.undrawnSpeciesAnswer](local/product/presentation.md#undrawnspeciesanswer).
- [product.undrawnSpeciesAnswer](local/product/presentation.md#undrawnspeciesanswer) — Active, new.
- [compareUrlContract](#compareurlcontract)(b) — stands unchanged for `/compare`.

## seaShareDiffers

**The share at sea is not the share on land, and [product.shareOfRecords](local/product/presentation.md#shareofrecords) presents them as one number**

Date: 2026-10-05
Decided by: Sameer
Status: amended by honestSeaDenominator
Affects: [product.shareOfRecords](local/product/presentation.md#shareofrecords), [product.regionAndMonthFilter](local/product/filtering.md#regionandmonthfilter)
Subject: `presentation.share`, `dataset.region.totals`
Kind: Derived / emergent — [product.shareOfRecords](local/product/presentation.md#shareofrecords) and [product.regionAndMonthFilter](local/product/filtering.md#regionandmonthfilter) are each satisfiable alone; the defect is in a ratio computed across both
Found by: the owner, reading `/species/yenalb2?region=NorthAtlantic` against `/?region=NorthAtlantic` and noticing the two pages disagreed about where the bird is commonest

### What a reader sees

| page | says |
| --- | --- |
| `/?region=NorthAtlantic` | Atlantic Yellow-nosed Albatross, **1 in 4,521** bird records in North Atlantic |
| `/species/yenalb2` | Commonest in **Maine — 1 in 1,177,757** bird records there |

The obvious reading is that the detail page is wrong. It is not. **The index figure is the
wrong one**, and the detail page's answer survives a correct recomputation.

### Three defects, stacked

**(a) The peak-share region is computed over land only.** `birdview-status` walks
`region.LandIDs()` ([main.go:92](../../cmd/birdview-status/main.go)) and has no marine pass;
`birdview-regions` walks land *and then* boxes the five seas
([main.go:211](../../cmd/birdview-regions/main.go)). So `Species.Regions`, `RegionCounts` and
`RegionTotals` cover 69 land + 5 marine, while `Species.Commonest` covers 69 land.
**Zero of 1,426 species name a marine region as commonest**, and none can.

**(b) A marine region's total counts a subset of species, a land region's counts all of them.**
`birdview-regions` only boxes the seas for species already judged **pelagic** — at least
`minPelagicShare` = 0.5 of their records outside every land region — and accumulates
`totals[id]` over just those. A land total is every bird record in the region. The two are
different quantities, and [product.shareOfRecords](local/product/presentation.md#shareofrecords) divides by both as though they were one. Measured against GBIF
on 2026-10-05:

| sea | dataset total | all bird records in the same box | inflation |
| --- | --- | --- | --- |
| Arctic Ocean | 50,758 | 758,849 | 15× |
| Central Pacific | 19,586 | 2,638,631 | 135× |
| North Pacific | 1,596,171 | 233,491,816 | 146× |
| North Atlantic | 429,535 | 413,188,208 | **962×** |
| Gulf of Mexico | 24,897 | 111,327,757 | **4,472×** |

Every marine share the index has ever shown is inflated by its row above. The albatross's
`1 in 4,521` is honestly `1 in 4,349,349`, which is **worse** than Maine's `1 in 1,177,757` —
so Maine really is its peak.

**(c) The marine regions are lat/lon rectangles that contain a great deal of land.** The
Atlantic box is 10–66°N, 80–40°W: 413 million bird records, most of them on the eastern
seaboard. Even with the denominator of (b) fixed, "share of the North Atlantic" would partly
mean "share of Nova Scotia". The boxes work for the job they were built for — deciding *which
ocean* a bird already known to be pelagic belongs to — and do not support a share.

### Correcting the first analysis in this session

An earlier pass reported **54 of 55** species with a wrong peak, with errors up to 21,349×, and
called the albatross 710× wrong. Those figures divided a marine count by the pelagic-only
denominator of (b) and compared the result against a land share. They are not valid and are
withdrawn. Recomputed against all bird records in the box, **30 of 55** species have a peak that
genuinely moves to sea — among them Black-capped Petrel (`1 in 31,155` in the Atlantic against
the `1 in 626,418` in South Carolina now shown) and Townsend's Storm-Petrel (`1 in 53,837` in
the Pacific against `1 in 116,973,739`). Defect (a) is real and worth fixing; its size was
overstated by a factor it was itself an instance of.

### Why nothing caught it

**No test asserts that `Commonest` is the largest share.** The server fixture hardcodes
`commonestOf` and every test renders it back, so [product.shareOfRecords](local/product/presentation.md#shareofrecords)'s central claim — *the region where this
bird is the biggest share of what is recorded* — has never been evaluated against data. A value
the fixture supplies is a value no test can falsify.

### Options

The choice is (c); it decides (a) and (b).

| | What a marine region offers | Cost | Against |
| --- | --- | --- | --- |
| **A. Presence at sea, no share** | the region filter and "recorded in the North Atlantic"; no `1 in N`, and `Commonest` is explicitly the peak **land** region | 5 lines and a wording change; no re-run | Loses a real answer for 30 pelagics — the Black-capped Petrel's home genuinely is the Atlantic |
| **B. Honest denominator, box as declared** | a share comparable to a land share, over the box as defined | 5 facet requests for the box totals, a regenerated `region-presence.tsv` and `status.tsv`, and a marine `eventDate` facet that does not exist yet | "North Atlantic" keeps meaning a rectangle including Nova Scotia; honest arithmetic over a region whose name oversells it |
| **C. Leave it** | — | none | The index is actively wrong by 15–4,472× today |

C is rejected. **A is recommended**, on the grounds that B buys comparable arithmetic over a
region that is not really a region, and that [product.shareOfRecords](local/product/presentation.md#shareofrecords) already refuses to publish a number it
cannot stand behind. A pelagic's detail page can still say *recorded in the North Atlantic*
without claiming a ratio.

### Decision — A, and no commonest sea

**A.** The share is over land regions, and the pages say so. `RegionTotals` stops carrying
marine entries at all, so no caller can divide by one, and `shareIn` skips any region the
totals have no entry for — which keeps the numerator and the denominator over one set of
regions **by construction** rather than by two matching conditions that can drift apart. The
detail page's row is relabelled **Commonest land region**.

The cost is accepted and should be stated: a pelagic's index share is now computed from its
land records alone, so a Black-capped Petrel — 13,262 of its 13,601 records inside the Atlantic
box — reads as the vanishingly rare land bird it is. That is true as far as it goes, and this
project prefers a true narrow number to a wide one it cannot defend.

**No commonest sea.** This was asked and is rejected on a measurement. The boxes **overlap**
the land regions rather than abutting them: all 6 of the Black-capped Petrel's North Carolina
records fall inside the Atlantic box, confirmed against GBIF on 2026-10-05. A coastal record is
a marine record. For a true pelagic the contamination is negligible — 6 in 13,262 here — but
the Pacific box spans −150° to −115° and so contains most of coastal California, and the
Atlantic box the entire eastern seaboard. For anything short of a true pelagic, "commonest
sea" would name the coast the bird is watched from, which the land regions already say without
ambiguity.

The tempting middle — rank a bird's own marine records across the five seas, which needs no
cross-species denominator and so escapes defect (b) — fails on the same overlap. The boxes are
mutually disjoint, so sea-versus-sea ranking is safe from *each other*; it is not safe from the
coasts inside them, and a bird common on one shore would be ranked into the ocean beside it.
The metric would be sound only for a subset a reader cannot identify from the page.

**39 of the 55 species with marine records occur in two or more seas**, so this does decline a
real answer rather than a redundant one. It is declined anyway, because the answer would be
right for the 25 species that are ≥90% at sea and quietly wrong for the 29 that are not.

### Not yet established

Whether **scope** is affected. The recurrence test runs on `gadmGid`, and `DatesIn` has no
coordinate-box form, so pass 2 has never asked whether a pelagic recurs at sea. Four of the
nine [product.declaredExceptions](local/product/dataset.md#declaredexceptions) exceptions are pelagics whose written reason is that the
derivation "sees [it] in one episode because almost nobody is on the water where it is". All
four are 95–100% at sea with real marine records. The reason as written states a measurement
that was never made; **at minimum the wording is wrong**, and whether those species would clear
the bar on their own is unknown until a box form of the date facet exists. Tracked below.

### Experiment — is there a distance from shore that cleans up the boxes?

*Run 2026-10-05 against live GBIF, at the owner's request, to see whether the seas could be
tightened enough to carry a share after all.* **They can, and the instrument is not distance.**

**Distance from shore is not expressible in GBIF, and does not need to be.** A record on land
carries a GADM country; a record on water carries none. `facet=gadmLevel0Gid` therefore splits
any query into on-land and off-land, and the split is exactly the coastline — sharper than any
buffer we could declare, and it needs no coastline dataset. It is also **free**: the facet
response carries the unfiltered `count` alongside the facet rows, so one request yields both
halves. Adopting this costs **zero extra requests**.

**There is no gradient to tune.** Longitude bands at 35–40°N, walking east from the coast:

| band | km out | records | off land | top species |
| --- | --- | --- | --- | --- |
| −75.5 to −75 | 0–44 | 13,135,500 | 6.0% | Northern Cardinal, American Robin, Canada Goose |
| −75 to −74 | 44–133 | 13,707,437 | 4.9% | American Herring Gull, Northern Cardinal, Mourning Dove |
| **−74 to −73** | **133–222** | **42,227** | **100.0%** | **Wilson's Storm-Petrel, Great Shearwater, Audubon's Shearwater** |
| −73 to −71 | 222–400 | 33,047 | 100.0% | the same four tubenoses |
| −71 to −67 | 400–756 | 12,921 | 100.0% | the same |
| −67 to −60 | 756–1,378 | 154 | 100.0% | the same |

The bands measure longitude, not true distance — the 44–133 band still holds New York City,
which is why it is 4.9% off land. That is the finding: composition does not taper with
distance, it **flips between the band that contains coastline and the band that does not**.
A declared distance threshold would have nothing to interpolate between.

**The honest denominator, per sea:**

| sea | box | on land | off land | today's figure |
| --- | --- | --- | --- | --- |
| North Atlantic | 413,188,208 | 400,201,694 | **12,986,514** (3.1%) | 429,535 |
| North Pacific | 233,491,816 | 222,883,645 | **10,608,171** (4.5%) | 1,596,171 |
| Gulf of Mexico | 111,327,757 | 107,876,458 | **3,451,299** (3.1%) | 24,897 |

So the correction is **about 30×**, not the 962× quoted earlier against the whole box — the
whole box was never the right denominator either.

**What the rule reattributes**, as a share of each species' Atlantic-box records that sit on
land: Audubon's Shearwater 0.4%, Black-capped Petrel 1.0%, Cory's Shearwater 11.6%, Wilson's
Storm-Petrel 14.7%, Great Shearwater 20.2%, **Atlantic Puffin 53.2%, Northern Gannet 66.8%**.
The spread looks alarming and is not a loss: a gannet seen from a Massachusetts beach belongs
to Massachusetts, and keeps that record there. Even the gannet retains 142,000 records at sea.

**The residual is estuaries, not shore.** Water inside a bay has no GADM country either —
Chesapeake Bay reads 8.8% off land, Pamlico Sound 8.5%, Long Island Sound 4.1%. Those three
contribute on the order of 400,000 records against the Atlantic's 12.99 million, so the rule
delivers *on water* rather than *at sea*, overstating by roughly 3%. GBIF does accept a WKT
`geometry` polygon and combines it with facets (verified), so carving the open ocean by hand
is available later — but it chases a 3% effect against a 30× one, and the boxes are declared
approximations already.

**Consequence: [product.shareOfRecords](local/product/presentation.md#shareofrecords) could get its share back at sea.** "Of the records on
water in this box, how much is this bird" is a real quantity, comparable in kind to the land
figure and computable for free. That would reverse the part of option A that declines a marine
share — and would need a CR of its own, a regenerated `region-presence.tsv`, and a decision on
whether "on water" may be labelled with an ocean's name. Tracked below; nothing is changed on
the strength of this experiment alone.

## honestSeaDenominator

**A sea gets an honest denominator, and its share back**

Date: 2026-10-05
Decided by: Sameer
Status: accepted
Amends: seaShareDiffers
Affects: [product.shareOfRecords](local/product/presentation.md#shareofrecords)
Subject: `presentation.share`, `dataset.region.totals`
Kind: Reversal — [seaShareDiffers](#seasharediffers)'s option A declined a marine share; the measurement it declined for turns out to be available

### What changed

[seaShareDiffers](#seasharediffers) withheld the share at sea because no honest denominator existed for a sea.
Its own [experiment](#experiment--is-there-a-distance-from-shore-that-cleans-up-the-boxes),
run immediately afterwards, found one. GADM covers land, so **a record on a beach carries a
country and a record on the water carries none**: faceting a box by `gadmLevel0Gid` and
subtracting gives the off-land half exactly, with the coastline as the boundary. The facet
response carries the unfiltered count beside the facet rows, so it costs **no extra requests**.

A sea's total is now every bird record on open water inside its box, over all species — the
same quantity a land region's total is, measured the same way. `RegionTotals` carries all 74
regions again, and [product.shareOfRecords](local/product/presentation.md#shareofrecords)'s share ranges over every one.

| sea | was | now |
| --- | --- | --- |
| North Atlantic | 429,535 | **12,986,514** |
| North Pacific | 1,596,171 | **10,608,171** |
| Gulf of Mexico | 24,897 | **3,451,299** |
| Central Pacific | 19,586 | **127,341** |
| Arctic Ocean | 50,758 | **126,158** |

### The second half: where a bird is commonest now comes from one table

The peak region used to come from `data/status.tsv` and the index share from
`data/region-presence.tsv`. Two tables, two denominators, two membership rules — which is the
whole of the bug the owner reported. `/species/yenalb2` said *commonest in Maine* while
`/?region=NorthAtlantic` said something that read as a flat contradiction, and **nothing
compared them**, each being tested against its own source.

Worse, the status table's land-only pass meant no species could ever peak at sea, and its
peak could name a region the index does not list the bird in at all: Maine holds 12 albatross
records against 14 million, below [product.regionAndMonthFilter](local/product/filtering.md#regionandmonthfilter)'s one-per-million presence bar, so
the page named a place a reader could not then find in the filter.

`Commonest` is now computed in the ingest from `Regions`, `RegionCounts` and `RegionTotals` —
the same numbers the index divides. The two pages are the same arithmetic by construction, and
[acceptance.peakAndShareOneCalculation](acceptance.md#peakandshareonecalculation) asserts the round trip over all 1,059 placed species.
`status.CommonestBy` is deleted.

**The Atlantic Yellow-nosed Albatross now reads 1 in 282,315 in the North Atlantic**, which is
where the owner expected it, arrived at honestly rather than by the inflated figure that
originally seemed to show it. **45 species peak at sea.**

### What this does not fix

The result is **on water**, not **at sea**. An estuary has no GADM country either, so
Chesapeake Bay reads 8.8% off land, Pamlico Sound 8.5%, Long Island Sound 4.1% — roughly a 3%
overstatement of the open ocean against a 30× error corrected. GBIF accepts a WKT `geometry`
polygon alongside facets (verified), so carving the coast off properly is available if that 3%
ever matters. The boxes are declared approximations already and [product.shareOfRecords](local/product/presentation.md#shareofrecords) says
what the figure counts.

**"No land regions selected" is retired**, one turn after being added. Its condition no longer
occurs: every region has a denominator, so selecting seas yields a figure like anywhere else.
The branch survives with generic wording — *No share for this selection* — because a region
added to `internal/region` ahead of its data lands there, which is precisely how the marine
regions behaved for months.

**The pelagic set is frozen between full runs.** `birdview-regions -marine-only` refreshes the
sea rows for the species the table already places at sea, at about 300 requests against the
full run's several thousand, and leaves every land row byte-identical — so the diff of a change
to the marine rule is the marine rule. A species newly become pelagic waits for a full run.
That is a deliberate trade: re-measuring it costs a `CountFor` for every species in the dataset.

### Status of requirements touched

- [product.shareOfRecords](local/product/presentation.md#shareofrecords) — Active; amended, the share ranges over every region again.
- [seaShareDiffers](#seasharediffers) — option A's *no share at sea* half is superseded here. Its structural
  guard is not: `shareIn` still takes both sides of the ratio from one test, and
  [acceptance.commonestIsLargestShare](acceptance.md#commonestislargestshare) still asserts it. What was wrong there was the number, and
  what was right was the shape.

## frequencyMetric

**Frequency joins the metrics, and brings a second sectioning rule with it**

Date: 2026-10-05
Decided by: Sameer
Status: amended by frequencyGlobalScale
Affects: [product.oneSelectedMetric](local/product/ordering.md#oneselectedmetric), [product.sectionBreakRule](local/product/ordering.md#sectionbreakrule), [product.sectionThresholdDefault](local/product/ordering.md#sectionthresholddefault), [product.orderByFrequency](local/product/presentation.md#orderbyfrequency)
Subject: `ordering.metric`, `sectioning.rule`, `sectioning.step_percent`
Kind: Direct — a fourth ordering does not fit the value [product.oneSelectedMetric](local/product/ordering.md#oneselectedmetric) fixes, and cannot be sectioned by the rule [product.sectionBreakRule](local/product/ordering.md#sectionbreakrule) fixes

### What changed

[product.shareOfRecords](local/product/presentation.md#shareofrecords) put a number on every tile — *1 in 282,315 bird records in North
Atlantic* — and the index could not order by it. It can now: **Frequency** is a fourth choice
under Sort by.

Three subjects move:

- `ordering.metric` gains `frequency`.
- `sectioning.rule` gains a **second** rule, used only by frequency.
- `sectioning.step_percent` means something else under frequency, and the control shows it.

### Why the proportional rule cannot section frequency

[product.sectionBreakRule](local/product/ordering.md#sectionbreakrule)'s rule is already multiplicative — a section ends at `min × (1 + N/100)`
— so on the face of it 15% should work. It does not, for two reasons measured on the data.

**It produces 128 sections.** Frequency spans **1 in 43 to 1 in 1,252,475,518**, which is 7.5
decades. At 15% that is `log(1.25e9/43) / log(1.15)` ≈ 128 bands of a few hundred species each.
Length at 15% produces a page a reader can scroll.

**Its boundaries move when the filter does.** [product.sectionBreakRule](local/product/ordering.md#sectionbreakrule) anchors each band to the data — the first
section starts at the smallest bird present. Frequency follows the region filter
([product.shareOfRecords](local/product/presentation.md#shareofrecords)), so a data-anchored band would start at 1 in 43 nationally and
somewhere else entirely in Alaska, and *the same tier would mean a different thing on two
pages of the same site.* Size does not have this problem because a Gyrfalcon is the same
length wherever you select.

### The rule for frequency: decades, anchored at powers of ten

A section runs from `1 in 10^k` to `1 in 10^(k+1) − 1`. Measured over the whole dataset that
gives **nine sections**, and a distribution that is close to bell-shaped:

| tier | species |
| --- | --- |
| 1 in 10 – 99 | 23 |
| 1 in 100 – 999 | 174 |
| 1 in 1,000 – 9,999 | **267** |
| 1 in 10,000 – 99,999 | 223 |
| 1 in 100,000 – 999,999 | 163 |
| 1 in 1,000,000 – 9,999,999 | 125 |
| 1 in 10,000,000 – 99,999,999 | 74 |
| 1 in 100,000,000 – 999,999,999 | 9 |
| 1 in 1,000,000,000 and beyond | 1 |

Powers of ten rather than quantiles or a fitted break, because the tier boundary is then **the
same fact in every view**: 1 in 1,000–9,999 means that in Alaska and nationally alike. It also
matches the form the number is already printed in, so a reader can place a tile in its tier by
reading it.

**Section step becomes a band width**: `10×` (nine sections, the default) or `100×` (five). The
percentages are hidden under Frequency because they would be lies — 10× is +900%, and a
control reading "900%" explains nothing.

### What is deliberately NOT done

**No tier is given a name.** Not "Common", not "Scarce". [product.shareOfRecords](local/product/presentation.md#shareofrecords) says the
metric "is not a rarity tag and must not become one — no threshold divides common from rare;
the number is shown and the reader judges", and a named band is exactly that threshold wearing
a label. The headers state the range, as [product.sectionHeadersStateRange](local/product/ordering.md#sectionheadersstaterange) already requires of every
other metric.

**Species with no records in the selection are kept, in a final section of their own**, headed
*No records in the regions selected*. They cannot be placed in a tier — there is no ratio —
and dropping them would silently shorten the page by 24 birds nationally and more under a
filter.

### A known consequence — and this got it wrong

Under the size metrics a section holds birds of similar size, so
[product.scaleToggle](local/product/diagram.md#scaletoggle)'s per-section scale is both honest and legible. Frequency does not
correlate with size, so a frequency tier can hold a hummingbird and a swan.

*This record originally concluded that was "no worse than Global, which draws the hummingbird
just as small". That is false, and [frequencyGlobalScale](#frequencyglobalscale) corrects it with the measurements:
per-section rescales each tier by its own largest bird, which across the nine tiers swings by
0.15× to 1.36×, so a 140 mm parrotbill is drawn the height of a 1,201 mm swan six tiers up.
Global is the default under Frequency now.*

### Status of requirements touched

- [product.oneSelectedMetric](local/product/ordering.md#oneselectedmetric) — Active; value gains `frequency`.
- [product.sectionBreakRule](local/product/ordering.md#sectionbreakrule) — Active; scoped to the size metrics, pointing at [product.orderByFrequency](local/product/presentation.md#orderbyfrequency).
- [product.sectionThresholdDefault](local/product/ordering.md#sectionthresholddefault) — Active; scoped likewise.
- [product.orderByFrequency](local/product/presentation.md#orderbyfrequency) — Active, new.

## frequencyGlobalScale

**Frequency defaults to global scale, because per-section has nothing to stand on there**

Date: 2026-10-05
Decided by: Sameer
Status: accepted
Amends: frequencyMetric, perSectionDefault
Affects: [product.scaleToggle](local/product/diagram.md#scaletoggle)
Subject: `diagram.scale.mode`
Kind: Conditional — [product.scaleToggle](local/product/diagram.md#scaletoggle)'s default is right for the three size metrics and unsupported under the fourth

### What [frequencyMetric](#frequencymetric) got wrong

That record said per-section scaling under Frequency was "the per-section scale working as
specified rather than a new defect, and it is no worse than Global". **The second half is
false**, and the owner found it by preferring Global without being told to.

[product.sectionBreakRule](local/product/ordering.md#sectionbreakrule) justifies per-section scaling in one sentence: *"because a section is
by construction no more than N% wide, [it] makes the per-section scale of [product.scaleToggle](local/product/diagram.md#scaletoggle) honest."* That
guarantee is the entire basis for rescaling between sections, and **a frequency tier does not
have it** — tiers are banded by how often a bird is recorded, which is uncorrelated with size,
so one tier spans 72 mm to 1,201 mm.

### Measured

Each tier is drawn against its own largest bird. Across the nine tiers that reference is:

| tier | largest bird | scale vs. previous tier |
| --- | --- | --- |
| 1 in 10 – 99 | Great Blue Heron, 882 mm | — |
| 1 in 100 – 999 | American White Pelican, 1,198 mm | 1.36× |
| 1 in 1,000 – 9,999 | Trumpeter Swan, 1,201 mm | 1.00× |
| 1 in 10,000 – 99,999 | Whooping Crane, 1,134 mm | 0.94× |
| 1 in 100,000 – 999,999 | Whooper Swan, 1,135 mm | 1.00× |
| 1 in 1,000,000 – 9,999,999 | Great White Pelican, 1,400 mm | 1.23× |
| 1 in 10,000,000 – 99,999,999 | Emu, 1,789 mm | 1.28× |
| 1 in 100,000,000 – 999,999,999 | Scarlet Macaw, 912 mm | **0.51×** |
| 1 in 1,000,000,000 and beyond | Maui Parrotbill, 140 mm | **0.15×** |

The last tier holds one bird, a **140 mm** parrotbill, drawn to fill the box — the same height
on screen as the 1,201 mm Trumpeter Swan six tiers up. **An 8.6× lie between two sections of
one page**, and precisely the confusion [product.scaleToggle](local/product/diagram.md#scaletoggle) exists to prevent.

### Decision

**Global is the default when Order by is Frequency.** Per-section remains the default for
length, wingspan and mass, where a section really is at most N% wide and the rescaling is
honest. Either choice stays one click away under both.

The reader loses legibility, which is what [perSectionDefault](#persectiondefault) moved the default *to* per-section
for. That cost is accepted here because the alternative is not a less legible drawing but a
wrong one: under Frequency, per-section scaling answers "how big is this bird" differently in
adjacent tiers for no reason the reader can see.

**Both scale values are now recognised explicitly in parsing.** With a default that varies by
metric, reading only the non-default would leave *Per section* a dead button under Frequency —
the exact trap the comment on that code already warned about, written when perSectionDefault moved the
default the other way.

### Status of requirements touched

- [product.scaleToggle](local/product/diagram.md#scaletoggle) — Active; the default is per-section for the size metrics and
  global for frequency.
- [frequencyMetric](#frequencymetric) — its "no worse than Global" aside is withdrawn, above.

## sharedCountShownOnBoth

**A count that covers two species is shown on both, and says so**

Date: 2026-10-05
Decided by: Sameer
Status: accepted
Affects: [product.namesRecordsFiledUnder](local/product/presentation.md#namesrecordsfiledunder), [product.shareOfRecords](local/product/presentation.md#shareofrecords)
Subject: `presentation.share`, `dataset.name_provenance`
Kind: Derived / emergent — no requirement is violated alone; the defect is in what a tile means when the occurrence sources and AviList disagree about what a name denotes
Found by: the owner, asking why `manwar1` outranked `yelwar1` in New York

### The rejected fix, and why

The first proposal was to reassign eBird's 6,245,227 `Setophaga petechia` records to
*Setophaga aestiva*, on the grounds that nearly every Yellow Warbler in the US and Canada is
the northern bird. **The owner stopped it.** That is true and it is also a division nobody has
measured: eBird has not adopted AviList's split and distinguishes none of those records.
Attributing them all to one daughter replaces a wrong answer with a confident guess, which is
the trade this project refuses everywhere else.

### The sweep

Every one of the 1,083 drawn species was queried against both databases under the AviList
binomial BirdView serves (2,166 requests, 2026-10-05). **18 have zero eBird records under that
name**, and they fall into two groups that look identical in the data and mean opposite things:

**Name disagreement — the records exist under another name.** Nine species, all but one
already covered by `AviListRenames` or `CountedAs`:

| BirdView draws | eBird files it as | eBird records |
| --- | --- | --- |
| *Leuconotopicus villosus* (Hairy Woodpecker) | *Dryobates villosus* | 7,310,342 |
| *Setophaga aestiva* (N. Yellow Warbler) | *Setophaga petechia* | 6,245,227 |
| *Icterus bullockiorum* (Bullock's Oriole) | *Icterus bullockii* | 881,936 |
| *Hesperiphona vespertina* (Evening Grosbeak) | *Coccothraustes vespertinus* | 838,952 |
| *Numenius hudsonicus* (Whimbrel) | *Numenius phaeopus* | 556,578 |
| *Pica nuttallii* (Yellow-billed Magpie) | *Pica nuttalli* | 129,969 |
| **_Trogon ambiguus_ (Coppery-tailed Trogon)** | ***Trogon elegans*** | **34,913** |
| *Leuconotopicus albolarvatus*, *borealis*, *arizonae* | *Dryobates* | 3,576 / 2,007 / 1,545 |

**eBird withholds the species.** Nine more — Spotted Owl, Gyrfalcon, Northern Hawk Owl, Lesser
Prairie-Chicken, Gunnison Sage-Grouse, Puaiohi, Maui Parrotbill, Akohekohe, Desertas Petrel.
Confirmed by querying every plausible alias: *Strix occidentalis* and *Falco rusticolus* return
zero under any name. Their records are suppressed, not renamed, and
[product.namesRecordsFiledUnder](local/product/presentation.md#namesrecordsfiledunder) must not say otherwise.

**Two of the name disagreements are SHARED names**, where one eBird binomial covers two species
BirdView draws: *Setophaga petechia* (Northern + Mangrove Yellow Warbler) and *Trogon elegans*
(Elegant + Coppery-tailed Trogon). The Trogon case was guessed at in a code comment and is now
measured.

**One apparent finding was not one.** *Sternula albifrons* showed zero in both databases, but
its 106 records are on Midway, whose country code is `UM` — excluded by the sweep's own
`country=US,CA` filter. A blind spot in the measurement, not in the data.

### Decision

**A count covering two species is shown on both, labelled.** Not reassigned to one, not
withheld from both. The tile prints the figure it has and names the species it is shared with;
the detail page says which source does not separate them and how many records are involved.

The alternative considered was showing no figure on either, on the grounds that a number we
cannot attribute is a number we cannot stand behind. **[goals.honestData](../goals/goals.md#honestdata) decides it the other way**: suppressing the figure hides
6.2 million records and one of the commonest warblers in North America, and a reader has no way
to tell that view from a bird genuinely absent. Showing the count with its seam named tells the
reader strictly more, and nothing it tells them is false.

### What this does not do

It does not make the Frequency ordering right for Mangrove Yellow Warbler, which will still
sit in a tier its own records do not earn. That is the honest consequence of eBird not
separating the two: the records are real, the tier is real, and the attribution is the part
nobody knows. The label is what keeps the page from asserting the part nobody knows.

## regionJoinNormalizes

**The region join normalizes names, and says what it drops**

Date: 2026-10-05
Decided by: Sameer
Status: accepted
Affects: [product.joinsReportDrops](local/product/dataset.md#joinsreportdrops), [product.orderByFrequency](local/product/presentation.md#orderbyfrequency)
Subject: `build.join_loss`, `dataset.region.counts`
Kind: Direct — `birdview-status` and `birdview-regions` read the same facets under different taxonomies, so the two committed tables disagree about what a name means
Found by: the owner, asking why the Hairy Woodpecker's count looked wrong, then asking for an audit of everything lost through joins

### The audit

Every verbatim name eBird publishes for the US and Canada, against the species BirdView draws.
**1,402 names, 1,238,404,492 records:**

| | names | records | |
| --- | --- | --- | --- |
| match a drawn species | 1,062 | 1,228,366,591 | 99.19% |
| match a species BirdView rejects | 303 | 22,213 | correctly excluded |
| **match nothing** | **37** | **10,015,688** | **0.81%, silently dropped** |

Of the 10,015,688: **9,954,362 belong to ten species BirdView draws**, every one of them
already covered by `AviListRenames` or `CountedAs` — tables `birdview-regions` never called.
The remaining 61,326, under 27 names, are birds BirdView deliberately does not draw.

| name eBird uses | records | |
| --- | --- | --- |
| *Dryobates villosus* | **7,310,342** | Hairy Woodpecker |
| *Icterus bullockii* | 868,123 | Bullock's Oriole |
| *Coccothraustes vespertinus* | 838,952 | Evening Grosbeak |
| *Numenius phaeopus* | 556,569 | Whimbrel |
| *Pica nuttalli* | 125,766 | Yellow-billed Magpie |
| *Dryobates albolarvatus / arizonae / borealis* | 238,503 | three woodpeckers |
| *Gygis alba*, *Saxicola stejnegeri* | 16,107 | two lumps |

**0.81% of records overall; 99.3% of the records of the species affected.** The aggregate is
trivial and the per-species damage is near-total, which is why no summary statistic caught it.

### Why it was invisible

A discarded record still counts. `totals[gid]` sums **every** facet name, while a species'
count takes only names matching its own — so those 7.3 million Hairy Woodpecker records sat in
the denominator of every share on the site and in no numerator at all, including the Hairy
Woodpecker's. The bird showed 0.7% of its own record and nothing anywhere reported a loss.

### Two changes

**One normalization, shared.** `taxonomy.Occurrence` applies `AviListRenames` and `CountedAs`
and is now what both readers of the occurrence data call. The bug was not a missing table
entry — every entry already existed — but a second reader that did not consult them.

**The join reports.** `birdview-regions` lists every name that matched no drawn species with
its record count, and **refuses to write** above `maxOrphanFraction` = 0.05%, a bar set just
above the 61,326 records of legitimate residue. Before this change the figure was 0.81%.

### What this does not touch

The shared names of [sharedCountShownOnBoth](#sharedcountshownonboth). *Setophaga petechia* and *Trogon elegans* match a drawn
species exactly — they are not orphans and never were, which is why this audit does not list
them and why that one needed a different instrument entirely.

### Frequency sorts the other way from size, and that is a reciprocal, not a mistake

*Raised by the owner 2026-10-05, evaluated, nothing changed.* Length, wingspan and mass put the
smallest bird first; Frequency puts the commonest first. The quality being ordered therefore
runs **least → most** under three metrics and **most → least** under the fourth, which reads as
an inconsistency.

It is one, and it cannot be fixed by reversing the sort. **"1 in N" is a reciprocal**: N rises
as the bird gets scarcer, so ascending the printed number descends abundance. Every ordering
here obeys [product.ascendingOrder](local/product/ordering.md#ascendingorder) — ascending the selected metric — and the metric itself is
inverted relative to the other three.

Reversing it would move the inconsistency rather than remove it:

| | printed numbers | the bird quality |
| --- | --- | --- |
| **as built** — commonest first | **ascend**, 1 in 43 → 1 in 1,252,475,518 | descends |
| reversed — rarest first | descend | ascends, matching size |

The deciding argument is that **the order as built is visible and the reversed one is not.** A
reader scrolling sees the printed figure climb monotonically; nothing has to be inferred or
explained. Reversed, the numbers would count *down* under a control that ascends everywhere
else, and the justification — *we ascend abundance, which descends the printed number* — is
nowhere on the page.

The root of the tension is a decision taken for other reasons and still right:
[product.shareOfRecords](local/product/presentation.md#shareofrecords) shows a scarcity-shaped number and refuses to call it rarity. If the
control were named for the shape of the figure the tension would vanish, and that name is
exactly the one the requirement forbids. The felt inversion is the price of not shipping a
rarity tag, and it is worth paying.

**What is done instead is to say the direction rather than imply it:** the Sort by control
carries *most common first*, the help repeats it, and the first tier on the page reads
`1 in 10 – 1 in 99`. Naming the tiers — "the birds you see most" — was considered and rejected
on the same grounds as every other tier name ([frequencyMetric](#frequencymetric)).

### Two renames proposed and declined

Both were put forward to resolve the inversion. Each fails in its own way, and the three
candidates fail at different levels, which is why none of them is free.

**Rarity.** It would genuinely fix the direction: ascending rarity *is* commonest first, so
the name would finally agree with the order. It is declined because the figure is a
**reporting rate, not scarcity**, and the two come apart by orders of magnitude for the
eleven species eBird withholds. Labelled Rarity, Spotted Owl reads as the 927th rarest bird of
1,062 at 1 in 3,219,731 — while its own page says eBird suppresses it and the counts are
iNaturalist's alone. The label would contradict the disclosure. "Rarity" is also a term of art
in birding, and the word claims on its own whatever the project intends.

**Observations.** Accurate about the ordering: within one selection every species divides by
the same denominator, so ranking by share and ranking by raw count are **provably the same
order** — verified over all regions and over New York. It is declined because *observations*
is the **unit, not the quantity**. Length is measured in millimetres and the control does not
say "Millimetres"; mass is in grams and it does not say "Grams". Having just standardised
*observation* as the unit on every tile ([product.shareOfRecords](local/product/presentation.md#shareofrecords)), reusing it as the axis
name would conflate the two.

On direction the second proposal is a wash and was offered as such: ascending *Observations*
would mean fewest-first, exactly the inversion it was meant to cure.

**Infrequency.** Proposed third, and it is the only candidate correct on **both** axes at
once. Ascending infrequency *is* commonest first, so the direction agrees with the printed
number — and unlike *Rarity* it keeps the right referent, naming how seldom a bird is
**recorded** rather than asserting how scarce the bird is. It therefore escapes the trap the
other two fall into, that every frequency-polarity name points the wrong way and every
scarcity-polarity name overclaims.

It is declined on usability alone, which is the honest reason. *Infrequency* is a real English
word and an uncommon one, and it would sit in a row of four buttons beside Length, Wingspan
and Mass — three everyday nouns and one that makes a reader stop. The project has made this
trade before and in the same direction: [product.shareOfRecords](local/product/presentation.md#shareofrecords) prints *1 in N* rather than
a percentage because 0.000001% is unreadable, choosing the legible form over the conventional
one. Here the legible form is the vaguer one.

So the four candidates miss in four different places — **Frequency** is vague about the
quantity, **Rarity** is wrong about it, **Observations** names the unit instead, and
**Infrequency** is right about everything except the reader. Frequency is the least wrong.

### What is done instead of renaming

The inversion is **stated rather than removed**. The Order by row now carries the direction for
every metric — *smallest first* ruled under the three size metrics, *most common first*
under Frequency — so a reader who has learned the pattern from three of them is told where
the fourth departs, instead of inferring a bug.

**Both captions stand on every view**, whichever metric is selected. One caption that changed
with the choice would tell a reader what they had just clicked; two standing captions tell
them the controls behave differently *before* they click, which is the thing worth knowing.
Each sits under its own group behind a thin rule the width of the wider of the buttons or the
text — an inline-flex column sizes to its widest child, so the Frequency rule tracks its own
caption without anything being measured.

## monthPresence

**A month is present when the bird is, not when somebody was looking**

Date: 2026-10-07
Decided by: Sameer
Status: amended by ebirdSilenceNotAbsence
Affects: [product.regionAndMonthFilter](local/product/filtering.md#regionandmonthfilter)
Subject: `presentation.region_filter`, `dataset.region.counts`
Kind: Direct — [product.regionAndMonthFilter](local/product/filtering.md#regionandmonthfilter) decides membership per region and says nothing about when in the year, and extending the filter to months needs a rule it does not have
Found by: the owner, asking what it would take to add a Months control beside Regions

### The decision

The Regions control becomes **Regions and months**. Months filter **which species are drawn**,
not only the figure on the tile, because that is how every other filter in BirdView behaves
and a control that silently worked differently would be the worse surprise.

A species is present in a region **in a month** when

```
    (its records that month / all bird records that month)
  ÷ (its records that year  / all bird records that year )   >=  0.2
```

and it is present in that region at all by [product.regionAndMonthFilter](local/product/filtering.md#regionandmonthfilter)'s annual rule, which is
unchanged.

### Why not recurrence, which is the model everywhere else

[product.scopeIsRecurringAvifauna](local/product/dataset.md#scopeisrecurringavifauna) decides scope by **recurrence** — records falling in repeated
episodes across years — and that is the right instrument for "is this bird really here". It
cannot be had per month at an acceptable price. The `eventDate` facet caps at 1,000 distinct
dates and truncates for every common species, so months read off it would be whichever dates
GBIF returned first; and asking per month properly means faceting `year` under each of twelve
`month` filters, which is twelve times the slowest run in the project.

So the monthly rule works from **counts**, and this is a real departure from the model
recorded here rather than hidden in a threshold constant.

### Why the ratio, and not the share

Dividing by the region-month's own total is not optional. New York records **20.9%** of its
birding in May and **4.8%** in December, so a raw share would make every bird a May bird. The
ratio asks a different question — *is this species a bigger or smaller part of what people
record this month than it is across the year* — and 1.0 means exactly as present as usual.

It is [product.shareOfRecords](local/product/presentation.md#shareofrecords)'s own logic at month granularity, not a new concept.

### The measurements, 2026-10-07, eBird via GBIF

New York (62,467,434 records) and Hawaii (2,507,087). Each species' ratio, by month:

| species | behavior | months at or above 0.2 | correct? |
| --- | --- | --- | --- |
| Northern Cardinal | resident | all twelve (0.6–1.3) | yes |
| American Robin | partial migrant | all twelve (0.48–1.06) | yes — robins winter in NY |
| Dark-eyed Junco | winter visitor | all but August (Aug 0.18) | yes |
| Blackpoll Warbler | long-distance migrant | May, Jun, Sep, Oct | yes |
| Bobolink | breeder | May–Sep | yes |
| Semipalmated Sandpiper | passage shorebird | May–Oct | yes |
| Snow Bunting | winter visitor | Oct–Mar | yes |
| Warbling White-eye (HI) | resident | all twelve | yes |
| **Pacific Golden-Plover (HI)** | **breeds in Alaska** | **absent May–Jul** | **yes** |
| Nēnē (HI) | resident | all twelve | yes |

**The threshold is not load-bearing.** Across both regions the passing ratios cluster at
**0.31 and above** and the failing ones at **0.17 and below**. Nothing lands between, so any
cut from 0.2 to 0.3 gives the same answer — the property that made
[`minPelagicShare`](../../cmd/birdview-regions/main.go) defensible for the same reason.

**The Pacific Golden-Plover is the result worth keeping.** It leaves Hawaii in late April to
breed in Alaska and returns in August, and the rule recovered exactly that from record counts,
with nothing about the species, its migration or its breeding range encoded anywhere.

### Where the rule abstains, and why it must

Millerbird has **28 records** in Hawaii, all of them in March, April, August, September and
October. It is a Nihoa resident, present every month of the year. What those records show is
**when anyone last got to Nihoa**.

So below an evidence floor the month filter **abstains**: the species is shown in every
selected month, and the page says why. Hiding it from June would be a statement about the
bird, when the only fact available is a statement about the survey — the same distinction
[product.undrawnSpeciesAnswer](local/product/presentation.md#undrawnspeciesanswer) draws for a species with no records, and the one
`No data for these regions yet` draws for a region with no denominator.

**The floor is 200 of the species' own records in that region**, and it is a floor on
evidence, so it counts the species' records there rather than the region's total.

Calibrated 2026-10-07 against the 27,830 species-region cells of the committed table and the
real monthly effort of New York and Hawaii. The risk below is that a species present in
**every** month is wrongly called absent in at least one, purely by sampling:

| floor | cells abstaining | records abstaining | false absence, NY | false absence, HI |
| --- | --- | --- | --- | --- |
| 50 | 13.5% | — | 38.4% | 29.9% |
| 100 | 19.0% | 0.0145% | 10.2% | 7.9% |
| 150 | 22.4% | 0.0240% | 2.4% | 0.6% |
| **200** | **25.1%** | **0.0345%** | **0.5%** | **0.2%** |
| 300 | 28.6% | 0.0540% | 0.0% | 0.0% |

**A quarter of the cells, and three hundredths of one percent of the records.** That is the
shape that makes the floor cheap: abstention is concentrated almost entirely where there is
nothing to abstain from. 117 of 1,062 species hold under 200 records in every region they
occur in and will never be month-filtered at all; those are the vagrants, island endemics and
species eBird withholds, for which showing every month is the only honest answer.

**100 was the first guess and it is wrong**, which is why this was measured rather than
reasoned. A uniform month share makes 100 look safe at 0.3%, but effort is not uniform — New
York records 20.9% of its birding in May and 4.8% in December — and against the real profile
100 misfires on **10.2%** of year-round species. The knee is at 200.

**The threshold stays at the low end of its empty gap for the same reason.** A stricter cut
is easier to fail, so it would raise this risk; 0.2 rather than 0.3 buys margin here at no
cost to the verdicts measured above.

### What the interface says when it abstains

The tile marks the figure, as [product.namesRecordsFiledUnder](local/product/presentation.md#namesrecordsfiledunder)'s shared-name seam does, and the
detail page carries the same disclosure, as [product.tileOpensDetail](local/product/presentation.md#tileopensdetail) requires of anything a
tile shows. The marker names the reason rather than the mechanism: too few records to tell
the months the bird is here from the months nobody looked, so it is shown in all of them.

### Consequences

- All twelve months selected must reproduce the annual figures **exactly**. Every record
  carries a month — 65,504,204 checked across both sources, none missing — so this is a check
  and not an assumption.
- `birdview-regions` goes from about 140 queries to about 2,300, one to three hours, and
  therefore needs checkpoint and resume before it is run at all. **Done 2026-10-07**: the
  land pass caches each region's raw facet as it completes and skips what it already has.
  The cache holds the **unnormalized facet**, not presence or counts, for the reason
  `birdview-status` caches raw dates — asking GBIF is expensive and deciding what the answer
  means is free, and this very decision is changing the second.
- **The marine pass costs 2 queries per sea, not 24.** GBIF ORs repeated `gadmGid` values, so
  one month facet over the box and one restricted to `gadmGid=USA|CAN` give total and on-land
  per month, and the difference is water. Verified 2026-10-07. It counts **11 records fewer**
  in 108 million than today's `gadmLevel0Gid` facet, those being records inside a third
  country's polygon; that is 0.00001%, against the ~3% by which "on water" already overstates
  "at sea" (`CountOffLandInBox`).
- The committed table carries twelve counts per species-region cell and per region total.

### Rejected

**A fixed record count per month.** [product.regionAndMonthFilter](local/product/filtering.md#regionandmonthfilter) already rejected absolute bars
for region membership — eBird's volume doubled in five years, so any fixed count admits more
species over time for nothing — and a month is the same argument twelve times over.

**One in a million, applied to the region-month.** The bar degenerates where it is most
needed: Hawaii records roughly 208,000 birds a month, so one per million is 0.2 records and
any single sighting would read as presence.

**Season groups, named.** Rejected on [product.orderByFrequency](local/product/presentation.md#orderbyfrequency)'s existing rule — state the
range, do not name the band. "Winter" is wrong for Hawaii, Palmyra, Navassa, Wake and
Johnston whatever the hemisphere, and the Central Pacific box spans −10° to 40° and crosses
the equator. The twelve months stand unlabeled and ungrouped.

**A "This month" control.** Dropped by the owner: clicking **All** when every month is
selected clears them, so the current month is two clicks away and a dedicated control buys
one. It would also have made the index time-dependent, breaking
[acceptance.honestRevalidation](acceptance.md#honestrevalidation)'s invariant that a page is a pure function of its build and
its URL.

## ebirdSilenceNotAbsence

**eBird's silence is not a bird's absence**

Date: 2026-10-09
Decided by: Sameer
Status: amended by countEbirdFirst
Amends: monthPresence
Affects: [product.namesRecordsFiledUnder](local/product/presentation.md#namesrecordsfiledunder)
Subject: `presentation.region_filter`, `dataset.region.counts`, `presentation.name_provenance`
Kind: Direct — [monthPresence](#monthpresence)'s month filter treats a month with no records as a month without the bird, and for some species in some regions eBird publishes no records **on purpose**
Found by: the owner, asking whether dropping iNaturalist would skew BirdView; measuring that turned up Long-eared Owl with 15 eBird records in the US against 17,015 in Canada

### What was wrong

GBIF's eBird export applies eBird's [Sensitive Species list](imports.md#ebird-sensitive), and
the list has **regional and seasonal** entries as well as the nine global ones BirdView already
knew as `WithheldByEBird`. In a listed region and window, eBird's records are missing by policy.

The month filter read that as absence. **Painted Bunting was hidden from Florida in January**,
the heart of its Florida season, because eBird publishes none of its Florida records from
October to mid-April; **Short-eared Owl from New York in June**, the same way. Both are the
error [product.undrawnSpeciesAnswer](local/product/presentation.md#undrawnspeciesanswer) exists to prevent: an absence of measurement read as a
measurement of absence.

### The decision

1. **eBird's list is a source** ([`ebird-sensitive`](imports.md#ebird-sensitive)). The entries
   for species BirdView draws, in the US and Canada, are restated as `taxonomy.EBirdSensitive` —
   9 global and 68 regional, each a species, a region and a date window.
2. **No entry is acted on until the records confirm it.** `birdview-regions` marks a month as
   *withheld by eBird* for a species and region only when the list names it **and** eBird's
   records there show the gap:
   - **a seasonal window** is confirmed when eBird's monthly rate inside it is **at most 5%** of
     its rate in the same region's other months;
   - **a whole-year listing** is confirmed when iNaturalist holds **at least three times** what
     eBird does there — [acceptance.disagreementsDeclared](acceptance.md#disagreementsdeclared)'s existing test for a source that
     out-records eBird, applied to one region.
   Only months the window covers **entirely** are marked; a month it half covers still holds
   half its records and is judged on them.
3. **A withheld month is not judged.** The month filter abstains there — the species is shown —
   exactly as it does below 200 records, and the tile and the detail page say why:
   *Withheld by eBird in Florida, Oct–Mar*.

### Why confirm rather than trust the list

**The list describes eBird's policy now; the export reflects what was applied when it was
built**, and they differ. Of the 230 region cells the US and Canada entries name, **204 are
confirmed** and **26 are not** — among them Greater Sage-Grouse across the West from February
to May (Utah: 422 eBird records a month inside the window, 61 outside), Ferruginous
Pygmy-Owl in Texas (4,809 records) and Spruce Grouse in Vermont (1,303). Marking those would put
*withheld by eBird* on records eBird plainly did publish, which is a false statement on the
page. Where the list and the records agree, the disclosure is true; where they disagree, the
records win and the ordinary rules apply.

**The thresholds are not load-bearing for the cases that matter.** Painted Bunting in Florida
runs at 0 inside its window against thousands a month outside; Long-eared Owl in New York holds
no eBird records and 39 on iNaturalist. The cases that fall the other way are a handful of
small counts — Piping Plover in Wisconsin, 9 eBird against 19 iNaturalist — where no threshold
would make the evidence anything but thin.

### What this does not change

Presence ([product.regionAndMonthFilter](local/product/filtering.md#regionandmonthfilter)) and the figure on the tile are untouched. Counting eBird
alone and using iNaturalist only where eBird is confirmed silent is the larger change in work
arising; this decision builds the table and the confirmation it will need, and fixes the bug
that was live.

## countEbirdFirst

**Count eBird, and use iNaturalist only where eBird is withholding a bird**

Date: 2026-10-09
Decided by: Sameer
Status: accepted
Amends: scopeFromRecurrence, ebirdSilenceNotAbsence
Affects: [product.regionAndMonthFilter](local/product/filtering.md#regionandmonthfilter), [product.shareOfRecords](local/product/presentation.md#shareofrecords), [product.scopeIsRecurringAvifauna](local/product/dataset.md#scopeisrecurringavifauna), [product.namesRecordsFiledUnder](local/product/presentation.md#namesrecordsfiledunder), [product.statusIsPublished](local/product/dataset.md#statusispublished)
Subject: `dataset.region.counts`, `presentation.region_filter`, `dataset.scope`
Kind: Direct — [product.regionAndMonthFilter](local/product/filtering.md#regionandmonthfilter) admitted a species to a region on either source against that source's own total, and iNaturalist's totals are small enough that one record passes
Found by: the owner, concerned that iNaturalist was skewing BirdView, then asking what dropping it would cost once eBird's withholdings were known ([ebirdSilenceNotAbsence](#ebirdsilencenotabsence))

### What was wrong

Every count summed the two sources, and presence was judged per source. iNaturalist is 1.24% of
land records, so the sum barely moved any figure — a median 0.6% — but the per-source test did
real damage: iNaturalist's New York total is 429,693, its bar at 1 per million is **0.43
records**, and any single photo passed it. **1,228 land presences rested on iNaturalist
alone**, and 985 of them, across 406 species, were escapes, domestic birds and strays: Red
Junglefowl in 42 regions, Japanese Quail in 27, Zebra Finch, Indian Peafowl, Budgerigar, Emu.
Scope was the same: `birdview-status` unioned the two sources' dates, so a species iNaturalist
alone photographed in repeated years recurred.

### The decision

**eBird's records are counted, everywhere. iNaturalist's are counted only in a month eBird is
confirmed to withhold the species in that region** (ebirdSilenceNotAbsence's flags). That applies to the figure
on every tile, to region presence, to the month filter's counts, and to the recurrence that
decides scope.

- **Presence** is eBird's count against eBird's total — or, where eBird is withholding the
  species, iNaturalist's count in those months against iNaturalist's total. An eBird-sized
  denominator would make every iNaturalist count look about forty-eight times rarer than it is.
- **Shared names**: both species take the shared name's eBird count ([product.namesRecordsFiledUnder](local/product/presentation.md#namesrecordsfiledunder)).
  eBird holds Northern Yellow Warbler's records only under *Setophaga petechia*; without this the
  commonest warbler in North America would appear in no region. Region totals are taken before
  the copy, so the shared count is never divided into itself twice.
- **`WithheldByEBird` is derived from eBird's list** — a global entry, or year-round entries for
  both the US and Canada — rather than kept by hand. That removes **Desertas Petrel**, which is on
  no entry of eBird's list; "withheld by eBird" about it was false.

### Why not drop iNaturalist entirely

Because eBird publishes where it is silent, and silence is not absence. Without iNaturalist,
nine species eBird withholds everywhere — Spotted Owl, Great Gray Owl, Gyrfalcon, Northern Hawk
Owl, Lesser Prairie-Chicken, Gunnison Sage-Grouse and three Hawaiian endemics — would have **no
location at all**: eBird's list says it hides them everywhere, which says nothing about where
they are. The same holds for the 57 region cells where eBird is silent all year, Long-eared Owl in
about thirty states among them. In those places iNaturalist is the only evidence, and it is
used there and labelled.

### Why not keep both and fix only the threshold

Judging the combined count against the combined total removed most of the escapes, but it would
have dropped Long-eared Owl from every US state: eBird's silence there is total, and 39
iNaturalist records do not reach 1 per million of New York's 62 million. Fixing the threshold
without knowing where eBird withholds trades one error for its opposite.

---

## whiteTernBorrows

**Blue-billed White-Tern borrows from BirdLife's narrower White Tern**

Date: 2026-10-13
Decided by: Sameer
Status: amended by vegaGullBorrows
Amends: abaSpeciesWithoutAvonet
Affects: [product.lengthAndWingspanSources](local/product/measurements.md#lengthandwingspansources)
Subject: `dataset.join.near_match.sheet`
Kind: Refinement of [abaSpeciesWithoutAvonet](#abaspecieswithoutavonet) — a near-match donor exists in a narrower concept than the one the lookup order reaches first

### What was wrong

*Gygis candida* has no AVONET row of its own, so it borrows from *Gygis alba* under
[abaSpeciesWithoutAvonet](#abaspecieswithoutavonet). The near-match lookup tries the eBird sheet first, and `AVONET2_eBird`'s
*Gygis alba* is the 2022 Clements concept, which still contains all three white terns AviList now
separates: *alba*, *candida* and *microrhyncha*. `AVONET1_BirdLife` already splits off
*microrhyncha*, the Little White Tern, so its *Gygis alba* is *alba* plus *candida* only. The
crosswalk shows it: eBird *Gygis alba* maps to BirdLife *Gygis alba* and *Gygis microrhyncha*.

| Row | Culmen | Tail | Wing | Mass |
| --- | --- | --- | --- | --- |
| `AVONET2_eBird` *Gygis alba* (all three) | 44.4 | 90.2 | 233.4 | 117 |
| `AVONET1_BirdLife` *Gygis alba* (without *microrhyncha*) | 49.7 | 103.8 | 251.8 | 117 |
| `AVONET1_BirdLife` *Gygis microrhyncha* | 39.0 | 76.8 | 215.0 | 117 |

The small Little White Tern pulls the eBird row's linear traits down. Mass is identical across all
three rows, the same unsplit-mass caveat abaSpeciesWithoutAvonet recorded for the gulls.

### The decision

**Blue-billed White-Tern's traits come from `AVONET1_BirdLife`'s *Gygis alba*.** They remain
`near-match`, because that row still combines two species. The source sheet is named in the
provenance record as [product.provenancePerMetric](local/product/measurements.md#provenancepermetric) already requires. Bill and tail change; mass, and the head, neck and
trunk modeled from it, do not. Wingspan does not either: the crosswalk already led to Fox's
BirdLife *Gygis alba*.

The sheet preference is declared per species in `taxonomy.NearMatchSheets`, not made a general
rule. The crosswalk was checked for every other near-match parent. Eight map one-to-one, so no
sheet offers a narrower row. The exception is *Larus argentatus*, Vega Gull's donor: the eBird row
maps to BirdLife *Larus argentatus* plus *Larus smithsonianus*. Whether BirdLife files *vegae*
under its *argentatus* is a taxonomic question this record does not answer, so Vega Gull is left
on the eBird row.

### Status of requirements touched

- [product.lengthAndWingspanSources](local/product/measurements.md#lengthandwingspansources) — Active; the traits line names the exception.
- [product.provenancePerMetric](local/product/measurements.md#provenancepermetric) — Active and unchanged; the sheet already appears in `source`.
- [acceptance.whiteTernFromBirdlife](acceptance.md#whiteternfrombirdlife) — new.
- abaSpeciesWithoutAvonet's open taxonomic check on the four disguised splits is **not** closed by this record.
- Open: whether Vega Gull should borrow from BirdLife's *Larus argentatus* in the same way.

---

## vegaGullBorrows

**Vega Gull borrows from BirdLife's *Larus smithsonianus*, and shared drawings say so**

Date: 2026-10-13
Decided by: Sameer
Status: amended by herringGullEuropeanRow
Amends: whiteTernBorrows, abaSpeciesWithoutAvonet
Affects: [product.borrowedMeasurementsSaySo](local/product/measurements.md#borrowedmeasurementssayso)
Subject: `dataset.join.near_match`, `presentation.shared_measurements`
Kind: Resolves the open item of [whiteTernBorrows](#whiteternborrows), and adds [product.borrowedMeasurementsSaySo](local/product/measurements.md#borrowedmeasurementssayso)

### What was found

[whiteTernBorrows](#whiteternborrows) left open which of BirdLife's two herring gull rows — *Larus argentatus* or
*Larus smithsonianus* — holds Vega Gull. Both vendored CC BY 4.0 workbooks answer it:

1. **AviList v2025b extended** (`Proposal_number` `2025-1136`, `Decision_summary` on the four
   herring gull rows) splits the complex into four species — polytypic *L. argentatus* (including
   *argenteus*), monotypic *L. smithsonianus*, monotypic *L. vegae* and monotypic *L. mongolicus* —
   and notes that *"Species L. vegae is sometimes considered a subspecies of L. smithsonianus"*. Its
   `English_name_BirdLife_v10` column names `European Herring Gull` (`avibase-F002188E`) on *Larus
   argentatus* alone, and gives *Larus vegae*'s range as breeding across Arctic Siberia and
   wintering mostly in Japan, the Korean Peninsula and eastern China.
2. **AVONET's `AVONET_Raw_Data` sheet** shows which specimens each BirdLife row actually averages.
   All 6 specimens behind `AVONET1_BirdLife`'s *Larus argentatus* (`AVIBASE-F002188E`) are tagged
   `eBird.species.group = Larus argentatus argentatus/argenteus`. The 8 specimens behind
   `AVONET1_BirdLife`'s *Larus smithsonianus* are 4 from the United States and Canada
   (`eBird.species.group = Larus argentatus smithsonianus`, `AVIBASE-60F32AF8`) and **4 from Russia
   and Japan** (NHMUK `1965.M.3949`, `84.1.7.5`, `95.5.14.5`, `98.12.2.634`).

So BirdLife's *Larus argentatus* is the European bird alone and is not a donor for Vega Gull.
BirdLife's *Larus smithsonianus* is the narrowest AVONET row that contains it — and averages four
Russian and Japanese specimens alongside four North American ones.

*(Re-based 2026-10-14 onto the vendored `avilist` and `avonet` workbooks. The 2026-10-13 text cited
the archived HBW-BirdLife checklist on DataZone, which had been fetched in a way DataZone's terms
forbid; those quotations are removed and [`birdlife-checklist`](imports.md#birdlife-checklist) is
not used.)*

### The decision

1. **Vega Gull borrows from `AVONET1_BirdLife`'s *Larus smithsonianus*,** labeled `near-match`.
   American Herring Gull already uses that row, so the two are drawn identically.
2. **Identical drawings are acceptable for borrowed values, and are noted on the drawing**
   ([product.borrowedMeasurementsSaySo](local/product/measurements.md#borrowedmeasurementssayso)). abaSpeciesWithoutAvonet named identical drawings only as a symptom of *silent*
   borrowing. It never forbade two labeled species sharing a row. The owner asked that the
   coincidence be visible on the tile, in the way a shared name already is.
3. **Vega Gull's wingspan becomes modeled.** It was a near-match to Fox's *Larus argentatus*,
   reached through the eBird crosswalk: the European bird's wingspan. Fox has no empirical value
   for *Larus smithsonianus*, so American Herring Gull's wingspan is modeled from its wing chord,
   and Vega Gull, with the same chord, now gets the same modeled value, 1,376.5 mm. The two
   drawings are therefore identical in every segment and in the crossbar.

**Four pairs shared a row before this decision, and none said so:** Siberian and American Pipit,
Chihuahuan and Eastern Meadowlark, Cocos and Brown Booby, and Vega and European Herring Gull. The
last becomes Vega and American Herring Gull.

### Still open

- American Herring Gull's row is BirdLife's three-subspecies concept, which AviList divides into
  three species, but it is labeled `measured`.
- European Herring Gull is read from `AVONET2_eBird`'s *Larus argentatus*, the 2022 concept that
  contains every herring gull, and is also labeled `measured`. BirdLife's European-only row exists.

Both are recorded in [todo.md](../todo.md).

### Status of requirements touched

- [product.borrowedMeasurementsSaySo](local/product/measurements.md#borrowedmeasurementssayso) — new.
- [product.lengthAndWingspanSources](local/product/measurements.md#lengthandwingspansources), [product.provenancePerMetric](local/product/measurements.md#provenancepermetric) — Active and unchanged.
- [acceptance.whiteTernFromBirdlife](acceptance.md#whiteternfrombirdlife) — Vega Gull now expects `AVONET1_BirdLife`.
- [acceptance.sharedMeasurementsDisclosed](acceptance.md#sharedmeasurementsdisclosed) — new.

---

## herringGullEuropeanRow

**European Herring Gull is read from BirdLife's European-only row**

Date: 2026-10-13
Decided by: Sameer
Status: accepted
Amends: vegaGullBorrows
Affects: [product.lengthAndWingspanSources](local/product/measurements.md#lengthandwingspansources)
Subject: `dataset.join.key`
Kind: Resolves one of the two items [vegaGullBorrows](#vegagullborrows) left open

### What was wrong

The ingest reads a species' own row from `AVONET2_eBird` first. That sheet's *Larus argentatus*
is the 2022 Clements concept, which still contains every herring gull, and European Herring Gull
was drawn from it as `measured`. AVONET's raw data shows what that row averages: **16 specimens,
of which 6 are tagged as European subspecies**. The other ten are 8 that BirdLife assigns to
*Larus smithsonianus* and 2 assigned to no BirdLife species. `AVONET1_BirdLife`'s *Larus
argentatus* is those 6 alone.

| Row | Specimens | Culmen | Tail | Wing |
| --- | --- | --- | --- | --- |
| `AVONET2_eBird` *Larus argentatus* (was used) | 16, 6 European | 70.5 | 168.5 | 429.0 |
| `AVONET1_BirdLife` *Larus argentatus* | 6, all European | 74.0 | 170.2 | 430.8 |

### The decision

**European Herring Gull's traits come from `AVONET1_BirdLife`'s *Larus argentatus*,** labeled
`measured`: every specimen behind it is this species. Mass is 1,091 g in both rows, so the
modeled head, neck and body do not change. Wingspan is Fox's empirical *Larus argentatus* either
way.

The sheet is declared per species in `taxonomy.OwnRowSheets`, with no fallback, as
[whiteTernBorrows](#whiteternborrows) did for near-match donors.

**American Herring Gull stays as it is**, by the owner's decision: `AVONET1_BirdLife`'s
*Larus smithsonianus*, labeled `measured`. It is the narrowest row AVONET has, and 4 of its 8
specimens are tagged American Herring Gull; the other 4 carry no subspecies.

### Status of requirements touched

- [product.lengthAndWingspanSources](local/product/measurements.md#lengthandwingspansources) — Active; the traits line names this case.
- [acceptance.herringGullFromEuropeanRow](acceptance.md#herringgullfromeuropeanrow) — new.

---

## belowBarCounted

**Records below the presence bar are counted, not discarded**

Date: 2026-10-13
Decided by: Sameer
Status: accepted
Affects: [product.regionAndMonthFilter](local/product/filtering.md#regionandmonthfilter)
Subject: `dataset.region.counts`
Kind: Direct — [product.regionAndMonthFilter](local/product/filtering.md#regionandmonthfilter)'s presence bar was also deciding which records [product.shareOfRecords](local/product/presentation.md#shareofrecords)'s figure counts
Found by: the owner, asking why Emu's tile said it had zero observations

### What was wrong

`birdview-regions` wrote a species' count for a region only when the species reached 1 record
per million there. Below that, the records were dropped, so the figure on the tile never saw
them. Emu is `established` in Texas on 19 eBird records in 18 episodes across 11 years, which is
0.26 per million of Texas's birding, and its tile read *0 in 1,237,352,304 species observations*.
**47 drawn species read zero everywhere**, most of them escapes whose iNaturalist-only presences
[countEbirdFirst](#countebirdfirst) removed, and every other species lost whatever it had below the bar.

### The decision

**Every count is kept. The bar decides membership only.** `region-presence.tsv` writes a cell for
every region where a species has records, and marks the cells below the bar; the ingest carries
them separately from the regions the species belongs to, and the figure on the tile counts both.
Region and month filtering are unchanged: a species below the bar in Texas is still not listed
under Texas.

This is the half of the parked *Judge presence over the whole selection* item that decides
nothing about where the bar sits, so it is done on its own.

### Status of requirements touched

- [product.regionAndMonthFilter](local/product/filtering.md#regionandmonthfilter) — Active; membership no longer decides what is counted.
- [product.shareOfRecords](local/product/presentation.md#shareofrecords) — Active and unchanged; its figure now has all the records it names.
- [acceptance.belowBarCountsKept](acceptance.md#belowbarcountskept) — new.

---

## escapedBirdsOffIndex

**Escaped and kept birds come off the index**

Date: 2026-10-14
Decided by: Sameer
Status: accepted
Affects: [product.escapedBirdsListed](local/product/measurements.md#escapedbirdslisted), [product.statusIsPublished](local/product/dataset.md#statusispublished), [product.declaredExceptions](local/product/dataset.md#declaredexceptions)
Subject: `presentation.index.exotic`, `dataset.status.exceptions`
Kind: Resolves the open question on [product.declaredExceptions](local/product/dataset.md#declaredexceptions)'s second class, and implements [product.statusIsPublished](local/product/dataset.md#statusispublished)'s `exotic` flag

### What was wrong

Recurrence is blind to origin, as [scopeFromRecurrence](#scopefromrecurrence) said, and [product.declaredExceptions](local/product/dataset.md#declaredexceptions) required a declared list for
birds kept going by people. None was built, so escapes were drawn as ordinary birds. Once
[belowBarCounted](#belowbarcounted) counted every record, they showed up as figures too. The owner found them
distracting.

### Options

| Option | Effect | |
| --- | --- | --- |
| Escaped cage, aviary and farm birds | About 40 species, all from one recognizable class | **chosen** |
| Every introduced species | House Sparrow, European Starling, Rock Pigeon and much of Hawaii's avifauna; no open source records origin, so every entry would be a judgement | rejected — removes many of the commonest birds a reader meets |

### The decision

- **`status.Exotic` declares the list**, each entry with its reason. It started from the 47
  drawn species that reach the presence bar in no region and kept out the wild ones: Little
  Curlew, Ruddy Quail-Dove, Desertas Petrel, and — as doubtful rather than escapes —
  West Indian Whistling-Duck, Cuban Grassquit and Cinnamon-rumped Seedeater.
- **Listed birds stay in the dataset**, flagged, so their detail pages are served. The index and
  every control leave them out, and `/excluded` lists them, linked.
- **The ingest fails if a listed bird is not drawn**, so a list entry cannot outlive the bird.
- **Emu is on the list.** A note of 2026-10-13 in the open-work list read "Emu stays", written
  while [belowBarCounted](#belowbarcounted) was restoring its counts; the owner decided on 2026-10-14 that it is
  excluded with the other farm escapes, which replaces that note.

### Status of requirements touched

- [product.escapedBirdsListed](local/product/measurements.md#escapedbirdslisted) — new.
- [product.statusIsPublished](local/product/dataset.md#statusispublished) — Active; `exotic` exists, as a declared list.
- [product.declaredExceptions](local/product/dataset.md#declaredexceptions) — Active; its second class is now implemented for escapes.
- [acceptance.escapedBirdsListedAndServed](acceptance.md#escapedbirdslistedandserved) — new.

## stickFigure

**Readers are told the drawing is a stick figure**

Date: 2026-10-15
Decided by: Sameer
Status: accepted
Affects: [product.segmentComposition](local/product/diagram.md#segmentcomposition)
Subject: `diagram.segments` (wording)
Kind: Presentation; no behavior changes

### What was wrong

The site called each bird's drawing a *diagram*. That is correct, but it does not tell a reader
what they will see, and since [product.pipelineDiagram](local/product/measurements.md#pipelinediagram) the site has a second diagram, of the data
pipeline, linked from the same page as the first.

### The decision

**On the site and in the guides, the drawing is a *stick figure*.** It is one: lines for the bill,
neck, body and tail, an oval head, and a crossbar for the wings. The term also says up front that
the drawing is not a likeness, which is the honesty the rest of the page depends on. Because
*stick figure* can suggest crude or not to scale, it is paired with *to scale* where a reader
first meets it, and the contrast is the point. In running text *the figure* or *the drawing*
takes over, since the mass cube is not a stick and the full term repeated reads as a gimmick.

**The spec, the code and the tests keep *diagram*.** Template names, requirement titles and test
names are internal and correct, and renaming them would be churn with no reader benefit. On the
site, *diagram* now means only the pipeline diagram.

### Status of requirements touched

- [product.segmentComposition](local/product/diagram.md#segmentcomposition) — Active; states the reader-facing term.
- [acceptance.readersToldStickFigure](acceptance.md#readerstoldstickfigure) — new.

## Work arising

Moved to [todo.md](../todo.md) on 2026-10-13, with the design notes that supported it, so that
everything outstanding is in one file. Decisions stay here; open work does not.

---

## bundleMovesToRiver

**The bundle moves from spec/ to river/; the site keeps /spec**

Date: 2026-10-15
Decided by: Sameer
Status: amended by requirementsNotServed
Affects: [tech.bundleIsNormative](local/tech/process.md#bundleisnormative)
Subject: `process.spec_location`
Kind: Amends [tech.bundleIsNormative](local/tech/process.md#bundleisnormative)

### What changed

[tech.bundleIsNormative](local/tech/process.md#bundleisnormative) named `spec/` as where the normative bundle lives. On adopting River, whose
method gives the bundle the directory `river/`, it moves there. [tech.bundleIsNormative](local/tech/process.md#bundleisnormative)'s value is now `river/`.

### What did not change

**The site still serves the documents at `/spec`.** The directory names the method; the URL names
the thing a reader is looking at, and a birder following a link from a species page wants the
requirements, not the name of the process that produced them. The URL is also already linked
from outside. `repoLinks` in `cmd/birdview` maps one to the other, so neither name bends.

### What the move exposed

`TestServedDocumentsHaveNoBrokenLinks` ([acceptance.categoryOriginShown](acceptance.md#categoryoriginshown)) was written for exactly the bug a move like this
causes — a repository link like `../river/product.md` left unrewritten, so a browser gets a 404 —
and passed over 53 of them, because it followed only links that began with `/`. An unrewritten
link is relative, so it was skipped rather than fetched. The test could never have caught its own
founding bug.

It now resolves every link against the page it sits on, as a browser does, and was watched
failing twice before being trusted: once with the rewriter deliberately left unfixed, once with a
stale entry planted in its known-breakage list.

Making it honest surfaced **26 links that were broken before the move**, from served documents to
repository files the server does not serve: source files, data tables, vendored inputs, and the
provenance records under `river/sources/`. They are listed by name in the test rather than
excused, so a new broken link still fails and a fixed one must leave the list. What they should
become — links to the published repository, or plain code spans — is undecided and logged in
[todo.md](../todo.md).

---

## requirementsNotServed

**The requirements are no longer served; the guides link to the repository**

Date: 2026-10-16
Decided by: Sameer
Status: accepted
Amends: bundleMovesToRiver
Affects: [product.categoryProvenance](local/product/measurements.md#categoryprovenance)
Subject: `presentation.docs`
Kind: Amends [product.categoryProvenance](local/product/measurements.md#categoryprovenance); supersedes the half of [bundleMovesToRiver](#bundlemovestoriver) that kept `/spec`

### Why the spec was served

[product.categoryProvenance](local/product/measurements.md#categoryprovenance) gave two reasons, and the second rested on the first. The guides were
served by BirdView itself because *"the project has no public repository to link to"*; and the
requirement documents were then served too, in commit `a19bda7` on 2026-10-03, because the
guides cite them constantly and *"a citation to a 404 is worse than no citation."* Serving the
spec was never a goal of its own.

### What changed

**The first premise expired.** `github.com/google/birdview` has existed since 2026-10-08 and is
to be published. [product.categoryProvenance](local/product/measurements.md#categoryprovenance) also leaned on [tech.selfContainedServer](local/tech/delivery.md#selfcontainedserver), but [tech.selfContainedServer](local/tech/delivery.md#selfcontainedserver) limits the requests the
*server* makes, not the links a reader follows. Its rationale — a laptop in a cabin — argues for
serving the guides offline, which still holds, and only weakly for serving requirement
documents to an offline birder.

**And the cost grew.** Under River the bundle holds working documents: the decision log, the
open-work list, bindings, and River's own `process.md`. Serving them was also behind most of
the friction of the migration — the link rewriter's directory-to-URL mapping, the reason River's
nested layout would have needed a server change, and all 26 links that 404'd because a served
requirement linked to source code or a data table the site could not serve.

### The decision

1. **`/docs` stays served; `/spec` is not.** The embed package for `river/` is deleted, so the
   server holds no requirement document at all.
2. **A guide's link into the repository becomes a link to the published repository.** That
   covers requirements, source code, data tables and vendored inputs alike. A link to another
   guide becomes its `/docs` path.
3. **The three reader-facing links to the spec now go to the guide.** The index, the
   undrawn-species page and `/excluded` linked readers to [product.scopeIsRecurringAvifauna](local/product/dataset.md#scopeisrecurringavifauna) and [product.escapedBirdsListed](local/product/measurements.md#escapedbirdslisted) for scope; they now
   link to [Scope](../../docs/data.md) in the data guide, which explains it for a reader rather
   than as a requirement.

### Guardrails, and one coverage gap closed

The broken-links check ([acceptance.categoryOriginShown](acceptance.md#categoryoriginshown)) now confirms that every repository link names a file that
exists in the checkout, and was watched failing on a misspelled one. The 26 known-broken links
are gone, and the list that held them is deleted.

**Stopping serving would have quietly cut [acceptance.noContactAddressPublished](acceptance.md#nocontactaddresspublished)'s coverage.** The
personal-address check scanned the requirements because they were served; they are still
public with the repository. It now reads `river/*.md` from the checkout — the same files — and
was watched failing on an address planted in `river/goals.md`.

### Known limitation

GitHub builds a heading's anchor from its full text, so a link to `decisions.md#seasharediffers` opens
the file at the top rather than at the entry. Logged in [todo.md](../todo.md).

---

## descriptiveIdentifiers

**Every identifier is a descriptive name, and every heading is the identifier alone**

Date: 2026-10-16
Decided by: Sameer
Status: accepted
Affects: none
Subject: `process.identifiers`
Kind: Renames every goal, requirement, criterion and decision; no behavior changes

### What changed

All 230 identifiers — 6 goals, 62 product and 19 technical requirements, 97 acceptance criteria
and 46 decisions — went from letter-and-number codes, a `P`, `T`, `AC` or `CR` and three digits,
to descriptive camelCase:
[product.scopeIsRecurringAvifauna](local/product/dataset.md#scopeisrecurringavifauna),
[tech.fetchRestrictedInputs](local/tech/delivery.md#fetchrestrictedinputs),
[acceptance.bothCommandsBuild](acceptance.md#bothcommandsbuild),
[decisions.scopeFromRecurrence](#scopefromrecurrence). This is River's convention, adopted with
the method.

**A reference is qualified by kind outside the file that defines it** — `product.x`, `tech.x`,
`goals.x`, `decisions.x`, `acceptance.x` — and bare inside it. Requirement and criterion names are
unique across all files, because composition will merge them into the `app` and `dataset`
packages, where a shared name would be a silent conflict.

**An item's heading is its identifier alone, with the title in bold beneath.** A code host
anchors a heading by its whole text, so a heading of the old form — the code, a dash, then the
title — anchored as the code and the title run together, and a link to the code alone opened the
file at the top. That mattered once
[requirementsNotServed](#requirementsnotserved) sent the guides' citations to the published
repository. Now every anchor is the identifier lowercased, and a title can be reworded without
breaking a link.

### What it cost, and why it was done anyway

About 3,000 references across 96 files: Go comments, log output, generated data, the guides.
The alternative was to let River accept legacy identifiers. Sameer chose to rename: a migration
from specman happens once, and optimizing the method for it would leave the cost in every
project afterwards.

Two pipeline-diagram labels named requirements by identifier; they now say what they mean in
words, and the picture was re-rendered from its source.

### Checks

`refs.py` was rewritten: it now checks every tracked file rather than three, resolves every
qualified reference and every anchored link into the bundle, and fails on any identifier in the
superseded form. Its first run found three links that had been broken since specman's method was
replaced — two citing a rule River had dropped, which River has since restored.

---

## nestedLayout

**The bundle takes River's nested layout, with a package per source**

Date: 2026-10-16
Decided by: Sameer
Status: accepted
Affects: none
Subject: `process.layout`
Kind: Moves every file in the bundle; no requirement changes

### What moved

| From | To |
| --- | --- |
| `process.md` | `README.md` — River's method, under the name a code host shows first |
| `goals.md` | `goals/goals.md`, `package goals` |
| `product.md`, `tech.md` | `requirements/sources/product/` and `.../tech/`, one file per group, each declaring its package |
| `sources.md` | `requirements/requirements.md`, now declaring `module birdview` |
| `decisions.md`, `acceptance.md` | `requirements/` |
| `sources/<name>/` | `requirements/sources/<package>/` |

**Every directory follows Go's rules, and a package's directory carries its name.** A source
whose name has punctuation keeps it in the manifest and in the data, where readers see it, and
drops it for its package and directories: `inat-api` is `package inatapi` in
`requirements/sources/inatapi/`, and its vendored files moved to `third_party/inatapi/` so the
two trees still share a name. `fox-wingspan` and `inat-terms` likewise.

**Why `third_party/` and `requirements/sources/` both exist.** They do not divide local from
imported. `third_party/` holds the imported material itself, each file under its own licence,
outside the bundle because BirdView's Apache licence at the root would otherwise appear to cover
it. `requirements/sources/<package>/` holds BirdView's own writing about each source: its
provenance record and, where it has rules, BirdView's transcription of them. River now allows
exactly this.

### What it exposed

- **Seven technical requirements have no group.** They sit in `tech/ungrouped.md` until each is
  placed, rather than being filed by guess.
- **Every item's title ran into the line after it.** The rename put the title in bold beneath
  the heading but left no blank line before `Goals:`, so a code host rendered them as one
  paragraph. 178 separated.
- **Five Go comments still carried the old `inat-api` rule numbers**, which the rename missed
  because it looked only for the backticked form. `refs.py` now rejects both forms.
- **River's checker now runs here**, as `goals.py`, and found that it read a title followed by a
  status marker — *(withdrawn)*, *(hard rule)* — as missing. River was fixed.

---

## hereAndNowCaching

**The boundary file is cached for a year, and Now keeps the server's clock out of every page**

Date: 2026-10-18
Decided by: Sameer
Status: accepted
Affects: [tech.cacheLifetimeStated](local/tech/delivery.md#cachelifetimestated), [product.nowSelectsCurrentMonth](local/product/filtering.md#nowselectscurrentmonth), [product.hereSelectsLocalRegion](local/product/filtering.md#hereselectslocalregion)
Subject: `serving.cache_validation`
Kind: Amends [tech.cacheLifetimeStated](local/tech/delivery.md#cachelifetimestated); settles how [product.nowSelectsCurrentMonth](local/product/filtering.md#nowselectscurrentmonth) works without JavaScript

### What conflicted

Here and Now ([product.hereSelectsLocalRegion](local/product/filtering.md#hereselectslocalregion), [product.nowSelectsCurrentMonth](local/product/filtering.md#nowselectscurrentmonth)) ran into
[tech.cacheLifetimeStated](local/tech/delivery.md#cachelifetimestated) twice.

1. **The boundary file wants a long cache lifetime.** It loads with every index page
   ([product.locationStaysInBrowser](local/product/filtering.md#locationstaysinbrowser)), so revalidating it on every load is waste. The requirement
   said `no-cache` and an ETag on *every* response.
2. **A Now link carrying the server's month would put a clock into the page.** The requirement's
   ETag is a fingerprint of the build and the URL, sound only because a page has "no clock
   anywhere in the render path", which [acceptance.honestRevalidation](acceptance.md#honestrevalidation) checks. After a month ended,
   a browser revalidating a cached index would get 304 and keep last month's Now link.

### Decided

1. **A file whose URL carries a hash of its own content is served immutable for a year.** It can
   never be stale: new content is a new URL. The pages that load it keep `no-cache`.
2. **Without JavaScript, Now links to `/now`**, which redirects to the index for the server's
   current month and is marked `no-store`. With JavaScript the script sets the link to the
   reader's own month and `/now` is never requested. No page carries the server's month.

### Considered and not chosen

- **The month in the ETag.** Workable, but it amends the purity rationale and its check, and
  every cached page would refetch in full once a month.
- **Now only with JavaScript**, inactive otherwise like Here. It drops the server-month fallback
  the owner accepted.

---

## regionFilterReinstated

**Region filtering returns, because GBIF serves the per-region data noLocationData found missing**

Date: 2026-10-02
Recorded: 2026-10-21, from commit `396fcff`; the requirement was rewritten then, and no entry was written
Decided by: Sameer
Status: accepted
Supersedes: noLocationData
Affects: [product.regionAndMonthFilter](local/product/filtering.md#regionandmonthfilter), [product.filtersCompose](local/product/months.md#filterscompose)

### Decided

[product.regionAndMonthFilter](local/product/filtering.md#regionandmonthfilter) and [product.filtersCompose](local/product/months.md#filterscompose) are reinstated. noLocationData withdrew
them because no freely redistributable per-region occurrence data existed at useful granularity.
It does now: GBIF serves eBird and iNaturalist under CC BY 4.0 with GADM regions as a queryable
field, recorded as the `gbif` entry in [imports.md](imports.md#gbif). A species is listed in a
region when it reaches one record per million of that region's own records. Region and family
filters compose, and an empty intersection is a real outcome, distinct from nothing being
selected.

### Why

From the commit that made the change. Each region is asked what bird names were reported there,
faceting on `verbatimScientificName` rather than `speciesKey`: the `speciesKey` facet drops names
GBIF's backbone cannot resolve, and reports no records for Cooper's Hawk. iNaturalist is queried
alongside eBird because eBird withholds conservation-sensitive species outright; Spotted Owl is
placed in eight western states from iNaturalist alone. Membership is relative, not a fixed count,
because eBird doubled in five years and California's 114 million records cannot mean what
Hawaii's 2.5 million do; deciding per region also keeps endemics a continental cut would drop. The
first run placed 833 of 839 species across 69 regions. Two gaps it found: GADM files the
Northwestern Hawaiian Islands under the Minor Outlying Islands, so those were added for Laysan Duck
and Laysan Finch; and six pelagic species fall in no land region, which the requirement states so
they are not filtered away silently.

Later decisions built on the reinstated filter (regionsTwoWays, seaShareDiffers, monthPresence,
countEbirdFirst, belowBarCounted) without this entry existing; it records what they assumed.

---

## publicPersonalDeployment

**BirdView is served publicly at birdview.site, as Sameer's personal, non-commercial site**

Date: 2026-10-22
Recorded: 2026-10-23, from the launch plan in `deploymentPlumbing` (commit `aeb08a5`), where it was first written down
Decided by: Sameer
Status: accepted
Affects: [tech.localNowPublicLater](local/tech/posture.md#localnowpubliclater), [tech.servesOnLocalhost](local/tech/delivery.md#servesonlocalhost)

### Decided

BirdView is deployed publicly at `birdview.site`. The domain is Sameer's, the site is served at his
expense from his Google Cloud project on Cloud Run, and it is deployed from his one-way mirror,
`github.com/Sajmani/birdview`, while this repository stays the source of truth. It is a personal
project, not a service this organization operates. It stays non-commercial.

[tech.localNowPublicLater](local/tech/posture.md#localnowpubliclater)'s value moves from `local` to public, keeping its never-commercial clause
unchanged. [tech.servesOnLocalhost](local/tech/delivery.md#servesonlocalhost) keeps `localhost:8080` as the default, and the deployment sets
`-addr` instead.

### Why

As the launch plan records it: `birdview.site` belongs to Sameer personally and is served at his
own expense from his own clone, so the site is a personal project rather than something this
organization operates, and the Apache 2.0 license permits that. Nothing further was given.

What made it possible was already recorded: [tech.localNowPublicLater](local/tech/posture.md#localnowpubliclater) kept every source's obligations
against the public, non-commercial case so that going public "needs no licensing re-litigation",
and since [scopeFromRecurrence](#scopefromrecurrence) and [avilistReplacesAba](#avilistreplacesaba) nothing all-rights-reserved is read or served.

Keeping [tech.servesOnLocalhost](local/tech/delivery.md#servesonlocalhost)'s default was drafted by an agent on 2026-10-23, with this reason, for Sameer to
approve: a default is what a local run gets without anyone choosing, and a server listening on
every interface by default would expose a developer's machine to its network, so the deployment
states its address instead, `-addr 0.0.0.0:8080`, because Cloud Run must reach it.
