import type { InferGetStaticPropsType } from 'next'
import Layout from '../components/Layout'
import MainHeading from '../components/MainHeading'
import Projects from '../components/Project/Projects'
import { generateMainPageImage } from '../lib/meta-image'
import { getFilteredProjects } from '../lib/stuff'
import { ProjectInterface } from '../types'
import Link from 'next/link'
import Image from 'next/image'

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
        <h2>About me</h2>
        <div className="about">
          <div>
            <p>
              Hello, I am Arttu. I like &quot;dumb&quot; stuff like functional
              programming and using higher-order functions in JavaScript.
            </p>
            <p>
              I&apos;m currently working as a Junior Software Engineer at
              Hoxhunt and alongside completing an Information and Communication
              Technology Bachelor&apos;s degree at Metropolia.
            </p>
            <p>
              Here on my website, I post about the programming endeavors I find
              myself tinkering with.
            </p>
          </div>
          <aside>
            <Image
              alt='Black and white photo of myself giving thumbs-up outdoors."'
              width={310}
              height={200}
              src={'/images/peukku.jpg'}
            ></Image>
          </aside>
        </div>
      </section>
      <section>
        <h2>Pinned projects</h2>
        <Projects projects={projectsData} />
      </section>
      <section className="all-link">
        <Link href="/uncurated">List of uncurated projects</Link>
      </section>
    </Layout>
  )
}

export default Index
