# BirdView

**How big is that bird, really?**

BirdView draws North American birds as a scaled schematic — bill, head, neck, body,
tail, and a wing crossbar — at one shared scale, so you can *see* the difference in size instead of
reading it off a number and trying to imagine it.

It covers **1082 species across 99 families and 27 orders**: every bird whose records **recur**
in the regions it covers — repeated episodes across years, in at least one region — by a rule
BirdView derives and owns. Those regions are the US states, the Canadian provinces, five US
outlying islands and five stretches of open sea; they are BirdView's own list and not the ABA
Area, which includes St-Pierre et Miquelon and excludes four of the islands.

Each bird is drawn as a stick figure, to scale and deliberately abstract: a **top-down view of a
bird in flight** that carries only what the data supports — no plumage, no posture, no shape the
sources cannot justify.

---

## Three ways in

| If you are… | Read | It covers |
| --- | --- | --- |
| **using the site** | [docs/using.md](docs/using.md) | the controls, the stick figures, reading a tile |
| **checking the science** | [docs/data.md](docs/data.md) | what is measured, what is modeled, what is invented |
| **building or changing it** | [docs/building.md](docs/building.md) | build, run, and the spec-driven workflow |
| **tracing a number to its source** | [docs/pipeline.md](docs/pipeline.md) | every source, the fields read from it, and each step between |

---

## The honest summary

Most of what you see is **modeled**, and BirdView says so everywhere it can.

No openly licensed dataset publishes body length for North American birds. The one source that
does is under terms that forbid us using it ([decisions.ricklefsProhibited](river/requirements/decisions.md#ricklefsprohibited)), so **every length
in BirdView is a model** — built from measured parts, with each part labeled. Of the three numbers
on a tile, mass is measured for almost every species, wingspan is measured for about a third, and
length is modeled for all 1082.

That is why every value carries its own provenance — `measured`, `near-match` or `modeled` — and
why clicking any bird shows you the source or the formula behind each individual number.
[docs/data.md](docs/data.md) is the full accounting, including the parts that are frankly invented.

---

## Status

Runs locally. **The licensing blocker is gone**: nothing all-rights-reserved is read at build
time any more, so serving BirdView reproduces nobody else's work.

Names, families and taxonomic sequence come from [AviList](river/requirements/requirements.md#avilist)
([decisions.avilistReplacesAba](river/requirements/decisions.md#avilistreplacesaba)), and **scope is now BirdView's own**
([decisions.scopeFromRecurrence](river/requirements/decisions.md#scopefromrecurrence)). The ABA was asked for permission on 2026-09-27 and never
replied, so the project stopped waiting and derived a list it owns.

What replaced the checklist is not a better threshold on how *often* a bird is recorded. A record
count measures observer effort: it under-reports pelagic species and remote-island endemics,
because nobody is where they are, and over-reports vagrants, because everybody goes to where they
are. Those pull in opposite directions and no single cutoff corrects both. Scope instead asks
whether a species' records **recur** — repeated episodes across years, in at least one region —
which is insensitive to both. Akikiki has 376 records across 45 years and is a resident; the 2023
American Flamingo in Pennsylvania has 931 inside 13 days and is not.

What is left to do is listed in [river/todo.md](river/todo.md).

---

## Data and licensing

| Source | License | How it is used |
| --- | --- | --- |
| [AVONET](https://doi.org/10.6084/m9.figshare.16586228.v7) (Tobias et al. 2022) | CC BY 4.0 | morphology traits; vendored in this repo |
| [Fox et al. 2026 wingspan database](https://doi.org/10.6084/m9.figshare.32089083) | CC BY 4.0 | empirical wingspan; vendored in this repo |
| [GBIF](https://www.gbif.org) — [eBird Observation Dataset](https://doi.org/10.15468/aomfnb) and iNaturalist research-grade | CC BY 4.0 | which regions each bird is found in; only the derived list is kept, never a record |
| [iNaturalist](https://www.inaturalist.org) | terms vendored; taxon IDs are facts | one taxon ID per species, to link the detail page out for photographs |

How these become the dataset, field by field, is drawn in
[docs/pipeline.md](docs/pipeline.md).

Every source above is CC BY 4.0 and is redistributed here under its license, with attribution.
**No all-rights-reserved input remains.** The ABA Checklist supplied scope until 2026-10-05 and
was never committed; it is no longer read at all.

BirdView **links** to iNaturalist and copies nothing from it: no photograph is embedded, cached
or hotlinked, so no contributor's license is engaged. See
[decisions.inaturalistLinkAdopted](river/requirements/decisions.md#inaturalistlinkadopted).

BirdView's own code is in this repository, along with the tables it derives under `data/` —
the occurrence status, the region presence, and the two name bridges. The built dataset
(`web/data/species.json`) stays a build artifact and is not committed: it is reproducible from
the committed inputs in one command.

---

## License

BirdView's own code and documentation are licensed under the
[Apache License 2.0](LICENSE). See [CONTRIBUTING.md](CONTRIBUTING.md) for how to contribute.

**Everything under [`third_party/`](third_party) is not.** Each directory there carries its
own license and a README naming its source:

- [AviList](third_party/avilist), [AVONET](third_party/avonet) and the
  [Fox et al. wingspans](third_party/foxwingspan) are licensed by their authors under
  **CC BY 4.0**, and must be credited as each README describes.
- The [iNaturalist terms](third_party/inatterms) and [API practices](third_party/inatapi) are
  **© iNaturalist** and openly licensed by no one. They are kept so that every quotation of
  them in [river/](river) can be checked against the exact text it came from.

This is not an officially supported Google product.

---

## Security

Eligibility for the [Google Open Source Software Vulnerability Rewards
Program](https://bughunters.google.com/open-source-security) is determined by the [Google Open
Source Software Vulnerability Reward Program
Rules](https://bughunters.google.com/about/rules/open-source/google-open-source-software-vulnerability-reward-program-rules).

---

## Contact

Feedback and corrections: please open an issue.

Corrections to the biology are especially welcome — see
[docs/data.md](docs/data.md#where-this-is-most-likely-to-be-wrong) for where the models are
weakest and least validated.
