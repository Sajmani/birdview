# Decisions

Append-only. Entries are never rewritten; a decision that turns out badly is superseded by a new
entry, not edited.

Each record states the subject, the conflicting requirements, the kind of conflict, the options
considered, the decision, who made it, and the resulting status of everything it touched.

---

## CR-001 — No open source supplies both body length and wingspan

**Subject:** `measurement.provenance`
**Kind:** Over-constraint — no value satisfies every applicable constraint
**Date:** 2026-09-20 · **Decided by:** Sameer

### The conflict

Three requirements could not hold at once:

1. Lengths and wingspans shown to the user are real published measurements, not model output
   (owner's stated position, after rejecting the earlier design's invented allometry).
2. The dataset covers the regularly occurring ABA avifauna — 841 species (P-001).
3. Inputs are freely licensed and machine-readable (T-004, T-021).

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

The conditional matters: this only satisfies (1) because **P-061 makes the labeling
user-visible**. Without P-061 the decision degenerates into presenting modeled numbers as data,
which is the thing that was rejected. P-061 is therefore not a nice-to-have and must not be
dropped as a simplification.

### Resulting status

- **P-060** Active, as amended.
- **P-061** Active, and load-bearing for this decision.
- Honest expectation: most v1 lengths will be estimates. Ricklefs covers passerines only, so
  roughly half of the 841 species have no measured length available at any price.

---

## CR-002 — Ricklefs 2017 cannot be used; its terms prohibit it *(resolved)*

**Subject:** `dataset.length.source`
**Kind:** Cross-source — a mandatory source's terms versus a local requirement
**Date raised:** 2026-09-20 · **Resolved:** 2026-09-21 · **Decided by:** Sameer

### Resolution

**The source is withdrawn. BirdView will not use Ricklefs 2017.** Every total length in v1 is
`modeled`, tagged as such under P-062 and marked on the tile under P-061.

The terms were finally read — the maintainer printed the page from a browser, since the site 403s
automated requests — and four clauses each independently prohibit what BirdView would do. They are
transcribed with quotations in
[sources/wiley/requirements.md](sources/wiley/requirements.md). In brief:

| Clause | Effect |
| --- | --- |
| Rules of Use B | No storing material "in a retrieval system", no "systematic distribution" — and expressly "whether or not for commercial or non-profit use or for a fee or free of charge" |
| Rules of Use C | No integrating the material "with other material or otherwise creat[ing] derivative works" — which is exactly the AVONET join |
| Rules of Use G | No "automated means of any kind" to access the site, which forecloses the fetch-at-build mechanism of T-004 |
| Rules of Use F | No incorporation "into any institutional or other repository under any circumstances" |

The decisive detail is in B. This project's non-commercial posture (T-011) has unlocked other
sources — Macaulay embedding, iNaturalist's CC BY-NC corpus — and it was the natural hope here
too. The clause forecloses it in as many words.

### Consequence

Roughly half the dataset would have had a measured length; now none does. That is a real quality
loss and it is worth naming rather than absorbing quietly. What makes it survivable is P-062:
before per-metric provenance existed, an all-modeled length column meant presenting synthetic
numbers as data. Now it means a length column honestly labeled `modeled` throughout, with the
model named in the detail view.

### Still worth doing

**Ask Ricklefs or ESA.** These are platform terms accepted by using Wiley Online Library; they are
not the last word on what the author or the society may permit for the data itself. A direct grant
would come with none of the above restrictions.

**Deferred until BirdView runs**, by decision of 2026-09-21: a demo that shows the app crediting
its sources and labeling every modeled value as modeled makes a far better case than a description
of an intention. See [Work arising](#why-the-permission-asks-wait-for-a-demo). This is an upside
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

Per [process.md](process.md#some-sources-refuse-to-be-fetched), the gap is left unfilled rather
than guessed at. Specifically *not* relied upon: that the `RS-eco/traitdata` R package
redistributes the data under GPL-3 (a package license does not establish the data's license), and
the argument that measurements are uncopyrightable facts (a legal judgment nobody here is
qualified to make).

**Three exits, none yet taken:**

1. A human reads the Wiley terms and records what they permit.
2. Ask Ricklefs or ESA for permission for non-commercial open-source use.
3. Drop the measured-length half of P-060, making every length an estimate.

Until one is taken, the measured-length branch of P-060 is **Deferred (owner: Sameer)**. The
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

### Reframed by P-062 *(2026-09-21)*

This conflict was raised when provenance was binary — a value was either measured or the app was
misrepresenting it. P-062 removed that premise. If no measured length is ever obtained, every
length is labeled `modeled`, and the product is honest and complete.

CR-002 is therefore **not a blocker** and should not be treated as one. It governs how good v1's
length data is, not whether v1 can ship.

---

## CR-003 — Location filtering has no redistributable data source

**Subject:** `filter.location`
**Kind:** Over-constraint
**Date:** 2026-09-20 · **Decided by:** Sameer

The owner asked for filtering by location. The species-list source chosen for P-001 carries no
geography, and no freely redistributable per-region occurrence data was found at useful
granularity. eBird's region species-list endpoint would serve, but requires an API key, and
`api.ebird.org/robots.txt` is `Disallow: /` with the API terms page returning 403 — so what it
permits is unestablished.

**Decision:** withdraw P-011 and P-012 from v1 rather than defer them, and revisit with a range
source. Withdrawing rather than deferring is deliberate: a `Deferred` entry needs an owner and a
date, and there is no date until a source exists.

**Resulting status:** P-011 Withdrawn · P-012 Withdrawn · family is the only filter.

---

## CR-004 — The neck/body split has no measured basis

**Subject:** `diagram.neck_ratio`
**Kind:** Over-constraint, and a consistency question against CR-001
**Date:** 2026-09-20 · **Decided by:** Sameer

With a total length in hand, AVONET supplies culmen and tail directly and head follows from bill
depth — but nothing measures where the neck ends and the body begins. The earlier design used
hand-assigned per-family neck ratios, which is the same class of invented number that CR-001
rejected for lengths.

| Option | Consequence |
| --- | --- |
| Per-family ratio | Herons look like herons; the split is a guess with no source |
| One fixed ratio | No invented numbers; every bird the same shape, largely defeating G-003 |
| Find a source | No per-species neck-length data is believed to exist at this coverage |
| Merge neck into body | Drops a segment the owner explicitly asked to keep |

**Decision:** keep the per-family ratio, **declared as a visual design parameter** carrying no
measurement claim and never surfaced as data (P-045).

This is consistent with CR-001 rather than in tension with it, on one condition: CR-001 governs
*numbers presented to the user as measurements*, and this governs *pixels in a schematic*. The
distinction holds only while the ratio stays out of the dataset and out of the UI's numbers. If a
neck length ever appears as a value, this decision is void and must be reopened.

---

## CR-005 — Photographs withdrawn from v1

**Subject:** `presentation.mode`
**Kind:** Scope
**Date:** 2026-09-20 · **Decided by:** Sameer

Photographs were researched in full before being scoped out, and the findings are recorded here so
the work is not repeated:

- **Wikimedia Commons** is the recommended primary. Its licensing policy guarantees commercial-use
  compatibility, the Action API returns structured license metadata (`extmetadata`) sufficient to
  build a credit line automatically, no API key is needed, caching is permitted and preferred over
  hot-linking, and per-species `in flight` categories exist. There is **no "perched" equivalent**
  — 30 category hits in total, mostly cats — so in-flight is a positive filter and perched is the
  residual default.
- **iNaturalist** is a viable fallback *because BirdView is non-commercial* (T-011), which makes
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

**Decision:** ship v1 with diagrams only. P-050 through P-053 are withdrawn, not abandoned.

**Resulting status:** P-030 reduced to a single mode · P-050–P-053 Withdrawn.

---

## CR-006 — ABA species with no AVONET row of their own *(resolved)*

**Subject:** `dataset.join.key`
**Kind:** Derived — no pair of requirements contradicts; the conflict appears only when P-006's
mapping meets the provenance rule
**Date raised:** 2026-09-21 · **Resolved:** 2026-09-21 · **Decided by:** Sameer

### Resolution

Two decisions, together:

1. **Read each species from whichever AVONET taxonomy carries it as its own taxon**, recording the
   sheet and the filed name in its provenance record. This reduces the problem from 9 species to
   6.
2. **For the remaining 6, borrow the parent taxon's traits and label them `near-match`** under
   P-062. All 841 species stay in the dataset, and every borrowed value says whose measurement it
   actually is.

A value read from another sheet, or found under a pre-revision name, is **`measured`** — the
specimens are the right birds and only the filing changed. `near-match` is reserved for a value
measured on a genuinely different taxon.

This mirrors CR-001 exactly: mixed authority is acceptable when, and only when, each value is
labeled. P-062 now provides the vocabulary both decisions needed, and the detail view of P-034 /
P-063 is where it reaches the user.

**Resulting status:** P-006 Active as amended · P-062 Active · P-063 Active · the six near-match
species are `Larus vegae`, `Numenius hudsonicus`, `Sturnella lilianae`, `Vireo swainsoni`,
`Sula brewsteri`, `Anthus japonicus`.

**Carried forward, not closed by this:** the mass caveat below, and the four renames-in-disguise
at the end, still need a human taxonomic check before the mapping table is written.

---

### Original analysis

### What was measured

Joining ABA 8.19 codes 1–3 against `AVONET2_eBird` on normalized binomial: **819 of 841 = 97.38%**.
AVONET carries the 2022-era eBird taxonomy, so four years of revision separate the two sources.
Full classification is in [sources/aba/PROVENANCE.md](sources/aba/PROVENANCE.md#taxonomy).

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

It also collides with [CR-001](#cr-001). P-060 divides every value into *measured* or *estimated*;
a trait borrowed from a pre-split parent is neither. It is a real measurement of a different
taxon, which is a third category the provenance model does not currently have.

### Options

| Option | Effect |
| --- | --- |
| Borrow parent traits, tag as a third provenance class | All 841 species present; honest labeling; P-060/P-061 grow a category |
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
(P-020) would still show the pair as identical** — a smaller version of the original problem,
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

## CR-007 — The neck length is surfaced as a metric, superseding CR-004's condition

**Subject:** `diagram.neck_ratio`
**Kind:** Direct — a later decision reverses a condition an earlier one set
**Date:** 2026-09-21 · **Decided by:** Sameer

### What changed

[CR-004](#cr-004) kept the per-family neck ratio on one condition, stated plainly: *"If a neck
length ever appears as a value, this decision is void and must be reopened."* The owner has now
decided it should appear — in the detail view, alongside every other segment, with provenance
naming the specific modeling function and the ratio used for that bird's family.

This record reopens it, as CR-004 required.

### Why it resolves the other way now

CR-004 was decided before P-062 existed. At the time the choice lay between hiding the number and
presenting it as though it were measured, because there was no way to say "drawing convention" in
the data. Confining it to pixels was the least-bad option available.

Per-metric provenance removes the dilemma. A neck length labeled `kind: modeled` with
`model: "per-family neck ratio, Ardeidae = 0.35"` is not passing itself off as a measurement; it
states exactly what it is and what produced it.

And the concealment was never worth much. **The user sees the neck either way** — it is drawn, at
a length, in every diagram. Withholding the number withheld the explanation, not the claim. Naming
the function is strictly more honest than drawing an unexplained proportion.

This is the third time P-062 has dissolved a constraint that existed only because provenance was
binary, after [CR-001](#cr-001) on modeled lengths and [CR-006](#cr-006) on split taxa.

### Resulting status

- **P-045** amended: the ratio is a named per-family model, and the neck length is presented with
  its function named. No longer "never surfaced as data".
- **P-063** amended: every metric, including all five segments, is shown with its provenance.
- **P-034** amended: the detail view is a larger version of the same tile, plus per-segment
  measurements and provenance.
- **P-031** amended: the tile header carries scientific name as well as common name and badge.
- **AC-035** inverted: it asserted the neck length never leaves the drawing layer; it now asserts
  the neck length *is* serialized, with a `model` naming the family ratio.
- **CR-004** stands as history. Its reasoning was right for the constraints of its time; only its
  condition is superseded.

### Consequence worth watching

Surfacing the segments makes the composition of total length visible. A reader will see a
**measured** bill inside a **modeled** total and may reasonably ask how both are true. The answer
is that total length is the sum of five segments, two measured and three modeled — now stated in
P-060 so the detail view can say it rather than leave the reader to infer it.

---

## CR-008 — Fit the wingspan model rather than invent one

**Subject:** `measurement.wingspan.model`
**Kind:** Over-constraint, reopened by evidence
**Date:** 2026-09-22 · **Decided by:** Sameer

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
`wingspan_predicted_*`. But [sources.md](sources.md#fox-wingspan) excludes those columns, so
ingesting them would have needed an amendment — and spec2code does not get to reinterpret an
exclusion.

The resolution avoids the question. BirdView fits **its own** regression on the empirical column
it already adopts, which is modeling from data the project may use rather than ingesting someone
else's estimates as data. The exclusion stands unchanged.

### Validation, which is what made the choice safe

Fitted this way and asked for the same 9,194 species Fox predicted, BirdView's model agrees with
theirs to a **mean of 1.32%** and a median under 1%, with 1.4% of species differing by more than
10% — mostly extreme-aspect seabirds. It is, in effect, their model reconstructed.

That comparison is now a standing check (AC-045), so a regression in the fitting code shows up as
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

- **P-060** amended: the modeled-wingspan branch names the fitted regression.
- **P-062** amended: a `model` states the computation *and its measured error* where one exists.
- **AC-045** added: agreement with Fox's published predictions stays within 2% mean.
- `morphology.Wingspan`, the invented model, is deleted rather than left for someone to reach for.
- Coverage is unchanged at 301 measured, 7 near-match, 533 modeled; only the modeled values move.

---

## CR-009 — The exposed leg is modeled from the tarsus, reversing P-040

**Subject:** `diagram.legs`
**Kind:** Direct — a later decision reverses a rejection an earlier one recorded
**Date:** 2026-09-22 · **Decided by:** Sameer

### What changed

[P-040](product.md#p-040) drew the legs as the measured tarsus and explicitly rejected the
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

This is the same shape of argument as [CR-007](#cr-007), and the fourth time P-062 has dissolved a
constraint that existed only because there was no way to say "this is a model" in the data.

### What it costs, stated plainly

The diagram used to have three parts that were not modeled — bill, tail and tarsus. It now has
two. The factors are **invented**: no source measures exposed leg length, and unlike the wingspan
model of [CR-008](#cr-008) there is no empirical column to fit against or validate on. They are a
drawing parameter of the same kind as the neck ratio of P-045, and carry the same obligation —
named in the provenance of every value they produce, never presented as measurement.

### Consequence the renderer had to absorb

AC-066 recorded an invariant: feet never fall below the tail tip, which is what let the renderer
center a diagram on its bill-to-tail extent. **The modeled leg makes that false** for 11 of the
841 species — an American Flamingo by 30% of its own length, then Whooping Crane, Reddish Egret
and Black-necked Stilt. That is correct anatomy, so the renderer accommodates it: `VerticalExtent`
sizes each section's scale on the drawn extent rather than on total length, and the template
centers on `DrawnHeight`. The invariant is not merely dropped — its replacement is asserted in
both directions, so a bird standing below its own tail is still confined to the wader families.

Drawing order changed with it. The body stick is painted **after** the legs, so the legs pass
behind the body rather than crossing it and the junction mark.

### Resulting status

- **P-040** amended: legs are the modeled exposed leg, not the measured tarsus; the rejection
  above is superseded; the draw order is stated.
- **AC-066** rewritten: the feet-above-tail assertion is inverted into a drawn-extent invariant,
  and the leg model gains dataset-wide checks.
- **P-062** unchanged but newly load-bearing: `Species.Leg` is modeled, `Species.Tarsus` measured,
  and both are serialized and shown.
- The tile width reference stays the wingspan alone. A pair of legs never spreads wider than the
  wings — the widest case in the dataset reaches 35% of it — so that is asserted over the whole
  dataset rather than guarded by a branch nothing would ever exercise.

**Superseded in full by [CR-010](#cr-010).** The diagram is a bird in flight; legs are gone.

---

## CR-010 — The diagram depicts a bird in flight, and legs are removed entirely

**Subject:** `diagram.posture`, `diagram.legs`
**Kind:** Premise settled, and a scope cut following from it — supersedes [CR-009](#cr-009)
**Date:** 2026-09-23 · **Decided by:** Sameer

### The premise, which was never written down

The diagram has always been ambiguous about what it depicts. A vertical body axis reads as a
standing bird; a spread wing crossbar reads as a flying one. Nothing in the spec said which, and
every leg decision so far — the rake angle, the Z-shape proposal, CR-009 itself — was really an
argument about that unstated premise.

It is now stated: **the diagram is a top-down view of a bird in flight.** That reading is what the
crossbar already assumed, and it makes P-043 better founded than it was — seen from above, wingspan
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
| The modeled leg (CR-009) | 14 of 110 |

The true figure is far higher — every large heron, egret, crane, stork and ibis does it. Checking
the textbook field marks against the modeled leg: Whooping Crane +208 mm, Sandhill Crane +115,
Wood Stork +70, Black-necked Stilt +63, Great Egret +55 — but **White Ibis −14 mm**, a bird whose
trailing legs are unmistakable in flight. The sign comes out right for the extremes and wrong in
the middle, which is what systematic underprediction looks like. The plausible numbers were luck,
not model.

### What is lost, stated plainly

The tarsus was a **measured** trait spanning 115×, from a Rufous Hummingbird's 2.5 mm to a Whooping
Crane's 289 mm, and CR-009's rationale for drawing it at all still reads well. That range is real
and it is now not shown anywhere. The honest accounting is that it was only ever *visible* because
the diagram tacitly depicted a standing bird — the posture the project has now decided against. A
measurement that can only be displayed by implying a false posture is not a measurement the
diagram can carry.

Considered and rejected: a fixed-length, not-to-scale marker past the tail for the trailing-leg
families. It would capture a genuine field mark, but it puts an unscaled element into a diagram
whose entire premise is that everything in it is to scale. That cost is higher than the gain.

### Resulting status

- **P-040** amended: five segments and a crossbar. No legs. The flight premise is stated there.
- **P-042** reverted: the scale reference is the largest total length again, not the drawn extent.
- **P-002/P-003** amended: `Tarsus.Length` is no longer ingested, leaving five AVONET traits.
- **AC-066** withdrawn — the requirement it verified no longer exists.
- **AC-012** amended: five traits, not six.
- Deleted from the code: `morphology.LegFactors`, `DefaultLegFactor`, `LegLength`;
  `dataset.Species.Leg` and `.Tarsus`; `avonet.Traits.Tarsus`; `view.VerticalExtent`,
  `Geometry.DrawnHeight`, `LegCos`, `legRake`, `legSin` and the leg coordinates; the leg strokes,
  legend entry and CSS.
- CR-009's renderer machinery goes with it. That work was correct for its premise and survived
  about a day; the premise, not the code, was the thing that was wrong.

### Left open

The **neck** is still modeled on a standing posture, and under the flight premise it is wrong for
the families that fold it — Ardeidae and Pelecanidae fly with the head drawn back onto the
shoulders, while our Ardeidae neck ratio of 0.85 is the second-highest in the table. Fixing it
collides with P-041, since a folded-neck heron's drawn extent is genuinely shorter than its
catalogued total length. **Resolved by [CR-011](#cr-011).**

---

## CR-011 — A folded neck is drawn as an S; length as data is untouched

**Subject:** `diagram.neck.posture`
**Kind:** Open item from [CR-010](#cr-010), resolved
**Date:** 2026-09-23 · **Decided by:** Sameer

### What changed

For the families that fold the neck in flight — **Ardeidae** and **Pelecanidae** — the neck is
drawn as an **S of two semicircular arcs** rather than a straight stick. Everything else about the
neck is unchanged.

### The property that makes it cheap

The S is **two semicircles of radius r**. Its path length is `2πr` and its vertical extent is `4r`,
so a neck of length L is drawn with `r = L/2π`, the stroke is **exactly L long at the diagram's
scale**, and it occupies `2L/π ≈ 63.7%` of L vertically.

That ratio is a **consequence of the shape, not a parameter**. No new number enters the system —
which is exactly what separates this from the leg factors of CR-009, and why this change needs no
provenance record: like `headAspect`, it produces no metric. Nothing is rescaled, nothing is
hidden, and the ink you see is still the neck at the same pixels-per-millimeter as every other
part of the bird. It is the same neck, curved.

Tightening it further would mean arcs of more than 180°, and *that* would introduce a chosen
number. If 63.7% ever proves too loose to read as a fold, the cost of the next step is a parameter.

### What it costs

For these 14 species the **drawn vertical extent is no longer the total length**. P-041's rationale
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

- **P-040** amended: the neck is a stick or an S depending on the family's flight posture.
- **P-041** amended: the rationale is qualified — drawn extent equals total length except where the
  neck is folded, and the sum over the *metrics* is unaffected.
- **AC-067** added.
- No change to P-045, the neck-ratio model, or to any serialized value.

---

## CR-012 — Adopting iNaturalist for a species link, and what it is allowed to be

**Subject:** `presentation.detail_view.external_link`
**Kind:** Source adoption, with three sub-decisions
**Date:** 2026-09-23 · **Decided by:** Sameer

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
[CR-002](#cr-002).

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

`birdview-ingest` currently makes **no network request at all**, and [T-002](product.md) keeps the
server offline at runtime while `tech.md` binds tests to make none either. Resolving 841 names at
the one-request-per-second that `inat-api/R1` asks for would add roughly fifteen minutes to every
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
reproduced, so no contributor's license is engaged. [T-002](tech.md) forbids the embedding
independently, which is a useful accident: two unrelated rules point the same way.

### Resulting status

- **`sources.md`**: `inat-api` (governing) and `inat-terms` (mandatory) adopted, both vendored by
  copy from the maintainer's `birdsync` repository and verified byte-for-byte against the SHA-256
  hashes recorded there. Neither could have been fetched — `inaturalist.org/pages/*` answers an
  agent with a 403 and a JavaScript challenge.
- **P-064** added: the detail view links to iNaturalist when, and only when, the match is
  confident.
- **T-022** added: the mapping tool paces its requests and identifies itself.
- **AC-068** added.
- The existing **T-006** (identify the client) and **AC-052** already discharge `inat-api/R4`,
  which is why that requirement needed no new work.

---

## CR-013 — Contrast is checked for all text, at two floors rather than one

**Subject:** `text.contrast.ratio`
**Kind:** Source scope widened, after a defect the existing scope could not see
**Date:** 2026-09-23 · **Decided by:** Sameer

### What prompted it

The iNaturalist link of [CR-012](#cr-012) shipped with no `color` rule, so it fell back to the
browser default: `#0000EE` on a `#0f172a` background, which is **1.90:1**, and `#551A8B` once
visited, which is **1.62:1**. Effectively invisible. A reader spotted it immediately.

Nothing caught it, and nothing could have. [`wcag`](sources.md) was adopted with its scope set to
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
| Badges and chips | 7:1 against their own text (T-010, unchanged) | the colors are generated |
| **Links** | **7:1** against the page background | interactive, and the palette clears it easily at 8.33:1 |
| **All other text** | **4.5:1** against both surfaces | AA; `--dim` clears this comfortably |

Links get the strict floor because nothing makes it hard: `--accent` is already 8.33:1 on the page
background, and the new visited color is 9.67:1. The 7:1 link check is applied against the page
background specifically, having verified that no bare `<a>` sits inside `.help` or `.controls`,
whose surface is `--panel`. If one ever does, the 4.5:1 rule still governs it and the check still
runs.

### The second assertion, which is the one that matters

A contrast check over *declared* colors is blind to the colour nobody declared — which is precisely
how this bug shipped. So AC-070 also asserts the stylesheet carries top-level `a` and `a:visited`
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
- **T-023** added.
- **AC-070** added, watched failing three ways, including against the original defect.
- `--accent-visited` added; link colors moved to global `a` / `a:visited` rules; `.tile`, `.btn` and
  `.chip` given explicit `:visited` rules so the global rule cannot recolor them.
- Still excluded, and still claimed nowhere: keyboard access, focus order, timing, reading level.
  **BirdView claims no WCAG conformance level.**

**The two-floor scheme was superseded the next day by [CR-014](#cr-014).**

---

## CR-014 — One contrast floor, superseding CR-013's two

**Subject:** `text.contrast.ratio`
**Kind:** Direct — a later decision reverses a tradeoff an earlier one accepted
**Date:** 2026-09-23 · **Decided by:** Sameer

### What changed

[CR-013](#cr-013) set two floors: 7:1 for links and badges, 4.5:1 for everything else. There is now
one, **7:1**, and `--dim` and `--accent` moved to clear it.

| | was | now | on page bg | on panel |
| --- | --- | --- | --- | --- |
| `--dim` | `#94a3b8` | `#adbbcc` | 6.96 → **9.14** | 5.71 → **7.49** |
| `--accent` | `#38bdf8` | `#5ccbfa` | 8.33 → **9.67** | 6.83 → **7.93** |

### Why it resolves the other way now

CR-013's reasoning was explicit and, at the time, sound: raising the floor meant repainting the
interface's secondary text, which looked disproportionate to the one broken link that prompted the
check. So it recorded the gap instead of closing it.

What that reasoning was missing is that the gap was not hypothetical. **The owner independently
reported the dim text as hard to read** — scientific names, provenance detail, section counts, all
at 5.71:1 on a panel. A measured shortfall that a user has actually noticed is no longer a tradeoff
worth defending; it is the same defect as the link, one tier less severe.

Cheap, too, in the end. Two variables moved and nothing else did. `--accent` as the `.btn.on` fill
still clears its own dark text at 8.25:1, and `--accent-visited` was already at 7.92:1 on a panel.
The repaint CR-013 was avoiding turned out to be two hex values.

### What this says about recording a known gap

CR-013 did the right thing by writing the shortfall down with its numbers rather than rounding up
to a claim BirdView could not support. That is what made this decision a one-line lookup rather
than a rediscovery — the exact figures were already on the page.

But a recorded gap is a deferral, not a resolution, and this one lasted a day. Worth remembering
the next time an exception looks cheaper than a fix.

### Resulting status

- **T-023** amended: one floor, 7:1, links and text alike, against both surfaces.
- **AC-070** amended: the measured assertion is now a single floor, and the watched-failing list
  gains the two restored colors.
- The test's two-tier machinery — a separate link floor, and the selector tokenizer that decided
  which rules were links — is deleted rather than left switched off. It is recoverable from history
  if the floors ever diverge again.
- **`sources.md`** unchanged: both WCAG criteria were already adopted, and only which one governs
  ordinary text has moved.
- Still claimed nowhere: any WCAG conformance level.

---

## CR-015 — The drawing carries provenance, and the crossbar splits to do it

**Subject:** `diagram.provenance`
**Kind:** New capability, with one reversal and two exceptions worth defending
**Date:** 2026-09-24 · **Decided by:** Sameer

### What changed

Stroke style now encodes provenance: solid measured, dashed modeled, head hollow. The wing crossbar
is split at the wing chord, so its measured outer portions stay solid even when the span is
modeled. Both pages document the code. [P-065](product.md) states it.

### The reversal: what dash used to mean

Dash previously meant *"this is the tail, not the body"*. It now means *modeled*, and the tail —
being measured — becomes **solid**. The tail is still distinguishable: it keeps its own colour and
the junction mark, which was always doing most of the work. This is a strictly better use of the
channel, since tail-versus-body was a distinction the drawing already made twice.

### The crossbar split, and an argument that was wrong on the first pass

The initial objection to splitting was this: `span = a × chord^b` is multiplicative
([CR-008](#cr-008)), so the span does not decompose into a measured part and a modeled remainder,
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

- **P-065** added; **P-040** amended, since the tail's stroke style now means something else.
- **AC-072** added.
- `dataset.Species.WingChord` added: AVONET's `Wing.Length`, previously consumed by the wingspan
  model and discarded. Note this is the **ornithological** chord, carpal joint to primary tip, not
  the aerodynamic chord of leading edge to trailing edge — the two share a name, and only the first
  can be drawn along a crossbar.
- The `?prov=1` prototype flag and its head/body variant switches are **removed**; the rendering is
  unconditional.

---

## Work arising

Not conflicts; obligations this bundle creates that nothing yet discharges.

| Item | Owner | Gate |
| --- | --- | --- |
| Retrieve, hash and vendor `fox-wingspan`; write its `PROVENANCE.md` | — | before ingest reads it |
| Write `acceptance.md` — Gate 2, before any code | — | now |
| Ask Ricklefs or ESA for permission to use the measured-length data | Sameer | **after a working demo** |
| Ask ABA for written permission to redistribute the checklist | Sameer | **after a working demo** |
| ~~Decide the neck's flight posture~~ — done, [CR-011](#cr-011) | Sameer | closed 2026-09-23 |
| Assess whether a perched/standing toggle is achievable at any honest accuracy ([CR-011](#cr-011)) | — | before it is offered to users |
| Investigate cervical vertebra counts as an anchor for the neck model (P-045) | — | opportunistic; the biggest available win on the weakest model |

### Cervical vertebra counts, and why they are worth chasing

*Recorded 2026-09-23.* The per-family neck ratios of P-045 are the largest unforced assumption in
the project: invented, unvalidated, and visibly driving the silhouette of every long-necked family.
Three families have now been corrected by eye — Aramidae and Pelecanidae after a reader noticed a
short-necked Limpkin, Sulidae on review — which is a sign the table is guesswork with a slow error
signal, not that it is converging.

**Cervical vertebra count is the most promising anchor available.** Birds run from roughly 11 to 25
cervicals, against a mammalian fixed seven, and the count tracks neck length closely: swans sit at
the top of the range, passerines near the bottom, galliforms between. Unlike neck length it is a
discrete, published, comparable anatomical trait.

It would not give the ratio directly — ours is neck ÷ trunk, so it needs count × mean vertebra
length ÷ trunk — but it would give a defensible **ordering across all 91 families**. That is enough
to calibrate the table against the few anchors the project is most confident about, instead of
inventing each entry independently and waiting for someone to spot the next Limpkin.

Two things are unknown and should not be assumed: whether a dataset exists at BirdView's coverage,
and whether its license permits use. Both have sunk candidate sources before. Adoption would follow
the usual path — a `sources.md` entry and a compose-and-resolve pass.

### Sketch: one dash per vertebra

*Recorded 2026-09-23, owner's idea, not a decision.* If the counts existed, the neck could be drawn
as a **dashed line with one dash per cervical vertebra** — putting a measured, discrete count into
the one segment that currently carries the least support.

What makes it more than decoration: two birds can reach the same neck length by different means. A
swan is near the top of the vertebral range; a flamingo has a comparably long neck built from
fewer, longer vertebrae. That distinction is real, and the diagram cannot presently show it at all.

Three obstacles, recorded so the idea is picked up with its problems attached rather than
rediscovered:

- **The tail is already dashed**, to distinguish it from the body. Dash would then mean two
  different things on one figure. The tail would need another treatment; there are options.
- **It cannot work on the index tiles.** The neck is a fixed fraction of a bird that fills its tile,
  so its drawn length is roughly scale-independent: about 31 px for the longest-necked species in
  the set, and under 2 px for a median bird. Twenty dashes in 31 px is a grey blur.
- **On the detail view it is marginal**, at 1.9× the tile: a heron's neck is about 62 px, giving
  roughly 3 px per vertebra.

That last figure is the interesting one. Nobody will count 24 dashes, but 3 px is enough to read as
**texture** — finely divided against coarsely divided — which is exactly the swan-versus-flamingo
distinction the idea exists to surface. Comparability rather than countability may be the right
goal, and would be the thing to test.

### Why the perched toggle is not simply the other half of CR-011

*Recorded 2026-09-23.* Offering "show them perched" sounds like a rendering switch and is not one.
A perched bird needs two things BirdView does not have.

**Legs.** [CR-010](#cr-010) removed them because a flight silhouette does not show them; a perched
silhouette is mostly leg. Restoring them restores every problem CR-010 recorded — the femur, the
toes and the hip position are unmeasured, and the tarsus alone understates the visible leg most in
the birds whose legs define them. A perched view would need the exposed-leg model that CR-009
invented and CR-010 deleted, and it would still have no source to validate against.

**A second neck model.** The current ratios are a standing-posture model used, until CR-011, in a
flight drawing. A perched view needs the perched extension for every family, which is a different
invented table again — a perched heron's neck is neither folded flat nor fully extended, and where
it sits between those is not published at this coverage.

So the toggle is not a view of the same data from another angle. It is a second set of invented
per-family parameters with no empirical check, which is the thing this project has twice decided
not to buy. Worth revisiting only if a source appears that measures posture, not just skins.

### Why the permission asks wait for a demo

*Recorded 2026-09-21.* Both asks are deferred until BirdView runs and can be shown.

A request that describes an intended use is weaker than one that demonstrates it. Showing a
rights-holder the actual application — what it does with their data, how the data is credited,
that it is non-commercial and that every value carries visible provenance — answers most of what
they would otherwise have to take on trust. It also lets them see that BirdView labels a modeled
length as modeled, which is the strongest available evidence that their measured data would be
handled honestly.

Neither ask blocks v1, which is what makes waiting free:

- **Ricklefs / ESA** would *add* capability. Every length is currently `modeled`
  ([CR-002](#cr-002)); a grant would make roughly half of them measured. Nothing breaks without
  it.
- **ABA** would *simplify the build*. Fetch-at-build already works under T-004, so a grant would
  only remove the network step and let T-005 relax. Nothing breaks without it either.

Sending them together after a demo also means one conversation per organization rather than two
rounds of the same explanation.

**The cost of waiting**, recorded so it is a choice rather than a drift: if ESA grants permission
late, the length column changes from wholly `modeled` to half `measured`, which will move diagrams
and re-sort sections. Builders should treat the provenance model (P-062) as the thing that stays
fixed and the mix of kinds as the thing that moves.
