import { Metadata } from 'next'
import Image from 'next/image'
import { A } from '../../components/A'
import { JsonLd } from '../../components/JsonLd'
import { Stack } from '../../components/Stack'
import { generateProjectImage } from '../../meta/image'
import { formatDate, parseDateString } from '../lib'
import { baseUrl } from '../sitemap'
import { getProjectById, getProjects } from './lib'
import { generateProjectJsonLd } from '../../lib/json-ld'
import styles from './page.module.css'

export async function generateMetadata({
  params
}: {
  params: Promise<{ project: string }>
}): Promise<Metadata | undefined> {
  const id = (await params).project
  const project = getProjectById(id)

  if (!project) {
    return
  }
  const ogImage = await generateProjectImage(project)
  const title = 'Arttu Pennanen - ' + project.name
  const description = project.description
  const publishedTime = parseDateString(project.date)?.toISOString()

  return {
    title,
    description,
    openGraph: {
      title,
      description,
      type: 'article',
      publishedTime,
      url: `${baseUrl}/${project.id}`,
      images: [
        {
          url: `${baseUrl}/meta/${ogImage}`
        }
      ]
    },
    twitter: {
      card: 'summary_large_image',
      title,
      description,
      images: [`${baseUrl}/meta/${ogImage}`]
    }
  }
}

export default async function Page({
  params
}: {
  params: Promise<{ project: string }>
}) {
  const id = (await params).project
  const project = getProjectById(id)!
  const date = project.date ? new Date(project.date) : null
  const jsonld = generateProjectJsonLd(project)

  return (
    <Stack>
      <header>
        <Stack>
          <Stack>
            <h1>{project.name}</h1>
            {date && <time>{formatDate(date)}</time>}
          </Stack>
          <Stack horizontal alignItems="center">
            <div>
              <Image
                width={48}
                height={48}
                alt={project.icon ? `${project.name} icon` : ''}
                src={
                  project.icon
                    ? '/sub/' + project.id + '/' + project.icon
                    : '/images/placeholder.png'
                }
              />
            </div>
            {project.url && project.url.includes('github.com') && (
              <A href={project.url} rel="noreferrer">
                View source <span>[{project.url}]</span>
              </A>
            )}
            {project.url && project.url.includes('pennanen.dev') && (
              <A href={project.url} rel="noreferrer">
                Open project <span>[{project.url}]</span>
              </A>
            )}
            {project.url &&
              !project.url.includes('github.com') &&
              !project.url.includes('pennanen.dev') && (
                <A href={project.url} rel="noreferrer">
                  Open outbound link <span>[{project.url}]</span>
                </A>
              )}
            {!project.url && (
              <A href={`/sub/${project.id}/index.html`}>Launch project</A>
            )}
          </Stack>
          <p>{project.description}</p>
        </Stack>
      </header>

      {project.pretext && (
        <Stack>
          <p
            dangerouslySetInnerHTML={{
              __html: project.pretext
            }}
          ></p>
        </Stack>
      )}
      {project.images[0] && (
        <Stack alignItems="stretch">
          {project.images.map((url, i) => (
            <div key={project.id + i} className={styles.image}>
              {url.endsWith('.mp4') && (
                <video
                  src={'/sub/' + project.id + '/' + url}
                  muted={true}
                  autoPlay={true}
                  loop={true}
                />
              )}
              {!url.endsWith('.mp4') && url.includes('://') && (
                <Image
                  src={url}
                  alt=""
                  loading="lazy"
                  fill
                  objectFit="contain"
                />
              )}
              {!url.endsWith('.mp4') && !url.includes('://') && (
                <Image
                  src={'/sub/' + project.id + '/' + url}
                  alt=""
                  loading="lazy"
                  fill
                  objectFit="contain"
                />
              )}
            </div>
          ))}
        </Stack>
      )}
      {project.github && (
        <Stack>
          <A href={project.github} rel="noreferrer">
            View source
          </A>
        </Stack>
      )}
      <JsonLd object={jsonld} />
    </Stack>
  )
}

export function generateStaticParams() {
  return getProjects().map(({ id }) => ({ project: id }))
}

export const dynamicParams = false
