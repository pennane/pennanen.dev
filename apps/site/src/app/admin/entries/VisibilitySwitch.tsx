'use client'

import { Eye, EyeOff } from 'lucide-react'
import { useState } from 'react'
import type { Visibility } from '../../../lib/log/types'
import styles from './entries.module.css'

const OTHER: Record<Visibility, Visibility> = {
  private: 'public',
  public: 'private'
}

const ICONS = { private: EyeOff, public: Eye }

const ICON_SIZE = 14

type Save = { state: 'idle' } | { state: 'failed'; wanted: Visibility }

export const VisibilitySwitch = ({
  visibility,
  onChange
}: {
  visibility: Visibility
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
        aria-label={`${visibility}, make ${OTHER[visibility]}`}
        onClick={() => choose(OTHER[visibility])}
      >
        <Icon size={ICON_SIZE} aria-hidden />
        {visibility}
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
