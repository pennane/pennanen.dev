import Link from 'next/link'
import type { CSSProperties } from 'react'
import {
  formatPartialDate,
  itemPath,
  statusLabel,
  type ListedItem
} from '../../../../lib/log/present'
import type { Axis } from '../../../../lib/log/types'
import styles from './table.module.css'
import { tierColour } from './tierColours'
import { axisKey, divisionIndex, type Column, type Sort } from './view'

const yearOf = (date: string | undefined) => date?.slice(0, 4)

const META_COLUMNS = ['status', 'released', 'last-played']

const ARROWS = { forward: '↓', reverse: '↑' }

const gridFor = (axes: Axis[]) =>
  ({ '--axis-columns': `repeat(${axes.length}, 2.25rem)` }) as CSSProperties

const Heading = ({
  column,
  sort,
  onSort,
  className
}: {
  column: Column
  sort: Sort
  onSort: (sort: Sort) => void
  className?: string
}) => {
  const active = sort.by === column.key
  return (
    <button
      type="button"
      className={className}
      title={column.name}
      aria-pressed={active}
      onClick={() =>
        onSort({
          by: column.key,
          direction: active && sort.direction === 'forward' ? 'reverse' : 'forward'
        })
      }
    >
      {column.name}
      {active && <span className={styles.arrow}>{ARROWS[sort.direction]}</span>}
    </button>
  )
}

const AxisCell = ({ axis, item }: { axis: Axis; item: ListedItem }) => {
  const at = divisionIndex(axis, item)
  return (
    <span className={styles.axis}>
      {at !== undefined && (
        <span
          className={styles.badge}
          style={{ background: tierColour(at) }}
          title={`${axis.name}: ${axis.divisions[at].name}`}
        >
          {axis.divisions[at].name}
        </span>
      )}
    </span>
  )
}

export const ItemTable = ({
  items,
  axes,
  columns,
  sort,
  onSort
}: {
  items: ListedItem[]
  axes: Axis[]
  columns: Column[]
  sort: Sort
  onSort: (sort: Sort) => void
}) => {
  const columnFor = (key: string) => columns.find((column) => column.key === key)!
  const rank = columnFor('rank')
  const title = columnFor('title')
  const axisColumns = axes.map((axis) => columnFor(axisKey(axis)))
  const metaColumns = META_COLUMNS.map(columnFor)
  return (
    <div className={styles.table} style={gridFor(axes)}>
      <div className={`${styles.row} ${styles.head}`}>
        <Heading column={rank} sort={sort} onSort={onSort} className={styles.rank} />
        <span />
        <Heading column={title} sort={sort} onSort={onSort} className={styles.title} />
        {axisColumns.map((column) => (
          <Heading key={column.key} column={column} sort={sort} onSort={onSort} />
        ))}
        <span className={styles.meta}>
          {metaColumns.map((column) => (
            <Heading key={column.key} column={column} sort={sort} onSort={onSort} />
          ))}
        </span>
      </div>
      {items.length === 0 && <p className={styles.nothing}>Nothing matches.</p>}
      <ol className={styles.body}>
        {items.map((item) => (
          <li key={item.slug} className={styles.row}>
            <span className={styles.rank}>{item.rank?.place}</span>
            <Link href={itemPath(item)} className={styles.thumb} tabIndex={-1} aria-hidden>
              {item.picture && <img src={item.picture} alt="" loading="lazy" />}
            </Link>
            <Link href={itemPath(item)} className={styles.title}>
              {item.title}
            </Link>
            {axes.map((axis) => (
              <AxisCell key={axis.key} axis={axis} item={item} />
            ))}
            <span className={styles.meta}>
              <span>{statusLabel(item.status)}</span>
              <span>{yearOf(item.released)}</span>
              <span>{item.lastPlayed && formatPartialDate(item.lastPlayed.slice(0, 7))}</span>
            </span>
          </li>
        ))}
      </ol>
    </div>
  )
}
