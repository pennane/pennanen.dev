'use client'

import Image from 'next/image'
import { useEffect, useState } from 'react'
import styles from './igdb.module.css'

const SEARCH_AFTER_TYPING_MS = 350

export type Found = {
  key: string
  name: string
  detail: string
  poster?: string
  itemId?: number
}

type Search<R> =
  | { state: 'idle' }
  | { state: 'searching' }
  | { state: 'found'; query: string; results: R[] }
  | { state: 'failed'; error: string }

export const CatalogueSearch = <R,>({
  catalogue,
  search: searchCatalogue,
  shown,
  initialQuery = '',
  autoFocus,
  onPick,
  nothingFits
}: {
  catalogue: string
  search: (query: string) => Promise<R[]>
  shown: (result: R) => Found
  initialQuery?: string
  autoFocus?: boolean
  onPick: (result: R) => void
  nothingFits?: (query: string) => React.ReactNode
}) => {
  const [query, setQuery] = useState(initialQuery)
  const [search, setSearch] = useState<Search<R>>({ state: 'idle' })

  useEffect(() => {
    const wanted = query.trim()
    if (!wanted) {
      setSearch({ state: 'idle' })
      return
    }
    let current = true
    const timer = setTimeout(async () => {
      setSearch({ state: 'searching' })
      try {
        const results = await searchCatalogue(wanted)
        if (current) setSearch({ state: 'found', query: wanted, results })
      } catch (error) {
        if (current) setSearch({ state: 'failed', error: String(error) })
      }
    }, SEARCH_AFTER_TYPING_MS)
    return () => {
      current = false
      clearTimeout(timer)
    }
  }, [query, searchCatalogue])

  return (
    <div className={styles.search}>
      <input
        className={styles.input}
        type="search"
        value={query}
        autoFocus={autoFocus}
        placeholder={`search ${catalogue}`}
        enterKeyHint="search"
        onChange={(event) => setQuery(event.target.value)}
      />
      {search.state === 'searching' && <p className={styles.quiet}>searching…</p>}
      {search.state === 'failed' && <p className={styles.failed}>{search.error}</p>}
      {search.state === 'found' && (
        <ul className={styles.results}>
          {search.results.map((result) => {
            const found = shown(result)
            return (
              <li key={found.key}>
                <button type="button" className={styles.result} onClick={() => onPick(result)}>
                  <span className={styles.poster}>
                    {found.poster && (
                      <Image src={found.poster} alt="" width={45} height={64} unoptimized />
                    )}
                  </span>
                  <span className={styles.body}>
                    <span className={styles.name}>{found.name}</span>
                    <span className={styles.quiet}>{found.detail}</span>
                  </span>
                  {found.itemId !== undefined && <span className={styles.inLog}>in log</span>}
                </button>
              </li>
            )
          })}
          {search.results.length === 0 && (
            <li className={styles.quiet}>{catalogue} has nothing for that</li>
          )}
          {nothingFits && <li>{nothingFits(search.query)}</li>}
        </ul>
      )}
    </div>
  )
}
