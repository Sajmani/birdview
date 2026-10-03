# Product requirements

What BirdView must do, from a user's point of view. Normative: where this file and the code
disagree, the code is wrong — or this file needs an acknowledged amendment.

IDs are permanent. Withdrawn requirements stay in place so references to them still resolve.

---

## Dataset and taxonomy

### P-001 — Scope is the regularly occurring ABA Area avifauna
Subject: `dataset.scope`
Value: `ABA Area, rarity codes 1-3`
Goals: G-001, G-002

The dataset is the species of the ABA Area — North America north of Mexico, plus Hawaii —
restricted to ABA rarity codes 1 through 3, which is 841 of the 1,161 species in ABA Checklist
8.19.

Rationale: codes 4–6 are casual, accidental, extinct or extirpated — 320 species a birder will
essentially never encounter, many of them Old World strays whose presence among the common birds
makes the comparison less useful, not more. The rarity column is read at ingest purely as a scope
filter; it is not carried into the dataset and not shown in the UI (see P-002 for why that
distinction matters to the license).

### P-002 — Morphometric traits come from AVONET
Subject: `dataset.traits.source`
Value: `AVONET2_eBird`
Goals: G-001, G-003

Bill, wing, tail and mass measurements come from AVONET's `AVONET2_eBird` sheet, which uses
eBird/Clements taxonomy. See [sources.md](sources.md#avonet).

### P-003 — Traits carried per species
Goals: G-003

Each species record carries scientific name, family, and these AVONET traits: `Beak.Length_Culmen`,
`Tail.Length`, `Wing.Length`, `Secondary1`, `Mass`.

Rationale: this is the set the diagram and its derivations actually consume. AVONET's other
columns — `Beak.Width`, `Beak.Depth`, `Hand-Wing.Index` — are deliberately not ingested; carrying
unused data invites someone to invent a use for it.

The set moves both ways, and a column that stops earning its place leaves. `Beak.Depth` was
ingested until the cranium model left bill depth for mass. `Tarsus.Length` made the whole round
trip — excluded as unused, ingested when the diagram gained legs, then dropped again when the
flight premise removed them ([CR-010](decisions.md#cr-010)).

### P-004 — Names come from AviList
Subject: `dataset.names.source`
Value: `AviList v2025b`
Goals: G-002

English common names in **both locales** ([P-049](#p-049)), the **eBird species code**, family
scientific and English names, the order, and the family **sequence** all come from
[AviList](sources.md#avilist). Scientific names come from the scope list, and are the join key.

Rationale: **licensing, and nothing else.** The ABA checklist supplied all of this and supplied it
well, but it is all rights reserved — serving BirdView reproduces whatever names it, so the
checklist blocked public deployment outright ([CR-022](decisions.md#cr-022)). AviList is CC BY 4.0
and carries the same four things, so it can be served.

That it is *also* the consensus successor to IOC, Clements and BirdLife is a bonus, not the
reason.

**AviList recognizes 7 of BirdView's species under a different name and 2 not at all.** The seven
are one-to-one renamings and resolve through a synonym table. The two — Gray-headed Swamphen and
Western Warbling-Vireo, both ABA splits AviList lumps — are **dropped from the dataset**. That
follows from the decision rather than qualifying it: if AviList supplies the names it supplies the
species concepts, and keeping a bird AviList does not recognize would mean keeping the one name
the project may not redistribute.

Both tables are recorded in the code with this requirement cited beside them, and the ingest
**asserts the drop list** rather than discovering it: a species losing its name source quietly
would change its URL under a reader ([P-049](#p-049)).

### P-005 — Family badge colors are generated
Subject: `badge.color.derivation`
Value: `generated`
Goals: G-002

Colors are generated as **two palettes**: one per taxonomic order, and one per family *within* its
order. Both satisfy T-010, and families are visibly different from those listed next to them.

Two families in different orders may share a color. Their order badge distinguishes them.

Rationale: the ABA area spans 91 families. A hand-curated palette does not scale to that and
breaks on every taxonomic revision, so generation makes the palette a function of the taxonomy
rather than a table to maintain.

The two-palette split is what makes the problem solvable at all. A legible badge bounds the usable
gamut at roughly 43 mutually distinguishable colors with black text, or 69 if white text is also
allowed — against 91 families. No assignment fits 91 into 43. Scoping family colors to their order
reduces the requirement to the largest order, Passeriformes at 44 families, which does fit.

Distinctness must be measured perceptually, and between the entries a reader actually compares.
Evenly spaced hues satisfy "distinct colors" on paper and failed completely in practice: at 91
families that is a four-degree step, and the selector listed families in that same order, so every
neighbor was the same shade.

**Known residual:** within Passeriformes, 9 of 946 family pairs remain below ΔE 15. The dark half
of the gamut has less perceptual room than the light half, and 44 families is close to what the
space holds. Those pairs are never adjacent in the selector, and every badge carries its family
name.

### P-006 — Sources join on normalized scientific binomial
Subject: `dataset.join.key`
Value: `scientific binomial, normalized`
Goals: G-001

Species are joined across sources on scientific binomial, after stripping `(...)` parentheticals
and collapsing internal whitespace, with a committed mapping table covering the taxonomic
divergence between AOS/NACC (which ABA follows) and the 2022-era eBird/Clements taxonomy AVONET
carries.

Rationale: the measured join rate on binomial alone, over the P-001 scope, is **97.38%** — 819 of
841 — leaving 22 species needing a mapping entry. English name is the worse key: hyphenation and
Eastern/Western prefixes churn on every taxonomy update.

The 22 are not all renames, and that distinction is the subject of [CR-006](decisions.md#cr-006).
A rename may be mapped freely. A **split** cannot be mapped without borrowing the parent species'
measurements, which makes the value no longer a measurement *of that species* and therefore
engages P-060 and P-061.

### P-007 — Unmatched species are reported, not silently dropped
Goals: G-001

Ingest reports every species that fails to join, by name and count, and fails the build if the
unmatched fraction exceeds a stated threshold.

Rationale: ABA, AOS and eBird/Clements revise on independent annual cycles. Without a loud
failure, a taxonomy update silently shrinks the dataset and nobody notices.

---

## Filtering

### P-010 — Filter by family
Goals: G-002

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
a second filter exists — a region ([P-011](#p-011)) — "Parulidae 51" is false in Hawaii, and a
count that lies is worse than no count. Making every chip's count depend on every other filter is
possible and cheap, but it means 90 numbers that each need re-reading after every click, to answer
a question the reader asked once. One number above the list answers it once.

Rationale: the selector is where the common name earns its place. A birder looking for nightjars
should not have to know they are Caprimulgidae to find them, and this is the one surface where
they are choosing a family rather than reading about one they have already been shown.

Checklist sequence is how the ABA checklist, Sibley and every other field guide arrange birds, so
it is the order a birder already has in their head — and alphabetizing actively discards an
ordering the source file provides for free.

### P-011 — Filter by region
Subject: `presentation.region_filter`
Goals: G-002

**Reinstated 2026-10-02**, having been withdrawn because no freely redistributable per-region
occurrence data existed at useful granularity ([CR-003](decisions.md#cr-003)). It does: GBIF
serves eBird and iNaturalist under CC BY 4.0 with GADM regions as a queryable field.

The user filters to any subset of **US states and Canadian provinces**, multi-select, exactly as
the family filter works — everything selected on arrival, an **All** control, and group toggles.

**Membership is per region and relative.** A species belongs to a region when it reaches **1
record per million** of that region's bird records, in either source. Relative rather than a
fixed count for two measured reasons: eBird's volume doubled in five years, so any absolute bar
admits more species over time for no ecological reason; and effort differs enormously between
regions — California has 114 million records against Hawaii's 2.5 million — so one count cannot
mean the same thing in both.

**Deciding membership per region is what keeps the endemics.** A continental cut drops them by
construction: the Millerbird's handful of records is nothing against a billion, and 22 per
million within Hawaii. The regional rule needs no special case for them.

**Regions are gathered into zones so a reader can select an area in one click**, and zones mix
the two countries — British Columbia sits with Washington and Oregon, not with Nova Scotia. What
a zone *means* is [P-013](#p-013).

**A region appears in a URL as its postal abbreviation, or as its name** — `?region=CA`,
`?region=GulfOfMexico` ([CR-024](decisions.md#cr-024)). Where a standard abbreviation exists it
is used, and where none does the name is used in PascalCase. **No code is invented**: the four
obvious ones are already taken, since Midway cannot be `MI` and Palmyra cannot be `PA`, and a
code chosen only to dodge a collision would sit beside 64 official ones with nothing marking it
as this project's guess. A region with no standard code therefore shows **no code** on its chip,
only its name.

The GADM id stays the internal key — it keys the presence table and is what GBIF is queried
with — but it is a database key and does not belong in a URL. The slug is **derived** from the
code or the name rather than stored alongside them, so a renamed region cannot keep a URL that
no longer describes it.

**Six species belong to no region, and that is structural.** Bermuda Petrel, Desertas Petrel,
Trindade Petrel, Townsend's Storm-Petrel, Craveri's Murrelet and Guadalupe Murrelet are seen at
sea, where no land polygon reaches — Bermuda Petrel has 384 records in the ABA Area and **zero**
in any GADM region. They are not errors and must not be filtered away silently; a pelagic bird is
still a bird a reader may want.

### P-012 — Filters compose
Subject: `presentation.filters`
Goals: G-002

**Reinstated with [P-011](#p-011).** Region and family narrow the list together: what is shown is
what every active filter admits. Selecting Hawaii and Parulidae shows the warblers of Hawaii, and
shows nothing if there are none.

An empty intersection is a real outcome and is **not** the same as nothing being selected
([P-037](#p-037)). The empty state names each filter and what it is set to, because with two
filters a reader cannot otherwise tell which one emptied the page.

### P-013 — Regions are grouped two ways
Subject: `presentation.region_filter.zones`
Goals: G-002

The selector offers **two groupings** of the same regions, and the reader picks from either
([CR-023](decisions.md#cr-023)). Each is a **partition**: every region sits in exactly one zone
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

**Neither grouping may be presented as the other.** The UI labels which is which, because a
hand-drawn grouping shown as measurement is the failure [CR-023](decisions.md#cr-023) was
written to correct: the previous zones were defended with a statistic that turned out to be an
artifact of a broken query.

Zone membership is **presentation only**. It changes which chips a group toggle selects and
nothing about which birds a region contains — that is [P-011](#p-011)'s 1-ppm rule, and it is
computed per region, never per zone.

---

## Ordering and sectioning

### P-020 — One primary metric, user-selected
Subject: `ordering.metric`
Value: `length | mass | wingspan`
Goals: G-001, G-002

The user chooses whether the dataset is ordered by total length, body mass, or wingspan.

### P-021 — Ascending order
Goals: G-001

Species are sorted ascending by the selected metric.

### P-022 — Sections break at N% above the section minimum
Subject: `sectioning.rule`
Goals: G-001, G-002

A new section begins at the first species more than N% larger than the **minimum** species of the
previous section.

Rationale: bird size follows a power law. Fixed-width bins put hundreds of small birds in one row
and isolate large raptors in rows of one. A ratio rule produces sections of comparable visual
span at every scale — and, because a section is by construction no more than N% wide, makes the
per-section scale of P-042 honest.

### P-023 — N defaults to 15 and is adjustable
Subject: `sectioning.step_percent`
Value: `15`
Goals: G-002

Offered values run from 5% to 30%. Every value the control offers must be one the server accepts,
or the control silently falls back to the default.

### P-024 — Section headers state range and count
Goals: G-001

Each section shows its range in the currently selected metric and the number of species in it.

Rationale: under per-section scaling (P-042) the header is what tells the user that the scale has
changed between one section and the next. Without it, per-section scaling is misleading.

---

## Presentation

### P-030 — Diagram is the only presentation mode in v1
Subject: `presentation.mode`
Value: `diagram`
Goals: G-001, G-003

Rationale: photographs were scoped out of v1 (P-050). The diagram is what makes size and shape
comparable; photographs were always the enrichment.

### P-031 — Tile contents
Subject: `presentation.tile`
Goals: G-001, G-002, G-003

Each species is one **tile**, laid out top to bottom:

1. a header whose first line carries the **common name followed by its eBird species code in
   parentheses**, with the **scientific name right-justified on the same line**; and below it the
   **family badge**, showing the family's scientific name and its common name together;
2. the diagram, **vertically centered** in its area, carrying the **mass cube** in its lower-left
   corner ([CR-020](decisions.md#cr-020)) and **no labels of any kind**;
3. the main metrics — **length, wingspan, mass** — named in full and in the selected unit system.

The metrics are **spelled out**, not abbreviated to `L`, `W` and `M`. The initials cost a reader
who has not yet learned them, which is every reader on their first tile, and they save a line of
nothing: the three fit a tile at full width.

The **cube** is the one mark the tile shares with the annotated diagram. It needs no gutter and
no leader, only a corner the drawing does not use, so it does not cost the density that keeps
labels off the tile. It carries no label here — the footer already names the mass — and because
every tile in a section is drawn to one scale (P-042), the cubes are directly comparable down the
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

Rationale: this is a layout rule in service of G-001. A section is a run of birds drawn to one
scale, and reading down it — comparing diagrams, and comparing mass cubes against each other — is
what the index is for. A single tile standing taller than its neighbours puts every diagram in
that row at a different offset, which costs the comparison far more than an abbreviated word
does. Eight of the 841 common names are long enough to wrap, and the longest family badge comes
within a few pixels of the tile's inner width; one tile is enough to disturb a whole row.

In pick mode ([P-068](#p-068)) each tile also carries a **pick control**, beside the tile's own
link rather than inside it. In browsing mode it is absent entirely, so the tile a reader sees
almost all the time is unchanged.

Rationale for centering: under global scale a small bird occupies a fraction of the tile's height,
and bottom-aligning it reads as a rendering fault rather than as the size difference it is.
Centering makes the empty space symmetric, so it reads as headroom rather than as misalignment.

Family common names are **not** unique across families — several groups share a vernacular
label — so they are a reading aid beside the scientific name, never an identifier.

### P-032 — Imperial by default, metric available
Subject: `units.default`
Value: `imperial`
Goals: G-001

One toggle switches the whole interface between:

| | length, wingspan | mass |
| --- | --- | --- |
| imperial (default) | inches | ounces, pounds above 16 oz |
| metric | centimeters | grams, kilograms above 1 kg |

**Every** displayed number honors it — tile footers, section headers and the detail view alike.

Rationale: the ABA Area audience is predominantly US; the toggle serves Canadian users and anyone
who thinks in centimeters. Ounces below one are shown to two decimals, because a 3 g hummingbird
and a 12 g kinglet both round to 0.1 oz otherwise, erasing the distinction at exactly the size
where the app is trying to make it.

### P-033 — First load shows everything
Goals: G-002

The initial view has every family selected and shows all species, ordered by length.

Rationale: the full sweep from hummingbird to swan is the clearest statement of what the app is
for. Opening filtered or empty would hide it.

### P-038 — Toggling a control keeps your place
Subject: `presentation.scroll`
Goals: G-002

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
[arch.md](arch.md)). Without JavaScript the page behaves as it did before.

### P-037 — The empty state says what to do
Subject: `presentation.empty_state`
Goals: G-002

With nothing selected, the display area reads *"Select birds to display in Controls above"*.

**When something is selected and nothing matches, the message names the filters that are set.**
Not "no species match that filter" — that says a thing failed without saying which thing, and
leaves the reader to go and look. It reads as *"No birds match. Family: Parulidae, Icteridae."*,
listing each active filter and what it is set to, and offers the link that clears them.

The two messages stay distinct, and the distinction is the point: **nothing selected is not the
same as nothing matched.** In the first case nothing failed — the reader turned everything off,
and the fix is to turn something on. In the second the filters are each reasonable and their
intersection is empty, and the fix is to relax one. A single message cannot say both.

Rationale: an empty page is reachable in one click, and with two filters an empty *intersection*
is reachable almost as easily — Hawaii and Parulidae is a plausible thing to ask for and a dead
end to arrive at. The description is built once and used in both places it is needed: here, and
in the collapsed summary of a filter control, so the two cannot drift apart.

### P-036 — The controls collapse
Subject: `presentation.controls`
Goals: G-002

The controls are **several collapsible panels**, each starting open and each remembering its own
state across navigations (P-038): **Controls** for the display toggles, **Regions**, and
**Families**.

They are separate boxes rather than one, because the two filters are large — 90 family chips in
26 order groups, and 74 regions in 10 zones — and stacking both inside one panel pushes the birds
a long way below the fold while making "collapse the filter I am done with" impossible. Each
panel carries the documentation for its own controls (P-035), so an explanation sits with the
thing it explains.

Rationale: a reader who has chosen their filter should be able to put it away — and have it stay
away. With two filters that is not one decision but two, and one panel cannot express it.

### P-035 — The page explains itself
Subject: `presentation.help`
Goals: G-001, G-002, G-003

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
[AC-085](acceptance.md#ac-085) checks by comparing the two sets rather than by looking for names
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

### P-034 — A tile opens a detail view
Subject: `presentation.detail_view`
Goals: G-001, G-003

Clicking a tile opens a detail view that is **a larger version of the same tile** — same header,
same diagram, same main metrics, same layout — with the bird drawn **as large as the view allows,
whatever scale the index is using**, plus:

- the measurement of **every individual diagram segment**: bill, head, neck, body, tail;
- the **wing chord**, which the crossbar draws as its solid outer portions (P-065) and which would
  otherwise be the one length visible in the drawing with no row explaining it;
- the **provenance of every metric shown**, main metrics and segments alike (P-063).

Rationale: the tile is deliberately dense, because it exists to be compared against dozens of
others at a glance. That leaves no room to say where each number came from. Keeping the detail
view the *same* layout at a larger size means the user is reading a familiar thing with more of it
exposed, rather than a different screen they have to re-orient to.

The detail view does **not** inherit the index's scale. Under the global default a kinglet is a
few pixels tall, which is the honest comparison — but the one view devoted to that single bird has
nothing to compare it against, so carrying the scale there would leave nothing to look at.

### P-067 — A comparison view holds several birds to one scale
Subject: `presentation.compare_view`
Value: `/compare?species=<id>&species=<id>…`
Goals: G-002, G-001, G-003

A **comparison view** draws any number of chosen species together **on a single shared scale**, so
they can be read directly against one another. It answers G-002's own question — *is a Merlin
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
wingspan, exactly as P-042 does for a section. There is **no per-panel scale option**. Drawing each
bird to fill its own panel is what P-042 rejects outright, and in a view whose entire purpose is
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
is where the source or model behind each number is named (P-063).

This is a deliberate scoping of where provenance is *explained*, not a weakening of it. The compare
view still marks every value that is not measured — the stroke styles of P-065, the hollow head,
and the `~` on each label — which is exactly what P-061 requires of a surface showing numbers and a
diagram: mark the kind, *without having to explain why*. What it does not reproduce is the `model`
and `source` text of P-063. The alternative was to repeat a full provenance table for every bird on
the page, which would bury the comparison the view exists to make, under the same reasoning that
keeps the tile dense.

### P-068 — Birds are picked for comparison from the index, and the picks are the mode
Subject: `presentation.pick_mode`
Value: `?species=<id>&species=<id>… — present, even empty, means picking`
Goals: G-002

A reader assembles a comparison by **picking birds from the index**, where the diagrams and the
family filter already are, rather than from a separate chooser that would duplicate both.

**The presence of a `species` parameter is the mode.** There is no separate flag. With none, the
app behaves exactly as it always has. With one or more, every page is in **pick mode**: the index
offers a pick control on each tile, the index and the detail view carry a bar showing the picks,
and the detail view offers to pick the bird being read.

A **bare `species=`** — the parameter present with no value — means *picking, nothing picked yet*,
so unpicking the last bird does not silently drop the reader out of the mode mid-task. This is
the distinction P-010's family filter already draws between a parameter that is absent and one
that is explicitly empty, and it is drawn here the same way for the same reason.

That empty state is also an **entry point**. A **Compare birds** toggle in the controls turns
pick mode on with nothing picked, and off again — the same thing the bar's Clear does, by the
same link. It is for a reader who wants to compare before they have a particular bird in mind;
the detail view's control is for one who already does.

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
beside it, never a repurposing of the tile, so [P-034](#p-034) holds without exception. A reader
mid-pick can always open a bird to look at it properly, which is the main thing they will want to
do while deciding.

**Picking preserves everything else.** The filter, ordering, scale and units all keep working, so
a reader can narrow to one family and pick three birds from a page showing only that family. The
bar lists picks the filter is currently hiding, because a filter that appeared to discard them
would be worse than useless.

**Vocabulary: a bird is *picked*, a family is *selected*.** Both sets are live at once in this
mode and they mean unrelated things — a reader may have three families selected and two birds
picked — so they do not share a verb.

#### Unrecognised ids are dropped and named

An id that matches no species is **dropped, and the page says which ids it dropped**, on the index
and in the comparison view alike. Every link the page then builds omits it, so the set repairs
itself on the next click.

This is not typo-handling, it is rot-handling. Species ids are derived from scientific names and
this checklist renames them — the Sandhill Crane is `antigone_canadensis` and was
`grus_canadensis`. A comparison URL is made to be shared and kept, so it will outlive some of the
ids in it. See [CR-018](decisions.md#cr-018), which reverses an earlier 404.

#### What a tile link carries, and the one thing it does not

A tile link carries the picks and the display options — `species`, `units`, `by`, `scale`, `step`
— because the detail view cannot otherwise offer to pick the bird, build a link to the set plus
that bird, or format a single value in the reader's own units.

It deliberately **does not carry the family filter.** `family` is the only parameter with no bound
on its length: a 90-family selection is about 1.7 KB, which across 841 tiles would add roughly
1.4 MB to one page. The consequence, accepted rather than overlooked, is that the in-page back
link restores the picks, the ordering, the scale and the units, but **not the filter**. The
browser's own Back button restores that exactly, because it replays the previous URL.

---

## Diagram

### P-040 — Segment composition
Subject: `diagram.segments`
Value: `bill, head, neck, body, tail, wing crossbar`
Goals: G-003

**The diagram is a top-down view of a bird in flight.** Everything below follows from that, and it
is stated first because leaving it unstated is what made the leg question unanswerable for as long
as it was; see [CR-010](decisions.md#cr-010).

The diagram is a vertical cross. Top to bottom: a bill stick, a head oval, a neck, a body stick,
and a tail stick — plus one horizontal crossbar representing the wings. The tail is visually
distinguished from the body by stroke style, colour and a mark at the junction between them. That
remains true under P-065, but for a different reason: the tail is measured and so solid, the body
modeled and so dashed. The junction mark carries the distinction on its own if the two ever share a
stroke style again.

**The neck is a stick, except where the family folds it in flight.** Ardeidae and Pelecanidae fly
with the head drawn back onto the shoulders, and for those the neck is drawn as an **S of two
semicircular arcs**. The stroke is still exactly the neck long at the diagram's scale — the S is
the same neck, curved — so it occupies `2/π ≈ 63.7%` of that length vertically. That ratio is a
consequence of the shape and not a chosen parameter; see [CR-011](decisions.md#cr-011).

**There are no legs.** In flight most birds tuck them out of sight, and for the minority that
trail them — herons, cranes, storks, ibises, flamingos, stilts, rails, loons, grebes — what shows
is the leg's projection past the tail tip, a small residual of quantities BirdView does not have:
the femur, the toes and the hip's position are unmeasured by any source, and the feathered
tibiotarsus that straightens out in flight is exactly what a standing-posture leg model leaves out.
Legs were drawn from the measured tarsus, then from a modeled leg, and are now removed; CR-010
records the evidence and what the removal costs.

Rationale: the tail is a stick rather than a triangle because no trait in any source gives a tail
width, and a triangle would assert a fan shape the data does not support.

### P-041 — Segments abut exactly and sum to total length
Subject: `diagram.segments.sum`
Goals: G-001, G-003

The five vertical segments meet without gap or overlap, and their lengths sum to the species'
total length.

Rationale: this is what lets the vertical extent of the drawing *be* the bird's length, so that
two diagrams at the same scale can be compared directly.

**Qualification.** That equality is a property of the segment *lengths*, and it holds without
exception. It is also true of the drawn vertical extent for 827 of 841 species — but not for the
folded-neck families of P-040, whose S occupies less height than the neck is long. Their silhouette
is genuinely shorter than their catalogued length, which is a fact about a flying heron and not a
rendering error. Everything that treats length as data — ordering, sectioning, the scale reference
of P-042, and the detail view — uses the full total length regardless of how the neck is drawn, so
comparability is unaffected. See [CR-011](decisions.md#cr-011).

### P-042 — Scale is a user toggle, defaulting to per-section
Subject: `diagram.scale.mode`
Value: `per-section (default) | global`
Goals: G-001, G-002

The user switches between a single scale across the entire filtered set, and a scale shared by all
species within a section. **Per-section is the default, and is listed first.**

Whichever is chosen, the scale is set by the largest total length in the group. Since the segments
sum to total length (P-041) and nothing is drawn outside them, that is the whole vertical extent.

Rationale: this is the requirement G-001 lives or dies by. A per-tile scale — every bird drawn to
fill its tile — makes a hummingbird and a swan identical on screen and destroys size intuition
entirely.

**The default was global and is now per-section** ([CR-017](decisions.md)). Global still makes the
cross-family argument better than anything else, and it is one click away and named first in the
help. But the diagram has since acquired detail a reader has to be able to see — provenance in the
stroke style (P-065), the chord ticks, the labels of P-066 — and at global scale a median bird is a
few pixels tall, where none of it is legible. A default that hides the drawing's own vocabulary
teaches the reader nothing, and a reader who has learned to read one bird is far better placed to
be surprised by the swan.

### P-043 — The wing crossbar is drawn at the vertical scale
Subject: `diagram.wingspan.scale`
Value: `same as vertical axis`
Goals: G-001

The crossbar's length represents wingspan at the same scale as the vertical axis, so its length
relative to the body is literal.

Rationale: wingspan runs 1.0–2.5× body length, so honoring this forces a wide tile — roughly
2.6:1. That is a real layout cost, accepted because a crossbar on its own arbitrary scale tells
the user something false about proportion while looking authoritative.

### P-044 — Proportion differentiates species
Goals: G-003

A dowitcher reads as long-billed, an accipiter as long-tailed, a heron as long-necked.

### P-046 — The cranium is modeled from mass
Subject: `diagram.head_model`
Value: `isometric on mass`
Goals: G-003

Head length scales isometrically with body mass. Like every modeled value it names its function
(P-062) and claims no measurement.

Rationale: an earlier model derived the cranium from bill depth with a 12 mm floor, and the floor
bound for **442 of 841 species and 86% of birds under 200 mm**. Half the dataset drew one identical
head that varied with nothing, consumed a fixed share of every small bird's length, and
differentiated no species from another — the opposite of what a shape diagram is for. Bill depth is
in any case a poor predictor of skull length; unfloored it gave a sandpiper a 7 mm head.

Mass is the only whole-animal measurement available, and a linear dimension scaling as its cube
root is the natural default. There is deliberately **no floor**: a floor would only conceal a bad
coefficient.

Known limitation: because both cranium and trunk scale on mass, birds of similar mass get similar
heads. Within a size tier the head therefore differentiates little — the shape signal there comes
from bill, tail, neck and wingspan, which are measured or family-specific.

**The cube root is deliberately left alone by [CR-021](decisions.md#cr-021)**, which raised the
*trunk's* exponent on the evidence of measured skeletons. The same evidence supports isometry for
a cranium: single skeletal elements scale at 0.314–0.325, where chains of them scale higher. The
head is one element and keeps 1/3; the trunk is a chain and does not. See [P-047](#p-047).

Still open, and recorded because it is the one place the head model is visibly wrong: an Ostrich
draws a 360 mm cranium against a real skull of roughly 200 mm, which implies bird heads scale with
*negative* allometry rather than isometry. No head dataset exists at this coverage to fit against
— the one large skull collection is passerines-only — so the coefficient stands unvalidated. It
costs nothing inside the ABA scope, where the heaviest bird is 11 kg rather than 111 kg.

### P-047 — The trunk is modeled allometrically from mass
Subject: `diagram.trunk_model`
Value: `14.0329 x mass^0.37`
Goals: G-001, G-003

Trunk length scales with body mass by a power law. Like every modeled value it names its function
(P-062) and claims no measurement.

The **exponent is calibrated against measured skeletons** rather than chosen. It must lie inside
the 95% confidence interval that published skeletal data give for a *multi-element* linear
dimension ([`bohmer-neck`](sources.md#bohmer-neck)) — **0.338 to 0.401**, the intersection of the
intervals for neck length and total leg length. The distinction matters: single bones scale near
isometry, chains of them scale higher, and a trunk is a chain.

The **coefficient is a consequence, not a second free parameter.** It is set so the model is
unchanged at **63.4 g**, the median mass of the ABA dataset when the exponent was calibrated.
Recalibration is therefore a change of *proportion* across the size range rather than of overall
scale: birds near the middle are drawn as they were, small birds slightly shorter, large birds
substantially longer.

That anchor is a **fixed constant, not a value read from the dataset at build time.** If scope
changes the median moves, and the drawing must not move with it silently; re-anchoring would be a
decision, recorded as one.

Rationale: three of the five segments are driven by mass, so this exponent governs how much longer
a big bird is drawn than a small one — which is the one thing G-001 exists to convey. The original
0.31 was an unvalidated constant and it was **too shallow**, understating the large end of the
current scope by about 1.65×. See [CR-021](decisions.md#cr-021) for the measurement, and for what
it does not establish.

### P-045 — The neck/body split is a named per-family model
Subject: `diagram.neck_ratio`
Value: `per-family modeling function`
Goals: G-001, G-003, G-004

The division of the remaining body axis between neck and body uses a per-family ratio. The
resulting **neck length is presented as a metric like any other**, in the detail view, carrying a
provenance record whose `kind` is `modeled` and whose `model` **names the specific function and
the ratio applied to that bird's family** — for example, "per-family neck ratio, Ardeidae = 0.35".

The ratio table is recorded in the code with this requirement cited beside it.

Rationale: no source measures avian neck length at this coverage, and one ratio for every species
would make every bird the same shape, defeating G-003. What makes a per-family constant
acceptable is not hiding it but naming it: a user who wants to know why a heron is drawn
long-necked can read the exact function that made it so. Hiding the number while still drawing it
would be the less honest option — the reader sees the neck either way, and only the explanation
would be missing. Supersedes the condition in [CR-004](decisions.md#cr-004); see
[CR-007](decisions.md#cr-007).

### P-048 — One diagram vocabulary, on every surface that draws one
Subject: `diagram.vocabulary.shared`
Goals: G-001, G-002

Four surfaces draw the diagram: the **favicon**, the **index** (its tiles and its legend), the
**detail view**, and the **comparison view**. A reader meets the same figure on all four, so the
figure means the same thing on all four.

**Colour is the shared part, and it is shared exactly.** Each part of the diagram — bill, head,
neck, body, tail, wing crossbar, mass cube — has one colour, and every surface uses that colour
for that part. A surface may not pick its own shade of a part, however good its local reason.

**Size-dependent geometry is not shared, and is not expected to be.** Stroke weight, dash pattern
and which parts appear at all may be tuned per surface, because a 16px icon and a 1000px detail
diagram are not the same drawing problem: the favicon drops the dashes entirely and collapses
neck, body and tail into one stroke, and the detail view retunes the body's dash so it stays
legible at its own scale. Any such divergence carries its reason beside it.

**The mechanism is most of the guarantee.** The index, detail and comparison views cannot diverge:
one `diagram` template emits the markup and one stylesheet colours it, so there is nowhere for a
per-page shade to live. The favicon is the exception, and the only one — a standalone asset that
cannot reference the stylesheet and therefore restates the palette by hand. **It is the restatement
that has to be checked**, which is what [AC-081](acceptance.md#ac-081) does.

Rationale: this requirement exists because the thing it forbids had already happened. The favicon
drew its crossbar `#5ccbfa` while the app drew `#0ea5e9` — a hand-lightened variant adopted
because the app's own blue was too dark to clear the icon's contrast standard. That is the
failure in miniature: a surface hit a problem with the shared vocabulary and fixed it locally
instead of fixing the vocabulary. The app's blue was *also* too dark for its own labels (T-023),
so the local fix concealed a real defect for as long as it stood.

The lesson generalizes past colour. When one surface cannot live with a shared decision, that is
evidence about the decision, not licence to fork it.

### P-049 — The reader chooses which English name, and the URL does not change
Subject: `presentation.names.locale`
Value: `na (default) | world`
Goals: G-001, G-002

English bird names are regional. The same bird is **Dovekie** to a North American birder and
**Little Auk** to everyone else, and neither is a mistake. BirdView serves both, as a toggle
beside Units and Scale:

| locale | source | *Alle alle* |
| --- | --- | --- |
| **`na`**, the default | eBird/Clements | Dovekie |
| `world` | AviList, following IOC | Little Auk |

**North American is the default**, this being an ABA-area dataset whose reader calls that bird a
Dovekie. Over the current scope the two differ for **82 species**, most of them hyphenation; the
genuine vocabulary changes are a much smaller set.

Both names are compiled into the dataset, because the locale is a request-time choice and
re-ingesting per locale would be absurd. Five species have no eBird/Clements concept and so no
North American name; they **fall back to the world name** rather than rendering blank.

**Family English names are not localized.** AviList publishes one per family, so "Fringilline and
Cardueline Finches and Allies" becomes "Finches, Euphonias, and Allies" in either locale.
Accepted: a family label is a reading aid, not an identifier.

#### The species id is the eBird code, and is locale-independent

A species' URL is `/species/amerob`, and the comparison view names its birds the same way:
`/compare?species=amerob&species=doveki`. The id is the **eBird species code** from
[P-004](#p-004).

**This is what makes the locale toggle possible at all.** The four-letter alpha code it replaced
was *derived from the English name*, so under the world locale `DOVE` would have to become
`LIAU` — and an identifier that moves when a display preference moves is not an identifier. The
eBird code is assigned rather than derived, so `doveki` is `doveki` under both names.

It is also simply better as a URL: `/species/amerob` over `/species/turdus_migratorius` is
shorter, typable, and survives a genus move — and genus moves are common, with seven in this
dataset alone.

Three species have no eBird code. They keep a **scientific-name slug** as their id, which is the
old scheme surviving where the new one has nothing to offer.

**Ids must be unique and stable across rebuilds.** Uniqueness is not automatic: AviList lumps two
BirdView species into taxa already in scope, and lumping them in place rather than dropping them
would have given two birds one URL. That is one of the two reasons P-004 drops them.

### P-050 — Species photographs
Status: **Withdrawn from v1.** See [CR-005](decisions.md#cr-005) for the research and the reasons
it is deferred rather than abandoned.

### P-051 — Perched and in-flight selection
Status: **Withdrawn from v1** with P-050. Note for whoever revives it: Wikimedia Commons has
per-species `in flight` categories but no "perched" equivalent, so the design must treat in-flight
as a positive filter and perched as the residual default.

### P-052 — Photograph attribution
Status: **Withdrawn from v1** with P-050.

### P-053 — Local caching of photographs
Status: **Withdrawn from v1** with P-050.

---

## Measurements

### P-060 — Where length and wingspan come from
Subject: `measurement.source`
Goals: G-001

- **Wingspan, measured** — Fox et al. 2026, filtered to `needs_prediction = 0` so only empirical
  values are taken. Stored as the reported min/mid/max triple rather than a point value.
- **Wingspan, modeled** — where no empirical value exists, a log-log regression of wingspan on
  wing chord, fitted **by the ingest** on those same empirical rows, per taxonomic order, with a
  global fit for orders holding fewer than five empirical species. See
  [CR-008](decisions.md#cr-008).
- **Length** — **modeled for every species.** No usable source of measured avian total length
  was found: Ricklefs 2017 was the only candidate and its terms prohibit the use
  ([CR-002](decisions.md#cr-002)).
- **Traits** — AVONET, read from whichever of its three taxonomy sheets carries the species as its
  own taxon ([CR-006](decisions.md#cr-006)).
- **Everything not covered above** — modeled.

**Total length is a composite.** It is the sum of the five segments (P-041), of which bill and
tail are measured from AVONET while head, neck and body are modeled. Its `kind` is therefore
`modeled`, and its `model` says so — a reader who sees a measured bill inside a modeled total
should be able to find out how both are true at once.

Rationale: AVONET records neither macroscopic dimension, and no open source supplies both across
the ABA avifauna. Mixing sources of differing authority is acceptable *only* under P-062. See
[CR-001](decisions.md#cr-001).

### P-061 — The tile marks any metric that is not measured
Goals: G-001

On the tile itself — where numbers and the scaled diagram appear — any metric whose `kind` is not
`measured` is visibly marked as such, without the tile having to explain why.

Rationale: this is what makes P-060 honest rather than merely convenient. If it is dropped, the
provenance record becomes metadata nobody sees and BirdView presents modeled numbers as data,
which is the thing the project explicitly rejected. It is load-bearing, not decoration.

### P-062 — Every presented metric carries a provenance record
Subject: `measurement.provenance.record`
Value: `{kind, source, sourceTaxon, model}`
Goals: G-001, G-004

Every metric BirdView presents — total length, wingspan, mass, and each diagram segment — carries
a record of where that specific value came from:

| Field | Meaning |
| --- | --- |
| `kind` | `measured`, `near-match`, or `modeled` |
| `source` | The dataset and its pinned version, including the sheet where that distinguishes rows |
| `sourceTaxon` | The binomial the value was actually filed under |
| `model` | For `modeled` values, the named model that produced it |

The three kinds:

- **`measured`** — a published measurement *of this species*, whatever name it is filed under and
  whichever sheet it sits in. A value read from `AVONET1_BirdLife`, or found under a pre-revision
  name such as *Charadrius nivosus* for *Anarhynchus nivosus*, is measured: the specimens are the
  right birds. `sourceTaxon` records the name so the claim is auditable.
- **`near-match`** — a published measurement of a **different taxon**, used because none exists
  for this species. Typically a pre-split parent: *Larus argentatus* standing in for *Larus
  vegae*. Real data, wrong bird, and the record says so.
- **`modeled`** — computed rather than measured. `model` names the computation, and where the
  model's accuracy has been measured it states that too, so a reader can judge how far to trust
  the number rather than only knowing that it is an estimate.

Rationale: this generalizes what CR-001 established for lengths and CR-006 for split taxa. Those
were solved twice, differently, because there was no single place to say how confident a number
is. One record covers all of it, and makes each later gap a labeling problem rather than a
redesign.

### P-065 — The diagram shows which lengths are measured and which are modeled
Subject: `diagram.provenance`
Goals: G-001, G-004

Stroke style carries provenance. A **solid** stroke is a measured length; a **dashed** stroke is a
modeled one. `near-match` draws solid: it is a real measurement, of a real bird, taken with
callipers — of a sibling taxon, which the detail view says and the drawing does not.

That gives, for every species: **bill and tail solid**, **neck and body dashed**, and the **head
drawn hollow** — an unfilled outline, because dash is meaningless on a closed shape and absence of
fill says the same thing.

**The wing crossbar is split.** Its outer portions are the measured wing chord — carpal joint to
the tip of the longest primary — and are **always solid**. Only the inner span between them is
dashed, and only when the wingspan is modeled. A tick marks where the two meet. Where no chord is
available the crossbar is drawn whole, dashed if the span is modeled.

**Both pages explain the convention**, because a visual code nobody documents is decoration.

Rationale: the diagram asserted five proportions and explained none of them. A reader saw a heron's
long neck with no cue that it is entirely invented — the `~` markers of P-061 sit on the *numbers*,
and a reader looking at the shape never meets them. For a project whose central commitment is that
a modeled value is never presented as a measurement (CR-001), the drawing being silent was the
largest remaining gap.

The crossbar split earns its complexity: the outer segment's length stays true even when the span
model is wrong, so all the model's error is attributed to the inner segment, where it belongs. See
[CR-015](decisions.md).

### P-066 — The detail diagram is labeled with each segment and its measurement
Subject: `presentation.detail_view.diagram_labels`
Goals: G-001, G-003

**Every metric the detail view reports appears in the drawing.** The provenance table below it may
explain more about a number, but it may not be the only place a number appears: a reader looking
at the diagram should not have to discover that something was measured and left out of it. Adding
a metric to the table therefore means adding it to the diagram. This is stated as a rule rather
than a list because mass was reported and undrawn for days, and every check at the time named the
metrics it expected — mass among them nowhere.

In the detail view, each of the five segments is named and measured **beside the part it
describes**, joined to it by a leader line. The **wing chord** is labeled too, on the crossbar
rather than in the stack, because it names a horizontal length and belongs beside one. The
**total length** and the **mass** are labeled as well — see below. An estimated value carries the same `~` the tile footer
uses. The provenance table below the diagram is unchanged: it remains the place where the source or
model behind each number is explained.

The `~` must be in the **markup**. Writing it as CSS generated content is the obvious way and does
not work: `::after` generates no box on an SVG `<text>`, so the rule is silently inert. This is not
a hypothetical — it is how the marker was built in `a826b96`, the commit that introduced these
labels, and it shipped drawing every modeled value identically to a measured one. See
[AC-073](acceptance.md#ac-073).

Labels never overlap one another. Where two segments are too close together for their labels to
sit at both midpoints, the text moves and the leader line stretches — the leader, not the text
position, is what ties a label to its segment.

**Text draws over lines.** Every line in the diagram is emitted before every label, so a leader
that crosses a label passes behind the words rather than through them.

Rationale: the table already lists every segment with its value, but reading it means mapping a row
onto a part of a drawing and holding the correspondence in your head. Labeling the drawing directly
removes that step, and the detail view is the one place with room to do it — the tile is
deliberately dense because it exists to be compared against dozens of others.

The **wing chord** is labeled centred over the **left** outer segment — both outer segments are the
measured chord, and the left keeps the label clear of the wingspan. The **wingspan** sits just
beyond the right wingtip and on the crossbar's own line, so it reads as naming the whole bar rather
than any part of it.

This is why the detail diagram is drawn wider than the tile's proportions, and why the bird is laid
out in only part of that width. The right of the box is reserved for the label column and for the
wingspan label beyond the wingtip; a wide-winged bird is drawn slightly smaller rather than running
under its own annotations.

#### The whole-bird labels: total length and mass

**Total length and mass are labeled horizontally, centred on the axis, below the tail tip.** They
are the labels that name the whole bird rather than a part of it, and their placement says so:
under the foot of the figure, on the centreline, spanning nothing. They take no leader line —
there is no part to point at, and proximity to the end of the drawing is the whole of their
meaning.

**Mass is drawn as a number, and also as water.** The **cube of water that weighs what the bird
weighs** is drawn in the **lower-left corner**, in oblique projection, at the diagram's own
scale. Its side is ∛(mass), which is a unit conversion and not a model — one gram of water is
one cubic centimetre by definition — so it assumes nothing about the density of birds and
introduces no second scale. It is drawn away from the figure, and named in the diagram key,
because an unexplained shape beside a bird would read as part of the bird.

**The front face stands on the floor of the drawing, in the same place in every diagram**, and
the projection grows up and to the right from there. That fixed corner is what makes two cubes
comparable at a glance: only their size differs between birds, so the eye has a common origin to
measure from. A base that moved with the bird would cost that without changing a single number.
The depth is 0.4 of the side; cabinet projection's conventional 0.5 also clears the crossbar
everywhere, but leaves 14px where this leaves 21, and no unit test can see that margin because
the fixture holds five birds.

The **Presentation panel explains the cube**, not only the diagram key: that panel is where a
reader is taught to read the drawing ([P-035](#p-035)), and the cube is the one mark on the
figure that is not a length. It says what the cube is, that every cube stands on the floor of its
own diagram so two can be held against each other, and that a cube root flattens — a cube twice
as wide is a bird eight times as heavy.

**Mass is labeled under its cube, not in the stack under the tail**, so the number sits with the
shape it describes. On a small cube the label is wider than the face and is held clear of the
left edge rather than centred past it.

The **index tiles carry the cube too**, unlabelled ([P-031](#p-031), CR-020). It is there for
what the rest of the figure cannot say: **bulk against elongation**. Read against
the bird's own length it is about **27%** of the drawing height for alcids, which are dense
divers, and **10–13%** for long-tailed light-bodied species like the Scissor-tailed Flycatcher
and the magpie. Length and wingspan cannot tell those apart.

*(Re-measured after [CR-021](decisions.md#cr-021), which lengthened heavy birds and so shrank the
cube's share of them; the alcid figure was 28–29%. The ratio is scale-invariant, so it is the same
under either mode of P-042.)*

**The cube compresses, and the number is there because it does.** A Bushtit and a Common Raven
differ in mass by **175×** and their cubes differ by 5.6×, which is correct for equal-volume cubes
and understates the gap badly. So the number carries the magnitude and the square carries the
sense of it; a reader who takes the squares alone as the mass comparison will read them wrong.
See [CR-019](decisions.md#cr-019), which reverses an earlier decision to draw no shape at all, and
[P-043](#p-043) for why a disc or bar sized by mass — a second, arbitrary scale — is still
refused.

Mass also behaves differently from its neighbours on provenance, and the drawing shows it:
**measured for 831 of 841 species**, against a total length modeled for every one. It is usually
the only unmarked number on the figure, and the only solid-stroked whole-bird mark.

It belongs on the diagram for the reason the segments do, only more so. Total length is the number
BirdView exists to give a reader intuition about (G-001); every other label names a component of
it. Leaving it only in the table below made the drawing annotate its own parts while staying silent
about the thing it is a drawing *of*.

A **bottom strip is added** for it. Without one the label is clipped: the detail view scales the
bird to fill the box, so a bird whose height binds has its tail tip exactly on the bottom edge and
nothing below it.

The strip **extends the box downward rather than shrinking the bird into it** — the opposite of
what the label gutter does horizontally, and deliberately so. Width is fixed by the page column, so
the gutter has to come out of the bird. Height is not: the drawing can simply be taller, and the
bird keeps the size [P-034](#p-034) requires. Taking it out of the bird instead would cost 8.5% of
the detail view's height to buy a one-line label, which is a bad trade and would put this
requirement in conflict with P-034 rather than beside it.

**Total length states the bird's length, not the drawing's height.** For the folded-neck families of
[CR-011](decisions.md#cr-011) these differ: the neck's S spans 2/π of its length, so the figure is
6–9% shorter than the labeled number across the 14 species affected. That is the rule the neck
label already follows — CR-011 reports neck length in full while drawing it compressed — and the
total length label inherits it rather than introducing it. A label derived from the drawn extent
would be a different and worse thing: it would name a quantity no source states and no table lists.

### P-064 — The detail view links to the species on iNaturalist
Subject: `presentation.detail_view.external_link`
Goals: G-001, G-004

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
[CR-016](decisions.md).

BirdView **links and never copies**. No photograph is embedded, hotlinked, cached or reproduced;
the reader arrives on iNaturalist, where each image carries its own license. T-002 forbids the
embedding independently.

Rationale: BirdView will never have photographs, and a reader surprised by a size wants to see the
bird. Linking is the whole answer. The confidence gate exists because a link to the *wrong* bird is
worse than no link on a page whose purpose is being careful about what it claims — and an absent
link is legible, where a search URL that always "works" hides the failure. See
[CR-012](decisions.md).

### P-063 — The detail view explains each metric's provenance
Goals: G-001, G-004

In the detail view (P-034), **every** metric — the three main metrics and all five diagram
segments — is shown with its value and its provenance stated in plain language: naming the source,
naming the taxon actually measured for a `near-match`, and naming the model for a `modeled` value.

Rationale: the tile marks *that* a value is uncertain; this says *how*. "Wingspan estimated from
wing chord", "mass measured on Herring Gull, a species this was split from in 2024", and "neck
length from the per-family ratio for Ardeidae, 0.35 of the body axis" are very different caveats,
and a birder who cares enough to click deserves the difference.

Showing the segments individually is what makes the diagram auditable. Without it, the drawing
asserts five proportions and explains none of them.
