# Requirements — BirdView

What BirdView must do and how it must be built, and where each requirement came from.

| Path | What it holds |
| --- | --- |
| [`local/`](local/) | Sources written here: [`product`](local/product/product.md) (what BirdView must do) and [`tech`](local/tech/tech.md) (how it must be built) |
| [`imported/`](imported/) | Sources written elsewhere, at `imported/<origin>/<name>/`, each with a `PROVENANCE.md`; `<origin>` is the host of the source's `Origin` field |
| [`composed/internal/app/`](composed/internal/app/app.md) | Everything the web application must satisfy, and its criteria |
| [`acceptance.md`](acceptance.md) | Criteria not yet moved into a composed package |
| [`imports.md`](imports.md) | Every source BirdView is subject to or has evaluated: version, tier and status |
| [`precedence.md`](precedence.md) | The standing rules for which tier of source outranks which |
| [`decisions.md`](decisions.md) | The log of rulings on particular conflicts and judgments, append-only |

**One source is one package.** A source BirdView transcribed rules from is named for the source
without punctuation — `inat-api` is `package inatapi` in
[imported/inaturalist.org/inatapi/](imported/inaturalist.org/inatapi/) — so it keeps its readable
name in `imports.md` and its package follows Go's rules. A source that contributes no requirements,
such as a dataset, has a directory holding only its provenance.

**Composed packages are built, not written.** `app` is composed from `local/` and `imported/` under
`precedence.md` and `decisions.md`, says so in its first lines, and is changed by recomposing.

**Where a composed package sits is the record of who may import it.** Nothing is exported:
`app` is under `composed/internal/`.

## Module

```
module birdview
```
