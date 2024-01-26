import type { InferGetStaticPropsType } from 'next'
import Layout from '../components/Layout'
import MainHeading from '../components/MainHeading'
import Projects from '../components/Project/Projects'
import { generateMainPageImage } from '../lib/meta-image'
import { getFilteredProjects } from '../lib/stuff'
import { ProjectInterface } from '../types'
import Link from 'next/link'

const description =
  'Arttu Pennanen - the digital playground pennanen.dev. Sharing my programming endeavours, from individual explorations to collaborative projects. Each piece is a step towards unraveling the vast world of code, crafted with desire to learn. Dive in to see where theory meets practice, through the lens of my experiences.'

export const getStaticProps = async () => {
  const projectsData: ProjectInterface[] = getFilteredProjects()
  const metaImage = await generateMainPageImage({
    title: 'pennanen.dev',
    description,
  })
  return {
    props: {
      projectsData,
      metaImage,
    },
  }
}

const Index = ({
  projectsData,
  metaImage,
}: InferGetStaticPropsType<typeof getStaticProps>) => {
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
