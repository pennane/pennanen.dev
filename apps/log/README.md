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
| `GET /admin/items?kind=&status=&visibility=&hint=&tag=&q=` | items with verdicts and entries |
| `POST /admin/items` | `{kind, title, fields}` |
| `GET /admin/items/{id}` | item with claims, moments, relations |
| `DELETE /admin/items/{id}` | |
| `PATCH /admin/items/{id}/fields` | override fields; `null` removes |
| `PATCH /admin/items/{id}/entry` | `{status, visibility, notes, blurb, scores, addTags, removeTags}` |
| `POST /admin/entries/change` | `{itemIds, change}` for bulk triage |
| `POST /admin/items/{id}/moments` | `{at, note, visibility, scores, refs, scoresEffect}` |
| `DELETE /admin/moments/{id}` | |
| `POST`/`DELETE /admin/relations` | `{fromId, type, toId}` |
| `GET`/`POST /admin/sync/steam` | sync status / start |
| `POST /admin/backup` | back up now |
| `GET /public/items?kind=` | public entries |
| `GET /public/items/{kind}/{slug}` | one public entry with its visible moments |
