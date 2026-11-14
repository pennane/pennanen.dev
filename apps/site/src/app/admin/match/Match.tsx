'use client'

import { useEffect, useState } from 'react'
import { IGDB_MATCH } from '../../../lib/log/syncs'
import type { SyncReport } from '../../../lib/log/types'
import { useSync } from '../_sync/useSync'
import { formatPlaytime, QueueHeader, SaveStatus, type GameCard } from '../_queue/card'
import { IGDBSearch } from '../_queue/IGDBSearch'
import styles from '../_queue/queue.module.css'
import { useSaveQueue } from '../_queue/useSaveQueue'
import { matchIGDB, matchIGDBNone } from '../actions'

export const Match = ({ cards, report }: { cards: GameCard[]; report?: SyncReport }) => {
  const [queue, setQueue] = useState(cards)
  const [open, setOpen] = useState<number>()
  const saves = useSaveQueue()

  useEffect(() => setQueue(cards), [cards])

  const settle = (card: GameCard, label: string, run: () => Promise<unknown>) => {
    setQueue((q) => q.filter((c) => c.id !== card.id))
    setOpen(undefined)
    saves.enqueue(label, run)
  }

  return (
    <div className={styles.queue}>
      <QueueHeader current="match" count={queue.length}>
        <MatchButton initial={report} />
      </QueueHeader>

      {queue.length === 0 ? (
        <p className={styles.empty}>Every game has an IGDB match or a “no match”.</p>
      ) : (
        <ul className={styles.bulkList}>
          {queue.map((card) => (
            <li key={card.id} className={styles.matchRow}>
              <button
                type="button"
                className={styles.matchTitle}
                aria-expanded={open === card.id}
                onClick={() => setOpen(open === card.id ? undefined : card.id)}
              >
                <span>{card.title}</span>
                {card.playtimeMinutes !== undefined && (
                  <span className={styles.count}>{formatPlaytime(card.playtimeMinutes)}</span>
                )}
              </button>
              {open === card.id && (
                <div className={styles.matchPanel}>
                  <IGDBSearch
                    initialQuery={card.title}
                    onPick={(result) =>
                      settle(card, `${card.title} → ${result.name}`, () =>
                        matchIGDB(card.id, result.igdbId)
                      )
                    }
                  />
                  <button
                    type="button"
                    className={styles.quiet}
                    onClick={() =>
                      settle(card, `${card.title} → no match`, () => matchIGDBNone(card.id))
                    }
                  >
                    igdb doesn’t have it: no match
                  </button>
                </div>
              )}
            </li>
          ))}
        </ul>
      )}

      <footer className={styles.footer}>
        <SaveStatus state={saves.state} />
      </footer>
    </div>
  )
}

const MatchButton = ({ initial }: { initial?: SyncReport }) => {
  const { report, start } = useSync(IGDB_MATCH, initial)
  if (!initial) return null
  return (
    <button
      className={styles.quiet}
      disabled={report?.state === 'running'}
      title={report?.problems.at(-1)}
      onClick={start}
    >
      {matchLabel(report)}
    </button>
  )
}

const matchLabel = (report: SyncReport | undefined) => {
  const matched = report?.counts.matched ?? 0
  const checked = report?.counts.checked ?? 0
  switch (report?.state) {
    case 'running':
      return `matching… ${matched} of ${checked}`
    case 'failed':
      return 'match failed, retry'
    case 'succeeded':
      return `last run matched ${matched} of ${checked} · run again`
    default:
      return 'match steam games'
  }
}
