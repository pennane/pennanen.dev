import Link from 'next/link'
import Image from 'next/image'
import style from '../styles/navigation-bar.module.css'
import ThemeToggler from './theme-toggler'
import { useTheme } from 'next-themes'

const NavigationBar = () => {
    const { theme } = useTheme()

    return (
        <div className={style['container']}>
            <nav className={style['main']}>
                <Link href="/">
                    <a className={style['brand'] + ' ' + style['link']}>
                        <div className={style['brand-image']}>
                            {theme === 'dark' && <Image src={'/images/white-icon.png'} alt="" width={20} height={20} />}
                            {theme !== 'dark' && <Image src={'/images/black-icon.png'} alt="" width={20} height={20} />}
                        </div>
                        <span className="title">pennanen.dev</span>
                    </a>
                </Link>
                <ThemeToggler />
            </nav>
        </div>
    )
}

export default NavigationBar
