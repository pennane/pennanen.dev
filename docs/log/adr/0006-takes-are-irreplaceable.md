# 6. Takes are the only irreplaceable data

**Status:** accepted · 2026-11-15

## Context

Records can mostly be fetched again and everything else is derived from
them. What ap thought, wrote and decided can't be. Syncs also come across far
more than ap means to log: every album in a top list, two thousand tracks,
the guest artists on a soundtrack.

## Decision

- A **take** is ap's opinion of a thing, stored with when ap gave it:
  - the **entry**: visibility, scores, rank, blurb, prose, time spent, and
    which source logged it, if one did;
  - a **status choice**;
  - **tag choices**, adding or removing a tag;
  - **dismissal**;
  - a **parked entry**, the losing side of a merge until ap answers.
- ap's corrections to the world (field overrides, match answers, link
  edits, merge answers) are not takes but records from the source `ap`
  ([ADR 4](./0004-records-are-lossless.md)). They are just as irreplaceable,
  and backed up the same way.
- Images in prose are stored next to the database, named by a hash of their
  content, and copied to R2 at upload. Prose refers to them as
  `/log/images/<name>`, so stored text never names a host. The database
  backup and the image copies together restore everything.
- An item has at most one entry. An item with an entry is **logged**; one
  without is only **known**. Nothing else tells them apart.
- **Dismissal** is a take, not a status: "this doesn't belong in my log".
  Nothing applies it but ap: a hint may suggest it (`likely-not-a-game`),
  and only ap's yes dismisses. Nothing overrides it. A dismissed item never
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
