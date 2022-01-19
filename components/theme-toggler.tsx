import { useTheme } from 'next-themes'
import style from '../styles/theme-toggler.module.css'
import { faMoon } from '@fortawesome/free-regular-svg-icons'
import { FontAwesomeIcon } from '@fortawesome/react-fontawesome'

const ThemeToggler = () => {
    const { resolvedTheme, theme, setTheme } = useTheme()
    const isDark = resolvedTheme === 'dark'
    let label

    if (theme === 'dark') {
        label = 'dark'
    } else if (theme === 'light') {
        label = 'light'
    } else {
        label = 'auto'
    }

    let timeout: NodeJS.Timeout

    const toggleTheme = () => {
        document.documentElement.classList.add('theme-transition')
        clearTimeout(timeout)
        timeout = setTimeout(() => {
            document.documentElement.classList.remove('theme-transition')
        }, 1000)
        setTheme(isDark ? 'light' : 'dark')
    }

    return (
        <div className={style['main']} onClick={() => toggleTheme()}>
            <button className={style['button']} aria-label={label} title="Toggles between light and dark theme">
                <FontAwesomeIcon icon={faMoon} />
            </button>
        </div>
    )
}

export default ThemeToggler
