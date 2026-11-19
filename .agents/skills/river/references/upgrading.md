# Upgrading River

Part of River's method. Search `^## ` for the sections; read only the one you need: *Every
upgrade*, then the section for each step `river upgrade` says it ran.

## Every upgrade

`river upgrade` takes a module from the version it declares to the latest. Upgrading across several
versions runs each step below in order, in one run; the report names the steps it ran, and each
has a section here.

1. **See the plan:** `river upgrade --check` from the repository root writes nothing and prints
   what it would change and what it would leave for you.
2. **Upgrade on its own branch,** apart from other work, and open its own pull request:
   `river upgrade`. It makes every edit a tool can, refreshes the skill copy and its pin, declares
   the latest version, and commits nothing. The module declares the new version even while
   failures remain: finish them on the same branch.
3. **Work through "needs a person, after upgrading".** Every item says where it is and, after
   `To do:`, what to do. The sections below explain each kind.
4. **Recompose** every composed package whose inputs or files you changed, then run `river sum -w`
   and `river compose <package>` ([`composition.md`](composition.md), *Procedure*).
5. **Run `river check` until it reports no problems.**
6. **Open the pull request** with its *Requirements impact* section
   ([`pull-requests.md`](pull-requests.md)). An upgrade's format changes move no requirement; put
   anything you decided along the way (a package recomposed differently, a criterion composed, a
   guardrail let go) under *Needs your call*.

## 1.x to 2.0

**What changes.** A module declares its River version in a `## Module` block in `README.md` at its
root. The CLI writes `requirements/sums.md`. A composed package names its sources under
`## Composed from`, and marks each excluded item `*(excluded)*` after its link with the decision
that excluded it. 1.x's notes for unaccounted items and unresolved references become failures.

**What `river upgrade` does for you.** It moves the `## Module` block to `README.md`, adds
`## Composed from` from what each composed package accounts for, adds the `*(excluded)*` markers,
moves `Approves:` entries out of `decisions.md`, and writes `sums.md`. A 1.x module then continues
through *2.x to 3.0* in the same run.

**What's left for you.** Problems the upgraded module still has, each with its `To do:`:

- **An unaccounted source item** (`unaccounted-requirement`): recompose the package, including,
  composing from, or excluding it.
- **An excluded item with no decision** (`excluded-without-decision`): link the decision that
  excluded it, or ask the owner and record one.
- **A reference that names nothing or is not a link:** fix the name, or link it to its definition.

## 2.0 to 2.1

**What changes.** River 2.1 added goals that serve goals, imported goals, and goal versions. River
3.0 later left goals dormant, so this step is passed through: a 2.0 module upgrades straight through
*2.x to 3.0*, which this River runs for 2.0 and 2.1 alike.

**What `river upgrade` does for you.** Nothing of its own; see *2.x to 3.0*.

**What's left for you.** Nothing of its own.

## 2.x to 3.0

**What changes.** Source control approves: `requirements/approvals.md` goes, and each composed
package carries a `composed.sum` stamp that `river compose` writes. Decisions record no
`### Sums`. Goals are dormant and gradients set aside. The decisions, precedence, imports,
conformance, and bindings packages become directories. Open work (`todo`, `proposals.md`) leaves
the module.

**What `river upgrade` does for you.** It stamps each composed package whose 2.x approval was
current, deletes `approvals.md`, removes every `### Sums` block from decisions, moves each one-file
built-in package into its directory with its package line, moves open work to the repository
root, rewrites links to what moved, and rewrites `sums.md`.

**What's left for you.**

- **A package left unstamped:** its 2.x approval was stale or missing, so the upgrade could not
  vouch for it. Recompose it from its sources, then run `river sum -w` and `river compose <package>`.
- **A built-in file not moved,** because its directory already holds the target file: merge the two
  by hand, keeping the `package` line, delete the old file, then run `river links -w`.
- **Open work not moved,** because the repository root already has a file by that name: merge them,
  then delete the module's copy. River 3.0 doesn't check open work.
- **Problems the upgraded module still has,** each with its `To do:`, such as
  `composition-unstamped` or `stale-composition`: recompose, then `river compose <package>`.

## 3.0 to 3.1

**What changes.** River 3.1 is a minor version: a module declaring `river 3.0` that passed still
passes, and 3.1's new failures are notes for it. Declaring `river 3.1` makes them failures:

| Code | What it holds you to |
| --- | --- |
| `verifies-not-criterion` | A test's `Verifies:` comment names criteria, never requirements ([`verification.md`](verification.md)) |
| `guardrail-missing` | A guardrail a conformance entry names as `path:TestName` exists |
| `criterion-unguarded`, `guardrail-dropped` | A criterion that blocks unless met has a guardrail; a met one keeps it |
| `blocks-malformed`, `implementation-malformed`, `implementation-commit-unknown` | `Blocks:` and `Implements:` are well formed and resolve |
| `horizontal-only-narrowed` | A horizontal requirement is stated over its class, not only in narrowed items ([`composition.md`](composition.md), *Projecting horizontal requirements*) |
| `unaccounted-requirement`, `applies-to-routing` | Every composed package accounts for every source item; `Applies to:` names only a class, never a package ([`composition.md`](composition.md), *Accounted for*) |
| `decision-undecided` | A decision `river upgrade` drafted is confirmed by its owner |

**What `river upgrade` does for you.** It converts routing: each composed package excludes the
source items its 3.0 routing sent only to other packages, in one group under `## Excluded`
headed `### Excluded by [decisions.<package>Scope](…)`, and it drafts that decision in
`requirements/decisions/` with `Decided by: UNDECIDED`. It removes routing values (package names,
and what a composed package binds) from `Applies to:` on local items, and from interpretation
decisions as a format migration that keeps their classes and content. It re-stamps each package
whose stamp was current, and reports the counts. It also repoints each test's `Verifies:` name that exactly one
in-force criterion covers, rewrites `Verifies: <requirements>. Criterion: <criteria>.` (one line)
as `Verifies: <criteria>`, and declares `river 3.1`. Production code's `Verifies:` comments may keep
naming requirements; it lists them for information only.

**What's left for you.**

- **A drafted scope decision** (`decision-undecided`): ask the owner to confirm it and write their
  name in `Decided by:`. If an excluded item does govern the package, move it out of the group and
  include it or compose from it, then recompose and run `river compose <package>`.
- **An item no package accounted for** (`unaccounted-requirement`): it was a gap before, not
  routing; recompose each package to include, compose from, or exclude it.
- **A routing value in an imported item's `Applies to:`:** imported text is not edited; record the
  class it should name in an interpretation decision, and exclude the item where it doesn't apply.
- **A requirement several criteria cover:** the item lists them. Read the test, replace the
  requirement with the criterion it checks, or name each if it checks several.
- **A requirement no criterion covers:** compose a criterion covering it first
  ([`composition.md`](composition.md), step 6), then name it in the test.
- **`Criterion:` on another line:** rewrite the comment as `Verifies: <criteria>`.
- **`guardrail-missing`:** repoint the entry's `Guardrails:` at the test that implements the
  criterion now, or restore the test.
- **`criterion-unguarded`:** write the guardrail, name it, watch it fail, and record
  `Watched failing:`; or ask the owner whether the criterion may block once met.
- **`guardrail-dropped`:** restore the guardrail; or, if it is gone on purpose, ask the owner,
  record why in the conformance entry, and say so in the pull request.
- **`horizontal-only-narrowed`:** the item names the class and the rules composed from the
  horizontal item. If one of those rules already states it over the whole class, add
  `Applies to: <class>` to that rule. Otherwise include the item unchanged, or compose a
  class-level rule carrying the same `Applies to:`. Then add a class-level criterion whose
  guardrail discovers the instances from the system.
- **`blocks-malformed`, `implementation-*`:** fix the field; `river implement <package>`
  writes `Implements:`.
