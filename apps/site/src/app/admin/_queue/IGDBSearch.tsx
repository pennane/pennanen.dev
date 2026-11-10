'use client'

import Image from 'next/image'
import { useEffect, useState } from 'react'
import type { IGDBResult } from '../../../lib/log/types'
import { searchIGDB } from '../actions'
import styles from './igdb.module.css'

const SEARCH_AFTER_TYPING_MS = 350

type Search =
  | { state: 'idle' }
  | { state: 'searching' }
  | { state: 'found'; query: string; results: IGDBResult[] }
  | { state: 'failed'; error: string }

export const IGDBSearch = ({
  initialQuery = '',
  autoFocus,
  onPick,
  nothingFits
}: {
  initialQuery?: string
  autoFocus?: boolean
  onPick: (result: IGDBResult) => void
  nothingFits?: (query: string) => React.ReactNode
}) => {
  const [query, setQuery] = useState(initialQuery)
  const [search, setSearch] = useState<Search>({ state: 'idle' })

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
        const results = await searchIGDB(wanted)
        if (current) setSearch({ state: 'found', query: wanted, results })
      } catch (error) {
        if (current) setSearch({ state: 'failed', error: String(error) })
      }
    }, SEARCH_AFTER_TYPING_MS)
    return () => {
      current = false
      clearTimeout(timer)
    }
  }, [query])

  return (
    <div className={styles.search}>
      <input
        className={styles.input}
        type="search"
        value={query}
        autoFocus={autoFocus}
        placeholder="search igdb"
        enterKeyHint="search"
        onChange={(event) => setQuery(event.target.value)}
      />
      {search.state === 'searching' && <p className={styles.quiet}>searching…</p>}
      {search.state === 'failed' && <p className={styles.failed}>{search.error}</p>}
      {search.state === 'found' && (
        <ul className={styles.results}>
          {search.results.map((result) => (
            <li key={result.igdbId}>
              <button type="button" className={styles.result} onClick={() => onPick(result)}>
                <span className={styles.poster}>
                  {result.poster && (
                    <Image src={result.poster} alt="" width={45} height={64} unoptimized />
                  )}
                </span>
                <span className={styles.body}>
                  <span className={styles.name}>{result.name}</span>
                  <span className={styles.quiet}>
                    {[result.year, result.platforms.slice(0, 4).join(', ')]
                      .filter(Boolean)
                      .join(' · ')}
                  </span>
                </span>
                {result.itemId !== undefined && <span className={styles.inLog}>in log</span>}
              </button>
            </li>
          ))}
          {search.results.length === 0 && (
            <li className={styles.quiet}>igdb has nothing for that</li>
          )}
          {nothingFits && <li>{nothingFits(search.query)}</li>}
        </ul>
      )}
    </div>
  )
}
