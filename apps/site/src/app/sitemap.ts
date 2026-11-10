import { itemPath, PUBLIC_KINDS } from '../lib/log/present'
import { logPublic } from '../lib/log/public-api'
import { getProjects } from '../lib/projects'
import { getBlogPosts } from './(site)/entries/lib'
import { parseDateString } from './(site)/lib'

export const baseUrl = 'https://pennanen.dev'

const formatDateForSitemap = (date: Date | null | undefined) => {
  if (!date) return undefined
  return date.toISOString().split('T')[0]
}

export default async function sitemap() {
  const entries = getBlogPosts().map((post) => ({
    url: `${baseUrl}/entries/${post.slug}`,
    lastModified: formatDateForSitemap(parseDateString(post.metadata.date))
  }))

  const projects = (await getProjects()).map((project) => ({
    url: `${baseUrl}/${project.id}`,
    lastModified: formatDateForSitemap(project.date)
  }))

  const logItems = (
    await Promise.all(PUBLIC_KINDS.map((kind) => logPublic.items(kind)))
  )
    .flat()
    .map((item) => ({
      url: `${baseUrl}${itemPath(item)}`,
      lastModified: formatDateForSitemap(new Date(item.updatedAt))
    }))

  const logRoutes = ['/log', ...PUBLIC_KINDS.map((kind) => `/log/${kind}`)]

  const routes = [
    '',
    '/uncurated',
    '/entries',
    '/projects',
    ...logRoutes
  ].map((route) => ({
    url: `${baseUrl}${route}`,
    lastModified: formatDateForSitemap(new Date())
  }))

  return [...routes, ...entries, ...projects, ...logItems]
}
