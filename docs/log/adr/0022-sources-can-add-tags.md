# 22. Sources can add tags, on the same terms as statuses

**Status:** accepted · 2026-11-14 · amends [ADR 17](./0017-outside-takes-set-statuses.md)

## Context

Wanting to see a film and having seen it aren't exclusive: ap puts watched
films back on the Letterboxd watchlist to rewatch them. A status can't say
both, and `another-shot` already means "wants to try again, whatever its
status". But tags were ap's alone, so the sync dropped the watchlist fact.
Letterboxd also takes a film off the watchlist once it's logged, so a tag the
sync added and never removed would go stale after every rewatch.

## Decision

- A tag records who added it: `ap` or a source. ADR 17's rule for statuses
  holds for tags: a sync adds and removes only tags it added itself, and once
  ap adds or removes one in the app, it's ap's.
- The Letterboxd take records since when a film has been on the watchlist.
  A watched film arriving there gets `another-shot` from Letterboxd. It loses
  it when it leaves: a diary watch on or after that day, a film first
  watched, or an export upload without it on the watchlist.
- Statuses don't change: a film on both lists stays `watched`.

## Consequences

- Admin shows which source added a tag.
- Nightly syncs see only page one of the watchlist, so they notice arrivals
  near its top; an export upload settles the rest.
- No sync re-adds a tag ap removed until the film leaves the watchlist and
  comes back.
