import style from '../styles/main-header.module.css'

const MainHeading = () => {
    return (
        <div className={style['container']}>
            <div className={style['gradient']}>
                <img src="/images/gradient.png" />
            </div>
            <div className={style['heading']}>
                <h1>Welcome</h1>
            </div>
        </div>
    )
}

export default MainHeading
