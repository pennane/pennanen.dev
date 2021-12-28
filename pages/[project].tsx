import type { GetStaticPaths, GetStaticProps, GetStaticPropsContext, InferGetStaticPropsType } from 'next'
import Layout from '../components/layout'
import style from '../styles/project.module.css'
import linkStyle from '../styles/project-link.module.css'
import { getProjectById, getProjectIds } from '../lib/stuff'
import { isString, monthIndexToName } from '../lib/util'
import { Project, Project as ProjectPage } from '../types'

export const getStaticPaths: GetStaticPaths = async () => {
    const paths = getProjectIds()
    return {
        paths,
        fallback: false
    }
}

export const getStaticProps: GetStaticProps = ({ params }: GetStaticPropsContext) => {
    if (!params || !isString(params.project)) {
        return {
            notFound: true
        }
    }

    const project = getProjectById(params.project)

    if (!project) {
        return {
            notFound: true
        }
    }

    return {
        props: {
            project
        }
    }
}

const ProjectPage = ({ project }: InferGetStaticPropsType<typeof getStaticProps>) => {
    if (!project) return null
    project = project as Project

    const date = project.date ? new Date(project.date) : null

    return (
        <Layout title={project.name} description={project.description || undefined}>
            <div className={style.project}>
                <header className={style['header']}>
                    <h1>{project.name}</h1>
                    <p>{project.description}</p>
                    {date && (
                        <span className={linkStyle['date']}>{`${monthIndexToName(
                            date.getMonth()
                        )}, ${date.getFullYear()}`}</span>
                    )}
                </header>
                <section className={style['lift']}>
                    {project.icon && (
                        <div>
                            <img className={linkStyle['image']} src={`/stuff/${project.id}/${project.icon}`} />
                        </div>
                    )}
                    {project.url && project.url.includes('github') && (
                        <a href={project.url} rel="noreferrer">
                            Open github link <span className={linkStyle['date']}>[{project.url}]</span>
                        </a>
                    )}
                    {project.url && !project.url.includes('github') && (
                        <a href={project.url} rel="noreferrer">
                            Open outbound link <span className={linkStyle['date']}>[{project.url}]</span>
                        </a>
                    )}
                    {!project.url && <a href={`/stuff/${project.id}/index.html`}>Launch project</a>}
                </section>
            </div>
        </Layout>
    )
}

export default ProjectPage
