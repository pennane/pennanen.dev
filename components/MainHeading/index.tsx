import style from './mainheading.module.css'
import Image from 'next/image'
import SequentialAnimation from '../SequentialAnimation'
import { useState } from 'react'

const MainHeading = ({ text }: { text?: string }) => {
    const [splashImageLoaded, setSplashImageLoaded] = useState(false)

    return (
        <div className={style['container']}>
            <SequentialAnimation
                animationKey="header"
                once={true}
                initialDelay={0}
                delayBetween={200}
                animationDuration={800}
                stopped={!splashImageLoaded}
            >
                <div className={style['gradient']}>
                    <div className={style['image-wrapper']}>
                        <Image
                            height={350}
                            width={350}
                            src="/images/gradient.png"
                            alt=""
                            onLoadingComplete={() => setSplashImageLoaded(true)}
                        />
                    </div>
                </div>
                <div className={style['heading']}>
                    <h1>{text || 'Welcome'}</h1>
                </div>
            </SequentialAnimation>
        </div>
    )
}

export default MainHeading
