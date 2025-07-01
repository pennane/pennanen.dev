import { GithubIcon, LinkedinIcon } from 'lucide-react'
import Image from 'next/image'
import Link from 'next/link'
import { Gap, Stack } from '../../Stack'
import { ThemeToggle } from '../../ThemeToggle'
import styles from './navbar.module.css'

const ICON_SIZE = 18

const NavItem = ({ children, ...props }: React.ComponentProps<'div'>) => (
  <div className={styles.navItem} {...props}>
    {children}
  </div>
)

const Navbar = () => {
  return (
    <nav className={styles.nav}>
      <Stack
        horizontal
        gap={{ column: Gap.large, row: Gap.medium }}
        justifyContent="space-between"
        alignItems="center"
      >
        <Stack horizontal alignItems="flex-start" gap={Gap.large}>
          <NavItem>
            <Link href="/" style={{ marginLeft: '-.125rem' }}>
              <Stack
                horizontal
                alignItems="flex-start"
                className={styles.logoStack}
              >
                <Image
                  alt=""
                  src={'/icons/favicon-96.png'}
                  height={ICON_SIZE}
                  width={ICON_SIZE}
                />
                <span>pennanen.dev</span>
              </Stack>
            </Link>
          </NavItem>
          <NavItem>
            <Link className={styles.link} href="/entries">
              entries
            </Link>
          </NavItem>
          <NavItem>
            <Link className={styles.link} href="/projects">
              projects
            </Link>
          </NavItem>
        </Stack>

        <Stack horizontal alignItems="center" gap={Gap.mSmall}>
          <Stack horizontal alignItems="center" gap={Gap.small}>
            <NavItem>
              <Link
                href="https://github.com/pennane"
                rel="noopener noreferrer"
                target="_blank"
                className={styles.brand}
                aria-label="Visit Arttu Pennanen's GitHub profile"
                aria-describedby="github-link-description"
                title="Visit Arttu Pennanen's GitHub profile"
              >
                <GithubIcon height={ICON_SIZE} />
                <span id="github-link-description" className="sr-only">
                  (opens in a new tab)
                </span>
              </Link>
            </NavItem>

            <NavItem>
              <Link
                href="https://www.linkedin.com/in/arttu-pennanen/"
                rel="noopener noreferrer"
                target="_blank"
                className={styles.brand}
                aria-label="Visit Arttu Pennanen's LinkedIn profile"
                aria-describedby="linkedin-link-description"
                title="Visit Arttu Pennanen's LinkedIn profile"
              >
                <LinkedinIcon height={ICON_SIZE} />
                <span id="linkedin-link-description" className="sr-only">
                  (opens in a new tab)
                </span>
              </Link>
            </NavItem>
          </Stack>
          <ThemeToggle />
        </Stack>
      </Stack>
    </nav>
  )
}

export default Navbar
