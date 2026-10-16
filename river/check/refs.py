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

"""acceptance.decisionReferencesExist - every reference to a goal, requirement, criterion or decision resolves.

Verifies: process integrity. Catches a name that was renamed, one invented in a
decision record and never written down, and a link whose anchor lands nowhere.

Three checks, across every tracked text file outside third_party/:

  - every qualified reference -- goals.x, product.x, tech.x, decisions.x,
    acceptance.x -- names an item that its file defines;
  - every anchored link into river/ lands on a heading that exists, where an
    item's anchor is its identifier lowercased (decisions.descriptiveIdentifiers);
  - no identifier in the superseded P-001 / CR-001 form survives anywhere.
"""
import re, sys, pathlib, subprocess

spec = pathlib.Path(__file__).resolve().parents[1]
repo = spec.parent
KINDS = {'goals': 'goals.md', 'product': 'product.md', 'tech': 'tech.md',
         'decisions': 'decisions.md', 'acceptance': 'acceptance.md'}
ITEM = re.compile(r'^#{2,3} ([a-z][A-Za-z]+)[ \t]*$', re.M)
HEADING = re.compile(r'^#{1,6} (.+?)[ \t]*$', re.M)

def slug(text):
    """A heading's anchor, as GitHub derives it."""
    text = re.sub(r'[`*]', '', text.strip().lower())  # underscores inside words survive
    return re.sub(r'[^\w\- ]', '', text).replace(' ', '-')

defined, anchors = {}, {}
for kind, name in KINDS.items():
    text = (spec / name).read_text(encoding='utf-8')
    defined[kind] = set(ITEM.findall(text))
for path in spec.glob('*.md'):
    anchors[path.name] = {slug(h) for h in HEADING.findall(path.read_text(encoding='utf-8'))}

tracked = subprocess.run(['git', 'ls-files'], cwd=repo, capture_output=True, text=True).stdout.split()
EXTENSIONS = {'md', 'py', 'go', 'svg', 'json', 'tsv', 'html', 'css'}
QUALIFIED = re.compile(r'\b(goals|product|tech|decisions|acceptance)\.([a-z][A-Za-z]+)\b')
LEGACY = re.compile(r'\b(?:G|P|T|AC|CR)-\d{3}\b')
LINK = re.compile(r'\]\(([^)\s#]*)#([^)\s]+)\)')

problems, refs = [], 0
for rel in tracked:
    path = repo / rel
    if rel.startswith('third_party/') or not path.is_file():
        continue
    try:
        text = path.read_text(encoding='utf-8')
    except (UnicodeDecodeError, OSError):
        continue
    # process.md is River's method, copied in unchanged; its examples are River's.
    for m in (QUALIFIED.finditer(text) if rel != 'river/process.md' else ()):
        kind, name = m.groups()
        if name in EXTENSIONS:
            continue  # goals.md, refs.py: a file, not a reference
        refs += 1
        if name not in defined[kind]:
            problems.append(f'{rel}: {kind}.{name} names nothing {KINDS[kind]} defines')
    # The checks have to describe the form they reject.
    for m in (LEGACY.finditer(text) if not rel.startswith('river/check/') and rel != 'river/process.md' else ()):
        problems.append(f'{rel}: {m.group(0)} is a superseded identifier')
    if rel.endswith('.md'):
        for target, frag in LINK.findall(text):
            dest = (path.parent / target).resolve() if target else path.resolve()
            if dest.parent != spec or not dest.name.endswith('.md'):
                continue  # only links into the bundle are held to its anchors
            if frag not in anchors.get(dest.name, set()):
                problems.append(f'{rel}: link to {dest.name}#{frag} lands on no heading')

for p in problems:
    print('FAIL:', p)
if problems:
    sys.exit(1)
counts = ', '.join(f'{len(v)} {k}' for k, v in defined.items())
print(f'ok: {counts}; {refs} references resolve')
