package todo

# Open work

Search `## <id>` for one item, or `grep -n -A1 '^## ' river/todo.md` for the list. `Status:` is
active, open, parked, or deferred; `Gate:` says when it is due. **What is in flight:**
`grep -n -B2 '^Status: active' river/todo.md`. An active item has an `Issue:` line linking the GitHub
issue that tracks it, and the issue links back ([tech.activeWorkHasIssue](requirements/local/tech/process.md#activeworkhasissue)).

Everything outstanding in BirdView: unfinished work, open questions, known defects, criteria not
yet watched failing, and stale spec or doc text. Each item points at the requirement, decision, or
criterion that holds its reasoning.

**Adding an item.** Add it the moment the work is found: a camelCase identifier as the heading,
the title in bold, then `Status:`, `Owner:`, and `Gate:`. Never leave a `TODO` in code or a "left
open" note in a decision record instead.

**Closing an item.** Set `Status: closed`, replace `Gate:` with `Closed:` and what closed it, and
move it to [todo-closed.md](todo-closed.md). Items are never deleted: a closed one records why a
question that looks open is not. Design notes behind items are in [notes.md](notes.md).

## scalableInterface

**The interface is sized for North America by state, not for the world by county**

Status: open
Owner: Sameer
Gate: before worldwide species or county-sized regions; the readers' complaints make it due sooner

The dataset can grow to the whole world (about ten times the species) and to county-sized regions
(a hundred times the regions or more). The interface would not survive either: every region is a
chip listed under its zone, every family a chip under its order, the index is one scroll of every
tile, and the Presentation panel explains everything at once.

The first readers' feedback (2026-10-25) says it is already too much at 1,041 species and about 75
regions. Ryan found it "pretty intimidating with all of that text" and suggested a simpler interface
that just compares stick figures, picked from a list or a search bar; advanced filters behind a menu,
as in the iNaturalist app; and the explanations on their own page, as in the Sibley app. Joe doubted
the average field birder wants "so much info". Karen and Ryan both hit the phone layout
([#2](https://github.com/google/birdview/issues/2), fixed by
[product.fitsAPhone](requirements/local/product/presentation.md#fitsaphone)).

A redesign needs a decision first. The panels already start closed
([product.collapsibleControls](requirements/local/product/presentation.md#collapsiblecontrols)), and the 2026-10-25 prose pass cut the page text; beyond that,
from smallest to largest: move Presentation to the guide and leave a one-line key, or make
search-and-compare the landing view and the full index the advanced one. Region choice at county scale needs search or a map, not chips.
The search box ([product.searchBox](requirements/local/product/filtering.md#searchbox), [#3](https://github.com/google/birdview/issues/3)) is the first
step, done 2026-10-25. Next, decided 2026-10-25: move Presentation to the guide and leave a one-line
key (option (a)); option (b) waits for a discussion of options and for
[simulatedUserTests](#simulatedusertests).
[tightenProse](#tightenprose) is the text half of the same complaint.

Done 2026-10-26, option (b), by the owner's call and without the simulated comparison: the one-set
view built at `/next` became the home page ([product.nextView](requirements/local/product/presentation.md#nextview)), with search-and-pin, one
scale, a Show setting, and each setting on one count line; the index stayed a day as the classic
view at `/classic` and was removed 2026-10-27. Deployed as `82821a6`. Left: region choice at
county scale (search or a map, not chips) and the worldwide dataset;
[homeSpeckScale](todo-closed.md#homespeckscale) followed from the redesign and is closed.

## simulatedUserTests

**Test the interface with simulated users**

Status: open
Owner: Sameer
Gate: before a redesign under [scalableInterface](#scalableinterface) is accepted

Agents given a persona and a task drive the site in headless Chrome, at phone and laptop widths,
and report what they did, where they got stuck, and their answer with a confidence; answers are
scored against expected ones, as in River's concept suite (`usertests/concepts` in River). Personas:
a new birder on a phone, an experienced field birder, an ornithologist, someone outside North
America. Tasks such as: which is bigger, a Cooper's Hawk or an American Crow; which birds are around
you in October; compare three ducks; what does a dashed line mean.

Calibrate before trusting it: run it against `4f25193`, the revision the first readers saw, and see
whether the simulated users report what they did (the phone layout, the density, the missing
search). Then use it to compare designs.

Built and first run 2026-10-25 ([usertests/](../usertests/README.md),
[summary](../usertests/results/2026-10-25/summary.md)): four readers on `4f25193` and on
`b3b2f93`. The old run found the missing search, the density, and the backwards month toggle, but
saw the phone overflow only weakly; mean ease rose from 2.25 to 3.1. Left: run it before and after
Presentation moves to the guide, and decide whether results stay in the repository. Since
2026-10-27 the classic view is gone, so the run is the home page (`/`) against `4f25193` and
`b3b2f93`; the readers' tasks will need new instructions for pins and the Show setting.

## testersAskedBack

**Ask the first readers back to the new home page**

Status: open
Owner: Sameer
Gate: owner's call; [homeSpeckScale](todo-closed.md#homespeckscale) is settled

Ryan, Karen, and Joe gave the first feedback (2026-10-25), and Shai Mitra is also to be asked; none
has seen the home page that answers it. They were deliberately not contacted while the redesign
was in progress. Ask them to try birdview.site on a phone and a laptop, with the same questions as
before, and record what they say under [scalableInterface](#scalableinterface).

## homePageReaderFeedback

**A reader's suggestions for the home page, 2026-10-30**

Status: open
Owner: Sameer
Gate: owner's call on each; requirement first

Feedback from a reader of the home page ([product.nextView](requirements/local/product/presentation.md#nextview)), relayed by the owner:

1. **Month:** say that several months can be chosen ("click as many as you want").
2. **Tile layout:** move length, wingspan, and mass up beside or below the name, above the stick
   figure, since they are what the page is for; move the banding codes and Latin name below the
   line. Set the measurements in a face designed for small text ("old eyes"). **Layout done
   2026-10-30, revised 2026-10-31** ([decisions.tileNumbersFirst](requirements/decisions.md#tilenumbersfirst),
   [decisions.tileMeasurementLines](requirements/decisions.md#tilemeasurementlines)). **Type
   enlarged 2026-11-01**: nothing below 13px, measurements in the system sans-serif with even-width
   figures ([decisions.textSizeFloor](requirements/decisions.md#textsizefloor)). Left: whether a bundled legibility face such as Atkinson
   Hyperlegible is worth an `imports.md` entry, and whether the measurements, still the dim gray,
   should be full-contrast ink.
3. **Pins:** keep the pinned bird at the top of the results as a reference, whatever the Show
   setting. **Done 2026-11-01** as the Pinned group ([decisions.pinnedGroupLeads](requirements/decisions.md#pinnedgroupleads)).
4. **Mass cube:** at first it looked like a button for a 3-D view, and among similar birds (gulls)
   the cubes were all about the same size, so their meaning did not come through. Show the hover
   text sooner, or draw a vessel of water (a fish tank, a water-cooler bottle) so the volume reads
   as water ([decisions.massAsWaterCube](requirements/decisions.md#massaswatercube); compare
   [massAsLine](#massasline)).
5. **iNaturalist link on every tile** (the owner's addition), not only on the detail view
   ([product.inaturalistLink](requirements/local/product/presentation.md#inaturalistlink)). The
   tile is one link to the detail view, so this one must be a sibling, like the pin. **Not with
   iNaturalist's logo:** its Terms of Use name "the iNaturalist Bird Logo Design" as a trademark
   and grant "no right or license to reproduce" it, and
   [inatterms](requirements/imported/inaturalist.org/inatterms/inatterms.md) adopted iNaturalist on
   the finding that BirdView names it in words rather than reproducing a logo. Use a word or a
   glyph of BirdView's own, unless iNaturalist grants permission in writing. **Done 2026-10-30**,
   as the word **iNat**, in the tile's lower-right corner since 2026-10-31. Replaced 2026-11-01 by
   an **eBird** link ([decisions.ebirdLinkReplacesInaturalist](requirements/decisions.md#ebirdlinkreplacesinaturalist)).

The reader also asked what "1 in 6,322 species observations" in New York means, so the line does
not explain itself ([product.shareOfRecords](requirements/local/product/presentation.md#shareofrecords)).

## massUnitConsistent

**A bird's mass changes unit between pages**

Status: open
Owner: unassigned
Gate: opportunistic

The unit follows the birds on screen, so the Mallard is 29.8 oz on the home page and 1.9 lb when
pinned or on its own page. Four of eight readers on 2026-11-01 tripped on it. Decide whether one
bird should read the same everywhere ([product.imperialByDefault](requirements/local/product/presentation.md#imperialbydefault)).

## harnessOpensSettings

**The simulated readers cannot open the home page's settings**

Status: open
Owner: unassigned
Gate: before the next reader run

`usertests/tools/page --open` knows only the classic view's panel names and `shot` cannot click, so
no reader on 2026-11-01 could open Month, Region, Units, or Names, and each reported the settings
broken. Teach `page` to open a `<details>` setting by its name, and say so in `prompt.md`, which
also still tells readers the find box sends `q=` rather than `find=`.

## ebirdTermsUnread

**eBird's terms of use are not vendored, and the eBird link was adopted without reading them**

Status: open
Owner: Sameer
Gate: before any use of eBird beyond a plain link

[product.ebirdLink](requirements/local/product/presentation.md#ebirdlink) links to eBird species pages and copies nothing, which
needs no license. But iNaturalist's adoption read and transcribed its terms
([inatterms](requirements/imported/inaturalist.org/inatterms/inatterms.md)), and eBird's never
were. Vendor eBird's terms of use under `third_party/`, transcribe what binds a site that links to
it, and record the finding as `inatterms` did. Until then, anything more than a link (a logo, an
embedded map or photograph, a lookup against eBird's API) waits.

## withdrawnWithClassic

**One behavior went with the classic view; the owner may want it back**

Status: open
Owner: Sameer
Gate: owner's call

Removed 2026-10-27 under [decisions.classicViewRemoved](requirements/decisions.md#classicviewremoved):
the "Common birds" region grouping ([decisions.regionsTwoWays](requirements/decisions.md#regionstwoways))
no longer appears anywhere. Restore it on request. The other, the Here-and-now link using the
reader's month rather than the server's, was restored the same day under
[decisions.homeShowsEveryMonth](requirements/decisions.md#homeshowseverymonth). `internal/view`'s `ScaleMode` and `Scales` are left
over from the classic scale control and nothing calls them; delete them, since
[homeSpeckScale](todo-closed.md#homespeckscale) closed without bringing back a scale choice.

## cleanupLeftovers

**Two criteria from the classic-view cleanup need a second look**

Status: open
Owner: unassigned
Gate: opportunistic

Found by the 2026-10-27 cleanup. The honestRevalidation criterion says every response carries an
ETag, which is untrue for `/?month=now` (bare `/` has carried one since
[decisions.homeShowsEveryMonth](requirements/decisions.md#homeshowseverymonth)); correct the wording. Some Habitat and Diet
criteria are now decided only through [app.nextViewDrawsOneSet](requirements/composed/internal/app/acceptance.md#nextviewdrawsoneset); check that each still has a test
that fails when its behavior breaks.

## tightenProse

**Tighten all prose and page text; less is more**

Status: open
Owner: Sameer
Gate: when convenient

Second pass 2026-10-25, on what readers see: the page templates, the text built in Go
(`main.go` notes, verdicts, tooltips, `/added` and `/excluded`), the region grouping notes, and
`docs/using.md`. Design reasoning left the pages for river/; two requirements were amended to
allow it ([product.labeledDetailDiagram](requirements/local/product/measurements.md#labeleddetaildiagram)'s cube and folded-neck notes). First pass done 2026-10-20: README, AGENTS.md, CONTRIBUTING, CHANGELOG, the four guides, every
product and tech requirement, `app.md`, and the page templates, about 30% shorter overall. Left:
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

## craniumAtLargeMass

**The cranium model overdraws the largest birds, and one is now in scope**

Status: open
Owner: Sameer
Gate: when convenient

[product.craniumFromMass](requirements/local/product/diagram.md#craniumfrommass) said its
overestimate at large mass cost nothing because the heaviest bird in scope was 11 kg. Emu is now in
scope at 34 kg and draws a 243 mm cranium. Check it against a measured emu skull, and if it is far
off, fit or cap the coefficient.

## sexSpecificDrawings

**Separate drawings for males and females are not supported by the data; investigated 2026-10-18 and deferred**

Status: open
Owner: Sameer
Gate: when a usable sex-specific mass source is found

Mass is the blocker: it is pooled per species in every vendored source, and head, body, and the mass
cube are modeled from it
([product.craniumFromMass](requirements/local/product/diagram.md#craniumfrommass),
[product.trunkFromMass](requirements/local/product/diagram.md#trunkfrommass)), so per-sex bill,
wing, and tail on a shared body would draw a bird nobody measured
([product.tileMarksModeled](requirements/local/product/measurements.md#tilemarksmodeled)). What
exists: AVONET's unused `AVONET_Raw_Data` sheet sexes each specimen for bill, wing chord, secondary,
tail, and tarsus, but only **213 of 1,082** drawn species have five or more of each sex, and **9 of
61** drawn raptors; Fox's sex-specific wingspan columns are filled for **2** drawn species. Where
samples suffice, raptor females run larger in wing chord (Sharp-shinned Hawk +15%, Peregrine +9%)
and males larger in grackles, cowbirds, Ruff, and pheasant (+11–22%); a 5–9% gap in a passerine from
five specimens per sex is likely noise. **Leads for sex-specific mass, none checked:** Myhrvold et
al. 2015's amniote database (believed public domain, with a figshare copy; license and coverage
unverified). Lislevand et al. 2007 and Ecology's data papers generally sit on **Wiley, which
BirdView may not read**
([decisions.ricklefsProhibited](requirements/decisions.md#ricklefsprohibited)); check where any
candidate is hosted before touching it. Dunning's handbook and Birds of the World are closed. A
cheaper step that needs no new source: per-sex wing chord and tail as numbers, with specimen counts,
on the detail page

## nonTextContrastCheck

**WCAG 1.4.11 has no check**

Status: open
Owner: Sameer
Gate: before any accessibility claim

[app.pagesMeetNonTextContrast](requirements/composed/internal/app/app.md#pagesmeetnontextcontrast) is uncovered: the
stylesheet checks skip shapes on purpose and the rendered check measures text only. Extending the
rendered check to strokes and fills against adjacent colors is the obvious route; "required to
understand the content" needs a ruling per diagram part

## composeDatasetAndRepo

**Compose `dataset` and `internal/repo`**

Status: open
Owner: Sameer
Gate: ready; `app` was approved 2026-10-25

, which must account for the 24 local requirements no package lists yet (`TestCompositionResolves`
logs them), then move each criterion out of [acceptance.md](requirements/acceptance.md) into its
package with `Covers:` and delete the Traceability table once nothing is left in it. **Four wait on
the morphometric-model ruling** — [product.craniumFromMass](requirements/local/product/diagram.md#craniumfrommass), [product.trunkFromMass](requirements/local/product/diagram.md#trunkfrommass),
[product.neckModelPerFamily](requirements/local/product/diagram.md#neckmodelperfamily), and [product.generatedBadgeColors](requirements/local/product/dataset.md#generatedbadgecolors): each produces a value the app draws,
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

**Statement-first: one requirement still leads with a license condition**

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
[tech.badgeContrast](requirements/local/tech/accessibility.md#badgecontrast), and
[tech.textContrastFloor](requirements/local/tech/accessibility.md#textcontrastfloor) (WCAG 2.2 SC
1.4.6 and 1.4.11). These belong in a composed package citing the source, not in `tech`. **Mixed — a
local requirement with an imported obligation inside it, to split:**
[tech.fetchRestrictedInputs](requirements/local/tech/delivery.md#fetchrestrictedinputs) (*do not
redistribute* is the licenses'; *fetch at build time* is BirdView's mechanism),
[product.selfDescriptionStaysTrue](requirements/local/product/presentation.md#selfdescriptionstaystrue)
(*credit every CC BY source* is a license condition; docs currency is local),
and [tech.localNowPublicLater](requirements/local/tech/posture.md#localnowpubliclater) (*never
commercial* follows the non-commercial licenses). **The other 73 are local intent**, including
those that name a source — choosing AviList for names is BirdView's decision, made with licenses in
mind, not an obligation AviList imposes. Two findings on the way: the WCAG manifest entry still says
other text is held to 4.5:1, which
[decisions.oneContrastFloor](requirements/decisions.md#onecontrastfloor) replaced with one 7:1
floor; and Wikimedia and Macaulay Library are cited as rationale from the photograph era, with no
code reading either. **Progress 2026-10-16:** WCAG is now vendored and transcribed (`wcag`), the
stale manifest entry is fixed, and the two contrast requirements serve as the guarantee half of
[app.stylesheetContract](requirements/composed/internal/app/app.md#stylesheetcontract), with their WCAG obligation
composed separately as [app.pagesMeetTextContrast](requirements/composed/internal/app/app.md#pagesmeettextcontrast).
The three mixed requirements are still to split. **2026-11-01:** three of the original list
went with iNaturalist ([decisions.inaturalistSourcesRetired](requirements/decisions.md#inaturalistsourcesretired)): [tech.pacedInaturalistMapping](requirements/local/tech/process.md#pacedinaturalistmapping) and
[product.inaturalistLink](requirements/local/product/presentation.md#inaturalistlink) are withdrawn, and [tech.identifiedClient](requirements/local/tech/delivery.md#identifiedclient) no longer restates an
imported source

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
**+11.9%**, `Anser cygnoides` **+20.0%**, and `Anser anser` **+6.9%** — the same signature as the
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
still misses Wood Stork/NS, Spoonbill/ME, Fork-tailed Flycatcher/DE, and Baltimore Oriole/NT.
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
602-checklist probe already separated waterbirds, woodland songbirds, open country, feeder birds,
and **town birds** (pigeons, House Sparrows, starlings), leaving hawks and crows unattached because
they occur with everything. Expect the rare pelagic and Hawaiian families to be unplaceable, as
the atolls were

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
identity columns), `measurements.csv` (one row per species per metric, with its kind, source, and
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
Gate: opportunistic; [acceptance.tileMarksNonMeasured](requirements/acceptance.md#tilemarksnonmeasured) and [acceptance.nothingRestrictedTracked](requirements/acceptance.md#nothingrestrictedtracked) guard hard rules

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
(outbound requests identify the client), and
[acceptance.abaFileHashMatches](requirements/acceptance.md#abafilehashmatches) (the fetched ABA file
matches its hash). [acceptance.joinCoversEverySpecies](requirements/acceptance.md#joincoverseveryspecies) and [acceptance.proportionsDifferentiate](requirements/acceptance.md#proportionsdifferentiate) are
already cited by tests, so they may only need the mutation run and a status;
[acceptance.abaFileHashMatches](requirements/acceptance.md#abafilehashmatches) is probably obsolete since
[decisions.scopeFromRecurrence](requirements/decisions.md#scopefromrecurrence) removed the ABA
input, and wants withdrawing rather than verifying

## riverReleaseFindings

**What adopting River's Release stage taught, to send upstream**

Status: open
Owner: Sameer
Gate: next River session

Found 2026-10-22 and 23, taking birdview.site public and adopting Release. Each is a place River's
README was silent or assumed something BirdView did not have:

- **A decision recorded after it was made.** [decisions.publicPersonalDeployment](requirements/decisions.md#publicpersonaldeployment) and
  [decisions.regionFilterReinstated](requirements/decisions.md#regionfilterreinstated) needed a `Recorded:` field River's entry format does not define.
- **Guardrails in a private repository.** River says the bundle cites guardrails from a mutable
  record. BirdView's health probe and uptime check live in `Sajmani/birdview-deploy`, which is
  private, so the citation names something most readers cannot open. River should say whether
  that is acceptable, and what the bundle must then state itself.
- **Code owner and release owner differ.** The code is this organization's; the deployment is
  Sameer's own. River's "Who decides" assumes one owner over every stage.
- **Spec checks need version control.** River's checks read what the VCS tracks, so they cannot run
  in a build context. BirdView runs them in CI and the rest in the image build; River could name
  that split.
- **Adopting a stage that was informal.** River says to declare an unadopted stage informal, but
  not what triggers a revisit or what adopting it requires. BirdView's trigger was going public;
  adoption meant requirements, criteria, and the deployment's own settings.
- **A one-way mirror as the deployed copy.** Deploying from a fork that is synced, never committed
  to, kept one source of truth; River's Release stage might name it as a pattern

## inatPacingFlake

**The GBIF client's pacing test can fail under load, as the iNaturalist one did**

Status: open
Owner: unassigned
Gate: if CI goes red on it

Found 2026-10-23. `TestLookupHonorsTheSourceRequirements` measured two requests 114 ms apart against
a 120 ms interval and a 5 ms tolerance, during a full `go test ./...`; alone it passed five runs of
five. The gaps are timed at the fake server, so a first request delayed in arriving shortens the
next gap without the client pacing any faster. The GBIF client's test has the same shape and
flaked the same way on 2026-10-26: `TestTaxonIdentifiesItselfAndPaces` measured 53.8 ms against a
60 ms interval in a full run, and passed three runs alone. Now that
[tech.checksOnEveryChange](requirements/local/tech/release.md#checksoneverychange) fails CI on any failure, a flake here would read as a broken
change. Measure the gap at the client's send instead, or widen the tolerance with that reason.
By 2026-10-27 both flaked in most local full runs, though not yet in CI. The iNaturalist test was
deleted 2026-11-01 with the client it tested ([decisions.inaturalistSourcesRetired](requirements/decisions.md#inaturalistsourcesretired)); the GBIF one remains.

## attributionCriterionNamesAba

**[acceptance.attributionComplete](requirements/acceptance.md#attributioncomplete) still asserts an ABA credit**

Status: open
Owner: unassigned
Gate: opportunistic

Found 2026-10-23 during step 4 of [deploymentPlumbing](#deploymentplumbing). Its assertions say "the ABA checklist is credited
for the species list and *only* that", but
[decisions.scopeFromRecurrence](requirements/decisions.md#scopefromrecurrence) removed ABA on
2026-10-05 and the footer credits GBIF for scope. Check what `TestAttributionIsComplete` actually
asserts, then correct whichever of the two is wrong, and add GBIF to the named credits

## nearMatchSplits

**Four near-matches are splits awaiting a human taxonomic check**

Status: open
Owner: Sameer
Gate: needs the owner's call

[decisions.abaSpeciesWithoutAvonet](requirements/decisions.md#abaspecieswithoutavonet), *Still
unresolved*. *Astur atricapillus*, *Ardea ibis*, *Gygis candida*, and *Tyto furcata* borrow from
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
borealis*/*diomedea*, and *Pterodroma deserta*/*feae* each share one mass, so their linear traits
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
   deletion and force-pushes; issues, pull requests, wiki, and projects are off.
2. ~~**Deploy to Cloud Run**~~, decided 2026-10-22, from a Google Cloud project billed to Sameer.
   **Done 2026-10-22:** project `birdview-site`, region `us-central1`, service `birdview`,
   revision `birdview-00001-hfm` at `f1fa369`, serving <https://birdview-570018079643.us-central1.run.app>.
   Deployed by hand from a clone of the mirror (`~/src/birdview-site`, SSH remote) with:

   ```
   gcloud run deploy birdview --source . --region us-central1 --allow-unauthenticated \
     --args=-addr,0.0.0.0:8080,-data,/data/species.json,-feedback,https://github.com/google/birdview/issues
   ```

   Pass every flag: `--args` replaces the image's `CMD` entirely, so `--args=-feedback,<url>` alone
   would leave the server on `localhost:8080`, unreachable. The server reads `-addr`, never `PORT`,
   so leave Cloud Run's port at 8080. `us-central1` was chosen with step 7 in mind; confirm it
   supports domain mappings before relying on one. Checked on the live service: every page
   answers, an unknown species 404s, the index carries the same 1,041 species links as a local
   build and arrives gzipped (277 KB) with `Vary`, `Cache-Control: no-cache`, and an ETag that
   revalidates to 304, the boundary file is served `immutable`, HTTP redirects to HTTPS, no
   cookie is set, no script, stylesheet, or image on the index loads from another host, and the
   feedback link names `google/birdview` issues.
3. ~~**Build the dataset for deployment.**~~ Done: the `Dockerfile`'s first stage runs the ingest
   and copies `species.json` into the image, so no deploy step supplies it
   ([tech.datasetNeverCommitted](requirements/local/tech/delivery.md#datasetnevercommitted)).
4. ~~**Amend the requirements for public serving**~~, requirement first:
   [tech.localNowPublicLater](requirements/local/tech/posture.md#localnowpubliclater) (its value is `local`) and
   [tech.servesOnLocalhost](requirements/local/tech/delivery.md#servesonlocalhost). The stable surface is done:
   [tech.stableUrls](requirements/local/tech/delivery.md#stableurls), 2026-10-23. **Done 2026-10-23:** both amended under
   [decisions.publicPersonalDeployment](requirements/decisions.md#publicpersonaldeployment), which records the 2026-10-22 decision to go public; the
   stage-adoption row of `bindings.md` no longer says BirdView is undeployed.
5. ~~**Write the first release criteria**~~, BirdView's first use of River's Release stage: a health
   check, monitoring and alerts, a rollback path, and what happens when an ingest input changes.
   **Done 2026-10-23** (`7fcea0b`): [tech/release.md](requirements/local/tech/release.md) — CI on every change, a Dockerfile test stage
   gating every image, and `birdview-smoke` after every deploy; rollback documented in
   `docs/building.md`; Release adopted in `bindings.md`. The deployment's own settings are in
   place since 2026-10-23, recorded in the private `Sajmani/birdview-deploy`: a startup probe on
   `/robots.txt`, an uptime check on a species page from three regions every 15 minutes, and an
   email alert when it fails.
6. **Operational checks:** TLS, caching behavior
   ([decisions.hereAndNowCaching](requirements/decisions.md#hereandnowcaching)), no tracking, a feedback channel that exposes
   no personal address, basic rate limiting. If step 7 puts Firebase Hosting in front, check
   through it too: its CDN caches by `Cache-Control`, so confirm pages still revalidate
   ([tech.cacheLifetimeStated](requirements/local/tech/delivery.md#cachelifetimestated)) and arrive gzipped with `Vary` intact
   ([tech.compressedResponses](requirements/local/tech/delivery.md#compressedresponses)).
7. ~~**Point `birdview.site` at the host.**~~ Three ways onto Cloud Run with TLS: a Cloud Run domain
   mapping (simplest, but preview and in some regions only); Firebase Hosting rewriting to the
   service (free, certificate included); or a global external load balancer (about $18 a month).
   **Done 2026-10-23 with Cloud Run domain mappings** in `us-central1`, which supports them:
   `birdview.site` (four A and four AAAA records at Squarespace, replacing the "Squarespace
   Defaults" preset, including its `HTTPS` record) and `www.birdview.site` (CNAME to
   `ghs.googlehosted.com`). Ownership verified in Search Console by the `google-site-verification`
   TXT on `@`, which must stay. Certificates from Google Trust Services; `www` was issued within
   23 minutes, `birdview.site` after 1 h 41 m, reporting "challenge data was not visible" until
   then. The likeliest cause: its mapping was created before the DNS change, while the old
   records (4-hour TTL) were still published. **Next time, change DNS first and map second.**
   Checked on both hosts on 2026-10-23: the same results as the `run.app` checks in step 2, and
   HTTP redirects to HTTPS on each. The service is capped at 2 instances
   (`--max-instances=2`) and the project has a $10 budget alert. ~~`www` serves the site itself~~:
   since 2026-10-23 the server takes `-canonical` ([tech.oneCanonicalHost](requirements/local/tech/delivery.md#onecanonicalhost)), so deploying with
   `--args=-addr,0.0.0.0:8080,-data,/data/species.json,-feedback,https://github.com/google/birdview/issues,-canonical,birdview.site`
   redirects `www` and the Cloud Run address to `birdview.site`; deployed 2026-10-23 as revision
   `birdview-00004-z2b`, and checked: both answer 301 with path and query intact, plain HTTP takes
   one more hop through Google's own HTTPS redirect, and `birdview.site` serves as before. ~~And there is no `robots.txt`~~: added 2026-10-23 under
   [tech.crawlersSkipFilteredViews](requirements/local/tech/delivery.md#crawlersskipfilteredviews), asking crawlers to skip every URL with a query string.
8. **Sync and deploy automatically.** A push to the mirror's `main` fires a Cloud Build trigger,
   connected through the Cloud Build GitHub app installed on the `Sajmani` account, that builds
   the image and deploys it. The sync that feeds it should also run from Google Cloud: Cloud
   Scheduler starts a small Cloud Build job, hourly say, that runs `gh repo sync` with a
   fine-grained token for the mirror alone, kept in Secret Manager. Nothing is committed to
   either repository. A scheduled GitHub Action was considered and is worse here: its workflow
   file would have to be committed to the mirror, which breaks the one-way rule, or to this
   repository, which puts personal deployment plumbing in this organization's repository; and
   GitHub turns scheduled workflows off in forks until enabled. **Grant the token Contents and
   Workflows, read and write, on the mirror alone**: without Workflows, the sync fails the day
   this repository adds a file under `.github/workflows/`. If [sourceOfTruth](#sourceoftruth) moves
   development to the mirror, the sync job goes and the deploy trigger stays.

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
links to photographs and never embeds, hotlinks, or caches one

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
Whistling-Duck (15 records: Alabama, Texas, Florida, Virginia), Cuban Grassquit (22, all Florida),
and Cinnamon-rumped Seedeater (131, mostly California and Arizona) off `status.Exotic` because their
records may be wild vagrants. Each needs a judgment: add it to the list with a reason, or leave it.
The seedeater may also be a naming question, since AviList's *Sporophila torqueola* is the west
Mexican bird and Morelet's Seedeater is the one native to Texas

## countyRegions

**Regions down to counties, coded as the data codes them**

Status: active
Issue: [#11](https://github.com/google/birdview/issues/11)
Owner: Sameer
Gate: ships to birdview.site, with -counties turned on, only when every US state and Canadian
province and territory has counties ([decisions.countiesBehindFlag](requirements/decisions.md#countiesbehindflag)); work in progress is pushed

[decisions.regionCodesFromData](requirements/decisions.md#regioncodesfromdata) names regions by GADM code on land and Marine Regions id at sea, and
redirects every value BirdView wrote before. Step 1, land regions by GADM code with postal codes
and island names redirecting, was done 2026-11-03. Left, in order:

2. *Done 2026-11-03:* both downloads recorded in `data/gbif-downloads.tsv` (New York counties,
   10.15468/dl.ttkwue; uncoded cells, 10.15468/dl.4jngjn). The uncoded cells turned out to be mostly
   inland and near-shore water, not open sea: lakes, rivers, and bays GADM leaves out, with Lake
   Saint-Louis at Montreal the busiest cell. Step 6 has to separate inland, coastal, and open water.
   Original plan: **a GBIF download** in place of per-region queries, which county scale would multiply about
   fortyfold and which is kinder to GBIF: species by county by month for New York, and species by
   grid cell by month for records with no GADM code, offshore. Needs a GBIF account and SQL download
   access; each download has a DOI to cite, which changes [imports.gbif](requirements/imports.md#gbif) and the credits.
3. *Done 2026-11-03 (eBird only):* 63 units, 112 k to 4.9 M eBird records each; at today's rule a
   county lists 235 to 410 species, 9% of listings on fewer than 5 records; 50 species reach a county
   list but not the state's, cage birds and vagrants; NYC's five counties together list 428 of the
   state's 450, so in dense counties the share ordering matters more than the set; ~40 MB of
   presence for the US by county. Details in [notes: The New York county spike](notes.md#the-new-york-county-spike).
   **Answered 2026-11-03** ([decisions.countyListingRule](requirements/decisions.md#countylistingrule)): a floor of 5 records over all years
   with the state list required, counties narrowing and ordering like states, and per-state tables
   for the server; alternatives to the floor parked in [countyFloorRevisit](#countyfloorrevisit). The questions as asked:
   - *Minimum record count:* a fixed floor (5 or 10 records) or one per year, since records
     accumulate over decades.
   - *Vagrants:* whether a county listing also requires the bird to be on its state's list, or to
     recur as [product.scopeIsRecurringAvifauna](requirements/local/product/dataset.md#scopeisrecurringavifauna) requires, rather than only to reach a count.
   - *What a county view is for:* narrowing the set, or chiefly ordering it by local share; it
     shapes the picker more than the bar does.
   - *Storage and delivery:* about 40 MB for the US, which the server cannot embed as it does
     today: per-state files loaded on demand, or a compact binary table.
   - *Loose ends:* the download's eBird total is 2.6% under October's state table, unexplained; and
     GADM's Lake Ontario unit takes in shoreline (American Robin is its sixth-commonest bird).
   Original plan: a New York spike on that data: records per county, file size, and a minimum record count for
   small counties, where one record per million is less than one record.
4. *Prototype done 2026-11-03:* New York's county table (`data/county-presence/USA.33_1.tsv`, built
   by `birdview-regions -counties`), joined by the ingest and served as regions; a **County**
   setting follows Region when New York or its counties are chosen. Left: searching counties by name,
   since the setting lists 63 alphabetically; Here to a county, which needs step 5; how the setting
   scales to Texas's 254; and the remaining states' downloads and tables, and Canada's census
   divisions (the owner's call, 2026-11-03). Served only with `-counties` until then
   ([decisions.countiesBehindFlag](requirements/decisions.md#countiesbehindflag)).
   Original plan: a county picker prototype (drill-down, search, Here to a county) on the New York data, sized so
   no link lists many county codes.
5. Boundaries keyed to GADM codes for Here at county scale, since
   [tech.regionBoundaries](requirements/local/tech/delivery.md#regionboundaries) uses Natural Earth.
6. Seas: adopt [imports.marine-regions](requirements/imports.md#marine-regions), assign the download's offshore grid cells to its
   areas while building the data, replace the five sea boxes, and redirect their names. Asking VLIZ
   about serving its boundaries gates only Here at sea, the one use that sends them to readers.

## countyFloorRevisit

**Revisit the county floor if county lists grow with eBird**

Status: parked
Owner: Sameer
Gate: when a regenerated county table lists noticeably more species than the one before it, or
before counties outside the US and Canada

[decisions.countyListingRule](requirements/decisions.md#countylistingrule) lists a bird in a county at 5 records over all years. A fixed count gets
easier to reach as eBird grows (New York added 6.6 M records in 2025 alone), though the requirement
that the state also list the bird bounds what can get through. Alternatives the owner considered,
each needing the year in the GBIF download (a `"year"` column in `data/gbif-queries/ny-counties.sql`):

1. **A recent window:** count only the last 10 years, so old records age out and the floor means
   the same thing each year.
2. **Distinct years:** require records in at least N different years, as
   [product.scopeIsRecurringAvifauna](requirements/local/product/dataset.md#scopeisrecurringavifauna) does for scope, which separates a resident from a bird that visited once and
   drew a crowd.
3. **A relative floor:** a share of the county's records larger than one per million, so it scales
   with effort as the state bar does; the spike's table in
   [notes: The New York county spike](notes.md#the-new-york-county-spike) gives the counts it would trade against.

## traitAudit

**Audit every bird's habitat and diet, so the filters say what a birder expects**

Status: active
Issue: [#12](https://github.com/google/birdview/issues/12)
Owner: Sameer
Gate: before the Habitat and Diet filters are relied on further, and again before worldwide species

Habitat and Diet come from AVONET, one primary value per species, with BirdView's additions for
nine birds ([product.habitatAndDietFilters](requirements/local/product/months.md#habitatanddietfilters)). A reader found Least Sandpiper under Marsh and lake and
missing, with the other peeps, from Shore and estuary. The owner's judgment: without intuitively
correct values these filters are not only useless but misleading. Findings so far, for habitat, in
[notes: The habitat audit, pass 1](notes.md#the-habitat-audit-pass-1):

1. AVONET's single habitat is in practice the breeding habitat for migrants: Black-bellied Plover,
   Semipalmated and Purple Sandpiper are Grassland (tundra), the scoters and every loon and grebe
   Wetland. AVONET's own `Migration` column is 3 for all of them, a useful hint but not a rule:
   residents can span several habitats and diets too.
2. It is range-wide, not North American (Black Skimmer Riverine), and sometimes contradicts its own
   definitions (Clapper Rail Wetland, though AVONET puts salt marsh under Coastal).
3. The occurrence data cannot settle it: records with no GADM code do not measure time on water.

Left:

- **Habitat:** reviewed species by species in `data/habitat-review.tsv`, which the build holds
  complete ([decisions.habitatReviewedPerSpecies](requirements/decisions.md#habitatreviewedperspecies)): 9 reviewed (the former `trait.Supplementary`
  entries) and 60 more (pass 1's water birds, approved by the owner on 2026-11-07, with Shore and
  estuary added to Killdeer), so 69 reviewed and 1,013 pending. Next: the other families, landbirds
  mostly; raptors, swallows, and blackbirds first.
- **Diet:** the same audit for `Trophic.Niche`, which has the same one-value limit (a bird that eats
  insects in summer and seeds in winter has one niche); `mixedDiet` covers only AVONET's own
  Omnivore-level disagreements.
- **The label:** whether the filters should say they choose a bird's main habitat or diet.
- **Sources:** no openly licensed per-species, multi-habitat source is known; IUCN's habitats come
  through BirdLife, whose terms BirdView cannot accept. Additions are BirdView's own, reviewed by the
  owner (and perhaps Dennis and Bobbi).
- **Worldwide species:** repeat the audit for every species added.

## anchorBirds

**An anchor bird, chosen by the reader, to measure other birds against**

Status: parked
Owner: Sameer
Gate: after [traitAudit](#traitaudit), and once a group such as "all shorebirds" can be chosen in one step

From issue #8 and a conversation with Dennis Hrehowsik, who with Bobbi Manian teaches gull
identification by comparison with an anchor, a first-year American Herring Gull ("The Brooklyn Gull
Guide", *The Clapper Rail*, Spring 2025, pp. 8–11): "anchor bird means suggesting an easily
identifiable species as a bench mark for size comparison. Like how bb plover is the anchor or middle
sized shore bird." Explored 2026-11-07 (the review artifacts `pin-twins-spec.md`,
`cross-family-twins-10.md`):

- **Twins**, birds within 10% in length, wingspan, bill, and tail, are mostly a bird's relatives,
  or across families trivia in which a birder knows one bird of the pair. Dropped.
- **Grouping larger, same, smaller** than an anchor adds little to the tiers.
- **[product.reportingFrequencyFilter](requirements/local/product/filtering.md#reportingfrequencyfilter)**, built from this, does most of the work: familiar birds sorted by size line
  up across families.

Left for an anchor: a pin-like control with a ship's-anchor icon marking one bird, and every tile
stating its length and wingspan as a share of the anchor's. Two things stand in the way. Habitat
cannot yet choose the shorebirds ([traitAudit](#traitaudit): Black-bellied Plover is Grassland). And Family offers
shorebirds only family by family, under Charadriiformes ("Shorebirds, Gulls, and Auks").

## placeAsHabitat

**Let the place stand in for habitat: birds recorded near here, this month**

Status: parked
Owner: Sameer
Gate: after counties ship ([countyRegions](#countyregions))

Proposed 2026-11-07 as the longer-term answer to "which birds am I likely to see here", which
eBird's hotspots answer by place rather than by habitat (though Plumb Beach is beach, salt marsh,
and bay at once). GBIF records carry coordinates, so a SQL download grouped by small grid cells (a
kilometer or two) and month would give the birds recorded near a point, and a beach cell and a park
cell would differ with no habitat classification at all. Open: cell size; the minimum record count,
as for counties ([decisions.countyListingRule](requirements/decisions.md#countylistingrule)); keeping Here private, since the reader's location stays in
the browser ([product.locationStaysInBrowser](requirements/local/product/filtering.md#locationstaysinbrowser)); and the data's size.


## photoProportions

**Check each species' drawn proportions against photographs, as a pipeline of its own**

Status: withdrawn 2026-11-09 ([decisions.photoReviewUnreliable](requirements/decisions.md#photoreviewunreliable))
Owner: Sameer
Gate: none; revive only with a source of photographs taken from directly below, wings spread

**Withdrawn.** A local review tool put each genus's figure beside Commons photographs with an overlay
and sliders (2026-11-09). Trying it, the owner found the evidence cannot be had: the figure is a plan
view, body and spread wings in one plane, and almost no photograph shows a bird that way. Few are
exactly broadside, and only a bird photographed from directly below with wings spread shows the
wingspan undistorted, which is too rare to cover many species. Any other angle foreshortens some part,
so values set by matching photographs, by hand or by a pipeline, would measure the camera's angle.
The cost estimate on issue #7 assumed usable photographs and no longer applies. What follows is the
record as it stood.

[GitHub issue #7](https://github.com/google/birdview/issues/7): the Lesser Scaup's figure had the
longer neck, and the Horned Grebe's the shorter neck and the larger head (0.18 of its length, from
the mass model), the opposite of photographs of both in flight. The cause was family-wide models:
`NeckRatios` gives Anatidae 0.70, set for swans and geese. **The quick fix is done** (2026-11-08,
[decisions.duckAndGrebeProportions](requirements/decisions.md#duckandgrebeproportions)): genus ratios for the diving ducks and grebes and a head factor
for grebes, set by judgment. Still on the family's 0.70: the dabbling ducks, mergansers, geese, and
swans, whose necks all draw about a quarter of their length, so a Canada Goose's is no longer than
a Mallard's ([reviewNeckRatios](#reviewneckratios)). Cost estimate for the per-species version:
on the issue, 2026-11-08.

The ideal is per-species: measure neck, head, bill, and tail as shares of total length from photos
of each bird in flight, and record them as a BirdView dataset with its method and sources, marked
measured-from-photo. That is a pipeline: finding usable side-on flight photos, measuring them
(by hand, or by a model with a reviewed sample), and recording the results with provenance. The
owner's view (2026-11-08): measuring a photo, which copies nothing of it, seems acceptable, but
fair use of the photos found has to be settled first. BirdView copies nothing from eBird or
iNaturalist, least of all photographs ([product.ebirdLink](requirements/local/product/presentation.md#ebirdlink)), and automated access must respect
each site's terms; openly licensed photos (Wikimedia Commons, or GBIF's media under CC0 and CC BY)
are the likelier sources. Cost scales with species: about 1,082 now, about 11,000 worldwide.

## habitatDataset

**Research each species' usual habitats as a BirdView dataset of its own**

Status: active
Issue: [#10](https://github.com/google/birdview/issues/10)
Owner: Sameer
Gate: a decision on the classification and on which sources may be used; before worldwide species

[GitHub issue #10](https://github.com/google/birdview/issues/10): AVONET's Coastal and Marine do not
separate three groups a birder does: birds of the open ocean seen from boats (shearwaters, petrels),
birds of the sea that are seen from shore (sea ducks, loons), and birds of beaches and flats
(gulls, terns, shorebirds). AVONET codes every pelagic family Marine, splits gulls and terns 24
Coastal to 21 Marine, and files sea ducks, loons, and grebes under Wetland, by breeding habitat.
The 2026-11-07 review ([traitAudit](#traitaudit)) added habitats by hand, and the Shore and sea
shortcut hides the inconsistency rather than fixing it.

The fix is a pipeline that looks up each species' usual habitats, through the year and where North
American birders see it, and records them as a BirdView dataset with sources and reasons, perhaps
in BirdView's own classification (shore, nearshore water, offshore) defined over AVONET's terms.
It is mostly text and data analysis, so cheaper than [photoProportions](#photoproportions), but
still per-species work, and roughly ten times as much worldwide. BirdLife's habitats cannot be used
(DataZone's terms forbid automated access), and text sources' licenses decide whether facts drawn
from them may be recorded.

**The cheap version is done as research** (2026-11-08,
[notes: Distance from shore](notes.md#distance-from-shore-as-evidence-for-the-open-sea-question)): the share of a
bird's water records more than 10 km from shore splits AVONET's Marine cleanly into open-ocean
birds (30% or more, about 75 species) and birds of the sea near shore (under 20%, 26), but cannot
separate beach birds from sea ducks. Next: the owner's decision on whether to split Open sea by it,
and if so, a committed tool and a requirement. Cost estimate for the full pipeline: on the issue.
