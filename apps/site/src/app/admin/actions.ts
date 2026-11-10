'use server'

import { requireAdmin } from '../../lib/admin/require-admin'
import { logAdmin } from '../../lib/log/admin-api'
import { textField } from '../../lib/log/present'
import type { Axis, EntryChange, KindSettings, NewMoment } from '../../lib/log/types'
import type { EntrySubject } from './entries/Editor'

export async function changeEntry(itemId: number, change: EntryChange) {
  await requireAdmin()
  await logAdmin.changeEntry(itemId, change)
}

export async function openEntry(itemId: number): Promise<EntrySubject> {
  await requireAdmin()
  const item = await logAdmin.item(itemId)
  if (!item.entry) throw new Error(`item ${itemId} has no entry`)
  return {
    itemId: item.id,
    title: textField(item.verdicts.values, 'title') ?? item.slug,
    cover: textField(item.verdicts.values, 'cover') ?? textField(item.verdicts.values, 'image'),
    publicPath: `/log/${item.kind}/${item.slug}`,
    entry: item.entry,
    moments: item.moments ?? []
  }
}

export async function placeEntry(itemId: number, division: string, index: number) {
  await requireAdmin()
  await logAdmin.place(itemId, division, index)
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

export async function saveAxes(axes: Axis[]) {
  await requireAdmin()
  return logAdmin.putAxes(axes)
}

export async function saveKindSettings(kind: string, settings: KindSettings) {
  await requireAdmin()
  return logAdmin.putKindSettings(kind, settings)
}

export async function saveTagTitles(titles: Record<string, string>) {
  await requireAdmin()
  return logAdmin.putTagTitles(titles)
}
