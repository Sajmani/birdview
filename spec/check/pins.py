#!/usr/bin/env python3
"""AC-003 - every vendored input matches the hash recorded in its PROVENANCE.md.

Verifies: T-021, G-004.

Scope: committed inputs only. `wiley` is quoted rather than vendored and has no file
to hash; `aba` is fetched at build and verified by the ingest instead (AC-053).
"""
import re, sys, hashlib, pathlib

spec = pathlib.Path(__file__).resolve().parents[1]
repo = spec.parent
rows, problems = [], []

for prov in sorted(spec.glob('sources/*/PROVENANCE.md')):
    text = prov.read_text(encoding='utf-8')
    # table rows of the form | `//path` | origin | `sha256` |
    for m in re.finditer(r'\|\s*`//([^`]+)`\s*\|[^|]*\|\s*`([0-9a-f]{64})`\s*\|', text):
        rel, want = m.group(1), m.group(2)
        f = repo / rel
        if not f.exists():
            problems.append(f'{prov.parent.name}: missing file {rel}')
            continue
        got = hashlib.sha256(f.read_bytes()).hexdigest()
        rows.append((prov.parent.name, rel, got == want))
        if got != want:
            problems.append(f'{prov.parent.name}: {rel}\n    recorded {want}\n    actual   {got}')

if not rows:
    print('FAIL: no vendored files found to check')
    sys.exit(1)
for name, rel, ok in rows:
    print(f'  {"ok  " if ok else "FAIL"} {name:14} {rel}')
for p in problems:
    print('FAIL:', p)
sys.exit(1 if problems else 0)
