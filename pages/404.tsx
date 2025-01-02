import Image from 'next/image'
import Layout from '../components/Layout'
import Link from 'next/link'
import { useState, useRef, useEffect } from 'react'
import style from './404.module.css'
import { Stack } from '../components/Stack'

function Error() {
	const [imageLoaded, setImageLoaded] = useState(false)
	const mounted = useRef(false)

	useEffect(() => {
		mounted.current = true
		return () => {
			mounted.current = false
		}
	}, [])

	return (
		<Layout wrapperClassName="full" gap="large">
			<Stack className={style.main} gap="none">
				<header>
					<Stack className={style['header']}>
						<h1>404 - Not found</h1>
					</Stack>
				</header>
				<Stack>
					<Link href="/">
						<a className={style['link']}>To home</a>
					</Link>
				</Stack>
				<div className={style['background']}>
					<Image
						className={[
							style['image'],
							!imageLoaded && style['loading'],
						]
							.filter(Boolean)
							.join(' ')}
						layout="fill"
						src="/images/starrynight.jpg"
						alt=""
						loading="lazy"
						onLoadingComplete={() =>
							mounted.current && setImageLoaded(true)
						}
					/>
				</div>
			</Stack>
		</Layout>
	)
}

export default Error
