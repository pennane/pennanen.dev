import type { NextConfig } from 'next'
import createMDX from '@next/mdx'
import rehypeHighlight from 'rehype-highlight'
import remarkFrontmatter from 'remark-frontmatter'
import remarkMdxFrontmatter from 'remark-mdx-frontmatter'
import rehypeAutolinkHeadings from 'rehype-autolink-headings'
import rehypeSlug from 'rehype-slug'
import { parseFrontmatterContent } from './src/app/entries/lib'
import { formatDate, parseDateString } from './src/app/lib'
import { Heading, Parent } from 'mdast'

const nextConfig: NextConfig = {
  pageExtensions: ['md', 'mdx', 'ts', 'tsx']
}

const addTitleAndTimestamp = () => {
  return (tree: Parent) => {
    const matter = tree.children.find((node) => node.type === 'yaml')?.value

    if (!matter) return

    const metadata = parseFrontmatterContent(matter)
    const title = metadata?.title || 'Untitled'
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
      addTitleAndTimestamp
    ]
  }
})

export default withMDX(nextConfig)
