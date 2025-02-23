import type { NextConfig } from 'next'
import createMDX from '@next/mdx'
import rehypeHighlight from 'rehype-highlight'
import remarkFrontmatter from 'remark-frontmatter'
import remarkMdxFrontmatter from 'remark-mdx-frontmatter'

const nextConfig: NextConfig = {
  pageExtensions: ['md', 'mdx', 'ts', 'tsx']
}

const withMDX = createMDX({
  options: {
    rehypePlugins: [rehypeHighlight],
    remarkPlugins: [
      remarkFrontmatter,
      [remarkMdxFrontmatter, { name: 'matter' }]
    ]
  }
})

export default withMDX(nextConfig)
