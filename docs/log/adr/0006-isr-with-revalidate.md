# 6. Public pages use ISR, revalidated by the VPS

**Status:** accepted · 2026-11-09

## Context
pennanen.dev is on Vercel; the data lives on a small VPS that may go down.

## Decision
`/log`, `/log/<kind>`, `/log/<kind>/<slug>` (singular kind, slug unique per
kind) fetch from unauthenticated read endpoints that only return entries with
`public` visibility and their non-private moments. Pages are ISR-cached; the
VPS calls a Vercel revalidate hook (own secret) when public data changes. Covers use `next/image` with
`unoptimized` and the source's pre-sized variants.

## Consequences
- Visitors never depend on the VPS being up; they get the last cached page.
- Public/private filtering happens on the VPS; Next never sees private data
  through the public endpoints.
- Image optimization quota is untouched.
