import { getBlogPosts } from './lib'
import styles from './page.module.css'
import { Stack } from '../../components/Stack'
import { BlogLink } from './components/BlogLink'
import { ItemGroups } from '../../components/ProjectGroups'

export default function Blog() {
  return (
    <Stack className={styles.page}>
      <ItemGroups
        items={getBlogPosts()}
        getDate={(x) => x.metadata.date}
        render={(post) => <BlogLink post={post} key={post.slug} />}
      />
    </Stack>
  )
}
