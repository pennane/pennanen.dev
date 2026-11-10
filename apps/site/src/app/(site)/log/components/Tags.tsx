import clsx from 'clsx'
import { isFavorite, tagLabel } from '../../../../lib/log/present'
import styles from './log.module.css'

export const Tags = ({ tags }: { tags: string[] }) => {
  if (tags.length === 0) return null
  return (
    <ul className={styles.tags}>
      {tags.map((tag) => (
        <li
          key={tag}
          className={clsx(styles.tag, isFavorite([tag]) && styles.favorite)}
        >
          {tagLabel(tag)}
        </li>
      ))}
    </ul>
  )
}
