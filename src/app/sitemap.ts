import { getProjects } from './[project]/lib'
import { getBlogPosts } from './entries/lib'

export const baseUrl = 'https://pennanen.dev'

export default async function sitemap() {
  const entries = getBlogPosts().map((post) => ({
    url: `${baseUrl}/entries/${post.slug}`,
    lastModified: post.metadata.publishedAt || post.metadata.date
  }))

  const projects = getProjects().map((project) => ({
    url: `${baseUrl}/${project.id}`,
    lastModified: project.date
      ? new Date(project.date).toISOString().split('T')[0]
      : undefined
  }))

  const routes = ['', '/blog'].map((route) => ({
    url: `${baseUrl}${route}`,
    lastModified: new Date().toISOString().split('T')[0]
  }))

  return [...routes, ...entries, ...projects]
}
