'use client'

import Image from 'next/image'
import Link from 'next/link'
import { useEffect, useMemo, useState } from 'react'
import {
  divisionName,
  findAxis,
  PRESTIGE,
  presenceOf,
  statusLabel,
  tagIsWithin,
  tagLabel,
  tagWithAncestors
} from '../../../lib/log/present'
import type { EntryChange, Kind, ListedEntry, Visibility } from '../../../lib/log/types'
import { changeEntry, openEntry } from '../actions'
import { Editor, type EntrySubject } from './Editor'
import styles from './entries.module.css'
import { VISIBILITY_LABELS, VisibilitySwitch } from './VisibilitySwitch'

type Panel =
  | { state: 'closed' }
  | { state: 'loading'; id: number }
  | { state: 'open'; id: number; subject: EntrySubject }
  | { state: 'failed'; id: number }

const CLOSED: Panel = { state: 'closed' }

const DISMISSED = 'dismissed'

type Filter = {
  status: string | undefined
  visibility: Visibility | undefined
  tag: string | undefined
}

const VISIBILITIES: Visibility[] = ['auto', 'public', 'private']

const matches = (entry: ListedEntry, { status, visibility, tag }: Filter, query: string) =>
  (!status || entry.status === status) &&
  (!visibility || entry.visibility === visibility) &&
  (!tag || entry.tags.some((t) => tagIsWithin(t, tag))) &&
  (!query || entry.title.toLowerCase().includes(query))

const tagsByUse = (entries: ListedEntry[]) => {
  const uses = new Map<string, number>()
  for (const entry of entries) {
    const tags = new Set(entry.tags.flatMap(tagWithAncestors))
    for (const tag of tags) uses.set(tag, (uses.get(tag) ?? 0) + 1)
  }
  return [...uses].sort((a, b) => b[1] - a[1] || a[0].localeCompare(b[0]))
}

const openIdOf = (panel: Panel) => (panel.state === 'closed' ? undefined : String(panel.id))

const withScores = (scores: Record<string, string>, changed: Record<string, string | null>) => {
  const next = { ...scores }
  for (const [axis, value] of Object.entries(changed)) {
    if (value === null) delete next[axis]
    else next[axis] = value
  }
  return next
}

const applyChange = (entry: ListedEntry, change: EntryChange): ListedEntry => ({
  ...entry,
  status: change.status ?? entry.status,
  visibility: change.visibility ?? entry.visibility,
  scores: change.scores ? withScores(entry.scores, change.scores) : entry.scores,
  rank: change.scores || change.status === DISMISSED ? undefined : entry.rank,
  tags: [
    ...entry.tags.filter((t) => !change.removeTags?.includes(t)),
    ...(change.addTags ?? []).filter((t) => !entry.tags.includes(t))
  ]
})

const isPlainClick = (event: React.MouseEvent) =>
  event.button === 0 && !event.metaKey && !event.ctrlKey && !event.shiftKey && !event.altKey

const syncUrl = (params: Record<string, string | undefined>) => {
  const url = new URL(window.location.href)
  for (const [name, value] of Object.entries(params)) {
    if (value) url.searchParams.set(name, value)
    else url.searchParams.delete(name)
  }
  window.history.replaceState(null, '', url)
}

export const Entries = ({
  entries: initialEntries,
  kind,
  statuses,
  initial,
  initiallyOpen
}: {
  entries: ListedEntry[]
  kind: Kind
  statuses: string[]
  initial: Filter
  initiallyOpen: number | undefined
}) => {
  const [entries, setEntries] = useState(initialEntries)
  const [filter, setFilter] = useState(initial)
  const [query, setQuery] = useState('')
  const [panel, setPanel] = useState<Panel>(
    initiallyOpen === undefined ? CLOSED : { state: 'loading', id: initiallyOpen }
  )
  const statusesShown = kind.statusesShown

  const choose = (change: Partial<Filter>) => {
    const next = { ...filter, ...change }
    setFilter(next)
    syncUrl({ ...next, open: openIdOf(panel) })
  }

  const showPanel = (next: Panel) => {
    setPanel(next)
    syncUrl({ ...filter, open: openIdOf(next) })
  }

  const loadingId = panel.state === 'loading' ? panel.id : undefined
  useEffect(() => {
    if (loadingId === undefined) return
    let current = true
    openEntry(loadingId).then(
      (subject) => current && setPanel({ state: 'open', id: loadingId, subject }),
      () => current && setPanel({ state: 'failed', id: loadingId })
    )
    return () => {
      current = false
    }
  }, [loadingId])

  const allTags = useMemo(() => tagsByUse(entries).map(([tag]) => tag), [entries])

  const patchEntry = (id: number, change: EntryChange) =>
    setEntries((all) => all.map((e) => (e.id === id ? applyChange(e, change) : e)))

  const tags = useMemo(() => {
    const counted = tagsByUse(entries.filter((e) => matches(e, { ...filter, tag: undefined }, '')))
    return !filter.tag || counted.some(([tag]) => tag === filter.tag)
      ? counted
      : [...counted, [filter.tag, 0] as [string, number]]
  }, [entries, filter])
  const shown = useMemo(
    () => entries.filter((e) => matches(e, filter, query.trim().toLowerCase())),
    [entries, filter, query]
  )

  const showVisibility = (id: number, visibility: Visibility) =>
    setEntries((all) => all.map((e) => (e.id === id ? { ...e, visibility } : e)))

  const changeVisibility = async (entry: ListedEntry, visibility: Visibility) => {
    showVisibility(entry.id, visibility)
    try {
      await changeEntry(entry.id, { visibility })
    } catch (error) {
      showVisibility(entry.id, entry.visibility)
      throw error
    }
  }

  return (
    <div className={styles.workspace} data-panel={panel.state}>
      <div className={styles.listSide}>
        <header className={styles.header}>
          <h1>
            entries <span className={styles.count}>{shown.length}</span>
          </h1>
          <Link href="/admin/triage" className={styles.quiet}>
            triage
          </Link>
          <input
            className={styles.search}
            type="search"
            value={query}
            placeholder="find"
            onChange={(event) => setQuery(event.target.value)}
          />
        </header>
        <Choices
          chosen={filter.status}
          anyLabel="any status"
          options={statuses.map((s) => ({ value: s, label: statusLabel(s) }))}
          onChoose={(status) => choose({ status })}
        />
        <Choices
          chosen={filter.visibility}
          anyLabel="anyone"
          options={VISIBILITIES.map((v) => ({ value: v, label: VISIBILITY_LABELS[v] }))}
          onChoose={(visibility) => choose({ visibility })}
        />
        {tags.length > 0 && (
          <Choices
            chosen={filter.tag}
            anyLabel="any tag"
            options={tags.map(([tag, uses]) => ({ value: tag, label: `${tagLabel(tag)} ${uses}` }))}
            onChoose={(tag) => choose({ tag })}
          />
        )}
        {shown.length === 0 ? (
          <p className={styles.empty}>No entries match.</p>
        ) : (
          <ul className={styles.list}>
            {shown.map((entry) => (
              <li
                key={entry.id}
                className={styles.row}
                aria-current={openIdOf(panel) === String(entry.id) ? 'true' : undefined}
              >
                <div className={styles.thumb}>
                  {entry.cover && (
                    <Image
                      alt=""
                      src={entry.cover}
                      width={92}
                      height={43}
                      unoptimized
                      loading="lazy"
                    />
                  )}
                </div>
                <div className={styles.rowBody}>
                  <Link
                    href={`/admin/entries/${entry.id}`}
                    className={styles.title}
                    onClick={(event) => {
                      if (!isPlainClick(event)) return
                      event.preventDefault()
                      showPanel({ state: 'loading', id: entry.id })
                    }}
                  >
                    {entry.title}
                  </Link>
                  <span className={styles.meta}>
                    {statusLabel(entry.status)}
                    {entry.scores[PRESTIGE] !== undefined &&
                      ` · ${divisionName(findAxis(kind.axes, PRESTIGE), entry.scores[PRESTIGE])}`}
                    {entry.rank !== undefined && ` #${entry.rank}`}
                    {entry.tags.length > 0 && ` · ${entry.tags.map(tagLabel).join(', ')}`}
                  </span>
                </div>
                <VisibilitySwitch
                  visibility={entry.visibility}
                  presence={presenceOf(entry, statusesShown)}
                  onChange={(visibility) => changeVisibility(entry, visibility)}
                />
              </li>
            ))}
          </ul>
        )}
      </div>
      <aside className={styles.panel}>
        {panel.state === 'loading' && <p className={styles.empty}>Opening…</p>}
        {panel.state === 'failed' && (
          <p className={styles.empty}>
            Could not open this entry.{' '}
            <button type="button" className={styles.quiet} onClick={() => showPanel(CLOSED)}>
              close
            </button>
          </p>
        )}
        {panel.state === 'open' && (
          <Editor
            key={panel.id}
            subject={panel.subject}
            kind={kind}
            knownTags={allTags}
            onClose={() => showPanel(CLOSED)}
            onChange={(change) => patchEntry(panel.id, change)}
          />
        )}
      </aside>
    </div>
  )
}

const Choices = <T extends string>({
  chosen,
  anyLabel,
  options,
  onChoose
}: {
  chosen: T | undefined
  anyLabel: string
  options: { value: T; label: string }[]
  onChoose: (value: T | undefined) => void
}) => (
  <nav className={styles.filters}>
    {[{ value: undefined, label: anyLabel }, ...options].map((o) => (
      <button
        key={o.label}
        type="button"
        className={styles.filter}
        aria-current={o.value === chosen ? 'true' : undefined}
        onClick={() => onChoose(o.value)}
      >
        {o.label}
      </button>
    ))}
  </nav>
)
