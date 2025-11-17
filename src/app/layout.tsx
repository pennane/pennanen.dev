/* eslint-disable @typescript-eslint/no-unused-vars */
import 'highlight.js/styles/github-dark.css'
import type { Metadata } from 'next'
import { Fira_Code, Inter } from 'next/font/google'
import Script from 'next/script'
import Layout from '../components/Layout'
import './globals.css'

const inter = Inter({
  subsets: ['latin'],
  variable: '--font-inter',
  display: 'swap'
})

const firaCode = Fira_Code({
  subsets: ['latin'],
  display: 'block'
})

export const metadata: Metadata = {
  title: 'Arttu Pennanen',
  description:
    'Digital playground pennanen.dev - sharing some of personal projects and writings.'
}

export default function RootLayout({
  children
}: Readonly<{
  children: React.ReactNode
}>) {
  return (
    <html lang="en" suppressHydrationWarning>
      <head>
        <link rel="icon" href="/favicon.ico" sizes="any" />
        <link rel="icon" href="/favicon.svg" type="image/svg+xml" />
        <link rel="apple-touch-icon" href="/icons/favicon-180.png" />
        <link rel="manifest" href="/site.webmanifest" />
        <Script id="theme-init" strategy="beforeInteractive">
          {`
            (function () {
              try {
                var theme = localStorage.getItem('theme');
                var systemTheme = window.matchMedia('(prefers-color-scheme: dark)').matches ? 'dark' : 'light';
                var initialTheme = theme || systemTheme;
                document.documentElement.setAttribute('data-color-scheme', initialTheme);
              } catch (e) {
                var systemTheme = window.matchMedia('(prefers-color-scheme: dark)').matches ? 'dark' : 'light';
                document.documentElement.setAttribute('data-color-scheme', systemTheme);
              }
            })();
          `}
        </Script>
      </head>
      <body>
        <Layout>{children}</Layout>
      </body>
    </html>
  )
}
