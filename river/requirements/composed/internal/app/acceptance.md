package app

# app — acceptance criteria

Composed with [app.md](app.md); do not edit, recompose.

How each requirement in [app.md](app.md) is decided. Each criterion names what it covers in a
`Covers:` field; nothing lists criteria per requirement.

**Uncovered:** pagesMeetNonTextContrast. No criterion decides WCAG 1.4.11 yet; the stylesheet
checks deliberately skip shapes, and the rendered check measures text only.

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

**Run:** `go test ./cmd/birdview/ -run TestEveryKindOfPageIsDiscovered`
**Status: verified 2026-10-16, passing, and watched failing** — 43 pages across 8 routes and 5
templates. Registering a route nothing links to failed it, naming the route; dropping the seed for
an undrawn bird failed it, naming `undrawn.html`, the one template only that page renders.

## pagesKeepToTheStylesheet

**Every templated page meets the stylesheet contract's assumption**

Covers: stylesheetContract

**Method:** on every discovered page rendered from a template: it links `styles.css`, has no
`<style>` element, sets no color as an SVG or HTML attribute, and sets a color inline only as one
of the palette's generated pairs. There is no list of exempt pages, because the contract is closed.

**Run:** `go test ./cmd/birdview/ -run TestPagesKeepToTheStylesheetContract`
**Status: verified 2026-10-16, passing.** Its first run failed on every badge in the test fixture,
which gave species a generated background and no generated text color — a fixture defect, since
production pages carry both; fixed in the fixture. **Watched failing:** adding `color:#777777` to
an inline style in the species page template failed it on every species page.

## renderedTextContrast

**Text measured in a real browser clears 7:1 on every kind of page**

Covers: pagesMeetTextContrast

**Method:** render pages in headless Chrome with every `<details>` open, and measure every visible
run of text against what is drawn behind it, found by hit-testing under the text rather than by
inheritance. Rendered: two pages of every kind (route and template), and every page outside the
stylesheet contract — the SVG documents — in full. A plain-text page is not rendered: it sets no
colors.

This is the end-to-end check on stylesheetContract. It gates where Chrome is installed and skips
without it ([bindings.md](../../../../bindings.md)).

**Run:** `go test ./cmd/birdview/ -run TestRenderedTextContrast`
**Status: verified 2026-10-16, passing, and watched failing twice:**

- Changing the selected button's text to `#2a7fa8` failed it at 2.42:1 on every selected button.
  **The stylesheet check passed over the same change**, because [tech.textContrastFloor](../../../local/tech/accessibility.md#textcontrastfloor) exempted
  every rule with its own background. The exemption is gone and the stylesheet check now fails on
  that change too; this is the case the end-to-end check exists for.
- Changing the pipeline diagram's dark text to `#aaa` failed it at 1.99:1 on
  `/docs/pipeline.svg`. No other check reads that page.

Its first runs failed for reasons in the check, each fixed: a compressed response the script was
spliced into, a test server closed before the parallel subtests ran, and text inside closed
panels, laid out but not drawn, measured against whatever was drawn there.

**Known limits:** a background image or gradient counts as no background; text whose middle lands
on a thin line drawn behind it is measured against the line; an SVG drawn into a page as an image
is measured only as its own page.

## pagesFitAPhone

**Pages drawn 360 pixels wide stay inside the window and inside their boxes**

Covers: [product.fitsAPhone](../../../local/product/presentation.md#fitsaphone)

**Method:** render two HTML pages of every kind (route and template) in headless Chrome, each in a
frame 360 pixels wide, which is a viewport of its own (Chrome will not open a window narrower than
500), with every `<details>` open. Fail on any element that reaches past the frame's right edge,
reporting only the outermost, and on any run of text that extends outside the nearest ancestor with
a background or a border, unless an element between them clips or scrolls. Fail too if the page is
wider than the frame, and refuse to judge a page laid out at any other width.

It gates where Chrome is installed and skips without it, like renderedTextContrast.

**Run:** `go test ./cmd/birdview/ -run TestPagesFitAPhone`
**Status: verified 2026-10-25, passing, and watched failing:**

- Its first run, before the fix, failed on 7 of the 12 pages. The index was 588 pixels wide: region and
  family rows ran to 499, the Presentation panel's columns to 419 with their text outside the
  panel, the diet and habitat definitions to 387, and the tiles to 420. Species pages' tables ran to
  393, `/added`'s to 415, a guide's table to 366, and an example address on the empty comparison
  page to 423. These are the failures [#2](https://github.com/google/birdview/issues/2) reported.
- Restoring the Presentation panel's fixed 380-pixel column minimum failed it on the index: the
  panel's section at 419 pixels and three runs of its text outside the panel.

Its first runs failed for reasons in the check, each fixed: Chrome's 500-pixel minimum window,
caught by the width guard, and text clipped with an ellipsis by its own element counted as bleeding.

**Known limits:** two pages of each kind, from the test fixture, so a long name only the full
dataset has is not drawn; a page at its default settings, so a view only some filters produce is
not drawn either; and text overlapping other text inside one box is not looked for.

## searchMatchesInGoAndTheBrowser

**The search keeps the same birds in Go and in the browser, and travels on the next link**

Covers: [product.searchBox](../../../local/product/filtering.md#searchbox)

**Method:** in Go, ten queries against the fixture with the birds each must keep: a name, a word
two names share, the end of one word and the start of the next, a scientific name, a code in
capitals, a name typed without its punctuation, a family; and none for words out of order, for a
query running from one field into the next, and for nonsense; an order's scientific name matches, and no family's or order's
English description is searchable. In
headless Chrome, the same queries typed into the box: the birds left showing must be exactly the
Go matcher's, the count line must count them and name the search, no empty section may show and
no section with birds may have a 0% bar, and the nothing-matches note shows exactly when nothing
does. Then, with a search in the box, a control, a tile, and Here must carry it, and the welcome
line's links, reset to defaults, and a guide must not. It gates where Chrome is installed.

**Run:** `go test ./cmd/birdview/ -run 'TestSearchMatchesNamesCodesAndTaxa|TestSearchInTheBrowser'`
**Status: verified 2026-10-25, passing.**
Watched failing: 2026-10-25, by not lowercasing what is typed in the script ("AMECRO" kept all six).
Watched failing: 2026-10-25, by removing the script's click handler (the control, tile, and Here went without the search).
Watched failing: 2026-10-25, by leaving the order's English name out of the searchable text (when the rule included it).
Watched failing: 2026-10-25, by restoring the families' and orders' English descriptions to the searchable text.
Watched failing: 2026-10-25, by matching each word anywhere, in Go and then in the script ("heron great" kept the heron).

## searchFromTheAddress

**An address with a search shows only the matches, without JavaScript**

Covers: [product.searchBox](../../../local/product/filtering.md#searchbox)

**Method:** render the index with `q` and check, in the HTML alone: only the matching tiles are
unhidden while every tile carries its search text; the count line counts the matches and names the
search; a section is hidden exactly when it counts none; the form carries every other setting and
the box shows the search; the controls carry `q`, while the welcome line's links and *Clear the
filters* do not; and a search that matches nothing says so.

**Run:** `go test ./cmd/birdview/ -run TestSearchFromTheAddress`
**Status: verified 2026-10-25, passing.**
Watched failing: 2026-10-25, by never hiding a tile the search does not match.

## boundariesResolveKnownPoints

**Known places resolve to the regions they are in**

Covers: [tech.regionBoundaries](../../../local/tech/delivery.md#regionboundaries)

**Method:** look up known coordinates in the committed boundary file with the Go reference
lookup — land polygons first, then sea boxes — and require: Sacramento → CA, Toronto → ON,
Honolulu → HI, Anchorage → AK, Washington → DC, Midway → MidwayAtoll, a point in the Gulf off
Louisiana → GulfOfMexico, a point in the open North Atlantic → NorthAtlantic; Mexico City and
London → no region. The expected answers come from geography, not from a run of the code.

**Run:** `go test ./cmd/birdview-boundaries/ -run TestKnownPoints`
**Status: verified 2026-10-18, passing, and watched failing:** making the lookup try the sea boxes
before land failed it on Tampa, which lies inside the Gulf box; deleting Midway's boundary failed
it on Sand Island, which then fell to the Central Pacific box.

## boundariesCoverEveryRegion

**Every land region has a boundary, and the file stays small**

Covers: [tech.regionBoundaries](../../../local/tech/delivery.md#regionboundaries)

**Method:** every land region in `internal/region` has at least one polygon in the boundary file
and every polygon belongs to one; the file's sea boxes equal `internal/region`'s; and the file
compressed is at most 200 KB.

**Run:** `go test ./cmd/birdview-boundaries/ -run TestCoverage`
**Status: verified 2026-10-18, passing** at 154,956 bytes compressed (425,003 uncompressed),
simplified to about 450 m on the ground. **Watched failing:** deleting Midway's boundary failed
it, and regenerating at about 100 m failed it on size, 228,788 bytes compressed.

## hereAndNowSitBesideAll

**Here and Now sit beside All, and keep every other setting**

Covers: [product.hereSelectsLocalRegion](../../../local/product/filtering.md#hereselectslocalregion), [product.nowSelectsCurrentMonth](../../../local/product/filtering.md#nowselectscurrentmonth)

**Method:** on the rendered index, Here is the control after All in Regions and Now the control
after All in Months. Now links to `/now` with every other setting kept, and `/now` redirects to
the index with the server's month as the only month, marked `no-store`. Here carries a URL
template keeping every other setting, and without JavaScript is marked inactive, focusable, and
labeled with why.

**The welcome line** ([product.pageExplainsItself](../../../local/product/presentation.md#pageexplainsitself)): it ends with *reset to defaults*, linking the
bare address (watched failing before the link existed); every other link in it is also in the
guide's *Some things to try*, its link to that list lands on the heading, and the `#here-and-now` page's
template, filled in, equals choosing the region and month by hand (`TestWelcomeLine`, watched
failing on a welcome link the guide lacks and on a template missing the month).

The Months control folds into two lines — the label, All, Now, and the explanation, then the
twelve months (`TestMonthsFoldIntoTwoLines`, watched failing against the one-line layout).

**Run:** `go test ./cmd/birdview/ -run 'TestHereAndNow|TestMonthsFoldIntoTwoLines|TestWelcomeLine'`
**Status: verified 2026-10-18, passing, and watched failing:** a template that added the reader's
region to the old one instead of replacing it failed on every region; letting `/now` be
revalidated failed on its `no-cache` and ETag.

## hereLookupAgreesInTheBrowser

**The browser finds the same region Go does, and says when it cannot**

Covers: [product.hereSelectsLocalRegion](../../../local/product/filtering.md#hereselectslocalregion)

**Method:** in headless Chrome with the location stubbed, press Here at each point of
boundariesResolveKnownPoints and require the navigation the Go lookup predicts; with permission
refused, location unavailable, and a point outside every region, require the matching message and
no navigation. Now, pressed, goes to the reader's month without requesting `/now`. Arriving by a
`#here-and-now` link goes to the page choosing the reader's region and month by hand gives; refused,
it stays and says so in the welcome line (`TestHereAndNowLink`, watched failing on a link that set
the wrong month and on a refusal message that did not say refused).

**Run:** `go test ./cmd/birdview/ -run TestHereInTheBrowser`
**Status: verified 2026-10-18, passing, and watched failing:** a browser lookup that tried the sea
boxes first failed on Tampa, and one that read a refusal as an ordinary failure failed on the
message. Needs Chrome, and skips without it.

## hereSendsNothingExtra

**The server cannot tell that Here was used**

Covers: [product.locationStaysInBrowser](../../../local/product/filtering.md#locationstaysinbrowser)

**Method:** in headless Chrome, record every request: the boundary file is requested when the
index loads, before anything is pressed; pressing Here then makes exactly one request, identical
to the one choosing that region's chip would make after Only, and no request carries a
coordinate.

**Run:** `go test ./cmd/birdview/ -run TestHereInTheBrowser`
**Status: verified 2026-10-18, passing, and watched failing:** adding the latitude to the URL a
press requests failed it on the coordinate in the request and on the page requested. A press that
navigates leaves headless Chrome waiting under the flags that keep it off the network, so that
case reads the server's request log and then stops the browser.

## tierHistogramMatchesTiers

**The histogram has one bar per tier, sized and labeled from the tiers themselves**

Covers: [product.tierHistogram](../../../local/product/ordering.md#tierhistogram)

**Method:** on the rendered index under several orderings and a filter, the histogram has exactly
one bar per tier section, in order; each bar links to its tier's section; each bar's height is its
tier's count as a percentage of the largest tier's, rounded; and each bar's label names the tier
and ends with its count. Under the size orderings, a stretch with no birds is one empty bucket per
whole 15% step it spans, between the bars on either side.

**Run:** `go test ./cmd/birdview/ -run TestTierHistogram`
**Status: verified 2026-10-19, passing, and watched failing** before the histogram and the tier
ids existed, and with bars sized by raw count instead of by share of the largest tier. The empty
buckets were watched failing by removing them: "0 empty buckets before tier 2, want 4".

## nextViewDrawsOneSet

**The home page draws the set its address names, at one scale, and every control on its count line keeps to the rule**

Covers: [product.nextView](../../../local/product/presentation.md#nextview)

**Method:** over the real dataset, with the clock fixed to October. With no month, the page shows
exactly the birds the month filter admits for October; with May, May and June, or `month=all`,
those months' birds or every bird; each most often recorded first. October narrows the set from
all. The count line has no Show setting and starts with Order by, set to frequency.
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
The home page draws the first 100 of October's 1,006 and a link to all of them (`show=all`) in
the same order, in the gap row at the end; the herons draw every heron with no link. `show=50`
draws 50, says 50 in the box, and is kept by a filter but not Reset; `show=all` is kept by a
filter too. The Show setting's First choice is a form reading First [box] that sends every other
setting but show and the find box, beside All with the count the settings match and Pinned only
(2), the chosen one marked, and its summary names the choice. The count line is two rows: what is
drawn, Show, Clear pins, and Reset, then Order by through Names; it has no Show all of its own.
`show=100`, `show=`, `show=0`, `show=-3`, and `show=zero` redirect to the address without show;
`show=50`, `show=all`, and `show=pinned` do not. The header ends with Click a bird and How to read a drawing and has no
key line; New here? Try's four comparisons each pin six real birds shown alone by length, herons
and ducks draw only those families in every month, and the birds around you now is
/#here-and-now, which in headless Chrome lands on /?region=CA from California and, refused, says
to choose under Region. In headless Chrome, the script's × goes where the server's does, and a
tile's pin clicked with text in the box carries the text. In headless Chrome, clicking the First button or the First choice's edge
sends the typed count, and clicking into its box does not; without the script, First is a submit
button. The chosen order links to its reverse and back, every other
order links to itself unreversed, reversed mass runs largest first, and reversed frequency puts
the most recorded duck last. Without a month the page is
no-store with no validator; with one it carries an ETag. The classic view is at /classic. Every
count-line setting is a `<details>` sharing one name, so the browser opens one at a time, and
Reset links to /. The pinned list names each pinned bird in alphabetical order by displayed name, linking to its tile, with a pin
that unpins it; no pin link jumps the page, and in headless Chrome a page loaded by a pin opens
scrolled where the pin was clicked. Reset keeps the pins, in order, and drops every setting; Clear pins,
before it and only when something is pinned, drops the pins and keeps every setting. The phone-width and contrast checks reach the home page through the crawl, open
every setting at once, and measure after the placing script has run. Pinning the October set's
least recorded bird, past the first 100, draws it after one gap row and none after it, and the count line says "plus 1 pin" ("plus 2 pins" for two); pinning two mid-order birds by length draws gap rows before, between, and after them summing to Hidden, each linking to Show all, and Show all draws no gaps. Pinned
only links to `show=pinned` with the order kept and draws nothing; with two pinned it
draws those two by mass, unpinning one leaves the other, and First brings the
October birds back with both; `show=all` matches what the default does. Here, filled with FL, is
the page with only Florida chosen, every other setting kept, and the pinned-only
view left, and the page carries the button, its status, and here.js. Month's choices are one
line, All, Now, and the twelve months, with a rule only before January; Now is chosen with no
month named and links to none, and All is chosen with `month=all`. In headless Chrome, choosing
May in Month loads a page with Month open and nothing else, and loading it again after no choice
opens nothing. In headless Chrome at 1280 and 390 px, every pin on October's birds by length
stands within a pixel of its tile's mass cube, tiles with a note under the footer and tiles
stretched to a taller neighbor included.

**Run:** `go test ./cmd/birdview/ -run 'TestNext'` (the browser check gates where Chrome is installed)
**Status: verified 2026-10-26, passing.**
Watched failing: 2026-10-26, by dropping the frequency default ("coshum (0.000207) follows a bird at 0.000015").
Watched failing: 2026-10-26, by dropping the month default ("1041 birds, the month filter admits 1006"), and by leaving Order by off the count line.
Watched failing: 2026-10-26, by scaling each bird to its own length.
Watched failing: 2026-10-26, by ordering frequency least recorded first.
Watched failing: 2026-10-26, by letting an empty family filter mean none ("/next?family= shows 0 birds").
Watched failing: 2026-10-26, by a remove link that removed nothing, and by an add link that kept the box.
Watched failing: 2026-10-26, by a toggle that kept the find box and Show all.
Watched failing: 2026-10-26, by ignoring all=1 ("draws 200 of 1041").
Watched failing: 2026-10-26, by ignoring rev=1 ("reversed mass: gnwtea after gargan"), by a chosen order that never reverses back, and by other orders keeping rev.
Watched failing: 2026-10-26, by dropping Show all from the count line.
Watched failing: 2026-10-26, by capping added birds with the rest ("after adding tersan from past the first 100, it is not drawn"), and by a toggle that only undid a removal.
Watched failing: 2026-10-26, by Clear that cleared nothing ("cleared: 100 drawn of 1006"), and by a filter that kept the cleared view.
Watched failing: 2026-10-26, by Here keeping clear=1, by renaming its button, and by dropping Now.
Watched failing: 2026-10-26, by show=pinned that hid nothing, and by Here keeping show=pinned.
Watched failing: 2026-10-26, by a Reset that dropped the pins, and by a Clear pins that kept them.
Watched failing: 2026-10-26, by a pin that jumps to its tile, by not keeping the scroll ("the new page opened at 0"), and by dropping the pinned list or its entries.
Watched failing: 2026-10-26, by a comparison without show=pinned ("106 drawn"), by an unknown bird in one ("5 known"), by keeping "Drawn to scale", and by here.js ignoring the page's Region setting.
Watched failing: 2026-10-26, by a tile's pin dropping the typed text, by the script's list links leaving it out, and by pinning emptying the box ("pinning went to /?family=Ardeidae&pin=shshaw").
Watched failing: 2026-10-26, by dropping the edge click's script, by sending on every click ("box 1"), and by First not being a submit button.
Watched failing: 2026-10-26, by a filter dropping every show, by ignoring a count ("show=50: 100 drawn"), by the First form sending show twice, by All counting only the drawn birds ("under Pinned only, All offers 0 birds"), by one count-line row, and by an unmarked First.
Watched failing: 2026-10-26, by the count line saying "and 1 added" instead of "plus 1 pin".
Watched failing: 2026-10-26, by ignoring cap, by a filter dropping it, by not redirecting cap=100, by Show all naming cap=all, and by the cap form sending cap twice.
Watched failing: 2026-10-26, by placing the pin a fixed 60px above the tile's bottom ("28 of 100 pins are not level with the mass cube", 16px high), and by sharing a stretched tile's height among all its rows ("20 of 100", 4px).
Watched failing: 2026-10-26, by dropping the gap after the last pin ("has 104 tiles and gaps, want 102 tiles and 3 gaps"), by dropping gaps before pins, by keeping pin order in the pinned list, and by sorting it by id instead of displayed name.
Watched failing: 2026-10-26, by not reopening the setting ("with [] open"), by never forgetting it ("a plain reload opened [Month]"), and by dropping the rule and the one line.
Watched failing: 2026-10-26, by restoring the 1,200 cap ("/next has 1006 birds; the check needs more than 1200").
Watched failing: 2026-10-26, by caching /next without a month like any page.
Watched failing: 2026-10-26, by anchoring each popover to its own setting with no placing script (the phone-width check: "the page is 543px wide in a 360px window").
Watched failing: 2026-10-26, by never marking a pin pinned, by a pin that never unpins, by dropping the families' colors, and by the script drawing every pin unpinned.
Watched failing: 2026-10-26, by dropping the shared name ("0 of 9 count-line settings share the name ctl").
Watched failing: 2026-10-26, by not lowercasing what is typed in the script ("AMECRO": the script says "", the server "1 bird matches.").
Watched failing: 2026-10-26, by the script listing matches over the limit, and by its offering Add for a bird already shown.
Watched failing: 2026-10-26, by the toggle not removing a bird the filters chose, and by ignoring exclude=.
Watched failing: 2026-10-26, by the server listing 23 matches for "hawk".
