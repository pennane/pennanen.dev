import { requireAdmin } from '../../../lib/admin/require-admin'
import { logAdmin } from '../../../lib/log/admin-api'
import type { ItemView } from '../../../lib/log/types'
import { Triage, type TriageCard } from './Triage'

export const dynamic = 'force-dynamic'

const TRIAGED_KIND = 'game'

const numberField = (item: ItemView, name: string) => {
  const value = item.verdicts.values[name]
  return typeof value === 'number' ? value : undefined
}

const stringField = (item: ItemView, name: string) => {
  const value = item.verdicts.values[name]
  return typeof value === 'string' ? value : undefined
}

const toCard = (item: ItemView): TriageCard => {
  const unlocked = numberField(item, 'achievements_unlocked')
  const total = numberField(item, 'achievements_total')
  const categories = item.verdicts.values.steam_categories
  return {
    id: item.id,
    title: stringField(item, 'title') ?? item.slug,
    cover: stringField(item, 'cover'),
    playtimeMinutes: numberField(item, 'playtime_minutes'),
    lastPlayed: stringField(item, 'last_played'),
    achievements:
      unlocked !== undefined && total ? { unlocked, total } : undefined,
    steamCategories: Array.isArray(categories)
      ? categories.filter((c): c is string => typeof c === 'string')
      : [],
    hint: item.hint
  }
}

export default async function TriagePage() {
  await requireAdmin()
  const [kinds, items, steamSync] = await Promise.all([
    logAdmin.kinds(),
    logAdmin.items({ kind: TRIAGED_KIND, status: 'untriaged' }),
    logAdmin.steamSync().catch(() => undefined)
  ])
  const kind = kinds.find((k) => k.name === TRIAGED_KIND)
  if (!kind) throw new Error(`kind ${TRIAGED_KIND} is missing from the log`)

  const cards = items
    .map(toCard)
    .sort((a, b) => (b.playtimeMinutes ?? 0) - (a.playtimeMinutes ?? 0))

  return <Triage kind={kind} cards={cards} steamSync={steamSync} />
}
