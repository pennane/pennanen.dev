import Link from 'next/link'
import Image from 'next/image'
import style from '../styles/navigation-bar.module.css'

const NavigationBar = () => {
    return (
        <div className={style['container']}>
            <nav className={style['main']}>
                <Link href="/">
                    <a className={style['brand'] + ' ' + style['link']}>
                        <div className={style['brand-image']}>
                            <Image src="/images/black-icon.png" alt="" width={20} height={20} />
                        </div>

                        <span className="title">pennanen.dev</span>
                    </a>
                </Link>
            </nav>
        </div>
    )
}

export default NavigationBar
