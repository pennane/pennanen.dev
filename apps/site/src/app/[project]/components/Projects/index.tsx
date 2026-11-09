import { A } from '../../../../components/A'
import { Stack } from '../../../../components/Stack'
import { getFeaturedProjects } from '../../../../lib/projects'
import { ProjectLink } from '../ProjectLink'

export const Projects = async () => {
  const projects = await getFeaturedProjects()
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
