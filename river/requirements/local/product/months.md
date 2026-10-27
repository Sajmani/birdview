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
the tile says *Withheld by eBird in Florida, Oct–Mar*. Judging them hid Painted Bunting from
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
Goals: [goals.flexibleComparison](../../../goals/goals.md#flexiblecomparison)

What is shown is what every active filter admits: region and family narrow the list together.
Selecting Hawaii and Parulidae shows the warblers of Hawaii, and shows nothing if there are none.

Reinstated with [regionAndMonthFilter](filtering.md#regionandmonthfilter).

An empty intersection is a real outcome: the page says no birds match, its count line naming each
filter and what it is set to, so a reader can tell which one emptied the page.

## regionsGroupedTwoWays

**Regions are grouped two ways**

Subject: `presentation.region_filter.zones`
Goals: [goals.localRelevance](../../../goals/goals.md#localrelevance), [goals.flexibleComparison](../../../goals/goals.md#flexiblecomparison)
Status: **Withdrawn 2026-10-27**, with the classic view, which the owner removed when the home page replaced it ([decisions.classicViewRemoved](../../decisions.md#classicviewremoved)).

The selector offers **two groupings** of the same regions, and the reader picks from either
([decisions.regionsTwoWays](../../decisions.md#regionstwoways)). Each is a **partition**: every region sits in exactly one zone
of each grouping, so a region appears exactly twice in the selector and no region is unreachable
from either.

**Common birds** is derived, not drawn: regions are clustered on how much their species lists
overlap, from the same presence table the filter uses. Four regions are left alone because
nothing is close enough — Hawaii's best match is 0.239 where every accepted placement is at least
0.652. A grouping that swallowed them would claim a similarity it did not measure.

**Geographic** is drawn, not derived, and says so. It is the grouping a reader already knows, so
someone looking for Alaska finds it under Alaska; the derived grouping scatters the Arctic.

**Geographic is offered first**, because a reader arrives wanting their own state.

**Neither grouping may be presented as the other.** The UI labels which is which, because a
hand-drawn grouping shown as measurement is the failure [decisions.regionsTwoWays](../../decisions.md#regionstwoways) was
written to correct.

Zone membership is **presentation only**. It changes which chips a group toggle selects and
nothing about which birds a region contains — that is [regionAndMonthFilter](filtering.md#regionandmonthfilter)'s 1-ppm rule, and it is
computed per region, never per zone.

## familiesByOrder

**Families are grouped by taxonomic order**

Subject: `presentation.family_filter.groups`
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
Goals: [goals.flexibleComparison](../../../goals/goals.md#flexiblecomparison)

Two more filters, **assigned per species**: where a bird mostly lives and what it mostly eats.
Both are multi-select and compose with region and family ([filtersCompose](#filterscompose)). Selecting two
options shows birds in **either**, because picking two places asks for the birds of both places
and not for the birds that use both.

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
