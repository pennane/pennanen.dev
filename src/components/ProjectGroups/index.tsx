import { groupEntries } from '../../app/[project]/lib'
import styles from './project-groups.module.css'
import { WrapInBox } from '../WrapInBox'
import { formatDate } from '../../app/lib'
import { ReactNode } from 'react'

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
