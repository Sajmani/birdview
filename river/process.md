> Project copy of River's `process.md`. Canonical origin: `sso://user/sameer/river`, pin `a1f29c2`.
> Local amendments, if any, are listed in [bindings.md](bindings.md).

# River

**River is a software development methodology with four stages — Requirements, Implementation,
Verification, Release — that run continuously rather than in sequence.** It keeps intent in
version control, next to the code, in a form an agent can read and a person can argue with.

This file is the method. It is identical in every project that uses River, so it says nothing
about any particular project; what is true of yours goes in `bindings.md`.

---

## Adopting River

**You do not have to change any tool you already use.** River is files in your repository plus
checks that run in whatever already runs your tests. If adopting a methodology requires
adopting a toolchain first, you have been handed a migration, not a method.

**Adopt one stage at a time.** Requirements alone is worth doing. A stage you have not adopted
is declared **informal** in `bindings.md` rather than left unmentioned, so a partial adoption is
distinguishable from an abandoned one.

One dependency is unavoidable: **Verification depends on Requirements**, because a criterion
with nothing behind it has nothing to check against. The others are free.

**Brownfield is the normal case.** A codebase that already exists has requirements — they are
implicit in the code, the tests and the documentation. Recover them first and map them to goals
afterwards. You will start with every requirement unmapped and a test suite nobody wrote
requirements for, and neither is a failure state. Both are counted.

### Copying this file

Copy `process.md` into your project unchanged and record where it came from:

```markdown
> Project copy of River's `process.md`. Canonical origin: <url>, pin `<commit>`.
> Local amendments, if any, are listed in `bindings.md`.
```

Keeping the file byte-identical to its origin is what makes refreshing it a reviewable diff
instead of an archaeology exercise. Put amendments in `bindings.md` and say what they change.

**The method is not a requirements source.** It is a pinned external document, but it
contributes no requirements to your composed set, and its pin belongs in the banner above and
nowhere else — recording it in the source manifest as well gives you two records of one pin,
which is the thing a single recorded hash exists to prevent.

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

The stages are not phases. They run continuously, and a project is normally doing all four at
once on different changes.

---

## The bundle

```
river/
  process.md                 this file
  bindings.md                how THIS project works
  implements.md              what is actually built
  goals/                     package goals
  gradients/                 package gradients
  check/                     the checks, in your language
  requirements/
    requirements.md          module declaration and precedence
    decisions.md             how each conflict and judgment was settled
    sources/                 RESERVED
      <name>/                one source, one package
      vendor/                vendored imported modules
    internal/                RESERVED — unexported composed packages
    <name>/                  exported composed package, and its criteria
```

**Put the bundle in `river/` at the repository root**, unless you have a reason not to; if you
put it elsewhere, say where in `bindings.md`. Paths inside it are relative to its root, and your
agent instruction file (`AGENTS.md` or equivalent) points at `process.md` — that pointer is how
an agent finds the method at all.

**Everything is plain Markdown.** No database, no server, no format that needs a tool to read.
Checks are written in a language the project already builds with, so adopting River adds no
runtime.

---

## Goals

A goal is an **objective**: why the software exists. It is what a requirement is in service of,
not the requirement restated in looser words.

**A goal need not be measurable, but it may be.** Measurement enters through gradients.

**What separates a goal from a requirement is attainment, not measurability.** Ask whether a
single change could make it true on its own. *p99 under 200 ms* could, so it is a requirement.
*Reach 10,000 weekly users* could not, however good the change, so it is a goal. Revenue is the
clearest case: you do not require it in order to ship, and only by shipping can you earn it.

That is why **a goal never gates a change and a requirement can.**

Requirements cite goals by package-qualified ID in a `Goals:` field. A goal never lists its
requirements, so the two cannot drift apart.

A goal nothing cites yet, and a requirement citing no goal, are both **counted, not rejected**.
A brownfield project starts with the second; a goal added because the market moved starts with
the first. Requiring either to arrive with the other attached produces invented ones.

---

## Gradients

A **gradient** is a metric with a declared **direction of improvement**. The direction is the
whole of the difference from a KPI: a KPI reports where you are, a gradient says which of two
candidates is better — which is the operation an agent needs when it has produced several.

A gradient may carry a **target**, a **backstop**, both, or neither.

```
  worse  ├──────────────●───────────→──────────┤  better
      Backstop        current               Target
```

- A **target** is absolute: where you are trying to get to. A gradient with one is an SLO.
- A **backstop** is a fixed limit the value must not cross. It does not move when you
  improve: reach 90 ms against a 100 ms backstop and the backstop stays at 100 ms, and the 10 ms
  between is slack you may spend elsewhere, as an error budget is spent. Performance budgets are
  backstops.
- **Both** brackets the work: *reach 100 ms, fail above 300 ms, improve in between.*

**The kind of threshold does not decide whether you have written a goal or a requirement** —
attainment does. Either kind can produce either, though backstops are usually requirements,
because every change can be checked against one.

**A ratchet is not a backstop.** A ratchet moves the limit to each new best, which throws the
slack away. It is occasionally the right policy, but moving a backstop is a decision, and a
person's to make.

**Most gradients carry no threshold, and should not.** Where several exist, nearly every change
regresses one of them; thresholding all of them stops delivery, and what follows is not stricter
engineering but a team that has learned to override gates. Online-experiment practice calls a
thresholded metric a **guardrail metric** and warns against having many, for the same reason.

**Achieving a target may promote it to a backstop.** A target gives direction until you reach
it and then gives none; converting it holds the position while effort moves to goals still
unmet. That is what makes a set of goals prioritizable — an agent reads the achieved ones as
constraints and the rest as direction. Promotion is a person's call: it needs a firing rule or
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
diff rather than a surprise. Record the hash in exactly one place.

**Vendoring is one way to pin a source and is often not available.** A licence may forbid
redistribution, terms may forbid automated access, or the source may be an API with nothing to
copy. Record an **import style** per source — vendored, fetched at build, queried, quoted,
absorbed, consulted — and pin it by whatever the style allows. What is constant is that a
change is detectable, not the mechanism.

**A source you cannot retrieve is a recorded gap, never a reconstruction.** Terms pages refuse
automated clients; publishers go offline. Write down what could not be read and leave it
unread. Never supply its contents from memory — a plausible rate limit or licence term is
indistinguishable from a real one, and an agent will produce one readily.

**A source has a lifecycle, and refusal is part of it.** A source you evaluated and
**rejected** is kept, with the reason: the refusal is a durable decision, usually worth more
than the source would have been, because otherwise the next person evaluates it again. A
source you adopted and later **retired** is kept too, with what depended on it. Neither is the
same as a source nobody considered.

**One source is one package.** A source may span several files in its directory without becoming
several packages; files organize, packages namespace. If a grouping deserves a prefix it
deserves to be a package, and if it does not, it is a file.

### Precedence

`requirements.md` ranks sources in tiers, highest first. Composition uses the ranking to resolve
a conflict mechanically; anything it cannot resolve goes to a person. A **Governing** tier
marks a source you may not locally override.

### Composition

**Composition** applies precedence and recorded decisions to your sources and produces one or
more **composed requirements packages** — explicit, versioned artifacts rather than an implied
state. Each records which sources it came from, at which pins, under which precedence order.

A composed package is named for its **audience or scope**, the way a Go package is. Never for a
version: versions are tags and branches, so a directory called `v1/` is the one name that is
always wrong. Once tagged, a composed version is never edited; a later composition makes a new
tag, and simultaneous milestones live on branches.

Downstream repositories import your **exported composed packages**, never your raw sources, so
every cross-repository obligation has been through composition with explicit precedence and
human sign-off.

### Conflicts

**Never resolve a conflict silently.** Two requirements that disagree go to `decisions.md`:
what conflicted, which sources, what was decided, by whom, and on what authority. Records are
append-only — a decision that turns out badly is superseded by a new one, never edited, because
the record of having been wrong is worth keeping.

### Anatomy

```markdown
## noSilentFailures

**Errors surface to the caller**

Subject: `api.errors`
Goals: `goals.trustworthyApi`

A swallowed error is indistinguishable from success, and the caller acts on a lie.
```

Identifiers are descriptive CamelCase — `noSilentFailures`, not `P-014`. Inside a package its
own names are bare; everywhere else they are qualified, `product.noSilentFailures`. Package
names are unique within a module, which is what lets a bare prefix resolve without an import
block.

**The heading is the identifier and nothing else; the title is the bold line beneath it.** A
heading's anchor is derived from its text, so this makes every item's anchor its identifier
lowercased — `product.md#nosilentfailures` — and lets the title be reworded without breaking
a single link to it. Write link fragments in lowercase; anchors are generated that way and
matched exactly. Decisions follow the same form.

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
Equally: a criterion's expected value comes from the requirement, never from a run of the
current code — a check built from existing output can only confirm that behavior has not
changed, including when it was wrong.

### When a requirement cannot be verified

Verification reports a requirement it cannot turn into a deciding criterion. It is one of three
things, and telling them apart is judgment:

1. **Underspecified.** Vague, not undecidable. *Responsive* becomes *p99 under 200 ms*.
2. **Exact but undecidable.** *Never leaks PII* is precise and cannot be established — instances
   are testable, absence is not. Attach the coverage you can get, state in the criterion what it
   does **not** cover, and count the requirement as uncovered.
3. **Requires judgment.** *Complies with the spirit of the regulation.* The guardrail is a human
   review with a named owner and a cadence. Still a guardrail; simply not automated.

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
a guardrail in exactly this sense, and citing it is the whole of the integration.

The citation does not live in the composed package. A criterion is immutable with its version;
the guardrail implementing it is operational, and swapping one tool for another must not force a
new requirements version. Cite guardrails from a mutable, descriptive record.

**A guardrail may have no criterion, and is never removed for lacking one.** A brownfield
project begins with a suite full of orphans. An orphan is one of three things and you cannot
tell which without looking: an undocumented requirement, a test of an implementation detail, or
a test of behavior nobody wants any more. The default is **keep and count**. A method that
rewards deleting unexplained tests has made the codebase less safe in exchange for a tidier
table.

---

## Implementation

Implementation produces the code and the **designs that describe it**, generated together and
stored at the same commit. A design describes the code at the commit it sits in — not a record
written afterwards, not a plan written before. It is one half of an implementation, kept in step
with the other, so an agent can read it as grounding context.

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
requirement. Say which — only one of them answers to an acceptance criterion.

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
the source, or on the goal, and composition runs again.

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
at all, which is the most common and the easiest to get wrong in the direction of writing
everything down.

**Unsatisfiable** has three dispositions: weaken it to what is achievable, withdraw it, or
**accept the gap** — keep the requirement and record in `decisions.md` that it is not met, by
whom, on what grounds, and when it will be looked at again. An accepted gap with no review date
is indistinguishable from a forgotten one.

**Release feedback is not a weaker verification.** A guardrail catches only the violation of a
criterion that exists, so everything outside the written set is invisible to it by construction
— and that is where most of what a running system teaches you lives. Bug reports, feature
requests, UX studies, traces, profiles and exceptions are the one channel that reaches the
unknown. It is also the only feedback that is not always about a defect: the answer to an
opportunity may be a new **goal**.

**Goal movement has three shapes and only one invalidates anything.** A goal **added** needs new
requirements and touches none. A goal **expanded** means the requirements serving it still serve
it and no longer suffice. A goal **changed or withdrawn** puts every requirement citing it in
doubt. Treating all movement as invalidation makes teams reluctant to move goals at all.

**Revision is not composition.** Revision edits a source or a goal; composition then produces a
new requirements version from it.

---

## Who decides

**An agent drafts; a person decides.**

An agent transcribes sources, drafts requirements, surfaces conflicts, generates criteria, runs
guardrails, clusters unmapped requirements and proposes candidate goals.

A person decides whether those are the right requirements, which one loses a conflict, whether a
gap may be accepted, and whether a target should become a backstop.

The division is by accountability, not capability. **A criterion an agent both wrote and
approved verifies nothing**, because the agent will simply have satisfied itself.

---

## Checks

Run them before calling a change done. At minimum a bundle should check that:

- every requirement's goal citation resolves, and every criterion's requirement citation does
- identifiers are well formed and unique within their package
- a pinned source still matches its recorded hash, by whatever means its style allows
- counts that are claimed are computed, not maintained by hand

**Every check must have been watched failing.** Plant a violation, see it reported, remove it.
A check nobody has seen fail is a check nobody has established.

What a check cannot do is tell you whether a requirement genuinely *serves* the goal it cites.
Service is a semantic relation, not a measurement. A green run is not a validated one.

---

## What is new here, and what is not

River assembles a great deal of existing practice and says so: requirements management,
acceptance criteria, CI gates, progressive rollout, SLOs and error budgets, guardrail metrics,
dependency pinning, and Go's package and module conventions, which it adopts directly rather
than inventing syntax.

What differs is mostly joinery. Requirements are composed from many sources with explicit
precedence and a durable record of every resolution. Criteria are generated from requirements
and cite them, so coverage is derived rather than asserted. Gradients make "better" orderable so
an agent can choose between candidates. And the whole of it is files in your repository, legible
to a person and loadable as grounding context by an agent.

---

## When this file is wrong

It will be. Write the amendment in `bindings.md`, say what it changes and why, and if the change
is general rather than local, send it upstream. A project copy that silently diverges from its
origin is worse than one that never refreshed, because the diff no longer means anything.
