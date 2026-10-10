import fs from 'fs'
import path from 'path'
import { z } from 'zod'

const PostMetadataSchema = z.object({
  title: z.string().min(1),
  date: z.string().regex(/^\d{1,2}\.\d{1,2}\.\d{4}$/, 'Expected dd.mm.yyyy'),
  summary: z.string().min(1),
  draft: z
    .enum(['true', 'false'])
    .optional()
    .transform((v) => v === 'true')
})

type Metadata = z.infer<typeof PostMetadataSchema>

export type Post = {
  metadata: Metadata
  slug: string
  content: string
}

export function parseFrontmatterContent(frontMatterBlock: string): Metadata {
  const fields: Record<string, string> = {}

  for (const line of frontMatterBlock.trim().split('\n')) {
    const [key, ...valueArr] = line.split(': ')
    fields[key.trim()] = valueArr
      .join(': ')
      .trim()
      .replace(/^['"](.*)['"]$/, '$1')
  }

  return PostMetadataSchema.parse(fields)
}

function parseFrontmatter(fileContent: string) {
  const frontmatterRegex = /---\s*([\s\S]*?)\s*---/
  const match = frontmatterRegex.exec(fileContent)
  const frontMatterBlock = match![1]
  const content = fileContent.replace(frontmatterRegex, '').trim()
  const metadata = parseFrontmatterContent(frontMatterBlock)

  return { metadata, content }
}

function getMdxFiles(dir: string) {
  return fs.readdirSync(dir).filter((file) => path.extname(file) === '.mdx')
}

function readMdxFile(filePath: string) {
  const rawContent = fs.readFileSync(filePath, 'utf-8')
  try {
    return parseFrontmatter(rawContent)
  } catch (error) {
    throw new Error(`Invalid frontmatter in ${filePath}`, { cause: error })
  }
}

function getMdxData(dir: string) {
  const mdxFiles = getMdxFiles(dir)
  return mdxFiles.map((file): Post => {
    const { metadata, content } = readMdxFile(path.join(dir, file))
    const slug = path.basename(file, path.extname(file))

    return {
      metadata,
      slug,
      content
    }
  })
}

export function getBlogPosts() {
  return getMdxData(path.join(process.cwd(), 'src', 'content')).filter(
    (post) => !post.metadata.draft
  )
}

export function getBlogPostsWithDrafts() {
  return getMdxData(path.join(process.cwd(), 'src', 'content'))
}
