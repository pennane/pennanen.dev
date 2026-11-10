import {
  byRank,
  byTitle,
  PRESTIGE,
  statusLabel,
  tagLabel,
  tagWithAncestors,
  type ListedItem
} from '../../../../lib/log/present'
import type { Axis } from '../../../../lib/log/types'

export const MODES = ['tiers', 'list'] as const

export type Mode = (typeof MODES)[number]

export type Direction = 'forward' | 'reverse'

export type Sort = { by: string; direction: Direction }

export type View = {
  mode: Mode
  axis: string
  sort: Sort
  query: string
  chosen: Record<string, string[]>
}

type ReadableParams = { get: (name: string) => string | null }

const PARAMS = { mode: 'view', axis: 'axis', sort: 'sort', direction: 'order', query: 'q' }

const REVERSED = 'reverse'

const VALUE_SEPARATOR = ','

const RANK_SORT = 'rank'

const DEFAULT_SORT: Sort = { by: RANK_SORT, direction: 'forward' }

const isMode = (text: string | null): text is Mode => MODES.some((mode) => mode === text)

export const rankedFirst = (axis: string) => (axis === PRESTIGE ? byRank : byTitle)

export const divisionIndex = (axis: Axis, item: ListedItem) => {
  const at = axis.divisions.findIndex((division) => division.key === item.scores[axis.key])
  return at === -1 ? undefined : at
}

export type Facet = {
  key: string
  name: string
  valuesOf: (item: ListedItem) => string[]
  labelOf: (value: string) => string
  order: (a: string, b: string) => number
}

const decadeOf = (date: string | undefined) => {
  const year = Number(date?.slice(0, 4))
  return year ? [`${Math.floor(year / 10) * 10}s`] : []
}

const PLAYTIME_BANDS = [
  { key: 'under-2h', label: 'under 2 h', below: 2 * 60 },
  { key: '2-10h', label: '2–10 h', below: 10 * 60 },
  { key: '10-50h', label: '10–50 h', below: 50 * 60 },
  { key: '50h-plus', label: '50 h +', below: Number.POSITIVE_INFINITY }
]

const bandOf = (minutes: number | undefined) =>
  minutes === undefined ? [] : [PLAYTIME_BANDS.find((band) => minutes < band.below)!.key]

const bandOrder = (key: string) => PLAYTIME_BANDS.findIndex((band) => band.key === key)

const alphabetical = (a: string, b: string) => a.localeCompare(b)

const HAS_VISIONS = '+visions'

const SCORE_PREFIX = 'score-'

export const axisKey = (axis: Axis) => SCORE_PREFIX + axis.key

export const facetsFor = (
  axes: Axis[],
  statuses: string[],
  tagTitle: (tag: string) => string | undefined
): Facet[] => [
  {
    key: 'status',
    name: 'status',
    valuesOf: (item) => [item.status],
    labelOf: statusLabel,
    order: (a, b) => statuses.indexOf(a) - statuses.indexOf(b) || alphabetical(a, b)
  },
  ...axes.map((axis) => ({
    key: axisKey(axis),
    name: axis.name,
    valuesOf: (item: ListedItem) => (item.scores[axis.key] ? [item.scores[axis.key]] : []),
    labelOf: (value: string) => axis.divisions.find((d) => d.key === value)?.name ?? value,
    order: (a: string, b: string) =>
      axis.divisions.findIndex((d) => d.key === a) - axis.divisions.findIndex((d) => d.key === b)
  })),
  {
    key: 'genre',
    name: 'genre',
    valuesOf: (item) => item.genres,
    labelOf: (value) => value.toLowerCase(),
    order: alphabetical
  },
  {
    key: 'released',
    name: 'released',
    valuesOf: (item) => decadeOf(item.released),
    labelOf: (value) => value,
    order: alphabetical
  },
  {
    key: 'playtime',
    name: 'playtime',
    valuesOf: (item) => bandOf(item.playtimeMinutes),
    labelOf: (value) => PLAYTIME_BANDS.find((band) => band.key === value)?.label ?? value,
    order: (a, b) => bandOrder(a) - bandOrder(b)
  },
  {
    key: 'tag',
    name: 'tag',
    valuesOf: (item) => [
      ...(item.proseWords > 0 ? [HAS_VISIONS] : []),
      ...new Set(item.tags.flatMap(tagWithAncestors))
    ],
    labelOf: (value) =>
      value === HAS_VISIONS ? 'has visions' : (tagTitle(value) ?? tagLabel(value)),
    order: (a, b) => Number(b === HAS_VISIONS) - Number(a === HAS_VISIONS) || alphabetical(a, b)
  }
]

export const valuesIn = (facet: Facet, items: ListedItem[]) => {
  const counts = new Map<string, number>()
  for (const item of items) {
    for (const value of facet.valuesOf(item)) counts.set(value, (counts.get(value) ?? 0) + 1)
  }
  return [...counts]
    .map(([value, count]) => ({ value, count }))
    .sort((a, b) => facet.order(a.value, b.value))
}

export const splitsItems = (facet: Facet, items: ListedItem[]) => {
  const values = valuesIn(facet, items)
  return values.length > 1 || (values.length === 1 && values[0].count < items.length)
}

export const parseView = (params: ReadableParams, axes: Axis[], facets: Facet[]): View => {
  const mode = params.get(PARAMS.mode)
  const axis = params.get(PARAMS.axis)
  const sort = params.get(PARAMS.sort)
  return {
    mode: isMode(mode) ? mode : 'tiers',
    axis: axes.some((a) => a.key === axis) && axis ? axis : (axes[0]?.key ?? PRESTIGE),
    sort: sort
      ? {
          by: sort,
          direction: params.get(PARAMS.direction) === REVERSED ? 'reverse' : 'forward'
        }
      : DEFAULT_SORT,
    query: params.get(PARAMS.query) ?? '',
    chosen: Object.fromEntries(
      facets.flatMap((facet) => {
        const values = params.get(facet.key)?.split(VALUE_SEPARATOR).filter(Boolean) ?? []
        return values.length > 0 ? [[facet.key, values]] : []
      })
    )
  }
}

export const searchFor = (view: View, axes: Axis[]) => {
  const query = new URLSearchParams()
  if (view.mode !== 'tiers') query.set(PARAMS.mode, view.mode)
  if (view.axis !== axes[0]?.key) query.set(PARAMS.axis, view.axis)
  if (view.sort.by !== DEFAULT_SORT.by || view.sort.direction !== DEFAULT_SORT.direction) {
    query.set(PARAMS.sort, view.sort.by)
    if (view.sort.direction === 'reverse') query.set(PARAMS.direction, REVERSED)
  }
  if (view.query) query.set(PARAMS.query, view.query)
  for (const [key, values] of Object.entries(view.chosen)) {
    if (values.length > 0) query.set(key, values.join(VALUE_SEPARATOR))
  }
  const search = query.toString()
  return search ? `?${search}` : ''
}

const matchesQuery = (query: string) => {
  const needle = query.trim().toLowerCase()
  return (item: ListedItem) =>
    !needle ||
    item.title.toLowerCase().includes(needle) ||
    item.developers.some((developer) => developer.toLowerCase().includes(needle))
}

const matchesFacet = (facet: Facet, chosen: string[] | undefined) => (item: ListedItem) =>
  !chosen || chosen.length === 0 || facet.valuesOf(item).some((value) => chosen.includes(value))

export const filterItems = (items: ListedItem[], view: View, facets: Facet[], except?: string) =>
  items
    .filter(matchesQuery(view.query))
    .filter((item) =>
      facets.every(
        (facet) => facet.key === except || matchesFacet(facet, view.chosen[facet.key])(item)
      )
    )

type Comparable = string | number

export type Column = {
  key: string
  name: string
  valueOf: (item: ListedItem) => Comparable | undefined
  order: 'ascending' | 'descending'
}

const compareValues = (a: Comparable, b: Comparable) =>
  typeof a === 'number' && typeof b === 'number' ? a - b : String(a).localeCompare(String(b))

export const columnsFor = (axes: Axis[]): Column[] => [
  { key: RANK_SORT, name: '#', valueOf: (item) => item.rank?.place, order: 'ascending' },
  { key: 'title', name: 'title', valueOf: (item) => item.title, order: 'ascending' },
  ...axes.map((axis): Column => ({
    key: axisKey(axis),
    name: axis.name,
    valueOf: (item) => divisionIndex(axis, item),
    order: 'ascending'
  })),
  { key: 'status', name: 'status', valueOf: (item) => item.status, order: 'ascending' },
  { key: 'released', name: 'released', valueOf: (item) => item.released, order: 'descending' },
  {
    key: 'last-played',
    name: 'last played',
    valueOf: (item) => item.lastPlayed,
    order: 'descending'
  }
]

const SIGNS = { ascending: 1, descending: -1 }

const DIRECTION_SIGNS = { forward: 1, reverse: -1 }

export const sortItems = (items: ListedItem[], sort: Sort, columns: Column[]) => {
  const column = columns.find((c) => c.key === sort.by) ?? columns[0]
  const sign = SIGNS[column.order] * DIRECTION_SIGNS[sort.direction]
  return items.toSorted((a, b) => {
    const va = column.valueOf(a)
    const vb = column.valueOf(b)
    if (va === undefined || vb === undefined) {
      return Number(va === undefined) - Number(vb === undefined) || byRank(a, b)
    }
    return sign * compareValues(va, vb) || byRank(a, b)
  })
}
