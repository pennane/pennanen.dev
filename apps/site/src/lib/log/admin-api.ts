import { requiredEnv } from './env'
import type {
  EntryChange,
  IGDBResult,
  ItemView,
  ListedEntry,
  Kind,
  Axis,
  KindSettings,
  Presentation,
  SyncReport,
  TakeFieldVisibility,
  UploadTicket
} from './types'

class LogApiError extends Error {
  constructor(
    readonly status: number,
    message: string
  ) {
    super(message)
  }
}

const call = async <T>(path: string, init: RequestInit = {}): Promise<T> => {
  const response = await fetch(requiredEnv('LOG_API_URL') + path, {
    ...init,
    cache: 'no-store',
    headers: {
      Authorization: `Bearer ${requiredEnv('LOG_ADMIN_TOKEN')}`,
      'Content-Type': 'application/json',
      ...init.headers
    }
  })
  if (!response.ok) {
    const body = (await response.json().catch(() => ({}))) as { error?: string }
    throw new LogApiError(response.status, body.error ?? response.statusText)
  }
  if (response.status === 204) return undefined as T
  return (await response.json()) as T
}

const send = (method: string, body: unknown): RequestInit => ({
  method,
  body: JSON.stringify(body)
})

export const logAdmin = {
  kinds: () => call<Kind[]>('/admin/kinds'),
  settings: () => call<Presentation>('/admin/settings'),
  putAxes: (axes: Axis[]) => call<Axis[]>('/admin/settings/axes', send('PUT', axes)),
  putKindSettings: (kind: string, settings: KindSettings) =>
    call<KindSettings>(`/admin/settings/kinds/${encodeURIComponent(kind)}`, send('PUT', settings)),
  putTagTitles: (titles: Record<string, string>) =>
    call<Record<string, string>>('/admin/settings/tag-titles', send('PUT', titles)),
  items: (filter: Record<string, string>) =>
    call<ItemView[]>(`/admin/items?${new URLSearchParams(filter)}`),
  entries: (kind: string) => call<ListedEntry[]>(`/admin/entries?${new URLSearchParams({ kind })}`),
  place: (itemId: number, division: string, index: number) =>
    call<void>(`/admin/items/${itemId}/place`, send('PUT', { division, index })),
  createItem: (kind: string, title: string) =>
    call<ItemView>('/admin/items', send('POST', { kind, title })),
  searchIGDB: (query: string) =>
    call<IGDBResult[]>(`/admin/igdb/search?${new URLSearchParams({ q: query })}`),
  addFromIGDB: (igdbId: number) =>
    call<ItemView>('/admin/items/from-igdb', send('POST', { igdbId })),
  matchIGDB: (itemId: number, igdbId: number) =>
    call<ItemView>(`/admin/items/${itemId}/igdb`, send('PUT', { igdbId })),
  matchTMDB: (itemId: number, tmdbId: number) =>
    call<ItemView>(`/admin/items/${itemId}/tmdb`, send('PUT', { tmdbId })),
  matchTMDBNone: (itemId: number) =>
    call<ItemView>(`/admin/items/${itemId}/tmdb/none`, { method: 'PUT' }),
  matchIGDBNone: (itemId: number) =>
    call<ItemView>(`/admin/items/${itemId}/igdb/none`, { method: 'PUT' }),
  item: (itemId: number) => call<ItemView>(`/admin/items/${itemId}`),
  uploadTicket: async (): Promise<UploadTicket> => {
    const { ticket, path } = await call<{ ticket: string; path: string }>('/admin/images/ticket', {
      method: 'POST'
    })
    return { ticket, url: requiredEnv('LOG_API_URL') + path }
  },
  changeEntry: (itemId: number, change: EntryChange) =>
    call<ItemView>(`/admin/items/${itemId}/entry`, send('PATCH', change)),
  setTakeFieldVisibility: (
    itemId: number,
    source: string,
    field: string,
    visibility: TakeFieldVisibility
  ) =>
    call<ItemView>(
      `/admin/items/${itemId}/takes/${encodeURIComponent(source)}/${encodeURIComponent(field)}`,
      send('PUT', { visibility })
    ),
  changeFields: (itemId: number, changes: Record<string, unknown>) =>
    call<ItemView>(`/admin/items/${itemId}/fields`, send('PATCH', changes)),
  changeEntries: (itemIds: number[], change: EntryChange) =>
    call<void>('/admin/entries/change', send('POST', { itemIds, change })),
  syncReports: () => call<SyncReport[]>('/admin/sync'),
  syncReport: (name: string) => call<SyncReport>(`/admin/sync/${name}`),
  startSync: (name: string) => call<SyncReport>(`/admin/sync/${name}`, { method: 'POST' }),
  importLetterboxdExport: (archive: ArrayBuffer) =>
    call<SyncReport>('/admin/letterboxd/export', {
      method: 'POST',
      body: archive,
      headers: { 'Content-Type': 'application/zip' }
    })
}
