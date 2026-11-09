import Image, { StaticImageData } from 'next/image'
import { isUnoptimizable } from '../../lib/images'
import styles from './media.module.css'

export const MdxImage = ({
  src,
  alt
}: {
  src?: StaticImageData | string
  alt?: string
}) => {
  if (!src) return null
  if (typeof src !== 'string') {
    return <Image className={styles.media} src={src} alt={alt ?? ''} />
  }
  // Public and remote images have no known size, so they fill a block and keep their ratio
  return (
    <span className={styles.fill}>
      <Image
        src={src}
        alt={alt ?? ''}
        fill
        sizes="(max-width: 48rem) 100vw, 48rem"
        unoptimized={isUnoptimizable(src)}
      />
    </span>
  )
}

export const Video = ({ src }: { src: string }) => (
  <video className={styles.media} src={src} muted autoPlay loop playsInline />
)

export const Embed = ({ src, title }: { src: string; title?: string }) => (
  <iframe
    className={styles.embed}
    src={src}
    title={title ?? 'Embedded demo'}
    loading="lazy"
    allowFullScreen
  />
)
