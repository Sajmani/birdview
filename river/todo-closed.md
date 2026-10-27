package todo

# Closed work

Items moved here from [todo.md](todo.md) when they closed. Search `## <id>`.

## approveAppComposition

**Approve the first composition, `app`, and decide whether to tag it**

Status: closed
Owner: Sameer
Closed: 2026-10-25, approved by Sameer

[requirements/composed/internal/app/](requirements/composed/internal/app/app.md) composes accessibility only: WCAG 1.4.6 and 1.4.11 over
every web page as [tech.webPages](requirements/local/tech/accessibility.md#webpages) defines one,
and the stylesheet's contract. An agent composed it; River requires a person to sign off. The
rulings it embeds: plain-text error pages are web pages (WCAG's definition leaves no choice), large
text gets no allowance ([decisions.oneContrastFloor](requirements/decisions.md#onecontrastfloor)), and closed panels are measured open

**done 2026-10-25**: approved as composed, rulings included. Not tagged: the commit that records the
approval identifies the published version.

## contrastCheckNeedsChrome

**The rendered contrast check gates only where Chrome is installed**

Status: closed
Owner: Sameer
Closed: 2026-10-23, by [tech.checksOnEveryChange](requirements/local/tech/release.md#checksoneverychange)

**done 2026-10-23**: CI (`.github/workflows/checks.yml`) runs on every push and pull request with
Chrome installed and fails on any skipped test. Its first run, on `7fcea0b`, ran
[app.renderedTextContrast](requirements/composed/internal/app/acceptance.md#renderedtextcontrast) over 15 pages and passed, with no test skipped

## stableSurface

**State what BirdView promises to keep stable, so the rest is free to change**

Status: closed
Owner: Sameer
Closed: 2026-10-23, by [tech.stableUrls](requirements/local/tech/delivery.md#stableurls) and [tech.oneCanonicalHost](requirements/local/tech/delivery.md#onecanonicalhost)

**done 2026-10-23**: [tech.stableUrls](requirements/local/tech/delivery.md#stableurls) names the
public interface (the index, species pages, the comparison, `/now`, the guides, the added and
excluded lists, and the eleven query parameters the pages write) and declares everything else
internal, `species.json` included. [tech.oneCanonicalHost](requirements/local/tech/delivery.md#onecanonicalhost)
gives a deployment one address. Its expected tension with publishing the dataset stays with
[publishDataset](todo.md#publishdataset): a publication is a separate, documented schema

## dockerBuildMissesData

**The container image does not build: its server stage omits `data/`**

Status: closed
Owner: Sameer
Closed: 2026-10-22, by `COPY data ./data` in the `builder` stage and [tech.containerImageBuilds](requirements/local/tech/delivery.md#containerimagebuilds)

Found and fixed 2026-10-22. `cmd/birdview` has imported `birdview/data`, which embeds
`region-boundaries.json`, since the Here button (`db8d1e7`, 2026-10-18), and the `builder` stage
did not copy it, so `docker build` failed. [acceptance.dockerStagesCopyWhatTheyBuild](requirements/acceptance.md#dockerstagescopywhattheybuild) now checks
every stage's `COPY` list against what its command imports. Docker is not installed on the
maintainer's machine, so the fix was confirmed by building both stages from exactly the directories
they copy, running the ingest as the image does (1,082 species), and serving from the binary and
dataset alone. Nothing has yet run `docker build` itself

## marineShareDefinition

**Decide how a marine region's share is defined, or that it has none**

Status: closed
Owner: Sameer
Closed: 2026-10-05, by [decisions.honestSeaDenominator](requirements/decisions.md#honestseadenominator); moved here 2026-10-22

**done 2026-10-05**, option B: a sea's total is every bird record on open water inside its box,
taken by faceting the box by `gadmLevel0Gid` and subtracting the land (`gbif.CountOffLandInBox`), so
it is the same quantity a land region's total is. All five seas carry `total` rows in
`data/region-presence.tsv`, and `Commonest` is computed in the ingest from the same numbers the
index divides ([acceptance.peakAndShareOneCalculation](requirements/acceptance.md#peakandshareonecalculation)).
The item stayed open by oversight after the decision landed. What the fix does not cover, records on
estuaries counting as sea, is in that decision

## vegaGullEvidence

**[decisions.vegaGullBorrows](requirements/decisions.md#vegagullborrows)'s evidence was gathered in a way BirdLife's terms forbid**

Status: closed
Owner: Sameer
Closed: 2026-10-14

**done 2026-10-14**: [decisions.vegaGullBorrows](requirements/decisions.md#vegagullborrows) re-based
onto the two CC BY 4.0 workbooks already vendored (`AVONET_Raw_Data` shows all 6 specimens of
`AVONET1_BirdLife`'s *Larus argentatus* are `argentatus/argenteus` while 4 of the 8 specimens of its
*Larus smithsonianus* are from Russia and Japan; AviList v2025b's `Decision_summary` for Proposal
`2025-1136` states that *L. vegae* is sometimes treated as a subspecies of *L. smithsonianus* and
that BirdLife's *L. argentatus* is European Herring Gull alone). The DataZone quotations were
removed and [`birdlife-checklist`](requirements/imports.md#birdlife-checklist) moved to
`rejected`

## dependabotAlerts

**GitHub reports 15 Dependabot alerts on `google/birdview`'s default branch, 9 high and 6 moderate, as of the push of 2026-10-13. Not yet triaged**

Status: closed
Owner: unassigned
Closed: 2026-10-14

**done 2026-10-14**: all vulnerabilities reported by `govulncheck` were in `golang.org/x/net`
(bumped `v0.56.0` → `v0.60.0`) and standard-library `net/http` (`go.mod` bumped to `go 1.26.9`).
`govulncheck ./...` now reports zero vulnerabilities

## staleAbaDocs

**Spec and docs still describe the ABA era, or older counts**

Status: closed
Owner: unassigned
Closed: 2026-10-14

**done 2026-10-14**:
[tech.fetchRestrictedInputs](requirements/local/tech/delivery.md#fetchrestrictedinputs),
[tech.datasetNeverCommitted](requirements/local/tech/delivery.md#datasetnevercommitted),
[tech.goVersionFollowsDeps](requirements/local/tech/delivery.md#goversionfollowsdeps),
[tech.localNowPublicLater](requirements/local/tech/posture.md#localnowpubliclater),
[tech.pacedInaturalistMapping](requirements/local/tech/process.md#pacedinaturalistmapping) and
[product.selfDescriptionStaysTrue](requirements/local/product/presentation.md#selfdescriptionstaystrue)
updated for post-[decisions.scopeFromRecurrence](requirements/decisions.md#scopefromrecurrence);
[product.traitsFromAvonet](requirements/local/product/dataset.md#traitsfromavonet) qualified to
match
[product.lengthAndWingspanSources](requirements/local/product/measurements.md#lengthandwingspansources)'s
three-sheet lookup;
[acceptance.scopeMatchesStatusTable](requirements/acceptance.md#scopematchesstatustable),
[acceptance.noRarityCodes](requirements/acceptance.md#noraritycodes),
[acceptance.allFiveTraits](requirements/acceptance.md#allfivetraits),
[acceptance.joinCoversEverySpecies](requirements/acceptance.md#joincoverseveryspecies) and the "no
code" notes in [acceptance.md](requirements/acceptance.md) updated to the 1,082-species derived
scope; `AGENTS.md`, `README.md`, `web/web.go`, [docs/building.md](../docs/building.md)
([decisions.speciesUrlOutlivesSpecies](requirements/decisions.md#speciesurloutlivesspecies) bookmark
note), [docs/data.md](../docs/data.md) (1,082 counts, withdrawn Böhmer neck derivation,
dropped-species count, [decisions.scopeFromRecurrence](requirements/decisions.md#scopefromrecurrence) status) and [docs/using.md](../docs/using.md)
(rarity codes sentence) updated

## keptSpeciesEstablished

**Species kept going by people are drawn as established, which [product.declaredExceptions](requirements/local/product/dataset.md#declaredexceptions) forbids**

Status: closed
Owner: Sameer
Closed: 2026-10-14

**done 2026-10-14**,
[decisions.escapedBirdsOffIndex](requirements/decisions.md#escapedbirdsoffindex),
[product.escapedBirdsListed](requirements/local/product/measurements.md#escapedbirdslisted),
[acceptance.escapedBirdsListedAndServed](requirements/acceptance.md#escapedbirdslistedandserved). 41
escaped cage, aviary and farm birds, Emu among them, are declared in `status.Exotic`, left off the
index and its controls, and listed at `/excluded`; their pages stay. Long-feral populations such as
Indian Peafowl in Florida and Red Junglefowl on Kauai are not listed

## hbwConsulted

**Register the HBW-BirdLife checklist as consulted, not ingested**

Status: closed
Owner: Sameer
Closed: 2026-10-13

**done 2026-10-13**: [`birdlife-checklist`](requirements/imports.md#birdlife-checklist).
Checking its terms found the problem closed above

## europeanHerringGullRow

**European Herring Gull is read from the combined 2022 row**

Status: closed
Owner: Sameer
Closed: 2026-10-13

**done 2026-10-13**,
[decisions.herringGullEuropeanRow](requirements/decisions.md#herringgulleuropeanrow),
[acceptance.herringGullFromEuropeanRow](requirements/acceptance.md#herringgullfromeuropeanrow). It
now reads `AVONET1_BirdLife`'s European-only *Larus argentatus*, labeled `measured`

## americanHerringGullRow

**American Herring Gull is labeled `measured` from a three-species row**

Status: closed
Owner: Sameer
Closed: 2026-10-13

**decided 2026-10-13: kept as it is**,
[decisions.herringGullEuropeanRow](requirements/decisions.md#herringgulleuropeanrow).
`AVONET1_BirdLife`'s *Larus smithsonianus* is the narrowest row AVONET has; 4 of its 8 specimens are
tagged American Herring Gull and the other 4 carry no subspecies

## vegaGullDonor

**Vega Gull's donor**

Status: closed
Owner: Sameer
Closed: 2026-10-13

**done 2026-10-13**, [decisions.vegaGullBorrows](requirements/decisions.md#vegagullborrows)
(evidence re-based 2026-10-14 on `AVONET_Raw_Data` and `AviList`). `AVONET1_BirdLife`'s *Larus
smithsonianus* holds the four Russian and Japanese specimens alongside four North American ones, so
Vega Gull borrows that row, as American Herring Gull uses it, and both tiles say *Same measurements
as* the other
([product.borrowedMeasurementsSaySo](requirements/local/product/measurements.md#borrowedmeasurementssayso),
[acceptance.sharedMeasurementsDisclosed](requirements/acceptance.md#sharedmeasurementsdisclosed))

## regionsTrinomials

**`birdview-regions` does not fold iNaturalist trinomials**

Status: closed
Owner: Sameer
Closed: 2026-10-06

**done 2026-10-06**,
[acceptance.occurrenceReadersAgree](requirements/acceptance.md#occurrencereadersagree). The rank
reduction moved into `taxonomy.Binomial` and runs inside `taxonomy.Occurrence` *before* either
naming table is consulted, so `dryobates villosus harrisi` reaches `Leuconotopicus villosus` instead
of nothing. The last of the three asymmetries
[decisions.regionJoinNormalizes](requirements/decisions.md#regionjoinnormalizes) found, and the
criterion now asserts the property the three shared — that both readers land on the same AviList
species — rather than the three bugs. The orphan residue was 0.0431% before the fold; what stays is
hybrids, which must never match a species, and subspecies of birds BirdView does not draw

## regionsRenames

**`birdview-regions` applies neither `AviListRenames` nor `CountedAs`**

Status: closed
Owner: Sameer
Closed: 2026-10-05

**done 2026-10-05**,
[decisions.regionJoinNormalizes](requirements/decisions.md#regionjoinnormalizes). Both readers of
the occurrence data now call `taxonomy.Occurrence`, and the join reports every name it could not
match, refusing to write above 0.05% of records. The audit that justified it: 1,402 eBird names, 37
matching nothing, 9,954,362 recoverable records across ten species

## ebirdFirstCounting

**Count eBird alone, and use iNaturalist only where eBird says it is hiding a bird**

Status: closed
Owner: Sameer
Closed: 2026-10-12

Planned 2026-10-09; one change in four parts, in this order. **Done: parts (1) and (2) 2026-10-09**
([decisions.ebirdSilenceNotAbsence](requirements/decisions.md#ebirdsilencenotabsence),
[acceptance.withheldMonthsExplained](requirements/acceptance.md#withheldmonthsexplained)) — the list
restated and confirmed, 202 cells flagged, the month filter fixed; **parts (3) and (4) 2026-10-12**
([decisions.countEbirdFirst](requirements/decisions.md#countebirdfirst),
[acceptance.ebirdCountedInatFillsGaps](requirements/acceptance.md#ebirdcountedinatfillsgaps)) —
eBird counted, iNaturalist only in confirmed gaps, shared names carrying eBird's count,
`WithheldByEBird` derived, scope re-run: 1,083 drawn species to 1,082, Mottled Owl the one to leave.
**(1) A live bug comes first.** GBIF's eBird export applies eBird's published [Sensitive Species
list](https://support.ebird.org/en/support/solutions/articles/48000803210-sensitive-species-in-ebird),
including its *regional and seasonal* entries, so eBird is silent for some species in some regions
and months — and [decisions.monthPresence](requirements/decisions.md#monthpresence)'s month filter
reads that silence as absence. **Painted Bunting is hidden from Florida in January**, the heart of
its Florida season (eBird: 0 records October–March, 12,483 in April), and **Short-eared Owl from New
York in June** (0 in May and June, 3,709 in January). Suppressed region-months must abstain and say
*withheld by eBird*, not count as absent. **(2) Encode the list** for the US and Canada as a
committed table: species, region, date window. Of BirdView's species, 9 are on its global list —
exactly the mainland and Hawaiian entries of `WithheldByEBird` — and **26 more on its regional
list**, among them Long-eared, Short-eared, Burrowing, Boreal and Barn Owls, Golden Eagle, American
Goshawk, King Rail, Loggerhead Shrike, Painted Bunting and Whooping Crane; only Black Rail was
known. Verified against the export: Long-eared Owl has 15 eBird records in the US against 17,015 in
Canada. The list's formatting is irregular (a scrape missed Barn Owl's regional entry), so the table
needs building carefully, and eBird's terms checking before anything derived from it is committed.
**(3) Change what counts.** eBird alone everywhere; iNaturalist only in cells eBird declares hidden
— the 9 globally withheld species (100 presences) and the 57 region cells where eBird is silent all
year (Long-eared Owl in about 30 states, Whooping Crane, Burrowing Owl in the prairie provinces,
Spruce Grouse in New York, Black Rail) — judged on iNaturalist alone and labelled as such. This
drops **985 of the 1,228 presences that rest on iNaturalist today**: 406 species, almost all escapes
and domestic birds (Red Junglefowl in 42 regions, Japanese Quail in 27, Zebra Finch, Indian Peafowl,
Budgerigar, Emu) and out-of-range strays. The figures barely move — median +0.6% — because
iNaturalist is 1.24% of land records. It supersedes
[product.regionAndMonthFilter](requirements/local/product/filtering.md#regionandmonthfilter)'s "in
either source", which judges iNaturalist against its own small total so one observation passes.
**(4) Shared names and scope.** With iNaturalist out of the counts, Northern Yellow Warbler loses
all 53,887 records (eBird files them under *Setophaga petechia*), so both halves of a shared name
take eBird's shared count and regions, as
[product.namesRecordsFiledUnder](requirements/local/product/presentation.md#namesrecordsfiledunder)
already describes. `WithheldByEBird` is derived from the list rather than kept by hand, which
removes **Desertas Petrel**: it is not on eBird's list, so its missing records are probably a naming
issue, the error Blue-billed White-Tern was. And `birdview-status` must be re-run to see which
species leave scope without iNaturalist recurrence; the six mainland globally withheld species
probably need declaring under
[product.declaredExceptions](requirements/local/product/dataset.md#declaredexceptions), as the
three Hawaiian ones already are

## seaShareDenominator

**Restore the share at sea on an off-land denominator**

Status: closed
Owner: Sameer
Closed: 2026-10-05

**done 2026-10-05**,
[decisions.honestSeaDenominator](requirements/decisions.md#honestseadenominator). Sea totals are now
every bird on open water in the box: North Atlantic 12,986,514 against the 429,535 shipped.
`Commonest` moved onto the same table the index divides, so the two pages cannot disagree again, and
**45 species peak at sea**. Remaining: the figure is *on water* rather than *at sea*, overstating by
~3% through estuaries, and a WKT polygon would fix that if it ever matters

## shoreDistance

**Is there a distance from shore that tightens the marine boxes?**

Status: closed
Owner: Sameer
Closed: 2026-10-05

**answered 2026-10-05**, no: distance is the wrong instrument and there is no gradient to tune.
Composition does not taper offshore, it flips between the band holding coastline and the band that
does not — 4.9% off land and Northern Cardinal at one step, 100% off land and pure tubenoses at the
next. GADM membership does the job exactly and for free. Written up under
[decisions.seaShareDiffers](requirements/decisions.md#seasharediffers)

## pelagicExceptionWording

**Reword the four pelagic exceptions in [product.declaredExceptions](requirements/local/product/dataset.md#declaredexceptions)**

Status: closed
Owner: Sameer
Closed: 2026-10-05

**done 2026-10-05** with [decisions.seaShareDiffers](requirements/decisions.md#seasharediffers).
They said the derivation sees these birds rarely "because almost nobody is on the water where it
is", which stated a measurement nobody made: the recurrence test runs on `gadmGid`, GBIF has no
`gadmGid` for open ocean, and `DatesIn` has no coordinate-box form, so pass 2 has never asked the
water anything. Each now says the records are at sea *where the test cannot yet look*, and names
what the derivation does see on land. **The exceptions stand** — whether these species would clear
the bar unaided remains unknown, and that is now what the table says

## vendorFoxWingspan

**Retrieve, hash and vendor `fox-wingspan`; write its `PROVENANCE.md`**

Status: closed
Owner: unassigned
Closed: 2026-10-04

done: `spec/sources/foxwingspan/PROVENANCE.md` exists, `spec/check/pins.py` verifies the hash, and
the ingest reads it for 300 measured wingspans

## writeAcceptance

**Write `acceptance.md` — Gate 2, before any code**

Status: closed
Owner: unassigned
Closed: 2026-10-04

done: 68 criteria, and Gate 2 passed long before the current work

## scopeChangeUrls

**Decide what happens to a URL when the derived scope changes**

Status: closed
Owner: Sameer
Closed: 2026-10-05

**done 2026-10-05**,
[decisions.speciesUrlOutlivesSpecies](requirements/decisions.md#speciesurloutlivesspecies) and
[product.undrawnSpeciesAnswer](requirements/local/product/presentation.md#undrawnspeciesanswer).
`/species/<id>` for a bird the status table classified but left out of scope now answers **404 with
the bird named and its measured reason**, covering **328 of the 334 species** the derivation
rejected — 124 vagrant, 167 absent, 37 historical; the other six have no AviList row, so no name and
no stable id. [decisions.compareUrlContract](requirements/decisions.md#compareurlcontract)'s *name what was lost* carries over; its **200 does
not**, because a comparison minus one bird is still a comparison and a detail page minus its bird is
nothing. 410 was rejected for overclaiming: scope is regenerated, so a bird out today can be in
after the next run. Cost a tenth column in `data/status.tsv` — the date last recorded — without
which "historical" could be asserted and not shown

## abaDecision

**ABA has not replied. Decide between chasing the permission and publishing a derived list**

Status: closed
Owner: Sameer
Closed: 2026-10-02

decided: stop waiting, use public data. Names moved to AviList
([decisions.avilistReplacesAba](requirements/decisions.md#avilistreplacesaba)); **scope is what
remains**

## derivedScope

**Replace ABA scope with a list BirdView derives and owns**

Status: closed
Owner: Sameer
Closed: 2026-10-05

**done 2026-10-05, and this was the deployment blocker.**
[decisions.scopeFromRecurrence](requirements/decisions.md#scopefromrecurrence): scope is recurrence,
not abundance. `birdview-ingest` reads the committed [`data/status.tsv`](../data/status.tsv);
`internal/aba` is deleted; nothing reads `fetched/`; the ABA credit is gone from all three footers
because nothing on a page comes from it. **1,083 species in 99 families**, up from 839 — 830 of the
old list kept on measured evidence, 9 kept by hand under
[product.declaredExceptions](requirements/local/product/dataset.md#declaredexceptions), and the
rest admitted because they recur. Every input the build reads is now CC BY 4.0 or the project's own
work

## abaPermission

**Ask ABA for written permission to redistribute the checklist**

Status: closed
Owner: Sameer
Closed: 2026-10-02, superseded by abaDecision

**sent 2026-09-27**, awaiting reply

## neckFlightPosture

**Decide the neck's flight posture**

Status: closed
Owner: Sameer
Closed: 2026-09-23

done, [decisions.foldedNeckAsS](requirements/decisions.md#foldedneckass)

## cervicalVertebrae

**Investigate cervical vertebra counts as an anchor for the neck model**

Status: closed
Owner: unassigned
Closed: 2026-10-04

answered twice, and both answers are no. As a **calibration** anchor: counts are dimensionless and
the trunk's are conserved, so they cannot do it;
[`bohmer-neck`](requirements/imports.md#bohmer-neck) measures neck *lengths* instead and now
calibrates [product.trunkFromMass](requirements/local/product/diagram.md#trunkfrommass). As a
**display** device — one dash per vertebra in the neck stroke — the neck is too short to carry them,
measured 2026-10-04 over all 839 species with each bird filling its box: the median neck is **14.8px
on the detail page and 7.8px on a tile**, which at a typical fourteen cervicals is **1.06px per
vertebra**. Only American Flamingo (76px), Trumpeter Swan (68px) and Great Blue Heron (65px) reach
even 4px per vertebra, and only on the detail page. Two further blocks if the geometry had worked:
dash style already carries provenance
([product.measuredVersusModeled](requirements/local/product/measurements.md#measuredversusmodeled)),
so dash *count* would overload the one mark that says "not a measurement"; and Böhmer covers 103
species against BirdView's 839, so ~88% of counts would be invented

## cubeClearanceTest

**Promote the cube/crossbar clearance invariant from prose into a test over the real dataset**

Status: closed
Owner: unassigned
Closed: 2026-10-04

done 2026-10-04: `TestTileCubeNeverReachesTheCrossbar` asserts it over 839 species at both scales
and all three metrics, and **logs the tightest margin** (33–41px) rather than asserting a figure
read off a run. A boolean check was tried first and rejected as too weak: the cube is bottom-left
and the crossbar upper-middle, so only a gross change breaks the conjunction

## comparisonSort

**Add a Sort by control to the comparison view**

Status: closed
Owner: Sameer
Closed: 2026-10-01

built, [acceptance.comparisonSortRewritesUrl](requirements/acceptance.md#comparisonsortrewritesurl)

## massWaterCube

**Decide whether mass gets an equal-mass cube of water drawn in the diagram**

Status: closed
Owner: Sameer
Closed: 2026-10-01

built, [decisions.massAsWaterCube](requirements/decisions.md#massaswatercube)

## cubeColor

**Give the mass cube a colour that is not the crossbar's**

Status: closed
Owner: Sameer
Closed: 2026-10-02

resolved the other way round: the **crossbar** moved to green `#4ade80`, because the cube's blue
depicts the water and should keep it. The recolour also fixed a live 7:1 violation, the labels
sharing that colour being text

## massOneUnit

**Mass units switch at 16 oz, so one section can print `14.5 oz` beside `1.2 lb`**

Status: closed
Owner: unassigned
Closed: 2026-10-04

done 2026-10-04,
[product.oneUnitInComparison](requirements/local/product/presentation.md#oneunitincomparison) and
[acceptance.massInOneUnit](requirements/acceptance.md#massinoneunit): the mass unit is now a
property of the comparison **scope** — a section or the whole set on the index, following the scale
setting; the birds named in the address on the comparison page — and the **smallest** bird in the
scope chooses it. A swan beside a sparrow reads `390.5 oz` and `0.93 oz`; two large birds keep
pounds. Taking the largest instead was tried and rejected: it crushes a 27 g sparrow to `0.1 lb`

## staleDocs

**`README.md` and `docs/` do not describe the current product**

Status: closed
Owner: Sameer
Closed: 2026-10-04

done 2026-10-04. `using.md` rewritten (it said Global was the default scale, the opposite of the
page, and documented none of Regions, Diet and Habitat, compare, `/added` or the served spec);
`building.md` had a mangled code fence rendering four paragraphs of prose as code on the served
page, listed two of six commands and eight of fourteen packages, and said the ABA permission request
was "drafted and unsent" when it was sent 2026-09-27; `data.md` claimed "Taxonomy is ABA's" two
paragraphs after saying every name comes from AviList. README gained the GBIF row and the
[decisions.scopeFromRecurrence](requirements/decisions.md#scopefromrecurrence) status.
[product.selfDescriptionStaysTrue](requirements/local/product/presentation.md#selfdescriptionstaystrue)
keeps the reading gate: none of this was catchable by
[acceptance.creditsComplete](requirements/acceptance.md#creditscomplete), which compares counts and
credits, not whether prose is true

## deadSwatchRules

**Delete the dead `.sw.*` swatch rules in `styles.css`**

Status: closed
Owner: unassigned
Closed: 2026-10-04

done 2026-10-04: eight rules removed, no template has emitted `class="sw …"` since
[decisions.perSectionDefault](requirements/decisions.md#persectiondefault)

## classicViewFate

**Keep, change, or retire the classic view at /classic**

Status: closed
Owner: Sameer
Closed: 2026-10-27, retired by Sameer

The index moved to `/classic` on 2026-10-26, unchanged, with its own links, `/now`, compare mode
(`/compare`), and the guide's examples (`docs/using.md`'s Some things to try) pointing at it. It is
not linked from the home page except in the credits. Decide whether it stays, and if it goes, what
replaces compare mode and the guide's examples. Either way, the requirements still say "the index"
in about 37 places; [product.nextView](requirements/local/product/presentation.md#nextview) says that now means the classic view, but a pass
should name the classic view or the home page in each. `docs/using.md` also still describes the
classic view's panels (Regions and months, Presentation) as the way in.

Closed 2026-10-27: the owner removed the classic view, `/compare`, `/now`, and pick mode
([decisions.classicViewRemoved](requirements/decisions.md#classicviewremoved)). The home page's pins replace compare mode, the guide's examples and panels now describe the
home page, and the requirements that only the classic view satisfied are withdrawn; the rest name
the home page.

## togglesStartSelected

**Tapping a month or region when all are selected removes it**

Status: closed
Owner: Sameer
Closed: 2026-10-27

All four simulated readers, on both revisions, tapped October or New York and lost it, because
everything starts selected and a tap drops one ([product.regionAndMonthFilter](requirements/local/product/filtering.md#regionandmonthfilter)); three of the
four built the address by hand instead. Karen's wording complaint
([#1](https://github.com/google/birdview/issues/1)) was the same problem seen from the help text.
One answer: when everything is selected, a tap on one item selects only it, as Only does, and
later taps add and remove. That changes the requirement, and families too. The home page adopted
this rule on 2026-10-26 ([product.nextView](requirements/local/product/presentation.md#nextview)); what is left is the classic view, if it stays
([classicViewFate](#classicviewfate)).

Closed 2026-10-27: the classic view went ([decisions.classicViewRemoved](requirements/decisions.md#classicviewremoved)), so the home page's rule is the only one left.

## compareNumbersUnreadable

**The comparison's numbers cannot be read on a phone**

Status: closed
Owner: unassigned
Closed: 2026-10-27

Both phone readers, on both revisions, could not read the length, wingspan, and mass labels inside
the comparison's drawings, and opened each bird's page instead
([product.comparisonOneScale](requirements/local/product/presentation.md#comparisononescale)). Show the three numbers as text under each panel, or scale the
labels with the drawing. The home page writes the three numbers as text under every drawing;
what is left is `/compare`, which the classic view's compare mode reaches
([classicViewFate](#classicviewfate)).

Closed 2026-10-27: `/compare` went with the classic view ([decisions.classicViewRemoved](requirements/decisions.md#classicviewremoved)); the home page's comparisons, made
with pins, write the three numbers as text under every drawing.

## wingspanSectionWidth

**Per-section scale under Wingspan is not bounded in total length**

Status: closed
Owner: Sameer
Closed: 2026-10-27

Found 2026-10-19 while setting Bill and Tail's default scale
([product.scaleToggle](requirements/local/product/diagram.md#scaletoggle)): with 15% sections, one
wingspan section runs from the 146 mm Elf Owl to the 1,789 mm Emu, 12.3× in total length, because a
flightless bird's short wings sort it among small owls. Length sections stay within 1.1× and mass
within 3.1×. Either Wingspan should default to Global as Bill, Tail, and Frequency do, or flightless
birds need handling in the wingspan sort.

Closed 2026-10-27: per-section scale went with the classic view
([decisions.classicViewRemoved](requirements/decisions.md#classicviewremoved)); the home page draws
every bird to one scale, so no section's width decides a scale.
