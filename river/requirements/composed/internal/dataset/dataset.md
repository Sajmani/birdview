package dataset

# dataset — composed requirements for the BirdView data pipeline and dataset

**Composed, not authored.** Change it by changing a source and recomposing, never by editing files in
this directory by hand. Composed from `gbifterms`, `wiley`, `inatapi`, `inatterms`, `product`,
`tech`, and `decisions` under [precedence.md](../../../precedence.md); approvals are recorded in
[approvals.md](../../../approvals.md).

Everything the BirdView data pipeline and built dataset (`cmd/birdview-ingest`,
`cmd/birdview-status`, `cmd/birdview-regions`, `cmd/birdview-sea`, `cmd/birdview-names`,
`cmd/birdview-gbifmap`, `cmd/birdview-gbifdownload`, `data/`, and the ingest packages under
`internal/`) must satisfy. Its acceptance criteria are in [`acceptance.md`](acceptance.md), and the
guardrails and verification observations covering them are in
[conformance.md](../../../../conformance.md).

## Composed from

- gbifterms (Mandatory)
- wiley (Mandatory)
- inatapi (Governing)
- inatterms (Mandatory)
- product (Local)
- tech (Local)
- decisions
- precedence

## Composition record

| Source | Tier | Pin |
| --- | --- | --- |
| [`gbifterms`](../../../imported/gbif.org/gbifterms/gbifterms.md) | mandatory | GBIF Data User Agreement (15 September 2026) and Citation Guidelines; SHA-256 in [PROVENANCE.md](../../../imported/gbif.org/gbifterms/PROVENANCE.md) |
| [`wiley`](../../../imported/onlinelibrary.wiley.com/wiley/wiley.md) | mandatory | Wiley Online Library Terms and Conditions, retrieved 2026-09-21; SHA-256 in [PROVENANCE.md](../../../imported/onlinelibrary.wiley.com/wiley/PROVENANCE.md) |
| [`inatterms`](../../../imported/inaturalist.org/inatterms/inatterms.md) | mandatory (retired) | Retrieved 2026-08-09; SHA-256 in [PROVENANCE.md](../../../imported/inaturalist.org/inatterms/PROVENANCE.md) |
| [`inatapi`](../../../imported/inaturalist.org/inatapi/inatapi.md) | governing (retired) | Revised 2025-02-27; SHA-256 in [PROVENANCE.md](../../../imported/inaturalist.org/inatapi/PROVENANCE.md) |
| [`product`](../../../local/product/product.md) | local | the commit this file is in |
| [`tech`](../../../local/tech/tech.md) | local | the commit this file is in |
| [`decisions`](../../../decisions.md) | — | the commit this file is in |

Precedence is the ranking in [precedence.md](../../../precedence.md): mandatory sources outrank
governing and local ones. Where Wiley's Rules of Use prohibited automated retrieval and repository
incorporation, Ricklefs 2017 was rejected ([decisions.ricklefsProhibited](../../../decisions.md#ricklefsprohibited)). Where direct iNaturalist API and
link use was retired ([decisions.inaturalistSourcesRetired](../../../decisions.md#inaturalistsourcesretired)), `inatapi` and `inatterms` are excluded.

## Applies unchanged

### Scope, names, and joins

- [product.scopeIsRecurringAvifauna](../../../local/product/dataset.md#scopeisrecurringavifauna) — Scope is the recurring avifauna of the regions BirdView covers
- [product.statusIsPublished](../../../local/product/dataset.md#statusispublished) — Status is published, not just used
- [product.declaredExceptions](../../../local/product/dataset.md#declaredexceptions) — Where the data cannot speak, the exception is declared and named
- [product.namesFromAvilist](../../../local/product/dataset.md#namesfromavilist) — Names come from AviList
- [product.bandingCodes](../../../local/product/dataset.md#bandingcodes) — Each bird carries its four-letter banding code, where the Bird Banding Laboratory gives one
- [product.joinOnBinomial](../../../local/product/dataset.md#joinonbinomial) — Sources join on normalized scientific binomial
- [product.unmatchedSpeciesReported](../../../local/product/dataset.md#unmatchedspeciesreported) — Unmatched species are reported, not silently dropped
- [product.joinsReportDrops](../../../local/product/dataset.md#joinsreportdrops) — Every join reports what it dropped
- [product.sourceNamingMeasured](../../../local/product/dataset.md#sourcenamingmeasured) — How each source names a species is measured, not assumed

### Measurements, morphometric models, and palette

- [product.traitsFromAvonet](../../../local/product/dataset.md#traitsfromavonet) — Morphometric traits come from AVONET
- [product.traitsPerSpecies](../../../local/product/dataset.md#traitsperspecies) — Traits carried per species
- [product.lengthAndWingspanSources](../../../local/product/measurements.md#lengthandwingspansources) — Where length and wingspan come from
- [product.provenancePerMetric](../../../local/product/measurements.md#provenancepermetric) — Every presented metric carries a provenance record
- [product.craniumFromMass](../../../local/product/diagram.md#craniumfrommass) — The cranium is modeled from mass
- [product.trunkFromMass](../../../local/product/diagram.md#trunkfrommass) — The trunk is modeled allometrically from mass
- [product.neckModelPerFamily](../../../local/product/diagram.md#neckmodelperfamily) — The neck/body split is a named per-family model
- [product.generatedBadgeColors](../../../local/product/dataset.md#generatedbadgecolors) — Family badge colors are generated

### Regions, sea distance, and categories

- [product.regionAndMonthFilter](../../../local/product/filtering.md#regionandmonthfilter) — Filter by region and month
- [product.habitatAndDietFilters](../../../local/product/months.md#habitatanddietfilters) — Filter by habitat and by diet
- [product.seaDistanceDerived](../../../local/product/dataset.md#seadistancederived) — Each bird of the sea carries where its records lie, and its open-ocean refinement follows from it
- [tech.seaDistanceTable](../../../local/tech/delivery.md#seadistancetable) — Where each species' records lie relative to shore is derived once and committed

### Pipeline etiquette and provenance

- [tech.identifiedClient](../../../local/tech/delivery.md#identifiedclient) — Identify the client on every outbound request
- [tech.gbifDownloadsRecorded](../../../local/tech/process.md#gbifdownloadsrecorded) — Every GBIF download is requested from a committed query and recorded before it is read

## Excluded requirements

- [tech.pacedInaturalistMapping](../../../local/tech/process.md#pacedinaturalistmapping) *(excluded)* — The iNaturalist mapping is built by a separate, paced tool. Excluded by [decisions.inaturalistSourcesRetired](../../../decisions.md#inaturalistsourcesretired)
- [tech.proportionsReviewTool](../../../local/tech/process.md#proportionsreviewtool) *(excluded)* — A local tool for reviewing proportions against photographs. Excluded by [decisions.photoReviewUnreliable](../../../decisions.md#photoreviewunreliable)
- [gbifterms.RetainOwnershipIdentifier](../../../imported/gbif.org/gbifterms/gbifterms.md#retainownershipidentifier) *(excluded)* — A record shared onward keeps the identifier of its owner. Excluded by [decisions.riverV2PackagesAndSplits](../../../decisions.md#riverv2packagesandsplits)
- [inatapi.RequestRate](../../../imported/inaturalist.org/inatapi/inatapi.md#requestrate) *(excluded)* — About one request per second, about 10,000 a day. Excluded by [decisions.inaturalistSourcesRetired](../../../decisions.md#inaturalistsourcesretired)
- [inatapi.LargestPageSize](../../../imported/inaturalist.org/inatapi/inatapi.md#largestpagesize) *(excluded)* — Use the largest supported page size. Excluded by [decisions.inaturalistSourcesRetired](../../../decisions.md#inaturalistsourcesretired)
- [inatapi.IdAboveBeyondPageCap](../../../imported/inaturalist.org/inatapi/inatapi.md#idabovebeyondpagecap) *(excluded)* — Page numbers cap at 10,000 results; use `id_above` beyond that. Excluded by [decisions.inaturalistSourcesRetired](../../../decisions.md#inaturalistsourcesretired)
- [inatapi.UserAgent](../../../imported/inaturalist.org/inatapi/inatapi.md#useragent) *(excluded)* — Identify the client with a User-Agent. Excluded by [decisions.inaturalistSourcesRetired](../../../decisions.md#inaturalistsourcesretired)
- [inatapi.JwtAuthorization](../../../imported/inaturalist.org/inatapi/inatapi.md#jwtauthorization) *(excluded)* — JWTs, expiring in 24 hours, in an Authorization header. Excluded by [decisions.inaturalistSourcesRetired](../../../decisions.md#inaturalistsourcesretired)
- [inatapi.NoGroupAccounts](../../../imported/inaturalist.org/inatapi/inatapi.md#nogroupaccounts) *(excluded)* — Users log in as themselves; no group accounts. Excluded by [decisions.inaturalistSourcesRetired](../../../decisions.md#inaturalistsourcesretired)
- [inatapi.AuthenticateOnlyWhenNecessary](../../../imported/inaturalist.org/inatapi/inatapi.md#authenticateonlywhennecessary) *(excluded)* — Authenticate only when necessary. Excluded by [decisions.inaturalistSourcesRetired](../../../decisions.md#inaturalistsourcesretired)
- [inatapi.MediaDownloadCaps](../../../imported/inaturalist.org/inatapi/inatapi.md#mediadownloadcaps) *(excluded)* — One IP; media download volume caps. Excluded by [decisions.inaturalistSourcesRetired](../../../decisions.md#inaturalistsourcesretired)
- [inatterms.ContributorKeepsRights](../../../imported/inaturalist.org/inatterms/inatterms.md#contributorkeepsrights) *(excluded)* — A contributor keeps their rights, and licenses by default as CC BY-NC. Excluded by [decisions.inaturalistSourcesRetired](../../../decisions.md#inaturalistsourcesretired)
- [inatterms.AccountHolderResponsible](../../../imported/inaturalist.org/inatterms/inatterms.md#accountholderresponsible) *(excluded)* — The account holder is responsible for everything done under the account. Excluded by [decisions.inaturalistSourcesRetired](../../../decisions.md#inaturalistsourcesretired)
- [inatterms.RateDoesNotHinderOthers](../../../imported/inaturalist.org/inatterms/inatterms.md#ratedoesnothinderothers) *(excluded)* — Post at a rate and volume that doesn't hinder other users. Excluded by [decisions.inaturalistSourcesRetired](../../../decisions.md#inaturalistsourcesretired)
- [inatterms.NoMachineGeneratedContent](../../../imported/inaturalist.org/inatterms/inatterms.md#nomachinegeneratedcontent) *(excluded)* — Machine generated content is prohibited. Excluded by [decisions.inaturalistSourcesRetired](../../../decisions.md#inaturalistsourcesretired)
- [inatterms.DoNotBehaveLikeMachine](../../../imported/inaturalist.org/inatterms/inatterms.md#donotbehavelikemachine) *(excluded)* — Don't behave like a machine. Excluded by [decisions.inaturalistSourcesRetired](../../../decisions.md#inaturalistsourcesretired)
- [inatterms.DuplicatesUndesirable](../../../imported/inaturalist.org/inatterms/inatterms.md#duplicatesundesirable) *(excluded)* — Duplicates are undesirable but not an offense. Excluded by [decisions.inaturalistSourcesRetired](../../../decisions.md#inaturalistsourcesretired)

---

## Composed requirements

## gbifOccurrenceAndApiTerms

**GBIF API and download tools pace requests, record DOIs, and count CC BY-NC iNaturalist observations only under BirdView's non-commercial posture**

Composed from: [gbifterms.CiteDownloadDoi](../../../imported/gbif.org/gbifterms/gbifterms.md#citedownloaddoi), [gbifterms.CiteApiDataByDerivedDataset](../../../imported/gbif.org/gbifterms/gbifterms.md#citeapidatabyderiveddataset), [gbifterms.PublisherLicencesPrevail](../../../imported/gbif.org/gbifterms/gbifterms.md#publisherlicencesprevail), [gbifterms.ApiWithinSpecifications](../../../imported/gbif.org/gbifterms/gbifterms.md#apiwithinspecifications),
  [gbifterms.AccurateRegistration](../../../imported/gbif.org/gbifterms/gbifterms.md#accurateregistration), [gbifterms.DataNotGuaranteed](../../../imported/gbif.org/gbifterms/gbifterms.md#datanotguaranteed), [tech.localNowPublicLater](../../../local/tech/posture.md#localnowpubliclater), [decisions.gbifTermsMet](../../../decisions.md#gbiftermsmet), [decisions.ebirdSilenceNotAbsence](../../../decisions.md#ebirdsilencenotabsence)
Goals: [goals.goodDataCitizen](../../../../goals/goals.md#gooddatacitizen), [goals.honestData](../../../../goals/goals.md#honestdata)

Every BirdView command that queries GBIF identifies itself and paces its requests so as never to
disrupt GBIF ([gbifterms.ApiWithinSpecifications](../../../imported/gbif.org/gbifterms/gbifterms.md#apiwithinspecifications), [tech.identifiedClient](../../../local/tech/delivery.md#identifiedclient)). SQL downloads use an
accurately registered maintainer account whose credentials stay out of the repository
([gbifterms.AccurateRegistration](../../../imported/gbif.org/gbifterms/gbifterms.md#accurateregistration)), and record every download's key and DOI in
`data/gbif-downloads.tsv` before any code reads it ([gbifterms.CiteDownloadDoi](../../../imported/gbif.org/gbifterms/gbifterms.md#citedownloaddoi),
[tech.gbifDownloadsRecorded](../../../local/tech/process.md#gbifdownloadsrecorded)); tables built from the search API name the underlying dataset DOIs in
their headers ([gbifterms.CiteApiDataByDerivedDataset](../../../imported/gbif.org/gbifterms/gbifterms.md#citeapidatabyderiveddataset)). Because dataset publishers' licenses
prevail ([gbifterms.PublisherLicencesPrevail](../../../imported/gbif.org/gbifterms/gbifterms.md#publisherlicencesprevail)) and iNaturalist's research-grade dataset is CC BY-NC
4.0, BirdView counts iNaturalist records only while its deployment remains personal and
non-commercial ([tech.localNowPublicLater](../../../local/tech/posture.md#localnowpubliclater), [decisions.gbifTermsMet](../../../decisions.md#gbiftermsmet)). Because GBIF guarantees
neither data completeness nor uninterrupted service ([gbifterms.DataNotGuaranteed](../../../imported/gbif.org/gbifterms/gbifterms.md#datanotguaranteed)), the pipeline
measures suppression and naming disagreements rather than assuming silence is absence
([decisions.ebirdSilenceNotAbsence](../../../decisions.md#ebirdsilencenotabsence)).

## noWileyMaterialIngestedOrAutomated

**No Wiley Online Library material is fetched, stored in the repository, processed by AI, or integrated into the dataset**

Composed from: [wiley.NoRetrievalStorage](../../../imported/onlinelibrary.wiley.com/wiley/wiley.md#noretrievalstorage), [wiley.NoIntegrationOrDerivatives](../../../imported/onlinelibrary.wiley.com/wiley/wiley.md#nointegrationorderivatives), [wiley.NoAutomatedAccess](../../../imported/onlinelibrary.wiley.com/wiley/wiley.md#noautomatedaccess),
  [wiley.NoRepositoryIncorporation](../../../imported/onlinelibrary.wiley.com/wiley/wiley.md#norepositoryincorporation), [wiley.NoAiProcessing](../../../imported/onlinelibrary.wiley.com/wiley/wiley.md#noaiprocessing), [wiley.WhatIsPermitted](../../../imported/onlinelibrary.wiley.com/wiley/wiley.md#whatispermitted), [wiley.MiningNeedsAgreement](../../../imported/onlinelibrary.wiley.com/wiley/wiley.md#miningneedsagreement),
  [decisions.ricklefsProhibited](../../../decisions.md#ricklefsprohibited)
Goals: [goals.goodDataCitizen](../../../../goals/goals.md#gooddatacitizen)

Without a separate written agreement ([wiley.MiningNeedsAgreement](../../../imported/onlinelibrary.wiley.com/wiley/wiley.md#miningneedsagreement)), BirdView neither fetches nor
scrapes any material from Wiley Online Library by program or agent ([wiley.NoAutomatedAccess](../../../imported/onlinelibrary.wiley.com/wiley/wiley.md#noautomatedaccess)),
passes none to AI tools ([wiley.NoAiProcessing](../../../imported/onlinelibrary.wiley.com/wiley/wiley.md#noaiprocessing)), stores none in git or `third_party/`
([wiley.NoRetrievalStorage](../../../imported/onlinelibrary.wiley.com/wiley/wiley.md#noretrievalstorage), [wiley.NoRepositoryIncorporation](../../../imported/onlinelibrary.wiley.com/wiley/wiley.md#norepositoryincorporation)), and integrates none — including
Ricklefs 2017 — into the dataset ([wiley.NoIntegrationOrDerivatives](../../../imported/onlinelibrary.wiley.com/wiley/wiley.md#nointegrationorderivatives),
[decisions.ricklefsProhibited](../../../decisions.md#ricklefsprohibited)). Only the brief scholarly quotations of Wiley's own Rules of Use in
`wiley.md` are kept, under [wiley.WhatIsPermitted](../../../imported/onlinelibrary.wiley.com/wiley/wiley.md#whatispermitted).
