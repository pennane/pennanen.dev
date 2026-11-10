import Link from 'next/link'
import { requireAdmin } from '../../../lib/admin/require-admin'
import { logAdmin } from '../../../lib/log/admin-api'
import { QueueHeader } from '../_queue/card'
import { Order } from './Order'
import { toSections } from './sections'
import styles from './order.module.css'

export const dynamic = 'force-dynamic'

const DEFAULT_KIND = 'game'

type SearchParams = Promise<Record<string, string | string[] | undefined>>

export default async function OrderPage({ searchParams }: { searchParams: SearchParams }) {
  await requireAdmin()
  const requested = (await searchParams).kind
  const kinds = (await logAdmin.kinds()).filter((k) => k.rankedAxis)
  const kind = kinds.find((k) => k.name === requested) ?? kinds.find((k) => k.name === DEFAULT_KIND) ?? kinds[0]
  const entries = await logAdmin.entries(kind.name)
  const axis = kind.rankedAxis
  const sections = toSections(kind.divisions[axis] ?? [], entries, axis)
  const unrated = entries.filter((e) => e.scores[axis] === undefined).length

  return (
    <div className={styles.page}>
      <QueueHeader current="order" count={sections.reduce((n, s) => n + s.ids.length, 0)} />
      <nav className={styles.kinds}>
        {kinds.map((k) => (
          <Link
            key={k.name}
            href={`/admin/order?kind=${k.name}`}
            className={styles.kind}
            aria-current={k.name === kind.name ? 'true' : undefined}
          >
            {k.name}
          </Link>
        ))}
      </nav>
      <Order key={kind.name} sections={sections} entries={entries} />
      {unrated > 0 && (
        <p className={styles.hint}>
          {unrated} more without a division. <Link href="/admin/rate">rate them</Link>
        </p>
      )}
    </div>
  )
}
