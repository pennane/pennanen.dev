import style from '../styles/main-header.module.css'
import Image from 'next/image'

const MainHeading = () => {
    return (
        <div className={style['container']}>
            <div className={style['gradient']}>
                <Image height={350} width={350} src="/images/gradient.png" alt="multicolored splash" />
            </div>
            <div className={style['heading']}>
                <h1>Welcome</h1>
            </div>
        </div>
    )
}

export default MainHeading
