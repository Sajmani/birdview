package app

# app — composed requirements for the BirdView web application

**Composed, not authored.** Change it by changing a source and recomposing, never by editing files in
this directory by hand. Composed from `wcag`, `product`, `tech`, and `decisions` under
[precedence.md](../../../precedence.md); approvals are recorded in
[approvals.md](../../../approvals.md).

Everything the BirdView web application (`cmd/birdview`, `web/`, `docs/`, and
`cmd/birdview-boundaries`) must satisfy. Its acceptance criteria are in [`acceptance.md`](acceptance.md),
and the guardrails and verification observations covering them are in
[conformance.md](../../../../conformance.md).

## Composed from

- wcag (Mandatory)
- gbifterms (Mandatory)
- product (Local)
- tech (Local)
- decisions
- precedence

## Composition record

| Source | Tier | Pin |
| --- | --- | --- |
| [`wcag`](../../../imported/w3.org/wcag/wcag.md) | mandatory | WCAG 2.2, W3C Recommendation of 12 December 2024; SHA-256 in [PROVENANCE.md](../../../imported/w3.org/wcag/PROVENANCE.md) |
| [`gbifterms`](../../../imported/gbif.org/gbifterms/gbifterms.md) | mandatory (page-credit clauses) | GBIF Data User Agreement (15 September 2026) and Citation Guidelines; SHA-256 in [PROVENANCE.md](../../../imported/gbif.org/gbifterms/PROVENANCE.md) |
| [`product`](../../../local/product/product.md) | local | the commit this file is in |
| [`tech`](../../../local/tech/tech.md) | local | the commit this file is in |
| [`decisions`](../../../decisions.md) | — | the commit this file is in |

Precedence is the ranking in [precedence.md](../../../precedence.md): a mandatory source beats a
local one. **No conflict arose.** BirdView's local requirements only tighten what WCAG and source
licenses ask, which a lower tier may do.

## Applies unchanged

### Diagram

- [product.segmentComposition](../../../local/product/diagram.md#segmentcomposition) — Segment composition
- [product.segmentsSumToLength](../../../local/product/diagram.md#segmentssumtolength) — Segments abut exactly and sum to total length
- [product.crossbarAtVerticalScale](../../../local/product/diagram.md#crossbaratverticalscale) — The wing crossbar is drawn at the vertical scale
- [product.proportionDifferentiates](../../../local/product/diagram.md#proportiondifferentiates) — Proportion differentiates species
- [product.oneDiagramVocabulary](../../../local/product/diagram.md#onediagramvocabulary) — One diagram vocabulary, on every surface that draws one
- [product.measuredVersusModeled](../../../local/product/measurements.md#measuredversusmodeled) — The diagram shows which lengths are measured and which are modeled
- [product.labeledDetailDiagram](../../../local/product/measurements.md#labeleddetaildiagram) — The detail tile names each segment and its measurement, keyed to the drawing by color

### What the reader is told about the numbers

- [product.tileMarksModeled](../../../local/product/measurements.md#tilemarksmodeled) — The tile marks any metric that is not measured
- [product.provenancePerMetric](../../../local/product/measurements.md#provenancepermetric) — Every presented metric carries a provenance record
- [product.everydayDefinitions](../../../local/product/measurements.md#everydaydefinitions) — Every measurement says what it is, in everyday words, before where it came from
- [product.metricProvenanceExplained](../../../local/product/measurements.md#metricprovenanceexplained) — The detail view explains each metric's provenance
- [product.categoryProvenance](../../../local/product/measurements.md#categoryprovenance) — The detail view lists every category the bird is in, and where each came from
- [product.borrowedMeasurementsSaySo](../../../local/product/measurements.md#borrowedmeasurementssayso) — A drawing made from another drawn species' measurements says so
- [product.namesRecordsFiledUnder](../../../local/product/presentation.md#namesrecordsfiledunder) — A species says which names its records were filed under
- [product.shareOfRecords](../../../local/product/presentation.md#shareofrecords) — Each bird says how much of what people record it is
- [product.statusIsPublished](../../../local/product/dataset.md#statusispublished) — Status is published, not just used
- [product.pipelineDiagram](../../../local/product/measurements.md#pipelinediagram) — The documentation draws the data pipeline, field by field
- [product.selfDescriptionStaysTrue](../../../local/product/presentation.md#selfdescriptionstaystrue) — What the project says about itself stays true

### Names and units

- [product.englishNameChoice](../../../local/product/diagram.md#englishnamechoice) — The reader chooses which English name, and the URL does not change
- [product.imperialByDefault](../../../local/product/presentation.md#imperialbydefault) — Imperial by default, metric available
- [product.oneUnitInComparison](../../../local/product/presentation.md#oneunitincomparison) — One unit across whatever is being compared

### Filters

- [product.familyFilter](../../../local/product/filtering.md#familyfilter) — Filter by family
- [product.regionAndMonthFilter](../../../local/product/filtering.md#regionandmonthfilter) — Filter by region and month
- [product.reportingFrequencyFilter](../../../local/product/filtering.md#reportingfrequencyfilter) — Filter by how often a bird is reported where and when the reader is looking
- [product.filtersCompose](../../../local/product/months.md#filterscompose) — Filters compose
- [product.familiesByOrder](../../../local/product/months.md#familiesbyorder) — Families are grouped by taxonomic order
- [product.habitatAndDietFilters](../../../local/product/months.md#habitatanddietfilters) — Filter by habitat and by diet
- [product.hereSelectsLocalRegion](../../../local/product/filtering.md#hereselectslocalregion) — Here selects the region the reader is in
- [product.locationStaysInBrowser](../../../local/product/filtering.md#locationstaysinbrowser) — The reader's location never leaves the browser

### Ordering and sections

- [product.oneSelectedMetric](../../../local/product/ordering.md#oneselectedmetric) — One primary metric, user-selected
- [product.ascendingOrder](../../../local/product/ordering.md#ascendingorder) — Ascending order
- [product.orderByFrequency](../../../local/product/presentation.md#orderbyfrequency) — Birds can be ordered by how often they are recorded
- [product.sectionBreakRule](../../../local/product/ordering.md#sectionbreakrule) — Sections break at N% above the section minimum
- [product.sectionThresholdDefault](../../../local/product/ordering.md#sectionthresholddefault) — N defaults to 10 and is adjustable
- [product.sectionHeadersStateRange](../../../local/product/ordering.md#sectionheadersstaterange) — Section headers state range and count
- [product.tierHistogram](../../../local/product/ordering.md#tierhistogram) — A histogram shows how many species each tier holds

### Pages and views

- [product.diagramOnlyMode](../../../local/product/presentation.md#diagramonlymode) — Diagram is the only presentation mode in v1
- [product.tileContents](../../../local/product/presentation.md#tilecontents) — Tile contents
- [product.tileOpensDetail](../../../local/product/presentation.md#tileopensdetail) — A tile opens a detail view
- [product.undrawnSpeciesAnswer](../../../local/product/presentation.md#undrawnspeciesanswer) — A species BirdView does not draw still answers for itself
- [product.escapedBirdsListed](../../../local/product/measurements.md#escapedbirdslisted) — Escaped and kept birds are left off the home page, and listed
- [product.fitsAPhone](../../../local/product/presentation.md#fitsaphone) — Every page fits a phone held upright
- [tech.textSizeFloor](../../../local/tech/accessibility.md#textsizefloor) — No text renders below 13px, and running text is 15px, 16px on a phone
- [product.nextView](../../../local/product/presentation.md#nextview) — The home page: one view of birds
- [product.birdPresets](../../../local/product/presentation.md#birdpresets) — A Birds setting chooses a kind of bird by filling in Family, Habitat, and Diet
- [product.chooseOneSettings](../../../local/product/presentation.md#chooseonesettings) — A setting that takes one choice looks like one: a joined row, one segment filled
- [product.bandingCodes](../../../local/product/dataset.md#bandingcodes) — Each bird carries its four-letter banding code, where the Bird Banding Laboratory gives one
- [product.ebirdLink](../../../local/product/presentation.md#ebirdlink) — The detail view and every tile link to the species on eBird
- [product.noPersonalAddress](../../../local/product/presentation.md#nopersonaladdress) — No personal address is published
- [product.reportDataIssue](../../../local/product/presentation.md#reportdataissue) — A species page asks for corrections where its data is shown

### Serving

- [tech.selfContainedServer](../../../local/tech/delivery.md#selfcontainedserver) — The server is self-contained at runtime
- [tech.servesOnlyItsPages](../../../local/tech/delivery.md#servesonlyitspages) — The server serves its pages and what they load, and nothing else
- [tech.crawlersSkipFilteredViews](../../../local/tech/delivery.md#crawlersskipfilteredviews) — Crawlers are asked to skip filtered views
- [tech.stableUrls](../../../local/tech/delivery.md#stableurls) — These URLs keep working; everything else may change
- [tech.oneCanonicalHost](../../../local/tech/delivery.md#onecanonicalhost) — A deployment answers at one address
- [tech.servesOnLocalhost](../../../local/tech/delivery.md#servesonlocalhost) — Listens on localhost:8080 unless told otherwise
- [tech.compressedResponses](../../../local/tech/delivery.md#compressedresponses) — Responses are compressed when the client asks
- [tech.cacheLifetimeStated](../../../local/tech/delivery.md#cachelifetimestated) — Every response says how long it may be reused
- [tech.regionBoundaries](../../../local/tech/delivery.md#regionboundaries) — Region boundaries are generated from Natural Earth and kept small

## Excluded requirements

- [product.speciesPhotographs](../../../local/product/diagram.md#speciesphotographs) *(excluded)* — Species photographs. Excluded by [decisions.photographsWithdrawn](../../../decisions.md#photographswithdrawn)
- [product.perchedOrInFlight](../../../local/product/diagram.md#perchedorinflight) *(excluded)* — Perched and in-flight selection. Excluded by [decisions.photographsWithdrawn](../../../decisions.md#photographswithdrawn)
- [product.photoAttribution](../../../local/product/diagram.md#photoattribution) *(excluded)* — Photograph attribution. Excluded by [decisions.photographsWithdrawn](../../../decisions.md#photographswithdrawn)
- [product.photoCaching](../../../local/product/diagram.md#photocaching) *(excluded)* — Local caching of photographs. Excluded by [decisions.photographsWithdrawn](../../../decisions.md#photographswithdrawn)
- [product.comparisonOneScale](../../../local/product/presentation.md#comparisononescale) *(excluded)* — A comparison view holds several birds to one scale. Excluded by [decisions.classicViewRemoved](../../../decisions.md#classicviewremoved)
- [product.pickFromIndex](../../../local/product/presentation.md#pickfromindex) *(excluded)* — Birds are picked for comparison from the index, and the picks are the mode. Excluded by [decisions.classicViewRemoved](../../../decisions.md#classicviewremoved)
- [product.collapsibleControls](../../../local/product/presentation.md#collapsiblecontrols) *(excluded)* — The controls collapse. Excluded by [decisions.classicViewRemoved](../../../decisions.md#classicviewremoved)
- [product.togglesKeepPlace](../../../local/product/presentation.md#toggleskeepplace) *(excluded)* — Toggling a control keeps your place. Excluded by [decisions.classicViewRemoved](../../../decisions.md#classicviewremoved)
- [product.onlyPerGroup](../../../local/product/presentation.md#onlypergroup) *(excluded)* — Every group offers "Only". Excluded by [decisions.classicViewRemoved](../../../decisions.md#classicviewremoved)
- [product.helpfulEmptyState](../../../local/product/presentation.md#helpfulemptystate) *(excluded)* — The empty state says what to do. Excluded by [decisions.classicViewRemoved](../../../decisions.md#classicviewremoved)
- [product.pageExplainsItself](../../../local/product/presentation.md#pageexplainsitself) *(excluded)* — The page explains itself. Excluded by [decisions.classicViewRemoved](../../../decisions.md#classicviewremoved)
- [product.firstLoadShowsAll](../../../local/product/presentation.md#firstloadshowsall) *(excluded)* — First load shows everything. Excluded by [decisions.classicViewRemoved](../../../decisions.md#classicviewremoved)
- [product.searchBox](../../../local/product/filtering.md#searchbox) *(excluded)* — A search box narrows the index as you type. Excluded by [decisions.classicViewRemoved](../../../decisions.md#classicviewremoved)
- [product.scaleToggle](../../../local/product/diagram.md#scaletoggle) *(excluded)* — Scale is a user toggle, defaulting to per-section. Excluded by [decisions.classicViewRemoved](../../../decisions.md#classicviewremoved)
- [product.nowSelectsCurrentMonth](../../../local/product/filtering.md#nowselectscurrentmonth) *(excluded)* — Now selects the current month. Excluded by [decisions.classicViewRemoved](../../../decisions.md#classicviewremoved)
- [product.regionsGroupedTwoWays](../../../local/product/months.md#regionsgroupedtwoways) *(excluded)* — Regions are grouped two ways. Excluded by [decisions.classicViewRemoved](../../../decisions.md#classicviewremoved)
- [product.inaturalistLink](../../../local/product/presentation.md#inaturalistlink) *(excluded)* — The detail view and every tile link to the species on iNaturalist. Excluded by [decisions.ebirdLinkReplacesInaturalist](../../../decisions.md#ebirdlinkreplacesinaturalist)

---

## Composed requirements

## pagesMeetTextContrast

**Every BirdView web page meets WCAG 1.4.6 in full, without the large-text allowance**

Composed from: [wcag.WebPage](../../../imported/w3.org/wcag/wcag.md#webpage), [wcag.FullPages](../../../imported/w3.org/wcag/wcag.md#fullpages), [wcag.ContrastRatio](../../../imported/w3.org/wcag/wcag.md#contrastratio), [wcag.ContrastMinimum](../../../imported/w3.org/wcag/wcag.md#contrastminimum), [wcag.ContrastEnhanced](../../../imported/w3.org/wcag/wcag.md#contrastenhanced), [tech.webPages](../../../local/tech/accessibility.md#webpages),
  [decisions.oneContrastFloor](../../../decisions.md#onecontrastfloor)
Goals: [goals.openToEveryone](../../../../goals/goals.md#opentoeveryone)

Every web page, as [tech.webPages](../../../local/tech/accessibility.md#webpages) defines one, meets [wcag.ContrastEnhanced](../../../imported/w3.org/wcag/wcag.md#contrastenhanced) over every part of the
page ([wcag.FullPages](../../../imported/w3.org/wcag/wcag.md#fullpages)), measured as [wcag.ContrastRatio](../../../imported/w3.org/wcag/wcag.md#contrastratio) defines. All text is held to 7:1, large text
included: [wcag.ContrastEnhanced](../../../imported/w3.org/wcag/wcag.md#contrastenhanced) allows 4.5:1 for large text, and [decisions.oneContrastFloor](../../../decisions.md#onecontrastfloor) gives
that up. Meeting this meets [wcag.ContrastMinimum](../../../imported/w3.org/wcag/wcag.md#contrastminimum) too.

**No page is excepted.** A page that sets no colors, such as a plain-text error response, is drawn
in the reader's own colors and meets this with nothing to measure.

## pagesMeetNonTextContrast

**Every BirdView web page meets WCAG 1.4.11 in full**

Composed from: [wcag.WebPage](../../../imported/w3.org/wcag/wcag.md#webpage), [wcag.FullPages](../../../imported/w3.org/wcag/wcag.md#fullpages), [wcag.NonTextContrast](../../../imported/w3.org/wcag/wcag.md#nontextcontrast), [tech.webPages](../../../local/tech/accessibility.md#webpages)
Goals: [goals.openToEveryone](../../../../goals/goals.md#opentoeveryone)

Every web page meets [wcag.NonTextContrast](../../../imported/w3.org/wcag/wcag.md#nontextcontrast) over every part of the page: the graphics a reader needs
to read a diagram, and the controls and their states, have 3:1 against adjacent colors.

## stylesheetContract

**The stylesheet and the palette carry text contrast for every page that uses only them**

Composed from: [tech.textContrastFloor](../../../local/tech/accessibility.md#textcontrastfloor), [tech.badgeContrast](../../../local/tech/accessibility.md#badgecontrast), [wcag.ContrastRatio](../../../imported/w3.org/wcag/wcag.md#contrastratio)
Goals: [goals.openToEveryone](../../../../goals/goals.md#opentoeveryone)

The stylesheet meets pagesMeetTextContrast for any page that takes its colors from it, so that
requirement is checked once on the stylesheet rather than once per page.

- **Guarantee.** Every text color `styles.css` declares clears 7:1 against both the page and the
  panel background, or against the rule's own background where it sets one, and no link goes
  uncolored ([tech.textContrastFloor](../../../local/tech/accessibility.md#textcontrastfloor)). Every badge's generated background clears 7:1 against the
  text color generated with it ([tech.badgeContrast](../../../local/tech/accessibility.md#badgecontrast)).
- **Assumption.** The page links `styles.css`, has no `<style>` element, and sets a color inline
  only as a badge's generated pair. A page that breaks this is outside the contract and is
  checked on its own.
- **Scope.** Text, in pages rendered from BirdView's templates. **Not** an SVG document served at
  its own URL, which carries its own styles, and **not** non-text contrast.
- **Override.** Closed. A template may not set a color except through the palette; one that does
  has left the contract, which the assumption check reports.

## sourceCreditsAndLicenses

**Every page credits the CC BY and GBIF sources whose output it shows, naming each dataset's license and DOI**

Composed from: [gbifterms.CiteApiDataByDerivedDataset](../../../imported/gbif.org/gbifterms/gbifterms.md#citeapidatabyderiveddataset), [gbifterms.CiteDownloadDoi](../../../imported/gbif.org/gbifterms/gbifterms.md#citedownloaddoi), [gbifterms.PublisherLicencesPrevail](../../../imported/gbif.org/gbifterms/gbifterms.md#publisherlicencesprevail), [tech.localNowPublicLater](../../../local/tech/posture.md#localnowpubliclater),
  [decisions.gbifTermsMet](../../../decisions.md#gbiftermsmet), [decisions.riverV2PackagesAndSplits](../../../decisions.md#riverv2packagesandsplits)
Goals: [goals.goodDataCitizen](../../../../goals/goals.md#gooddatacitizen), [goals.honestData](../../../../goals/goals.md#honestdata)

Every CC BY source whose output appears on a page is credited on that page: AviList, AVONET, and Fox
et al. under CC BY 4.0, including `detail.html` (whose names, families, and species codes come from
AviList) as well as the home page; and the USGS Bird Banding Laboratory as requested. Each GBIF
dataset is credited with its own prevailing license ([gbifterms.PublisherLicencesPrevail](../../../imported/gbif.org/gbifterms/gbifterms.md#publisherlicencesprevail),
[tech.localNowPublicLater](../../../local/tech/posture.md#localnowpubliclater)) — the eBird Observation Dataset under CC BY 4.0 and iNaturalist's
research-grade observations under CC BY-NC 4.0 — wherever its output appears, and in the home page's
credits by publisher, title, DOI as a URL, retrieval date, and "accessed via GBIF.org"
([gbifterms.CiteApiDataByDerivedDataset](../../../imported/gbif.org/gbifterms/gbifterms.md#citeapidatabyderiveddataset), [decisions.gbifTermsMet](../../../decisions.md#gbiftermsmet)). Any GBIF download used is cited by its
own DOI URL ([gbifterms.CiteDownloadDoi](../../../imported/gbif.org/gbifterms/gbifterms.md#citedownloaddoi)).
