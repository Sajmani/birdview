package product

# Filtering

## familyFilter

**Filter by family**

Goals: [goals.flexibleComparison](../../../goals/goals.md#flexiblecomparison)

The user can filter to any subset of families, multi-select. **Everything is selected on arrival**
— All, every order and every family show as on. An **All** control toggles like an order does,
over the whole set: untoggling it deselects everything and **no birds are displayed**.

Selection is therefore always literal: what is shown is what is selected. There is no separate
"unfiltered" state that happens to look like everything being chosen.

The selector is **grouped by taxonomic order, in the checklist's own sequence** rather than
alphabetically. Each group is headed by a **toggle that selects or clears every family in that
order**, labeled with the order's scientific name and its common name, and styled like the other
controls rather than like a family chip.

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
Goals: [goals.localRelevance](../../../goals/goals.md#localrelevance), [goals.flexibleComparison](../../../goals/goals.md#flexiblecomparison)

The reader filters to any subset of US states and Canadian provinces, multi-select, as the family
filter works. A species is listed in a region when it reaches 1 record per million of that region's
bird observations, counted from eBird — or from iNaturalist only where eBird is confirmed to
withhold it there — and its records below the bar still count in its figure.

As with families, everything is selected on arrival, with an **All** control and group toggles.
The data is GBIF's, which serves eBird and iNaturalist under CC BY 4.0 with GADM regions as a
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

**Regions are gathered into zones so a reader can select an area in one click**, and zones mix
the two countries — British Columbia sits with Washington and Oregon, not with Nova Scotia. What
a zone *means* is [regionsGroupedTwoWays](months.md#regionsgroupedtwoways).

**A region appears in a URL as its postal abbreviation, or as its name** — `?region=CA`,
`?region=GulfOfMexico` ([decisions.regionUrlIsPostalCode](../../decisions.md#regionurlispostalcode)). Where a standard abbreviation exists it
is used, and where none does the name is used in PascalCase. **No code is invented**: the obvious
ones are taken (Midway cannot be `MI`, Palmyra cannot be `PA`), and an invented code would sit
beside 64 official ones with nothing marking it as a guess. A region with no standard code shows
**no code** on its chip, only its name.

The GADM id stays the internal key — it keys the presence table and is what GBIF is queried
with — but does not belong in a URL. The slug is **derived** from the code or the name rather
than stored, so a renamed region cannot keep a URL that no longer describes it.

## hereSelectsLocalRegion

**Here selects the region the reader is in**

Subject: `presentation.region_filter.here`
Goals: [goals.localRelevance](../../../goals/goals.md#localrelevance)

A **Here** button beside **All** in Regions replaces the region selection with the one region the
reader is in, every other setting kept. If the reader refuses permission, their location cannot
be found, or it falls in no region BirdView covers, the page says which of the three happened and
leaves the region selection unchanged.

The reader's position is tested against the land regions first and the sea regions second, so a
reader on a ferry in the Gulf gets the Gulf and a reader in Tampa gets Florida. The boundaries come
from a different source than the occurrence records ([tech.regionBoundaries](../tech/delivery.md#regionboundaries)), so within a few
hundred meters of a border or coast, Here may pick the neighbor.

**A link ending in `#here-and-now` presses Here and Now on arrival**, together: the browser asks
for the reader's location and goes to the page with their region and the current month, exactly
as choosing both by hand would. The fragment is never sent to the server, so the link reveals
nothing locationStaysInBrowser does not allow. If the region cannot be found the page stays where
the link landed and says why, beside the link that was followed rather than inside the Regions
panel, which starts closed. Without JavaScript the link simply lands.

**Without JavaScript, Here is shown inactive** — dotted border, no background, text at the full
contrast floor ([decisions.oneContrastFloor](../../decisions.md#onecontrastfloor)) — and focusable, with a tooltip saying it needs
JavaScript and permission to use the reader's location. It does nothing when pressed. Dotted,
because a dashed border already marks the Only buttons, which are active.

## nowSelectsCurrentMonth

**Now selects the current month**

Subject: `presentation.month_filter.now`
Goals: [goals.localRelevance](../../../goals/goals.md#localrelevance)

A **Now** button beside **All** in Months replaces the month selection with the current month on
the reader's own clock, every other setting kept. Without JavaScript it uses the server's month.
The Months control is two lines: the label, All, Now and the explanation of what a month means,
then the twelve months.

Without JavaScript the button links to `/now`, which redirects to the index for the server's
current month and is never cached; with it, the script points the button at the reader's month
directly, and `/now` is not requested. Either way no page carries a month taken from the server's
clock, which [tech.cacheLifetimeStated](../tech/delivery.md#cachelifetimestated) depends on ([decisions.hereAndNowCaching](../../decisions.md#hereandnowcaching)).

## locationStaysInBrowser

**The reader's location never leaves the browser**

Subject: `privacy.location`
Goals: [goals.readerPrivacy](../../../goals/goals.md#readerprivacy)

The reader's position is used only inside the browser to choose a region, and is never sent,
stored or logged. The region boundaries load with every index page whether or not Here is
pressed, and pressing Here produces exactly the request that choosing that region by hand would.

So the server cannot tell that Here was used, let alone where the reader is. Loading the
boundaries only when Here is pressed would leak the first; sending the position to the server
would leak both.

## searchBox

**A search box narrows the index as you type**

Subject: `presentation.search`
Goals: [goals.flexibleComparison](../../../goals/goals.md#flexiblecomparison)

**A text box above the count line keeps the birds that match what is typed and hides the rest.** A
bird matches when what is typed, ignoring case and punctuation, appears unbroken within one of its
English names (North American and world), its species code, its scientific name, or the scientific
name of its family or order. So "d hawk" finds Sharp-shinned Hawk and the other "…d hawk"s, but
not the nighthawks, whose family is Caprimulgidae. Families and orders are matched by their
scientific names only, since their English descriptions list other birds: "Kites, Old World
Vultures, Eagles, and Hawks" would make "hawk" find every kite (Sameer, 2026-10-25, on both). The
count line, each section's count and the histogram count only the matches, a section with none is
hidden, and the count line names the search.

**It filters in the browser as the reader types**, with no request to the server. Matches keep
their places: sections and scales are the ones the other filters made, so what remains does not
move while the reader types.

**The search goes into the address as `q` on the reader's next click that changes the address**
— a control, a tile, a pick — not while typing. The welcome line's things to try, *reset to
defaults* and *Clear the filters* drop it. Opened from an address with `q`, the page shows the same
matches with or without JavaScript; without it, Enter in the box loads the page with the search.

Rationale: the first readers asked for a search bar to pick birds by name
([#3](https://github.com/google/birdview/issues/3)), and a box keeps working at ten times the
species where scrolling and chips do not. The words to match are prepared by the server, so the
rule lives in Go; the browser only looks them up, and a test holds it to the Go matcher (Sameer,
2026-10-25: filter in the browser; add `q` on the next click).
