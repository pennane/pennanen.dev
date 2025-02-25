import { getBlogPosts } from './lib'
import styles from './page.module.css'
import { Stack } from '../../components/Stack'
import { BlogLink } from './components/BlogLink'

export default function Blog() {
  return (
    <Stack className={styles.page}>
      {getBlogPosts().map((post) => (
        <BlogLink key={post.slug} post={post} />
      ))}
    </Stack>
  )
}
