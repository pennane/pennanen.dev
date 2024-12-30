import type { InferGetStaticPropsType } from 'next'
import Layout from '../components/Layout'
import Projects from '../components/Project/Projects'
import { generateMainPageImage } from '../lib/meta-image'
import { getProjects } from '../lib/stuff'
import { IProject } from '../types'

const description =
	'All of the prior Web projects. Unfiltered and uncurated. Functionality not guaranteed.'

export const getStaticProps = async () => {
	const projectsData: IProject[] = getProjects()
	const metaImage = await generateMainPageImage({
		title: 'All | pennanen.dev',
		description,
	})
	return {
		props: {
			projectsData,
			metaImage,
		},
	}
}

const All = ({
	projectsData,
	metaImage,
}: InferGetStaticPropsType<typeof getStaticProps>) => {
	return (
		<Layout description={description} metaImage={metaImage}>
			<section>
				<h2>Uncurated projects</h2>
				<Projects
					projects={projectsData.filter(
						(project) => project.ignoreInListing
					)}
				/>
			</section>
		</Layout>
	)
}

export default All
