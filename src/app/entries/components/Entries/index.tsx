import { Stack } from '../../../../components/Stack'
import { getBlogPosts } from '../../lib'
import { BlogLink } from '../BlogLink'

export const Entries = () => {
  const projects = getBlogPosts()
  return (
    <Stack>
      <h3>Entries</h3>
      <Stack gap={'.75rem'}>
        {projects.map((post) => (
          <BlogLink key={post.slug} post={post} />
        ))}
      </Stack>
      {}
    </Stack>
  )
}
