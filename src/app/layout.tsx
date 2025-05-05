import 'highlight.js/styles/github-dark.css'
import type { Metadata } from 'next'
import { Fira_Code, Inter } from 'next/font/google'
import Layout from '../components/Layout'
import './globals.css'

export const inter = Inter({
  subsets: ['latin'],
  variable: '--font-inter',
  display: 'swap'
})

export const firaCode = Fira_Code({
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
    <html lang="en">
      <head>
        <link rel="icon" href="/favicon.ico" sizes="any" />
        <link rel="icon" type="image/svg+xml" href="/icons/favicon.svg" />
        <link rel="apple-touch-icon" href="/images/apple-touch-icon.png" />
      </head>
      <body>
        <Layout>{children}</Layout>
      </body>
    </html>
  )
}
