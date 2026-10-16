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

"""acceptance.everyRequirementCitesGoal - every active requirement cites a declared goal, and every goal is cited.

Verifies: the structural rules in river/process.md.

An item's heading is its identifier alone, with its title in bold beneath it
(decisions.descriptiveIdentifiers); a heading in the superseded `### P-001 — Title`
form is reported rather than skipped, since skipping it would let a requirement
quietly drop out of the check.

NOTE: this checks STRUCTURE only. Whether a requirement genuinely serves the goal it
names is not measurable and belongs to Gate 1 review. A green run here is not a
validated one.
"""
import re, sys, pathlib

root = pathlib.Path(__file__).resolve().parents[1]
ITEM = re.compile(r'^### ([a-z][A-Za-z]+)[ \t]*$', re.M)
OLD = re.compile(r'^#{2,3} [A-Z]{1,2}-\d{3} — ', re.M)

fail = []
for name in ('goals.md', 'product.md', 'tech.md'):
    for m in OLD.finditer((root / name).read_text(encoding='utf-8')):
        fail.append(f'{name} has a heading in the superseded form: {m.group(0).strip()}')

goals = set(ITEM.findall((root / 'goals.md').read_text(encoding='utf-8')))
cited, missing, items = set(), [], 0
for name in ('product.md', 'tech.md'):
    text = (root / name).read_text(encoding='utf-8')
    heads = list(ITEM.finditer(text))
    for i, m in enumerate(heads):
        block = text[m.end(): heads[i + 1].start() if i + 1 < len(heads) else len(text)]
        items += 1
        if re.search(r'^Status: \*\*Withdrawn', block, re.M):
            continue
        g = re.search(r'^Goals: (.+)$', block, re.M)
        if not g:
            missing.append(m.group(1))
            continue
        cited |= set(re.findall(r'\bgoals\.([a-z][A-Za-z]+)', g.group(1)))

if not goals or items < 10:
    fail.append(f'found {len(goals)} goals and {items} requirements; the parser or the layout has changed')
if missing:
    fail.append(f'requirements with no Goals: field -> {" ".join(missing)}')
if goals - cited:
    fail.append(f'goals cited by nothing -> {" ".join(sorted(goals - cited))}')
if cited - goals:
    fail.append(f'citations to undeclared goals -> {" ".join(sorted(cited - goals))}')

for f in fail:
    print('FAIL:', f)
if fail:
    sys.exit(1)
print(f'ok: {len(goals)} goals declared, all cited; every active requirement cites one')
