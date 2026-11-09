import { Gap, Stack } from '../Stack'
import BackgroundLines from './BackgroundLines'
import Footer from './Footer'
import styles from './layout.module.css'
import Navbar from './Navbar'

const Layout = ({ children }: { children: React.ReactNode }) => {
  return (
    <Stack alignItems="center" className={styles.wrapper}>
      <BackgroundLines />
      <Stack className={styles.layout} gap={Gap.xLarge}>
        <Navbar />
        <main className={styles.main}>{children}</main>
        <Footer />
      </Stack>
    </Stack>
  )
}

export default Layout
