import type { Metadata } from 'next'
import { A } from '../../../components/A'
import { Gap, Stack } from '../../../components/Stack'
import { WrapInBox } from '../../../components/WrapInBox'
import { logPublic } from '../../../lib/log/public-api'
import { baseUrl } from '../../sitemap'
import { ItemList } from './components/ItemRow'
import styles from './components/log.module.css'
import { logOverview } from './components/overview'
import { logSummary } from './components/summaries'

const TITLE = 'things as ap saw them'

export async function generateMetadata(): Promise<Metadata> {
  const description = logSummary(await logOverview(await logPublic.presentation()))
  return {
    metadataBase: new URL(baseUrl),
    title: TITLE,
    description,
    openGraph: { title: TITLE, description, type: 'website', url: `${baseUrl}/log` },
    twitter: { card: 'summary_large_image', title: TITLE, description }
  }
}

export default async function LogPage() {
  const filled = await logOverview(await logPublic.presentation())

  return (
    <Stack gap={Gap.xLarge} alignItems="stretch">
      <h1>{TITLE}</h1>
      {filled.length === 0 && <p className={styles.empty}>Nothing here yet.</p>}
      {filled.map(({ kind, title, items, shown }) => (
        <section key={kind}>
          <Stack gap={Gap.large} alignItems="stretch">
            <header className={styles.sectionHeader}>
              <WrapInBox>
                <h3>{title}</h3>
              </WrapInBox>
              <A href={`/log/${kind}`} className={styles.more}>
                all {items.length} {title}
              </A>
            </header>
            <ItemList items={items.slice(0, shown)} />
          </Stack>
        </section>
      ))}
    </Stack>
  )
}
