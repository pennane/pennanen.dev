// components/Stack.tsx
import React from 'react'

type StackProps = {
  direction?: 'row' | 'column'
  align?: 'flex-start' | 'center' | 'flex-end' | 'stretch' | 'baseline'
  justify?:
    | 'flex-start'
    | 'center'
    | 'flex-end'
    | 'space-between'
    | 'space-around'
    | 'space-evenly'
  wrap?: boolean | 'wrap' | 'wrap-reverse'
  reverse?: boolean
  alignContent?:
    | 'flex-start'
    | 'center'
    | 'flex-end'
    | 'stretch'
    | 'space-between'
    | 'space-around'
  gap?: string
  children: React.ReactNode
} & React.DetailedHTMLProps<
  React.HTMLAttributes<HTMLDivElement>,
  HTMLDivElement
>

export const Stack: React.FC<StackProps> = ({
  direction = 'column',
  align = 'baseline',
  justify = 'flex-start',
  wrap = false,
  reverse = false,
  alignContent = 'baseline',
  gap = '0.5rem',
  children,
  ...rest
}) => {
  const wrapStyle =
    wrap === true ? 'wrap' : wrap === 'wrap-reverse' ? 'wrap-reverse' : 'nowrap'

  const style: React.CSSProperties = {
    display: 'flex',
    flexDirection: reverse ? `${direction}-reverse` : direction,
    gap: gap,
    alignItems: align,
    justifyContent: justify,
    flexWrap: wrapStyle,
    alignContent: alignContent
  }

  return (
    <div {...rest} style={style}>
      {children}
    </div>
  )
}
