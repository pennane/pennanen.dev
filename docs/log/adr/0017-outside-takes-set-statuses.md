# 17. Outside takes are claims, and they can set a status

**Status:** accepted · 2026-11-11 · amends [ADR 4](./0004-claims-and-verdicts.md)

## Context

ap already rates films on Letterboxd, shelves books on Goodreads and rates
wine on Vivino, and won't do it twice. Those actions say for sure what a
hint only guesses: a rated film was watched. Some sites also mix ap's take
with what they say about the thing, while a claim is one per item and source.

## Decision

- ap's take on another site is an **outside take**, stored as a claim from
  its own source: `letterboxd-take`, `goodreads-take`, `vivino-take`, apart
  from the site's descriptive source. Each kind marks which sources are
  takes and which of their fields are private (Vivino price, personal note);
  private fields show only in admin until ap makes one public.
- Outside takes show alongside the entry and its prose; neither replaces
  the other.
- An entry records who set its status: `ap` or the source that did. A sync
  creates or changes the status only when it set it itself: Letterboxd
  watched → `watched`, watchlist → `want`; Goodreads shelves → `want`,
  `reading`, `read`; Vivino rated → `tasted`, wishlist → `want`.
- When ap undoes the action on the site, an entry whose status that site set
  goes back to `untriaged`. An entry ap set is never touched.
- Any change ap makes to a status in the app, triage included, makes it
  set by ap. Existing triaged entries start set by ap; untriaged ones, by
  the source that made them.

## Consequences

- Most films, books and wine skip triage entirely.
- Hints still never set a status; outside takes are not guesses.
- An entry stays put when its status goes back to untriaged, so prose and
  scores are never lost to a sync.
