import Link from 'next/link'
import { Project } from '../../lib'
import { formatDate } from '../../../lib'
import { Stack } from '../../../../components/Stack'

export const ProjectLink = ({ project }: { project: Project }) => {
  return (
    <Link href={`/${project.id}`}>
      <Stack>
        <h3>{project.name}</h3>{' '}
        {project.description && `- ${project.description}`}
        <time>{formatDate(project.date)}</time>
      </Stack>
    </Link>
  )
}
