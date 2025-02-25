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
          Writing code that can not break. Currently building large-scale
          integrated systems at{' '}
          <Link href="https://hoxhunt.com/" rel="noreferrer" target="_blank">
            Hoxhunt
          </Link>
          . Alumni of{' '}
          <Link
            href="https://www.metropolia.fi/en"
            rel="noreferrer"
            target="_blank"
          >
            Metropolia
          </Link>{' '}
          with BEng in Software.
        </p>
        <p>
          I thrive on building fail-proof systems enjoyed by developers and end
          users.
        </p>
      </Stack>
      <Entries />
      <Projects />
    </Stack>
  )
}
