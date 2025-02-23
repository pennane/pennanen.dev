import Link from 'next/link'
import { Stack } from '../../Stack'

const Navbar = () => {
  return (
    <nav>
      <Stack direction="row" gap={'1rem'}>
        <Link href="/">pennanen.dev</Link>
        <Link href="/blog">Blog</Link>
      </Stack>
    </nav>
  )
}

export default Navbar
