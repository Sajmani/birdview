# BirdView

**How big is that bird, really?**

BirdView draws North American birds as stick figures — bill, head, neck, body, tail, and a wing
crossbar — at one shared scale, so you can see size differences instead of imagining them from
numbers. Is a Merlin bigger than a Blue Jay? Line them up and look.

It covers over a thousand species: those whose records recur in the US states, the Canadian
provinces, five US outlying islands and five stretches of open sea. Each figure is an abstract
top-down view of a bird in flight, showing only what the data supports: no plumage, posture or
shape the sources cannot justify.

## Try it

You need [Go](https://go.dev/dl/). From a clone of this repository:

```
go run ./cmd/birdview-ingest     # builds the dataset from the sources in this repo
go run ./cmd/birdview            # serves http://localhost:8080
```

Then open <http://localhost:8080> and follow one of the suggestions under the title.

## Read more

| If you are… | Read | It covers |
| --- | --- | --- |
| **using the site** | [docs/using.md](docs/using.md) | the controls, the stick figures, reading a tile, things to try |
| **checking the science** | [docs/data.md](docs/data.md) | what is measured, what is modeled, what is invented |
| **tracing a number to its source** | [docs/pipeline.md](docs/pipeline.md) | every source, the fields read from it, and each step between |
| **building or changing it** | [docs/building.md](docs/building.md) | build, run, deploy, the data's licenses, and the spec-driven workflow |

## What is measured and what is modeled

No openly licensed dataset publishes body length for these birds, so **every length in BirdView is
a model** built from measured parts. Bill, tail and mass are measured for almost every species;
wingspan is measured for about a third and estimated for the rest; head, neck and body are modeled
from mass. Every value is marked `measured`, `near-match` or `modeled`, and clicking a bird shows
the source or formula behind each number. [docs/data.md](docs/data.md) gives the full accounting,
including what is invented and
[where the models are most likely to be wrong](docs/data.md#where-this-is-most-likely-to-be-wrong).

## Data

BirdView is built from sources under CC BY 4.0: names and taxonomy from AviList, body measurements
from AVONET, wingspans from Fox et al. 2026, and where each bird is found from eBird and
iNaturalist records via GBIF. It links to iNaturalist for photographs and copies none. Every source
is credited on the site; [docs/building.md](docs/building.md#the-data-and-its-licenses) lists them
with licenses and what this repository keeps.

## License

BirdView's own code and documentation are licensed under the [Apache License 2.0](LICENSE).
**Everything under [`third_party/`](third_party) is not**: each directory there carries its own
license and a README naming its source.

This is not an officially supported Google product.

## Contributing

Report problems and corrections as issues; corrections to the biology are especially welcome. See
[CONTRIBUTING.md](CONTRIBUTING.md) to contribute code and [CHANGELOG.md](CHANGELOG.md) for major
changes.

## Security

Eligibility for the [Google Open Source Software Vulnerability Rewards
Program](https://bughunters.google.com/open-source-security) is determined by the [Google Open
Source Software Vulnerability Reward Program
Rules](https://bughunters.google.com/about/rules/open-source/google-open-source-software-vulnerability-reward-program-rules).
