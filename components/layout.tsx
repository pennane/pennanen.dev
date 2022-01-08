import Head from 'next/head'
import BackgroundLines from './background-lines'
import Footer from './footer'
import NavigationBar from './navigation-bar'

export const siteTitle = 'Arttu Pennanen'

const Layout = ({
    children,
    title,
    description
}: {
    children: React.ReactNode
    title?: string
    description?: string
}) => {
    return (
        <div className="layout-container">
            <Head>
                <title>{title ? `${title} | ${siteTitle}` : siteTitle}</title>
                {description && (
                    <>
                        <meta name="og:description" content={description} />
                        <meta name="description" content={description} />
                    </>
                )}
                <meta name="theme-color" content="#1c4ed8"></meta>
                <link rel="icon" href="/favicon.ico" />
                <link rel="apple-touch-icon" href="/images/apple-touch-icon.png" />
            </Head>
            <div className="page-container">
                <BackgroundLines />
                <NavigationBar />
                <main>{children}</main>
                <Footer />
            </div>
        </div>
    )
}

export default Layout
