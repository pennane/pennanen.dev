import type { Metadata } from 'next'
import { A } from '../../../components/A'
import { Gap, Stack } from '../../../components/Stack'
import { WrapInBox } from '../../../components/WrapInBox'
import { logPublic } from '../../../lib/log/public-api'
import { byRank, kindPlural, PUBLIC_KINDS, toListed } from '../../../lib/log/present'
import { ItemList } from './components/ItemRow'
import styles from './components/log.module.css'

const SHOWN_PER_KIND = 5

export const metadata: Metadata = {
  title: 'things as ap saw them',
  description: 'A log of games and other things Arttu Pennanen has seen.'
}

export default async function LogPage() {
  const sections = await Promise.all(
    PUBLIC_KINDS.map(async (kind) => ({
      kind,
      items: (await logPublic.items(kind)).map(toListed).sort(byRank)
    }))
  )
  const filled = sections.filter((section) => section.items.length > 0)

  return (
    <Stack gap={Gap.xLarge} alignItems="stretch">
      <h1>things as ap saw them</h1>
      {filled.length === 0 && <p className={styles.empty}>Nothing here yet.</p>}
      {filled.map(({ kind, items }) => (
        <section key={kind}>
          <Stack gap={Gap.large} alignItems="stretch">
            <header className={styles.sectionHeader}>
              <WrapInBox>
                <h3>{kindPlural(kind)}</h3>
              </WrapInBox>
              <A href={`/log/${kind}`} className={styles.more}>
                all {items.length} {kindPlural(kind)}
              </A>
            </header>
            <ItemList items={items.slice(0, SHOWN_PER_KIND)} />
          </Stack>
        </section>
      ))}
    </Stack>
  )
}
