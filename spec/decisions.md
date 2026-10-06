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

## CR-016 — Total iNaturalist coverage, reached by widening the evidence rather than lowering the bar

**Subject:** `presentation.detail_view.external_link`
**Kind:** Requirement tightened, after diagnosis showed the shortfall was artifactual
**Date:** 2026-09-27 · **Decided by:** Sameer

### What prompted it

[CR-012](#cr-012) shipped with 811 of 841 species linked and treated the other 30 as correct
absences: species iNaturalist had not confidently matched, mostly assumed to be splits it had not
followed. The owner asked for **100% coverage**.

That request sits directly against CR-012's central rule, that a wrong bird is worse than no link.
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

- **P-064** amended: both names must still agree, but each is compared more thoroughly.
- **AC-068** extended: a case per bucket, plus the assertion that a differing genus is refused when
  the common name does not agree.
- Coverage is now a **required** 841 of 841, asserted over the built dataset rather than reported.
  A species that stops resolving is a fault, not an absence.
- The normalizer drops separators rather than collapsing them, because an okina sits *inside* a
  word and collapsing it to a space merely moves the mismatch: `Hawaiʻi` became `hawai i`.

---

## CR-017 — Per-section becomes the default, and the legend becomes a diagram

**Subject:** `diagram.scale.mode`, `presentation.help`
**Kind:** Direct — a later decision reverses a default an earlier one argued for
**Date:** 2026-09-27 · **Decided by:** Sameer

### What changed

The scale default returns to **per-section**, listed first. The help's diagram section becomes a
real labeled diagram instead of a list of colour swatches, and the three sections are reordered:
the diagram, where the numbers come from, then the controls.

### Why the default moves, against P-042's own argument

P-042 argued the case for global plainly and it has not become wrong: *"the cross-family surprise
**is** the product… a reader could scroll the whole list without ever seeing that a Trumpeter Swan
dwarfs a kinglet."* That is still true, and global is still one click away and still named first
in the help text.

What changed is the diagram. Since that default was chosen it has acquired provenance in the stroke
style ([CR-015](#cr-015)), a split crossbar with chord ticks, and labels ([P-066](product.md)). At
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
"dashed, to tell it from the body", which stopped being true when CR-015 gave dash a different
meaning, and nothing caught it because the text and the stylesheet had no connection. A rendered
diagram cannot drift from the renderer.

### Resulting status

- **P-042** amended: per-section is the default and is listed first.
- **P-035** amended: three sections in a stated order, and the diagram section is a rendered,
  labeled diagram.
- **AC-031** and **AC-059** amended, and the tests that asserted the global default inverted.
- The help diagram is drawn from a real species with a **modeled** wingspan, so the dashed inner
  crossbar of P-065 is visible in the legend rather than only in the prose beside it.

---

## CR-018 — The compare URL contract changes when the index becomes the picker

**Subject:** `presentation.compare_view`, `presentation.pick_mode`
**Kind:** Direct — a later decision reverses two conditions an earlier one set
**Date:** 2026-10-01 · **Decided by:** Sameer

### What changed

Two behaviors settled days ago now resolve the other way, both because the same premise
disappeared. [P-068](product.md#p-068) gives the index a picker, which means the index now reads
the `species` parameter that the compare view writes.

**a. Leaving the compare view keeps the species list.** Commit `10e938d` removed it, and
[AC-076](acceptance.md#ac-076) asserted that no link back to the index may carry `species`.

**b. An unrecognised id is dropped and named, not a 404.** `/compare` refused the whole page; it
now renders the species it recognises and says which it could not.

### Why (a) reverses

The fix in `10e938d` was correct and its reasoning was sound: *"the index does not read `species`,
so the parameter is copied into every control link on that page and follows the reader through
every toggle afterwards."* A parameter that means nothing where it lands is litter, and litter
that propagates is worse.

That argument rests entirely on the index not reading it. P-068 makes the index read it — the
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

## CR-019 — Mass gets a shape after all: the cube of water

**Subject:** `presentation.detail_view.diagram_labels`
**Kind:** Direct — a later decision reverses one an earlier one argued for, on finding one of its
two reasons did not hold
**Date:** 2026-10-01 · **Decided by:** Sameer

### What changed

[P-066](product.md#p-066) said mass is drawn as a number **and not as a shape**, one day before
this. It now also carries a square in the lower-left corner: one face of the cube of water that
weighs what the bird weighs, drawn at the diagram's own scale.

### Why it reverses

P-066 gave two reasons and they are not equally good. The first — that a mass glyph puts a
second, arbitrary scale on a figure whose every other dimension shares one, which
[P-043](product.md#p-043) refuses — **does not apply to this glyph.** One gram of water is one
cubic centimetre by the historical definition of the gram, so ∛(mass) is a unit conversion and
not a model: it assumes nothing about the density of birds, and it is a real length that draws at
the diagram's own px/mm. The objection was to arbitrary scales, and this is not one. The earlier
text conflated the general case with the specific proposal and rejected both.

The second reason stands and is recorded in P-066 as a limitation rather than deleted: **the cube
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

*(Re-measured after [CR-021](#cr-021): the alcid end read 28–29% before heavy birds lengthened.
The comparison and its direction are unchanged.)*

Measured before it was built, at the detail view's geometry: 28 px smallest, 52 px median, 80 px
largest, in a drawing area 668 × 281. In the lower-left corner at an 8 px margin the crossbar
intrudes for **0 of 841 species**, the tightest clearance being 50 px on Western Grebe.

### The risk being accepted

A reader who takes the cubes as *the* mass comparison reads them wrong, by a factor of the cube
root. The mitigation is that the number is right beside it and the diagram key names what the
square is. If it proves misleading in use, the thing to remove is the square and not the number —
the number is what P-066 requires.

## CR-020 — The mass cube goes on the index tiles too

**Subject:** `presentation.tile`
**Kind:** Direct — a later decision reverses a scope an earlier one set
**Date:** 2026-10-01 · **Decided by:** Sameer

### What changed

[CR-019](#cr-019) put the cube of water on the annotated diagrams and kept it off the index
tiles, and [AC-078](acceptance.md#ac-078) asserted that exclusion. The tiles now carry it as
well, in the same lower-left corner, **without the mass label** — a tile carries no labels at all
([P-031](product.md#p-031)), and the mass is already in its footer.

### Why it reverses

The exclusion rested on two claims. One was crowding: *"the tile is dense by design and this is
annotated-diagram furniture, as P-066's labels already are."* That reasoning holds for **labels**,
which need a gutter a third of the width and a leader line to the part they name. It does not
transfer to the cube, which needs neither: it sits in a corner the drawing does not use, and
measured across the whole dataset it never reaches the crossbar — **0 overlaps in 841**, which is
the claim that matters and which [CR-021](#cr-021) re-verified after every length moved.

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
scale (P-042), so cubes on tiles are directly comparable down a whole section — and the thing the
cube shows, bulk against elongation, is exactly what a reader scanning a section of similar-length
birds cannot otherwise see. Confining it to the detail view meant it could only ever be read one
bird at a time, or two or three on a comparison page assembled by hand.

### What this does not change

The cube stays off the tile in no other way: no label, no leader, no footer row of its own. The
tile's contract is unchanged apart from the shape itself.

## CR-021 — The trunk's mass exponent is calibrated against skeletons, not invented

**Subject:** `diagram.trunk_model`
**Kind:** Direct — a measurement contradicts a constant the project chose by eye
**Date:** 2026-10-02 · **Decided by:** Sameer

### What changed

`BodyLength` modeled the trunk as `18.0 x mass^0.31`, an unvalidated constant carried since the
first ingest. It is now **`14.0329 x mass^0.37`**, and the exponent is pinned to published skeletal
measurement by [P-047](product.md#p-047) and [AC-080](acceptance.md#ac-080).

### Why

The question was raised by an Ostrich. Measuring what a worldwide scope would do to the models
produced a bird with a 360 mm head on a 119 mm neck and a total length of 1,539 mm against a real
2.1–2.75 m, which prompted the right follow-up: *is the model also wrong for the 841 we already
have?*

It is, and the evidence converged from three independent directions. Regressing log dimension on
log mass:

| dimension | n | exponent | 95% CI |
| --- | ---: | ---: | :---: |
| measured wingspan ([`fox-wingspan`](sources.md#fox-wingspan)) | 301 | 0.368 | — |
| neck length ([`bohmer-neck`](sources.md#bohmer-neck)) | 103 | **0.380** | 0.338 – 0.421 |
| total leg length ([`bohmer-neck`](sources.md#bohmer-neck)) | 103 | **0.364** | 0.328 – 0.401 |
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
wingspan estimate at 0.368 corroborates it. It sits comfortably inside the band P-047 requires.

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
inside the body contour, where [P-045](product.md#p-045) models the *exposed* neck of a
silhouette. A raven's cervical column is 106 mm; its visible neck is not. The apparent mass trend
in that join (+0.115, CI +0.001 to +0.229) points the opposite way from the exponent argument and
is an artifact of which families were sampled — its large taxa sit in long-neck families (Anatidae
0.70, Phoenicopteridae 0.95, Gaviidae 0.50), its small ones in short-neck families (Alcedinidae
0.10). The family table drives it, not biology.

The decision rests only on the pure scaling fits, which involve no BirdView model at all.

**The head model is untouched** ([P-046](product.md#p-046)). Single-element data supports isometry
for a cranium, so the cube root stays — along with its own unresolved problem, recorded there.

**Total length is still `modeled`.** A better exponent is not a measurement. [CR-001](#cr-001) and
[P-062](product.md#p-062) are unaffected, and every length stays dashed and marked.

### What it costs

Every length in the dataset moves, so section boundaries, the global scale anchor and the ordering
within sections all shift. That churn is why this lands as its own commit, ahead of the AviList
names work rather than beside it: two large dataset changes at once would be impossible to tell
apart in review or in a screenshot.

## CR-022 — AviList replaces the ABA checklist for names, and the eBird code becomes the URL

**Subject:** `dataset.names.source`
**Kind:** Direct — a licensing constraint forces a source change, which then enables a feature
**Date:** 2026-10-02 · **Decided by:** Sameer

### What changed

Every name BirdView serves now comes from [AviList](sources.md#avilist) (CC BY 4.0) instead of
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
what turned a loss into [P-049](product.md#p-049).

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
**scope** — which 841 species are in — and scope is the remaining dependency. T-004 already
permits fetching it at build; what it does not permit is serving it. The next step is a list
BirdView derives and owns.

**It does not touch any measurement.** AVONET and Fox are unaffected, every provenance record is
unchanged, and no number moved.

**It is not a route into Cornell's data.** [sources.md](sources.md#avilist) records why the
Clements and Cornell columns are usable here when eBird/Clements itself is rejected: AviList is a
separate CC BY 4.0 publication, and nothing is taken from Cornell's own distribution.

## CR-023 — Regions are grouped two ways: by shared birds, and by geography

**Subject:** `presentation.region_filter.zones`
**Kind:** Reversal — [P-011](product.md#p-011) said zones "claim nothing about birds". They now
claim something about birds, because a corrected measurement says they can.
**Date:** 2026-10-03 · **Decided by:** Sameer

### What changed

The region selector offers **two groupings** instead of one. **Common birds** is derived by
clustering regions on how much their species lists overlap. **Geographic** is the existing
hand-made grouping, kept as a convenience. A region sits in exactly one zone of each, so it
appears twice in the selector.

### Why the old claim was wrong

P-011 said there was "no data-driven way to subdivide the east", citing a 42-member block
covering everything from Maine to Georgia. **That measurement was an artifact.** It came from
faceting GBIF's `/occurrence/search` on `speciesKey`, which silently drops records whose names
the backbone cannot resolve — the same trap recorded in [sources.md](sources.md#gbif), which
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
file, so a regeneration can move them. [AC-087](acceptance.md#ac-087) recomputes cohesion from
that table rather than pinning the names, so a drift that makes the grouping dishonest fails a
test instead of going unnoticed.

## CR-024 — A region's URL is its postal code, or its name, never an invented code

**Subject:** `presentation.region_filter.url`
**Kind:** Direct — the URL carried an internal database key, and the obvious fix would have
required inventing identifiers the project has no authority to mint
**Date:** 2026-10-03 · **Decided by:** Sameer

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
renamed region cannot keep a stale URL ([AC-088](acceptance.md#ac-088)).

**Old URLs break, and no alias is kept.** Nothing links to BirdView yet, so this is free now and
would not be later — the same argument [CR-022](#cr-022) made for species URLs. Accepting both
forms would mean two identifiers for one region and a second thing to keep in step.

## CR-025 — Families are grouped two ways, and habitat is not one of them

**Subject:** `presentation.family_filter.groups`
**Kind:** Direct — the taxonomic grouping is unusable for navigation, and the obvious
replacement turned out to be unsupportable
**Date:** 2026-10-03 · **Decided by:** Sameer

### What changed

The family selector offers **two groupings**. **Birder groups** is hand-drawn and offered
first — Shorebirds, Birds of Prey, Owls & Nightbirds. **Taxonomic order** is the existing
grouping from AviList. A family sits in exactly one group of each, so it appears twice in the
selector, exactly as a region now does ([CR-023](#cr-023)).

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

**The table is keyed on the scientific family name.** [CR-022](#cr-022) renamed **55 of 90**
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
([P-031](product.md#p-031)); groupings beyond family are navigation. Every tile is one height
and a second badge costs a line on all of them.

## CR-026 — A family may be in two birder groups, and that is not a route to habitat

**Subject:** `presentation.family_filter.groups`
**Kind:** Direct — [P-014](product.md#p-014) required a partition, and one family has two
honest homes
**Date:** 2026-10-03 · **Decided by:** Sameer

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
[AC-089](acceptance.md#ac-089) caps the number of duplicated placements rather than trusting
restraint.

## CR-027 — Habitat and diet become species-level filters, and the birder grouping is retired

**Subject:** `presentation.trait_filter`
**Kind:** Reversal — [P-014](product.md#p-014)'s hand-drawn family grouping is withdrawn and
replaced by published per-species classifiers
**Date:** 2026-10-03 · **Decided by:** Sameer

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

Measured earlier and recorded in CR-025: **only 12% of families with ten or more species share
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
lookup that misses is silent, so [AC-092](acceptance.md#ac-092) now checks every key against the
dataset.

### What it costs

**The familiar birding vocabulary is gone from the controls.** "Shorebirds", "Sparrows" and
"Finches" are not habitats or diets, and AVONET has no vocabulary for them. A reader wanting
the shorebirds selects four families in the taxonomic block. That is the real price of this
change, and it is accepted because a grouping that is *wrong* about the Anhinga is worse than
one that is merely less convenient.

**One value per species is still a simplification**, and the control says so rather than
implying completeness.

## CR-028 — Scope becomes a consequence of a derived status label, and recurrence replaces abundance

**Subject:** `dataset.scope`
**Kind:** Replacement — [P-001](product.md#p-001)'s ABA rarity-code filter is withdrawn in favor
of a status vocabulary BirdView derives from open occurrence data and owns
**Date:** 2026-10-04 · **Decided by:** Sameer

### What changes

Two things, and the second is the point.

**Scope stops being a threshold.** Every attempt to derive scope from a *volume* of
observations has failed, and the next one would fail too. What replaces it is **recurrence**:
a species is in scope when its records in at least one region recur across years, however few
of them there are.

**Status becomes a product.** The ABA codes 1–6 that P-001 filters on are not a scope rule at
all — they are a *status vocabulary*, and P-001 uses three of its values as a cutoff. BirdView
deriving its own status labels makes scope a consequence of labeling rather than a knife-edge,
and the labels themselves become a navigation facet alongside Habitat and Diet
([CR-027](#cr-027)). A vagrant is no longer *excluded*; it is *labeled*, and the default view
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

**Pelagic species need the `SEA.*` regions**, which [`region`](../internal/region/region.go)
already defines. A query for Black-capped Petrel against North Carolina's land polygon returns
6 records; the species has 13,601 in the ABA Area, nearly all offshore.

### Why the weights are declared and never fitted

The obvious way to build this is to fit coefficients so the output reproduces the ABA's 839.
**That is rejected.** It would re-target exactly what [`sources.md`](sources.md#why-occurrence-data-cannot-reproduce-the-aba-scope)
disclaimed when the owner chose to own the list, and it would launder the ABA's editorial
judgement into model weights — a list that is theirs in substance and ours in format is not a
list BirdView owns. There is no other labeled training signal available, so there is no
supervised version of this model that is not that one.

Everything here is therefore either **measured** (episodes, dates, spans) or **declared** — a
stated coefficient with a written reason, as [`OrderCommonNames`](../internal/taxonomy/mapping.go)
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
[P-071](product.md#p-071) and nothing else: three with no open record at all (Maui Parrotbill,
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
deeper and better curated — while AviList is already the authority for **names** (CR-022). So
eBird's records are credited to the AviList name whose birds they are, which is what
`AviListRenames` does for genus moves; `taxonomy.LumpedInto` now does it for lumps.

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
   [`birdview-regions`](../cmd/birdview-regions) already does.
2. An `eventDate` facet for each species in its **top three regions by record count** — about
   4,100 queries. If a species does not recur where it is seen most, it does not recur anywhere.

**The established/rare boundary is set after the run, not before.** Picking it now would be
inventing a coefficient and then finding data to justify it; the honest order is to measure the
distribution first and then declare the cut with a written reason. This is the one piece of the
model still owed a number.

## CR-029 — A species URL outlives the species, but does not keep its 200

**Subject:** `presentation.unknown_species_id`
**Kind:** Conditional — [CR-018](#cr-018)(b) set this subject for `/compare`, and the detail
view resolves it the same way in substance and the opposite way in status code
**Date:** 2026-10-05 · **Decided by:** Sameer

### The question CR-018 left open

CR-018(b) decided that an unrecognised species id in `/compare` is **dropped and named**, not a
404: *"render what is recognised and name what is not."* `/species/<id>` was never revisited, and
it still calls `http.NotFound`, which writes `404 page not found` and nothing else.

That was harmless while scope came from a checklist nobody intended to change. It stopped being
harmless with [CR-028](#cr-028): scope is now **derived from live occurrence data and
regenerated**, so the set of species URLs that work is a function of a GBIF query, and a future
run can take `/species/<code>` away from a bird that has one today.

### Why it is not simply CR-018 again

CR-018's answer has two halves — *render what you can* and *name what you cannot* — and only
the second half survives the move. A comparison of four birds minus one is still a comparison
of three; a detail page minus its one bird is nothing. There is no partial page to render.

That difference decides the status code, which is the only point on which this record departs
from CR-018. `/compare` returns **200** because it delivered a comparison and merely delivered
a smaller one than was asked for. `/species/<id>` has delivered **no bird at all**, and
returning 200 would tell every cache and crawler that an address which serves nothing is fine.

### Options considered

| | Status | What the reader sees | Why not |
| --- | --- | --- | --- |
| Keep `http.NotFound` | 404 | `404 page not found` | Throws away a measured verdict the derivation already holds. This is the status quo being changed |
| Redirect to the index | 302 | The index, no explanation | Loses the question. The reader asked about one bird and is handed 1,083 |
| **Named 404** | **404** | The bird, its status, its evidence | **Chosen** |
| Named 410 Gone | 410 | The same page | Overclaims. 410 asserts the resource is permanently gone and should be de-indexed; scope is regenerated, and a bird that fails the recurrence bar on one run can clear it on the next. Three of the nine [P-071](product.md#p-071) exceptions exist precisely because sparse evidence moves |
| Serve the bird anyway, marked out of scope | 200 | A drawing | Rejected outright. [P-001](product.md#p-001) decides what BirdView draws, and a page that draws a bird the dataset excludes makes the scope rule a suggestion |

### Decision

[P-075](product.md#p-075): **404 with a page that names the bird and gives the measured reason.**
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

Harmless while nothing read the region column; not harmless once [P-075](product.md#p-075)
prints the date on a page as the reason a bird is not drawn. A reason that changes when nothing
changed is not a reason.

The tiebreak is now declared rather than incidental: status, then episodes, then **years, then
records, then the region id**. Years before records because the model is recurrence and not
abundance ([CR-028](#cr-028)) — between two regions a bird recurs in equally often, the one it
came back to across more years is the better evidence, whatever the crowds did. The id settles
what nothing else can: arbitrary, but the same arbitrary answer every run. Fifteen rows moved,
most of them onto strictly more evidence; **no species changed status, and scope is unchanged at
1,083.**

### Status of requirements touched

- [P-001](product.md#p-001) — Active; amended to name `historical` in the out-of-scope list,
  which it had omitted, and to point at P-075.
- [P-075](product.md#p-075) — Active, new.
- [CR-018](#cr-018)(b) — stands unchanged for `/compare`.

## CR-030 — The share at sea is not the share on land, and P-074 presents them as one number

**Subject:** `presentation.share`, `dataset.region.totals`
**Kind:** Derived / emergent — [P-074](product.md#p-074) and [P-011](product.md#p-011) are each
satisfiable alone; the defect is in a ratio computed across both
**Date:** 2026-10-05 · **Decided by:** Sameer — **option A**, and no commonest sea
**Found by:** the owner, reading `/species/yenalb2?region=NorthAtlantic` against
`/?region=NorthAtlantic` and noticing the two pages disagreed about where the bird is commonest

### What a reader sees

| page | says |
| --- | --- |
| `/?region=NorthAtlantic` | Atlantic Yellow-nosed Albatross, **1 in 4,521** bird records in North Atlantic |
| `/species/yenalb2` | Commonest in **Maine — 1 in 1,177,757** bird records there |

The obvious reading is that the detail page is wrong. It is not. **The index figure is the
wrong one**, and the detail page's answer survives a correct recomputation.

### Three defects, stacked

**(a) The peak-share region is computed over land only.** `birdview-status` walks
`region.LandIDs()` ([main.go:92](../cmd/birdview-status/main.go)) and has no marine pass;
`birdview-regions` walks land *and then* boxes the five seas
([main.go:211](../cmd/birdview-regions/main.go)). So `Species.Regions`, `RegionCounts` and
`RegionTotals` cover 69 land + 5 marine, while `Species.Commonest` covers 69 land.
**Zero of 1,426 species name a marine region as commonest**, and none can.

**(b) A marine region's total counts a subset of species, a land region's counts all of them.**
`birdview-regions` only boxes the seas for species already judged **pelagic** — at least
`minPelagicShare` = 0.5 of their records outside every land region — and accumulates
`totals[id]` over just those. A land total is every bird record in the region. The two are
different quantities, and P-074 divides by both as though they were one. Measured against GBIF
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
`commonestOf` and every test renders it back, so P-074's central claim — *the region where this
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
region that is not really a region, and that P-074 already refuses to publish a number it
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
nine [P-071](product.md#p-071) exceptions are pelagics whose written reason is that the
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

**Consequence: [P-074](product.md#p-074) could get its share back at sea.** "Of the records on
water in this box, how much is this bird" is a real quantity, comparable in kind to the land
figure and computable for free. That would reverse the part of option A that declines a marine
share — and would need a CR of its own, a regenerated `region-presence.tsv`, and a decision on
whether "on water" may be labelled with an ocean's name. Tracked below; nothing is changed on
the strength of this experiment alone.

## CR-031 — A sea gets an honest denominator, and its share back

**Subject:** `presentation.share`, `dataset.region.totals`
**Kind:** Reversal — [CR-030](#cr-030)'s option A declined a marine share; the measurement it
declined for turns out to be available
**Date:** 2026-10-05 · **Decided by:** Sameer

### What changed

[CR-030](#cr-030) withheld the share at sea because no honest denominator existed for a sea.
Its own [experiment](#experiment--is-there-a-distance-from-shore-that-cleans-up-the-boxes),
run immediately afterwards, found one. GADM covers land, so **a record on a beach carries a
country and a record on the water carries none**: faceting a box by `gadmLevel0Gid` and
subtracting gives the off-land half exactly, with the coastline as the boundary. The facet
response carries the unfiltered count beside the facet rows, so it costs **no extra requests**.

A sea's total is now every bird record on open water inside its box, over all species — the
same quantity a land region's total is, measured the same way. `RegionTotals` carries all 74
regions again, and [P-074](product.md#p-074)'s share ranges over every one.

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
records against 14 million, below [P-011](product.md#p-011)'s one-per-million presence bar, so
the page named a place a reader could not then find in the filter.

`Commonest` is now computed in the ingest from `Regions`, `RegionCounts` and `RegionTotals` —
the same numbers the index divides. The two pages are the same arithmetic by construction, and
[AC-100](acceptance.md#ac-100) asserts the round trip over all 1,059 placed species.
`status.CommonestBy` is deleted.

**The Atlantic Yellow-nosed Albatross now reads 1 in 282,315 in the North Atlantic**, which is
where the owner expected it, arrived at honestly rather than by the inflated figure that
originally seemed to show it. **45 species peak at sea.**

### What this does not fix

The result is **on water**, not **at sea**. An estuary has no GADM country either, so
Chesapeake Bay reads 8.8% off land, Pamlico Sound 8.5%, Long Island Sound 4.1% — roughly a 3%
overstatement of the open ocean against a 30× error corrected. GBIF accepts a WKT `geometry`
polygon alongside facets (verified), so carving the coast off properly is available if that 3%
ever matters. The boxes are declared approximations already and [P-074](product.md#p-074) says
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

- [P-074](product.md#p-074) — Active; amended, the share ranges over every region again.
- [CR-030](#cr-030) — option A's *no share at sea* half is superseded here. Its structural
  guard is not: `shareIn` still takes both sides of the ratio from one test, and
  [AC-099](acceptance.md#ac-099) still asserts it. What was wrong there was the number, and
  what was right was the shape.

## CR-032 — Frequency joins the metrics, and brings a second sectioning rule with it

**Subject:** `ordering.metric`, `sectioning.rule`, `sectioning.step_percent`
**Kind:** Direct — a fourth ordering does not fit the value [P-020](product.md#p-020) fixes, and
cannot be sectioned by the rule [P-022](product.md#p-022) fixes
**Date:** 2026-10-05 · **Decided by:** Sameer

### What changed

[P-074](product.md#p-074) put a number on every tile — *1 in 282,315 bird records in North
Atlantic* — and the index could not order by it. It can now: **Frequency** is a fourth choice
under Sort by.

Three subjects move:

- `ordering.metric` gains `frequency`.
- `sectioning.rule` gains a **second** rule, used only by frequency.
- `sectioning.step_percent` means something else under frequency, and the control shows it.

### Why the proportional rule cannot section frequency

[P-022](product.md#p-022)'s rule is already multiplicative — a section ends at `min × (1 + N/100)`
— so on the face of it 15% should work. It does not, for two reasons measured on the data.

**It produces 128 sections.** Frequency spans **1 in 43 to 1 in 1,252,475,518**, which is 7.5
decades. At 15% that is `log(1.25e9/43) / log(1.15)` ≈ 128 bands of a few hundred species each.
Length at 15% produces a page a reader can scroll.

**Its boundaries move when the filter does.** P-022 anchors each band to the data — the first
section starts at the smallest bird present. Frequency follows the region filter
([P-074](product.md#p-074)), so a data-anchored band would start at 1 in 43 nationally and
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

**No tier is given a name.** Not "Common", not "Scarce". [P-074](product.md#p-074) says the
metric "is not a rarity tag and must not become one — no threshold divides common from rare;
the number is shown and the reader judges", and a named band is exactly that threshold wearing
a label. The headers state the range, as [P-024](product.md#p-024) already requires of every
other metric.

**Species with no records in the selection are kept, in a final section of their own**, headed
*No records in the regions selected*. They cannot be placed in a tier — there is no ratio —
and dropping them would silently shorten the page by 24 birds nationally and more under a
filter.

### A known consequence — and this got it wrong

Under the size metrics a section holds birds of similar size, so
[P-042](product.md#p-042)'s per-section scale is both honest and legible. Frequency does not
correlate with size, so a frequency tier can hold a hummingbird and a swan.

*This record originally concluded that was "no worse than Global, which draws the hummingbird
just as small". That is false, and [CR-033](#cr-033) corrects it with the measurements:
per-section rescales each tier by its own largest bird, which across the nine tiers swings by
0.15× to 1.36×, so a 140 mm parrotbill is drawn the height of a 1,201 mm swan six tiers up.
Global is the default under Frequency now.*

### Status of requirements touched

- [P-020](product.md#p-020) — Active; value gains `frequency`.
- [P-022](product.md#p-022) — Active; scoped to the size metrics, pointing at P-076.
- [P-023](product.md#p-023) — Active; scoped likewise.
- [P-076](product.md#p-076) — Active, new.

## CR-033 — Frequency defaults to global scale, because per-section has nothing to stand on there

**Subject:** `diagram.scale.mode`
**Kind:** Conditional — [P-042](product.md#p-042)'s default is right for the three size metrics
and unsupported under the fourth
**Date:** 2026-10-05 · **Decided by:** Sameer, who picked Global by hand and said so

### What [CR-032](#cr-032) got wrong

That record said per-section scaling under Frequency was "the per-section scale working as
specified rather than a new defect, and it is no worse than Global". **The second half is
false**, and the owner found it by preferring Global without being told to.

[P-022](product.md#p-022) justifies per-section scaling in one sentence: *"because a section is
by construction no more than N% wide, [it] makes the per-section scale of P-042 honest."* That
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
one page**, and precisely the confusion [P-042](product.md#p-042) exists to prevent.

### Decision

**Global is the default when Order by is Frequency.** Per-section remains the default for
length, wingspan and mass, where a section really is at most N% wide and the rescaling is
honest. Either choice stays one click away under both.

The reader loses legibility, which is what [CR-017](#cr-017) moved the default *to* per-section
for. That cost is accepted here because the alternative is not a less legible drawing but a
wrong one: under Frequency, per-section scaling answers "how big is this bird" differently in
adjacent tiers for no reason the reader can see.

**Both scale values are now recognised explicitly in parsing.** With a default that varies by
metric, reading only the non-default would leave *Per section* a dead button under Frequency —
the exact trap the comment on that code already warned about, written when CR-017 moved the
default the other way.

### Status of requirements touched

- [P-042](product.md#p-042) — Active; the default is per-section for the size metrics and
  global for frequency.
- [CR-032](#cr-032) — its "no worse than Global" aside is withdrawn, above.

## CR-034 — A count that covers two species is shown on both, and says so

**Subject:** `presentation.share`, `dataset.name_provenance`
**Kind:** Derived / emergent — no requirement is violated alone; the defect is in what a tile
means when the occurrence sources and AviList disagree about what a name denotes
**Date:** 2026-10-05 · **Decided by:** Sameer
**Found by:** the owner, asking why `manwar1` outranked `yelwar1` in New York

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
already covered by `AviListRenames` or `LumpedInto`:

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
[P-077](product.md#p-077) must not say otherwise.

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

The alternative considered was showing no figure on either, which is defensible under
[G-005](goals.md#g-005) alone — a number we cannot attribute is a number we cannot stand
behind. **[G-006](goals.md#g-006) decides it the other way**: suppressing the figure hides
6.2 million records and one of the commonest warblers in North America, and a reader has no way
to tell that view from a bird genuinely absent. Showing the count with its seam named tells the
reader strictly more, and nothing it tells them is false.

### What this does not do

It does not make the Frequency ordering right for Mangrove Yellow Warbler, which will still
sit in a tier its own records do not earn. That is the honest consequence of eBird not
separating the two: the records are real, the tier is real, and the attribution is the part
nobody knows. The label is what keeps the page from asserting the part nobody knows.

## CR-035 — The region join normalizes names, and says what it drops

**Subject:** `build.join_loss`, `dataset.region.counts`
**Kind:** Direct — `birdview-status` and `birdview-regions` read the same facets under
different taxonomies, so the two committed tables disagree about what a name means
**Date:** 2026-10-05 · **Decided by:** Sameer
**Found by:** the owner, asking why the Hairy Woodpecker's count looked wrong, then asking for
an audit of everything lost through joins

### The audit

Every verbatim name eBird publishes for the US and Canada, against the species BirdView draws.
**1,402 names, 1,238,404,492 records:**

| | names | records | |
| --- | --- | --- | --- |
| match a drawn species | 1,062 | 1,228,366,591 | 99.19% |
| match a species BirdView rejects | 303 | 22,213 | correctly excluded |
| **match nothing** | **37** | **10,015,688** | **0.81%, silently dropped** |

Of the 10,015,688: **9,954,362 belong to ten species BirdView draws**, every one of them
already covered by `AviListRenames` or `LumpedInto` — tables `birdview-regions` never called.
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

**One normalization, shared.** `taxonomy.Occurrence` applies `AviListRenames` and `LumpedInto`
and is now what both readers of the occurrence data call. The bug was not a missing table
entry — every entry already existed — but a second reader that did not consult them.

**The join reports.** `birdview-regions` lists every name that matched no drawn species with
its record count, and **refuses to write** above `maxOrphanFraction` = 0.05%, a bar set just
above the 61,326 records of legitimate residue. Before this change the figure was 0.81%.

### What this does not touch

The shared names of [CR-034](#cr-034). *Setophaga petechia* and *Trogon elegans* match a drawn
species exactly — they are not orphans and never were, which is why this audit does not list
them and why that one needed a different instrument entirely.

### Frequency sorts the other way from size, and that is a reciprocal, not a mistake

*Raised by the owner 2026-10-05, evaluated, nothing changed.* Length, wingspan and mass put the
smallest bird first; Frequency puts the commonest first. The quality being ordered therefore
runs **least → most** under three metrics and **most → least** under the fourth, which reads as
an inconsistency.

It is one, and it cannot be fixed by reversing the sort. **"1 in N" is a reciprocal**: N rises
as the bird gets scarcer, so ascending the printed number descends abundance. Every ordering
here obeys [P-021](product.md#p-021) — ascending the selected metric — and the metric itself is
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
[P-074](product.md#p-074) shows a scarcity-shaped number and refuses to call it rarity. If the
control were named for the shape of the figure the tension would vanish, and that name is
exactly the one the requirement forbids. The felt inversion is the price of not shipping a
rarity tag, and it is worth paying.

**What is done instead is to say the direction rather than imply it:** the Sort by control
carries *most common first*, the help repeats it, and the first tier on the page reads
`1 in 10 – 1 in 99`. Naming the tiers — "the birds you see most" — was considered and rejected
on the same grounds as every other tier name ([CR-032](#cr-032)).

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
*observation* as the unit on every tile ([P-074](product.md#p-074)), reusing it as the axis
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
trade before and in the same direction: [P-074](product.md#p-074) prints *1 in N* rather than
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

## Work arising

Not conflicts; obligations this bundle creates that nothing yet discharges.

| Item | Owner | Gate |
| --- | --- | --- |
| ~~**`birdview-regions` does not fold iNaturalist trinomials**~~ — **done 2026-10-06**, [AC-106](acceptance.md#ac-106). The rank reduction moved into `taxonomy.Binomial` and runs inside `taxonomy.Occurrence` *before* either naming table is consulted, so `dryobates villosus harrisi` reaches `Leuconotopicus villosus` instead of nothing. The last of the three asymmetries [CR-035](#cr-035) found, and the criterion now asserts the property the three shared — that both readers land on the same AviList species — rather than the three bugs. The orphan residue was 0.0431% before the fold; what stays is hybrids, which must never match a species, and subspecies of birds BirdView does not draw | Sameer | closed 2026-10-06 |
| ~~**`birdview-regions` applies neither `AviListRenames` nor `LumpedInto`**~~ — **done 2026-10-05**, [CR-035](#cr-035). Both readers of the occurrence data now call `taxonomy.Occurrence`, and the join reports every name it could not match, refusing to write above 0.05% of records. The audit that justified it: 1,402 eBird names, 37 matching nothing, 9,954,362 recoverable records across ten species | Sameer | closed 2026-10-05 |
| **Three more domestic forms behave like the two that are declared.** The rank fold of [AC-106](acceptance.md#ac-106) moved iNaturalist's trinomials into their wild species, and the proportional gainers were `Gallus gallus` **+11.9%**, `Anser cygnoides` **+20.0%** and `Anser anser` **+6.9%** — the same signature as the declared `Columba livia` (+1.3%) and `Cairina moschata` (+10.0%). `taxonomy.DomesticForms` names only the latter two, so [T-028](tech.md#t-028)'s ratio check reads the other three as undeclared disagreements if it ever flags them. Either declare them or record why the two are special; measuring it needs a `cmd/birdview-names` run | — | opportunistic |
| **Nothing reads `data/gbif-taxa.tsv`.** It is generated by `cmd/birdview-gbifmap`, committed, and consumed by no code. It carries a GBIF-to-AviList name bridge with an `ambiguous` tier that would be the natural home for [CR-034](#cr-034)'s shared names — except that it files `Setophaga petechia -> Setophaga petechia` as `exact`, correctly, which is why the bridge could not find them. Either wire it in or retire it | — | opportunistic |
| **The name sweep is a throwaway and should not be.** Finding [CR-034](#cr-034)'s cases took 2,166 queries comparing what each database reports under each AviList binomial — the only method that catches a name whose strings match and whose meanings differ. It ran once from a scratch binary. As a committed command it would catch the next split the week it lands, rather than when a reader notices a mangrove warbler in Ontario. Note its blind spot: a `country=US,CA` filter hides the outlying islands, which is why *Sternula albifrons* looked like a finding and was not | Sameer | before the next AviList release |
| ~~**Restore the share at sea on an off-land denominator**~~ — **done 2026-10-05**, [CR-031](#cr-031). Sea totals are now every bird on open water in the box: North Atlantic 12,986,514 against the 429,535 shipped. `Commonest` moved onto the same table the index divides, so the two pages cannot disagree again, and **45 species peak at sea**. Remaining: the figure is *on water* rather than *at sea*, overstating by ~3% through estuaries, and a WKT polygon would fix that if it ever matters | Sameer | closed 2026-10-05 |
| ~~**Is there a distance from shore that tightens the marine boxes?**~~ — **answered 2026-10-05**, no: distance is the wrong instrument and there is no gradient to tune. Composition does not taper offshore, it flips between the band holding coastline and the band that does not — 4.9% off land and Northern Cardinal at one step, 100% off land and pure tubenoses at the next. GADM membership does the job exactly and for free. Written up under [CR-030](#cr-030) | Sameer | closed 2026-10-05 |
| **Decide how a marine region's share is defined, or that it has none** — [CR-030](#cr-030). The index inflates every marine share by 15–4,472× because a sea's denominator sums only the pelagic species, where a land region's counts every bird. Option A (presence at sea, no share) is recommended and costs almost nothing; option B costs a regenerated `region-presence.tsv` and `status.tsv` plus a coordinate-box form of the `eventDate` facet. **Until this lands the index is wrong for any marine selection** | Sameer | next — it is live and misleading |
| ~~**Reword the four pelagic exceptions in [P-071](product.md#p-071)**~~ — **done 2026-10-05** with [CR-030](#cr-030). They said the derivation sees these birds rarely "because almost nobody is on the water where it is", which stated a measurement nobody made: the recurrence test runs on `gadmGid`, GBIF has no `gadmGid` for open ocean, and `DatesIn` has no coordinate-box form, so pass 2 has never asked the water anything. Each now says the records are at sea *where the test cannot yet look*, and names what the derivation does see on land. **The exceptions stand** — whether these species would clear the bar unaided remains unknown, and that is now what the table says | Sameer | closed 2026-10-05 |
| **Give the recurrence test a coordinate-box form, and find out whether the pelagics clear the bar on their own.** `DatesIn` filters by `gadmGid`; GBIF assigns none to open ocean, so no pelagic's at-sea records have ever been tested for recurrence. Four [P-071](product.md#p-071) exceptions exist in that blind spot, and [CR-028](#cr-028)'s claim that recurrence is insensitive to observer effort has never been tested where observers are scarcest. The `eventDate` facet does accept a lat/lon box — probed 2026-10-05 — so the client method is small; the cost is a re-run of `birdview-status` against live GBIF. Note the boxes overlap the coast ([CR-030](#cr-030)), so a naive run would count shore records as sea | Sameer | opportunistic; it could retire hand-declared exceptions |
| **Assert that `Commonest` is the largest share.** Nothing does. The server fixture hardcodes `commonestOf` and every test renders it back, so [P-074](product.md#p-074)'s central claim has never met data — which is why [CR-030](#cr-030)(a) shipped. The check belongs over the real dataset, as [AC-098](acceptance.md#ac-098)'s id-collision half is | — | with CR-030 |
| **Region membership counts records, so a one-off vagrant can outscore a resident.** [P-011](product.md#p-011)'s 1-ppm rule is anti-correlated with what it measures: the more surprising a bird, the bigger the crowd, the more records. **Two examples are in the data as shipped** — American Flamingo in Pennsylvania (931 records, one pond, thirteen days, post-Idalia) and **Lucy's Warbler, a Sonoran desert bird, in Virginia** (53 records, one location, six days). **Three candidate remedies have been tried and all three fail; they are recorded so the next attempt does not repeat them.** *(1) Records in ≥2 of 3 four-year windows* — Lucy's Warbler walks through it, present 2016-12-31 to 2017-01-06, so one continuous event straddles a window boundary and scores 2/3. *(2) A threshold on distinct places* — cannot work at any value: Akohekohe, a Hawaiian endemic, has 4 places and Wood Stork in Nova Scotia, a vagrant, has 7. A critically endangered endemic with four records and a vagrant with four records are the same shape. *(3) Local abundance against continental abundance* — Flamingo/PA scores 1.61, above several birds that must be kept, because flamingos are rare everywhere so the baseline is tiny. A composite — drop if ≥20 records **and** (≤2 places **or** ≤1 year) — keeps every endemic and withheld species correctly and catches Lucy's and the flamingo, but still misses Wood Stork/NS, Spoonbill/ME, Fork-tailed Flycatcher/DE and Baltimore Oriole/NT. **Measurement caveat for whoever resumes**: these figures came from one GBIF page of ≤300 records, which is a *contiguous slice in GBIF's own ordering, not a random sample* — all 300 American Robin records for Pennsylvania fell in January 2024, and a page-derived year count read 1 where the true count is 60. Any real rule must take both signals from facets over all records, which raises the cost. Costs a regenerated `data/region-presence.tsv` and dataset. **Knock-on: the "Common birds" zones were clustered *from* that table ([CR-023](#cr-023)), so they must be re-derived rather than left as a stale snapshot.** Documented for readers meanwhile under [Regions](../docs/data.md) | Sameer | deferred 2026-10-04; before anyone reads a region list as a range map |
| Group families by **what is seen together**: crawl eBird checklists and cluster families by co-occurrence, as a third grouping ([CR-025](#cr-025)). This is the honest version of the habitat idea — co-occurrence is natively a *family*-level property, so nothing is averaged and nothing is misfiled. Two findings make it tractable, both measured 2026-10-03: GBIF's eBird export carries **no `eventID`**, but a checklist reconstructs from `recordedBy` + `eventDate` + rounded coordinates **only from a complete slice** — a 0.1% sample of a state-month gives 96% singleton checklists, while one county-day (1–3k records) gives real ones, median 19 species. A 602-checklist probe already separated waterbirds, woodland songbirds, open country, feeder birds and **town birds** (pigeons, House Sparrows, starlings), leaving hawks and crows unattached because they occur with everything. Expect the rare pelagic and Hawaiian families to be unplaceable, as the atolls were | Sameer | opportunistic; needs a stratified crawl of ~400–800 county-days into a committed table |
| ~~Retrieve, hash and vendor `fox-wingspan`; write its `PROVENANCE.md`~~ — done: `spec/sources/fox-wingspan/PROVENANCE.md` exists, `spec/check/pins.py` verifies the hash, and the ingest reads it for 300 measured wingspans | — | closed 2026-10-04 |
| ~~Write `acceptance.md` — Gate 2, before any code~~ — done: 68 criteria, and Gate 2 passed long before the current work | — | closed 2026-10-04 |
| **Publish the BirdView dataset as a source in its own right.** The project consumes four published datasets and credits each by DOI; it now derives work worth the same treatment — the occurrence status ([CR-028](#cr-028)), the region presence and shares, the supplementary habitat and diet assignments, the neck and trunk models, and the wingspan regression. None of it is reachable except by scraping the app. **What the shape should be is the open question, not whether.** `web/data/species.json` is a serving artifact, not a publication: it nests provenance inside every metric, repeats source strings across 1,083 species, and carries presentation fields — `familyColor`, `familyTextColor`, `id` — that mean nothing outside this app. A publication probably wants CSV, one table per thing, keys that join: `species.csv` (the identity columns), `measurements.csv` (one row per species per metric, with its kind, source and model, so provenance is a column rather than a nest), `categories.csv` (habitat and diet with origin), `regions.csv` (species, region, records), `status.csv` (what `data/status.tsv` already is). Plus the thing none of our own sources gave us and all of them needed: a `PROVENANCE.md` stating the version, the license, what is measured against what is modeled, and the known distortions — eBird withholding sensitive species, crowd-driven over-reporting, the 4 species with no AVONET row. **Licensing needs deciding first**: the derived values are built from CC BY 4.0 inputs, so attribution flows through, and BirdView should pick its own license rather than leave it implied | Sameer | after the repo exists |
| ~~**Decide what happens to a URL when the derived scope changes.**~~ — **done 2026-10-05**, [CR-029](#cr-029) and [P-075](product.md#p-075). `/species/<id>` for a bird the status table classified but left out of scope now answers **404 with the bird named and its measured reason**, covering **328 of the 334 species** the derivation rejected — 124 vagrant, 167 absent, 37 historical; the other six have no AviList row, so no name and no stable id. CR-018's *name what was lost* carries over; its **200 does not**, because a comparison minus one bird is still a comparison and a detail page minus its bird is nothing. 410 was rejected for overclaiming: scope is regenerated, so a bird out today can be in after the next run. Cost a tenth column in `data/status.tsv` — the date last recorded — without which "historical" could be asserted and not shown | Sameer | closed 2026-10-05 |
| Ask Ricklefs or ESA for permission to use the measured-length data | Sameer | **after a working demo** |
| ~~ABA has not replied. Decide between chasing the permission and publishing a derived list~~ — decided: stop waiting, use public data. Names moved to AviList ([CR-022](#cr-022)); **scope is what remains** | Sameer | decided 2026-10-02 |
| ~~Replace ABA **scope** with a list BirdView derives and owns~~ — **done 2026-10-05, and this was the deployment blocker.** [CR-028](#cr-028): scope is recurrence, not abundance. `birdview-ingest` reads the committed [`data/status.tsv`](../data/status.tsv); `internal/aba` is deleted; nothing reads `fetched/`; the ABA credit is gone from all three footers because nothing on a page comes from it. **1,083 species in 99 families**, up from 839 — 830 of the old list kept on measured evidence, 9 kept by hand under [P-071](product.md#p-071), and the rest admitted because they recur. Every input the build reads is now CC BY 4.0 or the project's own work | Sameer | closed 2026-10-05 |
| ~~Ask ABA for written permission to redistribute the checklist~~ — **sent 2026-09-27**, awaiting reply | Sameer | done |
| ~~Decide the neck's flight posture~~ — done, [CR-011](#cr-011) | Sameer | closed 2026-09-23 |
| Assess whether a perched/standing toggle is achievable at any honest accuracy ([CR-011](#cr-011)) | — | before it is offered to users |
| ~~Investigate cervical vertebra counts as an anchor for the neck model~~ — answered twice, and both answers are no. As a **calibration** anchor: counts are dimensionless and the trunk's are conserved, so they cannot do it; [`bohmer-neck`](sources.md#bohmer-neck) measures neck *lengths* instead and now calibrates [P-047](product.md#p-047). As a **display** device — one dash per vertebra in the neck stroke — the neck is too short to carry them, measured 2026-10-04 over all 839 species with each bird filling its box: the median neck is **14.8px on the detail page and 7.8px on a tile**, which at a typical fourteen cervicals is **1.06px per vertebra**. Only American Flamingo (76px), Trumpeter Swan (68px) and Great Blue Heron (65px) reach even 4px per vertebra, and only on the detail page. Two further blocks if the geometry had worked: dash style already carries provenance ([P-065](product.md#p-065)), so dash *count* would overload the one mark that says "not a measurement"; and Böhmer covers 103 species against BirdView's 839, so ~88% of counts would be invented | — | closed 2026-10-04 |
| **Review `NeckRatios` against the full family list** — 24 of 90 families carry a ratio and **66 take the default 0.18, covering 548 species**. This is the half of the old entry that is actually tractable, and the one that catches the next short-necked Limpkin. ~~Derive the values from [`bohmer-neck`](sources.md#bohmer-neck) instead~~ — **withdrawn 2026-10-04**: [`sources.md`](sources.md#bohmer-neck) already records why, and the old entry ignored it. Böhmer measures the **full cervical column including the part inside the body contour**, where [P-045](product.md#p-045) models the *exposed* neck of a silhouette. Converting one to the other needs the fraction of the column that shows in flight, per family, unmeasured — which is exactly what `NeckRatios` already is. It would move the guess, not remove it, and add a dependency on a 103-species `.docx` that is consulted-not-ingested | Sameer | opportunistic; still the weakest model |
| Find a source that measures a **trunk**. CR-021's exponent is inferred from neck and leg, and that is its weakest link | — | opportunistic; would upgrade a calibration to a fit |
| ~~Promote the cube/crossbar clearance invariant from prose into a test over the real dataset~~ — done 2026-10-04: `TestTileCubeNeverReachesTheCrossbar` asserts it over 839 species at both scales and all three metrics, and **logs the tightest margin** (33–41px) rather than asserting a figure read off a run. A boolean check was tried first and rejected as too weak: the cube is bottom-left and the crossbar upper-middle, so only a gross change breaks the conjunction | — | closed 2026-10-04 |
| ~~Add a **Sort by** control to the comparison view~~ — built, [AC-079](acceptance.md#ac-079) | Sameer | closed 2026-10-01 |
| ~~Decide whether mass gets an **equal-mass cube of water** drawn in the diagram~~ — built, [CR-019](#cr-019) | Sameer | closed 2026-10-01 |
| **Parked:** replace the cube with a mass **line** proportional to mass. Measured and written up in [Mass as a proportional line](#mass-as-a-proportional-line-replacing-the-cube--parked); needs a CR, because it reverses CR-019 | Sameer | parked 2026-10-02 — resume from that section, do not re-derive |
| ~~Give the mass cube a colour that is not the crossbar's~~ — resolved the other way round: the **crossbar** moved to green `#4ade80`, because the cube's blue depicts the water and should keep it. The recolour also fixed a live 7:1 violation, the labels sharing that colour being text | Sameer | closed 2026-10-02 |
| ~~Mass units switch at 16 oz, so one section can print `14.5 oz` beside `1.2 lb`~~ — done 2026-10-04, [P-073](product.md#p-073) and [AC-094](acceptance.md#ac-094): the mass unit is now a property of the comparison **scope** — a section or the whole set on the index, following the scale setting; the birds named in the address on the comparison page — and the **smallest** bird in the scope chooses it. A swan beside a sparrow reads `390.5 oz` and `0.93 oz`; two large birds keep pounds. Taking the largest instead was tried and rejected: it crushes a 27 g sparrow to `0.1 lb` | — | closed 2026-10-04 |
| ~~**`README.md` and `docs/` do not describe the current product**~~ — done 2026-10-04. `using.md` rewritten (it said Global was the default scale, the opposite of the page, and documented none of Regions, Diet and Habitat, compare, `/added` or the served spec); `building.md` had a mangled code fence rendering four paragraphs of prose as code on the served page, listed two of six commands and eight of fourteen packages, and said the ABA permission request was "drafted and unsent" when it was sent 2026-09-27; `data.md` claimed "Taxonomy is ABA's" two paragraphs after saying every name comes from AviList. README gained the GBIF row and the CR-028 status. [T-024](tech.md#t-024) keeps the reading gate: none of this was catchable by [AC-090](acceptance.md#ac-090), which compares counts and credits, not whether prose is true | Sameer | closed 2026-10-04 |
| ~~Delete the dead `.sw.*` swatch rules in `styles.css`~~ — done 2026-10-04: eight rules removed, no template has emitted `class="sw …"` since [CR-017](#cr-017) | — | closed 2026-10-04 |

### A cube of water for mass

*Evaluated 2026-10-01, owner's idea, not a decision.* [P-066](product.md#p-066) rejects drawing
mass as a shape and gives two reasons. The owner then proposed a specific shape that answers one
of them completely, so the rejection is narrower than it reads and the measurements are recorded
here rather than left to be redone.

**Draw the cube of water that weighs what the bird weighs**, side = ∛(mass). This is not a model
of anything: one gram of water is one cubic centimetre, by the historical definition of the gram,
so the side is a unit conversion and carries no assumption about avian density. It is a real
length, so it can be drawn at the diagram's own scale and does not introduce the second,
arbitrary scale P-043 refuses. P-066's objection on that count does not apply to it.

**Measured on the real data, at the detail view's geometry** (drawing area 668 × 281):

| | side |
| --- | --- |
| smallest (Fork-tailed Flycatcher, 32 g) | 28 px — 10% of the drawing height |
| median | 52 px — 18% |
| largest (Thick-billed Murre, 964 g) | 80 px — 29% |

It fits. Placed in the lower-left corner with an 8px margin, the crossbar intrudes on it for
**0 of 841 species**; the tightest clearance is 50 px, on Western Grebe.

**What it would add that nothing else shows.** The extremes are not arbitrary: the largest cubes
belong to alcids — Thick-billed Murre, Parakeet Auklet, Crested Auklet, Marbled Murrelet, all
dense diving birds — and the smallest to long-tailed, light-bodied species, Fork-tailed and
Scissor-tailed Flycatcher and Black-billed Magpie. The cube read against the bird's own length is
**bulk against elongation**, which is exactly the thing a length-and-wingspan diagram cannot say.

**What it still does not do**, and the reason P-066's second objection stands: it compresses. A
Bushtit and a Common Raven differ 175× in mass and their cubes differ 5.6×. That is arithmetically
correct for equal-volume cubes and visually understates the gap. It matters less now that the
number is on the figure — the label can carry the magnitude and the cube the tangibility — but a
reader who takes the cubes as the mass comparison will read them wrong.

Open if it is taken up: what labels it, whether it appears on index tiles (it should not — the
tile is dense by design and this is detail-view furniture, as P-066's labels already are), and
whether it is drawn solid or dashed for the 10 near-match masses, which P-065's vocabulary already
answers.

### Mass as a proportional line, replacing the cube — parked

*Evaluated 2026-10-02, owner's idea, parked rather than decided.* Everything measured here is
still true; pick it up from this section rather than re-deriving it.

**The complaint is correct and is already in the spec.** The cube root flattens: a Bushtit and a
Common Raven differ in mass by 175× and their cubes by 5.6×. P-066 admits this and answers it with
the printed number. The proposal was to answer it with the drawing instead — a horizontal line
below the bird, from the lower-left, ticks at both ends, and the Mass label at the far end as the
Wingspan label already sits at the end of the crossbar.

**Then the owner looked at a rendered section and reported the opposite: the cubes draw the eye to
the heavier birds, and work.** That is evidence, and it is worth more than the argument above, so
it is recorded first. Measured on the section in that screenshot — #9, 51 species:

| | |
| --- | --- |
| length spread | **1.15×** (272–312 mm), by construction |
| mass spread | **8.3×** — Mangrove Cuckoo 63.9 g to Pigeon Guillemot 530 g |
| cube edge spread | **2.02×** |
| apparent area spread | **4.10×** |

A cube twice as wide as its neighbour, in a row of birds that are all the same length, is not
subtle. **The cube root is mild compression over the range a reader actually compares across.**

That reframes the Bushtit-and-Raven example this section opens with. Those two differ 175× in
mass — and they are 11 sections apart, never drawn side by side at a shared scale. The complaint
is strongest exactly where the comparison is not being made, and weakest where it is. Per-section
scale is the default (P-042), so the common case is the favourable one.

**It also rescues a comparison the numbers break.** In that same section the Pigeon Guillemot's
mass prints as `1.2 lb` and the Eared Grebe's as `14.5 oz`, because the formatter switches unit at
16 oz. Two birds of the same length, one third heavier, in units a reader has to convert before
comparing. The cubes are directly comparable where the labels are not — an argument for the shape
that nobody made when CR-019 adopted it.

**Where that leaves the line.** The case for it is weaker than when this section was written: it
would buy proportionality over a range where the cube is already legible, and pay the price set
out below. Not closed — the measurements stand and the design is worked out — but anyone picking
it up should start by disagreeing with the screenshot.

**The finding that decides the shape of the work: a line linear in mass cannot be drawn at one
global scale.** The mass range is **4,258×** (2.6 g – 11,071 g). A 380 px tile has a usable
dynamic range of roughly **180×**. With the Trumpeter Swan's line at 300 px the Calliope
Hummingbird's is **0.07 px**.

Nor is there a physically honest way out. The cube's whole elegance is that one gram of water is
one cubic centimetre *by definition*, so its edge is a true length drawn at the figure's own
scale and invents no second scale. The line equivalent is a water column of fixed cross-section,
where length in centimetres equals mass in grams — and the swan's column is **92× its own body
length**. No choice of cross-section helps, because the range to be spanned exceeds the range a
tile can draw by 24×.

**So a proportional mass line must be normalized per group, and that is the trade.** Measured
within the 20 sections at the default 15% step:

| within one length section | |
| --- | --- |
| median mass range | **4.0×** |
| worst section | **20.0×** — 57 species, 31.9–637 g |
| sections above 50× | **0 of 20** |

Comfortably drawable. Note what is and is not given up: **the bird is already drawn at a
per-section scale** (P-042), and so is the cube, so a per-section mass scale is not a new kind of
relativity. What the cube has that a normalized bar would not is that its size *means* something
absolute — you could measure it against the bird with one ruler. A normalized bar is a bar-chart
row: very readable, making no physical claim.

**It therefore needs a CR, because it reverses rather than extends CR-019.** P-066 currently
states that "a disc or bar sized by mass — a second, arbitrary scale — is still refused", and
P-043's rationale is that a mark on its own convenient scale "looks authoritative and says
something false". Those are the sentences a CR would have to overturn, not work around.

**The hard case is the comparison view**, where a reader may pick a hummingbird against a swan and
the lighter line goes sub-pixel. That is arguably the correct message — the cube is precisely what
hides it — but invisible reads as broken. The mitigation is to draw the origin tick
unconditionally, so there is always a mark, and to let the number carry the magnitude as P-066
already says it must.

**Styling, as far as it was worked out:**

- **Solid.** Mass is measured for 831 of 841, and P-065 draws `near-match` solid too.
- **Not a heavier weight.** Stroke weight carries no meaning in this diagram today, and the
  heaviest mark on a bird diagram should not be the one that is not the bird.
- **A different colour, and this is the load-bearing one.** A horizontal solid ticked line will
  otherwise read as a second crossbar — another length to scale, which is the exact false claim
  P-043 exists to prevent. Contrast measured against the `#0f172a` background, and against the
  wing's `#0ea5e9`, because a hue-only difference fails for colour-blind readers:

  | candidate | vs background | vs wing |
  | --- | ---: | ---: |
  | lime `#a3e635` | 11.8:1 | **1.8:1** |
  | emerald `#34d399` | 9.3:1 | 1.4:1 |
  | fuchsia `#e879f9` | 7.3:1 | 1.1:1 |
  | violet `#a78bfa` | 6.6:1 | 1.0:1 |

  Lime has the best luminance separation from the crossbar; the others differ by hue alone.
- **Ticks at both ends** — consistent with the chord-tick vocabulary of P-065.
- **Label at the far end** — consistent with the wingspan label in P-066.

**Two things found along the way that are worth acting on whether or not this is ever built:**

1. **The cube is `#38bdf8` and the crossbar `#0ea5e9`.** The one mark that is *not* a length is
   near-identical in colour to the mark whose entire point is that it *is* a length to scale.
   That is backwards, and it is a defect in what ships today.
2. **There is no clear band below the bird on a tile.** The tallest bird in each section fills the
   tile height exactly. The cube works because it occupies the empty lower-left *beside* a narrow
   figure; a full-width horizontal line would cross the tail unless the tile grows downward, as
   the detail view's bottom strip already does for P-066's labels.

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

**Update, 2026-09-27.** The ABA request was sent, with screenshots rather than a live demo. The
obvious plan — deploy somewhere they can reach and link it — was circular: hosting BirdView
publicly is the very thing the email asks permission for. Screenshots reproduce almost nothing, to
the rights-holder, about their own data, and need no server to stay up. The Ricklefs/ESA ask is
still outstanding.

**The cost of waiting**, recorded so it is a choice rather than a drift: if ESA grants permission
late, the length column changes from wholly `modeled` to half `measured`, which will move diagrams
and re-sort sections. Builders should treat the provenance model (P-062) as the thing that stays
fixed and the mix of kinds as the thing that moves.
