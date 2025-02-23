import Link from 'next/link'
import { Post } from '../../lib'
import { formatDate, parseDateString } from '../../../lib'
import { Stack } from '../../../../components/Stack'

export const BlogLink = ({ post }: { post: Post }) => {
  return (
    <Link href={`/blog/${post.slug}`} key={post.slug}>
      <Stack direction="column">
        <h3>{post.metadata.title}</h3>
        {formatDate(parseDateString(post.metadata.date))}
      </Stack>
    </Link>
  )
}
