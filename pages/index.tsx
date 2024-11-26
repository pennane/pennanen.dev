import type { InferGetStaticPropsType } from 'next'
import Layout from '../components/Layout'
import MainHeading from '../components/MainHeading'
import Projects from '../components/Project/Projects'
import { generateMainPageImage } from '../lib/meta-image'
import { getFilteredProjects } from '../lib/stuff'
import { IProject } from '../types'
import Link from 'next/link'
import Image from 'next/image'
import SequentialAnimation from '../components/SequentialAnimation'
import { useState } from 'react'

const description =
  'Arttu Pennanen - the digital playground pennanen.dev. Sharing my programming endeavours, from individual explorations to collaborative projects. Each piece is a step towards unraveling the vast world of code, crafted with desire to learn. Dive in to see where theory meets practice, through the lens of my experiences.'

export const getStaticProps = async () => {
  const projectsData: IProject[] = getFilteredProjects()
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
  const [imgLoaded, setImgLoaded] = useState(false)
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
              Hello, I&apos;m Arttu. I enjoy dumb stuff like functional
              programming and thinking through problems in an exceedingly
              generalized manner.
            </p>
            <p>
              I&apos;m currently working as a Software Engineer at
              Hoxhunt and alongside completing an Information and Communication
              Technology Bachelor&apos;s degree at Metropolia UAS.
            </p>
            <p>
              Here on this website, I post about some of the programming
              I&apos;ve fought with.
            </p>
          </div>
          <aside>
            <SequentialAnimation
              animationKey="face"
              once={true}
              initialDelay={300}
              delayBetween={0}
              animationDuration={500}
              stopped={!imgLoaded}
            >
              <Image
                alt="Black and white photo of myself giving thumbs-up outdoors"
                width={310}
                height={200}
                src={'/images/peukku.jpg'}
                onLoadingComplete={() => setImgLoaded(true)}
              ></Image>
            </SequentialAnimation>
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
