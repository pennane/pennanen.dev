import type { InferGetStaticPropsType } from 'next'
import Layout from '../components/layout'
import Projects from '../components/projects'
import { generateMainPageImage } from '../lib/meta-image'
import { getProjects } from '../lib/stuff'
import { Project } from '../types'

const description = 'Every single web-project. Unfiltered and uncurated. Functionality not guaranteed.'

export const getStaticProps = async () => {
    const projectsData: Project[] = getProjects()
    const metaImage = await generateMainPageImage({ title: 'All | pennanen.dev', description })
    return {
        props: {
            projectsData,
            metaImage
        }
    }
}

const All = ({ projectsData, metaImage }: InferGetStaticPropsType<typeof getStaticProps>) => {
    return (
        <Layout description={description} metaImage={metaImage}>
            <section className="margin-top">
                <Projects projects={[...projectsData].reverse()} />
            </section>
        </Layout>
    )
}

export default All
