// components/Stack.tsx
import React from 'react'

export enum Gap {
  none = '0',
  tiny = '0.125rem',
  small = '0.25rem',
  medium = '0.5rem',
  large = '1rem',
  xLarge = '2rem'
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
  gap?: string | { row: string; column: string }
  children: React.ReactNode
} & React.DetailedHTMLProps<
  React.HTMLAttributes<HTMLDivElement>,
  HTMLDivElement
>

export const Stack: React.FC<StackProps> = ({
  vertical = false,
  alignItems = 'baseline',
  justifyContent = 'flex-start',
  wrap = false,
  reverse = false,
  gap = Gap.medium,
  children,
  ...rest
}) => {
  const direction = vertical ? 'row' : 'column'

  const wrapStyle =
    wrap === true ? 'wrap' : wrap === 'wrap-reverse' ? 'wrap-reverse' : 'nowrap'

  const flexGap =
    typeof gap === 'string'
      ? { gap }
      : { columnGap: gap.column, rowGap: gap.row }

  const style: React.CSSProperties = {
    display: 'flex',
    flexDirection: reverse ? `${direction}-reverse` : direction,
    ...flexGap,
    alignItems: alignItems,
    justifyContent: justifyContent,
    flexWrap: wrapStyle
  }

  return (
    <div {...rest} style={style}>
      {children}
    </div>
  )
}
