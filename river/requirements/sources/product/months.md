package product

# Months

**The reader also filters to any subset of the twelve months**, multi-select, on the same
control and the same rules as regions — everything selected on arrival, an **All** toggle,
and the two composing as [filtersCompose](#filterscompose) requires. A month appears in a URL by its
three-letter abbreviation: `?month=Jun&month=Jul`.

**Months decide which birds are drawn, not only the figure.** Every other filter in BirdView
narrows the list, and a control that silently did something else would be the worse surprise.

**A species is present in a region in a month** when it is present in that region at all by
the annual rule above, **and**

```
    (its records that month / all bird records that month)
  ÷ (its records that year  / all bird records that year )   >=  0.2
```

Dividing by the month's own total is not optional: New York records **20.9%** of its birding
in May and **4.8%** in December, so a raw share would make every bird a May bird. The ratio
asks whether a species is a bigger or smaller part of what people record this month than
across the year, and 1.0 means exactly as present as usual. It is [shareOfRecords](presentation.md#shareofrecords)'s own logic
at month granularity.

This is **not** the recurrence rule of [scopeIsRecurringAvifauna](dataset.md#scopeisrecurringavifauna), which is the better instrument and
cannot be had at the price — see [decisions.monthPresence](../../decisions.md#monthpresence) for the measurements, the
calibration, and why 0.2 is not load-bearing.

**Below 200 records in a region, the month filter abstains.** Millerbird has 28 records in
Hawaii, all in five months, and lives on Nihoa year round: what those months show is when
anyone last got to Nihoa. So a species with too little evidence is shown in **every** selected
month and the page says why, rather than making a statement about the bird out of a fact about
the survey. It costs a quarter of the species-region cells and **0.0345% of the records**,
because thin cells are thin.

**A month eBird withholds is not judged either.** Where eBird's Sensitive Species list keeps a
species out of its records in a region and season, and the records confirm it
([decisions.ebirdSilenceNotAbsence](../../decisions.md#ebirdsilencenotabsence)), those months are missing by policy: the species is shown, and
the tile says *Withheld by eBird in Florida, Oct–Mar*. Judging them was how Painted Bunting came
to be hidden from Florida in January, the heart of its season there.

**All twelve months selected must reproduce the annual figures exactly.** Every record carries
a month — 65,504,204 checked across both sources, none missing — so this is a check and not an
assumption.

**No month is grouped or named.** No seasons, no "breeding", no "migration".
[orderByFrequency](presentation.md#orderbyfrequency) already forbids naming a band where the range can be stated, and the reasons
compound here: season names invert between hemispheres, the Central Pacific region spans
−10° to 40° and crosses the equator, and for Hawaii, Palmyra, Navassa, Wake and Johnston
temperate season names are the wrong *kind* of category whatever the latitude.

**Six species belong to no region, and that is structural.** Bermuda Petrel, Desertas Petrel,
Trindade Petrel, Townsend's Storm-Petrel, Craveri's Murrelet and Guadalupe Murrelet are seen at
sea, where no land polygon reaches — Bermuda Petrel has 384 records offshore and **zero**
in any GADM region. They are not errors and must not be filtered away silently; a pelagic bird is
still a bird a reader may want.

## filtersCompose

**Filters compose**

Subject: `presentation.filters`
Goals: goals.flexibleComparison

What is shown is what every active filter admits: region and family narrow the list together.
Selecting Hawaii and Parulidae shows the warblers of Hawaii, and shows nothing if there are none.

Reinstated with [regionAndMonthFilter](filtering.md#regionandmonthfilter).

An empty intersection is a real outcome and is **not** the same as nothing being selected
([helpfulEmptyState](presentation.md#helpfulemptystate)). The empty state names each filter and what it is set to, because with two
filters a reader cannot otherwise tell which one emptied the page.

## regionsGroupedTwoWays

**Regions are grouped two ways**

Subject: `presentation.region_filter.zones`
Goals: goals.localRelevance, goals.flexibleComparison

The selector offers **two groupings** of the same regions, and the reader picks from either
([decisions.regionsTwoWays](../../decisions.md#regionstwoways)). Each is a **partition**: every region sits in exactly one zone
of each grouping, so a region appears exactly twice in the selector and no region is unreachable
from either.

**Common birds** is derived, not drawn. Regions are clustered on how much their species lists
overlap, from the same presence table the filter itself uses. Four regions are left alone
because nothing is close enough to group them with — Hawaii's best match is 0.239 where every
accepted placement is at least 0.652, and Navassa shares *zero* species with any other island.
Leaving them alone is part of the honesty: a grouping that swallows a region it cannot place is
claiming a similarity it did not measure.

**Geographic** is drawn, not derived, and says so. It is the grouping a reader already carries
in their head, there so that someone looking for Alaska finds it under Alaska. The derived
grouping scatters the Arctic — Alaska with the Pacific Northwest, Yukon and the Northwest
Territories with the Prairies, Nunavut alone — which is true of the birds and useless for
navigation.

**Geographic is offered first.** A reader arrives wanting their own state, and finds it fastest
on the map they already know; the derived grouping is the more interesting answer, but it is
the one you go looking for rather than the one you need on arrival.

**Neither grouping may be presented as the other.** The UI labels which is which, because a
hand-drawn grouping shown as measurement is the failure [decisions.regionsTwoWays](../../decisions.md#regionstwoways) was
written to correct: the previous zones were defended with a statistic that turned out to be an
artifact of a broken query.

Zone membership is **presentation only**. It changes which chips a group toggle selects and
nothing about which birds a region contains — that is [regionAndMonthFilter](filtering.md#regionandmonthfilter)'s 1-ppm rule, and it is
computed per region, never per zone.

## familiesByOrder

**Families are grouped by taxonomic order**

Subject: `presentation.family_filter.groups`
Goals: goals.flexibleComparison

The family selector groups families by **AviList's taxonomic order**, in checklist sequence.
Every family has exactly one order, so the grouping is a partition and no family is
unreachable.

**It is lopsided, and that is tolerated.** 43 of 90 families sit in Passeriformes under one
label and 15 orders hold exactly one family. For the other 25 orders the taxonomic name is
already what a birder would say — Waterfowl, Owls, Tubenoses, Grebes — so the whole awkwardness
is one order.

**A hand-drawn "birder groups" grouping was built to fix that, and withdrawn**
([decisions.habitatDietSpeciesLevel](../../decisions.md#habitatdietspecieslevel)). It was a family-level answer to a species-level question and
failed in that shape repeatedly: an Anhinga filed under seabirds, a frugivorous Phainopepla
among insect-eating flycatchers, a finch family called seed-eaters when only 13 of its 37
species eat seeds. The job it was trying to do — let a reader ask for birds of a kind — is done
per species by [habitatAndDietFilters](#habitatanddietfilters), where the data supports it.

**What survives is the lesson.** A category attached to a family must be true of *every*
species in it. Where the property varies within the family, no family-level tag is honest in
either direction: assigning one misfiles the rest, and assigning all of them makes the filter
return birds that were not asked for — tagging families with every habitat their species use
drops the precision of "Fresh Water" to **38%**, because one wetland warbler brings all fifty.
Such a property is a species trait and belongs on a species filter.

## habitatAndDietFilters

**Filter by habitat and by diet**

Subject: `presentation.trait_filter`
Goals: goals.flexibleComparison

Two more filters, **assigned per species**: where a bird mostly lives and what it mostly eats.
Both are multi-select and compose with region and family ([filtersCompose](#filterscompose)). Selecting two
options shows birds in **either**, because picking two places asks for the birds of both places
and not for the birds that use both.

**The classification is AVONET's, taken verbatim, and the definitions are quoted.** The values
are AVONET's own strings; the control shows a plain-English label beside each — "Eats insects"
for `Invertivore` — and AVONET's own definition as the description, **led by AVONET's own term
in bold**: *Invertivore: Species obtaining at least 60% of food resources from invertebrates…*

Naming the term does two jobs at once. It attributes the definition to the word it belongs to,
rather than letting a quoted sentence look like BirdView's own prose. And the term **is the
value in the URL** — `?niche=Invertivore` — so the control explains its own query string
without a separate note about it. Translating the word while
keeping the classification is a far lighter liberty than inventing the classification, and this
project's record on inventing classifications is [decisions.familiesTwoWays](../../decisions.md#familiestwoways).

**Per species is why this works where a family grouping did not.** A family-level habitat
misfiles 23–35% of birds, because habitat is a property of a species: Double-crested Cormorant
is marine and Neotropic Cormorant is not, Apapane sips nectar while American Goldfinch eats
seeds, and an Anhinga is a swamp bird whatever its relatives do. The same data at species level
is right about all of them.

**AVONET records one primary value per species; BirdView may add others.** A bird that uses two
habitats, or eats two things, gets one of each from AVONET. BirdView adds **as many further
habitats and diets as apply** — there is no limit of one — and tags each **`birdview`**, never
merged into the published value ([categoryProvenance](measurements.md#categoryprovenance)).

**BirdView may also add a term to the vocabulary, under a narrower rule: only where AVONET has
none that fits at all.** Not "none that fits well" — a bird that eats berries is a `Frugivore`
even if the word is clumsy. The case this exists for is **sap**: AVONET's ten niches have no
category for exudates, so the four sapsuckers are filed under `Invertivore` or `Omnivore` and
the thing the genus is named for disappears. A coined term **defines itself rather than quoting AVONET**, is excepted by name where the
control claims AVONET's classification, and can only ever be assigned with origin `birdview` —
a category this project invented cannot be published by anyone. It carries no badge: the
definition already says whose category it is, and a tag on one row of a list reads as a defect
rather than as provenance.

The threshold is **the bird's own living**, not contact with the food. Sapsuckers excavate
wells and defend them; chickadees, nuthatches, kinglets and several warblers are documented
drinking from those wells opportunistically, and they are **not** given this category. Sap is a
supplement to them and a livelihood to Sphyrapicus, and a category that cannot tell those apart
is not worth having — the same ≥60% standard AVONET holds its own niches to. A test caps how
many exist, because the difference between a gap and a habit is counted, not promised.

**Everything BirdView adds is listed in one place, at `/added`.** The detail pages say it bird
by bird, which answers *why is this robin a frugivore*; this answers *what has this project
decided on its own account*, which is the question someone auditing it would ask. A claim that
can only be found one species at a time is not really disclosed. Published assignments are
excluded — there are 1,678 and they are AVONET's.

What keeps this honest is not scarcity but **provenance plus evidence**. Every addition is
visibly this project's, and every addition must rest on something stateable: a different
AVONET taxonomy sheet that says so, or a fact derivable from BirdView's own data. An addition
whose only support is that it seemed right to whoever typed it is the failure
[decisions.familiesTwoWays](../../decisions.md#familiestwoways) was written about, and is not permitted however it is tagged.

**Counts beside each option are of the whole dataset, not the current view.** How many birds
there are of a kind is a fact about the data; how many are still showing is what the species
count above the list already says.

