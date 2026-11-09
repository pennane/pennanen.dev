# Conventions

- No code comments. Not doc comments, not inline, not in SQL. Say it with names: extract a function or a constant instead of explaining.
- Avoid booleans and impossible states. Prefer enums and data shapes that cannot be wrong.

# Layout

- `apps/site` — Next.js site on Vercel
- `apps/log` — Go service on the VPS
- `docs/log` — glossary and ADRs for the log; use its words in code
