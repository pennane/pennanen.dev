'use client'

import { useRouter } from 'next/navigation'
import { useState } from 'react'
import { LETTERBOXD_EXPORT_FIELD } from '../../../lib/log/syncs'
import type { SpotifyConnection, SyncReport } from '../../../lib/log/types'
import { useSync } from '../_sync/useSync'
import { importLetterboxdExport } from '../actions'
import { Section } from './Settings'
import styles from './settings.module.css'

const countsOf = (report: SyncReport) =>
  Object.entries(report.counts)
    .map(([count, n]) => `${n} ${count}`)
    .join(', ')

const summaryOf = (report: SyncReport) => {
  switch (report.state) {
    case 'never-ran':
      return 'never ran'
    case 'running':
      return `running… ${countsOf(report)}`
    case 'failed':
      return `failed: ${report.problems.at(-1) ?? ''}`
    case 'succeeded':
      return `${report.finishedAt?.slice(0, 16).replace('T', ' ')} · ${countsOf(report) || 'nothing to do'}`
  }
}

const SyncRow = ({ initial }: { initial: SyncReport }) => {
  const { report = initial, start } = useSync(initial.name, initial)
  return (
    <div className={styles.row}>
      <span className={styles.rowKey}>{report.name}</span>
      <span className={styles.hint} title={report.problems.join('\n')}>
        {summaryOf(report)}
      </span>
      <button type="button" disabled={report.state === 'running'} onClick={start}>
        run
      </button>
    </div>
  )
}

type Import =
  | { state: 'waiting' }
  | { state: 'importing' }
  | { state: 'imported'; report: SyncReport }
  | { state: 'failed'; error: string }

const LetterboxdExport = () => {
  const [result, setResult] = useState<Import>({ state: 'waiting' })
  const router = useRouter()

  const upload = async (file: File) => {
    setResult({ state: 'importing' })
    const form = new FormData()
    form.set(LETTERBOXD_EXPORT_FIELD, file)
    try {
      setResult({ state: 'imported', report: await importLetterboxdExport(form) })
      router.refresh()
    } catch (error) {
      setResult({ state: 'failed', error: String(error) })
    }
  }

  return (
    <Section
      label="letterboxd export"
      hint="the zip from letterboxd settings → data. brings in films older than the first page of your films; once is enough"
    >
      <input
        type="file"
        accept=".zip,application/zip"
        disabled={result.state === 'importing'}
        onChange={(event) => {
          const file = event.target.files?.[0]
          if (file) upload(file)
        }}
      />
      {result.state === 'importing' && <p className={styles.hint}>importing…</p>}
      {result.state === 'failed' && <p className={styles.hint}>{result.error}</p>}
      {result.state === 'imported' && (
        <p className={styles.hint} title={result.report.problems.join('\n')}>
          {countsOf(result.report)}
          {result.report.problems.length > 0 && `, ${result.report.problems.length} problems`}. tmdb
          matching runs next.
        </p>
      )}
    </Section>
  )
}

const SPOTIFY_CONNECT_PATH = '/admin/spotify/connect'

const SpotifyConnect = ({
  connection,
  outcome
}: {
  connection: SpotifyConnection
  outcome?: string
}) => (
  <Section label="spotify">
    {connection === 'not-configured' && (
      <p className={styles.hint}>
        set SPOTIFY_CLIENT_ID and SPOTIFY_CLIENT_SECRET on the log service
      </p>
    )}
    {connection === 'not-connected' && <a href={SPOTIFY_CONNECT_PATH}>connect</a>}
    {connection === 'connected' && (
      <p className={styles.hint}>
        connected · <a href={SPOTIFY_CONNECT_PATH}>reconnect</a>
      </p>
    )}
    {outcome && <p className={styles.hint}>spotify: {outcome}</p>}
  </Section>
)

export const SourcesForm = ({
  reports,
  spotifyConnection,
  spotifyOutcome
}: {
  reports: SyncReport[]
  spotifyConnection: SpotifyConnection
  spotifyOutcome?: string
}) => (
  <article className={styles.kind}>
    <h2 className={styles.kindName}>sources</h2>
    <Section label="syncs" hint="each runs nightly; run one now when you can't wait">
      {reports.map((report) => (
        <SyncRow key={report.name} initial={report} />
      ))}
    </Section>
    <SpotifyConnect connection={spotifyConnection} outcome={spotifyOutcome} />
    <LetterboxdExport />
  </article>
)
