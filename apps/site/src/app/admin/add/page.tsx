import { requireAdmin } from '../../../lib/admin/require-admin'
import { logAdmin } from '../../../lib/log/admin-api'
import { toRanked } from '../_queue/placing'
import { Add } from './Add'

export const dynamic = 'force-dynamic'

const ADDED_KIND = 'game'

export default async function AddPage() {
  await requireAdmin()
  const [kinds, entries] = await Promise.all([logAdmin.kinds(), logAdmin.entries(ADDED_KIND)])
  const kind = kinds.find((k) => k.name === ADDED_KIND)
  if (!kind) throw new Error(`kind ${ADDED_KIND} is missing from the log`)
  return <Add kind={kind} ranked={toRanked(kind, entries)} />
}
