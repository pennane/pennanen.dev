import { Image, CanvasRenderingContext2D } from 'canvas'

export function wrapLines(
  context: CanvasRenderingContext2D,
  text: string,
  maxWidth: number,
  maxLines: number
) {
  const fits = (line: string) => context.measureText(line).width <= maxWidth
  const lines: string[] = []
  let line = ''

  for (const word of text.split(/\s+/).filter(Boolean)) {
    const candidate = line ? `${line} ${word}` : word
    if (line && !fits(candidate)) {
      lines.push(line)
      line = word
    } else {
      line = candidate
    }
  }
  if (line) lines.push(line)

  if (lines.length <= maxLines) return lines

  const kept = lines.slice(0, maxLines)
  let last = kept[maxLines - 1]
  while (last.includes(' ') && !fits(`${last}…`)) {
    last = last.slice(0, last.lastIndexOf(' '))
  }
  kept[maxLines - 1] = `${last.replace(/[\s\p{P}]+$/u, '')}…`
  return kept
}

export function drawImageCover(
  context: CanvasRenderingContext2D,
  image: Image,
  x: number,
  y: number,
  width: number,
  height: number
) {
  const scale = Math.max(width / image.width, height / image.height)
  const sourceWidth = width / scale
  const sourceHeight = height / scale

  context.drawImage(
    image,
    (image.width - sourceWidth) / 2,
    (image.height - sourceHeight) / 2,
    sourceWidth,
    sourceHeight,
    x,
    y,
    width,
    height
  )
}
