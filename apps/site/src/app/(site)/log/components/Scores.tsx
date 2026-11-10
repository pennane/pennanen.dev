'use client'

import { divisionName, orderedScores } from '../../../../lib/log/present'
import styles from './log.module.css'
import { usePresentation } from './PresentationContext'

export const Score = ({ kind, axis, division }: { kind: string; axis: string; division: string }) => {
  const settings = usePresentation().kinds[kind]
  return (
    <span className={styles.score}>
      <span className={styles.axis}>{settings?.axisLabels[axis] ?? axis}</span>
      <span className={styles.word}>{divisionName(settings?.divisions, axis, division)}</span>
    </span>
  )
}

export const Scores = ({ kind, scores }: { kind: string; scores: Record<string, string> }) => {
  const ordered = orderedScores(scores)
  if (ordered.length === 0) return null
  return (
    <span className={styles.scores}>
      {ordered.map(({ axis, division }) => (
        <Score key={axis} kind={kind} axis={axis} division={division} />
      ))}
    </span>
  )
}
