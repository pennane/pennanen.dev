# 8. Rankings are hand-picked, ordered and always public

**Status:** superseded by [ADR 12](./0012-divisions-and-ranks.md) · 2026-11-10

## Context

Scores answer "how good was it". ap also wants to say "these, in this order":
tier lists or a plain best-to-worst list, mostly to single out games.

## Decision

- A ranking holds placements ap picked by hand. Filters in the admin help find
  candidates but are not saved; nothing joins a ranking on its own.
- A ranking's layout is `numbered` (one order) or `tiered` (named tiers in
  order, each with its own order). The order is replaced whole on every save.
- Kinds may mix. Rankings and scores never derive from each other.
- Rankings have no visibility: they are public. A placement always shows the
  item's title, cover, kind and note; it links to the item page only when that
  entry is public. Dismissed items cannot be placed and never show.
- No history: only the current order and `updatedAt`.
- `ranking` is reserved and cannot be a kind name, so `/log/ranking/<slug>`
  never collides with `/log/<kind>/<slug>`.

## Consequences

- Placing a private entry in a ranking publishes its title and cover. That is
  the point of placing it.
- Deleting an item removes its placements.
