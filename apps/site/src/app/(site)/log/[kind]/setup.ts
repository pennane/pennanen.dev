import { fieldsUsedAs, type ListedItem } from '../../../../lib/log/present'
import type { Axis, KindSettings, PublicField } from '../../../../lib/log/types'
import { usePresentation, useTagTitle } from '../components/PresentationContext'
import {
  axisKey,
  columnAxes,
  defaultModeFor,
  facetsFor,
  overviewSort,
  searchedFields,
  sortsFor,
  splitsItems,
  tierAxes,
  type Column,
  type Facet,
  type ViewDefaults
} from './view'

export type Setup = {
  axes: Axis[]
  tiers: Axis[]
  listedAxes: Axis[]
  facets: Facet[]
  searched: PublicField[]
  sorts: Column[]
  columns: Column[]
  defaults: ViewDefaults
}

export const useKindAxes = (kind: string): Axis[] => {
  const { axes, kinds } = usePresentation()
  return (kinds[kind]?.axes ?? []).flatMap((key) => axes.filter((axis) => axis.key === key))
}

export const useSetup = (items: ListedItem[], kind: string): Setup => {
  const presentation = usePresentation()
  const tagTitle = useTagTitle()
  const settings: KindSettings | undefined = presentation.kinds[kind]
  const axes = useKindAxes(kind)
  const tiers = tierAxes(axes, settings)
  const listedAxes = columnAxes(axes, settings)
  const unlistedAxisKeys = new Set(axes.filter((axis) => !listedAxes.includes(axis)).map(axisKey))
  const fields = presentation.fields[kind] ?? []
  const facets = facetsFor({
    axes,
    statuses: settings?.statusesShown ?? [],
    tagTitle,
    fields: fieldsUsedAs(presentation, kind, 'filter'),
    settings,
    items
  }).filter((facet) => splitsItems(facet, items))
  const sorts = sortsFor(axes, fields, settings)
  const columnKeys = new Set(fieldsUsedAs(presentation, kind, 'column').map((field) => field.key))
  const columns = sorts.filter(
    (sort) => !unlistedAxisKeys.has(sort.key) && (!sort.field || columnKeys.has(sort.key))
  )
  return {
    axes,
    tiers,
    listedAxes,
    facets,
    searched: searchedFields(fields),
    sorts,
    columns,
    defaults: {
      tiers,
      facets,
      mode: defaultModeFor(items, tiers, settings?.defaultView),
      sort: { by: overviewSort(items, axes, sorts, settings).key, direction: 'forward' }
    }
  }
}
