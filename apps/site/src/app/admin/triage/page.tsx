import { requireAdmin } from '../../../lib/admin/require-admin'
import { logAdmin } from '../../../lib/log/admin-api'
import { byPlaytime, toGameCard } from '../_queue/toGameCard'
import { Triage } from './Triage'

export const dynamic = 'force-dynamic'

const TRIAGED_KIND = 'game'

export default async function TriagePage() {
  await requireAdmin()
  const [kinds, items, steamSync] = await Promise.all([
    logAdmin.kinds(),
    logAdmin.items({ kind: TRIAGED_KIND, status: 'untriaged' }),
    logAdmin.steamSync().catch(() => undefined)
  ])
  const kind = kinds.find((k) => k.name === TRIAGED_KIND)
  if (!kind) throw new Error(`kind ${TRIAGED_KIND} is missing from the log`)

  return <Triage kind={kind} cards={items.map(toGameCard).sort(byPlaytime)} steamSync={steamSync} />
}
