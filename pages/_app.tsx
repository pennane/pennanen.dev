// noinspection JSUnusedGlobalSymbols

import '../styles/globals.css'
import type { AppProps } from 'next/app'
import { AppContextProvider } from '../lib/context'
import { ThemeProvider } from 'next-themes'
import { GoogleAnalytics, usePagesViews } from 'nextjs-google-analytics'

function App({ Component, pageProps }: AppProps) {
  usePagesViews()
  return (
    <ThemeProvider>
      <AppContextProvider>
        <GoogleAnalytics />
        <Component {...pageProps} />
      </AppContextProvider>
    </ThemeProvider>
  )
}

export default App
