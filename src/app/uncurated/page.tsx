import { getProjects } from '../[project]/lib'
import { Stack } from '../../components/Stack'
import styles from './page.module.css'
import { ItemGroups } from '../../components/ProjectGroups'
import { ProjectLink } from '../[project]/components/ProjectLink'

export default function Page() {
  const projects = getProjects().filter((p) => p.ignoreInListing)

  return (
    <Stack className={styles.page}>
      <ItemGroups
        items={projects}
        getDate={(x) => x.date}
        render={(project) => <ProjectLink project={project} key={project.id} />}
      />
    </Stack>
  )
}
