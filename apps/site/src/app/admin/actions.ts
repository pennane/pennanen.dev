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

export async function searchIGDB(query: string) {
  await requireAdmin()
  return logAdmin.searchIGDB(query)
}

export async function addFromIGDB(igdbId: number) {
  await requireAdmin()
  const item = await logAdmin.addFromIGDB(igdbId)
  return item.id
}

export async function addByTitle(kind: string, title: string) {
  await requireAdmin()
  const item = await logAdmin.createItem(kind, title)
  return item.id
}

export async function matchIGDB(itemId: number, igdbId: number) {
  await requireAdmin()
  await logAdmin.matchIGDB(itemId, igdbId)
}

export async function matchIGDBNone(itemId: number) {
  await requireAdmin()
  await logAdmin.matchIGDBNone(itemId)
}

export async function igdbMatchStatus() {
  await requireAdmin()
  return logAdmin.igdbMatch()
}

export async function startIGDBMatch() {
  await requireAdmin()
  return logAdmin.startIGDBMatch()
}

export async function startSteamSync() {
  await requireAdmin()
  return logAdmin.startSteamSync()
}

export async function steamSyncStatus() {
  await requireAdmin()
  return logAdmin.steamSync()
}
