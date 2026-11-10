import { orderedScores } from '../../../../lib/log/present'
import styles from './log.module.css'

export const Scores = ({ scores }: { scores: Record<string, number> }) => {
  const ordered = orderedScores(scores)
  if (ordered.length === 0) return null
  return (
    <span className={styles.scores}>
      {ordered.map(({ axis, score }) => (
        <span key={axis} className={styles.score}>
          <span className={styles.axis}>{axis}</span>
          <span className={styles.number}>{score}</span>
        </span>
      ))}
    </span>
  )
}
