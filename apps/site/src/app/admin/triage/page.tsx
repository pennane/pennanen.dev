import { requireAdmin } from '../../../lib/admin/require-admin'
import { logAdmin } from '../../../lib/log/admin-api'
import type { Kind } from '../../../lib/log/types'
import { catalogueOf } from '../_queue/catalogue'
import { QueueHeader } from '../_queue/card'
import styles from '../_queue/queue.module.css'
import { toRanked } from '../_queue/placing'
import { byOrder, toCard } from '../_queue/toCard'
import { STEAM_SYNC, IGDB_MATCH } from '../../../lib/log/syncs'
import { Triage, type KindQueue } from './Triage'

export const dynamic = 'force-dynamic'

const UNTRIAGED = 'untriaged'
const GAME = 'game'

type SearchParams = Promise<Record<string, string | string[] | undefined>>

const single = (value: string | string[] | undefined) =>
  typeof value === 'string' ? value : undefined

const queueOf = async (kind: Kind) => {
  const [untriaged, needingMatch] = await Promise.all([
    kind.statuses.includes(UNTRIAGED) ? logAdmin.items({ kind: kind.name, status: UNTRIAGED }) : [],
    catalogueOf(kind) ? logAdmin.items({ kind: kind.name, needs: 'match' }) : []
  ])
  return { kind, untriaged, needingMatch }
}

export default async function TriagePage({ searchParams }: { searchParams: SearchParams }) {
  await requireAdmin()
  const params = await searchParams
  const queues = await Promise.all((await logAdmin.kinds()).map(queueOf))
  const withWork = queues.filter((q) => q.untriaged.length + q.needingMatch.length > 0)
  const chosen = withWork.find((q) => q.kind.name === single(params.kind)) ?? withWork[0]
  const kindQueues: KindQueue[] = withWork.map((q) => ({
    kind: q.kind.name,
    untriaged: q.untriaged.length,
    needingMatch: q.needingMatch.length
  }))

  if (!chosen)
    return (
      <div className={styles.queue}>
        <QueueHeader current="triage" />
        <p className={styles.empty}>Nothing left to triage.</p>
      </div>
    )

  const { kind } = chosen
  const isGame = kind.name === GAME
  const [entries, steamSync, matchSync] = await Promise.all([
    logAdmin.entries(kind.name),
    isGame ? logAdmin.syncReport(STEAM_SYNC).catch(() => undefined) : undefined,
    isGame ? logAdmin.syncReport(IGDB_MATCH).catch(() => undefined) : undefined
  ])

  return (
    <Triage
      key={kind.name}
      kind={kind}
      kindQueues={kindQueues}
      cards={[
        ...chosen.untriaged.map((item) => toCard(kind, 'untriaged', item)).sort(byOrder),
        ...chosen.needingMatch.map((item) => toCard(kind, 'match', item)).sort(byOrder)
      ]}
      ranked={toRanked(kind, entries)}
      steamSync={steamSync}
      matchSync={matchSync}
    />
  )
}
