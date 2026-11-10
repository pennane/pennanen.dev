import { Link2 } from 'lucide-react'
import type { Metadata } from 'next'
import Image from 'next/image'
import Link from 'next/link'
import { notFound } from 'next/navigation'
import { Lede, MetaLine, MetaLink } from '../../../../../components/MetaLine'
import { Gap, Stack } from '../../../../../components/Stack'
import { WrapInBox } from '../../../../../components/WrapInBox'
import {
  formatPartialDate,
  freeFormFields,
  isPublicKind,
  orderedScores,
  itemPath,
  kindPlural,
  pictureOf,
  statusLabel,
  textField,
  TIME_SPENT_LABELS,
  titleOf
} from '../../../../../lib/log/present'
import {
  logPublic,
  type PublicItem,
  type PublicMoment
} from '../../../../../lib/log/public-api'
import { baseUrl } from '../../../../sitemap'
import { Score, Scores } from '../../components/Scores'
import { Tags } from '../../components/Tags'
import logStyles from '../../components/log.module.css'
import styles from './item.module.css'

type Params = Promise<{ kind: string; slug: string }>

const findItem = async (params: Params) => {
  const { kind, slug } = await params
  if (!isPublicKind(kind)) return undefined
  return logPublic.item(kind, slug)
}

export async function generateStaticParams({
  params
}: {
  params: { kind: string }
}) {
  const items = await logPublic.items(params.kind)
  return items.map(({ slug }) => ({ slug }))
}

const summaryOf = (item: PublicItem) =>
  item.blurb || undefined

export async function generateMetadata({
  params
}: {
  params: Params
}): Promise<Metadata> {
  const item = await findItem(params)
  if (!item) return {}
  const title = titleOf(item)
  const description = summaryOf(item)
  return {
    metadataBase: new URL(baseUrl),
    title: `${title} · things as ap saw them`,
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
      {orderedScores(item.scores).map(({ axis, score }) => (
        <Score key={axis} kind={item.kind} axis={axis} score={score} />
      ))}
      {released && <span>{formatPartialDate(released)}</span>}
      {item.timeSpent && <span>{TIME_SPENT_LABELS[item.timeSpent]}</span>}
      {lastPlayed && <span>last played {formatPartialDate(lastPlayed)}</span>}
      {link && <MetaLink href={link} icon={Link2} />}
    </MetaLine>
  )
}

const byNewest = (a: PublicMoment, b: PublicMoment) => b.at.localeCompare(a.at)

const Moments = ({ kind, moments }: { kind: string; moments: PublicMoment[] }) => (
  <section>
    <Stack gap={Gap.large} alignItems="stretch">
      <header>
        <WrapInBox>
          <h3>moments</h3>
        </WrapInBox>
      </header>
      <ol className={styles.moments}>
        {moments.toSorted(byNewest).map((moment) => (
          <li key={`${moment.at}${moment.note}`} className={styles.moment}>
            <time dateTime={moment.at}>{formatPartialDate(moment.at)}</time>
            <div className={styles.momentBody}>
              {moment.note && <p>{moment.note}</p>}
              <Scores kind={kind} scores={moment.scores} />
              {moment.refs.length > 0 && (
                <p className={styles.refs}>
                  {moment.refs.map((ref) => (
                    <Link key={itemPath(ref)} href={itemPath(ref)}>
                      {ref.title}
                    </Link>
                  ))}
                </p>
              )}
            </div>
          </li>
        ))}
      </ol>
    </Stack>
  </section>
)

export default async function ItemPage({ params }: { params: Params }) {
  const item = await findItem(params)
  if (!item || !isPublicKind(item.kind)) notFound()
  const picture = pictureOf(item)
  const extras = freeFormFields(item.fields)

  return (
    <article>
      <header>
        <Stack gap={Gap.large} alignItems="stretch">
          <nav className={logStyles.crumbs}>
            <Link href="/log">log</Link>
            <span aria-hidden>/</span>
            <Link href={`/log/${item.kind}`}>{kindPlural(item.kind)}</Link>
            <span aria-hidden>/</span>
          </nav>
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
      {item.moments.length > 0 && <Moments kind={item.kind} moments={item.moments} />}
    </article>
  )
}
