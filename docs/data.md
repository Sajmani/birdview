# What is measured, what is modeled, and what is invented

BirdView draws precise-looking figures from data that is, in places, thin. This page says which
parts of each figure are observations and which are arithmetic.

In short: of the five segments in each figure, **bill and tail are measured; head, neck, and body
are modeled.** Every total length is modeled, and so are nearly two thirds of the wingspans. The
UI labels each value with its status.

For every source, field, join, filter, and model in one place, see
[how BirdView's data is built](pipeline.md).

---

## The provenance system

Every number carries one of three kinds, shown on the detail page for any bird:

| Kind | Means |
| --- | --- |
| **measured** | taken from a source, for *this* species |
| **near-match** | measured on a **different** species this one was recently split from, because no measurement exists for it yet |
| **modeled** | computed. The `model` field names the formula and, where one exists, its error |

A `~` in the tile footer marks anything not measured. A modeled value is never presented as a
measurement; see [the decision on length and wingspan
sources](../river/requirements/decisions.md#noopenlengthandwingspan).

**Near-match** applies to nine drawn species that are recent splits AVONET does not yet treat
separately: American Goshawk, Vega Gull, Hudsonian Whimbrel, Western Cattle-Egret, American Barn
Owl, Chihuahuan Meadowlark, Cocos Booby, Siberian Pipit, and Blue-billed White-Tern. Each borrows
the parent taxon's measurements, and the detail page names that taxon. For the goshawk,
cattle-egret, barn owl, and white-tern, whether AviList's change is a split or only a rename still
needs a taxonomic check; near-match is the conservative label until then.

Where the parent is drawn too, the two drawings are identical, and both tiles say *Same
measurements as* the other bird: Siberian and American Pipit, Chihuahuan and Eastern Meadowlark,
Cocos and Brown Booby, and Vega and American Herring Gull.

---

## Segment by segment

| Part | Status | Where it comes from |
| --- | --- | --- |
| **Bill** | **measured** — 1073 direct, 9 near-match | AVONET `Beak.Length_Culmen` |
| **Tail** | **measured** — 1073 direct, 9 near-match | AVONET `Tail.Length` |
| **Mass** | **measured** — 1073 direct, 9 near-match | AVONET `Mass` |
| **Wingspan** | **391 measured**, 6 near-match, **685 modeled** | Fox et al. 2026 where empirical; otherwise a fitted regression |
| **Head** | **modeled, unvalidated** — every species | `7.5 × mass^(1/3)` |
| **Body** | **modeled, calibrated** — every species | `14.0329 × mass^0.37` |
| **Neck** | **modeled, invented** — every species | a per-family ratio × the trunk |
| **Total length** | **modeled** — every species | `bill + head + neck + body + tail` |

Counts are over all 1082 species in the dataset. Mass is in grams, lengths in millimeters. The
five segments sum to total length; mass and wingspan are not part of the sum.

### Why every length is modeled

AVONET, the only openly licensed morphology dataset with this coverage, **has no body length and
no wingspan**. Field-guide lengths are not openly licensed, and Ricklefs (2017), the one
comprehensive academic source, prohibits extraction, storage, and redistribution.

So BirdView builds length from parts, two measured and three modeled. The segments sum exactly to
the total, so a figure's height is the bird's length and figures at one scale are directly
comparable. **These lengths will not match your field guide**: they are a different quantity, a
consistent basis for comparison, not authoritative measurements.

### The wingspan model

For the 685 species without an empirical wingspan, BirdView fits its **own** order-level log-log
regression of wingspan on wing chord, trained on the empirical column of the Fox et al. dataset.
Five-fold cross-validation, as absolute percentage error:

| Model | mean | median | p90 |
| --- | --- | --- | --- |
| Global log-log on wing chord | 9.0% | 7.3% | 19.0% |
| **Order-level log-log (adopted)** | **6.0%** | **4.6%** | **12.6%** |

For the 9,194 species Fox et al. predicted, it matches their published predictions to a **mean of
1.32%**; a test checks that agreement, so a fault in the fitting code shows up against an external
reference. Fox et al. give no prediction for species with an empirical value, so there is no
held-out comparison: the 1.32% is agreement between two models, and the error figures are our own
cross-validation. See [the decision record](../river/requirements/decisions.md#fittedwingspanmodel).

Each modeled wingspan states its own error. A passerine reads:

> log-log regression of wingspan on wing chord, fitted on 360 empirical species in Passeriformes
> from the Fox et al. 2026 dataset; cross-validated median error 4.7%

AVONET's `Wing.Length` is the **wing chord**, one folded wing. Doubling it is wrong by a wide,
species-dependent margin, and BirdView never does it.

**No gamebird (Galliformes) is in the Fox dataset**, so every gamebird wingspan is modeled.

### The three modeled segments, in descending order of confidence

**Body (trunk) — `14.0329 × mass^0.37`.** Mass is AVONET's only whole-animal measurement, and no
dataset of trunk lengths exists to fit against, so the exponent is **calibrated** against measured
skeletons: wingspan 0.368 (n=301), neck length 0.380, and total leg length 0.364 (n=103, Böhmer et
al. 2019) agree on how a linear measure grows with mass, and the model sits inside the interval
they imply. The coefficient keeps a 63.4 g bird's trunk where an earlier, uncalibrated model put
it, so calibration changed proportions across sizes but not overall scale. There is still no
measured *trunk*.

**Head (cranium) — `7.5 × mass^(1/3)`.** Isometric scaling, with the coefficient putting a 20 g
passerine near 20 mm. A shape function, not a fit, and **unvalidated**. There is deliberately no
minimum: a 2.5 g hummingbird lands near 10 mm, and a floor would only hide a bad coefficient.

**Neck — a per-family ratio of the trunk. This one is invented.** No source measures neck length
at this coverage. The ratios are a drawing parameter: Ardeidae 0.85, Phoenicopteridae 0.95,
Anhingidae 0.90, Gruidae 0.85, down to Strigidae and Tytonidae at 0.08, every other family 0.18.
They are shown anyway — a heron's page reads *"neck modeled as 0.85 × trunk (family-specific
ratio for Ardeidae)"* — because every figure draws a neck, and hiding the number would hide only
the explanation.

---

## Deliberate fictions in the drawing

These are rendering choices; none produces a number.

**The figure is a top-down view of a bird in flight**, where wingspan and body length lie in one
plane, so one scale for both is accurate.

**There are no legs.** Most birds tuck them in flight. For those that trail them — herons, cranes,
storks, ibises, flamingos, stilts, rails, loons, grebes — the projection past the tail depends on
the femur, toes, hip, and *full* tibiotarsus, but AVONET measures only the tarsometatarsus; a leg
modeled from it gives a White Ibis −14 mm of projection, though its legs visibly trail.
[The decision record](../river/requirements/decisions.md#inflightnolegs) has the numbers.

**Herons and pelicans get an S-shaped neck**, because they fly with the head drawn back, the field
mark that separates a heron from a crane, stork, or ibis. Two semicircles fit the full neck length
into `2/π ≈ 63.7%` of that height; the stroke stays to scale. Sorting, sectioning, and the detail
view use the full length.

**The head is an oval 0.72 as wide as long**; tail and body are single strokes, because no source
gives their widths.

**Colors encode taxonomy only**: related hues within an order, distinguishable between families.

---

## The iNaturalist link

BirdView has no photographs, so the detail page links to the species on iNaturalist — only when
iNaturalist's scientific name matches exactly *and* its common name agrees after normalizing
punctuation and case. A taxonomic revision on either side then removes the link rather than
misdirecting it. Failing species get no link, not a search URL, which never 404s and so always
looks as if it worked.

On a 55-species sample including the near-match splits, all 55 matched on scientific name and 54
on common name (the miss was a hyphen); Wikimedia Commons failed three of six hard cases because it
lags recent splits. See
[the decision record](../river/requirements/decisions.md#inaturalistlinkadopted).

BirdView **links and never copies**: no photograph is embedded, hotlinked, or cached. Images on
iNaturalist carry their own licenses, often CC BY-NC, which covers contributors' work, not the
taxonomy we take an identifier from.

## Scope

**1082 species**: every bird whose records recur in the regions BirdView covers, by BirdView's
own rule rather than anyone else's checklist. The regions are the 50 US states and DC, 13 Canadian
provinces and territories, 5 US outlying islands, and 5 stretches of open sea. **They are not the
ABA Area**, which includes St-Pierre et Miquelon (no BirdView region) and excludes Johnston,
Palmyra, Wake, and Navassa (covered by BirdView). A species is in scope when, in at least one
region, its observations fall in repeated episodes across years. Nine are kept by hand because
the open record cannot speak for them, each named with its reason.

**41 of them are not on the index**, which shows 1041. Recurrence cannot tell a wild population
from a steady supply of escapes, so parrots, waxbills, ornamental waterfowl, zoo birds, and a
farmed Emu pass the rule too. They are on a declared list with reasons and left off the index and
its controls; [the list](/excluded) links to their pages, which stay up. Long-feral populations
and wild vagrants are not on it.

Gray-headed Swamphen (*Porphyrio poliocephalus*) passes the rule but is dropped: eBird splits it
and AviList lumps it into Purple Swamphen, and every name BirdView serves comes from AviList.

**Out-of-scope birds still have a page.** 326 named species are out of scope: 124 `vagrant`, 38
`historical`, 164 `absent`. A URL such as `/species/paspig` returns a 404 that names the bird and
gives the measured reason: Passenger Pigeon recurs across 89 episodes in 52 years in Ontario, the
last on 18 May 1902. An `absent` species shows no counts: no date query was run for it, so its
zeros mean *not measured*, not *measured zero*. The status is 404, not 410, because scope is
re-derived from live records and a bird can clear the bar on a later run.

**Names** come from AviList v2025b: eBird/Clements (Dovekie) by default, AviList/IOC (Little Auk)
on request. The **species id in the URL is the eBird code** (`amerob`, not
`turdus_migratorius`), which is assigned rather than derived from the English name, so it does not
change with the vocabulary.

**Taxonomy is AviList's**: names, families, and family order. Where AVONET's BirdLife taxonomy
disagrees, AviList wins, joined through explicit synonym tables rather than fuzzy matching: one
maps a current name back to the older label AVONET measured under, the other maps BirdView's name
forward to a newer label AviList adopted.

## Regions

Which regions list a bird is **derived**: no source states "this species occurs in Pennsylvania"
under a license BirdView can serve, so the list is computed from occurrence records.

**Regions are GADM level-1 units** — US states, Canadian provinces and territories, and the
outlying islands — plus five ocean boxes of BirdView's own.

**A species is listed in a region when it reaches 1 record per million of that region's bird
records**, from GBIF under CC BY 4.0. The bar decides only the listing: records under it still
count in *1 in N*. Emu is listed nowhere, yet its 38 records, 19 in Texas, are on its tile.

**The bar is relative** because eBird's volume doubled in five years, so an absolute bar would
admit more species over time for no ecological reason, and because effort differs enormously:
California has about 114 million species observations, Hawaii 2.5 million.

**eBird first, iNaturalist only where eBird is deliberately silent.** eBird's records are reviewed
before publication and make up 98.8% of the two sources combined. But eBird withholds sensitive
species (listed under [When the sources disagree about a
name](#when-the-sources-disagree-about-a-name)); where its list says a bird is withheld and the
records confirm it, BirdView counts iNaturalist instead, and the page says so. eBird alone would
omit the birds readers most often look up; iNaturalist everywhere would count escaped pets.

**Names are matched verbatim**, not through GBIF's taxonomic backbone, which silently drops what it
cannot resolve and reports **zero** US and Canadian records for Cooper's Hawk.

**Some species are correctly listed only at sea**: Townsend's Storm-Petrel, Bermuda Petrel,
Trindade Petrel, Craveri's Murrelet, Guadalupe Murrelet, and Atlantic Yellow-nosed Albatross are
seen beyond any land polygon. Six species on the index, among them Desertas Petrel and Little
Curlew, reach the bar nowhere; their pages show no Regions row.

**The ocean boxes are approximate**: disjoint coordinate rectangles, so a record falls in at most
one, drawn to separate the birds rather than to follow any published maritime limit.

### Known problems with the region rule

**The rule counts records, not birds, so a popular rarity outscores a quiet resident**: the more
surprising a bird, the more people go to see it and file checklists. **American Flamingo is listed
in Pennsylvania** on 931 records of one or two birds at one pond over thirteen days in September
2023, after Hurricane Idalia; Lucy's Warbler, a Sonoran desert bird, is listed in Virginia on 53
records from one site over six days. **A region list is where a bird has been reported often
enough to clear the bar, not a range map.**

No threshold fixes this, because a count measures observer effort: remote and pelagic species are
under-reported and vagrants over-reported, errors in opposite directions. **Recurrence** across
years separates them — the flamingo's 931 records span thirteen days; Akikiki, a Hawaiian endemic,
has 376 across forty-five years — which is why **Scope** uses it. **Per-region presence** still
uses the record bar, a known limitation.

## How common a bird is

Each bird shows how many of the observations recorded in the selected regions are of that bird,
as **one in N**.

### What an observation is

**One observation is one species on one checklist**, both sources' own unit: eBird's id is
`OBS1912782147`, iNaturalist's is `inaturalist.org/observations/333181443`. A list of thirty
species is thirty observations.

**The number of individuals is deliberately ignored**: twelve White-throated Sparrows, one Northern
Cardinal, and fifty thousand Snow Geese are one observation each. So *1 in N* measures **how often a
bird is logged**, not how many there are. It understates flocking birds such as waterfowl and
blackbirds and overstates chickadees and cardinals. The denominator also grows with how many
*kinds* of bird people report; dividing by each region's own total cancels that out.

### Turning *1 in N* into *how often you would see it*

**Divide the average checklist length by N.** A bird on a fraction *f* of checklists of *S*
species holds *f/S* of the observations, so *f = S/N*. **American Robin is 1 in 32 observations
in New York**; on two complete New York county-days the average checklist ran 15.6 and 19.1
species, so the robin is on roughly **half to three-fifths of outings**. Where fewer species are
about, lists are shorter and the same *1 in N* means fewer outings; the ordering is unaffected.

BirdView does not publish this figure because GBIF's eBird export has no checklist id. Checklists
must be reconstructed from observer, date, and rounded coordinates over a complete slice of the
data (a partial slice makes almost every checklist look like one species), and doing that for
every region would mean downloading all the records.

**On the index the scope is the regions you have selected.** A Gyrfalcon is about 1 in 3.6 million
species observations across all regions and 1 in 82,000 in Alaska. Selecting a region where a bird
is absent lowers its figure, because those records join only the denominator.

**On the detail page** the scope is every region, and the page names the region where the species
makes up the most observations — for a Sharp-shinned Hawk, *Top region: Yukon — 1 in 249 species
observations there*. A bird with no records on screen reads *0 in* the full count of more than a
billion, because zero in a billion records and zero in nine hundred are different findings.

**Seas are measured on open water.** The marine boxes include coastline (the Pacific box, 150°W to
115°W, takes in most of coastal California), so a sea counts only records inside it on **no land**:
GADM covers land, so a beach record carries a country and a record on the water none. A sea's
denominator is thus comparable with a state's, 44 species are commonest at sea, and selecting
both double-counts nothing. See
[the decision record](../river/requirements/decisions.md#honestseadenominator). Bay water also
has no GADM country, so 8.8% of Chesapeake Bay records count as off land, overstating the
open-ocean figure by roughly 3%.

### When the sources disagree about a name

BirdView serves AviList's names; the records come from eBird and iNaturalist, which do not always
agree. Where they disagree the species page says so.

**One name, two birds.** eBird has not adopted AviList's split of the Yellow Warbler: its more than
six million records under *Setophaga petechia* cover both the northern and the mangrove bird, with
none under *Setophaga aestiva*, and **nobody knows how they divide.** Both birds show the shared
count and name each other. The Elegant and Coppery-tailed Trogons are the same case.

**eBird withholds some species** to protect them from trapping, disturbance, or hunting, and
publishes a list. Ten are withheld everywhere BirdView looks: Spotted Owl, Great Gray Owl,
Gyrfalcon, Northern Hawk Owl, Lesser Prairie-Chicken, Gunnison Sage-Grouse, Black Rail, and three
Hawaiian endemics. Others only in some places or seasons: Long-eared Owl across the United States,
Painted Bunting in Florida from October to mid-April, Short-eared Owl in New York in spring and
summer. There the counts are iNaturalist's alone and far below the real record, the month filter
does not judge them, and the page says *withheld by eBird* — a gap in what BirdView can see, not a
statement about the bird.

**The list is checked, not trusted**: it states current policy, while records reflect the policy
when published, and for 26 of the 230 places it names eBird's records are present anyway. A bird is
marked withheld only where the records show the gap. See [the decision
record](../river/requirements/decisions.md#ebirdsilencenotabsence). Because withholding and a
different name both return nothing, names were also checked against every plausible alternative.

**The same bird under another name.** eBird files seven species under names AviList has replaced,
and the page names both: five genus moves (Hairy Woodpecker is *Leuconotopicus* in AviList; eBird
files its more than seven million records under *Dryobates villosus*) and two corrected epithets
(*Pica nuttallii* against eBird's *Pica nuttalli*). **No count is affected**, and the page says so.
iNaturalist follows AviList on the genus moves and neither uses AviList's epithets, so the page
names **eBird**, the database that holds nothing under the AviList name in every case.

**Where the species boundary falls.** eBird files three species under a name that AviList gives to
a different species, and the page says which case each is. Two are splits eBird has not adopted:
North American Hudsonian Whimbrels are under *Numenius phaeopus* in eBird (556,569 records, none
under *hudsonicus*), while iNaturalist has 17,923 under *hudsonicus* and nine under *phaeopus*;
*Gygis candida* against *Gygis alba* is the same. The third is the only true **lump**: AviList does
not recognize Stejneger's Stonechat, which both databases do, so its 95 eBird and 28 iNaturalist
records belong to *Saxicola maurus*. In each case neither database holds much under the other's
name, so it is one set of birds under two names, and BirdView **adds** the records.

**Feral and domestic birds.** iNaturalist files feral Rock Pigeons under *Columba livia
domestica* and domestic Muscovy Ducks under *Cairina moschata domestica*, below species rank.
BirdView folds these into the species, and the page says so; no source separates them.

### A proportion, not a count

**A count says where the birdwatchers are; a proportion says where the bird is.** The two pick
different top regions for more than half the dataset. California has 35 times Yukon's
Sharp-shinned Hawk records but about 117 times its bird observations, so the hawk is about 1 in 250
in Yukon and 1 in 830 in California.

**"1 in N", not a percentage**, because the figure runs from 1 in 2 to beyond 1 in 100 million:
50% to below 0.000001%. The median bird in its top region is about 0.26%, and many figures round to
zero as percentages.

**Both numbers are shown**: the species' records and the region's total, so a figure resting on
nine records says so, instead of an arbitrary minimum-evidence cutoff. A couple of dozen species
have a top region resting on fewer than ten records.

### What this is not

It is **not abundance** or a rarity rank, but how often a bird is *reported* relative to others
reported in the same place: the closest open data gets to "how often would I see this", and not
the same claim.

Two distortions are uncorrected. Species eBird **withholds**, such as Spotted Owl, Gyrfalcon, and
Northern Hawk Owl, look rarer than they are, because the denominator counts observers whose
sightings of them were suppressed. And a bird that draws a crowd is over-reported against a quiet
resident, the bias that keeps record counts out of the scope rule; see [Scope](#scope).

## Habitat and diet

Both come from **AVONET**, with its definitions quoted in the control. AVONET records **one**
primary habitat and one primary diet per species, so a bird that uses two is filed under one.

BirdView adds values where that hides something, and **marks every addition**: `birdview` where
BirdView decided it, `derived` where it follows from AVONET's own data by a stated rule. All are
listed at [/added](/added).

**The derived rule reads AVONET against itself.** Where AVONET's trophic *level* is `Omnivore` but
its trophic *niche* names one food, the diet is mixed and the niche names the larger part. For a
seed-eater the rest is invertebrates, which follows necessarily; for an insect-eater the plant part
is a judgment from the family, so it is tagged `birdview`.

**One category is BirdView's own.** AVONET's ten niches have no term for exudates, so the four
sapsuckers would be filed under `Invertivore` or `Omnivore` without their staple food. `Sap` fills
that gap. The bar for a new term is that **no** AVONET term fits, not that none fits well: a
berry-eater is a `Frugivore`. Birds that only drink from sapsucker wells — chickadees, nuthatches,
kinglets, several warblers — do **not** get it; sap is only a supplement for them.

---

## Where this is most likely to be wrong

In rough order of how much it would change the picture:

1. **The neck ratios are invented.** They are BirdView's largest unforced assumption and set the
   silhouette of every long-necked family. About a third of families have their own ratio; the
   rest take 0.18, each non-passerine family individually reviewed and passerines as a group, being
   compact songbirds with little exposed neck. All are judged from flight silhouettes. Böhmer et
   al. 2019 measures the full cervical column, including the part inside the body contour, while
   BirdView models the *exposed* neck, and converting between them needs an unmeasured per-family
   fraction ([source notes](../river/requirements/imports.md#bohmer-neck)). Vertebra counts
   are dimensionless: proportions, never millimeters.
2. **The head model is unvalidated, and the body model is calibrated but not fitted.** The head
   coefficient is fitted against nothing; the trunk exponent rests on *necks and legs*, since no
   one has published trunk lengths at this coverage. If either is off, every length is off in the
   same direction.
3. **Lengths are not comparable to published lengths.** They are internally consistent, which is
   what the tool needs, but do not quote them.
4. **Nearly two thirds of wingspans are modeled**, with a stated error.
5. **The flight posture is uniform**: the same implied wing position and body axis for every bird.
6. **One figure per species.** No sex, age, subspecies, or individual variation, though some
   species differ by 30% between sexes.

**If you know a source that fixes any of these**, particularly a validation set for head or trunk
length or an openly licensed body-length dataset, that is the most useful contribution anyone
could make. Adopting a source follows the process in
[river/requirements/imports.md](../river/requirements/imports.md) and the
source-adoption path in [docs/building.md](building.md).

Corrections are welcome, especially to the biology. Report them on the project's issue tracker.
