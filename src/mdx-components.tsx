import type { MDXComponents } from 'mdx/types'
import { A } from './components/A'

export function useMDXComponents(components: MDXComponents): MDXComponents {
  return {
    ...components,
    a: (props) => <A rel="noopener noreferrer" target="_blank" {...props} />
  }
}
