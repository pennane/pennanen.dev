import styles from './page.module.css'
import { Stack } from '../components/Stack'
import { Projects } from './[project]/components/Projects'
import { Entries } from './entries/components/Entries'

export default function Page() {
  return (
    <Stack className={styles.page} gap="1rem">
      <Stack>
        <Stack gap="0">
          <h1>Arttu Pennanen</h1>
          <h2>Software Engineer</h2>
        </Stack>
        <p>
          Writing code that can not break. Currently at Hoxhunt. Metropolia
          alumni. Here on this website, I post about some of the problems
          I&apos;ve fought.
        </p>
      </Stack>
      <Entries />
      <Projects />
    </Stack>
  )
}
