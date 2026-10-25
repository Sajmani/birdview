# AGENTS.md

**BirdView** draws each bird species as a scaled stick figure so birders can compare real sizes. One
Go module: `birdview-ingest` builds the dataset, `birdview` serves it.

**The requirements in [river/](river/) are normative; where this file disagrees, `river/` wins.**
Names like `product.scaleToggle` are citations, linked to their definitions; follow the link, or search `river/` for `## scaleToggle`.

## Where to look

| Need | Read |
| --- | --- |
| The method and the change loop | [river/README.md](river/README.md) |
| Web app requirements | [river/requirements/composed/internal/app/app.md](river/requirements/composed/internal/app/app.md) |
| Open work | [river/todo.md](river/todo.md) (`grep -n -A1 '^## '`) |
| Code shape, gotchas, local server | [river/arch.md](river/arch.md) |
| Settled conflicts | [river/requirements/decisions.md](river/requirements/decisions.md) |

Read a requirement by its heading, not its file; its first paragraph is the rule.

## Commands

```
go build ./... && go vet ./... && gofmt -l . && go test ./...   # gofmt must print nothing
```

All spec checks are Go tests in `river/check/`. `TestRenderedTextContrast` and
`TestPagesFitAPhone` skip without Chrome (`BIRDVIEW_CHROME`); a skip is not a pass.

**End every turn with a freshly built server on <http://localhost:8080>**, started from the repo
root: `go build -o /tmp/birdview-serve ./cmd/birdview`, stop the old one, then start it **detached**
so the command returns: `nohup /tmp/birdview-serve > /tmp/birdview-serve.log 2>&1 &`. Never leave
it running as a foreground or tracked background task; that holds the turn open.

## Making a change

- Requirement first (statement first), then criterion, then code, in one commit.
- **Watch every new check fail**, make the mutation fail loudly if its anchor is missing, and read
  the failure message.
- A data-pipeline change updates [docs/pipeline.md](docs/pipeline.md) and re-renders it
  (product.pipelineDiagram).
- Found or left work goes in `river/todo.md` as an item, never as a `TODO` in code.
- **Never resolve a conflict in code**; take it to `decisions.md`.
- Cite requirements in code as `// Verifies: product.scaleToggle.`

## Hard rules: licensing

- **Never reintroduce the ABA checklist** (all rights reserved; decisions.scopeFromRecurrence).
  `fetched/` stays gitignored.
- **Never read or process Wiley Online Library material**, including Ricklefs 2017
  (decisions.ricklefsProhibited).
- **Never fetch or process anything from BirdLife's DataZone**; its terms forbid automated access.
- **Never commit the built dataset** (tech.datasetNeverCommitted).
- **Only AVONET, the Fox wingspan data and AviList may be committed as source data.** Check
  `river/requirements/imports.md` before adding any other. The generated tables in `data/`
  are committed on purpose.
- **Third-party files live under `third_party/<name>/`** with their own license, no Apache header;
  provenance in `river/requirements/imported/<origin>/<name>/PROVENANCE.md`.
- **Link to iNaturalist; copy nothing from it**, least of all photographs
  (product.inaturalistLink).
- **Never present a modeled value as a measurement** (product.tileMarksModeled).

## Conventions

American spellings, but **never edit a quotation or a vendored document**, and keep formatters
away from `third_party/`, `river/requirements/imported/` and `river/README.md`.
