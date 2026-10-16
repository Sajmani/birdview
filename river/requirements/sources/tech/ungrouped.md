package tech

# Not yet grouped

Requirements added after the groups above were laid out, which sat under the bindings section
only because they were appended after it. Each belongs in one of the groups above; placing them
is part of splitting this source into files.

## selfDescriptionStaysTrue

**What the project says about itself stays true**

Subject: `docs.currency`
Goals: goals.shapeNotOnlySize

**Every CC BY source whose output appears on a page is credited on that page.** This is a licence
condition, not a courtesy: AviList, AVONET and Fox are all CC BY 4.0, and attribution is the
thing they ask for in return.

It was not being met. The index page carried full credits, while **`detail.html` and
`compare.html` credited AVONET, Fox and the ABA and omitted AviList** — on pages whose every
English name, scientific name, family and species code comes from AviList. A reader could use
BirdView for an hour and never see the source of the words on the screen.

**User-facing documentation describes the project as it is, not as it was.** `README.md` and
`docs/` are read before the code and are the first thing a new reader trusts, so a stale claim
there is worse than no claim. Two kinds of staleness are forbidden:

- **A count that no longer matches the dataset.** `docs/data.md` claimed 841 species after
  [decisions.avilistReplacesAba](../../decisions.md#avilistreplacesaba) cut the list to 839.
- **A source credited for something it no longer supplies.** `README.md` listed the ABA Checklist
  as the source of "species list, names, families" after decisions.avilistReplacesAba moved names and families to
  AviList (at that point the ABA checklist supplied **scope** alone; [decisions.scopeFromRecurrence](../../decisions.md#scopefromrecurrence)
  then replaced scope too).

**Before any public deployment, the docs are checked against the spec**, since a release is the
moment the claims become promises to strangers rather than notes to ourselves.

The check is mechanical where it can be — counts and credits are comparable against the dataset
and the page — and a reading task where it cannot. [acceptance.creditsComplete](../../acceptance.md#creditscomplete) does the
mechanical part, and records plainly that prose accuracy is not something a test can hold.

## noPersonalAddress

**No personal address is published**

Subject: `deploy.contact`
Goals: goals.goodDataCitizen

No page BirdView serves carries a contact address. Reader feedback goes to the project's issue
tracker, whose URL is a deployment setting (`-feedback`) rather than a constant in the markup.

**Unset publishes nothing**, not a dead link. A site may legitimately go up before its tracker
exists; what it must not do is invite a reader somewhere that does not.

Rationale: the header carried the owner's personal mailbox for the project's whole life, spelled
out in words rather than written plainly. **That obfuscation is the argument against it.** It
was written that way because a plain address on a public page is a harvested mailbox — which is
a reason not to publish the address at all, not a reason to spell it oddly. Beyond the spam, an
issue tracker is the better channel: a bug report in public can be read, searched, corrected by
someone else and closed, where mail to one person is none of those things. The owner's personal
mailbox is not the project's contact point.

*(This requirement does not quote the address it removed, and that is not squeamishness. The
requirements are published with the repository, so a quotation here would republish exactly
what the rule forbids — which is how [acceptance.noContactAddressPublished](../../acceptance.md#nocontactaddresspublished) first failed, back when they
were also served at `/spec`.)*

**The `User-Agent` contact stays, and is a different thing.** [identifiedClient](delivery.md#identifiedclient) requires outbound
requests to name a contact, and iNaturalist's API guidance asks for it specifically. That
address reaches a handful of API operators who may need to reach whoever is making the requests;
it is not published to visitors. Removing it would be discourteous and risks the project being
blocked.

**This extends to addresses that are not ours.** The source manifest recorded a source owner's
mailbox, which the served spec then republished to every visitor. A provenance record names who
owns a source; it does not need to hand out their mail.

## compressedResponses

**Responses are compressed when the client asks**

Subject: `http.content_encoding`
Value: `gzip on Accept-Encoding, Vary on every response`
Goals: goals.sizeIntuition, goals.goodDataCitizen

Every response is gzipped when the request carries `Accept-Encoding: gzip`, and every response
carries `Vary: Accept-Encoding` whether or not it was compressed.

Rationale: the index inlines a full SVG diagram for each of 1,083 species, which is **3.3 MB of
markup compressing to 270 KB — a 92% saving**. Nothing else BirdView could do moves a number
that far. It matters beyond politeness because egress is the one axis of a serverless
deployment that bills: against 1 GiB of free monthly transfer, this is the difference between
about **320 index loads and about 3,980**.

**Done in the application, not left to a CDN or a load balancer**, because the saving should
not depend on which of several deployment shapes is chosen, and because the shape is not chosen
yet.

**`Vary` goes on uncompressed responses too.** Without it a cache that stored the plain reply
would serve it to clients asking for gzip, and — the damaging direction — hand a gzip stream to
a client that cannot read one.

**The headers are decided at `WriteHeader`, not at the first `Write`.** Any handler that sets a
status itself has already committed the header map by the time a byte is written, so a
compressor that labels from `Write` emits a correctly compressed body with no
`Content-Encoding` and the browser renders binary. Both static files and every
[product.undrawnSpeciesAnswer](../product/presentation.md#undrawnspeciesanswer) 404 shipped that way in the first draft of this.

**A status that may carry no body — 204, 304 — is not compressed at all**, an empty gzip
stream's twenty-byte header being bytes the protocol says must not be there.

## joinsReportDrops

**Every join reports what it dropped**

Subject: `build.join_loss`
Value: `named and counted on every run; a declared bar above which the run refuses`
Goals: goals.goodDataCitizen, goals.nothingDisappearsQuietly

Every join between two datasets reports the records it could not match: how many, under what
names, and whether the loss was intended. Where a loss can be bounded, the bar is declared and
the run **refuses to write** rather than exceed it.

Rationale: [goals.nothingDisappearsQuietly](../../../goals/goals.md#nothingdisappearsquietly) at the point where data actually disappears. A join is the
one place in this pipeline where a record can cease to exist without anyone deciding it should,
and the failure is invisible on both sides — the input still has the record, the output simply
never mentions it.

**The ingest's joins already did this and the region pass did not**, which is exactly where the
loss was. `birdview-ingest` prints what the AviList join dropped and asserts it against a
declared list, prints what AVONET could not match and fails above a threshold, and requires
iNaturalist coverage to be total. `birdview-regions` matched raw facet names against AviList
names and discarded the rest in silence: **37 names, 10,015,688 records, of which 9,954,362
belonged to ten species BirdView draws** ([decisions.regionJoinNormalizes](../../decisions.md#regionjoinnormalizes)).

**A discarded record is worse than absent, because it still counts.** Every one of those names
reached the denominator of [product.shareOfRecords](../product/presentation.md#shareofrecords)'s share and no species' numerator — so
Hairy Woodpecker's 7,310,342 eBird records were counted as *some other bird* on every tile
including its own, and the bird showed 0.7% of its own record.

**The bar is a fraction of records, not a count of names.** A new vagrant adds a name and six
records; a genus move adds one name and seven million.

**Normalizing is part of the join, not a step before it.** Two commands read the occurrence
data, and the loss above was not a missing table entry — every entry already existed — but a
second reader that never consulted them. So every reader normalizes through **one function**,
and a criterion asserts the two agree on the species rather than asserting either in isolation
([acceptance.occurrenceReadersAgree](../../acceptance.md#occurrencereadersagree)). The three gaps [decisions.regionJoinNormalizes](../../decisions.md#regionjoinnormalizes) found surfaced
one at a time over two days precisely because nothing compared the readers, so each fix left
the next one standing: no rename, then no lump, then no reduction of rank.

**Rank is reduced before any name is looked up**, the naming tables being keyed on binomials
that `dryobates villosus harrisi` matches none of. The ordering that is load-bearing the other
way is the hybrid test, which runs *before* the truncation: cut to two tokens, `larus
smithsonianus × hyperboreus` is an ordinary species name that joins cleanly and credits a real
bird with records of something that is not it.

Measured 2026-10-06, the region pass's unmatched residue fell from **0.0431% under 1,754
names to 0.0113% under 727**, moving **387,515 records into 430 species**. Region totals came
back byte-identical, so this reattributes records rather than changing what they are divided
by, and **114 species-region presences were added and none removed** — the only shape the
change can have, since a trinomial either joins a drawn species or stays where it was. The
largest single beneficiaries are the two declared domestic forms, feral Rock Pigeon
(+109,401) and Muscovy Duck (+37,470), and one genuine subspecies recovery, Yellow-rumped
Warbler (+30,274) from `coronata` and `auduboni`.

## cacheLifetimeStated

**Every response says how long it may be reused**

Subject: `serving.cache_validation`
Value: `Cache-Control: no-cache and an ETag on every response`
Goals: goals.sizeIntuition, goals.nothingDisappearsQuietly

Every response carries a `Cache-Control` directive and a validator, so a cache knows whether
it may reuse what it holds and can ask cheaply when it may not.

Rationale: **sending no cache headers is not the same as sending "do not cache".** A response
with no freshness information and no validator lets a browser apply its own heuristic, and
leaves it nothing to revalidate against when it does. The failure is a reader who reloads
after a rebuild, is served the old page, and has no way to tell — which is indistinguishable
from a bug in the page and was reported as one.

**`no-cache` means store it and ask before reusing it**, not "do not store". That is the right
directive for every page here: validation is cheap and staleness is never acceptable, since
the point of a figure like [product.shareOfRecords](../product/presentation.md#shareofrecords)'s is that it is current.

**The ETag is a fingerprint of the build and the URL, not a hash of the body.** Hashing the
body means buffering it, and the index is 3.3 MB before compression — the memory this would
cost on every concurrent request is the thing [compressedResponses](#compressedresponses) exists to avoid spending. The
substitute is sound only because a page is a pure function of the dataset, the templates, the
static files and the request URI, with no clock anywhere in the render path, and
[acceptance.honestRevalidation](../../acceptance.md#honestrevalidation) asserts that rather than trusting it.

**One fingerprint over all assets, not one per asset.** A stylesheet edit invalidates the HTML
that loads it. Over-invalidating costs one refetch after a deploy; under-invalidating serves a
page styled by a stylesheet it was not written for, which is the failure hardest to recognize
as a caching problem.

**The compressed representation carries a different validator**, suffixed rather than dropped.
A gzipped body is a different representation of the resource, and a cache that let the two
share a tag would eventually hand gzip bytes to a client that asked for none — the same class
of error `Vary: Accept-Encoding` exists to prevent, one layer down.

**`-dev` sends `no-store` and no validator at all.** It re-reads templates per request, so no
fingerprint taken at startup stays true, and a stale page under the one flag that exists to
make edits visible would defeat the flag.

## sourceNamingMeasured

**How each source names a species is measured, not assumed**

Subject: `build.name_correspondence`
Value: `committed per-source record counts; a declared ratio band`
Goals: goals.goodDataCitizen, goals.dataAsItIs, goals.nothingDisappearsQuietly

A committed table records how many records each occurrence source holds for every species
BirdView draws, under the name BirdView serves. A species far outside the usual ratio between
the two sources is reported, and every exception is declared with a reason.

Rationale: a name can fail to join in two ways and only one is visible. A name that **misses**
leaves a species at zero, which somebody eventually notices — Bullock's Oriole in no region at
all. A name that **matches the wrong concept** joins cleanly and stays wrong: eBird's
*Setophaga petechia* is AviList's Mangrove Yellow Warbler by string and the undivided Yellow
Warbler by meaning, and **no name-matching check can see it** —
[`data/gbif-taxa.tsv`](../../../../data/gbif-taxa.tsv) files that pair as `exact`, correctly.

What separates them is asking each database how many records it holds. The ratio is
unexpectedly stable — **median 48×, 5th percentile 10.7×, 95th 148×** — because iNaturalist is
uniformly shallower than eBird rather than differently distributed. Every kind of disagreement
shows as an outlier, and in opposite directions:

| shape | signature |
| --- | --- |
| a rename or lump | eBird ≈ 0; its records are under another name |
| a shared name | ratio enormous — *Setophaga petechia* is **117,834×** |
| a withheld species | eBird ≈ 0 under *every* name, iNaturalist normal |
| a domestic form | ratio large — Rock Pigeon 11,809×, iNaturalist filing under a trinomial |

**Only 7 of 1,019 species fall outside 1×–1000×, and every one is a real finding.** Measured
2026-10-05, the band immediately caught two species nothing had recorded as withheld —
**Great Gray Owl** (eBird 1, iNaturalist 2,859) and **Black Rail** (16 against 101) — and
exposed an error in [`sources.md`](../../requirements.md), which listed Blue-billed White-Tern among the
withheld when eBird holds 16,008 records of it under *Gygis alba*.

**The declared exceptions are the point, not an escape hatch.** Each names which of the four
shapes it is, so the table reads as a statement about taxonomy rather than a list of numbers
to ignore.
