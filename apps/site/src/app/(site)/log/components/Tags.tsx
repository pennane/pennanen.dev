'use client'

import clsx from 'clsx'
import { isFavorite, tagLabel } from '../../../../lib/log/present'
import styles from './log.module.css'
import { useTagTitle } from './PresentationContext'

export const Tags = ({ tags }: { tags: string[] }) => {
  const titleOf = useTagTitle()
  if (tags.length === 0) return null
  return (
    <ul className={styles.tags}>
      {tags.map((tag) => (
        <li key={tag} className={clsx(styles.tag, isFavorite([tag]) && styles.favorite)}>
          {titleOf(tag) ?? tagLabel(tag)}
        </li>
      ))}
    </ul>
  )
}
