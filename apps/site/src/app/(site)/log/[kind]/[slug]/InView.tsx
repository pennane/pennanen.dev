'use client'

import Link from 'next/link'
import { useSearchParams } from 'next/navigation'
import { Suspense, type ReactNode } from 'react'
import { itemPath, type ListedItem } from '../../../../../lib/log/present'
import { useSetup } from '../setup'
import { orderIn, parseView, searchFor } from '../view'
import styles from './item.module.css'

type ReadableParams = { get: (name: string) => string | null }

type Place = { kind: string; slug: string; items: ListedItem[] }

const NO_PARAMS = new URLSearchParams()

const usePlaceInView = ({ kind, slug, items }: Place, params: ReadableParams) => {
  const setup = useSetup(items, kind)
  const view = parseView(params, setup.defaults)
  const search = searchFor(view, setup.defaults)
  const ordered = orderIn(items, view, setup)
  const at = ordered.findIndex((item) => item.slug === slug)
  const linkTo = (index: number) => {
    const item = at >= 0 ? ordered[index] : undefined
    return item && { title: item.title, href: itemPath(item) + search }
  }
  return { search, previous: linkTo(at - 1), next: linkTo(at + 1) }
}

const withParams = (render: (params: ReadableParams) => ReactNode) => {
  const FromURL = () => render(useSearchParams())
  return (
    <Suspense fallback={render(NO_PARAMS)}>
      <FromURL />
    </Suspense>
  )
}

const KindLinkIn = ({
  place,
  title,
  params
}: {
  place: Place
  title: string
  params: ReadableParams
}) => {
  const { search } = usePlaceInView(place, params)
  return <Link href={`/log/${place.kind}${search}`}>{title}</Link>
}

export const KindLink = ({ place, title }: { place: Place; title: string }) =>
  withParams((params) => <KindLinkIn place={place} title={title} params={params} />)

const NEIGHBOURS_PLACES = { top: styles.neighboursTop, bottom: styles.neighboursBottom }

type Spot = keyof typeof NEIGHBOURS_PLACES

const NeighboursIn = ({
  place,
  spot,
  params
}: {
  place: Place
  spot: Spot
  params: ReadableParams
}) => {
  const { previous, next } = usePlaceInView(place, params)
  return (
    <nav className={`${styles.neighbours} ${NEIGHBOURS_PLACES[spot]}`}>
      {previous && (
        <Link href={previous.href} rel="prev" className={styles.previous}>
          ← {previous.title}
        </Link>
      )}
      {next && (
        <Link href={next.href} rel="next" className={styles.next}>
          {next.title} →
        </Link>
      )}
    </nav>
  )
}

export const Neighbours = ({ place, spot }: { place: Place; spot: Spot }) =>
  withParams((params) => <NeighboursIn place={place} spot={spot} params={params} />)
