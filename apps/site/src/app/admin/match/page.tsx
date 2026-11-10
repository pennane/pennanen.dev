import { requireAdmin } from '../../../lib/admin/require-admin'
import { logAdmin } from '../../../lib/log/admin-api'
import { byPlaytime, toGameCard } from '../_queue/toGameCard'
import { Match } from './Match'

export const dynamic = 'force-dynamic'

const MATCHED_KIND = 'game'

export default async function MatchPage() {
  await requireAdmin()
  const [items, report] = await Promise.all([
    logAdmin.items({ kind: MATCHED_KIND, needs: 'match' }),
    logAdmin.igdbMatch().catch(() => undefined)
  ])
  return <Match cards={items.map(toGameCard).sort(byPlaytime)} report={report} />
}
