package product

# Filtering

## familyFilter

**Filter by family**

Goals: goals.comparisonAcrossFamilies

The user can filter to any subset of families, multi-select. **Everything is selected on arrival**
— All, every order and every family show as on. An **All** control toggles like an order does,
over the whole set: untoggling it deselects everything and **no birds are displayed**.

Selection is therefore always literal: what is shown is what is selected, and the controls say so.
There is no separate "unfiltered" state that happens to look like everything being chosen.

The selector is **grouped by taxonomic order, in the checklist's own sequence** rather than
alphabetically. Each group is headed by a **toggle that selects or clears every family in that
order**, labeled with the order's scientific name and its common name, and styled like the other
controls rather than like a family chip.

Order common names are **the project's own editorial labels**. No source BirdView uses publishes
vernacular names for orders, and several are deliberately loose — Charadriiformes really is
shorebirds plus gulls plus auks. Nothing is derived from them. Each family is labeled with
**both its scientific and its common name**.

**No control carries a species count.** One count is reported instead, **above the list**, naming
how many species the current filters match out of the whole dataset.

Rationale for moving it: a per-control count is only true while it is the only filter. The moment
a second filter exists — a region ([regionAndMonthFilter](#regionandmonthfilter)) — "Parulidae 51" is false in Hawaii, and a
count that lies is worse than no count. Making every chip's count depend on every other filter is
possible and cheap, but it means 90 numbers that each need re-reading after every click, to answer
a question the reader asked once. One number above the list answers it once.

Rationale: the selector is where the common name earns its place. A birder looking for nightjars
should not have to know they are Caprimulgidae to find them, and this is the one surface where
they are choosing a family rather than reading about one they have already been shown.

Checklist sequence is how the ABA checklist, Sibley and every other field guide arrange birds, so
it is the order a birder already has in their head — and alphabetizing actively discards an
ordering the source file provides for free.

## regionAndMonthFilter

**Filter by region and month**

Subject: `presentation.region_filter`
Goals: goals.comparisonAcrossFamilies

The reader filters to any subset of US states and Canadian provinces, multi-select, as the family
filter works. A species is listed in a region when it reaches 1 record per million of that region's
bird observations, counted from eBird — or from iNaturalist only where eBird is confirmed to
withhold it there — and its records below the bar still count in its figure.

**Reinstated 2026-10-02**, having been withdrawn because no freely redistributable per-region
occurrence data existed at useful granularity ([decisions.noLocationData](../../decisions.md#nolocationdata)). It does: GBIF
serves eBird and iNaturalist under CC BY 4.0 with GADM regions as a queryable field.

The user filters to any subset of **US states and Canadian provinces**, multi-select, exactly as
the family filter works — everything selected on arrival, an **All** control, and group toggles.

**Membership is per region and relative.** A species belongs to a region when it reaches **1
record per million** of that region's bird observations, **counted from eBird** — and, only where
eBird is confirmed to withhold the species there, from iNaturalist judged against its own total
([decisions.countEbirdFirst](../../decisions.md#countebirdfirst)). It used to read "in either source", which judged iNaturalist
against its own small total everywhere, so one photo of an escaped pet cleared a bar of 0.43
records in New York. Relative rather than a
fixed count for two measured reasons: eBird's volume doubled in five years, so any absolute bar
admits more species over time for no ecological reason; and effort differs enormously between
regions — California has 114 million records against Hawaii's 2.5 million — so one count cannot
mean the same thing in both.

**Membership decides where a species is listed, not what it counts.** Records in a region where
the species stays below the bar are kept and counted in its figure ([shareOfRecords](presentation.md#shareofrecords)); they only
leave it off that region's list. Discarding them made the figure contradict the status: Emu is
established on 19 eBird records in Texas, 0.26 per million, and its tile read *0 in
1,237,352,304* ([decisions.belowBarCounted](../../decisions.md#belowbarcounted)).

**Deciding membership per region is what keeps the endemics.** A continental cut drops them by
construction: the Millerbird's handful of records is nothing against a billion, and 22 per
million within Hawaii. The regional rule needs no special case for them.

**Regions are gathered into zones so a reader can select an area in one click**, and zones mix
the two countries — British Columbia sits with Washington and Oregon, not with Nova Scotia. What
a zone *means* is [regionsGroupedTwoWays](months.md#regionsgroupedtwoways).

**A region appears in a URL as its postal abbreviation, or as its name** — `?region=CA`,
`?region=GulfOfMexico` ([decisions.regionUrlIsPostalCode](../../decisions.md#regionurlispostalcode)). Where a standard abbreviation exists it
is used, and where none does the name is used in PascalCase. **No code is invented**: the four
obvious ones are already taken, since Midway cannot be `MI` and Palmyra cannot be `PA`, and a
code chosen only to dodge a collision would sit beside 64 official ones with nothing marking it
as this project's guess. A region with no standard code therefore shows **no code** on its chip,
only its name.

The GADM id stays the internal key — it keys the presence table and is what GBIF is queried
with — but it is a database key and does not belong in a URL. The slug is **derived** from the
code or the name rather than stored alongside them, so a renamed region cannot keep a URL that
no longer describes it.
