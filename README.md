# BirdView

**How big is that bird, really?**

BirdView draws North American birds as a scaled schematic — bill, head, neck, body,
tail, and a wing crossbar — at one shared scale, so you can *see* the difference in size instead of
reading it off a number and trying to imagine it.

It covers **839 species across 90 families and 26 orders**: everything on
[ABA Checklist 8.19](https://www.aba.org/aba-checklist/) at rarity codes 1–3, which is to say every
bird you have a realistic chance of seeing in North America.

The diagrams are deliberately abstract. They are a **top-down view of a bird in flight**, and they
carry only what the data supports: no plumage, no posture, no shape the sources cannot justify.

---

## Three ways in

| If you are… | Read | It covers |
| --- | --- | --- |
| **using the site** | [docs/using.md](docs/using.md) | the controls, the diagram, reading a tile |
| **checking the science** | [docs/data.md](docs/data.md) | what is measured, what is modeled, what is invented |
| **building or changing it** | [docs/building.md](docs/building.md) | build, run, and the spec-driven workflow |

---

## The honest summary

Most of what you see is **modeled**, and BirdView says so everywhere it can.

No openly licensed dataset publishes body length for North American birds. The one source that
does is under terms that forbid us using it ([CR-002](spec/decisions.md)), so **every length
in BirdView is a model** — built from measured parts, with each part labeled. Of the three numbers
on a tile, mass is measured for almost every species, wingspan is measured for about a third, and
length is modeled for all 839.

That is why every value carries its own provenance — `measured`, `near-match` or `modeled` — and
why clicking any bird shows you the source or the formula behind each individual number.
[docs/data.md](docs/data.md) is the full accounting, including the parts that are frankly invented.

---

## Status

Runs locally; **not deployed publicly**, and blocked from being so.

Every **name** now comes from [AviList](spec/sources.md#avilist), which is CC BY 4.0, so the ABA
checklist no longer reaches a page through naming. One thing still comes from it: **scope**, the
choice of which 839 species are in. The ABA was asked for permission on 2026-09-27 and has not replied,
so the next step is a species list BirdView derives and owns. See
[CR-022](spec/decisions.md) and the Work arising table in
[spec/decisions.md](spec/decisions.md).

---

## Data and licensing

| Source | License | How it is used |
| --- | --- | --- |
| [AVONET](https://doi.org/10.6084/m9.figshare.16586228.v7) (Tobias et al. 2022) | CC BY 4.0 | morphology traits; vendored in this repo |
| [Fox et al. 2026 wingspan database](https://doi.org/10.6084/m9.figshare.32089083) | CC BY 4.0 | empirical wingspan; vendored in this repo |
| [ABA Checklist 8.19](https://www.aba.org/aba-checklist/) | **all rights reserved** | which species are in scope, and nothing else; **fetched at build, never committed** |
| [iNaturalist](https://www.inaturalist.org) | terms vendored; taxon IDs are facts | one taxon ID per species, to link the detail page out for photographs |

The two CC BY sources are redistributed here under their license, with attribution. The ABA
checklist is not redistributed at all — you fetch your own copy, and `.gitignore` enforces that.

BirdView **links** to iNaturalist and copies nothing from it: no photograph is embedded, cached
or hotlinked, so no contributor's license is engaged. See
[CR-012](spec/decisions.md).

BirdView's own code is in this repository; the built dataset (`web/data/species.json`) is a build
artifact and is not committed, because it carries values derived from the checklist.

---

## Contact

Feedback: **birdview at ajmani dot net**.

Corrections to the biology are especially welcome — see
[docs/data.md](docs/data.md#where-this-is-most-likely-to-be-wrong) for where the models are
weakest and least validated.
