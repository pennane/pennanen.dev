import style from '../styles/main-header.module.css'
import Image from 'next/image'
import SequentialAnimation from './sequential-animation'
import { useState } from 'react'

const MainHeading = () => {
    const [splashImageLoaded, setSplashImageLoaded] = useState(false)

    return (
        <div className={style['container']}>
            <SequentialAnimation
                animationKey="header"
                childClass="fade-in"
                once={true}
                initialDelay={100}
                delayBetween={200}
                animationDuration={800}
                stopped={!splashImageLoaded}
            >
                <div className={style['gradient']}>
                    <Image
                        height={350}
                        width={350}
                        src="/images/gradient.png"
                        alt="multicolored splash"
                        onLoadingComplete={() => setSplashImageLoaded(true)}
                    />
                </div>
                <div className={style['heading']}>
                    <h1>Welcome</h1>
                </div>
            </SequentialAnimation>
        </div>
    )
}

export default MainHeading
