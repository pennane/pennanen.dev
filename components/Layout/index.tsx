import Head from 'next/head'
import BackgroundLines from '../BackgroundLines'
import Footer from '../Footer'
import NavigationBar from '../NavigationBar'
import React from 'react'
import style from './layout.module.css'
import { Stack, TStackProps } from '../Stack'

const siteTitle = 'Arttu Pennanen'

const LayoutHead = ({
	title,
	description,
	metaImage,
}: {
	title?: string
	description?: string
	metaImage?: string
}) => (
	<Head>
		{title && (
			<>
				<title>{`${title} | ${siteTitle}`}</title>
				<meta name="title" content={`${title} | ${siteTitle}`} />
				<meta property="og:title" content={`${title} | ${siteTitle}`} />
				<meta
					name="twitter:title"
					content={`${title} | ${siteTitle}`}
				/>
			</>
		)}
		{!title && (
			<>
				<title>{siteTitle}</title>
				<meta name="title" content={siteTitle} />
				<meta property="og:title" content={siteTitle} />
				<meta name="twitter:title" content={siteTitle} />
			</>
		)}
		{description && (
			<>
				<meta name="description" content={description} />
				<meta property="og:description" content={description} />
				<meta name="twitter:description" content={description} />
			</>
		)}

		<meta property="twitter:card" content="summary_large_image" />

		{metaImage && (
			<>
				<meta property="og:image" content={'/meta/' + metaImage} />
				<meta property="twitter:image" content={'/meta/' + metaImage} />
			</>
		)}
		{!metaImage && (
			<>
				<meta
					property="og:image"
					content="https://pennanen.dev/logo.jpg"
				/>
				<meta
					property="twitter:image"
					content="https://pennanen.dev/logo.jpg"
				/>
			</>
		)}
		<meta property="twitter:url" content="https://pennanen.dev/" />
		<meta name="theme-color" content="#011e17"></meta>
		<meta name="robots" content="index, follow"></meta>
		<meta name="author" content="Arttu Pennanen"></meta>
		<link rel="icon" href="/favicon.ico" />
		<link rel="apple-touch-icon" href="/images/apple-touch-icon.png" />
	</Head>
)

const LayoutBody = ({ children, ...rest }: TStackProps) => (
	<Stack className={style['page']} alignItems="center">
		<BackgroundLines />
		<NavigationBar />
		<Stack className={`${style['main']} ${rest.className || ''}`} {...rest}>
			{children}
		</Stack>
		<Footer />
	</Stack>
)

const Layout = ({
	children,
	title,
	description,
	metaImage,
	wrapperClassName,
	...rest
}: {
	children: React.ReactNode
	title?: string
	description?: string
	metaImage?: string
	wrapperClassName?: string
} & TStackProps) => {
	const wrapperClass = [style['container'], wrapperClassName]
		.filter(Boolean)
		.join(' ')

	return (
		<Stack className={wrapperClass}>
			<LayoutHead
				description={description}
				title={title}
				metaImage={metaImage}
			/>
			<LayoutBody {...rest}>{children}</LayoutBody>
		</Stack>
	)
}

export default Layout
