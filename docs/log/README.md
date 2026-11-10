# things as ap saw them

A log of things ap has seen — games first, then whatever else. Mainly for ap,
partly public at `/log`. Words: [GLOSSARY](./GLOSSARY.md). Decisions: [adr/](./adr).

## v1

1. **Service** — Go in `log/`: SQLite, REST, kinds `game` and `thing`,
   verdicts, public read endpoints, nightly R2 backup, nightly Steam sync with
   idle hints. Deployed to `/opt/log`, `api.log.pennanen.dev` → `:3005`.
2. **Identity** — IGDB matching for games; manual-match queue for misses ([ADR 10](./adr/0010-igdb.md)).
3. **Admin** — `/admin` PWA: triage (with bulk actions) first, then the
   add-anything search box with one-tap prestige, then entry editing.
4. **Public** — `/log`, `/log/game`, `/log/game/<slug>`.
5. **Ranks** — prestige divisions and a strict order per kind ([ADR 12](./adr/0012-divisions-and-ranks.md)); axes set in admin ([ADR 13](./adr/0013-axes-as-data.md)).
6. **Rating** — public by default, triage then rate ([ADR 9](./adr/0009-public-by-default.md)); what shows and how is set in admin ([ADR 11](./adr/0011-presentation-settings.md)).

## Later

`want` status · `person` kind and redacting private items on public pages ·
`tabletop` (BoardGameGeek) · wine · books · music · relations on public pages ·
importers beyond Steam · offline support.
