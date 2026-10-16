package tech

# Process

## bundleIsNormative

**The spec bundle is normative**

Subject: `process.spec_location`
Value: `river/`
Goals: goals.goodDataCitizen

The repository follows the loop in [process.md](../../../README.md). `AGENTS.md` names `river/` as
normative so agents reading the instruction file are pointed at it.

## pacedInaturalistMapping

**The iNaturalist mapping is built by a separate, paced tool**

Goals: goals.goodDataCitizen

Resolving species to iNaturalist taxon IDs happens in `cmd/birdview-inatmap`, never in
`birdview-ingest` and never while serving. The tool waits at least one second between requests
(`inatapi.RequestRate`) and identifies itself (identifiedClient). Its output is a table committed to the repository;
the ingest joins against that table and makes no network request.

Rationale: 1082 requests at the rate iNaturalist asks for is a quarter of an hour, which does not
belong in a build — and a build that needs the network is a build that fails on a plane. Keeping
resolution separate also keeps selfContainedServer and the no-network-in-tests binding intact without exception.
The table going stale is the accepted cost; a stale ID fails visibly and the refresh is one
command.

---

## textContrastFloor

**Link and text colors meet a stated contrast floor**

Subject: `text.contrast.ratio`
Goals: goals.sizeIntuition

Every text color the stylesheet declares for text on a bare surface clears **7:1** (WCAG 2.2
SC 1.4.6, Contrast Enhanced) against **both** the page background and the panel background. Links
and ordinary text are held to the same floor, which is also the floor badgeContrast holds badges to.

**"Text color" means `fill` as well as `color`.** SVG text is painted with `fill`, so a diagram
label is text and is held to this floor exactly like a paragraph. Shapes are not: a stroke or a
fill on a line, a polygon or the mass cube is a graphical object, governed by WCAG 1.4.11's 3:1
rather than by this requirement. What separates the two is whether the rule styles a text element,
which the stylesheet already declares by giving that element a font property.

That distinction went unwritten, and the check read `color` alone. The wing chord and wingspan
labels of [product.labeledDetailDiagram](../product/measurements.md#labeleddetaildiagram) shipped painted `#0ea5e9` — **6.44:1 on the page background
and 5.28:1 on a panel** — and nothing saw them, because SVG text does not use `color`.

Rules that set their own background are excluded and governed by badgeContrast instead: badges and chips
pair a generated background with a matching text color, and the pair is what has to clear 7:1.

**No link may go uncolored.** A link with no rule inherits the browser default — `#0000EE`
unvisited and `#551A8B` visited — which on this background is 1.90:1 and 1.62:1.

Rationale: contrast was adopted for badges only, because badge colors are *generated* and an
unchecked generator is obviously dangerous. Hand-written CSS turned out to be just as dangerous and
had no check at all: the iNaturalist link of product.inaturalistLink shipped with no color rule and was effectively
invisible.

One floor rather than two. The first version of this requirement set 4.5:1 for ordinary text,
because `--dim` sat at 6.96:1 on the page background and 5.71:1 on a panel and repainting the
interface's secondary text looked disproportionate to a single broken link. The owner then reported
the dim text as hard to read too, which settled it: `--dim` and `--accent` moved, and the exception
went with them. See [decisions.oneContrastFloor](../../decisions.md#onecontrastfloor).

---

## recordBeforeConsuming

**No source is consumed before it is recorded**

Goals: goals.goodDataCitizen

Every input source appears in [sources.md](../../requirements.md) with its tier, pin, license and status
*before* code that reads it is written.

Rationale: this project found, after designing around three datasets, that two of them could not
legally be redistributed and a third did not contain the fields it was chosen for. Recording first
is cheaper than discovering later.

---

Project bindings moved to [bindings.md](../../../bindings.md) on adopting River, whose layout gives them
their own file.
