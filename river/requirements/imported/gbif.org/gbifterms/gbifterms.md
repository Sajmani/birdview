package gbifterms

# gbif-terms — transcribed requirements

Tier: **mandatory**. Scope: BirdView's use of data obtained through GBIF: today, eBird and
iNaturalist occurrence counts read from GBIF's occurrence search API, and GADM and taxon lookups;
proposed, SQL occurrence downloads ([todo.countyRegions](../../../../todo.md#countyregions)).

GBIF's terms govern access and attribution; **the licenses of the datasets prevail over them**
([PublisherLicencesPrevail](#publisherlicencesprevail)), so the two datasets' own licenses are
part of what this file has to answer. See [PROVENANCE.md](PROVENANCE.md).

---

## CiteDownloadDoi

**Cite every download BirdView uses by its DOI**

> "Users must publicly acknowledge, following the scientific convention of citing sources in
> conjunction with the use of the data, the Data Publishers whose biodiversity data they have
> used, where appropriate through use of a Digital Object Identifier (DOI) applying to the
> dataset(s) and/or data downloads."
> — Data user agreement, Provisions, 3

> "Under the terms of the GBIF data user agreement, users who download individual datasets or
> search results and use them in research or policy agree to cite them using a DOI, or Digital
> Object Identifier."
> — Citation guidelines

> "However, please do include each element of content—most importantly the **DOI** expressed as a
> URL."
> — Citation guidelines

**For BirdView: applies as soon as it uses a download.** Each SQL download is assigned a DOI;
BirdView must cite it as a URL, `https://doi.org/10.15468/dl.…`, where it credits its sources:
the page footer and the data guide. Several downloads are each cited, or listed together:

> "If you have used multiple downloads, you may not be able to include all citations in the
> reference list of your article. In this case, we recommend including a supplementary list or
> addendum of all downloads used."
> — Citation guidelines, Citing multiple downloads

## CiteApiDataByDerivedDataset

**Data read from the search API, which has no DOI, is cited by a derived dataset**

> "Tools returning results directly from the GBIF search API (e.g. spocc, dismo and the
> *occ_data()* and *occ_search()* functions of rgbif) will **not** assign single DOIs for data
> downloaded. It is up to the user to identify dataset publishers and properly acknowledge each of
> them when citing the data."
> — Citation guidelines, GBIF data accessed using third-party tools

> "For data obtained via occurrence search API-based tools, we recommend using a derived dataset
> as an easy way of obtaining a DOI for citing the data."
> — Citation guidelines, GBIF data accessed using third-party tools

**For BirdView: applies today, and is not yet met.** BirdView's region presence, recurrence, and
counts come from the occurrence search API, so no DOI covers them. The requirement is to
acknowledge each publisher (the eBird Observation Dataset and iNaturalist research-grade
observations, by their dataset DOIs); a derived dataset, which needs a GBIF account and a list of
contributing datasets with counts, is GBIF's recommendation for a single DOI. Moving to downloads
replaces this with [CiteDownloadDoi](#citedownloaddoi) for whatever the downloads cover.

## PublisherLicencesPrevail

**Each dataset's license governs, and prevails over GBIF's terms**

> "Users must comply with the terms and conditions included in the licence selected by each Data
> Publisher, and the licensing information included with each data download. If any provision of
> this Use Agreement conflicts with the terms and conditions within the licences selected by the
> Data Publisher, licences selected by the Data Publisher shall prevail."
> — Data user agreement, Provisions, 4

**For BirdView: the licenses differ, and one is non-commercial.** The eBird Observation Dataset is
CC BY 4.0. iNaturalist's research-grade observations are **CC BY-NC 4.0** (both from the datasets'
GBIF records, `api.gbif.org/v1/dataset/{key}`, read 2026-11-03). BirdView already counts
iNaturalist records where eBird withholds a species ([decisions.countEbirdFirst](../../../decisions.md#countebirdfirst)), so the
non-commercial term reaches it now, not only with downloads. **Whether BirdView's use is
non-commercial in the license's sense is open** and is the owner's to decide in `decisions.md`
([todo.gbifLicenseTerms](../../../../todo.md#gbiflicenseterms)).

## RetainOwnershipIdentifier

**A record shared onward keeps the identifier of its owner**

> "In order to make attribution of use for owners of the data possible, the identifier of
> ownership of data must be retained with every data record shared onward for reuse."
> — Data user agreement, Provisions, 2

**Not applicable as written.** BirdView shares no occurrence record onward: no coordinate, date,
observer, or record reaches its dataset ([imports.gbif](../../../imports.md#gbif)), only per-region lists and counts
derived from them. If it ever publishes records, each keeps its `datasetKey`, which the citation
guidelines also ask be preserved through filtering.

## ApiWithinSpecifications

**Use the APIs as specified, and never so as to disrupt GBIF**

> "Data use through the available Application Programming Interfaces (API) shall comply with the
> technical specifications provided in the website of the GBIF Secretariat. In particular, it is
> forbidden to use APIs to create denial of service, disruption in the service, and in general to
> interfere with the normal operations of GBIF and the Data Publishers."
> — Data user agreement, Provisions, 7

**For BirdView: met by pacing, and better met by downloads.** Every BirdView client paces its
requests and identifies itself ([tech.identifiedClient](../../../local/tech/delivery.md#identifiedclient)). County-scale regions would multiply
the search requests about fortyfold, which is the reason to move to downloads.

## AccurateRegistration

**The account behind BirdView's downloads is registered accurately**

> "Users must provide accurate and up to date personal data at the time of registration. Failure
> to do so implies automatic termination of this Agreement."
> — Data user agreement, Provisions, 6

**For BirdView: applies to the maintainer's GBIF account**, which requests the downloads. Its
credentials never enter the repository.

## DataNotGuaranteed

**GBIF guarantees neither the data nor the service**

> "GBIF Secretariat cannot guarantee the quality or completeness of data, nor does it guarantee
> uninterrupted data access services. Users employ these data and services at their own risk."
> — Data user agreement, Provisions, 1

**For BirdView: informative.** It is why BirdView measures what it derives rather than trusting it
([decisions.ebirdSilenceNotAbsence](../../../decisions.md#ebirdsilencenotabsence)), and why its pipeline
retries and records retrieval dates.

## Not adopted

- **Provision 5**, benefit-sharing for digital sequence information: BirdView uses none.
- **Provisions 8 to 10**, the website's cookies, GBIF's privacy policy, and Danish law: they govern
  GBIF.org and the agreement itself, not anything BirdView does.
- **Citing individual occurrences, species pages, media, and non-data content**: BirdView cites
  none of these. It shows no media from occurrence records.
