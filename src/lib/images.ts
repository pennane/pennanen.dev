import type { ImageSource } from './projects'

// next/image cannot resize these formats, so they are served as they are
export const isUnoptimizable = (src: ImageSource) =>
  typeof src === 'string' && (src.includes('://') || /\.(svg|ico)$/i.test(src))
