import { requireAdmin } from '../../../lib/admin/require-admin'
import { logAdmin } from '../../../lib/log/admin-api'
import { byPlaytime, toGameCard } from '../_queue/toGameCard'
import { Rate } from './Rate'

export const dynamic = 'force-dynamic'

const RATED_KIND = 'game'

export default async function RatePage() {
  await requireAdmin()
  const items = await logAdmin.items({ kind: RATED_KIND, needs: 'score' })
  return <Rate cards={items.map(toGameCard).sort(byPlaytime)} />
}
