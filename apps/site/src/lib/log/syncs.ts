import type { SyncReport } from './types'

export const STEAM_SYNC = 'steam'
export const IGDB_MATCH = 'igdb-match'
export const LETTERBOXD_EXPORT_FIELD = 'export'

const DAY_MS = 24 * 60 * 60 * 1000
const OVERDUE_AFTER_MS = 2 * DAY_MS

export type SyncTrouble = 'failed' | 'overdue'

export const troubleOf = (report: SyncReport, now: number): SyncTrouble | undefined => {
  if (report.state === 'failed') return 'failed'
  if (report.lastSucceededAt && now - Date.parse(report.lastSucceededAt) > OVERDUE_AFTER_MS)
    return 'overdue'
  return undefined
}
