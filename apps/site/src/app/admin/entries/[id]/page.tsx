import { notFound } from 'next/navigation'
import { requireAdmin } from '../../../../lib/admin/require-admin'
import { logAdmin } from '../../../../lib/log/admin-api'
import { textField } from '../../../../lib/log/present'
import { Editor } from './Editor'

export const dynamic = 'force-dynamic'

type Params = Promise<{ id: string }>

export default async function EntryPage({ params }: { params: Params }) {
  await requireAdmin()
  const id = Number((await params).id)
  if (!Number.isInteger(id)) notFound()
  const [kinds, item] = await Promise.all([
    logAdmin.kinds(),
    logAdmin.item(id).catch(() => undefined)
  ])
  const kind = item && kinds.find((k) => k.name === item.kind)
  if (!item?.entry || !kind) notFound()

  return (
    <Editor
      key={item.id}
      itemId={item.id}
      title={textField(item.verdicts.values, 'title') ?? item.slug}
      cover={
        textField(item.verdicts.values, 'cover') ?? textField(item.verdicts.values, 'image')
      }
      publicPath={`/log/${item.kind}/${item.slug}`}
      kind={kind}
      entry={item.entry}
      moments={item.moments ?? []}
    />
  )
}
