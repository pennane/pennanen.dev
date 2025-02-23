import { getBlogPosts } from '../lib'

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
  return getBlogPosts().map(({ slug }) => ({ slug }))
}

export const dynamicParams = false
