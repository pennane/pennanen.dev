'use client'

import { useSearchParams } from 'next/navigation'
import type { ListedItem } from '../../../../lib/log/present'
import type { Axis } from '../../../../lib/log/types'
import { usePresentation, useTagTitle } from '../components/PresentationContext'
import { ItemTable } from './ItemTable'
import styles from './kind.module.css'
import { TierBoard } from './TierBoard'
import { Toolbar } from './Toolbar'
import {
  columnsFor,
  facetsFor,
  filterItems,
  MODES,
  parseView,
  searchFor,
  sortItems,
  splitsItems,
  type Facet,
  type View
} from './view'

const MODE_LABELS = { tiers: 'tiers', list: 'list' }

export const useKindAxes = (kind: string): Axis[] => {
  const { axes, kinds } = usePresentation()
  return (kinds[kind]?.axes ?? []).flatMap((key) => axes.filter((axis) => axis.key === key))
}

const Switch = <T extends string>({
  options,
  chosen,
  label,
  onChoose
}: {
  options: { value: T; label: string }[]
  chosen: T
  label: string
  onChoose: (value: T) => void
}) => (
  <div className={styles.switch} role="group" aria-label={label}>
    {options.map((option) => (
      <button
        key={option.value}
        type="button"
        aria-pressed={option.value === chosen}
        onClick={() => onChoose(option.value)}
      >
        {option.label}
      </button>
    ))}
  </div>
)

const useFacets = (items: ListedItem[], kind: string, axes: Axis[]) => {
  const { kinds } = usePresentation()
  const tagTitle = useTagTitle()
  return facetsFor(axes, kinds[kind]?.statusesShown ?? [], tagTitle).filter((facet) =>
    splitsItems(facet, items)
  )
}

const NO_PARAMS = new URLSearchParams()

const Browser = ({
  items,
  axes,
  facets,
  view,
  onChange
}: {
  items: ListedItem[]
  axes: Axis[]
  facets: Facet[]
  view: View
  onChange: (view: View) => void
}) => {
  const axis = axes.find((a) => a.key === view.axis)
  const filtered = filterItems(items, view, facets)
  const columns = columnsFor(axes)
  return (
    <div className={styles.browser}>
      <div className={styles.bar}>
        <Switch
          label="view"
          options={MODES.map((mode) => ({ value: mode, label: MODE_LABELS[mode] }))}
          chosen={view.mode}
          onChoose={(mode) => onChange({ ...view, mode })}
        />
        {view.mode === 'tiers' && axes.length > 1 && (
          <Switch
            label="axis"
            options={axes.map((a) => ({ value: a.key, label: a.name }))}
            chosen={view.axis}
            onChoose={(key) => onChange({ ...view, axis: key })}
          />
        )}
      </div>
      <Toolbar items={items} facets={facets} view={view} onChange={onChange} />
      <p className={styles.count}>
        {filtered.length === items.length
          ? `${items.length}`
          : `${filtered.length} of ${items.length}`}
      </p>
      {view.mode === 'tiers' && axis && <TierBoard items={filtered} axis={axis} />}
      {view.mode === 'list' && (
        <ItemTable
          items={sortItems(filtered, view.sort, columns)}
          axes={axes}
          columns={columns}
          sort={view.sort}
          onSort={(sort) => onChange({ ...view, sort })}
        />
      )}
    </div>
  )
}

export const KindBrowser = ({ items, kind }: { items: ListedItem[]; kind: string }) => {
  const axes = useKindAxes(kind)
  const facets = useFacets(items, kind, axes)
  const view = parseView(useSearchParams(), axes, facets)
  const change = (next: View) =>
    window.history.replaceState(null, '', `${window.location.pathname}${searchFor(next, axes)}`)
  return <Browser items={items} axes={axes} facets={facets} view={view} onChange={change} />
}

export const DefaultBoard = ({ items, kind }: { items: ListedItem[]; kind: string }) => {
  const axes = useKindAxes(kind)
  const facets = useFacets(items, kind, axes)
  return (
    <Browser
      items={items}
      axes={axes}
      facets={facets}
      view={parseView(NO_PARAMS, axes, facets)}
      onChange={() => undefined}
    />
  )
}
