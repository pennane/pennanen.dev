import Navbar from './Navbar'
import Footer from './Footer'
import { Stack } from '../Stack'
import styles from './layout.module.css'

const Layout = ({ children }: { children: React.ReactNode }) => {
  return (
    <Stack align="center" className={styles.wrapper}>
      <Stack className={styles.layout} gap="2rem">
        <Navbar />
        <main className={styles.main}>{children}</main>
        <Footer />
      </Stack>
    </Stack>
  )
}

export default Layout
