'use client'

import Image from 'next/image'
import Link from 'next/link'
import { useMemo, useState } from 'react'
import {
  presenceOf,
  statusLabel,
  tagIsWithin,
  tagLabel,
  tagWithAncestors
} from '../../../lib/log/present'
import type { Visibility } from '../../../lib/log/types'
import { changeEntry } from '../actions'
import styles from './entries.module.css'
import { VISIBILITY_LABELS, VisibilitySwitch } from './VisibilitySwitch'

export type ListedEntry = {
  id: number
  title: string
  cover: string | undefined
  status: string
  headlineScore: number | undefined
  visibility: Visibility
  tags: string[]
}

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

const syncUrl = (filter: Filter) => {
  const url = new URL(window.location.href)
  for (const [name, value] of Object.entries(filter)) {
    if (value) url.searchParams.set(name, value)
    else url.searchParams.delete(name)
  }
  window.history.replaceState(null, '', url)
}

export const Entries = ({
  entries: initialEntries,
  statuses,
  statusesShown,
  initial
}: {
  entries: ListedEntry[]
  statuses: string[]
  statusesShown: string[]
  initial: Filter
}) => {
  const [entries, setEntries] = useState(initialEntries)
  const [filter, setFilter] = useState(initial)
  const [query, setQuery] = useState('')

  const choose = (change: Partial<Filter>) => {
    const next = { ...filter, ...change }
    setFilter(next)
    syncUrl(next)
  }

  const tags = useMemo(() => tagsByUse(initialEntries), [initialEntries])
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
    <>
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
            <li key={entry.id} className={styles.row}>
              <div className={styles.thumb}>
                {entry.cover && (
                  <Image alt="" src={entry.cover} width={92} height={43} unoptimized loading="lazy" />
                )}
              </div>
              <div className={styles.rowBody}>
                <Link href={`/admin/entries/${entry.id}`} className={styles.title}>
                  {entry.title}
                </Link>
                <span className={styles.meta}>
                  {statusLabel(entry.status)}
                  {entry.headlineScore !== undefined && ` · ${entry.headlineScore}`}
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
    </>
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
