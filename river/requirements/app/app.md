package app

# app — composed requirements for the BirdView web application

Everything the web application must satisfy. Implementations and releases of the server are
verified against this package; its criteria are in [acceptance.md](acceptance.md).

**Two parts.** First, every source requirement that applies to the application **unchanged**,
linked to its full text. Then the requirements **composition wrote**, in full, each naming its
sources in `Composed from:`. A source requirement not listed here belongs to another package or
is not yet composed.

**Status: composed by an agent, awaiting the owner's approval.** Not tagged.

## Composition record

| Source | Tier | Pin |
| --- | --- | --- |
| [`wcag`](../sources/wcag/requirements.md) | mandatory | WCAG 2.2, W3C Recommendation of 12 December 2024; SHA-256 in [PROVENANCE.md](../sources/wcag/PROVENANCE.md) |
| [`product`](../sources/product/product.md) | local | the commit this file is in |
| [`tech`](../sources/tech/tech.md) | local | the commit this file is in |
| [`decisions`](../decisions.md) | — | the commit this file is in |

Precedence is the ranking in [requirements.md](../requirements.md): a mandatory source beats a
local one. **No conflict arose.** BirdView's local requirements only tighten what WCAG asks, which
a lower tier may always do; nothing local loosens it.

## Applies unchanged

### Diagram

- [product.segmentComposition](../sources/product/diagram.md#segmentcomposition) — Segment composition
- [product.segmentsSumToLength](../sources/product/diagram.md#segmentssumtolength) — Segments abut exactly and sum to total length
- [product.scaleToggle](../sources/product/diagram.md#scaletoggle) — Scale is a user toggle, defaulting to per-section
- [product.crossbarAtVerticalScale](../sources/product/diagram.md#crossbaratverticalscale) — The wing crossbar is drawn at the vertical scale
- [product.proportionDifferentiates](../sources/product/diagram.md#proportiondifferentiates) — Proportion differentiates species
- [product.oneDiagramVocabulary](../sources/product/diagram.md#onediagramvocabulary) — One diagram vocabulary, on every surface that draws one
- [product.measuredVersusModeled](../sources/product/measurements.md#measuredversusmodeled) — The diagram shows which lengths are measured and which are modeled
- [product.labeledDetailDiagram](../sources/product/measurements.md#labeleddetaildiagram) — The detail diagram is labeled with each segment and its measurement

### What the reader is told about the numbers

- [product.tileMarksModeled](../sources/product/measurements.md#tilemarksmodeled) — The tile marks any metric that is not measured
- [product.metricProvenanceExplained](../sources/product/measurements.md#metricprovenanceexplained) — The detail view explains each metric's provenance
- [product.categoryProvenance](../sources/product/measurements.md#categoryprovenance) — The detail view lists every category the bird is in, and where each came from
- [product.borrowedMeasurementsSaySo](../sources/product/measurements.md#borrowedmeasurementssayso) — A drawing made from another drawn species' measurements says so
- [product.namesRecordsFiledUnder](../sources/product/presentation.md#namesrecordsfiledunder) — A species says which names its records were filed under
- [product.shareOfRecords](../sources/product/presentation.md#shareofrecords) — Each bird says how much of what people record it is
- [product.statusIsPublished](../sources/product/dataset.md#statusispublished) — Status is published, not just used
- [product.pipelineDiagram](../sources/product/measurements.md#pipelinediagram) — The documentation draws the data pipeline, field by field
- [product.selfDescriptionStaysTrue](../sources/product/presentation.md#selfdescriptionstaystrue) — What the project says about itself stays true

### Names and units

- [product.englishNameChoice](../sources/product/diagram.md#englishnamechoice) — The reader chooses which English name, and the URL does not change
- [product.imperialByDefault](../sources/product/presentation.md#imperialbydefault) — Imperial by default, metric available
- [product.oneUnitInComparison](../sources/product/presentation.md#oneunitincomparison) — One unit across whatever is being compared

### Filters

- [product.familyFilter](../sources/product/filtering.md#familyfilter) — Filter by family
- [product.regionAndMonthFilter](../sources/product/filtering.md#regionandmonthfilter) — Filter by region and month
- [product.filtersCompose](../sources/product/months.md#filterscompose) — Filters compose
- [product.regionsGroupedTwoWays](../sources/product/months.md#regionsgroupedtwoways) — Regions are grouped two ways
- [product.familiesByOrder](../sources/product/months.md#familiesbyorder) — Families are grouped by taxonomic order
- [product.habitatAndDietFilters](../sources/product/months.md#habitatanddietfilters) — Filter by habitat and by diet
- [product.onlyPerGroup](../sources/product/presentation.md#onlypergroup) — Every group offers "Only"
- [product.firstLoadShowsAll](../sources/product/presentation.md#firstloadshowsall) — First load shows everything
- [product.helpfulEmptyState](../sources/product/presentation.md#helpfulemptystate) — The empty state says what to do
- [product.hereSelectsLocalRegion](../sources/product/filtering.md#hereselectslocalregion) — Here selects the region the reader is in
- [product.nowSelectsCurrentMonth](../sources/product/filtering.md#nowselectscurrentmonth) — Now selects the current month
- [product.locationStaysInBrowser](../sources/product/filtering.md#locationstaysinbrowser) — The reader's location never leaves the browser

### Ordering and sections

- [product.oneSelectedMetric](../sources/product/ordering.md#oneselectedmetric) — One primary metric, user-selected
- [product.ascendingOrder](../sources/product/ordering.md#ascendingorder) — Ascending order
- [product.orderByFrequency](../sources/product/presentation.md#orderbyfrequency) — Birds can be ordered by how often they are recorded
- [product.sectionBreakRule](../sources/product/ordering.md#sectionbreakrule) — Sections break at N% above the section minimum
- [product.sectionThresholdDefault](../sources/product/ordering.md#sectionthresholddefault) — N defaults to 15 and is adjustable
- [product.sectionHeadersStateRange](../sources/product/ordering.md#sectionheadersstaterange) — Section headers state range and count
- [product.tierHistogram](../sources/product/ordering.md#tierhistogram) — A histogram shows how many species each tier holds

### Pages and views

- [product.diagramOnlyMode](../sources/product/presentation.md#diagramonlymode) — Diagram is the only presentation mode in v1
- [product.tileContents](../sources/product/presentation.md#tilecontents) — Tile contents
- [product.tileOpensDetail](../sources/product/presentation.md#tileopensdetail) — A tile opens a detail view
- [product.comparisonOneScale](../sources/product/presentation.md#comparisononescale) — A comparison view holds several birds to one scale
- [product.pickFromIndex](../sources/product/presentation.md#pickfromindex) — Birds are picked for comparison from the index, and the picks are the mode
- [product.undrawnSpeciesAnswer](../sources/product/presentation.md#undrawnspeciesanswer) — A species BirdView does not draw still answers for itself
- [product.escapedBirdsListed](../sources/product/measurements.md#escapedbirdslisted) — Escaped and kept birds are left off the index, and listed
- [product.pageExplainsItself](../sources/product/presentation.md#pageexplainsitself) — The page explains itself
- [product.collapsibleControls](../sources/product/presentation.md#collapsiblecontrols) — The controls collapse
- [product.togglesKeepPlace](../sources/product/presentation.md#toggleskeepplace) — Toggling a control keeps your place
- [product.inaturalistLink](../sources/product/presentation.md#inaturalistlink) — The detail view links to the species on iNaturalist
- [product.noPersonalAddress](../sources/product/presentation.md#nopersonaladdress) — No personal address is published

### Serving

- [tech.selfContainedServer](../sources/tech/delivery.md#selfcontainedserver) — The server is self-contained at runtime
- [tech.servesOnlyItsPages](../sources/tech/delivery.md#servesonlyitspages) — The server serves its pages and what they load, and nothing else
- [tech.servesOnLocalhost](../sources/tech/delivery.md#servesonlocalhost) — Serves on localhost:8080
- [tech.compressedResponses](../sources/tech/delivery.md#compressedresponses) — Responses are compressed when the client asks
- [tech.cacheLifetimeStated](../sources/tech/delivery.md#cachelifetimestated) — Every response says how long it may be reused
- [tech.regionBoundaries](../sources/tech/delivery.md#regionboundaries) — Region boundaries are generated from Natural Earth and kept small

## Excluded

Withdrawn with the photograph design; see [decisions.photographsWithdrawn](../decisions.md#photographswithdrawn).

- [product.speciesPhotographs](../sources/product/diagram.md#speciesphotographs) — Species photographs. Excluded by [decisions.photographsWithdrawn](../decisions.md#photographswithdrawn)
- [product.perchedOrInFlight](../sources/product/diagram.md#perchedorinflight) — Perched and in-flight selection. Excluded by [decisions.photographsWithdrawn](../decisions.md#photographswithdrawn)
- [product.photoAttribution](../sources/product/diagram.md#photoattribution) — Photograph attribution. Excluded by [decisions.photographsWithdrawn](../decisions.md#photographswithdrawn)
- [product.photoCaching](../sources/product/diagram.md#photocaching) — Local caching of photographs. Excluded by [decisions.photographsWithdrawn](../decisions.md#photographswithdrawn)

---

## Composed

## pagesMeetTextContrast

**Every BirdView web page meets WCAG 1.4.6 in full, without the large-text allowance**

Composed from: [wcag.ContrastEnhanced](../sources/wcag/requirements.md#contrastenhanced), [wcag.FullPages](../sources/wcag/requirements.md#fullpages), [wcag.ContrastRatio](../sources/wcag/requirements.md#contrastratio), [tech.webPages](../sources/tech/accessibility.md#webpages), [decisions.oneContrastFloor](../decisions.md#onecontrastfloor)

Every web page, as tech.webPages defines one, meets wcag.ContrastEnhanced over every part of the
page (wcag.FullPages), measured as wcag.ContrastRatio defines. All text is held to 7:1, large text
included: wcag.ContrastEnhanced allows 4.5:1 for large text, and decisions.oneContrastFloor gives
that up. Meeting this meets wcag.ContrastMinimum too.

**No page is excepted.** A page that sets no colors, such as a plain-text error response, is drawn
in the reader's own colors and meets this with nothing to measure.

## pagesMeetNonTextContrast

**Every BirdView web page meets WCAG 1.4.11 in full**

Composed from: [wcag.NonTextContrast](../sources/wcag/requirements.md#nontextcontrast), [wcag.FullPages](../sources/wcag/requirements.md#fullpages), [tech.webPages](../sources/tech/accessibility.md#webpages)

Every web page meets wcag.NonTextContrast over every part of the page: the graphics a reader needs
to read a diagram, and the controls and their states, have 3:1 against adjacent colors.

## stylesheetContract

**The stylesheet and the palette carry text contrast for every page that uses only them**

Composed from: [tech.textContrastFloor](../sources/tech/accessibility.md#textcontrastfloor), [tech.badgeContrast](../sources/tech/accessibility.md#badgecontrast), [wcag.ContrastRatio](../sources/wcag/requirements.md#contrastratio)

The stylesheet is an intermediary: it meets pagesMeetTextContrast for any page that takes its
colors from it, so that requirement is checked once on the stylesheet rather than once per page.

- **Guarantee.** Every text color `styles.css` declares clears 7:1 against both the page and the
  panel background, or against the rule's own background where it sets one, and no link goes
  uncolored (tech.textContrastFloor). Every badge's generated
  background clears 7:1 against the text color generated with it (tech.badgeContrast).
- **Assumption.** The page links `styles.css`, has no `<style>` element, and sets a color inline
  only as a badge's generated pair. A page that breaks this is outside the contract and is
  checked on its own.
- **Scope.** Text, in pages rendered from BirdView's templates. **Not** an SVG document served at
  its own URL, which carries its own styles, and **not** non-text contrast, which this contract
  does not guarantee at all.
- **Override.** Closed. A template may not set a color except through the palette; one that does
  has left the contract, which the assumption check reports.
