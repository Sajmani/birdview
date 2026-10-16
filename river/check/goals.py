#!/usr/bin/env python3
# Copyright 2026 Google LLC
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#     https://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.

"""Validate Go package layout, descriptive identifiers, and goal citations across the bundle.

Usage:  python3 river/check/goals.py

Verifies: requirementsCiteGoals, goPackageConventions

Structural failures checked:

  - requirements.md missing or lacking a `## Module` fenced block with `module <path>`
  - a .md file in goals/, gradients/, or requirements/sources/ missing `package <name>`
  - .md files in the same directory declaring different package names
  - a package whose directory is named differently from it
  - two directories declaring the same package name, which would make a bare prefix ambiguous
  - a goal, gradient, or requirement ID that is not a descriptive CamelCase Go identifier
  - duplicate IDs within a package
  - a gradient citing no goal in `Serves:`
  - a `goals.<id>` citation naming no declared goal

Counted but never failed: a requirement citing no goal, and a goal nothing cites. product.requirementsCiteGoals permits it
gradients.unjustifiedRequirements and gradients.unservedGoals measure them. A brownfield project
starts with every requirement unmapped; a goal added because intent moved starts with no
requirements serving it. Neither should have to invent the other to satisfy a checker.
  - a `goals.<id>` citation naming no declared goal in goals/

What this does NOT check is whether a requirement genuinely *serves* the goal it names. Service
is a semantic relation rather than a measurement, so no check can read it -- true even of the
goals that do have a gradient and a target. That judgment belongs to a person at the gate, and
requirementsCiteGoals says so. A green run here is not a validated one.

Paths resolve relative to this file, per relativeArtifactPaths, so the bundle can live anywhere.
"""

import pathlib
import re
import sys

BUNDLE = pathlib.Path(__file__).resolve().parent.parent
GOALS_DIR = BUNDLE / "goals"
GRADIENTS_DIR = BUNDLE / "gradients"
REQ_ROOT = BUNDLE / "requirements"
REQ_MD = REQ_ROOT / "requirements.md"
SOURCES_DIR = REQ_ROOT / "sources"

# An item heading is the identifier and nothing else, so its anchor is the identifier
# lowercased and survives any rewording of the title. The title is the bold line under it.
H2_ITEM = re.compile(r"^## ([A-Za-z]+)[ \t]*$", re.M)
H3_ITEM = re.compile(r"^### ([A-Za-z]+)[ \t]*$", re.M)
# A status marker may follow the title: **Legs are modeled from the tarsus** *(withdrawn)*.
TITLE_LINE = re.compile(r"\s*\*\*(.+?)\*\*(?:[ \t]+\*\([a-z ]+\)\*)?[ \t]*$", re.M)
# The superseded form, `## id — Title`, which would otherwise be skipped silently.
OLD_FORM = re.compile(r"^#{2,3} [A-Za-z]+ — .+$", re.M)
DESCRIPTIVE_GO_IDENT = re.compile(r"^[A-Za-z]+$")
PACKAGE_DECL = re.compile(r"^package\s+([A-Za-z][A-Za-z0-9]*)\s*$")
MODULE_BLOCK = re.compile(r"^## Module\s+```[^\n]*\n(.*?)^```", re.M | re.S)
MODULE_DECL = re.compile(r"^module\s+\S+\s*$", re.M)
GOALS_FIELD = re.compile(r"^Goals:\s*(.+)$", re.M)
SERVES_FIELD = re.compile(r"^Serves:\s*(.+)$", re.M)
GOAL_REF = re.compile(r"\bgoals\.([A-Za-z0-9_]+)\b")


def check_package_dir(directory, heading_re, field_re=None, expected_pkg=None):
    """Parse items in a package directory and return (items, problems)."""
    items, problems = [], []
    seen_ids = {}
    pkg_by_dir = {}
    # PROVENANCE.md is a record about a source, not part of its package (goPackageConventions).
    md_files = sorted(
        p for p in directory.rglob("*.md")
        if "vendor" not in p.relative_to(directory).parts and p.name != "PROVENANCE.md"
    )
    for path in md_files:
        rel = path.relative_to(BUNDLE)
        text = path.read_text(encoding="utf-8")
        first_line = next((line.strip() for line in text.splitlines() if line.strip()), "")
        pm = PACKAGE_DECL.match(first_line)
        if not pm:
            problems.append(f"{rel}: first non-empty line must be `package <name>`")
            pkg = None
        else:
            pkg = pm.group(1)
            # A package's directory carries its name, so a reader looks in the right place.
            if path.parent.name != pkg:
                problems.append(f"{rel}: declares package {pkg} in a directory named {path.parent.name}; "
                                f"a package's directory carries its name")
            if expected_pkg and pkg != expected_pkg:
                problems.append(f"{rel}: declares package {pkg}, expected {expected_pkg}")
            prev_pkg, prev_file = pkg_by_dir.get(path.parent, (pkg, rel))
            if pkg != prev_pkg:
                problems.append(f"{rel}: declares package {pkg}, but {prev_file} declares {prev_pkg}")
            else:
                pkg_by_dir[path.parent] = (pkg, rel)

        for old in OLD_FORM.finditer(text):
            problems.append(
                f"{rel}: `{old.group(0)[:60]}` uses the superseded heading form; the heading is the "
                f"identifier alone and the title goes on the bold line beneath it"
            )
        heads = list(heading_re.finditer(text))
        for i, m in enumerate(heads):
            ident = m.group(1)
            tm = TITLE_LINE.match(text, m.end())
            title = tm.group(1) if tm else ""
            if not tm:
                problems.append(f"{rel}: `{ident}` has no title; put it in bold on the line beneath the heading")
            if not DESCRIPTIVE_GO_IDENT.match(ident):
                problems.append(
                    f"{rel}: identifier `{ident}` must be a descriptive CamelCase Go identifier (letters only)"
                )
            key = (pkg, ident)
            if key in seen_ids:
                problems.append(f"{rel}: duplicate identifier `{ident}` in package {pkg} (also in {seen_ids[key]})")
            else:
                seen_ids[key] = rel

            cited = []
            if field_re is not None:
                end = heads[i + 1].start() if i + 1 < len(heads) else len(text)
                field = field_re.search(text[m.end():end])
                cited = GOAL_REF.findall(field.group(1)) if field else []
            items.append((str(rel), ident, title, cited))
    return items, problems


def check_unique_package_names():
    """No two directories may declare the same package name.

    River files carry no import block, so a qualified reference like `goals.noSilentBreakage`
    resolves by package name alone. Go does not need this rule because every Go file names its
    imports explicitly; River does, and descriptiveIdentifiers records why. Two directories
    claiming one name would make every bare prefix to it ambiguous.
    """
    problems, dirs_by_pkg = [], {}
    for path in sorted(BUNDLE.rglob("*.md")):
        if "vendor" in path.relative_to(BUNDLE).parts:
            continue
        text = path.read_text(encoding="utf-8")
        first_line = next((line.strip() for line in text.splitlines() if line.strip()), "")
        m = PACKAGE_DECL.match(first_line)
        if m:
            dirs_by_pkg.setdefault(m.group(1), set()).add(path.parent.relative_to(BUNDLE))
    for pkg, dirs in sorted(dirs_by_pkg.items()):
        if len(dirs) > 1:
            where = ", ".join(str(d) for d in sorted(dirs))
            problems.append(
                f"package {pkg} is declared in {len(dirs)} directories ({where}); "
                f"package names must be unique within a module so a bare prefix resolves"
            )
    return problems


def check_bundle():
    problems = check_unique_package_names()
    if not REQ_MD.exists():
        problems.append(f"{REQ_MD.relative_to(BUNDLE)}: missing requirements.md")
    else:
        block = MODULE_BLOCK.search(REQ_MD.read_text(encoding="utf-8"))
        if not block or not MODULE_DECL.search(block.group(1)):
            problems.append(
                f"{REQ_MD.relative_to(BUNDLE)}: missing `## Module` fenced block with `module <path>`"
            )

    goal_items, g_probs = check_package_dir(GOALS_DIR, H2_ITEM, expected_pkg="goals")
    grad_items, gr_probs = check_package_dir(
        GRADIENTS_DIR, H3_ITEM, field_re=SERVES_FIELD, expected_pkg="gradients"
    )
    req_items, r_probs = check_package_dir(SOURCES_DIR, H2_ITEM, field_re=GOALS_FIELD)
    problems.extend(g_probs + gr_probs + r_probs)

    goals = {ident: title for _, ident, title, _ in goal_items}
    cited_anywhere = set()

    unjustified = []
    for src, rid, title, cited in req_items:
        if not cited:
            # Permitted and counted, never a failure: product.requirementsCiteGoals, measured by
            # gradients.unjustifiedRequirements. A brownfield project starts with all of them.
            unjustified.append(f"{src}: {rid} ({title})")
            continue
        for g in cited:
            cited_anywhere.add(g)
            if g not in goals:
                problems.append(f"{src}: {rid} cites goals.{g}, which package goals does not declare")

    for src, grid, title, cited in grad_items:
        if not cited:
            problems.append(f"{src}: {grid} ({title}) cites no `goals.<id>` in Serves:")
            continue
        for g in cited:
            if g not in goals:
                problems.append(f"{src}: {grid} serves goals.{g}, which package goals does not declare")

    unserved = []
    for gid, title in sorted(goals.items()):
        if gid not in cited_anywhere:
            # Counted, never failed: gradients.unservedGoals. A goal added because intent moved
            # should not have to arrive with invented requirements attached.
            unserved.append(f"goals.{gid} ({title})")

    return goals, grad_items, req_items, problems, unjustified, unserved


def main():
    goals, grads, reqs, problems, unjustified, unserved = check_bundle()

    # No goals at all is a valid state -- a brownfield project adopting Requirements before it
    # has recovered its goals (product.requirementsCiteGoals). No goals *while goal files exist*
    # means the parser or the layout changed, which is a failure.
    if not goals and any(GOALS_DIR.rglob("*.md")):
        print(f"no goals parsed from {GOALS_DIR} -- parser or layout has changed", file=sys.stderr)
        return 1
    if not reqs:
        print(f"no requirements found under {SOURCES_DIR} -- parser or layout has changed", file=sys.stderr)
        return 1

    for p in problems:
        print(p, file=sys.stderr)
    summary = (
        f"{len(reqs)} requirement(s), {len(goals)} goal(s), {len(grads)} gradient(s), "
        f"{len(problems)} problem(s)"
    )
    if unjustified:
        summary += f", {len(unjustified)} citing no goal"
    if unserved:
        summary += f", {len(unserved)} goal(s) uncited"
    print(summary)
    return 1 if problems else 0


if __name__ == "__main__":
    sys.exit(main())
