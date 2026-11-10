import { requireAdmin } from '../../../lib/admin/require-admin'
import { logAdmin } from '../../../lib/log/admin-api'
import { QueueHeader } from '../_queue/card'
import styles from '../_queue/queue.module.css'
import { KindSettingsForm, TagTitlesForm } from './Settings'

export const dynamic = 'force-dynamic'

const NEVER_SHOWABLE = ['untriaged', 'dismissed']

export default async function SettingsPage() {
  await requireAdmin()
  const [kinds, presentation] = await Promise.all([logAdmin.kinds(), logAdmin.settings()])
  const itemsByKind = await Promise.all(kinds.map((k) => logAdmin.items({ kind: k.name })))
  const tagsInUse = [
    ...new Set([
      ...itemsByKind.flat().flatMap((item) => item.entry?.tags ?? []),
      ...Object.keys(presentation.tagTitles)
    ])
  ].sort()

  return (
    <div className={styles.queue}>
      <QueueHeader current="settings" />
      {kinds.map((kind) => (
        <KindSettingsForm
          key={kind.name}
          kind={kind.name}
          statuses={kind.statuses.filter((s) => !NEVER_SHOWABLE.includes(s))}
          axes={kind.axes}
          initial={presentation.kinds[kind.name]}
        />
      ))}
      <TagTitlesForm tags={tagsInUse} initial={presentation.tagTitles} />
    </div>
  )
}
