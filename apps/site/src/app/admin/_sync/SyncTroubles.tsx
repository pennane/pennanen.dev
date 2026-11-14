import { cookies } from 'next/headers'
import { isValidSession, SESSION_COOKIE } from '../../../lib/admin/session'
import { logAdmin } from '../../../lib/log/admin-api'
import { troubleOf } from '../../../lib/log/syncs'
import type { SyncReport } from '../../../lib/log/types'
import styles from '../admin.module.css'

const describe = (report: SyncReport, now: number) => {
  switch (troubleOf(report, now)) {
    case 'failed':
      return `${report.name} sync failed: ${report.problems.at(-1) ?? 'no reason given'}`
    case 'overdue':
      return `${report.name} has not synced since ${report.lastSucceededAt?.slice(0, 10)}`
    default:
      return undefined
  }
}

export const SyncTroubles = async () => {
  const session = (await cookies()).get(SESSION_COOKIE)?.value
  if (!(await isValidSession(session))) return null
  const reports = await logAdmin.syncReports().catch(() => undefined)
  if (!reports) return <p className={styles.syncTrouble}>the log service did not answer</p>
  const now = Date.now()
  const troubles = reports.flatMap((report) => describe(report, now) ?? [])
  if (troubles.length === 0) return null
  return (
    <ul className={styles.syncTroubles}>
      {troubles.map((trouble) => (
        <li key={trouble} className={styles.syncTrouble}>
          {trouble}
        </li>
      ))}
    </ul>
  )
}
