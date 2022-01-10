import style from '../styles/project-link.module.css'
import { Project } from '../types'
import ProjectLink from './project-link'
import SequentialAnimation from './sequential-animation'

const Projects = ({ projects }: { projects: Project[] }) => {
    return (
        <div className={style['projects']}>
            <SequentialAnimation childClass="fade-in" initialDelay={80} once={true} animationKey="projects">
                {projects
                    .filter((p) => !p.ignoreInListing)
                    .map((p, i) => (
                        <ProjectLink project={p} key={p.id + '-' + i} />
                    ))}
            </SequentialAnimation>
        </div>
    )
}

export default Projects
