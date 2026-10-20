# What is measured, what is modeled, and what is invented

BirdView draws precise-looking figures from data that is, in places, thin. This page says which
parts of each figure are observations and which are arithmetic, so you can decide how far to
trust them.

The short version: of the five segments in each figure, **bill and tail are measured; head, neck
and body are modeled.** Every total length is modeled, and so are nearly two thirds of the
wingspans. The UI labels each value with its status.

For the whole pipeline at once — every source, the fields BirdView reads from each, and every
join, filter and model between them — see [how BirdView's data is built](pipeline.md).

---

## The provenance system

Every number carries one of three kinds, shown on the detail page for any bird:

| Kind | Means |
| --- | --- |
| **measured** | taken from a source, for *this* species |
| **near-match** | measured on a **different** species this one was recently split from, because no measurement exists for it yet |
| **modeled** | computed. The `model` field names the formula and, where one exists, its error |

A `~` in the tile footer marks anything not measured. A modeled value is never presented as a
measurement; the reasoning is recorded in [the decision on length and wingspan
sources](../river/requirements/decisions.md#noopenlengthandwingspan).

**Near-match** applies to nine drawn species that are recent taxonomic splits AVONET does not yet
treat separately: American Goshawk, Vega Gull, Hudsonian Whimbrel, Western Cattle-Egret, American
Barn Owl, Chihuahuan Meadowlark, Cocos Booby, Siberian Pipit and Blue-billed White-Tern. Each
borrows the parent taxon's measurements, and the detail page names the taxon it borrowed from.
For four of them (the goshawk, cattle-egret, barn owl and white-tern), whether AviList's change is
a split or only a rename still needs a taxonomic check; near-match is the conservative label
until then.

Where the taxon a species borrows from is drawn too, the two drawings are identical, because the
measurements are. Both tiles then say *Same measurements as* the other bird, and both detail pages
name the row they share. Four pairs do: Siberian and American Pipit, Chihuahuan and Eastern
Meadowlark, Cocos and Brown Booby, and Vega and American Herring Gull.

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

Counts are over all 1082 species in the dataset. Mass is in grams and lengths in millimeters.

Those five — bill, head, neck, body, tail — are the figure's segments, and total length is their
sum. Mass and wingspan are drawn or reported but are not part of the sum.

### Why every length is modeled

AVONET, the only openly licensed morphology dataset with this coverage, **has no body length and
no wingspan**. Field-guide lengths are not openly licensed. The one comprehensive academic source,
Ricklefs (2017), is published under terms that prohibit extraction, storage and redistribution,
so BirdView does not use it.

BirdView therefore builds length from parts: bill + head + neck + body + tail, with the two
measured parts real and the three modeled parts labeled. The segments sum exactly to the total,
so the height of a figure is the bird's length and two figures at one scale are directly
comparable.

As a result, **BirdView's lengths will not match your field guide.** They are a different
quantity, built a different way. Read them as a consistent basis for comparison, not as
authoritative measurements.

### The wingspan model

For the 685 species without an empirical wingspan, BirdView fits its **own** order-level log-log
regression of wingspan on wing chord, trained on the empirical column of the Fox et al. dataset.
Five-fold cross-validation, as absolute percentage error:

| Model | mean | median | p90 |
| --- | --- | --- | --- |
| Global log-log on wing chord | 9.0% | 7.3% | 19.0% |
| **Order-level log-log (adopted)** | **6.0%** | **4.6%** | **12.6%** |

Asked for the same 9,194 species Fox et al. predicted, it agrees with their published predictions
to a **mean of 1.32%**: it is, in effect, their model reconstructed from the data they published.
A test checks that agreement, so a fault in the fitting code shows up as disagreement with an
external reference. The full reasoning is in
[the decision record](../river/requirements/decisions.md#fittedwingspanmodel).

Each modeled wingspan states its own error. A passerine reads:

> log-log regression of wingspan on wing chord, fitted on 360 empirical species in Passeriformes
> from the Fox et al. 2026 dataset; cross-validated median error 4.7%

**What is not claimed.** Fox et al. report no predicted value for species that already have an
empirical one, so there is no held-out set on which to compare their predictions with ours
directly. The 1.32% is agreement between two models; the error figures are our own
cross-validation.

AVONET's `Wing.Length` is the **wing chord**: one folded wing. It is not half a wingspan, and
doubling it is wrong by a wide and species-dependent margin. BirdView never does that.

**Galliformes have no empirical wingspan at all.** None of them is in the Fox dataset, so every
gamebird's wingspan is modeled.

### The three modeled segments, in descending order of confidence

**Body (trunk) — `14.0329 × mass^0.37`.** An allometric relation. Mass is the only whole-animal
measurement AVONET carries, so the trunk is derived from it rather than from a linear trait.

There is no dataset of avian trunk lengths to fit against, so this is not a fit. The exponent is
**calibrated** against measured skeletons instead. Three independent dimensions agree on how a
linear measure grows with mass — wingspan 0.368 (n=301), neck length 0.380 and total leg length
0.364 (n=103, Böhmer et al. 2019) — and the model must sit inside the interval those imply. The
coefficient is not chosen separately: it keeps the trunk of a 63.4 g bird at the length an
earlier, uncalibrated model gave it, so calibrating the exponent changed proportions across the
size range but not overall scale. The evidence comes from necks and legs; there is still no
measured *trunk*.

**Head (cranium) — `7.5 × mass^(1/3)`.** Isometric: a linear dimension scales as the cube root
of mass. The coefficient puts a 20 g passerine near 20 mm. It is a shape function, not a fit, and
it is **unvalidated**. There is deliberately no minimum: a 2.5 g hummingbird lands near 10 mm,
and a floor would only hide a bad coefficient.

**Neck — a per-family ratio of the trunk. This one is invented.** No source measures avian neck
length at anything like this coverage. The ratios are a drawing parameter: Ardeidae 0.85,
Phoenicopteridae 0.95, Anhingidae 0.90, Gruidae 0.85, down to Strigidae and Tytonidae at 0.08,
with every other family at 0.18.

The ratios are shown anyway, with the function named: a heron's detail page reads *"neck modeled
as 0.85 × trunk (family-specific ratio for Ardeidae)"*. The neck is drawn at some length in every
figure, so the reader sees the assumption either way; hiding the number would hide only the
explanation.

---

## Deliberate fictions in the drawing

These are choices about rendering, not data. None of them produces a number.

**The stick figure is a top-down view of a bird in flight.** Seen from above, wingspan and body
length lie in one plane, so drawing them at one scale is accurate rather than a convention.

**There are no legs.** In flight most birds tuck the legs away. For the minority that trail them
— herons, cranes, storks, ibises, flamingos, stilts, rails, loons, grebes — what shows is the
projection past the tail tip, which depends on the femur, the toes, the hip position and the
*full* tibiotarsus. AVONET measures only the tarsometatarsus, and a leg modeled from it gives a
White Ibis −14 mm of projection, though its trailing legs are obvious in the field.
[The decision record](../river/requirements/decisions.md#inflightnolegs) has the numbers.

**Herons and pelicans get an S-shaped neck.** Ardeidae and Pelecanidae fly with the head drawn
back onto the shoulders, the field mark that separates a heron from a crane, a stork or an ibis
in the air. Their neck is drawn as two semicircles, which fit the neck's full length of stroke
into `2/π ≈ 63.7%` of that height. The stroke is still to scale; the compression follows from the
shape, not from a chosen number. Sorting, sectioning and the detail view all use the full neck
length.

**The head is an oval 0.72 as wide as it is long**, and the tail and body are single strokes. No
source gives a tail width or a body width, and drawing a fan or a teardrop would assert a shape
the data does not support.

**Colors encode taxonomy only**: related hues within an order, distinguishable between families.
They say nothing about the bird.

---

## The iNaturalist link

The detail page links to the species on iNaturalist, because BirdView has no photographs.

The link appears only where the match is **confident**: iNaturalist's scientific name matches
ours exactly, *and* the common name it returns agrees with the one BirdView shows once punctuation
and case are normalized. Requiring both means a taxonomic revision on either side removes the link
rather than pointing it at the wrong bird. Species that fail either test show no link at all —
not a search URL, which would never 404 and would therefore always look as if it had worked.

On a 55-species sample including the near-match splits, 55 resolved exactly on scientific name
and 54 also agreed on common name, the one difference being a hyphen. Wikimedia Commons failed on
three of six of the hard cases, because it lags the recent splits.
[The decision record](../river/requirements/decisions.md#inaturalistlinkadopted) has the full
comparison.

BirdView **links and never copies**. No photograph is embedded, hotlinked or cached; the reader
arrives on iNaturalist, where each image carries its own license — often CC BY-NC, which covers
contributors' work and not the taxonomy we take an identifier from.

## Scope

**1082 species**: every bird whose records recur in the regions BirdView covers, by BirdView's
own rule rather than anyone else's checklist. Those regions are the 50 US states and DC, 13
Canadian provinces and territories, 5 US outlying islands and 5 stretches of open sea. **They are
not the ABA Area**, which includes St-Pierre et Miquelon (BirdView has no region for it) and
excludes Johnston, Palmyra, Wake and Navassa (which BirdView covers). A species is in scope when,
in at least one region, its observations fall in repeated episodes across years. Nine are kept by
hand because the open record cannot speak for them, each named with its reason.

**41 of them are not on the index**, which shows 1041. Recurrence cannot tell a wild population
from a steady supply of escapes, so parrots, waxbills, ornamental waterfowl, zoo birds and a
farmed Emu pass the rule too. They are on a declared list, each with its reason, and are left off
the index and its controls; [the list](/excluded) names them all and links to their pages, which
stay up. Long-feral populations and wild vagrants are not on it.

One species that passes the rule is dropped: Gray-headed Swamphen (*Porphyrio poliocephalus*),
which eBird splits and AviList lumps into Purple Swamphen. Every name BirdView serves comes from
AviList, so a species AviList does not recognize has no name BirdView can publish.

**Out-of-scope birds still have a page.** The derivation leaves 326 named species out of scope as
`vagrant`, `historical` or `absent` (124 vagrant, 38 historical, 164 absent). Ask for one by URL
— `/species/paspig` — and you get a 404 whose body names the bird and gives the measured reason,
rather than `404 page not found`. Passenger Pigeon recurs across 89 episodes in 52 years in
Ontario, and the last of those records is dated 18 May 1902; the page says so. An `absent`
species shows no counts at all, because none were measured: the derivation never ran a date query
for it, and the zeros in its row mean *not measured*, not *measured zero*.

The status is 404 rather than 410 on purpose. Scope is re-derived from live occurrence records,
so a bird that misses the bar on one run can clear it on the next as more observations are
published. "Gone permanently" is not a claim the derivation supports.

**Names** come from AviList v2025b, in two vocabularies: eBird/Clements (Dovekie) by default and
AviList/IOC (Little Auk) on request. The **species id in the URL is the eBird code** — `amerob`,
not `turdus_migratorius` — which does not change when you switch vocabulary, because it is
assigned rather than derived from the English name.

**Taxonomy is AviList's**: names, families, and the order families are listed in. Where AVONET's
BirdLife taxonomy disagrees, AviList wins, and the join is bridged by explicit synonym tables
rather than fuzzy matching. One table maps a current name back to the older label AVONET measured
under; the other maps BirdView's name forward to a newer label AviList adopted.

## Regions

Which regions a bird is listed in is **derived**, not published. No source states "this species
occurs in Pennsylvania" under a license BirdView can serve, so the list is computed from
occurrence records, and the rule is stated here.

**Regions are GADM level-1 units** — US states, Canadian provinces and territories, and the
outlying islands — plus five ocean boxes that are BirdView's own, described below.

**A species is listed in a region when it reaches 1 record per million of that region's bird
records**, from GBIF under CC BY 4.0. The bar decides only where it is listed. Records in a
region where it stays under the bar still count in its *1 in N*: Emu is listed nowhere, and its
38 records, 19 of them in Texas, are on its tile.

**The bar is relative rather than a fixed count**, for two measured reasons. eBird's volume
doubled in five years, so any absolute bar admits more species over time for no ecological
reason. And effort differs enormously between regions — California has about 114 million species
observations against Hawaii's 2.5 million — so one count cannot mean the same thing in both.

**eBird first, and iNaturalist only where eBird is deliberately silent.** eBird's records are
reviewed before they are published, and they are 98.8% of the two sources' records combined, so
BirdView counts eBird. But eBird **withholds sensitive species** — Spotted Owl, Gyrfalcon and
three Hawaiian endemics everywhere; Long-eared Owl across the United States; Painted Bunting in
Florida from October to mid-April — and publishes its list of them. Where that list says eBird is
withholding a bird and the records confirm it, BirdView counts iNaturalist's records instead, and
the page says so. An eBird-only table would omit exactly the birds a reader is most likely to look
up; adding iNaturalist everywhere would also count escaped pets.

**Names are matched verbatim**, as the publisher wrote them, not through GBIF's taxonomic
backbone. The backbone silently drops what it cannot resolve and reports **zero** US and Canadian
records for Cooper's Hawk.

**Some species are listed only at sea, and that is not an error.** Townsend's Storm-Petrel,
Bermuda Petrel, Trindade Petrel, Craveri's Murrelet, Guadalupe Murrelet and Atlantic Yellow-nosed
Albatross are seen offshore, where no land polygon reaches, so their only regions are ocean
boxes. Six species on the index reach the bar in no region at all, Desertas Petrel and Little
Curlew among them; their detail pages show no Regions row rather than a row claiming the rule
placed them nowhere.

**The five ocean boxes are BirdView's own and approximate.** They are coordinate rectangles,
disjoint so a record falls in at most one, chosen to separate the birds rather than to reproduce
any published maritime limit.

### Known problems with the region rule

**The rule counts records, not birds, so a popular rarity outscores a quiet resident.** The more
surprising a bird is, the more people travel to see it, and each of them files a checklist.

Two examples in the current data: **American Flamingo is listed in Pennsylvania** on 931 records —
one or two birds at a single pond over thirteen days in September 2023, after Hurricane Idalia
scattered flamingos up the eastern seaboard. **Lucy's Warbler, a Sonoran desert bird, is listed in
Virginia** on 53 records from one location over six days around New Year 2016–17.

So **a region list is a list of places a bird has been reported often enough to clear the bar,
not a range map**, and at the margin it contains birds that turned up once and drew a crowd.

**A better threshold would not fix this.** Every rule that counts *how many* records a bird has
fails the same way, because a count measures observer effort: pelagic species and remote-island
endemics are under-reported, since few observers are where they are, while vagrants are
over-reported, since many observers go to see them. Those errors run in opposite directions, so
no single threshold corrects both. A critically endangered endemic with four records and a
vagrant with four records look the same to a count.

What separates them is **recurrence**: not how many records, but whether they repeat across
years. The 2023 flamingo has 931 records inside thirteen days; Akikiki, a Hawaiian endemic, has
376 across forty-five years and is clearly the resident of the two. **Scope** uses recurrence for
this reason. **Per-region presence** still uses the 1-per-million record bar, and that is a known
limitation.

## How common a bird is

Each bird shows how many of the observations recorded in the selected regions are of that bird,
as **one in N**.

### What an observation is

**One observation is one species on one checklist.** That is both sources' own unit and their own
word for it: eBird's id is `OBS1912782147`, iNaturalist's is
`inaturalist.org/observations/333181443`. A birder who submits a list of thirty species makes
thirty observations, each with its own id.

**The number of birds is recorded and deliberately ignored.** An observation notes how many
individuals were seen, and BirdView does not read it. Twelve White-throated Sparrows on one list
is one observation, exactly like a single Northern Cardinal. A flock of fifty thousand Snow Geese
is one observation.

So *1 in N* measures **how often a bird is logged**, not how many birds there are, and the two
can differ widely: a species seen in ones and twos but on nearly every list outranks one that
turns up rarely in enormous flocks. The measure understates waterfowl and blackbirds and
overstates chickadees and cardinals.

It also means the denominator grows with how *many kinds* of bird people report, not with how
many birds are present. A region where observers list thirty species a visit contributes more
records per visit than one where they list five, and dividing by that region's own total cancels
that out.

### Turning *1 in N* into *how often you would see it*

**Divide the average checklist length by N.** If the typical outing logs *S* species, a bird that
is 1 in *N* of all observations appears on roughly *S/N* of outings. The arithmetic is exact: a
bird on a fraction *f* of checklists holds *f/S* of the observations, so *f = S/N*.

**American Robin is 1 in 32 observations in New York.** Measured on two complete New York
county-days, the average checklist ran 15.6 and 19.1 species (reconstructed from observer, date
and place, because GBIF's eBird export carries no checklist id). So the robin is on roughly
**half to three-fifths of New York outings**, which matches a birder's experience better than
*1 in 32* suggests.

Where fewer species are about, checklists are shorter, and the same *1 in N* means a smaller
fraction of outings. The ordering is unaffected, being the same division by the same denominator
everywhere, but the intuitive reading of the number changes.

BirdView does not publish the per-outing figure, because it does not know the checklist length for
every region. Without a checklist id, each checklist must be reconstructed from observer, date and
rounded coordinates over a complete slice of the data, and a reconstruction from a partial slice
reports almost every checklist as a single species. Two county-days can be measured this way;
every region cannot without downloading all of the records.

**On the index the scope is whatever regions you have selected**, so the figure follows your
filter. A Gyrfalcon is about 1 in 3.6 million species observations across all regions and about
1 in 82,000 in Alaska, roughly forty-four times commoner where it actually lives. Select a region
the bird is absent from and the figure falls, because those records join the denominator and none
join the numerator.

**On the detail page** the scope is every region at once, and the page also names the region
where the species makes up the most observations — for a Sharp-shinned Hawk, *Top region:
Yukon — 1 in 249 species observations there*. A bird with no records in the regions on screen reads *0
in* the full count of more than a billion species observations: a stated zero against a stated
denominator, because zero in a billion records and zero in nine hundred are different findings.

**Birds at sea get a figure too, measured on open water.** The five marine regions are
latitude/longitude boxes, and they take in coastlines along with the water: the Pacific box spans
150°W to 115°W, which includes most of coastal California. So a sea counts only the records inside
it that fall on **no land at all**. GADM covers land, so a record on a beach carries a country and
a record on the water carries none. All six of the Black-capped Petrel's North Carolina records
sit inside the Atlantic box and stay with North Carolina.

A sea's denominator is therefore the same kind of number as a state's — every bird observation
there — so the two can be compared, and 44 species are commonest at sea. The two kinds never
overlap, one being *inside a GADM region* and the other *inside none*, so selecting both
double-counts nothing. The reasoning is recorded in
[the decision record](../river/requirements/decisions.md#honestseadenominator).

The figure is **on water** rather than strictly **at sea**: bay water carries no GADM country
either, so 8.8% of Chesapeake Bay records count as off land, and the open-ocean figure is
overstated by roughly 3%.

### When the sources disagree about a name

BirdView serves AviList's names; the records come from eBird and iNaturalist, and the three do not
always agree. Where they disagree the species page says so, because otherwise the number would
look ordinary when it is not.

**One name, two birds.** eBird has not adopted AviList's split of the Yellow Warbler: it files
both the northern and the mangrove bird under *Setophaga petechia*, and holds no records under
*Setophaga aestiva*. So its more than six million records cover two species, and **nobody knows
how they divide.** BirdView does not guess: both birds show the shared count, and each names the
other bird it shares the count with. The Elegant and Coppery-tailed Trogons are the same case.

**eBird withholds some species** to protect birds from trapping, disturbance or targeted hunting,
and publishes a list of which. Ten are withheld everywhere BirdView looks: Spotted Owl, Great Gray
Owl, Gyrfalcon, Northern Hawk Owl, Lesser Prairie-Chicken, Gunnison Sage-Grouse, Black Rail and
three Hawaiian endemics. Others are withheld only in some places or seasons: Long-eared Owl across
the United States, Painted Bunting in Florida from October to mid-April, Short-eared Owl in New
York in spring and summer. There, the counts are iNaturalist's alone and far below the real
record, the month filter does not judge them, and the page says *withheld by eBird*. That is a
gap in what BirdView can see, not a statement about the bird.

**The list is checked, not trusted.** It describes eBird's current policy, while the records
reflect what was applied when they were published; for 26 of the 230 places it names, eBird's
records are present anyway. BirdView marks a bird withheld only where the records show the gap.
The reasoning is recorded in
[the decision record](../river/requirements/decisions.md#ebirdsilencenotabsence).

Withholding and a different name look identical in the data — eBird returns nothing either way —
so the names were also checked by querying every plausible alternative.

**The same bird under another name.** Seven species are filed by eBird under a name AviList has
replaced, and the page names both. Five are genus moves: AviList puts the Hairy Woodpecker in
*Leuconotopicus*, and eBird files its more than seven million records under *Dryobates villosus*.
Two are corrected epithets, where the names differ by a letter or two: *Pica nuttallii* against
eBird's *Pica nuttalli*. **No count is affected either way**, and the page says so, so that a
reader is not left wondering what the disagreement changes.

The databases do not move together. iNaturalist follows AviList for all five genus moves, and for
the two spelling corrections neither database uses AviList's form. So the page names **eBird**,
which is the one thing true of every case: it holds nothing under the AviList name, which is why
the join needs the older one.

**Where the species boundary falls.** Three species are filed by eBird under a name that, in
AviList's taxonomy, belongs to a different species. **These are not all the same event, and the
page says which each one is.**

Two are splits eBird has not adopted. AviList separates the Hudsonian Whimbrel from the Eurasian,
and eBird files North American birds under *Numenius phaeopus* — 556,569 of them, against none
under *hudsonicus*. iNaturalist follows the split, with 17,923 records under *hudsonicus* and nine
under *phaeopus*. The same holds for *Gygis candida* against *Gygis alba*.

The third runs the other way, and is the only true **lump** here: AviList does not recognize
Stejneger's Stonechat as separate from the Siberian Stonechat, while both databases do. Its 95
eBird and 28 iNaturalist records are *Saxicola maurus*'s birds under the taxonomy BirdView serves.

In each case the two names do not overlap — each database holds almost nothing under the other's
name — so these are not two populations but one set of birds recorded under two names. BirdView
**adds** the records rather than choosing, and names the other label on the page.

**Feral and domestic birds.** iNaturalist files feral Rock Pigeons under *Columba livia
domestica* and domestic Muscovy Ducks under *Cairina moschata domestica*, below species rank.
BirdView folds a form into the species it belongs to, so those birds are counted here, and the
page says so. They are not separated, because no source separates them.

### A proportion, not a count

**A count says where the birdwatchers are; a proportion says where the bird is.** The two pick
different top regions for more than half the dataset. California holds 35 times as many
Sharp-shinned Hawk records as Yukon, but about 117 times as many bird observations overall. About
one bird observation in 250 from Yukon is a Sharp-shinned Hawk; in California it is about one in
830. The hawk is a larger part of what observers record in Yukon, even though far more have been
logged in California. Dividing by the region's own total cancels out how hard anyone was looking.

The Millerbird shows the difference clearly. 28 of its records are filed under Hawaii and 22 under
Nihoa, so by count its home reads "Hawaii", a state with 2.5 million species observations, where
the Millerbird is about 1 in 90,000. Nihoa has under 50,000 records and the bird is about 1 in
2,200 of them. Nihoa is the island it actually lives on.

**Why "1 in N" and not a percentage.** Across species and regions the figure runs from 1 in 2 to
more than 1 in 100 million. As percentages that is 50% down to below 0.000001%; the median bird
in its top region is about 0.26%, and many figures elsewhere round to zero. "1 in 514" stays
legible the whole way down.

**Both numbers are kept, not just the ratio.** The detail page gives the species' records and the
region's total, so a figure resting on nine records says so. This avoids a minimum-evidence
cutoff, which would be one more arbitrary threshold; a couple of dozen species have a top region
resting on fewer than ten records.

### What this is not

It is **not abundance**, and it is not a rarity rank. It measures how often a bird is *reported*
relative to other birds reported in the same place. That is the closest thing to "how often would
I see this" that open data supports, and it is still not the same claim.

Two distortions are known and uncorrected. Species whose records eBird **withholds**, such as
Spotted Owl, Gyrfalcon and Northern Hawk Owl, look rarer than they are, because the denominator
counts observers whose sightings of that species were suppressed. And a bird that draws a crowd
when it appears is over-reported relative to a quiet resident, the same observer-behavior bias
that keeps record counts out of the scope rule; see [Scope](#scope).

## Habitat and diet

Both come from **AVONET**, assigned per species, with AVONET's own definitions quoted in the
control. AVONET records **one** primary habitat and one primary diet per species, so a bird that
uses two of either is filed under one.

BirdView adds further values where that hides something, and **every addition is marked**: as
`birdview` where BirdView decided it, or `derived` where it follows from AVONET's own data by a
stated rule. Everything BirdView adds is listed together at [/added](/added).

**The derived rule reads AVONET against itself.** AVONET records a trophic *level* as well as a
trophic *niche*. Where the level is `Omnivore` but the niche names one specific food, AVONET is
saying the diet is mixed and the niche names only the larger part. For a seed-eater the other part
is invertebrates, which follows necessarily; for an insect-eater the plant part is a judgment
based on the family, and so is tagged `birdview` rather than `derived`.

**One category is BirdView's own.** AVONET's ten niches have no term for exudates, so the four
sapsuckers, a genus named for drilling and tending sap wells, would be filed under `Invertivore`
or `Omnivore` with the food they are known for missing. `Sap` fills that gap. The bar for adding a
term is that **no** AVONET term fits, not that none fits well: a bird that eats berries is a
`Frugivore` even if the word is clumsy. Birds that only drink from sapsucker wells — chickadees,
nuthatches, kinglets, several warblers — do **not** get it; sap is a supplement for them and a
staple for the sapsuckers.

---

## Where this is most likely to be wrong

In rough order of how much it would change the picture:

1. **The neck ratios are invented.** They are the largest unforced assumption in BirdView, and
   they drive the silhouette of every long-necked family. About a third of families carry their
   own ratio. The rest take the 0.18 default: each non-passerine family on it has been reviewed
   individually, and passerine families take it as a group, being compact songbirds with little
   exposed neck. All of it is judged from flight silhouettes, not measured. Böhmer et al. 2019
   does not solve this: it measures the full cervical column, including the part inside the body
   contour, while BirdView models the *exposed* neck of a silhouette, and converting one to the
   other needs an unmeasured per-family fraction
   ([source notes](../river/requirements/requirements.md#bohmer-neck)). Cervical vertebra counts
   do not work either: a count is dimensionless, so it can give proportions but never
   millimeters.
2. **The head model is unvalidated, and the body model is calibrated but not fitted.** The
   cranium's coefficient is a shape function fitted against nothing. The trunk's exponent is
   pinned to measured skeletons, but to *necks and legs*; nobody has published trunk lengths at
   this coverage. If either coefficient is off, every length is off in the same direction.
3. **Lengths are not comparable to published lengths.** They are internally consistent, which is
   what the tool needs, but do not quote them.
4. **Nearly two thirds of wingspans are modeled**: accurately, with a stated error, but modeled.
5. **The flight posture is uniform.** Every bird is drawn with the same implied wing position and
   the same body axis. Real birds differ greatly.
6. **One figure per species.** No sex, age, subspecies or individual variation, in a dataset where
   some species differ by 30% between sexes.

**If you know a source that fixes any of these**, particularly a validation set for head or trunk
length or an openly licensed body-length dataset, that would be the most useful contribution
anyone could make. Adopting a source follows a defined process: see
[river/requirements/requirements.md](../river/requirements/requirements.md) and the
source-adoption path in [docs/building.md](building.md).

Corrections are welcome, especially corrections to the biology: the models in this document are
where BirdView is most likely to be wrong. Report them on the project's issue tracker.
