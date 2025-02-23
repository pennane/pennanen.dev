import { Stack } from '../../components/Stack'
import { ProjectLink } from '../[project]/components/ProjectLink'

import { getProjects } from '../[project]/lib'

const Page = () => {
  return (
    <Stack>
      {getProjects().map((post) => (
        <ProjectLink key={post.id} project={post} />
      ))}
    </Stack>
  )
}

export default Page
