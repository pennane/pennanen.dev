import '../styles/globals.css'
import type { AppProps } from 'next/app'
import { AppContextProvider } from '../lib/context'
import { ThemeProvider } from 'next-themes'

function App({ Component, pageProps }: AppProps) {
  return (
    <ThemeProvider>
      <AppContextProvider>
        <Component {...pageProps} />
      </AppContextProvider>
    </ThemeProvider>
  )
}

export default App
