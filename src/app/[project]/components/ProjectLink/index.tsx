import Link from 'next/link'
import { Project } from '../../lib'
import { formatDate } from '../../../lib'
import { Gap, Stack } from '../../../../components/Stack'
import Image from 'next/image'

export const ProjectLink = ({ project }: { project: Project }) => {
  return (
    <Stack vertical alignItems="flex-start">
      <time style={{ width: '6.5rem', flexShrink: 0 }}>
        {formatDate(project.date)}
      </time>
      <Link href={`/${project.id}`}>
        <Stack vertical alignItems="flex-start">
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
            <h4>{project.name}</h4>{' '}
          </Stack>
        </Stack>
      </Link>
    </Stack>
  )
}
