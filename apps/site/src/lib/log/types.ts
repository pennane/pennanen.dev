import type { TimeSpent } from './present'

export type Hint = 'likely-idled' | 'likely-played' | 'likely-never-played' | 'likely-not-a-game'

export type Visibility = 'auto' | 'public' | 'private'

export type Division = { key: string; name: string }

export type AxisSort = 'own' | 'outside' | 'crowd'

export type Reading = {
  sort: Exclude<AxisSort, 'own'>
  source: string
  field: string
  floors: number[]
}

export const TMDB_ID = 'tmdb_id'

export type ReadFrom = 'take' | 'description'

export type Readable = { source: string; field: string; from: ReadFrom }

export const READ_FROM_BY_SORT: Record<Reading['sort'], ReadFrom> = {
  outside: 'take',
  crowd: 'description'
}

export type Axis = { key: string; name: string; divisions: Division[]; reading?: Reading }

export const sortOf = (axis: Axis): AxisSort => axis.reading?.sort ?? 'own'

export const DEFAULT_VIEWS = ['auto', 'tiers', 'list'] as const

export type DefaultView = (typeof DEFAULT_VIEWS)[number]

export const FIELD_USES = ['filter', 'column', 'page'] as const

export type FieldUse = (typeof FIELD_USES)[number]

export const FIELD_TYPES = [
  'text',
  'int',
  'number',
  'partial-date',
  'url',
  'string-list',
  'external-id'
] as const

export type FieldType = (typeof FIELD_TYPES)[number]

export const DEFAULT_OVERVIEW_COUNT = 5

export type PublicField = {
  key: string
  name: string
  type: FieldType
  take?: string
  defaultUses: FieldUse[]
}

export type Overview = { sort: string; count: number }

export type KindSettings = {
  title: string
  description: string
  statusesShown: string[]
  axes: string[]
  defaultView: DefaultView
  axesWithoutTiers: string[]
  axesWithoutColumns: string[]
  fieldUses: Record<string, FieldUse[]>
  fieldNames: Record<string, string>
  overview: Overview
}

export type Presentation = {
  axes: Axis[]
  kinds: Record<string, KindSettings>
  fields: Record<string, PublicField[]>
  tagTitles: Record<string, string>
  readables: Readable[]
}

export type Kind = {
  name: string
  statuses: string[]
  statusesToScore: string[]
  statusesShown: string[]
  rankedAxis: string
  axes: Axis[]
  fields: { name: string; type: string }[]
  takes: { source: string; privateFields: string[] }[]
}

export type Take = { source: string; values: Record<string, unknown>; fetchedAt: string }

export type TakeView = Take & { hidden: string[] }

export type TakeFieldVisibility = 'shown' | 'hidden'

export const SET_BY_AP = 'ap'

export type Entry = {
  status: string
  statusSetBy: string
  visibility: Visibility
  blurb: string
  prose: string
  timeSpent?: TimeSpent
  scores: Record<string, string>
  readScores: Record<string, string>
  rank?: number
  tags: string[]
  tagsAddedBySources?: Record<string, string>
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
  takes: TakeView[]
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

export type SyncState = 'never-ran' | 'running' | 'succeeded' | 'failed'

export type SyncReport = {
  name: string
  state: SyncState
  startedAt?: string
  finishedAt?: string
  lastSucceededAt?: string
  counts: Record<string, number>
  problems: string[]
}
