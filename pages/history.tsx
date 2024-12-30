import type { InferGetStaticPropsType } from 'next'
import Layout from '../components/Layout'
import fs from 'fs'
import path from 'path'
import Image from 'next/image'

const description = 'Site design history'

export const getStaticProps = async () => {
	const images = fs
		.readdirSync(path.join((process.cwd(), 'public/history')))
		.filter((i) => i.toLowerCase().endsWith('.png'))
		.sort()
	return {
		props: {
			images,
		},
	}
}

const All = ({ images }: InferGetStaticPropsType<typeof getStaticProps>) => {
	return (
		<Layout
			description={description}
			className="design-history"
			pageClassName="design-history-page"
		>
			<section>
				<h2>Design history</h2>
				<div className="history-images">
					{images.map((url, i) => (
						<Image
							width={320}
							height={566}
							key={i}
							src={'/history/' + url}
							loading="lazy"
						></Image>
					))}
				</div>
			</section>
		</Layout>
	)
}

export default All
