# Alignment — BirdView

> **Draft, 2026-10-25**, trying River's proposed alignment record on a real project; River has since
> renamed its own `conformance.md`. Not yet part of River's method; generated, so regenerate rather
> than edit. Generated from the bundle, the tests' `Criterion:` comments, and the history;
> regenerated 2026-11-10.

Which version of each layer the layer below it is aligned with, since when, and who said so; and,
for the implementation, where each criterion of the targeted requirements stands. A version is a
commit. Search `## ` for the sections.

## Chain

| Layer | Aligned with | Since | Said by | State |
| --- | --- | --- | --- | --- |
| Goals, [`goals/`](goals/goals.md) | Nothing above them: BirdView imports no values. Committed at [`dbe250e`](https://github.com/google/birdview/commit/dbe250e); reworded at [`f744b83`](https://github.com/google/birdview/commit/f744b83) (American English) and [`cc4cd67`](https://github.com/google/birdview/commit/cc4cd67) ("how often each one is reported" for "how common") | 2026-10-22 | sameer@, who committed them; no separate review recorded | 14 goals |
| Requirements: [`app`](requirements/composed/internal/app/app.md) | [Goals](goals/goals.md) at [`dbe250e`](https://github.com/google/birdview/commit/dbe250e) | 2026-10-25 | Composed by an agent; approved by Sameer at [`99c0811`](https://github.com/google/birdview/commit/99c0811) | Edited since its approval ([below](#chain)); [criteria](requirements/composed/internal/app/acceptance.md); `dataset` and `internal/repo` not yet composed |
| Implementation, at the commit this file is in | [`app`](requirements/composed/internal/app/app.md) as it stands at that commit | 2026-11-10 | Generated from the criteria's recorded statuses | 87 of 94 criteria in force met, and 8 withdrawn ([below](#implementation-app-as-it-stands)) |
| Release, [birdview.site](https://birdview.site) | Not recorded here; release guardrails in [`tech/release.md`](requirements/local/tech/release.md) | — | — | The deployed version is recorded only in the private `Sajmani/birdview-deploy` ([bindings](bindings.md)) |

`app` was edited in 55 commits after its approval at `99c0811`, from [`5e105c8`](https://github.com/google/birdview/commit/5e105c8) to [`6b97f05`](https://github.com/google/birdview/commit/6b97f05), including the recomposition at [`7d47cd8`](https://github.com/google/birdview/commit/7d47cd8) that removed the classic view; no later approval recorded.

Implementation targets `app` as it stands, not the approved `99c0811`: the edits since are too many
and too far-reaching (the classic view is gone) for the approved version to describe the
implementation. Approving the current `app` would make the two agree.

## Implementation: `app` as it stands

One entry per acceptance criterion that verifies a requirement `app` includes, unchanged or as a
source it composes from, or that `app` composed. **Blocks** is *blocks once met*, the default,
everywhere; no criterion records otherwise yet. A criterion is **met since** its first
`Watched failing:` line, once its status says verified, and from then on a regression blocks:
*(currently blocking)*. *Recorded in prose* means the watched failure is told in the status rather
than in the field. A person has not yet confirmed any of them under the new rule.

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
  - [`TestSpeciesCodesAreUniqueAndURLSafe`](../cmd/birdview-ingest/main_test.go#L218)
- Blocks: blocks once met
- State: met since 2026-10-02 (currently blocking)

### [`acceptance.nameLocaleToggle`](requirements/acceptance.md#namelocaletoggle)

- Verifies requirements:
  - [`product.englishNameChoice`](requirements/local/product/diagram.md#englishnamechoice)
- Guardrails:
  - [`TestNameLocaleTogglesNamesAndNothingElse`](../cmd/birdview/main_test.go#L1608)
  - [`TestNameFallsBackToWorldWhenNoNorthAmericanName`](../cmd/birdview/main_test.go#L1642)
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
- Guardrails:
  - [`TestBillAndTailOrderings`](../cmd/birdview/herenow_test.go#L457)
- Blocks: blocks once met
- State: met since 2026-10-19 (currently blocking)

### [`acceptance.tierHeadingsNameTheOrdering`](requirements/acceptance.md#tierheadingsnametheordering)

- Verifies requirements:
  - [`product.sectionHeadersStateRange`](requirements/local/product/ordering.md#sectionheadersstaterange)
- Guardrails:
  - [`TestTierHeadersNameTheOrdering`](../cmd/birdview/herenow_test.go#L491)
- Blocks: blocks once met
- State: met since 2026-10-19 (currently blocking)

### [`acceptance.tierHeadingsHoldTheirBirds`](requirements/acceptance.md#tierheadingsholdtheirbirds)

- Verifies requirements:
  - [`product.sectionHeadersStateRange`](requirements/local/product/ordering.md#sectionheadersstaterange)
- Guardrails:
  - [`TestTierRangesHoldTheirBirds`](../cmd/birdview/herenow_test.go#L544)
- Blocks: blocks once met
- State: met since 2026-10-19 (currently blocking)

### [`acceptance.tierHeadingDigits`](requirements/acceptance.md#tierheadingdigits)

- Verifies requirements:
  - [`product.sectionHeadersStateRange`](requirements/local/product/ordering.md#sectionheadersstaterange)
- Guardrails:
  - [`TestHeadingDigits`](../cmd/birdview/herenow_test.go#L519)
- Blocks: blocks once met
- State: met since 2026-10-19 (currently blocking)

### [`acceptance.tierCountFollowsTheRange`](requirements/acceptance.md#tiercountfollowstherange)

- Verifies requirements:
  - [`product.sectionHeadersStateRange`](requirements/local/product/ordering.md#sectionheadersstaterange)
- Guardrails:
  - [`TestTierCountFollowsTheRange`](../cmd/birdview/herenow_test.go#L598)
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
  - [`TestSegmentsSumToTotalExactly`](../cmd/birdview-ingest/main_test.go#L257)
  - [`TestSegmentsSumExactly`](../internal/morphology/morphology_test.go#L36)
- Blocks: blocks once met
- State: met since 2026-09-22 (currently blocking)

### [`acceptance.crossbarToScale`](requirements/acceptance.md#crossbartoscale)

- Verifies requirements:
  - [`product.crossbarAtVerticalScale`](requirements/local/product/diagram.md#crossbaratverticalscale)
- Guardrails:
  - [`TestCrossbarIsToScale`](../internal/view/view_test.go#L183)
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
  - [`TestDetailShowsEverySegmentWithProvenance`](../cmd/birdview/main_test.go#L598)
- Blocks: blocks once met
- State: met since 2026-09-22 (currently blocking)

### [`acceptance.totalIsSumOfSegments`](requirements/acceptance.md#totalissumofsegments)

- Verifies requirements:
  - [`product.lengthAndWingspanSources`](requirements/local/product/measurements.md#lengthandwingspansources)
  - [`product.provenancePerMetric`](requirements/local/product/measurements.md#provenancepermetric)
  - [`product.segmentsSumToLength`](requirements/local/product/diagram.md#segmentssumtolength)
- Guardrails:
  - [`TestTotalLengthSaysItIsASum`](../cmd/birdview-ingest/main_test.go#L371)
- Blocks: blocks once met
- State: passing since 2026-09-22, never recorded failing: not yet met

### [`acceptance.noServerNetworkRequest`](requirements/acceptance.md#noservernetworkrequest)

- Verifies requirements:
  - [`tech.selfContainedServer`](requirements/local/tech/delivery.md#selfcontainedserver)
- Guardrails:
  - [`TestNoExternalResources`](../cmd/birdview/main_test.go#L398)
- Blocks: blocks once met
- State: met since 2026-09-22 (currently blocking)

### [`acceptance.bothCommandsBuild`](requirements/acceptance.md#bothcommandsbuild)

- Verifies requirements:
  - [`tech.twoCommandsOneModule`](requirements/local/tech/delivery.md#twocommandsonemodule)
  - [`tech.servesOnLocalhost`](requirements/local/tech/delivery.md#servesonlocalhost)
- Guardrails:
  - [`TestRoutesServe`](../cmd/birdview/main_test.go#L360)
- Blocks: blocks once met
- State: passing since 2026-09-22, never recorded failing: not yet met

### [`acceptance.tileCarriesHeaderAndMetrics`](requirements/acceptance.md#tilecarriesheaderandmetrics)

- Verifies requirements:
  - [`product.tileContents`](requirements/local/product/presentation.md#tilecontents)
- Guardrails:
  - [`TestTileShowsHeaderAndMetrics`](../cmd/birdview/main_test.go#L409)
  - [`TestTileMassLabelSitsOnTheCube`](../cmd/birdview/main_test.go#L475)
  - [`TestDiagramStaysInsideItsBox`](../cmd/birdview/main_test.go#L526)
  - [`TestDiagramIsVerticallyCentered`](../cmd/birdview/main_test.go#L570)
  - [`TestMassLabelSetLikeTheMeasurements`](../web/tilewrap_test.go#L125)
- Blocks: blocks once met
- State: met since 2026-10-01 (currently blocking)

### [`acceptance.familyFilterSelectsAndClears`](requirements/acceptance.md#familyfilterselectsandclears)

- Verifies requirements:
  - [`product.familyFilter`](requirements/local/product/filtering.md#familyfilter)
- Guardrails:
  - [`TestFamilyFilter`](../cmd/birdview/main_test.go#L624)
- Blocks: blocks once met
- State: met since 2026-09-22 (currently blocking)

### [`acceptance.unitToggleSwitches`](requirements/acceptance.md#unittoggleswitches)

- Verifies requirements:
  - [`product.imperialByDefault`](requirements/local/product/presentation.md#imperialbydefault)
- Guardrails:
  - [`TestUnitToggle`](../cmd/birdview/main_test.go#L723)
  - [`TestMassUnitsBySystem`](../cmd/birdview/main_test.go#L772)
  - [`TestNearMatchTooltipNamesTheDonor`](../cmd/birdview/main_test.go#L2293)
- Blocks: blocks once met
- State: met since 2026-09-22 (currently blocking)

### [`acceptance.detailDrawsFullScale`](requirements/acceptance.md#detaildrawsfullscale)

- Verifies requirements:
  - [`product.tileOpensDetail`](requirements/local/product/presentation.md#tileopensdetail)
- Guardrails:
  - [`TestDetailDrawsFullScale`](../cmd/birdview/main_test.go#L800)
- Blocks: blocks once met
- State: met since 2026-09-22 (currently blocking)

### [`acceptance.badgeTextInheritsColor`](requirements/acceptance.md#badgetextinheritscolor)

- Verifies requirements:
  - [`product.generatedBadgeColors`](requirements/local/product/dataset.md#generatedbadgecolors)
  - [`tech.badgeContrast`](requirements/local/tech/accessibility.md#badgecontrast)
- Guardrails:
  - [`TestChipTextInheritsItsBackgroundsTextColor`](../cmd/birdview/main_test.go#L832)
- Blocks: blocks once met
- State: met since 2026-09-22 (currently blocking)

### [`acceptance.emptyStateGuides`](requirements/acceptance.md#emptystateguides)

- Verifies requirements:
  - [`product.helpfulEmptyState`](requirements/local/product/presentation.md#helpfulemptystate)
  - [`product.familyFilter`](requirements/local/product/filtering.md#familyfilter)
- Guardrails:
  - none cited
- Blocks: blocks once met
- State: withdrawn 2026-10-27; met from 2026-09-22

### [`acceptance.scrollScriptScoped`](requirements/acceptance.md#scrollscriptscoped)

- Verifies requirements:
  - [`product.togglesKeepPlace`](requirements/local/product/presentation.md#toggleskeepplace)
  - [`tech.selfContainedServer`](requirements/local/tech/delivery.md#selfcontainedserver)
- Guardrails:
  - none cited
- Blocks: blocks once met
- State: withdrawn 2026-10-27; met from 2026-10-04

### [`acceptance.controlsMatchServer`](requirements/acceptance.md#controlsmatchserver)

- Verifies requirements:
  - [`product.sectionThresholdDefault`](requirements/local/product/ordering.md#sectionthresholddefault)
  - [`product.scaleToggle`](requirements/local/product/diagram.md#scaletoggle)
- Guardrails:
  - none cited
- Blocks: blocks once met
- State: withdrawn 2026-10-27; met from 2026-09-22

### [`acceptance.foldedNeckSameStroke`](requirements/acceptance.md#foldednecksamestroke)

- Verifies requirements:
  - [`product.segmentComposition`](requirements/local/product/diagram.md#segmentcomposition)
  - [`product.segmentsSumToLength`](requirements/local/product/diagram.md#segmentssumtolength)
  - [`product.tileOpensDetail`](requirements/local/product/presentation.md#tileopensdetail)
- Guardrails:
  - [`TestFoldedNeckRendersAsAnArc`](../cmd/birdview/main_test.go#L1029)
  - [`TestFoldedNeckIsTheSameStrokeInLessHeight`](../internal/view/view_test.go#L299)
  - [`TestFoldedNeckDoesNotChangeOrdering`](../internal/view/view_test.go#L356)
- Blocks: blocks once met
- State: met since 2026-09-23 (currently blocking)

### [`acceptance.foldingFamiliesHaveNecks`](requirements/acceptance.md#foldingfamilieshavenecks)

- Verifies requirements:
  - [`product.segmentComposition`](requirements/local/product/diagram.md#segmentcomposition)
  - [`product.neckModelPerFamily`](requirements/local/product/diagram.md#neckmodelperfamily)
- Guardrails:
  - [`TestEveryFoldedFamilyHasALongNeck`](../internal/view/view_test.go#L403)
- Blocks: blocks once met
- State: met since 2026-09-23 (currently blocking)

### [`acceptance.noTextBelowContrastFloor`](requirements/acceptance.md#notextbelowcontrastfloor)

- Verifies requirements:
  - [`tech.textContrastFloor`](requirements/local/tech/accessibility.md#textcontrastfloor)
  - [`wcag.ContrastEnhanced`](requirements/imported/w3.org/wcag/wcag.md#contrastenhanced)
- Guardrails:
  - [`TestStylesheetContrast`](../web/contrast_test.go#L87)
  - [`TestSVGTextContrast`](../web/contrast_test.go#L188)
  - [`TestNoLinkUsesTheBrowserDefault`](../web/contrast_test.go#L270)
- Blocks: blocks once met
- State: met since 2026-09-23 (currently blocking)

### [`acceptance.ownRenderableFavicon`](requirements/acceptance.md#ownrenderablefavicon)

- Verifies requirements:
  - [`tech.selfContainedServer`](requirements/local/tech/delivery.md#selfcontainedserver)
- Guardrails:
  - [`TestFaviconIsServedLocally`](../cmd/birdview/main_test.go#L1198)
- Blocks: blocks once met
- State: met since 2026-09-23 (currently blocking)

### [`acceptance.diagramDistinguishesModeled`](requirements/acceptance.md#diagramdistinguishesmodeled)

- Verifies requirements:
  - [`product.measuredVersusModeled`](requirements/local/product/measurements.md#measuredversusmodeled)
  - [`product.segmentComposition`](requirements/local/product/diagram.md#segmentcomposition)
- Guardrails:
  - [`TestCrossbarSplitsAtTheChord`](../cmd/birdview/main_test.go#L1248)
  - [`TestBothPagesExplainTheStrokeCode`](../cmd/birdview/main_test.go#L1296)
  - [`TestChordTickMarksTheMeasuredPortion`](../internal/view/view_test.go#L426)
  - [`TestStrokeStyleCarriesProvenance`](../web/contrast_test.go#L294)
- Blocks: blocks once met
- State: met since 2026-09-24 (currently blocking)

### [`acceptance.foldedFigureExplained`](requirements/acceptance.md#foldedfigureexplained)

- Verifies requirements:
  - [`product.labeledDetailDiagram`](requirements/local/product/measurements.md#labeleddetaildiagram)
- Guardrails:
  - [`TestFoldedFigureExplained`](../cmd/birdview/main_test.go#L4652)
- Blocks: blocks once met
- State: met since 2026-10-22 (currently blocking)

### [`acceptance.segmentLabelsDontCollide`](requirements/acceptance.md#segmentlabelsdontcollide)

- Verifies requirements:
  - [`product.labeledDetailDiagram`](requirements/local/product/measurements.md#labeleddetaildiagram)
  - [`product.tileOpensDetail`](requirements/local/product/presentation.md#tileopensdetail)
- Guardrails:
  - [`TestDetailTileNamesEverySegment`](../cmd/birdview/main_test.go#L1335)
- Blocks: blocks once met
- State: met since 2026-09-28 (currently blocking)

### [`acceptance.innerSpanDerived`](requirements/acceptance.md#innerspanderived)

- Verifies requirements:
  - [`product.labeledDetailDiagram`](requirements/local/product/measurements.md#labeleddetaildiagram)
  - [`product.tileContents`](requirements/local/product/presentation.md#tilecontents)
- Guardrails:
  - [`TestInnerSpanIsDerived`](../cmd/birdview/main_test.go#L4678)
- Blocks: blocks once met
- State: met since 2026-11-01 (currently blocking)

### [`acceptance.totalLengthLabeled`](requirements/acceptance.md#totallengthlabeled)

- Verifies requirements:
  - [`product.labeledDetailDiagram`](requirements/local/product/measurements.md#labeleddetaildiagram)
- Guardrails:
  - [`TestDetailTileStatesTotalLength`](../cmd/birdview/main_test.go#L1399)
- Blocks: blocks once met
- State: met since 2026-09-28 (currently blocking)

### [`acceptance.familyListWraps`](requirements/acceptance.md#familylistwraps)

- Verifies requirements:
  - [`product.familyFilter`](requirements/local/product/filtering.md#familyfilter)
- Guardrails:
  - none cited
- Blocks: blocks once met
- State: withdrawn 2026-10-27; met from 2026-09-28

### [`acceptance.pickModeSurvivesNavigation`](requirements/acceptance.md#pickmodesurvivesnavigation)

- Verifies requirements:
  - [`product.pickFromIndex`](requirements/local/product/presentation.md#pickfromindex)
  - [`product.tileContents`](requirements/local/product/presentation.md#tilecontents)
- Guardrails:
  - none cited
- Blocks: blocks once met
- State: withdrawn 2026-10-27; met from 2026-10-01

### [`acceptance.massDrawnAsWaterCube`](requirements/acceptance.md#massdrawnaswatercube)

- Verifies requirements:
  - [`product.labeledDetailDiagram`](requirements/local/product/measurements.md#labeleddetaildiagram)
- Guardrails:
  - [`TestMassCubeIsWaterToScale`](../cmd/birdview/main_test.go#L1486)
  - [`TestMassCubeIsRenderedAndExplained`](../cmd/birdview/main_test.go#L1555)
  - [`TestTileCubeNeverReachesTheCrossbar`](../cmd/birdview/main_test.go#L2138)
- Blocks: blocks once met
- State: met since 2026-10-01 (currently blocking)

### [`acceptance.faviconFromPalette`](requirements/acceptance.md#faviconfrompalette)

- Verifies requirements:
  - [`product.oneDiagramVocabulary`](requirements/local/product/diagram.md#onediagramvocabulary)
- Guardrails:
  - none cited
- Blocks: blocks once met
- State: met since 2026-10-02 (currently blocking)

### [`acceptance.partColorsHoldApart`](requirements/acceptance.md#partcolorsholdapart)

- Verifies requirements:
  - [`product.oneDiagramVocabulary`](requirements/local/product/diagram.md#onediagramvocabulary)
  - [`product.labeledDetailDiagram`](requirements/local/product/measurements.md#labeleddetaildiagram)
- Guardrails:
  - [`TestPartColorsHoldApart`](../web/partcolors_test.go#L57)
- Blocks: blocks once met
- State: met since 2026-11-01 (currently blocking)

### [`acceptance.tileHeaderNeverWraps`](requirements/acceptance.md#tileheaderneverwraps)

- Verifies requirements:
  - [`product.tileContents`](requirements/local/product/presentation.md#tilecontents)
- Guardrails:
  - [`TestBandingCodeIsOutsideTheTruncatingElement`](../web/tilewrap_test.go#L88)
- Blocks: blocks once met
- State: met since 2026-10-02 (currently blocking)

### [`acceptance.regionFilterBehaves`](requirements/acceptance.md#regionfilterbehaves)

- Verifies requirements:
  - [`product.regionAndMonthFilter`](requirements/local/product/filtering.md#regionandmonthfilter)
  - [`product.filtersCompose`](requirements/local/product/months.md#filterscompose)
- Guardrails:
  - [`TestRegionFilter`](../cmd/birdview/main_test.go#L1674)
- Blocks: blocks once met
- State: met since 2026-10-03 (currently blocking)

### [`acceptance.habitatGroupsFilter`](requirements/acceptance.md#habitatgroupsfilter)

- Verifies requirements:
  - [`product.habitatAndDietFilters`](requirements/local/product/months.md#habitatanddietfilters)
  - [`decisions.habitatGroupShortcuts`](requirements/decisions.md#habitatgroupshortcuts)
  - [`tech.stableUrls`](requirements/local/tech/delivery.md#stableurls)
- Guardrails:
  - [`TestHabitatGroupsFilter`](../cmd/birdview/next_test.go#L2036)
  - [`TestHabitatGroupsPartitionAVONET`](../internal/trait/review_test.go#L68)
  - [`TestDietGroupsPartitionNiches`](../internal/trait/review_test.go#L112)
- Blocks: blocks once met
- State: met since 2026-11-07 (currently blocking)

### [`acceptance.allAlwaysOffered`](requirements/acceptance.md#allalwaysoffered)

- Verifies requirements:
  - [`product.nextView`](requirements/local/product/presentation.md#nextview)
- Guardrails:
  - [`TestAllAlwaysOffered`](../cmd/birdview/next_test.go#L2186)
- Blocks: blocks once met
- State: met since 2026-11-07 (currently blocking)

### [`acceptance.habitatReviewComplete`](requirements/acceptance.md#habitatreviewcomplete)

- Verifies requirements:
  - [`product.habitatAndDietFilters`](requirements/local/product/months.md#habitatanddietfilters)
  - [`decisions.habitatReviewedPerSpecies`](requirements/decisions.md#habitatreviewedperspecies)
- Guardrails:
  - [`TestEveryDrawnBirdHasAHabitatReview`](../cmd/birdview-ingest/main_test.go#L970)
  - [`TestHabitatReviewParsesAndApplies`](../internal/trait/review_test.go#L30)
- Blocks: blocks once met
- State: met since 2026-11-07 (currently blocking)

### [`acceptance.countyListingApplied`](requirements/acceptance.md#countylistingapplied)

- Verifies requirements:
  - [`product.regionAndMonthFilter`](requirements/local/product/filtering.md#regionandmonthfilter)
  - [`decisions.countyListingRule`](requirements/decisions.md#countylistingrule)
- Guardrails:
  - [`TestCountiesJoinedAsRegions`](../cmd/birdview-ingest/main_test.go#L921)
  - [`TestCountyListingRule`](../cmd/birdview-regions/county_test.go#L33)
  - [`TestCountyRowsFromDownload`](../cmd/birdview-regions/county_test.go#L100)
  - [`TestCountyIsARegion`](../cmd/birdview/main_test.go#L4869)
  - [`TestCountySettingDrillsDown`](../cmd/birdview/main_test.go#L4914)
  - [`TestCountiesOffByDefault`](../cmd/birdview/main_test.go#L4986)
  - [`TestPlaceAndMonthsAreNotJoinedByAComma`](../cmd/birdview/main_test.go#L5021)
  - [`TestCountyParentFromCode`](../internal/region/region_test.go#L329)
- Blocks: blocks once met
- State: met since 2026-11-03 (currently blocking)

### [`acceptance.regionUrlRoundTrips`](requirements/acceptance.md#regionurlroundtrips)

- Verifies requirements:
  - [`product.regionAndMonthFilter`](requirements/local/product/filtering.md#regionandmonthfilter)
  - [`tech.stableUrls`](requirements/local/tech/delivery.md#stableurls)
- Guardrails:
  - [`TestRegionURLsCarryDataCodes`](../cmd/birdview/main_test.go#L1736)
  - [`TestSlugsAreUniqueAndResolve`](../internal/region/region_test.go#L194)
  - [`TestSlugIsDataCode`](../internal/region/region_test.go#L234)
  - [`TestRetiredValuesResolve`](../internal/region/region_test.go#L262)
- Blocks: blocks once met
- State: met since 2026-10-03 (currently blocking)

### [`acceptance.familyGroupingComplete`](requirements/acceptance.md#familygroupingcomplete)

- Verifies requirements:
  - [`product.familiesByOrder`](requirements/local/product/months.md#familiesbyorder)
  - [`product.tileContents`](requirements/local/product/presentation.md#tilecontents)
- Guardrails:
  - [`TestTileCarriesExactlyOneBadge`](../cmd/birdview/main_test.go#L1788)
- Blocks: blocks once met
- State: met since 2026-10-03 (currently blocking)

### [`acceptance.creditsComplete`](requirements/acceptance.md#creditscomplete)

- Verifies requirements:
  - [`product.selfDescriptionStaysTrue`](requirements/local/product/presentation.md#selfdescriptionstaystrue)
- Guardrails:
  - [`TestEveryPageCreditsWhatItShows`](../web/docs_test.go#L112)
  - [`TestDocsCountsMatchTheDataset`](../web/docs_test.go#L145)
  - [`TestNoDocCreditsASourceForWhatItNoLongerSupplies`](../web/docs_test.go#L179)
- Blocks: blocks once met
- State: met since 2026-10-03 (currently blocking)

### [`acceptance.categoryOriginShown`](requirements/acceptance.md#categoryoriginshown)

- Verifies requirements:
  - [`product.categoryProvenance`](requirements/local/product/measurements.md#categoryprovenance)
- Guardrails:
  - [`TestDetailListsEveryCategoryWithItsOrigin`](../cmd/birdview/main_test.go#L1822)
  - [`TestServedDocumentsHaveNoBrokenLinks`](../cmd/birdview/main_test.go#L1929)
  - [`TestVendoredTermsAreNotServed`](../cmd/birdview/main_test.go#L2013)
  - [`TestDocumentTitlesAreHeadings`](../cmd/birdview/main_test.go#L2035)
  - [`TestAddedListsEveryUnpublishedAssignment`](../cmd/birdview/main_test.go#L2058)
  - [`TestCitedAnchorsResolve`](../cmd/birdview/main_test.go#L2334)
- Blocks: blocks once met
- State: met since 2026-10-03 (currently blocking)

### [`acceptance.habitatDietFilterPerSpecies`](requirements/acceptance.md#habitatdietfilterperspecies)

- Verifies requirements:
  - [`product.habitatAndDietFilters`](requirements/local/product/months.md#habitatanddietfilters)
- Guardrails:
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
  - [`TestStaticServesAssetsOnly`](../cmd/birdview/main_test.go#L4632)
- Blocks: blocks once met
- State: met since 2026-10-16 (currently blocking)

### [`acceptance.robotsKeepsCrawlersToPages`](requirements/acceptance.md#robotskeepscrawlerstopages)

- Verifies requirements:
  - [`tech.crawlersSkipFilteredViews`](requirements/local/tech/delivery.md#crawlersskipfilteredviews)
  - [`tech.servesOnlyItsPages`](requirements/local/tech/delivery.md#servesonlyitspages)
- Guardrails:
  - [`TestRobotsKeepsCrawlersToPages`](../cmd/birdview/robots_test.go#L94)
- Blocks: blocks once met
- State: met since 2026-10-23 (currently blocking)

### [`acceptance.stableUrlsKeepWorking`](requirements/acceptance.md#stableurlskeepworking)

- Verifies requirements:
  - [`tech.stableUrls`](requirements/local/tech/delivery.md#stableurls)
- Guardrails:
  - [`TestStableUrlsKeepWorking`](../cmd/birdview/stable_test.go#L47)
- Blocks: blocks once met
- State: met since 2026-10-23 (currently blocking)

### [`acceptance.canonicalHostRedirects`](requirements/acceptance.md#canonicalhostredirects)

- Verifies requirements:
  - [`tech.oneCanonicalHost`](requirements/local/tech/delivery.md#onecanonicalhost)
- Guardrails:
  - [`TestCanonicalHostRedirects`](../cmd/birdview/stable_test.go#L79)
- Blocks: blocks once met
- State: met since 2026-10-23 (currently blocking)

### [`acceptance.wingspanErrorStated`](requirements/acceptance.md#wingspanerrorstated)

- Verifies requirements:
  - [`product.lengthAndWingspanSources`](requirements/local/product/measurements.md#lengthandwingspansources)
  - [`product.metricProvenanceExplained`](requirements/local/product/measurements.md#metricprovenanceexplained)
- Guardrails:
  - [`TestEveryFitStatesItsError`](../internal/wingspan/model_test.go#L116)
- Blocks: blocks once met
- State: met since 2026-11-09 (currently blocking)

### [`acceptance.openOceanFilter`](requirements/acceptance.md#openoceanfilter)

- Verifies requirements:
  - [`product.habitatAndDietFilters`](requirements/local/product/months.md#habitatanddietfilters)
  - [`tech.stableUrls`](requirements/local/tech/delivery.md#stableurls)
  - [`product.categoryProvenance`](requirements/local/product/measurements.md#categoryprovenance)
- Guardrails:
  - [`TestOpenOceanFilter`](../cmd/birdview/next_test.go#L2233)
- Blocks: blocks once met
- State: met since 2026-11-09 (currently blocking)

### [`acceptance.actionsColoredApart`](requirements/acceptance.md#actionscoloredapart)

- Verifies requirements:
  - [`product.nextView`](requirements/local/product/presentation.md#nextview)
- Guardrails:
  - [`TestActionsColoredApart`](../cmd/birdview/chooseone_test.go#L75)
- Blocks: blocks once met
- State: met since 2026-11-09 (currently blocking)

### [`acceptance.massInOneUnit`](requirements/acceptance.md#massinoneunit)

- Verifies requirements:
  - [`product.oneUnitInComparison`](requirements/local/product/presentation.md#oneunitincomparison)
- Guardrails:
  - [`TestUnitsForTakesTheSmallestBird`](../cmd/birdview/main_test.go#L2201)
  - [`TestNoSectionMixesMassUnits`](../cmd/birdview/main_test.go#L2246)
- Blocks: blocks once met
- State: met since 2026-10-04 (currently blocking)

### [`acceptance.commonestShareIsShare`](requirements/acceptance.md#commonestshareisshare)

- Verifies requirements:
  - [`product.shareOfRecords`](requirements/local/product/presentation.md#shareofrecords)
  - [`product.categoryProvenance`](requirements/local/product/measurements.md#categoryprovenance)
- Guardrails:
  - [`TestRegionCountsLineUp`](../cmd/birdview-ingest/main_test.go#L536)
  - [`TestTileShareFollowsTheRegionFilter`](../cmd/birdview/main_test.go#L2405)
- Blocks: blocks once met
- State: met since 2026-10-05 (currently blocking)

### [`acceptance.reportDataIssueLinks`](requirements/acceptance.md#reportdataissuelinks)

- Verifies requirements:
  - [`product.reportDataIssue`](requirements/local/product/presentation.md#reportdataissue)
  - [`product.noPersonalAddress`](requirements/local/product/presentation.md#nopersonaladdress)
- Guardrails:
  - [`TestReportDataIssueLinks`](../cmd/birdview/report_test.go#L31)
  - [`TestHomeReportLink`](../cmd/birdview/report_test.go#L86)
  - [`TestEveryReportLinkFilledIn`](../cmd/birdview/report_test.go#L124)
  - [`TestFeedbackMustBeFillable`](../cmd/birdview/report_test.go#L176)
- Blocks: blocks once met
- State: met since 2026-11-08 (currently blocking)

### [`acceptance.noContactAddressPublished`](requirements/acceptance.md#nocontactaddresspublished)

- Verifies requirements:
  - [`product.noPersonalAddress`](requirements/local/product/presentation.md#nopersonaladdress)
- Guardrails:
  - [`TestNoPersonalAddressIsPublished`](../cmd/birdview/main_test.go#L2535)
- Blocks: blocks once met
- State: met since 2026-10-05 (currently blocking)

### [`acceptance.outOfScopeExplained`](requirements/acceptance.md#outofscopeexplained)

- Verifies requirements:
  - [`product.undrawnSpeciesAnswer`](requirements/local/product/presentation.md#undrawnspeciesanswer)
  - [`decisions.speciesUrlOutlivesSpecies`](requirements/decisions.md#speciesurloutlivesspecies)
- Guardrails:
  - [`TestUndrawnSpeciesNamesItselfAt404`](../cmd/birdview/main_test.go#L2621)
  - [`TestNoIDIsBothDrawnAndNot`](../cmd/birdview/main_test.go#L2710)
  - [`TestAbsentSpeciesCarryNoEvidence`](../cmd/birdview/main_test.go#L2746)
- Blocks: blocks once met
- State: met since 2026-10-05 (currently blocking)

### [`acceptance.commonestIsLargestShare`](requirements/acceptance.md#commonestislargestshare)

- Verifies requirements:
  - [`product.shareOfRecords`](requirements/local/product/presentation.md#shareofrecords)
  - [`decisions.seaShareDiffers`](requirements/decisions.md#seasharediffers)
- Guardrails:
  - [`TestRegionCountsLineUp`](../cmd/birdview-ingest/main_test.go#L536)
  - [`TestCommonestIsTheLargestShare`](../cmd/birdview/main_test.go#L2795)
  - [`TestShareRangesOverOneSetOfRegions`](../cmd/birdview/main_test.go#L2845)
  - [`TestShareStatesZeroRatherThanDashing`](../cmd/birdview/main_test.go#L2929)
  - [`TestMissingTotalsDoNotClaimTheBirdIsAbsent`](../cmd/birdview/main_test.go#L3066)
  - [`TestInterfaceNeverSaysShare`](../cmd/birdview/main_test.go#L3134)
- Blocks: blocks once met
- State: met since 2026-10-05 (currently blocking)

### [`acceptance.peakAndShareOneCalculation`](requirements/acceptance.md#peakandshareonecalculation)

- Verifies requirements:
  - [`product.shareOfRecords`](requirements/local/product/presentation.md#shareofrecords)
  - [`decisions.honestSeaDenominator`](requirements/decisions.md#honestseadenominator)
- Guardrails:
  - [`TestCommonestRoundTripsThroughTheIndex`](../cmd/birdview/main_test.go#L2983)
- Blocks: blocks once met
- State: met since 2026-10-05 (currently blocking)

### [`acceptance.responsesCompress`](requirements/acceptance.md#responsescompress)

- Verifies requirements:
  - [`tech.compressedResponses`](requirements/local/tech/delivery.md#compressedresponses)
- Guardrails:
  - [`TestResponsesCompressWhenAsked`](../cmd/birdview/main_test.go#L3169)
- Blocks: blocks once met
- State: met since 2026-10-05 (currently blocking)

### [`acceptance.frequencyTiers`](requirements/acceptance.md#frequencytiers)

- Verifies requirements:
  - [`product.orderByFrequency`](requirements/local/product/presentation.md#orderbyfrequency)
  - [`decisions.frequencyMetric`](requirements/decisions.md#frequencymetric)
- Guardrails:
  - [`TestFrequencyOrdersAndBandsByDecade`](../cmd/birdview/main_test.go#L3275)
  - [`TestFrequencyTiersAreNotNamed`](../cmd/birdview/main_test.go#L3445)
- Blocks: blocks once met
- State: met since 2026-10-05 (currently blocking)

### [`acceptance.sharedCountDisclosed`](requirements/acceptance.md#sharedcountdisclosed)

- Verifies requirements:
  - [`product.namesRecordsFiledUnder`](requirements/local/product/presentation.md#namesrecordsfiledunder)
  - [`decisions.sharedCountShownOnBoth`](requirements/decisions.md#sharedcountshownonboth)
- Guardrails:
  - [`TestSharedNamesAreDisclosed`](../cmd/birdview/main_test.go#L3475)
- Blocks: blocks once met
- State: met since 2026-10-05 (currently blocking)

### [`acceptance.figureSaysWhatItCounts`](requirements/acceptance.md#figuresayswhatitcounts)

- Verifies requirements:
  - [`product.shareOfRecords`](requirements/local/product/presentation.md#shareofrecords)
- Guardrails:
  - [`TestTheUnitSaysSpeciesAndNotBird`](../cmd/birdview/main_test.go#L2459)
- Blocks: blocks once met
- State: met since 2026-10-06 (currently blocking)

### [`acceptance.detailSaysWhatTileSays`](requirements/acceptance.md#detailsayswhattilesays)

- Verifies requirements:
  - [`product.tileOpensDetail`](requirements/local/product/presentation.md#tileopensdetail)
  - [`product.shareOfRecords`](requirements/local/product/presentation.md#shareofrecords)
- Guardrails:
  - [`TestTheDetailPageSaysEverythingTheTileSays`](../cmd/birdview/main_test.go#L3577)
- Blocks: blocks once met
- State: met since 2026-10-07 (currently blocking)

### [`acceptance.disputedNamesShown`](requirements/acceptance.md#disputednamesshown)

- Verifies requirements:
  - [`product.namesRecordsFiledUnder`](requirements/local/product/presentation.md#namesrecordsfiledunder)
- Guardrails:
  - [`TestEveryNameDisagreementIsOnThePage`](../cmd/birdview/main_test.go#L3685)
- Blocks: blocks once met
- State: met since 2026-10-07 (currently blocking)

### [`acceptance.honestRevalidation`](requirements/acceptance.md#honestrevalidation)

- Verifies requirements:
  - [`tech.cacheLifetimeStated`](requirements/local/tech/delivery.md#cachelifetimestated)
- Guardrails:
  - [`TestResponsesCarryAValidator`](../cmd/birdview/main_test.go#L3794)
- Blocks: blocks once met
- State: met since 2026-10-07 (currently blocking)

### [`acceptance.nameDisclosureLegible`](requirements/acceptance.md#namedisclosurelegible)

- Verifies requirements:
  - [`product.namesRecordsFiledUnder`](requirements/local/product/presentation.md#namesrecordsfiledunder)
- Guardrails:
  - [`TestTheNameDisclosureExplainsItselfAndLinksSomewhereUseful`](../cmd/birdview/main_test.go#L3895)
- Blocks: blocks once met
- State: met since 2026-10-07 (currently blocking)

### [`acceptance.monthsFilterAndAbstain`](requirements/acceptance.md#monthsfilterandabstain)

- Verifies requirements:
  - [`product.regionAndMonthFilter`](requirements/local/product/filtering.md#regionandmonthfilter)
  - [`product.shareOfRecords`](requirements/local/product/presentation.md#shareofrecords)
- Guardrails:
  - [`TestMonthScopeWraps`](../cmd/birdview/main_test.go#L4009)
  - [`TestAllTwelveMonthsIsTheWholeYear`](../cmd/birdview/main_test.go#L4070)
  - [`TestMonthsNarrowTheListAndTheFigureTogether`](../cmd/birdview/main_test.go#L4161)
  - [`TestAbstainedTilesSayWhy`](../cmd/birdview/main_test.go#L4227)
  - [`TestPresentRecoversKnownSeasons`](../internal/month/month_test.go#L55)
  - [`TestPresentAbstainsBelowTheFloor`](../internal/month/month_test.go#L121)
- Blocks: blocks once met
- State: met since 2026-10-07 (currently blocking)

### [`acceptance.withheldMonthsExplained`](requirements/acceptance.md#withheldmonthsexplained)

- Verifies requirements:
  - [`product.regionAndMonthFilter`](requirements/local/product/filtering.md#regionandmonthfilter)
  - [`product.namesRecordsFiledUnder`](requirements/local/product/presentation.md#namesrecordsfiledunder)
- Guardrails:
  - [`TestWithheldMonthsAreNotJudged`](../cmd/birdview/main_test.go#L4278)
  - [`TestPresenceReadsWithheldFlags`](../internal/region/region_test.go#L298)
  - [`TestEBirdSensitiveIsWellFormed`](../internal/taxonomy/sensitive_test.go#L28)
  - [`TestWindowsCoverWholeMonthsOnly`](../internal/taxonomy/sensitive_test.go#L83)
  - [`TestConfirmedMonthsBelievesTheRecordsOverTheList`](../internal/taxonomy/sensitive_test.go#L144)
- Blocks: blocks once met
- State: met since 2026-10-09 (currently blocking)

### [`acceptance.ebirdCountedInatFillsGaps`](requirements/acceptance.md#ebirdcountedinatfillsgaps)

- Verifies requirements:
  - [`product.scopeIsRecurringAvifauna`](requirements/local/product/dataset.md#scopeisrecurringavifauna)
  - [`product.regionAndMonthFilter`](requirements/local/product/filtering.md#regionandmonthfilter)
  - [`product.shareOfRecords`](requirements/local/product/presentation.md#shareofrecords)
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
  - [`TestVegaGullSharesAmericanHerringGullsRow`](../cmd/birdview-ingest/main_test.go#L617)
  - [`TestSameMeasurementsAreNoted`](../cmd/birdview/main_test.go#L4376)
  - [`TestSameMeasurementsNeedNoTable`](../cmd/birdview/main_test.go#L4433)
- Blocks: blocks once met
- State: met since 2026-10-13 (currently blocking)

### [`acceptance.belowBarCountsKept`](requirements/acceptance.md#belowbarcountskept)

- Verifies requirements:
  - [`product.regionAndMonthFilter`](requirements/local/product/filtering.md#regionandmonthfilter)
  - [`product.shareOfRecords`](requirements/local/product/presentation.md#shareofrecords)
- Guardrails:
  - [`TestBelowBarRegionsAreCountedNotListed`](../cmd/birdview-ingest/main_test.go#L701)
  - [`TestBelowBarCellsRoundTrip`](../cmd/birdview-regions/below_test.go#L30)
  - [`TestRecordsBelowTheBarCount`](../cmd/birdview/main_test.go#L4465)
- Blocks: blocks once met
- State: met since 2026-10-13 (currently blocking)

### [`acceptance.escapedBirdsListedAndServed`](requirements/acceptance.md#escapedbirdslistedandserved)

- Verifies requirements:
  - [`product.escapedBirdsListed`](requirements/local/product/measurements.md#escapedbirdslisted)
  - [`product.statusIsPublished`](requirements/local/product/dataset.md#statusispublished)
- Guardrails:
  - [`TestExoticsAreFlaggedFromTheList`](../cmd/birdview-ingest/main_test.go#L728)
  - [`TestExoticsAreOffTheIndexAndListed`](../cmd/birdview/main_test.go#L4509)
- Blocks: blocks once met
- State: met since 2026-10-14 (currently blocking)

### [`acceptance.readersToldStickFigure`](requirements/acceptance.md#readerstoldstickfigure)

- Verifies requirements:
  - [`product.segmentComposition`](requirements/local/product/diagram.md#segmentcomposition)
  - [`decisions.stickFigure`](requirements/decisions.md#stickfigure)
- Guardrails:
  - [`TestReadersAreToldItIsAStickFigure`](../cmd/birdview/main_test.go#L4592)
- Blocks: blocks once met
- State: met since 2026-10-15 (currently blocking)

### [`app.everyPageIsFound`](requirements/composed/internal/app/acceptance.md#everypageisfound)

- Verifies requirements:
  - [`tech.webPages`](requirements/local/tech/accessibility.md#webpages)
- Guardrails:
  - [`TestEveryKindOfPageIsDiscovered`](../cmd/birdview/pages_test.go#L159)
- Blocks: blocks once met
- State: met since 2026-10-16 (currently blocking); recorded in prose

### [`app.pagesKeepToTheStylesheet`](requirements/composed/internal/app/acceptance.md#pageskeeptothestylesheet)

- Verifies requirements:
  - none named
- Guardrails:
  - [`TestPagesKeepToTheStylesheetContract`](../cmd/birdview/pages_test.go#L235)
- Blocks: blocks once met
- State: met since 2026-10-16 (currently blocking); recorded in prose

### [`app.renderedTextContrast`](requirements/composed/internal/app/acceptance.md#renderedtextcontrast)

- Verifies requirements:
  - none named
- Guardrails:
  - [`TestRenderedTextContrast`](../cmd/birdview/pages_test.go#L332)
- Blocks: blocks once met
- State: met since 2026-10-16 (currently blocking); recorded in prose

### [`app.pagesFitAPhone`](requirements/composed/internal/app/acceptance.md#pagesfitaphone)

- Verifies requirements:
  - [`product.fitsAPhone`](requirements/local/product/presentation.md#fitsaphone)
- Guardrails:
  - [`TestPagesFitAPhone`](../cmd/birdview/phone_test.go#L61)
- Blocks: blocks once met
- State: met since 2026-10-25 (currently blocking); recorded in prose

### [`app.searchMatchesInGoAndTheBrowser`](requirements/composed/internal/app/acceptance.md#searchmatchesingoandthebrowser)

- Verifies requirements:
  - [`product.searchBox`](requirements/local/product/filtering.md#searchbox)
- Guardrails:
  - none cited
- Blocks: blocks once met
- State: withdrawn 2026-10-27; met from 2026-10-25

### [`app.searchFromTheAddress`](requirements/composed/internal/app/acceptance.md#searchfromtheaddress)

- Verifies requirements:
  - [`product.searchBox`](requirements/local/product/filtering.md#searchbox)
- Guardrails:
  - none cited
- Blocks: blocks once met
- State: withdrawn 2026-10-27; met from 2026-10-25

### [`app.boundariesResolveKnownPoints`](requirements/composed/internal/app/acceptance.md#boundariesresolveknownpoints)

- Verifies requirements:
  - [`tech.regionBoundaries`](requirements/local/tech/delivery.md#regionboundaries)
- Guardrails:
  - [`TestKnownPoints`](../cmd/birdview-boundaries/main_test.go#L39)
- Blocks: blocks once met
- State: met since 2026-10-18 (currently blocking); recorded in prose

### [`app.boundariesCoverEveryRegion`](requirements/composed/internal/app/acceptance.md#boundariescovereveryregion)

- Verifies requirements:
  - [`tech.regionBoundaries`](requirements/local/tech/delivery.md#regionboundaries)
- Guardrails:
  - [`TestCoverage`](../cmd/birdview-boundaries/main_test.go#L53)
- Blocks: blocks once met
- State: met since 2026-10-18 (currently blocking); recorded in prose

### [`app.hereAndNowSitBesideAll`](requirements/composed/internal/app/acceptance.md#hereandnowsitbesideall)

- Verifies requirements:
  - [`product.hereSelectsLocalRegion`](requirements/local/product/filtering.md#hereselectslocalregion)
  - [`product.nowSelectsCurrentMonth`](requirements/local/product/filtering.md#nowselectscurrentmonth)
- Guardrails:
  - none cited
- Blocks: blocks once met
- State: withdrawn 2026-10-27

### [`app.hereLookupAgreesInTheBrowser`](requirements/composed/internal/app/acceptance.md#herelookupagreesinthebrowser)

- Verifies requirements:
  - [`product.hereSelectsLocalRegion`](requirements/local/product/filtering.md#hereselectslocalregion)
- Guardrails:
  - [`TestHereInTheBrowser`](../cmd/birdview/herenow_test.go#L84)
  - [`TestHereAndNowLink`](../cmd/birdview/herenow_test.go#L306)
- Blocks: blocks once met
- State: met since 2026-10-18 (currently blocking); recorded in prose

### [`app.hereSendsNothingExtra`](requirements/composed/internal/app/acceptance.md#heresendsnothingextra)

- Verifies requirements:
  - [`product.locationStaysInBrowser`](requirements/local/product/filtering.md#locationstaysinbrowser)
- Guardrails:
  - [`TestHereInTheBrowser`](../cmd/birdview/herenow_test.go#L84)
- Blocks: blocks once met
- State: met since 2026-10-18 (currently blocking); recorded in prose

### [`app.tierHistogramMatchesTiers`](requirements/composed/internal/app/acceptance.md#tierhistogrammatchestiers)

- Verifies requirements:
  - [`product.tierHistogram`](requirements/local/product/ordering.md#tierhistogram)
- Guardrails:
  - [`TestTierHistogram`](../cmd/birdview/herenow_test.go#L630)
- Blocks: blocks once met
- State: met since 2026-10-28 (currently blocking)

### [`app.nextViewDrawsOneSet`](requirements/composed/internal/app/acceptance.md#nextviewdrawsoneset)

- Verifies requirements:
  - [`product.nextView`](requirements/local/product/presentation.md#nextview)
- Guardrails:
  - [`TestNowIsTheBrowsersMonth`](../cmd/birdview/herenow_test.go#L414)
  - [`TestDiagramStaysInsideItsBox`](../cmd/birdview/main_test.go#L526)
  - [`TestFoundBirdLinksToItsPage`](../cmd/birdview/main_test.go#L4838)
  - [`TestNextLandingIsThisMonthByFrequency`](../cmd/birdview/next_test.go#L67)
  - [`TestNextDrawsOneScale`](../cmd/birdview/next_test.go#L138)
  - [`TestNextOrders`](../cmd/birdview/next_test.go#L171)
  - [`TestNextFiltersCombineAndEmptyIsAll`](../cmd/birdview/next_test.go#L198)
  - [`TestNextPins`](../cmd/birdview/next_test.go#L268)
  - [`TestNextFindCountsThenLists`](../cmd/birdview/next_test.go#L322)
  - [`TestNextFindInTheBrowser`](../cmd/birdview/next_test.go#L350)
  - [`TestNextToggle`](../cmd/birdview/next_test.go#L452)
  - [`TestNextShowAll`](../cmd/birdview/next_test.go#L481)
  - [`TestNextOrderReverses`](../cmd/birdview/next_test.go#L621)
  - [`TestNextWithoutAMonthIsNotStored`](../cmd/birdview/next_test.go#L669)
  - [`TestNextControlsOpenOneAtATime`](../cmd/birdview/next_test.go#L704)
  - [`TestNextAddedPastTheFirstIsDrawn`](../cmd/birdview/next_test.go#L723)
  - [`TestNextShowPinned`](../cmd/birdview/next_test.go#L811)
  - [`TestNextHereAndNow`](../cmd/birdview/next_test.go#L850)
  - [`TestNextSettingStaysOpen`](../cmd/birdview/next_test.go#L927)
  - [`TestNextFamilyBadges`](../cmd/birdview/next_test.go#L971)
  - [`TestNextResetKeepsPins`](../cmd/birdview/next_test.go#L1016)
  - [`TestNextPinnedList`](../cmd/birdview/next_test.go#L1041)
  - [`TestNextHeading`](../cmd/birdview/next_test.go#L1092)
  - [`TestNextPinKeepsPlace`](../cmd/birdview/next_test.go#L1118)
  - [`TestNextTileCornerAndRows`](../cmd/birdview/next_test.go#L1168)
  - [`TestNextFirstCommitsOnClick`](../cmd/birdview/next_test.go#L1268)
  - [`TestNextWelcomeLine`](../cmd/birdview/next_test.go#L1320)
  - [`TestNextKeepsEveryParameter`](../cmd/birdview/next_test.go#L1416)
  - [`TestNextShortValues`](../cmd/birdview/next_test.go#L1547)
  - [`TestSpeciesBackUsesHistory`](../cmd/birdview/next_test.go#L1595)
  - [`TestNextTiers`](../cmd/birdview/next_test.go#L1654)
  - [`TestSpeciesPinAsksTheHomePage`](../cmd/birdview/next_test.go#L1822)
  - [`TestReportingFrequencyFilters`](../cmd/birdview/next_test.go#L1948)
  - [`TestShowRowFitsAPhone`](../cmd/birdview/phone_test.go#L161)
  - [`TestReferenceIsTheLongestAndWidest`](../internal/view/view_test.go#L152)
  - [`TestLargerBirdsDrawLarger`](../internal/view/view_test.go#L205)
  - [`TestSharedScale`](../internal/view/view_test.go#L252)
- Blocks: blocks once met
- State: met since 2026-10-26 (currently blocking)

### [`app.ebirdLinkOnTilesAndPages`](requirements/composed/internal/app/acceptance.md#ebirdlinkontilesandpages)

- Verifies requirements:
  - [`product.ebirdLink`](requirements/local/product/presentation.md#ebirdlink)
  - [`product.tileContents`](requirements/local/product/presentation.md#tilecontents)
- Guardrails:
  - [`TestEbirdLinks`](../cmd/birdview/main_test.go#L1121)
  - [`TestNextTileCornerAndRows`](../cmd/birdview/next_test.go#L1168)
- Blocks: blocks once met
- State: met since 2026-11-01 (currently blocking)

### [`app.everySpeciesHasAnEBirdCode`](requirements/composed/internal/app/acceptance.md#everyspecieshasanebirdcode)

- Verifies requirements:
  - [`product.ebirdLink`](requirements/local/product/presentation.md#ebirdlink)
- Guardrails:
  - [`TestEverySpeciesHasAnEBirdCode`](../cmd/birdview-ingest/main_test.go#L845)
  - [`TestSpeciesCodesChooseBySimplerRule`](../internal/taxonomy/mapping_test.go#L130)
- Blocks: blocks once met
- State: met since 2026-11-01 (currently blocking)

### [`app.everyCodeReachesItsSpecies`](requirements/composed/internal/app/acceptance.md#everycodereachesitsspecies)

- Verifies requirements:
  - [`product.ebirdLink`](requirements/local/product/presentation.md#ebirdlink)
  - [`tech.stableUrls`](requirements/local/tech/delivery.md#stableurls)
- Guardrails:
  - [`TestEveryCodeReachesItsSpecies`](../cmd/birdview/main_test.go#L1081)
- Blocks: blocks once met
- State: met since 2026-11-01 (currently blocking)

### [`app.textMeetsSizeFloor`](requirements/composed/internal/app/acceptance.md#textmeetssizefloor)

- Verifies requirements:
  - [`tech.textSizeFloor`](requirements/local/tech/accessibility.md#textsizefloor)
- Guardrails:
  - [`TestRenderedTextSize`](../cmd/birdview/size_test.go#L46)
- Blocks: blocks once met
- State: met since 2026-11-01 (currently blocking)

### [`app.pinnedGroupLeads`](requirements/composed/internal/app/acceptance.md#pinnedgroupleads)

- Verifies requirements:
  - [`product.nextView`](requirements/local/product/presentation.md#nextview)
- Guardrails:
  - [`TestPinnedGroupLeadsThePage`](../cmd/birdview/next_test.go#L1896)
- Blocks: blocks once met
- State: met since 2026-11-01 (currently blocking)

### [`app.bandingCodesShown`](requirements/composed/internal/app/acceptance.md#bandingcodesshown)

- Verifies requirements:
  - [`product.bandingCodes`](requirements/local/product/dataset.md#bandingcodes)
- Guardrails:
  - [`TestBandingCodes`](../cmd/birdview-ingest/main_test.go#L760)
  - [`TestBandingJoin`](../cmd/birdview-ingest/main_test.go#L796)
  - [`TestBandingCodesOnThePages`](../cmd/birdview/next_test.go#L1783)
  - [`TestParseTheVendoredList`](../internal/bbl/bbl_test.go#L27)
  - [`TestParseSkipsWhatIsNotOneSpecies`](../internal/bbl/bbl_test.go#L61)
- Blocks: blocks once met
- State: met since 2026-10-27 (currently blocking)

## Not in this target

52 criteria verify requirements no composed package includes; they belong to `dataset` and `internal/repo` when those are composed (River `TODO.md`:
`composeDatasetAndRepo`), or are withdrawn with what they verified.

- [`acceptance.everyRequirementCitesGoal`](requirements/acceptance.md#everyrequirementcitesgoal)
- [`acceptance.vendoredHashesMatch`](requirements/acceptance.md#vendoredhashesmatch)
- [`acceptance.quotationsAreVerbatim`](requirements/acceptance.md#quotationsareverbatim)
- [`acceptance.compositionResolves`](requirements/acceptance.md#compositionresolves)
- [`acceptance.decisionReferencesExist`](requirements/acceptance.md#decisionreferencesexist)
- [`acceptance.scopeMatchesStatusTable`](requirements/acceptance.md#scopematchesstatustable)
- [`acceptance.noRarityCodes`](requirements/acceptance.md#noraritycodes)
- [`acceptance.allFiveTraits`](requirements/acceptance.md#allfivetraits)
- [`acceptance.namesNeverDefaulted`](requirements/acceptance.md#namesneverdefaulted)
- [`acceptance.joinCoversEverySpecies`](requirements/acceptance.md#joincoverseveryspecies)
- [`acceptance.unmatchedFailsBuild`](requirements/acceptance.md#unmatchedfailsbuild)
- [`acceptance.sharedScale`](requirements/acceptance.md#sharedscale)
- [`acceptance.neckLengthSerialized`](requirements/acceptance.md#necklengthserialized)
- [`acceptance.provenanceWellFormed`](requirements/acceptance.md#provenancewellformed)
- [`acceptance.provenanceMatchesSources`](requirements/acceptance.md#provenancematchessources)
- [`acceptance.wingspanModelReconstructsFox`](requirements/acceptance.md#wingspanmodelreconstructsfox)
- [`acceptance.nothingRestrictedTracked`](requirements/acceptance.md#nothingrestrictedtracked)
- [`acceptance.outboundRequestsIdentified`](requirements/acceptance.md#outboundrequestsidentified)
- [`acceptance.abaFileHashMatches`](requirements/acceptance.md#abafilehashmatches)
- [`acceptance.dockerStagesCopyWhatTheyBuild`](requirements/acceptance.md#dockerstagescopywhattheybuild)
- [`acceptance.pageExplainsControls`](requirements/acceptance.md#pageexplainscontrols)
- [`acceptance.countLineNamesEverySetting`](requirements/acceptance.md#countlinenameseverysetting)
- [`acceptance.nothingMatchedOffersAWayOut`](requirements/acceptance.md#nothingmatchedoffersawayout)
- [`acceptance.attributionComplete`](requirements/acceptance.md#attributioncomplete)
- [`acceptance.craniumVaries`](requirements/acceptance.md#craniumvaries)
- [`acceptance.inaturalistLinkConfidentOnly`](requirements/acceptance.md#inaturalistlinkconfidentonly)
- [`acceptance.tileValuesNameTheirSource`](requirements/acceptance.md#tilevaluesnametheirsource)
- [`acceptance.comparisonSharesScale`](requirements/acceptance.md#comparisonsharesscale)
- [`acceptance.comparisonSortRewritesUrl`](requirements/acceptance.md#comparisonsortrewritesurl)
- [`acceptance.trunkScalesLikeSkeletons`](requirements/acceptance.md#trunkscaleslikeskeletons)
- [`acceptance.everyControlDocumented`](requirements/acceptance.md#everycontroldocumented)
- [`acceptance.zonesAnswerToData`](requirements/acceptance.md#zonesanswertodata)
- [`acceptance.everydayDefinitionsShown`](requirements/acceptance.md#everydaydefinitionsshown)
- [`acceptance.gbifDownloadsRecorded`](requirements/acceptance.md#gbifdownloadsrecorded)
- [`acceptance.checksWorkflowIsSound`](requirements/acceptance.md#checksworkflowissound)
- [`acceptance.dockerTestStageGates`](requirements/acceptance.md#dockerteststagegates)
- [`acceptance.duckAndGrebeProportions`](requirements/acceptance.md#duckandgrebeproportions)
- [`acceptance.proportionsReviewTable`](requirements/acceptance.md#proportionsreviewtable)
- [`acceptance.activeTodoHasIssue`](requirements/acceptance.md#activetodohasissue)
- [`acceptance.seaDistanceTable`](requirements/acceptance.md#seadistancetable)
- [`acceptance.seaRefinementsDerived`](requirements/acceptance.md#searefinementsderived)
- [`acceptance.birdPresets`](requirements/acceptance.md#birdpresets)
- [`acceptance.chooseOneSettings`](requirements/acceptance.md#chooseonesettings)
- [`acceptance.imageRecordsRevision`](requirements/acceptance.md#imagerecordsrevision)
- [`acceptance.smokeChecksDeployment`](requirements/acceptance.md#smokechecksdeployment)
- [`acceptance.onlyNarrowsOneGroup`](requirements/acceptance.md#onlynarrowsonegroup)
- [`acceptance.neckRatiosReviewed`](requirements/acceptance.md#neckratiosreviewed)
- [`acceptance.regionJoinReportsUnmatched`](requirements/acceptance.md#regionjoinreportsunmatched)
- [`acceptance.disagreementsDeclared`](requirements/acceptance.md#disagreementsdeclared)
- [`acceptance.occurrenceReadersAgree`](requirements/acceptance.md#occurrencereadersagree)
- [`acceptance.whiteTernFromBirdlife`](requirements/acceptance.md#whiteternfrombirdlife)
- [`acceptance.herringGullFromEuropeanRow`](requirements/acceptance.md#herringgullfromeuropeanrow)
