package repo

# repo — composed requirements for repository structure, build, and release

**Composed, not authored.** Change it by changing a source and recomposing, never by editing files in
this directory by hand. Composed from `tech` and `decisions` under
[precedence.md](../../../precedence/precedence.md); its
`composed.sum`, written by `river compose`, records what it was composed from.

Everything BirdView's repository layout, build inputs and artifacts, container image, continuous
integration, deployment posture, and development process must satisfy. Its acceptance criteria are
in [`acceptance.md`](acceptance.md), and the guardrails and verification observations covering them
are in [conformance.md](../../../../conformance/conformance.md).

## Composed from

- tech (Local)
- decisions
- precedence

## Composition record

| Source | Tier | Pin |
| --- | --- | --- |
| [`tech`](../../../local/tech/tech.md) | local | the commit this file is in |
| [`decisions`](../../../decisions/decisions.md) | — | the commit this file is in |

Precedence is the ranking in [precedence.md](../../../precedence/precedence.md). **No conflict arose.**

## Applies unchanged

### Module, build inputs, and artifacts

- [tech.twoCommandsOneModule](../../../local/tech/delivery.md#twocommandsonemodule) — Two commands, one module
- [tech.goVersionFollowsDeps](../../../local/tech/delivery.md#goversionfollowsdeps) — The Go version tracks whatever dependencies require
- [tech.fetchRestrictedInputs](../../../local/tech/delivery.md#fetchrestrictedinputs) — Build inputs that are not vendored are fetched into `fetched/`, hash-checked, and never committed
- [tech.datasetNeverCommitted](../../../local/tech/delivery.md#datasetnevercommitted) — The built dataset is an artifact, never committed
- [tech.containerImageBuilds](../../../local/tech/delivery.md#containerimagebuilds) — The container image builds from a clean checkout

### Process and posture

- [tech.localNowPublicLater](../../../local/tech/posture.md#localnowpubliclater) — Public as a personal site; never commercial
- [tech.bundleIsNormative](../../../local/tech/process.md#bundleisnormative) — The spec bundle is normative
- [tech.recordBeforeConsuming](../../../local/tech/process.md#recordbeforeconsuming) — No source is consumed before it is recorded
- [tech.activeWorkHasIssue](../../../local/tech/process.md#activeworkhasissue) — Work in progress is marked active in the todo list and tracked by a GitHub issue

### Continuous integration and release gates

- [tech.checksOnEveryChange](../../../local/tech/release.md#checksoneverychange) — Every push to main and every pull request runs the standing checks
- [tech.testedBeforeDeploy](../../../local/tech/release.md#testedbeforedeploy) — A deploy builds only code whose tests pass
- [tech.imageRecordsRevision](../../../local/tech/release.md#imagerecordsrevision) — The deployed server knows the commit it was built from
- [tech.smokeCheckAfterDeploy](../../../local/tech/release.md#smokecheckafterdeploy) — Every deploy is checked from outside

## Excluded

### Excluded by [decisions.repoScope](../../../decisions/decisions.md#reposcope)

- [gbifterms.AccurateRegistration](../../../imported/gbif.org/gbifterms/gbifterms.md#accurateregistration)
- [gbifterms.ApiWithinSpecifications](../../../imported/gbif.org/gbifterms/gbifterms.md#apiwithinspecifications)
- [gbifterms.CiteApiDataByDerivedDataset](../../../imported/gbif.org/gbifterms/gbifterms.md#citeapidatabyderiveddataset)
- [gbifterms.CiteDownloadDoi](../../../imported/gbif.org/gbifterms/gbifterms.md#citedownloaddoi)
- [gbifterms.DataNotGuaranteed](../../../imported/gbif.org/gbifterms/gbifterms.md#datanotguaranteed)
- [gbifterms.PublisherLicencesPrevail](../../../imported/gbif.org/gbifterms/gbifterms.md#publisherlicencesprevail)
- [gbifterms.RetainOwnershipIdentifier](../../../imported/gbif.org/gbifterms/gbifterms.md#retainownershipidentifier)
- [inatapi.AuthenticateOnlyWhenNecessary](../../../imported/inaturalist.org/inatapi/inatapi.md#authenticateonlywhennecessary)
- [inatapi.IdAboveBeyondPageCap](../../../imported/inaturalist.org/inatapi/inatapi.md#idabovebeyondpagecap)
- [inatapi.JwtAuthorization](../../../imported/inaturalist.org/inatapi/inatapi.md#jwtauthorization)
- [inatapi.LargestPageSize](../../../imported/inaturalist.org/inatapi/inatapi.md#largestpagesize)
- [inatapi.MediaDownloadCaps](../../../imported/inaturalist.org/inatapi/inatapi.md#mediadownloadcaps)
- [inatapi.NoGroupAccounts](../../../imported/inaturalist.org/inatapi/inatapi.md#nogroupaccounts)
- [inatapi.RequestRate](../../../imported/inaturalist.org/inatapi/inatapi.md#requestrate)
- [inatapi.UserAgent](../../../imported/inaturalist.org/inatapi/inatapi.md#useragent)
- [inatterms.AccountHolderResponsible](../../../imported/inaturalist.org/inatterms/inatterms.md#accountholderresponsible)
- [inatterms.ContributorKeepsRights](../../../imported/inaturalist.org/inatterms/inatterms.md#contributorkeepsrights)
- [inatterms.DoNotBehaveLikeMachine](../../../imported/inaturalist.org/inatterms/inatterms.md#donotbehavelikemachine)
- [inatterms.DuplicatesUndesirable](../../../imported/inaturalist.org/inatterms/inatterms.md#duplicatesundesirable)
- [inatterms.NoMachineGeneratedContent](../../../imported/inaturalist.org/inatterms/inatterms.md#nomachinegeneratedcontent)
- [inatterms.RateDoesNotHinderOthers](../../../imported/inaturalist.org/inatterms/inatterms.md#ratedoesnothinderothers)
- [product.ascendingOrder](../../../local/product/ordering.md#ascendingorder)
- [product.bandingCodes](../../../local/product/dataset.md#bandingcodes)
- [product.birdPresets](../../../local/product/presentation.md#birdpresets)
- [product.borrowedMeasurementsSaySo](../../../local/product/measurements.md#borrowedmeasurementssayso)
- [product.categoryProvenance](../../../local/product/measurements.md#categoryprovenance)
- [product.chooseOneSettings](../../../local/product/presentation.md#chooseonesettings)
- [product.craniumFromMass](../../../local/product/diagram.md#craniumfrommass)
- [product.crossbarAtVerticalScale](../../../local/product/diagram.md#crossbaratverticalscale)
- [product.declaredExceptions](../../../local/product/dataset.md#declaredexceptions)
- [product.diagramOnlyMode](../../../local/product/presentation.md#diagramonlymode)
- [product.ebirdLink](../../../local/product/presentation.md#ebirdlink)
- [product.englishNameChoice](../../../local/product/diagram.md#englishnamechoice)
- [product.escapedBirdsListed](../../../local/product/measurements.md#escapedbirdslisted)
- [product.everydayDefinitions](../../../local/product/measurements.md#everydaydefinitions)
- [product.familiesByOrder](../../../local/product/months.md#familiesbyorder)
- [product.familyFilter](../../../local/product/filtering.md#familyfilter)
- [product.filtersCompose](../../../local/product/months.md#filterscompose)
- [product.fitsAPhone](../../../local/product/presentation.md#fitsaphone)
- [product.generatedBadgeColors](../../../local/product/dataset.md#generatedbadgecolors)
- [product.habitatAndDietFilters](../../../local/product/months.md#habitatanddietfilters)
- [product.hereSelectsLocalRegion](../../../local/product/filtering.md#hereselectslocalregion)
- [product.imperialByDefault](../../../local/product/presentation.md#imperialbydefault)
- [product.joinOnBinomial](../../../local/product/dataset.md#joinonbinomial)
- [product.joinsReportDrops](../../../local/product/dataset.md#joinsreportdrops)
- [product.labeledDetailDiagram](../../../local/product/measurements.md#labeleddetaildiagram)
- [product.lengthAndWingspanSources](../../../local/product/measurements.md#lengthandwingspansources)
- [product.locationStaysInBrowser](../../../local/product/filtering.md#locationstaysinbrowser)
- [product.measuredVersusModeled](../../../local/product/measurements.md#measuredversusmodeled)
- [product.metricProvenanceExplained](../../../local/product/measurements.md#metricprovenanceexplained)
- [product.namesFromAvilist](../../../local/product/dataset.md#namesfromavilist)
- [product.namesRecordsFiledUnder](../../../local/product/presentation.md#namesrecordsfiledunder)
- [product.neckModelPerFamily](../../../local/product/diagram.md#neckmodelperfamily)
- [product.nextView](../../../local/product/presentation.md#nextview)
- [product.noPersonalAddress](../../../local/product/presentation.md#nopersonaladdress)
- [product.oneDiagramVocabulary](../../../local/product/diagram.md#onediagramvocabulary)
- [product.oneSelectedMetric](../../../local/product/ordering.md#oneselectedmetric)
- [product.oneUnitInComparison](../../../local/product/presentation.md#oneunitincomparison)
- [product.orderByFrequency](../../../local/product/presentation.md#orderbyfrequency)
- [product.pipelineDiagram](../../../local/product/measurements.md#pipelinediagram)
- [product.proportionDifferentiates](../../../local/product/diagram.md#proportiondifferentiates)
- [product.provenancePerMetric](../../../local/product/measurements.md#provenancepermetric)
- [product.regionAndMonthFilter](../../../local/product/filtering.md#regionandmonthfilter)
- [product.reportDataIssue](../../../local/product/presentation.md#reportdataissue)
- [product.reportingFrequencyFilter](../../../local/product/filtering.md#reportingfrequencyfilter)
- [product.scopeIsRecurringAvifauna](../../../local/product/dataset.md#scopeisrecurringavifauna)
- [product.seaDistanceDerived](../../../local/product/dataset.md#seadistancederived)
- [product.sectionBreakRule](../../../local/product/ordering.md#sectionbreakrule)
- [product.sectionHeadersStateRange](../../../local/product/ordering.md#sectionheadersstaterange)
- [product.sectionThresholdDefault](../../../local/product/ordering.md#sectionthresholddefault)
- [product.segmentComposition](../../../local/product/diagram.md#segmentcomposition)
- [product.segmentsSumToLength](../../../local/product/diagram.md#segmentssumtolength)
- [product.selfDescriptionStaysTrue](../../../local/product/presentation.md#selfdescriptionstaystrue)
- [product.shareOfRecords](../../../local/product/presentation.md#shareofrecords)
- [product.sourceNamingMeasured](../../../local/product/dataset.md#sourcenamingmeasured)
- [product.statusIsPublished](../../../local/product/dataset.md#statusispublished)
- [product.tierHistogram](../../../local/product/ordering.md#tierhistogram)
- [product.tileContents](../../../local/product/presentation.md#tilecontents)
- [product.tileMarksModeled](../../../local/product/measurements.md#tilemarksmodeled)
- [product.tileOpensDetail](../../../local/product/presentation.md#tileopensdetail)
- [product.traitsFromAvonet](../../../local/product/dataset.md#traitsfromavonet)
- [product.traitsPerSpecies](../../../local/product/dataset.md#traitsperspecies)
- [product.trunkFromMass](../../../local/product/diagram.md#trunkfrommass)
- [product.undrawnSpeciesAnswer](../../../local/product/presentation.md#undrawnspeciesanswer)
- [product.unmatchedSpeciesReported](../../../local/product/dataset.md#unmatchedspeciesreported)
- [tech.badgeContrast](../../../local/tech/accessibility.md#badgecontrast)
- [tech.cacheLifetimeStated](../../../local/tech/delivery.md#cachelifetimestated)
- [tech.compressedResponses](../../../local/tech/delivery.md#compressedresponses)
- [tech.crawlersSkipFilteredViews](../../../local/tech/delivery.md#crawlersskipfilteredviews)
- [tech.filteredViewsBounded](../../../local/tech/delivery.md#filteredviewsbounded)
- [tech.gbifDownloadsRecorded](../../../local/tech/process.md#gbifdownloadsrecorded)
- [tech.identifiedClient](../../../local/tech/delivery.md#identifiedclient)
- [tech.oneCanonicalHost](../../../local/tech/delivery.md#onecanonicalhost)
- [tech.regionBoundaries](../../../local/tech/delivery.md#regionboundaries)
- [tech.seaDistanceTable](../../../local/tech/delivery.md#seadistancetable)
- [tech.selfContainedServer](../../../local/tech/delivery.md#selfcontainedserver)
- [tech.servesOnLocalhost](../../../local/tech/delivery.md#servesonlocalhost)
- [tech.servesOnlyItsPages](../../../local/tech/delivery.md#servesonlyitspages)
- [tech.stableUrls](../../../local/tech/delivery.md#stableurls)
- [tech.textContrastFloor](../../../local/tech/accessibility.md#textcontrastfloor)
- [tech.textSizeFloor](../../../local/tech/accessibility.md#textsizefloor)
- [tech.webPages](../../../local/tech/accessibility.md#webpages)
- [wcag.ContrastEnhanced](../../../imported/w3.org/wcag/wcag.md#contrastenhanced)
- [wcag.ContrastMinimum](../../../imported/w3.org/wcag/wcag.md#contrastminimum)
- [wcag.ContrastRatio](../../../imported/w3.org/wcag/wcag.md#contrastratio)
- [wcag.FullPages](../../../imported/w3.org/wcag/wcag.md#fullpages)
- [wcag.NonTextContrast](../../../imported/w3.org/wcag/wcag.md#nontextcontrast)
- [wcag.WebPage](../../../imported/w3.org/wcag/wcag.md#webpage)
- [wiley.MiningNeedsAgreement](../../../imported/onlinelibrary.wiley.com/wiley/wiley.md#miningneedsagreement)
- [wiley.NoAiProcessing](../../../imported/onlinelibrary.wiley.com/wiley/wiley.md#noaiprocessing)
- [wiley.NoAutomatedAccess](../../../imported/onlinelibrary.wiley.com/wiley/wiley.md#noautomatedaccess)
- [wiley.NoIntegrationOrDerivatives](../../../imported/onlinelibrary.wiley.com/wiley/wiley.md#nointegrationorderivatives)
- [wiley.NoRepositoryIncorporation](../../../imported/onlinelibrary.wiley.com/wiley/wiley.md#norepositoryincorporation)
- [wiley.NoRetrievalStorage](../../../imported/onlinelibrary.wiley.com/wiley/wiley.md#noretrievalstorage)
- [wiley.WhatIsPermitted](../../../imported/onlinelibrary.wiley.com/wiley/wiley.md#whatispermitted)
