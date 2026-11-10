# 5. Kinds are defined in code, data is generic

**Status:** accepted · 2026-11-09 · axes moved to admin by [ADR 13](./0013-axes-as-data.md)

## Context

Games first, but books, music, tabletop, wine (regions, grapes, vintages)
and things that fit nowhere should be addable without migrations.

## Decision

Generic tables: `item` (kind, title, identity, `attrs` JSON), `relation`
(typed edge between items), `entry`, `score` (entry, axis, value), `moment`,
`claim`. Each kind is a definition in Go: attrs schema, axes, statuses,
identity source, claim ranking. The admin fetches them from `GET /kinds` and
renders forms from them. `thing` is the unchecked fallback kind.

## Consequences

- New kind = new definition (+ optional importer), no schema change.
- Scores are rows, so per-kind axes (enjoyment, quality, later aspects) are free.
- A `want` status can be added to any kind later by editing its status set.
- Validation of attrs happens in Go, not the database.
- Avoid booleans and impossible states: enums (status, visibility) and
  entry tags (`favorite/all-time`) instead of flags.
