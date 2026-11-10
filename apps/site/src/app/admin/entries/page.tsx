import Link from 'next/link'
import { requireAdmin } from '../../../lib/admin/require-admin'
import { logAdmin } from '../../../lib/log/admin-api'
import type { Visibility } from '../../../lib/log/types'
import { Entries } from './Entries'
import styles from './entries.module.css'

export const dynamic = 'force-dynamic'

const DEFAULT_KIND = 'game'
const NOT_YET_TRIAGED = 'untriaged'
const NEVER_PUBLIC = 'dismissed'
const VISIBILITIES: Visibility[] = ['auto', 'public', 'private']

type SearchParams = Promise<Record<string, string | string[] | undefined>>

const single = (value: string | string[] | undefined) =>
  typeof value === 'string' ? value : undefined

const entryId = (value: string | undefined) => {
  const id = Number(value)
  return value && Number.isInteger(id) ? id : undefined
}

export default async function EntriesPage({ searchParams }: { searchParams: SearchParams }) {
  await requireAdmin()
  const params = await searchParams
  const kinds = await logAdmin.kinds()
  const kindName = single(params.kind) ?? DEFAULT_KIND
  const kind = kinds.find((k) => k.name === kindName) ?? kinds[0]
  const entries = (await logAdmin.entries(kind.name)).sort((a, b) => a.title.localeCompare(b.title))
  const visibility = single(params.visibility)

  return (
    <div className={styles.page}>
      <nav className={styles.filters}>
        {kinds.map((k) => (
          <Link
            key={k.name}
            href={`/admin/entries?kind=${k.name}`}
            className={styles.filter}
            aria-current={k.name === kind.name ? 'true' : undefined}
          >
            {k.name}
          </Link>
        ))}
      </nav>
      <Entries
        key={kind.name}
        entries={entries}
        kind={kind}
        statuses={kind.statuses.filter((s) => s !== NOT_YET_TRIAGED && s !== NEVER_PUBLIC)}
        initial={{
          status: single(params.status),
          visibility: VISIBILITIES.find((v) => v === visibility),
          tag: single(params.tag)
        }}
        initiallyOpen={entryId(single(params.open))}
      />
    </div>
  )
}
