package app

# app — acceptance criteria

How each requirement in [app.md](app.md) is decided. Each criterion names what it covers in a
`Covers:` field; nothing lists criteria per requirement.

**Uncovered:** pagesMeetNonTextContrast. No criterion decides WCAG 1.4.11 yet; the stylesheet
checks deliberately skip shapes, and the rendered check measures text only.

---

## everyPageIsFound

**Every route and every page template yields a page the checks see**

Covers: tech.webPages

**Method:** crawl the server from a handful of seeds — the pages nothing links to — following
every same-origin link and source, and require that every route the server registers, every page
template it can render, and each of the three content types tech.webPages names produced at least
one page.

It decides the domain, not a criterion over it: it covers tech.webPages, and the page
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
without it ([bindings.md](../../bindings.md)).

**Run:** `go test ./cmd/birdview/ -run TestRenderedTextContrast`
**Status: verified 2026-10-16, passing, and watched failing twice:**

- Changing the selected button's text to `#2a7fa8` failed it at 2.42:1 on every selected button.
  **The stylesheet check passed over the same change**, because tech.textContrastFloor exempted
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

## boundariesResolveKnownPoints

**Known places resolve to the regions they are in**

Covers: tech.regionBoundaries

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

Covers: tech.regionBoundaries

**Method:** every land region in `internal/region` has at least one polygon in the boundary file
and every polygon belongs to one; the file's sea boxes equal `internal/region`'s; and the file
compressed is at most 200 KB.

**Run:** `go test ./cmd/birdview-boundaries/ -run TestCoverage`
**Status: verified 2026-10-18, passing** at 154,956 bytes compressed (425,003 uncompressed),
simplified to about 450 m on the ground. **Watched failing:** deleting Midway's boundary failed
it, and regenerating at about 100 m failed it on size, 228,788 bytes compressed.

## hereAndNowSitBesideAll

**Here and Now sit beside All, and keep every other setting**

Covers: product.hereSelectsLocalRegion, product.nowSelectsCurrentMonth

**Method:** on the rendered index, Here is the control after All in Regions and Now the control
after All in Months. Now links to `/now` with every other setting kept, and `/now` redirects to
the index with the server's month as the only month, marked `no-store`. Here carries a URL
template keeping every other setting, and without JavaScript is marked inactive, focusable and
labeled with why.

**The welcome line** (product.pageExplainsItself): it ends with *reset to defaults*, linking the
bare address (watched failing before the link existed); every other link in it is also in the
guide's *Some things to try*, its link to that list lands on the heading, and the `#here-and-now` page's
template, filled in, equals choosing the region and month by hand (`TestWelcomeLine`, watched
failing on a welcome link the guide lacks and on a template missing the month).

The Months control folds into two lines — the label, All, Now and the explanation, then the
twelve months (`TestMonthsFoldIntoTwoLines`, watched failing against the one-line layout).

**Run:** `go test ./cmd/birdview/ -run 'TestHereAndNow|TestMonthsFoldIntoTwoLines|TestWelcomeLine'`
**Status: verified 2026-10-18, passing, and watched failing:** a template that added the reader's
region to the old one instead of replacing it failed on every region; letting `/now` be
revalidated failed on its `no-cache` and ETag.

## hereLookupAgreesInTheBrowser

**The browser finds the same region Go does, and says when it cannot**

Covers: product.hereSelectsLocalRegion

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

Covers: product.locationStaysInBrowser

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

Covers: product.tierHistogram

**Method:** on the rendered index under several orderings and a filter, the histogram has exactly
one bar per tier section, in order; each bar links to its tier's section; each bar's height is its
tier's count as a percentage of the largest tier's, rounded; and each bar's label names the tier
and ends with its count.

**Run:** `go test ./cmd/birdview/ -run TestTierHistogram`
**Status: verified 2026-10-19, passing, and watched failing** before the histogram and the tier
ids existed, and with bars sized by raw count instead of by share of the largest tier.

