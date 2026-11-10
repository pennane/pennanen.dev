export type Hint =
  | 'likely-idled'
  | 'likely-played'
  | 'likely-never-played'
  | 'likely-not-a-game'

export type Visibility = 'private' | 'public'

export type Kind = {
  name: string
  statuses: string[]
  statusesToScore: string[]
  statusesShown: string[]
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

export type MomentVisibility = 'inherit' | 'private'

export type Moment = {
  id: number
  at: string
  note: string
  visibility: MomentVisibility
  scores: Record<string, number> | null
}

export type ItemDetail = ItemView & { moments?: Moment[] }

export type NewMoment = {
  at: string
  note: string
  visibility: MomentVisibility
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
