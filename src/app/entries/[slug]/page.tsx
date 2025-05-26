import { Metadata } from 'next'
import { BlogPosting, WithContext } from 'schema-dts'
import { JsonLd } from '../../../components/JsonLd'
import { getFullPostImageUrl } from '../../../meta/image'
import { parseDateString } from '../../lib'
import { baseUrl } from '../../sitemap'
import { getBlogPostsWithDrafts, Post } from '../lib'

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

  const ogImage = getFullPostImageUrl(post)
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
          url: ogImage
        }
      ]
    },
    twitter: {
      card: 'summary_large_image',
      title,
      description,
      images: [ogImage]
    }
  }
}

export const generatePostJsonLd = (post: Post) => {
  const datePublished = parseDateString(post.metadata.date)?.toISOString()
  const jsonLd: WithContext<BlogPosting> = {
    '@context': 'https://schema.org',
    '@type': 'BlogPosting',
    headline: post.metadata.title,
    description: post.metadata.summary,
    datePublished,
    author: {
      '@type': 'Person',
      name: 'Arttu Pennanen'
    },
    image: getFullPostImageUrl(post),
    url: `${baseUrl}/entries/${post.slug}`,
    mainEntityOfPage: {
      '@type': 'WebPage',
      '@id': `${baseUrl}/entries/${post.slug}`
    },
    isPartOf: {
      '@type': 'CreativeWork',
      name: 'pennanen.dev'
    }
  }
  return jsonLd
}

export default async function Page({
  params
}: {
  params: Promise<{ slug: string }>
}) {
  const slug = (await params).slug
  const { default: Post } = await import(`@/content/${slug}.mdx`)
  const post = getBlogPostsWithDrafts().find((p) => p.slug === slug)!
  const jsonLd = generatePostJsonLd(post)

  return (
    <article>
      <Post />
      {jsonLd && <JsonLd object={jsonLd} />}
    </article>
  )
}

export function generateStaticParams() {
  return getBlogPostsWithDrafts().map(({ slug }) => ({ slug }))
}

export const dynamicParams = false
