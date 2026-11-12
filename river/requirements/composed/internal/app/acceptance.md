package app

# app — acceptance criteria

**Composed, not authored.** Composed with [app.md](app.md); do not edit by hand, recompose.

How each requirement in [app.md](app.md) is decided. Each criterion names what it covers in
`Covers:`; mutable guardrail bindings and `Watched failing:` observations live in
[conformance.md](../../../../conformance.md).

**Uncovered:** [pagesMeetNonTextContrast](app.md#pagesmeetnontextcontrast). No criterion decides
WCAG 1.4.11 yet; the stylesheet checks skip shapes and the rendered check measures text only.

---

## everyPageIsFound

**Every route and every page template yields a page the checks see**

Covers: [tech.webPages](../../../local/tech/accessibility.md#webpages)

**Method:** crawl the server from a handful of seeds — the pages nothing links to — following
every same-origin link and source, and require that every route the server registers, every page
template it can render, and each of the three content types [tech.webPages](../../../local/tech/accessibility.md#webpages) names produced at least
one page.

It decides the domain, not a criterion over it: it covers [tech.webPages](../../../local/tech/accessibility.md#webpages), and the page
requirements rely on it without being decided by it. This is what lets them say *every page*
without listing the pages. A route or a
template added later that the crawl cannot reach fails here rather than going unchecked.

## pagesKeepToTheStylesheet

**Every templated page meets the stylesheet contract's assumption**

Covers: [stylesheetContract](app.md#stylesheetcontract)

**Method:** on every discovered page rendered from a template: it links `styles.css`, has no
`<style>` element, sets no color as an SVG or HTML attribute, and sets a color inline only as one
of the palette's generated pairs. There is no list of exempt pages, because the contract is closed.

## renderedTextContrast

**Text measured in a real browser clears 7:1 on every kind of page**

Covers: [pagesMeetTextContrast](app.md#pagesmeettextcontrast)

**Method:** render pages in headless Chrome with every `<details>` open, and measure every visible
run of text against what is drawn behind it, found by hit-testing under the text rather than by
inheritance. Rendered: two pages of every kind (route and template), and every page outside the
stylesheet contract — the SVG documents — in full. A plain-text page is not rendered: it sets no
colors.

This is the end-to-end check on [stylesheetContract](app.md#stylesheetcontract). It gates where Chrome is installed and skips
without it ([bindings.md](../../../../bindings.md)).

## pagesFitAPhone

**Pages drawn 360 pixels wide stay inside the window and inside their boxes**

Covers: [product.fitsAPhone](../../../local/product/presentation.md#fitsaphone)

**Method:** render two HTML pages of every kind (route and template) in headless Chrome, each in a
frame 360 pixels wide, which is a viewport of its own (Chrome will not open a window narrower than
500), with every `<details>` open. Fail on any element that reaches past the frame's right edge,
reporting only the outermost, and on any run of text that extends outside the nearest ancestor with
a background or a border, unless an element between them clips or scrolls. Fail too if the page is
wider than the frame, and refuse to judge a page laid out at any other width.

It gates where Chrome is installed and skips without it, like [renderedTextContrast](#renderedtextcontrast).

## boundariesResolveKnownPoints

**Known places resolve to the regions they are in**

Covers: [tech.regionBoundaries](../../../local/tech/delivery.md#regionboundaries)

**Method:** look up known coordinates in the committed boundary file with the Go reference
lookup — land polygons first, then sea boxes — and require: Sacramento → CA, Toronto → ON,
Honolulu → HI, Anchorage → AK, Washington → DC, Midway → MidwayAtoll, a point in the Gulf off
Louisiana → GulfOfMexico, a point in the open North Atlantic → NorthAtlantic; Mexico City and
London → no region. The expected answers come from geography, not from a run of the code.

## boundariesCoverEveryRegion

**Every land region has a boundary, and the file stays small**

Covers: [tech.regionBoundaries](../../../local/tech/delivery.md#regionboundaries)

**Method:** every land region in `internal/region` has at least one polygon in the boundary file
and every polygon belongs to one; the file's sea boxes equal `internal/region`'s; and the file
compressed is at most 200 KB.

## hereLookupAgreesInTheBrowser

**The browser finds the same region Go does, and says when it cannot**

Covers: [product.hereSelectsLocalRegion](../../../local/product/filtering.md#hereselectslocalregion)

**Method:** on the home page, in headless Chrome with the location stubbed, press Here at each
point of boundariesResolveKnownPoints and require the navigation the Go lookup predicts; with
permission refused, location unavailable, and a point outside every region, require the matching
message and no navigation. Arriving by a `#here-and-now` link goes to the page choosing the
reader's region by hand gives, with the month set to the browser's; refused, it stays and says so in the welcome
line (`TestHereAndNowLink`, watched failing on a link that set the wrong month and on a refusal
message that did not say refused). Until 2026-10-27 this ran on the classic view, and also pressed
Now.

## hereSendsNothingExtra

**The server cannot tell that Here was used**

Covers: [product.locationStaysInBrowser](../../../local/product/filtering.md#locationstaysinbrowser)

**Method:** in headless Chrome, record every request: the boundary file is requested when the
home page loads, before anything is pressed; pressing Here then makes exactly one request,
identical to the one choosing that region alone in the Region setting would make, and no request
carries a coordinate.

## tierHistogramMatchesTiers

**The histogram has one bar per tier, sized and labeled from the tiers themselves**

Covers: [product.tierHistogram](../../../local/product/ordering.md#tierhistogram)

**Method:** on the rendered home page, every bird shown, under several orderings and a filter, the
histogram has exactly one bar per tier, in order, a single tier included (one pinned bird draws
one bar at 100%; [#6](https://github.com/google/birdview/issues/6)); each bar links to its
tier's heading; each bar's height is its
tier's count as a percentage of the largest tier's, rounded; and each bar's label names the tier
and ends with its count. Under the size orderings, a stretch with no birds is one empty bucket per
whole 15% step it spans, between the bars on either side.

## nextViewDrawsOneSet

**The home page draws the set its address names, at one scale, and every control on its count line keeps to the rule**

Covers: [product.nextView](../../../local/product/presentation.md#nextview)

**Method:** over the real dataset, with the clock fixed to October. With no month or `month=all`,
the page shows every bird; with October, May, or May and June, exactly those months' birds; each
most often recorded first. October narrows the set from all. `month=now` answers 303 to the same
address with `month=Oct`, never stored. The count line has no Show setting and starts with Order by, set to frequency.
Every bird on four pages is drawn at one pixels-per-millimeter, and the longest fits its tile.
Ordering by mass runs smallest first; ordering by frequency runs most often recorded first.
Family and region narrow together, and an emptied filter shows all, not nothing. Every drawn
heron has an unpinned pin, and no pin excludes a bird. Sharp-shinned Hawk, found by name, pins
from the find list with the family and the box's text kept, so the list stays open with the hawk
pinned and its × goes to the page without the text; it is drawn pinned and unpins to the address
it came from, text kept; a heron already drawn pins in place and unpins
back. Without JavaScript, five queries give the matcher's count and list the birds only at 10 or
fewer. In headless Chrome, typing six queries (one in capitals, one over a lowered limit, one
matching nothing) shows the same count line and lists the same birds, with the same pins, on or
off, as the server's page for each. The Family setting draws every family in its badge's colors
with its description. Choosing from all selects
only that one, choosing again removes it, and changing a filter drops the find box and keeps how many are shown.
The home page draws the first 100 of its 1,041 and a link to all of them (`show=all`) in
the same order, in the gap row at the end; the herons draw every heron with no link. `show=50`
draws 50, says 50 in the box, and is kept by a filter but not Reset; `show=all` is kept by a
filter too. Under the heading, the Show row reads First [box], a form that sends every other
setting but show and the find box, then All with the count the settings match, Pinned
(2), and Clear pins, the chosen one marked, all on one line at 360px (`TestShowRowFitsAPhone`); Show is no longer a closed setting. The page runs, top
to bottom: the header, the heading, the Show row, the pinned list, the find box, one count row
(Order by through Names, then Reset), the histogram, the sentence, and the birds. The sentence
names the order, what the histogram counts, and what is drawn: "Mass distribution of all", "Frequency
distribution of all N birds", "Length distribution of N pinned birds", "Tail length distribution of
all N birds, showing the first 100", and with a pin past the first, ", plus 1 pin"; and the
histogram's tiers count all N birds while the first 100 are drawn (`TestNextShowAll`). The heading reads "Showing pinned birds" under Pinned only, and
otherwise "Showing birds in" the regions and months, then "· Family:", "· Habitat:", and "· Diet:"
with what is chosen in each (`TestNextHeading`).
`show=100`, `show=`, `show=0`, `show=-3`, and `show=zero` redirect to the address without show;
`show=50`, `show=all`, and `show=pinned` do not. By length, reversed, and by mass with every bird shown, the tiers hold every
matching bird, each drawn bird in order under its tier's heading, one heading per drawn tier;
the histogram has a bar per tier linking to its heading, with show=all when the tier is not
drawn; reversed tiers mirror the forward ones, headings included; frequency tiers read "1 in N". Every length tier holds birds within 10% of its first, the next tier starting past that, and
each drawn tier's heading has a back to top link to the header. From an address setting every parameter, under Pinned only and under a
count, every link on the page changes only its own setting (Order by, also rev), none but × and
Reset drops the find box's text, Reset keeps only the pins, Clear pins drops them (and Pinned
only), × drops the text, the forms send every parameter but their own, Here replaces only the
region, the birds around you now replaces the region and month (and leaves Pinned only), and a
tile's species page carries only the units and names; in headless Chrome, a species page opened
from the home page shows Pin or Unpin as the home page pins the bird, and its Pin goes back and
leaves the bird, so that the home page, loaded again, requests its own address with bird= that
bird; without the script Pin links to /?pin=<id>, and a bird left off the home page has none; Back
on a species page reached from the home page goes back by history, and from a shared link follows its link to
/?units=metric; choosing Nov from Pinned only stays
Pinned only. The summaries read frequency ↓, length ↑, mass ↓ reversed, all by default, Oct, May or Jun,
all, imperial or metric, and North Amer or world, never "(now)", and Units offers Imperial and Metric; the header links stick figures to the guide. The header ends with Click a bird and How to read a drawing and has no
key line; its Try line has no "New here?" before it, its five comparisons each pin six real birds shown alone by length, the
grassland hunters draw only grassland birds that hunt animals, herons
and ducks draw only those families in every month, and the birds around you now is the address
with #here-and-now, its template replacing the region and setting the month to a slot for the
browser's month, and keeping
every other setting; in headless Chrome, from /?by=mass&month=May in California it lands on
/?by=mass&month=<the browser's month>&region=CA and, refused, says the page is unchanged and to choose under Region. The
same holds when the link is clicked on the page, which changes only the fragment. In headless
Chrome, after here.js runs, Now links to the browser's month with every other setting kept
(`TestNowIsTheBrowsersMonth`). In headless Chrome, the script's × goes where the server's does, and a
tile's pin clicked with text in the box carries the text. In headless Chrome, clicking the First button or the First choice's edge
sends the typed count, and clicking into its box does not; without the script, First is a submit
button. The chosen order links to its reverse and back, every other
order links to itself unreversed, reversed mass runs largest first, and reversed frequency puts
the most recorded duck last. Every home page, with no month included, carries an ETag; the
`month=now` redirect is no-store with no validator, and keeps every other setting. Every
count-line setting is a `<details>` sharing one name, so the browser opens one at a time, and
Reset links to /. The pinned list names each pinned bird in alphabetical order by displayed name, linking to its tile, with a pin
that unpins it, and has no Compare of its own; no pin link jumps the page, and in headless Chrome a page loaded by a pin opens
scrolled where the pin was clicked. Reset keeps the pins, in order, and drops every setting; Clear pins,
before it and only when something is pinned, drops the pins and keeps every setting. The phone-width and contrast checks reach the home page through the crawl, open
every setting at once, and measure after the placing script has run. Pinning the default set's
least recorded bird, past the first 100, draws it after one gap row and none after it, and the count line says "plus 1 pin" ("plus 2 pins" for two); pinning two mid-order birds by length draws gap rows before, between, and after them summing to Hidden, each linking to Show all, and Show all draws no gaps. Pinned
only links to `show=pinned` with the order kept and draws nothing; with two pinned it
draws those two by mass, unpinning one leaves the other, and First brings the
default birds back with both; `show=all` matches what the default does. Here, filled with FL, is
the page with only Florida chosen, every other setting kept, and the pinned-only
view left, and the page carries the button, its status, and here.js. Month's choices are one
line, All, Now, and the twelve months, with a rule only before January; All is chosen with no
month named or with `month=all` and links to no month; Now is never chosen, links to `month=now`,
and is the one link marked `data-now` for here.js; with October named, October alone is chosen. In headless Chrome, choosing
May in Month loads a page with Month open and nothing else, and loading it again after no choice
opens nothing. In headless Chrome at 1280 px and in a 360 px frame, on the birds by length with
the two longest names pinned, every pin stands in its tile's top-right corner, within the tile's
11px of border and padding of the top and right edges, and the common name ends short of it;
every eBird link stands in the lower-right corner, its foot within a pixel of the bottom line's and
its top below the rule, and the scientific name ends short of it; every mass label's baseline is
within a pixel of its cube's foot; and every drawing starts within
a pixel of the others in its row, tiles with notes and tiles stretched to a taller neighbor
included ([decisions.tileNumbersFirst](../../../decisions.md#tilenumbersfirst), [decisions.tileMeasurementLines](../../../decisions.md#tilemeasurementlines)).

## ebirdLinkOnTilesAndPages

**Every tile and species page with an eBird code links to the species on eBird, in words**

Covers: [product.ebirdLink](../../../local/product/presentation.md#ebirdlink), [product.tileContents](../../../local/product/presentation.md#tilecontents)

**Method:** on the home page with every bird shown, the heron's and the robin's tiles each carry a
link reading **eBird** whose href is exactly `https://ebird.org/species/<code>`, named the bird's
name and "on eBird", and placed after the tile's own link closes rather than inside it; their
species pages link to the same address in a sentence naming the bird and eBird. With the stilt's
code removed, neither its tile nor its page links to eBird. No page links to an iNaturalist taxon
page, and neither page carries an image ([decisions.ebirdLinkReplacesInaturalist](../../../decisions.md#ebirdlinkreplacesinaturalist)). In headless
Chrome, at least nine in ten measured tiles carry the link, in the lower-right corner
(`TestNextTileCornerAndRows`, under [nextViewDrawsOneSet](acceptance.md#nextviewdrawsoneset)).

## everySpeciesHasAnEBirdCode

**Every species has an eBird code, its simplest, and answers to the rest**

Covers: [product.ebirdLink](../../../local/product/presentation.md#ebirdlink)

**Method:** over the built dataset, every species carries an eBird code, which is also its id.
Ruddy Duck, White-necked Petrel, and American Barn Owl, which AviList gives none, have `rudduc`,
`whnpet`, and `brnowl`, with their former scientific-name addresses as aliases. Fork-tailed Swift
`fotswi`, Bronzed Cowbird `brocow`, and Siberian Stonechat `sibsto1` take the simpler of two codes
and keep AviList's as an alias; Marsh Wren `marwre` and Willet `willet1` keep AviList's and take
Cornell's as an alias; Mallard keeps `mallar3` with none. No code, id or alias, answers for two
species, drawn or not. `taxonomy.SimplerEBirdCode` takes fewer digits, then the shorter, then
AviList's, and `taxonomy.SpeciesCodes` applies it ([decisions.ebirdCodeAliases](../../../decisions.md#ebirdcodealiases)).

## everyCodeReachesItsSpecies

**Every code a species has reaches it**

Covers: [product.ebirdLink](../../../local/product/presentation.md#ebirdlink), [tech.stableUrls](../../../local/tech/delivery.md#stableurls)

**Method:** with a second code on the fixture heron and a former address on Passenger Pigeon,
`/species/<alias>` answers 301 to the species' own address, query kept, for a drawn bird and an
undrawn one; `pin=<alias>` and `bird=<alias>` redirect to an address pinning the own code, which
the alias no longer appears in; the find box lists the heron alone for its alias; and a code no
species has is a 404.

## textMeetsSizeFloor

**No text renders below 13px, and running text is 15px, 16px on a phone**

Covers: [tech.textSizeFloor](../../../local/tech/accessibility.md#textsizefloor)

**Method:** two HTML pages of every kind, found by the crawl of [everyPageIsFound](acceptance.md#everypageisfound),
render in headless Chrome at 1280px and framed at 360px with every `<details>` open; every visible
run of text is measured at its drawn size, its font size times the scale of any SVG it sits in,
and the page's own font size is read. It fails on any run under 13px and on running text under
15px, or 16px in the frame. SVG documents served on their own, the pipeline diagram, are drawn
images rather than pages of text and are outside it.

## pinnedGroupLeads

**Pinned birds lead the home page as a group, and stay in their tiers**

Covers: [product.nextView](../../../local/product/presentation.md#nextview)

**Method:** with the hummingbird and the heron pinned, by length, a heading naming its 2 birds
comes before the first tier, then their two tiles, shortest first, with anchors `p-<id>`; both
are still drawn in their tiers below; the pinned list links to `#p-<id>`; the histogram is the
same as with nothing pinned; with nothing pinned there is no group; and under Pinned only,
reversed, the page is the group alone, heron first ([decisions.pinnedGroupLeads](../../../decisions.md#pinnedgroupleads)).

## bandingCodesShown

**Birds carry the Bird Banding Laboratory's four-letter codes, and the site uses them**

Covers: [product.bandingCodes](../../../local/product/dataset.md#bandingcodes)

**Method:** the vendored list parses to over 1,100 one-species rows, each code once, with American
Robin AMRO, Blue Jay BLJA, Western Grebe WEGR, and Bicknell's Thrush BITH, and without the
unidentified Gray-cheeked/Bicknell's Thrush. The built dataset gives AMRO, BLJA, AMCR, NOCA, and
BAEA to their birds, at least 850 species a code, and every code is four capital letters on one
species. On a small list, the join takes the scientific name, picks the row named as the species
where two share one, falls back to the one row with the English name, and gives a code two
species claim to neither.

American Robin's tile and species page show "AMRO (amerob)" beside its name; the find box finds it
alone by AMRO or amro and lists it the same way; in headless Chrome the script's list reads as the
server's, codes included; `/species/AMRO` is a 404, since addresses stay the eBird code.

## stableSpeciesId

**A species id is unique, stable, and fit for a URL**

Covers: [product.englishNameChoice](../../../local/product/diagram.md#englishnamechoice)

**Method:** property test over the built dataset, plus
assertions on the served routes

**Assertions, uniqueness:** no two species share an id. This is not a formality — AviList lumps
two BirdView species into taxa already in scope, and had they been lumped in place rather than
dropped, two birds would have shared one URL.

**Assertions, URL fitness:** every id survives `url.PathEscape` unchanged and none is empty, so
no species needs encoding to be linked.

**Assertions, the fallback:** the three species with no eBird code carry a scientific-name slug
instead, and are reachable at it.

**Assertions, the routes:** `/species/<code>` resolves. Until 2026-10-27 the comparison view also
accepted codes in `species=`; it went with the classic view.

## nameLocaleToggle

**The name locale is a toggle, and only names move**

Covers: [product.englishNameChoice](../../../local/product/diagram.md#englishnamechoice)

**Method:** assertions on the rendered pages under both
locales

**Assertions, both names are served:** with no parameter the page shows the **North American**
name; with `names=world` the same URL shows the **world** name. The fixture gives every species a
world name that differs from its North American one, so a toggle that silently serves the same
string either way cannot pass.

**Assertions, the id does not move:** the species code and the URL are **identical under both
locales**. This is the property the whole change rests on, and the one an alpha code could not
have had.

**Assertions, the control:** the home page's Names setting offers both locales, marks the active
one, and its links preserve the rest of the view state. Its check went with the classic view on
2026-10-27; this is now decided through [app.nextViewDrawsOneSet](../../../composed/internal/app/acceptance.md#nextviewdrawsoneset) (`TestNext*`).

**Assertions, the fallback:** a species with no North American name renders its world name under
either locale, rather than an empty header.

## sortsAscending

**Sorting is ascending by the selected metric**

Covers: [product.oneSelectedMetric](../../../local/product/ordering.md#oneselectedmetric), [product.ascendingOrder](../../../local/product/ordering.md#ascendingorder)

**Method:** property test for each of the three
metrics

## billAndTailOrderings

**The home page sorts by bill and by tail, and says so**

Covers: [product.oneSelectedMetric](../../../local/product/ordering.md#oneselectedmetric), [product.ascendingOrder](../../../local/product/ordering.md#ascendingorder)

**Method:** rendering test. The home page sorts
ascending by each and names it on the count line.

## tierHeadingsNameTheOrdering

**Every section heading names the ordering it belongs to**

Covers: [product.sectionHeadersStateRange](../../../local/product/ordering.md#sectionheadersstaterange)

**Method:** rendering test. Every heading names the ordering
as Order by labels it: *Tail Tier 1* under Tail.

## tierHeadingsHoldTheirBirds

**Each tier heading states only its own birds, and the headings rise strictly**

Covers: [product.sectionHeadersStateRange](../../../local/product/ordering.md#sectionheadersstaterange)

**Method:** rendering test. Each heading states only its own
birds, and the page's headings rise strictly in one precision, so a gap shows as a gap.

## tierHeadingDigits

**Tier headings use the fewest digits that keep every value apart**

Covers: [product.sectionHeadersStateRange](../../../local/product/ordering.md#sectionheadersstaterange)

**Method:** unit test. The headings' precision is the fewest
that keep every stated value apart from its neighbors, as well as each tier's own ends.

## tierCountFollowsTheRange

**A tier's species count follows its range, in the count line's style**

Covers: [product.sectionHeadersStateRange](../../../local/product/ordering.md#sectionheadersstaterange)

**Method:** stylesheet test. The tier's *N birds* sits right
after the range in the count line's style.

## sectionBoundariesObeyRule

**Section boundaries obey the N% rule**

Covers: [product.sectionBreakRule](../../../local/product/ordering.md#sectionbreakrule), [product.sectionThresholdDefault](../../../local/product/ordering.md#sectionthresholddefault)

**Method:** property test over generated inputs

**Assertions:** for every section, `max / min < 1 + N/100`; and the first member of each section
is `>= (1 + N/100) x` the **minimum** of the previous section. Both hold for N across its full
adjustable range, and for N = 15 by default.

## sectionHeadersMatch

**Section headers match their contents**

Covers: [product.sectionHeadersStateRange](../../../local/product/ordering.md#sectionheadersstaterange)

**Method:** unit test

**Assertion:** each header's stated range equals the min and max of its members in the selected
metric, and its count equals the number of members.

## segmentsSumExactly

**Segments sum exactly to total length**

Covers: [product.segmentsSumToLength](../../../local/product/diagram.md#segmentssumtolength)

**Method:** property test over every record

**Assertion:** the five **serialized** segment values sum **exactly** to the serialized total
length — not within a tolerance.

## crossbarToScale

**The wing crossbar is to scale with the vertical axis**

Covers: [product.crossbarAtVerticalScale](../../../local/product/diagram.md#crossbaratverticalscale)

**Method:** property test over computed geometry

**Assertion:** `crossbarLength / bodyLength == wingspan / totalLength` for every tile, within
floating-point tolerance.

## diagramPartsInOrder

**The diagram has the specified parts in the specified order**

Covers: [product.segmentComposition](../../../local/product/diagram.md#segmentcomposition)

**Method:** unit test over rendered output

**Assertion:** five vertical segments top to bottom — bill, head, neck, body, tail — plus one
crossbar; the tail is rendered with a stroke style distinct from the body.

## proportionsDifferentiate

**Proportions differentiate species as the requirement claims**

Covers: [product.proportionDifferentiates](../../../local/product/diagram.md#proportiondifferentiates)

**Method:** ordering assertions over a small set of
species pairs

**Assertion:** a dowitcher's culmen fraction exceeds a robin's; a heron's neck fraction exceeds a
sparrow's; an accipiter's tail fraction exceeds a duck's.

## tileMarksNonMeasured

**The tile marks every non-measured metric**

Covers: [product.tileMarksModeled](../../../local/product/measurements.md#tilemarksmodeled)

**Method:** rendering test asserting a marker is present
for each metric whose `kind` is not `measured`, and absent otherwise

## detailShowsProvenance

**The detail view shows every metric and segment with provenance**

Covers: [product.metricProvenanceExplained](../../../local/product/measurements.md#metricprovenanceexplained), [product.tileOpensDetail](../../../local/product/presentation.md#tileopensdetail)

**Method:** rendering test

**Assertions:** the detail view shows all three main metrics **and all five diagram segments**
with their values; for each it names the source; for a `near-match` it names the taxon actually
measured; for a `modeled` value it names the model. The detail view reproduces the tile's header,
diagram, and main metrics — it is the same layout enlarged, not a different screen.

## noServerNetworkRequest

**The server makes no network request**

Covers: [tech.selfContainedServer](../../../local/tech/delivery.md#selfcontainedserver)

**Method:** rendering test scanning for external
**subresources** — any `src`, and `href` on a `<link>`

## tileCarriesHeaderAndMetrics

**The tile carries the header, diagram, and main metrics**

Covers: [product.tileContents](../../../local/product/presentation.md#tilecontents)

**Method:** rendering test

**Assertions:** every tile shows the common name, the eBird code in parentheses, the scientific
name, a family badge, an SVG diagram, and the main metrics **named in full** — `Length`,
`Wingspan`, `Mass`, not `L`, `W`, `M` ([decisions.massCubeOnTiles](../../../decisions.md#masscubeontiles)). On the crow's tile they run in
order: the common name, the banding code, the eBird code, length, bill, tail, wingspan, chord, and
the share line, then the drawing with the mass label in it, then the family badge and the
scientific name; the badge reads the family's scientific name alone
([decisions.tileMeasurementLines](../../../decisions.md#tilemeasurementlines)). On every fixture tile the mass label's baseline is on the cube's foot, it
starts right of the cube, and it is the drawing's only text; and the stylesheet sets `.tl` in
`.ft`'s font.

## familyFilterSelectsAndClears

**The family filter selects and clears**

Covers: [product.familyFilter](../../../local/product/filtering.md#familyfilter)

**Method:** rendering test over the home page's `?family=` and its Family setting

**Assertions:** filtering to one family shows its species and excludes others; no filter shows
all; every entry is labeled with its scientific name and its common name; **no entry carries a
species count**, and a single count sits above the list instead; the selector is grouped by
taxonomic order; groups appear in checklist sequence rather than
alphabetically; and each family sits inside its own order's group. Until 2026-10-27 each order was also a
toggle for its families, and **All** one for the whole set; both went with the classic view.

## unitToggleSwitches

**The unit toggle switches imperial and metric**

Covers: [product.imperialByDefault](../../../local/product/presentation.md#imperialbydefault)

**Method:** rendering test over the tile footer, plus a
unit test pinning the conversions

**Assertions:** the tile's two measurement lines carry every length on them, at least four, and
its mass label one mass value, in the selected system's units, and they differ between the two
systems ([decisions.tileMeasurementLines](../../../decisions.md#tilemeasurementlines)); mass converts as `1 oz = 28.349523125 g`.

## detailDrawsFullScale

**The detail view draws its bird full scale**

Covers: [product.tileOpensDetail](../../../local/product/presentation.md#tileopensdetail)

**Method:** rendering test over the smallest fixture
species, with and without an explicit `?scale=`

**Assertion:** the diagram fills the detail view's height whatever `scale=` the address carries.

## badgeTextInheritsColor

**Text on a generated background inherits that background's text color**

Covers: [stylesheetContract](app.md#stylesheetcontract)

**Method:** structural check over `styles.css`

**Assertions:** no rule targeting a descendant of `.chip` or `.badge` sets its own `color`, and no
such rule dims text with `opacity`.

## attributionComplete

**Attribution is complete, and claims nothing unearned**

Covers: [sourceCreditsAndLicenses](app.md#sourcecreditsandlicenses)

**Method:** rendering test over the footer

**Assertions:** each CC BY source is credited with creator, a link to its deposit and a link to the
license — **AVONET, Fox, and AviList by name and DOI**; the ABA checklist is never credited, scope having moved to BirdView's recurrence derivation over GBIF ([decisions.scopeFromRecurrence](../../../decisions.md#scopefromrecurrence)); the footer discloses that lengths are modeled; the detail view
carries credit too, being a shareable URL of its own; and **no page claims a permission the
project does not hold**. **The two GBIF datasets** are cited in the home page's credits by
publisher, title, DOI as a URL, and "accessed via GBIF.org", and the home page, the species page,
and the page for an undrawn bird each give iNaturalist's license as CC BY-NC 4.0, never as CC BY
4.0 with eBird's ([decisions.gbifTermsMet](../../../decisions.md#gbiftermsmet)).
by reinstating a "used by permission" claim. **AviList added 2026-10-02**
([decisions.avilistReplacesAba](../../../decisions.md#avilistreplacesaba)) and watched failing by deleting its credit while leaving its names
on every tile — which is the shape the obligation actually has: CC BY permits the use *because*
of the attribution, so an unasserted credit is a license breach waiting to happen.

The second half exists because the first version of this footer said the ABA checklist was "used by
permission". It is not — that request has not been sent. Claiming a right you have not been granted
is worse than claiming none, because it misleads the rights holder's own readers. Remove that
assertion only when a written grant exists and is recorded in [imports.md](../../../imports.md).

## foldedNeckSameStroke

**A folded neck is the same stroke in less height**

Covers: [product.segmentComposition](../../../local/product/diagram.md#segmentcomposition), [product.segmentsSumToLength](../../../local/product/diagram.md#segmentssumtolength), [product.tileOpensDetail](../../../local/product/presentation.md#tileopensdetail)

**Method:** unit tests over the geometry plus
assertions on the rendered page

**Assertions, geometry:** for a folding family the neck is an S and for every other family it is a
stick; the S's arc length `2πr` equals the scaled neck exactly; its vertical span is `2/π` of that
length, checked against the ratio rather than against whatever the code produces; the two arcs join
halfway; and the whole drawing is shorter by exactly what the neck saved and by nothing else, so no
other segment may shift.

**Assertions, length as data:** a folding bird and a non-folding bird of identical total length
share a section, and the scale reference is the full total length in both scale modes — checked
against a section holding **only** the folding bird as well as a mixed one.

**Assertions, rendering:** the heron's neck reaches the page as `<path class="neck">` and every
other bird's as `<line class="neck">`, never both and never neither; and the Neck row of the detail
view still reports the full modeled length, read out of that row rather than searched for in the
whole page.

## foldingFamiliesHaveNecks

**A family that folds its neck has a neck to fold**

Covers: [product.segmentComposition](../../../local/product/diagram.md#segmentcomposition)

**Method:** unit test cross-checking the two neck
tables

**Assertion:** every family in the folded-neck set has an **explicit** entry in
`morphology.NeckRatios`, rather than falling back to the default ratio.

## noTextBelowContrastFloor

**No text or link falls below its contrast floor**

Covers: [stylesheetContract](app.md#stylesheetcontract), [pagesMeetTextContrast](app.md#pagesmeettextcontrast)

**Method:** two checks parsing
the stylesheet **as embedded in the binary**, not off disk

**Assertions, measured:** every `color` a rule declares — with custom properties resolved — clears
**7:1** against both `--bg` and `--panel`, or, for a rule that also sets a `background`, against
that background. **Until 2026-10-16 such rules were excluded** as self-contained and governed by
[tech.badgeContrast](../../../local/tech/accessibility.md#badgecontrast); that was true of badges and false of the selected button and the pick
control, whose hand-written pairs nothing read. Watched failing since: selected-button text set to
`#2a7fa8` is reported at 2.42:1 against its own `--accent` background. That change was first
caught by [app.renderedTextContrast](../../../composed/internal/app/acceptance.md#renderedtextcontrast), with this check passing.

**Since 2026-11-01 no SVG text is styled** ([decisions.textSizeFloor](../../../decisions.md#textsizefloor)): the drawing's one label is page text,
held to the floor as `color`. The SVG text check stays and reports that it found nothing to check;
a `fill` on a text rule added later is held to the floor again. Its old guard, failing when it
checked nothing, now cannot tell a broken parser from an absent subject, so the parser is shown
working by the text tokens it finds instead.

**Assertions, SVG text:** every `fill` a rule declares **on a text element** clears the same 7:1
against both surfaces. Which rules those are is derived rather than listed: a class token that any
rule gives a font property to is a text token, and any rule naming one is styling text. So
`.dlabel { font-size: 11px }` makes `dlabel` a text token, and the `fill` on `.dlabel.chord` is a
text color. A `fill` on a shape — the head oval, the mass cube — is untouched, being a graphical
object under WCAG 1.4.11's 3:1 instead.

Deriving it is the point. An explicit list of text selectors would need maintaining by the same
person who just added an unchecked one, which is the failure this assertion exists to fix.

**Assertions, structural:** the stylesheet carries top-level `a` and `a:visited` rules, so no link
can fall back to the browser default.

## ownRenderableFavicon

**The favicon is BirdView's own, and is actually renderable**

Covers: [tech.selfContainedServer](../../../local/tech/delivery.md#selfcontainedserver)

**Method:** assertions over the served asset and both pages

**Assertions:** `/static/favicon.svg` is served with status 200; it **parses as XML** and its root
element is `svg` with a `viewBox`; both the home page and a detail page reference it; and
`/favicon.ico` returns 404 rather than the index page, which the catch-all route makes worth
pinning.

**The palette assertion moved to [faviconFromPalette](acceptance.md#faviconfrompalette)**, where it is derived from the stylesheet. It
lived here as a search for three hardcoded literals, one of which existed only in the favicon, so
it compared the icon against a copy of itself and could not fail. This criterion keeps what it can
actually see from the server: that the asset is served, parseable, and referenced.

## diagramDistinguishesModeled

**The diagram distinguishes measured from modeled, and says so**

Covers: [product.measuredVersusModeled](../../../local/product/measurements.md#measuredversusmodeled), [product.segmentComposition](../../../local/product/diagram.md#segmentcomposition)

**Method:** geometry unit tests, a structural check
over the stylesheet, and assertions on both rendered pages

**Assertions, geometry:** `ChordTickX` equals the scaled half-span minus the scaled chord exactly;
it is zero when no chord is available or when the chord would reach the axis, so the renderer falls
back to a whole crossbar rather than drawing a tick outside the wing.

**Assertions, stroke style:** in the stylesheet, every modeled element — neck, body, and the
modeled crossbar — carries a `stroke-dasharray`; every measured one — bill, tail, and the plain
crossbar — does not; and the head is drawn with `fill: none`. The **tail must not be dashed**, which
is the assertion that catches a revert to the old meaning of dash.

**Assertions, the neck exception:** the neck's dash pattern differs from the body's. Stated as a
requirement of the check, not merely tolerated by it, because the two look like an inconsistency
and tidying them into agreement makes short necks render solid ([decisions.drawingCarriesProvenance](../../../decisions.md#drawingcarriesprovenance)).

**Assertions, rendering:** a species with a **measured** wingspan emits no dashed crossbar segment;
one with a **modeled** wingspan emits exactly one, flanked by two solid segments and marked by two
ticks; and the user guide (`/docs/using`) and the species page each carry text explaining what
solid, dashed, and the tick mean.

## foldedFigureExplained

**A folded-neck bird's detail page explains why its figure is shorter than its total length**

Covers: [product.labeledDetailDiagram](../../../local/product/measurements.md#labeleddetaildiagram)

**Method:** assertions on rendered detail pages

**Assertions:** a heron's detail page says its neck is drawn as an S and gives the species' own neck
length and total length in the reader's units; a straight-necked bird's page carries no such note.
(Until 2026-10-25 it also had to say the full length is what sorting, sections, and measurements
use.)

## segmentLabelsDontCollide

**Every metric the species page reports is named on its tile, keyed to its part by color**

Covers: [product.labeledDetailDiagram](../../../local/product/measurements.md#labeleddetaildiagram), [product.tileOpensDetail](../../../local/product/presentation.md#tileopensdetail)

**Method:** assertions on the rendered species page

**Assertions:** the species page draws one large tile, the home page's template; its measurement
lines and mass label name **every row of the measurements table**, total length as Length and
wing chord as Chord, with the table's value and a `~` exactly where the table's kind is not
`measured`, in both unit systems. Written against the table rather than a list of metric names, so
it maintains itself. Bill, Inner, Chord, Tail, and Mass are keyed to their parts' colors and
Wingspan is not ([decisions.innerSpanDerived](../../../decisions.md#innerspanderived)); the drawing
holds no text and no leader line; and the badge carries the family's common name
([decisions.textSizeFloor](../../../decisions.md#textsizefloor)). Until 2026-11-01 this criterion held the labels inside the drawing to
a layout, a paint order, and a column; those assertions went with the labels.

## tileValuesNameTheirSource

**Every value on a tile says where it came from in its tooltip, measured ones included**

Covers: [product.provenancePerMetric](../../../local/product/measurements.md#provenancepermetric)

**Method:** assertions on the rendered species page and on the tooltip function

**Assertions:** on the species page's tile, Length, Bill, Head, Neck, Body, Tail, Wingspan, Chord,
and Mass each carry a tooltip, and a measured one begins `measured, from` and its source; a
measurement filed under another name adds `filed as` and that name, and one filed under the
species' own name does not repeat it; a near-match names its donor and its source.

## innerSpanDerived

**The inner span is the wingspan less two chords, and never passes for a measurement**

Covers: [product.labeledDetailDiagram](../../../local/product/measurements.md#labeleddetaildiagram), [product.tileContents](../../../local/product/presentation.md#tilecontents)

**Method:** assertions on the rendered pages and on a tile built for a bird whose chords meet

**Assertions:** the species page's tile names an Inner, keyed emerald, whose value is the
wingspan less two chords in the page's unit, marked `~`, with a tooltip saying it is derived from
the wingspan and the chord; the measurements table has an Inner span row of kind `derived`; the
home page's tiles carry it the same way; and a bird whose chord is half its wingspan gets no inner
span, as its bar is drawn whole.

## totalLengthLabeled

**The species page states the whole bird's total length, not the drawing's height**

Covers: [product.labeledDetailDiagram](../../../local/product/measurements.md#labeleddetaildiagram)

**Method:** assertions on the rendered species page

**Assertions:** the species page's tile gives Length as the **total length from the measurements
table**, read out of the rendered table. The fixture is a **folded-neck species** (Ardeidae), whose
drawn height is 6–9% short of its total length, so a Length taken from the drawing would differ;
the test stops if the fixture ever stops folding its neck. Until 2026-11-01 the total length was a
label under the tail, held to a placement and a room in the box; those assertions went with the
label ([decisions.textSizeFloor](../../../decisions.md#textsizefloor)).

## massDrawnAsWaterCube

**Mass is drawn as the cube of water that weighs the same**

Covers: [product.labeledDetailDiagram](../../../local/product/measurements.md#labeleddetaildiagram)

**Method:** geometry assertions over every fixture
species, plus assertions on the rendered detail page and home page tiles

**Assertions, arithmetic:** the square's side is `10 × ∛(mass in grams) × px/mm`, to within a
twentieth of a pixel, for every species. Derived from the definition of the gram — one gram of
water is one cubic centimeter — and never read off a run. This is the criterion that makes the
square's claim checkable: it is a real volume at the diagram's own scale, not a glyph on a scale
of its own, and only the arithmetic can tell those apart.

**Assertions, placement:** the front face's **base and left edge are identical in every
diagram** — that fixed corner is the whole point of the cube form, since it gives the eye a
common origin and lets two cubes be compared directly; a base that drifted with the bird would
cost that without changing any number. The projection has **non-zero depth**, or it is the square
this replaced. It stays inside the drawing area above and never reaches the bird's axis to the
right. Its label, page text over the drawing since 2026-11-01, starts right of the cube and stands
on its foot, placed in percent of the drawing ([decisions.textSizeFloor](../../../decisions.md#textsizefloor)); before, it hung under the
face and was held clear of the left edge, and a strip under the drawing was reserved for it.

**Assertions, provenance:** solid when the mass is measured, dashed when it is not, as [product.measuredVersusModeled](../../../local/product/measurements.md#measuredversusmodeled)
requires of every other stroke.

**Assertions, rendering:** the square reaches the detail page, the diagram key appears **exactly
once** on it, and it **names the square**. The comparison page was checked too until it went
with the classic view on 2026-10-27. An unexplained shape beside a bird reads as part of the bird, so the key naming it is
not documentation of the feature, it is the feature.

**Assertions, the home page tiles:** **every** tile carries a cube — counted against the number of
tiles, not merely found once — and **no tile carries a label of any kind**, the mass being named
in the footer instead ([decisions.massCubeOnTiles](../../../decisions.md#masscubeontiles)). This reverses what this criterion asserted
when it was written, which was that tiles carry none; the reasoning is in [decisions.massCubeOnTiles](../../../decisions.md#masscubeontiles).

**Assertions, clearance:** across the real dataset the tile cube never reaches the crossbar, at
every scale and metric the index offers — 839 species × 6 combinations, asserted rather than
recorded. This was prose, **"0 overlaps in 841"**, and it rotted the moment lengths moved:
[decisions.trunkExponentCalibrated](../../../decisions.md#trunkexponentcalibrated) invalidated the figure and nothing failed, because nothing ran. It
was caught by reading.

What is asserted is a **clearance greater than zero**, taken from the requirement; the tightest
margin measured is logged rather than asserted, so a change that halves it is visible without
pinning today's number as a rule. A boolean overlap check proved a poor guard — the cube sits
bottom-left and the crossbar crosses the upper middle, so floating the cube to the top of the
tile, sliding it 120px right, tripling the projection depth, and scaling it by 1.6 all leave it
technically clear while moving the margin a long way.

**Assertions, the Presentation panel** (gone with the classic view on 2026-10-27, and with it
[product.pageExplainsItself](../../../local/product/presentation.md#pageexplainsitself))**:** it explains the cube in prose, naming what it is: the
bird's mass drawn as water. The diagram key alone is not enough — the Presentation panel
is where a reader learns to read the drawing ([product.pageExplainsItself](../../../local/product/presentation.md#pageexplainsitself)), and the cube is the one mark that is not a
length. (Until 2026-10-25 it also had to warn that a cube root flattens; that warning left the
page with the requirement's amendment.)

## faviconFromPalette

**The favicon is painted from the app's palette, not its own**

Covers: [product.oneDiagramVocabulary](../../../local/product/diagram.md#onediagramvocabulary)

**Method:** parse both assets out of the embedded FS and
compare the favicon's colors against the stylesheet's

**Assertions, no invented colors:** every color literal in `favicon.svg` is one the stylesheet
declares — as a diagram part's `stroke`/`fill`, or as a surface custom property. A shade that
exists only in the icon fails, which is the exact drift this criterion exists for.

**Assertions, the parts it does draw:** the favicon contains the **crossbar** color and the
**bill** color, both read from the stylesheet's `.wing` and `.bill` rules rather than written
here. Change either in the stylesheet alone and this goes red.

**Assertions, the comparison can still bite:** the stylesheet must yield a non-empty palette and
the favicon a non-empty set of colors, so a parser that silently matches nothing fails loudly
instead of passing.

## partColorsHoldApart

**The diagram's part colors are text colors, name their parts on the tile, and stay apart for color-blind readers**

Covers: [product.oneDiagramVocabulary](../../../local/product/diagram.md#onediagramvocabulary), [product.labeledDetailDiagram](../../../local/product/measurements.md#labeleddetaildiagram)

**Method:** a check over the stylesheet

**Assertions:** each measured part's stroke (bill, tail, crossbar, chord, cube) clears 7:1 on the
page and the panel; each name on the tile (`.k-bill`, `.k-tail`, `.k-wing`, `.k-chord`, `.k-mass`)
is set in exactly its part's stroke color; and every two parts, with the gray head, neck, and body,
differ by a CIEDE2000 ΔE of at least 10 under normal vision and under protanopia, deuteranopia, and
tritanopia simulated with Machado, Oliveira, and Fernandes's 2009 matrices at full severity
([decisions.partColorsByMeasurement](../../../decisions.md#partcolorsbymeasurement)), except that the mass cube is held apart only under normal vision
([decisions.massCubeSkyBlue](../../../decisions.md#masscubeskyblue)). The ΔE formula is pinned on a color against itself and black against
white.

## tileHeaderNeverWraps

**Nothing in a tile header can wrap to a second line**

Covers: [product.tileContents](../../../local/product/presentation.md#tilecontents)

**Method:** a structural check over the stylesheet, plus
assertions on the rendered markup — the pattern `familyListWraps` used before it was retired uses, for the same reason

**Assertions, the name line cannot wrap:** `.nameline` declares `flex-wrap: nowrap`, and the
common name sits in its own `.cnname` element carrying `white-space: nowrap` with
`text-overflow: ellipsis`. Both halves are needed: without the element the ellipsis would eat the
banding code at the end of `.cn`, which is the one part that must survive.

**Assertions, the code never truncates:** `.cn .code` declares `flex: none`, so it keeps its
width while everything beside it gives way. The line under a tile's rule, `.idline`, and the
measurement lines, `.ft`, declare `flex-wrap: nowrap`, since a wrapped measurement line would push
the drawing down.

**Assertions, the badge cannot wrap:** `.badge` declares `white-space: nowrap` with
`text-overflow: ellipsis` and a `max-width`, a bare `nowrap` being worse than wrapping — it would
overflow the tile instead.

**Assertions, markup:** the species page's name line and the tile's each contain a `.cnname`,
and the codes are siblings of it rather than inside it ([decisions.tileMeasurementLines](../../../decisions.md#tilemeasurementlines)). A code nested in the truncating element is
the defect this criterion exists to prevent, and it looks identical until a name is long enough.

## regionFilterBehaves

**The region filter selects, groups, and composes**

Covers: [product.regionAndMonthFilter](../../../local/product/filtering.md#regionandmonthfilter), [product.filtersCompose](../../../local/product/months.md#filterscompose)

**Method:** rendering tests over the home page's `?region=`

**Assertions, filtering:** filtering to one region shows species reported there and excludes
others; no parameter shows all; and an explicitly empty `region=` shows all too, the home page's
rule. Until 2026-10-27 it showed none, as the classic view's family filter did.

**Assertions, the selector, until the classic view went on 2026-10-27:** the box offers **both groupings** of [product.regionsGroupedTwoWays](../../../local/product/months.md#regionsgroupedtwoways),
each labeled and each carrying a note saying whether it is measured or drawn, with
**Geographic first**; within each,
regions are grouped into zones, **Pelagic is the last zone**, and every region is present, each
grouping being a partition. Each zone header toggles its whole group; **All** toggles
everything; and each chip carries the standard postal abbreviation with the full name beside it.

Checking each grouping **separately** is the point. Reading the zone toggles across the whole
box and checking only the final one passed on the hand-drawn grouping alone and left the derived
grouping's ordering unchecked — one of the whole-page assertions this file keeps having to
scope.

**Assertions, composition:** region and family narrow together ([product.filtersCompose](../../../local/product/months.md#filterscompose)), and
an empty intersection renders the empty state naming **both** filters — which is now reachable
through the UI, where with one filter it was not.

**Assertions, the sea:** a pelagic species appears under its ocean, and a seabird that also comes
near shore appears under both its ocean and the states it is seen from. Species with no region at
all are not silently swallowed.

## habitatGroupsFilter

**Habitat's groups are shortcuts over AVONET's habitats, a line each**

Covers: [product.habitatAndDietFilters](../../../local/product/months.md#habitatanddietfilters), [tech.stableUrls](../../../local/tech/delivery.md#stableurls)

**Method:** the groups checked against AVONET's vocabulary, and the home page rendered and requested

**Assertions:** the five groups hold each of AVONET's eleven habitats exactly once, with the labels
and order the requirement gives; the setting offers one line per group, the group's choice
leading it in a column of leads (so the rule after them and the habitats start at one place, and a
long line wraps under its habitats) and Towns and farms led by its one habitat alone; the group's
choice is marked only when all its habitats are, adds the missing ones (to any others chosen, or
alone from all), and removes them all when all are chosen; each habitat still toggles alone;
choosing Shore and sea draws exactly the birds with Coastal or a sea refinement of Marine, AVONET's or BirdView's (the
heron, Coastal by BirdView's addition, among them); the heading and the setting's value name a fully
chosen group by its label and other habitats by theirs; and an address naming AVONET's habitats is
served, not redirected.

## allAlwaysOffered

**Every setting that can be all offers All once, first, and marks it when chosen**

Covers: [product.nextView](../../../local/product/presentation.md#nextview)

**Method:** the home page rendered with every such setting all, and with every one narrowed

**Assertions:** Month, Region, County, Reporting frequency, Family, Habitat, and Diet each offer exactly one
All, as their first choice, at an address where each is all (Region aside, set to a state so County
shows) and at one where each is narrowed; All is marked exactly when its setting is all.

## countyListingApplied

**A county lists only what its state lists, with at least 5 records there**

Covers: [product.regionAndMonthFilter](../../../local/product/filtering.md#regionandmonthfilter)

**Method:** the county table's builder over rows shaped like the download

**Assertions:** a species its state lists, with at least 5 records in a county and one per million
of that county's records, is listed there; one with 4 records, one its state does not list, and one
under one per million of a busy county are written under the bar, so they still count; a name
BirdView does not draw counts in the county's total and nowhere else; the twelve months sum to the
total; the download's zipped table is read by its header; and a state's eBird withholding covers its
counties.

Joined and served: the ingest carries a county table's listed and under-the-bar cells, its totals,
and each county's name and state into the dataset without moving the peak share off a state
(`TestCountiesJoinedAsRegions`); and on the home page a county is a region by its GADM code, listing
only its own birds, named with its state's code, giving each bird the county's own share, and never
counted twice, in every region or beside its own state (`TestCountyIsARegion`). The County setting
appears only for one state with counties or counties of one state, a county chosen from the state
replaces it, more counties add, and removing the last or All returns to the state
(`TestCountySettingDrillsDown`). Without `-counties`, the default, a county's code resolves to
nothing, there is no County setting, and every figure equals the one without county data
(`TestCountiesOffByDefault`); and a county is recognized by its code whether or not it is
registered (`TestCountyParentFromCode`). A county's months follow it in parentheses on the share
line, and a list of counties is separated by semicolons while a list of states keeps its commas
(`TestPlaceAndMonthsAreNotJoinedByAComma`).

## everydayDefinitionsShown

**Every measurement says what it is before where it came from, in the requirement's own words**

Covers: [product.everydayDefinitions](../../../local/product/measurements.md#everydaydefinitions)

**Method:** assertions on the rendered species page and the guide, against the requirement's table

**Assertions:** the definitions are read from [product.everydayDefinitions](../../../local/product/measurements.md#everydaydefinitions)'s table itself, so a page cannot drift
from it; on the species page's tile, the tooltip of Length, Bill, Head, Neck, Body, Tail, Wingspan,
Inner, Chord, and Mass each begins with its definition and goes on to say where the value came
from; each row of the measurements table begins with its definition; and the guide carries every
definition word for word.

## regionUrlRoundTrips

**A region's URL is the data's code, unique, round-trips, and old values redirect**

Covers: [product.regionAndMonthFilter](../../../local/product/filtering.md#regionandmonthfilter), [tech.stableUrls](../../../local/tech/delivery.md#stableurls)

**Method:** unit tests over the region table, plus a
rendering and redirect test over the home page

**Assertions, the URL value:** a land region's is its GADM code (`USA.5_1`, `CAN.2_1`, `UMI.4_1`)
and a sea box's its PascalCase name (`GulfOfMexico`), and no internal `SEA.*` id resolves in a URL
([decisions.regionCodesFromData](../../../decisions.md#regioncodesfromdata)); values are **unique across all 74 regions** and URL-safe
without escaping. Uniqueness is the load-bearing one — two regions sharing a value makes one of
them unreachable, and the collision would be silent.

**Assertions, round-trip:** resolving a region's own value returns that same region, for every
region. A value that renders but does not resolve is a dead link on every choice.

**Assertions, retired values:** every postal code and every island name BirdView wrote before
2026-11-03 resolves as retired to its region, none is also a current value, and the sea names are
not retired; on the home page `/?region=CA&by=mass` redirects permanently to
`/?by=mass&region=USA.5_1`, several retired values in one address are all rewritten, and current
values answer 200.

**Assertions, no invented codes:** every non-empty postal code is two characters, and the ten regions
with no standard abbreviation carry an **empty** code rather than a made-up one
([decisions.regionUrlIsPostalCode](../../../decisions.md#regionurlispostalcode)). The codes now only label choices, but the rule still holds.

**Assertions, rendering:** region links carry GADM codes for land and names for the sea boxes,
and no retired value; a choice with a postal code is named with the code as its tooltip, and a
choice without one is named alone.

## familyGroupingComplete

**The family grouping covers every family, and stays off the tile**

Covers: [product.familiesByOrder](../../../local/product/months.md#familiesbyorder), [product.tileContents](../../../local/product/presentation.md#tilecontents)

**Method:** unit tests over the grouping against
the compiled dataset, plus a rendering test over the selector and a tile

**Assertions, coverage:** every family of the dataset sits in exactly one taxonomic order, and
every order holds at least one family. Checked **against the compiled dataset, not a fixture** —
the family list comes from AviList, which renamed 55 of 90 English family names in
[decisions.avilistReplacesAba](../../../decisions.md#avilistreplacesaba), so a family that is added or renamed must fail a test rather
than silently vanish from the control.

**The hand-drawn birder grouping this criterion also covered is gone**, retired by
[decisions.habitatDietSpeciesLevel](../../../decisions.md#habitatdietspecieslevel) in favour of the per-species filters of
[product.habitatAndDietFilters](../../../local/product/months.md#habitatanddietfilters). Its distinctness and ordering assertions went with it, and so did
[decisions.familyInTwoGroups](../../../decisions.md#familyintwogroups)'s cap on duplicate placements — there is no hand-drawn grouping
left for a family to be placed in twice.

**Assertions, the tile:** a tile carries **exactly one badge**. This is the assertion that
keeps [product.tileContents](../../../local/product/presentation.md#tilecontents) true when a third grouping arrives: without a count, "the other
groupings need no badge" is a sentence in a decision record rather than a property of the code.
It is a count, not a list of forbidden badges, because a list invites one more entry.

## creditsComplete

**Credits are complete and the docs are not stale**

Covers: [product.selfDescriptionStaysTrue](../../../local/product/presentation.md#selfdescriptionstaystrue)

**Method:** tests over the rendered pages and over
`README.md` and `docs/`, against the compiled dataset

**Assertions, credits:** **every page that shows a CC BY source's output credits that source on
that page.** Checked per page rather than site-wide, because that is how the obligation works
and because the gap this criterion was written for was exactly a per-page one — the index
credited AviList and `detail.html` and `compare.html` did not, while showing AviList names,
families, and species codes throughout.

**Assertions, counts:** every species or family count stated in `README.md` or `docs/` matches
the compiled dataset. `docs/data.md` said 841 species after [decisions.avilistReplacesAba](../../../decisions.md#avilistreplacesaba) cut
the list to 839.

**Assertions, attribution claims:** no document credits a source for something it no longer
supplies. `README.md` listed the ABA Checklist as the source of "species list, names, families"
after names and families moved to AviList.

**What this cannot check, stated so nobody assumes otherwise:** whether the prose is *true*. A
test can compare a number against the dataset and a credit against a page; it cannot read
`docs/using.md` and notice that the control it describes was redesigned. That remains a reading
task before release, and [product.selfDescriptionStaysTrue](../../../local/product/presentation.md#selfdescriptionstaystrue) requires it as a release gate rather than
pretending a test covers it.

**Counts are checked only where a claim is about the whole dataset** — "N species across M
families", and a bolded "**N species**". Every number before the word "species" is deliberately
not checked: "533 species without an empirical wingspan" is a subset, "442 of 841 species"
records a measurement taken when the dataset was that size, and Böhmer's "68 families" is
someone else's study. A check that flagged those would be switched off within a week.

## categoryOriginShown

**Every category on the detail page says where it came from**

Covers: [product.categoryProvenance](../../../local/product/measurements.md#categoryprovenance)

**Method:** rendering test over the detail view

**Assertions:** the detail page lists the bird's family, order, regions, habitat, and niche, and
any group it is in; **each carries a provenance of published, derived, or ours**; and a published
one names its source. A category rendered with no provenance fails, which is the whole point —
the failure mode is an opinion of ours sitting unlabeled beside a published fact.

**Assertions, regions:** the regions a bird is listed in are **one row**, not a row apiece.
They share a single rule, so sixty copies of one sentence would bury the habitat rows where
BirdView sometimes differs from AVONET. A bird in no region gets **no** Regions row rather than
an empty one — six species are at sea where no land polygon reaches ([product.regionAndMonthFilter](../../../local/product/filtering.md#regionandmonthfilter))
and must not be shown as placed nowhere by a rule.

**Assertions, links:** where provenance is an algorithm, the cell names the rule and links to a
document **served by BirdView**, not to an external host. Every `/docs/` link on the page is
fetched and must answer 200, so a fabricated citation cannot pass — **and where the link names
a section, that section must exist**. A link to the right document and a wrong anchor resolves
with a 200 and lands the reader nowhere near the explanation, which is how the regions row came
to cite a document that said nothing whatever about regions.

**Assertions, between documents:** every internal link in every served document resolves. The
guides cite the spec as `../river/requirements/local/product/`, which is right in a checkout and a 404 in a
browser, and the per-species link check could not see it because these are links between
documents rather than on a species page. **Relative links are resolved as a browser resolves
them**, against the page they sit on; until 2026-10-15 the check followed only absolute links,
so it could not see an unrewritten link at all. A link a guide makes into the repository —
a requirement, source code, a data table — is rewritten to the published repository, and the
check confirms the file it names exists in the checkout.

**Assertions, titles:** a document is titled and indexed by its **first top-level heading**,
not by its first line. The case that forced it was the requirements' `process.md`, which opened
with an HTML comment and was listed in the index as `<!--`; the requirements are no longer
served, and the rule stands for the guides.

**Assertions, what is not served:** nothing under `river/` or `third_party/` is served.
`river/requirements/imported/` and `third_party/` hold other people's terms and must not be reachable
([tech.fetchRestrictedInputs](../../../local/tech/delivery.md#fetchrestrictedinputs)). Since [decisions.requirementsNotServed](../../../decisions.md#requirementsnotserved) nothing under `river/` is embedded
at all; the probes remain, so that serving the requirements again cannot quietly bring the
vendored terms back with them.

## staticServesAssetsOnly

**`/static/` serves the files pages load, never a template**

Covers: [tech.servesOnlyItsPages](../../../local/tech/delivery.md#servesonlyitspages)

**Method:** request the bare directory and
every page template under `/static/` and expect 404; request the stylesheet and the favicon and
expect 200

## robotsKeepsCrawlersToPages

**`/robots.txt` opens every query-free page, and closes every filtered view**

Covers: [tech.crawlersSkipFilteredViews](../../../local/tech/delivery.md#crawlersskipfilteredviews), [tech.servesOnlyItsPages](../../../local/tech/delivery.md#servesonlyitspages)

**Method:** parse the served policy and judge
every page the accessibility crawl discovers by RFC 9309's matching rules

**Assertions:** `/robots.txt` answers 200 as `text/plain`; under `User-agent: *`, every discovered
page without a query string is allowed, and every one with a query string is disallowed, as is
the boundary file (`/compare` and `/now` had rules until 2026-10-27); `/`, a species page, the guides, and the
stylesheet are allowed. The crawl must offer at least three of each kind, or the check fails
rather than passing on nothing. An unknown field in the policy fails it too, so a typo cannot drop
a rule. `TestRobotsMatcher` pins the matcher to RFC 9309's precedence: the longest match wins,
Allow wins a tie, `*` and `$` are wildcards, and paths are case-sensitive.
**Does not cover:** whether any crawler obeys it, which no test can decide; the deployment's
instance cap is what bounds one that does not.

## stableUrlsKeepWorking

**Every stable path answers, and every stable parameter is still read**

Covers: [tech.stableUrls](../../../local/tech/delivery.md#stableurls)

**Method:** request each stable path; set each stable
parameter to a value the pages write and compare the home page's settings summary with the default's

**Assertions:** `/`, a species page, `/birds.json`, `/robots.txt`, `/docs` and each of its four
guides, `/added` and `/excluded` answer 200; and each of `by`, `rev`, `step`, `units`, `names`,
`pin`, `show`, `family`, `region`, `month`, `habitat` and `niche` changes the home page's title,
count line, tier headings, or pinned list, which together name every setting. The comparison,
`/now`, `scale` and `species` left this list on 2026-10-27 ([decisions.classicViewRemoved](../../../decisions.md#classicviewremoved)). The count line
is compared rather than the page, because every link on the page carries the query, so any
parameter at all would change the page.
**Does not cover:** that a URL shows the same birds it did, which moves with the data by design;
and older value forms, of which none has been retired yet. The first rename adds its old form
here.

## canonicalHostRedirects

**Given a canonical host, every other host is redirected to it**

Covers: [tech.oneCanonicalHost](../../../local/tech/delivery.md#onecanonicalhost)

**Method:** requests with set `Host` headers

**Assertions:** with no canonical host, `www.birdview.site`, `birdview.site` and `localhost:8080`
are each served; with `birdview.site` canonical, `www.birdview.site`, the Cloud Run address and
`localhost:8080` each answer 301 to `https://birdview.site` with the path and query unchanged,
and `birdview.site` is served however it is cased and with a port.

## openOceanFilter

**The filter offers the refinements, and addresses naming Marine redirect to them**

Covers: [product.habitatAndDietFilters](../../../local/product/months.md#habitatanddietfilters), [tech.stableUrls](../../../local/tech/delivery.md#stableurls), [product.categoryProvenance](../../../local/product/measurements.md#categoryprovenance)

**Method:** the home page and a species page rendered and requested

**Assertions:** `habitat=Marine` redirects permanently to `Marine-inshore` and `Marine-offshore`,
each once, other parameters kept; each refinement draws exactly the birds carrying it; the Shore and
sea line reads Shore and estuary, Sea near shore, Open ocean; the heading names both "Open sea" and
one alone by its label; a bird's page shows its refinement as derived, with its evidence and a link
to the guide.

## birdPresets

**Birds fills in Family, Habitat, and Diet for a kind of bird, and keeps every other setting**

Covers: [product.birdPresets](../../../local/product/presentation.md#birdpresets)

**Method:** the home page rendered on the fixture, and every choice's address on the built dataset

**Assertions:** Birds is the first setting and reads all by default; Shorebirds from an address with
a region, month, reporting frequency, order, family, and diet keeps the first four and sets exactly
its families and habitats, no diet, and no name for the choice; at that address Birds names it, the
heading says "Showing shorebirds" without the settings, choosing it again or All clears Family and
Habitat and keeps the rest, and an added habitat reads custom; Wading birds draws the heron and not
the stilt, Shorebirds the stilt and not the heron; every link from Birds changes Family, Habitat,
and Diet only. On the built dataset every family a choice names is drawn, every habitat is in the
vocabulary, and every choice draws birds (Shorebirds 64, Raptors 42, Wading birds 30 across every
region and month).

## chooseOneSettings

**Each choose-one setting draws one joined row and marks the chosen choice for screen readers**

Covers: [product.chooseOneSettings](../../../local/product/presentation.md#chooseonesettings)

**Method:** the home page rendered and each setting's popover read

**Assertions:** Birds, Order by, Reporting frequency, Units, and Names each draw every one of their
choices inside one joined row (`role="group"`), Birds' All among them, with no words saying to
choose one; in each, exactly the chosen link is marked `aria-current="true"`; Month, Region, Family,
Habitat, and Diet draw no joined row. The contrast and phone-width checks cover the new styling.

## actionsColoredApart

**Here and Now are colored apart as actions, and nothing else is**

Covers: [product.nextView](../../../local/product/presentation.md#nextview)

**Method:** the home page rendered; the contrast check covers the color

**Assertions:** Now (Month) is a link with the act class, Here (Region) a button with it, and no
other control carries it.

## massInOneUnit

**Mass is written in one unit across a comparison**

Covers: [product.oneUnitInComparison](../../../local/product/presentation.md#oneunitincomparison)

**Method:** unit tests over
`unitsFor`, plus rendering tests over the home page

**Assertions, the rule:** the smallest mass in a scope chooses the unit, so a scope spanning the
switch point is written in the smaller one throughout. A scope entirely above it keeps the large
unit — the rule is "the smallest decides", not "always ounces". A scope with no recorded mass
does not crash and does not pick the large unit on no evidence.

**Assertions, the scopes:** the home page draws every bird at one scale, so its scope is every
bird it draws. Until 2026-10-27 the classic view's scope followed its Scale setting, and the
comparison page's was the set named in its address; both went with the classic view.

**Assertions, no mixing:** over the real dataset, the home page never renders two different
mass units at once. This is the assertion that would have caught the original defect,
which was found in a screenshot rather than by any check.

## commonestShareIsShare

**The commonest-region share is a share, and its citation resolves**

Covers: [product.shareOfRecords](../../../local/product/presentation.md#shareofrecords), [product.categoryProvenance](../../../local/product/measurements.md#categoryprovenance)

**Method:** unit tests over the metric, rendering tests over the tile and detail page, and a
walk of every anchored citation the site emits

**Assertions, the metric:** the region chosen is the one where the species is the largest
**share** of that region's bird records, not the one holding the most of its records — the two
differ for 564 of 1,081 species. Both halves of the ratio reach the dataset, so the detail page
can show the evidence a thin share rests on.

**Assertions, the tile:** the share **follows the region filter** — the same bird reads
differently under `?region=AK` than across everything, and narrowing to where it is densest
raises its share. Region ids and their counts travel as parallel slices, so their lengths are
pinned, no count exceeds its region's total, and every region a species is placed in has a
total to divide by. The line renders on **every** tile, em dash and all, because a line that
appears on some tiles and not others makes those tiles shorter and takes the whole flex row out
of step. The footer stays one line ([product.tileContents](../../../local/product/presentation.md#tilecontents)) — a fourth cell there wrapped it,
which is why the share has its own line.

**Assertions, the citation:** every `#fragment` the site emits resolves to a heading in the
document it points at. This is the assertion that matters most, and it did not exist: the
link-walking check drops the fragment first — *"fragments are within the page"* — so it verified
`/docs/data` and never `#commonest`. Writing the row before the section left the suite green.

Walking it found **176 broken cross-references in the served spec**: every `(product.md#scopeisrecurringavifauna)`
in the bundle, which `river/check/refs.py` validates in the *markdown*, rendered as
`#scopeisrecurringavifauna--scope-is-the-aba-area-avifauna-that-recurs-by-birdviews-own-measure` and landed every
reader at the top of the document. Headings opening with a requirement id now anchor on the id.

## reportDataIssueLinks

**Each table on a species page links to a filled-in issue, and only with a tracker set**

Covers: [product.reportDataIssue](../../../local/product/presentation.md#reportdataissue), [product.noPersonalAddress](../../../local/product/presentation.md#nopersonaladdress)

**Method:** a species page rendered with and without a tracker, and its links parsed

**Assertions:** with a tracker set, the page has exactly two Report a data issue links, one under each
of the Measurements and Categories headings; each opens the tracker's new-issue page with a title
naming the bird's English and scientific names and the table, and a body holding the page's address
with its query, that table's rows with label, value, and provenance tag, and the dataset's source
versions and the server's revision; each link is under 8,000 bytes; the home page's Report it opens a
new issue with no title and a body holding the page's address with its query, its heading with the
settings in view, the source versions, and the revision; with no tracker set, neither page has a
link. Across every page (the home page, species pages, the three kinds of page for a bird no longer
drawn, `/added`, `/excluded`, and every guide), every link to the tracker but a numbered issue or the
building guide's Contact opens a new issue whose body holds the page's address, the source versions,
and the revision; the home page, a species page, and each no-longer-drawn page have one, and the
Passenger Pigeon's names the bird and its verdict. A tracker setting that is not a GitHub issue list
over HTTPS is refused.

## noContactAddressPublished

**No served page publishes a contact address**

Covers: [product.noPersonalAddress](../../../local/product/presentation.md#nopersonaladdress)

**Method:** rendering test over every
served page, including every document the two indexes list

**Assertions:** no page matches a plain address, the spelled-out obfuscation, or a mail link.
The feedback link appears only when the tracker URL is configured, and an unset one renders
nothing rather than an empty href.

**Why it matches the obfuscation and not just the at-sign:** the address that shipped was
spelled out in words and contains no at-sign at all. A check for one would have passed over the
exact string it exists to catch.

**It is deliberately not limited to the owner's address.** Run the first time it failed twice.
Once on a source owner's mailbox in the served manifest, republished by BirdView because the
provenance table recorded it — a provenance record names who owns a source without handing out
their mail. And once on this bundle: the requirement and this criterion both quoted the address
as an example, and the spec is served, so they republished on `/spec/tech` and
`/spec/acceptance` precisely what the rule forbids. Neither quotes it now. Exempting the spec
was the wrong fix, because the spec is a page readers reach. Since [decisions.requirementsNotServed](../../../decisions.md#requirementsnotserved)
it is no longer served, but it is still published with the repository, so the check reads
`river/*.md` from the checkout instead — the same files, with no loss of coverage.

## outOfScopeExplained

**A species left out of scope is named and explained, at 404**

Covers: [product.undrawnSpeciesAnswer](../../../local/product/presentation.md#undrawnspeciesanswer)

**Method:** handler test over the real dataset, plus an ingest test over the real status table

**Assertions:**

1. `GET /species/<id>` for an out-of-scope species returns **404** and a body naming the bird,
   its scientific name, and its status. Not 200 and not 410.
2. The body carries the evidence the status rests on — episodes, years, and records for
   `vagrant` and `historical` — and for `historical` the **date last recorded**.
3. For `absent` the body carries **no counts at all**, because none were measured.
4. An id matching nothing gets the bare `404 page not found`, with no bird named.
5. The ingest carries every classified-but-not-drawn species into the dataset, and **no id
   collides with a drawn species'**. A collision would hide a living page behind an epitaph.

**Why it is checked over the real dataset and not a fixture:** assertion 5 is a statement about
1,411 real ids produced by one AviList join, and a fixture asserts it over whichever handful the
fixture holds. The fixture tests cover the rendering; the dataset tests cover the claim.

**Three tests, because the assertions sit at different levels.**
`TestUndrawnSpeciesNamesItselfAt404` renders all three verdicts and the typo case against the
fixture; `TestNoIDIsBothDrawnAndNot` and `TestAbsentSpeciesCarryNoEvidence` run over the
compiled dataset. The second of those asserts **both halves** — absent species carry no
evidence, and every other verdict does — because a check written one way round is satisfied by
an ingest that publishes no evidence at all.

## commonestIsLargestShare

**The commonest region is the largest share, and the seas are out of the ratio**

Covers: [product.shareOfRecords](../../../local/product/presentation.md#shareofrecords)

**Method:** dataset test over all 1,083 species, plus rendering tests for the wording

**Assertions:**

1. For every species carrying a `commonest`, **no land region it occupies has a larger share**.
   This is [product.shareOfRecords](../../../local/product/presentation.md#shareofrecords)'s central claim and nothing asserted it before.
2. `RegionTotals` carries **no marine entry**, so nothing can divide by one.
3. A species' share is computed over regions the totals have an entry for, on **both** sides:
   no region contributes to the numerator without contributing to the denominator.
4. The detail page names the row **Commonest land region**, not "Commonest in".
5. A selection whose regions have no record total reads `No data for these regions yet` —
   never anything of the form *no records*, which would state a gap in BirdView's data as a
   fact about the bird.
7. **No page of the interface, and no guide under `/docs`, uses the word "share"** as the
   name of the metric. The normative documents under `river/` are exempt; they define it, and
they are not served.
6. A species with no records in the regions on screen reads **`0 in <denominator>`** in the
   same sentence every other tile uses — never an em-dash. The denominator is stated, because
   zero in a billion records and zero in nine hundred are different findings.

**Why assertion 1 is over the real dataset.** The server fixture hardcodes `commonestOf` and
every test renders it back, so the claim has only ever been checked against a value the test
itself supplied. That is how [decisions.seaShareDiffers](../../../decisions.md#seasharediffers)(a) shipped: a value the fixture
provides is a value no fixture test can falsify.

**Why assertion 3 is stated as an invariant and not as a region list.** The numerator and the
denominator used to be two separate loops with two separate conditions — `sel[id]` over the
species' regions, `all || sel[id]` over the totals — which is exactly how they came to range
over different sets. Deriving both from "does this region have a denominator" makes the
mismatch unrepresentable rather than merely absent today.

## peakAndShareOneCalculation

**The peak region and the index share are one calculation**

Covers: [product.shareOfRecords](../../../local/product/presentation.md#shareofrecords)

**Method:** dataset test over every placed species, plus a client test for the measurement

**Assertions:**

1. For every species carrying a peak, **selecting that one region on the home page reproduces the
   detail page's figure exactly.**
2. The peak is always **a region the species is listed in**. The old figure could name one the
   reader would not find in the filter at all.
3. At least one species peaks at sea, or the case [decisions.honestSeaDenominator](../../../decisions.md#honestseadenominator) exists for is unexercised. (45 do.)
4. The row is labeled **Top region**, and none of its superseded names survives — not
   `Commonest in`, not `Commonest land region`. Rejecting the old forms as well as requiring
   the new one is what would have caught `Commonest land region` outliving
   [decisions.honestSeaDenominator](../../../decisions.md#honestseadenominator), which gave the seas the denominator that qualifier existed
   for.
4. Every region has a record total, seas included.
5. A sea's total **exceeds the sum of the species placed there**. A denominator built out of
   its own numerators is the shape of the [decisions.seaShareDiffers](../../../decisions.md#seasharediffers) bug, and the only part of it a test can see
   without the network.
6. `CountOffLandInBox` subtracts the land, and **refuses** a facet whose rows sum past the box
   rather than returning a negative count.

**Why assertion 1 is the one that matters.** The bug the owner reported was not in either
page's logic — it was that nothing compared them. The peak came from `data/status.tsv` and the
share from `data/region-presence.tsv`, two tables with different denominators and different
membership rules, and each was tested against its own source. A check can only catch that by
asserting the pairing.

**Why assertion 5 names a margin and not a number.** A sea's total runs 8× to 200× the sum of
what BirdView places there, the tightest being the North Pacific at 11.9%. Land is exempt:
there the drawn species really are almost all the records — 98–100%, and Navassa is 6 of 6 —
so no margin exists to assert.

## responsesCompress

**Responses compress, and say that they did**

Covers: [tech.compressedResponses](../../../local/tech/delivery.md#compressedresponses)

**Method:** handler test over a page,
a detail view, a static file, and a served document

**Assertions:**

1. With `Accept-Encoding: gzip`, the response carries `Content-Encoding: gzip` and the body is
   a valid gzip stream that decompresses **byte-for-byte** to the uncompressed response.
2. Without it, no `Content-Encoding` and the plain body.
3. `Vary: Accept-Encoding` on **both**.
4. No `Content-Length` describing the uncompressed body survives.
5. A [product.undrawnSpeciesAnswer](../../../local/product/presentation.md#undrawnspeciesanswer) 404 keeps its status **and** its `Content-Encoding`.
6. Each page saves at least 50%, with the real figure logged rather than asserted.

**The header is asserted before the body, and that order is the criterion.** Reading the body
with `gzip.NewReader` does not catch an unlabeled response — the stream is perfectly valid,
it is simply unannounced — so a check that looks at the body first passes over the exact bug
this was written during. It did: the first version of the middleware set `Content-Encoding`
from `Write`, which is too late for any handler that calls `WriteHeader` itself, and the
static files and all 328 undrawn-species pages went out as compressed bodies labeled as HTML.
Confirmed live before the fix: `HTTP/1.1 404`, no `Content-Encoding`, `file` reporting
`gzip compressed data`.

**Why 50% and not a measured figure.** What must not regress is that compression happens;
pinning 94% would fail the first time a template gained a paragraph. The measured savings are
logged on every run — 94% for the index, 70% for a detail page, 67% for the stylesheet, 62%
for a served document.

## frequencyTiers

**Frequency orders commonest first, in tiers of powers of ten**

Covers: [product.orderByFrequency](../../../local/product/presentation.md#orderbyfrequency)

**Method:** rendering test over the real dataset, reading the section headers

**Assertions:**

1. `?by=frequency` produces about nine sections, ascending, every edge a power of ten.
2. **Each tier holds exactly the species whose frequency falls in the band its header claims**,
   recomputed independently in the test.
3. The birds with no records in the selection are kept in **exactly one** final section.
4. `step=100` is strictly coarser than `step=10`.
5. **Every tier edge under a region filter also appears unfiltered** — the property the whole
   design exists for.
6. No tier header contains `common`, `scarce`, `rare`, `abundant`, `uncommon` or `everyday`.
7. Scale defaults to **global** under Frequency and **per-section** under the size metrics
   ([decisions.frequencyGlobalScale](../../../decisions.md#frequencyglobalscale)), and **both** values remain selectable under both.
8. The Order by row **aligns on the text baseline**, so the label, the size buttons, and the
   Frequency button sit on one line with the captions hanging below. Both the declaration and
   the class that carries it are asserted: either alone matches nothing.
9. The Order by row **states its direction for every metric** — *smallest first* under the
   three size metrics, *most common first* under Frequency, captioned under each group. Both
   stand on every view: a caption that followed the selection would report what the reader had
   just chosen, where two standing captions say the controls differ before they are clicked.

Assertions 7 to 9 went with the classic view on 2026-10-27, with its Scale setting and its
Order by row.

**Assertion 2 is the one with teeth, and it was added after the first version passed a
mutation it should have caught.** The header is written by `frequencyBand` and the banding done
by `SectionsBy`, so a tier can be labeled `1 in 1,000 – 1 in 9,999` while holding whatever the
sectioning rule actually put in it. Replacing the decade rule with a proportional one left
every label a tidy power of ten; only the tier *count* gave it away, and only because
`step=100` then stopped changing anything. Checking membership turns that into five explicit
failures — `tier "1 in 10 – 1 in 99" holds 102 species; 23 have a frequency in that band`.

## sharedCountDisclosed

**A count covering two species says so, on the tile and on the page**

Covers: [product.namesRecordsFiledUnder](../../../local/product/presentation.md#namesrecordsfiledunder)

**Method:** rendering test over the real dataset, both members of every declared shared name

**Assertions:**

1. For **both** species a shared name covers, the detail page carries a *Recorded as* row
   naming the source binomial and the other species.
2. The **home page tile** carries the marker too, naming the other bird.
3. A species eBird **withholds** says so, and is **not** reported as a naming disagreement.
4. An ordinary species carries none of it.

**Why the index and not just the detail page.** The bug that prompted this was visible as a
ranking — a Caribbean mangrove resident sitting 51st most-recorded in North America — and a
note on a page nobody opened would not have caught it. The marker goes where the misreading
happens.

**Why assertion 3 is separate.** Withholding and renaming are indistinguishable in the data:
eBird returns zero for both. They are opposite in meaning, and conflating them would tell a
reader that eBird calls the Spotted Owl something else, sending them after a name that does
not exist. Nine species are in that state.

## figureSaysWhatItCounts

**The figure says what it counts, and it is not individual birds**

Covers: [product.shareOfRecords](../../../local/product/presentation.md#shareofrecords)

**Method:** unit test over the
rendered tile cell and the tooltip beside it

**Assertions:**

1. A home page tile's `1 in N` is followed, **in the same cell**, by the unit `species
   observations`.
2. The tooltip that spells the figure out counts in the same noun, and no longer in `bird
   observations`.

**The noun was wrong in the one way a reader would not catch.** The denominator counts species
on checklists and deliberately ignores how many individuals each entry saw — twelve
White-throated Sparrows is one observation, and so is a flock of fifty thousand Snow Geese.
"bird observations" names precisely the quantity the figure is *not*, so someone reading
`1 in 32` as one bird in thirty-two has misread nothing except the unit, and the unit is the
only thing on the tile that could have told them.

**The cell, not the page.** `species observations` also appears in the help text and in
[`docs/data.md`](../../../../../docs/data.md), so a whole-page search would pass with the figure absent
from every tile. The assertion matches the number and its unit as one pattern inside the
robin's own tile — the vacuous-check shape this project has now paid for about a dozen times.

## detailSaysWhatTileSays

**The detail page says everything the tile says**

Covers: [product.tileOpensDetail](../../../local/product/presentation.md#tileopensdetail), [product.shareOfRecords](../../../local/product/presentation.md#shareofrecords)

**Method:** property test comparing rendered tiles against rendered detail pages, under two
region scopes

**Assertions:**

1. For a sample of species across the home page, the figure printed on the tile appears verbatim
   on that species' detail page.
2. It appears **under the scope the reader selected** — `(all regions)` unfiltered, `in
   Alaska` under `?region=AK` — not quietly swapped for another.
3. The same bird's figure **differs** between the two scopes, or both assertions above could
   be satisfied by a pair of pages that ignore the filter entirely.
4. The home page parse is floored per scope — 1,000 tiles unfiltered, 300 in Alaska — so a regex
   that silently matched nothing cannot make the rest vacuous.

**The gap this closes.** [product.tileOpensDetail](../../../local/product/presentation.md#tileopensdetail) described the detail view as a larger
version of the tile "plus" three additions, and the *plus* was read as permission to add to
the detail page rather than as a floor under it. So when [product.shareOfRecords](../../../local/product/presentation.md#shareofrecords)'s figure
went onto every tile it never reached the detail page: a reader saw `1 in 23,393 species
observations (all regions)` on the Northern Yellow Warbler's tile, opened the bird, and found
only `Top region — 1 in 2,986 there`. A different number, in the same unit, for a scope they
had not asked about.

**The scope half is the sharper of the two.** A figure that is absent is visibly absent. A
figure that silently changes what it ranges over between one page and the next is a wrong
answer with nothing on either page to reveal the substitution — which is why assertion 2
exists and why the row is built by `shareWithSeams`, the same function the tile calls, rather
than recomputed. Two call sites deriving one number from the same inputs is exactly how the
index and the detail page came to disagree about where a bird is commonest
([decisions.honestSeaDenominator](../../../decisions.md#honestseadenominator)).

## disputedNamesShown

**Every name the occurrence data disputes is named on the page**

Covers: [product.namesRecordsFiledUnder](../../../local/product/presentation.md#namesrecordsfiledunder)

**Method:** property test over the
five naming tables and the rendered detail pages

**Assertions:**

1. Every drawn species in `AviListRenames`, `CountedAs`, `SharedNames`, `DomesticForms` or
   `WithheldByEBird` carries a `Recorded as` row.
2. The row names the specific thing: the alias for a rename, the published binomial for a
   lump, the covering name for a shared name, `domestic and feral` for a form,
   `withheld by eBird` for a suppression.
3. **No species outside those tables carries the row.** 26 of 1,083 are in the tables — 27
   until [decisions.countEbirdFirst](../../../decisions.md#countebirdfirst) derived `WithheldByEBird` from eBird's list and Desertas
   Petrel, on no entry of it, left — plus 24 that eBird withholds only in some regions or
   seasons ([withheldMonthsExplained](acceptance.md#withheldmonthsexplained)). The other 1,033 must stay silent.
4. The expected set is floored at 25 species, so emptying the tables cannot make the rest
   vacuous.

**Two of the four kinds were never implemented**, though [product.namesRecordsFiledUnder](../../../local/product/presentation.md#namesrecordsfiledunder) has named
the alias case since it was written — *"the page names the alias so a reader can find the bird
in the source."* A visitor who knows the Hairy Woodpecker as *Dryobates villosus* and finds a
page headed *Leuconotopicus villosus* could not tell a taxonomic update from a mistake, nor go
back to eBird and check. That is the reader this requirement exists for.

**Assertion 3 is what makes the other three mean something.** A disclosure satisfied by
printing it everywhere says nothing and trains a reader to skip it, and a one-directional
check would accept exactly that.

**The all-rows rule is not reachable by test today.** No species is in two of the five tables,
so a regression to the previous `if/else-if` chain would still pass. It is recorded in
[product.namesRecordsFiledUnder](../../../local/product/presentation.md#namesrecordsfiledunder) as the required shape rather than asserted here, because the
alternative — inventing an overlap in the real tables to exercise it — would put a false entry
in a published taxonomy to satisfy a test.

## honestRevalidation

**Every response can be revalidated, and the validator is honest**

Covers: [tech.cacheLifetimeStated](../../../local/tech/delivery.md#cachelifetimestated)

**Method:** unit test over rendered
responses, their headers, and their conditional round trip

**Assertions:**

1. Every response — pages and static files alike — carries an `ETag` and
   `Cache-Control: no-cache`.
2. Two different URLs never share an `ETag`.
3. Re-requesting with the returned `ETag` yields **304 with an empty body**, so the validator
   saves the thing it costs a header to offer.
4. A page renders **byte-identically** three times running, which is the assumption the
   fingerprint rests on.
5. Changing the dataset changes the home page's `ETag`.

**Assertions 4 and 5 are the ones that matter**, because the ETag is computed from the build
and the URL rather than from the body. That substitute is what makes the scheme affordable —
hashing the body means buffering 3.3 MB of index per concurrent request — and it is sound
only if rendering is deterministic. A clock anywhere in the render path would make every
cached page wrong in a way no other check here would notice, so determinism is asserted and
not assumed. Assertion 5 closes the other end: without it a constant would satisfy the rest.

**What this fixes.** BirdView sent no `Cache-Control`, no `ETag` and no `Last-Modified` on
anything, which does not mean "do not cache" — it means the browser may apply its own
heuristic and has nothing to revalidate against when it does. It surfaced as a review round
on `/species/yelwar1`, reported as rendering old state when the server was serving the new
one correctly.

## nameDisclosureLegible

**The name disclosure is legible and its link lands on the explanation**

Covers: [product.namesRecordsFiledUnder](../../../local/product/presentation.md#namesrecordsfiledunder)

**Method:** unit test over every
renamed species' rendered row

**Assertions:**

1. Every `Recorded as` row links to `/docs/data#when-the-sources-disagree-about-a-name`, and
   that anchor exists in the rendered page.
2. The row gives eBird's name **in full** — the actionable half, since it is what a reader
   types back into eBird.
3. Where the genus is unchanged the text says **epithet** and must not say *moved*.
4. Where the genus changed the text says **genus**.
5. The row names eBird.

**Both halves were wrong and both were reported.** Every row linked to *how common a bird is*,
the section about the frequency figure, while its own label read "how names are joined". And
the text said *"AviList moved this species to Pica nuttallii and eBird has not followed, so
BirdView looks the records up under Pica nuttalli"* — which for a one-letter difference looks
like the same name printed twice, and says *moved* about a species whose genus did not change.
Five of the seven renames are genus moves; two are corrected epithets, and one sentence was
being used for both.

**Why the row says eBird and not "the occurrence data".** Naming the sources generically would
be wrong in both directions, and the committed counts say so: iNaturalist follows AviList for
all five genus moves — 53,705 Hairy Woodpecker records under *Leuconotopicus villosus* — and
for the two spelling corrections **neither** database uses AviList's form. eBird is the single
claim true of all seven, holding nothing under the AviList name, which is why the join needs
the old one at all.

**Assertion 2 is narrower than it first was.** It began as "both binomials appear in the row",
which failed on the genus-move rows: those name the genus and the epithet separately and never
the full AviList string, because the full string is the page heading two inches above. Writing
the stricter rule would have forced repetition to satisfy a check rather than a reader.

**A boundary disagreement names both databases, and only calls itself a lump when it is one.**
`CountedAs` holds three entries and they are not the same taxonomic event: AviList **split**
the Whimbrel and the white terns and eBird has not adopted either split, while AviList
**declines to split** the Siberian Stonechat where both databases do. One sentence covering all
three named the wrong event for two of them, and asserted that AviList "does not carry"
*Numenius phaeopus* — which it does, as the Eurasian bird. The reason is therefore carried per
entry, and the criterion asserts each row names eBird *and* iNaturalist, says `lump` only for
*Saxicola maurus*, and says `does not carry` only where that is true.

## monthsFilterAndAbstain

**Months filter the list, scope the figure, and abstain where they cannot tell**

Covers: [product.regionAndMonthFilter](../../../local/product/filtering.md#regionandmonthfilter), [product.shareOfRecords](../../../local/product/presentation.md#shareofrecords)

**Assertions:**

1. `month.Present` recovers the known season of seven species measured in New York —
   resident, partial migrant, winter visitor, long-distance migrant, breeder, passage
   shorebird — month for month.
2. It **abstains** below 200 records, and abstaining is distinguishable from twelve trues: the
   first says nothing is known, the second says the bird is there all year.
3. A month nobody birded in reads **present**, not absent: the missing records are everyone's.
4. All twelve months selected reproduce the annual figure **exactly**, for every species and
   for the denominator every tile divides by. `birdview-regions` refuses to write a table
   where the twelve do not reconstruct the year.
5. Months narrow the **list**, not only the figure: June shows fewer species than the year and
   a different set from January, while all twelve shows exactly what the year shows.
6. The scope label collapses runs and **wraps the year end** — a winter selection reads
   `Nov-Feb`, not `Jan, Feb, Nov, Dec`.

**Assertion 1 is the calibration, not a spot check.** The rule is effort-corrected because a
raw share would call every bird a May bird: an American Robin takes 22.1% of its New York
records in May and 2.6% in December, and is present in both. The seven profiles are measured
counts, so a change to the measure shows up as a wrong month for a named bird rather than as a
number moving.

**Assertion 3 is the half that is easy to get backwards.** A region with no birding in a month
says nothing about any species in it, so calling the bird absent would report an absence of
measurement as a measurement of absence — the distinction [product.undrawnSpeciesAnswer](../../../local/product/presentation.md#undrawnspeciesanswer) exists for.

**Assertion 6 earns its own function.** Winter is the selection a birder is most likely to
make and the one a renderer that walks January to December mangles worst.

## withheldMonthsExplained

**A month eBird withholds is shown and explained, not read as absence**

Covers: [product.regionAndMonthFilter](../../../local/product/filtering.md#regionandmonthfilter), [product.namesRecordsFiledUnder](../../../local/product/presentation.md#namesrecordsfiledunder)

**Assertions:**

1. `taxonomy.EBirdSensitive` is well formed: every window parses, every region is a GADM id or a
   country, and every species in `WithheldByEBird` is on the list — except Desertas Petrel,
   which is not, and whose missing records are a naming question.
2. A window marks only the months it covers **entirely**, and may wrap the year end. Painted
   Bunting's Florida window ends April 15, and April — which still holds 12,483 eBird records —
   is not marked.
3. **A listing is acted on only where the records confirm it.** Painted Bunting in Florida and
   Long-eared Owl in New York are confirmed; Greater Sage-Grouse in Utah, whose records run
   higher inside its window than outside, and Spruce Grouse in Vermont, with 1,303 eBird records,
   are not.
4. The table's withheld flags round-trip, and malformed flags are refused rather than read as
   "nothing withheld".
5. Under `?region=FL&month=Jan` Painted Bunting is **shown** and its tile says *Withheld by eBird
   in Florida (Jan)*; under `?region=NY&month=Jun` the same holds for Short-eared Owl. Without a
   month filter the Florida caveat reads *(Oct-Mar)*, wrapping the year end.
6. Greater Sage-Grouse in Wyoming in March carries **no** caveat: its records are in the export.
7. **Controls:** Bobolink in New York in January and Snow Bunting in New York in June stay
   hidden. They are not withheld, just gone.
8. The species page lists the withheld regions and months ([product.tileOpensDetail](../../../local/product/presentation.md#tileopensdetail)).

**Assertion 7 is what keeps the fix honest.** Abstaining is easy to get right by abstaining
everywhere, which would quietly switch the month filter off; the controls are the birds it must
still hide.

**What the build found.** 202 region cells across 34 species are marked withheld. No count
changed: the table was rebuilt from the same cached facets as the one it replaces, and every
cell's twelve monthly figures are identical. The flags are new information, not a new
measurement.

## pipelineDiagramServed

**The pipeline diagram is served, current, self-contained, and linked**

Covers: [product.pipelineDiagram](../../../local/product/measurements.md#pipelinediagram), [product.categoryProvenance](../../../local/product/measurements.md#categoryprovenance), [tech.selfContainedServer](../../../local/tech/delivery.md#selfcontainedserver)

**Level:** code · **Method:** unit tests over the docs embed and the served pages

**Assertions:**

1. **Current.** Every page in `docs/` with a Mermaid block has an SVG of the same name, whose
   first line records the SHA-256 of that block.
2. **Self-contained.** The SVG carries no `<script>`, no `<foreignObject>`, no event handler and
   no URL other than XML namespace declarations.
3. **Served.** `/docs/pipeline` shows the picture, linked to itself full size, in place of the
   Mermaid source; `/docs/pipeline.svg` answers as `image/svg+xml`; the index lists the page and
   not the picture.
4. **Faithful to the readers.** Every column the diagram names for AviList, AVONET, and the Fox
   wingspans is a column the reader code asks for, so a renamed or dropped column cannot leave
   the diagram describing a pipeline that no longer exists.
5. **Linked** from `docs/data.md`, `docs/building.md` and `docs/using.md` as served, from the
   home page's footer, and from the README.
6. **One field per line, one mark per field.** Every box with fields lists one field per line,
   followed by exactly one of ★ · ✗ and an optional note in italics; the lines under a box's name
   are italic. A box with fields written as a plain label, which cannot carry italics, fails.
7. **Tables read left-aligned.** Every box with fields, and the legend's key to the marks, is
   left-aligned in the rendered SVG; steps and commands stay centered. Mermaid
   centers every label, so the renderer moves them, and a render that skipped it would still
   match its source's hash.
8. **Every taxonomy table the ingest reads is in the diagram.** Each `taxonomy.X` that
   `cmd/birdview-ingest` refers to is named in `docs/pipeline.md`, except a listed few that are
   not data: the inverse of a table the diagram already names, and a list printed in the build
   report. A table added to the ingest without a line in the diagram fails.

**The column check found two mislabeled fields before it existed.** Writing the diagram from the
readers turned up AVONET's `Family1`–`3` and `Secondary1`, both read and neither used — `FamilyOf`
was kept to cross-check ABA's families, and ABA is gone — and a first draft that called
`Trophic.Level` a fallback for a missing niche, when it adds a second diet where AVONET's level says
Omnivore and its niche does not. The check itself first failed on its own diagram, on a field label
wrapped onto a second line; one field per line is now the rule it relies on.

**Assertion 6 came from review.** The first diagram used · both as a mark and as a separator
between fields on one line, so "species · region · date window" read as one field with two marks,
and the legend's "one field per line" held for some boxes and not others. Rewriting every box to
the rule also found what the diagram had left out: the server reads `internal/taxonomy`'s tables
directly for *Recorded as* and the withheld notes, three generators read the previous build's
`species.json` for their species list, and `birdview-gbifmap` reads AviList.

## sharedMeasurementsDisclosed

**Species drawn from the same measurements say so**

Covers: [product.borrowedMeasurementsSaySo](../../../local/product/measurements.md#borrowedmeasurementssayso)

**Method:** unit test over ingest
output, and rendering tests over the home page and the detail page

**Assertions:**

1. **The groups are derived.** Grouping the built dataset by the bill's `source` and
   `sourceTaxon` gives exactly the pairs the page notes, and includes Vega Gull with American
   Herring Gull, both at `AVONET1_BirdLife` *Larus smithsonianus*, with the same bill, tail, mass,
   and wingspan.
2. **Every tile in a group says so.** The tile of each member carries *Same measurements as* and
   the English name of every other member, inside that tile's own markup. A tile outside every
   group carries no such line.
3. **Every detail page in a group says so,** naming the other members, the source sheet, and the
   source taxon.
4. **No hand table.** A species given a near-match to a drawn relative in a test dataset gets
   the note on both tiles without any other change.

## belowBarCountsKept

**Records below the presence bar are counted**

Covers: [product.regionAndMonthFilter](../../../local/product/filtering.md#regionandmonthfilter), [product.shareOfRecords](../../../local/product/presentation.md#shareofrecords)

**Method:** unit tests on the table's reader and writer, and over the built dataset and the
rendered index

**Assertions:**

1. **The table keeps them.** A cell marked below the bar round-trips through the writer and
   `region.LoadPresence` with its counts, its months, and its mark, and an unmarked cell reads as
   present.
2. **The ingest separates them.** A below-bar cell never appears in `regions`, so region and
   month filtering are unchanged, and its counts appear in the species' below-bar fields.
3. **The figure counts them.** With every region selected, a species' figure is its records in
   every region it has any, present or not. Emu's is not zero, and no drawn species with records
   reads zero.

## escapedBirdsListedAndServed

**Escaped and kept birds are off the home page, listed, and still served**

Covers: [product.escapedBirdsListed](../../../local/product/measurements.md#escapedbirdslisted), [product.statusIsPublished](../../../local/product/dataset.md#statusispublished)

**Method:** unit test over ingest output, and rendering tests over the home page, `/excluded` and
the detail pages

**Assertions:**

1. **Flagged from the list.** Every species in `status.Exotic` is drawn and carries its reason as
   `exotic`; no other species carries one. A listed name that is not drawn fails the build.
2. **Off the home page.** No flagged species has a tile on the home page, and none is offered or
   counted by the family, habitat, diet, or region settings; the page reports *All N birds*
   without them. A family whose every species is flagged is
   not offered.
3. **Listed.** `/excluded` names every flagged species with its reason and links to its detail
   page, and the home page links to `/excluded`.
4. **Still served.** Each flagged species' detail page answers 200, says it is not on the home page
   with its reason, and links to `/excluded`.

## readersToldStickFigure

**Readers are told the drawing is a stick figure**

Covers: [product.segmentComposition](../../../local/product/diagram.md#segmentcomposition)

**Method:** rendering test over the visible text of the home page, a detail page, and the using
guide

**Assertions:**

1. No visible text on the home page or a detail page says *diagram* or *diagrams*, except the link to the data-pipeline diagram. Class names, template comments and
   scripts are not visible text and are not checked.
2. The home page and `docs/using.md` call the drawing a *stick figure*.

**The check found one the edit had missed:** the detail page's provenance for total length read
*sum of the five diagram segments*, a model string written by the ingest rather than the
templates. It now reads *sum of the five drawn segments*.
