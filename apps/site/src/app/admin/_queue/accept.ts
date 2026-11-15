import type { EntryChange, Suggestion } from '../../../lib/log/types'
import type { TriageCard } from './toCard'

export type Acceptance = { change: EntryChange; label: string; cards: TriageCard[] }

export const changeFromSuggestions = (suggestions: Suggestion[]): EntryChange => {
  const change: EntryChange = {}
  const tags: string[] = []
  for (const suggestion of suggestions) {
    if (suggestion.subject === 'status') change.status = suggestion.value
    if (suggestion.subject === 'tag') tags.push(suggestion.value)
    if (suggestion.subject === 'dismissal') change.dismissal = 'dismissed'
  }
  return tags.length > 0 ? { ...change, addTags: tags } : change
}

const labelOf = (change: EntryChange) =>
  [change.status, ...(change.addTags ?? []), change.dismissal].filter(Boolean).join(' + ')

export const acceptancesOf = (cards: TriageCard[], rule: string): Acceptance[] => {
  const byChange = new Map<string, Acceptance>()
  for (const card of cards) {
    const change = changeFromSuggestions(card.suggestions.filter((s) => s.rule === rule))
    if (Object.keys(change).length === 0) continue
    const key = JSON.stringify(change)
    const known = byChange.get(key)
    if (known) known.cards.push(card)
    else byChange.set(key, { change, label: labelOf(change), cards: [card] })
  }
  return [...byChange.values()]
}
