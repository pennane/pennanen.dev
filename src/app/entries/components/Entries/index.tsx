import { Stack } from '../../../../components/Stack'
import { getBlogPosts } from '../../lib'
import { BlogLink } from '../BlogLink'

export const Entries = () => {
  const projects = getBlogPosts()
  return (
    <Stack>
      <h3>Entries</h3>
      <Stack>
        {projects.map((post) => (
          <BlogLink key={post.slug} post={post} />
        ))}
      </Stack>
    </Stack>
  )
}
