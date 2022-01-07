import Link from 'next/link'
import Image from 'next/image'
import style from '../styles/footer.module.css'

const Footer = () => {
    const start = 2018
    const end = new Date().getFullYear()

    return (
        <div className={style['wrapper']}>
            <footer className={style['main']}>
                <span>&copy; Arttu Pennanen {`${start} - ${end}`}</span>
                {/* <span>
                    <Link href="https://github.com/Pennane">
                        <a className={style['icon-link']}>
                            <Image src="/images/GitHub-Mark-32px.png" alt="Github Logomark" width={24} height={24} />
                        </a>
                    </Link>
                </span> */}
            </footer>
        </div>
    )
}

export default Footer
