import Link from 'next/link'
import { Post } from '../../lib'
import { formatDate, parseDateString } from '../../../lib'
import { Stack } from '../../../../components/Stack'

export const BlogLink = ({ post }: { post: Post }) => {
  return (
    <Link href={`/entries/${post.slug}`} key={post.slug}>
      <Stack direction="row" wrap="wrap-reverse" align="flex-start">
        <time style={{ width: '6.5rem' }}>
          {formatDate(parseDateString(post.metadata.date))}
        </time>
        <h4>{post.metadata.title}</h4>
      </Stack>
    </Link>
  )
}
