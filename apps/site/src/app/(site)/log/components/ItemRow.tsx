import Image from 'next/image'
import Link from 'next/link'
import {
  itemPath,
  statusLabel,
  type ListedItem
} from '../../../../lib/log/present'
import { Scores } from './Scores'
import { Tags } from './Tags'
import styles from './log.module.css'

const THUMB_WIDTH = 96
const THUMB_HEIGHT = 45

export const ItemRow = ({ item }: { item: ListedItem }) => (
  <li className={styles.row}>
    <Link href={itemPath(item)} className={styles.thumb} tabIndex={-1} aria-hidden>
      {item.picture && (
        <Image
          alt=""
          src={item.picture}
          width={THUMB_WIDTH}
          height={THUMB_HEIGHT}
          unoptimized
        />
      )}
    </Link>
    <div className={styles.rowBody}>
      <Link href={itemPath(item)}>
        <h4>{item.title}</h4>
      </Link>
      <p className={styles.rowMeta}>
        <span className={styles.status}>{statusLabel(item.status)}</span>
        <Scores kind={item.kind} scores={item.scores} />
      </p>
      <Tags tags={item.tags} />
    </div>
  </li>
)

export const ItemList = ({ items }: { items: ListedItem[] }) => (
  <ul className={styles.list}>
    {items.map((item) => (
      <ItemRow key={`${item.kind}/${item.slug}`} item={item} />
    ))}
  </ul>
)
