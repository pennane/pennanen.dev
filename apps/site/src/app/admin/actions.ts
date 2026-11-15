'use server'

import { requireAdmin } from '../../lib/admin/require-admin'
import { LogApiError, logAdmin } from '../../lib/log/admin-api'
import { itemPath, textField } from '../../lib/log/present'
import { LETTERBOXD_EXPORT_FIELD } from '../../lib/log/syncs'
import type {
  Axis,
  EntryChange,
  FieldValue,
  InheritedValue,
  ItemLink,
  ItemView,
  KindSettings,
  TakeFieldVisibility,
  TMDBLink,
  TMDBType
} from '../../lib/log/types'
import { INHERITED_SOURCE_PREFIX, TMDB_ID, TMDB_TYPE_OF_SOURCE } from '../../lib/log/types'
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

const tmdbLinkOf = (item: ItemView): TMDBLink | undefined => {
  const tmdbId = item.verdicts.values[TMDB_ID]
  const tmdbType = TMDB_TYPE_OF_SOURCE[item.verdicts.sourceOf[TMDB_ID] ?? '']
  return typeof tmdbId === 'number' && tmdbType ? { tmdbType, tmdbId } : undefined
}

const inheritedRelationOf = (source: string) =>
  source.startsWith(INHERITED_SOURCE_PREFIX)
    ? source.slice(INHERITED_SOURCE_PREFIX.length)
    : undefined

const nearestRelated = (item: ItemView, relation: string) =>
  (item.relations ?? [])
    .filter((r) => r.fromId === item.id && r.type === relation)
    .map((r) => r.toId)
    .sort((a, b) => a - b)[0]

const shownValue = (value: unknown) => (Array.isArray(value) ? value.join(', ') : String(value))

const relatedTitle = async (itemId: number | undefined, relation: string) => {
  if (itemId === undefined) return `its ${relation}`
  const related = await logAdmin.item(itemId)
  return textField(related.verdicts.values, 'title') ?? related.slug
}

const inheritedOf = (item: ItemView): Promise<InheritedValue[]> =>
  Promise.all(
    Object.entries(item.verdicts.sourceOf).flatMap(([field, source]) => {
      const relation = inheritedRelationOf(source)
      if (!relation) return []
      return [
        relatedTitle(nearestRelated(item, relation), relation).then((from) => ({
          field,
          value: shownValue(item.verdicts.values[field]),
          from
        }))
      ]
    })
  )

const subjectOf = async (item: ItemView): Promise<EntrySubject> => {
  if (!item.entry) throw new Error(`item ${item.id} has no entry`)
  return {
    itemId: item.id,
    title: textField(item.verdicts.values, 'title') ?? item.slug,
    cover: textField(item.verdicts.values, 'cover') ?? textField(item.verdicts.values, 'image'),
    publicPath: `/log/${item.kind}/${item.slug}`,
    entry: item.entry,
    lastPlayed: lastPlayedOf(item),
    takes: item.takes,
    kind: item.kind,
    tmdb: tmdbLinkOf(item),
    inherited: await inheritedOf(item)
  }
}

export async function matchTMDB(itemId: number, link: TMDBLink) {
  await requireAdmin()
  return subjectOf(await logAdmin.matchTMDB(itemId, link))
}

export async function searchTMDB(tmdbType: TMDBType, query: string) {
  await requireAdmin()
  return logAdmin.searchTMDB(tmdbType, query)
}

export async function addFromTMDB(link: TMDBLink) {
  await requireAdmin()
  const item = await logAdmin.addFromTMDB(link)
  return item.id
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

export type Saved<T> = { saved: T } | { refused: string }

const refusedByTheLog = async <T>(save: () => Promise<T>): Promise<Saved<T>> => {
  try {
    return { saved: await save() }
  } catch (error) {
    if (error instanceof LogApiError) return { refused: error.message }
    throw error
  }
}

export async function saveAxes(axes: Axis[]) {
  await requireAdmin()
  return refusedByTheLog(() => logAdmin.putAxes(axes))
}

export async function saveKindSettings(kind: string, settings: KindSettings) {
  await requireAdmin()
  return refusedByTheLog(() => logAdmin.putKindSettings(kind, settings))
}

export async function saveTagTitles(titles: Record<string, string>) {
  await requireAdmin()
  return refusedByTheLog(() => logAdmin.putTagTitles(titles))
}

export async function saveKindOrder(order: string[]) {
  await requireAdmin()
  return refusedByTheLog(() => logAdmin.putKindOrder(order))
}
