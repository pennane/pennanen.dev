# 5. Kinds are code, data is generic

**Status:** accepted · 2026-11-15

## Context

Games came first; films, series and music followed, and more sorts of
things will. Each should cost a definition, not a schema change. A
catch-all kind with unchecked fields was tried and never used.

## Decision

- A **kind** is a Go definition: its fields, statuses, identity source, the
  ranking of sources per field, inheritance, unranked order, take sources and
  the rules it uses.
- Each field has a **field type** (text, number, partial date, list, link,
  id, picture…), a **reach** (everyone, or only ap) and, for a picture, a
  **shape**.
- Tables are generic. Admin reads kind definitions from the API and builds
  its forms from them.
- There is no catch-all kind and no undeclared field. Something that fits no
  kind waits until one is defined for it.

## Consequences

- A new kind is a definition plus, maybe, a source; no migration.
- Validation happens in Go, not in the database.
- A field only ap may see never leaves the API ([ADR 10](./0010-presentation-is-a-stored-projection.md)).
