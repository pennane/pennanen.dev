# log

The service behind *things as ap saw them*. Words: [docs/log/GLOSSARY.md](../../docs/log/GLOSSARY.md).

```sh
go test ./...
LOG_ADMIN_TOKEN=dev go run ./cmd/log
```

Settings are environment variables; see [deploy/env.example](deploy/env.example). Anything left empty switches its feature off.

## Deploy

Pushing to `main` under `apps/log/` tests, builds and ships the binary to `/opt/log/log` and restarts `log.service`. Data lives in `/var/lib/log/log.db`, uploaded images in `/var/lib/log/images`.

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
5. R2: a bucket with a lifecycle rule expiring `backups/` after 30 days. Images are copied to `images/` and never expire.

## API

Admin routes need `Authorization: Bearer $LOG_ADMIN_TOKEN`.

| | |
|---|---|
| `GET /admin/kinds` | kind definitions, shown statuses and the axes each kind uses, with their divisions |
| `GET /admin/settings` | axes, per-kind settings and tag titles |
| `PUT /admin/settings/axes` | every axis `[{key, name, divisions: [{key, name}]}]` in order, divisions best first; a new axis or division without a key gets one from its name; `prestige` stays, and an axis or division in use cannot be removed |
| `PUT /admin/settings/kinds/{kind}` | `{statusesShown, axes}`: shown statuses and the axis keys the kind uses, in order; a ranked kind keeps `prestige`, and an axis with scores cannot be dropped |
| `PUT /admin/settings/tag-titles` | `{tag: title}` |
| `GET /admin/entries?kind=` | triaged, undismissed entries with only what the admin list shows, `rank` on ranked ones |
| `GET /admin/items?kind=&status=&visibility=&hint=&tag=&q=&needs=score` | items with verdicts and entries; `needs=score` lists played entries with no score, `needs=match` (with `kind`) unmatched ones |
| `POST /admin/items` | `{kind, title, fields}` |
| `GET /admin/items/{id}` | item with claims and relations |
| `DELETE /admin/items/{id}` | |
| `PATCH /admin/items/{id}/fields` | override fields; `null` removes |
| `PATCH /admin/items/{id}/entry` | `{status, visibility, blurb, prose, timeSpent, scores: {axis: division}, addTags, removeTags}`; prose is Markdown; a new division joins the bottom of it, `dismissed` clears scores |
| `PUT /admin/items/{id}/place` | `{division, index}` on the ranked axis; index 0 is the top of the division |
| `POST /admin/entries/change` | `{itemIds, change}` for bulk triage |
| `POST`/`DELETE /admin/relations` | `{fromId, type, toId}` |
| `GET /admin/igdb/search?q=` | IGDB games, each with `itemId` when already in the log |
| `POST /admin/items/from-igdb` | `{igdbId}`; 201 new, 200 when already in the log |
| `PUT /admin/items/{id}/igdb` | `{igdbId}`; 409 when another item has that game |
| `PUT /admin/items/{id}/igdb/none` | IGDB has no match; never asked again |
| `GET`/`POST /admin/sync/igdb` | match Steam games by app id: status / start |
| `GET`/`POST /admin/sync/steam` | sync status / start |
| `POST /admin/backup` | back up now |
| `POST /admin/images/ticket` | `{ticket, path}`: lets a browser upload images for 10 minutes |
| `POST /images` | `Authorization: Ticket <ticket>`, the image as the body (jpg, png, webp or gif, at most 10 MB); `{name}`, named by its content |
| `GET /images/{name}` | an uploaded image, cached forever |
| `GET /public/items?kind=` | public entries; ranked ones carry `rank: {place, of}` counted among public entries |
| `GET /public/items/{kind}/{slug}` | one public entry with its prose |
| `GET /public/kinds/{kind}` | `{name, untriaged}` |
| `GET /public/presentation` | axes, per-kind settings and tag titles |
