import { ItemGroups } from '../../components/ProjectGroups'
import { Gap, Stack } from '../../components/Stack'
import { ProjectLink } from '../[project]/components/ProjectLink'

import { getProjects } from '../[project]/lib'

const Page = () => {
  return (
    <Stack gap={Gap.large}>
      <ItemGroups
        items={getProjects()}
        getDate={(x) => x.date}
        render={(project) => <ProjectLink project={project} key={project.id} />}
      />
    </Stack>
  )
}

export default Page
