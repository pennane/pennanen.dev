import { isPublicKind } from '../../../../lib/log/present'
import { logPublic } from '../../../../lib/log/public-api'
import { ogCard, ogSize } from '../components/ogCard'
import { kindOverview, postersOf } from '../components/overview'
import { kindSummary } from '../components/summaries'

export const size = ogSize
export const contentType = 'image/png'
export const alt = 'One kind of thing in the log of things as ap saw them'

export default async function OpenGraphImage({ params }: { params: Promise<{ kind: string }> }) {
  const { kind } = await params
  const presentation = await logPublic.presentation()
  if (!isPublicKind(presentation, kind)) {
    return ogCard({
      title: 'things as ap saw them',
      blurb: undefined,
      footer: '',
      backdrop: { posters: [] }
    })
  }
  const overview = await kindOverview(presentation, kind)
  return ogCard({
    title: `${overview.title} as ap saw them`,
    blurb: kindSummary(presentation, overview),
    footer: `${overview.items.length} ${overview.title}`,
    backdrop: { posters: postersOf(overview.items) }
  })
}
