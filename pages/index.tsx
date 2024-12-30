import type { InferGetStaticPropsType } from 'next'
import Layout from '../components/Layout'
import Projects from '../components/Project/Projects'
import { generateMainPageImage } from '../lib/meta-image'
import { getFilteredProjects } from '../lib/stuff'
import { IProject } from '../types'
import Link from 'next/link'

const description =
	"Arttu Pennanen - the digital playground pennanen.dev. Sharing some of the personal projects I've built for the sake of building something."

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
	return (
		<Layout description={description} metaImage={metaImage}>
			<section style={{ position: 'relative' }}>
				<header>
					<h2>Arttu Pennanen</h2>
					<h3>Software Engineer</h3>
				</header>

				<div className="about">
					<div>
						<p>
							Writing code that can not break. Currently at
							Hoxhunt. Metropolia alumni.
						</p>
						<p>
							Here on this website, I post about some of the
							problems I&apos;ve fought.
						</p>
					</div>
				</div>
			</section>
			<section>
				<h3>My pride and joy</h3>
				<Projects projects={projectsData} />
			</section>
			<section className="all-link">
				<Link href="/uncurated">List of uncurated projects</Link>
			</section>
		</Layout>
	)
}

export default Index
