import type { Division, ListedEntry } from '../../../lib/log/types'

export type Section = { division: Division; ids: number[] }

export type Spot = { division: string; index: number }

export const toSections = (divisions: Division[], entries: ListedEntry[], axis: string): Section[] => {
  const ranked = entries
    .filter((e) => e.rank !== undefined && e.scores[axis] !== undefined)
    .sort((a, b) => (a.rank ?? 0) - (b.rank ?? 0))
  return divisions.map((division) => ({
    division,
    ids: ranked.filter((e) => e.scores[axis] === division.key).map((e) => e.id)
  }))
}

export const spotOf = (sections: Section[], id: number): Spot | undefined => {
  for (const section of sections) {
    const index = section.ids.indexOf(id)
    if (index >= 0) return { division: section.division.key, index }
  }
  return undefined
}

export const moved = (sections: Section[], id: number, to: Spot): Section[] =>
  sections.map((section) => {
    const without = section.ids.filter((x) => x !== id)
    if (section.division.key !== to.division) return { ...section, ids: without }
    const index = Math.min(Math.max(to.index, 0), without.length)
    return { ...section, ids: [...without.slice(0, index), id, ...without.slice(index)] }
  })

const sectionIndex = (sections: Section[], division: string) =>
  sections.findIndex((s) => s.division.key === division)

export const oneStep = (sections: Section[], from: Spot, direction: 1 | -1): Spot | undefined => {
  const at = sectionIndex(sections, from.division)
  const size = sections[at].ids.length
  const index = from.index + direction
  if (index >= 0 && index < size) return { division: from.division, index }
  const neighbour = sections[at + direction]
  if (!neighbour) return undefined
  return { division: neighbour.division.key, index: direction === 1 ? 0 : neighbour.ids.length }
}

export const acrossLine = (sections: Section[], from: Spot, direction: 1 | -1): Spot | undefined => {
  const neighbour = sections[sectionIndex(sections, from.division) + direction]
  if (!neighbour) return undefined
  return { division: neighbour.division.key, index: direction === 1 ? 0 : neighbour.ids.length }
}

export const ranksOf = (sections: Section[]) => {
  const ranks = new Map<number, number>()
  sections.flatMap((s) => s.ids).forEach((id, i) => ranks.set(id, i + 1))
  return ranks
}

export const spotBefore = (sections: Section[], id: number, target: number): Spot | undefined => {
  if (id === target) return spotOf(sections, id)
  const section = sections.find((s) => s.ids.includes(target))
  if (!section) return undefined
  return { division: section.division.key, index: section.ids.filter((x) => x !== id).indexOf(target) }
}
