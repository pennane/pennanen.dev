'use client'

import { useCallback, useMemo, useState } from 'react'
import { PRESTIGE } from '../../../lib/log/present'
import type { Kind } from '../../../lib/log/types'
import { changeEntry, placeEntry } from '../actions'
import { Compare, compareKeys, useRanked, type Comparing } from '../_queue/Compare'
import type { Placement, Ranked, Search } from '../_queue/placing'
import {
  ANOTHER_SHOT,
  CardFace,
  prestigeDivisions,
  prestigeName,
  QueueHeader,
  SaveStatus,
  TagToggle,
  useKeys,
  type GameCard
} from '../_queue/card'
import styles from '../_queue/queue.module.css'
import { useSaveQueue } from '../_queue/useSaveQueue'

type Rated = { card: GameCard; blurb: string; addedTags: string[] }

type Step = { phase: 'division' } | { phase: 'compare'; search: Search }

const CHOOSING: Step = { phase: 'division' }

const RATE_TAGS = [{ tag: ANOTHER_SHOT, shortcut: 'a', label: 'worth another shot' }] as const

export const Rate = ({
  kind,
  cards,
  ranked: initialRanked
}: {
  kind: Kind
  cards: GameCard[]
  ranked: Ranked
}) => {
  const presets = useMemo(() => prestigeDivisions(kind), [kind])
  const { ranked, place, unplace, searchFrom } = useRanked(initialRanked)
  const [step, setStep] = useState<Step>(CHOOSING)
  const [queue, setQueue] = useState(cards)
  const [history, setHistory] = useState<Rated[]>([])
  const [blurb, setBlurb] = useState('')
  const [tagsToAdd, setTagsToAdd] = useState<string[]>([])
  const saves = useSaveQueue()
  const current = queue[0]

  const reset = () => {
    setBlurb('')
    setTagsToAdd([])
    setStep(CHOOSING)
  }

  const commit = useCallback(
    (placement: Placement) => {
      if (!current) return
      const written = blurb.trim()
      const addedTags = tagsToAdd.filter((t) => !current.tags.includes(t))
      setQueue((q) => q.slice(1))
      setHistory((h) => [...h, { card: current, blurb: written, addedTags }])
      place(current, placement)
      reset()
      const label = `${current.title} → ${prestigeName(kind, placement.division)} #${placement.index + 1}`
      saves.enqueue(label, async () => {
        await changeEntry(current.id, {
          scores: { [PRESTIGE]: placement.division },
          ...(written ? { blurb: written } : {}),
          ...(addedTags.length > 0 ? { addTags: addedTags } : {})
        })
        await placeEntry(current.id, placement.division, placement.index)
      })
    },
    [current, blurb, tagsToAdd, saves, place, kind]
  )

  const rate = useCallback(
    (division: string) => {
      if (!current) return
      const search = searchFrom(division, current.id)
      if (search.high === 0) commit({ division, index: 0 })
      else setStep({ phase: 'compare', search })
    },
    [current, searchFrom, commit]
  )

  const comparing = useMemo<Comparing | undefined>(
    () =>
      current && step.phase === 'compare'
        ? {
            ranked,
            search: step.search,
            id: current.id,
            divisionName: prestigeName(kind, step.search.division),
            onSearch: (search) => setStep({ phase: 'compare', search }),
            onPlace: commit,
            onBack: () => setStep(CHOOSING)
          }
        : undefined,
    [current, step, ranked, kind, commit]
  )

  const skip = useCallback(() => {
    if (!current) return
    setQueue((q) => [...q.slice(1), current])
    reset()
  }, [current])

  const undo = useCallback(() => {
    const last = history.at(-1)
    if (!last) return
    setHistory((h) => h.slice(0, -1))
    setQueue((q) => [last.card, ...q])
    unplace(last.card.id)
    reset()
    saves.enqueue(`undo ${last.card.title}`, () =>
      changeEntry(last.card.id, {
        scores: { [PRESTIGE]: null },
        ...(last.blurb ? { blurb: '' } : {}),
        ...(last.addedTags.length > 0 ? { removeTags: last.addedTags } : {})
      })
    )
  }, [history, saves, unplace])

  const toggleTag = useCallback(
    (tag: string) =>
      setTagsToAdd((tags) => (tags.includes(tag) ? tags.filter((t) => t !== tag) : [...tags, tag])),
    []
  )

  const keys = useMemo(() => {
    const handlers: Record<string, () => void> = { s: skip, z: undo }
    if (comparing) Object.assign(handlers, compareKeys(comparing))
    else presets.forEach((preset, i) => (handlers[String(i + 1)] = () => rate(preset.key)))
    RATE_TAGS.forEach((t) => (handlers[t.shortcut] = () => toggleTag(t.tag)))
    return handlers
  }, [presets, skip, undo, rate, toggleTag, comparing])
  useKeys(keys)

  return (
    <div className={styles.queue}>
      <QueueHeader current="rate" count={queue.length} />

      {current ? (
        <article className={styles.card} key={current.id}>
          <CardFace card={current} badge={current.status} />
          <input
            className={styles.blurb}
            value={blurb}
            onChange={(event) => setBlurb(event.target.value)}
            placeholder="one line for the public page (optional)"
            enterKeyHint="done"
          />
          {comparing ? (
            <Compare {...comparing} />
          ) : (
            <div className={styles.choices}>
              <p className={styles.prompt}>which tier?</p>
              {presets.map((preset, i) => (
                <button
                  key={preset.name}
                  className={styles.choice}
                  data-preset={preset.name}
                  onClick={() => rate(preset.key)}
                >
                  <kbd>{i + 1}</kbd> {preset.name}
                </button>
              ))}
            </div>
          )}
          <div className={styles.toggles}>
            {RATE_TAGS.map((t) => (
              <TagToggle
                key={t.tag}
                {...t}
                pressed={tagsToAdd.includes(t.tag) || current.tags.includes(t.tag)}
                onToggle={toggleTag}
              />
            ))}
          </div>
        </article>
      ) : (
        <p className={styles.empty}>Everything you played has a score.</p>
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
