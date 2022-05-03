// noinspection JSUnusedGlobalSymbols

import type { GetStaticPaths, GetStaticProps, GetStaticPropsContext } from 'next'
import Layout from '../components/Layout'
import { getProjectById, getProjectIds } from '../lib/stuff'
import { isString } from '../lib/util'
import { ProjectInterface } from '../types'
import { generateProjectImage } from '../lib/meta-image'
import Project from '../components/Project'

export const getStaticPaths: GetStaticPaths = async () => {
    const paths = getProjectIds()
    return {
        paths,
        fallback: false
    }
}

export const getStaticProps: GetStaticProps = async ({ params }: GetStaticPropsContext) => {
    if (!params || !isString(params.project)) return { notFound: true }

    const project = getProjectById(params.project)

    if (!project) return { notFound: true }

    const metaImage = await generateProjectImage(project)

    return { props: { project, metaImage } }
}

const ProjectPage = ({ project, metaImage }: { project: ProjectInterface; metaImage: string }) => {
    if (!project) return null
    project = project as ProjectInterface
    return (
        <Layout title={project.name} description={project.description || undefined} metaImage={metaImage || undefined}>
            <Project project={project} />
        </Layout>
    )
}

export default ProjectPage
