// noinspection JSUnusedGlobalSymbols

import type { GetStaticPaths, GetStaticProps, GetStaticPropsContext } from 'next'
import Image from 'next/image'
import Layout from '../components/layout'
import style from '../styles/project.module.css'
import linkStyle from '../styles/project-link.module.css'
import { getProjectById, getProjectIds } from '../lib/stuff'
import { isString, monthIndexToName } from '../lib/util'
import { Project } from '../types'

export const getStaticPaths: GetStaticPaths = async () => {
    const paths = getProjectIds()
    return {
        paths,
        fallback: false
    }
}

export const getStaticProps: GetStaticProps = ({ params }: GetStaticPropsContext) => {
    if (!params || !isString(params.project)) return { notFound: true }

    const project = getProjectById(params.project)

    if (!project) return { notFound: true }

    return { props: { project } }
}

const ProjectPage = ({ project }: { project: Project }) => {
    if (!project) return null
    project = project as Project

    const date = project.date ? new Date(project.date) : null

    return (
        <Layout title={project.name} description={project.description || undefined}>
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
                            src={project.icon ? '/sub/' + project.id + '/' + project.icon : '/images/placeholder.png'}
                        />
                    </div>
                    {project.url && project.url.includes('github') && (
                        <a href={project.url} rel="noreferrer">
                            View source <span className={linkStyle['date']}>[{project.url}]</span>
                        </a>
                    )}
                    {project.url && project.url.includes('pennanen.dev') && (
                        <a href={project.url} rel="noreferrer">
                            Launch project <span className={linkStyle['date']}>[{project.url}]</span>
                        </a>
                    )}
                    {project.url && !project.url.includes('github') && !project.url.includes('pennanen.dev') && (
                        <a href={project.url} rel="noreferrer">
                            Open outbound link <span className={linkStyle['date']}>[{project.url}]</span>
                        </a>
                    )}
                    {!project.url && <a href={`/sub/${project.id}/index.html`}>Launch project</a>}
                </section>
                {project.pretext && (
                    <section>
                        <p>{project.pretext}</p>
                    </section>
                )}
                {project.images[0] && (
                    <section>
                        {'asdffdas'}
                        {project.images.map((url, i) => (
                            <div className={style['insert']} key={project.id + i}>
                                {url.includes('://') && <img src={url} alt="" />}
                                {!url.includes('://') && (
                                    <Image src={'/sub/' + project.id + '/' + url} alt="" layout="fill" />
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
        </Layout>
    )
}

export default ProjectPage
