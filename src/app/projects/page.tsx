import { ProjectGroups } from '../../components/ProjectGroups'
import { Gap, Stack } from '../../components/Stack'

import { getProjects } from '../[project]/lib'

const Page = () => {
  return (
    <Stack gap={Gap.large}>
      <ProjectGroups projects={getProjects()} />
    </Stack>
  )
}

export default Page
