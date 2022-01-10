import style from '../styles/project-link.module.css'
import Link from 'next/link'
import Image from 'next/image'
import { Project } from '../types'
import { monthIndexToName } from '../lib/util'
import { useState } from 'react'

const ProjectLink = ({ project }: { project: Project }) => {
    const date = project.date && !project.ignoreDate ? new Date(project.date) : null
    const [imageLoaded, setImageLoaded] = useState(false)

    return (
        <div className={style['main']}>
            <Link href={'/' + project.id}>
                <a>
                    <div className={style['container']}>
                        <div className={style['image']}>
                            <Image
                                className={imageLoaded ? style['image'] : `${style['image']} ${style['loading']}`}
                                width={48}
                                height={48}
                                alt={project.icon ? `${project.name} icon` : ''}
                                src={
                                    project.icon ? '/sub/' + project.id + '/' + project.icon : '/images/placeholder.png'
                                }
                                onLoadingComplete={() => setImageLoaded(true)}
                            />
                        </div>

                        <div className={style['information']}>
                            <h2 className={style['heading']}>{project.name}</h2>
                            <p className={style['description']}>{project.description}</p>
                            {date && (
                                <span className={style['date']}>{`${monthIndexToName(
                                    date.getMonth()
                                )}, ${date.getFullYear()}`}</span>
                            )}
                        </div>
                    </div>
                </a>
            </Link>
        </div>
    )
}

export default ProjectLink
