import styles from './footer.module.css'
const Footer = () => {
  return (
    <footer className={styles.footer}>
      <p>&copy; 2018 – {new Date().getFullYear()} Arttu Pennanen</p>
    </footer>
  )
}

export default Footer
