# 31. Pictures are a field type with a shape

**Status:** accepted · 2026-11-15 · extends [ADR 23](./0023-public-pages-follow-field-types.md)

## Context

Public pages picked pictures by field name: `poster` was tall, `cover` and
`image` wide. Album covers and artist images are square, so the tier list
cut them into tall tiles and an item page would show them as a wide banner.

## Decision

- `picture` is a field type. Every picture field has a **shape**: `tall`,
  `wide` or `square`. A game's Steam cover is wide and its IGDB poster tall;
  a film's poster is tall and its backdrop wide; album covers and artist
  images are square.
- Presentation lists each kind's pictures with their shapes. Pages choose by
  shape, never by name: a **tile** (tier lists, the overview wall) is the
  first tall picture, else square, else wide; a **banner** (rows, tables,
  item pages, link previews) is the first wide picture, else square, else
  tall. A tile and a thumbnail take their picture's shape.
- An item page puts a wide picture above the title and a tall or square one
  beside it.
- Pictures are placed, not fields with uses: admin doesn't offer them as
  filters, columns or page fields.

## Consequences

- A new kind's pictures lay out right by declaring their shapes, with no
  site code.
- Claims and readings are unchanged; only the field declarations moved.
