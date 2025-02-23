import type { NextConfig } from 'next'
import createMDX from '@next/mdx'
import rehypeHighlight from 'rehype-highlight'
import remarkFrontmatter from 'remark-frontmatter'
import remarkMdxFrontmatter from 'remark-mdx-frontmatter'
import rehypeAutolinkHeadings from 'rehype-autolink-headings'
import rehypeSlug from 'rehype-slug'

const nextConfig: NextConfig = {
  pageExtensions: ['md', 'mdx', 'ts', 'tsx']
}

const withMDX = createMDX({
  options: {
    rehypePlugins: [
      rehypeHighlight,
      rehypeSlug,
      [rehypeAutolinkHeadings, { behavior: 'append' }]
    ],
    remarkPlugins: [
      remarkFrontmatter,
      [remarkMdxFrontmatter, { name: 'matter' }]
    ]
  }
})

export default withMDX(nextConfig)
