# 15. Verdicts inherit through a relation; wine and vintage are two kinds

**Status:** accepted · 2026-11-11 · amends [ADR 4](./0004-claims-and-verdicts.md)

## Context

Some items share most of their facts with another item. An album shares
genres with its artist; every year of a wine shares its winery, grapes,
style and region. Storing those facts on each item is redundant, and sources
often give them only once, on the shared item.

Separately, ap ranks bottles, and a bottle is one year of a wine. One kind
with the year as a field can't hold two years as separate entries.

## Decision

- A kind can mark a field **inherited** through one named relation. The
  verdict for that field is ap's override, then the item's own sources in
  order, and only then the related item's verdict. An album reaches its
  artist's genres; `vintage.grapes` comes from `vintage-of → wine`.
- Inheritance goes one step and no further.
- `wine` and `vintage` are separate kinds. A vintage is `vintage-of` its
  wine. A vintage is of a single year or **non-vintage** (blended across
  years); non-vintage is its own case, never a made-up year. Every bottle ap
  drinks is a vintage, so tasting, scores and ranks all live on vintages and
  every bottle ranks against every other.
- A wine is usually only known, and can be logged only as `want`: "try
  Krug, whatever year". Letting a wine be `tasted` would split bottles into
  two rankings.
- The full region chain stays a field of the wine, with Vivino's own ids,
  so it reaches every vintage in one step and can become items later
  without re-matching.

## Consequences

- Inheritance is part of the foundation; music uses it before wine does.
- Admin shows an inherited value as inherited, so it's clear where to fix it.
- Reading an item that inherits reads its related item too.
- "Every bottle from Bordeaux", once regions are items, is a walk along
  relations, not inheritance.
- Going further than one step later needs no data change, only a reader that
  follows more relations.
