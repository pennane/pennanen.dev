import { getProjects } from './[project]/lib'
import { getBlogPosts } from './entries/lib'
import { parseDateString } from './lib'

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

  const projects = getProjects().map((project) => ({
    url: `${baseUrl}/${project.id}`,
    lastModified: project.date
      ? formatDateForSitemap(new Date(project.date))
      : undefined
  }))

  const routes = ['', '/uncurated', '/entries', '/projects'].map((route) => ({
    url: `${baseUrl}${route}`,
    lastModified: formatDateForSitemap(new Date())
  }))

  return [...routes, ...entries, ...projects]
}
