import style from '../styles/project-link.module.css'
import { Project } from '../types'
import ProjectLink from './project-link'

const Projects = ({ projects }: { projects: Project[] }) => {
    return (
        <div className={style['projects']}>
            {projects && projects.map((p, i) => <ProjectLink project={p} key={p.id + '-' + i} />)}
        </div>
    )
}

export default Projects
