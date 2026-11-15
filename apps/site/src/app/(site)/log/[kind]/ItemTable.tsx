import Link from 'next/link'
import type { CSSProperties } from 'react'
import {
  bannerShapeOf,
  formatFieldValue,
  formatPartialDate,
  itemPath,
  statusLabel,
  type ListedItem
} from '../../../../lib/log/present'
import type { Axis } from '../../../../lib/log/types'
import styles from './table.module.css'
import { tierColour } from './tierColours'
import { axisKey, divisionIndex, type Column, type Sort } from './view'

const META_COLUMN_WIDTH = '4.75rem'

const shownValue = (column: Column, item: ListedItem) => {
  const value = column.valueOf(item)
  if (value === undefined || !column.field) return value
  if (column.field.type === 'partial-date' && typeof value === 'string') {
    return value.length > 4 ? formatPartialDate(value.slice(0, 7)) : value
  }
  return formatFieldValue(column.field, value)
}

const ARROWS = { forward: '↓', reverse: '↑' }

const gridFor = (axes: Axis[], metaColumns: Column[]) =>
  ({
    '--axis-columns': `repeat(${axes.length}, 2.25rem)`,
    '--meta-columns': `repeat(${metaColumns.length}, ${META_COLUMN_WIDTH})`
  }) as CSSProperties

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
  search,
  onSort
}: {
  items: ListedItem[]
  axes: Axis[]
  columns: Column[]
  sort: Sort
  search: string
  onSort: (sort: Sort) => void
}) => {
  const columnFor = (key: string) => columns.find((column) => column.key === key)!
  const rank = columnFor('rank')
  const title = columnFor('title')
  const axisColumns = axes.map((axis) => columnFor(axisKey(axis)))
  const axisKeys = new Set(axes.map(axisKey))
  const metaColumns = columns.filter(
    (column) => column !== rank && column !== title && !axisKeys.has(column.key)
  )
  return (
    <div
      className={styles.table}
      style={gridFor(axes, metaColumns)}
      data-shape={bannerShapeOf(items)}
    >
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
            <Link
              href={itemPath(item) + search}
              className={styles.thumb}
              data-shape={item.banner?.shape}
              tabIndex={-1}
              aria-hidden
            >
              {item.banner && <img src={item.banner.url} alt="" loading="lazy" />}
            </Link>
            <Link href={itemPath(item) + search} className={styles.title}>
              {item.title}
            </Link>
            {axes.map((axis) => (
              <AxisCell key={axis.key} axis={axis} item={item} />
            ))}
            <span className={styles.meta}>
              {metaColumns.map((column) => (
                <span key={column.key} className={styles.metaValue}>
                  {column.key === 'status' ? statusLabel(item.status) : shownValue(column, item)}
                </span>
              ))}
            </span>
          </li>
        ))}
      </ol>
    </div>
  )
}
