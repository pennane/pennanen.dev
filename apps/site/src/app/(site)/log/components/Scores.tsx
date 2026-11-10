'use client'

import {
  divisionName,
  HEADLINE_AXIS,
  isPublicKind,
  kindPlural,
  orderedScores,
  type Rank
} from '../../../../lib/log/present'
import styles from './log.module.css'
import { usePresentation } from './PresentationContext'

export type RankShown = 'place' | 'place-of-all'

const RankLabel = ({ kind, rank, shown }: { kind: string; rank: Rank; shown: RankShown }) => (
  <span>
    <span className={styles.number}>#{rank.place}</span>
    {shown === 'place-of-all' && (
      <span className={styles.scale}>
        {' '}
        of {rank.of}
        {isPublicKind(kind) && ` ${kindPlural(kind)}`}
      </span>
    )}
  </span>
)

export const Score = ({
  kind,
  axis,
  division,
  rank,
  rankShown
}: {
  kind: string
  axis: string
  division: string
  rank: Rank | undefined
  rankShown: RankShown
}) => {
  const settings = usePresentation().kinds[kind]
  return (
    <span className={styles.score}>
      <span className={styles.axis}>{settings?.axisLabels[axis] ?? axis}</span>
      <span className={styles.word}>{divisionName(settings?.divisions, axis, division)}</span>
      {axis === HEADLINE_AXIS && rank && <RankLabel kind={kind} rank={rank} shown={rankShown} />}
    </span>
  )
}

export const Scores = ({
  kind,
  scores,
  rank,
  rankShown
}: {
  kind: string
  scores: Record<string, string>
  rank: Rank | undefined
  rankShown: RankShown
}) => {
  const ordered = orderedScores(scores)
  if (ordered.length === 0) return null
  return (
    <span className={styles.scores}>
      {ordered.map(({ axis, division }) => (
        <Score
          key={axis}
          kind={kind}
          axis={axis}
          division={division}
          rank={rank}
          rankShown={rankShown}
        />
      ))}
    </span>
  )
}
