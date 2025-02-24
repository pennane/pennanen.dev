import fs from 'fs'
import path from 'path'

type Metadata = {
  title?: string
  date?: string
  summary?: string
}

export type Post = {
  metadata: Metadata
  slug: string
  content: string
}

export function parseFrontmatterContent(frontMatterBlock: string): Metadata {
  const frontMatterLines = frontMatterBlock.trim().split('\n')
  const metadata: Metadata = {}

  for (const line of frontMatterLines) {
    const [key, ...valueArr] = line.split(': ')
    let value = valueArr.join(': ').trim()
    value = value.replace(/^['"](.*)['"]$/, '$1')
    metadata[key.trim() as keyof Metadata] = value
  }

  return metadata
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
  return parseFrontmatter(rawContent)
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
  return getMdxData(path.join(process.cwd(), 'src', 'content'))
}
