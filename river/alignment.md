# Alignment — BirdView

> **Draft, 2026-10-25**, trying River's proposed alignment record on a real project. Not yet part
> of River's method; generated, so regenerate rather than edit.
> Generated from the bundle, the tests' `Criterion:` comments and the history.

Which version of each layer the layer below it is aligned with, since when, and who said so; and,
for the implementation, where each criterion of the targeted requirements stands. A version is a
commit. Search `## ` for the sections.

## Chain

| Layer | Aligned with | Since | Said by | State |
| --- | --- | --- | --- | --- |
| Goals, [`goals/`](goals/goals.md) | Nothing above them: BirdView imports no values. In force: goals at [`dbe250e`](https://github.com/google/birdview/commit/dbe250e) | 2026-10-22 | sameer@, who committed them; no separate review recorded | 14 goals |
| Requirements: [`app`](requirements/composed/internal/app/app.md) | [Goals](goals/goals.md) at [`dbe250e`](https://github.com/google/birdview/commit/dbe250e) | 2026-10-25 | Composed by an agent; approved by Sameer | Published at [`99c0811`](https://github.com/google/birdview/commit/99c0811); [criteria](requirements/composed/internal/app/acceptance.md); `dataset` and `internal/repo` not yet composed |
| Implementation, at the commit this file is in | [`app`](requirements/composed/internal/app/app.md) at [`99c0811`](https://github.com/google/birdview/commit/99c0811) | 2026-10-25 | Generated from the criteria's recorded statuses | 86 of 93 criteria met ([below](#implementation-app-at-99c0811)) |
| Release, [birdview.site](https://birdview.site) | Not recorded here; release guardrails in [`tech/release.md`](requirements/local/tech/release.md) | — | — | The deployed version is recorded only in the private `Sajmani/birdview-deploy` ([bindings](bindings.md)) |

Implementation targets only a published version: `app` at `99c0811`, the commit that approved it.

## Implementation: `app` at `99c0811`

One entry per acceptance criterion that verifies a requirement `app` includes, or that `app` composed.
**Blocks** is *blocks once met*, the default, everywhere; no criterion records otherwise yet. A
criterion is **met since** its first `Watched failing:` line, once its status says verified, and from
then on a regression blocks: *(currently blocking)*. A person has not yet confirmed any of them under
the new rule.

Each lists every requirement it verifies and every test that cites it, nothing elided, so each
link can be followed and checked. `grep -n '^### ' alignment.md` lists the criteria.

### [`acceptance.legibleDistinctColors`](requirements/acceptance.md#legibledistinctcolors)

- Verifies requirements:
  - [`tech.badgeContrast`](requirements/local/tech/accessibility.md#badgecontrast)
  - [`product.generatedBadgeColors`](requirements/local/product/dataset.md#generatedbadgecolors)
  - [`wcag.ContrastEnhanced`](requirements/imported/w3.org/wcag/wcag.md#contrastenhanced)
- Guardrails:
  - [`TestContrastAgainstOwnText`](../internal/palette/palette_test.go#L45)
  - [`TestContrast`](../internal/palette/palette_test.go#L99)
  - [`TestAdjacentFamiliesAreVisiblyDifferent`](../internal/palette/palette_test.go#L192)
- Blocks: blocks once met
- State: met since 2026-09-22 (currently blocking)

### [`acceptance.stableSpeciesId`](requirements/acceptance.md#stablespeciesid)

- Verifies requirements:
  - [`product.englishNameChoice`](requirements/local/product/diagram.md#englishnamechoice)
- Guardrails:
  - [`TestSpeciesCodesAreUniqueAndURLSafe`](../cmd/birdview-ingest/main_test.go#L187)
- Blocks: blocks once met
- State: met since 2026-10-02 (currently blocking)

### [`acceptance.nameLocaleToggle`](requirements/acceptance.md#namelocaletoggle)

- Verifies requirements:
  - [`product.englishNameChoice`](requirements/local/product/diagram.md#englishnamechoice)
- Guardrails:
  - [`TestNameLocaleTogglesNamesAndNothingElse`](../cmd/birdview/main_test.go#L3136)
  - [`TestNameLocaleControlOffersBothAndMarksTheActive`](../cmd/birdview/main_test.go#L3171)
  - [`TestNameFallsBackToWorldWhenNoNorthAmericanName`](../cmd/birdview/main_test.go#L3207)
- Blocks: blocks once met
- State: met since 2026-10-02 (currently blocking)

### [`acceptance.sortsAscending`](requirements/acceptance.md#sortsascending)

- Verifies requirements:
  - [`product.oneSelectedMetric`](requirements/local/product/ordering.md#oneselectedmetric)
  - [`product.ascendingOrder`](requirements/local/product/ordering.md#ascendingorder)
- Guardrails:
  - [`TestSortedAscending`](../internal/view/view_test.go#L55)
- Blocks: blocks once met
- State: passing since 2026-09-22, never recorded failing: not yet met

### [`acceptance.billAndTailOrderings`](requirements/acceptance.md#billandtailorderings)

- Verifies requirements:
  - [`product.oneSelectedMetric`](requirements/local/product/ordering.md#oneselectedmetric)
  - [`product.ascendingOrder`](requirements/local/product/ordering.md#ascendingorder)
  - [`product.scaleToggle`](requirements/local/product/diagram.md#scaletoggle)
- Guardrails:
  - [`TestBillAndTailOrderings`](../cmd/birdview/herenow_test.go#L578)
- Blocks: blocks once met
- State: met since 2026-10-19 (currently blocking)

### [`acceptance.tierHeadingsNameTheOrdering`](requirements/acceptance.md#tierheadingsnametheordering)

- Verifies requirements:
  - [`product.sectionHeadersStateRange`](requirements/local/product/ordering.md#sectionheadersstaterange)
- Guardrails:
  - [`TestTierHeadersNameTheOrdering`](../cmd/birdview/herenow_test.go#L621)
- Blocks: blocks once met
- State: met since 2026-10-19 (currently blocking)

### [`acceptance.tierHeadingsHoldTheirBirds`](requirements/acceptance.md#tierheadingsholdtheirbirds)

- Verifies requirements:
  - [`product.sectionHeadersStateRange`](requirements/local/product/ordering.md#sectionheadersstaterange)
- Guardrails:
  - [`TestTierRangesHoldTheirBirds`](../cmd/birdview/herenow_test.go#L674)
- Blocks: blocks once met
- State: met since 2026-10-19 (currently blocking)

### [`acceptance.tierHeadingDigits`](requirements/acceptance.md#tierheadingdigits)

- Verifies requirements:
  - [`product.sectionHeadersStateRange`](requirements/local/product/ordering.md#sectionheadersstaterange)
- Guardrails:
  - [`TestHeadingDigits`](../cmd/birdview/herenow_test.go#L649)
- Blocks: blocks once met
- State: met since 2026-10-19 (currently blocking)

### [`acceptance.tierCountFollowsTheRange`](requirements/acceptance.md#tiercountfollowstherange)

- Verifies requirements:
  - [`product.sectionHeadersStateRange`](requirements/local/product/ordering.md#sectionheadersstaterange)
- Guardrails:
  - [`TestTierCountFollowsTheRange`](../cmd/birdview/herenow_test.go#L728)
- Blocks: blocks once met
- State: met since 2026-10-19 (currently blocking)

### [`acceptance.sectionBoundariesObeyRule`](requirements/acceptance.md#sectionboundariesobeyrule)

- Verifies requirements:
  - [`product.sectionBreakRule`](requirements/local/product/ordering.md#sectionbreakrule)
  - [`product.sectionThresholdDefault`](requirements/local/product/ordering.md#sectionthresholddefault)
- Guardrails:
  - [`TestSectionRule`](../internal/view/view_test.go#L80)
  - [`TestSectionsPartition`](../internal/view/view_test.go#L112)
- Blocks: blocks once met
- State: met since 2026-09-22 (currently blocking)

### [`acceptance.sectionHeadersMatch`](requirements/acceptance.md#sectionheadersmatch)

- Verifies requirements:
  - [`product.sectionHeadersStateRange`](requirements/local/product/ordering.md#sectionheadersstaterange)
- Guardrails:
  - [`TestSectionRangeMatchesContents`](../internal/view/view_test.go#L134)
- Blocks: blocks once met
- State: passing since 2026-09-22, never recorded failing: not yet met

### [`acceptance.segmentsSumExactly`](requirements/acceptance.md#segmentssumexactly)

- Verifies requirements:
  - [`product.segmentsSumToLength`](requirements/local/product/diagram.md#segmentssumtolength)
- Guardrails:
  - [`TestSegmentsSumToTotalExactly`](../cmd/birdview-ingest/main_test.go#L226)
  - [`TestSegmentsSumExactly`](../internal/morphology/morphology_test.go#L36)
- Blocks: blocks once met
- State: met since 2026-09-22 (currently blocking)

### [`acceptance.sharedScale`](requirements/acceptance.md#sharedscale)

- Verifies requirements:
  - [`product.scaleToggle`](requirements/local/product/diagram.md#scaletoggle)
- Guardrails:
  - [`TestDefaultScaleIsPerSection`](../cmd/birdview/main_test.go#L758)
  - [`TestScaleModes`](../internal/view/view_test.go#L155)
  - [`TestLargerBirdsDrawLarger`](../internal/view/view_test.go#L219)
  - [`TestSharedScaleWithinSection`](../internal/view/view_test.go#L270)
- Blocks: blocks once met
- State: met since 2026-09-22 (currently blocking)

### [`acceptance.crossbarToScale`](requirements/acceptance.md#crossbartoscale)

- Verifies requirements:
  - [`product.crossbarAtVerticalScale`](requirements/local/product/diagram.md#crossbaratverticalscale)
- Guardrails:
  - [`TestCrossbarIsToScale`](../internal/view/view_test.go#L197)
- Blocks: blocks once met
- State: met since 2026-09-22 (currently blocking)

### [`acceptance.diagramPartsInOrder`](requirements/acceptance.md#diagrampartsinorder)

- Verifies requirements:
  - [`product.segmentComposition`](requirements/local/product/diagram.md#segmentcomposition)
- Guardrails:
  - none cited
- Blocks: blocks once met
- State: passing since 2026-09-22, never recorded failing: not yet met

### [`acceptance.proportionsDifferentiate`](requirements/acceptance.md#proportionsdifferentiate)

- Verifies requirements:
  - [`product.proportionDifferentiates`](requirements/local/product/diagram.md#proportiondifferentiates)
- Guardrails:
  - [`TestShapeDifferentiation`](../internal/morphology/morphology_test.go#L100)
- Blocks: blocks once met
- State: not yet met (planned)

### [`acceptance.tileMarksNonMeasured`](requirements/acceptance.md#tilemarksnonmeasured)

- Verifies requirements:
  - [`product.tileMarksModeled`](requirements/local/product/measurements.md#tilemarksmodeled)
- Guardrails:
  - none cited
- Blocks: blocks once met
- State: not yet met (planned)

### [`acceptance.detailShowsProvenance`](requirements/acceptance.md#detailshowsprovenance)

- Verifies requirements:
  - [`product.metricProvenanceExplained`](requirements/local/product/measurements.md#metricprovenanceexplained)
  - [`product.tileOpensDetail`](requirements/local/product/presentation.md#tileopensdetail)
- Guardrails:
  - [`TestDetailShowsEverySegmentWithProvenance`](../cmd/birdview/main_test.go#L567)
- Blocks: blocks once met
- State: met since 2026-09-22 (currently blocking)

### [`acceptance.totalIsSumOfSegments`](requirements/acceptance.md#totalissumofsegments)

- Verifies requirements:
  - [`product.lengthAndWingspanSources`](requirements/local/product/measurements.md#lengthandwingspansources)
  - [`product.provenancePerMetric`](requirements/local/product/measurements.md#provenancepermetric)
  - [`product.segmentsSumToLength`](requirements/local/product/diagram.md#segmentssumtolength)
- Guardrails:
  - [`TestTotalLengthSaysItIsASum`](../cmd/birdview-ingest/main_test.go#L340)
- Blocks: blocks once met
- State: passing since 2026-09-22, never recorded failing: not yet met

### [`acceptance.noServerNetworkRequest`](requirements/acceptance.md#noservernetworkrequest)

- Verifies requirements:
  - [`tech.selfContainedServer`](requirements/local/tech/delivery.md#selfcontainedserver)
- Guardrails:
  - [`TestNoExternalResources`](../cmd/birdview/main_test.go#L403)
- Blocks: blocks once met
- State: met since 2026-09-22 (currently blocking)

### [`acceptance.bothCommandsBuild`](requirements/acceptance.md#bothcommandsbuild)

- Verifies requirements:
  - [`tech.twoCommandsOneModule`](requirements/local/tech/delivery.md#twocommandsonemodule)
  - [`tech.servesOnLocalhost`](requirements/local/tech/delivery.md#servesonlocalhost)
- Guardrails:
  - [`TestRoutesServe`](../cmd/birdview/main_test.go#L365)
- Blocks: blocks once met
- State: passing since 2026-09-22, never recorded failing: not yet met

### [`acceptance.tileCarriesHeaderAndMetrics`](requirements/acceptance.md#tilecarriesheaderandmetrics)

- Verifies requirements:
  - [`product.tileContents`](requirements/local/product/presentation.md#tilecontents)
- Guardrails:
  - [`TestTileShowsHeaderAndMetrics`](../cmd/birdview/main_test.go#L414)
  - [`TestDiagramIsVerticallyCentered`](../cmd/birdview/main_test.go#L539)
- Blocks: blocks once met
- State: met since 2026-10-01 (currently blocking)

### [`acceptance.familyFilterSelectsAndClears`](requirements/acceptance.md#familyfilterselectsandclears)

- Verifies requirements:
  - [`product.familyFilter`](requirements/local/product/filtering.md#familyfilter)
- Guardrails:
  - [`TestFamilyFilter`](../cmd/birdview/main_test.go#L593)
  - [`TestOrderToggles`](../cmd/birdview/main_test.go#L977)
  - [`TestAllToggle`](../cmd/birdview/main_test.go#L1039)
  - [`TestDefaultSelectsEverything`](../cmd/birdview/main_test.go#L1075)
- Blocks: blocks once met
- State: met since 2026-09-22 (currently blocking)

### [`acceptance.unitToggleSwitches`](requirements/acceptance.md#unittoggleswitches)

- Verifies requirements:
  - [`product.imperialByDefault`](requirements/local/product/presentation.md#imperialbydefault)
- Guardrails:
  - [`TestUnitToggle`](../cmd/birdview/main_test.go#L685)
  - [`TestMassUnitsBySystem`](../cmd/birdview/main_test.go#L729)
  - [`TestNearMatchTooltipNamesTheDonor`](../cmd/birdview/main_test.go#L4411)
- Blocks: blocks once met
- State: met since 2026-09-22 (currently blocking)

### [`acceptance.detailDrawsFullScale`](requirements/acceptance.md#detaildrawsfullscale)

- Verifies requirements:
  - [`product.tileOpensDetail`](requirements/local/product/presentation.md#tileopensdetail)
- Guardrails:
  - [`TestDetailDrawsFullScale`](../cmd/birdview/main_test.go#L778)
- Blocks: blocks once met
- State: met since 2026-09-22 (currently blocking)

### [`acceptance.pageExplainsControls`](requirements/acceptance.md#pageexplainscontrols)

- Verifies requirements:
  - [`product.pageExplainsItself`](requirements/local/product/presentation.md#pageexplainsitself)
- Guardrails:
  - [`TestHelpCoversEveryControl`](../cmd/birdview/main_test.go#L807)
  - [`TestEveryControlEntryIsWrapped`](../cmd/birdview/main_test.go#L1715)
  - [`TestHeaderNamesThePanelsThatExist`](../cmd/birdview/main_test.go#L1775)
  - [`TestControlHelpEntriesAreUnbreakable`](../web/contrast_test.go#L352)
- Blocks: blocks once met
- State: met since 2026-09-27 (currently blocking)

### [`acceptance.badgeTextInheritsColor`](requirements/acceptance.md#badgetextinheritscolor)

- Verifies requirements:
  - [`product.generatedBadgeColors`](requirements/local/product/dataset.md#generatedbadgecolors)
  - [`tech.badgeContrast`](requirements/local/tech/accessibility.md#badgecontrast)
- Guardrails:
  - [`TestChipTextInheritsItsBackgroundsTextColor`](../cmd/birdview/main_test.go#L912)
- Blocks: blocks once met
- State: met since 2026-09-22 (currently blocking)

### [`acceptance.emptyStateGuides`](requirements/acceptance.md#emptystateguides)

- Verifies requirements:
  - [`product.helpfulEmptyState`](requirements/local/product/presentation.md#helpfulemptystate)
  - [`product.familyFilter`](requirements/local/product/filtering.md#familyfilter)
- Guardrails:
  - [`TestClearingEverythingShowsGuidance`](../cmd/birdview/main_test.go#L1105)
- Blocks: blocks once met
- State: met since 2026-09-22 (currently blocking)

### [`acceptance.countLineNamesEverySetting`](requirements/acceptance.md#countlinenameseverysetting)

- Verifies requirements:
  - [`product.helpfulEmptyState`](requirements/local/product/presentation.md#helpfulemptystate)
- Guardrails:
  - [`TestFiltersAreNamedOnTheCountLine`](../cmd/birdview/main_test.go#L3693)
  - [`TestCountLineNamesEverySetting`](../cmd/birdview/main_test.go#L6962)
- Blocks: blocks once met
- State: met since 2026-10-03 (currently blocking)

### [`acceptance.nothingMatchedOffersAWayOut`](requirements/acceptance.md#nothingmatchedoffersawayout)

- Verifies requirements:
  - [`product.helpfulEmptyState`](requirements/local/product/presentation.md#helpfulemptystate)
- Guardrails:
  - [`TestEmptyStateNamesTheFiltersThatEmptiedIt`](../cmd/birdview/main_test.go#L3228)
  - [`TestClearTheFiltersClearsThem`](../cmd/birdview/main_test.go#L3742)
- Blocks: blocks once met
- State: met since 2026-10-02 (currently blocking)

### [`acceptance.scrollScriptScoped`](requirements/acceptance.md#scrollscriptscoped)

- Verifies requirements:
  - [`product.togglesKeepPlace`](requirements/local/product/presentation.md#toggleskeepplace)
  - [`tech.selfContainedServer`](requirements/local/tech/delivery.md#selfcontainedserver)
- Guardrails:
  - [`TestScrollRestorationIsInlineAndSelfContained`](../cmd/birdview/main_test.go#L1147)
  - [`TestCollapsiblePanelsPersist`](../cmd/birdview/main_test.go#L1188)
- Blocks: blocks once met
- State: met since 2026-10-04 (currently blocking)

### [`acceptance.controlsMatchServer`](requirements/acceptance.md#controlsmatchserver)

- Verifies requirements:
  - [`product.sectionThresholdDefault`](requirements/local/product/ordering.md#sectionthresholddefault)
  - [`product.scaleToggle`](requirements/local/product/diagram.md#scaletoggle)
- Guardrails:
  - [`TestControlOrderAndOptions`](../cmd/birdview/main_test.go#L1227)
  - [`TestUnitButtonsNameTheUnitsRendered`](../cmd/birdview/main_test.go#L4168)
- Blocks: blocks once met
- State: met since 2026-09-22 (currently blocking)

### [`acceptance.foldedNeckSameStroke`](requirements/acceptance.md#foldednecksamestroke)

- Verifies requirements:
  - [`product.segmentComposition`](requirements/local/product/diagram.md#segmentcomposition)
  - [`product.segmentsSumToLength`](requirements/local/product/diagram.md#segmentssumtolength)
  - [`product.tileOpensDetail`](requirements/local/product/presentation.md#tileopensdetail)
- Guardrails:
  - [`TestFoldedNeckRendersAsAnArc`](../cmd/birdview/main_test.go#L1368)
  - [`TestFoldedNeckIsTheSameStrokeInLessHeight`](../internal/view/view_test.go#L321)
  - [`TestFoldedNeckDoesNotChangeOrdering`](../internal/view/view_test.go#L378)
- Blocks: blocks once met
- State: met since 2026-09-23 (currently blocking)

### [`acceptance.inaturalistLinkConfidentOnly`](requirements/acceptance.md#inaturalistlinkconfidentonly)

- Verifies requirements:
  - [`product.inaturalistLink`](requirements/local/product/presentation.md#inaturalistlink)
  - [`tech.pacedInaturalistMapping`](requirements/local/tech/process.md#pacedinaturalistmapping)
- Guardrails:
  - [`TestINatLinkOnlyForAConfidentMatch`](../cmd/birdview/main_test.go#L1422)
  - [`TestNoINaturalistImagesAreEmbedded`](../cmd/birdview/main_test.go#L1459)
  - [`TestINatTableMatchesTheDataset`](../cmd/birdview-ingest/main_test.go#L506)
  - [`TestAcceptRequiresBothNames`](../internal/inat/inat_test.go#L36)
  - [`TestAcceptUsesEveryEnglishName`](../internal/inat/inat_test.go#L84)
  - [`TestTableRoundTripsAndRejectsJunk`](../internal/inat/inat_test.go#L102)
  - [`TestTaxonURL`](../internal/inat/inat_test.go#L150)
  - [`TestLookupHonorsTheSourceRequirements`](../internal/inat/inat_test.go#L184)
  - [`TestLookupRejectsAWrongBird`](../internal/inat/inat_test.go#L222)
  - [`TestLookupReportsTransportErrorsSeparately`](../internal/inat/inat_test.go#L254)
- Blocks: blocks once met
- State: met since 2026-09-27 (currently blocking)

### [`acceptance.foldingFamiliesHaveNecks`](requirements/acceptance.md#foldingfamilieshavenecks)

- Verifies requirements:
  - [`product.segmentComposition`](requirements/local/product/diagram.md#segmentcomposition)
  - [`product.neckModelPerFamily`](requirements/local/product/diagram.md#neckmodelperfamily)
- Guardrails:
  - [`TestEveryFoldedFamilyHasALongNeck`](../internal/view/view_test.go#L427)
- Blocks: blocks once met
- State: met since 2026-09-23 (currently blocking)

### [`acceptance.noTextBelowContrastFloor`](requirements/acceptance.md#notextbelowcontrastfloor)

- Verifies requirements:
  - [`tech.textContrastFloor`](requirements/local/tech/accessibility.md#textcontrastfloor)
  - [`wcag.ContrastEnhanced`](requirements/imported/w3.org/wcag/wcag.md#contrastenhanced)
- Guardrails:
  - [`TestStylesheetContrast`](../web/contrast_test.go#L87)
  - [`TestSVGTextContrast`](../web/contrast_test.go#L188)
  - [`TestNoLinkUsesTheBrowserDefault`](../web/contrast_test.go#L266)
- Blocks: blocks once met
- State: met since 2026-09-23 (currently blocking)

### [`acceptance.ownRenderableFavicon`](requirements/acceptance.md#ownrenderablefavicon)

- Verifies requirements:
  - [`tech.selfContainedServer`](requirements/local/tech/delivery.md#selfcontainedserver)
- Guardrails:
  - [`TestFaviconIsServedLocally`](../cmd/birdview/main_test.go#L1482)
- Blocks: blocks once met
- State: met since 2026-09-23 (currently blocking)

### [`acceptance.diagramDistinguishesModeled`](requirements/acceptance.md#diagramdistinguishesmodeled)

- Verifies requirements:
  - [`product.measuredVersusModeled`](requirements/local/product/measurements.md#measuredversusmodeled)
  - [`product.segmentComposition`](requirements/local/product/diagram.md#segmentcomposition)
- Guardrails:
  - [`TestCrossbarSplitsAtTheChord`](../cmd/birdview/main_test.go#L1532)
  - [`TestBothPagesExplainTheStrokeCode`](../cmd/birdview/main_test.go#L1576)
  - [`TestChordTickMarksTheMeasuredPortion`](../internal/view/view_test.go#L450)
  - [`TestStrokeStyleCarriesProvenance`](../web/contrast_test.go#L290)
- Blocks: blocks once met
- State: met since 2026-09-24 (currently blocking)

### [`acceptance.foldedFigureExplained`](requirements/acceptance.md#foldedfigureexplained)

- Verifies requirements:
  - [`product.labeledDetailDiagram`](requirements/local/product/measurements.md#labeleddetaildiagram)
- Guardrails:
  - [`TestFoldedFigureExplained`](../cmd/birdview/main_test.go#L7005)
- Blocks: blocks once met
- State: met since 2026-10-22 (currently blocking)

### [`acceptance.segmentLabelsDontCollide`](requirements/acceptance.md#segmentlabelsdontcollide)

- Verifies requirements:
  - [`product.labeledDetailDiagram`](requirements/local/product/measurements.md#labeleddetaildiagram)
- Guardrails:
  - [`TestDetailDiagramIsLabeled`](../cmd/birdview/main_test.go#L1598)
  - [`TestWingLabelsClearTheColumn`](../cmd/birdview/main_test.go#L1677)
  - [`TestDiagramTextPaintsOverLines`](../cmd/birdview/main_test.go#L1744)
  - [`TestEveryReportedMetricIsLabelled`](../cmd/birdview/main_test.go#L2834)
  - [`TestDetailLabelsDoNotCollide`](../internal/view/view_test.go#L489)
- Blocks: blocks once met
- State: met since 2026-09-28 (currently blocking)

### [`acceptance.totalLengthLabeled`](requirements/acceptance.md#totallengthlabeled)

- Verifies requirements:
  - [`product.labeledDetailDiagram`](requirements/local/product/measurements.md#labeleddetaildiagram)
- Guardrails:
  - [`TestDetailDiagramLabelsTotalLength`](../cmd/birdview/main_test.go#L1818)
  - [`TestTotalLengthLabelStaysInTheBox`](../cmd/birdview/main_test.go#L1898)
- Blocks: blocks once met
- State: met since 2026-09-28 (currently blocking)

### [`acceptance.familyListWraps`](requirements/acceptance.md#familylistwraps)

- Verifies requirements:
  - [`product.familyFilter`](requirements/local/product/filtering.md#familyfilter)
- Guardrails:
  - [`TestEveryOrderGroupWrapsItsFamilies`](../cmd/birdview/main_test.go#L1964)
  - [`TestOrderGroupsIndentWrappedFamilies`](../web/contrast_test.go#L382)
- Blocks: blocks once met
- State: met since 2026-09-28 (currently blocking)

### [`acceptance.comparisonSharesScale`](requirements/acceptance.md#comparisonsharesscale)

- Verifies requirements:
  - [`product.comparisonOneScale`](requirements/local/product/presentation.md#comparisononescale)
- Guardrails:
  - [`TestCompareSharesOneScale`](../cmd/birdview/main_test.go#L2066)
  - [`TestCompareSelectsAndExplainsItself`](../cmd/birdview/main_test.go#L2146)
  - [`TestCompareBackLinkKeepsTheComparison`](../cmd/birdview/main_test.go#L2238)
  - [`TestComparePanelsWrap`](../web/contrast_test.go#L429)
- Blocks: blocks once met
- State: met since 2026-09-30 (currently blocking)

### [`acceptance.pickModeSurvivesNavigation`](requirements/acceptance.md#pickmodesurvivesnavigation)

- Verifies requirements:
  - [`product.pickFromIndex`](requirements/local/product/presentation.md#pickfromindex)
  - [`product.tileContents`](requirements/local/product/presentation.md#tilecontents)
- Guardrails:
  - [`TestTileLinksCarryViewState`](../cmd/birdview/main_test.go#L2285)
  - [`TestPickModeIsTheSpeciesParameter`](../cmd/birdview/main_test.go#L2365)
  - [`TestPickingIsReversibleAndKeepsTheFilter`](../cmd/birdview/main_test.go#L2411)
  - [`TestComparisonKeepsTheFilter`](../cmd/birdview/main_test.go#L2468)
  - [`TestPicksSurviveTheDetailRoundTrip`](../cmd/birdview/main_test.go#L2508)
  - [`TestOnlyClearLeavesPickMode`](../cmd/birdview/main_test.go#L2533)
  - [`TestUnrecognisedPicksAreDroppedAndNamed`](../cmd/birdview/main_test.go#L2598)
  - [`TestDetailPagePickStates`](../cmd/birdview/main_test.go#L2641)
  - [`TestCompareBirdsToggle`](../cmd/birdview/main_test.go#L2721)
  - [`TestPickLinksAnchorAtTheirTile`](../cmd/birdview/main_test.go#L2798)
  - [`TestComparisonOffersAWayOutOfPickMode`](../cmd/birdview/main_test.go#L6535)
- Blocks: blocks once met
- State: met since 2026-10-01 (currently blocking)

### [`acceptance.massDrawnAsWaterCube`](requirements/acceptance.md#massdrawnaswatercube)

- Verifies requirements:
  - [`product.labeledDetailDiagram`](requirements/local/product/measurements.md#labeleddetaildiagram)
  - [`product.pageExplainsItself`](requirements/local/product/presentation.md#pageexplainsitself)
- Guardrails:
  - [`TestMassCubeIsWaterToScale`](../cmd/birdview/main_test.go#L2897)
  - [`TestMassCubeIsRenderedAndExplained`](../cmd/birdview/main_test.go#L2960)
  - [`TestTileCubeNeverReachesTheCrossbar`](../cmd/birdview/main_test.go#L4231)
- Blocks: blocks once met
- State: met since 2026-10-01 (currently blocking)

### [`acceptance.comparisonSortRewritesUrl`](requirements/acceptance.md#comparisonsortrewritesurl)

- Verifies requirements:
  - [`product.comparisonOneScale`](requirements/local/product/presentation.md#comparisononescale)
- Guardrails:
  - [`TestCompareSortReordersTheURL`](../cmd/birdview/main_test.go#L3033)
- Blocks: blocks once met
- State: met since 2026-10-01 (currently blocking)

### [`acceptance.faviconFromPalette`](requirements/acceptance.md#faviconfrompalette)

- Verifies requirements:
  - [`product.oneDiagramVocabulary`](requirements/local/product/diagram.md#onediagramvocabulary)
- Guardrails:
  - [`diagramParts`](../web/favicon_test.go#L42)
- Blocks: blocks once met
- State: met since 2026-10-02 (currently blocking)

### [`acceptance.tileHeaderNeverWraps`](requirements/acceptance.md#tileheaderneverwraps)

- Verifies requirements:
  - [`product.tileContents`](requirements/local/product/presentation.md#tilecontents)
- Guardrails:
  - [`rule`](../web/tilewrap_test.go#L38)
  - [`TestBandingCodeIsOutsideTheTruncatingElement`](../web/tilewrap_test.go#L83)
- Blocks: blocks once met
- State: met since 2026-10-02 (currently blocking)

### [`acceptance.everyControlDocumented`](requirements/acceptance.md#everycontroldocumented)

- Verifies requirements:
  - [`product.pageExplainsItself`](requirements/local/product/presentation.md#pageexplainsitself)
- Guardrails:
  - [`TestEveryControlIsDocumented`](../cmd/birdview/main_test.go#L3289)
- Blocks: blocks once met
- State: met since 2026-10-02 (currently blocking)

### [`acceptance.regionFilterBehaves`](requirements/acceptance.md#regionfilterbehaves)

- Verifies requirements:
  - [`product.regionAndMonthFilter`](requirements/local/product/filtering.md#regionandmonthfilter)
  - [`product.filtersCompose`](requirements/local/product/months.md#filterscompose)
- Guardrails:
  - [`TestRegionFilter`](../cmd/birdview/main_test.go#L3401)
  - [`TestRegionSelectorIsGroupedWithPelagicLast`](../cmd/birdview/main_test.go#L3468)
- Blocks: blocks once met
- State: met since 2026-10-03 (currently blocking)

### [`acceptance.zonesAnswerToData`](requirements/acceptance.md#zonesanswertodata)

- Verifies requirements:
  - [`product.regionsGroupedTwoWays`](requirements/local/product/months.md#regionsgroupedtwoways)
- Guardrails:
  - none cited
- Blocks: blocks once met
- State: met since 2026-10-03 (currently blocking)

### [`acceptance.regionUrlRoundTrips`](requirements/acceptance.md#regionurlroundtrips)

- Verifies requirements:
  - [`product.regionAndMonthFilter`](requirements/local/product/filtering.md#regionandmonthfilter)
- Guardrails:
  - [`TestRegionURLsCarrySlugsNotGADMIDs`](../cmd/birdview/main_test.go#L3529)
  - [`TestSlugsAreUniqueAndResolve`](../internal/region/region_test.go#L246)
  - [`TestSlugIsCodeOrPascalName`](../internal/region/region_test.go#L289)
  - [`TestGADMIDsAreNotSlugs`](../internal/region/region_test.go#L323)
- Blocks: blocks once met
- State: met since 2026-10-03 (currently blocking)

### [`acceptance.familyGroupingComplete`](requirements/acceptance.md#familygroupingcomplete)

- Verifies requirements:
  - [`product.familiesByOrder`](requirements/local/product/months.md#familiesbyorder)
  - [`product.tileContents`](requirements/local/product/presentation.md#tilecontents)
- Guardrails:
  - [`TestTileCarriesExactlyOneBadge`](../cmd/birdview/main_test.go#L3654)
- Blocks: blocks once met
- State: met since 2026-10-03 (currently blocking)

### [`acceptance.creditsComplete`](requirements/acceptance.md#creditscomplete)

- Verifies requirements:
  - [`product.selfDescriptionStaysTrue`](requirements/local/product/presentation.md#selfdescriptionstaystrue)
- Guardrails:
  - [`TestEveryPageCreditsWhatItShows`](../web/docs_test.go#L113)
  - [`TestDocsCountsMatchTheDataset`](../web/docs_test.go#L146)
  - [`TestNoDocCreditsASourceForWhatItNoLongerSupplies`](../web/docs_test.go#L180)
- Blocks: blocks once met
- State: met since 2026-10-03 (currently blocking)

### [`acceptance.categoryOriginShown`](requirements/acceptance.md#categoryoriginshown)

- Verifies requirements:
  - [`product.categoryProvenance`](requirements/local/product/measurements.md#categoryprovenance)
- Guardrails:
  - [`TestDetailListsEveryCategoryWithItsOrigin`](../cmd/birdview/main_test.go#L3778)
  - [`TestServedDocumentsHaveNoBrokenLinks`](../cmd/birdview/main_test.go#L3885)
  - [`TestVendoredTermsAreNotServed`](../cmd/birdview/main_test.go#L3969)
  - [`TestDocumentTitlesAreHeadings`](../cmd/birdview/main_test.go#L3991)
  - [`TestAddedListsEveryUnpublishedAssignment`](../cmd/birdview/main_test.go#L4014)
  - [`TestCitedAnchorsResolve`](../cmd/birdview/main_test.go#L4452)
- Blocks: blocks once met
- State: met since 2026-10-03 (currently blocking)

### [`acceptance.habitatDietFilterPerSpecies`](requirements/acceptance.md#habitatdietfilterperspecies)

- Verifies requirements:
  - [`product.habitatAndDietFilters`](requirements/local/product/months.md#habitatanddietfilters)
- Guardrails:
  - [`TestTraitAndFamilyPanels`](../cmd/birdview/main_test.go#L3577)
  - [`TestSupplementaryKeysExist`](../internal/trait/trait_test.go#L54)
  - [`TestSupplementaryReachesTheDataset`](../internal/trait/trait_test.go#L78)
  - [`TestMixedDietRuleReachesTheDataset`](../internal/trait/trait_test.go#L116)
  - [`TestEveryAssignmentHasAnOriginAndTheVocabularyKnowsIt`](../internal/trait/trait_test.go#L150)
  - [`TestLabelsAreNotJargon`](../internal/trait/trait_test.go#L202)
  - [`TestCoinedTermsAreRareAndSaySo`](../internal/trait/trait_test.go#L233)
  - [`TestCoinedTermsAreNeverPublished`](../internal/trait/trait_test.go#L259)
- Blocks: blocks once met
- State: met since 2026-10-03 (currently blocking)

### [`acceptance.staticServesAssetsOnly`](requirements/acceptance.md#staticservesassetsonly)

- Verifies requirements:
  - [`tech.servesOnlyItsPages`](requirements/local/tech/delivery.md#servesonlyitspages)
- Guardrails:
  - [`TestStaticServesAssetsOnly`](../cmd/birdview/main_test.go#L6942)
- Blocks: blocks once met
- State: met since 2026-10-16 (currently blocking)

### [`acceptance.robotsKeepsCrawlersToPages`](requirements/acceptance.md#robotskeepscrawlerstopages)

- Verifies requirements:
  - [`tech.crawlersSkipFilteredViews`](requirements/local/tech/delivery.md#crawlersskipfilteredviews)
  - [`tech.servesOnlyItsPages`](requirements/local/tech/delivery.md#servesonlyitspages)
- Guardrails:
  - [`TestRobotsKeepsCrawlersToPages`](../cmd/birdview/robots_test.go#L95)
- Blocks: blocks once met
- State: met since 2026-10-23 (currently blocking)

### [`acceptance.stableUrlsKeepWorking`](requirements/acceptance.md#stableurlskeepworking)

- Verifies requirements:
  - [`tech.stableUrls`](requirements/local/tech/delivery.md#stableurls)
- Guardrails:
  - [`TestStableUrlsKeepWorking`](../cmd/birdview/stable_test.go#L45)
- Blocks: blocks once met
- State: met since 2026-10-23 (currently blocking)

### [`acceptance.canonicalHostRedirects`](requirements/acceptance.md#canonicalhostredirects)

- Verifies requirements:
  - [`tech.oneCanonicalHost`](requirements/local/tech/delivery.md#onecanonicalhost)
- Guardrails:
  - [`TestCanonicalHostRedirects`](../cmd/birdview/stable_test.go#L81)
- Blocks: blocks once met
- State: met since 2026-10-23 (currently blocking)

### [`acceptance.onlyNarrowsOneGroup`](requirements/acceptance.md#onlynarrowsonegroup)

- Verifies requirements:
  - [`product.onlyPerGroup`](requirements/local/product/presentation.md#onlypergroup)
- Guardrails:
  - [`TestOnlyNarrowsToOneGroup`](../cmd/birdview/main_test.go#L4077)
  - [`TestLumpsMergeIntoTheSpeciesBirdViewCarries`](../cmd/birdview-status/main_test.go#L106)
- Blocks: blocks once met
- State: met since 2026-10-04 (currently blocking)

### [`acceptance.massInOneUnit`](requirements/acceptance.md#massinoneunit)

- Verifies requirements:
  - [`product.oneUnitInComparison`](requirements/local/product/presentation.md#oneunitincomparison)
- Guardrails:
  - [`TestUnitsForTakesTheSmallestBird`](../cmd/birdview/main_test.go#L4318)
  - [`TestNoSectionMixesMassUnits`](../cmd/birdview/main_test.go#L4363)
- Blocks: blocks once met
- State: met since 2026-10-04 (currently blocking)

### [`acceptance.commonestShareIsShare`](requirements/acceptance.md#commonestshareisshare)

- Verifies requirements:
  - [`product.shareOfRecords`](requirements/local/product/presentation.md#shareofrecords)
  - [`product.categoryProvenance`](requirements/local/product/measurements.md#categoryprovenance)
- Guardrails:
  - [`TestTileShareFollowsTheRegionFilter`](../cmd/birdview/main_test.go#L4523)
  - [`TestRegionCountsLineUp`](../cmd/birdview-ingest/main_test.go#L578)
- Blocks: blocks once met
- State: met since 2026-10-05 (currently blocking)

### [`acceptance.noContactAddressPublished`](requirements/acceptance.md#nocontactaddresspublished)

- Verifies requirements:
  - [`product.noPersonalAddress`](requirements/local/product/presentation.md#nopersonaladdress)
- Guardrails:
  - [`TestNoPersonalAddressIsPublished`](../cmd/birdview/main_test.go#L4653)
- Blocks: blocks once met
- State: met since 2026-10-05 (currently blocking)

### [`acceptance.outOfScopeExplained`](requirements/acceptance.md#outofscopeexplained)

- Verifies requirements:
  - [`product.undrawnSpeciesAnswer`](requirements/local/product/presentation.md#undrawnspeciesanswer)
  - [`decisions.speciesUrlOutlivesSpecies`](requirements/decisions.md#speciesurloutlivesspecies)
- Guardrails:
  - [`TestUndrawnSpeciesNamesItselfAt404`](../cmd/birdview/main_test.go#L4740)
  - [`TestNoIDIsBothDrawnAndNot`](../cmd/birdview/main_test.go#L4829)
  - [`TestAbsentSpeciesCarryNoEvidence`](../cmd/birdview/main_test.go#L4865)
- Blocks: blocks once met
- State: met since 2026-10-05 (currently blocking)

### [`acceptance.commonestIsLargestShare`](requirements/acceptance.md#commonestislargestshare)

- Verifies requirements:
  - [`product.shareOfRecords`](requirements/local/product/presentation.md#shareofrecords)
  - [`decisions.seaShareDiffers`](requirements/decisions.md#seasharediffers)
- Guardrails:
  - [`TestCommonestIsTheLargestShare`](../cmd/birdview/main_test.go#L4912)
  - [`TestShareRangesOverOneSetOfRegions`](../cmd/birdview/main_test.go#L4959)
  - [`TestShareStatesZeroRatherThanDashing`](../cmd/birdview/main_test.go#L5043)
  - [`TestMissingTotalsDoNotClaimTheBirdIsAbsent`](../cmd/birdview/main_test.go#L5180)
  - [`TestInterfaceNeverSaysShare`](../cmd/birdview/main_test.go#L5248)
- Blocks: blocks once met
- State: met since 2026-10-05 (currently blocking)

### [`acceptance.peakAndShareOneCalculation`](requirements/acceptance.md#peakandshareonecalculation)

- Verifies requirements:
  - [`product.shareOfRecords`](requirements/local/product/presentation.md#shareofrecords)
  - [`decisions.honestSeaDenominator`](requirements/decisions.md#honestseadenominator)
- Guardrails:
  - [`TestCommonestRoundTripsThroughTheIndex`](../cmd/birdview/main_test.go#L5097)
- Blocks: blocks once met
- State: met since 2026-10-05 (currently blocking)

### [`acceptance.responsesCompress`](requirements/acceptance.md#responsescompress)

- Verifies requirements:
  - [`tech.compressedResponses`](requirements/local/tech/delivery.md#compressedresponses)
- Guardrails:
  - [`TestResponsesCompressWhenAsked`](../cmd/birdview/main_test.go#L5283)
- Blocks: blocks once met
- State: met since 2026-10-05 (currently blocking)

### [`acceptance.frequencyTiers`](requirements/acceptance.md#frequencytiers)

- Verifies requirements:
  - [`product.orderByFrequency`](requirements/local/product/presentation.md#orderbyfrequency)
  - [`decisions.frequencyMetric`](requirements/decisions.md#frequencymetric)
- Guardrails:
  - [`TestFrequencyOrdersAndBandsByDecade`](../cmd/birdview/main_test.go#L5389)
  - [`TestFrequencyTiersAreNotNamed`](../cmd/birdview/main_test.go#L5559)
  - [`TestFrequencyDefaultsToGlobalScale`](../cmd/birdview/main_test.go#L5590)
  - [`TestOrderByStatesItsDirection`](../cmd/birdview/main_test.go#L5737)
  - [`TestOrderingRowAlignsOnTheBaseline`](../cmd/birdview/main_test.go#L5793)
- Blocks: blocks once met
- State: met since 2026-10-05 (currently blocking)

### [`acceptance.sharedCountDisclosed`](requirements/acceptance.md#sharedcountdisclosed)

- Verifies requirements:
  - [`product.namesRecordsFiledUnder`](requirements/local/product/presentation.md#namesrecordsfiledunder)
  - [`decisions.sharedCountShownOnBoth`](requirements/decisions.md#sharedcountshownonboth)
- Guardrails:
  - [`TestSharedNamesAreDisclosed`](../cmd/birdview/main_test.go#L5631)
- Blocks: blocks once met
- State: met since 2026-10-05 (currently blocking)

### [`acceptance.figureSaysWhatItCounts`](requirements/acceptance.md#figuresayswhatitcounts)

- Verifies requirements:
  - [`product.shareOfRecords`](requirements/local/product/presentation.md#shareofrecords)
- Guardrails:
  - [`TestTheUnitSaysSpeciesAndNotBird`](../cmd/birdview/main_test.go#L4577)
- Blocks: blocks once met
- State: met since 2026-10-06 (currently blocking)

### [`acceptance.detailSaysWhatTileSays`](requirements/acceptance.md#detailsayswhattilesays)

- Verifies requirements:
  - [`product.tileOpensDetail`](requirements/local/product/presentation.md#tileopensdetail)
  - [`product.shareOfRecords`](requirements/local/product/presentation.md#shareofrecords)
- Guardrails:
  - [`TestTheDetailPageSaysEverythingTheTileSays`](../cmd/birdview/main_test.go#L5840)
- Blocks: blocks once met
- State: met since 2026-10-07 (currently blocking)

### [`acceptance.disputedNamesShown`](requirements/acceptance.md#disputednamesshown)

- Verifies requirements:
  - [`product.namesRecordsFiledUnder`](requirements/local/product/presentation.md#namesrecordsfiledunder)
- Guardrails:
  - [`TestEveryNameDisagreementIsOnThePage`](../cmd/birdview/main_test.go#L5943)
- Blocks: blocks once met
- State: met since 2026-10-07 (currently blocking)

### [`acceptance.honestRevalidation`](requirements/acceptance.md#honestrevalidation)

- Verifies requirements:
  - [`tech.cacheLifetimeStated`](requirements/local/tech/delivery.md#cachelifetimestated)
- Guardrails:
  - [`TestResponsesCarryAValidator`](../cmd/birdview/main_test.go#L6052)
- Blocks: blocks once met
- State: met since 2026-10-07 (currently blocking)

### [`acceptance.nameDisclosureLegible`](requirements/acceptance.md#namedisclosurelegible)

- Verifies requirements:
  - [`product.namesRecordsFiledUnder`](requirements/local/product/presentation.md#namesrecordsfiledunder)
- Guardrails:
  - [`TestTheNameDisclosureExplainsItselfAndLinksSomewhereUseful`](../cmd/birdview/main_test.go#L6153)
- Blocks: blocks once met
- State: met since 2026-10-07 (currently blocking)

### [`acceptance.monthsFilterAndAbstain`](requirements/acceptance.md#monthsfilterandabstain)

- Verifies requirements:
  - [`product.regionAndMonthFilter`](requirements/local/product/filtering.md#regionandmonthfilter)
  - [`product.shareOfRecords`](requirements/local/product/presentation.md#shareofrecords)
  - [`decisions.monthPresence`](requirements/decisions.md#monthpresence)
- Guardrails:
  - [`TestMonthScopeWraps`](../cmd/birdview/main_test.go#L6267)
  - [`TestAllTwelveMonthsIsTheWholeYear`](../cmd/birdview/main_test.go#L6328)
  - [`TestMonthsNarrowTheListAndTheFigureTogether`](../cmd/birdview/main_test.go#L6416)
  - [`TestAbstainedTilesSayWhy`](../cmd/birdview/main_test.go#L6482)
  - [`TestPresentRecoversKnownSeasons`](../internal/month/month_test.go#L55)
  - [`TestPresentAbstainsBelowTheFloor`](../internal/month/month_test.go#L121)
- Blocks: blocks once met
- State: met since 2026-10-07 (currently blocking)

### [`acceptance.withheldMonthsExplained`](requirements/acceptance.md#withheldmonthsexplained)

- Verifies requirements:
  - [`product.regionAndMonthFilter`](requirements/local/product/filtering.md#regionandmonthfilter)
  - [`product.namesRecordsFiledUnder`](requirements/local/product/presentation.md#namesrecordsfiledunder)
  - [`decisions.ebirdSilenceNotAbsence`](requirements/decisions.md#ebirdsilencenotabsence)
- Guardrails:
  - [`TestWithheldMonthsAreNotJudged`](../cmd/birdview/main_test.go#L6602)
  - [`TestPresenceReadsWithheldFlags`](../internal/region/region_test.go#L336)
  - [`TestEBirdSensitiveIsWellFormed`](../internal/taxonomy/sensitive_test.go#L28)
  - [`TestWindowsCoverWholeMonthsOnly`](../internal/taxonomy/sensitive_test.go#L83)
  - [`TestConfirmedMonthsBelievesTheRecordsOverTheList`](../internal/taxonomy/sensitive_test.go#L139)
- Blocks: blocks once met
- State: met since 2026-10-09 (currently blocking)

### [`acceptance.ebirdCountedInatFillsGaps`](requirements/acceptance.md#ebirdcountedinatfillsgaps)

- Verifies requirements:
  - [`product.scopeIsRecurringAvifauna`](requirements/local/product/dataset.md#scopeisrecurringavifauna)
  - [`product.regionAndMonthFilter`](requirements/local/product/filtering.md#regionandmonthfilter)
  - [`product.shareOfRecords`](requirements/local/product/presentation.md#shareofrecords)
  - [`product.namesRecordsFiledUnder`](requirements/local/product/presentation.md#namesrecordsfiledunder)
  - [`decisions.countEbirdFirst`](requirements/decisions.md#countebirdfirst)
- Guardrails:
  - [`TestMeasureCountsEBirdAndFillsOnlyItsDeclaredGaps`](../cmd/birdview-regions/main_test.go#L54)
  - [`TestScopeCountsEBirdAndFillsOnlyItsDeclaredGaps`](../cmd/birdview-status/main_test.go#L314)
  - [`TestScopeBringsBothHalvesOfASharedName`](../internal/status/status_test.go#L305)
- Blocks: blocks once met
- State: met since 2026-10-12 (currently blocking)

### [`acceptance.pipelineDiagramServed`](requirements/acceptance.md#pipelinediagramserved)

- Verifies requirements:
  - [`product.pipelineDiagram`](requirements/local/product/measurements.md#pipelinediagram)
  - [`product.categoryProvenance`](requirements/local/product/measurements.md#categoryprovenance)
  - [`tech.selfContainedServer`](requirements/local/tech/delivery.md#selfcontainedserver)
- Guardrails:
  - [`TestDiagramsAreCurrent`](../cmd/birdview/diagram_test.go#L35)
  - [`TestDiagramsAreSelfContained`](../cmd/birdview/diagram_test.go#L76)
  - [`TestPipelineDiagramIsServed`](../cmd/birdview/diagram_test.go#L114)
  - [`TestPipelineDiagramNamesRealColumns`](../cmd/birdview/diagram_test.go#L153)
  - [`TestPipelineDiagramNamesEveryTaxonomyTable`](../cmd/birdview/diagram_test.go#L200)
  - [`TestPipelineDiagramIsLinked`](../cmd/birdview/diagram_test.go#L237)
  - [`TestPipelineFieldsCarryOneMark`](../cmd/birdview/diagram_test.go#L264)
  - [`TestPipelineTablesAreLeftAligned`](../cmd/birdview/diagram_test.go#L307)
- Blocks: blocks once met
- State: met since 2026-10-12 (currently blocking)

### [`acceptance.sharedMeasurementsDisclosed`](requirements/acceptance.md#sharedmeasurementsdisclosed)

- Verifies requirements:
  - [`product.borrowedMeasurementsSaySo`](requirements/local/product/measurements.md#borrowedmeasurementssayso)
- Guardrails:
  - [`TestSameMeasurementsAreNoted`](../cmd/birdview/main_test.go#L6691)
  - [`TestSameMeasurementsNeedNoTable`](../cmd/birdview/main_test.go#L6748)
  - [`TestVegaGullSharesAmericanHerringGullsRow`](../cmd/birdview-ingest/main_test.go#L659)
- Blocks: blocks once met
- State: met since 2026-10-13 (currently blocking)

### [`acceptance.belowBarCountsKept`](requirements/acceptance.md#belowbarcountskept)

- Verifies requirements:
  - [`product.regionAndMonthFilter`](requirements/local/product/filtering.md#regionandmonthfilter)
  - [`product.shareOfRecords`](requirements/local/product/presentation.md#shareofrecords)
- Guardrails:
  - [`TestRecordsBelowTheBarCount`](../cmd/birdview/main_test.go#L6780)
  - [`TestBelowBarRegionsAreCountedNotListed`](../cmd/birdview-ingest/main_test.go#L743)
  - [`TestBelowBarCellsRoundTrip`](../cmd/birdview-regions/below_test.go#L30)
- Blocks: blocks once met
- State: met since 2026-10-13 (currently blocking)

### [`acceptance.escapedBirdsListedAndServed`](requirements/acceptance.md#escapedbirdslistedandserved)

- Verifies requirements:
  - [`product.escapedBirdsListed`](requirements/local/product/measurements.md#escapedbirdslisted)
  - [`product.statusIsPublished`](requirements/local/product/dataset.md#statusispublished)
- Guardrails:
  - [`TestExoticsAreOffTheIndexAndListed`](../cmd/birdview/main_test.go#L6819)
  - [`TestExoticsAreFlaggedFromTheList`](../cmd/birdview-ingest/main_test.go#L770)
- Blocks: blocks once met
- State: met since 2026-10-14 (currently blocking)

### [`acceptance.readersToldStickFigure`](requirements/acceptance.md#readerstoldstickfigure)

- Verifies requirements:
  - [`product.segmentComposition`](requirements/local/product/diagram.md#segmentcomposition)
  - [`decisions.stickFigure`](requirements/decisions.md#stickfigure)
- Guardrails:
  - [`TestReadersAreToldItIsAStickFigure`](../cmd/birdview/main_test.go#L6902)
- Blocks: blocks once met
- State: met since 2026-10-15 (currently blocking)

### [`app.everyPageIsFound`](requirements/composed/internal/app/acceptance.md#everypageisfound)

- Verifies requirements:
  - [`tech.webPages`](requirements/local/tech/accessibility.md#webpages)
- Guardrails:
  - [`TestEveryKindOfPageIsDiscovered`](../cmd/birdview/pages_test.go#L159)
- Blocks: blocks once met
- State: met since 2026-10-16 (currently blocking); recorded in prose, awaiting the field when `app` is next recomposed

### [`app.pagesKeepToTheStylesheet`](requirements/composed/internal/app/acceptance.md#pageskeeptothestylesheet)

- Verifies requirements:
  - none named
- Guardrails:
  - [`TestPagesKeepToTheStylesheetContract`](../cmd/birdview/pages_test.go#L223)
- Blocks: blocks once met
- State: met since 2026-10-16 (currently blocking); recorded in prose, awaiting the field when `app` is next recomposed

### [`app.renderedTextContrast`](requirements/composed/internal/app/acceptance.md#renderedtextcontrast)

- Verifies requirements:
  - none named
- Guardrails:
  - [`TestRenderedTextContrast`](../cmd/birdview/pages_test.go#L320)
- Blocks: blocks once met
- State: met since 2026-10-16 (currently blocking); recorded in prose, awaiting the field when `app` is next recomposed

### [`app.pagesFitAPhone`](requirements/composed/internal/app/acceptance.md#pagesfitaphone)

- Verifies requirements:
  - [`product.fitsAPhone`](requirements/local/product/presentation.md#fitsaphone)
- Guardrails:
  - [`TestPagesFitAPhone`](../cmd/birdview/phone_test.go#L61)
- Blocks: blocks once met
- State: met since 2026-10-25 (currently blocking); recorded in prose, awaiting the field when `app` is next recomposed

### [`app.boundariesResolveKnownPoints`](requirements/composed/internal/app/acceptance.md#boundariesresolveknownpoints)

- Verifies requirements:
  - [`tech.regionBoundaries`](requirements/local/tech/delivery.md#regionboundaries)
- Guardrails:
  - [`TestKnownPoints`](../cmd/birdview-boundaries/main_test.go#L39)
- Blocks: blocks once met
- State: met since 2026-10-18 (currently blocking); recorded in prose, awaiting the field when `app` is next recomposed

### [`app.boundariesCoverEveryRegion`](requirements/composed/internal/app/acceptance.md#boundariescovereveryregion)

- Verifies requirements:
  - [`tech.regionBoundaries`](requirements/local/tech/delivery.md#regionboundaries)
- Guardrails:
  - [`TestCoverage`](../cmd/birdview-boundaries/main_test.go#L53)
- Blocks: blocks once met
- State: met since 2026-10-18 (currently blocking); recorded in prose, awaiting the field when `app` is next recomposed

### [`app.hereAndNowSitBesideAll`](requirements/composed/internal/app/acceptance.md#hereandnowsitbesideall)

- Verifies requirements:
  - [`product.hereSelectsLocalRegion`](requirements/local/product/filtering.md#hereselectslocalregion)
  - [`product.nowSelectsCurrentMonth`](requirements/local/product/filtering.md#nowselectscurrentmonth)
- Guardrails:
  - [`TestHereAndNow`](../cmd/birdview/herenow_test.go#L55)
  - [`TestMonthsFoldIntoTwoLines`](../cmd/birdview/herenow_test.go#L335)
  - [`TestWelcomeLine`](../cmd/birdview/herenow_test.go#L372)
  - [`TestWelcomeLineWhilePicking`](../cmd/birdview/herenow_test.go#L551)
- Blocks: blocks once met
- State: met since 2026-10-18 (currently blocking); recorded in prose, awaiting the field when `app` is next recomposed

### [`app.hereLookupAgreesInTheBrowser`](requirements/composed/internal/app/acceptance.md#herelookupagreesinthebrowser)

- Verifies requirements:
  - [`product.hereSelectsLocalRegion`](requirements/local/product/filtering.md#hereselectslocalregion)
- Guardrails:
  - [`TestHereAndNowLink`](../cmd/birdview/herenow_test.go#L435)
- Blocks: blocks once met
- State: met since 2026-10-18 (currently blocking); recorded in prose, awaiting the field when `app` is next recomposed

### [`app.hereSendsNothingExtra`](requirements/composed/internal/app/acceptance.md#heresendsnothingextra)

- Verifies requirements:
  - [`product.locationStaysInBrowser`](requirements/local/product/filtering.md#locationstaysinbrowser)
- Guardrails:
  - none cited
- Blocks: blocks once met
- State: met since 2026-10-18 (currently blocking); recorded in prose, awaiting the field when `app` is next recomposed

### [`app.tierHistogramMatchesTiers`](requirements/composed/internal/app/acceptance.md#tierhistogrammatchestiers)

- Verifies requirements:
  - [`product.tierHistogram`](requirements/local/product/ordering.md#tierhistogram)
- Guardrails:
  - [`TestTierHistogram`](../cmd/birdview/herenow_test.go#L755)
- Blocks: blocks once met
- State: met since 2026-10-19 (currently blocking); recorded in prose, awaiting the field when `app` is next recomposed


## Changed since the target

`app` was edited after its approval, at [`5e105c8`](https://github.com/google/birdview/commit/5e105c8), and that version is not
approved: composed files change by recomposing, and a published version by a new one. Until it is
approved, implementation still targets `99c0811`, and these are reported, not counted:

- [`app.searchMatchesInGoAndTheBrowser`](requirements/composed/internal/app/acceptance.md#searchmatchesingoandthebrowser): met since 2026-10-25 (not blocking: outside the target)
- [`app.searchFromTheAddress`](requirements/composed/internal/app/acceptance.md#searchfromtheaddress): met since 2026-10-25 (not blocking: outside the target)

## Not in this target

32 criteria verify requirements no composed package includes yet; they belong to
`dataset` and `internal/repo` when those are composed (River `TODO.md`: `composeDatasetAndRepo`).
