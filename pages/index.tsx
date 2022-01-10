import type { InferGetStaticPropsType } from 'next'
import Layout from '../components/layout'
import MainHeading from '../components/main-header'
import Projects from '../components/projects'
import { getFilteredProjects } from '../lib/stuff'
import { Project } from '../types'

export const getStaticProps = async () => {
    const projectsData: Project[] = getFilteredProjects()
    return {
        props: {
            projectsData
        }
    }
}

const Index = ({ projectsData }: InferGetStaticPropsType<typeof getStaticProps>) => {
    return (
        <Layout description="Arttu Pennanen Web-like projects">
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
