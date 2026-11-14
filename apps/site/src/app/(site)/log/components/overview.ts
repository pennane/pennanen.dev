import { logPublic } from '../../../../lib/log/public-api'
import { kindTitle, publicKinds, toListed, type ListedItem } from '../../../../lib/log/present'
import type { Presentation } from '../../../../lib/log/types'
import { overviewSort, sortItems, sortsFor } from '../[kind]/view'

export type KindOverview = { kind: string; title: string; items: ListedItem[]; shown: number }

export const kindOverview = async (
  presentation: Presentation,
  kind: string
): Promise<KindOverview> => {
  const settings = presentation.kinds[kind]
  const fields = presentation.fields[kind] ?? []
  const axes = settings.axes.flatMap((key) => presentation.axes.filter((a) => a.key === key))
  const items = (await logPublic.items(kind)).map((item) => toListed(item, fields))
  const sorts = sortsFor(axes, fields, settings)
  const sort = overviewSort(items, axes, sorts, settings)
  return {
    kind,
    title: kindTitle(presentation, kind),
    items: sortItems(items, { by: sort.key, direction: 'forward' }, sorts),
    shown: settings.overview.count
  }
}

export const logOverview = async (presentation: Presentation) =>
  (
    await Promise.all(publicKinds(presentation).map((kind) => kindOverview(presentation, kind)))
  ).filter((overview) => overview.items.length > 0)

export const listed = (words: string[]) =>
  words.length < 2 ? words.join('') : `${words.slice(0, -1).join(', ')} and ${words.at(-1)}`

export const countsOf = (overviews: KindOverview[]) =>
  overviews.map(({ items, title }) => `${items.length} ${title}`)

export const postersOf = (items: ListedItem[]) => items.flatMap((item) => item.poster ?? [])

export const roundRobin = <T>(lists: T[][]): T[] =>
  Array.from({ length: Math.max(0, ...lists.map((list) => list.length)) }).flatMap((_, i) =>
    lists.flatMap((list) => (i < list.length ? [list[i]] : []))
  )
