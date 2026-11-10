package repo

# repo — composed requirements for repository structure, build, and release

**Composed, not authored.** Change it by changing a source and recomposing, never by editing files in
this directory by hand. Composed from `tech` and `decisions` under
[precedence.md](../../../precedence.md); approvals are recorded in
[approvals.md](../../../approvals.md).

Everything BirdView's repository layout, build inputs and artifacts, container image, continuous
integration, deployment posture, and development process must satisfy. Its acceptance criteria are
in [`acceptance.md`](acceptance.md), and the guardrails and verification observations covering them
are in [conformance.md](../../../../conformance.md).

## Composed from

- tech (Local)
- decisions
- precedence

## Composition record

| Source | Tier | Pin |
| --- | --- | --- |
| [`tech`](../../../local/tech/tech.md) | local | the commit this file is in |
| [`decisions`](../../../decisions.md) | — | the commit this file is in |

Precedence is the ranking in [precedence.md](../../../precedence.md). **No conflict arose.**

## Applies unchanged

### Module, build inputs, and artifacts

- [tech.twoCommandsOneModule](../../../local/tech/delivery.md#twocommandsonemodule) — Two commands, one module
- [tech.goVersionFollowsDeps](../../../local/tech/delivery.md#goversionfollowsdeps) — The Go version tracks whatever dependencies require
- [tech.fetchRestrictedInputs](../../../local/tech/delivery.md#fetchrestrictedinputs) — Build inputs that are not vendored are fetched into `fetched/`, hash-checked, and never committed
- [tech.datasetNeverCommitted](../../../local/tech/delivery.md#datasetnevercommitted) — The built dataset is an artifact, never committed
- [tech.containerImageBuilds](../../../local/tech/delivery.md#containerimagebuilds) — The container image builds from a clean checkout

### Process and posture

- [tech.localNowPublicLater](../../../local/tech/posture.md#localnowpubliclater) — Public as a personal site; never commercial
- [tech.bundleIsNormative](../../../local/tech/process.md#bundleisnormative) — The spec bundle is normative
- [tech.recordBeforeConsuming](../../../local/tech/process.md#recordbeforeconsuming) — No source is consumed before it is recorded
- [tech.activeWorkHasIssue](../../../local/tech/process.md#activeworkhasissue) — Work in progress is marked active in the todo list and tracked by a GitHub issue

### Continuous integration and release gates

- [tech.checksOnEveryChange](../../../local/tech/release.md#checksoneverychange) — Every push to main and every pull request runs the standing checks
- [tech.testedBeforeDeploy](../../../local/tech/release.md#testedbeforedeploy) — A deploy builds only code whose tests pass
- [tech.imageRecordsRevision](../../../local/tech/release.md#imagerecordsrevision) — The deployed server knows the commit it was built from
- [tech.smokeCheckAfterDeploy](../../../local/tech/release.md#smokecheckafterdeploy) — Every deploy is checked from outside
