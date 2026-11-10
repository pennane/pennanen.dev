'use client'

import Image from 'next/image'
import { useState } from 'react'
import {
  divisionName,
  presenceOf,
  statusLabel,
  tagLabel,
  TIME_SPENT_CHOICES,
  TIME_SPENT_LABELS,
  type TimeSpent
} from '../../../lib/log/present'
import type {
  Axis,
  Division,
  Entry,
  EntryChange,
  Kind,
  Moment,
  MomentVisibility,
  Visibility
} from '../../../lib/log/types'
import { SaveStatus, useKeys } from '../_queue/card'
import { Compare, compareKeys, type Comparing } from '../_queue/Compare'
import {
  divisionOf,
  neighboursIn,
  ranksIn,
  startSearch,
  type Placement,
  type Ranked,
  type Rival,
  type Search
} from '../_queue/placing'
import { useSaveQueue } from '../_queue/useSaveQueue'
import { addMoment, changeEntry, deleteMoment, placeEntry } from '../actions'
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

export type EntrySubject = {
  itemId: number
  title: string
  cover: string | undefined
  publicPath: string
  entry: Entry
  moments: Moment[]
}

export const Editor = ({
  subject: { itemId, title, cover, publicPath, entry, moments: initialMoments },
  kind,
  knownTags,
  ranked,
  onClose,
  onChange,
  onPlace
}: {
  subject: EntrySubject
  kind: Kind
  knownTags: string[]
  ranked: Ranked
  onClose: () => void
  onChange: (change: EntryChange) => void
  onPlace: (rival: Rival, placement: Placement) => void
}) => {
  const [status, setStatus] = useState(entry.status)
  const [visibility, setVisibility] = useState(entry.visibility)
  const [tags, setTags] = useState(entry.tags)
  const [timeSpent, setTimeSpent] = useState<TimeSpent | typeof UNSAID>(entry.timeSpent ?? UNSAID)
  const [moments, setMoments] = useState(initialMoments)
  const saves = useSaveQueue()

  const save = (label: string, change: EntryChange) => {
    onChange(change)
    saves.enqueue(label, () => changeEntry(itemId, change))
  }

  const presence = presenceOf({ status, visibility }, kind.statusesShown)

  return (
    <div className={styles.editor}>
      <header className={styles.header}>
        <button type="button" className={styles.quiet} onClick={onClose}>
          ← entries
        </button>
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

      {kind.axes.map((axis) =>
        axis.key === kind.rankedAxis ? (
          <RankedPicker
            key={axis.key}
            axis={axis}
            rival={{ id: itemId, title, cover }}
            ranked={ranked}
            onPlace={(placement) => {
              onPlace({ id: itemId, title, cover }, placement)
              const where = `${divisionName(axis, placement.division)} #${placement.index + 1}`
              saves.enqueue(`${axis.name} → ${where}`, () =>
                placeEntry(itemId, placement.division, placement.index)
              )
            }}
            onClear={() => save(`${axis.name} → none`, { scores: { [axis.key]: null } })}
          />
        ) : (
          <DivisionPicker
            key={axis.key}
            label={axis.name}
            divisions={axis.divisions}
            initial={entry.scores[axis.key]}
            onSave={(division) =>
              save(`${axis.name} → ${division ?? 'none'}`, {
                scores: { [axis.key]: division ?? null }
              })
            }
          />
        )
      )}

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

      <Section label="tags">
        <Tags
          tags={tags}
          suggestions={knownTags.filter((t) => !tags.includes(t))}
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
            saves.enqueue(`moment ${moment.at}`, async () =>
              setMoments(await addMoment(itemId, moment))
            )
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

const NO_DIVISION = ''

const NO_KEYS = {}

const neighbourLine = (ranked: Ranked, id: number) => {
  const neighbours = neighboursIn(ranked, id)
  if (!neighbours) return undefined
  const { above, below } = neighbours
  if (above && below) return `below ${above.title} · above ${below.title}`
  if (below) return `top of all, above ${below.title}`
  if (above) return `bottom of all, below ${above.title}`
  return 'the only one ranked'
}

const RankedPicker = ({
  axis,
  rival,
  ranked,
  onPlace,
  onClear
}: {
  axis: Axis
  rival: Rival
  ranked: Ranked
  onPlace: (placement: Placement) => void
  onClear: () => void
}) => {
  const [search, setSearch] = useState<Search>()
  const division = divisionOf(ranked, rival.id)
  const rank = ranksIn(ranked).get(rival.id)

  const place = (placement: Placement) => {
    setSearch(undefined)
    onPlace(placement)
  }

  const begin = (key: string) => {
    const started = startSearch(ranked, key, rival.id)
    if (started.high === 0) place({ division: key, index: 0 })
    else setSearch(started)
  }

  const comparing: Comparing | undefined = search && {
    ranked,
    search,
    id: rival.id,
    divisionName: divisionName(axis, search.division),
    onSearch: setSearch,
    onPlace: place,
    onBack: () => setSearch(undefined)
  }
  useKeys(comparing ? compareKeys(comparing) : NO_KEYS)

  return (
    <Section label={rank === undefined ? axis.name : `${axis.name} · #${rank}`}>
      {comparing ? (
        <Compare {...comparing} />
      ) : (
        <>
          <Chips
            options={[
              ...axis.divisions.map((d) => ({ value: d.key, label: d.name })),
              { value: NO_DIVISION, label: NO_SCORE }
            ]}
            chosen={division ?? NO_DIVISION}
            onChoose={(next) => (next === NO_DIVISION ? onClear() : begin(next))}
          />
          {division && (
            <p className={styles.standing}>
              <span>{neighbourLine(ranked, rival.id)}</span>
              <button type="button" className={styles.quiet} onClick={() => begin(division)}>
                re-place in {divisionName(axis, division)}
              </button>
            </p>
          )}
        </>
      )}
    </Section>
  )
}

const DivisionPicker = ({
  label,
  divisions,
  initial,
  onSave
}: {
  label: string
  divisions: Division[]
  initial: string | undefined
  onSave: (division: string | undefined) => void
}) => {
  const [chosen, setChosen] = useState(initial ?? NO_DIVISION)
  return (
    <Section label={label}>
      <Chips
        options={[
          ...divisions.map((d) => ({ value: d.key, label: d.name })),
          { value: NO_DIVISION, label: NO_SCORE }
        ]}
        chosen={chosen}
        onChoose={(next) => {
          setChosen(next)
          onSave(next === NO_DIVISION ? undefined : next)
        }}
      />
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

const TAG_SUGGESTIONS = 'tag-suggestions'

const Tags = ({
  tags,
  suggestions,
  onAdd,
  onRemove
}: {
  tags: string[]
  suggestions: string[]
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
        list={TAG_SUGGESTIONS}
        enterKeyHint="done"
        onChange={(event) => setDraft(event.target.value)}
        onKeyDown={(event) => {
          if (event.key === 'Enter') add()
        }}
        onBlur={add}
      />
      <datalist id={TAG_SUGGESTIONS}>
        {suggestions.map((tag) => (
          <option key={tag} value={tag} label={tagLabel(tag)} />
        ))}
      </datalist>
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
        <input
          type="date"
          className={styles.date}
          value={at}
          onChange={(e) => setAt(e.target.value)}
        />
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
