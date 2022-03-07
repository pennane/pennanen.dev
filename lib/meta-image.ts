import path from 'path'
import fs from 'fs'
import { createCanvas, loadImage, registerFont } from 'canvas'
import { wrapText } from './util'
import { Project } from '../types'

registerFont(path.join(process.cwd(), 'public/fonts/Inter-Medium.ttf'), { family: 'Inter', weight: '300' })
registerFont(path.join(process.cwd(), 'public/fonts/Inter-Bold.ttf'), { family: 'Inter', weight: '700' })

const logoPath = fs.readFileSync(path.join(process.cwd(), 'public/icons/icon-white-128.png'))
const logo = loadImage(logoPath)
const bgPath = fs.readFileSync(path.join(process.cwd(), 'public/images/metabg.png'))
const bg = loadImage(bgPath)
const facePath = fs.readFileSync(path.join(process.cwd(), 'public/images/hl_pennanen.png'))
const face = loadImage(facePath)

async function computeImage({ title, description }: { title?: string; description?: string }): Promise<Buffer> {
    let width = 1200
    let height = 630
    let margin = 40
    let squareX = 0
    let squareY = 0

    let logoFontSize = 18
    let logoFontWeight = 300
    let logoLineHeight = logoFontWeight * 1.45

    let headingFontSize = 44
    let headingFontWeight = 700
    let headingLineHeight = headingFontSize * 1.45

    let textFontSize = 20
    let textFontWeight = 300
    let textLineHeight = textFontSize * 1.5

    let footerFontSize = 20
    let footerFontWeight = 300
    let footerLineHeight = footerFontWeight * 1.5

    let headingPosition = null

    const canvas = createCanvas(width, height)
    const context = canvas.getContext('2d')

    // BG IMAGE
    context.drawImage(await bg, 0, 0, canvas.width, canvas.height)

    // BLUE BG OVER BG IMAGE
    context.fillStyle = 'rgba(0, 0, 0, 0.86)'
    context.fillRect(squareX, squareY, canvas.width / 2, canvas.height)

    // LOGO
    context.drawImage(await logo, margin, margin, 32, 32)
    context.font = `${logoFontWeight} ${logoFontSize}pt Inter`
    context.fillStyle = 'rgba(255, 255, 255, 0.8)'
    context.textBaseline = 'top'
    wrapText(context, 'pennanen.dev', margin + 32 + 10, margin - 2, canvas.width / 2 - 2 * margin, logoLineHeight)

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
            squareY + margin + 170,
            canvas.width / 2 - margin,
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
                : squareY + margin + 170 + headingLineHeight * 1.4,
            canvas.width / 2 - margin,
            textLineHeight
        )
    }

    // FOOTER

    context.font = `${footerFontWeight} ${footerFontSize}pt Inter`
    context.fillStyle = 'rgba(255, 255, 255, 0.95)'
    context.textAlign = 'left'
    context.textBaseline = 'top'
    context.drawImage(await face, margin, canvas.height - 3 * margin, 80, 80)
    wrapText(
        context,
        'Arttu Pennanen',
        squareX + margin + 80 + margin / 2,
        canvas.height - 2.35 * margin,
        canvas.width / 2 - margin,
        footerLineHeight
    )
    const buffer = canvas.toBuffer('image/png')
    return buffer
}

export async function generateImage(project: Project) {
    const target = './public/meta'
    if (!fs.existsSync(target)) {
        fs.mkdirSync(target)
    }
    const image = await computeImage({ title: project.name || project.id, description: project.description as string })
    const fileName = project.id.toLowerCase() + '.png'
    const directory = path.join(target, fileName)
    fs.writeFileSync(directory, image)

    return fileName
}
