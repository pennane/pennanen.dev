import path from 'path'
import fs from 'fs'
import {
	createCanvas,
	Image,
	loadImage,
	registerFont,
	CanvasRenderingContext2D,
} from 'canvas'
import { TProject } from '../models'
import { wrapText, drawImageProp } from './lib'

const ASSET_PATH = path.join(process.cwd(), 'public')
const META_PATH = path.join(ASSET_PATH, 'meta')
const FONT_PATH = path.join(ASSET_PATH, 'fonts')
const ICON_PATH = path.join(ASSET_PATH, 'icons')
const IMAGE_PATH = path.join(ASSET_PATH, 'images')

const FONTS = [
	{ file: 'Inter-Medium.ttf', family: 'Inter', weight: '300' },
	{ file: 'Inter-Bold.ttf', family: 'Inter', weight: '700' },
]

FONTS.forEach(({ file, family, weight }) => {
	registerFont(path.join(FONT_PATH, file), { family, weight })
})

const loadAsset = async (filePath: string) =>
	loadImage(fs.readFileSync(filePath))

const ASSETS = {
	logo: loadAsset(path.join(ICON_PATH, 'icon-white-128.png')),
	darkLogo: loadAsset(path.join(ICON_PATH, 'icon-black-128.png')),
	bg: loadAsset(path.join(IMAGE_PATH, 'metabg.png')),
}

const createCanvasContext = (width: number, height: number) => {
	const canvas = createCanvas(width, height)
	return { canvas, context: canvas.getContext('2d') }
}

const drawText = (
	context: CanvasRenderingContext2D,
	text: string,
	x: number,
	y: number,
	maxWidth: number,
	lineHeight: number,
	font: string,
	color: string,
	align: CanvasTextAlign = 'center'
) => {
	context.font = font
	context.fillStyle = color
	context.textAlign = align
	context.textBaseline = 'top'
	return wrapText(context, text, x, y, maxWidth, lineHeight)
}

const generateImageBuffer = async (
	width: number,
	height: number,
	format: 'image/png' | 'image/jpeg',
	draw: (context: CanvasRenderingContext2D) => Promise<void>
) => {
	const { canvas, context } = createCanvasContext(width, height)
	await draw(context)
	const buffer = canvas.toBuffer(format as any)
	return buffer
}

async function computeMainPageImage(title?: string, description?: string) {
	return generateImageBuffer(1200, 630, 'image/png', async (context) => {
		context.drawImage(await ASSETS.bg, 0, 0, 1200, 630)
		context.fillStyle = 'rgba(255, 255, 255, 0.5)'
		context.fillRect(0, 0, 1200, 630)

		context.drawImage(await ASSETS.darkLogo, 40, 40, 32, 32)

		let headingY = 210
		if (title) {
			headingY = drawText(
				context,
				title,
				600,
				headingY,
				960,
				69.6,
				'700 48pt Inter',
				'#004dff'
			).y
		}

		if (description) {
			drawText(
				context,
				description,
				600,
				headingY + 96,
				960,
				33,
				'300 22pt Inter',
				'#004dff'
			)
		}
	})
}

async function computeProjectImage({
	title,
	image,
	description,
	notAuthor,
}: {
	title: string | null
	image: Image | null
	description: string | null
	notAuthor: boolean
}) {
	return generateImageBuffer(1200, 630, 'image/jpeg', async (context) => {
		if (image) {
			drawImageProp(context, image, 0, 0, 1200, 630)
		} else {
			context.drawImage(await ASSETS.bg, 0, 0, 1200, 630)
		}

		context.fillStyle = 'rgba(0, 0, 0, 0.86)'
		context.fillRect(0, 0, 600, 630)
		context.drawImage(await ASSETS.logo, 40, 40, 32, 32)

		drawText(
			context,
			'pennanen.dev',
			82,
			38,
			300,
			520,
			'300 18pt Inter',
			'rgba(255, 255, 255, 0.8)',
			'left'
		)

		let headingY = 190
		if (title) {
			headingY = drawText(
				context,
				title,
				40,
				headingY,
				520,
				69.6,
				'700 48pt Inter',
				'#4980ff',
				'left'
			).y
		}

		if (description) {
			drawText(
				context,
				description,
				40,
				headingY + 80,
				520,
				33,
				'300 22pt Inter',
				'#9fbcff',
				'left'
			)
		}

		if (!notAuthor) {
			drawText(
				context,
				'Arttu Pennanen',
				40,
				headingY + 350,
				520,
				30,
				'300 20pt Inter',
				'rgba(255, 255, 255, 0.95)',
				'left'
			)
		}
	})
}

const ensureMetaDirectory = () => {
	if (!fs.existsSync(META_PATH)) fs.mkdirSync(META_PATH)
}

const saveImage = (fileName: string, buffer: Buffer) => {
	const filePath = path.join(META_PATH, fileName)
	fs.writeFileSync(filePath, buffer)
	return fileName
}

export async function generateMainPageImage<
	T extends { title: string; description?: string },
>(data: T) {
	ensureMetaDirectory()
	const buffer = await computeMainPageImage(
		data.title || 'pennanen.dev',
		data.description
	)
	const fileName = `${data.title.replace(/[| .-]/g, '_').toLowerCase()}.jpg`
	return saveImage(fileName, buffer)
}

export async function generateProjectImage(project: TProject) {
	ensureMetaDirectory()
	let projectImage: Image | null = null

	if (project.largeImage) {
		try {
			const imageFile = fs.readFileSync(
				path.join(ASSET_PATH, 'sub', project.id, project.largeImage)
			)
			projectImage = await loadImage(imageFile)
		} catch {}
	}

	const buffer = await computeProjectImage({
		description: project.description,
		image: projectImage,
		notAuthor: !!project.notAuthor,
		title: project.name || project.id,
	})
	return saveImage(`${project.id.toLowerCase()}.png`, buffer)
}
