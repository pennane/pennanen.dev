'use client'

import { useSearchParams } from 'next/navigation'
import type { ListedItem } from '../../../../lib/log/present'
import { ItemTable } from './ItemTable'
import styles from './kind.module.css'
import { TierBoard } from './TierBoard'
import { Toolbar } from './Toolbar'
import { useSetup, type Setup } from './setup'
import { filterItems, MODES, parseView, searchFor, sortItems, type View } from './view'

const MODE_LABELS = { tiers: 'tiers', list: 'list' }

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
  const { tiers, listedAxes, facets, searched, sorts, columns, defaults } = setup
  const search = searchFor(view, defaults)
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
      {mode === 'tiers' && axis && <TierBoard items={filtered} axis={axis} search={search} />}
      {mode === 'list' && (
        <ItemTable
          items={sortItems(filtered, view.sort, sorts)}
          axes={listedAxes}
          columns={columns}
          sort={view.sort}
          search={search}
          onSort={(sort) => onChange({ ...view, sort })}
        />
      )}
    </div>
  )
}

export const KindBrowser = ({ items, kind }: { items: ListedItem[]; kind: string }) => {
  const setup = useSetup(items, kind)
  const view = parseView(useSearchParams(), setup.defaults)
  const change = (next: View) =>
    window.history.replaceState(
      null,
      '',
      `${window.location.pathname}${searchFor(next, setup.defaults)}`
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
      view={parseView(NO_PARAMS, setup.defaults)}
      onChange={() => undefined}
    />
  )
}
