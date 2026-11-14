import { plainText, type ListedItem } from '../../../../lib/log/present'
import type { Presentation } from '../../../../lib/log/types'
import { countsOf, listed, type KindOverview } from './overview'

const NAMED_IN_SUMMARY = 3

const leaders = (items: ListedItem[]) => items.slice(0, NAMED_IN_SUMMARY).map((item) => item.title)

export const kindSummary = (presentation: Presentation, overview: KindOverview) => {
  const written = plainText(presentation.kinds[overview.kind]?.description ?? '')
  if (written) return written
  const named = leaders(overview.items)
  const count = `${overview.items.length} ${overview.title} in Arttu Pennanen's log`
  return named.length > 0 ? `${count}, led by ${listed(named)}.` : `${count}.`
}

export const logSummary = (overviews: KindOverview[]) =>
  overviews.length > 0
    ? `${listed(countsOf(overviews))}, as Arttu Pennanen saw them.`
    : 'A log of things Arttu Pennanen has seen.'
