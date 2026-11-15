# 9. Links

**Status:** accepted · 2026-11-15

## Context

The relational model is the point: an album is by an artist, a track is on
an album at a position, ap's prose says one game goes with another. These
were three tables with three sets of code: relations, parts and mentions.

## Decision

- A **link** is `from, to, type, origin`, plus a **position** for ordered
  types.
- Link types are declared in code (`by`, `with`, `part-of`, `mentions`, …).
  A type is ordered or not; only an ordered type has a position, so a
  position can't be missing or stray.
- The **origin** is a source, `ap` or `prose`. Source links are derived from
  records. **Mentions** are derived from prose. ap's link edits, adding a
  link or suppressing one, are takes, and resolve by the newest word
  ([ADR 7](./0007-proposals-and-the-newest-word.md)).
- A **part** is an item linked `part-of` a whole at a position: a track of an
  album, and later a season of a series or an episode of a season. One item can be a
  part of several wholes, at a position in each. Parts are items like any
  other: known until ap logs one, ranked against their own kind.
- **Inheritance**: a field can declare a link type it inherits through.
  When no source of the item has a value, the verdict comes from the item at
  the other end, one step and no further.

## Consequences

- Tracklists, "mentioned in" and "everything by" are one query shape.
- _ap suppresses a wrong `by` that Spotify keeps sending unchanged; it stays
  suppressed. If Spotify's credits change, the new links show._
- Rederiving an item rederives its one-step neighbours, which is how
  inheritance and parts stay current.
