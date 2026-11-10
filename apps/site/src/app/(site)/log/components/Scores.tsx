'use client'

import { orderedScores, scoreWordFor } from '../../../../lib/log/present'
import styles from './log.module.css'
import { usePresentation } from './PresentationContext'

const SCORE_SCALE = 100

export const Score = ({ kind, axis, score }: { kind: string; axis: string; score: number }) => {
  const settings = usePresentation().kinds[kind]
  const label = settings?.axisLabels[axis] ?? axis
  const display = settings?.scoreDisplay ?? 'number'
  const word = settings && scoreWordFor(score, settings.scoreWords)
  const number = (
    <span>
      <span className={styles.number}>{score}</span>
      <span className={styles.scale}> / {SCORE_SCALE}</span>
    </span>
  )
  return (
    <span className={styles.score}>
      <span className={styles.axis}>{label}</span>
      {display !== 'number' && word && <span className={styles.word}>{word}</span>}
      {(display !== 'word' || !word) && number}
    </span>
  )
}

export const Scores = ({ kind, scores }: { kind: string; scores: Record<string, number> }) => {
  const ordered = orderedScores(scores)
  if (ordered.length === 0) return null
  return (
    <span className={styles.scores}>
      {ordered.map(({ axis, score }) => (
        <Score key={axis} kind={kind} axis={axis} score={score} />
      ))}
    </span>
  )
}
