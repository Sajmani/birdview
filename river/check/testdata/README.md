# Test cases for River's checks

Each directory is one case: a small bundle under `bundle/`, and `want.txt` stating what a correct
implementation of River's checker reports for it. Every implementation runs every case — the
Python reference through `check/cases.py`, and any translation into a project's own language
through that language's test runner (`decisions.checksInRepoLanguage`).

`want.txt` holds one fact per line; a line starting with `#` says why the case exists.

| Line | Meaning |
| --- | --- |
| `exit 0` or `exit 1` | whether the checker passes |
| `requirements N`, `goals N`, `gradients N` | items found |
| `unjustified N`, `unserved N` | requirements citing no goal, and goals nothing cites |
| `problem CODE PATH` | one reported problem; the set must match exactly |

`PATH` is relative to the bundle, with forward slashes: a file, a directory for a problem about a
directory, `-` for a problem about the bundle as a whole, or `../AGENTS.md` for the instruction
file beside the bundle. Length limits count characters (Unicode code points), not bytes. A case with no count lines is one where the checker stops before counting.

**The code and the path are the contract; the message is not.** A translation may word its
messages however its language and audience suit, so long as it reports the same problems.

A translation copies this directory and records the River version it copied from, alongside the
pin its translated code records.
