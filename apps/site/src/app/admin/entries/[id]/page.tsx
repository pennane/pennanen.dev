import { notFound, redirect } from 'next/navigation'
import { requireAdmin } from '../../../../lib/admin/require-admin'
import { logAdmin } from '../../../../lib/log/admin-api'

export const dynamic = 'force-dynamic'

type Params = Promise<{ id: string }>

export default async function EntryPage({ params }: { params: Params }) {
  await requireAdmin()
  const id = Number((await params).id)
  if (!Number.isInteger(id)) notFound()
  const item = await logAdmin.item(id).catch(() => undefined)
  if (!item?.entry) notFound()
  redirect(`/admin/entries?${new URLSearchParams({ kind: item.kind, open: String(id) })}`)
}
