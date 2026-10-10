import { ArrowUpRight, LucideIcon } from 'lucide-react'
import { ReactNode } from 'react'
import { A } from '../A'
import styles from './meta-line.module.css'

export const Lede = ({ children }: { children: ReactNode }) => (
  <p className={styles.lede}>{children}</p>
)

export const MetaLine = ({ children }: { children: ReactNode }) => (
  <div className={styles.clip}>
    <p className={styles.meta}>{children}</p>
  </div>
)

const linkText = (href: string) => {
  if (!href.includes('://')) {
    return `pennanen.dev${href.replace(/\/index\.html$/, '')}`
  }
  const url = new URL(href)
  const host = url.host.replace(/^www\./, '')
  const pathname = url.pathname.replace(/\/$/, '')
  return host === 'github.com' ? pathname.slice(1) : host + pathname
}

export const MetaLink = ({
  href,
  icon: Icon
}: {
  href: string
  icon: LucideIcon
}) => {
  const external = href.includes('://')
  return (
    <span className={styles.item}>
      <Icon size={14} aria-hidden />
      <A href={href} {...(external && { rel: 'noreferrer', target: '_blank' })}>
        {linkText(href)}
        {external && (
          <ArrowUpRight
            className={styles.arrow}
            size={12}
            strokeWidth={2.5}
            aria-hidden
          />
        )}
      </A>
    </span>
  )
}
