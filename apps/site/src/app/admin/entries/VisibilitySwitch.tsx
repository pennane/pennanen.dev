'use client'

import { CircleDashed, Eye, EyeOff } from 'lucide-react'
import { useState } from 'react'
import type { Presence } from '../../../lib/log/present'
import type { Visibility } from '../../../lib/log/types'
import styles from './entries.module.css'

const NEXT: Record<Visibility, Visibility> = {
  auto: 'public',
  public: 'private',
  private: 'auto'
}

const ICONS: Record<Visibility, typeof Eye> = { auto: CircleDashed, public: Eye, private: EyeOff }

export const VISIBILITY_LABELS: Record<Visibility, string> = {
  auto: 'auto',
  public: 'always',
  private: 'never'
}

const ICON_SIZE = 14

type Save = { state: 'idle' } | { state: 'failed'; wanted: Visibility }

export const VisibilitySwitch = ({
  visibility,
  presence,
  onChange
}: {
  visibility: Visibility
  presence: Presence
  onChange: (visibility: Visibility) => Promise<void>
}) => {
  const [save, setSave] = useState<Save>({ state: 'idle' })

  const choose = async (wanted: Visibility) => {
    setSave({ state: 'idle' })
    try {
      await onChange(wanted)
    } catch {
      setSave({ state: 'failed', wanted })
    }
  }

  const Icon = ICONS[visibility]

  return (
    <div className={styles.switchArea}>
      <button
        type="button"
        className={styles.visibility}
        data-visibility={visibility}
        data-presence={presence}
        aria-label={`${VISIBILITY_LABELS[visibility]}, make ${VISIBILITY_LABELS[NEXT[visibility]]}`}
        onClick={() => choose(NEXT[visibility])}
      >
        <Icon size={ICON_SIZE} aria-hidden />
        {visibility === 'auto' ? `auto · ${presence}` : VISIBILITY_LABELS[visibility]}
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
