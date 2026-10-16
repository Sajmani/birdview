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
