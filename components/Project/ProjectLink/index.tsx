import style from './projectlink.module.css'
import Link from 'next/link'
import Image from 'next/image'
import { TProject } from '../../../models'
import { useContext, useLayoutEffect, useState } from 'react'
import { monthIndexToName } from '../../../lib/date'
import { Section } from '../../Section'
import { AppContext } from '../../../context'
import { animationComplete } from '../../../context/reducer'

const ProjectLink = ({ project }: { project: TProject }) => {
	const date =
		project.date && !project.ignoreDate ? new Date(project.date) : null

	const { dispatch, state } = useContext(AppContext)

	const animationKey = `project-${project.id}`

	const runPreviously = !!state.animations[animationKey]

	const [imageLoaded, setImageLoaded] = useState(runPreviously)

	const [mounted, setMounted] = useState(false)
	useLayoutEffect(() => setMounted(true), [])

	const handleLoadComplete = () => {
		if (!mounted || imageLoaded) return

		dispatch(animationComplete(animationKey))
		setImageLoaded(true)
	}

	return (
		<div className={style['main']}>
			<Link href={'/' + project.id}>
				<div className={style['container']}>
					<div className={style['image']}>
						<Image
							className={[
								style['image'],
								!imageLoaded && style['loading'],
							]
								.filter(Boolean)
								.join(' ')}
							width={48}
							height={48}
							alt={project.icon ? `${project.name} icon` : ''}
							src={
								project.icon
									? '/sub/' + project.id + '/' + project.icon
									: '/images/placeholder.png'
							}
							onLoad={handleLoadComplete}
						/>
					</div>

					<Section gap="tiny">
						<h2 className={style['heading']}>{project.name}</h2>
						<p className={style['description']}>
							{project.description}
						</p>
						{date && (
							<span
								className={style['date']}
							>{`${monthIndexToName(
								date.getMonth()
							)}, ${date.getFullYear()}`}</span>
						)}
					</Section>
				</div>
			</Link>
		</div>
	)
}

export default ProjectLink
