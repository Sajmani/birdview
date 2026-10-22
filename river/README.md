> Project copy of River's `README.md`, pin `fc69e6c`, licensed under Apache 2.0 with the rest of BirdView.
> River is not yet published; its canonical origin will be recorded here when it is. Local amendments,
> if any, are listed in [bindings.md](bindings.md).

# River

**River is a software development methodology with four stages — Requirements, Implementation,
Verification, Release — that run continuously rather than in sequence.** It keeps intent in
version control, next to the code, in a form an agent can read and a person can argue with.

This file is the method. It is identical in every project that uses River, so it says nothing
about any particular project; what is true of yours goes in `bindings.md`.

---

## Adopting River

**You do not have to change any tool you already use.** River is files in your repository plus
checks written in your language and run by whatever already runs your tests.

**Adopt one stage at a time.** Requirements alone is worth doing. A stage you have not adopted
is declared **informal** in `bindings.md` rather than left unmentioned, so a partial adoption is
distinguishable from an abandoned one.

One dependency is unavoidable: **Verification depends on Requirements**, because a criterion
with nothing behind it has nothing to check against. The others are free.

**Brownfield is the normal case.** An existing codebase already has requirements, implicit in
its code, tests and documentation. Recover them first and map them to goals afterwards. You will
start with every requirement unmapped and a test suite nobody wrote requirements for; neither is
a failure state, and both are counted.

### Copying this file

Copy `README.md` into your project unchanged and record where it came from:

```markdown
> Project copy of River's `README.md`. Canonical origin: <url>, pin `<commit>`.
> Local amendments, if any, are listed in `bindings.md`.
```

Keeping the file byte-identical to its origin makes refreshing it a reviewable diff. Put
amendments in `bindings.md` and say what they change.

**The method is not a requirements source.** It is a pinned external document, but it
contributes no requirements to your composed set. Its pin belongs in the banner above and
nowhere else: recording it in `imports.md` too gives two records of one pin, which a
single recorded hash exists to prevent.

---

## The loop

Four stages, and four processes that act on the arcs between them.

| Stage | Produces |
| --- | --- |
| **Requirements** | What must be true, composed from sources, with acceptance criteria |
| **Implementation** | Code and the designs that describe it |
| **Verification** | Guardrails that check criteria |
| **Release** | A running system, ramped under rollout policy |

| Process | Where it acts |
| --- | --- |
| **Composition** | Sources → Requirements. Assemble, resolve conflicts, generate criteria |
| **Optimization** | Every stage. Reduce the cost of future change |
| **Experimentation** | Release → Goals. Align releases with intent |
| **Revision** | Back to Goals and Sources. Return what the other stages learned |

The stages are not phases: a project is normally doing all four at once, on different changes.

---

## The bundle

```
river/
  README.md                 this file
  bindings.md                how THIS project works
  implements.md              what is actually built
  goals/                     package goals
  gradients/                 package gradients
  check/                     the checks, in your language
  requirements/
    README.md                what is here, and the module declaration
    imports.md               every imported source, with its pinned version
    precedence.md            standing rules: which source outranks which
    decisions.md             log of how each conflict and judgment was settled
    local/<name>/            a source written here: one source, one package
    imported/<import path>/  another module's source, pinned, with PROVENANCE.md
    composed/
      exported/<name>/       a composed package others may import, and its criteria
      internal/<name>/       a composed package private to this repository
```

Each composed file says it is composed and from what, and is changed by recomposing, never by
hand. Where a composed package sits is the record of who may import it: moving it into
`exported/` exports it, and no other file lists exports. These names were chosen by testing candidate layouts on zero-context agents acting as new
engineers.

**Put the bundle in `river/` at the repository root**, unless you have a reason not to; if you
put it elsewhere, say where in `bindings.md`. Paths inside it are relative to its root, and your
agent instruction file (`AGENTS.md` or equivalent) points at `README.md` — that pointer is how
an agent finds the method at all.

**Everything is plain Markdown.** No database, no server, no format that needs a tool to read.

**Write every agent-read file to be read in part** (`product.seekableArtifacts`). An agent
searches and opens ranges; it should not have to read a file whole to find one item.

- **Keep the instruction file short** — under 6,000 characters. It is loaded in every session,
  so it holds commands, the rules whose breach costs a commit, and a map. No history.
- **One item per heading, the identifier as the heading**, its title in bold beneath. Then
  `grep -n -A1 '^## '` is the table of contents and `## <id>` finds one item. Keep no
  hand-maintained index.
- **No line over 500 characters** outside a code block. A paragraph in a table row is one search
  hit that cannot be read by range; give the item a heading instead.
- **Fixed field names** (`Status:`, `Owner:`) so a search can filter items without opening them.
- **Separate by lifecycle.** Closed items and background notes go in their own files, out of the
  path of the open ones.
- **Say how to search a long file** in one line at its top, and link rather than copy between
  files.
- **Link every reference to its definition.** A name such as `product.noSilentFailures` is written
  as a Markdown link to the heading that defines it, in prose and in `Goals:` fields alike, so a
  reader follows it in one step and a renamed item breaks a link the checker reports. Exported and
  imported packages, which move between modules, link only within their own directory.

### Keeping the bundle's files

What each file under `requirements/` must contain, so that anyone adding to it keeps it
consistent. **Read this before writing to any of them.** An existing entry is a good example of
the format, but it cannot show the rules below that only apply when something changes.

**`decisions.md`** — one entry per decision, oldest first, appended at the end. Each is
`## <id>`, a bold line stating the decision, then one field per line:

```
Date: 2026-10-07
Decided by: <person>
Status: accepted | amended by <id>, <id> | superseded by <id>
Supersedes: <id>     (only when it replaces an earlier decision in whole)
Amends: <id>, <id>   (only when it changes earlier decisions in part)
Affects: product.x   (the requirements it changes or settles, not ones it merely cites)
Expires: 2027-01-01  (only for an accepted gap: when it must be looked at again)
```

Then `### Decided`, the decision itself, before the reasons; `---` ends the entry. If no reason
was given, say so rather than supplying one. Other fields may follow these, and `Status:` may
carry a note after a semicolon. **Entries are append-only with one exception:** when a new entry
supersedes or amends an earlier one, change the earlier entry's `Status:` line to say so, and
nothing else in it, so that a reader of the old entry learns it is no longer the whole story.
Accepting a gap is a decision: `Affects:` names the unmet requirement, and `Expires:` is required.
To seek: `awk '/^## /{id=$2} /^Affects:.*\.offlineMode(,|$)/{print id}' decisions.md` lists what
touches a requirement, and `awk '/^## /{id=$2} /^Status: (amended|superseded)/{print id": "$0}'`
lists what is no longer in force as written.

**`imports.md`** — one entry per imported source, `` ## `<name>` ``, then a field table: Name,
Origin, Version, Integrity, Tier, Import style (`vendored`, `fetched at build`, `queried`,
`quoted`, `absorbed`, `consulted`), Scope, Adopted parts, Owner, Status. The version is recorded
here and nowhere else. **Status** is `adopted` when every part named in Adopted parts is in force,
`partial` when some named part is not yet (not transcribed, not pinned, not checked), `rejected`
when evaluated and declined, and `retired` when adopted and later removed; adopting one criterion
of a large standard is a narrow Adopted parts, not `partial`.

**`imported/<import path>/`** — one directory per imported source, at its import path: the host
that publishes it, then the path naming the publisher's project, ending in the package name, the
way a Go import path does: `github.com/OWASP/asvs`, `w3.org/wcag`. A source known by a DOI uses
the publisher behind it (`figshare.com/avonet`), not the resolver. The directory holds
`<name>.md` (first line `package <name>`, then the transcribed items, each `## <Identifier>` with
a bold title) and `PROVENANCE.md` (each file's path, origin and SHA-256, and how and when it was
retrieved). Mark anything not yet retrieved as such; never fill it in from memory. The source's
tier goes in `precedence.md`.

**`precedence.md`** — one table of tiers, highest first, naming the sources in each. A source is
added to a tier when it is imported.

**Composed packages** — each opens by saying what it is composed from and that it is changed by
recomposing. **Without a composition tool, recompose by hand:** rewrite the file from its sources,
`precedence.md` and `decisions.md`, in a commit that says it recomposed and why. A requirement
accepted as unmet stays listed, with the decision beside it, as an excluded one is (below):
`— <title>. Accepted unmet by decisions.<id>`, the decision linked.

---

## Goals

A goal is an **objective**: why the software exists. It is what a requirement is in service of,
not the requirement restated in looser words.

**A goal need not be measurable, but it may be.** Measurement enters through gradients.

**What separates a goal from a requirement is attainment, not measurability.** Ask whether a
single change could make it true on its own. *p99 under 200 ms* could, so it is a requirement.
*Reach 10,000 weekly users* could not, however good the change, so it is a goal.

That is why **a goal never gates a change and a requirement can.**

**A goal names who benefits and what they gain.** The usual beneficiaries are the software's
users, the communities it depends on, and the people who build it; a project with no goal for one
of them is usually missing one. A goal that names a mechanism is a requirement in disguise: it
should survive a rewrite of the implementation. One paragraph is enough.

Requirements cite goals by package-qualified ID in a `Goals:` field. A goal never lists its
requirements, so the two cannot drift apart.

A goal nothing cites yet, and a requirement citing no goal, are both **counted, not rejected**.
A brownfield project starts with the second; a goal added because the market moved starts with
the first. Requiring either to arrive with the other attached produces invented ones.

---

## Gradients

A **gradient** is a metric with a declared **direction of improvement**. The direction is the
only difference from a KPI: a KPI reports where you are; a gradient says which of two candidates
is better, which is what an agent needs when it has produced several.

A gradient may carry a **target**, a **backstop**, both, or neither.

```
  worse  ├──────────────●───────────→──────────┤  better
      Backstop        current               Target
```

- A **target** is absolute: where you are trying to get to. A gradient with one is an SLO.
- A **backstop** is a fixed limit the value must not cross. It does not move when you
  improve: reach 90 ms against a 100 ms backstop and the 10 ms between is slack you may spend
  elsewhere, as an error budget is spent. Performance budgets are backstops.
- **Both** brackets the work: *reach 100 ms, fail above 300 ms, improve in between.*

**The kind of threshold does not decide whether you have written a goal or a requirement** —
attainment does. Either kind can produce either, though backstops are usually requirements,
because every change can be checked against one.

**A ratchet is not a backstop.** A ratchet moves the limit to each new best, which throws the
slack away. It is occasionally the right policy, but moving a backstop is a decision, and a
person's to make.

**Most gradients carry no threshold, and should not.** Where several exist, nearly every change
regresses one of them; thresholding all of them stops delivery, and teams learn to override
gates. Online-experiment practice calls a thresholded metric a **guardrail metric** and warns
against having many, for the same reason.

**Achieving a target may promote it to a backstop.** A target gives direction until you reach
it and then gives none; converting it holds the position while effort moves to goals still
unmet. That is what makes a set of goals prioritizable: an agent reads the achieved ones as
constraints and the rest as direction. Promotion is a person's call. It needs a firing rule or
it flaps, a level below the target or it gates on the first ordinary dip, and a cap, because
ten achieved goals would otherwise become ten permanent gates.

Not every goal has a gradient. A fabricated measure is worse than none, because it looks like
evidence.

---

## Requirements

### Sources

A **source** provides requirements. A **local** source is authored in your repository. An
**imported** source is written by someone else — a compliance standard, a terms of service, a
contract — and is **pinned and integrity-checked**, so a change to it produces a reviewable
diff. Record the hash in exactly one place.

**Vendoring is one way to pin a source and is often not available.** A license may forbid
redistribution, terms may forbid automated access, or the source may be an API with nothing to
copy. Record an **import style** per source — vendored, fetched at build, queried, quoted,
absorbed, consulted — and pin it by whatever the style allows. What must hold is that a change
is detectable, whatever the mechanism.

**A source you cannot retrieve is a recorded gap, never a reconstruction.** Terms pages refuse
automated clients; publishers go offline. Record what could not be read and leave it unread.
Never supply its contents from memory: a plausible rate limit or license term is
indistinguishable from a real one, and an agent will produce one readily.

**A source has a lifecycle, and refusal is part of it.** A source you evaluated and
**rejected** is kept, with the reason: the refusal is a durable decision, and without it the
next person evaluates the source again. A source you adopted and later **retired** is kept too,
with what depended on it. Neither is the same as a source nobody considered.

**One source is one package.** A source may span several files in its directory without becoming
several packages; files organize, packages namespace. If a grouping deserves a prefix it
deserves to be a package, and if it does not, it is a file.

### Precedence

`precedence.md` ranks sources in tiers, highest first. Composition uses the ranking to resolve
a conflict mechanically; anything it cannot resolve goes to a person. A **Governing** tier
marks a source you may not locally override.

### Composition

**Composition** applies precedence and recorded decisions to your sources and produces one or
more **composed requirements packages** — explicit, versioned artifacts rather than an implied
state. Each records which sources it came from, at which pins, under which precedence order.

A composed package is named for its **audience or scope**, the way a Go package is, never for a
version: a version is a commit, changelist or revision, so a directory called `v1/` is always
wrong. Once published, a composed version is never edited; a later composition is a new version,
and simultaneous milestones live on separate lines of development. Identify a version by what
your version-control system never reassigns (a commit hash, changelist or revision number) and
check it by its content hash. A tag or label is a readable alias, worth protecting where your
system allows, but no system makes one immutable by default, so never trust it alone.

Downstream repositories import your **exported composed packages**, never your raw sources, so
every cross-repository obligation has been through composition with explicit precedence and
human sign-off.

**A composed package names everything that applies, and writes out only what composition made.**
Requirements that apply unchanged are a list grouped by area, one line each, linking the source
item and giving its title:

```markdown
### Diagram

- [product.scaleToggle](../../../local/product/diagram.md#scaletoggle) — Diagrams scale per section or globally, never per tile

## Excluded

- [tech.localNowPublicLater](../../../local/tech/posture.md#localnowpubliclater) — Local now, public later. Excluded by [decisions.publicDeploymentDeferred](../../../decisions.md#publicdeploymentdeferred)
```

What composition wrote — a rule over a domain, an intermediary's contract, a requirement a
higher tier changed — follows in full, each item naming its sources in `Composed from:`, every
entry a link to the item it names. **Every source requirement is included, excluded, or composed
from, in some package**; one that is none of these has silently dropped out. Included
requirements keep their source names. Do not commit a flattened copy with everything inlined
beside the sources: every search would find each requirement twice. If an importer or auditor
needs one, generate it when you tag.

**A requirement over every instance of something is composed once.** A source names the domain
a requirement applies to in an `Applies to:` field, in its own term — `Applies to:
wcag.WebPage`. You define what counts as an instance in a local requirement whose `Binds:` field
names that domain, and decide the edge cases there: is an error page a web page? The composed
package states the rule once, *every web page meets `wcag.ContrastEnhanced`*, and records any
excluded instance as an accepted gap. The local definition makes the source's term concrete and
**never narrows it**: if the source's own definition makes something an instance, leaving it out
is an exception, recorded instance by instance. **Nothing in the bundle lists the instances.**
Verification discovers them from the system, so a new page inherits its obligations by existing
and cannot escape by being left off a list. Discover from what the system declares — its routes,
its templates — as well as by following links: an instance nothing links to is still an instance.

**An intermediary carries a contract.** A stylesheet, a framework or a middleware that makes its
users meet a requirement states the terms: given an assumption about how it is used, it
guarantees the requirement, over a stated scope, and says whether its users may override it. The
stylesheet guarantees 7:1 text, given that a page takes its colors only from it; the framework
guarantees scrubbed request logs, given that nothing logs around it. A contract from another
repository is imported at a pin like any other requirement. Layered intermediaries each carry
their own, and the order they run in is part of the contract, checked where they are assembled.

### Conflicts

**Never resolve a conflict silently.** Two requirements that disagree go to `decisions.md`:
what conflicted, which sources, what was decided, by whom, and on what authority. Records are
append-only: a decision that turns out badly is superseded by a new one, never edited, because
the record of having been wrong is worth keeping. Only its `Status:` line changes, to point at
what replaced it (*Keeping the bundle's files*, above).

### Anatomy

```markdown
## noSilentFailures

**Errors surface to the caller**

Subject: `api.errors`
Goals: `goals.trustworthyApi`

Every error reaches the caller. None is logged and dropped, and none is replaced by a default
value.

A swallowed error is indistinguishable from success, and the caller acts on a lie.
```

**The first paragraph after the fields is the complete statement**: what must be true, readable
on its own. Rationale, history and examples follow it, so a reader who needs only the rule can
stop there. Keep the rationale, after the rule: it tells a future reader a constraint is
deliberate, and it holds the incident a good failing test is built from.

Identifiers are descriptive CamelCase — `noSilentFailures`, not `P-014`. Inside a package its
own names are bare; everywhere else they are qualified, `product.noSilentFailures`. Package
names are unique within a module, which is what lets a bare prefix resolve without an import
block.

**The heading is the identifier and nothing else; the title is the bold line beneath it.** A
heading's anchor is derived from its text, so this makes every item's anchor its identifier
lowercased — `product.md#nosilentfailures` — and lets the title be reworded without breaking
a single link to it. Write link fragments in lowercase; anchors are generated that way and
matched exactly. Decisions follow the same form.

Tell your agents to find one requirement by searching for `## noSilentFailures` and to read only
that section; a source file holding twenty requirements costs twenty times as much to read whole.

---

## Acceptance criteria

An **acceptance criterion** is a prose statement of what must be true. **Composition produces
criteria**, so they live in the composed package beside the requirements they decide — they
belong to that version, and an obligation is only *demonstrably* preserved if whoever imports it
also receives how to decide it.

**A criterion cites what it covers**, in a `Covers:` field. A requirement never lists its
criteria and no traceability table is kept; a stored reverse index drifts from what it indexes,
and this one is derivable.

A criterion citing no requirement is an error, not an orphan. A guardrail may predate the
bundle; a criterion cannot, because composition produces it from a requirement.

**A criterion must have been watched failing before it is trusted.** A check that cannot fail is
indistinguishable from one that passes, and it is worse than nothing because it reports safety.
Likewise, a criterion's expected value comes from the requirement, never from a run of the
current code: a check built from existing output can only confirm that behavior has not
changed, including when it was wrong.

### When a requirement cannot be verified

Verification reports a requirement it cannot turn into a deciding criterion. It is one of three
things, and telling them apart is judgment:

1. **Underspecified.** Vague, not undecidable. *Responsive* becomes *p99 under 200 ms*.
2. **Exact but undecidable.** *Never leaks PII* is precise and cannot be established — instances
   are testable, absence is not. Attach the coverage you can get, state in the criterion what it
   does **not** cover, and count the requirement as uncovered.
3. **Requires judgment.** *Complies with the spirit of the regulation.* The guardrail is a human
   review with a named owner and a cadence. It is still a guardrail, just not automated.

**Never remove a requirement for being unverifiable**, and **never substitute an easier
criterion.** *Must be accessible*, checked by a contrast checker, reports green on a product
nobody can use. A requirement counts as covered only when its criterion **decides** it.

---

## Verification and guardrails

A **guardrail** is the thing that runs: a test or checker on an implementation, a monitoring or
rollout rule on a release. Criteria say what; guardrails enforce it.

A guardrail either **gates** — the change does not proceed — or **informs**, meaning the result
is recorded and weighed. Which one may differ by stage: an expensive check can inform at code
generation and gate at release. Gating is a property of the guardrail, not a separate kind of
thing.

**Guardrails are not bundle artifacts.** They live wherever you already keep your tests and your
deployment configuration. The bundle **cites** them; it does not contain them, because requiring
otherwise would mean moving your tests as a condition of adoption. A scanner you already run is
a guardrail in this sense, and citing it is the whole of the integration.

The citation does not live in the composed package. A criterion is immutable with its version;
the guardrail implementing it is operational, and swapping one tool for another must not force a
new requirements version. Cite guardrails from a mutable, descriptive record.

**Check exhaustively what varies, and sample the instances.** A thousand pages rendered from one
template need every template checked, and every data-dependent input to the property checked in
full — not a thousand renders. Where an intermediary carries the requirement, three checks cover
it and none is per instance: the intermediary's **guarantee**, once; its **assumption**, for each
kind of user; and the **scope**, so the part it does not cover has its own criterion.

**Then check the result end to end anyway**, gating on a sample and informing on the full set.
The contract checks can all pass over a real violation when the contract is wrong — a stylesheet
check that reads `color` and never sees SVG text painted with `fill` — or when layers interfere,
as when a later stylesheet overrides a color or a middleware logs before the scrubber runs. The
end-to-end check tests the result regardless of how the work was divided, so the more layers
there are, the more it matters.

**A guardrail may have no criterion, and is never removed for lacking one.** A brownfield
project begins with a suite full of orphans. An orphan is one of three things and you cannot
tell which without looking: an undocumented requirement, a test of an implementation detail, or
a test of behavior nobody wants any more. The default is **keep and count**. Deleting
unexplained tests makes the codebase less safe in exchange for a tidier table.

---

## Implementation

Implementation produces the code and the **designs that describe it**, kept in step at every
commit. A design may lead the code: written and reviewed before the code exists, it says it is a
proposal, and it becomes the code's description in the commit where the code lands. From then on
a change that makes the two disagree updates both, and a guardrail can detect drift between them.
A design is never a record written afterwards. This follows the Elephant-Goldfish Model, which
reviews a design with zero-context agents before code is written and reports where they drift.

**A design lives with the code, not in the bundle.** The bundle holds intent and conformance;
keeping a design in it would put it where people edit requirements rather than where they edit
code, which is backwards for a document whose only job is to stay in step. `implements.md` cites
where the designs are.

---

## Release and Experimentation

Release produces a release version from an implementation version, verifies it against the
release guardrails, and ramps traffic according to rollout policy.

**Experimentation** aligns releases with goals using feature flags, A/B tests, impact analyses
and traffic shifts. It hill-climbs gradients toward their targets and rejects candidates that
hit a backstop.

---

## Optimization

**Optimization reduces the cost, time and risk of future changes** to Requirements,
Implementation, Verification and Release. It is continuous and belongs to no single stage. It
acts on four things: **structure** (refactoring, debt, reorganizing an overlarge source),
**tooling** at any stage, the **tech stack**, and **agent capability** — the skills and
instruction files an agent reads.

**Optimization is distinguished from Implementation by purpose, not by what it touches.** Both
change code. Implementation changes the system to satisfy a requirement; Optimization changes it
to make future satisfaction cheaper. Porting to a faster language is Optimization when it makes
later performance work tractable, and Implementation when it meets a stated performance
requirement. Say which: only one of them answers to an acceptance criterion.

**It acts on the bundle too.** Restructuring requirements so that writing code and guardrails
from them costs less — a composed package that lists rather than copies, a requirement whose
rule comes first — is Optimization: nothing required changes.

**It has no target.** Cost, time and risk are gradients with a direction and no finish line.

**It depends on Verification more than any other stage.** Restructuring is safe only because
guardrails hold behavior still while the structure moves. A project with no guardrails cannot
optimize; it can only rewrite and hope.

---

## Revision

**Revision** is the loop's return arc: what a later stage learns becomes a change to intent.
Without it a bundle records what someone believed once.

**Revision edits what was authored, never what was derived.** A source and a goal are written by
people; a composed package is produced from them and is immutable once tagged. An edit there is
discarded by the next composition and invisible to anyone who imported it. So feedback lands on
the source or the goal, and composition runs again.

Where the source is **imported** you cannot edit it. The change is either a local requirement
overriding it under precedence, or a request to whoever owns it — and under a Governing tier,
only the request.

| From | What comes back |
| --- | --- |
| **Implementation** | Newly discovered requirements; requirements that cannot be satisfied as stated |
| **Verification** | Requirements that cannot be verified |
| **Release** | Telemetry and product feedback |
| **Goals** | Intent moved |

**Newly discovered** is one of three things: a requirement, so add it; an existing requirement
that was too vague, so sharpen that one; or an implementation decision that is not a requirement
at all. The last is the most common, and the usual mistake is to write it down anyway.

**Unsatisfiable** has three dispositions: weaken it to what is achievable, withdraw it, or
**accept the gap** — keep the requirement and record in `decisions.md` that it is not met, by
whom, on what grounds, and when it will be looked at again, in its `Expires:` field. An accepted
gap with no review date is indistinguishable from a forgotten one.

**Release feedback is not a weaker verification.** A guardrail catches only violations of
criteria that exist, so it cannot see anything outside the written set, and that is where most
of what a running system teaches you lies. Bug reports, feature requests, UX studies, traces,
profiles and exceptions are the one channel that reaches the unknown. It is also the only
feedback that is not always about a defect: the answer to an opportunity may be a new **goal**.

**Goal movement has three shapes and only one invalidates anything.** A goal **added** needs new
requirements and touches none. A goal **expanded** means the requirements serving it still serve
it and no longer suffice. A goal **changed or withdrawn** puts every requirement citing it in
doubt. Treating all movement as invalidation makes teams reluctant to move goals at all.

**Realign by re-citing, not by editing.** When goals move, and from time to time regardless, cite
every requirement afresh rather than adjusting its old `Goals:` line, and mark the citations that
fit only by stretching. Weak fits that cluster are a missing goal. A goal cited by unrelated
requirements is a catch-all hiding them. Two goals cited by the same requirements are one goal.
Two goals that pull against each other are a tension to record, not a mistake: each new
requirement says which side it serves.

**Revision is not composition.** Revision edits a source or a goal; composition then produces a
new requirements version from it.

---

## Who decides

**An agent drafts; a person decides.**

An agent transcribes sources, drafts requirements, surfaces conflicts, generates criteria, runs
guardrails, clusters unmapped and weakly mapped requirements and proposes candidate goals.

A person decides whether those are the right requirements, which one loses a conflict, whether a
gap may be accepted, and whether a target should become a backstop.

The division is by accountability, not capability. **A criterion an agent both wrote and
approved verifies nothing**, because the agent will simply have satisfied itself.

---

## Checks

**Write them in your project's dominant language, and run them in its own test suite.** Checks
that run wherever your tests already run cannot be skipped and add no toolchain. River's
reference checks are in Python: use them as they are if your project has no better home for
them, or translate them into your language and record which River version you translated, so a
later change to River's checks is visible. What is normative is what a check asserts, not the
language it is written in. Write your own checks in the same language.

**Before River's checks replace yours, list what yours enforced.** River's checker counts some
things a project's own checks may have failed — a requirement citing no goal, say — because a
brownfield project may start that way. If your project was stricter, keep the stricter rule as
your own check; otherwise swapping checkers is how a rule disappears without anyone deciding it
should.

Run them before calling a change done. At minimum a bundle should check that:

- every requirement's goal citation resolves, and every criterion's requirement citation does
- identifiers are well formed and unique within their package
- a pinned source still matches its recorded hash, by whatever means its style allows
- counts that are claimed are computed, not maintained by hand
- the instruction file and every line in the bundle stay within `product.seekableArtifacts`' limits

**Every check must have been watched failing.** Plant a violation, see it reported, remove it.

**Some symptoms are notes, never failures:** a goal cited by one requirement, a goal whose
requirements all cite one other goal too, a goal longer than a paragraph, and a goal that contains
code or names another item. Each can be right. Each is where to look first.

What a check cannot do is tell you whether a requirement genuinely *serves* the goal it cites.
Service is a semantic relation, not a measurement. A green run is not a validated one.

---

## What is new here, and what is not

River assembles a great deal of existing practice and says so: requirements management,
acceptance criteria, CI gates, progressive rollout, SLOs and error budgets, guardrail metrics,
dependency pinning, Go's package and module conventions, which it adopts directly rather than
inventing syntax, and design-first development kept in step with the code, from the
Elephant-Goldfish Model.

What is new is mostly how these are joined. Requirements are composed from many sources with
explicit precedence and a durable record of every resolution. Criteria are generated from
requirements and cite them, so coverage is derived rather than asserted. Gradients make "better"
orderable so an agent can choose between candidates. And the whole of it is files in your
repository, legible to a person and loadable as grounding context by an agent.

---

## When this file is wrong

It will be. Write the amendment in `bindings.md`, say what it changes and why, and if the change
is general rather than local, send it upstream. A project copy that silently diverges from its
origin is worse than one that never refreshed, because the diff no longer means anything.
