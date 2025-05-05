import Image from 'next/image'
import Link from 'next/link'
import { Gap, Stack } from '../../../../components/Stack'
import { formatDate } from '../../../lib'
import { Project } from '../../lib'

export const ProjectLink = ({ project }: { project: Project }) => {
  return (
    <Stack horizontal alignItems="flex-start" gap={Gap.small}>
      <time>{formatDate(project.date)}</time>
      <Link href={`/${project.id}`}>
        <Stack horizontal alignItems="flex-start">
          <Image
            alt=""
            src={
              project.icon
                ? `/sub/${project.id}/${project.icon}`
                : `/images/placeholder.png`
            }
            width={24}
            height={24}
          />
          <Stack gap={Gap.none} wrap>
            <h4 style={{ whiteSpace: 'nowrap' }}>{project.name}</h4>
          </Stack>
        </Stack>
      </Link>
    </Stack>
  )
}
