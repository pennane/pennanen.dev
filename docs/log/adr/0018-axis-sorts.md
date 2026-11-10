# 18. Own, outside and crowd axes

**Status:** accepted · 2026-11-11 · amends [ADR 13](./0013-axes-as-data.md)

## Context

ADR 13 left axes filled from source claims for later. Outside takes carry
ap's ratings on other sites, and the same sites carry public averages. Both
are worth showing and filtering on, but neither is ap's opinion here.

## Decision

- Every axis has a sort: **own**, **outside** or **crowd**. Own axes are as
  in ADR 13: set only by ap, prestige the only ranked one.
- An outside axis is tied to one take source (`letterboxd`, `vivino`,
  `goodreads`); a crowd axis to one descriptive source (Vivino average,
  Goodreads average, TMDB vote). Each has a mapping from the source's scale
  onto its divisions, editable in admin.
- Outside and crowd scores are read from claims when read, like verdicts.
  Nothing stores them and ap can't set them. Neither sort ranks.
- An outside axis suggests a division on ap's own axes while rating. A crowd
  axis never suggests; the crowd's opinion isn't ap's.
- Crowd scores keep only their latest value; no history.

## Consequences

- Changing a mapping changes every score on that axis at once, with no
  migration.
- Tier lists and filters can use outside and crowd axes like any other.
