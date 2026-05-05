import BackgroundLines from '../BackgroundLines'
import styles from './room.module.css'

const Room = () => {
  return (
    <>
      <div aria-hidden="true" className={`${styles.wall} ${styles.back}`}>
        <BackgroundLines />
      </div>
      <div aria-hidden="true" className={`${styles.wall} ${styles.left}`}>
        <BackgroundLines />
      </div>
      <div aria-hidden="true" className={`${styles.wall} ${styles.right}`}>
        <BackgroundLines />
      </div>
      <div aria-hidden="true" className={`${styles.wall} ${styles.floor}`}>
        <BackgroundLines />
      </div>
      <div aria-hidden="true" className={`${styles.wall} ${styles.ceiling}`}>
        <BackgroundLines />
      </div>
    </>
  )
}

export default Room
