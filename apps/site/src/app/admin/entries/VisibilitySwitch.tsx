'use client'

import { Eye, EyeOff } from 'lucide-react'
import { useOptimistic, useState, useTransition } from 'react'
import type { Visibility } from '../../../lib/log/types'
import { setVisibility } from './actions'
import styles from './entries.module.css'

const OTHER: Record<Visibility, Visibility> = {
  private: 'public',
  public: 'private'
}

const ICONS = { private: EyeOff, public: Eye }

const ICON_SIZE = 14

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

  const Icon = ICONS[shown]

  return (
    <div className={styles.switchArea}>
      <button
        type="button"
        className={styles.visibility}
        data-visibility={shown}
        aria-label={`${shown}, make ${OTHER[shown]}`}
        onClick={() => choose(OTHER[shown])}
      >
        <Icon size={ICON_SIZE} aria-hidden />
        {shown}
      </button>
      {save.state === 'failed' && (
        <button
          type="button"
          className={styles.retry}
          onClick={() => choose(save.wanted)}
        >
          not saved, retry
        </button>
      )}
    </div>
  )
}
