import type { Metadata } from 'next'
import { Fira_Code, Inter } from 'next/font/google'
import './globals.css'
import Layout from '../components/Layout'
import 'highlight.js/styles/github-dark.css'

const inter = Inter({
  subsets: ['latin'],
  display: 'swap'
})

// eslint-disable-next-line @typescript-eslint/no-unused-vars
const firacodeMono = Fira_Code({
  subsets: ['latin'],
  display: 'block'
})

export const metadata: Metadata = {
  title: 'pennanen.dev',
  description:
    "Arttu Pennanen - the digital playground pennanen.dev. Sharing some of the personal projects I've built for the sake of building something."
}

export default function RootLayout({
  children
}: Readonly<{
  children: React.ReactNode
}>) {
  return (
    <html lang="en">
      <body className={`${inter.className}`}>
        <Layout>{children}</Layout>
      </body>
    </html>
  )
}
