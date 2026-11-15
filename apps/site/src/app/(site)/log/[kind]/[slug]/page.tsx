import { Link2 } from 'lucide-react'
import type { Metadata } from 'next'
import Image from 'next/image'
import Link from 'next/link'
import { notFound } from 'next/navigation'
import { Lede, MetaLine, MetaLink } from '../../../../../components/MetaLine'
import { Gap, Stack } from '../../../../../components/Stack'
import { WrapInBox } from '../../../../../components/WrapInBox'
import {
  isPublicKind,
  itemPath,
  kindTitle,
  pageFieldsOf,
  pictureOf,
  shownTake,
  statusLabel,
  textField,
  TIME_SPENT_LABELS,
  titleOf,
  toListed
} from '../../../../../lib/log/present'
import { logPublic, type PublicItem, type PublicParts } from '../../../../../lib/log/public-api'
import { baseUrl } from '../../../../sitemap'
import { Prose } from '../../components/Prose'
import { ScoreParts } from '../../components/Scores'
import { Tags } from '../../components/Tags'
import logStyles from '../../components/log.module.css'
import { KindLink, Neighbours } from './InView'
import styles from './item.module.css'

type Params = Promise<{ kind: string; slug: string }>

const findItem = async (params: Params) => {
  const { kind, slug } = await params
  if (!isPublicKind(await logPublic.presentation(), kind)) return undefined
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
  const link = textField(item.fields, 'link')
  return (
    <MetaLine>
      <span>{statusLabel(item.status)}</span>
      <ScoreParts kind={item.kind} scores={item.scores} />
      {item.timeSpent && <span>{TIME_SPENT_LABELS[item.timeSpent]}</span>}
      {link && <MetaLink href={link} icon={Link2} />}
    </MetaLine>
  )
}

const PARAGRAPH_BREAK = /\n\s*\n/

const OutsideTakes = ({ item }: { item: PublicItem }) => {
  const takes = item.takes.map(shownTake).filter((take) => take.marks.length > 0 || take.text)
  if (takes.length === 0) return null
  return (
    <>
      {takes.map((take) => (
        <section key={take.source} className={styles.take}>
          <Stack gap={Gap.medium} alignItems="stretch">
            <header>
              <WrapInBox>
                <h3>{take.source}</h3>
              </WrapInBox>
            </header>
            {take.marks.length > 0 && <p className={styles.takeMarks}>{take.marks.join(' ')}</p>}
            {take.text?.split(PARAGRAPH_BREAK).map((paragraph, i) => (
              <p key={i}>{paragraph}</p>
            ))}
          </Stack>
        </section>
      ))}
    </>
  )
}

const SECONDS_PER_MINUTE = 60

const lengthOf = (fields: Record<string, unknown>) => {
  const seconds = fields.duration
  if (typeof seconds !== 'number') return undefined
  const minutes = Math.floor(seconds / SECONDS_PER_MINUTE)
  return `${minutes}:${String(seconds % SECONDS_PER_MINUTE).padStart(2, '0')}`
}

const Parts = ({ parts }: { parts: PublicParts }) => (
  <section>
    <Stack gap={Gap.large} alignItems="stretch">
      <header>
        <WrapInBox>
          <h3>{parts.title}</h3>
        </WrapInBox>
      </header>
      <ol className={styles.parts}>
        {parts.items.map((part, i) => (
          <li key={i}>
            <span>{textField(part.fields, 'title')}</span>
            <MetaLine>
              <ScoreParts kind={part.kind} scores={part.scores} />
              {lengthOf(part.fields) && <span>{lengthOf(part.fields)}</span>}
            </MetaLine>
          </li>
        ))}
      </ol>
    </Stack>
  </section>
)

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

export default async function ItemPage({ params }: { params: Params }) {
  const item = await findItem(params)
  if (!item) notFound()
  const picture = pictureOf(item)
  const [listed, presentation] = await Promise.all([
    logPublic.items(item.kind),
    logPublic.presentation()
  ])
  const place = {
    kind: item.kind,
    slug: item.slug,
    items: listed.map((other) => toListed(other, presentation.fields[item.kind] ?? []))
  }
  const extras = pageFieldsOf(item, presentation)

  return (
    <article>
      <header>
        <Stack gap={Gap.large} alignItems="stretch">
          <nav className={logStyles.crumbs}>
            <Link href="/log">log</Link>
            <span aria-hidden>/</span>
            <KindLink place={place} title={kindTitle(presentation, item.kind)} />
            <span aria-hidden>/</span>
          </nav>
          <Neighbours place={place} spot="top" />
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
          {extras.map(({ key, name, value }) => (
            <div key={key}>
              <dt>{name}</dt>
              <dd>{value}</dd>
            </div>
          ))}
        </dl>
      )}
      {item.parts && <Parts parts={item.parts} />}
      <OutsideTakes item={item} />
      {item.prose.trim() && <Visions prose={item.prose} />}
      <Neighbours place={place} spot="bottom" />
    </article>
  )
}
