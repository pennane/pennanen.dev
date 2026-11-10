import {
  byLastPlayed,
  byScore,
  byTitle,
  tagIsWithin,
  tagWithAncestors,
  type ListedItem
} from '../../../../lib/log/present'

export const SORTS = ['score', 'last-played', 'title'] as const

export type Sort = (typeof SORTS)[number]

export const SORT_LABELS: Record<Sort, string> = {
  score: 'score',
  'last-played': 'last played',
  title: 'title'
}

export type View = {
  sort: Sort
  status: string | undefined
  tag: string | undefined
}

export const DEFAULT_VIEW: View = {
  sort: 'score',
  status: undefined,
  tag: undefined
}

type ReadableParams = { get: (name: string) => string | null }

const isSort = (text: string | null): text is Sort =>
  SORTS.some((sort) => sort === text)

export const parseView = (params: ReadableParams): View => {
  const sort = params.get('sort')
  return {
    sort: isSort(sort) ? sort : DEFAULT_VIEW.sort,
    status: params.get('status') || undefined,
    tag: params.get('tag') || undefined
  }
}

export const hrefFor = (path: string, view: View) => {
  const query = new URLSearchParams()
  if (view.sort !== DEFAULT_VIEW.sort) query.set('sort', view.sort)
  if (view.status) query.set('status', view.status)
  if (view.tag) query.set('tag', view.tag)
  const search = query.toString()
  return search ? `${path}?${search}` : path
}

const COMPARATORS: Record<Sort, (a: ListedItem, b: ListedItem) => number> = {
  score: byScore,
  'last-played': byLastPlayed,
  title: byTitle
}

const hasStatus = (status: string | undefined) => (item: ListedItem) =>
  !status || item.status === status

const hasTagWithin = (parent: string | undefined) => (item: ListedItem) =>
  !parent || item.tags.some((tag) => tagIsWithin(tag, parent))

export const applyView = (items: ListedItem[], view: View) =>
  items
    .filter(hasStatus(view.status))
    .filter(hasTagWithin(view.tag))
    .sort(COMPARATORS[view.sort])

const countBy = (values: string[]) =>
  [...Map.groupBy(values, (value) => value)]
    .map(([value, all]) => ({ value, count: all.length }))
    .sort((a, b) => a.value.localeCompare(b.value))

export const statusesIn = (items: ListedItem[]) =>
  countBy(items.map((item) => item.status))

export const tagsIn = (items: ListedItem[]) =>
  countBy(
    items.flatMap((item) => [...new Set(item.tags.flatMap(tagWithAncestors))])
  )
