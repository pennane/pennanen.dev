'use client'

import type { IGDBResult } from '../../../lib/log/types'
import { searchIGDB } from '../actions'
import { CatalogueSearch } from './CatalogueSearch'

const shownGame = (result: IGDBResult) => ({
  key: String(result.igdbId),
  name: result.name,
  detail: [result.year, result.platforms.slice(0, 4).join(', ')].filter(Boolean).join(' · '),
  poster: result.poster,
  itemId: result.itemId
})

export const IGDBSearch = (props: {
  initialQuery?: string
  autoFocus?: boolean
  onPick: (result: IGDBResult) => void
  nothingFits?: (query: string) => React.ReactNode
}) => <CatalogueSearch catalogue="igdb" search={searchIGDB} shown={shownGame} {...props} />
