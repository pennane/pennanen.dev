import style from './contact.module.css'
import { FontAwesomeIcon } from '@fortawesome/react-fontawesome'
import { faGithub, faLinkedinIn } from '@fortawesome/free-brands-svg-icons'

export default function ContactLinks() {
    return (
        <div className={style.container}>
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
        </div>
    )
}
