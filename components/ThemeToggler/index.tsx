import { useTheme } from 'next-themes'
import style from './themetoggler.module.css'
import { faMoon } from '@fortawesome/free-regular-svg-icons'
import { FontAwesomeIcon } from '@fortawesome/react-fontawesome'
import { Stack } from '../Stack'

const ThemeToggler = () => {
	const { resolvedTheme, setTheme } = useTheme()
	const isDark = resolvedTheme === 'dark'

	const toggleTheme = () => {
		document.documentElement.classList.add('theme-transition')
		setTimeout(() => {
			document.documentElement.classList.remove('theme-transition')
		}, 1000)
		setTheme(isDark ? 'light' : 'dark')
	}

	return (
		<Stack
			className={style.main}
			direction="row"
			alignItems="center"
			justifyContent="flex-end"
		>
			<button
				onClick={toggleTheme}
				className={style.button}
				aria-label={resolvedTheme || 'auto'}
				aria-live="polite"
				title="Switch between light and dark theme"
			>
				<FontAwesomeIcon icon={faMoon} />
			</button>
		</Stack>
	)
}

export default ThemeToggler
