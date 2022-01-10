import { useRouter } from 'next/router'
import React, { useContext, useEffect, useState } from 'react'
import { AppContext } from '../lib/context'
import { animationComplete } from '../lib/reducer'
import style from '../styles/sequential-animation.module.css'

interface SequentialProps {
    children: React.ReactNode
    initialDelay?: number
    delayBetween?: number
    animationDuration?: number
    childClass?: string
    once?: boolean
    animationKey?: string
    stopped?: boolean
}

const SequentialAnimation = ({
    children,
    initialDelay = 0,
    delayBetween = 40,
    stopped: wait = false,
    ...props
}: SequentialProps) => {
    const router = useRouter()
    const { dispatch, state } = useContext(AppContext)

    const shouldAnimate =
        !props.animationKey || !(props.animationKey in state.animations) ? true : !state.animations[props.animationKey]

    const [animationFinished, setAnimationFinished] = useState<boolean>(false)

    const childrenArray = React.Children.toArray(children)

    const [mounted, setMounted] = useState(false)
    useEffect(() => {
        setMounted(true)
    }, [])

    useEffect(() => {
        if (!props.once || !props.animationKey) return

        const timeout = setTimeout(() => {
            setAnimationFinished(true)
        }, (initialDelay + childrenArray.length * delayBetween + (props.animationDuration || 0)) / 2)

        const onHistoryChange = () => {
            if (shouldAnimate && animationFinished && props.animationKey) {
                dispatch(animationComplete(props.animationKey))
            }
        }
        router.events.on('beforeHistoryChange', onHistoryChange)

        return () => {
            router.events.off('beforeHistoryChange', onHistoryChange)
            clearInterval(timeout)
        }
    }, [animationFinished, shouldAnimate])

    if (!shouldAnimate || !mounted) return <div className={style['main']}> {children} </div>

    if (wait) {
        return <div className={style['main'] + ' ' + style['waiting']}>{children}</div>
    }

    return (
        <div className={style['main']}>
            {childrenArray.map((child, i) => (
                <div
                    className={props.childClass ? `${style['child']} ${props.childClass}` : style['child']}
                    style={
                        props.animationDuration
                            ? {
                                  animationDelay: initialDelay + delayBetween * i + 'ms',
                                  animationDuration: props.animationDuration + 'ms'
                              }
                            : { animationDelay: initialDelay + delayBetween * i + 'ms' }
                    }
                    key={'sequential-child-' + i}
                >
                    {child}
                </div>
            ))}
        </div>
    )
}

export default SequentialAnimation
