#!/usr/bin/env python3
"""AC-004 - every requirement cited in decisions.md and acceptance.md exists.

Verifies: process integrity. Catches an ID that was renamed, or one invented in a
decision record and never written down as a requirement.
"""
import re, sys, pathlib

spec = pathlib.Path(__file__).resolve().parents[1]
defined = set()
for name in ('product.md', 'tech.md'):
    defined |= set(re.findall(r'^### ([PT]-\d+)', (spec / name).read_text(encoding='utf-8'), re.M))
crs = set(re.findall(r'^## (CR-\d+)', (spec / 'decisions.md').read_text(encoding='utf-8'), re.M))
acs = set(re.findall(r'^### (AC-\d+)', (spec / 'acceptance.md').read_text(encoding='utf-8'), re.M))

problems = []
for name, valid in (('decisions.md', defined | crs | acs), ('acceptance.md', defined | crs | acs)):
    text = (spec / name).read_text(encoding='utf-8')
    for rid in sorted(set(re.findall(r'\b([PT]-\d{3})\b', text))):
        if rid not in defined:
            problems.append(f'{name} cites {rid}, which no requirement defines')
    for cid in sorted(set(re.findall(r'\b(CR-\d{3})\b', text))):
        if cid not in crs:
            problems.append(f'{name} cites {cid}, which no decision record defines')

for p in problems:
    print('FAIL:', p)
if problems:
    sys.exit(1)
print(f'ok: {len(defined)} requirements, {len(crs)} decisions, {len(acs)} criteria; all references resolve')
