import Layout from '../../components/Layout'

export default function SiteLayout({ children }: Readonly<{ children: React.ReactNode }>) {
  return <Layout>{children}</Layout>
}
