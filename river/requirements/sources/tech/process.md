package tech

# Process

## bundleIsNormative

**The spec bundle is normative**

Subject: `process.spec_location`
Value: `river/`
Goals: goals.giveBack, goals.honestData

The repository follows the loop in [README.md](../../../README.md). `AGENTS.md` names `river/` as
normative so agents reading the instruction file are pointed at it.

## pacedInaturalistMapping

**The iNaturalist mapping is built by a separate, paced tool**

Goals: goals.goodDataCitizen, goals.rapidIteration

Resolving species to iNaturalist taxon IDs happens in `cmd/birdview-inatmap`, never in
`birdview-ingest` and never while serving. The tool waits at least one second between requests
(`inatapi.RequestRate`) and identifies itself (identifiedClient). Its output is a table committed to the repository;
the ingest joins against that table and makes no network request.

Rationale: 1082 requests at the rate iNaturalist asks for is a quarter of an hour, which does not
belong in a build — and a build that needs the network is a build that fails on a plane. Keeping
resolution separate also keeps selfContainedServer and the no-network-in-tests binding intact without exception.
The table going stale is the accepted cost; a stale ID fails visibly and the refresh is one
command.

---

## recordBeforeConsuming

**No source is consumed before it is recorded**

Goals: goals.goodDataCitizen

Every input source appears in [requirements.md](../../requirements.md) with its tier, pin, license and status
*before* code that reads it is written.

Rationale: this project found, after designing around three datasets, that two of them could not
legally be redistributed and a third did not contain the fields it was chosen for. Recording first
is cheaper than discovering later.

---

Project bindings moved to [bindings.md](../../../bindings.md) on adopting River, whose layout gives them
their own file.
