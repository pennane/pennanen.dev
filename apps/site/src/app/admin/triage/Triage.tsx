'use client'

import { useRouter } from 'next/navigation'
import { useCallback, useEffect, useMemo, useRef, useState } from 'react'
import { PRESTIGE } from '../../../lib/log/present'
import type { EntryChange, Hint, Kind, SteamSyncReport } from '../../../lib/log/types'
import {
  ANOTHER_SHOT,
  CardFace,
  prestigeDivisions,
  formatPlaytime,
  HINT_LABELS,
  IDLED,
  QueueHeader,
  SaveStatus,
  TagToggle,
  useKeys,
  type GameCard
} from '../_queue/card'
import styles from '../_queue/queue.module.css'
import { changeEntries, changeEntry, startSteamSync, steamSyncStatus } from '../actions'
import { useSaveQueue } from '../_queue/useSaveQueue'

type Filter = 'everything' | Hint | 'no-hint'

type Step = { phase: 'status' } | { phase: 'score'; status: string }

type Triaged = { cards: GameCard[] }

const UNTRIAGED = 'untriaged'
const NEVER_PLAYED = 'never-played'
const SYNC_POLL_MS = 3000

const TRIAGE_TAGS = [
  { tag: ANOTHER_SHOT, shortcut: 'a', label: 'worth another shot' },
  { tag: IDLED, shortcut: 'i', label: 'idled for cards' }
] as const

const BULK_FOR_HINT: Partial<Record<Hint, { label: string; change: EntryChange }>> = {
  'likely-idled': { label: 'never-played + idled', change: { status: NEVER_PLAYED, addTags: [IDLED] } },
  'likely-never-played': { label: NEVER_PLAYED, change: { status: NEVER_PLAYED } },
  'likely-not-a-game': { label: 'dismissed', change: { status: 'dismissed' } }
}

const matches = (filter: Filter) => (card: GameCard) =>
  filter === 'everything' ||
  (filter === 'no-hint' ? card.hint === undefined : card.hint === filter)

export const Triage = ({
  kind,
  cards,
  steamSync
}: {
  kind: Kind
  cards: GameCard[]
  steamSync?: SteamSyncReport
}) => {
  const presets = useMemo(() => prestigeDivisions(kind), [kind])
  const [queue, setQueue] = useState(cards)
  const [history, setHistory] = useState<Triaged[]>([])
  const [filter, setFilter] = useState<Filter>('everything')
  const [step, setStep] = useState<Step>({ phase: 'status' })
  const [bulkArmed, setBulkArmed] = useState<Hint>()
  const [tagsToAdd, setTagsToAdd] = useState<string[]>([])
  const saves = useSaveQueue()
  const settledIds = useRef(new Set<number>())

  useEffect(() => {
    setQueue((current) => {
      const known = new Set(current.map((c) => c.id))
      const arrived = cards.filter((c) => !known.has(c.id) && !settledIds.current.has(c.id))
      return [...current, ...arrived]
    })
  }, [cards])

  const visible = useMemo(() => queue.filter(matches(filter)), [queue, filter])
  const current = visible[0]
  const statuses = kind.statuses.filter((s) => s !== UNTRIAGED)

  const countByFilter = useMemo(() => {
    const counts: Record<Filter, number> = {
      everything: queue.length,
      'no-hint': 0,
      'likely-played': 0,
      'likely-idled': 0,
      'likely-never-played': 0,
      'likely-not-a-game': 0
    }
    for (const card of queue) counts[card.hint ?? 'no-hint']++
    return counts
  }, [queue])

  const settle = useCallback(
    (triaged: GameCard[], label: string, save: () => Promise<unknown>) => {
      const ids = new Set(triaged.map((c) => c.id))
      ids.forEach((id) => settledIds.current.add(id))
      setQueue((q) => q.filter((c) => !ids.has(c.id)))
      setHistory((h) => [...h, { cards: triaged }])
      setStep({ phase: 'status' })
      setBulkArmed(undefined)
      setTagsToAdd([])
      saves.enqueue(label, save)
    },
    [saves]
  )

  const commit = useCallback(
    (status: string, prestige?: string) => {
      if (!current) return
      settle([current], `${current.title} → ${status}`, () =>
        changeEntry(current.id, {
          status,
          ...(prestige === undefined ? {} : { scores: { [PRESTIGE]: prestige } }),
          ...(tagsToAdd.length > 0 ? { addTags: tagsToAdd } : {})
        })
      )
    },
    [current, settle, tagsToAdd]
  )

  const choose = useCallback(
    (status: string) => {
      if (kind.statusesToScore.includes(status)) setStep({ phase: 'score', status })
      else commit(status)
    },
    [kind.statusesToScore, commit]
  )

  const skip = useCallback(() => {
    if (!current) return
    setQueue((q) => [...q.filter((c) => c.id !== current.id), current])
    setStep({ phase: 'status' })
    setTagsToAdd([])
  }, [current])

  const undo = useCallback(() => {
    const last = history.at(-1)
    if (!last) return
    setHistory((h) => h.slice(0, -1))
    setQueue((q) => [...last.cards, ...q])
    setStep({ phase: 'status' })
    const ids = last.cards.map((c) => c.id)
    ids.forEach((id) => settledIds.current.delete(id))
    saves.enqueue(`undo ${last.cards.length}`, () =>
      changeEntries(ids, {
        status: UNTRIAGED,
        scores: { [PRESTIGE]: null },
        removeTags: TRIAGE_TAGS.map((t) => t.tag)
      })
    )
  }, [history, saves])

  const bulk = useCallback(
    (hint: Hint) => {
      const action = BULK_FOR_HINT[hint]
      if (!action) return
      if (bulkArmed !== hint) {
        setBulkArmed(hint)
        return
      }
      const triaged = queue.filter((c) => c.hint === hint)
      settle(triaged, `${triaged.length} → ${action.label}`, () =>
        changeEntries(
          triaged.map((c) => c.id),
          action.change
        )
      )
    },
    [bulkArmed, queue, settle]
  )

  const toggleTag = useCallback(
    (tag: string) =>
      setTagsToAdd((tags) => (tags.includes(tag) ? tags.filter((t) => t !== tag) : [...tags, tag])),
    []
  )

  const keys = useMemo(() => {
    const handlers: Record<string, () => void> = {
      s: skip,
      z: undo
    }
    TRIAGE_TAGS.forEach((t) => (handlers[t.shortcut] = () => toggleTag(t.tag)))
    if (step.phase === 'status') {
      statuses.forEach((status, i) => (handlers[String(i + 1)] = () => choose(status)))
    } else {
      presets.forEach(
        (preset, i) => (handlers[String(i + 1)] = () => commit(step.status, preset.key))
      )
      handlers['0'] = () => commit(step.status)
      handlers.Escape = () => setStep({ phase: 'status' })
    }
    return handlers
  }, [presets, step, statuses, choose, commit, skip, undo, toggleTag])
  useKeys(keys)

  const bulkAction = filter !== 'everything' && filter !== 'no-hint' && BULK_FOR_HINT[filter]

  return (
    <div className={styles.queue}>
      <QueueHeader current="triage" count={queue.length}>
        <SteamSyncButton initial={steamSync} />
      </QueueHeader>

      <nav className={styles.filters}>
        {(
          [
            'everything',
            'likely-played',
            'no-hint',
            'likely-idled',
            'likely-never-played',
            'likely-not-a-game'
          ] as const
        ).map((f) => (
          <button
            key={f}
            className={styles.filter}
            aria-pressed={filter === f}
            onClick={() => {
              setFilter(f)
              setBulkArmed(undefined)
            }}
          >
            {f === 'everything' ? 'all' : f === 'no-hint' ? 'no hint' : HINT_LABELS[f]}{' '}
            <span className={styles.count}>{countByFilter[f]}</span>
          </button>
        ))}
      </nav>

      {bulkAction && countByFilter[filter] > 0 && (
        <section className={styles.bulkPanel}>
          <button
            className={styles.bulk}
            data-armed={bulkArmed === filter}
            onClick={() => bulk(filter as Hint)}
          >
            {bulkArmed === filter
              ? `Tap again: mark these ${countByFilter[filter]} as ${bulkAction.label}`
              : `Mark all ${countByFilter[filter]} as ${bulkAction.label}…`}
          </button>
          {bulkArmed === filter && (
            <>
              <ul className={styles.bulkList}>
                {visible.map((card) => (
                  <li key={card.id}>
                    <span>{card.title}</span>
                    {card.playtimeMinutes !== undefined && (
                      <span className={styles.count}>{formatPlaytime(card.playtimeMinutes)}</span>
                    )}
                  </li>
                ))}
              </ul>
              <button className={styles.quiet} onClick={() => setBulkArmed(undefined)}>
                cancel
              </button>
            </>
          )}
        </section>
      )}

      {bulkArmed ? null : current ? (
        <article className={styles.card} key={current.id}>
          <CardFace card={current} badge={current.hint && HINT_LABELS[current.hint]} />
          {step.phase === 'status' ? (
            <div className={styles.choices}>
              {statuses.map((status, i) => (
                <button key={status} className={styles.choice} onClick={() => choose(status)}>
                  <kbd>{i + 1}</kbd> {status}
                </button>
              ))}
            </div>
          ) : (
            <div className={styles.choices}>
              <p className={styles.prompt}>
                {step.status} — how much did you enjoy it?
              </p>
              {presets.map((preset, i) => (
                <button
                  key={preset.name}
                  className={styles.choice}
                  data-preset={preset.name}
                  onClick={() => commit(step.status, preset.key)}
                >
                  <kbd>{i + 1}</kbd> {preset.name}
                </button>
              ))}
              <button className={styles.quiet} onClick={() => commit(step.status)}>
                <kbd>0</kbd> no score
              </button>
              <button className={styles.quiet} onClick={() => setStep({ phase: 'status' })}>
                <kbd>esc</kbd> back
              </button>
            </div>
          )}
          <div className={styles.toggles}>
            {TRIAGE_TAGS.map((t) => (
              <TagToggle
                key={t.tag}
                {...t}
                pressed={tagsToAdd.includes(t.tag)}
                onToggle={toggleTag}
              />
            ))}
          </div>
        </article>
      ) : (
        <p className={styles.empty}>
          {queue.length === 0 ? 'Nothing left to triage.' : 'Nothing here. Try another filter.'}
        </p>
      )}

      <footer className={styles.footer}>
        <button className={styles.quiet} onClick={skip} disabled={!current}>
          <kbd>s</kbd> skip
        </button>
        <button className={styles.quiet} onClick={undo} disabled={history.length === 0}>
          <kbd>z</kbd> undo
        </button>
        <SaveStatus state={saves.state} />
      </footer>
    </div>
  )
}

const SteamSyncButton = ({ initial }: { initial?: SteamSyncReport }) => {
  const [report, setReport] = useState(initial)
  const router = useRouter()

  useEffect(() => {
    if (report?.state !== 'syncing') return
    const timer = setInterval(async () => {
      const next = await steamSyncStatus()
      setReport(next)
      if (next.state !== 'syncing') router.refresh()
    }, SYNC_POLL_MS)
    return () => clearInterval(timer)
  }, [report?.state, router])

  if (!initial) return null
  const syncing = report?.state === 'syncing'
  return (
    <button
      className={styles.quiet}
      disabled={syncing}
      title={report?.problems.at(-1)}
      onClick={async () => setReport(await startSteamSync())}
    >
      {syncing
        ? `syncing steam… ${report.gamesSeen ? `${report.gamesAdded} new` : ''}`
        : report?.state === 'failed'
          ? 'steam sync failed, retry'
          : 'sync steam'}
    </button>
  )
}
