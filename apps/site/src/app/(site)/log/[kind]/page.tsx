import type { Metadata } from 'next'
import Link from 'next/link'
import { notFound } from 'next/navigation'
import { Suspense } from 'react'
import { Gap, Stack } from '../../../../components/Stack'
import { logPublic } from '../../../../lib/log/public-api'
import { isPublicKind, kindTitle, publicKinds, toListed } from '../../../../lib/log/present'
import { baseUrl } from '../../../sitemap'
import { Prose } from '../components/Prose'
import { kindOverview } from '../components/overview'
import { kindSummary } from '../components/summaries'
import logStyles from '../components/log.module.css'
import { DefaultBoard, KindBrowser } from './KindBrowser'

type Params = Promise<{ kind: string }>

export const dynamicParams = false

export async function generateStaticParams() {
  return publicKinds(await logPublic.presentation()).map((kind) => ({ kind }))
}

export async function generateMetadata({ params }: { params: Params }): Promise<Metadata> {
  const { kind } = await params
  const presentation = await logPublic.presentation()
  if (!isPublicKind(presentation, kind)) return {}
  const overview = await kindOverview(presentation, kind)
  const title = `${overview.title} as ap saw them`
  const description = kindSummary(presentation, overview)
  return {
    metadataBase: new URL(baseUrl),
    title,
    description,
    openGraph: { title, description, type: 'website', url: `${baseUrl}/log/${kind}` },
    twitter: { card: 'summary_large_image', title, description }
  }
}

export default async function KindPage({ params }: { params: Params }) {
  const { kind } = await params
  const presentation = await logPublic.presentation()
  if (!isPublicKind(presentation, kind)) notFound()
  const [listed, { untriaged }] = await Promise.all([logPublic.items(kind), logPublic.kind(kind)])
  const title = kindTitle(presentation, kind)
  const description = presentation.kinds[kind]?.description.trim() ?? ''
  const items = listed.map((item) => toListed(item, presentation))

  return (
    <Stack gap={Gap.large} alignItems="stretch">
      <nav className={logStyles.crumbs}>
        <Link href="/log">log</Link>
        <span aria-hidden>/</span>
      </nav>
      <h1>{title}</h1>
      {description && <Prose markdown={description} />}
      {items.length === 0 ? (
        <p className={logStyles.empty}>Nothing here yet.</p>
      ) : (
        <Suspense fallback={<DefaultBoard items={items} kind={kind} />}>
          <KindBrowser items={items} kind={kind} />
        </Suspense>
      )}
      {untriaged > 0 && (
        <p className={logStyles.unsorted}>
          and {untriaged} {title} still unsorted
        </p>
      )}
    </Stack>
  )
}
