import { A } from '../components/A'
import { Gap, Stack } from '../components/Stack'
import { JsonLd } from '../components/JsonLd'
import { generateHomepageJsonLd } from '../lib/json-ld'
import { Projects } from './[project]/components/Projects'
import { getFeaturedProjects } from './[project]/lib'
import { Entries } from './entries/components/Entries'
import { getBlogPosts } from './entries/lib'
import styles from './page.module.css'

export default function Page() {
  const jsonLd = generateHomepageJsonLd(getFeaturedProjects(), getBlogPosts())

  return (
    <Stack className={styles.page} gap={Gap.large}>
      <Stack gap={Gap.large}>
        <Stack gap={Gap.none}>
          <h1>Arttu Pennanen</h1>
          <h2>Software Engineer</h2>
        </Stack>

        <p>
          Writing code that can not break. Currently building large-scale
          integrated systems at{' '}
          <A href="https://hoxhunt.com/" rel="noreferrer" target="_blank">
            Hoxhunt
          </A>
          . Studying MSc at{' '}
          <A href="https://www.aalto.fi/en" rel="noreferrer" target="_blank">
            Aalto
          </A>
          . Alumni of{' '}
          <A
            href="https://www.metropolia.fi/en"
            rel="noreferrer"
            target="_blank"
          >
            Metropolia
          </A>{' '}
          with BEng in Software.
        </p>
        <p>
          I focus on building reliable systems that serve both developers and
          end users.
        </p>
      </Stack>
      <Stack
        className={styles.split}
        gap={{ column: Gap.large, row: Gap.large }}
      >
        <Entries />
        <Projects />
      </Stack>
      <JsonLd object={jsonLd} />
    </Stack>
  )
}
