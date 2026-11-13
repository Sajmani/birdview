# Conformance

What BirdView's code is checked against, and how. This is the mutable record of which guardrails
implement each acceptance criterion in [`composed/internal/app`](requirements/composed/internal/app/acceptance.md),
[`composed/internal/dataset`](requirements/composed/internal/dataset/acceptance.md), and
[`composed/internal/repo`](requirements/composed/internal/repo/acceptance.md), and whether each has
been watched failing. Search `^## ` for the criteria, or `## <pkg>.<name>` for one.

## Built against

- Packages: [`composed/internal/app`](requirements/composed/internal/app/app.md), [`composed/internal/dataset`](requirements/composed/internal/dataset/dataset.md), [`composed/internal/repo`](requirements/composed/internal/repo/repo.md)
- Approvals: recorded in [`requirements/approvals.md`](requirements/approvals.md)
- Standing checks: `go build ./...`, `go vet ./...`, `gofmt -l .`, `go test ./...`, and `river check river`

## app.everyPageIsFound

Criterion: [app.everyPageIsFound](requirements/composed/internal/app/acceptance.md#everypageisfound)
Guardrails: `cmd/birdview/pages_test.go:TestEveryKindOfPageIsDiscovered`
Gates: yes, in `go test ./...` (CI and image build)
Watched failing: 2026-10-16,  — 43 pages across 8 routes and 5 templates. Registering a route nothing links to failed it, naming the route; dropping the seed for an undrawn bird failed it, naming `undrawn.html`, the one template only that page renders.

## app.pagesKeepToTheStylesheet

Criterion: [app.pagesKeepToTheStylesheet](requirements/composed/internal/app/acceptance.md#pageskeeptothestylesheet)
Guardrails: `cmd/birdview/pages_test.go:TestPagesKeepToTheStylesheetContract`
Gates: yes, in `go test ./...` (CI and image build)

## app.renderedTextContrast

Criterion: [app.renderedTextContrast](requirements/composed/internal/app/acceptance.md#renderedtextcontrast)
Guardrails: `cmd/birdview/pages_test.go:TestRenderedTextContrast`
Gates: yes, in `go test ./...` (CI and image build)
Watched failing: 2026-10-16, by setting selected-button text to `#2a7fa8` (2.42:1 on every selected button) and by setting the pipeline diagram's dark text to `#aaa` (1.99:1).

## app.pagesFitAPhone

Criterion: [app.pagesFitAPhone](requirements/composed/internal/app/acceptance.md#pagesfitaphone)
Guardrails: `cmd/birdview/phone_test.go:TestPagesFitAPhone`
Gates: yes, in `go test ./...` (CI and image build)
Watched failing: 2026-10-25, on 7 of 12 pages before the layout fix (588px wide in a 360px frame) and by restoring the 380px column minimum.

## app.boundariesResolveKnownPoints

Criterion: [app.boundariesResolveKnownPoints](requirements/composed/internal/app/acceptance.md#boundariesresolveknownpoints)
Guardrails: `cmd/birdview-boundaries/main_test.go:TestKnownPoints`
Gates: yes, in `go test ./...` (CI and image build)
Watched failing: 2026-10-18,  making the lookup try the sea boxes before land failed it on Tampa, which lies inside the Gulf box; deleting Midway's boundary failed it on Sand Island, which then fell to the Central Pacific box.

## app.boundariesCoverEveryRegion

Criterion: [app.boundariesCoverEveryRegion](requirements/composed/internal/app/acceptance.md#boundariescovereveryregion)
Guardrails: `cmd/birdview-boundaries/main_test.go:TestCoverage`;
  `cmd/birdview/herenow_test.go:TestBoundaryFileIsServedByHash`
Gates: yes, in `go test ./...` (CI and image build)

## app.hereLookupAgreesInTheBrowser

Criterion: [app.hereLookupAgreesInTheBrowser](requirements/composed/internal/app/acceptance.md#herelookupagreesinthebrowser)
Guardrails: `cmd/birdview/herenow_test.go:TestHereInTheBrowser`; `cmd/birdview/herenow_test.go:TestHereAndNowLink`
Gates: yes, in `go test ./...` (CI and image build)
Watched failing: 2026-10-18,  a browser lookup that tried the sea boxes first failed on Tampa, and one that read a refusal as an ordinary failure failed on the message. Needs Chrome, and skips without it.

## app.hereSendsNothingExtra

Criterion: [app.hereSendsNothingExtra](requirements/composed/internal/app/acceptance.md#heresendsnothingextra)
Guardrails: `cmd/birdview/herenow_test.go:TestHereInTheBrowser`
Gates: yes, in `go test ./...` (CI and image build)
Watched failing: 2026-10-18,  adding the latitude to the URL a press requests failed it on the coordinate in the request and on the page requested. A press that navigates leaves headless Chrome waiting under the flags that keep it off the network, so that case reads the server's request log and then stops the browser.

## app.tierHistogramMatchesTiers

Criterion: [app.tierHistogramMatchesTiers](requirements/composed/internal/app/acceptance.md#tierhistogrammatchestiers)
Guardrails: `cmd/birdview/herenow_test.go:TestTierHistogram`
Gates: yes, in `go test ./...` (CI and image build)
Watched failing: 2026-10-28, by the page hiding the histogram for a single tier (`one pinned bird, one tier: the histogram has 0 bars`).

## app.nextViewDrawsOneSet

Criterion: [app.nextViewDrawsOneSet](requirements/composed/internal/app/acceptance.md#nextviewdrawsoneset)
Guardrails: `cmd/birdview/herenow_test.go:TestWelcomeLine`;
  `cmd/birdview/herenow_test.go:TestNowIsTheBrowsersMonth`;
  `cmd/birdview/main_test.go:TestDiagramStaysInsideItsBox`; `cmd/birdview/main_test.go:TestFoundBirdLinksToItsPage`;
  `cmd/birdview/next_test.go:TestNextLandingIsThisMonthByFrequency`;
  `cmd/birdview/next_test.go:TestNextDrawsOneScale`; `cmd/birdview/next_test.go:TestNextOrders`;
  `cmd/birdview/next_test.go:TestNextFiltersCombineAndEmptyIsAll`; `cmd/birdview/next_test.go:TestNextPins`;
  `cmd/birdview/next_test.go:TestNextFindCountsThenLists`; `cmd/birdview/next_test.go:TestNextFindInTheBrowser`;
  `cmd/birdview/next_test.go:TestNextToggle`; `cmd/birdview/next_test.go:TestNextShowAll`;
  `cmd/birdview/next_test.go:TestNextOrderReverses`; `cmd/birdview/next_test.go:TestNextWithoutAMonthIsNotStored`;
  `cmd/birdview/next_test.go:TestNextControlsOpenOneAtATime`;
  `cmd/birdview/next_test.go:TestNextAddedPastTheFirstIsDrawn`; `cmd/birdview/next_test.go:TestNextShowPinned`;
  `cmd/birdview/next_test.go:TestNextHereAndNow`; `cmd/birdview/next_test.go:TestNextSettingStaysOpen`;
  `cmd/birdview/next_test.go:TestNextFamilyBadges`; `cmd/birdview/next_test.go:TestNextResetKeepsPins`;
  `cmd/birdview/next_test.go:TestNextPinnedList`; `cmd/birdview/next_test.go:TestNextHeading`;
  `cmd/birdview/next_test.go:TestNextPinKeepsPlace`; `cmd/birdview/next_test.go:TestNextTileCornerAndRows`;
  `cmd/birdview/next_test.go:TestNextFirstCommitsOnClick`; `cmd/birdview/next_test.go:TestNextWelcomeLine`;
  `cmd/birdview/next_test.go:TestNextKeepsEveryParameter`; `cmd/birdview/next_test.go:TestNextShortValues`;
  `cmd/birdview/next_test.go:TestSpeciesBackUsesHistory`; `cmd/birdview/next_test.go:TestNextTiers`;
  `cmd/birdview/next_test.go:TestSpeciesPinAsksTheHomePage`;
  `cmd/birdview/next_test.go:TestReportingFrequencyFilters`; `cmd/birdview/phone_test.go:TestShowRowFitsAPhone`;
  `cmd/birdview/search_test.go:TestSearchMatchesNamesCodesAndTaxa`;
  `internal/view/view_test.go:TestReferenceIsTheLongestAndWidest`;
  `internal/view/view_test.go:TestLargerBirdsDrawLarger`; `internal/view/view_test.go:TestSharedScale`
Gates: yes, in `go test ./...` (CI and image build)
Watched failing: 2026-10-26, by dropping the frequency and month defaults, scaling each bird to its own length instead of one shared scale, ignoring `all=1` and `rev=1`, and letting pins or filters drop active URL parameters.
Watched failing: 2026-10-27, by the old 15% tier step, by `here.js` leaving `month=now` unresolved, by the server returning 200 instead of 303 on `/?month=now`, and by tiles stretching past their boxes at 360px.
Watched failing: 2026-11-08, with the `Reporting frequency` threshold unapplied (`freq=100` admitting rarer birds) and before found-bird names linked to their species pages.
Watched failing: 2026-11-09, by `view.Reference` taking the last bird's wingspan rather than the widest (`Reference is 800.0 and 500.0`) after the page consolidated onto one shared scale.

## app.ebirdLinkOnTilesAndPages

Criterion: [app.ebirdLinkOnTilesAndPages](requirements/composed/internal/app/acceptance.md#ebirdlinkontilesandpages)
Guardrails: `cmd/birdview/main_test.go:TestEbirdLinks`; `cmd/birdview/next_test.go:TestNextTileCornerAndRows`
Gates: yes, in `go test ./...` (CI and image build)
Watched failing: 2026-11-01, against the iNaturalist links as they stood ("greblu's tile has no eBird link", "a page still links to an iNaturalist taxon page").

## app.everySpeciesHasAnEBirdCode

Criterion: [app.everySpeciesHasAnEBirdCode](requirements/composed/internal/app/acceptance.md#everyspecieshasanebirdcode)
Guardrails: `cmd/birdview-ingest/main_test.go:TestEverySpeciesHasAnEBirdCode`;
  `internal/taxonomy/mapping_test.go:TestSpeciesCodesChooseBySimplerRule`
Gates: yes, in `go test ./...` (CI and image build)
Watched failing: 2026-11-01, against AviList's codes alone ("3 species have no eBird code", "Apus pacificus carries eBird code \"fotswi1\", want \"fotswi\"").
Watched failing: 2026-11-01, by the first rule, any digits against none ("the rule picks \"y00427\"").
Watched failing: 2026-11-01, by ids that stay AviList's code or the slug ("Oxyura jamaicensis has code \"oxyura_jamaicensis\" and id \"oxyura_jamaicensis\", want both \"rudduc\"", "answers also to [], want [oxyura_jamaicensis]"); this also found that a species with no code at all would have shown its slug as its eBird code, now fixed.

## app.everyCodeReachesItsSpecies

Criterion: [app.everyCodeReachesItsSpecies](requirements/composed/internal/app/acceptance.md#everycodereachesitsspecies)
Guardrails: `cmd/birdview/main_test.go:TestEveryCodeReachesItsSpecies`
Gates: yes, in `go test ./...` (CI and image build)
Watched failing: 2026-11-01, against the server before aliases ("GET /species/grbher3?units=metric = 404", "GET /?pin=grbher3 = 200", "find=grbher3 lists []"), and with bird= taking two redirects ("301 to \"/?bird=greblu\"").

## app.textMeetsSizeFloor

Criterion: [app.textMeetsSizeFloor](requirements/composed/internal/app/acceptance.md#textmeetssizefloor)
Guardrails: `cmd/birdview/size_test.go:TestRenderedTextSize`
Gates: yes, in `go test ./...` (CI and image build)
Watched failing: 2026-11-01, against the stylesheet as it stood: all 20 page renders, among them the species page's diagram labels at 3.5px in the frame ("3.5px below 13px at 360px wide: \"Wingspan 15.4 in\""), the measurement lines at 10.5px, and running text at 14px.

## app.pinnedGroupLeads

Criterion: [app.pinnedGroupLeads](requirements/composed/internal/app/acceptance.md#pinnedgroupleads)
Guardrails: `cmd/birdview/next_test.go:TestPinnedGroupLeadsThePage`
Gates: yes, in `go test ./...` (CI and image build)
Watched failing: 2026-11-01, against the page before the group ("the Pinned group heading is at -1").

## app.bandingCodesShown

Criterion: [app.bandingCodesShown](requirements/composed/internal/app/acceptance.md#bandingcodesshown)
Guardrails: `cmd/birdview-ingest/main_test.go:TestBandingCodes`; `cmd/birdview-ingest/main_test.go:TestBandingJoin`;
  `cmd/birdview/next_test.go:TestBandingCodesOnThePages`; `internal/bbl/bbl_test.go:TestParseTheVendoredList`;
  `internal/bbl/bbl_test.go:TestParseSkipsWhatIsNotOneSpecies`
Gates: yes, in `go test ./...` (CI and image build)
Watched failing: 2026-10-31, by the tile's codes moved off the name line ("American Robin's tile does not show AMRO and (amerob) beside its name").
Watched failing: 2026-10-27, by leaving the code out of the search text ("find=AMRO lists []"), and by a species page without the code.
Watched failing: 2026-10-27, by the script's list without the banding code, by the names file without it, by a tile without it, and by the server's list without it.

## repo.everyRequirementCitesGoal

Criterion: [repo.everyRequirementCitesGoal](requirements/composed/internal/repo/acceptance.md#everyrequirementcitesgoal)
Guardrails: `river/check/birdview_test.go:TestEveryLocalRequirementCitesAGoal`
Gates: yes, in `go test ./...` (CI and image build)
Watched failing: 2026-10-16, by deleting `familyFilter`'s `Goals:` line.
Watched failing: 2026-09-21, by deleting [tech.servesOnLocalhost](requirements/local/tech/delivery.md#servesonlocalhost)'s `Goals:` line: exit 1, `requirements with no Goals: field -> tech.servesOnLocalhost`.

## dataset.legibleDistinctColors

Criterion: [dataset.legibleDistinctColors](requirements/composed/internal/dataset/acceptance.md#legibledistinctcolors)
Guardrails: `internal/palette/palette_test.go:TestContrastAgainstOwnText`;
  `internal/palette/palette_test.go:TestContrast`;
  `internal/palette/palette_test.go:TestAdjacentFamiliesAreVisiblyDifferent`
Gates: yes, in `go test ./...` (CI and image build)
Watched failing: 2026-09-22, on both halves: lowering the contrast floor to 3.0 produced four violations, all blues and violets, where relative luminance is lowest at a given HSL lightness; and restoring even hue stepping dropped adjacent families to ΔE 2–6.

## repo.vendoredHashesMatch

Criterion: [repo.vendoredHashesMatch](requirements/composed/internal/repo/acceptance.md#vendoredhashesmatch)
Guardrails: `river/check/pins_test.go:TestVendoredHashesMatch`
Gates: yes, in `go test ./...` (CI and image build)
Watched failing: 2026-09-21, zeroing AVONET's recorded hash produced exit 1 and a recorded-vs-actual diff; re-watched 2026-10-16 after translating the check to Go.

## repo.quotationsAreVerbatim

Criterion: [repo.quotationsAreVerbatim](requirements/composed/internal/repo/acceptance.md#quotationsareverbatim)
Guardrails: `river/check/quotes_test.go:TestQuotationsAreVerbatim`
Gates: yes, in `go test ./...` (CI and image build)
Watched failing: 2026-10-16, changing the large-text ratio in the quotation of WCAG 1.4.6 from 4.5:1 to 3:1 failed it, naming the quotation.

## repo.compositionResolves

Criterion: [repo.compositionResolves](requirements/composed/internal/repo/acceptance.md#compositionresolves)
Guardrails: `river/check/composition_test.go:TestCompositionResolves`
Gates: yes, in `go test ./...` (CI and image build)
Watched failing: 2026-10-16, deleting `Binds: wcag.WebPage` from [tech.webPages](requirements/local/tech/accessibility.md#webpages) failed it, naming the three WCAG criteria left applying to nothing; misspelling a `Covers:` target failed it, naming the criterion.

## repo.decisionReferencesExist

Criterion: [repo.decisionReferencesExist](requirements/composed/internal/repo/acceptance.md#decisionreferencesexist)
Guardrails: `river/check/refs_test.go:TestReferencesResolve`; `river/check/refs_test.go:TestRelativeLinksResolve`
Gates: yes, in `go test ./...` (CI and image build)
Watched failing: 2026-09-21, appending an undefined requirement id in the P-9xx range produced exit 1.
Watched failing: 2026-10-16, on first run against the moved bundle (16 broken links across vendored READMEs and index files).

## dataset.scopeMatchesStatusTable

Criterion: [dataset.scopeMatchesStatusTable](requirements/composed/internal/dataset/acceptance.md#scopematchesstatustable)
Guardrails: `cmd/birdview-ingest/main_test.go:TestScopeIsWhatTheStatusTableSaysRecurs`;
  `cmd/birdview-status/main_test.go:TestLumpsMergeIntoTheSpeciesBirdViewCarries`
Gates: yes, in `go test ./...` (CI and image build)

## dataset.noRarityCodes

Criterion: [dataset.noRarityCodes](requirements/composed/internal/dataset/acceptance.md#noraritycodes)
Guardrails: `cmd/birdview-ingest/main_test.go:TestRarityCodeNeverReachesTheDataset`
Gates: yes, in `go test ./...` (CI and image build)

## dataset.allFiveTraits

Criterion: [dataset.allFiveTraits](requirements/composed/internal/dataset/acceptance.md#allfivetraits)
Guardrails: none yet (planned)
Gates: no

## dataset.namesNeverDefaulted

Criterion: [dataset.namesNeverDefaulted](requirements/composed/internal/dataset/acceptance.md#namesneverdefaulted)
Guardrails: `cmd/birdview-ingest/main_test.go:TestSpeciesCodesArePresent`;
  `cmd/birdview-ingest/main_test.go:TestNamesArePopulatedNotDefaulted`
Gates: yes, in `go test ./...` (CI and image build)
Watched failing: 2026-10-02, by blanking the world name, which leaves the fallback with nothing to fall back to; and by adding a fourth species to the uncoded set, which the named list catches where a percentage threshold would have absorbed it.
Watched failing: 2026-11-01, by dropping American Barn Owl's declared code ("1 species have no eBird code, and no declared one: [Tyto furcata]").

## app.stableSpeciesId

Criterion: [app.stableSpeciesId](requirements/composed/internal/app/acceptance.md#stablespeciesid)
Guardrails: `cmd/birdview-ingest/main_test.go:TestSpeciesCodesAreUniqueAndURLSafe`
Gates: yes, in `go test ./...` (CI and image build)
Watched failing: 2026-10-02, by mapping both lumped species onto their AviList taxon instead of dropping them, which collides *Vireo swainsoni* onto `warvir`, already held by the Eastern Warbling-Vireo; and by blanking the fallback so the three uncoded species get an empty id.

## app.nameLocaleToggle

Criterion: [app.nameLocaleToggle](requirements/composed/internal/app/acceptance.md#namelocaletoggle)
Guardrails: `cmd/birdview/main_test.go:TestNameLocaleTogglesNamesAndNothingElse`;
  `cmd/birdview/main_test.go:TestNameFallsBackToWorldWhenNoNorthAmericanName`
Gates: yes, in `go test ./...` (CI and image build)
Watched failing: 2026-10-02, by serving the North American name for both locales; by dropping the `names` parameter from the link builder, so the toggle renders and does nothing; by hardcoding the on-state; and by removing the fallback, which blanks five headers under the default locale.

## dataset.joinCoversEverySpecies

Criterion: [dataset.joinCoversEverySpecies](requirements/composed/internal/dataset/acceptance.md#joincoverseveryspecies)
Guardrails: `cmd/birdview-ingest/main_test.go:TestNearMatchesAreExactlyTheMappedOnes`
Gates: yes, in `go test ./...` (CI and image build)

## dataset.unmatchedFailsBuild

Criterion: [dataset.unmatchedFailsBuild](requirements/composed/internal/dataset/acceptance.md#unmatchedfailsbuild)
Guardrails: none yet (planned)
Gates: no

## app.sortsAscending

Criterion: [app.sortsAscending](requirements/composed/internal/app/acceptance.md#sortsascending)
Guardrails: `internal/view/view_test.go:TestSortedAscending`
Gates: yes, in `go test ./...` (CI and image build)

## app.billAndTailOrderings

Criterion: [app.billAndTailOrderings](requirements/composed/internal/app/acceptance.md#billandtailorderings)
Guardrails: `cmd/birdview/herenow_test.go:TestBillAndTailOrderings`
Gates: yes, in `go test ./...` (CI and image build)
Watched failing: 2026-10-19, `TestBillAndTailOrderings`, before the orderings existed.

## app.tierHeadingsNameTheOrdering

Criterion: [app.tierHeadingsNameTheOrdering](requirements/composed/internal/app/acceptance.md#tierheadingsnametheordering)
Guardrails: `cmd/birdview/herenow_test.go:TestTierHeadersNameTheOrdering`
Gates: yes, in `go test ./...` (CI and image build)
Watched failing: 2026-10-19, `TestTierHeadersNameTheOrdering`, on the bare "Tier 1".

## app.tierHeadingsHoldTheirBirds

Criterion: [app.tierHeadingsHoldTheirBirds](requirements/composed/internal/app/acceptance.md#tierheadingsholdtheirbirds)
Guardrails: `cmd/birdview/herenow_test.go:TestTierRangesHoldTheirBirds`
Gates: yes, in `go test ./...` (CI and image build)
Watched failing: 2026-10-19, `TestTierRangesHoldTheirBirds`, against headings that ran on to the next tier, as "4.97 in – 8.84 in" did for herons below the pelicans.

## app.tierHeadingDigits

Criterion: [app.tierHeadingDigits](requirements/composed/internal/app/acceptance.md#tierheadingdigits)
Guardrails: `cmd/birdview/herenow_test.go:TestHeadingDigits`
Gates: yes, in `go test ./...` (CI and image build)
Watched failing: 2026-10-19, `TestHeadingDigits`, when only each tier's own ends were kept apart, which let "0.20 – 0.24" sit before "0.24 – 0.28".

## app.tierCountFollowsTheRange

Criterion: [app.tierCountFollowsTheRange](requirements/composed/internal/app/acceptance.md#tiercountfollowstherange)
Guardrails: `cmd/birdview/herenow_test.go:TestTierCountFollowsTheRange`
Gates: yes, in `go test ./...` (CI and image build)
Watched failing: 2026-10-19, `TestTierCountFollowsTheRange`, while it sat at the far edge in a smaller size.

## app.sectionBoundariesObeyRule

Criterion: [app.sectionBoundariesObeyRule](requirements/composed/internal/app/acceptance.md#sectionboundariesobeyrule)
Guardrails: `internal/view/view_test.go:TestSectionRule`; `internal/view/view_test.go:TestSectionsPartition`
Gates: yes, in `go test ./...` (CI and image build)
Watched failing: 2026-09-22, keying the break on the previous *maximum* instead of the minimum let a section span 7% at a 5% step.

## app.sectionHeadersMatch

Criterion: [app.sectionHeadersMatch](requirements/composed/internal/app/acceptance.md#sectionheadersmatch)
Guardrails: `internal/view/view_test.go:TestSectionRangeMatchesContents`
Gates: yes, in `go test ./...` (CI and image build)

## app.segmentsSumExactly

Criterion: [app.segmentsSumExactly](requirements/composed/internal/app/acceptance.md#segmentssumexactly)
Guardrails: `cmd/birdview-ingest/main_test.go:TestSegmentsSumToTotalExactly`;
  `internal/morphology/morphology_test.go:TestSegmentsSumExactly`
Gates: yes, in `go test ./...` (CI and image build)
Watched failing: 2026-09-22, rounding each segment independently instead of balancing broke the sum for both the property test and the real dataset, by 0.1 mm.

## app.crossbarToScale

Criterion: [app.crossbarToScale](requirements/composed/internal/app/acceptance.md#crossbartoscale)
Guardrails: `internal/view/view_test.go:TestCrossbarIsToScale`
Gates: yes, in `go test ./...` (CI and image build)
Watched failing: 2026-09-22, restoring the earlier design's formula, with its extra `availableH × scaleY` factor, broke the ratio for every species.

## app.diagramPartsInOrder

Criterion: [app.diagramPartsInOrder](requirements/composed/internal/app/acceptance.md#diagrampartsinorder)
Guardrails: none yet (planned)
Gates: no

## app.proportionsDifferentiate

Criterion: [app.proportionsDifferentiate](requirements/composed/internal/app/acceptance.md#proportionsdifferentiate)
Guardrails: `internal/morphology/morphology_test.go:TestShapeDifferentiation`
Gates: yes, in `go test ./...` (CI and image build)

## dataset.neckLengthSerialized

Criterion: [dataset.neckLengthSerialized](requirements/composed/internal/dataset/acceptance.md#necklengthserialized)
Guardrails: `cmd/birdview-ingest/main_test.go:TestNeckIsModeledAndNamesItsFamilyRatio`;
  `internal/morphology/morphology_test.go:TestNeckModelNamesTheFamilyRatio`
Gates: yes, in `go test ./...` (CI and image build)

## dataset.provenanceWellFormed

Criterion: [dataset.provenanceWellFormed](requirements/composed/internal/dataset/acceptance.md#provenancewellformed)
Guardrails: `cmd/birdview-ingest/main_test.go:TestEveryMetricHasWellFormedProvenance`
Gates: yes, in `go test ./...` (CI and image build)
Watched failing: 2026-09-22, blanking the neck model string produced a violation for every species.

## dataset.provenanceMatchesSources

Criterion: [dataset.provenanceMatchesSources](requirements/composed/internal/dataset/acceptance.md#provenancematchessources)
Guardrails: `cmd/birdview-ingest/main_test.go:TestProvenanceMatchesTheSourcesUsed`
Gates: yes, in `go test ./...` (CI and image build)
Watched failing: 2026-09-22, removing the near-match inheritance reported 308, i.e. seven borrowed values mislabeled as measured.

## app.tileMarksNonMeasured

Criterion: [app.tileMarksNonMeasured](requirements/composed/internal/app/acceptance.md#tilemarksnonmeasured)
Guardrails: none yet (planned)
Gates: no

## app.detailShowsProvenance

Criterion: [app.detailShowsProvenance](requirements/composed/internal/app/acceptance.md#detailshowsprovenance)
Guardrails: `cmd/birdview/main_test.go:TestDetailShowsEverySegmentWithProvenance`
Gates: yes, in `go test ./...` (CI and image build)
Watched failing: 2026-09-22, removing the neck disclosure was caught.

## dataset.totalIsSumOfSegments

Criterion: [dataset.totalIsSumOfSegments](requirements/composed/internal/dataset/acceptance.md#totalissumofsegments)
Guardrails: `cmd/birdview-ingest/main_test.go:TestTotalLengthSaysItIsASum`
Gates: yes, in `go test ./...` (CI and image build)

## dataset.wingspanModelReconstructsFox

Criterion: [dataset.wingspanModelReconstructsFox](requirements/composed/internal/dataset/acceptance.md#wingspanmodelreconstructsfox)
Guardrails: `internal/wingspan/model_test.go:TestModelAgreesWithFoxPublishedPredictions`
Gates: yes, in `go test ./...` (CI and image build)

## repo.nothingRestrictedTracked

Criterion: [repo.nothingRestrictedTracked](requirements/composed/internal/repo/acceptance.md#nothingrestrictedtracked)
Guardrails: none yet (planned)
Gates: no

## app.noServerNetworkRequest

Criterion: [app.noServerNetworkRequest](requirements/composed/internal/app/acceptance.md#noservernetworkrequest)
Guardrails: `cmd/birdview/main_test.go:TestNoExternalResources`;
  `cmd/birdview/main_test.go:TestFaviconIsServedLocally`
Gates: yes, in `go test ./...` (CI and image build)
Watched failing: 2026-09-22, adding the old design's `cdn.tailwindcss.com` script tag was caught.

## dataset.outboundRequestsIdentified

Criterion: [dataset.outboundRequestsIdentified](requirements/composed/internal/dataset/acceptance.md#outboundrequestsidentified)
Guardrails: none yet (planned)
Gates: no

## repo.bothCommandsBuild

Criterion: [repo.bothCommandsBuild](requirements/composed/internal/repo/acceptance.md#bothcommandsbuild)
Guardrails: `cmd/birdview/main_test.go:TestRoutesServe`
Gates: yes, in `go test ./...` (CI and image build)

## repo.dockerStagesCopyWhatTheyBuild

Criterion: [repo.dockerStagesCopyWhatTheyBuild](requirements/composed/internal/repo/acceptance.md#dockerstagescopywhattheybuild)
Guardrails: `cmd/birdview/dockerfile_test.go:TestDockerStagesCopyWhatTheyBuild`
Gates: yes, in `go test ./...` (CI and image build)
Watched failing: 2026-10-22, `TestDockerStagesCopyWhatTheyBuild`): against the `Dockerfile` as it stood, it reported that the `builder` stage needs `birdview/data` and does not copy it, the defect it was written for; with `go build` misspelled in the file, it reported finding no stage for either command.

## app.tileCarriesHeaderAndMetrics

Criterion: [app.tileCarriesHeaderAndMetrics](requirements/composed/internal/app/acceptance.md#tilecarriesheaderandmetrics)
Guardrails: `cmd/birdview/main_test.go:TestTileShowsHeaderAndMetrics`;
  `cmd/birdview/main_test.go:TestTileMassLabelSitsOnTheCube`;
  `cmd/birdview/main_test.go:TestDiagramStaysInsideItsBox`;
  `cmd/birdview/main_test.go:TestDiagramIsVerticallyCentered`;
  `web/tilewrap_test.go:TestMassLabelSetLikeTheMeasurements`
Gates: yes, in `go test ./...` (CI and image build)
Watched failing: 2026-10-01, by restoring the initials.
Watched failing: 2026-10-30, against the tiles as they stood, names first and measurements last ("on the crow's tile the drawing comes before the share line").
Watched failing: 2026-10-31, by the full badge on a tile ("the crow's badge reads … Corvidae <em>Corvidae family</em>"), by the mass label 2 units above the cube's foot ("baseline is at 138.0, the cube's foot at 140.0"), by a second label in the drawing ("carries 2 labels"), and by the mass label at 11px against the measurements' 10.5px.

## app.familyFilterSelectsAndClears

Criterion: [app.familyFilterSelectsAndClears](requirements/composed/internal/app/acceptance.md#familyfilterselectsandclears)
Guardrails: `cmd/birdview/main_test.go:TestFamilyFilter`
Gates: yes, in `go test ./...` (CI and image build)
Watched failing: 2026-09-22, by alphabetizing the family list, and separately by making the order toggle add only its first family.
Watched failing: 2026-10-02, by putting a count back on the chips — which is the direction the mistake would come from, since the count reads as helpful right up until a second filter makes it false.

## app.unitToggleSwitches

Criterion: [app.unitToggleSwitches](requirements/composed/internal/app/acceptance.md#unittoggleswitches)
Guardrails: `cmd/birdview/main_test.go:TestUnitToggle`; `cmd/birdview/main_test.go:TestMassUnitsBySystem`;
  `cmd/birdview/main_test.go:TestNearMatchTooltipNamesTheDonor`
Gates: yes, in `go test ./...` (CI and image build)
Watched failing: 2026-09-22, on both a hardcoded-inches footer and a kilograms-under-imperial conversion.
Watched failing: 2026-10-31, by a bill hardcoded to inches on the tile, under metric.

## app.detailDrawsFullScale

Criterion: [app.detailDrawsFullScale](requirements/composed/internal/app/acceptance.md#detaildrawsfullscale)
Guardrails: `cmd/birdview/main_test.go:TestDetailDrawsFullScale`
Gates: yes, in `go test ./...` (CI and image build)
Watched failing: 2026-09-22, by making the detail view adopt the dataset-wide reference, which inset the hummingbird by 128 px.

## app.badgeTextInheritsColor

Criterion: [app.badgeTextInheritsColor](requirements/composed/internal/app/acceptance.md#badgetextinheritscolor)
Guardrails: `cmd/birdview/main_test.go:TestChipTextInheritsItsBackgroundsTextColor`
Gates: yes, in `go test ./...` (CI and image build)
Watched failing: 2026-09-22, by restoring the fixed gray.

## app.attributionComplete

Criterion: [app.attributionComplete](requirements/composed/internal/app/acceptance.md#attributioncomplete)
Guardrails: `cmd/birdview/main_test.go:TestAttributionIsComplete`;
  `cmd/birdview/main_test.go:TestGBIFDatasetsCitedWithTheirLicenses`;
  `cmd/birdview/main_test.go:TestNoUnearnedPermissionClaim`
Gates: yes, in `go test ./...` (CI and image build)
Watched failing: 2026-09-22, by stripping the license links and by reinstating a "used by permission" claim.
Watched failing: 2026-10-02, by deleting its credit while leaving its names on every tile — which is the shape the obligation actually has: CC BY permits the use *because* of the attribution, so an unasserted credit is a license breach waiting to happen.
Watched failing: 2026-11-03, against the footers as they stood (`TestGBIFDatasetsCitedWithTheirLicenses`: "the full credits do not cite \"https://doi.org/10.15468/ab3s5x\"", "the footer does not give iNaturalist's license, CC BY-NC 4.0" on all three pages, "the footer licenses iNaturalist's records CC BY 4.0").

## dataset.craniumVaries

Criterion: [dataset.craniumVaries](requirements/composed/internal/dataset/acceptance.md#craniumvaries)
Guardrails: `cmd/birdview-ingest/main_test.go:TestHeadIsNotAConstant`;
  `internal/morphology/morphology_test.go:TestHeadScalesWithMass`
Gates: yes, in `go test ./...` (CI and image build)
Watched failing: 2026-09-22, by reinstating a floor, which pinned 809 of 841 species at one value.

## app.foldedNeckSameStroke

Criterion: [app.foldedNeckSameStroke](requirements/composed/internal/app/acceptance.md#foldednecksamestroke)
Guardrails: `cmd/birdview/main_test.go:TestFoldedNeckRendersAsAnArc`;
  `internal/view/view_test.go:TestFoldedNeckIsTheSameStrokeInLessHeight`;
  `internal/view/view_test.go:TestFoldedNeckDoesNotChangeOrdering`
Gates: yes, in `go test ./...` (CI and image build)
Watched failing: 2026-09-23, by: emptying the folding family table; setting the arc radius to `neck/4` so the stroke is rescaled; leaking the fold into `Scales`; forcing the template to always draw a line; and reporting the drawn neck instead of the modeled one in the detail view.

## app.foldingFamiliesHaveNecks

Criterion: [app.foldingFamiliesHaveNecks](requirements/composed/internal/app/acceptance.md#foldingfamilieshavenecks)
Guardrails: `internal/view/view_test.go:TestEveryFoldedFamilyHasALongNeck`
Gates: yes, in `go test ./...` (CI and image build)
Watched failing: 2026-09-23, against a real bug, which is how it was written: Pelecanidae was in the folded set with no ratio entry, so a pelican drew a 5%-of-length stub tucked into an S.

## app.noTextBelowContrastFloor

Criterion: [app.noTextBelowContrastFloor](requirements/composed/internal/app/acceptance.md#notextbelowcontrastfloor)
Guardrails: `web/contrast_test.go:TestStylesheetContrast`; `web/contrast_test.go:TestSVGTextContrast`;
  `web/contrast_test.go:TestNoLinkUsesTheBrowserDefault`
Gates: yes, in `go test ./...` (CI and image build)
Watched failing: 2026-10-16, by setting selected-button text to `#2a7fa8`; it reported 2.42:1 against its own `--accent` background.
Watched failing: 2026-09-23, by removing the global link rules (which reproduces the original defect exactly), by setting the visited color to the browser's `#551A8B`, by picking a link color that clears AA at 4.90:1 but misses 7:1, and — after the floor rose to 7:1 for everything ([decisions.oneContrastFloor](requirements/decisions.md#onecontrastfloor)) — by restoring the previous `--dim` and `--accent`, which fail at 5.71:1 and 6.83:1 on a.
Watched failing: 2026-10-02, by restoring `#0ea5e9` on SVG label fills and by dropping `font-size` from `.dlabel` (`checked == 0`).

## app.ownRenderableFavicon

Criterion: [app.ownRenderableFavicon](requirements/composed/internal/app/acceptance.md#ownrenderablefavicon)
Guardrails: `cmd/birdview/main_test.go:TestFaviconIsServedLocally`
Gates: yes, in `go test ./...` (CI and image build)
Watched failing: 2026-09-23, by removing the asset while leaving the reference, by removing the reference while leaving the asset, by pointing the reference at a CDN (which trips noServerNetworkRequest's existing scan, as intended), and by reintroducing the XML defect below.

## app.diagramDistinguishesModeled

Criterion: [app.diagramDistinguishesModeled](requirements/composed/internal/app/acceptance.md#diagramdistinguishesmodeled)
Guardrails: `cmd/birdview/main_test.go:TestCrossbarSplitsAtTheChord`;
  `cmd/birdview/main_test.go:TestBothPagesExplainTheStrokeCode`;
  `internal/view/view_test.go:TestChordTickMarksTheMeasuredPortion`;
  `web/contrast_test.go:TestStrokeStyleCarriesProvenance`
Gates: yes, in `go test ./...` (CI and image build)
Watched failing: 2026-09-24, by: restoring the tail's dash; tidying the neck's pattern to match the body's; giving the head a fill back; placing the tick by half the chord instead of the chord; removing the overrun fallback so a tick is drawn outside the wing; and deleting the explanatory text from the index.

## app.foldedFigureExplained

Criterion: [app.foldedFigureExplained](requirements/composed/internal/app/acceptance.md#foldedfigureexplained)
Guardrails: `cmd/birdview/main_test.go:TestFoldedFigureExplained`
Gates: yes, in `go test ./...` (CI and image build)
Watched failing: 2026-10-22, before the note existed (`TestFoldedFigureExplained`).

## app.segmentLabelsDontCollide

Criterion: [app.segmentLabelsDontCollide](requirements/composed/internal/app/acceptance.md#segmentlabelsdontcollide)
Guardrails: `cmd/birdview/main_test.go:TestDetailTileNamesEverySegment`
Gates: yes, in `go test ./...` (CI and image build)
Watched failing: 2026-09-28, by removing the mass label — which is the state this check was written to catch — and by labeling mass with the wingspan's value.
Watched failing: 2026-09-28, by anchoring every label at the same y, by removing the minimum-gap pass, by dropping a segment from the list, by removing the reserved gutter and by inflating it, by moving the wing labels back inside the centering group — which reproduces the bug where the neck's leader struck through "Wing chord" and "Wingspan" — and, for the marker, by reverting the `~` to the `.dlabel.est::after` rule it came from and by mar.
Watched failing: 2026-11-01, against the labeled diagram ("the table reports Mass but the tile does not name it", and Head and Neck likewise), and by the short badge on the species page ("carries the family's common name").

## app.tileValuesNameTheirSource

Criterion: [app.tileValuesNameTheirSource](requirements/composed/internal/app/acceptance.md#tilevaluesnametheirsource)
Guardrails: `cmd/birdview/main_test.go:TestTileValuesNameTheirSource`
Gates: yes, in `go test ./...` (CI and image build)
Watched failing: 2026-11-01, before measured values had tooltips ("Bill (measured) has no tooltip saying where it came from", and Tail, Chord, and Mass likewise; "a measurement filed under an older name has the tooltip \"estimated\""; "a near-match's tooltip is \"measured, but on Sturnella magna — not this species\"; it names the donor and the source").

## app.innerSpanDerived

Criterion: [app.innerSpanDerived](requirements/composed/internal/app/acceptance.md#innerspanderived)
Guardrails: `cmd/birdview/main_test.go:TestInnerSpanIsDerived`
Gates: yes, in `go test ./...` (CI and image build)
Watched failing: 2026-11-01, with the tile's field present and never set ("the species page's tile has no Inner span marked estimated"), and with the tile still saying Inner span ("has no Inner marked estimated").

## app.totalLengthLabeled

Criterion: [app.totalLengthLabeled](requirements/composed/internal/app/acceptance.md#totallengthlabeled)
Guardrails: `cmd/birdview/main_test.go:TestDetailTileStatesTotalLength`
Gates: yes, in `go test ./...` (CI and image build)
Watched failing: 2026-10-01, by sizing the strip for one line, which overran by 4px on every species.
Watched failing: 2026-09-28, by removing the label from the template.
Watched failing: 2026-09-28, by centering it on the box center, which put it at x=494 against the bird's axis at x=334, exactly half the gutter adrift.

## app.massDrawnAsWaterCube

Criterion: [app.massDrawnAsWaterCube](requirements/composed/internal/app/acceptance.md#massdrawnaswatercube)
Guardrails: `cmd/birdview/main_test.go:TestMassCubeIsWaterToScale`;
  `cmd/birdview/main_test.go:TestMassCubeIsRenderedAndExplained`;
  `cmd/birdview/main_test.go:TestTileCubeNeverReachesTheCrossbar`
Gates: yes, in `go test ./...` (CI and image build)
Watched failing: 2026-10-01, by removing the cube from the index tiles, which left 0 across 6.
Watched failing: 2026-10-01, by giving a tile the label too.
Watched failing: 2026-10-01, by removing that prose.

## dataset.trunkScalesLikeSkeletons

Criterion: [dataset.trunkScalesLikeSkeletons](requirements/composed/internal/dataset/acceptance.md#trunkscaleslikeskeletons)
Guardrails: `internal/morphology/morphology_test.go:TestTrunkScalesLikeMeasuredSkeletons`
Gates: yes, in `go test ./...` (CI and image build)
Watched failing: 2026-10-02, by restoring `18.0 * mass^0.31`, keeping `0.37` with the `18.0` coefficient, overcorrecting the exponent to `0.45`, and widening the band to `0.30–0.45`.

## app.faviconFromPalette

Criterion: [app.faviconFromPalette](requirements/composed/internal/app/acceptance.md#faviconfrompalette)
Guardrails: `web/favicon_test.go:TestFaviconUsesTheAppPalette`
Gates: yes, in `go test ./...` (CI and image build)
Watched failing: 2026-10-02, by restoring the old `#5ccbfa` crossbar, changing `.wing` in the stylesheet alone, and renaming `{{define "diagram"}}`.

## app.partColorsHoldApart

Criterion: [app.partColorsHoldApart](requirements/composed/internal/app/acceptance.md#partcolorsholdapart)
Guardrails: `web/partcolors_test.go:TestPartColorsHoldApart`
Gates: yes, in `go test ./...` (CI and image build)
Watched failing: 2026-11-01, against the palette as it stood ("bill #f59e0b is 6.81:1 on the panel", "tail #94a3b8 is 5.71:1 on the panel", no `.k-tail`), and with the rose tail of the scheme the decision rejected ("deuteranopia: tail and crossbar are ΔE00 3.8 apart").
Watched failing: 2026-11-03, with the sky-blue cube and no exemption ("protanopia: chord and mass cube are ΔE00 2.4 apart", "tritanopia: crossbar and mass cube are ΔE00 6.8 apart"), and with the exemption and the original `#38bdf8` stroke, which the exemption does not excuse ("mass cube #38bdf8 is 6.83:1 on the panel", "its name is colored … #7dd3fc, its stroke #38bdf8").

## app.tileHeaderNeverWraps

Criterion: [app.tileHeaderNeverWraps](requirements/composed/internal/app/acceptance.md#tileheaderneverwraps)
Guardrails: `cmd/birdview/main_test.go:TestTileCarriesExactlyOneBadge`;
  `web/tilewrap_test.go:TestTileHeaderCannotWrap`;
  `web/tilewrap_test.go:TestBandingCodeIsOutsideTheTruncatingElement`
Gates: yes, in `go test ./...` (CI and image build)
Watched failing: 2026-10-30, by dropping `flex-wrap: nowrap` from `.idline` and from `.ft`.
Watched failing: 2026-10-31, by moving the banding code inside the tile's `.cnname` ("a code is inside the truncating name on its .nameline line").
Watched failing: 2026-10-02, by removing `flex-wrap: nowrap` from `.nameline`, which is the shipped bug; by dropping `white-space: nowrap` from `.cnname`; by dropping `flex: none` from `.code`; by moving the code inside `.cnname` in the template, where the markup assertion catches what no stylesheet check can see; and by reverting `.badge` to its wrapping form.

## app.regionFilterBehaves

Criterion: [app.regionFilterBehaves](requirements/composed/internal/app/acceptance.md#regionfilterbehaves)
Guardrails: `cmd/birdview/main_test.go:TestRegionFilter`
Gates: yes, in `go test ./...` (CI and image build)
Watched failing: 2026-10-03, by ignoring the parameter.
Watched failing: 2026-10-03, by treating an empty `region=` as no filter, which is the bug the family filter had first.
Watched failing: 2026-10-03, by moving Pelagic out of last place.

## app.habitatGroupsFilter

Criterion: [app.habitatGroupsFilter](requirements/composed/internal/app/acceptance.md#habitatgroupsfilter)
Guardrails: `cmd/birdview/next_test.go:TestHabitatGroupsFilter`;
  `internal/trait/review_test.go:TestHabitatGroupsPartitionAVONET`;
  `internal/trait/review_test.go:TestDietGroupsPartitionNiches`
Gates: yes, in `go test ./...` (CI and image build)
Watched failing: 2026-11-07, with the filter matching AVONET's values instead of their groups (eight failures, among them "Coast and sea omits the heron"), and before the groups existed.
Watched failing: 2026-11-09, with Diet as one line ("Diet is not All, a rule, and its choices on one line"), and with Drinks sap left out of Plants ("the niche Sap is in no diet group"; "Diet offers … want … Drinks sap").
Watched failing: 2026-11-07, with the leads in a column: against the one-line layout, "a Habitat line has no lead". The alignment itself was checked by screenshot at 900, 560, and 400 pixels wide. The leads aligned right, with the rule running beside Towns and farmland too, were checked by screenshot at 560 pixels wide on 2026-11-07.

## app.allAlwaysOffered

Criterion: [app.allAlwaysOffered](requirements/composed/internal/app/acceptance.md#allalwaysoffered)
Guardrails: `cmd/birdview/next_test.go:TestAllAlwaysOffered`
Gates: yes, in `go test ./...` (CI and image build)
Watched failing: 2026-11-07, with All hidden when its setting is all and How common's All link restored: "County offers All 0 times, want once" for County, Family, Habitat, and Diet, and "How common offers All 2 times, want once".

## dataset.habitatReviewComplete

Criterion: [dataset.habitatReviewComplete](requirements/composed/internal/dataset/acceptance.md#habitatreviewcomplete)
Guardrails: `cmd/birdview-ingest/main_test.go:TestEveryDrawnBirdHasAHabitatReview`;
  `internal/trait/review_test.go:TestHabitatReviewParsesAndApplies`
Gates: yes, in `go test ./...` (CI and image build)
Watched failing: 2026-11-07, with Least Sandpiper's row removed ("Calidris minutilla is drawn and has no row in data/habitat-review.tsv") and American Robin's AVONET habitat edited ("AVONET's habitat is \"Forest\", but the review was of \"Desert\"; review the row again"), and before the parser existed.

## app.countyListingApplied

Criterion: [app.countyListingApplied](requirements/composed/internal/app/acceptance.md#countylistingapplied)
Guardrails: `cmd/birdview-ingest/main_test.go:TestCountiesJoinedAsRegions`;
  `cmd/birdview-regions/county_test.go:TestCountyListingRule`;
  `cmd/birdview-regions/county_test.go:TestCountyRowsFromDownload`;
  `cmd/birdview-regions/county_test.go:TestCountiesSplitByState`; `cmd/birdview/main_test.go:TestCountyIsARegion`;
  `cmd/birdview/main_test.go:TestCountySettingDrillsDown`; `cmd/birdview/main_test.go:TestCountiesOffByDefault`;
  `cmd/birdview/main_test.go:TestPlaceAndMonthsAreNotJoinedByAComma`;
  `internal/region/region_test.go:TestCountyParentFromCode`
Gates: yes, in `go test ./...` (CI and image build)
Watched failing: 2026-11-03, with the state-list condition removed ("Nymphicus hollandicus in USA.33.24_1: listed true … want listed false"), and before a state's withholding covered its counties ("a state's entry must cover its own counties").
Watched failing: 2026-11-03, before the server knew counties ("Los Angeles draws [amerob amecro annhum thimur blasti greblu]", "the heading reads \"Showing birds in every month\"") and while "every region" summed a county with its state ("1 in 1,461 with the county and 1 in 2,275 without"); and before the County setting existed ("choosing California, which has counties, shows no County setting").
Watched failing: 2026-11-04, with a county's months after a comma ("the hummingbird's share line does not read \"in Los Angeles, CA (Jun)\"") and a list of counties joined by commas ("five counties list as \"Bronx, NY, Kings, NY, New York, NY, or 2 more\"").
Watched failing: 2026-11-12, with each county's rows filed under its own code instead of its state's ("split into 4 states, want 3"; "USA.33_1 holds [], want [USA.33.24_1 USA.33.42_1]"; "a state's code in the county column was accepted as a county").

## app.everydayDefinitionsShown

Criterion: [app.everydayDefinitionsShown](requirements/composed/internal/app/acceptance.md#everydaydefinitionsshown)
Guardrails: `cmd/birdview/main_test.go:TestTooltipsSayWhatEachMeasures`
Gates: yes, in `go test ./...` (CI and image build)
Watched failing: 2026-11-03, against tooltips that named only the source ("Chord's tooltip is \"measured, from AVONET v7, sheet AVONET2_eBird\"; it begins with its definition"), the table likewise, and a guide without the definitions ("the guide does not define Wingspan as …").

## dataset.gbifDownloadsRecorded

Criterion: [dataset.gbifDownloadsRecorded](requirements/composed/internal/dataset/acceptance.md#gbifdownloadsrecorded)
Guardrails: `go test ./internal/gbif/ -run TestDownloadSubmitsWaitsAndRecords`
Gates: yes, in `go test ./...`
Watched failing: 2026-11-03, with the query check disabled ("a download was recorded under a query it was not requested from"), and before Validate returned GBIF's formatting.

## app.regionUrlRoundTrips

Criterion: [app.regionUrlRoundTrips](requirements/composed/internal/app/acceptance.md#regionurlroundtrips)
Guardrails: `cmd/birdview/main_test.go:TestRegionURLsCarryDataCodes`;
  `internal/region/region_test.go:TestSlugsAreUniqueAndResolve`;
  `internal/region/region_test.go:TestSlugIsDataCode`; `internal/region/region_test.go:TestRetiredValuesResolve`
Gates: yes, in `go test ./...` (CI and image build)
Watched failing: 2026-10-03, by returning a stored slug instead of deriving it, so a renamed region kept its old URL; by emitting GADM ids in links again; by giving Palmyra the code `PM`, which the no-invented-code assertion rejects; and by rendering the empty code, which puts a leading separator on ten chips.
Watched failing: 2026-11-03, against postal-code URLs ("USA.5_1 (California) has the URL value \"CA\", want \"USA.5_1\"") and with no retired lookup ("the retired value \"CA\" resolves to (false), want USA.5_1", 64 postal codes not resolving), and with the home page's redirect disabled ("GET /?region=CA&by=mass = 200 to \"\", want 301 to /?by=mass&region=USA.5_1").

## app.familyGroupingComplete

Criterion: [app.familyGroupingComplete](requirements/composed/internal/app/acceptance.md#familygroupingcomplete)
Guardrails: `cmd/birdview/main_test.go:TestTileCarriesExactlyOneBadge`
Gates: yes, in `go test ./...` (CI and image build)
Watched failing: 2026-10-03, by adding an order badge beside the family badge, which makes 6 birds carry 12 badges.

## app.creditsComplete

Criterion: [app.creditsComplete](requirements/composed/internal/app/acceptance.md#creditscomplete)
Guardrails: `web/docs_test.go:TestEveryPageCreditsWhatItShows`; `web/docs_test.go:TestDocsCountsMatchTheDataset`;
  `web/docs_test.go:TestNoDocCreditsASourceForWhatItNoLongerSupplies`
Gates: yes, in `go test ./...` (CI and image build)
Watched failing: 2026-10-03, first failed against the real defects: the missing AviList credit on `detail.html` and `compare.html`, and `README.md` and `docs/building.md` crediting the ABA Checklist with names and families.
Watched failing: 2026-10-03, again after the fixes, by removing the AviList credit, by restoring the 841 count, and by re-crediting the ABA.

## app.categoryOriginShown

Criterion: [app.categoryOriginShown](requirements/composed/internal/app/acceptance.md#categoryoriginshown)
Guardrails: `cmd/birdview/main_test.go:TestDetailListsEveryCategoryWithItsOrigin`;
  `cmd/birdview/main_test.go:TestServedDocumentsHaveNoBrokenLinks`;
  `cmd/birdview/main_test.go:TestVendoredTermsAreNotServed`;
  `cmd/birdview/main_test.go:TestDocumentTitlesAreHeadings`;
  `cmd/birdview/main_test.go:TestAddedListsEveryUnpublishedAssignment`;
  `cmd/birdview/main_test.go:TestCitedAnchorsResolve`
Gates: yes, in `go test ./...` (CI and image build)
Watched failing: 2026-10-16, by planting a misspelled repository link in a guide — `river/decisons.md does not exist in the repository` — and, for the coverage [app.noContactAddressPublished](requirements/composed/internal/app/acceptance.md#nocontactaddresspublished) would otherwise have lost, by planting an address in `river/goals/goals.md`.
Watched failing: 2026-10-03, by giving each region its own row.
Watched failing: 2026-10-03, by rendering a category with an empty origin.

## dataset.habitatDietFilterPerSpecies

Criterion: [dataset.habitatDietFilterPerSpecies](requirements/composed/internal/dataset/acceptance.md#habitatdietfilterperspecies)
Guardrails: `internal/trait/trait_test.go:TestSupplementaryKeysExist`;
  `internal/trait/trait_test.go:TestSupplementaryReachesTheDataset`;
  `internal/trait/trait_test.go:TestMixedDietRuleReachesTheDataset`;
  `internal/trait/trait_test.go:TestEveryAssignmentHasAnOriginAndTheVocabularyKnowsIt`;
  `internal/trait/trait_test.go:TestLabelsAreNotJargon`;
  `internal/trait/trait_test.go:TestCoinedTermsAreRareAndSaySo`;
  `internal/trait/trait_test.go:TestCoinedTermsAreNeverPublished`
Gates: yes, in `go test ./...` (CI and image build)
Watched failing: 2026-10-03, by coining two further terms for niches AVONET already covers, which trips the cap; by assigning the coined `Sap` category with origin `published`; and by the two supplementary keys that were dead on arrival: `Phalacrocorax auritus` and `Larus argentatus`, where AviList says `Nannopterum auritum` and `Larus smithsonianus`.

## app.staticServesAssetsOnly

Criterion: [app.staticServesAssetsOnly](requirements/composed/internal/app/acceptance.md#staticservesassetsonly)
Guardrails: `cmd/birdview/main_test.go:TestStaticServesAssetsOnly`
Gates: yes, in `go test ./...` (CI and image build)
Watched failing: 2026-10-16, against the code as found, every template returned 200 with its source, `/static/` returned `index.html`'s, and `/static/index.html` redirected to it.

## app.robotsKeepsCrawlersToPages

Criterion: [app.robotsKeepsCrawlersToPages](requirements/composed/internal/app/acceptance.md#robotskeepscrawlerstopages)
Guardrails: `cmd/birdview/robots_test.go:TestRobotsKeepsCrawlersToPages`
Gates: yes, in `go test ./...` (CI and image build)
Watched failing: 2026-10-23, before the route existed, `GET /robots.txt` answered 404; with `Disallow: /*?` deleted, it reported each filtered view of the index as allowed.

## app.stableUrlsKeepWorking

Criterion: [app.stableUrlsKeepWorking](requirements/composed/internal/app/acceptance.md#stableurlskeepworking)
Guardrails: `cmd/birdview/stable_test.go:TestStableUrlsKeepWorking`
Gates: yes, in `go test ./...` (CI and image build)
Watched failing: 2026-10-23, with the index reading `diet` where it read `niche`, it reported `?niche=Frugivore` as no longer read.

## app.canonicalHostRedirects

Criterion: [app.canonicalHostRedirects](requirements/composed/internal/app/acceptance.md#canonicalhostredirects)
Guardrails: `cmd/birdview/stable_test.go:TestCanonicalHostRedirects`
Gates: yes, in `go test ./...` (CI and image build)
Watched failing: 2026-10-23, before the redirect existed, every other host answered 200.

## repo.checksWorkflowIsSound

Criterion: [repo.checksWorkflowIsSound](requirements/composed/internal/repo/acceptance.md#checksworkflowissound)
Guardrails: `cmd/birdview/release_test.go:TestChecksWorkflow`
Gates: yes, in `go test ./...` (CI and image build)
Watched failing: 2026-10-23, with `actions/checkout` pinned to the tag `v7` instead of a commit, it reported the unpinned action.

## repo.dockerTestStageGates

Criterion: [repo.dockerTestStageGates](requirements/composed/internal/repo/acceptance.md#dockerteststagegates)
Guardrails: `cmd/birdview/release_test.go:TestDockerTestStageGatesTheBuild`
Gates: yes, in `go test ./...` (CI and image build)
Watched failing: 2026-10-23, with the builder's `COPY --from=test` deleted, it reported that nothing makes the tests run.

## dataset.duckAndGrebeProportions

Criterion: [dataset.duckAndGrebeProportions](requirements/composed/internal/dataset/acceptance.md#duckandgrebeproportions)
Guardrails: `internal/morphology/morphology_test.go:TestDivingDucksAndGrebes`
Gates: yes, in `go test ./...` (CI and image build)
Watched failing: 2026-11-08, with the hooks in place but no values: "Lesser Scaup's neck is 0.26 of its length and Horned Grebe's 0.19; the scaup's must be the shorter", and five more, matching issue #7.

## dataset.proportionsReviewTable

Criterion: [dataset.proportionsReviewTable](requirements/composed/internal/dataset/acceptance.md#proportionsreviewtable)
Guardrails: `cmd/birdview-ingest/main_test.go:TestNeckReviewAppliesToDrawnTaxa`;
  `cmd/birdview-ingest/main_test.go:TestReviewFactorsReachTheBuild`;
  `internal/morphology/review_test.go:TestProportionReviewParsesAndApplies`
Gates: yes, in `go test ./...` (CI and image build)
Watched failing: 2026-11-09, for body and wingspan: before BodyLengthFor existed (the test did not build), and with the wingspan factor never applied ("Clark's Grebe's modeled wingspan is 829.8 (factor 0) … want 1.1 times the model's 829.8").
Watched failing: 2026-11-09, before the parser existed (the tests did not build: "undefined: Review"), and with the ingest passing no table ("Lesser Scaup's neck model reads \"neck modeled as 0.70 x trunk (family-specific ratio for Anatidae)\"; want the Aythya row").

## dataset.wingspanErrorStated

Criterion: [dataset.wingspanErrorStated](requirements/composed/internal/dataset/acceptance.md#wingspanerrorstated)
Guardrails: `internal/wingspan/model_test.go:TestEveryFitStatesItsError`
Gates: yes, in `go test ./...` (CI and image build)
Watched failing: 2026-11-09, before the fix: "the fit for Gaviiformes_Podicipediformes (9 species) states no error", and likewise for Otidiformes (7) and Piciformes (8); "Piciformes's wingspan reads … cross-validated median error NaN%".

## repo.activeTodoHasIssue

Criterion: [repo.activeTodoHasIssue](requirements/composed/internal/repo/acceptance.md#activetodohasissue)
Guardrails: `river/check/birdview_test.go:TestActiveTodoHasIssue`
Gates: yes, in `go test ./...` (CI and image build)
Watched failing: 2026-11-09, with countyRegions marked active before its issue existed (the test reported countyRegions as active with no Issue: line linking its GitHub issue).

## dataset.seaDistanceTable

Criterion: [dataset.seaDistanceTable](requirements/composed/internal/dataset/acceptance.md#seadistancetable)
Guardrails: `cmd/birdview-sea/main_test.go:TestMeasureBandsWaterRecords`;
  `internal/geo/geo_test.go:TestShoreLandAndDistance`
Gates: yes, in `go test ./...` (CI and image build)
Watched failing: 2026-11-09, before the geometry existed (the test did not build: "undefined: NewShore"), and with the land test skipped in the generator ("Cory's Shearwater's bands are [3 5 18], want [3 5 7]").

## dataset.seaRefinementsDerived

Criterion: [dataset.seaRefinementsDerived](requirements/composed/internal/dataset/acceptance.md#searefinementsderived)
Guardrails: `cmd/birdview-ingest/main_test.go:TestSeaRefinementsDerived`
Gates: yes, in `go test ./...` (CI and image build)
Watched failing: 2026-11-09, before the step existed (the test did not build: "undefined: deriveSea"), and without the 80% rule ("Phaethon lepturus is [Marine-offshore], want \"Marine-inshore,Marine-offshore\"").

## app.openOceanFilter

Criterion: [app.openOceanFilter](requirements/composed/internal/app/acceptance.md#openoceanfilter)
Guardrails: `cmd/birdview/next_test.go:TestOpenOceanFilter`
Gates: yes, in `go test ./...` (CI and image build)
Watched failing: 2026-11-09, before the server knew the refinements ("GET /?habitat=Marine&by=mass = 200 to \"\", want 301"; "Habitat offers … Marine-inshore Marine-offshore"; "headed … Habitat: Marine-inshore or Marine-offshore").

## app.birdPresets

Criterion: [app.birdPresets](requirements/composed/internal/app/acceptance.md#birdpresets)
Guardrails: `cmd/birdview/presets_test.go:TestBirdPresets`;
  `cmd/birdview/presets_test.go:TestEveryPresetDrawsBirds`
Gates: yes, in `go test ./...` (CI and image build)
Watched failing: 2026-11-09, before the setting existed (the tests did not build: "undefined: birdPresets"), and with the Shorebirds check reading an address that still named New York ("Shorebirds must draw the stilt and not the heron", a fault in the test, corrected).

## app.chooseOneSettings

Criterion: [app.chooseOneSettings](requirements/composed/internal/app/acceptance.md#chooseonesettings)
Guardrails: `cmd/birdview/chooseone_test.go:TestChooseOneSettings`
Gates: yes, in `go test ./...` (CI and image build)
Watched failing: 2026-11-09, before the joined row existed ("Birds draws no joined row", and likewise for the other four), and with the removed words still drawn ("Birds still says Choose one, which the owner removed").

## app.actionsColoredApart

Criterion: [app.actionsColoredApart](requirements/composed/internal/app/acceptance.md#actionscoloredapart)
Guardrails: `cmd/birdview/chooseone_test.go:TestActionsColoredApart`
Gates: yes, in `go test ./...` (CI and image build)
Watched failing: 2026-11-09, without the class ("Now is not colored apart as an action", "Here is not colored apart as an action").

## repo.imageRecordsRevision

Criterion: [repo.imageRecordsRevision](requirements/composed/internal/repo/acceptance.md#imagerecordsrevision)
Guardrails: `cmd/birdview/dockerfile_test.go:TestImageRecordsRevision`
Gates: yes, in `go test ./...` (CI and image build)
Watched failing: 2026-11-08, in the first deploy's image build, where `.gcloudignore` had left itself out of the upload (".gcloudignore must include .gitignore and re-include !/REVISION … it reads:" with nothing after it); the deploy stopped, as [tech.testedBeforeDeploy](requirements/local/tech/release.md#testedbeforedeploy) intends.
Watched failing: 2026-11-08, before the Dockerfile and ignore files changed ("stage \"builder\" copies [go.mod go.sum cmd/birdview internal data web docs]; want REVISION*", and three more).

## repo.smokeChecksDeployment

Criterion: [repo.smokeChecksDeployment](requirements/composed/internal/repo/acceptance.md#smokechecksdeployment)
Guardrails: `cmd/birdview/smoke_test.go:TestSmokeChecksTheServer`
Gates: yes, in `go test ./...` (CI and image build)
Watched failing: 2026-11-08, the revision check before it existed ("with revision not recorded, check \"the server names its revision\" passed").
Watched failing: 2026-10-23, by the broken variants above, which the test keeps.

## app.massInOneUnit

Criterion: [app.massInOneUnit](requirements/composed/internal/app/acceptance.md#massinoneunit)
Guardrails: `cmd/birdview/main_test.go:TestUnitsForTakesTheSmallestBird`;
  `cmd/birdview/main_test.go:TestNoSectionMixesMassUnits`
Gates: yes, in `go test ./...` (CI and image build)
Watched failing: 2026-10-04, by choosing the unit from the largest mass instead of the smallest — a 27 g House Sparrow beside a Trumpeter Swan then renders `0.1 lb`, one significant figure for a whole bird — and by reverting the index to a per-value unit, which makes sections 36, 37, and 38 render both `oz` and `lb`.

## dataset.neckRatiosReviewed

Criterion: [dataset.neckRatiosReviewed](requirements/composed/internal/dataset/acceptance.md#neckratiosreviewed)
Guardrails: `internal/morphology/morphology_test.go:TestEveryFamilyHasBeenReviewed`
Gates: yes, in `go test ./...` (CI and image build)
Watched failing: 2026-10-04, by removing Stercorariidae from `NeckRatios`, by removing Procellariidae from `NeckDefaultReviewed`, and by listing Alcidae in both.

## app.commonestShareIsShare

Criterion: [app.commonestShareIsShare](requirements/composed/internal/app/acceptance.md#commonestshareisshare)
Guardrails: `cmd/birdview-ingest/main_test.go:TestRegionCountsLineUp`;
  `cmd/birdview/main_test.go:TestTileShareFollowsTheRegionFilter`
Gates: yes, in `go test ./...` (CI and image build)
Watched failing: 2026-10-05, by ignoring the region filter in `shareOf`, which makes the Alaska and everywhere figures identical; by appending one extra count in `applyRegions`, which desyncs the parallel slices; by removing the heading-id transformer, which reports `#scopeisrecurringavifauna` and `#avilistreplacesaba` dangling, and by renaming the `Commonest` heading in `docs/data.md`.

## app.reportDataIssueLinks

Criterion: [app.reportDataIssueLinks](requirements/composed/internal/app/acceptance.md#reportdataissuelinks)
Guardrails: `cmd/birdview/report_test.go:TestReportDataIssueLinks`;
  `cmd/birdview/report_test.go:TestHomeReportLink`; `cmd/birdview/report_test.go:TestEveryReportLinkFilledIn`;
  `cmd/birdview/report_test.go:TestFeedbackMustBeFillable`
Gates: yes, in `go test ./...` (CI and image build)
Watched failing: 2026-11-08, for every page: the no-longer-drawn pages still linked to the bare tracker ("/species/paspig links to the tracker as https://github.com/google/birdview/issues"); and before checkFeedback existed (the test did not build: "undefined: checkFeedback").
Watched failing: 2026-11-08, for the home page and the revision: before either was filled in ("Measurements and sources's issue body lacks \"BirdView server: 0123456789ab\"").
Watched failing: 2026-11-08, before the links existed ("the species page has 0 Report a data issue links, want 2"), and with the fixture's source versions missing ("Measurements and sources's issue body lacks \"AVONET v7\"").

## app.noContactAddressPublished

Criterion: [app.noContactAddressPublished](requirements/composed/internal/app/acceptance.md#nocontactaddresspublished)
Guardrails: `cmd/birdview/main_test.go:TestNoPersonalAddressIsPublished`;
  `cmd/birdview/main_test.go:TestInterfaceNeverSaysShare`
Gates: yes, in `go test ./...` (CI and image build)
Watched failing: 2026-10-05, by restoring the old header sentence, which reports it on every page that renders the header.

## app.outOfScopeExplained

Criterion: [app.outOfScopeExplained](requirements/composed/internal/app/acceptance.md#outofscopeexplained)
Guardrails: `cmd/birdview/main_test.go:TestUndrawnSpeciesNamesItselfAt404`;
  `cmd/birdview/main_test.go:TestNoIDIsBothDrawnAndNot`;
  `cmd/birdview/main_test.go:TestAbsentSpeciesCarryNoEvidence`
Gates: yes, in `go test ./...` (CI and image build)
Watched failing: 2026-10-05, across 11 mutations (---; `detail` restored to a bare `http.NotFound`; the handler writes 200; the handler writes 410; and 7 more mutations).

## app.commonestIsLargestShare

Criterion: [app.commonestIsLargestShare](requirements/composed/internal/app/acceptance.md#commonestislargestshare)
Guardrails: `cmd/birdview-ingest/main_test.go:TestRegionCountsLineUp`;
  `cmd/birdview/main_test.go:TestCommonestIsTheLargestShare`;
  `cmd/birdview/main_test.go:TestShareRangesOverOneSetOfRegions`;
  `cmd/birdview/main_test.go:TestShareStatesZeroRatherThanDashing`;
  `cmd/birdview/main_test.go:TestMissingTotalsDoNotClaimTheBirdIsAbsent`;
  `cmd/birdview/main_test.go:TestInterfaceNeverSaysShare`
Gates: yes, in `go test ./...` (CI and image build)
Watched failing: 2026-10-05, across 10 mutations (---; `shareIn` back to two independent loops; a sea-only selection falls back to the em-dash; the ingest carries the marine totals again — the shipped bug; and 6 more mutations).

## app.peakAndShareOneCalculation

Criterion: [app.peakAndShareOneCalculation](requirements/composed/internal/app/acceptance.md#peakandshareonecalculation)
Guardrails: `cmd/birdview/main_test.go:TestCommonestRoundTripsThroughTheIndex`
Gates: yes, in `go test ./...` (CI and image build)
Watched failing: 2026-10-05, across 6 mutations (---; `Commonest` taken from `status.tsv` again — the reported bug, exactly; `CountOffLandInBox` returns the box count instead of the off-land count; the multi-GADM-level guard removed; and 2 more mutations).

## app.responsesCompress

Criterion: [app.responsesCompress](requirements/composed/internal/app/acceptance.md#responsescompress)
Guardrails: `cmd/birdview/main_test.go:TestResponsesCompressWhenAsked`
Gates: yes, in `go test ./...` (CI and image build)
Watched failing: 2026-10-05, across 4 mutations (---; label from `Write` rather than `WriteHeader` — the shipped bug; compress regardless of `Accept-Encoding`; drop the `Vary` header).

## app.frequencyTiers

Criterion: [app.frequencyTiers](requirements/composed/internal/app/acceptance.md#frequencytiers)
Guardrails: `cmd/birdview/main_test.go:TestFrequencyOrdersAndBandsByDecade`;
  `cmd/birdview/main_test.go:TestFrequencyTiersAreNotNamed`
Gates: yes, in `go test ./...` (CI and image build)
Watched failing: 2026-10-05, across 10 mutations (---; band proportionally (`min × 10`) rather than by decade; restore `Inf >= Inf` in `SectionsBy`; name a tier `Common — 1 in …`; and 6 more mutations).

## app.sharedCountDisclosed

Criterion: [app.sharedCountDisclosed](requirements/composed/internal/app/acceptance.md#sharedcountdisclosed)
Guardrails: `cmd/birdview/main_test.go:TestSharedNamesAreDisclosed`
Gates: yes, in `go test ./...` (CI and image build)
Watched failing: 2026-10-05, across 3 mutations (---; the tile marker is suppressed, leaving only the detail page; a withheld species is labeled as a shared name).

## dataset.regionJoinReportsUnmatched

Criterion: [dataset.regionJoinReportsUnmatched](requirements/composed/internal/dataset/acceptance.md#regionjoinreportsunmatched)
Guardrails: none yet (planned)
Gates: no
Watched failing: 2026-10-05, across 3 mutations (---; `Occurrence` skips `AviListRenames` — the shipped bug; `Occurrence` folds the shared name into *aestiva*).

## dataset.disagreementsDeclared

Criterion: [dataset.disagreementsDeclared](requirements/composed/internal/dataset/acceptance.md#disagreementsdeclared)
Guardrails: `internal/taxonomy/correspondence_test.go:TestSourceNamesAgree`
Gates: yes, in `go test ./...` (CI and image build)
Watched failing: 2026-10-05, across 8 mutations (direction; ---; iNaturalist out-records eBird; eBird dwarfs iNaturalist; and 4 more mutations).

## dataset.occurrenceReadersAgree

Criterion: [dataset.occurrenceReadersAgree](requirements/composed/internal/dataset/acceptance.md#occurrencereadersagree)
Guardrails: `cmd/birdview-status/main_test.go:TestBothReadersOfTheOccurrenceDataAgreeOnTheSpecies`;
  `internal/taxonomy/mapping_test.go:TestBinomialReducesRankWithoutInventingSpecies`
Gates: yes, in `go test ./...` (CI and image build)
Watched failing: 2026-10-06, across 5 mutations (---; `Occurrence` does not reduce rank — the state before this change; `Occurrence` consults the tables *before* reducing rank; the hybrid guard is removed from `Binomial`; and 1 more mutations).

## app.figureSaysWhatItCounts

Criterion: [app.figureSaysWhatItCounts](requirements/composed/internal/app/acceptance.md#figuresayswhatitcounts)
Guardrails: `cmd/birdview/main_test.go:TestTheUnitSaysSpeciesAndNotBird`
Gates: yes, in `go test ./...` (CI and image build)
Watched failing: 2026-10-06, across 3 mutations (---; the tile template reverts to `bird observations`; the tooltip reverts while the tile does not).

## app.detailSaysWhatTileSays

Criterion: [app.detailSaysWhatTileSays](requirements/composed/internal/app/acceptance.md#detailsayswhattilesays)
Guardrails: `cmd/birdview/main_test.go:TestTheDetailPageSaysEverythingTheTileSays`
Gates: yes, in `go test ./...` (CI and image build)
Watched failing: 2026-10-07, across 4 mutations (---; the `Frequency` row is not emitted — the state before this change; the row is built from `options{}`, ignoring the reader's selection; the value drops `cell.Where`, printing the number without its scope).
Watched failing: 2026-11-01, when ids moved the sample onto Ferruginous Pygmy-Owl, whose tile note ("Withheld by eBird in Arizona, all year") the page carries in lower case as a row continuing "Recorded as"; the comparison now ignores case, and still fails on a real difference in wording (the page without its comma).

## app.disputedNamesShown

Criterion: [app.disputedNamesShown](requirements/composed/internal/app/acceptance.md#disputednamesshown)
Guardrails: `cmd/birdview/main_test.go:TestEveryNameDisagreementIsOnThePage`
Gates: yes, in `go test ./...` (CI and image build)
Watched failing: 2026-10-07, across 4 mutations (---; the alias row is not emitted — two of the four kinds, as shipped until now; the lump row is not emitted; a row is emitted for every species).

## app.honestRevalidation

Criterion: [app.honestRevalidation](requirements/composed/internal/app/acceptance.md#honestrevalidation)
Guardrails: `cmd/birdview/main_test.go:TestResponsesCarryAValidator`
Gates: yes, in `go test ./...` (CI and image build)
Watched failing: 2026-10-07, across 4 mutations (---; no `Cache-Control` is set — the state before this change; the ETag is built from the build alone, ignoring the URL; the fingerprint is a constant rather than a hash of the build).

## app.nameDisclosureLegible

Criterion: [app.nameDisclosureLegible](requirements/composed/internal/app/acceptance.md#namedisclosurelegible)
Guardrails: `cmd/birdview/main_test.go:TestTheNameDisclosureExplainsItselfAndLinksSomewhereUseful`
Gates: yes, in `go test ./...` (CI and image build)
Watched failing: 2026-10-07, across 4 mutations (---; a boundary entry reverts to one vague sentence calling itself a lump; the link points back at `#how-common-a-bird-is`; the epithet case reverts to the single "moved" sentence).

## app.monthsFilterAndAbstain

Criterion: [app.monthsFilterAndAbstain](requirements/composed/internal/app/acceptance.md#monthsfilterandabstain)
Guardrails: `cmd/birdview/main_test.go:TestMonthScopeWraps`;
  `cmd/birdview/main_test.go:TestAllTwelveMonthsIsTheWholeYear`;
  `cmd/birdview/main_test.go:TestMonthsNarrowTheListAndTheFigureTogether`;
  `cmd/birdview/main_test.go:TestAbstainedTilesSayWhy`;
  `internal/month/month_test.go:TestPresentRecoversKnownSeasons`;
  `internal/month/month_test.go:TestPresentAbstainsBelowTheFloor`
Gates: yes, in `go test ./...` (CI and image build)
Watched failing: 2026-10-07, across 5 mutations (---; the ratio drops its effort correction, becoming a raw share of the species' own records; the 200-record floor is removed; the month filter stops narrowing the list; and 1 more mutations).
Watched failing: 2026-10-07, on the old wording (`Millerbird is shown ...

## app.withheldMonthsExplained

Criterion: [app.withheldMonthsExplained](requirements/composed/internal/app/acceptance.md#withheldmonthsexplained)
Guardrails: `cmd/birdview/main_test.go:TestWithheldMonthsAreNotJudged`;
  `internal/region/region_test.go:TestPresenceReadsWithheldFlags`;
  `internal/taxonomy/sensitive_test.go:TestEBirdSensitiveIsWellFormed`;
  `internal/taxonomy/sensitive_test.go:TestWindowsCoverWholeMonthsOnly`;
  `internal/taxonomy/sensitive_test.go:TestConfirmedMonthsBelievesTheRecordsOverTheList`
Gates: yes, in `go test ./...` (CI and image build)
Watched failing: 2026-10-09, across 6 mutations (---; the month filter judges withheld months again — the bug as shipped; the caveat ignores the month selection; confirmation is skipped and the list trusted; and 2 more mutations).

## dataset.ebirdCountedInatFillsGaps

Criterion: [dataset.ebirdCountedInatFillsGaps](requirements/composed/internal/dataset/acceptance.md#ebirdcountedinatfillsgaps)
Guardrails: `cmd/birdview-regions/main_test.go:TestMeasureCountsEBirdAndFillsOnlyItsDeclaredGaps`;
  `cmd/birdview-status/main_test.go:TestScopeCountsEBirdAndFillsOnlyItsDeclaredGaps`;
  `internal/status/status_test.go:TestScopeBringsBothHalvesOfASharedName`
Gates: yes, in `go test ./...` (CI and image build)
Watched failing: 2026-10-12, across 16 mutations (; ---; land presences; species placed in a region; and 12 more mutations).

## app.pipelineDiagramServed

Criterion: [app.pipelineDiagramServed](requirements/composed/internal/app/acceptance.md#pipelinediagramserved)
Guardrails: `cmd/birdview/diagram_test.go:TestDiagramsAreCurrent`;
  `cmd/birdview/diagram_test.go:TestDiagramsAreSelfContained`;
  `cmd/birdview/diagram_test.go:TestPipelineDiagramIsServed`;
  `cmd/birdview/diagram_test.go:TestPipelineDiagramNamesRealColumns`;
  `cmd/birdview/diagram_test.go:TestPipelineDiagramNamesEveryTaxonomyTable`;
  `cmd/birdview/diagram_test.go:TestPipelineDiagramIsLinked`;
  `cmd/birdview/diagram_test.go:TestPipelineFieldsCarryOneMark`;
  `cmd/birdview/diagram_test.go:TestPipelineTablesAreLeftAligned`
Gates: yes, in `go test ./...` (CI and image build)
Watched failing: 2026-10-12, first failed on its own diagram, on a field label wrapped onto a second line; one field per line is now the rule it relies on.
Watched failing: 2026-10-12, across 22 mutations (---; a field label edited, the SVG not re-rendered; the SVGs dropped from the embed; a `<script>` added to the SVG; and 18 more mutations).

## dataset.whiteTernFromBirdlife

Criterion: [dataset.whiteTernFromBirdlife](requirements/composed/internal/dataset/acceptance.md#whiteternfrombirdlife)
Guardrails: `cmd/birdview-ingest/main_test.go:TestNearMatchBorrowsFromTheDeclaredSheet`
Gates: yes, in `go test ./...` (CI and image build)
Watched failing: 2026-10-13, across 5 mutations (---; the `Gygis candida` entry removed from `NearMatchSheets`; `resolve` ignoring the declared sheet; the missing-row check in `checkNearMatchSheets` disabled; and 1 more mutations).

## app.sharedMeasurementsDisclosed

Criterion: [app.sharedMeasurementsDisclosed](requirements/composed/internal/app/acceptance.md#sharedmeasurementsdisclosed)
Guardrails: `cmd/birdview-ingest/main_test.go:TestVegaGullSharesAmericanHerringGullsRow`;
  `cmd/birdview/main_test.go:TestSameMeasurementsAreNoted`;
  `cmd/birdview/main_test.go:TestSameMeasurementsNeedNoTable`
Gates: yes, in `go test ./...` (CI and image build)
Watched failing: 2026-10-13, across 7 mutations (---; `sharedRows` replaced by an empty map; rows grouped by sheet alone, not sheet and taxon; the tile's note deleted from `index.html`; and 3 more mutations).

## dataset.herringGullFromEuropeanRow

Criterion: [dataset.herringGullFromEuropeanRow](requirements/composed/internal/dataset/acceptance.md#herringgullfromeuropeanrow)
Guardrails: `cmd/birdview-ingest/main_test.go:TestEuropeanHerringGullReadsItsOwnRow`
Gates: yes, in `go test ./...` (CI and image build)
Watched failing: 2026-10-13, across 5 mutations (---; the `Larus argentatus` entry removed from `OwnRowSheets`; `resolve` ignoring the declared sheet; the missing-row check in `checkOwnRowSheets` disabled; and 1 more mutations).

## app.belowBarCountsKept

Criterion: [app.belowBarCountsKept](requirements/composed/internal/app/acceptance.md#belowbarcountskept)
Guardrails: `cmd/birdview-ingest/main_test.go:TestBelowBarRegionsAreCountedNotListed`;
  `cmd/birdview-regions/below_test.go:TestBelowBarCellsRoundTrip`;
  `cmd/birdview/main_test.go:TestRecordsBelowTheBarCount`
Gates: yes, in `go test ./...` (CI and image build)
Watched failing: 2026-10-13, across 6 mutations (---; none: the real-data test run against the table built before this change; the writer dropping the `~` mark; the reader ignoring the mark; and 2 more mutations).

## app.escapedBirdsListedAndServed

Criterion: [app.escapedBirdsListedAndServed](requirements/composed/internal/app/acceptance.md#escapedbirdslistedandserved)
Guardrails: `cmd/birdview-ingest/main_test.go:TestExoticsAreFlaggedFromTheList`;
  `cmd/birdview/main_test.go:TestExoticsAreOffTheIndexAndListed`
Gates: yes, in `go test ./...` (CI and image build)
Watched failing: 2026-10-14, across 7 mutations (---; the index view keeping the exotics; families of flagged birds only still offered; `/excluded` reading the index view instead of the whole dataset; and 3 more mutations).

## app.readersToldStickFigure

Criterion: [app.readersToldStickFigure](requirements/composed/internal/app/acceptance.md#readerstoldstickfigure)
Guardrails: `cmd/birdview/main_test.go:TestReadersAreToldItIsAStickFigure`
Gates: yes, in `go test ./...` (CI and image build)
Watched failing: 2026-10-15, across 5 mutations (---; the index's heading back to *The diagram*; the figure key back to *lines in the diagram*; every *stick figure* in `docs/using.md` replaced; and 1 more mutations).

## repo.noUnreachableCode

Criterion: [repo.noUnreachableCode](requirements/composed/internal/repo/acceptance.md#nounreachablecode)
Guardrails: `river/check/deadcode_test.go:TestNoUnreachableCode`
Gates: yes, in `go test ./...` (CI)
Watched failing: 2026-11-10, both against the codebase before cleanup (`gbif.Load`, `gbif.validTier`, `gbif.Client.CountInBox`, `gbif.Client.CountOffLandInBox`, `morphology.IsFamily`, `morphology.Review.Write`, `trait.GroupOf`, and `wingspan.publishedPredictions` in a non-test file) and by planting `unusedCascadeProbe` in `internal/trait/trait.go` (`1 unreachable function(s) in non-test code: birdview/internal/trait.unusedCascadeProbe`).
