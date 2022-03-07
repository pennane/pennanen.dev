import Image from 'next/image'
import Layout from '../components/layout'
import style from '../styles/error.module.css'
import Link from 'next/link'

const Error = () => {
    return (
        <Layout description="Arttu Pennanen Web-like projects">
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
