import Link from 'next/link'
import { Stack } from '../Stack'
import style from './footer.module.css'

const Footer = () => {
	const start = 2018
	const end = new Date().getFullYear()

	return (
		<footer className={style.main}>
			<Stack direction="row" alignItems="center" justifyContent="center">
				<span>
					<Link href="https://github.com/Pennane/pennanen.dev">
						<a>Github</a>
					</Link>
				</span>
				<span>&copy; Arttu Pennanen {`${start} - ${end}`}</span>
			</Stack>
		</footer>
	)
}

export default Footer
