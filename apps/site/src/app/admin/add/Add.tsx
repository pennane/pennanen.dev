'use client'

import Link from 'next/link'
import { useRouter } from 'next/navigation'
import { useState } from 'react'
import { PRESTIGE, statusLabel } from '../../../lib/log/present'
import type { IGDBResult, Kind } from '../../../lib/log/types'
import { prestigeDivisions, prestigeName, QueueHeader, useKeys } from '../_queue/card'
import { IGDBSearch } from '../_queue/IGDBSearch'
import styles from '../_queue/queue.module.css'
import { addByTitle, addFromIGDB, changeEntry, placeEntry } from '../actions'
import { Compare, compareKeys, useRanked, type Comparing } from '../_queue/Compare'
import type { Placement, Ranked, Search } from '../_queue/placing'

const UNTRIAGED = 'untriaged'
const NO_KEYS = {}

type Step =
  | { phase: 'search' }
  | { phase: 'adding'; title: string }
  | { phase: 'status'; itemId: number; title: string }
  | { phase: 'score'; itemId: number; title: string; status: string }
  | { phase: 'compare'; itemId: number; title: string; status: string; search: Search }
  | { phase: 'saved'; itemId: number; title: string; outcome: string }
  | { phase: 'failed'; title: string; error: string }

export const Add = ({ kind, ranked: initialRanked }: { kind: Kind; ranked: Ranked }) => {
  const presets = prestigeDivisions(kind)
  const { ranked, place, searchFrom } = useRanked(initialRanked)
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

  const save = async (itemId: number, title: string, status: string, placement?: Placement) => {
    const outcome = placement
      ? `${statusLabel(status)}, ${prestigeName(kind, placement.division)} #${placement.index + 1}`
      : statusLabel(status)
    try {
      await changeEntry(itemId, {
        status,
        ...(placement ? { scores: { [PRESTIGE]: placement.division } } : {})
      })
      if (placement) {
        await placeEntry(itemId, placement.division, placement.index)
        place({ id: itemId, title }, placement)
      }
      setStep({ phase: 'saved', itemId, title, outcome })
    } catch (error) {
      setStep({ phase: 'failed', title, error: String(error) })
    }
  }

  const score = (itemId: number, title: string, status: string, division: string) => {
    const search = searchFrom(division, itemId)
    if (search.high === 0) save(itemId, title, status, { division, index: 0 })
    else setStep({ phase: 'compare', itemId, title, status, search })
  }

  const comparing: Comparing | undefined =
    step.phase === 'compare'
      ? {
          ranked,
          search: step.search,
          id: step.itemId,
          divisionName: prestigeName(kind, step.search.division),
          onSearch: (search) => setStep({ ...step, search }),
          onPlace: (placement) => save(step.itemId, step.title, step.status, placement),
          onBack: () =>
            setStep({ phase: 'score', itemId: step.itemId, title: step.title, status: step.status })
        }
      : undefined

  useKeys(comparing ? compareKeys(comparing) : NO_KEYS)

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
            <p className={styles.prompt}>{statusLabel(step.status)} — which tier?</p>
            {presets.map((preset) => (
              <button
                key={preset.name}
                className={styles.choice}
                data-preset={preset.name}
                onClick={() => score(step.itemId, step.title, step.status, preset.key)}
              >
                {preset.name}
              </button>
            ))}
            <button className={styles.quiet} onClick={() => save(step.itemId, step.title, step.status)}>
              no score
            </button>
          </div>
        </article>
      )}

      {comparing && step.phase === 'compare' && (
        <article className={styles.card}>
          <h2 className={styles.title}>{step.title}</h2>
          <Compare {...comparing} />
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
