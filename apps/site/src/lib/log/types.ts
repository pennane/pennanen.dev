import type { TimeSpent } from './present'

export type Visibility = 'auto' | 'public' | 'private'

export type Division = { key: string; name: string }

export type AxisSort = 'own' | 'outside' | 'crowd' | 'activity'

export type Reading = {
  sort: Exclude<AxisSort, 'own'>
  source: string
  field: string
  floors: number[]
}

export const TMDB_ID = 'tmdb_id'

export type TMDBType = 'movie' | 'tv'

export type TMDBLink = { tmdbType: TMDBType; tmdbId: number }

export const TMDB_TYPE_OF_SOURCE: Record<string, TMDBType> = { tmdb: 'movie', 'tmdb-tv': 'tv' }

export const tmdbPage = ({ tmdbType, tmdbId }: TMDBLink) =>
  `https://www.themoviedb.org/${tmdbType}/${tmdbId}`

export type ReadFrom = 'take' | 'description' | 'activity'

export type Readable = { source: string; field: string; from: ReadFrom }

export const READ_FROM_BY_SORT: Record<Reading['sort'], ReadFrom> = {
  outside: 'take',
  crowd: 'description',
  activity: 'activity'
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

export const SHAPES = ['tall', 'wide', 'square'] as const

export type Shape = (typeof SHAPES)[number]

export type PictureField = { key: string; shape: Shape }

export type Picture = { url: string; shape: Shape }

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
  pictures: Record<string, PictureField[]>
  tagTitles: Record<string, string>
  readables: Readable[]
  kindOrder: string[]
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
  identitySource?: string
  unrankedOrder?: string
}

export type Take = { source: string; values: Record<string, unknown>; fetchedAt: string }

export type TakeView = Take & { hidden: string[] }

export type TakeFieldVisibility = 'shown' | 'hidden'

export const SET_BY_AP = 'ap'

export const DECIDED_BY_NOBODY = 'nobody'

export type ChoiceStanding = 'decides' | 'outvoted' | 'set-aside'

export type StatusChoice = { status: string; chosenAt: string; standing: ChoiceStanding }

export type StatusProposal = { rule: string; value: string; changedAt?: string }

export type Suggestion = { rule: string; subject: string; value: string }

export type Entry = {
  status: string
  statusDecidedBy: string
  statusChoice: StatusChoice | null
  statusProposals: StatusProposal[]
  visibility: Visibility
  blurb: string
  prose: string
  timeSpent?: TimeSpent
  scores: Record<string, string>
  readScores: Record<string, string>
  rank?: number
  tags: string[]
  tagsFromRules: Record<string, string>
  loggedBy: string
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

export type Relation = { fromId: number; type: string; toId: number }

export const INHERITED_SOURCE_PREFIX = 'inherited:'

export type InheritedValue = { field: string; value: string; from: string }

export type ItemView = {
  id: number
  kind: string
  slug: string
  createdAt: string
  verdicts: { values: Record<string, unknown>; sourceOf: Record<string, string> }
  entry: Entry | null
  dismissedAt?: string
  suggestions: Suggestion[]
  takes: TakeView[]
  relations?: Relation[]
}

export type UploadTicket = { ticket: string; url: string }

export type ItemLink = { title: string; href: string }

export type DismissalChange = 'dismissed' | 'kept'

export type EntryChange = {
  status?: string
  visibility?: Visibility
  blurb?: string
  prose?: string
  timeSpent?: TimeSpent | ''
  scores?: Record<string, string | null>
  addTags?: string[]
  removeTags?: string[]
  dismissal?: DismissalChange
}

export type IGDBResult = {
  igdbId: number
  name: string
  year?: string
  platforms: string[]
  poster?: string
  itemId?: number
}

export type TMDBResult = TMDBLink & {
  name: string
  year?: string
  poster?: string
  itemId?: number
}

export type SpotifyConnection = 'not-configured' | 'not-connected' | 'connected'

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
