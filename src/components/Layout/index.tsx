import Navbar from './Navbar'
import Footer from './Footer'
import { Gap, Stack } from '../Stack'
import styles from './layout.module.css'
import BackgroundLines from './BackgroundLines'

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
