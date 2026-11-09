import type { EntryChange, ItemView, Kind, SteamSyncReport } from './types'

class LogApiError extends Error {
  constructor(
    readonly status: number,
    message: string
  ) {
    super(message)
  }
}

const requiredEnv = (name: string) => {
  const value = process.env[name]
  if (!value) throw new Error(`${name} is not set`)
  return value
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
  items: (filter: Record<string, string>) =>
    call<ItemView[]>(`/admin/items?${new URLSearchParams(filter)}`),
  changeEntry: (itemId: number, change: EntryChange) =>
    call<ItemView>(`/admin/items/${itemId}/entry`, send('PATCH', change)),
  changeEntries: (itemIds: number[], change: EntryChange) =>
    call<void>('/admin/entries/change', send('POST', { itemIds, change })),
  steamSync: () => call<SteamSyncReport>('/admin/sync/steam'),
  startSteamSync: () =>
    call<SteamSyncReport>('/admin/sync/steam', { method: 'POST' })
}
