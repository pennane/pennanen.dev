import Link from 'next/link'
import { getProjects } from './[project]/lib'
import styles from './page.module.css'
import { Stack } from '../components/Stack'
import { ProjectLink } from './[project]/components/ProjectLink'

export default function Page() {
  const projects = getProjects()
    .filter((p) => !p.ignoreInListing)
    .toSorted((a, b) => {
      if (!a.date) return 1
      if (!b.date) return -1
      return b.date - a.date
    })

  return (
    <Stack className={styles.page} gap="2rem">
      <Stack>
        {projects.map((project) => (
          <ProjectLink key={project.id} project={project} />
        ))}
      </Stack>
      <Link href="/uncurated">rest of stuff</Link>
    </Stack>
  )
}
