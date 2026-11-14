import type { Metadata } from 'next'
import { A } from '../../../components/A'
import { Gap, Stack } from '../../../components/Stack'
import { WrapInBox } from '../../../components/WrapInBox'
import { logPublic } from '../../../lib/log/public-api'
import { kindTitle, publicKinds, toListed } from '../../../lib/log/present'
import { overviewSort, sortItems, sortsFor } from './[kind]/view'
import { ItemList } from './components/ItemRow'
import styles from './components/log.module.css'

export const metadata: Metadata = {
  title: 'things as ap saw them',
  description: 'A log of games and other things Arttu Pennanen has seen.'
}

export default async function LogPage() {
  const presentation = await logPublic.presentation()
  const sections = await Promise.all(
    publicKinds(presentation).map(async (kind) => {
      const settings = presentation.kinds[kind]
      const fields = presentation.fields[kind] ?? []
      const axes = settings.axes.flatMap((key) => presentation.axes.filter((a) => a.key === key))
      const items = (await logPublic.items(kind)).map((item) => toListed(item, fields))
      const sorts = sortsFor(axes, fields, settings)
      const sort = overviewSort(items, axes, sorts, settings)
      return {
        kind,
        title: kindTitle(presentation, kind),
        items: sortItems(items, { by: sort.key, direction: 'forward' }, sorts),
        shown: settings.overview.count
      }
    })
  )
  const filled = sections.filter((section) => section.items.length > 0)

  return (
    <Stack gap={Gap.xLarge} alignItems="stretch">
      <h1>things as ap saw them</h1>
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
