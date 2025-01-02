import { useRouter } from 'next/router'
import React, { useContext, useLayoutEffect, useState } from 'react'
import { AppContext } from '../../context'
import { animationComplete } from '../../context/reducer'
import style from './sequential-animation.module.css'

interface SequentialProps {
	children: React.ReactNode
	initialDelay?: number
	delayBetween?: number
	animationDuration?: number
	childClass?: string
	once?: boolean
	animationKey?: string
}

const FINISHED_STYLE = {
	animationPlayState: 'paused',
	animationDirection: 'reverse',
}

const INITIAL_STYLE = {
	animationPlayState: 'paused',
}

const ONGOING_STYLE = ({
	initialDelay,
	delayBetween,
	animationDuration,
	i,
}: {
	initialDelay: number
	delayBetween: number
	animationDuration?: number
	i: number
}) => ({
	animationDelay: `${initialDelay + delayBetween * i}ms`,
	animationDuration: animationDuration ? `${animationDuration}ms` : undefined,
})

const getStyle = (props: {
	shouldAnimate: boolean
	mounted: boolean
	initialDelay: number
	delayBetween: number
	animationDuration: number | undefined
	i: number
}) => {
	if (!props.shouldAnimate) return FINISHED_STYLE
	if (!props.mounted) return INITIAL_STYLE
	return ONGOING_STYLE(props)
}

const SequentialAnimation: React.FC<SequentialProps> = ({
	children,
	initialDelay = 40,
	delayBetween = 30,
	once,
	animationKey,
	animationDuration,
	childClass,
}) => {
	const router = useRouter()
	const { dispatch, state } = useContext(AppContext)

	const shouldAnimate = !animationKey || !state.animations[animationKey]

	const [animationFinished, setAnimationFinished] = useState(false)
	const [mounted, setMounted] = useState(false)

	const childrenArray = React.Children.toArray(children)

	useLayoutEffect(() => setMounted(true), [])

	useLayoutEffect(() => {
		if (!once || !animationKey) return

		const totalAnimationTime =
			initialDelay +
			childrenArray.length * delayBetween +
			(animationDuration || 0)

		const timeout = setTimeout(
			() => setAnimationFinished(true),
			totalAnimationTime / 2
		)

		const onHistoryChange = () => {
			if (shouldAnimate && animationFinished && animationKey) {
				dispatch(animationComplete(animationKey))
			}
		}

		router.events.on('beforeHistoryChange', onHistoryChange)

		return () => {
			router.events.off('beforeHistoryChange', onHistoryChange)
			clearTimeout(timeout)
		}
	}, [animationFinished, shouldAnimate])

	return childrenArray.map((child, i) => (
		<div
			key={`sequential-child-${i}`}
			className={[style.child, childClass || style['fade-in']].join(' ')}
			style={getStyle({
				i,
				animationDuration,
				delayBetween,
				initialDelay,
				mounted,
				shouldAnimate,
			})}
		>
			{child}
		</div>
	))
}

export default SequentialAnimation
