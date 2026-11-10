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
  FieldValue,
  Kind,
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
import dynamic from 'next/dynamic'
import { Prose } from '../../(site)/log/components/Prose'
import {
  changeEntry,
  changeLastPlayed,
  findItemLinks,
  imageUploadTicket,
  placeEntry
} from '../actions'
import styles from './editor.module.css'

const VISIBILITY_CHOICES: { value: Visibility; label: string }[] = [
  { value: 'auto', label: 'auto, by status' },
  { value: 'public', label: 'always shown' },
  { value: 'private', label: 'never shown' }
]
const UNTRIAGED = 'untriaged'
const UNSAID = ''
const NO_SCORE = '–'

const MAX_IMAGE_BYTES = 10 * 1024 * 1024

const ProseEditor = dynamic(() => import('./ProseEditor').then((m) => m.ProseEditor), {
  ssr: false
})

type Writing = 'reading' | 'writing'

const uploadImage = async (file: File) => {
  const { ticket, url } = await imageUploadTicket()
  const response = await fetch(url, {
    method: 'POST',
    headers: { Authorization: `Ticket ${ticket}`, 'Content-Type': file.type },
    body: file
  })
  const body = (await response.json().catch(() => ({}))) as { name?: string; error?: string }
  if (!response.ok || !body.name) throw new Error(body.error ?? `upload failed: ${response.status}`)
  return `/log/images/${body.name}`
}

export type EntrySubject = {
  itemId: number
  title: string
  cover: string | undefined
  publicPath: string
  entry: Entry
  lastPlayed: FieldValue
}

export const Editor = ({
  subject: { itemId, title, cover, publicPath, entry, lastPlayed },
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
  const [prose, setProse] = useState(entry.prose)
  const [writing, setWriting] = useState<Writing>('reading')
  const [imageProblem, setImageProblem] = useState<string>()
  const saves = useSaveQueue()

  const save = (label: string, change: EntryChange) => {
    onChange(change)
    saves.enqueue(label, () => changeEntry(itemId, change))
  }

  const presence = presenceOf({ status, visibility }, kind.statusesShown)

  const saveProse = (markdown: string) => {
    if (markdown === prose) return
    setProse(markdown)
    save('visions', { prose: markdown })
  }

  const addImage = (file: File, insert: (src: string) => void) => {
    if (file.size > MAX_IMAGE_BYTES) {
      setImageProblem(`${file.name} is over 10 MB`)
      return
    }
    setImageProblem(undefined)
    saves.enqueue(`image ${file.name}`, async () => insert(await uploadImage(file)))
  }

  if (writing === 'writing') {
    return (
      <div
        className={styles.writing}
        onKeyDown={(event) => {
          if (event.key === 'Escape' && !event.defaultPrevented) setWriting('reading')
        }}
      >
        <header className={styles.header}>
          <button type="button" className={styles.quiet} onClick={() => setWriting('reading')}>
            ← done
          </button>
          <SaveStatus state={saves.state} />
        </header>
        <div className={styles.writingBody}>
          {cover && (
            <div className={styles.cover}>
              <Image src={cover} alt="" width={460} height={215} unoptimized />
            </div>
          )}
          <h1 className={styles.title}>{title}</h1>
          {imageProblem && <p className={styles.problem}>{imageProblem}</p>}
          <ProseEditor
            initial={prose}
            onSave={saveProse}
            onImage={addImage}
            findLinks={findItemLinks}
          />
        </div>
      </div>
    )
  }

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

      {kind.fields.some((field) => field.name === LAST_PLAYED_FIELD) && (
        <LastPlayed
          initial={lastPlayed}
          onSave={(date, show) =>
            saves.enqueue(`last played → ${date ?? "steam's"}`, async () =>
              show(await changeLastPlayed(itemId, date))
            )
          }
        />
      )}

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

      <Section label="visions">
        {prose.trim() ? (
          <div className={styles.visions}>
            <Prose markdown={prose} />
          </div>
        ) : (
          <p className={styles.quiet}>nothing written yet</p>
        )}
        <button type="button" className={styles.add} onClick={() => setWriting('writing')}>
          write
        </button>
      </Section>
    </div>
  )
}

const LAST_PLAYED_FIELD = 'last_played'

const PARTIAL_DATE = /^\d{4}(-\d{2}(-\d{2})?)?$/

const SOURCE_NOTES: Record<string, string> = { ap: 'set by you', steam: 'from steam' }

const today = () => {
  const now = new Date()
  const pad = (n: number) => String(n).padStart(2, '0')
  return `${now.getFullYear()}-${pad(now.getMonth() + 1)}-${pad(now.getDate())}`
}

const LastPlayed = ({
  initial,
  onSave
}: {
  initial: FieldValue
  onSave: (date: string | null, show: (saved: FieldValue) => void) => void
}) => {
  const [saved, setSaved] = useState(initial)
  const [text, setText] = useState(initial.value ?? '')
  const show = (next: FieldValue) => {
    setSaved(next)
    setText(next.value ?? '')
  }
  const commit = (date: string) => {
    const trimmed = date.trim()
    if (trimmed === (saved.value ?? '') || !PARTIAL_DATE.test(trimmed)) return
    setText(trimmed)
    onSave(trimmed, show)
  }
  const invalid = text.trim() !== '' && !PARTIAL_DATE.test(text.trim())
  return (
    <Section label="last played">
      <div className={styles.chips}>
        <input
          className={styles.tagInput}
          value={text}
          placeholder="2024, 2024-03 or 2024-03-05"
          aria-invalid={invalid}
          onChange={(event) => setText(event.target.value)}
          onBlur={() => commit(text)}
          onKeyDown={(event) => {
            if (event.key === 'Enter') commit(text)
          }}
        />
        <button type="button" className={styles.chip} onClick={() => commit(today())}>
          today
        </button>
        {saved.source === 'ap' && (
          <button type="button" className={styles.chip} onClick={() => onSave(null, show)}>
            use steam&apos;s
          </button>
        )}
      </div>
      <p className={styles.quiet}>
        {invalid
          ? 'write a year, year-month or year-month-day'
          : saved.source
            ? (SOURCE_NOTES[saved.source] ?? `from ${saved.source}`)
            : 'not known'}
      </p>
    </Section>
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
