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

"""Render the Mermaid diagrams in docs/*.md to SVG, for BirdView's own docs page.

GitHub draws a ```mermaid block straight from the Markdown. BirdView's server
cannot: drawing Mermaid in the browser means loading Mermaid's script, and
T-002 forbids the server needing the network -- vendoring three megabytes of
JavaScript to draw one picture is the wrong trade. So each diagram is
rendered once, here, to a static SVG committed beside its page, and the
server shows that.

The SVG records the SHA-256 of the source it was drawn from, and
TestDiagramsAreCurrent fails when the two disagree. Edit the Markdown,
run this, commit both.

Needs Node and Google Chrome; Mermaid's CLI is fetched by npx and drives the
installed Chrome rather than downloading its own:

    python3 docs/render-diagrams.py

Labels are plain SVG text rather than HTML, because a browser shows an SVG
loaded through <img> without its foreignObject content in some engines.
"""
import hashlib, json, os, pathlib, re, subprocess, sys, tempfile

DOCS = pathlib.Path(__file__).resolve().parent
CHROME = os.environ.get("CHROME",
    "/Applications/Google Chrome.app/Contents/MacOS/Google Chrome")
BLOCK = re.compile(r"```mermaid\n(.*?)\n```", re.S)


def render(md: pathlib.Path) -> None:
    blocks = BLOCK.findall(md.read_text(encoding="utf-8"))
    if not blocks:
        return
    if len(blocks) > 1:
        sys.exit(f"{md.name}: {len(blocks)} diagrams; one per page is all the server shows")
    src = blocks[0]
    digest = hashlib.sha256(src.encode("utf-8")).hexdigest()
    with tempfile.TemporaryDirectory() as tmp:
        t = pathlib.Path(tmp)
        (t / "in.mmd").write_text(src, encoding="utf-8")
        (t / "pptr.json").write_text(json.dumps(
            {"executablePath": CHROME, "args": ["--no-sandbox"]}))
        # markdownAutoWrap off, or Mermaid rewraps every label at a fixed
        # width and a schema's one-field-per-line layout breaks mid-field.
        (t / "cfg.json").write_text(json.dumps(
            {"htmlLabels": False, "markdownAutoWrap": False,
             "flowchart": {"htmlLabels": False, "wrappingWidth": 2000}}))
        env = dict(os.environ, PUPPETEER_SKIP_DOWNLOAD="1")
        subprocess.run(["npx", "-y", "@mermaid-js/mermaid-cli@11.6.0",
                        "-p", str(t / "pptr.json"), "-c", str(t / "cfg.json"),
                        "-b", "white", "-i", str(t / "in.mmd"), "-o", str(t / "out.svg")],
                       check=True, env=env)
        svg = (t / "out.svg").read_text(encoding="utf-8")
    # The digest goes first, as a comment, so the check needs no XML parser.
    out = md.with_suffix(".svg")
    out.write_text(f"<!-- source sha256 {digest} -->\n" + svg, encoding="utf-8")
    print(f"wrote {out.relative_to(DOCS.parent)} from {md.name} ({digest[:12]})")


for md in sorted(DOCS.glob("*.md")):
    render(md)
