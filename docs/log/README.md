# things as ap saw them

A log of things ap has seen — games first, then whatever else. Mainly for ap,
partly public at `/log`. Words: [GLOSSARY](./GLOSSARY.md). Decisions: [adr/](./adr).

The service is Go in `apps/log`, deployed to `/opt/log`, with
`api.log.pennanen.dev` → `:3005` ([ADR 1](./adr/0001-one-binary-on-the-vps.md)).
Admin and public pages are in `apps/site` ([ADR 2](./adr/0002-next-renders-go-resolves.md)).

## The model

Five layers, each reading only those below ([ADR 3](./adr/0003-the-layered-model.md)):

1. **Records** — what sources said, kept as received ([ADR 4](./adr/0004-records-are-lossless.md)).
2. **World** — items, their kinds, fields and links, derived from records
   ([ADR 5](./adr/0005-kinds-are-code.md), [ADR 9](./adr/0009-links.md)).
3. **Takes** — what ap decided; the only thing ap writes ([ADR 6](./adr/0006-takes-are-irreplaceable.md)).
   Scoring is part of it ([ADR 8](./adr/0008-scoring.md)).
4. **Rules** — proposals from evidence; the newest word wins ([ADR 7](./adr/0007-proposals-and-the-newest-word.md)).
5. **Presentation** — what each public page shows, stored ([ADR 10](./adr/0010-presentation-is-a-stored-projection.md)).

## Sources

- **Steam** — the game library, playtime, achievements, ap's reviews.
- **IGDB** — game identity, titles, dates, genres, developers, pictures.
- **TMDB** — film and series identity, credits, pictures.
- **Letterboxd** — ap's films: diary, ratings, watchlist; export ZIP for the rest.
- **Spotify** — top lists, plays, saved albums; artists, albums, tracks.

## Later

`want` beyond films · `person` kind · tabletop · wine · books · offline support.

The ADRs were rewritten and renumbered on 2026-11-15; the old set is in git
history (`git log -- docs/log/adr`).
