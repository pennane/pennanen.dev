import type { PublicItem } from './public-api'

const KIND_PLURALS = { game: 'games', thing: 'things' } as const

export type PublicKind = keyof typeof KIND_PLURALS

export const PUBLIC_KINDS = Object.keys(KIND_PLURALS) as PublicKind[]

export const isPublicKind = (kind: string): kind is PublicKind =>
  Object.hasOwn(KIND_PLURALS, kind)

export const kindPlural = (kind: PublicKind) => KIND_PLURALS[kind]

type Fields = Record<string, unknown>

export const textField = (fields: Fields, name: string) => {
  const value = fields[name]
  return typeof value === 'string' && value !== '' ? value : undefined
}

export const numberField = (fields: Fields, name: string) => {
  const value = fields[name]
  return typeof value === 'number' ? value : undefined
}

export const listField = (fields: Fields, name: string) => {
  const value = fields[name]
  return Array.isArray(value)
    ? value.filter((v): v is string => typeof v === 'string')
    : []
}

const FIELDS_WITH_A_PLACE = new Set([
  'title',
  'description',
  'link',
  'image',
  'cover',
  'poster',
  'platforms',
  'released',
  'developers',
  'genres',
  'platform',
  'steam_categories',
  'steam_appid',
  'playtime_minutes',
  'last_played',
  'achievements_total',
  'achievements_unlocked'
])

export const freeFormFields = (fields: Fields) =>
  Object.entries(fields)
    .filter(([name]) => !FIELDS_WITH_A_PLACE.has(name))
    .flatMap(([name, value]) =>
      typeof value === 'string' || typeof value === 'number'
        ? [{ name: name.replaceAll('_', ' '), value: String(value) }]
        : []
    )

export const titleOf = (item: PublicItem) =>
  textField(item.fields, 'title') ?? item.slug

export const pictureOf = (item: PublicItem) =>
  textField(item.fields, 'cover') ?? textField(item.fields, 'image')

export const itemPath = (item: { kind: string; slug: string }) =>
  `/log/${item.kind}/${item.slug}`

export const statusLabel = (status: string) => status.replaceAll('-', ' ')

const TAG_SEPARATOR = '/'

export const tagLabel = (tag: string) => tag.split(TAG_SEPARATOR).join(' › ')

export const tagIsWithin = (tag: string, parent: string) =>
  tag === parent || tag.startsWith(parent + TAG_SEPARATOR)

export const tagWithAncestors = (tag: string) =>
  tag
    .split(TAG_SEPARATOR)
    .map((_, i, parts) => parts.slice(0, i + 1).join(TAG_SEPARATOR))

const FAVORITE_TAG = 'favorite'

export const isFavorite = (tags: string[]) =>
  tags.some((tag) => tagIsWithin(tag, FAVORITE_TAG))

const AXES_IN_ORDER = ['enjoyment', 'quality']

const axisRank = (axis: string) => {
  const rank = AXES_IN_ORDER.indexOf(axis)
  return rank === -1 ? AXES_IN_ORDER.length : rank
}

export const orderedScores = (scores: Record<string, number>) =>
  Object.entries(scores)
    .map(([axis, score]) => ({ axis, score }))
    .sort((a, b) => axisRank(a.axis) - axisRank(b.axis) || a.axis.localeCompare(b.axis))

export const HEADLINE_AXIS = 'enjoyment'

const monthYear = new Intl.DateTimeFormat('en-GB', {
  year: 'numeric',
  month: 'short',
  timeZone: 'UTC'
})

const dayMonthYear = new Intl.DateTimeFormat('en-GB', {
  year: 'numeric',
  month: 'short',
  day: 'numeric',
  timeZone: 'UTC'
})

export const formatPartialDate = (text: string) => {
  const [date] = text.split('T')
  const [year, month, day] = date.split('-').map(Number)
  if (!year) return text
  if (!month) return String(year)
  if (!day) return monthYear.format(Date.UTC(year, month - 1))
  return dayMonthYear.format(Date.UTC(year, month - 1, day))
}

export const TIME_SPENT_LABELS = {
  'not-enough': 'not enough time',
  'a-bit': 'a bit of time',
  'just-right': 'just the right amount of time',
  'a-lot': 'a lot of time',
  'too-much': 'too much time'
} as const

export type TimeSpent = keyof typeof TIME_SPENT_LABELS

export const TIME_SPENT_CHOICES = Object.keys(TIME_SPENT_LABELS) as TimeSpent[]

export const isTimeSpent = (value: unknown): value is TimeSpent =>
  typeof value === 'string' && Object.hasOwn(TIME_SPENT_LABELS, value)

export type ListedItem = {
  kind: string
  slug: string
  title: string
  picture: string | undefined
  status: string
  scores: Record<string, number>
  tags: string[]
  lastPlayed: string | undefined
}

export const toListed = (item: PublicItem): ListedItem => ({
  kind: item.kind,
  slug: item.slug,
  title: titleOf(item),
  picture: pictureOf(item),
  status: item.status,
  scores: item.scores,
  tags: item.tags,
  lastPlayed: textField(item.fields, 'last_played')
})

const UNSCORED = -1

const headlineScore = (item: ListedItem) =>
  item.scores[HEADLINE_AXIS] ?? UNSCORED

export const byTitle = (a: ListedItem, b: ListedItem) =>
  a.title.localeCompare(b.title)

const NEVER_PLAYED = ''

export const byLastPlayed = (a: ListedItem, b: ListedItem) =>
  (b.lastPlayed ?? NEVER_PLAYED).localeCompare(a.lastPlayed ?? NEVER_PLAYED) || byTitle(a, b)

export const byScore = (a: ListedItem, b: ListedItem) =>
  headlineScore(b) - headlineScore(a) || byTitle(a, b)

export const scoreWordFor = (score: number, words: { from: number; word: string }[]) =>
  words.findLast((w) => w.from <= score)?.word

export type Presence = 'shown' | 'hidden'

const NEVER_SHOWN = 'dismissed'

export const presenceOf = (
  entry: { status: string; visibility: 'auto' | 'public' | 'private' },
  statusesShown: string[]
): Presence => {
  if (entry.status === NEVER_SHOWN || entry.visibility === 'private') return 'hidden'
  if (entry.visibility === 'public') return 'shown'
  return statusesShown.includes(entry.status) ? 'shown' : 'hidden'
}
