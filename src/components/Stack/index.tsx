import React from 'react'
import clsx from 'clsx'

export enum Gap {
  none = '0',
  tiny = '0.125rem',
  small = '0.25rem',
  medium = '0.5rem',
  large = '1rem',
  xLarge = '2rem',
  mega = '3rem'
}

type StackProps = {
  vertical?: boolean
  alignItems?: 'flex-start' | 'center' | 'flex-end' | 'stretch' | 'baseline'
  justifyContent?:
    | 'flex-start'
    | 'center'
    | 'flex-end'
    | 'space-between'
    | 'space-around'
    | 'space-evenly'
  wrap?: boolean | 'wrap' | 'wrap-reverse'
  reverse?: boolean
  gap?: Gap | { row: Gap; column: Gap }
  children: React.ReactNode
} & React.HTMLAttributes<HTMLDivElement>

export const Stack: React.FC<StackProps> = ({
  vertical = false,
  alignItems = 'baseline',
  justifyContent = 'flex-start',
  wrap = false,
  reverse = false,
  gap = Gap.medium,
  children,
  className,
  ...rest
}) => {
  const direction = vertical ? 'flex-row' : 'flex-col'
  const wrapClass =
    wrap === true
      ? 'flex-wrap'
      : wrap === 'wrap-reverse'
      ? 'flex-wrap-reverse'
      : 'flex-nowrap'
  const flexDirection = reverse ? `${direction}-reverse` : direction

  let gapStyle = {}
  if (typeof gap === 'object') {
    gapStyle = { columnGap: gap.column, rowGap: gap.row }
  } else {
    gapStyle = { gap }
  }

  return (
    <div
      style={gapStyle}
      {...rest}
      className={clsx(
        'flex',
        flexDirection,
        wrapClass,
        `items-${alignItems}`,
        `justify-${justifyContent}`,
        className
      )}
    >
      {children}
    </div>
  )
}
