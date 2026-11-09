import { Code, Globe } from 'lucide-react'
import { Metadata } from 'next'
import Image from 'next/image'
import { JsonLd } from '../../components/JsonLd'
import { Lede, MetaLine, MetaLink } from '../../components/MetaLine'
import { Gap, Stack } from '../../components/Stack'
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

export default async function Page({ params }: Params) {
  const id = (await params).project
  const project = (await getProjectById(id))!
  const { default: Content } = await import(
    /* webpackExclude: /[\\/]_template[\\/]/ */ `@/content/projects/${id}/index.mdx`
  )

  return (
    <article>
      <header>
        <Stack gap={Gap.mSmall}>
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
          <Stack gap={Gap.medium}>
            {project.summary && <Lede>{project.summary}</Lede>}
            {(project.date || project.links.live || project.links.source) && (
              <MetaLine>
                {project.date && <time>{formatDate(project.date)}</time>}
                {project.links.live && (
                  <MetaLink href={project.links.live} icon={Globe} />
                )}
                {project.links.source && (
                  <MetaLink href={project.links.source} icon={Code} />
                )}
              </MetaLine>
            )}
          </Stack>
        </Stack>
      </header>
      <Content />
      <JsonLd object={generateProjectJsonLd(project)} />
    </article>
  )
}

export async function generateStaticParams() {
  return (await getProjects()).map(({ id }) => ({ project: id }))
}

export const dynamicParams = false
