import Link from 'next/link'
import style from '../styles/footer.module.css'

const Footer = () => {
    const start = 2018
    const end = new Date().getFullYear()

    return (
        <div className={style['wrapper']}>
            <footer className={style['main']}>
                <span>
                    <Link href="https://github.com/Pennane">
                        <a className={style['icon-link']}>Github</a>
                    </Link>
                </span>
                <span>&copy; Arttu Pennanen {`${start} - ${end}`}</span>
            </footer>
        </div>
    )
}

export default Footer
