# 6. Takes are the only irreplaceable data

**Status:** accepted · 2026-11-15

## Context

Records can mostly be fetched again and everything else is derived from
them. What ap thought, wrote and decided can't be. Syncs also come across far
more than ap means to log: every album in a top list, two thousand tracks,
the guest artists on a soundtrack.

## Decision

- A **take** is something ap decided, stored with when ap decided it:
  - the **entry**: visibility, scores, rank, blurb, prose, time spent;
  - a **status choice**;
  - **tag choices**, adding or removing a tag;
  - **dismissal**;
  - **link edits**, adding or suppressing a link ([ADR 9](./0009-links.md));
  - **match answers**: the item ap picked, or no match;
  - field overrides.
- An item has at most one entry. An item with an entry is **logged**; one
  without is only **known**. Nothing else tells them apart.
- **Dismissal** is a take, not a status: "this doesn't belong in my log".
  Nothing proposes it and nothing overrides it. A dismissed item never
  shows publicly and is never proposed for logging again.
- Logging is one-way. A source that logged an item (Steam owned, a top
  list, a saved album) never unlogs it, and never takes its scores or prose
  with it. Un-logging is dismissing.
- Only an item ap made by hand can be deleted. Anything a source knows is
  dismissed instead.
- When an item's kind changes, ap's status choice is kept. If the new kind
  has no such status, the choice is set aside, not lost, and comes back if
  the item moves back.

## Consequences

- The backup protects records and takes; everything else is rebuilt.
- Public pages see only logged items, so known ones cost lists nothing.
- Logging a known item creates its entry; nothing else moves.
