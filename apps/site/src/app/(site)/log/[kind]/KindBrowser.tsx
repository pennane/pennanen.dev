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

export const KindBrowser = ({ items, kind }: { items: ListedItem[]; kind: string }) => {
  const axes = useKindAxes(kind)
  const { kinds } = usePresentation()
  const tagTitle = useTagTitle()
  const facets = facetsFor(axes, kinds[kind]?.statusesShown ?? [], tagTitle).filter((facet) =>
    splitsItems(facet, items)
  )
  const view = parseView(useSearchParams(), axes, facets)
  const change = (next: View) =>
    window.history.replaceState(null, '', `${window.location.pathname}${searchFor(next, axes)}`)
  const axis = axes.find((a) => a.key === view.axis)

  return (
    <div className={styles.browser}>
      <div className={styles.bar}>
        <Switch
          label="view"
          options={MODES.map((mode) => ({ value: mode, label: MODE_LABELS[mode] }))}
          chosen={view.mode}
          onChoose={(mode) => change({ ...view, mode })}
        />
        {view.mode === 'tiers' && axes.length > 1 && (
          <Switch
            label="axis"
            options={axes.map((a) => ({ value: a.key, label: a.name }))}
            chosen={view.axis}
            onChoose={(key) => change({ ...view, axis: key })}
          />
        )}
      </div>
      {view.mode === 'tiers' && axis && <TierBoard items={items} axis={axis} />}
      {view.mode === 'list' && (
        <ListView items={items} axes={axes} facets={facets} view={view} onChange={change} />
      )}
    </div>
  )
}

const ListView = ({
  items,
  axes,
  facets,
  view,
  onChange
}: {
  items: ListedItem[]
  axes: Axis[]
  facets: ReturnType<typeof facetsFor>
  view: View
  onChange: (view: View) => void
}) => {
  const columns = columnsFor(axes)
  const shown = sortItems(filterItems(items, view, facets), view.sort, columns)
  return (
    <>
      <Toolbar items={items} facets={facets} view={view} onChange={onChange} />
      <p className={styles.count}>
        {shown.length === items.length ? `${items.length}` : `${shown.length} of ${items.length}`}
      </p>
      <ItemTable
        items={shown}
        axes={axes}
        columns={columns}
        sort={view.sort}
        onSort={(sort) => onChange({ ...view, sort })}
      />
    </>
  )
}

export const DefaultBoard = ({ items, kind }: { items: ListedItem[]; kind: string }) => {
  const [axis] = useKindAxes(kind)
  return (
    <div className={styles.browser}>
      <div className={styles.bar}>
        <Switch
          label="view"
          options={MODES.map((mode) => ({ value: mode, label: MODE_LABELS[mode] }))}
          chosen="tiers"
          onChoose={() => undefined}
        />
      </div>
      {axis && <TierBoard items={items} axis={axis} />}
    </div>
  )
}
