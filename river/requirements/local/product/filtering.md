package product

# Filtering

## familyFilter

**Filter by family**

Applies to: app
Goals: [goals.flexibleComparison](../../../goals/goals.md#flexiblecomparison)

The user can filter to any subset of families, multi-select, by the home page's rule
([nextView](presentation.md#nextview)): **nothing selected means every family**, choosing a
family from all selects only it, choosing more adds them, and choosing a selected family again
removes it. No address can empty the page by filtering.

The selector is **grouped by taxonomic order, in the checklist's own sequence** rather than
alphabetically, each group labeled with the order's scientific name and its common name.

Order common names are **the project's own editorial labels**, since no source BirdView uses
publishes them; several are deliberately loose, and nothing is derived from them. Each family is
labeled with **both its scientific and its common name**.

**No control carries a species count.** One count is reported instead, **above the list**, naming
how many species the current filters match out of the whole dataset.

Rationale: a per-control count is only true while it is the only filter; with a region selected
([regionAndMonthFilter](#regionandmonthfilter)), "Parulidae 51" is false in Hawaii. Recomputing every chip gives 90
numbers to re-read after every click, to answer a question one number answers.

A birder looking for nightjars should not have to know they are Caprimulgidae, so the selector
shows common names. Checklist sequence is how field guides arrange birds, so it is the order a
birder already knows, and the source provides it for free.

## regionAndMonthFilter

**Filter by region and month**

Subject: `presentation.region_filter`
Applies to: dataset, app
Goals: [goals.localRelevance](../../../goals/goals.md#localrelevance), [goals.flexibleComparison](../../../goals/goals.md#flexiblecomparison)

The reader filters to any subset of US states and Canadian provinces, multi-select, as the family
filter works. A species is listed in a region when it reaches 1 record per million of that region's
bird observations, counted from eBird — or from iNaturalist only where eBird is confirmed to
withhold it there — and its records below the bar still count in its figure.

**A US county is a region too, listed under a stricter rule** ([decisions.countyListingRule](../../decisions.md#countylistingrule)): a
species is listed in a county when its state lists it, and its records in the county reach both one
per million of the county's bird observations and **at least 5 records**, over all years, counted
as for a state. A county narrows the set and orders it by the county's own share of records, as a
state does. One record per million is less than one record in every New York county, so the bar
alone would list a bird on a single sighting; the state's list keeps out the cage birds and one-off
vagrants a county's small total would let in. **The region a bird is "commonest in" stays a state,
province, or sea**: a county is inside a state, and naming one of thousands would rarely tell a
reader where the bird is common. **Counties are served only when the server runs with `-counties`**,
until every US state and Canadian province has them ([decisions.countiesBehindFlag](../../decisions.md#countiesbehindflag)). **A bird's page names
its states and provinces, each with how many of its counties list the bird out of how many it has**
("CA 12/58"), rather than every county ([decisions.countiesServedApart](../../decisions.md#countiesservedapart)).

It follows the same rule as families: nothing selected means every region. The data is GBIF's, which serves eBird and iNaturalist under CC BY 4.0 with GADM regions as a
queryable field ([decisions.regionFilterReinstated](../../decisions.md#regionfilterreinstated)).

**Membership is per region and relative.** iNaturalist is judged against its own total, and only
where eBird is confirmed to withhold the species there
([decisions.countEbirdFirst](../../decisions.md#countebirdfirst)). The bar is relative because eBird's volume doubled in five
years, so an absolute bar would admit more species over time for no ecological reason, and
because effort differs between regions — California has 114 million records against Hawaii's
2.5 million.

**Membership decides where a species is listed, not what it counts.** Records in a region where
the species stays below the bar are kept and counted in its figure ([shareOfRecords](presentation.md#shareofrecords)); they only
leave it off that region's list. Otherwise Emu, established on 19 eBird records in Texas, would
read *0 in 1,237,352,304* ([decisions.belowBarCounted](../../decisions.md#belowbarcounted)).

**Deciding membership per region keeps the endemics** without a special case: the Millerbird's
handful of records is nothing against a continental billion, and 22 per million within Hawaii.

**Regions are listed by geographic zone**, and zones mix the two countries — British Columbia
sits with Washington and Oregon, not with Nova Scotia. Zones are presentation only: which birds a
region holds is decided per region, never per zone.

**A region appears in a URL as the code its records are filed under**
([decisions.regionCodesFromData](../../decisions.md#regioncodesfromdata)): a land region as its **GADM code**, as GBIF attaches it,
at any level — `?region=USA.33_1` for New York, `?region=USA.33.24_1` for Kings County — and a sea
region as its **Marine Regions id**, prefixed — `?region=MRGID:25281` for the United States part of
the Gulf of Mexico. **No code is invented**: every code is the data's own, and the same code keys
the presence table, queries GBIF, and names the region in the URL.

**Sea regions are IHO sea areas divided by exclusive economic zone**, with a "High Seas of" piece
for each sea area beyond any zone, as Marine Regions publishes them. A sea area holds its parts as
a state holds its counties. GADM covers land only and no occurrence record names a sea, so
BirdView assigns offshore records to these areas itself.

**A value BirdView no longer writes redirects to the codes that replace it**, permanently and
without a message: a postal code or sea name from before 2026-11 (`NY` to `USA.33_1`), and a code
its source retires (to GADM's successors, or to Marine Regions' accepted id). A value that maps to
several codes redirects to all of them. The page's heading names the regions shown, so the reader
sees what the link now means.

## reportingFrequencyFilter

**Filter by how often a bird is reported where and when the reader is looking**

Subject: `presentation.common_filter`
Applies to: app
Goals: [goals.localRelevance](../../../goals/goals.md#localrelevance), [goals.flexibleComparison](../../../goals/goals.md#flexiblecomparison)

A **Reporting frequency** setting keeps only the birds that are **at least 1 in 100, 1 in 1,000, or 1 in
10,000 of the species observations in the regions and months chosen**, the share each tile states
([presentation.shareOfRecords](presentation.md#shareofrecords)); **All**, the default, keeps every bird. It composes with every other filter and
with every order, so the familiar birds of a place and season can be sorted by length or wingspan,
and a familiar bird lines up with the familiar birds of other families its size. Pinned birds stay
drawn whatever it is set to, so an unusual bird can be compared with the familiar ones.

It is a threshold on the share, not a count of birds, so it means the same in a county as across
the continent, in the steps the frequency tiers use ([orderByFrequency](presentation.md#orderbyfrequency)). The address carries it as
`freq=100`, `freq=1000`, or `freq=10000`; any other value is ignored. The setting is named for what
it measures, reports, not birds present or individuals counted
([decisions.reportingFrequencyParam](../../decisions.md#reportingfrequencyparam)). The page's heading names
it ("· At least 1 in 1,000"), and the setting follows Region (and County, where shown) on the count
line.

Rationale: sorting by frequency already shows the commonest birds first, but only in that order;
sorted by length, "the first 100" is the 100 smallest. Choosing the set by how familiar its birds
are, and ordering it by size, is what lets a reader calibrate an unfamiliar bird against familiar
ones, as birders do with an anchor bird (Sameer, 2026-11-07).

## hereSelectsLocalRegion

**Here selects the region the reader is in**

Subject: `presentation.region_filter.here`
Applies to: app
Goals: [goals.localRelevance](../../../goals/goals.md#localrelevance)

A **Here** button in the home page's Region setting replaces the region selection with the one
region the reader is in, every other setting kept. If the reader refuses permission, their location cannot
be found, or it falls in no region BirdView covers, the page says which of the three happened and
leaves the region selection unchanged.

The reader's position is tested against the land regions first and the sea regions second, so a
reader on a ferry in the Gulf gets the Gulf and a reader in Tampa gets Florida. The boundaries come
from a different source than the occurrence records ([tech.regionBoundaries](../tech/delivery.md#regionboundaries)), so within a few
hundred meters of a border or coast, Here may pick the neighbor.

**A link ending in `#here-and-now` presses Here on arrival, and sets the month to Now**: the
browser asks for the reader's location and goes to the page with their region and the month set
to the browser's current month ([nextView](presentation.md#nextview)). The fragment is never sent to
the server, so the link reveals nothing locationStaysInBrowser does not allow. If the region cannot
be found the page stays where the link landed and says why, beside the link that was followed
rather than inside the Region setting, which starts closed. Without JavaScript the link simply
lands.

**Without JavaScript, Here is shown inactive** — dotted border, no background, text at the full
contrast floor ([decisions.oneContrastFloor](../../decisions.md#onecontrastfloor)) — and focusable, with a tooltip saying it needs
JavaScript and permission to use the reader's location. It does nothing when pressed. Dotted,
because a dashed border once marked buttons that were active.

## nowSelectsCurrentMonth

**Now selects the current month**

Subject: `presentation.month_filter.now`
Applies to: app
Goals: [goals.localRelevance](../../../goals/goals.md#localrelevance)
Status: **Withdrawn 2026-10-27**, with the classic view, which the owner removed when the home page replaced it ([decisions.classicViewRemoved](../../decisions.md#classicviewremoved)).

A **Now** button beside **All** in Months replaces the month selection with the current month on
the reader's own clock, every other setting kept. Without JavaScript it uses the server's month.
The Months control is two lines: the label, All, Now, and the explanation of what a month means,
then the twelve months.

Without JavaScript the button links to `/now`, which redirects to the index for the server's
current month and is never cached; with it, the script points the button at the reader's month
directly, and `/now` is not requested. Either way no page carries a month taken from the server's
clock, which [tech.cacheLifetimeStated](../tech/delivery.md#cachelifetimestated) depends on ([decisions.hereAndNowCaching](../../decisions.md#hereandnowcaching)).

## locationStaysInBrowser

**The reader's location never leaves the browser**

Subject: `privacy.location`
Applies to: app
Goals: [goals.readerPrivacy](../../../goals/goals.md#readerprivacy)

The reader's position is used only inside the browser to choose a region, and is never sent,
stored, or logged. The region boundaries load with every home page whether or not Here is
pressed, and pressing Here produces exactly the request that choosing that region by hand would.

So the server cannot tell that Here was used, let alone where the reader is. Loading the
boundaries only when Here is pressed would leak the first; sending the position to the server
would leak both.

## searchBox

**A search box narrows the index as you type**

Subject: `presentation.search`
Applies to: app
Goals: [goals.flexibleComparison](../../../goals/goals.md#flexiblecomparison)
Status: **Withdrawn 2026-10-27**, with the classic view, which the owner removed when the home page replaced it ([decisions.classicViewRemoved](../../decisions.md#classicviewremoved)).

**A text box above the count line keeps the birds that match what is typed and hides the rest.** A
bird matches when what is typed, ignoring case and punctuation, appears unbroken within one of its
English names (North American and world), its species code, its scientific name, or the scientific
name of its family or order. So "d hawk" finds Sharp-shinned Hawk and the other "…d hawk"s, but
not the nighthawks, whose family is Caprimulgidae. Families and orders are matched by their
scientific names only, since their English descriptions list other birds: "Kites, Old World
Vultures, Eagles, and Hawks" would make "hawk" find every kite (Sameer, 2026-10-25, on both). The
count line, each section's count, and the histogram count only the matches, a section with none is
hidden, and the count line names the search.

**It filters in the browser as the reader types**, with no request to the server. Matches keep
their places: sections and scales are the ones the other filters made, so what remains does not
move while the reader types.

**The search goes into the address as `q` on the reader's next click that changes the address**
— a control, a tile, a pick — not while typing. The welcome line's things to try, *reset to
defaults*, and *Clear the filters* drop it. Opened from an address with `q`, the page shows the same
matches with or without JavaScript; without it, Enter in the box loads the page with the search.

Rationale: the first readers asked for a search bar to pick birds by name
([#3](https://github.com/google/birdview/issues/3)), and a box keeps working at ten times the
species where scrolling and chips do not. The words to match are prepared by the server, so the
rule lives in Go; the browser only looks them up, and a test holds it to the Go matcher (Sameer,
2026-10-25: filter in the browser; add `q` on the next click).
