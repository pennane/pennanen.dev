import Link from 'next/link'
import { requireAdmin } from '../../../lib/admin/require-admin'
import { logAdmin } from '../../../lib/log/admin-api'
import { toRanked } from '../_queue/placing'
import styles from '../_queue/queue.module.css'
import { Add } from './Add'

export const dynamic = 'force-dynamic'

const DEFAULT_KIND = 'game'

type SearchParams = Promise<Record<string, string | string[] | undefined>>

export default async function AddPage({ searchParams }: { searchParams: SearchParams }) {
  await requireAdmin()
  const params = await searchParams
  const kinds = await logAdmin.kinds()
  const wanted = typeof params.kind === 'string' ? params.kind : DEFAULT_KIND
  const kind = kinds.find((k) => k.name === wanted) ?? kinds[0]
  const entries = await logAdmin.entries(kind.name)
  return (
    <Add key={kind.name} kind={kind} ranked={toRanked(kind, entries)}>
      <nav className={styles.filters}>
        {kinds.map((k) => (
          <Link
            key={k.name}
            href={`/admin/add?kind=${k.name}`}
            className={styles.filter}
            aria-current={k.name === kind.name ? 'true' : undefined}
          >
            {k.name}
          </Link>
        ))}
      </nav>
    </Add>
  )
}
