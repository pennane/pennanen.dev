import type { NextConfig } from 'next'
import path from 'path'
import createMDX from '@next/mdx'
import rehypeHighlight from 'rehype-highlight'
import remarkFrontmatter from 'remark-frontmatter'
import remarkMdxFrontmatter from 'remark-mdx-frontmatter'
import rehypeAutolinkHeadings from 'rehype-autolink-headings'
import rehypeSlug from 'rehype-slug'
import remarkMdxImages from 'remark-mdx-images'
import { parseFrontmatterContent } from './src/app/entries/lib'
import { formatDate, parseDateString } from './src/app/lib'
import { Heading, Parent } from 'mdast'

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

const addTitleAndTimestamp = () => {
  return (tree: Parent, file: { path: string }) => {
    if (file.path.includes(`${path.sep}content${path.sep}projects${path.sep}`)) {
      return
    }
    const matter = tree.children.find((node) => node.type === 'yaml')?.value

    if (!matter) return

    const metadata = parseFrontmatterContent(matter)
    const title = metadata.title
    const timestamp = formatDate(parseDateString(metadata.date))

    const titleNode: Heading = {
      type: 'heading',
      depth: 1,
      children: [
        {
          type: 'text',
          value: title
        }
      ]
    }

    const timestampNode = {
      type: 'mdxJsxFlowElement',
      name: 'time',
      attributes: [],
      children: [
        {
          type: 'text',
          value: timestamp
        }
      ]
    }

    // @ts-expect-error bad types
    tree.children.unshift(titleNode, timestampNode)
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
      remarkMdxImages,
      addTitleAndTimestamp
    ]
  }
})

export default withMDX(nextConfig)
