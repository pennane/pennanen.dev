import { logPublic } from '../../../lib/log/public-api'
import { ogCard, ogSize } from './components/ogCard'
import { countsOf, logOverview, postersOf, roundRobin } from './components/overview'
import { logSummary } from './components/summaries'

export const size = ogSize
export const contentType = 'image/png'
export const alt = 'The log of things as ap saw them'

export default async function OpenGraphImage() {
  const overviews = await logOverview(await logPublic.presentation())
  return ogCard({
    title: 'things as ap saw them',
    blurb: logSummary(overviews),
    footer: countsOf(overviews).join(' · '),
    backdrop: { posters: roundRobin(overviews.map(({ items }) => postersOf(items))) }
  })
}
