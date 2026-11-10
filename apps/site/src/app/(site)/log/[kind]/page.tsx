import type { Metadata } from 'next'
import Link from 'next/link'
import { notFound } from 'next/navigation'
import { Suspense } from 'react'
import { Gap, Stack } from '../../../../components/Stack'
import { logPublic } from '../../../../lib/log/public-api'
import { isPublicKind, kindPlural, PUBLIC_KINDS, toListed } from '../../../../lib/log/present'
import logStyles from '../components/log.module.css'
import { DefaultBoard, KindBrowser } from './KindBrowser'

type Params = Promise<{ kind: string }>

export const dynamicParams = false

export function generateStaticParams() {
  return PUBLIC_KINDS.map((kind) => ({ kind }))
}

export async function generateMetadata({ params }: { params: Params }): Promise<Metadata> {
  const { kind } = await params
  if (!isPublicKind(kind)) return {}
  return {
    title: `${kindPlural(kind)} as ap saw them`,
    description: `The ${kindPlural(kind)} in Arttu Pennanen's log.`
  }
}

export default async function KindPage({ params }: { params: Params }) {
  const { kind } = await params
  if (!isPublicKind(kind)) notFound()
  const [listed, { untriaged }] = await Promise.all([logPublic.items(kind), logPublic.kind(kind)])
  const items = listed.map(toListed)

  return (
    <Stack gap={Gap.large} alignItems="stretch">
      <nav className={logStyles.crumbs}>
        <Link href="/log">log</Link>
        <span aria-hidden>/</span>
      </nav>
      <h1>{kindPlural(kind)}</h1>
      {items.length === 0 ? (
        <p className={logStyles.empty}>Nothing here yet.</p>
      ) : (
        <Suspense fallback={<DefaultBoard items={items} kind={kind} />}>
          <KindBrowser items={items} kind={kind} />
        </Suspense>
      )}
      {untriaged > 0 && (
        <p className={logStyles.unsorted}>
          and {untriaged} {kindPlural(kind)} still unsorted
        </p>
      )}
    </Stack>
  )
}
