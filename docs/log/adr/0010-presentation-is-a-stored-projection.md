# 10. Presentation is a stored projection

**Status:** accepted · 2026-11-15

## Context

The public API took shortcuts:

- Every page fetched whole kinds: an item page listed its entire kind to
  find its neighbours, the overview listed every kind to show five, and each
  request recomputed visibility, standings and an O(n²) unranked order.
- `/public/presentation` was the admin settings dump. The site resolved
  defaults and overrides itself, and the endpoint listed every item on each
  call, which the layout made on every page.
- Privacy was a deny-list: a new field was public unless marked private.
- The site named sources and fields: Letterboxd stars, `reaction`,
  `recommendation`, `review`, `_unlocked`, and its own copy of the
  visibility rule.
- Values went over the wire untyped.

## Decision

- Go resolves presentation fully. The **projection**, what each public page
  shows, is a derived table, rebuilt for dirty items and, when settings
  change, for the affected kind. Requests serve stored rows.
- Endpoints are shaped by page:
  - `/public/log`: the overview, each kind with its title, count and top
    entries as list rows;
  - `/public/kinds/{kind}`: the kind, its resolved filters and columns, and
    light list rows (title, slug, tile, scores, place);
  - `/public/items/{kind}/{slug}`: the item, its neighbours, its parts, its
    page fields and its takes.
- Values are allow-listed and typed: only fields that may show are sent,
  each with its type and public name. Takes declare their fields' types as
  kinds do (a rating with its scale, a recommendation, text), so the site
  renders by type.
- **Visibility** is `auto` (shown when the kind shows its status),
  `public` or `private`. A dismissed item never shows.
- A field's type decides what it can do: dates and numbers filter, sort and
  show as columns; lists filter and search; text searches; ids and links do
  nothing. Admin overrides per kind: each field's uses and public name,
  which axes get a **tier list** or a column, the default view, the
  **overview**'s sort and count, the shown statuses, and tag titles.
- Pictures are placed by shape: a **tile** is the first tall picture, else
  square, else wide; a **banner** is the first wide, else square, else tall.
- Next keeps layout, rendering, and filtering and sorting over the rows it
  receives.

## Consequences

- No site code names a source or a field.
- A new field shows up without touching admin or the site.
- A settings change rebuilds one kind's projection; revalidation can target
  a kind or an item.
