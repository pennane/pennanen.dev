'use client'

import { shownScores } from '../../../../lib/log/present'
import styles from './log.module.css'
import { usePresentation } from './PresentationContext'

export const ScoreParts = ({ kind, scores }: { kind: string; scores: Record<string, string> }) => {
  const shown = shownScores(usePresentation(), kind, scores)
  return (
    <>
      {shown.map(({ axis, axisName, division }) => (
        <span key={axis} className={styles.score}>
          <span className={styles.axis}>{axisName}</span>
          <span className={styles.word}>{division}</span>
        </span>
      ))}
    </>
  )
}

export const Scores = ({ kind, scores }: { kind: string; scores: Record<string, string> }) =>
  Object.keys(scores).length === 0 ? null : (
    <span className={styles.scores}>
      <ScoreParts kind={kind} scores={scores} />
    </span>
  )
