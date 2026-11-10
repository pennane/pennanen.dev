import Link from 'next/link'
import { requireAdmin } from '../../../lib/admin/require-admin'
import { logAdmin } from '../../../lib/log/admin-api'
import { HEADLINE_AXIS, textField } from '../../../lib/log/present'
import type { ItemView, Visibility } from '../../../lib/log/types'
import { Entries, type ListedEntry } from './Entries'
import styles from './entries.module.css'

export const dynamic = 'force-dynamic'

const DEFAULT_KIND = 'game'
const NOT_YET_TRIAGED = 'untriaged'
const NEVER_PUBLIC = 'dismissed'
const VISIBILITIES: Visibility[] = ['auto', 'public', 'private']

type SearchParams = Promise<Record<string, string | string[] | undefined>>

const single = (value: string | string[] | undefined) =>
  typeof value === 'string' ? value : undefined

const toListedEntries = (items: ItemView[]): ListedEntry[] =>
  items
    .flatMap(({ id, slug, verdicts, entry }) =>
      entry && entry.status !== NOT_YET_TRIAGED && entry.status !== NEVER_PUBLIC
        ? [
            {
              id,
              title: textField(verdicts.values, 'title') ?? slug,
              cover: textField(verdicts.values, 'cover'),
              status: entry.status,
              headlineScore: entry.scores[HEADLINE_AXIS],
              visibility: entry.visibility,
              tags: entry.tags
            }
          ]
        : []
    )
    .sort((a, b) => a.title.localeCompare(b.title))

export default async function EntriesPage({
  searchParams
}: {
  searchParams: SearchParams
}) {
  await requireAdmin()
  const params = await searchParams
  const kinds = await logAdmin.kinds()
  const kindName = single(params.kind) ?? DEFAULT_KIND
  const kind = kinds.find((k) => k.name === kindName) ?? kinds[0]
  const entries = toListedEntries(await logAdmin.items({ kind: kind.name }))
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
        statuses={kind.statuses.filter((s) => s !== NOT_YET_TRIAGED && s !== NEVER_PUBLIC)}
        statusesShown={kind.statusesShown}
        initial={{
          status: single(params.status),
          visibility: VISIBILITIES.find((v) => v === visibility),
          tag: single(params.tag)
        }}
      />
    </div>
  )
}
