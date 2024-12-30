import { Stack } from '../Stack'
import style from './footer.module.css'

const Footer = () => {
	const start = 2018
	const end = new Date().getFullYear()

	return (
		<footer className={style.main}>
			<Stack direction="row" alignItems="center" justifyContent="center">
				<span>&copy; {`${start} - ${end}`} Arttu Pennanen </span>
			</Stack>
		</footer>
	)
}

export default Footer
