'use client'

import { useCallback, useEffect, useRef, useState } from 'react'

type Save = { key: number; label: string; run: () => Promise<unknown> }

export type SaveQueueState =
  | { kind: 'saved' }
  | { kind: 'saving'; unsaved: number }
  | { kind: 'retrying'; unsaved: number; error: string; label: string }

const MAX_RETRY_DELAY_MS = 30_000

export const useSaveQueue = () => {
  const [saves, setSaves] = useState<Save[]>([])
  const [failure, setFailure] = useState<{ error: string; attempts: number }>()
  const running = useRef(false)
  const nextKey = useRef(0)

  const enqueue = useCallback((label: string, run: () => Promise<unknown>) => {
    setSaves((current) => [...current, { key: nextKey.current++, label, run }])
  }, [])

  useEffect(() => {
    const save = saves[0]
    if (!save || running.current) return
    const attempts = failure?.attempts ?? 0
    const delay = attempts === 0 ? 0 : Math.min(MAX_RETRY_DELAY_MS, 1000 * 2 ** attempts)
    const timer = setTimeout(async () => {
      running.current = true
      try {
        await save.run()
        setFailure(undefined)
        setSaves((current) => current.filter((s) => s.key !== save.key))
      } catch (error) {
        setFailure({
          error: error instanceof Error ? error.message : String(error),
          attempts: attempts + 1
        })
      } finally {
        running.current = false
      }
    }, delay)
    return () => clearTimeout(timer)
  }, [saves, failure])

  useEffect(() => {
    if (saves.length === 0) return
    const warnBeforeLeaving = (event: BeforeUnloadEvent) => event.preventDefault()
    window.addEventListener('beforeunload', warnBeforeLeaving)
    return () => window.removeEventListener('beforeunload', warnBeforeLeaving)
  }, [saves.length])

  const state: SaveQueueState =
    saves.length === 0
      ? { kind: 'saved' }
      : failure
        ? { kind: 'retrying', unsaved: saves.length, error: failure.error, label: saves[0].label }
        : { kind: 'saving', unsaved: saves.length }

  return { enqueue, state }
}
