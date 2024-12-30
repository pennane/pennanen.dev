import style from './projectlink.module.css'
import Link from 'next/link'
import Image from 'next/image'
import { TProject } from '../../../models'
import { useEffect, useRef, useState } from 'react'
import { monthIndexToName } from '../../../lib/date'
import { Section } from '../../Section'

const ProjectLink = ({ project }: { project: TProject }) => {
	const date =
		project.date && !project.ignoreDate ? new Date(project.date) : null
	const [imageLoaded, setImageLoaded] = useState(false)
	const mounted = useRef(false)

	useEffect(() => {
		mounted.current = true
		return () => {
			mounted.current = false
		}
	}, [])

	return (
		<div className={style['main']}>
			<Link href={'/' + project.id}>
				<a>
					<div className={style['container']}>
						<div className={style['image']}>
							<Image
								className={
									imageLoaded
										? style['image']
										: `${style['image']} ${style['loading']}`
								}
								width={48}
								height={48}
								alt={project.icon ? `${project.name} icon` : ''}
								src={
									project.icon
										? '/sub/' +
											project.id +
											'/' +
											project.icon
										: '/images/placeholder.png'
								}
								onLoadingComplete={() =>
									mounted.current && setImageLoaded(true)
								}
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
				</a>
			</Link>
		</div>
	)
}

export default ProjectLink
