# Open work

Everything outstanding in BirdView, in one place: unfinished work, open questions, known defects,
criteria not yet verified, and statements in the spec or docs that have gone stale. Requirements
stay in [product.md](product.md) and [tech.md](tech.md), and decisions in
[decisions.md](decisions.md); this file only says what is left to do and where its reasoning
lives.

**Adding an item.** Add a row to *Open* the moment the work is found, with the requirement, CR or
AC it comes from, so the reasoning stays in one place and this row only points at it. Do not
leave a TODO in code or a "left open" note in a decision record without a row here.

**Closing an item.** Strike the title through, say what closed it and when — a CR, an AC, a
commit — and move the row to *Closed*. Rows are not deleted: a closed row records why a question
that looks open is not.

## Open

| Item | Owner | Gate |
| --- | --- | --- |
| **26 links from served documents point at files the site does not serve.** Found 2026-10-15 when [AC-091](acceptance.md#ac-091)'s crawler learned to resolve relative links ([CR-045](decisions.md#cr-045)); they were always broken in a browser and the old crawler skipped them. Two kinds: links to source code and data tables (`../internal/region/region.go`, `../data/status.tsv`, `../third_party/avonet`), and links into the provenance records under `river/sources/`, which the server refuses along with the vendored terms beside them. Each should become a link to the published repository, or a plain code span; the published-repo option waits on the repository going public. Listed by name in `knownBrokenRepoLinks` so the list can only shrink | Sameer | before the site links to the published repository |
| **Judge presence over the whole selection, against a threshold the reader sets.** Chosen in principle 2026-10-09 (option B of the regions-scaling design) and parked before implementation. A species is present when its share of the selection's records — the tile's own *1 in N* — beats a threshold shown as a control: *show birds at least* 1 in 10,000 · 100,000 · 1,000,000 · 10,000,000 · 100,000,000 · any record. Proposed defaults, **undecided**: 1 in 1,000,000; the arrival page left unfiltered so scope stays P-001's job; and CR-036's month rule applied to the summed counts of the selected regions. Measured at 1 in 1,000,000: New York 491 → 443, the Northeast's nine states 564 → 470, all land regions 1,057 → 836. Two defects it would fix, both real today: [P-011](product.md#p-011)'s "in either source" judges iNaturalist against its own small total, so one observation passes — 45 of New York's species are present on that alone, Diamond Dove among them; and `birdview-regions` discarded counts under the bar at write time — **fixed on its own 2026-10-13, [CR-042](decisions.md#cr-042)**: every count is now kept and the bar decides membership only, which left 47 drawn species reading zero before it. Keeping every count costs 11% more cells. Works on today's 74 regions; would make counties a data change rather than a rules change | Sameer | parked |
| **Every committed table comes from the live GBIF API and cannot be re-derived.** Each build command issues thousands of throttled HTTPS queries to `api.gbif.org`, so two runs a week apart give different numbers and no past run can be reproduced — the figures pinned in `river/` and in tests are evidence of one afternoon, not of a stated input. That is tolerable while BirdView only *consumes* the data and becomes a defect the moment it **publishes its own dataset**, where a reader must be able to check a number against the source it came from. Two routes, both of which also end the 11-to-40-minute runs: GBIF's SQL download service, which aggregates server-side and returns a small file with a citable DOI but needs an account; or a GBIF Parquet snapshot queried locally, which needs no account — the better fit if others are to work with BirdView directly — at the cost of tens of GB and a query engine. Either way the switch is **once, for the whole pipeline**: two sources at different freshness is the disagreement [CR-031](decisions.md#cr-031) already fixed. Every pinned figure must be re-derived from the new input rather than adjusted to match | Sameer | before publishing the dataset |
| **AVONET's `Family1`–`3` and `Secondary1` are read and never used.** Found drawing [P-078](product.md#p-078)'s pipeline diagram, which marks both ✗. `avonet.FamilyOf` is documented as a cross-check against ABA's family, and ABA is gone ([CR-028](decisions.md#cr-028)); `Secondary1` is a required column whose value nothing reads. Either use them — `FamilyOf` could cross-check AviList's family the way it once checked ABA's — or stop reading them, and update the diagram either way | — | opportunistic |
| **Three more domestic forms behave like the two that are declared.** The rank fold of [AC-106](acceptance.md#ac-106) moved iNaturalist's trinomials into their wild species, and the proportional gainers were `Gallus gallus` **+11.9%**, `Anser cygnoides` **+20.0%** and `Anser anser` **+6.9%** — the same signature as the declared `Columba livia` (+1.3%) and `Cairina moschata` (+10.0%). `taxonomy.DomesticForms` names only the latter two, so [T-028](tech.md#t-028)'s ratio check reads the other three as undeclared disagreements if it ever flags them. Either declare them or record why the two are special; measuring it needs a `cmd/birdview-names` run | — | opportunistic |
| **The record counts behind a name disagreement are not on the page.** [P-077](decisions.md#cr-034) asks for each source's binomial *and how many records it holds under it*, and [AC-109](acceptance.md#ac-109) ships only the binomial. [`data/source-names.tsv`](../data/source-names.tsv) has per-source counts and no reader, but it queries the **AviList** name only — for Hairy Woodpecker it records eBird 0, not the 7,310,342 under *Dryobates villosus*, which is the number a reader wants. Closing it needs `cmd/birdview-names` to query the alias as well, a GBIF run, and the counts plumbed through the ingest | Sameer | opportunistic |
| **Nothing reads `data/gbif-taxa.tsv`.** It is generated by `cmd/birdview-gbifmap`, committed, and consumed by no code. It carries a GBIF-to-AviList name bridge with an `ambiguous` tier that would be the natural home for [CR-034](decisions.md#cr-034)'s shared names — except that it files `Setophaga petechia -> Setophaga petechia` as `exact`, correctly, which is why the bridge could not find them. Either wire it in or retire it | — | opportunistic |
| **The name sweep is a throwaway and should not be.** Finding [CR-034](decisions.md#cr-034)'s cases took 2,166 queries comparing what each database reports under each AviList binomial — the only method that catches a name whose strings match and whose meanings differ. It ran once from a scratch binary. As a committed command it would catch the next split the week it lands, rather than when a reader notices a mangrove warbler in Ontario. Note its blind spot: a `country=US,CA` filter hides the outlying islands, which is why *Sternula albifrons* looked like a finding and was not | Sameer | before the next AviList release |
| **Decide how a marine region's share is defined, or that it has none** — [CR-030](decisions.md#cr-030). The index inflates every marine share by 15–4,472× because a sea's denominator sums only the pelagic species, where a land region's counts every bird. Option A (presence at sea, no share) is recommended and costs almost nothing; option B costs a regenerated `region-presence.tsv` and `status.tsv` plus a coordinate-box form of the `eventDate` facet. **Until this lands the index is wrong for any marine selection** | Sameer | next — it is live and misleading |
| **Give the recurrence test a coordinate-box form, and find out whether the pelagics clear the bar on their own.** `DatesIn` filters by `gadmGid`; GBIF assigns none to open ocean, so no pelagic's at-sea records have ever been tested for recurrence. Four [P-071](product.md#p-071) exceptions exist in that blind spot, and [CR-028](decisions.md#cr-028)'s claim that recurrence is insensitive to observer effort has never been tested where observers are scarcest. The `eventDate` facet does accept a lat/lon box — probed 2026-10-05 — so the client method is small; the cost is a re-run of `birdview-status` against live GBIF. Note the boxes overlap the coast ([CR-030](decisions.md#cr-030)), so a naive run would count shore records as sea | Sameer | opportunistic; it could retire hand-declared exceptions |
| **Assert that `Commonest` is the largest share.** Nothing does. The server fixture hardcodes `commonestOf` and every test renders it back, so [P-074](product.md#p-074)'s central claim has never met data — which is why [CR-030](decisions.md#cr-030)(a) shipped. The check belongs over the real dataset, as [AC-098](acceptance.md#ac-098)'s id-collision half is | — | with CR-030 |
| **Region membership counts records, so a one-off vagrant can outscore a resident.** [P-011](product.md#p-011)'s 1-ppm rule is anti-correlated with what it measures: the more surprising a bird, the bigger the crowd, the more records. **Two examples are in the data as shipped** — American Flamingo in Pennsylvania (931 records, one pond, thirteen days, post-Idalia) and **Lucy's Warbler, a Sonoran desert bird, in Virginia** (53 records, one location, six days). **Three candidate remedies have been tried and all three fail; they are recorded so the next attempt does not repeat them.** *(1) Records in ≥2 of 3 four-year windows* — Lucy's Warbler walks through it, present 2016-12-31 to 2017-01-06, so one continuous event straddles a window boundary and scores 2/3. *(2) A threshold on distinct places* — cannot work at any value: Akohekohe, a Hawaiian endemic, has 4 places and Wood Stork in Nova Scotia, a vagrant, has 7. A critically endangered endemic with four records and a vagrant with four records are the same shape. *(3) Local abundance against continental abundance* — Flamingo/PA scores 1.61, above several birds that must be kept, because flamingos are rare everywhere so the baseline is tiny. A composite — drop if ≥20 records **and** (≤2 places **or** ≤1 year) — keeps every endemic and withheld species correctly and catches Lucy's and the flamingo, but still misses Wood Stork/NS, Spoonbill/ME, Fork-tailed Flycatcher/DE and Baltimore Oriole/NT. **Measurement caveat for whoever resumes**: these figures came from one GBIF page of ≤300 records, which is a *contiguous slice in GBIF's own ordering, not a random sample* — all 300 American Robin records for Pennsylvania fell in January 2024, and a page-derived year count read 1 where the true count is 60. Any real rule must take both signals from facets over all records, which raises the cost. Costs a regenerated `data/region-presence.tsv` and dataset. **Knock-on: the "Common birds" zones were clustered *from* that table ([CR-023](decisions.md#cr-023)), so they must be re-derived rather than left as a stale snapshot.** Documented for readers meanwhile under [Regions](../docs/data.md) | Sameer | deferred 2026-10-04; before anyone reads a region list as a range map |
| Group families by **what is seen together**: crawl eBird checklists and cluster families by co-occurrence, as a third grouping ([CR-025](decisions.md#cr-025)). This is the honest version of the habitat idea — co-occurrence is natively a *family*-level property, so nothing is averaged and nothing is misfiled. Two findings make it tractable, both measured 2026-10-03: GBIF's eBird export carries **no `eventID`**, but a checklist reconstructs from `recordedBy` + `eventDate` + rounded coordinates **only from a complete slice** — a 0.1% sample of a state-month gives 96% singleton checklists, while one county-day (1–3k records) gives real ones, median 19 species. A 602-checklist probe already separated waterbirds, woodland songbirds, open country, feeder birds and **town birds** (pigeons, House Sparrows, starlings), leaving hawks and crows unattached because they occur with everything. Expect the rare pelagic and Hawaiian families to be unplaceable, as the atolls were | Sameer | opportunistic; needs a stratified crawl of ~400–800 county-days into a committed table |
| **Publish the BirdView dataset as a source in its own right.** The project consumes four published datasets and credits each by DOI; it now derives work worth the same treatment — the occurrence status ([CR-028](decisions.md#cr-028)), the region presence and shares, the supplementary habitat and diet assignments, the neck and trunk models, and the wingspan regression. None of it is reachable except by scraping the app. **What the shape should be is the open question, not whether.** `web/data/species.json` is a serving artifact, not a publication: it nests provenance inside every metric, repeats source strings across 1,083 species, and carries presentation fields — `familyColor`, `familyTextColor`, `id` — that mean nothing outside this app. A publication probably wants CSV, one table per thing, keys that join: `species.csv` (the identity columns), `measurements.csv` (one row per species per metric, with its kind, source and model, so provenance is a column rather than a nest), `categories.csv` (habitat and diet with origin), `regions.csv` (species, region, records), `status.csv` (what `data/status.tsv` already is). Plus the thing none of our own sources gave us and all of them needed: a `PROVENANCE.md` stating the version, the license, what is measured against what is modeled, and the known distortions — eBird withholding sensitive species, crowd-driven over-reporting, the 4 species with no AVONET row. **Licensing needs deciding first**: the derived values are built from CC BY 4.0 inputs, so attribution flows through, and BirdView should pick its own license rather than leave it implied | Sameer | ready: the repo exists as `google/birdview` since 2026-10-08 |
| Ask Ricklefs or ESA for permission to use the measured-length data | Sameer | **after a working demo** |
| Assess whether a perched/standing toggle is achievable at any honest accuracy ([CR-011](decisions.md#cr-011)) | — | before it is offered to users |
| **Review `NeckRatios` against the full family list** — 24 of 90 families carry a ratio and **66 take the default 0.18, covering 548 species**. This is the half of the old entry that is actually tractable, and the one that catches the next short-necked Limpkin. ~~Derive the values from [`bohmer-neck`](sources.md#bohmer-neck) instead~~ — **withdrawn 2026-10-04**: [`sources.md`](sources.md#bohmer-neck) already records why, and the old entry ignored it. Böhmer measures the **full cervical column including the part inside the body contour**, where [P-045](product.md#p-045) models the *exposed* neck of a silhouette. Converting one to the other needs the fraction of the column that shows in flight, per family, unmeasured — which is exactly what `NeckRatios` already is. It would move the guess, not remove it, and add a dependency on a 103-species `.docx` that is consulted-not-ingested | Sameer | opportunistic; still the weakest model |
| Find a source that measures a **trunk**. CR-021's exponent is inferred from neck and leg, and that is its weakest link | — | opportunistic; would upgrade a calibration to a fit |
| **Parked:** replace the cube with a mass **line** proportional to mass. Measured and written up in [Mass as a proportional line](#mass-as-a-proportional-line-replacing-the-cube--parked); needs a CR, because it reverses CR-019 | Sameer | parked 2026-10-02 — resume from that section, do not re-derive |
| **Eight acceptance criteria have never been watched failing.** Still `planned` in [acceptance.md](acceptance.md): [AC-012](acceptance.md#ac-012) (all five AVONET traits on every record), [AC-014](acceptance.md#ac-014) (join coverage monitored), [AC-015](acceptance.md#ac-015) (unmatched species fail the build above a threshold), [AC-034](acceptance.md#ac-034) (proportions differentiate species), [AC-042](acceptance.md#ac-042) (the tile marks every non-measured metric), [AC-050](acceptance.md#ac-050) (nothing restricted is tracked — the document calls it its highest-value check), [AC-052](acceptance.md#ac-052) (outbound requests identify the client) and [AC-053](acceptance.md#ac-053) (the fetched ABA file matches its hash). AC-014 and AC-034 are already cited by tests, so they may only need the mutation run and a status; AC-053 is probably obsolete since [CR-028](decisions.md#cr-028) removed the ABA input, and wants withdrawing rather than verifying | — | opportunistic; AC-042 and AC-050 guard hard rules |
| **Four near-matches are splits awaiting a human taxonomic check** — [CR-006](decisions.md#cr-006), *Still unresolved*. *Astur atricapillus*, *Ardea ibis*, *Gygis candida* and *Tyto furcata* borrow from AVONET rows for the combined pre-split taxon. Checked 2026-10-13: AviList v2025b lists each parent and daughter as separate species, and no AVONET sheet has a row for any daughter, so `near-match` is the right label. Closing it means recording that in a decision, folding the four into the main `NearMatches` block, and deleting `taxonomy.AwaitingTaxonomicCheck` and the ingest line that prints it | Sameer | needs the owner's call |
| **AVONET's mass column was not split where BirdLife splits a pair.** [CR-006](decisions.md#cr-006)'s caveat, carried forward and never closed: *Larus smithsonianus*/*argentatus*, *Calonectris borealis*/*diomedea* and *Pterodroma deserta*/*feae* each share one mass, so their linear traits differ and sorting by mass shows them as identical. The same holds for every white tern (117 g) | — | opportunistic |
| **Deployment plumbing.** ~~No container image~~ (`Dockerfile` and [docs/building.md](../docs/building.md) added 2026-10-14, binding `0.0.0.0:8080` in the container); no host chosen yet | Sameer | before a public deployment |
| **Hummingbirds and gamebirds take the global wingspan fit**, not an order-level one, because neither order has five empirical species and Galliformes has none. Their estimates carry roughly 7.3% median error against 4.6% for well-covered orders. Moved from [arch.md](arch.md) | — | opportunistic; needs a wingspan source for those orders |
| **A resolved row in the GBIF bridge can be name-correct and concept-wrong.** Where GBIF has not followed a split, its name maps to whichever half kept it: GBIF's `Circus cyaneus` maps to the Eurasian Hen Harrier while 3.8 million of its records are Northern Harriers. Nine such cases are listed in [sources.md](sources.md#gbif); the table has one AviList column and cannot express them, and nothing detects them automatically. Moved from [arch.md](arch.md) | — | opportunistic |
| **`headAspect` is an unexamined drawing constant.** It produces no metric, so it carries no provenance, but it is as invented as the neck ratio. Moved from [arch.md](arch.md) | — | opportunistic |
| **The detail view re-renders the tile at 1.9×** rather than sharing a sized component, so a tile layout change has to be made in both places. Moved from [arch.md](arch.md) | — | opportunistic |
| **The ingest reads the 21 MB AVONET workbook three times**, once per taxonomy sheet — about four seconds, which has not been worth optimizing. Moved from [arch.md](arch.md) | — | only if build time starts to matter |
| **Photographs ([P-050](product.md#p-050)–[P-053](product.md#p-053)) are withdrawn from v1, not abandoned** — [CR-005](decisions.md#cr-005). Any revival goes through the spec first and stays within [P-064](product.md#p-064) and [T-002](tech.md#t-002): BirdView links to photographs and never embeds, hotlinks or caches one | — | not planned |
| **Other species may be drawn from an eBird row broader than their AviList concept.** [CR-041](decisions.md#cr-041) found European Herring Gull labeled `measured` from `AVONET2_eBird`'s 2022 *Larus argentatus*, which averages 16 specimens of which 6 are European. The ingest prefers the eBird sheet everywhere, and nothing looks for other species where AviList's concept is narrower and a BirdLife row matches it better. AVONET's raw data, which tags each specimen with its BirdLife species and eBird subspecies group, is enough to find them | — | opportunistic |
| **Other clones of `google/birdview` predate the history rewrite of 2026-10-13.** The stray `birdview-regions` binary and a root-level copy of AVONET's `Supplementary dataset 2.xlsx` were removed from every commit and `main` was force-pushed, so any other clone or fork must re-clone or reset to `origin/main`, and an old branch pushed back would restore both. GitHub keeps the old objects until its own garbage collection, or until Support purges them | Sameer | until every other clone is replaced |
| **Three birds were kept on the index as doubtful, not as wild.** [CR-043](decisions.md#cr-043) left West Indian Whistling-Duck (15 records: Alabama, Texas, Florida, Virginia), Cuban Grassquit (22, all Florida) and Cinnamon-rumped Seedeater (131, mostly California and Arizona) off `status.Exotic` because their records may be wild vagrants. Each needs a judgement: add it to the list with a reason, or leave it. The seedeater may also be a naming question, since AviList's *Sporophila torqueola* is the west Mexican bird and Morelet's Seedeater is the one native to Texas | Sameer | needs the owner's call |

## Closed

| Item | Owner | Gate |
| --- | --- | --- |
| ~~**CR-040's evidence was gathered in a way BirdLife's terms forbid.**~~ — **done 2026-10-14**: [CR-040](decisions.md#cr-040) re-based onto the two CC BY 4.0 workbooks already vendored (`AVONET_Raw_Data` shows all 6 specimens of `AVONET1_BirdLife`'s *Larus argentatus* are `argentatus/argenteus` while 4 of the 8 specimens of its *Larus smithsonianus* are from Russia and Japan; AviList v2025b's `Decision_summary` for Proposal `2025-1136` states that *L. vegae* is sometimes treated as a subspecies of *L. smithsonianus* and that BirdLife's *L. argentatus* is European Herring Gull alone). The DataZone quotations were removed and [`birdlife-checklist`](sources.md#birdlife-checklist) moved to `rejected` | Sameer | closed 2026-10-14 |
| ~~**GitHub reports 15 Dependabot alerts** on `google/birdview`'s default branch, 9 high and 6 moderate, as of the push of 2026-10-13. Not yet triaged~~ — **done 2026-10-14**: all vulnerabilities reported by `govulncheck` were in `golang.org/x/net` (bumped `v0.56.0` → `v0.60.0`) and standard-library `net/http` (`go.mod` bumped to `go 1.26.9`). `govulncheck ./...` now reports zero vulnerabilities | — | closed 2026-10-14 |
| ~~**Spec and docs still describe the ABA era, or older counts.**~~ — **done 2026-10-14**: [T-004](tech.md#t-004), [T-005](tech.md#t-005), [T-007](tech.md#t-007), [T-011](tech.md#t-011), [T-022](tech.md#t-022) and [T-024](tech.md#t-024) updated for post-[CR-028](decisions.md#cr-028); [P-002](product.md#p-002) qualified to match [P-060](product.md#p-060)'s three-sheet lookup; [AC-010](acceptance.md#ac-010), [AC-011](acceptance.md#ac-011), [AC-012](acceptance.md#ac-012), [AC-014](acceptance.md#ac-014) and the "no code" notes in [acceptance.md](acceptance.md) updated to the 1,082-species derived scope; `AGENTS.md`, `README.md`, `web/web.go`, [docs/building.md](../docs/building.md) ([CR-029](decisions.md#cr-029) bookmark note), [docs/data.md](../docs/data.md) (1,082 counts, withdrawn Böhmer neck derivation, dropped-species count, CR-028 status) and [docs/using.md](../docs/using.md) (rarity codes sentence) updated | — | closed 2026-10-14 |
| ~~**Species kept going by people are drawn as established, which [P-071](product.md#p-071) forbids.**~~ — **done 2026-10-14**, [CR-043](decisions.md#cr-043), [P-080](product.md#p-080), [AC-120](acceptance.md#ac-120). 41 escaped cage, aviary and farm birds, Emu among them, are declared in `status.Exotic`, left off the index and its controls, and listed at `/excluded`; their pages stay. Long-feral populations such as Indian Peafowl in Florida and Red Junglefowl on Kauai are not listed | Sameer | closed 2026-10-14 |
| ~~**Register the HBW-BirdLife checklist as consulted, not ingested.**~~ — **done 2026-10-13**: [`birdlife-checklist`](sources.md#birdlife-checklist). Checking its terms found the problem closed above | Sameer | closed 2026-10-13 |
| ~~**European Herring Gull is read from the combined 2022 row**~~ — **done 2026-10-13**, [CR-041](decisions.md#cr-041), [AC-118](acceptance.md#ac-118). It now reads `AVONET1_BirdLife`'s European-only *Larus argentatus*, labeled `measured` | Sameer | closed 2026-10-13 |
| ~~**American Herring Gull is labeled `measured` from a three-species row**~~ — **decided 2026-10-13: kept as it is**, [CR-041](decisions.md#cr-041). `AVONET1_BirdLife`'s *Larus smithsonianus* is the narrowest row AVONET has; 4 of its 8 specimens are tagged American Herring Gull and the other 4 carry no subspecies | Sameer | closed 2026-10-13 |
| ~~**Vega Gull's donor**~~ — **done 2026-10-13**, [CR-040](decisions.md#cr-040) (evidence re-based 2026-10-14 on `AVONET_Raw_Data` and `AviList`). `AVONET1_BirdLife`'s *Larus smithsonianus* holds the four Russian and Japanese specimens alongside four North American ones, so Vega Gull borrows that row, as American Herring Gull uses it, and both tiles say *Same measurements as* the other ([P-079](product.md#p-079), [AC-117](acceptance.md#ac-117)) | Sameer | closed 2026-10-13 |
| ~~**`birdview-regions` does not fold iNaturalist trinomials**~~ — **done 2026-10-06**, [AC-106](acceptance.md#ac-106). The rank reduction moved into `taxonomy.Binomial` and runs inside `taxonomy.Occurrence` *before* either naming table is consulted, so `dryobates villosus harrisi` reaches `Leuconotopicus villosus` instead of nothing. The last of the three asymmetries [CR-035](decisions.md#cr-035) found, and the criterion now asserts the property the three shared — that both readers land on the same AviList species — rather than the three bugs. The orphan residue was 0.0431% before the fold; what stays is hybrids, which must never match a species, and subspecies of birds BirdView does not draw | Sameer | closed 2026-10-06 |
| ~~**`birdview-regions` applies neither `AviListRenames` nor `CountedAs`**~~ — **done 2026-10-05**, [CR-035](decisions.md#cr-035). Both readers of the occurrence data now call `taxonomy.Occurrence`, and the join reports every name it could not match, refusing to write above 0.05% of records. The audit that justified it: 1,402 eBird names, 37 matching nothing, 9,954,362 recoverable records across ten species | Sameer | closed 2026-10-05 |
| ~~**Count eBird alone, and use iNaturalist only where eBird says it is hiding a bird.**~~ Planned 2026-10-09; one change in four parts, in this order. **Done: parts (1) and (2) 2026-10-09** ([CR-037](decisions.md#cr-037), [AC-113](acceptance.md#ac-113)) — the list restated and confirmed, 202 cells flagged, the month filter fixed; **parts (3) and (4) 2026-10-12** ([CR-038](decisions.md#cr-038), [AC-114](acceptance.md#ac-114)) — eBird counted, iNaturalist only in confirmed gaps, shared names carrying eBird's count, `WithheldByEBird` derived, scope re-run: 1,083 drawn species to 1,082, Mottled Owl the one to leave. **(1) A live bug comes first.** GBIF's eBird export applies eBird's published [Sensitive Species list](https://support.ebird.org/en/support/solutions/articles/48000803210-sensitive-species-in-ebird), including its *regional and seasonal* entries, so eBird is silent for some species in some regions and months — and [CR-036](decisions.md#cr-036)'s month filter reads that silence as absence. **Painted Bunting is hidden from Florida in January**, the heart of its Florida season (eBird: 0 records October–March, 12,483 in April), and **Short-eared Owl from New York in June** (0 in May and June, 3,709 in January). Suppressed region-months must abstain and say *withheld by eBird*, not count as absent. **(2) Encode the list** for the US and Canada as a committed table: species, region, date window. Of BirdView's species, 9 are on its global list — exactly the mainland and Hawaiian entries of `WithheldByEBird` — and **26 more on its regional list**, among them Long-eared, Short-eared, Burrowing, Boreal and Barn Owls, Golden Eagle, American Goshawk, King Rail, Loggerhead Shrike, Painted Bunting and Whooping Crane; only Black Rail was known. Verified against the export: Long-eared Owl has 15 eBird records in the US against 17,015 in Canada. The list's formatting is irregular (a scrape missed Barn Owl's regional entry), so the table needs building carefully, and eBird's terms checking before anything derived from it is committed. **(3) Change what counts.** eBird alone everywhere; iNaturalist only in cells eBird declares hidden — the 9 globally withheld species (100 presences) and the 57 region cells where eBird is silent all year (Long-eared Owl in about 30 states, Whooping Crane, Burrowing Owl in the prairie provinces, Spruce Grouse in New York, Black Rail) — judged on iNaturalist alone and labelled as such. This drops **985 of the 1,228 presences that rest on iNaturalist today**: 406 species, almost all escapes and domestic birds (Red Junglefowl in 42 regions, Japanese Quail in 27, Zebra Finch, Indian Peafowl, Budgerigar, Emu) and out-of-range strays. The figures barely move — median +0.6% — because iNaturalist is 1.24% of land records. It supersedes [P-011](product.md#p-011)'s "in either source", which judges iNaturalist against its own small total so one observation passes. **(4) Shared names and scope.** With iNaturalist out of the counts, Northern Yellow Warbler loses all 53,887 records (eBird files them under *Setophaga petechia*), so both halves of a shared name take eBird's shared count and regions, as [P-077](product.md#p-077) already describes. `WithheldByEBird` is derived from the list rather than kept by hand, which removes **Desertas Petrel**: it is not on eBird's list, so its missing records are probably a naming issue, the error Blue-billed White-Tern was. And `birdview-status` must be re-run to see which species leave scope without iNaturalist recurrence; the six mainland globally withheld species probably need declaring under [P-071](product.md#p-071), as the three Hawaiian ones already are | Sameer | done 2026-10-12 |
| ~~**Restore the share at sea on an off-land denominator**~~ — **done 2026-10-05**, [CR-031](decisions.md#cr-031). Sea totals are now every bird on open water in the box: North Atlantic 12,986,514 against the 429,535 shipped. `Commonest` moved onto the same table the index divides, so the two pages cannot disagree again, and **45 species peak at sea**. Remaining: the figure is *on water* rather than *at sea*, overstating by ~3% through estuaries, and a WKT polygon would fix that if it ever matters | Sameer | closed 2026-10-05 |
| ~~**Is there a distance from shore that tightens the marine boxes?**~~ — **answered 2026-10-05**, no: distance is the wrong instrument and there is no gradient to tune. Composition does not taper offshore, it flips between the band holding coastline and the band that does not — 4.9% off land and Northern Cardinal at one step, 100% off land and pure tubenoses at the next. GADM membership does the job exactly and for free. Written up under [CR-030](decisions.md#cr-030) | Sameer | closed 2026-10-05 |
| ~~**Reword the four pelagic exceptions in [P-071](product.md#p-071)**~~ — **done 2026-10-05** with [CR-030](decisions.md#cr-030). They said the derivation sees these birds rarely "because almost nobody is on the water where it is", which stated a measurement nobody made: the recurrence test runs on `gadmGid`, GBIF has no `gadmGid` for open ocean, and `DatesIn` has no coordinate-box form, so pass 2 has never asked the water anything. Each now says the records are at sea *where the test cannot yet look*, and names what the derivation does see on land. **The exceptions stand** — whether these species would clear the bar unaided remains unknown, and that is now what the table says | Sameer | closed 2026-10-05 |
| ~~Retrieve, hash and vendor `fox-wingspan`; write its `PROVENANCE.md`~~ — done: `spec/sources/fox-wingspan/PROVENANCE.md` exists, `spec/check/pins.py` verifies the hash, and the ingest reads it for 300 measured wingspans | — | closed 2026-10-04 |
| ~~Write `acceptance.md` — Gate 2, before any code~~ — done: 68 criteria, and Gate 2 passed long before the current work | — | closed 2026-10-04 |
| ~~**Decide what happens to a URL when the derived scope changes.**~~ — **done 2026-10-05**, [CR-029](decisions.md#cr-029) and [P-075](product.md#p-075). `/species/<id>` for a bird the status table classified but left out of scope now answers **404 with the bird named and its measured reason**, covering **328 of the 334 species** the derivation rejected — 124 vagrant, 167 absent, 37 historical; the other six have no AviList row, so no name and no stable id. CR-018's *name what was lost* carries over; its **200 does not**, because a comparison minus one bird is still a comparison and a detail page minus its bird is nothing. 410 was rejected for overclaiming: scope is regenerated, so a bird out today can be in after the next run. Cost a tenth column in `data/status.tsv` — the date last recorded — without which "historical" could be asserted and not shown | Sameer | closed 2026-10-05 |
| ~~ABA has not replied. Decide between chasing the permission and publishing a derived list~~ — decided: stop waiting, use public data. Names moved to AviList ([CR-022](decisions.md#cr-022)); **scope is what remains** | Sameer | decided 2026-10-02 |
| ~~Replace ABA **scope** with a list BirdView derives and owns~~ — **done 2026-10-05, and this was the deployment blocker.** [CR-028](decisions.md#cr-028): scope is recurrence, not abundance. `birdview-ingest` reads the committed [`data/status.tsv`](../data/status.tsv); `internal/aba` is deleted; nothing reads `fetched/`; the ABA credit is gone from all three footers because nothing on a page comes from it. **1,083 species in 99 families**, up from 839 — 830 of the old list kept on measured evidence, 9 kept by hand under [P-071](product.md#p-071), and the rest admitted because they recur. Every input the build reads is now CC BY 4.0 or the project's own work | Sameer | closed 2026-10-05 |
| ~~Ask ABA for written permission to redistribute the checklist~~ — **sent 2026-09-27**, awaiting reply | Sameer | done |
| ~~Decide the neck's flight posture~~ — done, [CR-011](decisions.md#cr-011) | Sameer | closed 2026-09-23 |
| ~~Investigate cervical vertebra counts as an anchor for the neck model~~ — answered twice, and both answers are no. As a **calibration** anchor: counts are dimensionless and the trunk's are conserved, so they cannot do it; [`bohmer-neck`](sources.md#bohmer-neck) measures neck *lengths* instead and now calibrates [P-047](product.md#p-047). As a **display** device — one dash per vertebra in the neck stroke — the neck is too short to carry them, measured 2026-10-04 over all 839 species with each bird filling its box: the median neck is **14.8px on the detail page and 7.8px on a tile**, which at a typical fourteen cervicals is **1.06px per vertebra**. Only American Flamingo (76px), Trumpeter Swan (68px) and Great Blue Heron (65px) reach even 4px per vertebra, and only on the detail page. Two further blocks if the geometry had worked: dash style already carries provenance ([P-065](product.md#p-065)), so dash *count* would overload the one mark that says "not a measurement"; and Böhmer covers 103 species against BirdView's 839, so ~88% of counts would be invented | — | closed 2026-10-04 |
| ~~Promote the cube/crossbar clearance invariant from prose into a test over the real dataset~~ — done 2026-10-04: `TestTileCubeNeverReachesTheCrossbar` asserts it over 839 species at both scales and all three metrics, and **logs the tightest margin** (33–41px) rather than asserting a figure read off a run. A boolean check was tried first and rejected as too weak: the cube is bottom-left and the crossbar upper-middle, so only a gross change breaks the conjunction | — | closed 2026-10-04 |
| ~~Add a **Sort by** control to the comparison view~~ — built, [AC-079](acceptance.md#ac-079) | Sameer | closed 2026-10-01 |
| ~~Decide whether mass gets an **equal-mass cube of water** drawn in the diagram~~ — built, [CR-019](decisions.md#cr-019) | Sameer | closed 2026-10-01 |
| ~~Give the mass cube a colour that is not the crossbar's~~ — resolved the other way round: the **crossbar** moved to green `#4ade80`, because the cube's blue depicts the water and should keep it. The recolour also fixed a live 7:1 violation, the labels sharing that colour being text | Sameer | closed 2026-10-02 |
| ~~Mass units switch at 16 oz, so one section can print `14.5 oz` beside `1.2 lb`~~ — done 2026-10-04, [P-073](product.md#p-073) and [AC-094](acceptance.md#ac-094): the mass unit is now a property of the comparison **scope** — a section or the whole set on the index, following the scale setting; the birds named in the address on the comparison page — and the **smallest** bird in the scope chooses it. A swan beside a sparrow reads `390.5 oz` and `0.93 oz`; two large birds keep pounds. Taking the largest instead was tried and rejected: it crushes a 27 g sparrow to `0.1 lb` | — | closed 2026-10-04 |
| ~~**`README.md` and `docs/` do not describe the current product**~~ — done 2026-10-04. `using.md` rewritten (it said Global was the default scale, the opposite of the page, and documented none of Regions, Diet and Habitat, compare, `/added` or the served spec); `building.md` had a mangled code fence rendering four paragraphs of prose as code on the served page, listed two of six commands and eight of fourteen packages, and said the ABA permission request was "drafted and unsent" when it was sent 2026-09-27; `data.md` claimed "Taxonomy is ABA's" two paragraphs after saying every name comes from AviList. README gained the GBIF row and the CR-028 status. [T-024](tech.md#t-024) keeps the reading gate: none of this was catchable by [AC-090](acceptance.md#ac-090), which compares counts and credits, not whether prose is true | Sameer | closed 2026-10-04 |
| ~~Delete the dead `.sw.*` swatch rules in `styles.css`~~ — done 2026-10-04: eight rules removed, no template has emitted `class="sw …"` since [CR-017](decisions.md#cr-017) | — | closed 2026-10-04 |

## Background notes

Design notes the rows above refer to, moved here from decisions.md with the list on 2026-10-13.

### A cube of water for mass

*Evaluated 2026-10-01, owner's idea, not a decision.* [P-066](product.md#p-066) rejects drawing
mass as a shape and gives two reasons. The owner then proposed a specific shape that answers one
of them completely, so the rejection is narrower than it reads and the measurements are recorded
here rather than left to be redone.

**Draw the cube of water that weighs what the bird weighs**, side = ∛(mass). This is not a model
of anything: one gram of water is one cubic centimetre, by the historical definition of the gram,
so the side is a unit conversion and carries no assumption about avian density. It is a real
length, so it can be drawn at the diagram's own scale and does not introduce the second,
arbitrary scale P-043 refuses. P-066's objection on that count does not apply to it.

**Measured on the real data, at the detail view's geometry** (drawing area 668 × 281):

| | side |
| --- | --- |
| smallest (Fork-tailed Flycatcher, 32 g) | 28 px — 10% of the drawing height |
| median | 52 px — 18% |
| largest (Thick-billed Murre, 964 g) | 80 px — 29% |

It fits. Placed in the lower-left corner with an 8px margin, the crossbar intrudes on it for
**0 of 841 species**; the tightest clearance is 50 px, on Western Grebe.

**What it would add that nothing else shows.** The extremes are not arbitrary: the largest cubes
belong to alcids — Thick-billed Murre, Parakeet Auklet, Crested Auklet, Marbled Murrelet, all
dense diving birds — and the smallest to long-tailed, light-bodied species, Fork-tailed and
Scissor-tailed Flycatcher and Black-billed Magpie. The cube read against the bird's own length is
**bulk against elongation**, which is exactly the thing a length-and-wingspan diagram cannot say.

**What it still does not do**, and the reason P-066's second objection stands: it compresses. A
Bushtit and a Common Raven differ 175× in mass and their cubes differ 5.6×. That is arithmetically
correct for equal-volume cubes and visually understates the gap. It matters less now that the
number is on the figure — the label can carry the magnitude and the cube the tangibility — but a
reader who takes the cubes as the mass comparison will read them wrong.

Open if it is taken up: what labels it, whether it appears on index tiles (it should not — the
tile is dense by design and this is detail-view furniture, as P-066's labels already are), and
whether it is drawn solid or dashed for the 10 near-match masses, which P-065's vocabulary already
answers.

### Mass as a proportional line, replacing the cube — parked

*Evaluated 2026-10-02, owner's idea, parked rather than decided.* Everything measured here is
still true; pick it up from this section rather than re-deriving it.

**The complaint is correct and is already in the spec.** The cube root flattens: a Bushtit and a
Common Raven differ in mass by 175× and their cubes by 5.6×. P-066 admits this and answers it with
the printed number. The proposal was to answer it with the drawing instead — a horizontal line
below the bird, from the lower-left, ticks at both ends, and the Mass label at the far end as the
Wingspan label already sits at the end of the crossbar.

**Then the owner looked at a rendered section and reported the opposite: the cubes draw the eye to
the heavier birds, and work.** That is evidence, and it is worth more than the argument above, so
it is recorded first. Measured on the section in that screenshot — #9, 51 species:

| | |
| --- | --- |
| length spread | **1.15×** (272–312 mm), by construction |
| mass spread | **8.3×** — Mangrove Cuckoo 63.9 g to Pigeon Guillemot 530 g |
| cube edge spread | **2.02×** |
| apparent area spread | **4.10×** |

A cube twice as wide as its neighbour, in a row of birds that are all the same length, is not
subtle. **The cube root is mild compression over the range a reader actually compares across.**

That reframes the Bushtit-and-Raven example this section opens with. Those two differ 175× in
mass — and they are 11 sections apart, never drawn side by side at a shared scale. The complaint
is strongest exactly where the comparison is not being made, and weakest where it is. Per-section
scale is the default (P-042), so the common case is the favourable one.

**It also rescues a comparison the numbers break.** In that same section the Pigeon Guillemot's
mass prints as `1.2 lb` and the Eared Grebe's as `14.5 oz`, because the formatter switches unit at
16 oz. Two birds of the same length, one third heavier, in units a reader has to convert before
comparing. The cubes are directly comparable where the labels are not — an argument for the shape
that nobody made when CR-019 adopted it.

**Where that leaves the line.** The case for it is weaker than when this section was written: it
would buy proportionality over a range where the cube is already legible, and pay the price set
out below. Not closed — the measurements stand and the design is worked out — but anyone picking
it up should start by disagreeing with the screenshot.

**The finding that decides the shape of the work: a line linear in mass cannot be drawn at one
global scale.** The mass range is **4,258×** (2.6 g – 11,071 g). A 380 px tile has a usable
dynamic range of roughly **180×**. With the Trumpeter Swan's line at 300 px the Calliope
Hummingbird's is **0.07 px**.

Nor is there a physically honest way out. The cube's whole elegance is that one gram of water is
one cubic centimetre *by definition*, so its edge is a true length drawn at the figure's own
scale and invents no second scale. The line equivalent is a water column of fixed cross-section,
where length in centimetres equals mass in grams — and the swan's column is **92× its own body
length**. No choice of cross-section helps, because the range to be spanned exceeds the range a
tile can draw by 24×.

**So a proportional mass line must be normalized per group, and that is the trade.** Measured
within the 20 sections at the default 15% step:

| within one length section | |
| --- | --- |
| median mass range | **4.0×** |
| worst section | **20.0×** — 57 species, 31.9–637 g |
| sections above 50× | **0 of 20** |

Comfortably drawable. Note what is and is not given up: **the bird is already drawn at a
per-section scale** (P-042), and so is the cube, so a per-section mass scale is not a new kind of
relativity. What the cube has that a normalized bar would not is that its size *means* something
absolute — you could measure it against the bird with one ruler. A normalized bar is a bar-chart
row: very readable, making no physical claim.

**It therefore needs a CR, because it reverses rather than extends CR-019.** P-066 currently
states that "a disc or bar sized by mass — a second, arbitrary scale — is still refused", and
P-043's rationale is that a mark on its own convenient scale "looks authoritative and says
something false". Those are the sentences a CR would have to overturn, not work around.

**The hard case is the comparison view**, where a reader may pick a hummingbird against a swan and
the lighter line goes sub-pixel. That is arguably the correct message — the cube is precisely what
hides it — but invisible reads as broken. The mitigation is to draw the origin tick
unconditionally, so there is always a mark, and to let the number carry the magnitude as P-066
already says it must.

**Styling, as far as it was worked out:**

- **Solid.** Mass is measured for 831 of 841, and P-065 draws `near-match` solid too.
- **Not a heavier weight.** Stroke weight carries no meaning in this diagram today, and the
  heaviest mark on a bird diagram should not be the one that is not the bird.
- **A different colour, and this is the load-bearing one.** A horizontal solid ticked line will
  otherwise read as a second crossbar — another length to scale, which is the exact false claim
  P-043 exists to prevent. Contrast measured against the `#0f172a` background, and against the
  wing's `#0ea5e9`, because a hue-only difference fails for colour-blind readers:

  | candidate | vs background | vs wing |
  | --- | ---: | ---: |
  | lime `#a3e635` | 11.8:1 | **1.8:1** |
  | emerald `#34d399` | 9.3:1 | 1.4:1 |
  | fuchsia `#e879f9` | 7.3:1 | 1.1:1 |
  | violet `#a78bfa` | 6.6:1 | 1.0:1 |

  Lime has the best luminance separation from the crossbar; the others differ by hue alone.
- **Ticks at both ends** — consistent with the chord-tick vocabulary of P-065.
- **Label at the far end** — consistent with the wingspan label in P-066.

**Two things found along the way that are worth acting on whether or not this is ever built:**

1. **The cube is `#38bdf8` and the crossbar `#0ea5e9`.** The one mark that is *not* a length is
   near-identical in colour to the mark whose entire point is that it *is* a length to scale.
   That is backwards, and it is a defect in what ships today.
2. **There is no clear band below the bird on a tile.** The tallest bird in each section fills the
   tile height exactly. The cube works because it occupies the empty lower-left *beside* a narrow
   figure; a full-width horizontal line would cross the tail unless the tile grows downward, as
   the detail view's bottom strip already does for P-066's labels.

### Cervical vertebra counts, and why they are worth chasing

*Recorded 2026-09-23.* The per-family neck ratios of P-045 are the largest unforced assumption in
the project: invented, unvalidated, and visibly driving the silhouette of every long-necked family.
Three families have now been corrected by eye — Aramidae and Pelecanidae after a reader noticed a
short-necked Limpkin, Sulidae on review — which is a sign the table is guesswork with a slow error
signal, not that it is converging.

**Cervical vertebra count is the most promising anchor available.** Birds run from roughly 11 to 25
cervicals, against a mammalian fixed seven, and the count tracks neck length closely: swans sit at
the top of the range, passerines near the bottom, galliforms between. Unlike neck length it is a
discrete, published, comparable anatomical trait.

It would not give the ratio directly — ours is neck ÷ trunk, so it needs count × mean vertebra
length ÷ trunk — but it would give a defensible **ordering across all 91 families**. That is enough
to calibrate the table against the few anchors the project is most confident about, instead of
inventing each entry independently and waiting for someone to spot the next Limpkin.

Two things are unknown and should not be assumed: whether a dataset exists at BirdView's coverage,
and whether its license permits use. Both have sunk candidate sources before. Adoption would follow
the usual path — a `sources.md` entry and a compose-and-resolve pass.

### Sketch: one dash per vertebra

*Recorded 2026-09-23, owner's idea, not a decision.* If the counts existed, the neck could be drawn
as a **dashed line with one dash per cervical vertebra** — putting a measured, discrete count into
the one segment that currently carries the least support.

What makes it more than decoration: two birds can reach the same neck length by different means. A
swan is near the top of the vertebral range; a flamingo has a comparably long neck built from
fewer, longer vertebrae. That distinction is real, and the diagram cannot presently show it at all.

Three obstacles, recorded so the idea is picked up with its problems attached rather than
rediscovered:

- **The tail is already dashed**, to distinguish it from the body. Dash would then mean two
  different things on one figure. The tail would need another treatment; there are options.
- **It cannot work on the index tiles.** The neck is a fixed fraction of a bird that fills its tile,
  so its drawn length is roughly scale-independent: about 31 px for the longest-necked species in
  the set, and under 2 px for a median bird. Twenty dashes in 31 px is a grey blur.
- **On the detail view it is marginal**, at 1.9× the tile: a heron's neck is about 62 px, giving
  roughly 3 px per vertebra.

That last figure is the interesting one. Nobody will count 24 dashes, but 3 px is enough to read as
**texture** — finely divided against coarsely divided — which is exactly the swan-versus-flamingo
distinction the idea exists to surface. Comparability rather than countability may be the right
goal, and would be the thing to test.

### Why the perched toggle is not simply the other half of CR-011

*Recorded 2026-09-23.* Offering "show them perched" sounds like a rendering switch and is not one.
A perched bird needs two things BirdView does not have.

**Legs.** [CR-010](decisions.md#cr-010) removed them because a flight silhouette does not show them; a perched
silhouette is mostly leg. Restoring them restores every problem CR-010 recorded — the femur, the
toes and the hip position are unmeasured, and the tarsus alone understates the visible leg most in
the birds whose legs define them. A perched view would need the exposed-leg model that CR-009
invented and CR-010 deleted, and it would still have no source to validate against.

**A second neck model.** The current ratios are a standing-posture model used, until CR-011, in a
flight drawing. A perched view needs the perched extension for every family, which is a different
invented table again — a perched heron's neck is neither folded flat nor fully extended, and where
it sits between those is not published at this coverage.

So the toggle is not a view of the same data from another angle. It is a second set of invented
per-family parameters with no empirical check, which is the thing this project has twice decided
not to buy. Worth revisiting only if a source appears that measures posture, not just skins.

### Why the permission asks wait for a demo

*Recorded 2026-09-21.* Both asks are deferred until BirdView runs and can be shown.

A request that describes an intended use is weaker than one that demonstrates it. Showing a
rights-holder the actual application — what it does with their data, how the data is credited,
that it is non-commercial and that every value carries visible provenance — answers most of what
they would otherwise have to take on trust. It also lets them see that BirdView labels a modeled
length as modeled, which is the strongest available evidence that their measured data would be
handled honestly.

Neither ask blocks v1, which is what makes waiting free:

- **Ricklefs / ESA** would *add* capability. Every length is currently `modeled`
  ([CR-002](decisions.md#cr-002)); a grant would make roughly half of them measured. Nothing breaks without
  it.
- **ABA** would *simplify the build*. Fetch-at-build already works under T-004, so a grant would
  only remove the network step and let T-005 relax. Nothing breaks without it either.

Sending them together after a demo also means one conversation per organization rather than two
rounds of the same explanation.

**Update, 2026-09-27.** The ABA request was sent, with screenshots rather than a live demo. The
obvious plan — deploy somewhere they can reach and link it — was circular: hosting BirdView
publicly is the very thing the email asks permission for. Screenshots reproduce almost nothing, to
the rights-holder, about their own data, and need no server to stay up. The Ricklefs/ESA ask is
still outstanding.

**The cost of waiting**, recorded so it is a choice rather than a drift: if ESA grants permission
late, the length column changes from wholly `modeled` to half `measured`, which will move diagrams
and re-sort sections. Builders should treat the provenance model (P-062) as the thing that stays
fixed and the mix of kinds as the thing that moves.
