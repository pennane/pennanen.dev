import { useTheme } from 'next-themes'
import style from '../styles/theme-toggler.module.css'
import { faMoon } from '@fortawesome/free-regular-svg-icons'
import { FontAwesomeIcon } from '@fortawesome/react-fontawesome'

const ThemeToggler = () => {
    const { theme, setTheme } = useTheme()

    let timeout: NodeJS.Timeout

    const toggleTheme = () => {
        document.documentElement.classList.add('theme-transition')
        clearTimeout(timeout)
        timeout = setTimeout(() => {
            document.documentElement.classList.remove('theme-transition')
        }, 1000)
        setTheme(theme === 'dark' ? 'light' : 'dark')
    }

    return (
        <div className={style['main']} onClick={() => toggleTheme()}>
            <button
                className={theme === 'dark' ? `${style['button']} ${style['dark']}` : style['button']}
                aria-label="Toggle darkmode"
                title="Toggle darkmode"
            >
                <FontAwesomeIcon icon={faMoon} />
            </button>
        </div>
    )
}

export default ThemeToggler
