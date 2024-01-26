import Image from 'next/image'
import Layout from '../components/Layout'
import style from '../styles/error.module.css'
import Link from 'next/link'
import { useState, useRef, useEffect } from 'react'

function Error() {
  const [imageLoaded, setImageLoaded] = useState(false)
  const mounted = useRef(false)

  useEffect(() => {
    mounted.current = true
    return () => {
      mounted.current = false
    }
  }, [])

  return (
    <Layout className="pt0">
      <header className={style['header']}>
        <h1>404 - Not found</h1>
      </header>
      <section className={style['link-wrapper']}>
        <Link href="/">
          <a className={style['link']}>To home</a>
        </Link>
      </section>
      <div className={style['background']}>
        <Image
          className={
            imageLoaded
              ? style['image']
              : `${style['image']} ${style['loading']}`
          }
          layout="fill"
          src="/images/starrynight.jpg"
          alt=""
          loading="lazy"
          onLoadingComplete={() => mounted.current && setImageLoaded(true)}
        ></Image>
      </div>
    </Layout>
  )
}

export default Error
