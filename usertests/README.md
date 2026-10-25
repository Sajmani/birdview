# Simulated users

Agents that each play one reader, given only BirdView's pages and a few tasks, so a design can be
tried before real readers see it ([simulatedUserTests](../river/todo.md#simulatedusertests)).

- [personas.md](personas.md): who the readers are, one per heading.
- [tasks.md](tasks.md): what each is asked to do. This is all a reader is shown.
- [answers.md](answers.md): the expected answers, for scoring; never shown to a reader.
- [prompt.md](prompt.md): the prompt each agent is given, with the persona, the address and the
  tasks filled in.
- `tools/page URL [--open ID] [--find TEXT] [--tiles N]`: a page as text, with collapsed panels
  collapsed and links numbered, as a reader sees it. `--find` is find-in-page.
- `tools/shot URL OUT.png [phone|laptop] [screens]`: a screenshot at phone (390 px) or laptop
  (1280 px) width.
- `results/<date>/<version>/<persona>.md`: one report per run, and `results/<date>/summary.md`.

**Rules for a reader.** Only the two tools, only the address it is given, and no repository files:
a reader knows nothing of the code, the requirements or this directory beyond its own prompt. It
reaches a page the way a person would, by following links or typing an address it has seen on a
page; typing into the search box is the same as adding `q=` to the address.

**Calibrating.** Before a run is used to compare designs, the same readers run against the
revision real readers saw (`4f25193`, the first readers' feedback of 2026-10-25), and their
reports are checked against what those readers said: dense and intimidating, not made for
phones, no way to search.

**Running.** Serve each revision on its own port (the current tree on 8080; an old revision from a
`git worktree`, built and started with `-addr localhost:8081 -data web/data/species.json`), then
give each agent [prompt.md](prompt.md) filled in. Score each report against
[answers.md](answers.md): per task, correct, partly correct or wrong; the confidence given; and
whether the reader found the answer on the page or guessed.
