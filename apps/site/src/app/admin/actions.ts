'use server'

import { requireAdmin } from '../../lib/admin/require-admin'
import { logAdmin } from '../../lib/log/admin-api'
import type { EntryChange, NewMoment } from '../../lib/log/types'

export async function changeEntry(itemId: number, change: EntryChange) {
  await requireAdmin()
  await logAdmin.changeEntry(itemId, change)
}

export async function changeEntries(itemIds: number[], change: EntryChange) {
  await requireAdmin()
  await logAdmin.changeEntries(itemIds, change)
}

export async function addMoment(itemId: number, moment: NewMoment) {
  await requireAdmin()
  const item = await logAdmin.addMoment(itemId, moment)
  return item.moments ?? []
}

export async function deleteMoment(momentId: number) {
  await requireAdmin()
  await logAdmin.deleteMoment(momentId)
}

export async function startSteamSync() {
  await requireAdmin()
  return logAdmin.startSteamSync()
}

export async function steamSyncStatus() {
  await requireAdmin()
  return logAdmin.steamSync()
}
