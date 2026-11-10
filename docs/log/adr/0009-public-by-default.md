# 9. Entries are public by default; status decides what shows

**Status:** accepted · 2026-11-10

## Context

Entries started `private` and needed a separate publish step. Private entries
are rare, so the step only got in the way: ap rated games and none showed up.
Making everything public is not enough either, since a Steam library holds
hundreds of games ap never played.

`idled` was a status, but idling a game for cards says nothing about whether
ap also played it.

## Decision

- New entries are `public`. `private` is a rare per-entry opt-out. Every
  existing entry became public.
- Each kind lists the statuses it shows publicly. For `game`: `maybe-played`,
  `playing`, `played`, `finished`, `dropped`. For `thing`: `seen`.
  `untriaged`, `never-played` and `dismissed` stay off public pages.
- A score is not required to show. Rated entries sort first.
- Public pages show only a count of untriaged entries per kind, not the
  entries themselves.
- `idled` is a tag, like `another-shot`. Entries with status `idled` became
  `never-played` with the `idled` tag. The `likely-idled` hint stays.
- After triage comes rating: a one-card-at-a-time queue of shown entries with
  no score. There is no publish step; triage then rate.

## Consequences

- The schema default for `entry.visibility` stays `private`; the store sets
  `public` explicitly, because SQLite cannot change a column default without
  rebuilding the table and its foreign keys.
- Pages only update once the service tells the site to revalidate
  ([ADR 6](./0006-isr-with-revalidate.md)).
