import {
  CanvasRenderingContext2D,
  createCanvas,
  Image,
  loadImage,
  registerFont
} from 'canvas'
import fs from 'fs'
import path from 'path'
import { Post } from '../../app/entries/lib'
import { formatDate } from '../../app/lib'
import { baseUrl } from '../../app/sitemap'
import { Project } from '../../lib/projects'
import { drawImageCover, wrapLines } from './lib'

const ASSET_PATH = path.join(process.cwd(), 'public')
const META_PATH = path.join(ASSET_PATH, 'meta')
const FONT_PATH = path.join(ASSET_PATH, 'fonts')

const FONTS = [
  { file: 'Inter-Regular.ttf', family: 'Inter', weight: '400' },
  { file: 'Inter-Medium.ttf', family: 'Inter', weight: '500' },
  { file: 'Inter-Bold.ttf', family: 'Inter', weight: '700' }
]

FONTS.forEach(({ file, weight, family }) =>
  registerFont(path.join(FONT_PATH, file), { family, weight })
)

const ASSETS = {
  logo: loadImage(path.join(ASSET_PATH, 'icons', 'icon-white-128.png')),
  bg: loadImage(path.join(ASSET_PATH, 'images', 'metabg.png'))
}

const WIDTH = 1200
const HEIGHT = 630
const PANEL_WIDTH = 600
const PADDING = 48
const CONTENT_WIDTH = PANEL_WIDTH - PADDING * 2
const LOGO_SIZE = 32
const TITLE_GAP = 24

type TextStyle = {
  font: string
  color: string
  lineHeight: number
  maxLines: number
}

const BRAND: TextStyle = {
  font: '500 24px Inter',
  color: 'rgba(255, 255, 255, 0.8)',
  lineHeight: 32,
  maxLines: 1
}

const TITLE: TextStyle = {
  font: '700 60px Inter',
  color: '#4980ff',
  lineHeight: 64,
  maxLines: 3
}

const DESCRIPTION: TextStyle = {
  font: '400 28px Inter',
  color: '#9fbcff',
  lineHeight: 38,
  maxLines: 5
}

const FOOTER: TextStyle = {
  font: '500 24px Inter',
  color: 'rgba(255, 255, 255, 0.9)',
  lineHeight: 32,
  maxLines: 1
}

type Card = {
  title: string
  description?: string
  footer?: string
  image?: Image
}

const layoutText = (
  context: CanvasRenderingContext2D,
  text: string | undefined,
  style: TextStyle
) => {
  if (!text) return { lines: [], height: 0 }
  context.font = style.font
  const lines = wrapLines(context, text, CONTENT_WIDTH, style.maxLines)
  return { lines, height: lines.length * style.lineHeight }
}

const drawLines = (
  context: CanvasRenderingContext2D,
  lines: string[],
  style: TextStyle,
  top: number
) => {
  context.font = style.font
  context.fillStyle = style.color
  context.textAlign = 'left'
  context.textBaseline = 'middle'
  lines.forEach((line, i) => {
    context.fillText(
      line,
      PADDING,
      top + i * style.lineHeight + style.lineHeight / 2
    )
  })
}

async function renderCard({ title, description, footer, image }: Card) {
  const canvas = createCanvas(WIDTH, HEIGHT)
  const context = canvas.getContext('2d')

  drawImageCover(context, image ?? (await ASSETS.bg), 0, 0, WIDTH, HEIGHT)
  context.fillStyle = 'rgba(0, 0, 0, 0.86)'
  context.fillRect(0, 0, PANEL_WIDTH, HEIGHT)

  context.drawImage(await ASSETS.logo, PADDING, PADDING, LOGO_SIZE, LOGO_SIZE)
  context.font = BRAND.font
  context.fillStyle = BRAND.color
  context.textBaseline = 'middle'
  context.fillText(
    'pennanen.dev',
    PADDING + LOGO_SIZE + 12,
    PADDING + LOGO_SIZE / 2
  )

  const footerTop = HEIGHT - PADDING - FOOTER.lineHeight
  if (footer) drawLines(context, [footer], FOOTER, footerTop)

  // Title and description are centered between the brand row and the footer
  const titleBlock = layoutText(context, title, TITLE)
  const descriptionBlock = layoutText(context, description, DESCRIPTION)
  const gap = descriptionBlock.height ? TITLE_GAP : 0
  const areaTop = PADDING + LOGO_SIZE
  const areaHeight = footerTop - areaTop
  const top =
    areaTop +
    (areaHeight - titleBlock.height - gap - descriptionBlock.height) / 2

  drawLines(context, titleBlock.lines, TITLE, top)
  drawLines(
    context,
    descriptionBlock.lines,
    DESCRIPTION,
    top + titleBlock.height + gap
  )

  return canvas.toBuffer('image/jpeg', { quality: 0.9 })
}

const saveImage = (fileName: string, buffer: Buffer) => {
  fs.mkdirSync(META_PATH, { recursive: true })
  fs.writeFileSync(path.join(META_PATH, fileName), buffer)
  return fileName
}

export const getProjectImageFileName = (project: Project) =>
  `${project.id.toLowerCase()}.jpg` as const

export async function generateProjectImage(project: Project) {
  const buffer = await renderCard({
    title: project.title,
    description: project.summary,
    footer: project.notAuthor ? undefined : 'Arttu Pennanen',
    image: project.cover ? await loadImage(project.cover.filePath) : undefined
  })
  return saveImage(getProjectImageFileName(project), buffer)
}

const getPostImageUrl = (post: Post) => {
  return `${post.slug.replace(/[| .-]/g, '_').toLowerCase()}.jpg` as const
}

export const getFullPostImageUrl = (post: Post) => {
  return `${baseUrl}/meta/${getPostImageUrl(post)}` as const
}

export async function generatePostImage(post: Post) {
  const buffer = await renderCard({
    title: post.metadata.title,
    description: post.metadata.summary,
    footer: formatDate(post.metadata.date) || undefined
  })
  return saveImage(getPostImageUrl(post), buffer)
}
