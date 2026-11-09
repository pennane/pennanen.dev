export type Hint = 'likely-idled' | 'likely-played' | 'likely-never-played'

export type Visibility = 'private' | 'public'

export type Kind = {
  name: string
  statuses: string[]
  statusesToScore: string[]
  axes: string[]
}

export type Entry = {
  status: string
  visibility: Visibility
  notes: string
  blurb: string
  scores: Record<string, number>
  tags: string[]
  updatedAt: string
}

export type ItemView = {
  id: number
  kind: string
  slug: string
  createdAt: string
  verdicts: { values: Record<string, unknown>; sourceOf: Record<string, string> }
  entry: Entry | null
  hint?: Hint
}

export type EntryChange = {
  status?: string
  visibility?: Visibility
  notes?: string
  blurb?: string
  scores?: Record<string, number | null>
  addTags?: string[]
  removeTags?: string[]
}

export type SteamSyncReport = {
  state: 'never-synced' | 'syncing' | 'synced' | 'failed'
  startedAt?: string
  finishedAt?: string
  gamesSeen: number
  gamesAdded: number
  problems: string[]
}
