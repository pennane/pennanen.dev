'use client'

import { createContext, useContext, type ReactNode } from 'react'
import type { Presentation } from '../../../../lib/log/types'

const PresentationContext = createContext<Presentation>({ axes: [], kinds: {}, tagTitles: {} })

export const PresentationProvider = ({
  presentation,
  children
}: {
  presentation: Presentation
  children: ReactNode
}) => <PresentationContext.Provider value={presentation}>{children}</PresentationContext.Provider>

export const usePresentation = () => useContext(PresentationContext)

export const useTagTitle = () => {
  const { tagTitles } = usePresentation()
  return (tag: string) => tagTitles[tag]
}
