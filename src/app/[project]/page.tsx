import { Code, Globe, LucideIcon } from 'lucide-react'
import { Metadata } from 'next'
import Image from 'next/image'
import { A } from '../../components/A'
import { JsonLd } from '../../components/JsonLd'
import { Stack } from '../../components/Stack'
import { isUnoptimizable } from '../../lib/images'
import { generateProjectJsonLd } from '../../lib/json-ld'
import { getProjectById, getProjects } from '../../lib/projects'
import { generateProjectImage } from '../../meta/image'
import { formatDate } from '../lib'
import { baseUrl } from '../sitemap'
import styles from './page.module.css'

type Params = { params: Promise<{ project: string }> }

export async function generateMetadata({
  params
}: Params): Promise<Metadata | undefined> {
  const project = await getProjectById((await params).project)

  if (!project) {
    return
  }
  const ogImage = `${baseUrl}/meta/${await generateProjectImage(project)}`
  const title = 'Arttu Pennanen - ' + project.title
  const description = project.summary

  return {
    title,
    description,
    openGraph: {
      title,
      description,
      type: 'article',
      publishedTime: project.date?.toISOString(),
      url: `${baseUrl}/${project.id}`,
      images: [{ url: ogImage }]
    },
    twitter: {
      card: 'summary_large_image',
      title,
      description,
      images: [ogImage]
    }
  }
}

const linkText = (href: string) => {
  if (!href.includes('://')) {
    return `pennanen.dev${href.replace(/\/index\.html$/, '')}`
  }
  const url = new URL(href)
  const host = url.host.replace(/^www\./, '')
  const pathname = url.pathname.replace(/\/$/, '')
  return host === 'github.com' ? pathname.slice(1) : host + pathname
}

const MetaLink = ({ href, icon: Icon }: { href: string; icon: LucideIcon }) => {
  const external = href.includes('://')
  return (
    <span className={styles.metaItem}>
      <Icon size={14} aria-hidden />
      <A
        href={href}
        {...(external && { rel: 'noreferrer', target: '_blank' })}
      >
        {linkText(href)}
        {external && (
          <span className={styles.arrow} aria-hidden>
            ↗
          </span>
        )}
      </A>
    </span>
  )
}

export default async function Page({ params }: Params) {
  const id = (await params).project
  const project = (await getProjectById(id))!
  const { default: Content } = await import(
    /* webpackExclude: /[\\/]_template[\\/]/ */ `@/content/projects/${id}/index.mdx`
  )

  return (
    <Stack>
      <header>
        <Stack>
          <Stack horizontal alignItems="center">
            <Image
              className={styles.icon}
              width={44}
              height={44}
              alt=""
              src={project.icon}
              unoptimized={isUnoptimizable(project.icon)}
            />
            <h1>{project.title}</h1>
          </Stack>
          {(project.date || project.links.live || project.links.source) && (
            <p className={styles.meta}>
              {project.date && <time>{formatDate(project.date)}</time>}
              {project.links.live && (
                <MetaLink href={project.links.live} icon={Globe} />
              )}
              {project.links.source && (
                <MetaLink href={project.links.source} icon={Code} />
              )}
            </p>
          )}
          {project.summary && <p>{project.summary}</p>}
        </Stack>
      </header>

      <div className={styles.content}>
        <Content />
      </div>
      <JsonLd object={generateProjectJsonLd(project)} />
    </Stack>
  )
}

export async function generateStaticParams() {
  return (await getProjects()).map(({ id }) => ({ project: id }))
}

export const dynamicParams = false
