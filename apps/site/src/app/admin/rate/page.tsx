import { requireAdmin } from '../../../lib/admin/require-admin'
import { logAdmin } from '../../../lib/log/admin-api'
import { toRanked } from '../_queue/placing'
import { byPlaytime, toGameCard } from '../_queue/toGameCard'
import { Rate } from './Rate'

export const dynamic = 'force-dynamic'

const RATED_KIND = 'game'

export default async function RatePage() {
  await requireAdmin()
  const [kinds, items, entries] = await Promise.all([
    logAdmin.kinds(),
    logAdmin.items({ kind: RATED_KIND, needs: 'score' }),
    logAdmin.entries(RATED_KIND)
  ])
  const kind = kinds.find((k) => k.name === RATED_KIND)
  if (!kind) throw new Error(`kind ${RATED_KIND} is missing from the log`)
  return (
    <Rate
      kind={kind}
      cards={items.map(toGameCard).sort(byPlaytime)}
      ranked={toRanked(kind, entries)}
    />
  )
}
