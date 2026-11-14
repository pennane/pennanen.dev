'use client'

import { useRouter } from 'next/navigation'
import { useEffect, useState } from 'react'
import type { SyncReport } from '../../../lib/log/types'
import { startSync, syncStatus } from '../actions'

const SYNC_POLL_MS = 3000

export const useSync = (name: string, initial?: SyncReport) => {
  const [report, setReport] = useState(initial)
  const router = useRouter()

  useEffect(() => {
    if (report?.state !== 'running') return
    const timer = setInterval(async () => {
      const next = await syncStatus(name)
      setReport(next)
      if (next.state !== 'running') router.refresh()
    }, SYNC_POLL_MS)
    return () => clearInterval(timer)
  }, [name, report?.state, router])

  return { report, start: async () => setReport(await startSync(name)) }
}
