'use client'

import Link from 'next/link'
import { useRouter } from 'next/navigation'
import { useState } from 'react'
import { statusLabel } from '../../../lib/log/present'
import type { IGDBResult, Kind } from '../../../lib/log/types'
import { ENJOYMENT, ENJOYMENT_PRESETS, QueueHeader } from '../_queue/card'
import { IGDBSearch } from '../_queue/IGDBSearch'
import styles from '../_queue/queue.module.css'
import { addByTitle, addFromIGDB, changeEntry } from '../actions'

const UNTRIAGED = 'untriaged'

type Step =
  | { phase: 'search' }
  | { phase: 'adding'; title: string }
  | { phase: 'status'; itemId: number; title: string }
  | { phase: 'score'; itemId: number; title: string; status: string }
  | { phase: 'saved'; itemId: number; title: string; outcome: string }
  | { phase: 'failed'; title: string; error: string }

export const Add = ({ kind }: { kind: Kind }) => {
  const [step, setStep] = useState<Step>({ phase: 'search' })
  const router = useRouter()

  const add = async (title: string, create: () => Promise<number>) => {
    setStep({ phase: 'adding', title })
    try {
      setStep({ phase: 'status', itemId: await create(), title })
    } catch (error) {
      setStep({ phase: 'failed', title, error: String(error) })
    }
  }

  const pick = (result: IGDBResult) => {
    if (result.itemId !== undefined) {
      router.push(`/admin/entries/${result.itemId}`)
      return
    }
    add(result.name, () => addFromIGDB(result.igdbId))
  }

  const save = async (itemId: number, title: string, status: string, enjoyment?: number) => {
    const outcome = enjoyment === undefined ? statusLabel(status) : `${statusLabel(status)}, ${enjoyment}`
    try {
      await changeEntry(itemId, {
        status,
        ...(enjoyment === undefined ? {} : { scores: { [ENJOYMENT]: enjoyment } })
      })
      setStep({ phase: 'saved', itemId, title, outcome })
    } catch (error) {
      setStep({ phase: 'failed', title, error: String(error) })
    }
  }

  return (
    <div className={styles.queue}>
      <QueueHeader current="add" />

      {step.phase === 'search' && (
        <IGDBSearch
          autoFocus
          onPick={pick}
          nothingFits={(query) => (
            <button
              type="button"
              className={styles.quiet}
              onClick={() => add(query, () => addByTitle(kind.name, query))}
            >
              not here? add “{query}” by hand
            </button>
          )}
        />
      )}

      {step.phase === 'adding' && <p className={styles.empty}>adding {step.title}…</p>}

      {step.phase === 'status' && (
        <article className={styles.card}>
          <h2 className={styles.title}>{step.title}</h2>
          <div className={styles.choices}>
            {kind.statuses
              .filter((s) => s !== UNTRIAGED)
              .map((status) => (
                <button
                  key={status}
                  className={styles.choice}
                  onClick={() =>
                    kind.statusesToScore.includes(status)
                      ? setStep({ ...step, phase: 'score', status })
                      : save(step.itemId, step.title, status)
                  }
                >
                  {statusLabel(status)}
                </button>
              ))}
          </div>
          <Link href={`/admin/entries/${step.itemId}`} className={styles.quiet}>
            leave it untriaged, open the editor
          </Link>
        </article>
      )}

      {step.phase === 'score' && (
        <article className={styles.card}>
          <h2 className={styles.title}>{step.title}</h2>
          <div className={styles.choices}>
            <p className={styles.prompt}>{statusLabel(step.status)} — how much did you enjoy it?</p>
            {ENJOYMENT_PRESETS.map((preset) => (
              <button
                key={preset.label}
                className={styles.choice}
                data-preset={preset.label}
                onClick={() => save(step.itemId, step.title, step.status, preset.value)}
              >
                {preset.label}
              </button>
            ))}
            <button className={styles.quiet} onClick={() => save(step.itemId, step.title, step.status)}>
              no score
            </button>
          </div>
        </article>
      )}

      {step.phase === 'saved' && (
        <article className={styles.card}>
          <h2 className={styles.title}>{step.title}</h2>
          <p className={styles.prompt}>added as {step.outcome}</p>
          <div className={styles.choices}>
            <Link href={`/admin/entries/${step.itemId}`} className={styles.choice}>
              open the editor
            </Link>
            <button className={styles.choice} onClick={() => setStep({ phase: 'search' })}>
              add another
            </button>
          </div>
        </article>
      )}

      {step.phase === 'failed' && (
        <article className={styles.card}>
          <p className={styles.retrying}>
            {step.title}: {step.error}
          </p>
          <button className={styles.quiet} onClick={() => setStep({ phase: 'search' })}>
            back to search
          </button>
        </article>
      )}
    </div>
  )
}
