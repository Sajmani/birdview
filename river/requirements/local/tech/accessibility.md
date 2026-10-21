package tech

# Accessibility

## badgeContrast

**Badge contrast meets WCAG 2.2 SC 1.4.6**

Subject: `badge.contrast.ratio`
Value: `>= 7.0`
Goals: goals.openToEveryone

Every generated badge color yields a contrast ratio of at least 7:1 against **the text it actually
carries**, which is black or white per badge, computed per WCAG 2.2 relative luminance.

Requiring black text would be stricter than WCAG and would cost about a third of the
distinguishable colors available to product.generatedBadgeColors.

Rationale: WCAG is a **mandatory** source scoped to the contrast criteria only — see
[module.md](../../module.md#wcag). BirdView makes no broader conformance claim.

## textContrastFloor

**Link and text colors meet a stated contrast floor**

Subject: `text.contrast.ratio`
Goals: goals.openToEveryone

Every text color the stylesheet declares for text on a bare surface clears **7:1** (WCAG 2.2
SC 1.4.6, Contrast Enhanced) against **both** the page background and the panel background. Links
and ordinary text are held to the same floor, which is also the floor badgeContrast holds badges to.

**"Text color" means `fill` as well as `color`.** SVG text is painted with `fill`, so a diagram
label, such as those of
[product.labeledDetailDiagram](../product/measurements.md#labeleddetaildiagram), is text and is
held to this floor. Shapes are not: a stroke or fill on a line, a polygon or the
mass cube is a graphical object, governed by WCAG 1.4.11's 3:1. A rule styles text when it styles
an element the stylesheet gives a font property.

**A rule that sets its own background is held to 7:1 against that background**, not the page's.
A badge's background and text color are generated together and set inline, so that pair is
badgeContrast's; a pair written in the stylesheet, such as a selected button's dark text on
`--accent`, is this requirement's. Stylesheet checks once missed such a pair at 2.42:1; only
rendering the page caught it (app.renderedTextContrast).

**No link may go uncolored.** A link with no rule inherits the browser default — `#0000EE`
unvisited and `#551A8B` visited — which on this background is 1.90:1 and 1.62:1.

Rationale: hand-written CSS is as dangerous as a generated palette; the iNaturalist link of
product.inaturalistLink shipped with no color rule and was effectively invisible. One floor covers
all text, with no lower floor for secondary text; see
[decisions.oneContrastFloor](../../decisions.md#onecontrastfloor).

---

## webPages

**What counts as a BirdView web page**

Subject: `pages.domain`
Binds: wcag.WebPage
Goals: goals.openToEveryone

**Every response the server sends as `text/html`, `image/svg+xml` or `text/plain`, at any route,
is a BirdView web page**, with everything drawn into it. Today that is the index, the species
page in its drawn and undrawn forms, the comparison, the guides and their index, the added-values
and excluded-birds pages, the pipeline diagram and the favicon opened at their own URLs, and the
plain-text error responses.

The list is illustrative, not the definition: the checks discover pages from the server's routes,
so a page added later is a web page the day it exists.

**This binds WCAG's definition and does not narrow it.** WCAG makes a web page of any resource
obtained from a URI, so an error response or an SVG opened on its own is one. Leaving a page out
would excuse it from a mandatory source, which takes an accepted gap in
[decisions.md](../../decisions.md), page by page.

Rationale: treating "the pages" as what the stylesheet styles left out the pipeline diagram,
served as its own SVG document with its own colors, and the page templates, served as raw source
under `/static/` (tech.servesOnlyItsPages).
