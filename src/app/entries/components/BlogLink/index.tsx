import Link from 'next/link'
import { Stack } from '../../../../components/Stack'
import { formatDate, parseDateString } from '../../../lib'
import { Post } from '../../lib'

export const BlogLink = ({ post }: { post: Post }) => {
  return (
    <Stack horizontal alignItems="center">
      <time>{formatDate(parseDateString(post.metadata.date))}</time>
      <Link href={`/entries/${post.slug}`} key={post.slug}>
        <h4>{post.metadata.title}</h4>
      </Link>
    </Stack>
  )
}
