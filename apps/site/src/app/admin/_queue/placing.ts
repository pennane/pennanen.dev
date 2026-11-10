import { findAxis } from '../../../lib/log/present'
import type { Kind, ListedEntry } from '../../../lib/log/types'
import { toSections } from '../order/sections'

export type Rival = { id: number; title: string; cover?: string }

export type Ranked = Record<string, Rival[]>

export type Search = { division: string; low: number; high: number }

export type Placement = { division: string; index: number }

export type Answer = 'better' | 'worse'

export const toRanked = (kind: Kind, entries: ListedEntry[]): Ranked => {
  const byId = new Map(entries.map((e) => [e.id, e]))
  const divisions = findAxis(kind.axes, kind.rankedAxis)?.divisions ?? []
  return Object.fromEntries(
    toSections(divisions, entries, kind.rankedAxis).map((section) => [
      section.division.key,
      section.ids.flatMap((id) => {
        const entry = byId.get(id)
        return entry ? [{ id, title: entry.title, cover: entry.cover }] : []
      })
    ])
  )
}

const rivalsIn = (ranked: Ranked, division: string, id: number) =>
  (ranked[division] ?? []).filter((rival) => rival.id !== id)

export const startSearch = (ranked: Ranked, division: string, id: number): Search => ({
  division,
  low: 0,
  high: rivalsIn(ranked, division, id).length
})

export const midpoint = (search: Search) => Math.floor((search.low + search.high) / 2)

export const isSettled = (search: Search) => search.low >= search.high

export const answered = (search: Search, answer: Answer): Search =>
  answer === 'better'
    ? { ...search, high: midpoint(search) }
    : { ...search, low: midpoint(search) + 1 }

export const rivalOf = (ranked: Ranked, search: Search, id: number) =>
  rivalsIn(ranked, search.division, id)[midpoint(search)]

export const questionsLeft = (search: Search) => Math.ceil(Math.log2(search.high - search.low + 1))

export const placementOf = (search: Search): Placement => ({
  division: search.division,
  index: midpoint(search)
})

export const withPlaced = (ranked: Ranked, rival: Rival, placement: Placement): Ranked => {
  const without = withoutRival(ranked, rival.id)
  const division = without[placement.division] ?? []
  return {
    ...without,
    [placement.division]: division.toSpliced(placement.index, 0, rival)
  }
}

export const withoutRival = (ranked: Ranked, id: number): Ranked =>
  Object.fromEntries(
    Object.entries(ranked).map(([division, rivals]) => [
      division,
      rivals.filter((r) => r.id !== id)
    ])
  )

export const ranksIn = (ranked: Ranked) =>
  new Map(
    Object.values(ranked)
      .flat()
      .map((rival, i) => [rival.id, i + 1])
  )

export const divisionOf = (ranked: Ranked, id: number) =>
  Object.entries(ranked).find(([, rivals]) => rivals.some((r) => r.id === id))?.[0]

export const neighboursIn = (ranked: Ranked, id: number) => {
  const order = Object.values(ranked).flat()
  const at = order.findIndex((r) => r.id === id)
  return at < 0 ? undefined : { above: order[at - 1], below: order[at + 1] }
}
