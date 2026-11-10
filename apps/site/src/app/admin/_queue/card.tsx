'use client'

import Image from 'next/image'
import Link from 'next/link'
import { useEffect } from 'react'
import { divisionName, findAxis, PRESTIGE } from '../../../lib/log/present'
import type { Hint, Kind } from '../../../lib/log/types'
import styles from './queue.module.css'
import type { SaveQueueState } from './useSaveQueue'

export type GameCard = {
  id: number
  title: string
  cover?: string
  playtimeMinutes?: number
  lastPlayed?: string
  achievements?: { unlocked: number; total: number }
  steamCategories: string[]
  status: string
  tags: string[]
  hint?: Hint
}

export const ANOTHER_SHOT = 'another-shot'
export const IDLED = 'idled'
const STEAM_TRADING_CARDS = 'Steam Trading Cards'

export const prestigeDivisions = (kind: Kind) => findAxis(kind.axes, PRESTIGE)?.divisions ?? []

export const prestigeName = (kind: Kind, division: string) =>
  divisionName(findAxis(kind.axes, PRESTIGE), division)

export const HINT_LABELS: Record<Hint, string> = {
  'likely-played': 'looks played',
  'likely-idled': 'looks idled',
  'likely-never-played': 'never launched',
  'likely-not-a-game': 'not a game'
}

export const formatPlaytime = (minutes: number) => {
  if (minutes < 60) return `${minutes} min`
  const hours = Math.floor(minutes / 60)
  return minutes % 60 === 0 ? `${hours} h` : `${hours} h ${minutes % 60} min`
}

export const useKeys = (handlers: Record<string, () => void>) => {
  useEffect(() => {
    const onKey = (event: KeyboardEvent) => {
      if (event.metaKey || event.ctrlKey || event.altKey) return
      if (event.target instanceof HTMLInputElement) return
      const handler = handlers[event.key]
      if (!handler) return
      event.preventDefault()
      handler()
    }
    window.addEventListener('keydown', onKey)
    return () => window.removeEventListener('keydown', onKey)
  }, [handlers])
}

const QUEUES = [
  { href: '/admin/triage', label: 'triage' },
  { href: '/admin/rate', label: 'rate' },
  { href: '/admin/add', label: 'add' },
  { href: '/admin/match', label: 'match' },
  { href: '/admin/entries', label: 'entries' },
  { href: '/admin/order', label: 'order' },
  { href: '/admin/settings', label: 'settings' }
] as const

export const QueueHeader = ({
  current,
  count,
  children
}: {
  current: (typeof QUEUES)[number]['label']
  count?: number
  children?: React.ReactNode
}) => (
  <header className={styles.header}>
    <h1>
      {current} {count !== undefined && <span className={styles.count}>{count}</span>}
    </h1>
    <nav className={styles.queues}>
      {QUEUES.filter((q) => q.label !== current).map((q) => (
        <Link key={q.href} href={q.href} className={styles.quiet}>
          {q.label}
        </Link>
      ))}
      {children}
    </nav>
  </header>
)

export const CardFace = ({ card, badge }: { card: GameCard; badge?: string }) => {
  const facts = [
    card.playtimeMinutes !== undefined && formatPlaytime(card.playtimeMinutes),
    card.achievements && `${card.achievements.unlocked}/${card.achievements.total} achievements`,
    card.lastPlayed && `last played ${card.lastPlayed}`,
    card.steamCategories.includes(STEAM_TRADING_CARDS) && 'drops cards'
  ].filter(Boolean)
  return (
    <>
      <div className={styles.cover}>
        {card.cover ? (
          <Image src={card.cover} alt="" width={460} height={215} unoptimized priority />
        ) : (
          <span className={styles.noCover}>{card.title}</span>
        )}
      </div>
      <h2 className={styles.title}>{card.title}</h2>
      <p className={styles.facts}>
        {badge && <span className={styles.hint}>{badge}</span>}
        {facts.join(' · ')}
      </p>
    </>
  )
}

export const TagToggle = ({
  tag,
  shortcut,
  label,
  pressed,
  onToggle
}: {
  tag: string
  shortcut: string
  label: string
  pressed: boolean
  onToggle: (tag: string) => void
}) => (
  <button className={styles.filter} aria-pressed={pressed} onClick={() => onToggle(tag)}>
    <kbd>{shortcut}</kbd> {label}
  </button>
)

export const SaveStatus = ({ state }: { state: SaveQueueState }) => {
  if (state.kind === 'saved') return <span className={styles.saved}>saved</span>
  if (state.kind === 'saving') return <span className={styles.saving}>saving {state.unsaved}…</span>
  return (
    <span className={styles.retrying} title={state.error}>
      {state.unsaved} unsaved, retrying ({state.label})
    </span>
  )
}
