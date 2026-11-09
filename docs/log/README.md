# things as ap saw them

A log of things ap has seen — games first, then whatever else. Mainly for ap,
partly public at `/log`. Words: [GLOSSARY](./GLOSSARY.md). Decisions: [adr/](./adr).

## v1

1. **Service** — Go in `log/`: SQLite, REST, kinds `game` and `thing`,
   verdicts, public read endpoints, nightly R2 backup, nightly Steam sync with
   idle hints. Deployed to `/opt/log`, `api.log.pennanen.dev` → `:3005`.
2. **Identity** — IGDB matching for games; manual-match queue for misses.
3. **Admin** — `/admin` PWA: triage (with bulk actions) first, then the
   add-anything search box with one-tap enjoyment, then entry editing.
4. **Public** — `/log`, `/log/game`, `/log/game/<slug>`.

## Later

`want` status · `person` kind and redacting private items on public pages ·
`tabletop` (BoardGameGeek) · wine · books · music · relations on public pages ·
importers beyond Steam · offline support.

### Next feature: rankings

Ordered lists of preference (tier lists, or a plain best-to-worst order) as a
second way to judge things, next to scores. Still open:

- What a ranking ranks: any items, one kind, or a set picked by tag or relation
  (e.g. "soulslikes", "Margaux vintages").
- Tiers (S/A/B…) vs a full order vs both (tiers holding an order inside).
- Whether a ranking can produce scores, or the two stay separate judgements.
- One ranking per set or many, each with its own visibility.
