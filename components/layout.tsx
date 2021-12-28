import Head from 'next/head'
import BackgroundLines from './background-lines'
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
                        <link rel="icon" href="/favicon.ico" />
                    </>
                )}
            </Head>
            <div className="page-container">
                <BackgroundLines />
                <NavigationBar />
                <main>{children}</main>
            </div>
        </div>
    )
}

export default Layout
