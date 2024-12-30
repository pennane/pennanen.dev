import style from './projects.module.css'
import { TProject } from '../../../types/models'
import ProjectLink from '../ProjectLink'
import SequentialAnimation from '../../SequentialAnimation'

const Projects = ({ projects }: { projects: TProject[] }) => {
	return (
		<div className={style['main']}>
			<SequentialAnimation
				initialDelay={80}
				delayBetween={40}
				once={true}
				animationKey={`projects-${projects.length}`}
			>
				{projects.map((p, i) => (
					<ProjectLink project={p} key={p.id + '-' + i} />
				))}
			</SequentialAnimation>
		</div>
	)
}

export default Projects
