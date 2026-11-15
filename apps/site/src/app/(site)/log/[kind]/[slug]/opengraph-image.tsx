import { bannerOf, isPublicKind, shownScores, titleOf } from '../../../../../lib/log/present'
import { logPublic } from '../../../../../lib/log/public-api'
import { ogCard, ogSize } from '../../components/ogCard'

export const size = ogSize
export const contentType = 'image/png'
export const alt = 'An item in the log of things as ap saw them'

export default async function OpenGraphImage({
  params
}: {
  params: Promise<{ kind: string; slug: string }>
}) {
  const { kind, slug } = await params
  const presentation = await logPublic.presentation()
  const item = isPublicKind(presentation, kind) ? await logPublic.item(kind, slug) : undefined
  return ogCard({
    title: item ? titleOf(item) : 'things as ap saw them',
    blurb: item?.blurb,
    footer: item
      ? shownScores(presentation, kind, item.scores)
          .map(({ axisName, division }) => `${axisName} ${division}`)
          .join(' · ')
      : '',
    backdrop: { picture: item && bannerOf(presentation, item)?.url }
  })
}
