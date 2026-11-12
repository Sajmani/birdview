package product

# Months

**The reader also filters to any subset of the twelve months**, multi-select, by the same rule as
regions, the two composing as [filtersCompose](#filterscompose) requires. On the home page no month
named means the current month, and `month=all` every month ([nextView](presentation.md#nextview)).
A month appears in a URL by its three-letter abbreviation: `?month=Jun&month=Jul`.

**Months decide which birds are drawn, not only the figure**, as every other filter narrows the
list.

**A species is present in a region in a month** when it is present in that region at all by
the annual rule above, **and**

```
    (its records that month / all bird records that month)
  ÷ (its records that year  / all bird records that year )   >=  0.2
```

Dividing by the month's own total is required: New York records **20.9%** of its birding in May
and **4.8%** in December, so a raw share would make every bird a May bird. A ratio of 1.0 means
exactly as present as usual. It is [shareOfRecords](presentation.md#shareofrecords)'s logic at month granularity.

This is **not** the recurrence rule of [scopeIsRecurringAvifauna](dataset.md#scopeisrecurringavifauna), which is the better instrument and
cannot be had at the price — see [decisions.monthPresence](../../decisions.md#monthpresence) for the measurements, the
calibration, and why 0.2 is not load-bearing.

**Below 200 records in a region, the month filter abstains.** Millerbird has 28 records in
Hawaii, all in five months, and lives on Nihoa year round; those months show when anyone got to
Nihoa. So a species with too little evidence is shown in **every** selected month and the page
says why. This affects a quarter of the species-region cells and **0.0345% of the records**.

**A month eBird withholds is not judged either.** Where eBird's Sensitive Species list keeps a
species out of its records in a region and season, and the records confirm it
([decisions.ebirdSilenceNotAbsence](../../decisions.md#ebirdsilencenotabsence)), those months are missing by policy: the species is shown, and
the tile says *Withheld by eBird in Florida (Oct–Mar)*. Judging them hid Painted Bunting from
Florida in January, the heart of its season there.

**All twelve months selected must reproduce the annual figures exactly.** Every record carries
a month — 65,504,204 checked across both sources, none missing — so this is a check and not an
assumption.

**No month is grouped or named.** No seasons, no "breeding", no "migration".
[orderByFrequency](presentation.md#orderbyfrequency) already forbids naming a band where the range can be stated, and season
names invert between hemispheres: the Central Pacific region spans −10° to 40°, and for Hawaii,
Palmyra, Navassa, Wake, and Johnston temperate season names are the wrong kind of category.

**Six species belong to no region, and that is structural.** Bermuda Petrel, Desertas Petrel,
Trindade Petrel, Townsend's Storm-Petrel, Craveri's Murrelet, and Guadalupe Murrelet are seen at
sea, where no land polygon reaches. They are not errors and must not be filtered away silently.

## filtersCompose

**Filters compose**

Subject: `presentation.filters`
Applies to: app
Goals: [goals.flexibleComparison](../../../goals/goals.md#flexiblecomparison)

What is shown is what every active filter admits: region and family narrow the list together.
Selecting Hawaii and Parulidae shows the warblers of Hawaii, and shows nothing if there are none.

Reinstated with [regionAndMonthFilter](filtering.md#regionandmonthfilter).

An empty intersection is a real outcome: the page says no birds match, its count line naming each
filter and what it is set to, so a reader can tell which one emptied the page.

## regionsGroupedTwoWays

**Regions are grouped two ways**

Subject: `presentation.region_filter.zones`
Applies to: app
Goals: [goals.localRelevance](../../../goals/goals.md#localrelevance), [goals.flexibleComparison](../../../goals/goals.md#flexiblecomparison)
Status: **Withdrawn 2026-10-27**, with the classic view, which the owner removed when the home page replaced it ([decisions.classicViewRemoved](../../decisions.md#classicviewremoved)).

The classic view offered regions under two groupings, one by the map and one derived from how much
their species lists overlap ([decisions.regionsTwoWays](../../decisions.md#regionstwoways)). The home page keeps only the map's zones,
which help a reader find a region and claim nothing about birds
([regionAndMonthFilter](filtering.md#regionandmonthfilter)). The derived grouping's code was deleted on 2026-11-09, when the owner
chose not to restore it: Here, and looking up a destination, already narrow to the birds one is
likely to see.

## familiesByOrder

**Families are grouped by taxonomic order**

Subject: `presentation.family_filter.groups`
Applies to: app
Goals: [goals.flexibleComparison](../../../goals/goals.md#flexiblecomparison)

The family selector groups families by **AviList's taxonomic order**, in checklist sequence.
Every family has exactly one order, so the grouping is a partition and no family is
unreachable.

**It is lopsided, and that is tolerated.** 43 of 90 families sit in Passeriformes under one
label and 15 orders hold exactly one family. For the other orders the taxonomic name is already
what a birder would say — Waterfowl, Owls, Tubenoses, Grebes.

**A hand-drawn "birder groups" grouping was withdrawn**
([decisions.habitatDietSpeciesLevel](../../decisions.md#habitatdietspecieslevel)): it answered a species-level question at family level and
misfiled birds, such as an Anhinga under seabirds. Letting a reader ask for birds of a kind is
done per species by [habitatAndDietFilters](#habitatanddietfilters).

**A category attached to a family must be true of *every* species in it.** Where a property
varies within the family, it is a species trait and belongs on a species filter: tagging families
with every habitat their species use drops the precision of "Fresh Water" to **38%**.

## habitatAndDietFilters

**Filter by habitat and by diet**

Subject: `presentation.trait_filter`
Applies to: dataset, app
Goals: [goals.flexibleComparison](../../../goals/goals.md#flexiblecomparison)

Two more filters, **assigned per species**: where a bird mostly lives and what it mostly eats.
Both are multi-select and compose with region and family ([filtersCompose](#filterscompose)). Selecting two
options shows birds in **either**, because picking two places asks for the birds of both places
and not for the birds that use both.

**Habitat's choices are arranged in five groups of AVONET's habitats, one group to a line, each
line led by a choice that takes the whole group** ([decisions.habitatGroupShortcuts](../../decisions.md#habitatgroupshortcuts)). Each group is an
exact union of AVONET's terms, so nothing is invented:

| Group | AVONET's habitats |
| --- | --- |
| Shore and sea | Coastal, and Marine as Sea near shore and Open ocean |
| Fresh water | Wetland, Riverine |
| Woods | Forest, Woodland |
| Open country | Grassland, Shrubland, Desert, Rock |
| Towns and farms | Human Modified |

**AVONET's Marine is offered as two refinements, Sea near shore (`Marine-inshore`) and Open ocean
(`Marine-offshore`), derived from where the bird's records lie**
([decisions.openOceanSplit](../../decisions.md#openoceansplit), [product.seaDistanceDerived](dataset.md#seadistancederived)). A bird with Marine, AVONET's or
BirdView's, is **Open ocean** when at least 25% of its records on water are more than 10 km from
shore, and **Sea near shore** when fewer are, or when at least 80% of all its records are on land
or within 10 km of shore; it can be both, as a tropicbird seen from its nesting cliffs and far out
at sea is. A bird with fewer than 100 records on water is unmeasured and is both, saying so. A
reviewed row of the habitat review may set the refinements by hand, with a reason, in place of the
derived ones. An address naming `Marine`, as the filter wrote before 2026-11-09, redirects
permanently to both refinements, which show the same birds; the heading names the pair "Open sea".

**A group is a shortcut, not a value.** The address carries only AVONET's habitats
(`habitat=Coastal&habitat=Marine`), and the filter matches them as before. A group's choice is
marked when every habitat in it is chosen; choosing it adds those not yet chosen, and choosing it
again removes them all. Each habitat on its line can still be chosen alone. Towns and farms holds
one habitat, so its line is that habitat alone. The group names form a column, aligned right against
the rule after them, which runs as one vertical line down to the last group, so every group's
habitats start at the same place; a line too long for the screen wraps under its own habitats, not
under the group name (Sameer, 2026-11-07). **All heads that column**, aligned right with the group
names, without the rule beside it (Sameer, 2026-11-09).

**Diet's choices are grouped the same way, by AVONET's own trophic levels**, so nothing is invented:
**Plants** (AVONET's Herbivore level: Eats seeds, Eats fruit, Drinks nectar, Eats plants, Eats water
plants, and BirdView's Drinks sap, sap being a plant food) and **Animals** (its Carnivore level: Eats
insects, Hunts animals, Hunts in water), each a shortcut leading its line, then Scavenges and Eats
anything, each a level of one niche and so its own line with no shortcut. All heads the column, and
the heading names a fully chosen group, "Diet: Plants" (Sameer, 2026-11-09). The page's heading names a fully chosen group by its
label, and any other chosen habitat by its own.

The groups answer "which birds am I likely to see here", and a birder stands in a place, not a
category: a Long-tailed Duck five feet off a beach is AVONET's Marine, the Purple Sandpiper on the
jetty behind it is Grassland in AVONET and coastal in fact, and a salt marsh is Coastal by AVONET's
own definition. The species page still names AVONET's exact habitat and BirdView's additions with
their reasons.

**The classification is AVONET's, taken verbatim, and the definitions are quoted.** The values
are AVONET's own strings, and the value in the URL — `?niche=Invertivore`; the control shows a
plain-English label for each — "Eats insects" for `Invertivore` — with AVONET's own definition as
its description: *Species obtaining at least 60% of food resources from invertebrates…* Translating the word
while keeping the classification is a far lighter liberty than inventing a classification, which
is what [decisions.familiesTwoWays](../../decisions.md#familiestwoways) records.

**Per species is why this works.** A family-level habitat misfiles 23–35% of birds: Double-crested
Cormorant is marine and Neotropic Cormorant is not.

**AVONET records one primary value per species; BirdView may add others.** BirdView adds **as many
further habitats and diets as apply** — there is no limit of one — and tags each **`birdview`**,
never merged into the published value ([categoryProvenance](measurements.md#categoryprovenance)).

**Every bird's habitats are reviewed, species by species, for where a North American birder sees
it** ([decisions.habitatReviewedPerSpecies](../../decisions.md#habitatreviewedperspecies)). AVONET's one value is in practice where a migrant breeds:
Black-bellied Plover is Grassland, the peeps miss Shore and estuary, the loons miss Open sea. So
`data/habitat-review.tsv` holds a row for **every species BirdView draws**: AVONET's habitat as it
stood when reviewed, the habitats BirdView adds with a reason for each, and the review's status,
reviewer, and date. Only a reviewed row's additions reach the page; a pending or proposed row
leaves AVONET's value alone. **The build fails when a drawn species has no row, when a row names a
species BirdView no longer draws, or when AVONET's value has changed since the review**, so adding
species or updating AVONET brings the review back to the reviewer rather than leaving it stale.

**BirdView may also add a term to the vocabulary, under a narrower rule: only where AVONET has
none that fits at all.** Not "none that fits well" — a bird that eats berries is a `Frugivore`
even if the word is clumsy. The case this exists for is **sap**: AVONET's ten niches have no
category for exudates, so the four sapsuckers are filed under `Invertivore` or `Omnivore`. A
coined term **defines itself rather than quoting AVONET**, is excepted by name where the control
claims AVONET's classification, and can only ever be assigned with origin `birdview`. It carries
no badge: the definition already says whose category it is.

The threshold is **the bird's own living**, not contact with the food — the same ≥60% standard
AVONET holds its own niches to. Sapsuckers excavate and defend sap wells; chickadees, nuthatches,
kinglets, and warblers that drink from them opportunistically are **not** given this category. A
test caps how many coined terms exist.

**Everything BirdView adds is listed in one place, at `/added`.** The detail pages say it bird
by bird; this page answers what the project has decided on its own account, which is what an
auditor asks. Published assignments are excluded — there are 1,678 and they are AVONET's.

Every addition must be visibly this project's and rest on something stateable: a different
AVONET taxonomy sheet that says so, or a fact derivable from BirdView's own data. An addition
supported only by seeming right is the failure [decisions.familiesTwoWays](../../decisions.md#familiestwoways) was written about,
and is not permitted however it is tagged.

**No option carries a count.** How many birds are still showing is what the count line says.
