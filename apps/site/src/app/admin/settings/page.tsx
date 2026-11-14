import { requireAdmin } from '../../../lib/admin/require-admin'
import { logAdmin } from '../../../lib/log/admin-api'
import { QueueHeader } from '../_queue/card'
import styles from '../_queue/queue.module.css'
import { AxesForm, KindSettingsForm, TagTitlesForm } from './Settings'
import { SourcesForm } from './Sources'

export const dynamic = 'force-dynamic'

const NEVER_SHOWABLE = ['untriaged', 'dismissed']

export default async function SettingsPage() {
  await requireAdmin()
  const [kinds, presentation, syncReports] = await Promise.all([
    logAdmin.kinds(),
    logAdmin.settings(),
    logAdmin.syncReports()
  ])
  const entriesByKind = await Promise.all(kinds.map((k) => logAdmin.entries(k.name)))
  const tagsInUse = [
    ...new Set([
      ...entriesByKind.flat().flatMap((entry) => entry.tags),
      ...Object.keys(presentation.tagTitles)
    ])
  ].sort()

  return (
    <div className={styles.queue}>
      <QueueHeader current="settings" />
      <AxesForm initial={presentation.axes} />
      {kinds.map((kind) => (
        <KindSettingsForm
          key={kind.name}
          kind={kind.name}
          statuses={kind.statuses.filter((s) => !NEVER_SHOWABLE.includes(s))}
          axes={presentation.axes}
          rankedAxis={kind.rankedAxis}
          fields={presentation.fields[kind.name] ?? []}
          initial={presentation.kinds[kind.name]}
        />
      ))}
      <TagTitlesForm tags={tagsInUse} initial={presentation.tagTitles} />
      <SourcesForm reports={syncReports} />
    </div>
  )
}
