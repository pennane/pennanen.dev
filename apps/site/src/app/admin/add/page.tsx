import { requireAdmin } from '../../../lib/admin/require-admin'
import { logAdmin } from '../../../lib/log/admin-api'
import { Add } from './Add'

export const dynamic = 'force-dynamic'

const ADDED_KIND = 'game'

export default async function AddPage() {
  await requireAdmin()
  const kind = (await logAdmin.kinds()).find((k) => k.name === ADDED_KIND)
  if (!kind) throw new Error(`kind ${ADDED_KIND} is missing from the log`)
  return <Add kind={kind} />
}
