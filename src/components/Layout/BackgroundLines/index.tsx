import style from './background-lines.module.css'

const BackgroundLines = () => {
  return (
    <div className={style['container']} aria-hidden="true">
      <div className={style['item']}></div>
      <div className={style['item']}></div>
      <div className={style['item']}></div>
      <div className={style['item']}></div>
      <div className={style['item']}></div>
    </div>
  )
}

export default BackgroundLines
