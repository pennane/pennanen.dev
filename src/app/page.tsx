import styles from './page.module.css'
import { Gap, Stack } from '../components/Stack'
import { Projects } from './[project]/components/Projects'
import { Entries } from './entries/components/Entries'
import Link from 'next/link'

export default function Page() {
  return (
    <Stack className={styles.page} gap={Gap.xLarge}>
      <Stack vertical alignItems="flex-end" gap={Gap.large} wrap>
        <Stack gap={Gap.none}>
          <h1>Arttu Pennanen</h1>
          <h2>Software Engineer</h2>
        </Stack>

        <p>
          Writing code that can not break. Currently at{' '}
          <Link href="https://hoxhunt.com/" rel="noreferrer" target="_blank">
            Hoxhunt
          </Link>
          .{' '}
          <Link
            href="https://www.metropolia.fi/en"
            rel="noreferrer"
            target="_blank"
          >
            Metropolia
          </Link>{' '}
          alumni. Here on this website, I post about some of the problems
          I&apos;ve fought.
        </p>
      </Stack>
      <Entries />
      <Projects />
    </Stack>
  )
}
