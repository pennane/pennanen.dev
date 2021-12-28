import Link from 'next/link'
import style from '../styles/navigation-bar.module.css'

const NavigationBar = () => {
    return (
        <div className={style['container']}>
            <nav className={style['main']}>
                <Link href="/">
                    <a className={style['brand'] + ' ' + style['link']}>
                        <img src="/images/black-icon.png" alt="" />
                        <span className="title">Arttu Pennanen</span>
                    </a>
                </Link>
            </nav>
        </div>
    )
}

export default NavigationBar
