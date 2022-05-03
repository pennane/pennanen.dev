import Link from 'next/link'
import Image from 'next/image'
import style from './navigationbar.module.css'
import { useTheme } from 'next-themes'
import dynamic from 'next/dynamic'
import ContactLinks from '../ContactLinks'

const ThemeToggler = dynamic(() => import('../ThemeToggler'), { ssr: false })

const NavigationBar = () => {
    const { resolvedTheme } = useTheme()

    return (
        <div className={style['container']}>
            <nav className={style['main']}>
                <Link href="/">
                    <a className={style['link']}>
                        <div className={style['brand-image']}>
                            {resolvedTheme === 'dark' && (
                                <Image src={'/images/white-icon.png'} alt="" width={20} height={20} />
                            )}
                            {resolvedTheme !== 'dark' && (
                                <Image src={'/images/black-icon.png'} alt="" width={20} height={20} />
                            )}
                        </div>
                        <span className={style['title']}>pennanen.dev</span>
                    </a>
                </Link>
                <ContactLinks />
                <ThemeToggler />
            </nav>
        </div>
    )
}

export default NavigationBar
