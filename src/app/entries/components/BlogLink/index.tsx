import Link from 'next/link'
import { Post } from '../../lib'
import { formatDate, parseDateString } from '../../../lib'
import { Stack } from '../../../../components/Stack'

export const BlogLink = ({ post }: { post: Post }) => {
  return (
    <Link href={`/entries/${post.slug}`} key={post.slug}>
      <Stack direction="column">
        <h4>{post.metadata.title}</h4>
        {formatDate(parseDateString(post.metadata.date))}
      </Stack>
    </Link>
  )
}
