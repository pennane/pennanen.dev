import { monthIndexToName } from '../../lib/util'
import style from './project.module.css'
import linkStyle from './ProjectLink/projectlink.module.css'
import { ProjectInterface } from '../../types'
import Image from 'next/image'

export default function Project({ project }: { project: ProjectInterface }) {
  const date = project.date ? new Date(project.date) : null
  return (
    <div className={style['project']}>
      <header className={style['header']}>
        <h1>{project.name}</h1>
        {date && (
          <span className={linkStyle['date']}>
            From {`${monthIndexToName(date.getMonth())}, ${date.getFullYear()}`}
          </span>
        )}
        <p>{project.description}</p>
      </header>
      <section className={style['lift']}>
        <div className={linkStyle['image']}>
          <Image
            className={linkStyle['image']}
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
          <a href={project.url} rel="noreferrer">
            View source{' '}
            <span className={linkStyle['date']}>[{project.url}]</span>
          </a>
        )}
        {project.url && project.url.includes('pennanen.dev') && (
          <a href={project.url} rel="noreferrer">
            Launch project{' '}
            <span className={linkStyle['date']}>[{project.url}]</span>
          </a>
        )}
        {project.url &&
          !project.url.includes('github.com') &&
          !project.url.includes('pennanen.dev') && (
            <a href={project.url} rel="noreferrer">
              Open outbound link{' '}
              <span className={linkStyle['date']}>[{project.url}]</span>
            </a>
          )}
        {!project.url && (
          <a href={`/sub/${project.id}/index.html`}>Launch project</a>
        )}
      </section>
      {project.pretext && (
        <section>
          <p>{project.pretext}</p>
        </section>
      )}
      {project.images[0] && (
        <section>
          {project.images.map((url, i) => (
            <div className={style['insert']} key={project.id + i}>
              {url.endsWith('.mp4') && (
                <video
                  src={'/sub/' + project.id + '/' + url}
                  muted={true}
                  autoPlay={true}
                  loop={true}
                />
              )}
              {!url.endsWith('.mp4') && url.includes('://') && (
                <img src={url} alt="" loading="lazy" />
              )}
              {!url.endsWith('.mp4') && !url.includes('://') && (
                <img
                  src={'/sub/' + project.id + '/' + url}
                  alt=""
                  loading="lazy"
                />
              )}
            </div>
          ))}
        </section>
      )}
      {project.github && (
        <section>
          <a href={project.github} rel="noreferrer">
            View source
          </a>
        </section>
      )}
    </div>
  )
}
