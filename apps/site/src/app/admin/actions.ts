'use server'

import { requireAdmin } from '../../lib/admin/require-admin'
import { logAdmin } from '../../lib/log/admin-api'
import type { EntryChange } from '../../lib/log/types'

export async function changeEntry(itemId: number, change: EntryChange) {
  await requireAdmin()
  await logAdmin.changeEntry(itemId, change)
}

export async function changeEntries(itemIds: number[], change: EntryChange) {
  await requireAdmin()
  await logAdmin.changeEntries(itemIds, change)
}

export async function startSteamSync() {
  await requireAdmin()
  return logAdmin.startSteamSync()
}

export async function steamSyncStatus() {
  await requireAdmin()
  return logAdmin.steamSync()
}
