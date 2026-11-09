import type { NextConfig } from 'next'
import createMDX from '@next/mdx'
import rehypeHighlight from 'rehype-highlight'
import remarkFrontmatter from 'remark-frontmatter'
import remarkMdxFrontmatter from 'remark-mdx-frontmatter'
import rehypeAutolinkHeadings from 'rehype-autolink-headings'
import rehypeSlug from 'rehype-slug'
import remarkMdxImages from 'remark-mdx-images'

const nextConfig: NextConfig = {
  pageExtensions: ['md', 'mdx', 'ts', 'tsx'],
  webpack: (config, { isServer, dev }) => {
    config.module.rules.push({
      test: /\.mp4$/,
      type: 'asset/resource',
      generator: {
        filename: 'static/media/[name].[hash][ext]',
        // Server builds write under .next/server, but only .next/static is served.
        // Same offsets as Next's own image loader.
        outputPath: isServer ? (dev ? '../' : '../../') : '',
        publicPath: '/_next/'
      }
    })
    return config
  }
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
      [remarkMdxFrontmatter, { name: 'matter' }],
      remarkMdxImages
    ]
  }
})

export default withMDX(nextConfig)
