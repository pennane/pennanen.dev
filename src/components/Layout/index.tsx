import Navbar from './Navbar'
import Footer from './Footer'
import { Stack } from '../Stack'
import styles from './layout.module.css'

const Layout = ({ children }: { children: React.ReactNode }) => {
  return (
    <Stack className={styles.layout} gap="2rem">
      <Navbar />
      <main>{children}</main>
      <Footer />
    </Stack>
  )
}

export default Layout
