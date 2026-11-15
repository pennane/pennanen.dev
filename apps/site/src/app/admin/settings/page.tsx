import { requireAdmin } from '../../../lib/admin/require-admin'
import { logAdmin } from '../../../lib/log/admin-api'
import { QueueHeader } from '../_queue/card'
import styles from '../_queue/queue.module.css'
import { AxesForm, KindOrderForm, KindSettingsForm, TagTitlesForm } from './Settings'
import { SourcesForm } from './Sources'

export const dynamic = 'force-dynamic'

const NEVER_SHOWABLE = ['untriaged']

export default async function SettingsPage({
  searchParams
}: {
  searchParams: Promise<{ spotify?: string }>
}) {
  await requireAdmin()
  const { spotify: spotifyOutcome } = await searchParams
  const [kinds, presentation, syncReports, spotifyConnection] = await Promise.all([
    logAdmin.kinds(),
    logAdmin.settings(),
    logAdmin.syncReports(),
    logAdmin.spotifyConnection()
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
      <AxesForm initial={presentation.axes} readables={presentation.readables} />
      <KindOrderForm initial={presentation.kindOrder} />
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
      <SourcesForm
        reports={syncReports}
        spotifyConnection={spotifyConnection}
        spotifyOutcome={spotifyOutcome}
      />
    </div>
  )
}
