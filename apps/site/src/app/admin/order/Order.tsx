'use client'

import Image from 'next/image'
import { useEffect, useMemo, useState } from 'react'
import { statusLabel } from '../../../lib/log/present'
import type { ListedEntry } from '../../../lib/log/types'
import { SaveStatus, useKeys } from '../_queue/card'
import { useSaveQueue } from '../_queue/useSaveQueue'
import { placeEntry } from '../actions'
import {
  acrossLine,
  moved,
  oneStep,
  ranksOf,
  spotBefore,
  spotOf,
  type Section,
  type Spot
} from './sections'
import styles from './order.module.css'

type Picked = { state: 'none' } | { state: 'picked'; id: number }

const NONE: Picked = { state: 'none' }
const DOWN = 1
const UP = -1

const rowId = (id: number) => `order-${id}`

export const Order = ({
  sections: initial,
  entries
}: {
  sections: Section[]
  entries: ListedEntry[]
}) => {
  const [sections, setSections] = useState(initial)
  const [picked, setPicked] = useState<Picked>(NONE)
  const [dragged, setDragged] = useState<number>()
  const saves = useSaveQueue()
  const byId = useMemo(() => new Map(entries.map((e) => [e.id, e])), [entries])
  const ranks = useMemo(() => ranksOf(sections), [sections])
  const flat = useMemo(() => sections.flatMap((s) => s.ids), [sections])

  const move = (id: number, to: Spot | undefined) => {
    const from = spotOf(sections, id)
    if (!to || !from || (from.division === to.division && from.index === to.index)) return
    setSections(moved(sections, id, to))
    saves.enqueue(`${byId.get(id)?.title} → ${to.division} #${to.index + 1}`, () =>
      placeEntry(id, to.division, to.index)
    )
  }

  const movePicked = (step: (from: Spot) => Spot | undefined) => () => {
    if (picked.state !== 'picked') return
    const from = spotOf(sections, picked.id)
    if (from) move(picked.id, step(from))
  }

  const pickNeighbour = (direction: number) => () => {
    const at = picked.state === 'picked' ? flat.indexOf(picked.id) : -1
    const next = flat[Math.min(Math.max(at + direction, 0), flat.length - 1)]
    if (next !== undefined) setPicked({ state: 'picked', id: next })
  }

  const keys = {
    j: movePicked((from) => oneStep(sections, from, DOWN)),
    k: movePicked((from) => oneStep(sections, from, UP)),
    J: movePicked((from) => acrossLine(sections, from, DOWN)),
    K: movePicked((from) => acrossLine(sections, from, UP)),
    ArrowDown: pickNeighbour(DOWN),
    ArrowUp: pickNeighbour(UP),
    Escape: () => setPicked(NONE)
  }
  useKeys(keys)

  const pickedId = picked.state === 'picked' ? picked.id : undefined
  useEffect(() => {
    if (pickedId !== undefined)
      document.getElementById(rowId(pickedId))?.scrollIntoView({ block: 'nearest' })
  }, [pickedId, sections])

  const dropOn = (to: Spot | undefined) => {
    if (dragged !== undefined) move(dragged, to)
    setDragged(undefined)
  }

  return (
    <div className={styles.order}>
      <p className={styles.hint}>
        drag, or pick one and use <kbd>j</kbd>/<kbd>k</kbd> to move a place, <kbd>J</kbd>/
        <kbd>K</kbd> to cross a division · <kbd>↑</kbd>/<kbd>↓</kbd> pick ·{' '}
        <SaveStatus state={saves.state} />
      </p>
      {sections.map((section) => (
        <section key={section.division.key} className={styles.section}>
          <h2
            className={styles.division}
            onDragOver={(event) => event.preventDefault()}
            onDrop={() => dropOn({ division: section.division.key, index: 0 })}
          >
            {section.division.name} <span className={styles.count}>{section.ids.length}</span>
          </h2>
          <ol className={styles.rows}>
            {section.ids.map((id) => {
              const entry = byId.get(id)
              return (
                <li
                  key={id}
                  id={rowId(id)}
                  className={styles.row}
                  draggable
                  aria-current={pickedId === id ? 'true' : undefined}
                  data-dragged={dragged === id ? 'true' : undefined}
                  onClick={() => setPicked(pickedId === id ? NONE : { state: 'picked', id })}
                  onDragStart={() => setDragged(id)}
                  onDragEnd={() => setDragged(undefined)}
                  onDragOver={(event) => event.preventDefault()}
                  onDrop={() => dragged !== undefined && dropOn(spotBefore(sections, dragged, id))}
                >
                  <span className={styles.rank}>{ranks.get(id)}</span>
                  <span className={styles.thumb}>
                    {entry?.cover && (
                      <Image
                        alt=""
                        src={entry.cover}
                        width={64}
                        height={30}
                        unoptimized
                        loading="lazy"
                      />
                    )}
                  </span>
                  <span className={styles.title}>{entry?.title ?? id}</span>
                  <span className={styles.status}>{entry && statusLabel(entry.status)}</span>
                </li>
              )
            })}
            <li
              className={styles.end}
              onDragOver={(event) => event.preventDefault()}
              onDrop={() => dropOn({ division: section.division.key, index: section.ids.length })}
            />
          </ol>
        </section>
      ))}
    </div>
  )
}
