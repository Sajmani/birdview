package product

# Presentation

## diagramOnlyMode

**Diagram is the only presentation mode in v1**

Subject: `presentation.mode`
Value: `diagram`
Goals: [goals.sizeIntuition](../../../goals/goals.md#sizeintuition), [goals.shapeNotOnlySize](../../../goals/goals.md#shapenotonlysize)

The diagram is the only way BirdView presents a species in v1. There is no photograph mode.

Rationale: photographs were scoped out of v1 (speciesPhotographs). The diagram is what makes size
and shape comparable; photographs were always the enrichment.

## tileContents

**Tile contents**

Subject: `presentation.tile`
Goals: [goals.sizeIntuition](../../../goals/goals.md#sizeintuition), [goals.shapeNotOnlySize](../../../goals/goals.md#shapenotonlysize), [goals.flexibleComparison](../../../goals/goals.md#flexiblecomparison)

Each species is one **tile**, laid out top to bottom:

1. a header whose first line carries the **common name followed by its eBird species code in
   parentheses**, with the **scientific name right-justified on the same line**; and below it the
   **family badge**, showing the family's scientific name and its common name together;
2. the diagram, **vertically centered** in its area, carrying the **mass cube** in its lower-left
   corner ([decisions.massCubeOnTiles](../../decisions.md#masscubeontiles)) and **no labels of any kind**;
3. the main metrics — **length, wingspan, mass** — named in full and in the selected unit system.

**The family badge is the only badge, and the count is the rule.** Other groupings — the birder
groups and the taxonomic order of [familiesByOrder](months.md#familiesbyorder), and any later one
such as co-occurrence — are **navigation, and never appear on a tile**: a second badge would cost a
line on every tile, and the badge belongs to the single vocabulary
[oneDiagramVocabulary](diagram.md#onediagramvocabulary) holds the UI to. It is stated as a count
because a list invites one more entry.

The metrics are **spelled out**, not abbreviated to `L`, `W`, and `M`: they fit at full width, and
initials cost every new reader.

The **cube** needs only a corner the drawing does not use, and no label, since the footer names the
mass. Every tile in a section is drawn to one scale (scaleToggle), so the cubes compare down a
section, a run of birds of similar length, which is where bulk against elongation shows.

Tiles wrap to fill the available display width.

**Every tile is the same height, and nothing inside one may wrap to make it taller.** The header,
the family badge, and the footer are one line each and the diagram has a fixed aspect. Where text
does not fit, it **truncates on that one line**, giving way in order of what a reader can
reconstruct:

| | gives way |
| --- | --- |
| scientific name | first, and most readily |
| family common name, in the badge | next; the family's scientific name leads and survives |
| common name | last |
| **eBird species code** | **never** |

The species code is 4 to 8 characters, **not the fixed six its name suggests**, and cannot be
guessed from a truncation, so it never truncates. Nothing laying it out may assume a width.

Rationale: [goals.sizeIntuition](../../../goals/goals.md#sizeintuition). One taller tile offsets every diagram in its row and spoils reading
down a section. Eight of the 841 common names are long enough to wrap, and the longest family badge
comes within a few pixels of the tile's inner width.

In pick mode ([pickFromIndex](#pickfromindex)) each tile also carries a **pick control**, beside the
tile's own link rather than inside it. In browsing mode it is absent.

Rationale for centering: under global scale a small bird fills a fraction of the tile, and
bottom-aligned it reads as a rendering fault; centered, the empty space reads as headroom.

Family common names are **not** unique across families, so they are a reading aid beside the
scientific name, never an identifier.

## imperialByDefault

**Imperial by default, metric available**

Subject: `units.default`
Value: `imperial`
Goals: [goals.sizeIntuition](../../../goals/goals.md#sizeintuition), [goals.openToEveryone](../../../goals/goals.md#opentoeveryone)

One toggle switches the whole interface between:

| | length, wingspan | mass |
| --- | --- | --- |
| imperial (default) | inches | ounces, pounds above 16 oz |
| metric | centimeters | grams, kilograms above 1 kg |

**Every** displayed number honors it — tile footers, section headers, and the detail view alike.

Rationale: the audience is predominantly US; the toggle serves Canadian users and anyone who thinks
in centimeters. Ounces below one are shown to two decimals, because otherwise a 3 g hummingbird and
a 12 g kinglet both round to 0.1 oz.

## firstLoadShowsAll

**First load shows everything**

Goals: [goals.flexibleComparison](../../../goals/goals.md#flexiblecomparison)

The initial view has every family selected and shows all species, ordered by length.

Rationale: the full sweep from hummingbird to swan is the clearest statement of what the app is
for. Opening filtered or empty would hide it.

## togglesKeepPlace

**Toggling a control keeps your place**

Subject: `presentation.scroll`
Goals: [goals.openToEveryone](../../../goals/goals.md#opentoeveryone)

Changing any control preserves the scroll position **and the collapsed state of the Presentation
and Controls panels**, as does returning from a detail view.

Rationale: every control is a link, so each toggle loads a new URL at the top of a list of 841
birds, and the panel markup reopens a panel the reader collapsed. Panel state is restored
**before** the scroll, since collapsing a panel afterwards would move everything beneath it.

The restoration is the only client-side script in BirdView and holds no application logic;
ordering, sectioning, and geometry stay in Go where they are property-tested (see
[arch.md](../../../arch.md)). Without JavaScript the page behaves as it did before.

## helpfulEmptyState

**The empty state says what to do**

Subject: `presentation.empty_state`
Goals: [goals.openToEveryone](../../../goals/goals.md#opentoeveryone)

With nothing selected, the display area reads *"Select birds to display in Controls above"*.

**Every setting is named on the species-count line, always**, defaults included, one box each,
labeled as its control is and in the order the controls appear: Order by, Scale, Units, Names,
Section step, Comparing (only while picking birds to compare), Month, Region, Habitat, Diet,
Family. A setting at its default says so — *Order by Length*, *Region All* — so the line describes
the whole view, and a reader who has forgotten a selected region need not open a control to see it.

**When something is selected and nothing matches, the message says so and no more.** The filters
are already named on the count line above, so the two read as one statement: *"0 of 839 species ·
Region Hawaii · Family Parulidae"* then *"No birds match."*, with the link that clears them.

The two messages stay distinct: **nothing selected is not the same as nothing matched.** The first
is fixed by turning something on; the second by relaxing a filter.

Rationale: an empty page is one click away, and an empty intersection almost as close. The
description is **built once** and rendered on the count line, so the filled and empty states cannot
describe the same filters differently.

## collapsibleControls

**The controls collapse**

Subject: `presentation.controls`
Goals: [goals.openToEveryone](../../../goals/goals.md#opentoeveryone)

The controls are **several collapsible panels**, each starting **closed** and each remembering its
own state across navigations (togglesKeepPlace): **Controls** for the display toggles, **Regions**, **Diet
and Habitat**, and **Families**.

The filters are large — 90 family chips in 26 order groups, 74 regions in 10 zones — so one panel
would push the birds far below the fold and could not put away one finished filter. Each panel
carries the documentation for its own controls (pageExplainsItself).

**A closed panel names what it holds.** Its summary carries a one-line description of what opening
it gets the reader — *narrow to states, provinces, or stretches of open sea* — shown only while
closed.

Rationale: the birds are the product. Opening a panel is one click and is remembered, so the cost
falls on the reader who wants a filter rather than on every reader.

## onlyPerGroup

**Every group offers "Only"**

Subject: `presentation.filter_only`
Goals: [goals.flexibleComparison](../../../goals/goals.md#flexiblecomparison)

Beside each group selector in **Regions**, **Diet and Habitat**, and **Families** sits an
**Only** control that selects that group and deselects everything else **within its own
section**. Clicking Only beside *The High Arctic* leaves Nunavut selected and every other region
dropped; Only beside *Waterfowl* leaves that order's families and no others. Diet and habitat
are separate sections, so narrowing one leaves the other whole.

It is a link like every other control (helpfulEmptyState), carrying the resulting selection in the
URL, and its target is built from the group alone, so it lands on the same page whatever was
selected before.

Rationale: the group toggle adds to or removes from the selection, so reaching one group from the
arrival state took two steps, the first of which emptied the page.

## oneUnitInComparison

**One unit across whatever is being compared**

Subject: `presentation.units`
Value: `mass unit fixed per comparison scope, smallest bird decides`
Goals: [goals.sizeIntuition](../../../goals/goals.md#sizeintuition), [goals.flexibleComparison](../../../goals/goals.md#flexiblecomparison)

Mass is written in **one unit across everything a page is comparing**. The scope is what the
page claims to compare: on the index, a section, or the whole set, following the **scale**
setting ([scaleToggle](diagram.md#scaletoggle)); on the comparison page, the birds named in the address; on a detail
page, the one bird, which compares nothing.

**The smallest bird in the scope chooses the unit.** A scope holding one small bird is written
in the small unit throughout, so a Trumpeter Swan beside a House Sparrow reads `384 oz` and
`0.93 oz` rather than `24.0 lb` and `0.93 oz`.

Rationale: a unit per value prints `14.5 oz` beside `1.2 lb`, leaving the reader arithmetic on a
page meant for comparison. The *largest* bird would turn a 27 g House Sparrow into `0.1 lb`, losing
all precision, which is worse.

Length never switches unit, so only mass is affected. Decimal places within a unit still vary by
value — a 3 g hummingbird is `0.11 oz` — because that is precision, not a second scale.

## shareOfRecords

**Each bird says how much of what people record it is**

Subject: `presentation.share`
Value: `one species' observations over all species observations, for the regions and months in view`
Goals: [goals.localRelevance](../../../goals/goals.md#localrelevance), [goals.honestData](../../../goals/goals.md#honestdata)

Every species says how much of what people record in view is that bird, as **one in N**.

**An observation is one species on one checklist**, both sources' own unit (eBird's id is
`OBS1912782147`, iNaturalist's `inaturalist.org/observations/…`). The individual count is
**deliberately not read**: twelve White-throated Sparrows is one observation, and so is a flock of
fifty thousand Snow Geese. The figure measures how often a bird is *logged*, not how many there
are.

So *1 in N* reads smaller than the fraction of outings a bird is on: American Robin at 1 in 32 in
New York, where a checklist runs 12 to 19 species, is on roughly half to two-thirds of them.
BirdView does not publish that fraction — GBIF's eBird export lacks the checklist id, so checklist
length is not known per region — and [`docs/data.md`](../../../../docs/data.md) gives the
arithmetic.

**The unit on screen is “species observations”, not “bird observations”.** "Bird" reads as a count
of individuals, the quantity deliberately ignored; "species" also says the denominator grows with
how many kinds observers list, as [`docs/data.md`](../../../../docs/data.md) explains. *Records*
is internal, *checklist entries* names a mechanism the interface never shows, and *sightings*
reopens the individuals ambiguity ([decisions.frequencyMetric](../../decisions.md#frequencymetric)).

**The interface never uses the word “share”**, and neither does
[`docs/`](../../../../docs/data.md); on screen it read as jargon. This document and
[`decisions.md`](../../decisions.md) keep it as the technical term. The pages say what the number
*is* instead: *how much of what people record there is this bird, not how many of them are out
there.* This exemption does not follow
[acceptance.noContactAddressPublished](../../acceptance.md#nocontactaddresspublished)'s precedent,
where exempting the spec was the wrong fix: a published address is harm anywhere; a term inside the
document that defines it is not.

**The share ranges over every region, land and sea**
([decisions.honestSeaDenominator](../../decisions.md#honestseadenominator)). A sea's denominator is
every bird observation on **open water** inside its box: GADM covers land, so a record on the water
carries no country. Land and sea totals are the same quantity and **disjoint by construction**, so
they compare and sum without double counting.

A region BirdView cannot supply a denominator for contributes to **neither** side of the ratio;
deriving both sides from that one condition keeps them from drifting apart. No region is in that
state today.

**On the index the scope is the regions and months currently selected**: a Gyrfalcon is 1 in
3,631,056 species observations (all regions) and 1 in 82,946 in Alaska. Narrowing the months
narrows both halves of the ratio ([regionAndMonthFilter](filtering.md#regionandmonthfilter)), so a
month nobody was out in inflates nothing. The index therefore does not show where the bird peaks.

**Absence is a number, not a dash.** A bird with no records in the regions on screen reads
**`0 in 1,225,176,035 species observations across every land region`**. The denominator says how
hard anyone looked — zero in a billion records is a finding; zero in nine hundred is not — and a
dash cannot be told from a missing figure.

**Having no denominator at all is a different message**: **`No data for these regions yet`**.
It reports a gap in BirdView, not a fact about the bird, and must never be worded as one: a region
with no total is skipped on both sides, so its numerator reads zero only as an artifact (Buller's
Shearwater has 39,143 records in the North Pacific). No selection reaches this today.

**On the detail page the scope is every region at once**, and the page also names the one
where the species peaks, labeled **Top region**.

**The peak and the index share are one calculation, not two that agree.** Both use the same
per-region counts and totals, so selecting the named region on the index reproduces the detail
page's figure exactly, and the peak is always a region the bird is listed in
([decisions.honestSeaDenominator](../../decisions.md#honestseadenominator)).

**A bird watched from the shore counts to the shore.** The sea boxes overlap the land regions — all
six of the Black-capped Petrel's North Carolina records fall inside the Atlantic box — so a sea
counts only open water. **45 species peak at sea**; the Atlantic Yellow-nosed Albatross is 1 in
282,315 in the North Atlantic rather than 1 in 1,177,757 in Maine. Bay water carries no GADM
country either, so the figure is *on water* rather than *at sea* and overstates the open ocean by
roughly 3% ([decisions.honestSeaDenominator](../../decisions.md#honestseadenominator));
[`docs/data.md`](../../../../docs/data.md) says so.

**Share, not record count: a count says where the birdwatchers are, a share says where the bird
is.** California has 35× Yukon's Sharp-shinned Hawk records and 118× its birdwatching, so the hawk
is 1 in 250 in Yukon and 1 in 835 in California.

**One in N, not a percentage.** The share runs from 1 in 2 to 1 in 117 million; as percentages the
bottom quarter of the dataset rounds to nothing readable.

**Selecting a region a bird is absent from lowers its share**, correctly: that region's records
join the denominator and none join the numerator.

Rationale: this is the closest honest answer open data gives to "how common is this bird": how often
it is **reported** relative to other birds in the same place, which is not abundance. Two known,
uncorrected distortions — species eBird withholds look rarer, and a bird that draws a crowd is
over-reported — are stated in [`docs/data.md`](../../../../docs/data.md).

**It is not a rarity tag and must not become one.** No threshold divides common from rare; the
number is shown and the reader judges. Vagrancy stays a status
([statusIsPublished](dataset.md#statusispublished)), because the metric that would express it does
not work: the share of a species' records in one arrival reads 73% for the Hawaiian peafowl and 75%
for the 2023 Florida flamingo irruption, and 61% of per-region date facets are truncated, so an
episode-derived number would be a floor for most of the dataset.

**Nothing is precomputed per selection, and nothing needs to be.** The figure is a ratio of two
sums over the regions in view, and sums compose, so any selection is one pass over the per-region
counts.

## undrawnSpeciesAnswer

**A species BirdView does not draw still answers for itself**

Subject: `presentation.species_url.out_of_scope`
Value: `404 with a named bird and its measured reason`
Goals: [goals.honestData](../../../goals/goals.md#honestdata), [goals.shareableViews](../../../goals/goals.md#shareableviews)

`/species/<id>` for a bird the derivation classified but left out of scope returns **404 with a
page that names the bird and says why it is not drawn**, rather than the bare "404 page not
found" an unknown path gets.

The page carries both English names, the scientific name, the status
[scopeIsRecurringAvifauna](dataset.md#scopeisrecurringavifauna) derived, and the evidence that
status rests on:

- **vagrant** — the episodes, years, and records in the region that decided it, against the bar
  it did not clear.
- **historical** — the same, plus **the date it was last recorded**, which makes the verdict
  checkable: Passenger Pigeon recurs across 89 episodes in 52 years, every one ending in 1902.
- **absent** — no counts at all. The derivation never measured this bird's dates, so printing its
  zeros as a result would be the error [provenancePerMetric](measurements.md#provenancepermetric)
  exists to prevent. The page says the open record held nothing it could measure, and stops.

**404, not 410.** Scope is regenerated from live occurrence data, so a bird out today can be in
after the next run; 404 says the server has nothing at this address now, which is exactly true.

**An id belonging to no species at all still gets the bare 404**; this page would be a fabrication
for a typo. Nor is it served for the six rejected species AviList does not recognize, which have no
English name and no stable id.

Rationale: the status table classifies 1,426 species against the dataset's 1,083, so the derivation
knows *where did this bird go* for 334 rejected birds, 328 of which AviList can name.
[decisions.compareUrlContract](../../decisions.md#compareurlcontract) settled the same question for
comparison links: a link is made to be kept, and a reader who kept one is owed the reason it
stopped working. [decisions.speciesUrlOutlivesSpecies](../../decisions.md#speciesurloutlivesspecies)
records why the status code nevertheless differs from `/compare`'s.

## orderByFrequency

**Birds can be ordered by how often they are recorded**

Subject: `ordering.frequency`
Value: `ascending by one-in-N, banded by decade`
Goals: [goals.localRelevance](../../../goals/goals.md#localrelevance)

**Frequency** is a fourth ordering, over the number every tile already shows
([shareOfRecords](#shareofrecords)), and it follows the month selection exactly as it follows the
region selection: it sorts on the number on the tile, whatever that is scoped to. Species run from
the most often recorded to the least — ascending by the printed N, so *1 in 43* comes before *1 in
1,252,475,518*, consistent with [ascendingOrder](ordering.md#ascendingorder) ordering ascending by
what is displayed.

**Sections are decades, anchored at powers of ten**: `1 in 10 – 99`, `1 in 100 – 999`, and so on,
nine of them over the whole dataset. Anchored absolutely, not to the data, so a tier means the same
under every filter. See [decisions.frequencyMetric](../../decisions.md#frequencymetric) for the
measurements.

**No tier carries a name**, not "Common", not "Scarce": a named band is a rarity threshold, which
[shareOfRecords](#shareofrecords) refuses. Headers state the range, as
[sectionHeadersStateRange](ordering.md#sectionheadersstaterange) requires for every other metric.

**Birds with no records in the selection are kept**, in a final section headed *No records in
the regions selected*. They have no tier, but dropping them would silently shorten the page, and
[shareOfRecords](#shareofrecords) treats absence as a finding.

Rationale: [shareOfRecords](#shareofrecords) put the share on every tile but gave no way to ask
*what else is about that common*, the question a birder usually starts with.

## namesRecordsFiledUnder

**A species says which names its records were filed under**

Subject: `presentation.name_provenance`
Value: `per-source binomial and count, shown where the sources disagree`
Goals: [goals.honestData](../../../goals/goals.md#honestdata)

Where the occurrence sources do not use the name BirdView serves, the species page says so,
naming **each source's binomial and how many records it holds under it**.

Four kinds of disagreement, and they are not interchangeable:

- **An alias.** A source files the same concept under another binomial — eBird's *Dryobates
  villosus* for AviList's *Leuconotopicus villosus*. The page names the alias so a reader can find
  the bird in the source.
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

**Every kind that applies to a species gets a row, and they are not an either/or.** No species is
in two today, but a disclosure suppressed by another is the failure
[goals.honestData](../../../goals/goals.md#honestdata) names.

**The shared name is the case this requirement exists for.** Attributing those records to either
daughter asserts a division nobody has measured; eBird holds 6,245,227 of them in the US and Canada
and distinguishes none. Leaving them on whichever daughter inherited the binomial is worse: it made
Mangrove Yellow Warbler, a Caribbean mangrove resident, the 51st most-recorded bird in the dataset,
with 532,470 records in Ontario.

**A source that withholds a species is NOT a name disagreement, and must not be reported as
one.** eBird returns zero either way, but the meanings are opposite: it holds no record under
*Strix occidentalis* because it suppresses conservation-sensitive species, not because it calls
the Spotted Owl something else. Eleven species are in that state, and the page says the source
withholds them.

Rationale: [goals.honestData](../../../goals/goals.md#honestdata). A reader who sees two sources
disagree can discount the figure; one shown a reconciled number, or nothing, cannot. The
disagreement is not transient: eBird filed 827,098 records under the shared name in 2024 alone.

## pageExplainsItself

**The page explains itself**

Subject: `presentation.help`
Goals: [goals.openToEveryone](../../../goals/goals.md#opentoeveryone)

Below the title, a **Presentation** panel carries two sections side by side, each half the page:
**the diagram**, then **where the numbers come from**.

The diagram section is **a real labeled diagram**, the same one the detail view draws, rather than
a list of color swatches: every part named and measured against an actual bird. The numbers
section explains solid against dashed, the wing chord and its ticks, and the estimate marker.

**The controls explain themselves where they are.** Every control is described inside the Controls
panel, above the controls themselves, in two columns.

**A control is documented if and only if it exists**, which
[acceptance.everyControlDocumented](../../acceptance.md#everycontroldocumented) checks by comparing
the two sets. It is a rule rather than a list because a list cannot notice what it omits.
Documentation for a removed control is worse than none: it sends a reader looking for something
that is not there.

The help is collapsible and, like every panel, starts closed
([collapsibleControls](#collapsiblecontrols)); its summary says what opening it gets you.

**A welcome line under the title offers a few things to try**, each a link, then a link to the
full list in the guide's *Some things to try*, and last **reset to defaults**, a link to the bare
address, which carries no setting. Every thing to try is also in the guide's list, so the two
cannot drift apart; the guide says more about each. It is shown while picking birds to compare
too, and its things to try then carry the picks, so following one does not lose them; reset to
defaults drops them, compare mode being a setting like any other (pickFromIndex).

Rationale: nothing about the diagram is self-evident — that the amber stroke is the bill, that
dashed means estimated, that the crossbar is a true wingspan — and the scale toggle, which changes
what the app is *saying*, is the least discoverable control. A labeled example puts the answers on
the drawing itself; that needs two half-page columns, so the control descriptions sit with the
controls.

## fitsAPhone

**Every page fits a phone held upright**

Subject: `presentation.fit`
Goals: [goals.openToEveryone](../../../goals/goals.md#opentoeveryone)

**Down to 360 CSS pixels wide, nothing on any page reaches past the edge of the window, and no text
is drawn outside the box it sits in**, with every collapsible panel open. A box is anything drawn
with a background or a border. Text that is cut off with an ellipsis, or that scrolls inside its own
container, is not outside its box: a table too wide for a phone may scroll sideways inside itself,
but the page may not.

Rows laid out for a laptop stack on a phone instead of shrinking past reading: a group's header
shares its line with its Only button and its members wrap below, and a category's definition goes
under its toggle.

Rationale: most people who reach birdview.site are on phones, and the first readers' first
complaint was text running out of its panels and region rows running off the screen
([#2](https://github.com/google/birdview/issues/2)). 360 pixels is a small Android phone; an iPhone
is 375 to 430.

## tileOpensDetail

**A tile opens a detail view**

Subject: `presentation.detail_view`
Goals: [goals.sizeIntuition](../../../goals/goals.md#sizeintuition), [goals.shapeNotOnlySize](../../../goals/goals.md#shapenotonlysize)

Clicking a tile opens a detail view that is **a larger version of the same tile** — same header,
same diagram, same main metrics, same layout — with the bird drawn **as large as the view allows,
whatever scale the index is using**, plus:

- the measurement of **every individual diagram segment**: bill, head, neck, body, tail;
- the **wing chord**, which the crossbar draws as its solid outer portions (measuredVersusModeled) and which would
  otherwise be the one length visible in the drawing with no row explaining it;
- the **provenance of every metric shown**, main metrics and segments alike (metricProvenanceExplained).

**The detail view is a superset of the tile, and that is binding.** Every piece of information a
tile carries about a species appears on that species' detail page — not merely the header, diagram,
and main metrics named above, but anything a tile ever gains. The *plus* above is a floor: when
[shareOfRecords](#shareofrecords)'s figure reached the tiles but not the detail page, a reader
opening a bird found a different number for a different scope.

**Where the tile's figure depends on the region selection, the detail page carries that
selection and reports the same scope.** A figure that silently changes what it ranges over between
pages is worse than one that is absent.

Rationale: the tile is dense so it can be compared against dozens of others, which leaves no room
to say where each number came from. The same layout, larger, lets the reader see a familiar thing
with more exposed.

The detail view does **not** inherit the index's scale: under the global default a kinglet is a few
pixels tall, and the view of one bird has nothing to compare it against.

## comparisonOneScale

**A comparison view holds several birds to one scale**

Subject: `presentation.compare_view`
Value: `/compare?species=<id>&species=<id>…`
Goals: [goals.flexibleComparison](../../../goals/goals.md#flexiblecomparison), [goals.sizeIntuition](../../../goals/goals.md#sizeintuition), [goals.shapeNotOnlySize](../../../goals/goals.md#shapenotonlysize)

A **comparison view** draws any number of chosen species together **on a single shared scale**, so
they can be read directly against one another. It answers [goals.flexibleComparison](../../../goals/goals.md#flexiblecomparison)'s question — *is a Merlin
bigger than a Blue Jay?* — which the index answers only for birds that happen to land in the same
section, and the detail view cannot answer at all.

**The species are named in the URL, as repeated `species` parameters**, so a comparison is a
shareable link of **any number** of birds. They are drawn in the order given, which the page does
not reorder. A species named twice appears once.

**A *Sort by* control offers the index's three metrics** — length, wingspan, mass — and orders the
panels by the chosen one, ascending. It appears only with two or more birds.

**Sorting rewrites the order in the URL rather than adding a sort parameter**, so what is drawn is
always the order given, and a sorted comparison is an ordinary comparison URL. Nothing is stored:
a metric reads as **on** when the panels *currently* happen to be in its order, so the page never
claims an ordering the panels do not have.

**The scale is set by the largest total length in the set**, and the crossbar by the largest
wingspan, exactly as scaleToggle does for a section. There is **no per-panel scale option**: scaleToggle
rejects drawing each bird to fill its own panel, and in a comparison it would state something
false. A hummingbird beside a swan is a few pixels tall; its detail view is one click away.

**Panels wrap** onto as many lines as they need rather than squeezing into one row.

**Layout, in order:** the diagrams first, as a wrapping row of panels; then the diagram key — the
explanation of solid, dashed, hollow, and the chord ticks — **once for the page**, not once per
bird.

**There is no provenance table.** Each panel instead **links to that species' detail view**, which
is where the source or model behind each number is named (metricProvenanceExplained).

This limits where provenance is *explained*, not whether it is shown. Every value that is not
measured is still marked — the stroke styles of measuredVersusModeled, the hollow head, and the
`~` on each label — as tileMarksModeled requires. The `model` and `source` text of
metricProvenanceExplained is left out because a table per bird would bury the comparison.

## pickFromIndex

**Birds are picked for comparison from the index, and the picks are the mode**

Subject: `presentation.pick_mode`
Value: `?species=<id>&species=<id>… — present, even empty, means picking`
Goals: [goals.flexibleComparison](../../../goals/goals.md#flexiblecomparison), [goals.shareableViews](../../../goals/goals.md#shareableviews)

A reader assembles a comparison by **picking birds from the index**, where the diagrams and the
family filter already are, rather than from a separate chooser that would duplicate both.

**The presence of a `species` parameter is the mode.** There is no separate flag. With none, the
app behaves exactly as it always has. With one or more, every page is in **pick mode**: the index
offers a pick control on each tile, the index and the detail view carry a bar showing the picks,
and the detail view offers to pick the bird being read.

A **bare `species=`** — present with no value — means *picking, nothing picked yet*, so unpicking
the last bird does not end the mode; familyFilter's family filter draws the same distinction between
an absent and an explicitly empty parameter.

That empty state is also an **entry point**: a **Compare birds** toggle in the controls turns pick
mode on with nothing picked, and off again, by the same link as the bar's Clear.

**The comparison view offers both ways back, one above the other.** *Back to picking birds*
returns to the index with the picks intact. *Stop comparing*, the common case, returns to the index
with pick mode off, by the same URL as the bar's Clear.

**Reset to defaults**, at the end of the welcome line (pageExplainsItself), is the one other link
on the index that leaves pick mode, since it resets every setting and compare mode is one. Every
other link keeps the picks.

It is **its own labeled control set**, and the button reads **On** or **Off**; an unlabeled
button after the Section step percentages read as a sixth step option.

**Picking does not move the reader.** A pick link navigates to the picked bird's own tile, so the
browser restores the position as it loads, instead of painting the top of an 841-diagram list and
then jumping. The scroll script stands down whenever the URL names an anchor.

**The tile still opens the detail view in both modes.** The pick control is a separate target
beside it, so [tileOpensDetail](#tileopensdetail) holds without exception.

**Picking preserves everything else**: filter, ordering, scale, and units. The bar lists picks the
filter is hiding, because a filter that appeared to discard them would be worse than useless.

**Vocabulary: a bird is *picked*, a family is *selected*.** Both sets are live at once and mean
unrelated things, so they do not share a verb.

### Unrecognized ids are dropped and named

An id that matches no species is **dropped, and the page says which ids it dropped**, on the index
and in the comparison view alike. Every link the page then builds omits it, so the set repairs
itself on the next click. Species ids derive from scientific names, which change — the Sandhill
Crane was `grus_canadensis` and is `antigone_canadensis` — and a comparison URL is made to be kept.
See [decisions.compareUrlContract](../../decisions.md#compareurlcontract).

### What a tile link carries, and the one thing it does not

A tile link carries the picks and the display options — `species`, `units`, `by`, `scale`, `step`
— which the detail view needs to offer the pick, link to the set plus that bird, and format values
in the reader's units.

It deliberately **does not carry the family filter.** `family` is the only parameter with no bound
on its length: a 90-family selection is about 1.7 KB, which across 841 tiles adds roughly 1.4 MB
to a page. So the in-page back link restores everything except **the filter**; the browser's Back
button restores that.

**The Compare link does carry it**, and so do the comparison view's back, stop, and sort links. Each
appears once per page, and dropping `family` there would return a reader who picked from a
filtered index to an unfiltered one.

## selfDescriptionStaysTrue

**What the project says about itself stays true**

Subject: `docs.currency`
Goals: [goals.honestData](../../../goals/goals.md#honestdata), [goals.goodDataCitizen](../../../goals/goals.md#gooddatacitizen)

**Every CC BY source whose output appears on a page is credited on that page.** This is a license
condition, not a courtesy: AviList, AVONET, and Fox are all CC BY 4.0. It covers `detail.html` and
`compare.html`, whose names, families, and species codes all come from AviList.

**User-facing documentation describes the project as it is, not as it was.** `README.md` and
`docs/` are read before the code and trusted first. Two kinds of staleness are forbidden:

- **A count that no longer matches the dataset**, such as 841 species after
  [decisions.avilistReplacesAba](../../decisions.md#avilistreplacesaba) cut the list to 839.
- **A source credited for something it no longer supplies**, such as the ABA Checklist credited
  for names and families after [decisions.avilistReplacesAba](../../decisions.md#avilistreplacesaba), or for scope after
  [decisions.scopeFromRecurrence](../../decisions.md#scopefromrecurrence).

**Before any public deployment, the docs are checked against the spec**, since a release makes the
claims promises to strangers.

The check is mechanical where it can be — counts and credits against the dataset and the page —
and a reading task where it cannot. [acceptance.creditsComplete](../../acceptance.md#creditscomplete)
does the mechanical part, and records that prose accuracy is not something a test can hold.

## noPersonalAddress

**No personal address is published**

Subject: `deploy.contact`
Goals: [goals.giveBack](../../../goals/goals.md#giveback), [goals.maintainerPrivacy](../../../goals/goals.md#maintainerprivacy)

No page BirdView serves carries a contact address. Reader feedback goes to the project's issue
tracker, whose URL is a deployment setting (`-feedback`) rather than a constant in the markup.

**Unset publishes nothing**, not a dead link. A site may go up before its tracker exists; it must
not invite a reader somewhere that does not.

Rationale: a plain address on a public page gets harvested, which argues for not publishing one
rather than obfuscating it. A public issue tracker is also the better channel: a report there can
be read, searched, corrected, and closed. The owner's personal mailbox is not the project's contact
point.

*(This requirement does not quote the address it removed: the requirements are published with the
repository, so a quotation would republish what the rule forbids, which is how
[acceptance.noContactAddressPublished](../../acceptance.md#nocontactaddresspublished) first
failed.)*

**The `User-Agent` contact stays, and is a different thing.**
[tech.identifiedClient](../tech/delivery.md#identifiedclient) requires outbound requests to name a
contact, and iNaturalist's API guidance asks for it. It reaches a handful of API operators, not
visitors; removing it would be discourteous and risks the project being blocked.

**This extends to addresses that are not ours.** A provenance record names who owns a source; it
does not publish their mail.

## inaturalistLink

**The detail view links to the species on iNaturalist**

Subject: `presentation.detail_view.external_link`
Goals: [goals.goodDataCitizen](../../../goals/goals.md#gooddatacitizen)

The detail view carries a link to the species' iNaturalist taxon page, so a reader who has just
learned how big a bird is can go and see one.

The link is shown **only where the match is confident**, which requires **both** names to agree:

- **Scientific name** — identical, or differing only in genus while sharing a species epithet, which
  is a generic reassignment one authority has made and the other has not.
- **Common name** — ABA's matches **any** English name iNaturalist holds for the taxon, not merely
  its preferred one, once case, punctuation, and British spelling are normalized and any ABA
  parenthetical alternate is set aside.

A species that fails **either** test shows **no link at all** — not a search URL, not a genus page,
not a guess. The common name is the independent check that stops a near miss on the scientific
name becoming a link to the wrong bird.

**Coverage is expected to be total.** Every in-scope species resolves under these rules, and a
species that stops resolving is a fault to investigate rather than an absence to accept — see
[decisions.fullInaturalistCoverage](../../decisions.md#fullinaturalistcoverage).

BirdView **links and never copies**. No photograph is embedded, hotlinked, cached, or reproduced;
the reader arrives on iNaturalist, where each image carries its own license.
[tech.selfContainedServer](../tech/delivery.md#selfcontainedserver) forbids the embedding independently.

Rationale: BirdView will never have photographs, and a reader surprised by a size wants to see the
bird. A link to the *wrong* bird is worse than none, and an absent link is legible where a search
URL that always "works" hides the failure. See
[decisions.inaturalistLinkAdopted](../../decisions.md#inaturalistlinkadopted).

## nextView

**The home page: one view of birds**

Subject: `presentation.next`
Goals: [goals.sizeIntuition](../../../goals/goals.md#sizeintuition), [goals.flexibleComparison](../../../goals/goals.md#flexiblecomparison), [goals.localRelevance](../../../goals/goals.md#localrelevance), [goals.openToEveryone](../../../goals/goals.md#opentoeveryone)

**The home page, `/`, draws one set of birds, at one scale, with no sections or tiers**
(Sameer, 2026-10-26; [scalableInterface](../../../todo.md#scalableinterface)). It was built as a
prototype at `/next` and replaced the index as the home page the same day. The index, with its
sections and tiers, is now **the classic view** at `/classic`: unchanged, its links staying on it,
and not linked from the home page except in its credits. A species page's **Back** goes back by the
browser's history, so the home page returns as it was; species links carry only the units and
names, which the species page draws with, and from a shared link Back goes to the home page in
those. Elsewhere in these requirements, "the index" names
the classic view.

- **The header.** One line says what BirdView draws, its "stick figures" linking to the guide to
  reading a drawing, then "Click a bird for its numbers and sources"; what solid and dashed mean
  is left to that guide and to each species page. Under it, **New here? Try** offers the classic view's things to try, its
  comparisons made with pins: the birds around you now (this address with `#here-and-now`: the
  reader's region found in the browser replaces the region, the month is set to Now, and every
  other setting is kept), small, medium, large, and huge birds (six pinned and shown
  alone, by length), all herons and ducks (every month), and more ideas.
- **The set.** With no filter, it is every bird recorded in the month, most often recorded first.
  The month is the server's own unless the address names some; `month=all` is every month. Every
  filter (family, region, month, habitat, diet) narrows the set, and pinned birds join it.
- **Pins.** Every drawn bird has a pin, a push pin in a box, filled when the bird is pinned, in
  the drawing's lower right corner, level with the mass cube opposite it on every tile. A
  pinned bird (`pin=<id>`) is drawn whatever the settings, wherever the order puts it, even past
  the first 100, and the count line says how many are ("First 100 of 1006 birds, plus 2 pins"). Pinning or unpinning leaves the page where
  it was. The pinned birds are listed under the find box in alphabetical order: a bird's name goes
  to its tile, and its pin unpins it. Unpinning returns the bird to the settings'
  choice; nothing removes a bird the settings choose. A box above the set finds birds by name as
  the search box does ([searchBox](filtering.md#searchbox)): as the reader types, it says how many
  birds match, and once there are 10 or fewer it lists them, each with its pin. Without
  JavaScript, Enter does the same. Pinning keeps the box's text, so the list stays open and
  several of its birds can be pinned and unpinned in turn, from the list or from their tiles;
  the list's **×** closes it, dropping the text.
- **Filters are the count line.** The line under the box names each setting. Each is a control
  that opens the choices for that one setting directly under itself, as a `<details>`, so it
  works without JavaScript. **Clear pins**, there when something is pinned, unpins every bird and
  keeps the settings; **Reset**, after it, returns every setting to its default and keeps the
  pins. Family draws each family as its badge, with its description. Only one
  is open at a time, and a tap elsewhere closes it. A choice
  loads a new page, and the setting it was chosen in is open there, at the same place on the page,
  so several choices can be made in a row. Region
  offers **Here**, which finds the reader's region in the browser as on the index
  ([hereSelectsLocalRegion](filtering.md#hereselectslocalregion)). Month's choices sit on one
  line: **All**, **Now**, which names no month, so the page shows whatever month it is loaded in,
  a rule, and the twelve months. On a phone too narrow for them, the line scrolls sideways. A **Reset**
  control, colored apart from the settings, returns to the landing set. Choosing one option selects only it; choosing more adds them; choosing every
  selected option again returns to all. Nothing selected means all.
- **One scale.** Every bird shown is drawn to the scale of the longest. Its length, wingspan, and
  mass are text under the drawing.
- **Order.** By how often recorded (the default, most often first), or by length, wingspan,
  mass, bill, or tail (smallest first). Choosing the chosen order again reverses it (`rev=1`).
  Order by is the first setting on the count line.
- **Show.** Which birds are drawn is one setting, **Show**, with three answers, and one parameter,
  `show`: **First [100]** draws the first 100 of the order plus any pinned birds past them
  (`show=<n>`; 100 is the default, which the address leaves out), and the 100 is a box: typing
  another count and clicking First (anywhere on the choice but its box) or pressing Enter sets it,
  with or without JavaScript, and a blank, 100, or
  anything not a count returns to the default. **All N** draws every bird the settings match
  (`show=all`), N counting them with the pins. **Pinned only (k)** draws only the pinned birds
  (`show=pinned`), for a comparison of the reader's own picks. Wherever unpinned birds are left
  out (after the first, between pinned birds past them, or after the last pinned bird), a gap row
  says how many are hidden there and links to `show=all`. Choosing a setting keeps Show as it is,
  Pinned only included; Reset returns to the default. The count line is two rows: what is drawn
  ("First 100 of 1006 birds, plus 2 pins", "All 1006 birds", "2 pinned birds") with Show, Clear
  pins, and Reset; then Order by through Names. 100 is a reader's first
  screens, not a limit of speed: all 1,041 birds paint in 0.73 s on a phone-class CPU over 5G
  (measured 2026-10-26 under Lighthouse's mobile throttling, 229 KiB).
- **Every transition keeps what it does not change.** A setting's choice changes only that
  setting (Order by, also its direction); a pin, the pins; Show, show. Every other parameter is
  kept: Show, the pins, and the find box's text included. Only these drop more, on purpose:
  Reset keeps only the pins; Clear pins also leaves Pinned only; the find list's × drops its text;
  the birds around you now replaces the region, sets the month to Now, and leaves Pinned only. A
  tile's species page carries only the units and names; its Back returns by history.
- **Short summaries.** Each setting's summary is brief: the order's name and an arrow (↓ largest
  or most often recorded first, ↑ the reverse), months by three letters ("Oct (now)"), "all" for
  every region, "imperial" or "metric", "North Amer" or "world".
- **Stateless.** Every view is its address. The home page with no month changes with the month, so it is
  not stored or revalidated, as `/now` is not.

Rationale: the first readers and the simulated ones found the index dense and its toggles
backward, wanted search first, and could not read the comparison on a phone. A prototype beside
the index lets the two be compared before any requirement of the index changes. See the proposal
recorded with [scalableInterface](../../../todo.md#scalableinterface).
