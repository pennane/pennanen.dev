import Link from 'next/link'
import Markdown, { type Components } from 'react-markdown'
import { readImageWidth } from '../../../../lib/log/image-width'
import styles from './log.module.css'

const isOnSite = (href: string) => href.startsWith('/') && !href.startsWith('//')

const components: Components = {
  a: ({ href = '', children }) =>
    isOnSite(href) ? (
      <Link href={href}>{children}</Link>
    ) : (
      <a href={href} target="_blank" rel="noreferrer">
        {children}
      </a>
    ),
  img: ({ src, alt = '' }) => (typeof src === 'string' ? <SizedImg href={src} alt={alt} /> : null)
}

const SizedImg = ({ href, alt }: { href: string; alt: string }) => {
  const { src, width } = readImageWidth(href)
  return (
    <img
      src={src}
      alt={alt}
      loading="lazy"
      className={styles.proseImage}
      style={width ? { width } : undefined}
    />
  )
}

export const Prose = ({ markdown }: { markdown: string }) => (
  <div className={styles.prose}>
    <Markdown components={components}>{markdown}</Markdown>
  </div>
)
