# Requirements — BirdView

What BirdView must do and how it must be built, and where each requirement came from. The module
declaration lives in [`../README.md`](../README.md).

| Path | What it holds |
| --- | --- |
| [`local/`](local/) | Sources written here: [`product`](local/product/product.md) (what BirdView must do) and [`tech`](local/tech/tech.md) (how it must be built) |
| [`imported/`](imported/) | Sources written elsewhere, at their import path (`w3.org/wcag`, `gbif.org/gbifterms`, `figshare.com/avonet`), each with a `PROVENANCE.md` |
| [`composed/internal/app/`](composed/internal/app/app.md) | Everything the web application must satisfy, and its criteria |
| [`composed/internal/dataset/`](composed/internal/dataset/dataset.md) | Everything the data pipeline and dataset must satisfy, and its criteria |
| [`composed/internal/repo/`](composed/internal/repo/repo.md) | Everything repository layout, build, container image, and release must satisfy, and its criteria |
| [`imports.md`](imports.md) | The register of every source BirdView is subject to or has evaluated: version, tier, and status |
| [`precedence.md`](precedence.md) | The standing rules for which tier of source outranks which |
| [`decisions.md`](decisions.md) | The log of rulings, append-only: how each conflict, scope, or gap was settled, by whom and why |
| [`approvals.md`](approvals.md) | Machine-written approval records (`river approve --by <person>`) for each composed package |
| [`sums.md`](sums.md) | Machine-written content hashes (`river sum -w`) over every input to composition |

**One source is one package.** A source BirdView transcribed rules from is named for the source
without punctuation — `inat-api` is `package inatapi` in
[imported/inaturalist.org/inatapi/](imported/inaturalist.org/inatapi/) — so it keeps its readable
name in `imports.md` and its package follows Go's rules. A source that contributes no requirements,
such as a dataset, has a directory holding only its provenance.

**Composed packages are built, not written.** `app`, `dataset`, and `repo` are composed from
`local/` and `imported/` under `precedence.md` and `decisions.md`, say so in their first lines, and
are changed by recomposing. Mutable guardrail bindings and verification observations live in
[`../conformance.md`](../conformance.md).

**Where a composed package sits is the record of who may import it.** Nothing is exported: all
three packages live under `composed/internal/`.
