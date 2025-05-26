import { ItemGroups } from '../../components/ProjectGroups'
import { Gap, Stack } from '../../components/Stack'
import { JsonLd } from '../../components/JsonLd'
import { generateProjectsListingJsonLd } from '../../lib/json-ld'
import { ProjectLink } from '../[project]/components/ProjectLink'
import { getProjects } from '../[project]/lib'

const Page = () => {
  const projects = getProjects()
  const jsonLd = generateProjectsListingJsonLd(projects)

  return (
    <Stack gap={Gap.large}>
      <ItemGroups
        items={projects}
        getDate={(x) => x.date}
        render={(project) => <ProjectLink project={project} key={project.id} />}
      />
      <JsonLd object={jsonLd} />
    </Stack>
  )
}

export default Page
