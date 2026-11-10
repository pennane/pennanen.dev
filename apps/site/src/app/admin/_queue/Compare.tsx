'use client'

import Image from 'next/image'
import { useCallback, useState } from 'react'
import styles from './queue.module.css'
import {
  answered,
  isSettled,
  placementOf,
  questionsLeft,
  rivalOf,
  type Answer,
  type Placement,
  type Ranked,
  type Rival,
  type Search,
  startSearch,
  withoutRival,
  withPlaced
} from './placing'

export const useRanked = (initial: Ranked) => {
  const [ranked, setRanked] = useState(initial)
  const place = useCallback(
    (rival: Rival, placement: Placement) => setRanked((r) => withPlaced(r, rival, placement)),
    []
  )
  const unplace = useCallback((id: number) => setRanked((r) => withoutRival(r, id)), [])
  const searchFrom = useCallback(
    (division: string, id: number) => startSearch(ranked, division, id),
    [ranked]
  )
  return { ranked, place, unplace, searchFrom }
}

export type Comparing = {
  ranked: Ranked
  search: Search
  id: number
  divisionName: string
  onSearch: (search: Search) => void
  onPlace: (placement: Placement) => void
  onBack: () => void
}

export const answer = (comparing: Comparing, given: Answer) => {
  const next = answered(comparing.search, given)
  if (isSettled(next)) comparing.onPlace(placementOf(next))
  else comparing.onSearch(next)
}

export const compareKeys = (comparing: Comparing): Record<string, () => void> => ({
  ArrowUp: () => answer(comparing, 'better'),
  ArrowDown: () => answer(comparing, 'worse'),
  Enter: () => comparing.onPlace(placementOf(comparing.search)),
  Escape: comparing.onBack
})

export const Compare = (comparing: Comparing) => {
  const rival = rivalOf(comparing.ranked, comparing.search, comparing.id)
  if (!rival) return null
  const left = questionsLeft(comparing.search)
  return (
    <div className={styles.choices}>
      <p className={styles.prompt}>
        {comparing.divisionName}: better or worse than
      </p>
      <div className={styles.rival}>
        {rival.cover && <Image src={rival.cover} alt="" width={92} height={43} unoptimized />}
        <span>{rival.title}</span>
      </div>
      <button className={styles.choice} onClick={() => answer(comparing, 'better')}>
        <kbd>↑</kbd> better
      </button>
      <button className={styles.choice} onClick={() => answer(comparing, 'worse')}>
        <kbd>↓</kbd> worse
      </button>
      <button
        className={styles.quiet}
        onClick={() => comparing.onPlace(placementOf(comparing.search))}
      >
        <kbd>enter</kbd> place it here ({left} {left === 1 ? 'question' : 'questions'} left)
      </button>
      <button className={styles.quiet} onClick={comparing.onBack}>
        <kbd>esc</kbd> back
      </button>
    </div>
  )
}
