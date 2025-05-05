import Image from 'next/image'
import { A } from '../components/A'
import { Gap, Stack } from '../components/Stack'
import style from './not-found.module.css'

export default function NotFound() {
  return (
    <Stack className={style.main} gap={Gap.none}>
      <header>
        <Stack className={style['header']}>
          <h1>404 - Not found</h1>
        </Stack>
      </header>
      <Stack>
        <A href="/" className={style['link']}>
          To home
        </A>
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
