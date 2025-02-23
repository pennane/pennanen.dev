import Link from 'next/link'
import { getBlogPosts } from './lib'
import styles from './page.module.css'
import { Stack } from '../../components/Stack'

export default function Blog() {
  return (
    <Stack className={styles.page}>
      blog posts, bro:
      <Stack>
        {getBlogPosts().map((post) => (
          <Link href={`/blog/${post.slug}`} key={post.slug}>
            {post.metadata.title}
          </Link>
        ))}
      </Stack>
    </Stack>
  )
}
