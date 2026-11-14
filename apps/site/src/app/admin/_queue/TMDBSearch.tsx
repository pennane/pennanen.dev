'use client'

import { useCallback } from 'react'
import type { TMDBResult, TMDBType } from '../../../lib/log/types'
import { searchTMDB } from '../actions'
import { CatalogueSearch } from './CatalogueSearch'

const shownEntry = (result: TMDBResult) => ({
  key: `${result.tmdbType}/${result.tmdbId}`,
  name: result.name,
  detail: result.year ?? '',
  poster: result.poster,
  itemId: result.itemId
})

export const TMDBSearch = ({
  tmdbType,
  ...props
}: {
  tmdbType: TMDBType
  autoFocus?: boolean
  onPick: (result: TMDBResult) => void
  nothingFits?: (query: string) => React.ReactNode
}) => {
  const search = useCallback((query: string) => searchTMDB(tmdbType, query), [tmdbType])
  return <CatalogueSearch catalogue="tmdb" search={search} shown={shownEntry} {...props} />
}
