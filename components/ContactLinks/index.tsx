import style from './contact.module.css'
import { FontAwesomeIcon } from '@fortawesome/react-fontawesome'
import { faGithub, faLinkedinIn } from '@fortawesome/free-brands-svg-icons'
import { Stack } from '../Stack'

export default function ContactLinks() {
	return (
		<Stack gap="none" alignItems="center" direction="row">
			<a
				href="https://www.linkedin.com/in/arttu-pennanen/"
				aria-label="LinkedIn"
				title="LinkedIn"
				target="_blank"
				rel="noreferrer"
				className={style.link}
			>
				<FontAwesomeIcon icon={faLinkedinIn} />
			</a>
			<a
				href="https://github.com/pennane"
				aria-label="Github"
				title="Github"
				target="_blank"
				rel="noreferrer"
				className={style.link}
			>
				<FontAwesomeIcon icon={faGithub} />
			</a>
		</Stack>
	)
}
