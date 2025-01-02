type GapSize = 'none' | 'tiny' | 'small' | 'medium' | 'mediumlarge' | 'large'
type Direction = 'row' | 'column'
type JustifyContent =
	| 'flex-start'
	| 'flex-end'
	| 'center'
	| 'space-between'
	| 'space-around'
	| 'space-evenly'
type AlignItems = 'flex-start' | 'flex-end' | 'center' | 'stretch' | 'baseline'

export type TStackProps = {
	gap?: GapSize
	direction?: Direction
	justifyContent?: JustifyContent
	alignItems?: AlignItems
	wrap?: boolean
	className?: string
	fill?: 'width' | 'height' | 'both'
	children: React.ReactNode
}

export const Stack = ({
	gap = 'medium',
	direction = 'column',
	justifyContent,
	alignItems,
	wrap = false,
	className,
	fill,
	children,
}: TStackProps) => {
	const classes = [
		'stack',
		gap === 'none' ? '' : gap,
		direction,
		wrap,
		className,
	]
		.filter(Boolean)
		.join(' ')
	const style = { justifyContent, alignItems } as React.CSSProperties
	if (fill === 'both') {
		style.width = '100%'
		style.height = '100%'
	} else if (fill === 'height') {
		style.height = '100%'
	} else if (fill === 'width') {
		style.width = '100%'
	}

	return (
		<div className={classes} style={style}>
			{children}
		</div>
	)
}
