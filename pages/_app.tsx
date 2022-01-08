import '../styles/globals.css'
import type { AppProps } from 'next/app'
import { AppContextProvider } from '../lib/context'

function App({ Component, pageProps }: AppProps) {
    return (
        <AppContextProvider>
            <Component {...pageProps} />+{' '}
        </AppContextProvider>
    )
}

export default App
