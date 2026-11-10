const WIDTH_MARK = '#w='

export type SizedImage = { src: string; width: number | undefined }

const unsized = (href: string): SizedImage => ({ src: href, width: undefined })

export const readImageWidth = (href: string): SizedImage => {
  const at = href.lastIndexOf(WIDTH_MARK)
  if (at === -1) return unsized(href)
  const width = Number(href.slice(at + WIDTH_MARK.length))
  return Number.isInteger(width) && width > 0 ? { src: href.slice(0, at), width } : unsized(href)
}

export const withImageWidth = ({ src, width }: SizedImage) =>
  width ? `${src}${WIDTH_MARK}${Math.round(width)}` : src
