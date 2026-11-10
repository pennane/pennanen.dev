# log

The service behind *things as ap saw them*. Words: [docs/log/GLOSSARY.md](../../docs/log/GLOSSARY.md).

```sh
go test ./...
LOG_ADMIN_TOKEN=dev go run ./cmd/log
```

Settings are environment variables; see [deploy/env.example](deploy/env.example). Anything left empty switches its feature off.

## Deploy

Pushing to `main` under `apps/log/` tests, builds and ships the binary to `/opt/log/log` and restarts `log.service`. Data lives in `/var/lib/log/log.db`.

Once per VPS:

1. `/opt/log/.env` from `deploy/env.example`, `chmod 600`.
2. Caddy:
   ```
   api.log.pennanen.dev {
       reverse_proxy localhost:3005
   }
   ```
   then `systemctl reload caddy`.
3. Netlify DNS: `A api.log 134.209.254.153`.
4. This repo's GitHub settings: secret `VPS_SSH_KEY`, variables `VPS_HOST` and `VPS_HOST_KEY` (same values as the chess repo).
5. R2: a bucket with a lifecycle rule expiring `backups/` after 30 days.

## API

Admin routes need `Authorization: Bearer $LOG_ADMIN_TOKEN`.

| | |
|---|---|
| `GET /admin/kinds` | kind definitions |
| `GET /admin/items?kind=&status=&visibility=&hint=&tag=&q=&needs=score` | items with verdicts and entries; `needs=score` lists played entries with no score, `needs=match` (with `kind`) unmatched ones |
| `POST /admin/items` | `{kind, title, fields}` |
| `GET /admin/items/{id}` | item with claims, moments, relations |
| `DELETE /admin/items/{id}` | |
| `PATCH /admin/items/{id}/fields` | override fields; `null` removes |
| `PATCH /admin/items/{id}/entry` | `{status, visibility, notes, blurb, scores, addTags, removeTags}` |
| `POST /admin/entries/change` | `{itemIds, change}` for bulk triage |
| `POST /admin/items/{id}/moments` | `{at, note, visibility, scores, refs, scoresEffect}` |
| `DELETE /admin/moments/{id}` | |
| `POST`/`DELETE /admin/relations` | `{fromId, type, toId}` |
| `GET`/`POST /admin/rankings` | list / create `{name, description}` |
| `GET`/`PATCH`/`DELETE /admin/rankings/{id}` | ranking with its items / `{name, description}` / remove |
| `PUT /admin/rankings/{id}/order` | `{layout: "numbered", placements}` or `{layout: "tiered", tiers: [{name, placements}]}`; placement is `{itemId, note}` |
| `GET /admin/igdb/search?q=` | IGDB games, each with `itemId` when already in the log |
| `POST /admin/items/from-igdb` | `{igdbId}`; 201 new, 200 when already in the log |
| `PUT /admin/items/{id}/igdb` | `{igdbId}`; 409 when another item has that game |
| `PUT /admin/items/{id}/igdb/none` | IGDB has no match; never asked again |
| `GET`/`POST /admin/sync/igdb` | match Steam games by app id: status / start |
| `GET`/`POST /admin/sync/steam` | sync status / start |
| `POST /admin/backup` | back up now |
| `GET /public/items?kind=` | public entries |
| `GET /public/items/{kind}/{slug}` | one public entry with its visible moments |
| `GET /public/kinds/{kind}` | `{name, untriaged}` |
| `GET /public/rankings` | all rankings |
| `GET /public/rankings/{slug}` | one ranking; each placement has `kind, title, cover, note`, and `page` when the entry is public |
