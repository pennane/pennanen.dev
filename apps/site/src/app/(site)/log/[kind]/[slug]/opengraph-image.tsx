import { readFile } from 'fs/promises'
import { ImageResponse } from 'next/og'
import path from 'path'
import {
  isPublicKind,
  pictureOf,
  shownScores,
  titleOf
} from '../../../../../lib/log/present'
import { logPublic } from '../../../../../lib/log/public-api'

export const size = { width: 1200, height: 630 }
export const contentType = 'image/png'
export const alt = 'An item in the log of things as ap saw them'

const PANEL_WIDTH = 600
const PADDING = 48
const LOGO_SIZE = 32
const PUBLIC_DIR = path.join(process.cwd(), 'public')

const publicFile = (...parts: string[]) =>
  readFile(path.join(PUBLIC_DIR, ...parts))

const font = (file: string) => publicFile('fonts', file)

const pngDataUrl = async (...parts: string[]) =>
  `data:image/png;base64,${(await publicFile(...parts)).toString('base64')}`

export default async function OpenGraphImage({
  params
}: {
  params: Promise<{ kind: string; slug: string }>
}) {
  const { kind, slug } = await params
  const [item, presentation] = await Promise.all([
    isPublicKind(kind) ? logPublic.item(kind, slug) : undefined,
    logPublic.presentation()
  ])
  const [regular, medium, bold, logo, fallbackBackground] = await Promise.all([
    font('Inter-Regular.ttf'),
    font('Inter-Medium.ttf'),
    font('Inter-Bold.ttf'),
    pngDataUrl('icons', 'icon-white-128.png'),
    pngDataUrl('images', 'metabg.png')
  ])
  const title = item ? titleOf(item) : 'things as ap saw them'
  const blurb = item?.blurb
  const footer = item
    ? shownScores(presentation, kind, item.scores)
        .map(({ axisName, division }) => `${axisName} ${division}`)
        .join(' · ')
    : ''

  return new ImageResponse(
    (
      <div style={{ display: 'flex', width: '100%', height: '100%', position: 'relative' }}>
        <img
          alt=""
          src={(item && pictureOf(item)) ?? fallbackBackground}
          width={size.width}
          height={size.height}
          style={{ position: 'absolute', inset: 0, objectFit: 'cover' }}
        />
        <div
          style={{
            display: 'flex',
            flexDirection: 'column',
            justifyContent: 'space-between',
            width: PANEL_WIDTH,
            height: '100%',
            padding: PADDING,
            background: 'rgba(0, 0, 0, 0.86)',
            fontFamily: 'Inter'
          }}
        >
          <div style={{ display: 'flex', alignItems: 'center', gap: 12 }}>
            <img alt="" src={logo} width={LOGO_SIZE} height={LOGO_SIZE} />
            <span style={{ fontSize: 24, fontWeight: 500, color: 'rgba(255, 255, 255, 0.8)' }}>
              pennanen.dev
            </span>
          </div>
          <div style={{ display: 'flex', flexDirection: 'column', gap: 24 }}>
            <span
              style={{
                fontSize: 60,
                fontWeight: 700,
                lineHeight: 1.07,
                color: '#4980ff',
                lineClamp: 3
              }}
            >
              {title}
            </span>
            {blurb && (
              <span style={{ fontSize: 28, lineHeight: 1.36, color: '#9fbcff', lineClamp: 5 }}>
                {blurb}
              </span>
            )}
          </div>
          <span style={{ fontSize: 24, fontWeight: 500, color: 'rgba(255, 255, 255, 0.9)' }}>
            {footer}
          </span>
        </div>
      </div>
    ),
    {
      ...size,
      fonts: [
        { name: 'Inter', data: regular, weight: 400 },
        { name: 'Inter', data: medium, weight: 500 },
        { name: 'Inter', data: bold, weight: 700 }
      ]
    }
  )
}
