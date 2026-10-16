package inatapi

# inat-api — transcribed requirements

Tier: **governing**. Scope: every request BirdView makes to the iNaturalist API.

BirdView's use is narrow. `cmd/birdview-inatmap` makes **unauthenticated, read-only** `GET`
requests to `api.inaturalist.org/v1/taxa`, once per species, to resolve a scientific name to a
taxon ID. It runs when the mapping is rebuilt — not during `birdview-ingest`, and never while
serving. It posts nothing, authenticates as nobody, and downloads no media.

IDs match the transcription in the maintainer's `birdsync` repository, which vendored the same
page, so the two projects can be read side by side. The *readings* below are BirdView's own.
Each entry quotes the source so the reading can be checked. See [PROVENANCE.md](PROVENANCE.md).

---

## RequestRate

**About one request per second, about 10,000 a day**

> "Please keep requests to about 1 per second, and around 10k API requests a day"
>
> "Requests exceeding this limit might be throttled, and will return an HTTP 429 exception
> “Too Many Requests”... Please add delays into your code to keep under these limits...
> We may block IPs that consistently exceed these limits"

**Implemented as [tech.pacedInaturalistMapping](../tech/process.md#pacedinaturalistmapping).** The mapping tool waits a minimum interval between
requests. A full rebuild is 841 requests — one per in-scope species — which is under a tenth of
the daily figure, and the tool is run rarely rather than on every build.

## LargestPageSize

**Use the largest supported page size**

> "If using the API to fetch a lot of results, please use the highest supported per_page
> value. For example you can get up to 200 observations in a single request, which would be
> faster and more efficient than fetching the default 30 results at a time"

**Not applicable.** BirdView resolves one name at a time and reads only the matching taxon. There
is no bulk endpoint for name resolution to page through; `per_page` is left at its default and the
response is discarded after the ID is taken.

## IdAboveBeyondPageCap

**Page numbers cap at 10,000 results; use `id_above` beyond that**

> "The page and per_page parameters can be used to fetch up to (for many endpoints) 10k
> results. An error will be thrown if results beyond 10k are requested."

**Not applicable.** BirdView never pages.

## UserAgent

**Identify the client with a User-Agent**

> "please consider using a custom User Agent to identify your application... this could be
> set to the name of your app, an iNaturalist username, or anything we might use to
> differentiate your requests"

**Implemented as [tech.identifiedClient](../tech/delivery.md#identifiedclient)**, which BirdView already required of every outbound
request before this source was adopted, and which [acceptance.outboundRequestsIdentified](../../acceptance.md#outboundrequestsidentified) already checks.

## JwtAuthorization

**JWTs, expiring in 24 hours, in an Authorization header**

> "The preferred way to make an authenticated request in the newer API is to use a JSON Web
> Token (JWT), available at https://www.inaturalist.org/users/api_token"

**Not applicable.** BirdView holds no credential and sends no `Authorization` header.

## NoGroupAccounts

**Users log in as themselves; no group accounts**

> "If you are creating an application that will allow users to post data to iNaturalist,
> please do not do it using a group account (i.e. many individuals submitting data under a
> single username)"

**Not applicable.** BirdView allows no one to post anything.

## AuthenticateOnlyWhenNecessary

**Authenticate only when necessary**

> "Authenticate API requests only when required. Since the data returned from authenticated
> requests might contain private information, these requests are not cached, therefore they
> can require more server time to process"

**Satisfied by construction.** BirdView never authenticates, so every request it makes is
cacheable by iNaturalist. Worth recording rather than skipping: this is the one requirement that
`birdsync` records as a knowing departure, and BirdView gets it for free only because its needs
are so much smaller.

## MediaDownloadCaps

**One IP; media download volume caps**

> "Please use a single IP address for fetching data. If we think multiple IPs are being used
> in coordination to bypass rate limits, we may block those IPs regardless of query rate"
>
> "Downloading over 5 GB of media per hour or 24 GB of media per day may result in a
> permanent block"

**Satisfied by construction.** The mapping is rebuilt from one machine, and BirdView downloads no
media at all — it links to iNaturalist rather than reproducing anything from it.

---

## Not adopted

Bulk access routes — observation exports, the weekly GBIF dataset — serve consumers of
iNaturalist *observation* data. BirdView consumes no observations. It reads one integer per
species from the taxonomy and then stops.
