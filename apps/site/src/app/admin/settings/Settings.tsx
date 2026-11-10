'use client'

import { useRouter } from 'next/navigation'
import { useState } from 'react'
import { PRESTIGE, statusLabel } from '../../../lib/log/present'
import type { Axis, Division, KindSettings } from '../../../lib/log/types'
import { saveAxes, saveKindSettings, saveTagTitles } from '../actions'
import editor from '../entries/editor.module.css'
import styles from './settings.module.css'

const UNSAVED_KEY = ''

const TIERS = ['S', 'A', 'B', 'C', 'D', 'F']

const moved = <T,>(list: T[], from: number, to: number) => {
  if (to < 0 || to >= list.length) return list
  const next = [...list]
  const [item] = next.splice(from, 1)
  next.splice(to, 0, item)
  return next
}

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

const Section = ({
  label,
  hint,
  children
}: {
  label: string
  hint?: string
  children: React.ReactNode
}) => (
  <section className={editor.section}>
    <h3 className={editor.label}>{label}</h3>
    {hint && <p className={styles.hint}>{hint}</p>}
    {children}
  </section>
)

export const AxesForm = ({ initial }: { initial: Axis[] }) => {
  const router = useRouter()
  const [axes, setAxes] = useState(initial)
  const { save, changed, run } = useSave(saveAxes)

  const change = (next: Axis[]) => {
    setAxes(next)
    changed()
  }
  const changeAxis = (index: number, next: Partial<Axis>) =>
    change(axes.map((axis, i) => (i === index ? { ...axis, ...next } : axis)))

  return (
    <article className={styles.kind}>
      <h2 className={styles.kindName}>axes</h2>
      <p className={styles.hint}>
        shared by every kind that uses them; {PRESTIGE} orders things and stays
      </p>
      {axes.map((axis, index) => (
        <Section key={axis.key || `new-${index}`} label={axis.name || 'new axis'}>
          <div className={styles.row}>
            <input
              className={styles.input}
              value={axis.name}
              placeholder="name"
              onChange={(event) => changeAxis(index, { name: event.target.value })}
            />
            {axis.key !== UNSAVED_KEY && axis.key !== axis.name && (
              <span className={styles.rowKey}>{axis.key}</span>
            )}
            <button
              type="button"
              className={editor.quiet}
              aria-label="move up"
              onClick={() => change(moved(axes, index, index - 1))}
            >
              ↑
            </button>
            <button
              type="button"
              className={editor.quiet}
              aria-label="move down"
              onClick={() => change(moved(axes, index, index + 1))}
            >
              ↓
            </button>
            {axis.key !== PRESTIGE && (
              <button
                type="button"
                className={editor.quiet}
                onClick={() => change(axes.filter((_, i) => i !== index))}
              >
                remove
              </button>
            )}
          </div>
          <p className={styles.hint}>
            {axis.key === PRESTIGE
              ? 'divisions, best first; entries rank inside each'
              : 'divisions, best first'}
          </p>
          <DivisionList
            divisions={axis.divisions}
            onChange={(divisions) => changeAxis(index, { divisions })}
          />
        </Section>
      ))}
      <button
        type="button"
        className={editor.quiet}
        onClick={() =>
          change([
            ...axes,
            {
              key: UNSAVED_KEY,
              name: '',
              divisions: TIERS.map((name) => ({ key: UNSAVED_KEY, name }))
            }
          ])
        }
      >
        add an axis
      </button>
      <SaveButton
        save={save}
        onSave={() =>
          run(axes, (saved) => {
            setAxes(saved)
            router.refresh()
          })
        }
      />
    </article>
  )
}

export const KindSettingsForm = ({
  kind,
  statuses,
  axes,
  rankedAxis,
  initial
}: {
  kind: string
  statuses: string[]
  axes: Axis[]
  rankedAxis: string
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

  const nameOf = (key: string) => axes.find((axis) => axis.key === key)?.name ?? key
  const unused = axes.filter((axis) => !settings.axes.includes(axis.key))

  return (
    <article className={styles.kind}>
      <h2 className={styles.kindName}>{kind}</h2>

      <Section label="title" hint="the heading on its public page">
        <input
          className={styles.input}
          value={settings.title}
          aria-invalid={settings.title.trim() === ''}
          onChange={(event) => change({ title: event.target.value })}
        />
      </Section>

      <Section label="description" hint="markdown, shown under the title">
        <textarea
          className={editor.text}
          rows={4}
          value={settings.description}
          onChange={(event) => change({ description: event.target.value })}
        />
      </Section>

      <Section
        label="shown on public pages"
        hint="entries set to auto show when their status is picked here"
      >
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

      <Section label="axes" hint="the first one leads on list rows">
        {settings.axes.map((key, index) => (
          <div key={key} className={styles.row}>
            <span className={styles.input}>{nameOf(key)}</span>
            <button
              type="button"
              className={editor.quiet}
              aria-label="move up"
              onClick={() => change({ axes: moved(settings.axes, index, index - 1) })}
            >
              ↑
            </button>
            <button
              type="button"
              className={editor.quiet}
              aria-label="move down"
              onClick={() => change({ axes: moved(settings.axes, index, index + 1) })}
            >
              ↓
            </button>
            {key !== rankedAxis && (
              <button
                type="button"
                className={editor.quiet}
                onClick={() => change({ axes: settings.axes.filter((a) => a !== key) })}
              >
                remove
              </button>
            )}
          </div>
        ))}
        {unused.length > 0 && (
          <div className={editor.chips}>
            {unused.map((axis) => (
              <button
                key={axis.key}
                type="button"
                className={editor.chip}
                onClick={() => change({ axes: [...settings.axes, axis.key] })}
              >
                + {axis.name}
              </button>
            ))}
          </div>
        )}
      </Section>

      <SaveButton save={save} onSave={() => run(settings, setSettings)} />
    </article>
  )
}

const DivisionList = ({
  divisions,
  onChange
}: {
  divisions: Division[]
  onChange: (divisions: Division[]) => void
}) => {
  const rename = (index: number, name: string) =>
    onChange(divisions.map((d, i) => (i === index ? { ...d, name } : d)))
  return (
    <>
      {divisions.map((d, index) => (
        <div key={d.key || `new-${index}`} className={styles.row}>
          <input
            className={styles.input}
            value={d.name}
            placeholder="name"
            onChange={(event) => rename(index, event.target.value)}
          />
          {d.key !== UNSAVED_KEY && d.key !== d.name && (
            <span className={styles.rowKey}>{d.key}</span>
          )}
          <button
            type="button"
            className={editor.quiet}
            aria-label="move up"
            onClick={() => onChange(moved(divisions, index, index - 1))}
          >
            ↑
          </button>
          <button
            type="button"
            className={editor.quiet}
            aria-label="move down"
            onClick={() => onChange(moved(divisions, index, index + 1))}
          >
            ↓
          </button>
          <button
            type="button"
            className={editor.quiet}
            onClick={() => onChange(divisions.filter((_, i) => i !== index))}
          >
            remove
          </button>
        </div>
      ))}
      <button
        type="button"
        className={editor.quiet}
        onClick={() => onChange([...divisions, { key: UNSAVED_KEY, name: '' }])}
      >
        add a division
      </button>
    </>
  )
}

export const TagTitlesForm = ({
  tags,
  initial
}: {
  tags: string[]
  initial: Record<string, string>
}) => {
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
