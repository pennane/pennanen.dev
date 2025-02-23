import Link from 'next/link'
import { Project } from '../../lib'
import { formatDate } from '../../../lib'
import { Stack } from '../../../../components/Stack'
import Image from 'next/image'

export const ProjectLink = ({ project }: { project: Project }) => {
  return (
    <Link href={`/${project.id}`}>
      <Stack direction="row">
        <Image
          alt=""
          src={
            project.icon
              ? `/sub/${project.id}/${project.icon}`
              : `/images/placeholder.png`
          }
          width={32}
          height={32}
          style={{ marginTop: '.20rem' }}
        />
        <Stack gap=".25rem">
          <h4>{project.name}</h4>{' '}
          {project.description && <span>{project.description}</span>}
          <time>{formatDate(project.date)}</time>
        </Stack>
      </Stack>
    </Link>
  )
}
