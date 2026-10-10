import type { MDXComponents } from 'mdx/types'
import { A } from './components/A'
import { Embed, MdxImage, Video } from './components/Media'

export function useMDXComponents(components: MDXComponents): MDXComponents {
  return {
    ...components,
    img: MdxImage,
    Video,
    Embed,
    a: (props) => {
      const linkOnSamePage = props.href?.startsWith('#')
      return linkOnSamePage ? (
        <A {...props} />
      ) : (
        <A rel="noopener noreferrer" target="_blank" {...props} />
      )
    }
  }
}
