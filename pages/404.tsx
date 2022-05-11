import Image from 'next/image'
import Layout from '../components/Layout'
import style from '../styles/error.module.css'
import Link from 'next/link'

function Error() {
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
                <Image layout="fill" src="/images/starrynight.jpg" alt=""></Image>
            </div>
        </Layout>
    )
}

export default Error
