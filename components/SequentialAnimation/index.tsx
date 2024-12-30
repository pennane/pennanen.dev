import { useRouter } from 'next/router'
import React, { useContext, useLayoutEffect, useState } from 'react'
import { AppContext } from '../../lib/context'
import { animationComplete } from '../../lib/reducer'
import style from './sequential-animation.module.css'

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

const SequentialAnimation: React.FC<SequentialProps> = ({
	children,
	initialDelay = 40,
	delayBetween = 30,
	stopped = false,
	...props
}) => {
	const router = useRouter()
	const { dispatch, state } = useContext(AppContext)

	const shouldAnimate =
		!props.animationKey || !(props.animationKey in state.animations)
			? true
			: !state.animations[props.animationKey]

	const [animationFinished, setAnimationFinished] = useState(false)
	const [mounted, setMounted] = useState(false)

	const childrenArray = React.Children.toArray(children)

	useLayoutEffect(() => {
		setMounted(true)
	}, [])

	useLayoutEffect(() => {
		if (!props.once || !props.animationKey) return

		const totalAnimationTime =
			initialDelay +
			childrenArray.length * delayBetween +
			(props.animationDuration || 0)

		const timeout = setTimeout(
			() => setAnimationFinished(true),
			totalAnimationTime / 2
		)

		const onHistoryChange = () => {
			if (shouldAnimate && animationFinished && props.animationKey) {
				dispatch(animationComplete(props.animationKey))
			}
		}

		router.events.on('beforeHistoryChange', onHistoryChange)

		return () => {
			router.events.off('beforeHistoryChange', onHistoryChange)
			clearTimeout(timeout)
		}
	}, [animationFinished, shouldAnimate])

	if (!shouldAnimate || !mounted) return <>{children}</>

	if (stopped) {
		return <div className={style['waiting']}>{children}</div>
	}

	return (
		<>
			{childrenArray.map((child, i) => (
				<div
					key={`sequential-child-${i}`}
					className={
						props.childClass
							? `${style['child']} ${props.childClass}`
							: `${style['child']} ${style['fade-in']}`
					}
					style={{
						animationDelay: `${initialDelay + delayBetween * i}ms`,
						animationDuration: props.animationDuration
							? `${props.animationDuration}ms`
							: 'undefined',
					}}
				>
					{child}
				</div>
			))}
		</>
	)
}

export default SequentialAnimation
