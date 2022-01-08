import { useRouter } from 'next/router'
import { useContext, useEffect, useState } from 'react'
import { AppContext } from '../lib/context'
import { AppActionType } from '../lib/reducer'
import style from '../styles/project-link.module.css'
import { Project } from '../types'
import ProjectLink from './project-link'
import SequentialAnimation from './sequential-animation'

const Projects = ({ projects }: { projects: Project[] }) => {
    return (
        <div className={style['projects']}>
            <SequentialAnimation childClass="fade-in" initialDelay={80} once={true} animationKey="projects">
                {projects.map((p, i) => (
                    <ProjectLink project={p} key={p.id + '-' + i} />
                ))}
            </SequentialAnimation>
        </div>
    )
}

export default Projects
