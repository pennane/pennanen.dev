import { GithubIcon, LinkedinIcon } from 'lucide-react'
import Image from 'next/image'
import Link from 'next/link'
import { Gap, Stack } from '../../Stack'
import styles from './navbar.module.css'

const ICON_SIZE = 18

const Navbar = () => {
  return (
    <nav className={styles.nav}>
      <Stack
        horizontal
        gap={{ column: Gap.large, row: Gap.medium }}
        justifyContent="space-between"
        alignItems="center"
      >
        <Stack horizontal alignItems="center" gap={Gap.large}>
          <Link href="/" style={{ marginLeft: '-.125rem' }}>
            <Stack horizontal alignItems="center">
              <Image
                alt=""
                src={'/icons/favicon-96.png'}
                height={ICON_SIZE}
                width={ICON_SIZE}
              />
              <span>pennanen.dev</span>
            </Stack>
          </Link>
          <Link className={styles.link} href="/entries">
            entries
          </Link>
          <Link className={styles.link} href="/projects">
            projects
          </Link>
        </Stack>

        <Stack horizontal>
          <Link
            href="https://github.com/pennane"
            rel="noopener noreferrer"
            target="_blank"
            className={styles.brand}
            aria-label="Visit Arttu Pennanen's GitHub profile"
            aria-describedby="github-link-description"
          >
            <GithubIcon height={ICON_SIZE} />
            <span id="github-link-description" className="sr-only">
              (opens in a new tab)
            </span>
          </Link>

          <Link
            href="https://www.linkedin.com/in/arttu-pennanen/"
            rel="noopener noreferrer"
            target="_blank"
            className={styles.brand}
            aria-label="Visit Arttu Pennanen's LinkedIn profile"
            aria-describedby="linkedin-link-description"
          >
            <LinkedinIcon height={ICON_SIZE} />
            <span id="linkedin-link-description" className="sr-only">
              (opens in a new tab)
            </span>
          </Link>
        </Stack>
      </Stack>
    </nav>
  )
}

export default Navbar
