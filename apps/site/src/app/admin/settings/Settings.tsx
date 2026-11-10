'use client'

import { useState } from 'react'
import { scoreWordFor, statusLabel } from '../../../lib/log/present'
import type { KindSettings, ScoreDisplay, ScoreWord } from '../../../lib/log/types'
import { saveKindSettings, saveTagTitles } from '../actions'
import editor from '../entries/editor.module.css'
import styles from './settings.module.css'

const SCORE_DISPLAYS: { value: ScoreDisplay; label: string }[] = [
  { value: 'number', label: 'number' },
  { value: 'word', label: 'word' },
  { value: 'both', label: 'both' }
]

const SAMPLE_SCORES = [20, 55, 75, 92]

type Save =
  | { state: 'unchanged' }
  | { state: 'changed' }
  | { state: 'saving' }
  | { state: 'saved' }
  | { state: 'failed'; error: string }

const useSave = <T,>(store: (value: T) => Promise<T>) => {
  const [save, setSave] = useState<Save>({ state: 'unchanged' })
  const run = async (value: T, settle: (saved: T) => void) => {
    setSave({ state: 'saving' })
    try {
      settle(await store(value))
      setSave({ state: 'saved' })
    } catch (error) {
      setSave({ state: 'failed', error: String(error) })
    }
  }
  return { save, changed: () => setSave({ state: 'changed' }), run }
}

const SaveButton = ({ save, onSave }: { save: Save; onSave: () => void }) => (
  <div className={styles.saveRow}>
    <button
      type="button"
      className={styles.save}
      disabled={save.state === 'unchanged' || save.state === 'saving' || save.state === 'saved'}
      onClick={onSave}
    >
      {save.state === 'saving' ? 'saving…' : save.state === 'saved' ? 'saved' : 'save'}
    </button>
    {save.state === 'failed' && <span className={styles.error}>{save.error}</span>}
  </div>
)

const Section = ({ label, hint, children }: { label: string; hint?: string; children: React.ReactNode }) => (
  <section className={editor.section}>
    <h3 className={editor.label}>{label}</h3>
    {hint && <p className={styles.hint}>{hint}</p>}
    {children}
  </section>
)

export const KindSettingsForm = ({
  kind,
  statuses,
  axes,
  initial
}: {
  kind: string
  statuses: string[]
  axes: string[]
  initial: KindSettings
}) => {
  const [settings, setSettings] = useState(initial)
  const { save, changed, run } = useSave((value: KindSettings) => saveKindSettings(kind, value))

  const change = (next: Partial<KindSettings>) => {
    setSettings((s) => ({ ...s, ...next }))
    changed()
  }

  const toggleStatus = (status: string) =>
    change({
      statusesShown: settings.statusesShown.includes(status)
        ? settings.statusesShown.filter((s) => s !== status)
        : [...settings.statusesShown, status]
    })

  const changeWord = (index: number, word: Partial<ScoreWord>) =>
    change({ scoreWords: settings.scoreWords.map((w, i) => (i === index ? { ...w, ...word } : w)) })

  return (
    <article className={styles.kind}>
      <h2 className={styles.kindName}>{kind}</h2>

      <Section label="shown on public pages" hint="entries set to auto show when their status is picked here">
        <div className={editor.chips}>
          {statuses.map((status) => (
            <button
              key={status}
              type="button"
              className={editor.chip}
              aria-pressed={settings.statusesShown.includes(status)}
              onClick={() => toggleStatus(status)}
            >
              {statusLabel(status)}
            </button>
          ))}
        </div>
      </Section>

      <Section label="score names" hint="what each score is called on public pages">
        {axes.map((axis) => (
          <label key={axis} className={styles.row}>
            <span className={styles.rowKey}>{axis}</span>
            <input
              className={styles.input}
              value={settings.axisLabels[axis] ?? ''}
              placeholder={axis}
              onChange={(event) => change({ axisLabels: { ...settings.axisLabels, [axis]: event.target.value } })}
            />
          </label>
        ))}
      </Section>

      <Section label="scores show as">
        <div className={editor.chips}>
          {SCORE_DISPLAYS.map((d) => (
            <button
              key={d.value}
              type="button"
              className={editor.chip}
              aria-pressed={settings.scoreDisplay === d.value}
              onClick={() => change({ scoreDisplay: d.value })}
            >
              {d.label}
            </button>
          ))}
        </div>
      </Section>

      <Section label="score words" hint="a score gets the word of the highest start at or below it">
        {settings.scoreWords.map((w, index) => (
          <div key={index} className={styles.row}>
            <input
              className={styles.from}
              type="number"
              min={0}
              max={100}
              value={w.from}
              aria-label="from"
              onChange={(event) => changeWord(index, { from: Number(event.target.value) })}
            />
            <input
              className={styles.input}
              value={w.word}
              placeholder="word"
              onChange={(event) => changeWord(index, { word: event.target.value })}
            />
            <button
              type="button"
              className={editor.quiet}
              onClick={() => change({ scoreWords: settings.scoreWords.filter((_, i) => i !== index) })}
            >
              remove
            </button>
          </div>
        ))}
        <button
          type="button"
          className={editor.quiet}
          onClick={() => change({ scoreWords: [...settings.scoreWords, { from: 0, word: '' }] })}
        >
          add a word
        </button>
        <p className={styles.hint}>
          {SAMPLE_SCORES.map((score) => `${score} → ${scoreWordFor(score, sortedWords(settings.scoreWords)) ?? '–'}`).join(' · ')}
        </p>
      </Section>

      <SaveButton save={save} onSave={() => run(settings, setSettings)} />
    </article>
  )
}

const sortedWords = (words: ScoreWord[]) => words.toSorted((a, b) => a.from - b.from)

export const TagTitlesForm = ({ tags, initial }: { tags: string[]; initial: Record<string, string> }) => {
  const [titles, setTitles] = useState(initial)
  const { save, changed, run } = useSave(saveTagTitles)

  return (
    <article className={styles.kind}>
      <h2 className={styles.kindName}>tags</h2>
      <Section label="titles" hint="the id stays in the data, the title is what public pages show">
        {tags.map((tag) => (
          <label key={tag} className={styles.row}>
            <span className={styles.rowKey}>{tag}</span>
            <input
              className={styles.input}
              value={titles[tag] ?? ''}
              placeholder={tag}
              onChange={(event) => {
                setTitles((t) => ({ ...t, [tag]: event.target.value }))
                changed()
              }}
            />
          </label>
        ))}
      </Section>
      <SaveButton save={save} onSave={() => run(titles, setTitles)} />
    </article>
  )
}
