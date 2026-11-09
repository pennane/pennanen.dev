import type { Metadata, Viewport } from 'next'
import styles from './admin.module.css'

export const metadata: Metadata = {
  title: 'log admin',
  manifest: '/admin.webmanifest',
  robots: { index: false, follow: false },
  appleWebApp: { capable: true, title: 'log', statusBarStyle: 'black-translucent' }
}

export const viewport: Viewport = {
  width: 'device-width',
  initialScale: 1,
  viewportFit: 'cover'
}

export default function AdminLayout({
  children
}: Readonly<{ children: React.ReactNode }>) {
  return <div className={styles.shell}>{children}</div>
}
