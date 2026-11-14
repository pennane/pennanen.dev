'use client'

import { useSearchParams } from 'next/navigation'
import { fieldsUsedAs, type ListedItem } from '../../../../lib/log/present'
import type { Axis, KindSettings, PublicField } from '../../../../lib/log/types'
import { usePresentation, useTagTitle } from '../components/PresentationContext'
import { ItemTable } from './ItemTable'
import styles from './kind.module.css'
import { TierBoard } from './TierBoard'
import { Toolbar } from './Toolbar'
import {
  axisKey,
  columnAxes,
  defaultModeFor,
  facetsFor,
  filterItems,
  MODES,
  parseView,
  searchedFields,
  searchFor,
  sortItems,
  sortsFor,
  splitsItems,
  tierAxes,
  type Column,
  type Facet,
  type Mode,
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

type Setup = {
  axes: Axis[]
  tiers: Axis[]
  listedAxes: Axis[]
  facets: Facet[]
  searched: PublicField[]
  sorts: Column[]
  columns: Column[]
  defaultMode: Mode
}

const useSetup = (items: ListedItem[], kind: string): Setup => {
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
    defaultMode: defaultModeFor(items, tiers, settings?.defaultView)
  }
}

const Browser = ({
  items,
  setup,
  view,
  onChange
}: {
  items: ListedItem[]
  setup: Setup
  view: View
  onChange: (view: View) => void
}) => {
  const { tiers, listedAxes, facets, searched, sorts, columns } = setup
  const axis = tiers.find((a) => a.key === view.axis) ?? tiers[0]
  const filtered = filterItems(items, view, facets, searched)
  const mode = tiers.length > 0 ? view.mode : 'list'
  return (
    <div className={styles.browser}>
      <div className={styles.bar}>
        {tiers.length > 0 && (
          <Switch
            label="view"
            options={MODES.map((m) => ({ value: m, label: MODE_LABELS[m] }))}
            chosen={mode}
            onChoose={(next) => onChange({ ...view, mode: next })}
          />
        )}
        {mode === 'tiers' && tiers.length > 1 && axis && (
          <Switch
            label="axis"
            options={tiers.map((a) => ({ value: a.key, label: a.name }))}
            chosen={axis.key}
            onChoose={(key) => onChange({ ...view, axis: key })}
          />
        )}
      </div>
      <Toolbar items={items} facets={facets} searched={searched} view={view} onChange={onChange} />
      <p className={styles.count}>
        {filtered.length === items.length
          ? `${items.length}`
          : `${filtered.length} of ${items.length}`}
      </p>
      {mode === 'tiers' && axis && <TierBoard items={filtered} axis={axis} />}
      {mode === 'list' && (
        <ItemTable
          items={sortItems(filtered, view.sort, sorts)}
          axes={listedAxes}
          columns={columns}
          sort={view.sort}
          onSort={(sort) => onChange({ ...view, sort })}
        />
      )}
    </div>
  )
}

export const KindBrowser = ({ items, kind }: { items: ListedItem[]; kind: string }) => {
  const setup = useSetup(items, kind)
  const view = parseView(useSearchParams(), setup.tiers, setup.facets, setup.defaultMode)
  const change = (next: View) =>
    window.history.replaceState(
      null,
      '',
      `${window.location.pathname}${searchFor(next, setup.tiers, setup.defaultMode)}`
    )
  return <Browser items={items} setup={setup} view={view} onChange={change} />
}

const NO_PARAMS = new URLSearchParams()

export const DefaultBoard = ({ items, kind }: { items: ListedItem[]; kind: string }) => {
  const setup = useSetup(items, kind)
  return (
    <Browser
      items={items}
      setup={setup}
      view={parseView(NO_PARAMS, setup.tiers, setup.facets, setup.defaultMode)}
      onChange={() => undefined}
    />
  )
}
