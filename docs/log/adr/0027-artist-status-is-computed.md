# 27. Artists are logged from top lists, and their status is computed

**Status:** accepted · 2026-11-15 · closes the artist question in [ADR 16](./0016-known-items-and-suggestions.md) · exception to [ADR 17](./0017-outside-takes-set-statuses.md)

## Context

Spotify's top artists say what ap listens to, with no work from ap. Putting
every one of them through triage would make a queue out of a list, and an
artist is not something ap decides to start or stop in a sitting: the lists
decide it.

## Decision

- Every artist that has ever been in a top-artist list is logged. Once
  logged, an artist stays logged: dropping out of every list makes it
  `listened`, never gone.
- An artist's status is a **computed status**, set by `spotify`. It is
  `listening` when the artist is in the 4-week list of the newest day any
  top-artist list was recorded, and `listened` otherwise. A quiet month that
  leaves the newest 4-week list empty makes nobody `listening`.
- The nightly sync recomputes every such artist after refreshing claims.
- Artists never pass through triage; their statuses are `listening`,
  `listened` and `dismissed`.
- This departs from ADR 17, where a status ap sets is never touched again.
  For a kind with a computed status only `dismissed` sticks. Any other
  status ap picks hands the status back to `spotify`, and the next sync
  computes it again. The kind names the source that computes its statuses.

## Consequences

- ap can rank and score artists without sorting them first.
- ap can't pin an artist to `listening`; the lists decide, and ap can only
  rule an artist out.
- Admin says so on the status chips of such a kind.
