'use client'

import clsx from 'clsx'
import Link from 'next/link'
import { useSearchParams } from 'next/navigation'
import {
  statusLabel,
  tagLabel,
  type ListedItem
} from '../../../../lib/log/present'
import { ItemList } from '../components/ItemRow'
import { useTagTitle } from '../components/PresentationContext'
import styles from './kind.module.css'
import {
  applyView,
  hrefFor,
  parseView,
  SORT_LABELS,
  SORTS,
  statusesIn,
  tagsIn,
  type View
} from './view'

type Choice = { label: string; count?: number; view: View }

const ChoiceRow = ({
  name,
  path,
  current,
  choices
}: {
  name: string
  path: string
  current: View
  choices: Choice[]
}) => {
  const currentHref = hrefFor(path, current)
  return (
    <div className={styles.choiceRow}>
      <span className={styles.choiceName}>{name}</span>
      <ul className={styles.choices}>
        {choices.map(({ label, count, view }) => {
          const href = hrefFor(path, view)
          return (
            <li key={label}>
              <Link
                href={href}
                replace
                scroll={false}
                className={clsx(
                  styles.choice,
                  href === currentHref && styles.chosen
                )}
                aria-current={href === currentHref ? 'true' : undefined}
              >
                {label}
                {count !== undefined && (
                  <span className={styles.count}>{count}</span>
                )}
              </Link>
            </li>
          )
        })}
      </ul>
    </div>
  )
}

const MIN_TO_BE_WORTH_FILTERING = 2

export const KindBrowser = ({
  items,
  path
}: {
  items: ListedItem[]
  path: string
}) => {
  const view = parseView(useSearchParams())
  const titleOf = useTagTitle()
  const statuses = statusesIn(items)
  const tags = tagsIn(items)
  const shown = applyView(items, view)

  return (
    <>
      <div className={styles.controls}>
        <ChoiceRow
          name="sort"
          path={path}
          current={view}
          choices={SORTS.map((sort) => ({
            label: SORT_LABELS[sort],
            view: { ...view, sort }
          }))}
        />
        {statuses.length >= MIN_TO_BE_WORTH_FILTERING && (
          <ChoiceRow
            name="status"
            path={path}
            current={view}
            choices={[
              { label: 'any', view: { ...view, status: undefined } },
              ...statuses.map(({ value, count }) => ({
                label: statusLabel(value),
                count,
                view: { ...view, status: value }
              }))
            ]}
          />
        )}
        {tags.length > 0 && (
          <ChoiceRow
            name="tag"
            path={path}
            current={view}
            choices={[
              { label: 'any', view: { ...view, tag: undefined } },
              ...tags.map(({ value, count }) => ({
                label: titleOf(value) ?? tagLabel(value),
                count,
                view: { ...view, tag: value }
              }))
            ]}
          />
        )}
      </div>
      {shown.length > 0 ? (
        <ItemList items={shown} />
      ) : (
        <p className={styles.empty}>Nothing matches.</p>
      )}
    </>
  )
}
