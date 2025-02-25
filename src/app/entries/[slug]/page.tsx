import { Metadata } from 'next'
import { generatePostImage } from '../../../meta/image'
import { parseDateString } from '../../lib'
import { baseUrl } from '../../sitemap'
import { getBlogPostsWithDrafts } from '../lib'

export async function generateMetadata({
  params
}: {
  params: Promise<{ slug: string }>
}): Promise<Metadata | undefined> {
  const slug = (await params).slug
  const post = getBlogPostsWithDrafts().find((post) => post.slug === slug)

  if (!post) {
    return
  }

  const ogImage = await generatePostImage(post)
  const title = post.metadata.title
  const description = post.metadata.summary
  const publishedTime = parseDateString(post.metadata.date)?.toISOString()

  return {
    title,
    description,
    openGraph: {
      title,
      description,
      type: 'article',
      publishedTime,
      url: `${baseUrl}/entries/${post.slug}`,
      images: [
        {
          url: `${baseUrl}/meta/${ogImage}`
        }
      ]
    },
    twitter: {
      card: 'summary_large_image',
      title,
      description,
      images: [`${baseUrl}/meta/${ogImage}`]
    }
  }
}

export default async function Page({
  params
}: {
  params: Promise<{ slug: string }>
}) {
  const slug = (await params).slug
  const { default: Post } = await import(`@/content/${slug}.mdx`)

  return <Post />
}

export function generateStaticParams() {
  return getBlogPostsWithDrafts().map(({ slug }) => ({ slug }))
}

export const dynamicParams = false
