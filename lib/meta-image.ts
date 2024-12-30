import path from 'path'
import fs from 'fs'
import { createCanvas, Image, loadImage, registerFont } from 'canvas'
import { drawImageProp, wrapText } from './util'
import { IProject } from '../types'

registerFont(path.join(process.cwd(), 'public/fonts/Inter-Medium.ttf'), {
	family: 'Inter',
	weight: '300',
})
registerFont(path.join(process.cwd(), 'public/fonts/Inter-Bold.ttf'), {
	family: 'Inter',
	weight: '700',
})

const logoFile = fs.readFileSync(
	path.join(process.cwd(), 'public/icons/icon-white-128.png')
)
const logo = loadImage(logoFile)
const darkLogoFile = fs.readFileSync(
	path.join(process.cwd(), 'public/icons/icon-black-128.png')
)
const darkLogo = loadImage(darkLogoFile)
const bgFile = fs.readFileSync(
	path.join(process.cwd(), 'public/images/metabg.png')
)
const bg = loadImage(bgFile)

async function computeMainPageImage({
	title,
	description,
}: {
	title?: string
	description?: string
}): Promise<Buffer> {
	const width = 1200
	const height = 630
	const margin = 40
	const squareY = 0

	const logoFontSize = 18
	const logoFontWeight = 300

	const headingFontSize = 48
	const headingFontWeight = 700
	const headingLineHeight = headingFontSize * 1.45

	const textFontSize = 22
	const textFontWeight = 300
	const textLineHeight = textFontSize * 1.5

	let headingPosition = null

	const canvas = createCanvas(width, height)
	const context = canvas.getContext('2d')

	// BG IMAGE
	context.drawImage(await bg, 0, 0, canvas.width, canvas.height)

	// BLACK BG OVER BG IMAGE
	context.fillStyle = 'rgba(255, 255, 255, 0.5)'
	context.fillRect(0, 0, canvas.width, canvas.height)

	// LOGO
	context.drawImage(await darkLogo, margin, margin, 32, 32)
	context.font = `${logoFontWeight} ${logoFontSize}pt Inter`
	context.fillStyle = 'rgba(0, 0, 0, 0.8)'
	context.textBaseline = 'top'

	// TITLE
	if (title) {
		context.font = `${headingFontWeight} ${headingFontSize}pt Inter`
		context.fillStyle = '#004dff'
		context.textAlign = 'center'
		context.textBaseline = 'top'
		headingPosition = wrapText(
			context,
			Array.isArray(title) ? title[0] : title,
			canvas.width / 2,
			squareY + margin + 170,
			canvas.width / 1.25,
			headingLineHeight
		)
	}

	// DESCRIPTION
	if (description) {
		context.font = `${textFontWeight} ${textFontSize}pt Inter`
		context.fillStyle = '#004dff'
		context.textAlign = 'center'
		context.textBaseline = 'top'
		wrapText(
			context,
			Array.isArray(description) ? description[0] : description,
			canvas.width / 2,
			headingPosition
				? headingPosition.y + headingLineHeight * 2
				: squareY + margin + 170 + headingLineHeight * 2,
			canvas.width / 1.25,
			textLineHeight
		)
	}

	const buffer = canvas.toBuffer('image/png')
	return buffer
}

async function computeProjectImage({
	title,
	description,
	image,
	notAuthor,
}: {
	title?: string
	description?: string
	image?: Image
	notAuthor: boolean | null
}): Promise<Buffer> {
	const width = 1200
	const height = 630
	const margin = 40
	const squareX = 0
	const squareY = 0

	const logoFontSize = 18
	const logoFontWeight = 300
	const logoLineHeight = logoFontWeight * 1.45

	const headingFontSize = 48
	const headingFontWeight = 700
	const headingLineHeight = headingFontSize * 1.45

	const textFontSize = 22
	const textFontWeight = 300
	const textLineHeight = textFontSize * 1.5

	const footerFontSize = 20
	const footerFontWeight = 300
	const footerLineHeight = footerFontWeight * 1.5

	let headingPosition = null

	const canvas = createCanvas(width, height)
	const context = canvas.getContext('2d')

	context.imageSmoothingEnabled = true

	// BG
	if (image) {
		drawImageProp(context, image, 0, 0, canvas.width, canvas.height)
	} else {
		context.drawImage(await bg, 0, 0, canvas.width, canvas.height)
	}

	// BLACK BG OVER BG IMAGE
	context.fillStyle = 'rgba(0, 0, 0, 0.86)'
	context.fillRect(squareX, squareY, canvas.width / 2, canvas.height)

	// LOGO
	context.drawImage(await logo, margin, margin, 32, 32)
	context.font = `${logoFontWeight} ${logoFontSize}pt Inter`
	context.fillStyle = 'rgba(255, 255, 255, 0.8)'
	context.textBaseline = 'top'
	wrapText(
		context,
		'pennanen.dev',
		margin + 32 + 10,
		margin - 2,
		canvas.width / 2 - 2 * margin,
		logoLineHeight
	)

	// TITLE
	if (title) {
		context.font = `${headingFontWeight} ${headingFontSize}pt Inter`
		context.fillStyle = '#4980ff'
		context.textAlign = 'left'
		context.textBaseline = 'top'
		headingPosition = wrapText(
			context,
			Array.isArray(title) ? title[0] : title,
			squareX + margin,
			squareY + margin + 150,
			canvas.width / 2 - 2 * margin,
			headingLineHeight
		)
	}

	// DESCRIPTION
	if (description) {
		context.font = `${textFontWeight} ${textFontSize}pt Inter`
		context.fillStyle = '#9fbcff'
		context.textAlign = 'left'
		context.textBaseline = 'top'
		wrapText(
			context,
			Array.isArray(description) ? description[0] : description,
			squareX + margin,
			headingPosition
				? headingPosition.y + headingLineHeight * 1.4
				: squareY + margin + 150 + headingLineHeight * 1.4,
			canvas.width / 2 - 2 * margin,
			textLineHeight
		)
	}

	// FOOTER
	if (!notAuthor) {
		context.font = `${footerFontWeight} ${footerFontSize}pt Inter`
		context.fillStyle = 'rgba(255, 255, 255, 0.95)'
		context.textAlign = 'left'
		context.textBaseline = 'top'
		wrapText(
			context,
			'Arttu Pennanen',
			margin,
			canvas.height - 2.35 * margin,
			canvas.width / 2 - margin,
			footerLineHeight
		)
	}

	const buffer = canvas.toBuffer('image/jpeg')
	return buffer
}

export async function generateMainPageImage<
	T extends { title: string; description?: string },
>(data: T): Promise<string> {
	const target = './public/meta'
	try {
		fs.mkdirSync(target)
	} catch {
		//
	}

	const image = await computeMainPageImage({
		title: data.title || 'pennanen.dev',
		description: data.description,
	})

	// eslint-disable-next-line no-useless-escape
	const fileName =
		data.title.replace(/[\|\ \.\-]/g, '_').toLowerCase() + '.jpg'
	const directory = path.join(target, fileName)

	fs.writeFileSync(directory, image)

	return fileName
}

export async function generateProjectImage(project: IProject) {
	const target = './public/meta'
	if (!fs.existsSync(target)) {
		fs.mkdirSync(target)
	}

	let icon = null
	if (project.largeImage) {
		try {
			const imageFile = fs.readFileSync(
				path.join(
					process.cwd(),
					'public/sub/' + project.id + '/' + project.largeImage
				)
			)
			icon = await loadImage(imageFile)
		} catch {
			//
		}
	}

	const image = await computeProjectImage({
		title: project.name || project.id,
		description: project.description as string,
		image: icon || undefined,
		notAuthor: project.notAuthor,
	})

	const fileName = project.id.toLowerCase() + '.png'
	const directory = path.join(target, fileName)
	fs.writeFileSync(directory, image)

	return fileName
}
