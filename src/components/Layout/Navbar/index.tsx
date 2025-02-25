import Link from 'next/link'
import { Gap, Stack } from '../../Stack'
import Image from 'next/image'
import { GithubIcon, LinkedinIcon } from 'lucide-react'
import styles from './navbar.module.css'

const ICON_SIZE = 18

const Navbar = () => {
  return (
    <nav className={styles.nav}>
      <Stack
        vertical
        gap={{ column: Gap.large, row: Gap.medium }}
        justifyContent="space-between"
        alignItems="center"
      >
        <Stack vertical alignItems="center" gap={Gap.large}>
          <Link href="/" style={{ marginLeft: '-.125rem' }}>
            <Stack vertical alignItems="center">
              <Image
                alt=""
                src={'/icons/favicon-96.png'}
                height={ICON_SIZE}
                width={ICON_SIZE}
              />
              <span>pennanen.dev</span>
            </Stack>
          </Link>
          <Link href="/entries">entries</Link>
          <Link href="/projects">projects</Link>
        </Stack>

        <Stack vertical>
          <Link
            href="https://github.com/pennane"
            rel="noopener noreferrer"
            target="_blank"
            className={styles.brand}
            aria-label="Visit Arttu Pennanen's GitHub profile"
          >
            <GithubIcon height={ICON_SIZE} />
            <span className="sr-only">(opens in a new tab)</span>
          </Link>

          <Link
            href="https://www.linkedin.com/in/arttu-pennanen/"
            rel="noopener noreferrer"
            target="_blank"
            className={styles.brand}
            aria-label="Visit Arttu Pennanen's LinkedIn profile"
          >
            <LinkedinIcon height={ICON_SIZE} />
            <span className="sr-only">(opens in a new tab)</span>
          </Link>
        </Stack>
      </Stack>
    </nav>
  )
}

export default Navbar
