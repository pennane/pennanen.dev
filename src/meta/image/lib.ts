import { Image, CanvasRenderingContext2D } from 'canvas'

export function wrapText(
  context: CanvasRenderingContext2D,
  text: string,
  x: number,
  y: number,
  maxWidth: number,
  lineHeight: number
) {
  const words = text.split(' ')
  let line = ''

  for (let n = 0; n < words.length; n++) {
    const testLine = line + words[n] + ' '
    const metrics = context.measureText(testLine)
    const testWidth = metrics.width
    if (testWidth > maxWidth && n > 0) {
      context.fillText(line, x, y)
      line = words[n] + ' '
      y += lineHeight
    } else {
      line = testLine
    }
  }
  context.fillText(line, x, y)
  return { x, y }
}

/**
 * By Ken Fyrstenberg Nilsen
 *
 * drawImageProp(context, image [, x, y, width, height [,offsetX, offsetY]])
 *
 * If image and context are only arguments rectangle will equal canvas
 */
export function drawImageProp(
  ctx: CanvasRenderingContext2D,
  img: Image,
  x: number,
  y: number,
  w: number,
  h: number,
  offsetX?: number,
  offsetY?: number
) {
  if (arguments.length === 2) {
    x = y = 0
    w = ctx.canvas.width
    h = ctx.canvas.height
  }

  // default offset is center
  offsetX = typeof offsetX === 'number' ? offsetX : 0.5
  offsetY = typeof offsetY === 'number' ? offsetY : 0.5

  // keep bounds [0.0, 1.0]
  if (offsetX < 0) offsetX = 0
  if (offsetY < 0) offsetY = 0
  if (offsetX > 1) offsetX = 1
  if (offsetY > 1) offsetY = 1

  const iw = img.width,
    ih = img.height,
    r = Math.min(w / iw, h / ih)
  let nw = iw * r, // new prop. width
    nh = ih * r, // new prop. height
    cx,
    cy,
    cw,
    ch,
    ar = 1

  // decide which gap to fill
  if (nw < w) ar = w / nw
  if (Math.abs(ar - 1) < 1e-14 && nh < h) ar = h / nh // updated
  nw *= ar
  nh *= ar

  // calc source rectangle
  cw = iw / (nw / w)
  ch = ih / (nh / h)

  cx = (iw - cw) * offsetX
  cy = (ih - ch) * offsetY

  // make sure source rectangle is valid
  if (cx < 0) cx = 0
  if (cy < 0) cy = 0
  if (cw > iw) cw = iw
  if (ch > ih) ch = ih

  // fill image in dest. rectangle
  ctx.drawImage(img, cx, cy, cw, ch, x, y, w, h)
}
