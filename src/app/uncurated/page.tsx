import { JsonLd } from '../../components/JsonLd'
import { generateUncuratedProjectsJsonLd } from '../../lib/json-ld'
import { getProjects } from '../../lib/projects'
import { Stack } from '../../components/Stack'
import styles from './page.module.css'
import { ItemGroups } from '../../components/ProjectGroups'
import { ProjectLink } from '../[project]/components/ProjectLink'

export default async function Page() {
  const projects = (await getProjects()).filter((p) => !p.featured)
  const jsonLd = generateUncuratedProjectsJsonLd(projects)

  return (
    <Stack className={styles.page}>
      <ItemGroups
        items={projects}
        getDate={(x) => x.date}
        render={(project) => <ProjectLink project={project} key={project.id} />}
      />
      <JsonLd object={jsonLd} />
    </Stack>
  )
}
