import { A } from '../../../../components/A'
import { Stack } from '../../../../components/Stack'
import { getFeaturedProjects } from '../../lib'
import { ProjectLink } from '../ProjectLink'

export const Projects = () => {
  const projects = getFeaturedProjects()
  return (
    <Stack>
      <h3>Projects</h3>
      <Stack>
        {projects.map((project) => (
          <ProjectLink key={project.id} project={project} />
        ))}
      </Stack>
      <A style={{ marginTop: '.25rem' }} href="/uncurated">
        rest of stuff
      </A>
    </Stack>
  )
}
