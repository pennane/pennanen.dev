import React from 'react'
import style from '../styles/sequential-animation.module.css'

interface SequentialProps {
    children: React.ReactNode
    initialDelay?: number
    delayBetween?: number
    once?: boolean
    childClass?: string
    onFinish?: any
}

const SequentialAnimation = ({ children, initialDelay = 0, delayBetween = 40, ...props }: SequentialProps) => {
    const childrenArray = React.Children.toArray(children)

    if (props.onFinish) {
        setTimeout(() => {
            props.onFinish()
        }, initialDelay + childrenArray.length * delayBetween)
    }

    return (
        <div className={style['main']}>
            {childrenArray.map((child, i) => (
                <div
                    className={props.childClass ? `${style['child']} ${props.childClass}` : style['child']}
                    style={{ animationDelay: initialDelay + delayBetween * i + 'ms' }}
                    key={'sequential-child-' + i}
                >
                    {child}
                </div>
            ))}
        </div>
    )
}

export default SequentialAnimation
