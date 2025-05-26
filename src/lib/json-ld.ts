import {
  BlogPosting,
  CreativeWork,
  Graph,
  Person,
  SoftwareSourceCode,
  WebPage,
  WebSite,
  WithContext
} from 'schema-dts'
import { Project } from '../app/[project]/lib'
import { Post } from '../app/entries/lib'
import { parseDateString } from '../app/lib'
import { baseUrl } from '../app/sitemap'
import { getFullPostImageUrl } from '../meta/image'

const createAuthor = (): Person => ({
  '@type': 'Person',
  name: 'Arttu Pennanen'
})

const createWebSite = (): WebSite => ({
  '@type': 'WebSite',
  name: 'pennanen.dev',
  url: baseUrl
})

const createIsPartOf = () => ({
  '@type': 'CreativeWork' as const,
  name: 'pennanen.dev'
})

const createSoftwareSourceCode = (project: Project): SoftwareSourceCode => ({
  '@type': 'SoftwareSourceCode',
  name: project.name,
  description: project.description || '',
  codeRepository: project.github,
  url: `${baseUrl}/${project.id}`,
  image: project.largeImage
    ? `${baseUrl}/sub/${project.id}/${project.largeImage}`
    : undefined,
  author: createAuthor(),
  dateCreated: project.ignoreDate
    ? undefined
    : new Date(project.date!).toISOString(),
  isPartOf: createIsPartOf()
})

const createCreativeWork = (project: Project): CreativeWork => ({
  '@type': 'CreativeWork',
  name: project.name,
  description: project.description || '',
  url: `${baseUrl}/${project.id}`,
  image: project.largeImage
    ? `${baseUrl}/sub/${project.id}/${project.largeImage}`
    : undefined,
  author: createAuthor(),
  dateCreated: parseDateString(project.date)?.toISOString(),
  isPartOf: createIsPartOf()
})

const createBlogPosting = (post: Post): BlogPosting => ({
  '@type': 'BlogPosting',
  headline: post.metadata.title,
  description: post.metadata.summary,
  datePublished: parseDateString(post.metadata.date)?.toISOString(),
  author: createAuthor(),
  image: getFullPostImageUrl(post),
  url: `${baseUrl}/entries/${post.slug}`,
  mainEntityOfPage: {
    '@type': 'WebPage',
    '@id': `${baseUrl}/entries/${post.slug}`
  },
  isPartOf: createIsPartOf()
})

const createWebPage = (
  url: string,
  title: string,
  description?: string
): WebPage => ({
  '@type': 'WebPage',
  '@id': url,
  name: title,
  description,
  url,
  isPartOf: createWebSite(),
  author: createAuthor()
})

const selectProjectSchema = (project: Project) =>
  project.github
    ? createSoftwareSourceCode(project)
    : createCreativeWork(project)

const projectToJsonLd = (
  project: Project
): WithContext<SoftwareSourceCode | CreativeWork> =>
  ({
    '@context': 'https://schema.org',
    ...selectProjectSchema(project)
  } as WithContext<SoftwareSourceCode | CreativeWork>)

const postToJsonLd = (post: Post): WithContext<BlogPosting> =>
  ({
    '@context': 'https://schema.org',
    ...createBlogPosting(post)
  } as WithContext<BlogPosting>)

const buildGraphItems = (projects: Project[], posts: Post[]) => [
  ...projects.map(selectProjectSchema),
  ...posts.map(createBlogPosting)
]

const createGraph = (
  pageUrl: string,
  pageTitle: string,
  pageDescription?: string,
  projects: Project[] = [],
  posts: Post[] = []
): Graph => ({
  '@context': 'https://schema.org',
  '@graph': [
    createWebPage(pageUrl, pageTitle, pageDescription),
    ...buildGraphItems(projects, posts)
  ]
})

const createHomepageGraph = (
  featuredProjects: Project[],
  featuredPosts: Post[]
): Graph =>
  createGraph(
    baseUrl,
    'Arttu Pennanen - Software Engineer',
    'Digital playground pennanen.dev - sharing some of personal projects and writings.',
    featuredProjects,
    featuredPosts
  )

const createProjectsListingGraph = (projects: Project[]): Graph =>
  createGraph(
    `${baseUrl}/projects`,
    'Software Projects',
    undefined,
    projects,
    []
  )

const createEntriesListingGraph = (posts: Post[]): Graph =>
  createGraph(`${baseUrl}/entries`, 'Entries', undefined, [], posts)

const createUncuratedProjectsGraph = (projects: Project[]): Graph =>
  createGraph(
    `${baseUrl}/uncurated`,
    'Uncurated Projects',
    undefined,
    projects,
    []
  )

export const generateProjectJsonLd = projectToJsonLd
export const generatePostJsonLd = postToJsonLd
export const generateHomepageJsonLd = createHomepageGraph
export const generateProjectsListingJsonLd = createProjectsListingGraph
export const generateEntriesListingJsonLd = createEntriesListingGraph
export const generateUncuratedProjectsJsonLd = createUncuratedProjectsGraph
