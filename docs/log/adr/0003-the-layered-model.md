# 3. The layered model

**Status:** accepted · 2026-11-15

## Context

Data flows in from many sources and ap adds some by hand. Syncs used to
write straight into everything: `letterboxd/record.go` created items, set
statuses, added and removed tags and decided a film's kind; Spotify did the
same across `statuses.go`, `tracklists.go` and `listening.go`. Each source
carried its own copy of the rules for who may change what, so each new
source needed its own decision, and nobody could hold the whole in mind.

## Decision

The log is five layers:

```
5. Presentation   what shows and how           settings, field types, visibility
4. Rules          records + world → proposals   status/tag/match/entry proposals
3. Takes          ap's, irreplaceable           entry, status choice, scores, rank,
                                                tags, prose, link edits, dismissal
2. World          derived from records          items, identity, kind, fields
                                                (verdicts), links, inheritance
1. Records        lossless                      claims, observations, readings
```

- Each layer reads only the layers below it.
- Sources write only **records** ([ADR 4](./0004-records-are-lossless.md)).
- **Takes** are the only thing ap writes ([ADR 6](./0006-takes-are-irreplaceable.md)).
- The **world**, **rules** and **presentation** are derived. They are
  stored, not computed when read. Writing a record or a take marks what it
  touches **dirty**; one derivation pass rederives each dirty item and its
  one-step neighbours through links, then reruns rules and the projection
  for them.
- Items are part of the world, derived like everything else in it. An
  **item** is the records that share an **identity**. It has a stable id
  and slug and a set of identities; identities that arrive later, like an
  IGDB match on a Steam game, join the set without changing either.
  **Matching** links records across sources: a sure match is derived, an
  unsure one is a suggesting proposal ([ADR 7](./0007-proposals-and-the-newest-word.md)),
  and ap's answer is a take. An item's **kind** is derived too, so a film
  found to be a series **moves**, keeping its id and entry.
- Two items found to share an identity are one item. The one whose entry
  was edited last keeps its id and slug; the other's entry is parked and
  proposed for **merging** in triage. Yes folds it in, no keeps the items
  apart for good. Nothing is lost either way.
- A full **rebuild** runs at start, before serving, only when a derive,
  reader, rule or projection version changes, in batches with one
  transaction each.

## Consequences

- A new source is a fetch that writes records, a derive, and maybe rules. It
  can't touch takes, links or entries.
- Wiping the derived layers and rebuilding gives the same log.
- Requests stay cheap on the VPS ([ADR 1](./0001-one-binary-on-the-vps.md)):
  they only read stored rows.
- A change to a derive, reader, rule or projection needs its version bumped,
  or stored results stay as they were.
