import { useRouter } from 'next/router'
import { useContext, useEffect, useState } from 'react'
import { AppContext } from '../lib/context'
import { AppActionType } from '../lib/reducer'
import style from '../styles/project-link.module.css'
import { Project } from '../types'
import ProjectLink from './project-link'
import SequentialAnimation from './sequential-animation'

const Projects = ({ projects }: { projects: Project[] }) => {
    const router = useRouter()
    const { dispatch, state } = useContext(AppContext)

    const shouldAnimate = !state.finished

    const [animated, setAnimated] = useState<boolean>(false)

    useEffect(() => {
        const onHistoryChange = () => {
            if (shouldAnimate && animated) {
                dispatch({ type: AppActionType.ANIMATION_COMPLETE })
            }
        }
        router.events.on('beforeHistoryChange', onHistoryChange)

        return () => {
            router.events.off('beforeHistoryChange', onHistoryChange)
        }
    }, [animated])

    return (
        <div className={style['projects']}>
            {!shouldAnimate && projects && projects.map((p, i) => <ProjectLink project={p} key={p.id + '-' + i} />)}
            {shouldAnimate && projects && (
                <SequentialAnimation childClass="fade-in" initialDelay={80} onFinish={() => setAnimated(true)}>
                    {projects.map((p, i) => (
                        <ProjectLink project={p} key={p.id + '-' + i} />
                    ))}
                </SequentialAnimation>
            )}
        </div>
    )
}

export default Projects
