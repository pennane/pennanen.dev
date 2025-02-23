import Link from 'next/link'
import { Project } from '../../lib'
import { formatDate } from '../../../lib'
import { Stack } from '../../../../components/Stack'
import Image from 'next/image'

export const ProjectLink = ({ project }: { project: Project }) => {
  return (
    <Link href={`/${project.id}`}>
      <Stack direction="row" align="flex-start">
        <time style={{ width: '6.5rem', flexShrink: 0 }}>
          {formatDate(project.date)}
        </time>
        <Stack direction="row" align="flex-start">
          <Image
            alt=""
            src={
              project.icon
                ? `/sub/${project.id}/${project.icon}`
                : `/images/placeholder.png`
            }
            width={24}
            height={24}
            style={{ marginTop: '.20rem' }}
          />
          <Stack gap="0" wrap>
            <h4>{project.name}</h4>{' '}
            {/* {project.description && <span>{project.description}</span>} */}
          </Stack>
        </Stack>
      </Stack>
    </Link>
  )
}
