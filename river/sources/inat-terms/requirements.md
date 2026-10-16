# inat-terms — transcribed requirements

Tier: **mandatory**. Scope: BirdView's use of the iNaturalist platform, which consists entirely of
unauthenticated read requests to the public API and outbound hyperlinks to taxon pages.

BirdView contributes nothing to iNaturalist: no observations, no identifications, no comments, no
media, no account. Most of what these documents govern therefore cannot reach it, and saying so
explicitly is the point of this file — an unexamined "probably fine" is what this process exists
to replace.

IDs match the transcription in the maintainer's `birdsync` repository, which vendored the same
documents and is subject to far more of them. See [PROVENANCE.md](PROVENANCE.md).

---

## The finding this adoption rests on

**Nothing in the terms or the guidelines restricts reusing a taxon identifier, or linking to a
taxon page.**

That is a negative, so here is how it was checked rather than asserted. Searching the vendored
copies, `terms.html` contains no occurrence of *taxon*, *taxonomy*, *database*, *derivative*,
*hyperlink*, *deep link*, *link to*, or *scrape*. `community-guidelines.html` contains one
occurrence of "link to", in an example encouraging a user to share a helpful link with another
user.

What the terms *do* license is **Content a contributor posts** — see `ContributorKeepsRights` below. A
taxon record is not contributed Content; it is iNaturalist's own taxonomy. BirdView stores one
integer per species from it and reproduces nothing else.

The terms also state that the agreement "does not transfer to You any iNaturalist or third party
intellectual property", and reserve iNaturalist's trademarks. BirdView claims neither: it stores
identifiers, and the link text on a detail page names iNaturalist as the destination in plain
words rather than reproducing a logo.

---

### ContributorKeepsRights

**A contributor keeps their rights, and licenses by default as CC BY-NC**

> "If you own the Content prior to contributing the Content to the Site, and that Content is
> subject to Intellectual Property rights, you retain ownership of those rights. Unless you
> specify otherwise when you post Content, you agree to license Content you contribute to the
> Platform under the Creative Commons Attribution Noncommercial license (CC BY-NC)."
> — Terms of Use, Responsibility of Contributors

**For BirdView: governs the thing BirdView deliberately does not do.** The photographs a reader
sees after following a link are contributors' work under their own licenses, frequently CC BY-NC.
BirdView never copies, caches, hotlinks or embeds any of them — which is also what
[tech.selfContainedServer](../../tech.md#selfcontainedserver) requires for unrelated reasons, since an embedded image would be an
external subresource. The reader arrives on iNaturalist, where each photograph carries its own
license notice.

If BirdView ever wants to *show* a photograph rather than link to one, this requirement is the
first thing to re-read, and the answer will be per-photograph rather than blanket.

### AccountHolderResponsible

**The account holder is responsible for everything done under the account**

> "You are fully responsible for all activities that occur under the account and any other
> actions taken in connection with the account."
> — Terms of Use, Your iNaturalist Account

**Not applicable.** BirdView has no account.

### RateDoesNotHinderOthers

**Post at a rate and volume that doesn't hinder other users**

> "You will post only Content that is relevant to iNaturalist and at a rate and volume that
> does not hinder other Users' ability to use iNaturalist"
> — Terms of Use, Responsibility of Contributors

**Not applicable as written** — BirdView posts nothing. Its read rate is governed instead by
`inatapi.RequestRate`, implemented as [tech.pacedInaturalistMapping](../../tech.md#pacedinaturalistmapping).

### NoMachineGeneratedContent

**Machine generated content is prohibited** *(hard rule)*

> "(!) Machine generated observations, identifications, comments, and messages. We do not
> allow machines to generate and post content on iNat with no human oversight curating each
> piece of content, and any account suspected of doing so is subject to suspension and the
> removal of the content."
> — Community Guidelines, marked (!): "grounds for immediate suspension without warning"

**Not applicable.** The rule is about generating and *posting* content. BirdView posts nothing
and cannot; it has no write path and no credential.

Recorded rather than omitted because it is the sharpest rule in the document, and a future change
that gave BirdView any write capability at all would land squarely inside it. `birdsync` had to
resolve exactly this question and did so against the linked definition, vendored here as
`machine-generated-content.html`.

### DoNotBehaveLikeMachine

**Don't behave like a machine**

> "Any account that adds content we believe decreases the accuracy of iNaturalist data may be
> suspended, particularly if that account behaves like a machine, e.g. adds a lot of content
> very quickly and does not respond to comments and messages."
> — Community Guidelines

**Not applicable.** Scoped to accounts adding content.

### DuplicatesUndesirable

**Duplicates are undesirable but not an offense**

> "Duplicate observations. They're not ideal, but they're usually due to oversight or bugs."
> — Community Guidelines

**Not applicable.** BirdView creates no observations.

---

## Not adopted

Hate speech, harassment, sockpuppets, sexually explicit content, and the prohibition on commercial
AI training are binding on users of the platform and unreachable by a read-only client that posts
nothing.
