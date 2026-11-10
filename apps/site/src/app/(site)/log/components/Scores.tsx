import { orderedScores } from '../../../../lib/log/present'
import styles from './log.module.css'

const SCORE_SCALE = 100

export const Score = ({ axis, score }: { axis: string; score: number }) => (
  <span className={styles.score}>
    <span className={styles.axis}>{axis}</span>
    <span>
      <span className={styles.number}>{score}</span>
      <span className={styles.scale}> / {SCORE_SCALE}</span>
    </span>
  </span>
)

export const Scores = ({ scores }: { scores: Record<string, number> }) => {
  const ordered = orderedScores(scores)
  if (ordered.length === 0) return null
  return (
    <span className={styles.scores}>
      {ordered.map(({ axis, score }) => (
        <Score key={axis} axis={axis} score={score} />
      ))}
    </span>
  )
}
