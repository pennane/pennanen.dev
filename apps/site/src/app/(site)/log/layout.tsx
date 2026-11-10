import type { ReactNode } from 'react'
import { logPublic } from '../../../lib/log/public-api'
import { PresentationProvider } from './components/PresentationContext'

export default async function LogLayout({ children }: { children: ReactNode }) {
  const presentation = await logPublic.presentation()
  return <PresentationProvider presentation={presentation}>{children}</PresentationProvider>
}
