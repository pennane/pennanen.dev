import type { GetStaticPaths, GetStaticPropsContext, InferGetStaticPropsType } from 'next'
import Layout from '../components/layout'
import style from '../styles/project.module.css'
import linkStyle from '../styles/project-link.module.css'
import { getProjectById, getProjectIds } from '../lib/stuff'
import { isString } from '../lib/util'
import { Project } from '../types'

export const getStaticPaths: GetStaticPaths = async () => {
    const paths = await getProjectIds()
    return {
        paths,
        fallback: false
    }
}

export const getStaticProps = async ({ params }: GetStaticPropsContext) => {
    if (!params || !isString(params.project)) {
        return {
            redirect: {
                destination: '/',
                permanent: false
            }
        }
    }

    const project = await getProjectById(params.project)

    if (!project) {
        return {
            redirect: {
                destination: '/',
                permanent: false
            }
        }
    }

    return {
        props: {
            project
        }
    }
}

const Index = (context: InferGetStaticPropsType<typeof getStaticProps>) => {
    const project = context.project as Project
    if (!project) return null

    return (
        <Layout>
            <div className={style.project}>
                <header className={style['header']}>
                    <h1>{project.name}</h1>
                    <p>{project.description}</p>
                </header>
                <section className={style['lift']}>
                    {project.icon && (
                        <div>
                            <img className={linkStyle['image']} src={`/stuff/${project.id}/${project.icon}`} />
                        </div>
                    )}
                    {project.url && project.url.includes('github') && (
                        <a href={project.url} rel="noreferrer">
                            outbound github link
                        </a>
                    )}
                    {project.url && !project.url.includes('github') && (
                        <a href={project.url} rel="noreferrer">
                            outbound link
                        </a>
                    )}
                    {!project.url && <a href={`/stuff/${project.id}/index.html`}>link</a>}
                </section>
            </div>
        </Layout>
    )
}

export default Index
