# Precedence — BirdView

The standing rules for which source outranks which. Highest first. Composition uses this to resolve
a conflict mechanically; anything it cannot resolve escalates to a person and is recorded in
[decisions.md](decisions.md). Each source's tier, and why, is in its entry in
[module.md](module.md).

| Tier | Sources | Why |
| --- | --- | --- |
| **Governing** | `inat-api` | The operator can throttle or block a client that ignores it; BirdView may not locally override it |
| **Mandatory** | `avonet`, `fox-wingspan`, `bohmer-neck`, `avilist`, `gbif`, `wiley`, `wcag`, `inat-terms`, `ebird-sensitive`, `natural-earth` | Copyright licenses and platform terms. A local requirement may tighten one, never loosen it |
| **Local** | `product`, `tech` | BirdView's own requirements |

Retired and rejected sources (`aba`, `ricklefs`, `birdlife-checklist`) bind nothing and have no
tier.

**No order between `product` and `tech` has been decided.** No conflict between them has yet
needed one; the first that does goes to `decisions.md`.
