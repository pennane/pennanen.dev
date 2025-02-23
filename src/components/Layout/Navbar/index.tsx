import Link from 'next/link'
import { Stack } from '../../Stack'
import Image from 'next/image'

const Navbar = () => {
  return (
    <nav>
      <Stack direction="row" gap={'1rem'} align="center">
        <Link href="/">
          <Stack direction="row" align="center">
            <Image
              alt=""
              src={'/icons/favicon-96.png'}
              height={18}
              width={18}
            />
            <span>pennanen.dev</span>
          </Stack>
        </Link>
        <Link href="/entries">entries</Link>
        <Link href="/projects">projects</Link>
      </Stack>
    </nav>
  )
}

export default Navbar
