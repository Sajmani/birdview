#!/usr/bin/env python3
"""AC-001 - every active requirement cites a declared goal, and every goal is cited.

Verifies: the structural rules in spec/process.md.

NOTE: this checks STRUCTURE only. Whether a requirement genuinely serves the goal it
names is not measurable and belongs to Gate 1 review. A green run here is not a
validated one.
"""
import re, sys, pathlib

root = pathlib.Path(__file__).resolve().parents[1]
goals = set(re.findall(r'^### (G-\d+)', (root / 'goals.md').read_text(encoding='utf-8'), re.M))
cited, missing = set(), []

for name in ('product.md', 'tech.md'):
    for block in re.split(r'\n(?=### [PT]-\d+)', (root / name).read_text(encoding='utf-8')):
        m = re.match(r'### ([PT]-\d+)', block)
        if not m:
            continue
        if re.search(r'^Status: \*\*Withdrawn', block, re.M):
            continue
        g = re.search(r'^Goals: (.+)$', block, re.M)
        if not g:
            missing.append(m.group(1))
            continue
        cited |= set(re.findall(r'G-\d+', g.group(1)))

fail = []
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
