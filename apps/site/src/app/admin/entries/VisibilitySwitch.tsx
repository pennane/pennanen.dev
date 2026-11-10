'use client'

import { useOptimistic, useState, useTransition } from 'react'
import type { Visibility } from '../../../lib/log/types'
import { setVisibility } from './actions'
import styles from './entries.module.css'

const VISIBILITIES: Visibility[] = ['private', 'public']

type Save = { state: 'idle' } | { state: 'failed'; wanted: Visibility }

export const VisibilitySwitch = ({
  itemId,
  visibility
}: {
  itemId: number
  visibility: Visibility
}) => {
  const [shown, show] = useOptimistic(visibility)
  const [save, setSave] = useState<Save>({ state: 'idle' })
  const [, startTransition] = useTransition()

  const choose = (wanted: Visibility) =>
    startTransition(async () => {
      show(wanted)
      setSave({ state: 'idle' })
      try {
        await setVisibility(itemId, wanted)
      } catch {
        setSave({ state: 'failed', wanted })
      }
    })

  return (
    <div className={styles.switchArea}>
      <div className={styles.switch} role="group" aria-label="visibility">
        {VISIBILITIES.map((option) => (
          <button
            key={option}
            type="button"
            className={styles.option}
            aria-pressed={option === shown}
            onClick={() => option !== shown && choose(option)}
          >
            {option}
          </button>
        ))}
      </div>
      {save.state === 'failed' && (
        <button type="button" className={styles.retry} onClick={() => choose(save.wanted)}>
          not saved, retry
        </button>
      )}
    </div>
  )
}
