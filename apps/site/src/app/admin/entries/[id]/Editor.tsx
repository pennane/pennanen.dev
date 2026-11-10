'use client'

import Image from 'next/image'
import Link from 'next/link'
import { useState } from 'react'
import {
  presenceOf,
  statusLabel,
  tagLabel,
  TIME_SPENT_CHOICES,
  TIME_SPENT_LABELS,
  type TimeSpent
} from '../../../../lib/log/present'
import type {
  Entry,
  EntryChange,
  Kind,
  Moment,
  MomentVisibility,
  Visibility
} from '../../../../lib/log/types'
import { SaveStatus } from '../../_queue/card'
import { useSaveQueue } from '../../_queue/useSaveQueue'
import { addMoment, changeEntry, deleteMoment } from '../../actions'
import styles from './editor.module.css'

const VISIBILITY_CHOICES: { value: Visibility; label: string }[] = [
  { value: 'auto', label: 'auto, by status' },
  { value: 'public', label: 'always shown' },
  { value: 'private', label: 'never shown' }
]
const MOMENT_VISIBILITIES: { value: MomentVisibility; label: string }[] = [
  { value: 'inherit', label: 'same as entry' },
  { value: 'private', label: 'private' }
]
const UNTRIAGED = 'untriaged'
const UNSAID = ''
const NO_SCORE = '–'

const today = () => new Date().toISOString().slice(0, 10)

export const Editor = ({
  itemId,
  title,
  cover,
  publicPath,
  kind,
  entry,
  moments: initialMoments
}: {
  itemId: number
  title: string
  cover: string | undefined
  publicPath: string
  kind: Kind
  entry: Entry
  moments: Moment[]
}) => {
  const [status, setStatus] = useState(entry.status)
  const [visibility, setVisibility] = useState(entry.visibility)
  const [tags, setTags] = useState(entry.tags)
  const [timeSpent, setTimeSpent] = useState<TimeSpent | typeof UNSAID>(entry.timeSpent ?? UNSAID)
  const [moments, setMoments] = useState(initialMoments)
  const saves = useSaveQueue()

  const save = (label: string, change: EntryChange) =>
    saves.enqueue(label, () => changeEntry(itemId, change))

  const presence = presenceOf({ status, visibility }, kind.statusesShown)

  return (
    <div className={styles.editor}>
      <header className={styles.header}>
        <Link href="/admin/entries" className={styles.quiet}>
          ← entries
        </Link>
        <SaveStatus state={saves.state} />
      </header>

      {cover && (
        <div className={styles.cover}>
          <Image src={cover} alt="" width={460} height={215} unoptimized priority />
        </div>
      )}
      <h1 className={styles.title}>{title}</h1>
      <p className={styles.where}>
        {presence === 'shown' ? (
          <a href={publicPath} target="_blank" rel="noreferrer">
            shown at {publicPath}
          </a>
        ) : (
          'not on public pages'
        )}
      </p>

      <Section label="status">
        <Chips
          options={kind.statuses
            .filter((s) => s !== UNTRIAGED)
            .map((s) => ({ value: s, label: statusLabel(s) }))}
          chosen={status}
          onChoose={(next) => {
            setStatus(next)
            save(`status → ${next}`, { status: next })
          }}
        />
      </Section>

      <Section label="visibility">
        <Chips
          options={VISIBILITY_CHOICES}
          chosen={visibility}
          onChoose={(next) => {
            setVisibility(next)
            save(`visibility → ${next}`, { visibility: next })
          }}
        />
      </Section>

      {kind.axes.map((axis) => (
        <ScoreSlider
          key={axis}
          axis={axis}
          initial={entry.scores[axis]}
          onSave={(value) =>
            save(`${axis} → ${value ?? 'none'}`, { scores: { [axis]: value ?? null } })
          }
        />
      ))}

      <Section label="time spent">
        <Chips
          options={[
            ...TIME_SPENT_CHOICES.map((t) => ({ value: t, label: TIME_SPENT_LABELS[t] })),
            { value: UNSAID, label: 'not saying' }
          ]}
          chosen={timeSpent}
          onChoose={(next) => {
            setTimeSpent(next)
            save(`time → ${next || 'unsaid'}`, { timeSpent: next })
          }}
        />
      </Section>

      <TextField
        label="blurb"
        hint="public, a line or two for others"
        initial={entry.blurb}
        rows={3}
        onSave={(blurb) => save('blurb', { blurb })}
      />

      <TextField
        label="notes"
        hint="private, only you see these"
        initial={entry.notes}
        rows={6}
        onSave={(notes) => save('notes', { notes })}
      />

      <Section label="tags">
        <Tags
          tags={tags}
          onAdd={(tag) => {
            setTags((t) => (t.includes(tag) ? t : [...t, tag]))
            save(`tag ${tag}`, { addTags: [tag] })
          }}
          onRemove={(tag) => {
            setTags((t) => t.filter((x) => x !== tag))
            save(`untag ${tag}`, { removeTags: [tag] })
          }}
        />
      </Section>

      <Section label="moments">
        <Moments
          moments={moments}
          onAdd={(moment) =>
            saves.enqueue(`moment ${moment.at}`, async () => setMoments(await addMoment(itemId, moment)))
          }
          onDelete={(id) => {
            setMoments((m) => m.filter((x) => x.id !== id))
            saves.enqueue('delete moment', () => deleteMoment(id))
          }}
        />
      </Section>
    </div>
  )
}

const Section = ({ label, children }: { label: string; children: React.ReactNode }) => (
  <section className={styles.section}>
    <h2 className={styles.label}>{label}</h2>
    {children}
  </section>
)

const Chips = <T extends string>({
  options,
  chosen,
  onChoose
}: {
  options: { value: T; label: string }[]
  chosen: T
  onChoose: (value: T) => void
}) => (
  <div className={styles.chips}>
    {options.map((o) => (
      <button
        key={o.value}
        type="button"
        className={styles.chip}
        aria-pressed={o.value === chosen}
        onClick={() => o.value !== chosen && onChoose(o.value)}
      >
        {o.label}
      </button>
    ))}
  </div>
)

const ScoreSlider = ({
  axis,
  initial,
  onSave
}: {
  axis: string
  initial: number | undefined
  onSave: (value: number | undefined) => void
}) => {
  const [value, setValue] = useState(initial)
  const [saved, setSaved] = useState(initial)

  const commit = (next: number | undefined) => {
    if (next === saved) return
    setSaved(next)
    onSave(next)
  }

  return (
    <Section label={axis}>
      <div className={styles.score}>
        <input
          type="range"
          min={0}
          max={100}
          value={value ?? 50}
          data-empty={value === undefined}
          aria-label={axis}
          onChange={(event) => setValue(Number(event.target.value))}
          onPointerUp={() => commit(value)}
          onKeyUp={() => commit(value)}
          onBlur={() => commit(value)}
        />
        <span className={styles.scoreValue}>{value ?? NO_SCORE}</span>
        <button
          type="button"
          className={styles.quiet}
          disabled={value === undefined}
          onClick={() => {
            setValue(undefined)
            commit(undefined)
          }}
        >
          clear
        </button>
      </div>
    </Section>
  )
}

const TextField = ({
  label,
  hint,
  initial,
  rows,
  onSave
}: {
  label: string
  hint: string
  initial: string
  rows: number
  onSave: (text: string) => void
}) => {
  const [text, setText] = useState(initial)
  const [saved, setSaved] = useState(initial)
  return (
    <Section label={label}>
      <textarea
        className={styles.text}
        rows={rows}
        value={text}
        placeholder={hint}
        onChange={(event) => setText(event.target.value)}
        onBlur={() => {
          if (text.trim() === saved.trim()) return
          setSaved(text)
          onSave(text)
        }}
      />
    </Section>
  )
}

const Tags = ({
  tags,
  onAdd,
  onRemove
}: {
  tags: string[]
  onAdd: (tag: string) => void
  onRemove: (tag: string) => void
}) => {
  const [draft, setDraft] = useState('')
  const add = () => {
    const tag = draft.trim().toLowerCase()
    if (!tag) return
    onAdd(tag)
    setDraft('')
  }
  return (
    <div className={styles.chips}>
      {tags.map((tag) => (
        <button
          key={tag}
          type="button"
          className={styles.chip}
          aria-label={`remove ${tag}`}
          onClick={() => onRemove(tag)}
        >
          {tagLabel(tag)} ×
        </button>
      ))}
      <input
        className={styles.tagInput}
        value={draft}
        placeholder="add tag"
        enterKeyHint="done"
        onChange={(event) => setDraft(event.target.value)}
        onKeyDown={(event) => {
          if (event.key === 'Enter') add()
        }}
        onBlur={add}
      />
    </div>
  )
}

const Moments = ({
  moments,
  onAdd,
  onDelete
}: {
  moments: Moment[]
  onAdd: (moment: { at: string; note: string; visibility: MomentVisibility }) => void
  onDelete: (id: number) => void
}) => {
  const [at, setAt] = useState(today)
  const [note, setNote] = useState('')
  const [visibility, setVisibility] = useState<MomentVisibility>('inherit')
  return (
    <div className={styles.moments}>
      {moments.map((m) => (
        <div key={m.id} className={styles.moment}>
          <span className={styles.momentDate}>{m.at}</span>
          <span className={styles.momentNote}>
            {m.note}
            {m.visibility === 'private' && <span className={styles.quiet}> · private</span>}
          </span>
          <button type="button" className={styles.quiet} onClick={() => onDelete(m.id)}>
            delete
          </button>
        </div>
      ))}
      <div className={styles.newMoment}>
        <input type="date" className={styles.date} value={at} onChange={(e) => setAt(e.target.value)} />
        <textarea
          className={styles.text}
          rows={2}
          value={note}
          placeholder="replayed it, still great"
          onChange={(e) => setNote(e.target.value)}
        />
        <Chips options={MOMENT_VISIBILITIES} chosen={visibility} onChoose={setVisibility} />
        <button
          type="button"
          className={styles.add}
          disabled={!at || !note.trim()}
          onClick={() => {
            onAdd({ at, note: note.trim(), visibility })
            setNote('')
          }}
        >
          add moment
        </button>
      </div>
    </div>
  )
}
