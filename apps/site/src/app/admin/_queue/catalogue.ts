import { TMDB_TYPE_OF_SOURCE, type Kind, type TMDBType } from '../../../lib/log/types'

const IGDB_SOURCE = 'igdb'

export type Catalogue = { name: 'igdb' } | { name: 'tmdb'; tmdbType: TMDBType }

export const catalogueOf = (kind: Kind): Catalogue | undefined => {
  if (kind.identitySource === IGDB_SOURCE) return { name: 'igdb' }
  const tmdbType = TMDB_TYPE_OF_SOURCE[kind.identitySource ?? '']
  return tmdbType ? { name: 'tmdb', tmdbType } : undefined
}
