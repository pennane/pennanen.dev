import style from './project.module.css'
import linkStyle from './ProjectLink/projectlink.module.css'
import { TProject } from '../../models'
import Image from 'next/image'
import { monthIndexToName } from '../../lib/date'
import { Section } from '../Section'
import { Stack } from '../Stack'

export default function Project({ project }: { project: TProject }) {
	const date = project.date ? new Date(project.date) : null
	return (
		<Stack fill="width" gap="large">
			<header className={style['header']}>
				<Stack gap="medium">
					<Stack gap="none">
						<h1>{project.name}</h1>
						{date && (
							<span className={linkStyle['date']}>
								From{' '}
								{`${monthIndexToName(date.getMonth())}, ${date.getFullYear()}`}
							</span>
						)}
					</Stack>

					<Section direction="row" alignItems="center">
						<div className={linkStyle['image']}>
							<Image
								className={linkStyle['image']}
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
							/>
						</div>
						{project.url && project.url.includes('github.com') && (
							<a href={project.url} rel="noreferrer">
								View source{' '}
								<span className={linkStyle['date']}>
									[{project.url}]
								</span>
							</a>
						)}
						{project.url &&
							project.url.includes('pennanen.dev') && (
								<a href={project.url} rel="noreferrer">
									Open project{' '}
									<span className={linkStyle['date']}>
										[{project.url}]
									</span>
								</a>
							)}
						{project.url &&
							!project.url.includes('github.com') &&
							!project.url.includes('pennanen.dev') && (
								<a href={project.url} rel="noreferrer">
									Open outbound link{' '}
									<span className={linkStyle['date']}>
										[{project.url}]
									</span>
								</a>
							)}
						{!project.url && (
							<a href={`/sub/${project.id}/index.html`}>
								Launch project
							</a>
						)}
					</Section>
					<p className={style['description']}>
						{project.description}
					</p>
				</Stack>
			</header>

			{project.pretext && (
				<Section>
					<p
						dangerouslySetInnerHTML={{
							__html: project.pretext,
						}}
					></p>
				</Section>
			)}
			{project.images[0] && (
				<Section>
					{project.images.map((url, i) => (
						<div className={style['insert']} key={project.id + i}>
							{url.endsWith('.mp4') && (
								<video
									src={'/sub/' + project.id + '/' + url}
									muted={true}
									autoPlay={true}
									loop={true}
								/>
							)}
							{!url.endsWith('.mp4') && url.includes('://') && (
								<img src={url} alt="" loading="lazy" />
							)}
							{!url.endsWith('.mp4') && !url.includes('://') && (
								<img
									src={'/sub/' + project.id + '/' + url}
									alt=""
									loading="lazy"
								/>
							)}
						</div>
					))}
				</Section>
			)}
			{project.github && (
				<Section>
					<a href={project.github} rel="noreferrer">
						View source
					</a>
				</Section>
			)}
		</Stack>
	)
}
