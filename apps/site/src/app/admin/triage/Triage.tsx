'use client'

import Image from 'next/image'
import { useRouter } from 'next/navigation'
import { useCallback, useEffect, useMemo, useRef, useState } from 'react'
import type { Hint, Kind, SteamSyncReport } from '../../../lib/log/types'
import { changeEntries, changeEntry, startSteamSync, steamSyncStatus } from '../actions'
import styles from './triage.module.css'
import { useSaveQueue, type SaveQueueState } from './useSaveQueue'

export type TriageCard = {
  id: number
  title: string
  cover?: string
  playtimeMinutes?: number
  lastPlayed?: string
  achievements?: { unlocked: number; total: number }
  steamCategories: string[]
  hint?: Hint
}

type Filter = 'everything' | Hint | 'unsure'

type Step = { phase: 'status' } | { phase: 'score'; status: string }

type Triaged = { cards: TriageCard[] }

const UNTRIAGED = 'untriaged'
const ENJOYMENT = 'enjoyment'
const STEAM_TRADING_CARDS = 'Steam Trading Cards'
const SYNC_POLL_MS = 3000

const ENJOYMENT_PRESETS = [
  { label: 'loved', value: 92 },
  { label: 'liked', value: 80 },
  { label: 'meh', value: 60 },
  { label: 'disliked', value: 35 }
] as const

const HINT_LABELS: Record<Hint, string> = {
  'likely-played': 'looks played',
  'likely-idled': 'looks idled',
  'likely-never-played': 'never launched'
}

const BULK_STATUS_FOR_HINT: Partial<Record<Hint, string>> = {
  'likely-idled': 'idled',
  'likely-never-played': 'never-played'
}

const matches = (filter: Filter) => (card: TriageCard) =>
  filter === 'everything' ||
  (filter === 'unsure' ? card.hint === undefined : card.hint === filter)

const formatPlaytime = (minutes: number) => {
  if (minutes < 60) return `${minutes} min`
  const hours = Math.floor(minutes / 60)
  return minutes % 60 === 0 ? `${hours} h` : `${hours} h ${minutes % 60} min`
}

const useKeys = (handlers: Record<string, () => void>) => {
  useEffect(() => {
    const onKey = (event: KeyboardEvent) => {
      if (event.metaKey || event.ctrlKey || event.altKey) return
      if (event.target instanceof HTMLInputElement) return
      const handler = handlers[event.key]
      if (!handler) return
      event.preventDefault()
      handler()
    }
    window.addEventListener('keydown', onKey)
    return () => window.removeEventListener('keydown', onKey)
  }, [handlers])
}

export const Triage = ({
  kind,
  cards,
  steamSync
}: {
  kind: Kind
  cards: TriageCard[]
  steamSync?: SteamSyncReport
}) => {
  const [queue, setQueue] = useState(cards)
  const [history, setHistory] = useState<Triaged[]>([])
  const [filter, setFilter] = useState<Filter>('everything')
  const [step, setStep] = useState<Step>({ phase: 'status' })
  const [bulkArmed, setBulkArmed] = useState<Hint>()
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
      unsure: 0,
      'likely-played': 0,
      'likely-idled': 0,
      'likely-never-played': 0
    }
    for (const card of queue) counts[card.hint ?? 'unsure']++
    return counts
  }, [queue])

  const settle = useCallback(
    (triaged: TriageCard[], label: string, save: () => Promise<unknown>) => {
      const ids = new Set(triaged.map((c) => c.id))
      ids.forEach((id) => settledIds.current.add(id))
      setQueue((q) => q.filter((c) => !ids.has(c.id)))
      setHistory((h) => [...h, { cards: triaged }])
      setStep({ phase: 'status' })
      setBulkArmed(undefined)
      saves.enqueue(label, save)
    },
    [saves]
  )

  const commit = useCallback(
    (status: string, enjoyment?: number) => {
      if (!current) return
      settle([current], `${current.title} → ${status}`, () =>
        changeEntry(current.id, {
          status,
          ...(enjoyment === undefined ? {} : { scores: { [ENJOYMENT]: enjoyment } })
        })
      )
    },
    [current, settle]
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
      changeEntries(ids, { status: UNTRIAGED, scores: { [ENJOYMENT]: null } })
    )
  }, [history, saves])

  const bulk = useCallback(
    (hint: Hint) => {
      const status = BULK_STATUS_FOR_HINT[hint]
      if (!status) return
      if (bulkArmed !== hint) {
        setBulkArmed(hint)
        return
      }
      const triaged = queue.filter((c) => c.hint === hint)
      settle(triaged, `${triaged.length} → ${status}`, () =>
        changeEntries(
          triaged.map((c) => c.id),
          { status }
        )
      )
    },
    [bulkArmed, queue, settle]
  )

  const keys = useMemo(() => {
    const handlers: Record<string, () => void> = { s: skip, z: undo }
    if (step.phase === 'status') {
      statuses.forEach((status, i) => (handlers[String(i + 1)] = () => choose(status)))
    } else {
      ENJOYMENT_PRESETS.forEach(
        (preset, i) => (handlers[String(i + 1)] = () => commit(step.status, preset.value))
      )
      handlers['0'] = () => commit(step.status)
      handlers.Escape = () => setStep({ phase: 'status' })
    }
    return handlers
  }, [step, statuses, choose, commit, skip, undo])
  useKeys(keys)

  const bulkStatus = filter !== 'everything' && filter !== 'unsure' && BULK_STATUS_FOR_HINT[filter]

  return (
    <div className={styles.triage}>
      <header className={styles.header}>
        <h1>
          triage <span className={styles.count}>{queue.length}</span>
        </h1>
        <SteamSyncButton initial={steamSync} />
      </header>

      <nav className={styles.filters}>
        {(['everything', 'likely-played', 'unsure', 'likely-idled', 'likely-never-played'] as const).map((f) => (
          <button
            key={f}
            className={styles.filter}
            aria-pressed={filter === f}
            onClick={() => {
              setFilter(f)
              setBulkArmed(undefined)
            }}
          >
            {f === 'everything' ? 'all' : f === 'unsure' ? 'unsure' : HINT_LABELS[f]}{' '}
            <span className={styles.count}>{countByFilter[f]}</span>
          </button>
        ))}
      </nav>

      {bulkStatus && countByFilter[filter] > 0 && (
        <button
          className={styles.bulk}
          data-armed={bulkArmed === filter}
          onClick={() => bulk(filter as Hint)}
        >
          {bulkArmed === filter
            ? `Tap again: mark ${countByFilter[filter]} as ${bulkStatus}`
            : `Mark all ${countByFilter[filter]} as ${bulkStatus}`}
        </button>
      )}

      {current ? (
        <article className={styles.card} key={current.id}>
          <CardFace card={current} />
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
              {ENJOYMENT_PRESETS.map((preset, i) => (
                <button
                  key={preset.label}
                  className={styles.choice}
                  data-preset={preset.label}
                  onClick={() => commit(step.status, preset.value)}
                >
                  <kbd>{i + 1}</kbd> {preset.label}
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

const CardFace = ({ card }: { card: TriageCard }) => {
  const facts = [
    card.playtimeMinutes !== undefined && formatPlaytime(card.playtimeMinutes),
    card.achievements && `${card.achievements.unlocked}/${card.achievements.total} achievements`,
    card.lastPlayed && `last played ${card.lastPlayed}`,
    card.steamCategories.includes(STEAM_TRADING_CARDS) && 'drops cards'
  ].filter(Boolean)
  return (
    <>
      <div className={styles.cover}>
        {card.cover ? (
          <Image src={card.cover} alt="" width={460} height={215} unoptimized priority />
        ) : (
          <span className={styles.noCover}>{card.title}</span>
        )}
      </div>
      <h2 className={styles.title}>{card.title}</h2>
      <p className={styles.facts}>
        {card.hint && <span className={styles.hint}>{HINT_LABELS[card.hint]}</span>}
        {facts.join(' · ')}
      </p>
    </>
  )
}

const SaveStatus = ({ state }: { state: SaveQueueState }) => {
  if (state.kind === 'saved') return <span className={styles.saved}>saved</span>
  if (state.kind === 'saving') return <span className={styles.saving}>saving {state.unsaved}…</span>
  return (
    <span className={styles.retrying} title={state.error}>
      {state.unsaved} unsaved, retrying ({state.label})
    </span>
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
