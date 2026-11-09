# 4. Sources make claims; verdicts are picked when read

**Status:** accepted · 2026-11-09

## Context
The same item can come from several sources (Steam, IGDB, manual), which may
disagree. Re-syncing must never clobber ap's fixes.

## Decision
Each source's data is stored raw as a **claim**, one per (item, source),
replaced wholesale on sync and never edited. The shown value of each field,
the **verdict**, is picked at read time from a per-field ranking of sources.
ap's override is a claim that always ranks first.

## Consequences
- Data is stored redundantly on purpose; storage is cheap.
- Changing source priority needs no data migration.
- ap's takes (entries) never live on items or claims.
- Games are matched across sources by IGDB ID; misses go to a manual-match queue.
