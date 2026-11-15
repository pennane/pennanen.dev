# 28. Saved albums are logged, with a status from evidence

**Status:** accepted · 2026-11-15 · amends [ADR 17](./0017-outside-takes-set-statuses.md)

## Context

Saving an album on Spotify means ap has heard it or means to. Which one it
is shows in the listening: a track from it in a top list or in what ap
played. An album also credits its artists in order, and only the first one
is the album's own; the rest are guests, collaborators or a soundtrack's
cast.

## Decision

- Every album ap has saved on Spotify is logged, its status set by
  `spotify`. It is `listened` once any of its tracks has been in a top-track
  list or played, ever, and `want` until then. Observations are never
  deleted, so a `listened` album never goes back to `want`.
- This is not a computed status. ADR 17 applies: once ap picks an album's
  status, the sync leaves it alone.
- Unsaving an album changes nothing about its status; the unsave is kept as
  an observation. This amends ADR 17, where undoing the action on the site
  sends the status back to `untriaged`.
- An album is `by` its **main artist**, the first credited, and `with` every
  other credited artist. Spotify's "Various Artists" is never `by`, so a
  compilation has no main artist.
- The main artist of every saved album is logged, with the computed status
  of [ADR 27](./0027-artist-status-is-computed.md). The other credited
  artists stay only known.
- An album inherits `genres` through `by` ([ADR 15](./0015-inherited-verdicts.md)).

## Consequences

- Saved albums skip triage, and the public page shows only the heard ones.
- A soundtrack or collaboration doesn't put strangers on the artists page.
- A compilation inherits no genres; ap can give it some by hand.
- Relations made before this all said `by`; a migration turned every credit
  but the first into `with`.
