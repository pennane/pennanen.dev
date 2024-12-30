import style from './projects.module.css'
import { TProject } from '../../../models'
import ProjectLink from '../ProjectLink'
import SequentialAnimation from '../../SequentialAnimation'
import { Stack } from '../../Stack'

const Projects = ({ projects }: { projects: TProject[] }) => {
	return (
		<Stack className={style['main']}>
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
		</Stack>
	)
}

export default Projects
