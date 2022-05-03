import style from './projects.module.css'
import { ProjectInterface } from '../../../types'
import ProjectLink from '../ProjectLink'
import SequentialAnimation from '../../SequentialAnimation'

const Projects = ({ projects }: { projects: ProjectInterface[] }) => {
    return (
        <div className={style['main']}>
            <SequentialAnimation initialDelay={80} once={true} animationKey="projects">
                {projects.map((p, i) => (
                    <ProjectLink project={p} key={p.id + '-' + i} />
                ))}
            </SequentialAnimation>
        </div>
    )
}

export default Projects
