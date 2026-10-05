# BirdView

**How big is that bird, really?**

BirdView draws North American birds as a scaled schematic — bill, head, neck, body,
tail, and a wing crossbar — at one shared scale, so you can *see* the difference in size instead of
reading it off a number and trying to imagine it.

It covers **1083 species across 99 families and 27 orders**: every bird whose records **recur**
in the ABA Area — repeated episodes across years, in at least one region — by a rule BirdView
derives and owns.

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
length is modeled for all 1083.

That is why every value carries its own provenance — `measured`, `near-match` or `modeled` — and
why clicking any bird shows you the source or the formula behind each individual number.
[docs/data.md](docs/data.md) is the full accounting, including the parts that are frankly invented.

---

## Status

Runs locally. **The licensing blocker is gone**: nothing all-rights-reserved is read at build
time any more, so serving BirdView reproduces nobody else's work.

Names, families and taxonomic sequence come from [AviList](spec/sources.md#avilist)
([CR-022](spec/decisions.md#cr-022)), and **scope is now BirdView's own**
([CR-028](spec/decisions.md#cr-028)). The ABA was asked for permission on 2026-09-27 and never
replied, so the project stopped waiting and derived a list it owns.

What replaced the checklist is not a better threshold on how *often* a bird is recorded. A record
count measures observer effort: it under-reports pelagic species and remote-island endemics,
because nobody is where they are, and over-reports vagrants, because everybody goes to where they
are. Those pull in opposite directions and no single cutoff corrects both. Scope instead asks
whether a species' records **recur** — repeated episodes across years, in at least one region —
which is insensitive to both. Akikiki has 376 records across 45 years and is a resident; the 2023
American Flamingo in Pennsylvania has 931 inside 13 days and is not.

---

## Data and licensing

| Source | License | How it is used |
| --- | --- | --- |
| [AVONET](https://doi.org/10.6084/m9.figshare.16586228.v7) (Tobias et al. 2022) | CC BY 4.0 | morphology traits; vendored in this repo |
| [Fox et al. 2026 wingspan database](https://doi.org/10.6084/m9.figshare.32089083) | CC BY 4.0 | empirical wingspan; vendored in this repo |
| [GBIF](https://www.gbif.org) — [eBird Observation Dataset](https://doi.org/10.15468/aomfnb) and iNaturalist research-grade | CC BY 4.0 | which regions each bird is found in; only the derived list is kept, never a record |
| [iNaturalist](https://www.inaturalist.org) | terms vendored; taxon IDs are facts | one taxon ID per species, to link the detail page out for photographs |

Every source above is CC BY 4.0 and is redistributed here under its license, with attribution.
**No all-rights-reserved input remains.** The ABA Checklist supplied scope until 2026-10-05 and
was never committed; it is no longer read at all.

BirdView **links** to iNaturalist and copies nothing from it: no photograph is embedded, cached
or hotlinked, so no contributor's license is engaged. See
[CR-012](spec/decisions.md).

BirdView's own code is in this repository, along with the tables it derives under `data/` —
the occurrence status, the region presence, and the two name bridges. The built dataset
(`web/data/species.json`) stays a build artifact and is not committed: it is reproducible from
the committed inputs in one command.

---

## Contact

Feedback: **birdview at ajmani dot net**.

Corrections to the biology are especially welcome — see
[docs/data.md](docs/data.md#where-this-is-most-likely-to-be-wrong) for where the models are
weakest and least validated.
