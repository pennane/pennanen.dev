'use client'

import Link from 'next/link'
import { useCallback, useEffect, useMemo, useRef, useState } from 'react'
import { PRESTIGE } from '../../../lib/log/present'
import { IGDB_MATCH, STEAM_SYNC } from '../../../lib/log/syncs'
import type { Kind, SyncReport } from '../../../lib/log/types'
import { useSync } from '../_sync/useSync'
import {
  ANOTHER_SHOT,
  CardFace,
  prestigeDivisions,
  prestigeName,
  formatPlaytime,
  IDLED,
  QueueHeader,
  ruleLabel,
  SaveStatus,
  TagToggle,
  useKeys
} from '../_queue/card'
import styles from '../_queue/queue.module.css'
import { acceptancesOf } from '../_queue/accept'
import { catalogueOf, type Catalogue } from '../_queue/catalogue'
import { IGDBSearch } from '../_queue/IGDBSearch'
import { TMDBSearch } from '../_queue/TMDBSearch'
import type { TriageCard } from '../_queue/toCard'
import {
  changeEntries,
  changeEntry,
  matchIGDB,
  matchIGDBNone,
  matchTMDB,
  matchTMDBNone,
  placeEntry
} from '../actions'
import { Compare, compareKeys, useRanked, type Comparing } from '../_queue/Compare'
import type { Placement, Ranked, Search } from '../_queue/placing'
import { useSaveQueue } from '../_queue/useSaveQueue'

export type KindQueue = { kind: string; untriaged: number; needingMatch: number }

type Filter =
  | { by: 'everything' }
  | { by: 'no-suggestion' }
  | { by: 'needs-match' }
  | { by: 'rule'; rule: string }

const EVERYTHING: Filter = { by: 'everything' }

type Step =
  | { phase: 'status' }
  | { phase: 'score'; status: string }
  | { phase: 'compare'; status: string; search: Search }

type Triaged = { card: TriageCard; addedTags: string[] }[]

const UNTRIAGED = 'untriaged'

const TRIAGE_TAGS = [
  { tag: ANOTHER_SHOT, shortcut: 'a', label: 'worth another shot' },
  { tag: IDLED, shortcut: 'i', label: 'idled for cards' }
] as const

const triagedAs = (card: TriageCard, tags: string[]) => ({
  card,
  addedTags: tags.filter((tag) => !card.tags.includes(tag))
})

const cardKey = (card: TriageCard) => `${card.shape}:${card.id}`

const filterKey = (filter: Filter) => (filter.by === 'rule' ? `rule:${filter.rule}` : filter.by)

const matches = (filter: Filter) => (card: TriageCard) => {
  switch (filter.by) {
    case 'everything':
      return true
    case 'needs-match':
      return card.shape === 'match'
    case 'no-suggestion':
      return card.shape === 'untriaged' && card.suggestions.length === 0
    case 'rule':
      return card.shape === 'untriaged' && card.suggestions.some((s) => s.rule === filter.rule)
  }
}

const rulesOf = (cards: TriageCard[]) =>
  [...new Set(cards.flatMap((card) => card.suggestions.map((s) => s.rule)))].sort()

export const Triage = ({
  kind,
  kindQueues,
  cards,
  ranked: initialRanked,
  steamSync,
  matchSync
}: {
  kind: Kind
  kindQueues: KindQueue[]
  cards: TriageCard[]
  ranked: Ranked
  steamSync?: SyncReport
  matchSync?: SyncReport
}) => {
  const catalogue = catalogueOf(kind)
  const presets = useMemo(() => prestigeDivisions(kind), [kind])
  const { ranked, place, unplace, searchFrom } = useRanked(initialRanked)
  const [queue, setQueue] = useState(cards)
  const [history, setHistory] = useState<Triaged[]>([])
  const [filter, setFilter] = useState<Filter>(EVERYTHING)
  const [step, setStep] = useState<Step>({ phase: 'status' })
  const [bulkArmed, setBulkArmed] = useState<string>()
  const [tagsToAdd, setTagsToAdd] = useState<string[]>([])
  const saves = useSaveQueue()
  const settledKeys = useRef(new Set<string>())

  useEffect(() => {
    setQueue((current) => {
      const known = new Set(current.map(cardKey))
      const arrived = cards.filter(
        (c) => !known.has(cardKey(c)) && !settledKeys.current.has(cardKey(c))
      )
      return [...current, ...arrived]
    })
  }, [cards])

  const visible = useMemo(() => queue.filter(matches(filter)), [queue, filter])
  const current = visible[0]
  const statuses = kind.statuses.filter((s) => s !== UNTRIAGED)

  const countOf = useCallback((filter: Filter) => queue.filter(matches(filter)).length, [queue])
  const rules = useMemo(() => rulesOf(queue), [queue])

  const settle = useCallback(
    (settled: TriageCard[], label: string, save: () => Promise<unknown>) => {
      const keys = new Set(settled.map(cardKey))
      keys.forEach((key) => settledKeys.current.add(key))
      setQueue((q) => q.filter((c) => !keys.has(cardKey(c))))
      setStep({ phase: 'status' })
      setBulkArmed(undefined)
      setTagsToAdd([])
      saves.enqueue(label, save)
    },
    [saves]
  )

  const settleTriaged = useCallback(
    (triaged: Triaged, label: string, save: () => Promise<unknown>) => {
      setHistory((h) => [...h, triaged])
      settle(
        triaged.map((t) => t.card),
        label,
        save
      )
    },
    [settle]
  )

  const commit = useCallback(
    (status: string, placement?: Placement) => {
      if (current?.shape !== 'untriaged') return
      if (placement) place(current, placement)
      settleTriaged([triagedAs(current, tagsToAdd)], `${current.title} → ${status}`, async () => {
        await changeEntry(current.id, {
          status,
          ...(placement ? { scores: { [PRESTIGE]: placement.division } } : {}),
          ...(tagsToAdd.length > 0 ? { addTags: tagsToAdd } : {})
        })
        if (placement) await placeEntry(current.id, placement.division, placement.index)
      })
    },
    [current, settleTriaged, tagsToAdd, place]
  )

  const score = useCallback(
    (status: string, division: string) => {
      if (current?.shape !== 'untriaged') return
      const search = searchFrom(division, current.id)
      if (search.high === 0) commit(status, { division, index: 0 })
      else setStep({ phase: 'compare', status, search })
    },
    [current, searchFrom, commit]
  )

  const comparing = useMemo<Comparing | undefined>(
    () =>
      current?.shape === 'untriaged' && step.phase === 'compare'
        ? {
            ranked,
            search: step.search,
            id: current.id,
            divisionName: prestigeName(kind, step.search.division),
            onSearch: (search) => setStep({ phase: 'compare', status: step.status, search }),
            onPlace: (placement) => commit(step.status, placement),
            onBack: () => setStep({ phase: 'score', status: step.status })
          }
        : undefined,
    [current, step, ranked, kind, commit]
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
    setQueue((q) => [...q.filter((c) => cardKey(c) !== cardKey(current)), current])
    setStep({ phase: 'status' })
    setTagsToAdd([])
  }, [current])

  const undo = useCallback(() => {
    const last = history.at(-1)
    if (!last) return
    setHistory((h) => h.slice(0, -1))
    setQueue((q) => [...last.map((t) => t.card), ...q])
    setStep({ phase: 'status' })
    last.forEach((t) => settledKeys.current.delete(cardKey(t.card)))
    last.forEach((t) => unplace(t.card.id))
    const undoing = new Map<string, { removeTags: string[]; ids: number[] }>()
    for (const { card, addedTags } of last) {
      const group = undoing.get(addedTags.join()) ?? { removeTags: addedTags, ids: [] }
      group.ids.push(card.id)
      undoing.set(addedTags.join(), group)
    }
    saves.enqueue(`undo ${last.length}`, async () => {
      for (const { removeTags, ids } of undoing.values())
        await changeEntries(ids, {
          status: UNTRIAGED,
          dismissal: 'kept',
          scores: { [PRESTIGE]: null },
          removeTags
        })
    })
  }, [history, saves, unplace])

  const bulk = useCallback(
    (rule: string) => {
      const acceptances = acceptancesOf(queue.filter(matches({ by: 'rule', rule })), rule)
      if (acceptances.length === 0) return
      if (bulkArmed !== rule) {
        setBulkArmed(rule)
        return
      }
      const triaged = acceptances.flatMap((a) =>
        a.cards.map((card) => triagedAs(card, a.change.addTags ?? []))
      )
      settleTriaged(triaged, `${triaged.length} → ${rule}`, async () => {
        for (const { change, cards: accepted } of acceptances)
          await changeEntries(
            accepted.map((c) => c.id),
            change
          )
      })
    },
    [bulkArmed, queue, settleTriaged]
  )

  const matched = useCallback(
    (card: TriageCard, label: string, save: () => Promise<unknown>) =>
      settle([card], `${card.title} → ${label}`, save),
    [settle]
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
    if (current?.shape !== 'untriaged') return handlers
    TRIAGE_TAGS.forEach((t) => (handlers[t.shortcut] = () => toggleTag(t.tag)))
    if (comparing) {
      Object.assign(handlers, compareKeys(comparing))
    } else if (step.phase === 'status') {
      statuses.forEach((status, i) => (handlers[String(i + 1)] = () => choose(status)))
    } else {
      presets.forEach(
        (preset, i) => (handlers[String(i + 1)] = () => score(step.status, preset.key))
      )
      handlers['0'] = () => commit(step.status)
      handlers.Escape = () => setStep({ phase: 'status' })
    }
    return handlers
  }, [presets, step, statuses, choose, commit, score, skip, undo, toggleTag, comparing, current])
  useKeys(keys)

  const acceptances = useMemo(
    () => (filter.by === 'rule' ? acceptancesOf(queue.filter(matches(filter)), filter.rule) : []),
    [queue, filter]
  )
  const acceptable = acceptances.flatMap((a) => a.cards)
  const acceptLabel = acceptances.map((a) => a.label).join(', ')
  const filters: { filter: Filter; label: string }[] = [
    { filter: EVERYTHING, label: 'all' },
    ...rules.map((rule) => ({ filter: { by: 'rule', rule } as const, label: ruleLabel(rule) })),
    { filter: { by: 'no-suggestion' }, label: 'no suggestion' },
    ...(countOf({ by: 'needs-match' }) > 0
      ? [{ filter: { by: 'needs-match' } as const, label: 'needs match' }]
      : [])
  ]

  return (
    <div className={styles.queue}>
      <QueueHeader current="triage" count={queue.length}>
        {kind.name === 'game' && <SteamSyncButton initial={steamSync} />}
        {kind.name === 'game' && <MatchSyncButton initial={matchSync} />}
      </QueueHeader>

      {kindQueues.length > 1 && (
        <nav className={styles.filters}>
          {kindQueues.map((k) => (
            <Link
              key={k.kind}
              href={`/admin/triage?kind=${k.kind}`}
              className={styles.filter}
              aria-current={k.kind === kind.name ? 'true' : undefined}
            >
              {k.kind} <span className={styles.count}>{k.untriaged + k.needingMatch}</span>
            </Link>
          ))}
        </nav>
      )}

      <nav className={styles.filters}>
        {filters.map(({ filter: f, label }) => (
          <button
            key={filterKey(f)}
            className={styles.filter}
            aria-pressed={filterKey(filter) === filterKey(f)}
            onClick={() => {
              setFilter(f)
              setBulkArmed(undefined)
            }}
          >
            {label} <span className={styles.count}>{countOf(f)}</span>
          </button>
        ))}
      </nav>

      {filter.by === 'rule' && acceptable.length > 0 && (
        <section className={styles.bulkPanel}>
          <button
            className={styles.bulk}
            data-armed={bulkArmed === filter.rule}
            onClick={() => bulk(filter.rule)}
          >
            {bulkArmed === filter.rule
              ? `Tap again: accept ${filter.rule} for these ${acceptable.length} (${acceptLabel})`
              : `Accept ${filter.rule} for all ${acceptable.length}…`}
          </button>
          {bulkArmed === filter.rule && (
            <>
              <ul className={styles.bulkList}>
                {acceptable.map((card) => (
                  <li key={cardKey(card)}>
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

      {bulkArmed ? null : current?.shape === 'match' && catalogue ? (
        <article className={styles.card} key={cardKey(current)}>
          <CardFace card={current} badge="needs match" />
          <div className={styles.matchPanel}>
            <MatchSearch
              catalogue={catalogue}
              card={current}
              onMatched={(label, save) => matched(current, label, save)}
            />
          </div>
        </article>
      ) : current?.shape === 'untriaged' ? (
        <article className={styles.card} key={cardKey(current)}>
          <CardFace
            card={current}
            badge={[...new Set(current.suggestions.map((s) => ruleLabel(s.rule)))].join(' · ')}
          />
          {step.phase === 'status' ? (
            <div className={styles.choices}>
              {statuses.map((status, i) => (
                <button key={status} className={styles.choice} onClick={() => choose(status)}>
                  <kbd>{i + 1}</kbd> {status}
                </button>
              ))}
            </div>
          ) : comparing ? (
            <Compare {...comparing} />
          ) : (
            <div className={styles.choices}>
              <p className={styles.prompt}>{step.status} — which tier?</p>
              {presets.map((preset, i) => (
                <button
                  key={preset.name}
                  className={styles.choice}
                  data-preset={preset.name}
                  onClick={() => score(step.status, preset.key)}
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

const MatchSearch = ({
  catalogue,
  card,
  onMatched
}: {
  catalogue: Catalogue
  card: TriageCard
  onMatched: (label: string, save: () => Promise<unknown>) => void
}) => (
  <>
    {catalogue.name === 'igdb' ? (
      <IGDBSearch
        initialQuery={card.title}
        onPick={(result) => onMatched(result.name, () => matchIGDB(card.id, result.igdbId))}
      />
    ) : (
      <TMDBSearch
        tmdbType={catalogue.tmdbType}
        initialQuery={card.title}
        onPick={(result) =>
          onMatched(result.name, () =>
            matchTMDB(card.id, { tmdbType: result.tmdbType, tmdbId: result.tmdbId })
          )
        }
      />
    )}
    <button
      type="button"
      className={styles.quiet}
      onClick={() =>
        onMatched('no match', () =>
          catalogue.name === 'igdb' ? matchIGDBNone(card.id) : matchTMDBNone(card.id)
        )
      }
    >
      {catalogue.name} doesn’t have it: no match
    </button>
  </>
)

const SteamSyncButton = ({ initial }: { initial?: SyncReport }) => {
  const { report, start } = useSync(STEAM_SYNC, initial)
  if (!initial) return null
  return (
    <button
      className={styles.quiet}
      disabled={report?.state === 'running'}
      title={report?.problems.at(-1)}
      onClick={start}
    >
      {steamSyncLabel(report)}
    </button>
  )
}

const steamSyncLabel = (report: SyncReport | undefined) => {
  switch (report?.state) {
    case 'running':
      return report.counts['games-seen']
        ? `syncing steam… ${report.counts['games-added'] ?? 0} new`
        : 'syncing steam…'
    case 'failed':
      return 'steam sync failed, retry'
    default:
      return 'sync steam'
  }
}

const MatchSyncButton = ({ initial }: { initial?: SyncReport }) => {
  const { report, start } = useSync(IGDB_MATCH, initial)
  if (!initial) return null
  return (
    <button
      className={styles.quiet}
      disabled={report?.state === 'running'}
      title={report?.problems.at(-1)}
      onClick={start}
    >
      {matchSyncLabel(report)}
    </button>
  )
}

const matchSyncLabel = (report: SyncReport | undefined) => {
  const matched = report?.counts.matched ?? 0
  const checked = report?.counts.checked ?? 0
  switch (report?.state) {
    case 'running':
      return `matching… ${matched} of ${checked}`
    case 'failed':
      return 'match failed, retry'
    case 'succeeded':
      return `last run matched ${matched} of ${checked} · run again`
    default:
      return 'match steam games'
  }
}
