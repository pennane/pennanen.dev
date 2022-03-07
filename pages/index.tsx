import type { InferGetStaticPropsType } from 'next'
import Layout from '../components/layout'
import MainHeading from '../components/main-header'
import Projects from '../components/projects'
import { generateMainPageImage } from '../lib/meta-image'
import { getFilteredProjects } from '../lib/stuff'
import { Project } from '../types'

const description = 'Arttu Pennanen Web-like projects'

export const getStaticProps = async () => {
    const projectsData: Project[] = getFilteredProjects()
    const metaImage = await generateMainPageImage({ title: 'pennanen.dev', description })
    return {
        props: {
            projectsData,
            metaImage
        }
    }
}

const Index = ({ projectsData, metaImage }: InferGetStaticPropsType<typeof getStaticProps>) => {
    return (
        <Layout description={description} metaImage={metaImage}>
            <header>
                <MainHeading />
            </header>
            <section>
                <Projects projects={projectsData} />
            </section>
        </Layout>
    )
}

export default Index
