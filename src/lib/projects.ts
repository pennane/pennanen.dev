import fs from 'fs'
import type { StaticImageData } from 'next/image'
import path from 'path'
import { parse } from 'yaml'
import { z } from 'zod'

const PROJECTS_PATH = path.join(process.cwd(), 'src', 'content', 'projects')
const PUBLIC_PATH = path.join(process.cwd(), 'public')
const FAVICON_PATH = path.join(PUBLIC_PATH, 'meta', 'icons')
const PLACEHOLDER_ICON = '/images/placeholder.png'

const LinkSchema = z
  .string()
  .refine(
    (value) => /^https?:\/\//.test(value) || value.startsWith('/'),
    'Expected a full URL or a /path'
  )

const ProjectFrontmatterSchema = z
  .object({
    title: z.string().min(1),
    summary: z.string().min(1).optional(),
    date: z
      .string()
      .regex(/^\d{4}-\d{2}-\d{2}$/, 'Expected YYYY-MM-DD')
      .transform((value) => {
        const [year, month, day] = value.split('-').map(Number)
        return new Date(year, month - 1, day)
      })
      .optional(),
    featured: z.boolean().default(false),
    links: z
      .object({ live: LinkSchema.optional(), source: LinkSchema.optional() })
      .strict()
      .default({}),
    icon: z.string().min(1).optional(),
    cover: z.string().min(1).optional(),
    notAuthor: z.boolean().default(false)
  })
  .strict()

export type ImageSource = StaticImageData | string

export type Project = {
  id: string
  title: string
  summary?: string
  date?: Date
  featured: boolean
  links: { live?: string; source?: string }
  icon: ImageSource
  cover?: { src: ImageSource; filePath: string }
  notAuthor: boolean
}

const readFrontmatter = (id: string) => {
  const filePath = path.join(PROJECTS_PATH, id, 'index.mdx')
  const match = /^---\n([\s\S]*?)\n---/.exec(fs.readFileSync(filePath, 'utf-8'))
  const result = ProjectFrontmatterSchema.safeParse(parse(match?.[1] ?? ''))
  if (!result.success) {
    const issues = result.error.issues
      .map((issue) => `${issue.path.join('.') || '(root)'}: ${issue.message}`)
      .join('; ')
    throw new Error(`Invalid frontmatter in ${filePath}: ${issues}`)
  }
  return result.data
}

const resolveImage = async (id: string, file: string) => {
  const isPublic = file.startsWith('/')
  const name = file.replace(/^\.\//, '')
  const filePath = isPublic
    ? path.join(PUBLIC_PATH, file)
    : path.join(PROJECTS_PATH, id, name)

  if (!fs.existsSync(filePath)) {
    throw new Error(
      `Missing image "${file}" in ${path.join(PROJECTS_PATH, id, 'index.mdx')}`
    )
  }

  const src: ImageSource = isPublic
    ? file
    : (await import(`@/content/projects/${id}/${name}`)).default
  return { src, filePath }
}

const FAVICON_TYPES: Record<string, string> = {
  'image/png': 'png',
  'image/svg+xml': 'svg',
  'image/x-icon': 'ico',
  'image/vnd.microsoft.icon': 'ico',
  'image/jpeg': 'jpg',
  'image/gif': 'gif',
  'image/webp': 'webp'
}

const fetchFavicon = async (id: string, pageUrl: string) => {
  try {
    const page = await fetch(pageUrl, { signal: AbortSignal.timeout(3000) })
    const html = page.ok ? await page.text() : ''
    const tag = /<link[^>]+rel=["'][^"']*icon[^"']*["'][^>]*>/i.exec(html)?.[0]
    const href = tag && /href=["']([^"']+)["']/i.exec(tag)?.[1]

    const icon = await fetch(new URL(href || '/favicon.ico', pageUrl), {
      signal: AbortSignal.timeout(3000)
    })
    const type = icon.headers.get('content-type')?.split(';')[0].trim() ?? ''
    const extension = FAVICON_TYPES[type]
    if (!icon.ok || !extension) return undefined

    fs.mkdirSync(FAVICON_PATH, { recursive: true })
    const fileName = `${id}.${extension}`
    fs.writeFileSync(
      path.join(FAVICON_PATH, fileName),
      Buffer.from(await icon.arrayBuffer())
    )
    return `/meta/icons/${fileName}`
  } catch {
    return undefined
  }
}

const resolveIcon = async (
  id: string,
  icon: string | undefined,
  live: string | undefined
): Promise<ImageSource> => {
  if (icon) return (await resolveImage(id, icon)).src
  if (live?.startsWith('http')) {
    return (await fetchFavicon(id, live)) ?? PLACEHOLDER_ICON
  }
  return PLACEHOLDER_ICON
}

const loadProject = async (id: string): Promise<Project> => {
  const { icon, cover, ...frontmatter } = readFrontmatter(id)
  return {
    id,
    ...frontmatter,
    icon: await resolveIcon(id, icon, frontmatter.links.live),
    cover: cover ? await resolveImage(id, cover) : undefined
  }
}

const loadProjects = async () => {
  const ids = fs
    .readdirSync(PROJECTS_PATH, { withFileTypes: true })
    .filter((entry) => entry.isDirectory() && !entry.name.startsWith('_'))
    .map((entry) => entry.name)

  const projects = await Promise.all(ids.map(loadProject))
  return projects.sort((a, b) => {
    if (!a.date) return 1
    if (!b.date) return -1
    return b.date.getTime() - a.date.getTime()
  })
}

// Builds call this from many pages; development re-reads so edits show up
let cached: Promise<Project[]> | undefined

export function getProjects() {
  if (process.env.NODE_ENV !== 'production') return loadProjects()
  cached ??= loadProjects()
  return cached
}

export async function getFeaturedProjects() {
  return (await getProjects()).filter((project) => project.featured)
}

export async function getProjectById(id: string) {
  return (await getProjects()).find((project) => project.id === id)
}
