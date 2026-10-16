package product

# Presentation

## diagramOnlyMode

**Diagram is the only presentation mode in v1**

Subject: `presentation.mode`
Value: `diagram`
Goals: goals.sizeIntuition, goals.shapeNotOnlySize

Rationale: photographs were scoped out of v1 (speciesPhotographs). The diagram is what makes size and shape
comparable; photographs were always the enrichment.

## tileContents

**Tile contents**

Subject: `presentation.tile`
Goals: goals.sizeIntuition, goals.comparisonAcrossFamilies, goals.shapeNotOnlySize

Each species is one **tile**, laid out top to bottom:

1. a header whose first line carries the **common name followed by its eBird species code in
   parentheses**, with the **scientific name right-justified on the same line**; and below it the
   **family badge**, showing the family's scientific name and its common name together;
2. the diagram, **vertically centered** in its area, carrying the **mass cube** in its lower-left
   corner ([decisions.massCubeOnTiles](../../decisions.md#masscubeontiles)) and **no labels of any kind**;
3. the main metrics — **length, wingspan, mass** — named in full and in the selected unit system.

**The family badge is the only badge, and the count is the rule.** A family is the one grouping
a tile names. The others — the birder groups and the taxonomic order of [familiesByOrder](months.md#familiesbyorder), and any
later grouping such as co-occurrence — are **navigation, and never appear on a tile**. Two
reasons, both already load-bearing elsewhere: every tile is one height and a second badge costs
a line on all of them, and the badge is part of the single diagram vocabulary
[oneDiagramVocabulary](diagram.md#onediagramvocabulary) holds the UI to. Stated as a count rather than a list because a list invites
one more entry.

The metrics are **spelled out**, not abbreviated to `L`, `W` and `M`. The initials cost a reader
who has not yet learned them, which is every reader on their first tile, and they save a line of
nothing: the three fit a tile at full width.

The **cube** is the one mark the tile shares with the annotated diagram. It needs no gutter and
no leader, only a corner the drawing does not use, so it does not cost the density that keeps
labels off the tile. It carries no label here — the footer already names the mass — and because
every tile in a section is drawn to one scale (scaleToggle), the cubes are directly comparable down the
whole section. That is where bulk against elongation is most worth seeing, since a section is by
construction a run of birds of similar length.

Tiles wrap to fill the available display width.

**Every tile is the same height, and nothing inside one may wrap to make it taller.** The header
is one line, the family badge is one line, the diagram has a fixed aspect, and the footer is one
line — so uniform height is a consequence of the parts rather than something imposed on top of
them.

Where text does not fit, it **truncates on that one line** rather than taking a second. What
gives way is ordered, and the order is about what a reader can reconstruct:

| | gives way |
| --- | --- |
| scientific name | first, and most readily |
| family common name, in the badge | next; the family's scientific name leads and survives |
| common name | last |
| **eBird species code** | **never** |

The species code is short — 4 to 8 characters, **not the fixed six its name suggests** — and it is
the one element that cannot be guessed from a truncation, so it is the one element that never
truncates. Nothing laying it out may assume a width.

Rationale: this is a layout rule in service of goals.sizeIntuition. A section is a run of birds drawn to one
scale, and reading down it — comparing diagrams, and comparing mass cubes against each other — is
what the index is for. A single tile standing taller than its neighbours puts every diagram in
that row at a different offset, which costs the comparison far more than an abbreviated word
does. Eight of the 841 common names are long enough to wrap, and the longest family badge comes
within a few pixels of the tile's inner width; one tile is enough to disturb a whole row.

In pick mode ([pickFromIndex](#pickfromindex)) each tile also carries a **pick control**, beside the tile's own
link rather than inside it. In browsing mode it is absent entirely, so the tile a reader sees
almost all the time is unchanged.

Rationale for centering: under global scale a small bird occupies a fraction of the tile's height,
and bottom-aligning it reads as a rendering fault rather than as the size difference it is.
Centering makes the empty space symmetric, so it reads as headroom rather than as misalignment.

Family common names are **not** unique across families — several groups share a vernacular
label — so they are a reading aid beside the scientific name, never an identifier.

## imperialByDefault

**Imperial by default, metric available**

Subject: `units.default`
Value: `imperial`
Goals: goals.sizeIntuition

One toggle switches the whole interface between:

| | length, wingspan | mass |
| --- | --- | --- |
| imperial (default) | inches | ounces, pounds above 16 oz |
| metric | centimeters | grams, kilograms above 1 kg |

**Every** displayed number honors it — tile footers, section headers and the detail view alike.

Rationale: the audience is predominantly US; the toggle serves Canadian users and anyone
who thinks in centimeters. Ounces below one are shown to two decimals, because a 3 g hummingbird
and a 12 g kinglet both round to 0.1 oz otherwise, erasing the distinction at exactly the size
where the app is trying to make it.

## firstLoadShowsAll

**First load shows everything**

Goals: goals.comparisonAcrossFamilies

The initial view has every family selected and shows all species, ordered by length.

Rationale: the full sweep from hummingbird to swan is the clearest statement of what the app is
for. Opening filtered or empty would hide it.

## togglesKeepPlace

**Toggling a control keeps your place**

Subject: `presentation.scroll`
Goals: goals.comparisonAcrossFamilies

Changing any control preserves the scroll position **and the collapsed state of the Presentation
and Controls panels**, as does returning from a detail view.

Rationale: every control is a link, so each toggle loads a new URL and the browser lands at the
top — which discards your place in a list of 841 birds. It matters most for the controls that do
not change the list at all: switching units or scale should leave you looking at the same birds.

The panels are the same defect seen from another angle. Their markup says open, so every reload
reopened one the reader had deliberately collapsed and pushed the page down. Panel state is
restored **before** the scroll, since collapsing one afterwards would move everything beneath it.

The restoration is the only client-side script in BirdView and holds no application logic;
ordering, sectioning and geometry stay in Go where they are property-tested (see
[arch.md](../../../arch.md)). Without JavaScript the page behaves as it did before.

## helpfulEmptyState

**The empty state says what to do**

Subject: `presentation.empty_state`
Goals: goals.comparisonAcrossFamilies

With nothing selected, the display area reads *"Select birds to display in Controls above"*.

**The active filters are named on the species-count line, whenever any are set.** What is
narrowing the list is worth knowing while the list still has birds in it, not only once it is
empty — a reader who has forgotten that a region is still selected should not have to open a
control to find out. A filter that is narrowing nothing is absent, because naming the arrival
state would describe a default as a choice.

**When something is selected and nothing matches, the message says so and no more.** Not "no
species match that filter" — that says a thing failed without saying which thing. The filters
are already named on the count line immediately above, so the two read as one statement:
*"0 of 839 species · Region Hawaii · Family Parulidae"* then *"No birds match."*, with the link
that clears them.

The two messages stay distinct, and the distinction is the point: **nothing selected is not the
same as nothing matched.** In the first case nothing failed — the reader turned everything off,
and the fix is to turn something on. In the second the filters are each reasonable and their
intersection is empty, and the fix is to relax one. A single message cannot say both.

Rationale: an empty page is reachable in one click, and with two filters an empty *intersection*
is reachable almost as easily — Hawaii and Parulidae is a plausible thing to ask for and a dead
end to arrive at. The description is **built once** and rendered on the count line, so the
filled and empty states cannot describe the same filters differently. An earlier version of
this requirement claimed the description was also used in the collapsed summary of a filter
control; it never was, and the claim is withdrawn rather than left as a spec describing
something that does not exist.

## collapsibleControls

**The controls collapse**

Subject: `presentation.controls`
Goals: goals.comparisonAcrossFamilies

The controls are **several collapsible panels**, each starting **closed** and each remembering its
own state across navigations (togglesKeepPlace): **Controls** for the display toggles, **Regions**, **Diet
and Habitat**, and **Families**.

They are separate boxes rather than one, because the filters are large — 90 family chips in
26 order groups, and 74 regions in 10 zones — and stacking them inside one panel pushes the birds
a long way below the fold while making "collapse the filter I am done with" impossible. Each
panel carries the documentation for its own controls (pageExplainsItself), so an explanation sits with the
thing it explains.

**A closed panel names what it holds.** Its summary carries a one-line description of what
opening it gets the reader — *narrow to states, provinces or stretches of open sea* — shown only
while closed, since once open the contents say it better. Without it a column of five bare words
asks the reader to open each in turn to find out which one they wanted.

Rationale: the birds are the product, and four open panels put them below the fold before a
reader has chosen anything. Opening a panel is one click and is remembered, so the cost falls on
the reader who wants a filter rather than on every reader who does not. A reader who has chosen
their filter should also be able to put it away and have it stay away; with several filters that
is not one decision but several, and one panel cannot express it.

## onlyPerGroup

**Every group offers "Only"**

Subject: `presentation.filter_only`
Goals: goals.comparisonAcrossFamilies

Beside each group selector in **Regions**, **Diet and Habitat** and **Families** sits an
**Only** control that selects that group and deselects everything else **within its own
section**. Clicking Only beside *The High Arctic* leaves Nunavut selected and every other region
dropped; Only beside *Waterfowl* leaves that order's families and no others. Diet and habitat
are separate sections, so narrowing one leaves the other whole.

It is a link like every other control (helpfulEmptyState), carrying the resulting selection in the URL, and
its target is built from the group alone rather than from what is currently selected — so it
lands on the same page whatever was on before, and is as shareable as any other view.

Rationale: the group toggle adds to or removes from the current selection, which is the right
default and a slow way to reach one group. From everything selected — the arrival state — a
reader wanting one zone must first clear all, then find their zone: two steps, and the first
empties the page. Narrowing to one group is the common case and was the only one that took more
than a click.

## oneUnitInComparison

**One unit across whatever is being compared**

Subject: `presentation.units`
Value: `mass unit fixed per comparison scope, smallest bird decides`
Goals: goals.sizeIntuition, goals.comparisonAcrossFamilies

Mass is written in **one unit across everything a page is comparing**. The scope is what the
page claims to compare: on the index, a section or the whole set, following the **scale**
setting ([scaleToggle](diagram.md#scaletoggle)); on the comparison page, the birds named in the address; on a detail
page, the one bird, which compares nothing.

**The smallest bird in the scope chooses the unit.** A scope holding one small bird is written
in the small unit throughout, so a Trumpeter Swan beside a House Sparrow reads `384 oz` and
`0.93 oz` rather than `24.0 lb` and `0.93 oz`.

Rationale: a unit chosen per value prints `14.5 oz` beside `1.2 lb` for two birds in the same
section — arithmetic the reader has to do, on the page whose whole purpose is comparison, and
the one thing this project exists to spare them. Taking the *largest* bird instead would crush
every small one to a single figure — a 27 g House Sparrow becomes `0.1 lb` — which is worse: a
number with no precision left beats one that merely needs converting.

Length never switches unit, so only mass is affected. The decimal places within a unit still
vary by value — a 3 g hummingbird is `0.11 oz` and one decimal would erase the differences
among the smallest birds — because that is precision, not a second scale.

## shareOfRecords

**Each bird says how much of what people record it is**

Subject: `presentation.share`
Value: `one species' observations over all species observations, for the regions and months in view`
Goals: goals.sizeIntuition, goals.comparisonAcrossFamilies

Every species says how much of what people record in view is that bird, as **one in N**.

**An observation is one species on one checklist** — both sources' own unit, and their own
word for it: eBird's id is `OBS1912782147`, iNaturalist's `inaturalist.org/observations/…`. A list of thirty species is thirty observations. The
individual count each carries is **deliberately not read**: twelve White-throated Sparrows is
one observation, and so is a flock of fifty thousand Snow Geese. That is precisely why the figure measures how often a
bird is *logged* rather than how many there are, and it is the sharpest statement of the
"not abundance" caveat below — a species seen in twos on every list outranks one that appears
rarely in vast flocks.

**The unit also makes *1 in N* read smaller than the thing a birder recognizes.** Divide the
average checklist length by N and you get the fraction of outings the bird is on: American
Robin at 1 in 32 in New York, where a checklist runs 12 to 19 species, is on roughly half to
two-thirds of them. BirdView does not publish that figure — the checklist id is absent from
GBIF's eBird export, so checklist length is reconstructable but not known per region — and
[`docs/data.md`](../../../../docs/data.md) gives readers the arithmetic instead.

**The unit on screen is “species observations”, not “bird observations”.** The denominator
counts *species on checklists*, and "bird" names the wrong thing in the one way a reader is
most likely to get wrong: it reads as a count of individual birds, which is exactly the
quantity the paragraph above says is carried and deliberately ignored. Someone who takes
`1 in 32` to mean one bird in thirty-two has not misread the number, they have misread the
unit, and the unit is the only place that can tell them. "Species" also makes the denominator
self-describing — a region where observers list thirty kinds a visit contributes more to it
than one where they list five — which is otherwise a fact available only in
[`docs/data.md`](../../../../docs/data.md).

Preferred over the alternatives on the same grounds as the sort-direction wording
([decisions.frequencyMetric](../../decisions.md#frequencymetric)): *records* is this project's internal word and says nothing to
a birder, *checklist entries* is accurate but names a mechanism the interface never shows, and
*sightings* re-opens the individuals ambiguity it is meant to close.

**The interface never uses the word “share”**, and neither does
[`docs/`](../../../../docs/data.md). It is the right technical term and it is this requirement's
subject, so this document and [`decisions.md`](../../decisions.md) keep it — but a reader meeting
`1 in 282,315 species observations in North Atlantic` does not need a noun for the concept, and the
one word that was on screen, in a one-line cell with no room to define it, read as jargon. The
pages say what the number *is* instead: *how much of what people record there is this bird, not
how many of them are out there.*

The exemption for this bundle is narrower than it looks and does not follow
[acceptance.noContactAddressPublished](../../acceptance.md#nocontactaddresspublished)'s precedent, where exempting the spec was the wrong fix. An
address published anywhere is harm wherever it appears; a defined term inside the document
that defines it is not.

**The share ranges over every region, land and sea** ([decisions.honestSeaDenominator](../../decisions.md#honestseadenominator)). A
sea's denominator is every bird observation on **open water** inside its box: GADM covers land, so
a record on a beach carries a country and a record on the water carries none, and the
difference is the water. That is the same quantity a land region's total is, so the two may be
compared and summed.

The two are also **disjoint by construction** — one is *inside GADM region X*, the other is
*inside no GADM region at all* — so a selection mixing them double-counts nothing and the
sums still compose.

A region BirdView cannot supply a denominator for contributes to **neither** side of the
ratio. No region is in that state today; the rule stands because the numerator and the
denominator were once two separate conditions that drifted apart, and deriving both from
*does this region have a denominator* makes that unrepresentable rather than merely fixed.

**On the index the scope is the regions and months currently selected**, so the figure
follows the filter: a Gyrfalcon is 1 in 3,631,056 species observations (all regions) and 1 in
82,946 in Alaska. Narrowing the months narrows both halves of the ratio
([regionAndMonthFilter](filtering.md#regionandmonthfilter)) — a June figure is divided by June's birding, so a figure is never inflated
by a month nobody was out in. That is the question a filtered page poses — *of what people record where I am
looking, how much is this bird* — and it is why the index does not instead show where in North
America the bird peaks. Narrowing to Arizona and being told about Yukon answers something
nobody asked.

**Absence is a number, not a dash.** A bird with no records in the regions on screen reads
**`0 in 1,225,176,035 species observations across every land region`** — the same sentence as every
other tile, with a zero in it. The em-dash this replaced could not be told apart from a missing
figure, and it threw away the half of the answer that carries the weight: *how hard anyone
looked before concluding the bird was not there.* Zero in a billion records is a finding; zero
in nine hundred is not, and only the stated denominator distinguishes them.

**Having no denominator at all is a different message**: **`No data for these regions yet`**.
It reports a gap in BirdView, not a fact about the bird, and must never be worded as one. A
region with no total is skipped on **both** sides of the ratio, so the numerator comes back
zero as an artifact of the skip — Buller's Shearwater has 39,143 records in the North Pacific
and would read "no records" there if that sea's total went missing, which is what the marine
regions did for months. No selection reaches this today, every region having a total.

**On the detail page the scope is every region at once**, and the page also names the one
where the species peaks, labelled **Top region**. One bird has no filter to follow, and
where it is commonest is worth knowing once you are looking at it.

**The peak and the index share are one calculation, not two that agree.** Both come from the
same per-region counts and the same totals, so selecting the named region on the index
reproduces the detail page's figure exactly, and the peak is always a region the bird is
listed in. They were two tables with two denominators until
[decisions.honestSeaDenominator](../../decisions.md#honestseadenominator), and the pages contradicted each other in public.

**A bird watched from the shore counts to the shore.** The boxes overlap the land regions
rather than abutting them — all six of the Black-capped Petrel's North Carolina records fall
inside the Atlantic box — so a sea counts only what is on open water inside it, and the
coastal record stays with the coast. **45 species peak at sea** under that rule, and the
Atlantic Yellow-nosed Albatross is 1 in 282,315 in the North Atlantic rather than 1 in
1,177,757 in Maine, which was never where it lived.

What this does not reach is the estuaries: bay water carries no GADM country either, so the
figure is *on water* rather than *at sea* and overstates the open ocean by roughly 3%
([decisions.honestSeaDenominator](../../decisions.md#honestseadenominator)). [`docs/data.md`](../../../../docs/data.md) says so.

**Share, not record count: a count says where the birdwatchers are, a share says where the bird
is.** California holds 35× the Sharp-shinned Hawk records of Yukon and 118× the birdwatching,
so the hawk is 1 in 250 there and 1 in 835 in California. Dividing by the region's own total
cancels out how hard anyone was looking — the same normalization that stops an escaped
Budgerigar outranking an American Robin.

**One in N, not a percentage.** The share runs from 1 in 2 to 1 in 117 million; as percentages
that is 50% to 0.000001%, and the bottom quarter of the dataset rounds to nothing readable.

**Selecting a region a bird is absent from lowers its share**, because that region's records
join the denominator and none join the numerator. That is correct rather than a rounding
artifact: a bird is a smaller part of a wider world.

Rationale: a reader asking "how common is this bird" has had no answer at all. This is the
closest honest one open data supports — it measures how often a bird is **reported** relative
to other birds reported in the same place, which is not abundance. Two distortions are known
and uncorrected: species eBird withholds look rarer than they are, and a bird that draws a
crowd is over-reported against a quiet resident. [`docs/data.md`](../../../../docs/data.md) says both.

**It is not a rarity tag and must not become one.** No threshold divides common from rare; the
number is shown and the reader judges. Vagrancy stays a status ([statusIsPublished](dataset.md#statusispublished)), because the
metric that would express it does not work: the share of a species' records falling in one
arrival reads 73% for the Hawaiian peafowl and 75% for the 2023 Florida flamingo irruption, and
61% of per-region date facets are truncated, so any episode-derived number would be a floor for
most of the dataset rather than a count.

**Nothing is precomputed per selection, and nothing needs to be.** The figure is a ratio of two
sums over the regions in view, and sums compose from their parts, so 69 regions do not imply
2^69 subaggregates — the per-region counts already are the subaggregate, and any selection is
one pass over them.

## undrawnSpeciesAnswer

**A species BirdView does not draw still answers for itself**

Subject: `presentation.species_url.out_of_scope`
Value: `404 with a named bird and its measured reason`
Goals: goals.sizeIntuition, goals.comparisonAcrossFamilies, goals.goodDataCitizen

`/species/<id>` for a bird the derivation classified but left out of scope returns **404 with a
page that names the bird and says why it is not drawn**, rather than the bare "404 page not
found" an unknown path gets.

The page carries both English names, the scientific name, the status [scopeIsRecurringAvifauna](dataset.md#scopeisrecurringavifauna) derived,
and the evidence that status rests on:

- **vagrant** — the episodes, years and records in the region that decided it, against the bar
  it did not clear.
- **historical** — the same, plus **the date it was last recorded**, which is what makes the
  verdict checkable. Passenger Pigeon recurs across 89 episodes in 52 years and every one of
  them ends in 1902.
- **absent** — no counts at all. The derivation never measured this bird's dates, so printing
  its zeros as though they were a result would be the error [provenancePerMetric](measurements.md#provenancepermetric) exists to prevent.
  The page says the open record held nothing it could measure, and stops.

**404, not 410.** Scope is regenerated from live occurrence data, so a bird that is out today
can be in after the next run; "gone permanently" is a claim the derivation does not support.
404 says the server has nothing at this address now, which is exactly true.

**An id belonging to no species at all still gets the bare 404.** This page is for a bird
BirdView knows of and chose not to draw, and it would be a fabrication for a typo. The same
applies to the six rejected species AviList does not recognize: with no English name and no
stable id, the page would be an apology for a binomial, so none is served.

Rationale: the status table classifies far more species than the dataset holds — 1,426 against
1,083 — so the derivation already knows the answer to *where did this bird go* for 334 of the
birds it rejected, 328 of which AviList can also name. Throwing that away at the door and answering `404 page not found` discards a
measured result in favor of nothing. [decisions.compareUrlContract](../../decisions.md#compareurlcontract) settled the same question
for comparison links — name what was lost rather than refuse the page — and the reasoning
carries: a link is made to be kept, and a reader who kept one is owed the reason it stopped
working rather than left to wonder whether they mistyped.
[decisions.speciesUrlOutlivesSpecies](../../decisions.md#speciesurloutlivesspecies) records why the status code nevertheless differs from
`/compare`'s.

## orderByFrequency

**Birds can be ordered by how often they are recorded**

Subject: `ordering.frequency`
Value: `ascending by one-in-N, banded by decade`
Goals: goals.sizeIntuition, goals.comparisonAcrossFamilies

**Frequency** is a fourth ordering, over the number every tile already shows
([shareOfRecords](#shareofrecords)), and it follows the month selection exactly as it follows the region
selection — the number it sorts on is the one on the tile, whatever that is scoped to. Species run from the most often recorded to the least — ascending by the
printed N, so *1 in 43* comes before *1 in 1,252,475,518*, consistent with
[ascendingOrder](ordering.md#ascendingorder) ordering ascending by what is displayed.

**Sections are decades, anchored at powers of ten**: `1 in 10 – 99`, `1 in 100 – 999`, and so
on, nine of them over the whole dataset. Anchored absolutely rather than to the data, so a
tier means the same thing under every filter — the number follows the region selection, and a
band starting at whichever bird happens to be commonest today would make *1 in 1,000–9,999*
denote one thing nationally and another in Alaska. See [decisions.frequencyMetric](../../decisions.md#frequencymetric) for the
measurements.

**No tier carries a name.** Not "Common", not "Scarce". [shareOfRecords](#shareofrecords) refuses to let this
number become a rarity tag, and a named band is that threshold with a label on it. Headers
state the range, exactly as [sectionHeadersStateRange](ordering.md#sectionheadersstaterange) requires for every other metric.

**Birds with no records in the selection are kept**, in a final section headed *No records in
the regions selected*. They have no ratio and so no tier; dropping them would silently shorten
the page, and the whole point of [shareOfRecords](#shareofrecords)'s stated zero is that absence is a finding.

Rationale: the share has been on every tile since [shareOfRecords](#shareofrecords) and there was no way to
arrange the page by it — a reader could see that a bird was 1 in 282,315 but not ask *what else
is about that common*. Size answers "how big", and this answers the question a birder actually
starts with.

## namesRecordsFiledUnder

**A species says which names its records were filed under**

Subject: `presentation.name_provenance`
Value: `per-source binomial and count, shown where the sources disagree`
Goals: goals.goodDataCitizen, goals.dataAsItIs, goals.nothingDisappearsQuietly

Where the occurrence sources do not use the name BirdView serves, the species page says so,
naming **each source's binomial and how many records it holds under it**.

Four kinds of disagreement, and they are not interchangeable:

- **An alias.** A source files the same concept under another binomial — eBird's *Dryobates
  villosus* for AviList's *Leuconotopicus villosus*. The records are this species'. The page
  names the alias so a reader can find the bird in the source.
- **A name AviList does not carry.** eBird's *Saxicola stejnegeri* has no AviList concept; its
  records are *Saxicola maurus*'s under the taxonomy BirdView uses.
- **A shared name.** A source uses one binomial for what AviList treats as **two** species —
  eBird's *Setophaga petechia* covers both Northern and Mangrove Yellow Warbler, because eBird
  has not adopted the split. **The records cannot be divided**, and BirdView does not divide
  them. Both species say so, and both name the other.
- **A form filed below the species.** iNaturalist records the feral population under a
  trinomial — *Columba livia domestica* — which BirdView folds into the species it is a form
  of. The count on the page therefore includes domestic and feral birds, and says so.

**A source can also withhold a species in some places and seasons only.** eBird's Sensitive
Species list names regions and date windows as well as species — Painted Bunting in Florida
from October to mid-April, Short-eared Owl in New York from April to July. Where the records
confirm the withholding ([decisions.ebirdSilenceNotAbsence](../../decisions.md#ebirdsilencenotabsence)), the species page lists those regions
and months, and a tile whose regions include one says so: *Withheld by eBird in Florida,
Oct–Mar*. The count there is iNaturalist's alone, and far below the bird's real record.

**Every kind that applies to a species gets a row, and they are not an either/or.** The first
implementation was a chain that showed the shared name *or* the withheld note and would have
shown one of two had both ever applied. No species is in two of these tables today; the rule
is that the page states each one it is in, because a disclosure silently suppressed by another
disclosure is the failure [goals.nothingDisappearsQuietly](../../../goals/goals.md#nothingdisappearsquietly) names.

**The third case is the one this requirement exists for.** Attributing those records to either
daughter would assert a division nobody has measured — eBird holds 6,245,227 of them in the US
and Canada and distinguishes none. Silently leaving them on whichever daughter inherited the
binomial is worse still: that is how Mangrove Yellow Warbler, a Caribbean mangrove resident,
came to be the 51st most-recorded bird in a North American dataset, credited with 532,470
records in Ontario.

**A source that withholds a species is NOT a name disagreement, and must not be reported as
one.** The two are indistinguishable in the data — eBird returns zero either way — and mean
opposite things. eBird holds no record under *Strix occidentalis* because it suppresses
conservation-sensitive species, not because it calls the Spotted Owl something else; telling a
reader to look for another name would send them after one that does not exist. Eleven species
are in that state, and the page says the source withholds them.

Rationale: [goals.dataAsItIs](../../../goals/goals.md#dataasitis) and [goals.nothingDisappearsQuietly](../../../goals/goals.md#nothingdisappearsquietly). A reader who can see that two
sources disagree can discount the figure appropriately; a reader shown a reconciled number
cannot, and a reader shown nothing cannot even know to ask. The disagreement is also
not a transient to be waited out — eBird filed 827,098 records under the shared name in 2024
alone, so this grows by roughly that much a year.

## pageExplainsItself

**The page explains itself**

Subject: `presentation.help`
Goals: goals.sizeIntuition, goals.comparisonAcrossFamilies, goals.shapeNotOnlySize

Below the title, a **Presentation** panel carries two sections side by side, each half the page:
**the diagram**, then **where the numbers come from**.

The diagram section is **a real labeled diagram**, the same one the detail view draws, rather than
a list of colour swatches: every part named and measured against an actual bird. The numbers
section explains solid against dashed, the wing chord and its ticks, and the estimate marker.

**The controls explain themselves where they are.** Every control is described inside the Controls
panel, above the controls themselves, in two columns.

**Stated as a rule rather than a list, and the list is why.** This requirement used to enumerate
the controls it covered — "ordering metric, scale mode, section step, units, family filter" — and
by the time anyone looked, the panel had gained **Compare birds** and **Names** and the
enumeration had been updated for neither. An enumeration cannot notice the thing it omits: adding
a control and forgetting to document it leaves both the control and the requirement looking
complete. So the rule is that **a control is documented if and only if it exists**, which
[acceptance.everyControlDocumented](../../acceptance.md#everycontroldocumented) checks by comparing the two sets rather than by looking for names
written down here.

The "only if" half matters as much. Documentation for a control that has been removed is worse
than none, because it sends a reader looking for something that is not there.

The help is collapsible and starts open, so a returning user can dismiss it without losing it.

Rationale: nothing about the diagram is self-evident. A reader cannot be expected to know that the
amber stroke is the bill, that a dashed stroke means estimated, or that the crossbar is a true
wingspan rather than a decoration — and the scale toggle, the one control that changes what the
app is *saying*, is the least discoverable of all.

Showing the diagram rather than describing it is the point: a swatch list asks the reader to map
six colour chips onto a drawing, where a labeled example simply *is* the drawing with the answers
on it. That needs room, which is why the panel holds two half-page columns rather than three narrow
ones — and why the control descriptions moved down to sit with the controls they describe, where a
reader looking at a control finds its explanation without going back up the page.

## tileOpensDetail

**A tile opens a detail view**

Subject: `presentation.detail_view`
Goals: goals.sizeIntuition, goals.shapeNotOnlySize

Clicking a tile opens a detail view that is **a larger version of the same tile** — same header,
same diagram, same main metrics, same layout — with the bird drawn **as large as the view allows,
whatever scale the index is using**, plus:

- the measurement of **every individual diagram segment**: bill, head, neck, body, tail;
- the **wing chord**, which the crossbar draws as its solid outer portions (measuredVersusModeled) and which would
  otherwise be the one length visible in the drawing with no row explaining it;
- the **provenance of every metric shown**, main metrics and segments alike (metricProvenanceExplained).

**The detail view is a superset of the tile, and that is binding.** Every piece of information
a tile carries about a species appears on that species' detail page — not merely the header,
diagram and main metrics named above, but anything a tile ever gains. A reader who notices
something on a tile and opens the bird to learn more must never find less, and the page that
exists to be the complete record of a bird cannot be missing something the summary of it
shows.

Stated as an invariant rather than a list because the list is what failed. The *plus* above
was read as permission to add to the detail view and not as a floor under it, so when
[shareOfRecords](#shareofrecords)'s frequency figure was put on every tile it was never added to the detail
page: a reader could see `1 in 23,393 species observations (all regions)` on the Northern
Yellow Warbler's tile, open it, and find only `Top region — 1 in 2,986 there`. A different
number, in the same unit, for a scope they had not asked about.

**Where the tile's figure depends on the region selection, the detail page carries that
selection and reports the same scope.** A figure that silently changes what it ranges over
between one page and the next is worse than one that is absent, because nothing on either
page reveals the substitution.

Rationale: the tile is deliberately dense, because it exists to be compared against dozens of
others at a glance. That leaves no room to say where each number came from. Keeping the detail
view the *same* layout at a larger size means the user is reading a familiar thing with more of it
exposed, rather than a different screen they have to re-orient to.

The detail view does **not** inherit the index's scale. Under the global default a kinglet is a
few pixels tall, which is the honest comparison — but the one view devoted to that single bird has
nothing to compare it against, so carrying the scale there would leave nothing to look at.

## comparisonOneScale

**A comparison view holds several birds to one scale**

Subject: `presentation.compare_view`
Value: `/compare?species=<id>&species=<id>…`
Goals: goals.comparisonAcrossFamilies, goals.sizeIntuition, goals.shapeNotOnlySize

A **comparison view** draws any number of chosen species together **on a single shared scale**, so
they can be read directly against one another. It answers goals.comparisonAcrossFamilies's own question — *is a Merlin
bigger than a Blue Jay?* — which the index answers only for birds that happen to land in the same
section, and the detail view cannot answer at all.

**The species are named in the URL, as repeated `species` parameters**, so a comparison is a
shareable link like every other view. Repeated parameters rather than a path, because the view
takes **any number** of birds and a path would fix the count. They are drawn in the order given:
that order is the reader's choice and the page does not reorder it. A species named twice appears
once — a bird compared against itself is never what was meant.

**A *Sort by* control offers the index's three metrics** — length, wingspan, mass — and orders the
panels by the chosen one, ascending, as the index orders its sections. It appears only with two
or more birds, there being nothing to sort below that.

**Sorting rewrites the order in the URL rather than adding a sort parameter.** The link it builds
names the same species in the new order, so the rule above still holds without exception: what is
drawn is the order given, and there is no second place an ordering can live. A sorted comparison
is therefore an ordinary comparison URL, indistinguishable from one assembled bird by bird, and
shares and bookmarks as itself.

Because nothing is stored, a metric reads as **on** when the panels *currently* happen to be in
that metric's order — a statement about the arrangement on screen, not about a mode the page is
in. Remove a bird or add one and the highlight simply stops being true, which is the honest
result: the page never claims an ordering the panels do not have.

**The scale is set by the largest total length in the set**, and the crossbar by the largest
wingspan, exactly as scaleToggle does for a section. There is **no per-panel scale option**. Drawing each
bird to fill its own panel is what scaleToggle rejects outright, and in a view whose entire purpose is
comparison it would not merely fail to help, it would state something false.

Consequently a hummingbird compared against a swan draws the hummingbird a few pixels tall. That is
the honest answer and the view is for asking the question; a reader who wants to see the
hummingbird has its detail view one click away.

**Panels wrap.** With several birds they will not fit on one line at any useful size, so they flow
onto as many lines as they need. Comparing down a column works; comparing across a row that has
been silently squeezed does not.

**Layout, in order:** the diagrams first, as a wrapping row of panels; then the diagram key — the
explanation of solid, dashed, hollow and the chord ticks — **once for the page**, not once per
bird. It describes a drawing convention shared by every panel, and repeating it per bird would
crowd out the thing it explains.

**There is no provenance table.** Each panel instead **links to that species' detail view**, which
is where the source or model behind each number is named (metricProvenanceExplained).

This is a deliberate scoping of where provenance is *explained*, not a weakening of it. The compare
view still marks every value that is not measured — the stroke styles of measuredVersusModeled, the hollow head,
and the `~` on each label — which is exactly what tileMarksModeled requires of a surface showing numbers and a
diagram: mark the kind, *without having to explain why*. What it does not reproduce is the `model`
and `source` text of metricProvenanceExplained. The alternative was to repeat a full provenance table for every bird on
the page, which would bury the comparison the view exists to make, under the same reasoning that
keeps the tile dense.

## pickFromIndex

**Birds are picked for comparison from the index, and the picks are the mode**

Subject: `presentation.pick_mode`
Value: `?species=<id>&species=<id>… — present, even empty, means picking`
Goals: goals.comparisonAcrossFamilies

A reader assembles a comparison by **picking birds from the index**, where the diagrams and the
family filter already are, rather than from a separate chooser that would duplicate both.

**The presence of a `species` parameter is the mode.** There is no separate flag. With none, the
app behaves exactly as it always has. With one or more, every page is in **pick mode**: the index
offers a pick control on each tile, the index and the detail view carry a bar showing the picks,
and the detail view offers to pick the bird being read.

A **bare `species=`** — the parameter present with no value — means *picking, nothing picked yet*,
so unpicking the last bird does not silently drop the reader out of the mode mid-task. This is
the distinction familyFilter's family filter already draws between a parameter that is absent and one
that is explicitly empty, and it is drawn here the same way for the same reason.

That empty state is also an **entry point**. A **Compare birds** toggle in the controls turns
pick mode on with nothing picked, and off again — the same thing the bar's Clear does, by the
same link. It is for a reader who wants to compare before they have a particular bird in mind;
the detail view's control is for one who already does.

**The comparison view offers both ways back, one above the other.** *Back to picking birds*
returns to the index with the picks intact, to change them. *Stop comparing* returns to the
index with pick mode off — the same exit as the bar's Clear, by the same URL. The second is the
common case: a reader who has seen the comparison usually wants the index back, and offering
only the first left them to find Clear in a bar on another page.

It is **its own labelled control set**, and the button reads **On** or **Off**. Added as an
unlabelled button after the Section step percentages it read as a sixth step option, because
every other control in that row is a labelled set of choices. So the label names the control and
the button states the mode, rather than the button trying to do both.

**Picking does not move the reader.** A pick link navigates to the picked bird's own tile, so the
browser restores the reading position itself while loading the page. Every control here is a
link, so picking reloads an index of 841 inline diagrams; restoring the position in script
afterwards means painting the top of the list first and moving the reader after, which from deep
in the list is a long and visible jump. The scroll script stands down whenever the URL names an
anchor, so the two cannot fight.

**The tile still opens the detail view in both modes.** The pick control is a separate target
beside it, never a repurposing of the tile, so [tileOpensDetail](#tileopensdetail) holds without exception. A reader
mid-pick can always open a bird to look at it properly, which is the main thing they will want to
do while deciding.

**Picking preserves everything else.** The filter, ordering, scale and units all keep working, so
a reader can narrow to one family and pick three birds from a page showing only that family. The
bar lists picks the filter is currently hiding, because a filter that appeared to discard them
would be worse than useless.

**Vocabulary: a bird is *picked*, a family is *selected*.** Both sets are live at once in this
mode and they mean unrelated things — a reader may have three families selected and two birds
picked — so they do not share a verb.

### Unrecognised ids are dropped and named

An id that matches no species is **dropped, and the page says which ids it dropped**, on the index
and in the comparison view alike. Every link the page then builds omits it, so the set repairs
itself on the next click.

This is not typo-handling, it is rot-handling. Species ids are derived from scientific names and
this checklist renames them — the Sandhill Crane is `antigone_canadensis` and was
`grus_canadensis`. A comparison URL is made to be shared and kept, so it will outlive some of the
ids in it. See [decisions.compareUrlContract](../../decisions.md#compareurlcontract), which reverses an earlier 404.

### What a tile link carries, and the one thing it does not

A tile link carries the picks and the display options — `species`, `units`, `by`, `scale`, `step`
— because the detail view cannot otherwise offer to pick the bird, build a link to the set plus
that bird, or format a single value in the reader's own units.

It deliberately **does not carry the family filter.** `family` is the only parameter with no bound
on its length: a 90-family selection is about 1.7 KB, which across 841 tiles would add roughly
1.4 MB to one page. The consequence, accepted rather than overlooked, is that the in-page back
link restores the picks, the ordering, the scale and the units, but **not the filter**. The
browser's own Back button restores that exactly, because it replays the previous URL.

**The Compare link does carry it**, and so do the comparison view's back, stop and sort links. Each
appears once per page, so the size bound above does not apply, and dropping `family` there would
return a reader who picked from a filtered index to an unfiltered one.

## selfDescriptionStaysTrue

**What the project says about itself stays true**

Subject: `docs.currency`
Goals: goals.shapeNotOnlySize

**Every CC BY source whose output appears on a page is credited on that page.** This is a licence
condition, not a courtesy: AviList, AVONET and Fox are all CC BY 4.0, and attribution is the
thing they ask for in return.

It was not being met. The index page carried full credits, while **`detail.html` and
`compare.html` credited AVONET, Fox and the ABA and omitted AviList** — on pages whose every
English name, scientific name, family and species code comes from AviList. A reader could use
BirdView for an hour and never see the source of the words on the screen.

**User-facing documentation describes the project as it is, not as it was.** `README.md` and
`docs/` are read before the code and are the first thing a new reader trusts, so a stale claim
there is worse than no claim. Two kinds of staleness are forbidden:

- **A count that no longer matches the dataset.** `docs/data.md` claimed 841 species after
  [decisions.avilistReplacesAba](../../decisions.md#avilistreplacesaba) cut the list to 839.
- **A source credited for something it no longer supplies.** `README.md` listed the ABA Checklist
  as the source of "species list, names, families" after decisions.avilistReplacesAba moved names and families to
  AviList (at that point the ABA checklist supplied **scope** alone; [decisions.scopeFromRecurrence](../../decisions.md#scopefromrecurrence)
  then replaced scope too).

**Before any public deployment, the docs are checked against the spec**, since a release is the
moment the claims become promises to strangers rather than notes to ourselves.

The check is mechanical where it can be — counts and credits are comparable against the dataset
and the page — and a reading task where it cannot. [acceptance.creditsComplete](../../acceptance.md#creditscomplete) does the
mechanical part, and records plainly that prose accuracy is not something a test can hold.

## noPersonalAddress

**No personal address is published**

Subject: `deploy.contact`
Goals: goals.goodDataCitizen

No page BirdView serves carries a contact address. Reader feedback goes to the project's issue
tracker, whose URL is a deployment setting (`-feedback`) rather than a constant in the markup.

**Unset publishes nothing**, not a dead link. A site may legitimately go up before its tracker
exists; what it must not do is invite a reader somewhere that does not.

Rationale: the header carried the owner's personal mailbox for the project's whole life, spelled
out in words rather than written plainly. **That obfuscation is the argument against it.** It
was written that way because a plain address on a public page is a harvested mailbox — which is
a reason not to publish the address at all, not a reason to spell it oddly. Beyond the spam, an
issue tracker is the better channel: a bug report in public can be read, searched, corrected by
someone else and closed, where mail to one person is none of those things. The owner's personal
mailbox is not the project's contact point.

*(This requirement does not quote the address it removed, and that is not squeamishness. The
requirements are published with the repository, so a quotation here would republish exactly
what the rule forbids — which is how [acceptance.noContactAddressPublished](../../acceptance.md#nocontactaddresspublished) first failed, back when they
were also served at `/spec`.)*

**The `User-Agent` contact stays, and is a different thing.** [tech.identifiedClient](../tech/delivery.md#identifiedclient) requires outbound
requests to name a contact, and iNaturalist's API guidance asks for it specifically. That
address reaches a handful of API operators who may need to reach whoever is making the requests;
it is not published to visitors. Removing it would be discourteous and risks the project being
blocked.

**This extends to addresses that are not ours.** The source manifest recorded a source owner's
mailbox, which the served spec then republished to every visitor. A provenance record names who
owns a source; it does not need to hand out their mail.

## inaturalistLink

**The detail view links to the species on iNaturalist**

Subject: `presentation.detail_view.external_link`
Goals: goals.sizeIntuition, goals.goodDataCitizen

The detail view carries a link to the species' iNaturalist taxon page, so a reader who has just
learned how big a bird is can go and see one.

The link is shown **only where the match is confident**, which requires **both** names to agree:

- **Scientific name** — identical, or differing only in genus while sharing a species epithet, which
  is a generic reassignment one authority has made and the other has not.
- **Common name** — ABA's matches **any** English name iNaturalist holds for the taxon, not merely
  its preferred one, once case, punctuation and British spelling are normalized and any ABA
  parenthetical alternate is set aside.

A species that fails **either** test shows **no link at all** — not a search URL, not a genus page,
not a guess. Neither name alone is sufficient: the scientific name identifies the species and the
common name is the independent second opinion that stops a near miss becoming a link to the wrong
bird.

**Coverage is expected to be total.** Every in-scope species resolves under these rules, and a
species that stops resolving is a fault to investigate rather than an absence to accept — see
[decisions.fullInaturalistCoverage](../../decisions.md#fullinaturalistcoverage).

BirdView **links and never copies**. No photograph is embedded, hotlinked, cached or reproduced;
the reader arrives on iNaturalist, where each image carries its own license. tech.selfContainedServer forbids the
embedding independently.

Rationale: BirdView will never have photographs, and a reader surprised by a size wants to see the
bird. Linking is the whole answer. The confidence gate exists because a link to the *wrong* bird is
worse than no link on a page whose purpose is being careful about what it claims — and an absent
link is legible, where a search URL that always "works" hides the failure. See
[decisions.inaturalistLinkAdopted](../../decisions.md#inaturalistlinkadopted).
