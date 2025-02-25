/* eslint-disable @typescript-eslint/no-explicit-any */
import { cloneElement, isValidElement, ReactElement } from 'react'
import styles from './wrap-in-box.module.css'

export const WrapInBox = ({
  children
}: {
  children: ReactElement | ReactElement[]
}) => {
  if (!children) return null

  if (Array.isArray(children)) {
    return (
      <>
        {children.map((child) =>
          isValidElement(child)
            ? cloneElement(child as ReactElement<any>, {
                className: `${styles.box} ${
                  (child as ReactElement<any>).props.className || ''
                }`
              })
            : child
        )}
      </>
    )
  }

  if (isValidElement(children)) {
    return cloneElement(children as ReactElement<any>, {
      className: `${styles.box} ${
        (children as ReactElement<any>).props.className || ''
      }`
    })
  }

  return <>{children}</>
}
