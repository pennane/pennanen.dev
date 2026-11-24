import type { ItemView, Kind, Suggestion } from '../../../lib/log/types'
import type { GameCard } from './card'

const numberField = (item: ItemView, name: string) => {
  const value = item.verdicts.values[name]
  return typeof value === 'number' ? value : undefined
}

const stringField = (item: ItemView, name: string) => {
  const value = item.verdicts.values[name]
  return typeof value === 'string' ? value : undefined
}

const stringListField = (item: ItemView, name: string) => {
  const value = item.verdicts.values[name]
  return Array.isArray(value) ? value.filter((v): v is string => typeof v === 'string') : []
}

export const unrankedOrderOf = (kind: Kind, item: ItemView) =>
  (kind.unrankedOrder && numberField(item, kind.unrankedOrder)) || 0

export const toGameCard = (item: ItemView): GameCard => {
  const unlocked = numberField(item, 'achievements_unlocked')
  const total = numberField(item, 'achievements_total')
  return {
    id: item.id,
    title: stringField(item, 'title') ?? item.slug,
    cover: stringField(item, 'cover') ?? stringField(item, 'image'),
    artists: stringListField(item, 'artists'),
    playtimeMinutes: numberField(item, 'playtime_minutes'),
    lastPlayed: stringField(item, 'last_played'),
    achievements: unlocked !== undefined && total ? { unlocked, total } : undefined,
    steamCategories: stringListField(item, 'steam_categories'),
    status: item.entry?.status ?? 'untriaged',
    tags: item.entry?.tags ?? []
  }
}

export type TriageShape = 'untriaged' | 'match'

type CardBase = GameCard & { order: number; suggestions: Suggestion[] }

export type TriageCard = ({ shape: 'untriaged' } & CardBase) | ({ shape: 'match' } & CardBase)

export const toCard = (kind: Kind, shape: TriageShape, item: ItemView): TriageCard => ({
  shape,
  ...toGameCard(item),
  order: unrankedOrderOf(kind, item),
  suggestions: item.suggestions
})

export const byOrder = (a: TriageCard, b: TriageCard) => b.order - a.order
