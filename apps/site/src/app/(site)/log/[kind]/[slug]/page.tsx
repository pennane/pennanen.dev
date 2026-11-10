import { Link2 } from 'lucide-react'
import type { Metadata } from 'next'
import Image from 'next/image'
import Link from 'next/link'
import { notFound } from 'next/navigation'
import { Lede, MetaLine, MetaLink } from '../../../../../components/MetaLine'
import { Gap, Stack } from '../../../../../components/Stack'
import { WrapInBox } from '../../../../../components/WrapInBox'
import {
  byRank,
  formatPartialDate,
  freeFormFields,
  isPublicKind,
  itemPath,
  kindTitle,
  pictureOf,
  statusLabel,
  textField,
  TIME_SPENT_LABELS,
  titleOf,
  toListed
} from '../../../../../lib/log/present'
import { logPublic, type PublicItem } from '../../../../../lib/log/public-api'
import { baseUrl } from '../../../../sitemap'
import { Prose } from '../../components/Prose'
import { ScoreParts } from '../../components/Scores'
import { Tags } from '../../components/Tags'
import logStyles from '../../components/log.module.css'
import styles from './item.module.css'

type Params = Promise<{ kind: string; slug: string }>

const findItem = async (params: Params) => {
  const { kind, slug } = await params
  if (!isPublicKind(kind)) return undefined
  return logPublic.item(kind, slug)
}

export async function generateStaticParams({ params }: { params: { kind: string } }) {
  const items = await logPublic.items(params.kind)
  return items.map(({ slug }) => ({ slug }))
}

const summaryOf = (item: PublicItem) => item.blurb || undefined

export async function generateMetadata({ params }: { params: Params }): Promise<Metadata> {
  const item = await findItem(params)
  if (!item) return {}
  const title = titleOf(item)
  const description = summaryOf(item)
  return {
    metadataBase: new URL(baseUrl),
    title: `${title} as ap saw it`,
    description,
    openGraph: {
      title,
      description,
      type: 'article',
      url: `${baseUrl}${itemPath(item)}`
    },
    twitter: { card: 'summary_large_image', title, description }
  }
}

const PICTURE_WIDTH = 460
const PICTURE_HEIGHT = 215

const Facts = ({ item }: { item: PublicItem }) => {
  const released = textField(item.fields, 'released')
  const lastPlayed = textField(item.fields, 'last_played')
  const link = textField(item.fields, 'link')
  return (
    <MetaLine>
      <span>{statusLabel(item.status)}</span>
      <ScoreParts kind={item.kind} scores={item.scores} />
      {released && <span>{formatPartialDate(released)}</span>}
      {item.timeSpent && <span>{TIME_SPENT_LABELS[item.timeSpent]}</span>}
      {lastPlayed && <span>last played {formatPartialDate(lastPlayed)}</span>}
      {link && <MetaLink href={link} icon={Link2} />}
    </MetaLine>
  )
}

const Visions = ({ prose }: { prose: string }) => (
  <section>
    <Stack gap={Gap.large} alignItems="stretch">
      <header>
        <WrapInBox>
          <h3>visions</h3>
        </WrapInBox>
      </header>
      <Prose markdown={prose} />
    </Stack>
  </section>
)

type Neighbour = { title: string; href: string }

const neighboursOf = async (item: PublicItem) => {
  const ordered = (await logPublic.items(item.kind)).map(toListed).toSorted(byRank)
  const at = ordered.findIndex((listed) => listed.slug === item.slug)
  const linkTo = (index: number): Neighbour | undefined => {
    const listed = at >= 0 ? ordered[index] : undefined
    return listed && { title: listed.title, href: itemPath(listed) }
  }
  return { previous: linkTo(at - 1), next: linkTo(at + 1) }
}

const NEIGHBOURS_PLACES = { top: styles.neighboursTop, bottom: styles.neighboursBottom }

const Neighbours = ({
  previous,
  next,
  place
}: {
  previous?: Neighbour
  next?: Neighbour
  place: keyof typeof NEIGHBOURS_PLACES
}) => (
  <nav className={`${styles.neighbours} ${NEIGHBOURS_PLACES[place]}`}>
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

export default async function ItemPage({ params }: { params: Params }) {
  const item = await findItem(params)
  if (!item || !isPublicKind(item.kind)) notFound()
  const picture = pictureOf(item)
  const extras = freeFormFields(item.fields)
  const [neighbours, presentation] = await Promise.all([
    neighboursOf(item),
    logPublic.presentation()
  ])

  return (
    <article>
      <header>
        <Stack gap={Gap.large} alignItems="stretch">
          <nav className={logStyles.crumbs}>
            <Link href="/log">log</Link>
            <span aria-hidden>/</span>
            <Link href={`/log/${item.kind}`}>{kindTitle(presentation, item.kind)}</Link>
            <span aria-hidden>/</span>
          </nav>
          <Neighbours {...neighbours} place="top" />
          {picture && (
            <Image
              className={styles.picture}
              alt=""
              src={picture}
              width={PICTURE_WIDTH}
              height={PICTURE_HEIGHT}
              unoptimized
              priority
            />
          )}
          <Stack>
            <h1>{titleOf(item)}</h1>
            {item.blurb && <Lede>{item.blurb}</Lede>}
            <Facts item={item} />
          </Stack>
        </Stack>
      </header>
      <Tags tags={item.tags} />
      {extras.length > 0 && (
        <dl className={styles.fields}>
          {extras.map(({ name, value }) => (
            <div key={name}>
              <dt>{name}</dt>
              <dd>{value}</dd>
            </div>
          ))}
        </dl>
      )}
      {item.prose.trim() && <Visions prose={item.prose} />}
      <Neighbours {...neighbours} place="bottom" />
    </article>
  )
}
