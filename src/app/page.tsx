import Image from 'next/image'
import { A } from '../components/A'
import { JsonLd } from '../components/JsonLd'
import { Gap, Stack } from '../components/Stack'
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
        <div className={styles['intro']}>
          <Image
            alt="Arttu Pennanen thumbs up climbing a vegetated mountain slope"
            src="/images/peukku.jpg"
            width={160}
            height={160}
            className={styles['profile-image']}
          />

          <p>
            Writing code that cannot break. Currently building large-scale
            integrated systems at{' '}
            <A href="https://hoxhunt.com/" rel="noreferrer" target="_blank">
              Hoxhunt
            </A>
            .
          </p>
          <p>
            Pursuing Master&apos;s in Software and Service Engineering at{' '}
            <A href="https://www.aalto.fi/en" rel="noreferrer" target="_blank">
              Aalto University
            </A>
            . B.Eng. in Information Technology from{' '}
            <A
              href="https://www.metropolia.fi/en"
              rel="noreferrer"
              target="_blank"
            >
              Metropolia
            </A>
            .
          </p>
          <p>
            I focus on building reliable systems that delight users &
            developers&mdash;in a functionally pure way on special days, and
            pragmatically on most others.
          </p>
        </div>
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
