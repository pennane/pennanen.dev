import type { TimeSpent } from './present'

export type Hint = 'likely-idled' | 'likely-played' | 'likely-never-played' | 'likely-not-a-game'

export type Visibility = 'auto' | 'public' | 'private'

export type Division = { key: string; name: string }

export type Axis = { key: string; name: string; divisions: Division[] }

export type KindSettings = {
  title: string
  description: string
  statusesShown: string[]
  axes: string[]
}

export type Presentation = {
  axes: Axis[]
  kinds: Record<string, KindSettings>
  tagTitles: Record<string, string>
}

export type Kind = {
  name: string
  statuses: string[]
  statusesToScore: string[]
  statusesShown: string[]
  rankedAxis: string
  axes: Axis[]
  fields: { name: string; type: string }[]
}

export type Entry = {
  status: string
  visibility: Visibility
  blurb: string
  prose: string
  timeSpent?: TimeSpent
  scores: Record<string, string>
  rank?: number
  tags: string[]
  updatedAt: string
}

export type ListedEntry = {
  id: number
  title: string
  cover?: string
  status: string
  visibility: Visibility
  scores: Record<string, string>
  rank?: number
  tags: string[]
}

export type FieldValue = { value: string | undefined; source: string | undefined }

export type ItemView = {
  id: number
  kind: string
  slug: string
  createdAt: string
  verdicts: { values: Record<string, unknown>; sourceOf: Record<string, string> }
  entry: Entry | null
  hint?: Hint
}

export type UploadTicket = { ticket: string; url: string }

export type ItemLink = { title: string; href: string }

export type EntryChange = {
  status?: string
  visibility?: Visibility
  blurb?: string
  prose?: string
  timeSpent?: TimeSpent | ''
  scores?: Record<string, string | null>
  addTags?: string[]
  removeTags?: string[]
}

export type IGDBResult = {
  igdbId: number
  name: string
  year?: string
  platforms: string[]
  poster?: string
  itemId?: number
}

export type IGDBMatchReport = {
  state: 'never-matched' | 'matching' | 'matched' | 'failed'
  startedAt?: string
  finishedAt?: string
  checked: number
  matched: number
  problems: string[]
}

export type SteamSyncReport = {
  state: 'never-synced' | 'syncing' | 'synced' | 'failed'
  startedAt?: string
  finishedAt?: string
  gamesSeen: number
  gamesAdded: number
  problems: string[]
}
