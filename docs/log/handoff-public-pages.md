# Handoff: public log pages

Build the public side of *things as ap saw them* on pennanen.dev, plus the one admin control it depends on.

## Read first

- `CLAUDE.md`: no code comments, avoid booleans and impossible states.
- `docs/log/GLOSSARY.md`: use these words in code.
- `docs/log/adr/0006-isr-with-revalidate.md`: the caching design you are implementing.
- `docs/log/adr/0003-admin-in-next.md`, `docs/log/README.md` (scope, "later" list).

## Scope

1. **Pages** in `apps/site/src/app/(site)/log/` (the `(site)` group gives the navbar and footer):
   - `/log`: the public log.
   - `/log/[kind]`: one kind, e.g. `/log/game`.
   - `/log/[kind]/[slug]`: one item, with its public moments.
2. **Revalidate route** `POST /api/revalidate`:
   - Checks `Authorization: Bearer $REVALIDATE_SECRET`.
   - Takes the body `{"tags":["log"]}` and calls `revalidateTag` for each tag.
   - The VPS sends this, debounced by 5s, whenever data changes.
   - Public fetches use `next: { tags: ['log'] }`.
3. **Visibility control in the admin.** Every entry is private by default and nothing in the admin can change that yet, so prod has zero public entries.
   - The smallest useful version: `/admin/entries` lists triaged entries, filterable by kind and status, each with a private/public toggle.
   - Save through `PATCH /admin/items/{id}/entry` `{"visibility":"public"}`.
   - Follow the existing server-action pattern in `apps/site/src/app/admin/actions.ts`.
   - Full entry editing (notes, tags, moments) is a separate task, not this one.
4. **Wiring:**
   - Add the log pages to `src/app/sitemap.ts`.
   - Add a link from the site nav or home if ap wants one (ask).

Out of scope: rankings, IGDB, the `person` kind, editing notes, tags or moments.

## Decide with ap before building

These are ap's calls. Put them as questions with a recommendation each.

- What `/log` leads with: recent changes, per-kind sections, or favorites (`favorite/*` tags).
- How scores show: numbers (0–100), words (loved / liked / meh / disliked, i.e. 92 / 80 / 60 / 35 presets), or a bar. Whether `quality` shows next to `enjoyment`.
- Sorting and filtering on `/log/[kind]`: by score, by date, by status, by tag.
- Whether statuses show publicly (`finished`, `dropped`…), and how tags read (`favorite/all-time` is hierarchical on `/`).
- OG images for item pages. The site already generates meta images; see `src/meta/image`.

## Public API

Base URL comes from `LOG_API_URL`. No auth. These endpoints only ever return public entries and non-private moments, and never dismissed entries.

`GET /public/items?kind=game` returns `PublicItem[]`.
`GET /public/items/{kind}/{slug}` returns a `PublicItem` with `moments`, or 404.

```ts
type PublicItem = {
  kind: string
  slug: string
  fields: Record<string, unknown>
  status: string
  scores: Record<string, number>
  tags: string[]
  blurb: string
  updatedAt: string
  moments?: {
    at: string
    note: string
    scores: Record<string, number>
    refs: { kind: string; slug: string; title: string }[]
  }[]
}
```

Fields for `game` (any may be missing):
- `title`
- `cover` (Steam header URL, 460×215)
- `released` (partial date: `2020`, `2020-03` or `2020-03-17`)
- `description`
- `developers`, `genres`, `steam_categories` (lists)
- `playtime_minutes`, `last_played`
- `achievements_total`, `achievements_unlocked`

The `thing` kind has `title`, `description`, `link` and `image`, plus whatever ap added. Kind definitions come from `GET /admin/kinds` (admin only): statuses, axes, and field types. Go source: `apps/log/internal/kind/kinds.go` and `apps/log/internal/api/public.go`.

Covers use `next/image` with `unoptimized` (ADR 0006).

## Local dev

```sh
cd apps/log && go build -o "$TMPDIR/logbin" ./cmd/log
LOG_DB="$TMPDIR/dev.db" LOG_ADMIN_TOKEN=devtoken LOG_ADDR=127.0.0.1:3905 "$TMPDIR/logbin"
./scripts/seed-dev.sh
cd ../site && pnpm next dev
```

- The seed adds untriaged Steam games with every hint, two public games with scores, tags, a blurb and moments, a private game, and a public `thing`.
- `apps/site/.env.local` already points at `http://127.0.0.1:3905`, token `devtoken`, admin password `dev`. Add `REVALIDATE_SECRET` there for the route.
- Port 3905 may already be taken by another session's dev service. If so, use another port, and pass `LOG_URL` to the seed and `LOG_API_URL` to Next.

In the sandbox the Go build cache isn't writable, so prefix Go commands with `GOCACHE=$TMPDIR/gocache`.

## Conventions

- Next 15 app router, CSS modules, site color tokens: `--foreground`, `--foreground-secondary`, `--accent`, `--line-color`, `--background`, `--heading-gradient`.
- No semicolons, single quotes, two-space indent.
- `@total-typescript/ts-reset` is on, so `response.json()` returns `unknown`. Narrow it.
- Check with `pnpm tsc --noEmit`, `pnpm next lint` and `pnpm next build`. The build needs network for Google Fonts. CSS modules reject bare element selectors like `kbd {}`; scope them under a class.

## Shipping

- Work on a branch, not `main`. Pushing to `main` deploys both the site (Vercel) and, for `apps/log/**` changes, the VPS service.
- Commits carry no Claude attribution.
- Another session may be changing `apps/site/src/app/admin/triage/` and `apps/log/`. Keep away from those.
- After merge, someone sets `REVALIDATE_SECRET` in Vercel, and `REVALIDATE_URL=https://pennanen.dev/api/revalidate` plus the same secret in `/opt/log/.env` on the VPS, then restarts `log`. Leave that to ap; don't touch the VPS.
