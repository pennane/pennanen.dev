import Image from 'next/image'
import Link from 'next/link'
import { requireAdmin } from '../../../lib/admin/require-admin'
import { logAdmin } from '../../../lib/log/admin-api'
import {
  HEADLINE_AXIS,
  statusLabel,
  textField
} from '../../../lib/log/present'
import type { ItemView, Kind, Visibility } from '../../../lib/log/types'
import styles from './entries.module.css'
import { VisibilitySwitch } from './VisibilitySwitch'

export const dynamic = 'force-dynamic'

const DEFAULT_KIND = 'game'
const NOT_YET_TRIAGED = 'untriaged'
const NEVER_PUBLIC = 'dismissed'
const VISIBILITIES: Visibility[] = ['private', 'public']

type SearchParams = Promise<Record<string, string | string[] | undefined>>

type Filter = {
  kind: Kind
  status: string | undefined
  visibility: Visibility | undefined
}

const single = (value: string | string[] | undefined) =>
  typeof value === 'string' ? value : undefined

const listableStatuses = (kind: Kind) =>
  kind.statuses.filter((s) => s !== NOT_YET_TRIAGED && s !== NEVER_PUBLIC)

const readFilter = (
  kinds: Kind[],
  params: Record<string, string | string[] | undefined>
): Filter => {
  const kindName = single(params.kind) ?? DEFAULT_KIND
  const kind = kinds.find((k) => k.name === kindName) ?? kinds[0]
  const status = single(params.status)
  const visibility = single(params.visibility)
  return {
    kind,
    status: status && listableStatuses(kind).includes(status) ? status : undefined,
    visibility: VISIBILITIES.find((v) => v === visibility)
  }
}

const hrefFor = (filter: Filter) => {
  const query = new URLSearchParams({ kind: filter.kind.name })
  if (filter.status) query.set('status', filter.status)
  if (filter.visibility) query.set('visibility', filter.visibility)
  return `/admin/entries?${query}`
}

const apiFilterFor = (filter: Filter) => ({
  kind: filter.kind.name,
  ...(filter.status && { status: filter.status }),
  ...(filter.visibility && { visibility: filter.visibility })
})

type Choice = { label: string; filter: Filter }

const Choices = ({ current, choices }: { current: Filter; choices: Choice[] }) => (
  <nav className={styles.filters}>
    {choices.map(({ label, filter }) => (
      <Link
        key={label}
        href={hrefFor(filter)}
        className={styles.filter}
        aria-current={hrefFor(filter) === hrefFor(current) ? 'true' : undefined}
      >
        {label}
      </Link>
    ))}
  </nav>
)

type ListedEntry = {
  id: number
  title: string
  cover: string | undefined
  status: string
  headlineScore: number | undefined
  visibility: Visibility
}

const toListedEntries = (items: ItemView[]): ListedEntry[] =>
  items.flatMap(({ id, slug, verdicts, entry }) =>
    entry && entry.status !== NOT_YET_TRIAGED && entry.status !== NEVER_PUBLIC
      ? [
          {
            id,
            title: textField(verdicts.values, 'title') ?? slug,
            cover: textField(verdicts.values, 'cover'),
            status: entry.status,
            headlineScore: entry.scores[HEADLINE_AXIS],
            visibility: entry.visibility
          }
        ]
      : []
  )

export default async function EntriesPage({
  searchParams
}: {
  searchParams: SearchParams
}) {
  await requireAdmin()
  const kinds = await logAdmin.kinds()
  const filter = readFilter(kinds, await searchParams)
  const entries = toListedEntries(await logAdmin.items(apiFilterFor(filter)))
    .sort((a, b) => a.title.localeCompare(b.title))

  return (
    <div className={styles.page}>
      <header className={styles.header}>
        <h1>
          entries <span className={styles.count}>{entries.length}</span>
        </h1>
        <Link href="/admin/triage" className={styles.quiet}>
          triage
        </Link>
      </header>
      <Choices
        current={filter}
        choices={kinds.map((kind) => ({
          label: kind.name,
          filter: { kind, status: undefined, visibility: filter.visibility }
        }))}
      />
      <Choices
        current={filter}
        choices={[
          { label: 'any status', filter: { ...filter, status: undefined } },
          ...listableStatuses(filter.kind).map((status) => ({
            label: statusLabel(status),
            filter: { ...filter, status }
          }))
        ]}
      />
      <Choices
        current={filter}
        choices={[
          { label: 'anyone', filter: { ...filter, visibility: undefined } },
          ...VISIBILITIES.map((visibility) => ({
            label: visibility,
            filter: { ...filter, visibility }
          }))
        ]}
      />
      {entries.length === 0 ? (
        <p className={styles.empty}>No entries match.</p>
      ) : (
        <ul className={styles.list}>
          {entries.map((entry) => (
            <li key={entry.id} className={styles.row}>
              <div className={styles.thumb}>
                {entry.cover && (
                  <Image alt="" src={entry.cover} width={92} height={43} unoptimized />
                )}
              </div>
              <div className={styles.rowBody}>
                <span className={styles.title}>{entry.title}</span>
                <span className={styles.meta}>
                  {statusLabel(entry.status)}
                  {entry.headlineScore !== undefined && ` · ${entry.headlineScore}`}
                </span>
              </div>
              <VisibilitySwitch itemId={entry.id} visibility={entry.visibility} />
            </li>
          ))}
        </ul>
      )}
    </div>
  )
}
