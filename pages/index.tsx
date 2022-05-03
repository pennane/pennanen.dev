import type { InferGetStaticPropsType } from 'next'
import Layout from '../components/Layout'
import MainHeading from '../components/MainHeading'
import Projects from '../components/Project/Projects'
import { generateMainPageImage } from '../lib/meta-image'
import { getFilteredProjects } from '../lib/stuff'
import { ProjectInterface } from '../types'
import Link from 'next/link'

const description =
    'Arttu Pennanen Web projects porfolio. Applications built with Typescript, Javascript, Node, Deno, React and Vue.'

export const getStaticProps = async () => {
    const projectsData: ProjectInterface[] = getFilteredProjects()
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
                <h2>Projects</h2>
                <Projects projects={projectsData} />
            </section>
            <section className="all-link">
                <Link href="/uncurated">Open list of uncurated projects</Link>
            </section>
        </Layout>
    )
}

export default Index
