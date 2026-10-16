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

Across every tracked text file outside third_party/:

  - every qualified reference -- goals.x, product.x, tech.x, decisions.x,
    acceptance.x, and a transcribed source's inatapi.X -- names an item its
    package defines;
  - every anchored link into the bundle lands on a heading that exists, where an
    item's anchor is its identifier lowercased (decisions.descriptiveIdentifiers);
  - a link whose text names an item lands on that item: the file that defines
    it, at its anchor;
  - no identifier in a superseded form survives anywhere: a letter code and
    three digits, or a source name, a slash and an R number.

A package is a directory of files declaring it. The local sources are product
and tech; a transcribed source under requirements/sources/<pkg>/ is its own
package, with exported names.
"""
import re, sys, pathlib, subprocess

bundle = pathlib.Path(__file__).resolve().parents[1]
repo = bundle.parent
SOURCES = bundle / 'requirements' / 'sources'
LOWER_ITEM = re.compile(r'^#{2,3} ([a-z][A-Za-z]+)[ \t]*$', re.M)
UPPER_ITEM = re.compile(r'^#{2,3} ([A-Z][A-Za-z]+)[ \t]*$', re.M)
HEADING = re.compile(r'^#{1,6} (.+?)[ \t]*$', re.M)

def slug(text):
    """A heading's anchor, as GitHub derives it; underscores inside words survive."""
    text = re.sub(r'[`*]', '', text.strip().lower())
    return re.sub(r'[^\w\- ]', '', text).replace(' ', '-')

# package -> {name: defining file}
packages = {
    'goals': sorted((bundle / 'goals').glob('*.md')),
    'product': sorted((SOURCES / 'product').glob('*.md')),
    'tech': sorted((SOURCES / 'tech').glob('*.md')),
    'decisions': [bundle / 'requirements' / 'decisions.md'],
    'acceptance': [bundle / 'requirements' / 'acceptance.md'],
}
exported = {d.name: [d / 'requirements.md'] for d in sorted(SOURCES.iterdir())
            if (d / 'requirements.md').exists()}
defined = {}
for pkg, files in list(packages.items()) + list(exported.items()):
    item = UPPER_ITEM if pkg in exported else LOWER_ITEM
    defined[pkg] = {}
    for f in files:
        for name in item.findall(f.read_text(encoding='utf-8')):
            defined[pkg][name] = f.resolve()

anchors = {p.resolve(): {slug(h) for h in HEADING.findall(p.read_text(encoding='utf-8'))}
           for p in bundle.rglob('*.md')}

EXTENSIONS = {'md', 'py', 'go', 'svg', 'json', 'tsv', 'html', 'css'}
QUALIFIED = re.compile(r'\b(goals|product|tech|decisions|acceptance)\.([a-z][A-Za-z]+)\b')
EXPORTED = re.compile(r'\b(' + '|'.join(exported) + r')\.([A-Z][A-Za-z]+)\b') if exported else None
LEGACY = re.compile(r'\b(?:G|P|T|AC|CR)-\d{3}\b|\b[a-z][a-z-]*/R\d+\b')
LINK = re.compile(r'\]\(([^)\s#]*)#([^)\s]+)\)')
CITATION = re.compile(r'\[`?([a-z]+)\.([A-Za-z]+)`?\]\(([^)\s]*)\)')

FROM_RIVER = {'river/README.md', 'river/check/goals.py'}
tracked = subprocess.run(['git', 'ls-files'], cwd=repo, capture_output=True, text=True).stdout.split()
problems, refs = [], 0
for rel in tracked:
    path = repo / rel
    if rel.startswith('third_party/') or not path.is_file():
        continue
    try:
        text = path.read_text(encoding='utf-8')
    except (UnicodeDecodeError, OSError):
        continue
    # Copied from River unchanged -- the method and River's checker -- so their
    # names are River's, not BirdView's.
    if rel not in FROM_RIVER:
        for rx in (QUALIFIED, EXPORTED):
            for m in (rx.finditer(text) if rx else ()):
                pkg, name = m.groups()
                if name in EXTENSIONS:
                    continue  # goals.md, refs.py: a file, not a reference
                refs += 1
                if name not in defined[pkg]:
                    problems.append(f'{rel}: {pkg}.{name} names nothing package {pkg} defines')
    # The checks have to describe the form they reject.
    if not rel.startswith('river/check/') and rel not in FROM_RIVER:
        for m in LEGACY.finditer(text):
            problems.append(f'{rel}: {m.group(0)} is a superseded identifier')
    if rel.endswith('.md'):
        for pkg, name, url in CITATION.findall(text):
            if pkg not in defined or name in EXTENSIONS:
                continue
            target, _, frag = url.partition('#')
            dest = (path.parent / target).resolve() if target else path.resolve()
            want = defined[pkg].get(name)
            if want is not None and (dest != want or frag != name.lower()):
                problems.append(f'{rel}: [{pkg}.{name}]({url}) does not land on it; link to '
                                f'{want.relative_to(bundle)}#{name.lower()}')
        for target, frag in LINK.findall(text):
            dest = (path.parent / target).resolve() if target else path.resolve()
            if bundle not in dest.parents or not dest.name.endswith('.md'):
                continue  # only links into the bundle are held to its anchors
            if frag not in anchors.get(dest, set()):
                problems.append(f'{rel}: link to {dest.relative_to(bundle)}#{frag} lands on no heading')

for p in problems:
    print('FAIL:', p)
if problems:
    sys.exit(1)
counts = ', '.join(f'{len(v)} {k}' for k, v in defined.items())
print(f'ok: {counts}; {refs} references resolve')
