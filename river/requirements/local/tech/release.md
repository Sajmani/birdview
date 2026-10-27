package tech

# Release

How a change reaches birdview.site, and what checks it on the way. The deployment's own settings —
its health probe, uptime check, alerts, and runbook — belong to the deployment, not to this
repository, and live with it (`bindings.md`, Stage adoption).

## checksOnEveryChange

**Every push to main and every pull request runs the standing checks**

Subject: `release.ci`
Goals: [goals.secure](../../../goals/goals.md#secure), [goals.rapidIteration](../../../goals/goals.md#rapiditeration)

GitHub Actions runs the four standing checks on every push to `main` and every pull request, with
the dataset built and Chrome on the runner, and fails the run if any test skips. The workflow uses
only GitHub's own actions, pinned to commits, with read-only permissions and no secrets.

Rationale: the checks ran only where someone remembered to run them, and the rendered contrast
check only where Chrome was installed, so a skip could pass for a pass on any other machine. With
the dataset built and Chrome present, every test can run, so a skip means something broke. The
`google` organization scans workflows for supply-chain risk; pinned first-party actions and no
write access give it nothing to find.

## testedBeforeDeploy

**A deploy builds only code whose tests pass**

Subject: `release.test_gate`
Goals: [goals.secure](../../../goals/goals.md#secure), [goals.honestData](../../../goals/goals.md#honestdata)

Building the container image runs every test except the spec checks, against the dataset the image
will carry, and a failing test stops the build. A deployer may bypass it only by putting a file
named `SKIP_TESTS` in the build context, and the build log then says the tests were skipped.

Rationale: deploying builds the image from source, so the build is the one step every deploy
passes through, by hand or automated. The spec checks read git history, which a build context
does not carry; they gate every change through
[checksOnEveryChange](#checksoneverychange) instead. A changed input under `data/` or
`third_party/` goes through the same gate, since the ingest runs inside the build and the tests
run on what it produced. The bypass exists for an emergency fix that must ship past an unrelated
failure, and is a file rather than a deploy setting because a setting persists on the service
and would silently skip the tests of every later deploy.

## smokeCheckAfterDeploy

**Every deploy is checked from outside**

Subject: `release.smoke`
Goals: [goals.openToEveryone](../../../goals/goals.md#opentoeveryone), [goals.shareableViews](../../../goals/goals.md#shareableviews)

`birdview-smoke` checks a running deployment as a reader reaches it: the home page is served
gzipped and never stored, and with a month named it carries a validator that revalidates; a
species page and a guide answer; `robots.txt` serves its policy; an unknown species is not found;
the boundary file is cached as immutable; and each other address named redirects permanently to
the canonical one. It fetches nothing larger than a page, and it runs after every deploy.

Rationale: the tests check the code, not the deployment; flags, domains, and the platform in front
can each break a page the tests never see, as a missing `-addr` would have. Egress is what the
site pays for, so a check run after every deploy must not cost what a crawler would.
