package tech

# Process

## bundleIsNormative

**The spec bundle is normative**

Subject: `process.spec_location`
Value: `river/`
Goals: [goals.giveBack](../../../goals/goals.md#giveback), [goals.honestData](../../../goals/goals.md#honestdata)

The repository follows the loop in [README.md](../../../README.md). `AGENTS.md` names `river/` as
normative so agents reading the instruction file are pointed at it.

## pacedInaturalistMapping

**The iNaturalist mapping is built by a separate, paced tool**

Goals: [goals.goodDataCitizen](../../../goals/goals.md#gooddatacitizen), [goals.rapidIteration](../../../goals/goals.md#rapiditeration)

Resolving species to iNaturalist taxon IDs happens in `cmd/birdview-inatmap`, never in
`birdview-ingest` and never while serving. The tool waits at least one second between requests
([`inatapi.RequestRate`](../../imported/inaturalist.org/inatapi/inatapi.md#requestrate)) and identifies itself (identifiedClient). Its output is a table committed
to the repository; the ingest joins against that table and makes no network request.

Rationale: 1082 requests at the rate iNaturalist asks for take a quarter of an hour, too long for a
build, and a build that needs the network fails offline. Keeping resolution separate keeps
selfContainedServer and the no-network-in-tests binding free of exceptions. A stale table is the
accepted cost: a stale ID fails visibly and the refresh is one command.

---

## recordBeforeConsuming

**No source is consumed before it is recorded**

Goals: [goals.goodDataCitizen](../../../goals/goals.md#gooddatacitizen)

Every input source appears in [imports.md](../../imports.md) with its tier, pin, license,
and status *before* code that reads it is written.

Rationale: after designing around three datasets, this project found that two could not legally
be redistributed and the third lacked the fields it was chosen for.

---

Project bindings are in [bindings.md](../../../bindings.md).
