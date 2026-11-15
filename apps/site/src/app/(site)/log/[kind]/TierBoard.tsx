'use client'

import Link from 'next/link'
import { useState } from 'react'
import { itemPath, type ListedItem } from '../../../../lib/log/present'
import type { Axis } from '../../../../lib/log/types'
import { tierColour } from './tierColours'
import styles from './tiers.module.css'
import { tiersOf } from './view'

const WITHOUT_HOVER = '(hover: none)'

type Caption = 'shown' | 'hidden'

const Tile = ({
  item,
  search,
  caption,
  onReveal
}: {
  item: ListedItem
  search: string
  caption: Caption
  onReveal: () => void
}) => {
  const picture = item.tile
  return (
    <Link
      href={itemPath(item) + search}
      className={styles.tile}
      data-shape={picture?.shape}
      data-caption={caption}
      aria-label={item.title}
      onClick={(event) => {
        if (caption === 'shown' || !window.matchMedia(WITHOUT_HOVER).matches) return
        event.preventDefault()
        onReveal()
      }}
    >
      {picture ? (
        <img src={picture.url} alt="" loading="lazy" />
      ) : (
        <span className={styles.titleTile}>{item.title}</span>
      )}
      <span className={styles.caption}>{item.title}</span>
    </Link>
  )
}

export const TierBoard = ({
  items,
  axis,
  search
}: {
  items: ListedItem[]
  axis: Axis
  search: string
}) => {
  const [revealed, setRevealed] = useState<string>()
  const tiers = tiersOf(items, axis)
    .map((tier, i) => ({ ...tier, colour: tierColour(i) }))
    .filter((tier) => tier.items.length > 0)

  if (tiers.length === 0)
    return <p className={styles.nothing}>Nothing scored on {axis.name} yet.</p>

  return (
    <div className={styles.board}>
      {tiers.map(({ division, colour, items }) => (
        <section key={division.key} className={styles.tier} aria-label={division.name}>
          <div className={styles.label} style={{ background: colour }}>
            {division.name}
          </div>
          <div className={styles.tiles}>
            {items.map((item) => (
              <Tile
                key={item.slug}
                item={item}
                search={search}
                caption={revealed === item.slug ? 'shown' : 'hidden'}
                onReveal={() => setRevealed(item.slug)}
              />
            ))}
          </div>
        </section>
      ))}
    </div>
  )
}
