import Link from 'next/link'
import { Gap, Stack } from '../components/Stack'
import { Projects } from './[project]/components/Projects'
import { Entries } from './entries/components/Entries'
import styles from './page.module.css'

export default function Page() {
  return (
    <Stack className={styles.page} gap={Gap.mega}>
      <Stack gap={Gap.large}>
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
          . Studying MSc at{' '}
          <Link href="https://www.aalto.fi/en" rel="noreferrer" target="_blank">
            Aalto
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
      <Stack
        horizontal
        wrap="wrap-reverse"
        className={styles.split}
        gap={{ column: Gap.large, row: Gap.large }}
      >
        <Projects />
        <Entries />
      </Stack>
    </Stack>
  )
}
