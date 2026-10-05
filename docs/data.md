# What is measured, what is modeled, and what is invented

This page exists because BirdView draws confident-looking pictures out of data that is, in places,
very thin. If you are going to trust the diagrams, you should know exactly which parts are
observations and which are arithmetic.

The short version: of the five diagram segments, **bill and tail are measured; head, neck and body
are modeled.** Every total length is modeled, and two thirds of the wingspans are. BirdView labels
all of it, per value, in the UI.

---

## The provenance system

Every number carries one of three kinds, visible on the detail page for any bird:

| Kind | Means |
| --- | --- |
| **measured** | taken from a source, for *this* species |
| **near-match** | measured on a **different** species this one was recently split from, because no measurement exists for it yet |
| **modeled** | computed. The `model` field names the formula and, where one exists, its error |

A `~` in the tile footer marks anything not measured. The rule the project holds itself to is that
a modeled value is never presented as a measurement — see
[CR-001](../spec/decisions.md) for why that decision came first and shaped everything else.

**near-match** deserves a note. Ten species are recent taxonomic splits that AVONET has not caught
up with — American Goshawk, Vega Gull, Hudsonian Whimbrel, Western Cattle-Egret, American Barn Owl,
Chihuahuan Meadowlark, Cocos Booby, Siberian Pipit, Blue-billed White-Tern, Western
Warbling-Vireo. Each borrows the parent taxon's measurements, and the detail page names the taxon
it borrowed from. Four of those splits are still awaiting a taxonomic check
([CR-006](../spec/decisions.md)).

---

## Segment by segment

| Part | Status | Where it comes from |
| --- | --- | --- |
| **Bill** | **measured** — 1074 direct, 9 near-match | AVONET `Beak.Length_Culmen` |
| **Tail** | **measured** — 1074 direct, 9 near-match | AVONET `Tail.Length` |
| **Mass** | **measured** — 1074 direct, 9 near-match | AVONET `Mass` |
| **Wingspan** | **391 measured**, 7 near-match, **685 modeled** | Fox et al. 2026 where empirical; otherwise a fitted regression |
| **Head** | **modeled, unvalidated** — all 1083 | `7.5 × mass^(1/3)` |
| **Body** | **modeled, calibrated** — all 1083 | `14.0329 × mass^0.37` |
| **Neck** | **modeled, invented** — all 1083 | a per-family ratio × the trunk |
| **Total length** | **modeled** — all 1083 | `bill + head + neck + body + tail` |

Those five — bill, head, neck, body, tail — are the diagram's segments, and they are what total
length is built from. Mass and wingspan are drawn or reported but are not part of the sum.

Of the five, **bill and tail are observations. Head, neck and body are not.**

### Why every length is modeled

This is the project's central compromise and it is worth stating plainly.

AVONET — the only openly licensed morphology dataset with this coverage — **has no body length and
no wingspan**. Field-guide lengths are not openly licensed. The one comprehensive academic source,
Ricklefs (2017), is published under terms that prohibit extraction, storage and redistribution, so
it was evaluated and **rejected** ([CR-002](../spec/decisions.md)).

So BirdView builds length rather than borrowing it: bill + head + neck + body + tail, with the two
measured parts real and the three modeled parts labeled. The segments sum exactly to the total
([P-041](../spec/product.md)), which is what lets the height of a diagram *be* the bird's
length and makes two diagrams at one scale directly comparable.

The consequence is that **BirdView's lengths will not match your field guide.** They are a
different quantity, built a different way, and they should be read as a consistent basis for
comparison rather than as authoritative measurements.

### The wingspan model, which is the one that was actually validated

For the 533 species without an empirical wingspan, BirdView fits its **own** order-level log-log
regression of wingspan on wing chord, trained on the empirical column of the Fox dataset.

It replaced an invented formula whose mean absolute error was 18.7%. Five-fold cross-validation:

| Model | mean | median | p90 |
| --- | --- | --- | --- |
| The invented constants it replaced | 18.7% | 18.2% | 29.9% |
| Global log-log on wing chord | 9.0% | 7.3% | 19.0% |
| **Order-level log-log (adopted)** | **6.0%** | **4.6%** | **12.6%** |

Asked for the same species Fox et al. predicted, it agrees with their published predictions to a
**mean of 1.32%** — it is, in effect, their model reconstructed from the data they published. That
agreement is a standing check, so a regression in the fitting code shows up as disagreement with an
external reference rather than as a quietly worse dataset. Full reasoning:
[CR-008](../spec/decisions.md).

Each modeled wingspan states its own error. A Passeriform reads:

> log-log regression of wingspan on wing chord, fitted on 360 empirical species in Passeriformes
> from the Fox et al. 2026 dataset; cross-validated median error 4.7%

**Note what is not claimed.** Fox report no predicted value for species that already have an
empirical one, so there is no held-out set on which to compare their predictions against ours
directly. The 1.32% is agreement between two models; the error figures are our own
cross-validation.

One important trap avoided: AVONET's `Wing.Length` is the **wing chord** — a single folded wing. It
is not half a wingspan, and doubling it is wrong by a wide and species-dependent margin. BirdView
never does that.

### The three modeled segments, in descending order of confidence

**Body (trunk) — `14.0329 × mass^0.37`.** An allometric relation. Mass is the only whole-animal
measurement AVONET carries, so the trunk is derived from it rather than from a linear trait.

There is still no dataset of avian trunk lengths to fit against, so this is not a fit. But the
exponent is no longer guessed: it is **calibrated** against measured skeletons. Three independent
dimensions agree on how a linear measure grows with mass — wingspan 0.368 (n=301), neck length
0.380 and total leg length 0.364 (n=103, Böhmer et al. 2019) — and the model must sit inside the
interval those imply. The coefficient follows from holding a 63.4 g bird fixed, so the
recalibration changes proportion across the size range rather than overall scale.

The previous `18 × mass^0.31` was invented, and too shallow: it drew the largest birds in the set
about 1.65× too short relative to the smallest. What this still does not have is a measured
*trunk* — the evidence comes from necks and legs.

**Head (cranium) — `7.5 × mass^(1/3)`.** Isometric: a linear dimension scales as the cube root of
mass. The coefficient is chosen to put a 20 g passerine near 20 mm. It is a shape function, not a
fit, and it is **unvalidated**.

It replaced a model based on bill depth, `max(12 mm, depth × 1.65)`, which is worth recording as a
failure mode: the floor bound for **442 of 841 species** and 86% of birds under 200 mm, so half the
dataset shared one constant head that varied with nothing. The current model has no floor
deliberately — a 2.5 g hummingbird lands near 10 mm, so a floor would only mask a bad coefficient.

**Neck — a per-family ratio of the trunk. This one is invented.** No source measures avian neck
length at anything like this coverage. The ratios are a drawing parameter: Ardeidae 0.85,
Phoenicopteridae 0.95, Anhingidae 0.90, Gruidae 0.85, down to Strigidae and Tytonidae at 0.08, with
everything unlisted at 0.18.

They are surfaced anyway, with the function named — a heron's detail page reads *"neck modeled as
0.85 × trunk (family-specific ratio for Ardeidae)"*. The reasoning
([CR-007](../spec/decisions.md)) is that the user sees the neck either way, since it is
drawn at a length in every diagram. Hiding the number hides the explanation, not the claim.

---

## Deliberate fictions in the drawing

These are choices about rendering, not data. None of them produces a number.

**The diagram is a top-down view of a bird in flight.** Stated as a requirement
([P-040](../spec/product.md)) because leaving it unstated made several questions
unanswerable. It is also what makes the wing crossbar honest: seen from above, wingspan and body
length genuinely lie in one plane, so drawing them at one scale is a fact rather than a convention.

**There are no legs.** They were drawn from the measured tarsus, then from a modeled leg, and then
removed entirely. In flight most birds tuck the legs away; for the minority that trail them —
herons, cranes, storks, ibises, flamingos, stilts, rails, loons, grebes — what shows is the
projection past the tail tip, which needs the femur, the toes, the hip position and the *full*
tibiotarsus. AVONET measures only the tarsometatarsus. Attempting it anyway put a White Ibis at
−14 mm of projection, a bird whose trailing legs are unmistakable.
[CR-010](../spec/decisions.md) has the numbers and what the removal cost.

**Herons and pelicans get an S-shaped neck.** Ardeidae and Pelecanidae fly with the head drawn back
onto the shoulders — the field mark separating a heron from a crane, a stork or an ibis in the air.
Their neck is drawn as two semicircles, which run exactly the neck's length of stroke through
`2/π ≈ 63.7%` of that height. The stroke is still to scale; the compression is a consequence of the
shape, not a chosen number. Sorting, sectioning and the detail view all still use the full neck
length ([CR-011](../spec/decisions.md)).

**The head is an oval 0.72 as wide as it is long**, the tail is a stick, and the body is a stick.
No source gives a tail width or a body width, and drawing a fan or a teardrop would assert a shape
the data does not support.

**Colors encode taxonomy only** — related hues within an order, distinguishable between families.
They say nothing about the bird.

---

## The iNaturalist link

The detail page links out to the species on iNaturalist, because BirdView has no photographs.

The link is only shown where the match is **confident**: iNaturalist's scientific name matched
ours exactly, *and* the common name it returns agrees with the one BirdView shows once
punctuation and case are normalized. Requiring both means a taxonomic revision on either side drops the link rather than
quietly repointing it at the wrong bird. Species that fail either test show no link at all — not a
search URL, which would never 404 and would therefore always look like it had worked.

Matching was measured before adoption: on a 55-species sample containing all ten near-match
splits, 55 resolved exactly on scientific name and 54 also agreed on common name, the one
difference being a hyphen. Wikimedia Commons was tested too and failed on three of six of the hard
cases, because it lags the recent splits. [CR-012](../spec/decisions.md) has the full comparison.

BirdView **links and never copies**. No photograph is embedded, hotlinked or cached; the reader
arrives on iNaturalist, where each image carries its own license — frequently CC BY-NC, which
covers contributors' work and not the taxonomy we take an identifier from.

## Scope

**1083 species**: every bird whose records recur in the regions BirdView covers, by its own
rule — not by anyone else’s checklist. Those regions are 51 US states and DC, 13 Canadian
provinces and territories, 5 US outlying islands and 5 stretches of open sea. **They are not
the ABA Area**: that includes St-Pierre et Miquelon, which BirdView has no region for, and
excludes Johnston, Palmyra, Wake and Navassa, which BirdView covers. A species is in when, in at least one region, its observations fall
in repeated episodes across years. Nine more are kept by hand because the open record cannot
speak for them, each named with its reason.

Two species the occurrence data names are dropped: Gray-headed Swamphen and Western
Warbling-Vireo, both of which eBird splits and AviList lumps into a bird already in the list.
Every name BirdView serves comes from AviList, so one AviList has no concept for has no name
this project can publish.

**The birds that did not make it still have a page.** The derivation classifies 1426 species
and draws 1083; the other 334 were judged `vagrant`, `historical` or `absent`. Ask for one of
them by URL — `/species/paspig` — and you get a 404 whose body names the bird and gives the
measured reason, rather than `404 page not found`. Passenger Pigeon recurs across 89 episodes
in 52 years in Ontario and every one of those records ends on 18 May 1902; the page says so.
An `absent` species shows no counts at all, because none were measured — the derivation never
ran a date query for it, and the zeros in its row mean *not measured*, not *measured zero*.

It is 404 rather than 410 on purpose. Scope is re-derived from live occurrence records, so a
bird that misses the bar on one run can clear it on the next as more observations are
published. "Gone permanently" is not a claim this derivation supports.

**Names** come from AviList v2025b, in two vocabularies: eBird/Clements (Dovekie) by default and
AviList/IOC (Little Auk) on request. The **species id in the URL is the eBird code** — `amerob`,
not `turdus_migratorius` — which does not change when you switch vocabulary, because it is
assigned rather than derived from the English name.

**Galliformes have no empirical wingspan at all.** All 30 in-scope species are absent from the Fox
dataset, so every one is modeled.

**Taxonomy is AviList's** — names, families, and the sequence families are listed in
([CR-022](../spec/decisions.md#cr-022)). Where AVONET's BirdLife taxonomy disagrees, AviList
wins, and the join is bridged by explicit synonym tables rather than by fuzzy matching. The
tables run in both directions and are kept apart on purpose: one reaches *backward* from a
current name to the older label AVONET measured under, the other *forward* from BirdView's name
to a newer label AviList adopted, and merging them would make the direction of each entry a
matter of memory.

## Regions

Which regions a bird is listed in is **derived**, not published. No source states "this species
occurs in Pennsylvania" at a licence BirdView can serve, so the list is computed from occurrence
records and the rule is stated here rather than implied by the result.

**Regions are GADM level-1 units** — US states, Canadian provinces and the outlying islands —
plus five ocean boxes that are BirdView's own and are described below.

**A species is listed in a region when it reaches 1 record per million of that region's bird
records, in either eBird or iNaturalist**, both via GBIF under CC BY 4.0.

**Relative rather than a fixed count**, for two measured reasons. eBird's volume doubled in five
years, so any absolute bar admits more species over time for no ecological reason. And effort
differs enormously between regions — California has 114 million bird records against Hawaii's
2.5 million — so one count cannot mean the same thing in both.

**Both sources, because one is not enough.** eBird's GBIF export **withholds
conservation-sensitive species entirely**: Spotted Owl, Gyrfalcon, Northern Hawk Owl and the
critically endangered Hawaiian endemics have zero eBird records and hundreds in iNaturalist. An
eBird-only table would omit exactly the birds a reader is most likely to look up.

**Names are matched verbatim**, as the publisher wrote them, not through GBIF's taxonomic
backbone. The backbone silently drops what it cannot resolve and reports **zero** US and
Canadian records for Cooper's Hawk, which is a bird that exists.

**Six species belong to no region, and that is not an error.** Bermuda Petrel, Desertas Petrel,
Trindade Petrel, Townsend's Storm-Petrel, Craveri's Murrelet and Guadalupe Murrelet are seen at
sea, where no land polygon reaches. Bermuda Petrel has hundreds of records offshore and
none inside any GADM region. They are given ocean regions instead, and their detail pages show
no Regions row rather than a row claiming a rule placed them nowhere.

**The five ocean boxes are BirdView's own and approximate.** They are coordinate rectangles,
disjoint so a record falls in at most one, chosen to separate the birds rather than to reproduce
any published maritime limit.

### What is wrong with this, and it is known

**The rule counts records, not birds, and a popular rarity outscores a quiet resident.** The more
surprising a bird is, the more people travel to see it, and every one of them files a checklist.

Two examples that are in the data as shipped. **American Flamingo is listed in Pennsylvania** on
931 records — one or two birds at a single pond over thirteen days in September 2023, after
Hurricane Idalia scattered flamingos up the eastern seaboard. **Lucy's Warbler, a Sonoran desert
bird, is listed in Virginia** on 53 records from one location over six days around New Year
2016–17.

So **a region list is a list of places a bird has been reported often enough to clear the bar,
not a range map**, and at the margin it contains birds that turned up once and drew a crowd.

**The fix is designed and measured, and it is not a better threshold.** Every rule that counts
*how many* records a bird has fails the same way, because a count measures observer effort:
pelagic species and remote-island endemics are under-reported, since nobody is where they are,
while vagrants are over-reported, since everyone goes to where they are. Those pull in opposite
directions, so no single threshold corrects both — a critically endangered endemic with four
records and a vagrant with four records really are the same shape.

What separates them is **recurrence**: not how many records, but whether they repeat across
years. The 2023 flamingo has 931 records inside thirteen days; Akikiki, a Hawaiian endemic, has
376 across forty-five years and is unmistakably the resident of the two. That is
[CR-028](../spec/decisions.md#cr-028), in progress.

## How common a bird is

Each bird says how much of what people record in view is that bird, as **one in N**.

**On the index the scope is whatever regions you have selected**, so the figure follows
your filter. A Gyrfalcon is 1 in 3,631,056 bird records across every region and 1 in
82,946 in Alaska — forty-four times commoner where it actually lives. Select a region the bird
is absent from and the figure falls, because those records join the denominator and none join
the numerator; a bird is a smaller part of a wider world.

**On the detail page** the scope is every region at once, and the page also names the one
where the species peaks: *Commonest region: Yukon — 1 in 250 records there.* A bird with
no records in the regions on screen reads *0 in 1,225,176,035 bird records across every
region* — a stated zero against a stated denominator, because zero in a billion records and
zero in nine hundred are different findings.

**Birds at sea get a figure too, measured on open water.** The five marine regions are
latitude/longitude boxes, and they take in the coastlines along with the water: the Pacific box
spans 150°W to 115°W, which is most of coastal California. So a sea counts only the records
inside it that fall on **no land at all** — GADM covers land, so a record on a beach carries a
country and a record on the water carries none, and the difference is the water. All six of the
Black-capped Petrel's North Carolina records sit inside the Atlantic box and stay with North
Carolina.

That makes a sea's denominator the same kind of number as a state's — every bird record there
— so the two can be compared, and 45 species turn out to be commonest at sea. The two kinds
never overlap, one being *inside a GADM region* and the other *inside none*, so selecting both
double-counts nothing.

The figure is **on water** rather than strictly **at sea**: bay water carries no GADM country
either, so Chesapeake Bay reads 8.8% off land and the open ocean is overstated by roughly 3%.

Until 2026-10-05 a sea's denominator summed only the species already judged pelagic, while a
land region's counted every bird, and the two were divided alike — so every marine figure
shown was inflated, by 962× in the North Atlantic. Where a bird was commonest came from a different
table again, built from a land-only pass, so no bird could peak at sea and the detail page
could name a region the index did not list it in. See [CR-030](/spec/decisions#cr-030) and
[CR-031](/spec/decisions#cr-031).

### A proportion, not a count

**A count says where the birdwatchers are; a proportion says where the bird is.** The two disagree
for more than half the dataset. California
holds 35 times the Sharp-shinned Hawk records of Yukon — and 118 times the birdwatching. One
bird record in 250 from Yukon is a Sharp-shinned Hawk; in California it is one in 835. The hawk
is a bigger part of what you see in Yukon, even though far more have been logged in California.
Dividing by the region's own total cancels out how hard anyone was looking.

Counting records instead would just point at wherever the birders are. The Millerbird settles
it: 28 of its records are filed under Hawaii and 22 under Nihoa, so by count its home reads
"Hawaii" — a state with 2.6 million bird records, where the Millerbird is 1 in 93,090. Nihoa has
49,102 records and the bird is 1 in 2,231 of them. Nihoa is the island it actually lives on.

**Why "1 in N" and not a percentage.** Across the dataset the figure runs from 1 in 2 to 1 in
117 million. As percentages that is 50% down to 0.000001%; the median bird is 0.19% and the bottom
quarter rounds to nothing a reader can read. "1 in 514" stays legible the whole way down.

**Both numbers are kept, not just the ratio.** The detail page gives the species' records and
the region's total, so a figure resting on nine records says so. That is deliberate: the
alternative is a minimum-evidence cutoff, which is one more arbitrary threshold, and 50 of 1,083
species have a peak region resting on fewer than ten records.

### What this is not

It is **not abundance**, and it is not a rarity rank. It measures how often a bird is *reported*
relative to other birds reported in the same place — which is the closest honest thing to
"how often would I see this" that open data supports, and still not the same claim.

Two distortions are known and uncorrected. Species whose records eBird **withholds** — Spotted
Owl, Gyrfalcon, Northern Hawk Owl and nine others — look rarer than they are, because the
denominator counts observers whose sightings of that species were suppressed. And a bird that
draws a crowd when it appears is over-reported relative to a quiet resident, which is the same
observer-behavior bias that kept record counts out of the scope rule entirely; see
[Scope](#scope).

## Habitat and diet

Both come from **AVONET**, assigned per species, with AVONET's own definitions quoted in the
control. AVONET records **one** primary habitat and one primary diet per species, so a bird that
uses two of either is filed under one.

BirdView adds further ones where that hides something, and **every addition is marked** — as
`birdview` where this project decided it, or `derived` where it follows from AVONET's own data
by a stated rule. Everything BirdView adds is listed together at [/added](/added).

**The derived rule reads AVONET against itself.** AVONET records a trophic *level* as well as a
trophic *niche*. Where the level is `Omnivore` but the niche names one specific food, AVONET is
saying the diet is mixed and the niche only names the larger half. For a seed-eater the other
half is invertebrates, which is forced; for an insect-eater the plant half is a reading of the
family, and so is tagged `birdview` rather than `derived`.

**One category is BirdView's own invention.** AVONET's ten niches have no term for exudates, so
the four sapsuckers — a genus named for drilling and tending sap wells — were filed under
`Invertivore` or `Omnivore` with the thing they are known for missing. `Sap` fills that gap. The
bar for coining a term is that **no** AVONET term fits, not that none fits well: a bird that eats
berries is a `Frugivore` even if the word is clumsy. Birds that merely drink from sapsucker
wells — chickadees, nuthatches, kinglets, several warblers — are **not** given it; sap is a
supplement to them and a livelihood to the sapsuckers.

---

## Where this is most likely to be wrong

In rough order of how much it would change the picture:

1. **The neck ratios are invented.** They are the largest unforced assumption in the project, and
   they visibly drive the silhouette of every long-necked family. Three families have been
   corrected by eye so far — Aramidae and Pelecanidae after a reader noticed a short-necked
   Limpkin, then Sulidae on review — which says the table has a slow error signal rather than that
   it is converging. Böhmer et al. 2019 measures neck length across 68 families and is the most
   promising anchor for fixing it properly; it is logged under Work arising in
   [spec/decisions.md](../spec/decisions.md). (Cervical vertebra *counts* were the long-standing
   candidate and turn out not to work: a count is dimensionless, so it can give proportions but
   never millimetres.)
2. **The head model is unvalidated, and the body model is calibrated but not fitted.** The
   cranium's coefficient is a shape function fitted against nothing. The trunk's exponent is now
   pinned to measured skeletons, but to *necks and legs* — nobody has published trunk lengths at
   this coverage. If either coefficient is off, every length is off in the same direction.
3. **Lengths are not comparable to published lengths.** They are internally consistent, which is
   what the tool needs, but do not quote them.
4. **Two thirds of wingspans are modeled** — accurately, with a stated error, but modeled.
5. **The flight posture is uniform.** Every bird is drawn with the same implied wing position and
   the same body axis. Real birds differ enormously.
6. **One figure per species.** No sex, age, subspecies or individual variation, in a dataset where
   some species vary by 30% between sexes.

**If you can point at a source that fixes any of these** — particularly a validation set for head
or trunk length, or an openly licensed body-length dataset — that is the single most useful
contribution anyone could make. Adopting a source is a defined process, not an ad-hoc patch: see
[spec/sources.md](../spec/sources.md) and the source-adoption path in
[docs/building.md](building.md).

Corrections are welcome, and corrections to the biology especially: the models in this document
are where BirdView is most likely to be wrong. Report them on the project's issue tracker.
