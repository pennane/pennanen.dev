import Link from 'next/link'
import { Gap, Stack } from '../../../../components/Stack'
import { getProjects } from '../../lib'
import { ProjectLink } from '../ProjectLink'

export const Projects = () => {
  const projects = getProjects().filter((p) => !p.ignoreInListing)
  return (
    <Stack gap={Gap.large}>
      <h3>Projects</h3>
      <Stack>
        {projects.map((project) => (
          <ProjectLink key={project.id} project={project} />
        ))}
      </Stack>
      <Link href="/uncurated">rest of stuff</Link>
    </Stack>
  )
}
