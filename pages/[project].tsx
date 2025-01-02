import type {
	GetStaticPaths,
	GetStaticProps,
	GetStaticPropsContext,
} from 'next'
import Layout from '../components/Layout'
import { getProjectById, getProjectIds } from '../lib/staticProps'
import { isString } from '../lib/fp'
import { TProject } from '../models'
import { generateProjectImage } from '../metaImage'
import Project from '../components/Project'

export const getStaticPaths: GetStaticPaths = async () => {
	const paths = getProjectIds()
	return {
		paths,
		fallback: false,
	}
}

export const getStaticProps: GetStaticProps = async ({
	params,
}: GetStaticPropsContext) => {
	if (!params || !isString(params.project)) return { notFound: true }

	const project = getProjectById(params.project)

	if (!project) return { notFound: true }

	const metaImage = await generateProjectImage(project)

	return { props: { project, metaImage } }
}

const ProjectPage = ({
	project,
	metaImage,
}: {
	project: TProject
	metaImage: string
}) => {
	if (!project) return null
	return (
		<Layout
			title={project.name}
			description={project.description || undefined}
			metaImage={metaImage || undefined}
		>
			<Project project={project} />
		</Layout>
	)
}

export default ProjectPage
