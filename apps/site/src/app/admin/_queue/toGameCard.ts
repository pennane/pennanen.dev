import type { ItemView } from '../../../lib/log/types'
import type { GameCard } from './card'

const numberField = (item: ItemView, name: string) => {
  const value = item.verdicts.values[name]
  return typeof value === 'number' ? value : undefined
}

const stringField = (item: ItemView, name: string) => {
  const value = item.verdicts.values[name]
  return typeof value === 'string' ? value : undefined
}

export const toGameCard = (item: ItemView): GameCard => {
  const unlocked = numberField(item, 'achievements_unlocked')
  const total = numberField(item, 'achievements_total')
  const categories = item.verdicts.values.steam_categories
  return {
    id: item.id,
    title: stringField(item, 'title') ?? item.slug,
    cover: stringField(item, 'cover'),
    playtimeMinutes: numberField(item, 'playtime_minutes'),
    lastPlayed: stringField(item, 'last_played'),
    achievements: unlocked !== undefined && total ? { unlocked, total } : undefined,
    steamCategories: Array.isArray(categories)
      ? categories.filter((c): c is string => typeof c === 'string')
      : [],
    status: item.entry?.status ?? 'untriaged',
    tags: item.entry?.tags ?? [],
    hint: item.hint
  }
}

export const byPlaytime = (a: GameCard, b: GameCard) =>
  (b.playtimeMinutes ?? 0) - (a.playtimeMinutes ?? 0)
