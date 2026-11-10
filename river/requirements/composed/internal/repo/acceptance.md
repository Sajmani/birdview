package repo

# repo — acceptance criteria

**Composed, not authored.** Composed with [repo.md](repo.md); do not edit by hand, recompose.

How each requirement in [repo.md](repo.md) is decided. Each criterion names what it covers in
`Covers:`; mutable guardrail bindings and `Watched failing:` observations live in
[conformance.md](../../../../conformance.md).

---

## everyRequirementCitesGoal

**Every requirement cites a declared goal, and every goal is cited**

Covers: [tech.bundleIsNormative](../../../local/tech/process.md#bundleisnormative)

**Method:**
Go test over the bundle. River's checker counts a requirement citing no goal and a goal nothing
cites rather than failing them; BirdView is stricter, and `TestEveryLocalRequirementCitesAGoal`
fails both, exempting withdrawn requirements and the rules transcribed from imported sources

**This check cannot tell whether a requirement genuinely serves the goal it names.** Goals are not
measurable by construction. That judgment belongs to Gate 1 and to review, and a green run here
looks exactly like a validated one.

## vendoredHashesMatch

**Every vendored source matches its recorded hash**

Covers: [tech.recordBeforeConsuming](../../../local/tech/process.md#recordbeforeconsuming)

**Method:** hash each committed input and compare
against `river/requirements/imported/*/*/PROVENANCE.md`

**Scope exception:** `wiley` is quoted, not vendored, so it has no hash to check — see
[imports.md](../../../imports.md#wiley). `aba` is fetched at build and checked by sortsAscending instead.

## quotationsAreVerbatim

**Every quotation from a vendored source is word for word**

Covers: [tech.recordBeforeConsuming](../../../local/tech/process.md#recordbeforeconsuming)

**Method:** every block quotation in a
transcribed source's `<pkg>.md` is found in the text of a file that source vendors,
ignoring markup, whitespace, and the style of quotation mark, and splitting at an ellipsis

**Scope:** sources that vendor HTML — `inatapi`, `inatterms` and `wcag`. A source quoted but not
vendored, such as `wiley`, has nothing to check against.

## compositionResolves

**Every composition field names something that exists, and every applied domain is bound**

Covers: [tech.bundleIsNormative](../../../local/tech/process.md#bundleisnormative)

**Method:** read every `Applies to:`, `Binds:`,
`Composed from:` and `Covers:` field in the sources and composed packages; each must name an item
that exists, each domain must be bound by exactly one local requirement, and every domain an
adopted requirement applies to must be bound — otherwise that requirement silently applies to
nothing. Every line of a composed package's list of requirements that apply unchanged or are
excluded must name a requirement that exists, under its exact title, at most once per package,
and an exclusion must name a decision that exists. Composed requirements no criterion covers,
and local requirements no composed package accounts for, are counted and logged, not failed.

## decisionReferencesExist

**Decision records reference requirements that exist**

Covers: [tech.bundleIsNormative](../../../local/tech/process.md#bundleisnormative)

**Method:** extract every qualified reference
(`product.x`, `tech.x`, `goals.x`, `decisions.x`, `app.x`, `dataset.x`, `repo.x`) and every anchored link, and
confirm each resolves

*Known limitation, found by this check catching its own documentation:* the scanner cannot
distinguish a citation from a mention, so writing an example id in prose trips it. Worth
tolerating — the failure is loud, immediate, and safer than a checker that guesses intent.

## nothingRestrictedTracked

**Nothing restricted is tracked by version control**

Covers: [tech.fetchRestrictedInputs](../../../local/tech/delivery.md#fetchrestrictedinputs), [tech.datasetNeverCommitted](../../../local/tech/delivery.md#datasetnevercommitted)

**Method:** `git ls-files` assertion in CI

**Assertion:** no tracked path lies under `fetched/`, and the built dataset is not tracked.

## bothCommandsBuild

**Both commands build and the server serves**

Covers: [tech.twoCommandsOneModule](../../../local/tech/delivery.md#twocommandsonemodule)

**Method:** covered by the standing checks plus an
integration test that starts the server and requests `/`

## dockerStagesCopyWhatTheyBuild

**Every Dockerfile stage copies in what its command imports**

Covers: [tech.containerImageBuilds](../../../local/tech/delivery.md#containerimagebuilds)

**Method:** test reading the `Dockerfile` and the Go sources

**Assertions:** each stage that runs `go build ./cmd/…` copies `go.mod`, `go.sum` and the directory
of every package in this module the command imports, directly or not, reading every non-test file
whatever its build tags; and stages building both `./cmd/birdview` and `./cmd/birdview-ingest` are
found, so a `Dockerfile` the test can no longer parse fails rather than passing empty.
**Does not cover**, because tests must not need Docker or the network: that the base images exist,
that a `COPY` lands where the build expects it (the test assumes each copies to the same relative
path, as all do today), that the ingest finds its input files at run time inside the stage, and that
the final image starts. Counted as **partial** until something runs `docker build`.

## checksWorkflowIsSound

**The checks workflow runs the standing checks on every change, and is safe to run**

Covers: [tech.checksOnEveryChange](../../../local/tech/release.md#checksoneverychange)

**Method:** assertions over `.github/workflows/checks.yml`

**Assertions:** it runs on pushes to `main` and on pull requests, never on `pull_request_target`;
its permissions are `contents: read` and it reads no secret; every action is `actions/*` pinned to
a 40-character commit; it builds, vets, checks formatting, builds the dataset before testing, runs
`go test -v` under an explicitly named `bash`, which is what gives GitHub's shell `pipefail`, and
fails if the test log shows a skip.
**Does not cover:** that GitHub runs it as written; the first run's result is recorded here.

## dockerTestStageGates

**The image cannot be built unless the tests pass, or the bypass is visibly taken**

Covers: [tech.testedBeforeDeploy](../../../local/tech/release.md#testedbeforedeploy)

**Method:** assertions over the `Dockerfile`, and its test stage run by hand

**Assertions:** a `test` stage runs `go test` over every package but `river/check`, against the
dataset the `ingest` stage built; `SKIP_TESTS` is its only bypass and the bypass prints that the
tests were skipped; the `builder` stage copies from `test`, and the final image from `builder`, so
no build can leave the test stage out.
**Does not cover:** a real `docker build`, since Docker is not installed on the maintainer's
machine; tests that need Chrome skip in the image and gate in CI instead.

## activeTodoHasIssue

**Every active todo item links its GitHub issue**

Covers: [tech.activeWorkHasIssue](../../../local/tech/process.md#activeworkhasissue)

**Method:** `river/todo.md` read, offline

**Assertions:** every item whose status is `active` has an `Issue:` line linking one issue of
`google/birdview`, as `[#N](https://github.com/google/birdview/issues/N)`. Not checked by the test,
which does not reach GitHub: that the issue exists and links back, checked when an item is marked
active.

## imageRecordsRevision

**The image builds the deploy's REVISION into the server, and git never commits it**

Covers: [tech.imageRecordsRevision](../../../local/tech/release.md#imagerecordsrevision), [tech.testedBeforeDeploy](../../../local/tech/release.md#testedbeforedeploy)

**Method:** the Dockerfile and ignore files read; the upload listed with `gcloud meta list-files-for-upload`

**Assertions:** the stage that builds the server copies `REVISION*` (so its absence does not fail the
build) and links it with `-X main.revision=`; `.gitignore` ignores `/REVISION`; `.gcloudignore`
includes `.gitignore`, re-includes `!/REVISION`, and does not leave itself out (this test runs in
the image build, where it must find the file). Not run by the test: that gcloud uploads the
file, checked by hand on 2026-11-08 (`REVISION` listed, `.git` and `web/data/species.json` not),
and that the deployed server names it, which the smoke check confirms after every deploy.

## smokeChecksDeployment

**The smoke checks pass on a healthy deployment, and each fails for the breakage it guards**

Covers: [tech.smokeCheckAfterDeploy](../../../local/tech/release.md#smokecheckafterdeploy)

**Method:** run the checks over real HTTP against this server, a canonical-host redirect to it, and broken variants

**Assertions:** against the server and a second one redirecting to it, all of at least eight
checks pass; with `robots.txt` gone, the home page uncompressed or not revalidated, `/?month=now`
cacheable, an unknown species answered, a guide gone, or the home page's issue link naming no
revision, the check written for that breakage fails; and a host that serves the site instead of
redirecting fails the redirect check.

## noUnreachableCode

**Every unexported and internal function in BirdView has a live caller, or is a declared test oracle**

Covers: [tech.bundleIsNormative](../../../local/tech/process.md#bundleisnormative)

**Method:** parse and type-check every Go package in the module, trace reachability from every
`cmd/*` `main` and `init` entry point (plus `web/static/*.html` template method invocations) across
non-test code, and fail on any unreachable function or method outside the declared test-oracle
allowlist (`internal/palette` contrast helpers, `internal/boundary` `Parse` and `Lookup`,
`internal/dataset` `AllMetrics` and `(*Undrawn).Measured`, and `internal/region` `(Region).Zone`),
or on any allowlist entry that no longer exists or has gained a production caller
(`TestNoUnreachableCode`). Also flag functions kept alive only by tests when they sit outside the
oracle allowlist.
