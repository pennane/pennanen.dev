import Link from 'next/link'
import { CSSProperties, FC, PropsWithChildren } from 'react'
import styles from './a.module.css'

export const A: FC<
  PropsWithChildren<{
    href: string
    className?: string
    rel?: string
    target?: string
    style?: CSSProperties
  }>
> = ({ children, className, ...rest }) => {
  return (
    <Link
      className={[className, styles.link].filter(Boolean).join(' ')}
      {...rest}
    >
      {children}
    </Link>
  )
}
