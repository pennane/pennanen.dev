'use client'

import { useRouter } from 'next/navigation'
import { useEffect, useState } from 'react'
import type { IGDBMatchReport } from '../../../lib/log/types'
import { formatPlaytime, QueueHeader, SaveStatus, type GameCard } from '../_queue/card'
import { IGDBSearch } from '../_queue/IGDBSearch'
import styles from '../_queue/queue.module.css'
import { useSaveQueue } from '../_queue/useSaveQueue'
import { igdbMatchStatus, matchIGDB, matchIGDBNone, startIGDBMatch } from '../actions'

const MATCH_POLL_MS = 3000

export const Match = ({ cards, report }: { cards: GameCard[]; report?: IGDBMatchReport }) => {
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
                      settle(card, `${card.title} → ${result.name}`, () => matchIGDB(card.id, result.igdbId))
                    }
                  />
                  <button
                    type="button"
                    className={styles.quiet}
                    onClick={() => settle(card, `${card.title} → no match`, () => matchIGDBNone(card.id))}
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

const MatchButton = ({ initial }: { initial?: IGDBMatchReport }) => {
  const [report, setReport] = useState(initial)
  const router = useRouter()

  useEffect(() => {
    if (report?.state !== 'matching') return
    const timer = setInterval(async () => {
      const next = await igdbMatchStatus()
      setReport(next)
      if (next.state !== 'matching') router.refresh()
    }, MATCH_POLL_MS)
    return () => clearInterval(timer)
  }, [report?.state, router])

  if (!initial) return null
  const matching = report?.state === 'matching'
  return (
    <button
      className={styles.quiet}
      disabled={matching}
      title={report?.problems.at(-1)}
      onClick={async () => setReport(await startIGDBMatch())}
    >
      {matching
        ? `matching… ${report.matched} of ${report.checked}`
        : report?.state === 'failed'
          ? 'match failed, retry'
          : 'match steam games'}
    </button>
  )
}
