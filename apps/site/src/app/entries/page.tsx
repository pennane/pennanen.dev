import { JsonLd } from '../../components/JsonLd'
import { generateEntriesListingJsonLd } from '../../lib/json-ld'
import { getBlogPosts } from './lib'
import styles from './page.module.css'
import { Stack } from '../../components/Stack'
import { BlogLink } from './components/BlogLink'
import { ItemGroups } from '../../components/ProjectGroups'

export default function Blog() {
  const posts = getBlogPosts()
  const jsonLd = generateEntriesListingJsonLd(posts)

  return (
    <Stack className={styles.page}>
      <ItemGroups
        items={posts}
        getDate={(x) => x.metadata.date}
        render={(post) => <BlogLink post={post} key={post.slug} />}
      />
      <JsonLd object={jsonLd} />
    </Stack>
  )
}
