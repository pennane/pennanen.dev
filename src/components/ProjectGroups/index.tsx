import styles from './project-groups.module.css'
import { WrapInBox } from '../WrapInBox'
import { formatDate } from '../../app/lib'
import { ReactNode } from 'react'

const groupEntries = <T,>(
  items: T[],
  getDate: (x: T) => Parameters<typeof formatDate>[0]
) =>
  Object.entries(
    Object.groupBy(
      items,
      (item) =>
        formatDate(getDate(item))?.split(' ').at(-1) || 'Undated (2015-2018)'
    )
  ).sort((a, b) => Number(b[0]) - Number(a[0]))

export const ItemGroups = <T,>({
  items,
  render,
  getDate
}: {
  items: T[]
  getDate: (x: T) => Parameters<typeof formatDate>[0]
  render: (item: T) => ReactNode
}) => {
  return (
    <div className={styles.groups}>
      {groupEntries(items, getDate).map(([key, item]) => (
        <section className={styles.group} key={key}>
          <header className={styles.header}>
            <WrapInBox>
              <h3>{key}</h3>
            </WrapInBox>
          </header>

          <div className={styles.list}>{item!.map(render)}</div>
        </section>
      ))}
    </div>
  )
}
