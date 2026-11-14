import { readFile } from 'fs/promises'
import { ImageResponse } from 'next/og'
import path from 'path'

export const ogSize = { width: 1200, height: 630 }

const PANEL_WIDTH = 600
const PADDING = 48
const LOGO_SIZE = 32
const POSTER_COLUMNS = 3
const POSTER_ROWS = 2
const POSTER_WIDTH = (ogSize.width - PANEL_WIDTH) / POSTER_COLUMNS
const POSTER_HEIGHT = ogSize.height / POSTER_ROWS
export const POSTERS_ON_WALL = POSTER_COLUMNS * POSTER_ROWS
const PUBLIC_DIR = path.join(process.cwd(), 'public')

const publicFile = (...parts: string[]) => readFile(path.join(PUBLIC_DIR, ...parts))

const font = (file: string) => publicFile('fonts', file)

const pngDataUrl = async (...parts: string[]) =>
  `data:image/png;base64,${(await publicFile(...parts)).toString('base64')}`

export type Backdrop = { picture: string | undefined } | { posters: string[] }

const Wall = ({ posters }: { posters: string[] }) => (
  <div
    style={{
      display: 'flex',
      flexWrap: 'wrap',
      position: 'absolute',
      top: 0,
      right: 0,
      width: ogSize.width - PANEL_WIDTH,
      height: ogSize.height
    }}
  >
    {posters.slice(0, POSTERS_ON_WALL).map((poster, i) => (
      <img
        key={i}
        alt=""
        src={poster}
        width={POSTER_WIDTH}
        height={POSTER_HEIGHT}
        style={{ objectFit: 'cover' }}
      />
    ))}
  </div>
)

export const ogCard = async ({
  title,
  blurb,
  footer,
  backdrop
}: {
  title: string
  blurb: string | undefined
  footer: string
  backdrop: Backdrop
}) => {
  const [regular, medium, bold, logo, fallbackBackground] = await Promise.all([
    font('Inter-Regular.ttf'),
    font('Inter-Medium.ttf'),
    font('Inter-Bold.ttf'),
    pngDataUrl('icons', 'icon-white-128.png'),
    pngDataUrl('images', 'metabg.png')
  ])
  const wallIsFull = 'posters' in backdrop && backdrop.posters.length >= POSTERS_ON_WALL
  const picture = 'picture' in backdrop ? backdrop.picture : undefined

  return new ImageResponse(
    <div
      style={{
        display: 'flex',
        width: '100%',
        height: '100%',
        position: 'relative',
        background: '#000'
      }}
    >
      {wallIsFull ? (
        <Wall posters={backdrop.posters} />
      ) : (
        <img
          alt=""
          src={picture ?? fallbackBackground}
          width={ogSize.width}
          height={ogSize.height}
          style={{ position: 'absolute', inset: 0, objectFit: 'cover' }}
        />
      )}
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
    </div>,
    {
      ...ogSize,
      fonts: [
        { name: 'Inter', data: regular, weight: 400 },
        { name: 'Inter', data: medium, weight: 500 },
        { name: 'Inter', data: bold, weight: 700 }
      ]
    }
  )
}
