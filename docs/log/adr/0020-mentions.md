# 20. Mentions are kept apart from relations

**Status:** accepted · 2026-11-11 · amends [ADR 14](./0014-prose-replaces-moments.md)

## Context

The backend should hold a graph of items: what sources say (artist → album,
vintage → wine) and what ap writes ("Realm of the Mad God goes with
Pendulum"). ap's links live in prose, which is replaced wholesale on every
save.

## Decision

- A link to another item in prose is a **mention**. Mentions are read out of
  an entry's prose and stored in their own table, rebuilt whenever that
  prose is saved.
- `relation` keeps only structure that sources and ap's admin edits give.
  A prose save never touches it.
- The graph is the two read together.
- No history: mentions follow the prose as it stands.

## Consequences

- "Mentioned in" backlinks are a read of the mentions table.
- Rebuilding mentions needs no care about which edges came from where.
