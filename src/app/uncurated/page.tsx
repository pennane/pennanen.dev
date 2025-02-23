import { getProjects } from '../[project]/lib'
import { Stack } from '../../components/Stack'
import styles from './page.module.css'
import { ProjectLink } from '../[project]/components/ProjectLink'

export default function Page() {
  const projects = getProjects()
    .filter((p) => p.ignoreInListing)
    .toSorted((a, b) => {
      if (!a.date) return 1
      if (!b.date) return -1
      return b.date - a.date
    })

  return (
    <Stack className={styles.page}>
      {projects.map((project) => (
        <ProjectLink key={project.id} project={project} />
      ))}
    </Stack>
  )
}
