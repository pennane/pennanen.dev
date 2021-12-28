import style from '../styles/project-link.module.css'
import Link from 'next/link'
import { Project } from '../types'

const ProjectLink = ({ project }: { project: Project }) => {
    const date = project.date && !project.ignoreDate ? new Date(project.date) : null
    const monthNames = [
        'January',
        'February',
        'March',
        'April',
        'May',
        'June',
        'July',
        'August',
        'September',
        'October',
        'November',
        'December'
    ]

    return (
        <div className={style['main']}>
            <Link href={'/' + project.id}>
                <a>
                    <div className={style['container']}>
                        <img
                            className={style['image']}
                            src={project.icon ? '/stuff/' + project.id + '/' + project.icon : '/images/placeholder.png'}
                        />
                        <div className={style['information']}>
                            <h2 className={style['heading']}>{project.name}</h2>
                            <p className={style['description']}>{project.description}</p>
                            {date && (
                                <span className={style['date']}>{`${
                                    monthNames[date.getMonth()]
                                }, ${date.getFullYear()}`}</span>
                            )}
                        </div>
                    </div>
                </a>
            </Link>
        </div>
    )
}

export default ProjectLink
