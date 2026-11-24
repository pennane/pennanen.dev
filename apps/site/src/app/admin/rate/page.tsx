import { requireAdmin } from '../../../lib/admin/require-admin'
import { logAdmin } from '../../../lib/log/admin-api'
import type { ItemView, Kind } from '../../../lib/log/types'
import { QueueHeader } from '../_queue/card'
import { toRanked } from '../_queue/placing'
import styles from '../_queue/queue.module.css'
import { toGameCard, unrankedOrderOf } from '../_queue/toCard'
import { Rate, type KindQueue } from './Rate'

export const dynamic = 'force-dynamic'

type SearchParams = Promise<Record<string, string | string[] | undefined>>

const single = (value: string | string[] | undefined) =>
  typeof value === 'string' ? value : undefined

const queueOf = async (kind: Kind) => ({
  kind,
  items: await logAdmin.items({ kind: kind.name, needs: 'score' })
})

const byUnrankedOrder = (kind: Kind) => (a: ItemView, b: ItemView) =>
  unrankedOrderOf(kind, b) - unrankedOrderOf(kind, a)

export default async function RatePage({ searchParams }: { searchParams: SearchParams }) {
  await requireAdmin()
  const params = await searchParams
  const queues = await Promise.all((await logAdmin.kinds()).map(queueOf))
  const withWork = queues.filter((q) => q.items.length > 0)
  const chosen = withWork.find((q) => q.kind.name === single(params.kind)) ?? withWork[0]

  if (!chosen)
    return (
      <div className={styles.queue}>
        <QueueHeader current="rate" />
        <p className={styles.empty}>Everything has a score.</p>
      </div>
    )

  const { kind, items } = chosen
  const kindQueues: KindQueue[] = withWork.map((q) => ({
    kind: q.kind.name,
    count: q.items.length
  }))
  const entries = await logAdmin.entries(kind.name)
  return (
    <Rate
      key={kind.name}
      kind={kind}
      kindQueues={kindQueues}
      cards={[...items].sort(byUnrankedOrder(kind)).map(toGameCard)}
      ranked={toRanked(kind, entries)}
    />
  )
}
