import Image from 'next/image'
import Link from 'next/link'
import { Gap, Stack } from '../../../../components/Stack'
import { formatDate } from '../../../lib'
import { isUnoptimizable } from '../../../../lib/images'
import { Project } from '../../../../lib/projects'

export const ProjectLink = ({ project }: { project: Project }) => {
  return (
    <Stack horizontal alignItems="center" gap={Gap.small}>
      <time>{formatDate(project.date)}</time>
      <Link href={`/${project.id}`}>
        <Stack horizontal alignItems="center">
          <Image
            alt=""
            src={project.icon}
            unoptimized={isUnoptimizable(project.icon)}
            width={24}
            height={24}
          />
          <Stack gap={Gap.none} wrap>
            <h4>{project.title}</h4>
          </Stack>
        </Stack>
      </Link>
    </Stack>
  )
}
