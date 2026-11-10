'use client'

import { useEffect, useRef } from 'react'
import type { ListedItem } from '../../../../lib/log/present'
import styles from './toolbar.module.css'
import { filterItems, valuesIn, type Facet, type View } from './view'

const toggled = (values: string[], value: string) =>
  values.includes(value) ? values.filter((v) => v !== value) : [...values, value]

const useClosesOnOutsideClick = () => {
  const bar = useRef<HTMLDivElement>(null)
  useEffect(() => {
    const close = (event: Event) => {
      for (const open of bar.current?.querySelectorAll('details[open]') ?? []) {
        if (!(event.target instanceof Node) || !open.contains(event.target)) {
          open.removeAttribute('open')
        }
      }
    }
    const closeOnEscape = (event: KeyboardEvent) => {
      if (event.key === 'Escape') close(event)
    }
    document.addEventListener('click', close)
    document.addEventListener('keydown', closeOnEscape)
    return () => {
      document.removeEventListener('click', close)
      document.removeEventListener('keydown', closeOnEscape)
    }
  }, [])
  return bar
}

const FacetMenu = ({
  facet,
  items,
  view,
  facets,
  onChoose
}: {
  facet: Facet
  items: ListedItem[]
  view: View
  facets: Facet[]
  onChoose: (values: string[]) => void
}) => {
  const chosen = view.chosen[facet.key] ?? []
  const counted = valuesIn(facet, filterItems(items, view, facets, facet.key))
  const known = new Set(counted.map(({ value }) => value))
  const options = [
    ...counted,
    ...chosen.filter((value) => !known.has(value)).map((value) => ({ value, count: 0 }))
  ]
  return (
    <details className={styles.menu}>
      <summary className={styles.button} data-chosen={chosen.length > 0 ? 'some' : 'none'}>
        {facet.name}
        {chosen.length > 0 && <span className={styles.chosenCount}>{chosen.length}</span>}
        <span aria-hidden>▾</span>
      </summary>
      <ul className={styles.options}>
        {options.map(({ value, count }) => (
          <li key={value}>
            <label data-count={count === 0 ? 'none' : 'some'}>
              <input
                type="checkbox"
                checked={chosen.includes(value)}
                onChange={() => onChoose(toggled(chosen, value))}
              />
              <span>{facet.labelOf(value)}</span>
              <span className={styles.count}>{count}</span>
            </label>
          </li>
        ))}
      </ul>
    </details>
  )
}

export const Toolbar = ({
  items,
  facets,
  view,
  onChange
}: {
  items: ListedItem[]
  facets: Facet[]
  view: View
  onChange: (view: View) => void
}) => {
  const bar = useClosesOnOutsideClick()
  const choose = (facet: Facet, values: string[]) =>
    onChange({ ...view, chosen: { ...view.chosen, [facet.key]: values } })
  const pills = facets.flatMap((facet) =>
    (view.chosen[facet.key] ?? []).map((value) => ({ facet, value }))
  )
  return (
    <div className={styles.toolbar} ref={bar}>
      <div className={styles.controls}>
        <input
          className={styles.search}
          type="search"
          value={view.query}
          placeholder="search title or developer"
          aria-label="search"
          onChange={(event) => onChange({ ...view, query: event.target.value })}
        />
        {facets.map((facet) => (
          <FacetMenu
            key={facet.key}
            facet={facet}
            items={items}
            view={view}
            facets={facets}
            onChoose={(values) => choose(facet, values)}
          />
        ))}
      </div>
      {pills.length > 0 && (
        <div className={styles.pills}>
          {pills.map(({ facet, value }) => (
            <button
              key={`${facet.key}:${value}`}
              type="button"
              className={styles.pill}
              onClick={() => choose(facet, toggled(view.chosen[facet.key] ?? [], value))}
            >
              <span className={styles.pillName}>{facet.name}</span> {facet.labelOf(value)} ×
            </button>
          ))}
          <button
            type="button"
            className={styles.clear}
            onClick={() => onChange({ ...view, chosen: {}, query: '' })}
          >
            clear all
          </button>
        </div>
      )}
    </div>
  )
}
