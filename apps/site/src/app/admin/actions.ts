'use server'

import { requireAdmin } from '../../lib/admin/require-admin'
import { logAdmin } from '../../lib/log/admin-api'
import { itemPath, textField } from '../../lib/log/present'
import { LETTERBOXD_EXPORT_FIELD } from '../../lib/log/syncs'
import type {
  Axis,
  EntryChange,
  FieldValue,
  ItemLink,
  ItemView,
  KindSettings,
  TakeFieldVisibility
} from '../../lib/log/types'
import { TMDB_ID } from '../../lib/log/types'
import type { EntrySubject } from './entries/Editor'

const LAST_PLAYED = 'last_played'
export async function changeEntry(itemId: number, change: EntryChange) {
  await requireAdmin()
  await logAdmin.changeEntry(itemId, change)
}

export async function openEntry(itemId: number): Promise<EntrySubject> {
  await requireAdmin()
  return subjectOf(await logAdmin.item(itemId))
}

const subjectOf = (item: ItemView): EntrySubject => {
  if (!item.entry) throw new Error(`item ${item.id} has no entry`)
  const tmdbId = item.verdicts.values[TMDB_ID]
  return {
    itemId: item.id,
    title: textField(item.verdicts.values, 'title') ?? item.slug,
    cover: textField(item.verdicts.values, 'cover') ?? textField(item.verdicts.values, 'image'),
    publicPath: `/log/${item.kind}/${item.slug}`,
    entry: item.entry,
    lastPlayed: lastPlayedOf(item),
    takes: item.takes,
    tmdbId: typeof tmdbId === 'number' ? tmdbId : undefined
  }
}

export async function matchTMDB(itemId: number, tmdbId: number) {
  await requireAdmin()
  return subjectOf(await logAdmin.matchTMDB(itemId, tmdbId))
}

export async function matchTMDBNone(itemId: number) {
  await requireAdmin()
  return subjectOf(await logAdmin.matchTMDBNone(itemId))
}

export async function setTakeFieldVisibility(
  itemId: number,
  source: string,
  field: string,
  visibility: TakeFieldVisibility
) {
  await requireAdmin()
  return (await logAdmin.setTakeFieldVisibility(itemId, source, field, visibility)).takes
}

const lastPlayedOf = (item: ItemView): FieldValue => ({
  value: textField(item.verdicts.values, LAST_PLAYED),
  source: item.verdicts.sourceOf[LAST_PLAYED]
})

export async function changeLastPlayed(itemId: number, date: string | null) {
  await requireAdmin()
  return lastPlayedOf(await logAdmin.changeFields(itemId, { [LAST_PLAYED]: date }))
}

export async function placeEntry(itemId: number, division: string, index: number) {
  await requireAdmin()
  await logAdmin.place(itemId, division, index)
}

export async function changeEntries(itemIds: number[], change: EntryChange) {
  await requireAdmin()
  await logAdmin.changeEntries(itemIds, change)
}

export async function imageUploadTicket() {
  await requireAdmin()
  return logAdmin.uploadTicket()
}

const ITEM_LINKS_SHOWN = 8

export async function findItemLinks(query: string): Promise<ItemLink[]> {
  await requireAdmin()
  const items = await logAdmin.items({ q: query })
  return items
    .filter((item) => item.entry)
    .slice(0, ITEM_LINKS_SHOWN)
    .map((item) => ({
      title: textField(item.verdicts.values, 'title') ?? item.slug,
      href: itemPath(item)
    }))
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

export async function syncStatus(name: string) {
  await requireAdmin()
  return logAdmin.syncReport(name)
}

export async function startSync(name: string) {
  await requireAdmin()
  return logAdmin.startSync(name)
}

export async function importLetterboxdExport(form: FormData) {
  await requireAdmin()
  const archive = form.get(LETTERBOXD_EXPORT_FIELD)
  if (!(archive instanceof File)) throw new Error('no export file was sent')
  return logAdmin.importLetterboxdExport(await archive.arrayBuffer())
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
