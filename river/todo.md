# Open work

Search `## <id>` for one item, or `grep -n -A1 '^## ' river/todo.md` for the list. `Status:` is
open, parked or deferred; `Gate:` says when it is due.

Everything outstanding in BirdView: unfinished work, open questions, known defects, criteria not
yet watched failing, and stale spec or doc text. Each item points at the requirement, decision or
criterion that holds its reasoning.

**Adding an item.** Add it the moment the work is found: a camelCase identifier as the heading,
the title in bold, then `Status:`, `Owner:` and `Gate:`. Never leave a `TODO` in code or a "left
open" note in a decision record instead.

**Closing an item.** Set `Status: closed`, replace `Gate:` with `Closed:` and what closed it, and
move it to [todo-closed.md](todo-closed.md). Items are never deleted: a closed one records why a
question that looks open is not. Design notes behind items are in [notes.md](notes.md).

## tightenProse

**Tighten all prose and page text; less is more**

Status: open
Owner: Sameer
Gate: when convenient

First pass done 2026-10-20: README, AGENTS.md, CONTRIBUTING, CHANGELOG, the four guides, every
product and tech requirement, `app.md` and the page templates, about 30% shorter overall. Left:
`decisions.md` and `acceptance.md` (history and verification records; decide whether to tighten
them at all), `imports.md`, `notes.md`, `arch.md`, `bindings.md`, code comments, and page
text built in Go, chiefly `cmd/birdview/main.go` (group notes, the undrawn page's verdicts, region
and category row details, record-ratio tooltips, tile notes, and the `/added` and `/excluded`
intros) and the grouping notes in `internal/region/region.go`. Several of those strings are
asserted by tests; grep before editing.

## oneVersionRecord

**Record each imported source's version in imports.md only**

Status: open
Owner: Sameer
Gate: when convenient

River's decision `requirementsReadme` puts a source's version in `imports.md` and leaves
`PROVENANCE.md` with retrieval details and the hash. `aba` and `naturalearth` still carry their own
`Version` row in `PROVENANCE.md`, a second record of the same fact; drop it there, keeping any
detail `imports.md` lacks (the ABA species count, Natural Earth's `VERSION.txt` evidence). Versions
inside file names and URLs stay, since they are the files' names.

## massLabelPlacement

**labeledDetailDiagram says two different things about where the mass label goes**

Status: open
Owner: Sameer
Gate: needs the owner's call

[product.labeledDetailDiagram](requirements/local/product/measurements.md#labeleddetaildiagram)
says total length and mass are labeled below the tail tip, and later that mass is labeled under its
cube, not in the stack under the tail. Found during the prose pass; it predates it. Check what the
page does, and correct the statement that is wrong.

## stableSurface

**State what BirdView promises to keep stable, so the rest is free to change**

Status: open
Owner: Sameer
Gate: before a public deployment, and before [publishDataset](#publishdataset)

[goals.freedomToChange](goals/goals.md#freedomtochange) has one requirement
(servesOnlyItsPages). Write requirements naming the stable surface (likely `/species/<id>`,
`/compare` and the documented query parameters) and declaring everything else internal,
including the shape of `species.json`. This pulls against publishing the dataset: expect a
separate external schema, versioned and documented, with the internal one free to change.

## craniumAtLargeMass

**The cranium model overdraws the largest birds, and one is now in scope**

Status: open
Owner: Sameer
Gate: when convenient

[product.craniumFromMass](requirements/local/product/diagram.md#craniumfrommass) said its
overestimate at large mass cost nothing because the heaviest bird in scope was 11 kg. Emu is now in
scope at 34 kg and draws a 243 mm cranium. Check it against a measured emu skull, and if it is far
off, fit or cap the coefficient.

## wingspanSectionWidth

**Per-section scale under Wingspan is not bounded in total length**

Status: open
Owner: Sameer
Gate: when convenient

Found 2026-10-19 while setting Bill and Tail's default scale
([product.scaleToggle](requirements/local/product/diagram.md#scaletoggle)): with 15% sections, one
wingspan section runs from the 146 mm Elf Owl to the 1,789 mm Emu, 12.3× in total length, because a
flightless bird's short wings sort it among small owls. Length sections stay within 1.1× and mass
within 3.1×. Either Wingspan should default to Global as Bill, Tail and Frequency do, or flightless
birds need handling in the wingspan sort

## sexSpecificDrawings

**Separate drawings for males and females are not supported by the data; investigated 2026-10-18 and deferred**

Status: open
Owner: Sameer
Gate: when a usable sex-specific mass source is found

Mass is the blocker: it is pooled per species in every vendored source, and head, body and the mass
cube are modeled from it
([product.craniumFromMass](requirements/local/product/diagram.md#craniumfrommass),
[product.trunkFromMass](requirements/local/product/diagram.md#trunkfrommass)), so per-sex bill,
wing and tail on a shared body would draw a bird nobody measured
([product.tileMarksModeled](requirements/local/product/measurements.md#tilemarksmodeled)). What
exists: AVONET's unused `AVONET_Raw_Data` sheet sexes each specimen for bill, wing chord, secondary,
tail and tarsus, but only **213 of 1,082** drawn species have five or more of each sex, and **9 of
61** drawn raptors; Fox's sex-specific wingspan columns are filled for **2** drawn species. Where
samples suffice, raptor females run larger in wing chord (Sharp-shinned Hawk +15%, Peregrine +9%)
and males larger in grackles, cowbirds, Ruff and pheasant (+11–22%); a 5–9% gap in a passerine from
five specimens per sex is likely noise. **Leads for sex-specific mass, none checked:** Myhrvold et
al. 2015's amniote database (believed public domain, with a figshare copy; license and coverage
unverified). Lislevand et al. 2007 and Ecology's data papers generally sit on **Wiley, which
BirdView may not read**
([decisions.ricklefsProhibited](requirements/decisions.md#ricklefsprohibited)); check where any
candidate is hosted before touching it. Dunning's handbook and Birds of the World are closed. A
cheaper step that needs no new source: per-sex wing chord and tail as numbers, with specimen counts,
on the detail page

## approveAppComposition

**Approve the first composition, `app`, and decide whether to tag it**

Status: open
Owner: Sameer
Gate: before composing `dataset`

[requirements/composed/internal/app/](requirements/composed/internal/app/app.md) composes accessibility only: WCAG 1.4.6 and 1.4.11 over
every web page as [tech.webPages](requirements/local/tech/accessibility.md#webpages) defines one,
and the stylesheet's contract. An agent composed it; River requires a person to sign off. The
rulings it embeds: plain-text error pages are web pages (WCAG's definition leaves no choice), large
text gets no allowance ([decisions.oneContrastFloor](requirements/decisions.md#onecontrastfloor)), and closed panels are measured open

## nonTextContrastCheck

**WCAG 1.4.11 has no check**

Status: open
Owner: Sameer
Gate: before any accessibility claim

[app.pagesMeetNonTextContrast](requirements/composed/internal/app/app.md#pagesmeetnontextcontrast) is uncovered: the
stylesheet checks skip shapes on purpose and the rendered check measures text only. Extending the
rendered check to strokes and fills against adjacent colors is the obvious route; "required to
understand the content" needs a ruling per diagram part

## contrastCheckNeedsChrome

**The rendered contrast check gates only where Chrome is installed**

Status: open
Owner: Sameer
Gate: when CI exists

[app.renderedTextContrast](requirements/composed/internal/app/acceptance.md#renderedtextcontrast) skips without it,
and BirdView has no CI. A skipped check reads as a pass in `go test ./...` output unless someone
reads the skip line

## composeDatasetAndRepo

**Compose `dataset` and `internal/repo`**

Status: open
Owner: Sameer
Gate: after `app` is approved

, which must account for the 24 local requirements no package lists yet (`TestCompositionResolves`
logs them), then move each criterion out of [acceptance.md](requirements/acceptance.md) into its
package with `Covers:` and delete the Traceability table once nothing is left in it. **Four wait on
the morphometric-model ruling** — [product.craniumFromMass](requirements/local/product/diagram.md#craniumfrommass), [product.trunkFromMass](requirements/local/product/diagram.md#trunkfrommass),
[product.neckModelPerFamily](requirements/local/product/diagram.md#neckmodelperfamily) and [product.generatedBadgeColors](requirements/local/product/dataset.md#generatedbadgecolors): each produces a value the app draws,
and whether it is data (`dataset`) or a drawing instruction (`app`) is the owner's call

## largeSourceFiles

**Two source files are large enough to make reading by anchor matter**

Status: open
Owner: Sameer
Gate: when evidence warrants

`presentation.md` is about 13,000 tokens for 20 requirements and `measurements.md` about 6,700.
Agents are told to read by anchor (AGENTS.md), so splitting is deferred; split if trajectories show
whole-file reads

## licenseConditionFirst

**Statement-first: one requirement still leads with a licence condition**

Status: open
Owner: Sameer
Gate: with the split

[product.selfDescriptionStaysTrue](requirements/local/product/presentation.md#selfdescriptionstaystrue) opens with the CC BY credit rule, which is the imported half of a
mixed requirement; its statement is fixed when it is split (row above on separating restated
imports)

## separateRestatedRequirements

**Before composition, separate the requirements that restate an imported source**

Status: open
Owner: Sameer
Gate: before composition

Classified 2026-10-16, each of the 81 read individually. **Composed — wholly an imported obligation,
restated:**
[tech.pacedInaturalistMapping](requirements/local/tech/process.md#pacedinaturalistmapping)
([`inatapi.RequestRate`](requirements/imported/inaturalist.org/inatapi/inatapi.md#requestrate)),
[tech.identifiedClient](requirements/local/tech/delivery.md#identifiedclient)
([`inatapi.UserAgent`](requirements/imported/inaturalist.org/inatapi/inatapi.md#useragent)),
[tech.badgeContrast](requirements/local/tech/accessibility.md#badgecontrast) and
[tech.textContrastFloor](requirements/local/tech/accessibility.md#textcontrastfloor) (WCAG 2.2 SC
1.4.6 and 1.4.11). These belong in a composed package citing the source, not in `tech`. **Mixed — a
local requirement with an imported obligation inside it, to split:**
[tech.fetchRestrictedInputs](requirements/local/tech/delivery.md#fetchrestrictedinputs) (*do not
redistribute* is the licences'; *fetch at build time* is BirdView's mechanism),
[product.selfDescriptionStaysTrue](requirements/local/product/presentation.md#selfdescriptionstaystrue)
(*credit every CC BY source* is a licence condition; docs currency is local),
[tech.localNowPublicLater](requirements/local/tech/posture.md#localnowpubliclater) (*never
commercial* follows the non-commercial licences), and
[product.inaturalistLink](requirements/local/product/presentation.md#inaturalistlink) (*reproduce
no photograph* is [`inatterms.ContributorKeepsRights`](requirements/imported/inaturalist.org/inatterms/inatterms.md#contributorkeepsrights)). **The other 73 are local intent**, including
those that name a source — choosing AviList for names is BirdView's decision, made with licences in
mind, not an obligation AviList imposes. Two findings on the way: the WCAG manifest entry still says
other text is held to 4.5:1, which
[decisions.oneContrastFloor](requirements/decisions.md#onecontrastfloor) replaced with one 7:1
floor; and Wikimedia and Macaulay Library are cited as rationale from the photograph era, with no
code reading either. **Progress 2026-10-16:** WCAG is now vendored and transcribed (`wcag`), the
stale manifest entry is fixed, and the two contrast requirements serve as the guarantee half of
[app.stylesheetContract](requirements/composed/internal/app/app.md#stylesheetcontract), with their WCAG obligation
composed separately as [app.pagesMeetTextContrast](requirements/composed/internal/app/app.md#pagesmeettextcontrast).
The four mixed requirements are still to split

## presenceOverSelection

**Judge presence over the whole selection, against a threshold the reader sets**

Status: parked
Owner: Sameer
Gate: parked

Chosen in principle 2026-10-09 (option B of the regions-scaling design) and parked before
implementation. A species is present when its share of the selection's records — the tile's own *1
in N* — beats a threshold shown as a control: *show birds at least* 1 in 10,000 · 100,000 ·
1,000,000 · 10,000,000 · 100,000,000 · any record. Proposed defaults, **undecided**: 1 in 1,000,000;
the arrival page left unfiltered so scope stays [product.scopeIsRecurringAvifauna](requirements/local/product/dataset.md#scopeisrecurringavifauna)'s job; and
[decisions.monthPresence](requirements/decisions.md#monthpresence)'s month rule applied to the summed counts of the selected regions. Measured
at 1 in 1,000,000: New York 491 → 443, the Northeast's nine states 564 → 470, all land regions 1,057
→ 836. Two defects it would fix, both real today:
[product.regionAndMonthFilter](requirements/local/product/filtering.md#regionandmonthfilter)'s "in
either source" judges iNaturalist against its own small total, so one observation passes — 45 of New
York's species are present on that alone, Diamond Dove among them; and `birdview-regions` discarded
counts under the bar at write time — **fixed on its own 2026-10-13,
[decisions.belowBarCounted](requirements/decisions.md#belowbarcounted)**: every count is now kept
and the bar decides membership only, which left 47 drawn species reading zero before it. Keeping
every count costs 11% more cells. Works on today's 74 regions; would make counties a data change
rather than a rules change

## gbifTablesNotRederivable

**Every committed table comes from the live GBIF API and cannot be re-derived**

Status: open
Owner: Sameer
Gate: before publishing the dataset

Each build command issues thousands of throttled HTTPS queries to `api.gbif.org`, so two runs a week
apart give different numbers and no past run can be reproduced — the figures pinned in `river/` and
in tests are evidence of one afternoon, not of a stated input. That is tolerable while BirdView only
*consumes* the data and becomes a defect the moment it **publishes its own dataset**, where a reader
must be able to check a number against the source it came from. Two routes, both of which also end
the 11-to-40-minute runs: GBIF's SQL download service, which aggregates server-side and returns a
small file with a citable DOI but needs an account; or a GBIF Parquet snapshot queried locally,
which needs no account — the better fit if others are to work with BirdView directly — at the cost
of tens of GB and a query engine. Either way the switch is **once, for the whole pipeline**: two
sources at different freshness is the disagreement
[decisions.honestSeaDenominator](requirements/decisions.md#honestseadenominator) already fixed.
Every pinned figure must be re-derived from the new input rather than adjusted to match

## unusedAvonetColumns

**AVONET's `Family1`–`3` and `Secondary1` are read and never used**

Status: open
Owner: unassigned
Gate: opportunistic

Found drawing
[product.pipelineDiagram](requirements/local/product/measurements.md#pipelinediagram)'s pipeline
diagram, which marks both ✗. `avonet.FamilyOf` is documented as a cross-check against ABA's family,
and ABA is gone ([decisions.scopeFromRecurrence](requirements/decisions.md#scopefromrecurrence));
`Secondary1` is a required column whose value nothing reads. Either use them — `FamilyOf` could
cross-check AviList's family the way it once checked ABA's — or stop reading them, and update the
diagram either way

## moreDomesticForms

**Three more domestic forms behave like the two that are declared**

Status: open
Owner: unassigned
Gate: opportunistic

The rank fold of
[acceptance.occurrenceReadersAgree](requirements/acceptance.md#occurrencereadersagree) moved
iNaturalist's trinomials into their wild species, and the proportional gainers were `Gallus gallus`
**+11.9%**, `Anser cygnoides` **+20.0%** and `Anser anser` **+6.9%** — the same signature as the
declared `Columba livia` (+1.3%) and `Cairina moschata` (+10.0%). `taxonomy.DomesticForms` names
only the latter two, so
[product.sourceNamingMeasured](requirements/local/product/dataset.md#sourcenamingmeasured)'s ratio
check reads the other three as undeclared disagreements if it ever flags them. Either declare them
or record why the two are special; measuring it needs a `cmd/birdview-names` run

## nameDisagreementCounts

**The record counts behind a name disagreement are not on the page**

Status: open
Owner: Sameer
Gate: opportunistic

[product.namesRecordsFiledUnder](requirements/local/product/presentation.md#namesrecordsfiledunder)
asks for each source's binomial *and how many records it holds under it*, and
[acceptance.disputedNamesShown](requirements/acceptance.md#disputednamesshown) ships only the
binomial. [`data/source-names.tsv`](../data/source-names.tsv) has per-source counts and no reader,
but it queries the **AviList** name only — for Hairy Woodpecker it records eBird 0, not the
7,310,342 under *Dryobates villosus*, which is the number a reader wants. Closing it needs
`cmd/birdview-names` to query the alias as well, a GBIF run, and the counts plumbed through the
ingest

## gbifBridgeUnread

**Nothing reads `data/gbif-taxa.tsv`**

Status: open
Owner: unassigned
Gate: opportunistic

It is generated by `cmd/birdview-gbifmap`, committed, and consumed by no code. It carries a
GBIF-to-AviList name bridge with an `ambiguous` tier that would be the natural home for
[decisions.sharedCountShownOnBoth](requirements/decisions.md#sharedcountshownonboth)'s shared names
— except that it files `Setophaga petechia -> Setophaga petechia` as `exact`, correctly, which is
why the bridge could not find them. Either wire it in or retire it

## throwawayNameSweep

**The name sweep is a throwaway and should not be**

Status: open
Owner: Sameer
Gate: before the next AviList release

Finding [decisions.sharedCountShownOnBoth](requirements/decisions.md#sharedcountshownonboth)'s cases
took 2,166 queries comparing what each database reports under each AviList binomial — the only
method that catches a name whose strings match and whose meanings differ. It ran once from a scratch
binary. As a committed command it would catch the next split the week it lands, rather than when a
reader notices a mangrove warbler in Ontario. Note its blind spot: a `country=US,CA` filter hides
the outlying islands, which is why *Sternula albifrons* looked like a finding and was not

## recurrenceCoordinateBox

**Give the recurrence test a coordinate-box form, and find out whether the pelagics clear the bar on their own**

Status: open
Owner: Sameer
Gate: opportunistic; it could retire hand-declared exceptions

`DatesIn` filters by `gadmGid`; GBIF assigns none to open ocean, so no pelagic's at-sea records have
ever been tested for recurrence. Four
[product.declaredExceptions](requirements/local/product/dataset.md#declaredexceptions) exceptions
exist in that blind spot, and
[decisions.scopeFromRecurrence](requirements/decisions.md#scopefromrecurrence)'s claim that
recurrence is insensitive to observer effort has never been tested where observers are scarcest. The
`eventDate` facet does accept a lat/lon box — probed 2026-10-05 — so the client method is small; the
cost is a re-run of `birdview-status` against live GBIF. Note the boxes overlap the coast
([decisions.seaShareDiffers](requirements/decisions.md#seasharediffers)), so a naive run would count
shore records as sea

## commonestIsLargest

**Assert that `Commonest` is the largest share**

Status: open
Owner: unassigned
Gate: with [decisions.seaShareDiffers](requirements/decisions.md#seasharediffers)

Nothing does. The server fixture hardcodes `commonestOf` and every test renders it back, so
[product.shareOfRecords](requirements/local/product/presentation.md#shareofrecords)'s central
claim has never met data — which is why
[decisions.seaShareDiffers](requirements/decisions.md#seasharediffers)(a) shipped. The check belongs
over the real dataset, as
[acceptance.outOfScopeExplained](requirements/acceptance.md#outofscopeexplained)'s id-collision half
is

## vagrantOutscoresResident

**Region membership counts records, so a one-off vagrant can outscore a resident**

Status: deferred
Owner: Sameer
Gate: deferred 2026-10-04; before anyone reads a region list as a range map

[product.regionAndMonthFilter](requirements/local/product/filtering.md#regionandmonthfilter)'s
1-ppm rule is anti-correlated with what it measures: the more surprising a bird, the bigger the
crowd, the more records. **Two examples are in the data as shipped** — American Flamingo in
Pennsylvania (931 records, one pond, thirteen days, post-Idalia) and **Lucy's Warbler, a Sonoran
desert bird, in Virginia** (53 records, one location, six days). **Three candidate remedies have
been tried and all three fail; they are recorded so the next attempt does not repeat them.** *(1)
Records in ≥2 of 3 four-year windows* — Lucy's Warbler walks through it, present 2016-12-31 to
2017-01-06, so one continuous event straddles a window boundary and scores 2/3. *(2) A threshold on
distinct places* — cannot work at any value: Akohekohe, a Hawaiian endemic, has 4 places and Wood
Stork in Nova Scotia, a vagrant, has 7. A critically endangered endemic with four records and a
vagrant with four records are the same shape. *(3) Local abundance against continental abundance* —
Flamingo/PA scores 1.61, above several birds that must be kept, because flamingos are rare
everywhere so the baseline is tiny. A composite — drop if ≥20 records **and** (≤2 places **or** ≤1
year) — keeps every endemic and withheld species correctly and catches Lucy's and the flamingo, but
still misses Wood Stork/NS, Spoonbill/ME, Fork-tailed Flycatcher/DE and Baltimore Oriole/NT.
**Measurement caveat for whoever resumes**: these figures came from one GBIF page of ≤300 records,
which is a *contiguous slice in GBIF's own ordering, not a random sample* — all 300 American Robin
records for Pennsylvania fell in January 2024, and a page-derived year count read 1 where the true
count is 60. Any real rule must take both signals from facets over all records, which raises the
cost. Costs a regenerated `data/region-presence.tsv` and dataset. **Knock-on: the "Common birds"
zones were clustered *from* that table
([decisions.regionsTwoWays](requirements/decisions.md#regionstwoways)), so they must be re-derived
rather than left as a stale snapshot.** Documented for readers meanwhile under
[Regions](../docs/data.md)

## coOccurrenceGrouping

**Group families by what is seen together**

Status: open
Owner: Sameer
Gate: opportunistic; needs a stratified crawl of ~400–800 county-days into a committed table

Crawl eBird checklists and cluster families by co-occurrence, as a third grouping
([decisions.familiesTwoWays](requirements/decisions.md#familiestwoways)). This is the honest version
of the habitat idea — co-occurrence is natively a *family*-level property, so nothing is averaged
and nothing is misfiled. Two findings make it tractable, both measured 2026-10-03: GBIF's eBird
export carries **no `eventID`**, but a checklist reconstructs from `recordedBy` + `eventDate` +
rounded coordinates **only from a complete slice** — a 0.1% sample of a state-month gives 96%
singleton checklists, while one county-day (1–3k records) gives real ones, median 19 species. A
602-checklist probe already separated waterbirds, woodland songbirds, open country, feeder birds and
**town birds** (pigeons, House Sparrows, starlings), leaving hawks and crows unattached because they
occur with everything. Expect the rare pelagic and Hawaiian families to be unplaceable, as the
atolls were

## publishDataset

**Publish the BirdView dataset as a source in its own right**

Status: open
Owner: Sameer
Gate: ready: the repo exists as `google/birdview` since 2026-10-08

The project consumes four published datasets and credits each by DOI; it now derives work worth the
same treatment — the occurrence status
([decisions.scopeFromRecurrence](requirements/decisions.md#scopefromrecurrence)), the region
presence and shares, the supplementary habitat and diet assignments, the neck and trunk models, and
the wingspan regression. None of it is reachable except by scraping the app. **What the shape should
be is the open question, not whether.** `web/data/species.json` is a serving artifact, not a
publication: it nests provenance inside every metric, repeats source strings across 1,083 species,
and carries presentation fields — `familyColor`, `familyTextColor`, `id` — that mean nothing outside
this app. A publication probably wants CSV, one table per thing, keys that join: `species.csv` (the
identity columns), `measurements.csv` (one row per species per metric, with its kind, source and
model, so provenance is a column rather than a nest), `categories.csv` (habitat and diet with
origin), `regions.csv` (species, region, records), `status.csv` (what `data/status.tsv` already is).
Plus the thing none of our own sources gave us and all of them needed: a `PROVENANCE.md` stating the
version, the license, what is measured against what is modeled, and the known distortions — eBird
withholding sensitive species, crowd-driven over-reporting, the 4 species with no AVONET row.
**Licensing needs deciding first**: the derived values are built from CC BY 4.0 inputs, so
attribution flows through, and BirdView should pick its own license rather than leave it implied

## ricklefsPermission

**Ask Ricklefs or ESA for permission to use the measured-length data**

Status: open
Owner: Sameer
Gate: **after a working demo**

## perchedToggle

**Assess whether a perched/standing toggle is achievable at any honest accuracy**

Status: open
Owner: unassigned
Gate: before it is offered to users

([decisions.foldedNeckAsS](requirements/decisions.md#foldedneckass))

## reviewNeckRatios

**Review `NeckRatios` against the full family list**

Status: open
Owner: Sameer
Gate: opportunistic; still the weakest model

24 of 90 families carry a ratio and **66 take the default 0.18, covering 548 species**. This is the
half of the old entry that is actually tractable, and the one that catches the next short-necked
Limpkin. ~~Derive the values from [`bohmer-neck`](requirements/imports.md#bohmer-neck)
instead~~ — **withdrawn 2026-10-04**: [`imports.md`](requirements/imports.md#bohmer-neck)
already records why, and the old entry ignored it. Böhmer measures the **full cervical column
including the part inside the body contour**, where
[product.neckModelPerFamily](requirements/local/product/diagram.md#neckmodelperfamily) models the
*exposed* neck of a silhouette. Converting one to the other needs the fraction of the column that
shows in flight, per family, unmeasured — which is exactly what `NeckRatios` already is. It would
move the guess, not remove it, and add a dependency on a 103-species `.docx` that is
consulted-not-ingested

## trunkMeasurementSource

**Find a source that measures a trunk**

Status: open
Owner: unassigned
Gate: opportunistic; would upgrade a calibration to a fit

[decisions.trunkExponentCalibrated](requirements/decisions.md#trunkexponentcalibrated)'s exponent is inferred from neck and leg, and that is its weakest
link

## massAsLine

**Replace the cube with a mass line proportional to mass**

Status: parked
Owner: Sameer
Gate: parked 2026-10-02 — resume from that section, do not re-derive

Replace the cube with a mass **line** proportional to mass. Measured and written up in [Mass as a
proportional line](notes.md#mass-as-a-proportional-line-replacing-the-cube--parked); needs a CR,
because it reverses [decisions.massAsWaterCube](requirements/decisions.md#massaswatercube)

## criteriaNeverFailed

**Eight acceptance criteria have never been watched failing**

Status: open
Owner: unassigned
Gate: opportunistic; acceptance.tileMarksNonMeasured and acceptance.nothingRestrictedTracked guard hard rules

Still `planned` in [acceptance.md](requirements/acceptance.md):
[acceptance.allFiveTraits](requirements/acceptance.md#allfivetraits) (all five AVONET traits on
every record),
[acceptance.joinCoversEverySpecies](requirements/acceptance.md#joincoverseveryspecies) (join
coverage monitored),
[acceptance.unmatchedFailsBuild](requirements/acceptance.md#unmatchedfailsbuild) (unmatched species
fail the build above a threshold),
[acceptance.proportionsDifferentiate](requirements/acceptance.md#proportionsdifferentiate)
(proportions differentiate species),
[acceptance.tileMarksNonMeasured](requirements/acceptance.md#tilemarksnonmeasured) (the tile marks
every non-measured metric),
[acceptance.nothingRestrictedTracked](requirements/acceptance.md#nothingrestrictedtracked) (nothing
restricted is tracked — the document calls it its highest-value check),
[acceptance.outboundRequestsIdentified](requirements/acceptance.md#outboundrequestsidentified)
(outbound requests identify the client) and
[acceptance.abaFileHashMatches](requirements/acceptance.md#abafilehashmatches) (the fetched ABA file
matches its hash). acceptance.joinCoversEverySpecies and acceptance.proportionsDifferentiate are
already cited by tests, so they may only need the mutation run and a status;
acceptance.abaFileHashMatches is probably obsolete since
[decisions.scopeFromRecurrence](requirements/decisions.md#scopefromrecurrence) removed the ABA
input, and wants withdrawing rather than verifying

## nearMatchSplits

**Four near-matches are splits awaiting a human taxonomic check**

Status: open
Owner: Sameer
Gate: needs the owner's call

[decisions.abaSpeciesWithoutAvonet](requirements/decisions.md#abaspecieswithoutavonet), *Still
unresolved*. *Astur atricapillus*, *Ardea ibis*, *Gygis candida* and *Tyto furcata* borrow from
AVONET rows for the combined pre-split taxon. Checked 2026-10-13: AviList v2025b lists each parent
and daughter as separate species, and no AVONET sheet has a row for any daughter, so `near-match` is
the right label. Closing it means recording that in a decision, folding the four into the main
`NearMatches` block, and deleting `taxonomy.AwaitingTaxonomicCheck` and the ingest line that prints
it

## avonetMassSplit

**AVONET's mass column was not split where BirdLife splits a pair**

Status: open
Owner: unassigned
Gate: opportunistic

[decisions.abaSpeciesWithoutAvonet](requirements/decisions.md#abaspecieswithoutavonet)'s caveat,
carried forward and never closed: *Larus smithsonianus*/*argentatus*, *Calonectris
borealis*/*diomedea* and *Pterodroma deserta*/*feae* each share one mass, so their linear traits
differ and sorting by mass shows them as identical. The same holds for every white tern (117 g)

## deploymentPlumbing

**Launch birdview.site from a personal mirror of this repository**

Status: open
Owner: Sameer
Gate: the next step for a public deployment

Decided 2026-10-22: `birdview.site` belongs to Sameer personally and is served at his own expense
from his own clone, so the site is a personal project rather than something this organization
operates. The Apache 2.0 license permits that. While this repository stays the source of truth,
`github.com/Sajmani/birdview` is a **one-way mirror**: changes land here, are pulled into the
mirror and deployed from it, and nothing is ever committed there directly, or the two would drift.
A mirror made now postdates the history rewrite, so [prerewriteClones](#prerewriteclones) does not
apply to it. In order:

1. ~~**Create the mirror** and push this repository's `main` to it.~~ Done 2026-10-22:
   `github.com/Sajmani/birdview`, a GitHub fork. Sync it by pulling, never by pushing to it: after
   `git push origin main`, run `gh repo sync Sajmani/birdview --branch main` (or the fork's "Sync
   fork" button), which only fast-forwards from `google:main`. A ruleset on its `main` blocks
   deletion and force-pushes; issues, pull requests, wiki and projects are off.
2. **Choose a host** that runs the container image (`Dockerfile`, [docs/building.md](../docs/building.md),
   serving `0.0.0.0:8080`). Deploying from a local checkout with the host's own tooling is enough
   to launch; automate after.
3. **Build the dataset for deployment.** It is never committed
   ([tech.datasetNeverCommitted](requirements/local/tech/delivery.md#datasetnevercommitted)), so the image build or the deploy step runs
   the ingest and supplies `species.json`. Today that happens only on the maintainer's machine.
4. **Amend the requirements for public serving**, requirement first:
   [tech.localNowPublicLater](requirements/local/tech/posture.md#localnowpubliclater) (its value is `local`),
   [tech.servesOnLocalhost](requirements/local/tech/delivery.md#servesonlocalhost), and the stable surface
   ([stableSurface](#stablesurface)).
5. **Write the first release criteria**, BirdView's first use of River's Release stage: a health
   check, monitoring and alerts, a rollback path, and what happens when an ingest input changes.
6. **Operational checks:** TLS, caching behavior
   ([decisions.hereAndNowCaching](requirements/decisions.md#hereandnowcaching)), no tracking, a feedback channel that exposes
   no personal address, basic rate limiting.
7. **Point `birdview.site` at the host.**

## sourceOfTruth

**Move BirdView's source of truth to the personal repository**

Status: open
Owner: Sameer
Gate: before building BirdView's software factory; and once continuing development in a personal
repository is confirmed as approved

Planned 2026-10-22: an agent-driven software factory for BirdView, run from Sameer's own accounts
the way a customer would run one, belongs where the code lives; if this repository stayed the
source of truth, every change the factory made would be an outside contribution to it. So, once
approved, `github.com/Sajmani/birdview` becomes the source of truth and this repository is archived
with a note saying where development continues. Exactly one repository receives changes at any
time. **First, confirm** that continuing development there, on personal time and accounts, is
approved: the existing files carry this organization's copyright, and serving a clone is not the
same question as developing one. Afterwards BirdView is an outside adopter of River: re-pins copy
River's published parts until River itself is public, then come from River's public origin.

## smallOrderWingspanFit

**Hummingbirds and gamebirds take the global wingspan fit**

Status: open
Owner: unassigned
Gate: opportunistic; needs a wingspan source for those orders

, not an order-level one, because neither order has five empirical species and Galliformes has none.
Their estimates carry roughly 7.3% median error against 4.6% for well-covered orders. Moved from
[arch.md](arch.md)

## bridgeConceptMismatch

**A resolved row in the GBIF bridge can be name-correct and concept-wrong**

Status: open
Owner: unassigned
Gate: opportunistic

Where GBIF has not followed a split, its name maps to whichever half kept it: GBIF's `Circus
cyaneus` maps to the Eurasian Hen Harrier while 3.8 million of its records are Northern Harriers.
Nine such cases are listed in [imports.md](requirements/imports.md#gbif); the table has
one AviList column and cannot express them, and nothing detects them automatically. Moved from
[arch.md](arch.md)

## headAspectConstant

**`headAspect` is an unexamined drawing constant**

Status: open
Owner: unassigned
Gate: opportunistic

It produces no metric, so it carries no provenance, but it is as invented as the neck ratio. Moved
from [arch.md](arch.md)

## detailTileRerender

**The detail view re-renders the tile at 1.9×**

Status: open
Owner: unassigned
Gate: opportunistic

rather than sharing a sized component, so a tile layout change has to be made in both places. Moved
from [arch.md](arch.md)

## avonetReadThrice

**The ingest reads the 21 MB AVONET workbook three times**

Status: open
Owner: unassigned
Gate: only if build time starts to matter

, once per taxonomy sheet — about four seconds, which has not been worth optimizing. Moved from
[arch.md](arch.md)

## photographsWithdrawn

**Photographs ([product.speciesPhotographs](requirements/local/product/diagram.md#speciesphotographs)–[product.photoCaching](requirements/local/product/diagram.md#photocaching)) are withdrawn from v1, not abandoned**

Status: open
Owner: unassigned
Gate: not planned

[decisions.photographsWithdrawn](requirements/decisions.md#photographswithdrawn). Any revival goes
through the spec first and stays within
[product.inaturalistLink](requirements/local/product/presentation.md#inaturalistlink) and
[tech.selfContainedServer](requirements/local/tech/delivery.md#selfcontainedserver): BirdView
links to photographs and never embeds, hotlinks or caches one

## broadEbirdRows

**Other species may be drawn from an eBird row broader than their AviList concept**

Status: open
Owner: unassigned
Gate: opportunistic

[decisions.herringGullEuropeanRow](requirements/decisions.md#herringgulleuropeanrow) found European
Herring Gull labeled `measured` from `AVONET2_eBird`'s 2022 *Larus argentatus*, which averages 16
specimens of which 6 are European. The ingest prefers the eBird sheet everywhere, and nothing looks
for other species where AviList's concept is narrower and a BirdLife row matches it better. AVONET's
raw data, which tags each specimen with its BirdLife species and eBird subspecies group, is enough
to find them

## prerewriteClones

**Other clones of `google/birdview` predate the history rewrite of 2026-10-13**

Status: open
Owner: Sameer
Gate: until every other clone is replaced

The stray `birdview-regions` binary and a root-level copy of AVONET's `Supplementary dataset 2.xlsx`
were removed from every commit and `main` was force-pushed, so any other clone or fork must re-clone
or reset to `origin/main`, and an old branch pushed back would restore both. GitHub keeps the old
objects until its own garbage collection, or until Support purges them

## doubtfulWildBirds

**Three birds were kept on the index as doubtful, not as wild**

Status: open
Owner: Sameer
Gate: needs the owner's call

[decisions.escapedBirdsOffIndex](requirements/decisions.md#escapedbirdsoffindex) left West Indian
Whistling-Duck (15 records: Alabama, Texas, Florida, Virginia), Cuban Grassquit (22, all Florida)
and Cinnamon-rumped Seedeater (131, mostly California and Arizona) off `status.Exotic` because their
records may be wild vagrants. Each needs a judgement: add it to the list with a reason, or leave it.
The seedeater may also be a naming question, since AviList's *Sporophila torqueola* is the west
Mexican bird and Morelet's Seedeater is the one native to Texas
