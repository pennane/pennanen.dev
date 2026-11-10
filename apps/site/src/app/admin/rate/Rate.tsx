'use client'

import { useCallback, useMemo, useState } from 'react'
import type { Kind } from '../../../lib/log/types'
import { changeEntry } from '../actions'
import {
  ANOTHER_SHOT,
  CardFace,
  ENJOYMENT,
  enjoymentDivisions,
  QueueHeader,
  SaveStatus,
  TagToggle,
  useKeys,
  type GameCard
} from '../_queue/card'
import styles from '../_queue/queue.module.css'
import { useSaveQueue } from '../_queue/useSaveQueue'

type Rated = { card: GameCard; blurb: string; addedTags: string[] }

const RATE_TAGS = [{ tag: ANOTHER_SHOT, shortcut: 'a', label: 'worth another shot' }] as const

export const Rate = ({ kind, cards }: { kind: Kind; cards: GameCard[] }) => {
  const presets = useMemo(() => enjoymentDivisions(kind), [kind])
  const [queue, setQueue] = useState(cards)
  const [history, setHistory] = useState<Rated[]>([])
  const [blurb, setBlurb] = useState('')
  const [tagsToAdd, setTagsToAdd] = useState<string[]>([])
  const saves = useSaveQueue()
  const current = queue[0]

  const reset = () => {
    setBlurb('')
    setTagsToAdd([])
  }

  const rate = useCallback(
    (enjoyment: string) => {
      if (!current) return
      const written = blurb.trim()
      const addedTags = tagsToAdd.filter((t) => !current.tags.includes(t))
      setQueue((q) => q.slice(1))
      setHistory((h) => [...h, { card: current, blurb: written, addedTags }])
      reset()
      saves.enqueue(`${current.title} → ${enjoyment}`, () =>
        changeEntry(current.id, {
          scores: { [ENJOYMENT]: enjoyment },
          ...(written ? { blurb: written } : {}),
          ...(addedTags.length > 0 ? { addTags: addedTags } : {})
        })
      )
    },
    [current, blurb, tagsToAdd, saves]
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
    reset()
    saves.enqueue(`undo ${last.card.title}`, () =>
      changeEntry(last.card.id, {
        scores: { [ENJOYMENT]: null },
        ...(last.blurb ? { blurb: '' } : {}),
        ...(last.addedTags.length > 0 ? { removeTags: last.addedTags } : {})
      })
    )
  }, [history, saves])

  const toggleTag = useCallback(
    (tag: string) =>
      setTagsToAdd((tags) => (tags.includes(tag) ? tags.filter((t) => t !== tag) : [...tags, tag])),
    []
  )

  const keys = useMemo(() => {
    const handlers: Record<string, () => void> = { s: skip, z: undo }
    presets.forEach((preset, i) => (handlers[String(i + 1)] = () => rate(preset.key)))
    RATE_TAGS.forEach((t) => (handlers[t.shortcut] = () => toggleTag(t.tag)))
    return handlers
  }, [presets, skip, undo, rate, toggleTag])
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
          <div className={styles.choices}>
            <p className={styles.prompt}>how much did you enjoy it?</p>
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
