import type { PublicItem } from './public-api'
import type { Axis, FieldUse, KindSettings, Presentation, PublicField } from './types'

export type Rank = { place: number; of: number }

type Fields = Record<string, unknown>

type Takes = PublicItem['takes']

export const publicKinds = (presentation: Presentation) => Object.keys(presentation.kinds).sort()

export const isPublicKind = (presentation: Presentation, kind: string) =>
  Object.hasOwn(presentation.kinds, kind)

export const takeValues = (takes: Takes, source: string) =>
  takes.find((take) => take.source === source)?.values ?? {}

export const kindTitle = (presentation: Presentation, kind: string) =>
  presentation.kinds[kind]?.title.trim() || kind

export const plainText = (markdown: string) =>
  markdown
    .replace(/!?\[([^\]]*)\]\([^)]*\)/g, '$1')
    .replace(/[*_#>`]/g, '')
    .replace(/\s+/g, ' ')
    .trim()

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
  return Array.isArray(value) ? value.filter((v): v is string => typeof v === 'string') : []
}

export const titleOf = (item: PublicItem) => textField(item.fields, 'title') ?? item.slug

export const pictureOf = (item: PublicItem) =>
  textField(item.fields, 'cover') ?? textField(item.fields, 'image')

export const itemPath = (item: { kind: string; slug: string }) => `/log/${item.kind}/${item.slug}`

export const statusLabel = (status: string) => status.replaceAll('-', ' ')

const TAG_SEPARATOR = '/'

export const tagLabel = (tag: string) => tag.split(TAG_SEPARATOR).join(' › ')

export const tagIsWithin = (tag: string, parent: string) =>
  tag === parent || tag.startsWith(parent + TAG_SEPARATOR)

export const tagWithAncestors = (tag: string) =>
  tag.split(TAG_SEPARATOR).map((_, i, parts) => parts.slice(0, i + 1).join(TAG_SEPARATOR))

const FAVORITE_TAG = 'favorite'

export const isFavorite = (tags: string[]) => tags.some((tag) => tagIsWithin(tag, FAVORITE_TAG))

export const PRESTIGE = 'prestige'

export type ShownScore = { axis: string; axisName: string; division: string }

export const findAxis = (axes: Axis[], key: string) => axes.find((axis) => axis.key === key)

export const divisionName = (axis: Axis | undefined, key: string) =>
  axis?.divisions.find((d) => d.key === key)?.name ?? key

export const shownScores = (
  presentation: Presentation,
  kind: string,
  scores: Record<string, string>
): ShownScore[] => {
  const order = presentation.kinds[kind]?.axes ?? []
  const place = (axis: string) => (order.includes(axis) ? order.indexOf(axis) : order.length)
  return Object.entries(scores)
    .sort(([a], [b]) => place(a) - place(b) || a.localeCompare(b))
    .map(([key, division]) => {
      const axis = findAxis(presentation.axes, key)
      return { axis: key, axisName: axis?.name ?? key, division: divisionName(axis, division) }
    })
}

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

export type FieldValue = string | number | string[]

export type ListedItem = {
  kind: string
  slug: string
  title: string
  picture: string | undefined
  poster: string | undefined
  status: string
  scores: Record<string, string>
  rank: Rank | undefined
  tags: string[]
  proseWords: number
  values: Record<string, FieldValue>
}

const asFieldValue = (value: unknown): FieldValue | undefined => {
  if (typeof value === 'string') return value === '' ? undefined : value
  if (typeof value === 'number') return value
  if (Array.isArray(value)) {
    const strings = value.filter((v): v is string => typeof v === 'string')
    return strings.length > 0 ? strings : undefined
  }
  return undefined
}

export const fieldValueOf = (item: PublicItem, field: PublicField) =>
  asFieldValue(field.take ? takeValues(item.takes, field.take)[field.name] : item.fields[field.key])

export const toListed = (item: PublicItem, fields: PublicField[]): ListedItem => ({
  kind: item.kind,
  slug: item.slug,
  title: titleOf(item),
  picture: pictureOf(item),
  poster: textField(item.fields, 'poster'),
  status: item.status,
  scores: item.scores,
  rank: item.rank,
  tags: item.tags,
  proseWords: item.proseWords,
  values: Object.fromEntries(
    fields.flatMap((field) => {
      const value = fieldValueOf(item, field)
      return value === undefined ? [] : [[field.key, value]]
    })
  )
})

const TAKE_SUFFIX = '-take'

export const sourceName = (source: string) =>
  source.endsWith(TAKE_SUFFIX) ? source.slice(0, -TAKE_SUFFIX.length) : source

export const fieldName = (settings: KindSettings | undefined, field: PublicField) =>
  settings?.fieldNames[field.key] ??
  [field.take && sourceName(field.take), field.name.replaceAll('_', ' ')].filter(Boolean).join(' ')

export const usesOf = (settings: KindSettings | undefined, field: PublicField) =>
  settings?.fieldUses[field.key] ?? field.defaultUses

export const fieldsUsedAs = (
  presentation: Presentation,
  kind: string,
  use: FieldUse
): PublicField[] =>
  (presentation.fields[kind] ?? []).filter((field) =>
    usesOf(presentation.kinds[kind], field).includes(use)
  )

export const formatFieldValue = (field: PublicField, value: FieldValue) => {
  if (Array.isArray(value)) return value.join(', ')
  if (field.type === 'partial-date' && typeof value === 'string') return formatPartialDate(value)
  return String(value)
}

export const byTitle = (a: ListedItem, b: ListedItem) => a.title.localeCompare(b.title)

const placeOf = (item: ListedItem) => item.rank?.place ?? Number.MAX_SAFE_INTEGER

export const byRank = (a: ListedItem, b: ListedItem) => placeOf(a) - placeOf(b) || byTitle(a, b)

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

const HALF_STAR = 0.5

const starsOf = (rating: number) =>
  '★'.repeat(Math.floor(rating)) + (rating % 1 >= HALF_STAR ? '½' : '')

const RATED_IN_STARS = new Set(['letterboxd-take'])

const LIKED = 'liked'

const RECOMMENDATION_MARKS: Record<string, string> = {
  recommended: '👍',
  'not-recommended': '👎'
}

const TEXT_KEYS = ['review', 'note']

export type ShownTake = { source: string; marks: string[]; text: string | undefined }

export const shownTake = (take: PublicItem['takes'][number]): ShownTake => {
  const rating = numberField(take.values, 'rating')
  const ratingMark =
    rating === undefined ? [] : [RATED_IN_STARS.has(take.source) ? starsOf(rating) : String(rating)]
  const likedMark = textField(take.values, 'reaction') === LIKED ? ['♥'] : []
  const recommendation = textField(take.values, 'recommendation')
  const recommendationMark =
    recommendation && Object.hasOwn(RECOMMENDATION_MARKS, recommendation)
      ? [RECOMMENDATION_MARKS[recommendation]]
      : []
  const text = TEXT_KEYS.map((key) => textField(take.values, key)).find(Boolean)
  return {
    source: sourceName(take.source),
    marks: [...ratingMark, ...likedMark, ...recommendationMark],
    text
  }
}

const COUNT_SUFFIX = '_unlocked'

const TOTAL_SUFFIX = '_total'

export type PageField = { key: string; name: string; value: string }

export const pageFieldsOf = (item: PublicItem, presentation: Presentation): PageField[] => {
  const settings = presentation.kinds[item.kind]
  const shown = fieldsUsedAs(presentation, item.kind, 'page').flatMap((field) => {
    const value = fieldValueOf(item, field)
    return value === undefined ? [] : [{ field, value }]
  })
  const baseOfCount = (field: PublicField) =>
    field.key.endsWith(COUNT_SUFFIX) ? field.key.slice(0, -COUNT_SUFFIX.length) : undefined
  const totalOf = (field: PublicField) => {
    const base = baseOfCount(field)
    return base === undefined ? undefined : shown.find((s) => s.field.key === base + TOTAL_SUFFIX)
  }
  const totalsShownWithCounts = new Set(
    shown.flatMap(({ field }) => totalOf(field)?.field.key ?? [])
  )
  return shown.flatMap(({ field, value }) => {
    if (totalsShownWithCounts.has(field.key)) return []
    const total = totalOf(field)
    if (total === undefined) {
      return [
        { key: field.key, name: fieldName(settings, field), value: formatFieldValue(field, value) }
      ]
    }
    if (total.value === 0) return []
    const base = baseOfCount(field) ?? field.key
    return [
      {
        key: field.key,
        name: settings?.fieldNames[field.key] ?? base.replaceAll('_', ' '),
        value: `${value} / ${total.value}`
      }
    ]
  })
}
