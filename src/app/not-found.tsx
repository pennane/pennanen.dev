import Image from 'next/image'
import Link from 'next/link'
import style from './not-found.module.css'
import { Stack } from '../components/Stack'

export default function NotFound() {
  return (
    <Stack className={style.main} gap="0">
      <header>
        <Stack className={style['header']}>
          <h1>404 - Not found</h1>
        </Stack>
      </header>
      <Stack>
        <Link href="/" className={style['link']}>
          To home
        </Link>
      </Stack>
      <div className={style['background']}>
        <Image
          className={[style['image']].filter(Boolean).join(' ')}
          src="/images/starrynight.jpg"
          alt=""
          loading="lazy"
          fill
          sizes="100vw"
        />
      </div>
    </Stack>
  )
}
